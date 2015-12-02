package shaping

import (
	"fmt"
	"math"
	"net"
	"os/exec"
	"strings"
	"syscall"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/facebook/augmented-traffic-control/src/atc_thrift"
	"github.com/vishvananda/netlink"
)

var FILTER_IP_TYPE = []uint16{syscall.ETH_P_IP, syscall.ETH_P_IPV6}

var (
	// location of the iptables binaries
	IPTABLES  string
	IP6TABLES string

	// Names of the wan and lan interfaces (e.g. eth0, enp6s0)
	WAN_INT string
	LAN_INT string
)

/*
Sets up platform-specific flags for the shaper.
*/
func ShapingFlags() {
	kingpin.Flag("iptables", "location of the iptables binary").StringVar(&IPTABLES)
	kingpin.Flag("ip6tables", "location of the ip6tables binary").StringVar(&IP6TABLES)
	kingpin.Flag("wan", "name of the WAN interface").StringVar(&WAN_INT)
	kingpin.Flag("lan", "name of the LAN interface").StringVar(&LAN_INT)
}

/*
Returns a shaper suitable for the current platform.
This build of ATC is compiled with iptables support and only works on linux.
*/
func GetShaper() (Shaper, error) {
	// Make sure that the location of the iptables binaries are set
	// If they're not, pull them from $PATH
	if IPTABLES == "" {
		var err error
		IPTABLES, err = exec.LookPath("iptables")
		if err != nil {
			return nil, err
		}
	}
	if IP6TABLES == "" {
		var err error
		IP6TABLES, err = exec.LookPath("ip6tables")
		if err != nil {
			return nil, err
		}
	}
	return &netlinkShaper{}, nil
}

/*
The netlink shaper uses a combination of iptables and tc to achieve bandwidth traffic shaping.

Each group gets a unique identifier (provided when the group is created)
*/
type netlinkShaper struct{}

func (nl *netlinkShaper) GetPlatform() atc_thrift.PlatformType {
	return atc_thrift.PlatformType_LINUX
}

func (nl *netlinkShaper) Initialize() error {
	if WAN_INT == "eth0" && LAN_INT == "eth1" {
		Log.Println("-wan and -lan were not provided. Using defaults. This is probably not what you want!")
	}
	// Clean out mangle's FORWARD chain. There might be remaining
	// rules in here from past atcd instances.
	if err := mangle_flush(); err != nil {
		return err
	}
	// Setup the root qdisc
	wan, lan, err := lookupInterfaces()
	if err != nil {
		return err
	}
	if err := setupRootQdisc(wan); err != nil {
		return err
	}
	return setupRootQdisc(lan)
}

/*
Create a group. The ID used here is assumed to be unique to this group and won't change.
*/
func (nl *netlinkShaper) CreateGroup(id int64, member net.IP) error {
	return mark_packets_for(member, fmt.Sprintf("0x%x", id))
}

func (nl *netlinkShaper) JoinGroup(id int64, member net.IP) error {
	return mark_packets_for(member, fmt.Sprintf("0x%x", id))
}

func (nl *netlinkShaper) LeaveGroup(id int64, member net.IP) error {
	return remove_marking_for(member, fmt.Sprintf("0x%x", id))
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
	// Add class: (contains rate)
	//class htb 1:2 root leaf 8005: prio 0 rate 4194Mbit ceil 4194Mbit burst 1048b cburst 1048b

	// Rate is a required argument for HTB class. If we are given a value of 0,
	// rate was not set and is considered unlimited.
	// In that case, let set the rate as high as we can.
	// Note: currently, it is implemented as a u32 by the netlink library.
	rate := uint64(shaping.GetRate() * 1000)
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
	if err := netlink.ClassAdd(htbc); err != nil {
		return fmt.Errorf("Could not create htb class: %v", err)
	}

	// Add filter:
	// filter parent 1: protocol ip pref 1 fw handle 0x2 classid 1:2  police 0x5
	//     rate 4194Mbit burst 11010b mtu 2Kb action drop overhead 0b ref 1 bind 1
	// filters packets with mark 0x2 to classid 1:2
	// We need to add the filter for both IPv4 and IPv6
	for idx, proto := range FILTER_IP_TYPE {
		fw, err := netlink.NewFw(netlink.FilterAttrs{
			LinkIndex: link.Attrs().Index,
			Parent:    netlink.MakeHandle(1, 0),
			Handle:    uint32(id),
			Protocol:  proto,
			Priority:  uint16(idx + 1),
		}, netlink.FilterFwAttrs{
			ClassId: htbc.Attrs().Handle,
		})
		if err != nil {
			return fmt.Errorf("Could not create fw filter struct: %v", err)
		}
		if err := netlink.FilterAdd(fw); err != nil {
			return fmt.Errorf("Could not create fw filter: %v", err)
		}
	}

	// Add netem qdisc: (contains latency, packet drop, correlation, etc.)
	//qdisc netem 8001: parent 1:2 limit 1000
	htbq := netlink.NewNetem(netlink.QdiscAttrs{
		LinkIndex: link.Attrs().Index,
		// We can leave netlink assigning a handle for us
		// Handle:    netlink.MakeHandle(uint16(id+0x8000), 0),
		Parent: netlink.MakeHandle(1, uint16(id)),
	}, netlink.NetemQdiscAttrs{
		Latency:     uint32(shaping.GetDelay().Delay * 1000),   // in ms
		Jitter:      uint32(shaping.GetDelay().Jitter * 1000),  // in ms
		DelayCorr:   float32(shaping.GetDelay().Correlation),   // in %
		Loss:        float32(shaping.GetLoss().Percentage),     // in %
		LossCorr:    float32(shaping.GetLoss().Correlation),    // in %
		ReorderProb: float32(shaping.GetReorder().Percentage),  // in %
		ReorderCorr: float32(shaping.GetReorder().Correlation), // in %
		Gap:         uint32(shaping.GetReorder().Gap),
		CorruptProb: float32(shaping.GetCorruption().Percentage),  // in %
		CorruptCorr: float32(shaping.GetCorruption().Correlation), // in %
	})
	if err := netlink.QdiscAdd(htbq); err != nil {
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
		if err := netlink.FilterDel(fw); err != nil {
			return fmt.Errorf("Could not create fw filter: %v", err)
		}
	}

	if err := netlink.ClassDel(htbc); err != nil {
		return fmt.Errorf("Could not delete htb class: %v", err)
	}

	return nil
}

func mark_packets_for(net net.IP, mark string) error {
	if err := mark_int_packets_for("-d", net, WAN_INT, mark); err != nil {
		return err
	}
	return mark_int_packets_for("-s", net, LAN_INT, mark)
}

func mark_int_packets_for(flag string, net net.IP, int, mark string) error {
	if err := mangle_append(flag, net, "-i", int, "-j", "MARK", "--set-xmark", mark); err != nil {
		return fmt.Errorf("Could not mark packets for %s: %v", net, err)
	}
	return nil
}

func remove_marking_for(net net.IP, mark string) error {
	if err := remove_int_marking_for("-d", net, WAN_INT, mark); err != nil {
		return err
	}
	return remove_int_marking_for("-s", net, LAN_INT, mark)
}

func remove_int_marking_for(flag string, net net.IP, int, mark string) error {
	if err := mangle_delete(flag, net, "-i", int, "-j", "MARK", "--set-xmark", mark); err != nil {
		return fmt.Errorf("Could not remove marking for %s: %v", net, err)
	}
	return nil
}

func mangle_append(flag string, addr net.IP, args ...string) error {
	ipt := IPTABLES
	if addr.To4() == nil {
		ipt = IP6TABLES
	}
	args = append([]string{"-t", "mangle", "-A", "FORWARD", flag, addr.String()}, args...)
	return run_cmd(ipt, args...)
}

func mangle_delete(flag string, addr net.IP, args ...string) error {
	ipt := IPTABLES
	if addr.To4() == nil {
		ipt = IP6TABLES
	}
	args = append([]string{"-t", "mangle", "-D", "FORWARD", flag, addr.String()}, args...)
	return run_cmd(ipt, args...)
}

func mangle_flush() error {
	if err := run_cmd(IPTABLES, "-t", "mangle", "-F", "FORWARD"); err != nil {
		return err
	}
	return run_cmd(IP6TABLES, "-t", "mangle", "-F", "FORWARD")
}

func run_cmd(cmd string, args ...string) error {
	Log.Printf("Command: %s %s\n", cmd, strings.Join(args, " "))
	bytes, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		for _, s := range strings.Split(string(bytes), "\n") {
			if len(s) > 0 {
				Log.Println(s)
			}
		}
	}
	return err
}

func lookupInterfaces() (wan, lan netlink.Link, err error) {
	wan, err = netlink.LinkByName(LAN_INT)
	if err != nil {
		err = fmt.Errorf("Could not find lan(%s) interface: %v", LAN_INT, err)
		return
	}
	lan, err = netlink.LinkByName(WAN_INT)
	if err != nil {
		err = fmt.Errorf("Could not find wan(%s) interface: %v", WAN_INT, err)
		return
	}
	return
}

func setupRootQdisc(link netlink.Link) error {
	// Clean out old qdiscs
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

	if err := netlink.QdiscAdd(root_qdisc); err != nil {
		return fmt.Errorf("Could not create root qdisc (%s): %v", link.Attrs().Name, err)
	}
	return nil
}
