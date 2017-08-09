package shaping

import (
	"fmt"
	"math"
	"os/exec"
	"syscall"

	"github.com/facebook/augmented-traffic-control/src/atc_thrift"
	"github.com/facebook/augmented-traffic-control/src/iptables"
	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"
	"gopkg.in/alecthomas/kingpin.v2"
)

var FILTER_IP_TYPE = []uint16{syscall.ETH_P_IP, syscall.ETH_P_IPV6}

var (
	// location of the iptables binaries
	IPTABLES  string
	IP6TABLES string

	// Names of the wan and lan interfaces (e.g. eth0, enp6s0)
	WAN_INT string
	LAN_INT string

	DONT_DROP_PACKETS bool
)

/*
Sets up platform-specific flags for the shaper.
*/
func ShapingFlags() {
	kingpin.Flag("iptables", "location of the iptables binary").StringVar(&IPTABLES)
	kingpin.Flag("ip6tables", "location of the ip6tables binary").StringVar(&IP6TABLES)
	kingpin.Flag("wan", "name of the WAN interface").StringVar(&WAN_INT)
	kingpin.Flag("lan", "name of the LAN interface").StringVar(&LAN_INT)
	kingpin.Flag("dont-drop-packets", "Buffer packets that overflow the queue instead of dropping them").BoolVar(&DONT_DROP_PACKETS)
}

/*
Returns a shaper suitable for the current platform.
This build of ATC is compiled with iptables support and only works on linux.
*/
func GetShaper() (Shaper, error) {
	var err error
	// Make sure that the location of the iptables binaries are set
	// If they're not, pull them from $PATH
	if IPTABLES == "" {
		IPTABLES, err = exec.LookPath("iptables")
		if err != nil {
			return nil, err
		}
	}
	if IP6TABLES == "" {
		IP6TABLES, err = exec.LookPath("ip6tables")
		if err != nil {
			return nil, err
		}
	}
	Log.Debugf("Using iptables binary: %q\n", IPTABLES)
	Log.Debugf("Using ip6tables binary: %q\n", IP6TABLES)
	ip4t := iptables.New(IPTABLES)
	ip6t := iptables.New(IP6TABLES)
	return &netlinkShaper{ip4t, ip6t}, nil
}

/*
The netlink shaper uses a combination of iptables and tc to achieve bandwidth traffic shaping.

Each group gets a unique identifier (provided when the group is created)
*/
type netlinkShaper struct {
	ip4t *iptables.IPTables
	ip6t *iptables.IPTables
}

func (nl *netlinkShaper) GetPlatform() atc_thrift.PlatformType {
	return atc_thrift.PlatformType_LINUX
}

func (nl *netlinkShaper) Initialize() error {
	Log.Debugln("Initializing shaper using netlink")
	if WAN_INT == "eth0" && LAN_INT == "eth1" {
		Log.Println("-wan and -lan were not provided. Using defaults. This is probably not what you want!")
	}
	// Clean out mangle's FORWARD chain. There might be remaining
	// rules in here from past atcd instances.
	if err := nl.flush(); err != nil {
		return err
	}
	// Setup the root qdisc
	wan, lan, err := lookupInterfaces()
	if err != nil {
		return err
	}
	Log.Debugf("WAN interface: %s(%d)\n", wan.Attrs().Name, wan.Attrs().Index)
	Log.Debugf("LAN interface: %s(%d)\n", lan.Attrs().Name, lan.Attrs().Index)
	if err := setupRootQdisc(wan); err != nil {
		return err
	}
	if err := setupRootQdisc(lan); err != nil {
		return err
	}
	Log.Debugf("Done initializing netlink shaper")
	return nil
}

/*
Create a group. The ID used here is assumed to be unique to this group and won't change.
*/
func (nl *netlinkShaper) CreateGroup(id int64, member Target) (err error) {
	err = nl.mark_packets_for(member, id)
	if err != nil {
		err = errors.Wrapf(err, "failed to mark packets for group")
	}
	return
}

func (nl *netlinkShaper) JoinGroup(id int64, member Target) (err error) {
	err = nl.mark_packets_for(member, id)
	if err != nil {
		err = errors.Wrapf(err, "failed to mark packets for group")
	}
	return
}

func (nl *netlinkShaper) LeaveGroup(id int64, member Target) (err error) {
	err = nl.remove_marking_for(member, id)
	if err != nil {
		err = errors.Wrapf(err, "failed to remove marking for group")
	}
	return
}

func (nl *netlinkShaper) DeleteGroup(id int64) error {
	// This is a noop
	return nil
}

func (nl *netlinkShaper) Shape(id int64, shaping *atc_thrift.Shaping) error {
	wan, lan, err := lookupInterfaces()
	if err != nil {
		return err
	}
	// Unshape before shaping. (FIXME: #173)
	if err := shape_off(id, lan); err != nil {
		Log.Printf("Could not temporarily unshape lan(%s) interface: %v", LAN_INT, err)
	}
	if err := shape_off(id, wan); err != nil {
		Log.Printf("Could not temporarily unshape wan(%s) interface: %v", WAN_INT, err)
	}
	// Shape on the OUTBOUND side.
	// Traffic on the lan interface is incoming, so down.
	if err := shape_on(id, shaping.Down, lan); err != nil {
		return fmt.Errorf("Could not shape lan(%s) interface: %v", LAN_INT, err)
	}
	// Traffic on the wan interface is outgoing, so up.
	if err := shape_on(id, shaping.Up, wan); err != nil {
		return fmt.Errorf("Could not shape wan(%s) interface: %v", WAN_INT, err)
	}
	return nil
}

func (nl *netlinkShaper) Unshape(id int64) error {
	wan, lan, err := lookupInterfaces()
	if err != nil {
		return err
	}
	if err := shape_off(id, lan); err != nil {
		return fmt.Errorf("Could not unshape lan(%s) interface: %v", LAN_INT, err)
	}
	if err := shape_off(id, wan); err != nil {
		return fmt.Errorf("Could not unshape wan(%s) interface: %v", WAN_INT, err)
	}
	return nil
}

func shape_on(id int64, shaping *atc_thrift.LinkShaping, link netlink.Link) error {
	// Rate is a required argument for HTB class. If we are given a value of 0,
	// rate was not set and is considered unlimited.
	// In that case, let set the rate as high as we can.
	// Note: currently, it is implemented as a u32 by the netlink library.
	rate := uint64(shaping.Rate * 1000)
	if rate == 0 {
		// rate is given in bps but under the hood it expect Bps. Multiply by 8
		// because the netlink lib will divide it later on. That allows us to
		// provide up to 34 Gbit of traffic.
		rate = math.MaxUint32 * 8
	}
	htbc := netlink.NewHtbClass(netlink.ClassAttrs{
		LinkIndex: link.Attrs().Index,
		Handle:    netlink.MakeHandle(1, uint16(id)),
		Parent:    netlink.MakeHandle(1, 0),
	}, netlink.HtbClassAttrs{
		Rate: rate, // in kbps
		Ceil: rate,
	})
	Log.Debugf("Adding htb class: %#v\n", htbc)
	if err := netlink.ClassAdd(htbc); err != nil {
		return fmt.Errorf("Could not create htb class: %v", err)
	}

	// Add filter:
	// filter parent 1: protocol ip pref 1 fw handle 0x2 classid 1:2  police 0x5
	//     rate 4194Mbit burst 11010b mtu 2Kb action drop overhead 0b ref 1 bind 1
	// filters packets with mark 0x2 to classid 1:2
	// We need to add the filter for both IPv4 and IPv6
	action := netlink.TC_POLICE_SHOT
	if DONT_DROP_PACKETS {
		action = netlink.TC_POLICE_OK
	}
	for idx, proto := range FILTER_IP_TYPE {
		fw, err := netlink.NewFw(netlink.FilterAttrs{
			LinkIndex: link.Attrs().Index,
			Parent:    netlink.MakeHandle(1, 0),
			Handle:    uint32(id),
			Protocol:  proto,
			Priority:  uint16(idx + 1),
		}, netlink.FilterFwAttrs{
			ClassId:  htbc.Attrs().Handle,
			Rate:     uint32(rate),
			PeakRate: uint32(rate),
			Buffer:   calculateBufferSize(uint32(shaping.Rate), uint32(shaping.Delay.Delay * 1000)),
			Action:   action,
		})
		if err != nil {
			return fmt.Errorf("Could not create fw filter struct: %v", err)
		}
		Log.Debugf("Adding fw filter: %#v", fw)
		if err := netlink.FilterAdd(fw); err != nil {
			return fmt.Errorf("Could not create fw filter: %v", err)
		}
	}

	// Add netem qdisc: (contains latency, packet drop, correlation, etc.)
	//qdisc netem 8001: parent 1:2 limit 1000
	netem := netlink.NewNetem(netlink.QdiscAttrs{
		LinkIndex: link.Attrs().Index,
		// We can leave netlink assigning a handle for us
		// Handle:    netlink.MakeHandle(uint16(id+0x8000), 0),
		Parent: netlink.MakeHandle(1, uint16(id)),
	}, netlink.NetemQdiscAttrs{
		Latency:     uint32(shaping.Delay.Delay * 1000),   // in ms
		Jitter:      uint32(shaping.Delay.Jitter * 1000),  // in ms
		DelayCorr:   float32(shaping.Delay.Correlation),   // in %
		Loss:        float32(shaping.Loss.Percentage),     // in %
		LossCorr:    float32(shaping.Loss.Correlation),    // in %
		ReorderProb: float32(shaping.Reorder.Percentage),  // in %
		ReorderCorr: float32(shaping.Reorder.Correlation), // in %
		Gap:         uint32(shaping.Reorder.Gap),
		CorruptProb: float32(shaping.Corruption.Percentage),  // in %
		CorruptCorr: float32(shaping.Corruption.Correlation), // in %
	})
	Log.Debugf("Adding netem qdisc: %#v", netem)
	if err := netlink.QdiscAdd(netem); err != nil {
		return fmt.Errorf("Could not create htb qdisc: %v", err)
	}

	return nil
}

func shape_off(id int64, link netlink.Link) error {
	htbc := netlink.NewHtbClass(netlink.ClassAttrs{
		LinkIndex: link.Attrs().Index,
		Handle:    netlink.MakeHandle(1, uint16(id)),
		Parent:    netlink.MakeHandle(1, 0),
	}, netlink.HtbClassAttrs{})

	for idx, proto := range FILTER_IP_TYPE {
		fw, err := netlink.NewFw(netlink.FilterAttrs{
			LinkIndex: link.Attrs().Index,
			Parent:    netlink.MakeHandle(1, 0),
			Handle:    uint32(id),
			Protocol:  proto,
			Priority:  uint16(idx + 1),
		}, netlink.FilterFwAttrs{
			ClassId: htbc.Attrs().Handle, // This is not needed really
		})
		if err != nil {
			return fmt.Errorf("Could not create fw filter struct: %v", err)
		}
		Log.Debugf("Deleting fw filter: %#v\n", fw)
		if err := netlink.FilterDel(fw); err != nil {
			return fmt.Errorf("Could not delete fw filter: %v", err)
		}
	}

	Log.Debugf("Deleting htb class: %#v\n", htbc)
	if err := netlink.ClassDel(htbc); err != nil {
		return fmt.Errorf("Could not delete htb class: %v", err)
	}

	return nil
}

func (nl *netlinkShaper) mark_packets_for(target Target, mark int64) (err error) {
	ipt := nl.tablesFor(target)
	chain := ipt.Table("mangle").Chain("FORWARD")
	if err = chain.Append(iptables.Rule{Destination: target, In: WAN_INT}.SetMark(mark)); err != nil {
		err = errors.Wrapf(err, "Could not mark packets for %s: %v", target)
		return
	}
	if err = chain.Append(iptables.Rule{Source: target, In: LAN_INT}.SetMark(mark)); err != nil {
		err = errors.Wrapf(err, "Could not mark packets for %s", target)
		return
	}
	return
}

func (nl *netlinkShaper) remove_marking_for(target Target, mark int64) error {
	ipt := nl.tablesFor(target)
	chain := ipt.Table("mangle").Chain("FORWARD")
	if err := chain.Delete(iptables.Rule{Destination: target, In: WAN_INT}.SetMark(mark)); err != nil {
		return fmt.Errorf("Could not mark packets for %s: %v", target, err)
	}
	if err := chain.Delete(iptables.Rule{Source: target, In: LAN_INT}.SetMark(mark)); err != nil {
		return fmt.Errorf("Could not mark packets for %s: %v", target, err)
	}
	return nil
}

func (nl *netlinkShaper) tablesFor(target Target) *iptables.IPTables {
	if target.V6() {
		return nl.ip6t
	} else {
		return nl.ip4t
	}
}

func (nl *netlinkShaper) flush() error {
	if err := nl.ip4t.Table("mangle").Chain("FORWARD").Flush(); err != nil {
		return err
	}
	return nl.ip6t.Table("mangle").Chain("FORWARD").Flush()
}

func lookupInterfaces() (wan, lan netlink.Link, err error) {
	lan, err = netlink.LinkByName(LAN_INT)
	if err != nil {
		err = fmt.Errorf("Could not find lan(%s) interface: %v", LAN_INT, err)
		return
	}
	wan, err = netlink.LinkByName(WAN_INT)
	if err != nil {
		err = fmt.Errorf("Could not find wan(%s) interface: %v", WAN_INT, err)
		return
	}
	return
}

func setupRootQdisc(link netlink.Link) error {
	// Clean out old qdiscs
	Log.Debugf("Cleaning out old qdiscs on %s\n", link.Attrs().Name)
	qdiscs, err := netlink.QdiscList(link)
	if err != nil {
		return err
	}
	for _, q := range qdiscs {
		if err := netlink.QdiscDel(q); err != nil {
			Log.Printf("warning: Could not delete root qdisc (%s): %v\n", link.Attrs().Name, err)
		}
	}

	// Setup new HTB qdisc as root
	root_qdisc := netlink.NewHtb(netlink.QdiscAttrs{
		LinkIndex: link.Attrs().Index,
		Parent:    netlink.HANDLE_ROOT,
		Handle:    netlink.MakeHandle(1, 0),
	})
	Log.Debugf("Creating new root qdisc on %s: %#v\n", link.Attrs().Name, root_qdisc)

	if err := netlink.QdiscAdd(root_qdisc); err != nil {
		return fmt.Errorf("Could not create root qdisc (%s): %v", link.Attrs().Name, err)
	}
	return nil
}

func calculateBufferSize(rate uint32, latency uint32) uint32 {
	if rate == 0 || latency == 0 {
		return 12000 //FIXME: what to do when our rate or latency is 0?
	}
	bufsize := (2 * latency * rate) / 1000;
	Log.Debugf("Buffer size 2 * %d * %d / 1000 => %d", latency, rate, bufsize);
	return bufsize
}
