package daemon

import (
	"flag"
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/facebook/augmented-traffic-control/src/atc_thrift"
	"github.com/vishvananda/netlink"
)

var (
	// location of the iptables binaries
	IPTABLES  string
	IP6TABLES string

	// Names of the wan and lan interfaces
	WAN_INT string
	LAN_INT string
)

// Sets up platform-specific flags for the shaper.
func ShapingFlags() {
	flag.StringVar(&IPTABLES, "iptables", "", "iptables binary location")
	flag.StringVar(&IP6TABLES, "ip6tables", "", "ip6tables binary location")
	flag.StringVar(&WAN_INT, "wan", "eth0", "ATCD WAN interface")
	flag.StringVar(&LAN_INT, "lan", "eth1", "ATCD LAN interface")
}

/*
Returns a shaper suitable for the current platform.
(in this case, the shaper uses iptables)
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

// *netlinkShaper implements Shaper
type netlinkShaper struct{}

func (nl *netlinkShaper) GetPlatform() atc_thrift.PlatformType {
	return atc_thrift.PlatformType_LINUX
}

func (nl *netlinkShaper) Initialize() error {
	// Clean out mangle's FORWARD chain. There might be remaining
	// rules in here from past atcd instances.
	if err := mangle_flush(); err != nil {
		return err
	}
	// Setup the root qdisc
	if err := setupRootQdisc(WAN_INT); err != nil {
		return err
	}
	return setupRootQdisc(LAN_INT)
}

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
	return fmt.Errorf("netlink shaping is not implemented")
}

func (nl *netlinkShaper) Unshape(int64) error {
	return fmt.Errorf("netlink shaping is not implemented")
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

func setupRootQdisc(link_name string) error {
	link, err := netlink.LinkByName(link_name)
	if err != nil {
		return err
	}

	// Clean out old qdiscs
	qdiscs, err := netlink.QdiscList(link)
	if err != nil {
		return err
	}
	for _, q := range qdiscs {
		if err := netlink.QdiscDel(q); err != nil {
			return fmt.Errorf("Could not delete qdisc: %v", err)
		}
	}

	// Setup new HTB qdisc as root
	root_qdisc := netlink.NewHtb(netlink.QdiscAttrs{
		LinkIndex: link.Attrs().Index,
		Parent:    netlink.HANDLE_ROOT,
		//Handle:    0x1,
		Handle: 0x10000,
	})

	return netlink.QdiscAdd(root_qdisc)
}
