package shaping

import (
	crand "crypto/rand"
	"fmt"
	"math"
	mrand "math/rand"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"

	assertlib "github.com/alecthomas/assert"
	"github.com/facebook/augmented-traffic-control/src/atc_thrift"
	"github.com/facebook/augmented-traffic-control/src/iptables"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

func init() {
	IPTABLES = "iptables"
	IP6TABLES = "ip6tables"
}

func TestCreatesRootQdisc(t *testing.T) {
	tearDown, link := setUpNetlinkTest(t)
	defer tearDown()

	// Setup root qdisc.
	check(t, setupRootQdisc(link), "couldn't create root qdisc")
	// Check that the root qdisc exists.
	qdiscs, err := netlink.QdiscList(link)
	check(t, err, "couldn't list qdiscs")

	if testing.Verbose() {
		test_cmd(t, "tc", "qdisc", "show", "dev", link.Attrs().Name)
		test_cmd(t, "tc", "class", "show", "dev", link.Attrs().Name)
		test_cmd(t, "tc", "filter", "show", "dev", link.Attrs().Name)
	}

	// FIXME Better asserts

	if len(qdiscs) != 1 {
		t.Fatal("Failed to add qdisc")
	}
	_, ok := qdiscs[0].(*netlink.Htb)
	if !ok {
		t.Fatal("Qdisc is the wrong type")
	}
}

func TestShapeOn(t *testing.T) {
	tearDown, link := setUpNetlinkTest(t)
	defer tearDown()

	check(t, setupRootQdisc(link), "couldn't create root qdisc")

	// Set up class + filters (ipv4/ipv6) using shape_on
	mark := int64(5)
	shaping := &atc_thrift.LinkShaping{
		Rate:       10,
		Delay:      &atc_thrift.Delay{Delay: 10, Jitter: 15, Correlation: 1},
		Reorder:    &atc_thrift.Reorder{Percentage: 6, Gap: 1, Correlation: 0.6},
		Loss:       &atc_thrift.Loss{Percentage: 10, Correlation: 6.3},
		Corruption: &atc_thrift.Corruption{Percentage: 3.4},
	}
	check(t, shape_on(mark, shaping, link), "could not enable shaping")

	classes, err := netlink.ClassList(
		link, netlink.MakeHandle(0x1, uint16(mark)),
	)
	check(t, err, "couldn't list classes")
	if len(classes) != 1 {
		t.Fatal("Failed to add class")
	}

	filters, err := netlink.FilterList(link, netlink.MakeHandle(0x1, 0))
	check(t, err, "couldn't list filters")

	if testing.Verbose() {
		test_cmd(t, "tc", "qdisc", "show", "dev", link.Attrs().Name)
		test_cmd(t, "tc", "class", "show", "dev", link.Attrs().Name)
		test_cmd(t, "tc", "filter", "show", "dev", link.Attrs().Name)
	}

	// FIXME Better asserts

	// We expect 2 filters to be setup: 1 for ipv4, 1 for ipv6.
	if len(filters) != 2 {
		t.Fatal("Failed to add filter")
	}
}

func TestShapeRate0(t *testing.T) {
	tearDown, link := setUpNetlinkTest(t)
	defer tearDown()

	check(t, setupRootQdisc(link), "couldn't create root qdisc")

	// Set up class + filters (ipv4/ipv6) using shape_on
	mark := int64(5)
	shaping := &atc_thrift.LinkShaping{}
	check(t, shape_on(mark, shaping, link), "could not enable shaping")

	classes, err := netlink.ClassList(
		link, netlink.MakeHandle(0x1, uint16(mark)),
	)
	check(t, err, "couldn't list classes")
	if len(classes) != 1 {
		t.Fatal("Failed to add class")
	}
	class := classes[0].(*netlink.HtbClass)

	// When a rate of 0 is given, we default to not limiting the traffic by
	// allocating the biggest rate we can. (Currently this is u32 only until
	// the netlink implementation is supporting 64 bits).

	if testing.Verbose() {
		test_cmd(t, "tc", "qdisc", "show", "dev", link.Attrs().Name)
		test_cmd(t, "tc", "class", "show", "dev", link.Attrs().Name)
		test_cmd(t, "tc", "filter", "show", "dev", link.Attrs().Name)
	}
	if class.Rate != math.MaxUint32 {
		t.Fatal("Failed to set unlimited rate.")
	}
}

func TestShapeOff(t *testing.T) {
	tearDown, link := setUpNetlinkTest(t)
	defer tearDown()

	check(t, setupRootQdisc(link), "couldn't create root qdisc")

	// Set up class + filters (ipv4/ipv6) using shape_on
	mark := int64(5)
	shaping := &atc_thrift.LinkShaping{
		Rate:  10,
		Delay: &atc_thrift.Delay{Delay: 10},
		Loss:  &atc_thrift.Loss{Percentage: 10},
	}
	check(t, shape_on(mark, shaping, link), "could not enable shaping")
	check(t, shape_off(mark, link), "could not disable shaping")

	filters, err := netlink.FilterList(link, netlink.MakeHandle(0x1, 0))
	check(t, err, "could not list filters")

	if testing.Verbose() {
		test_cmd(t, "tc", "qdisc", "show", "dev", link.Attrs().Name)
		test_cmd(t, "tc", "class", "show", "dev", link.Attrs().Name)
		test_cmd(t, "tc", "filter", "show", "dev", link.Attrs().Name)
	}

	// FIXME Better asserts

	// We expect 0 filters to be setup
	if len(filters) != 0 {
		t.Fatal("Failed to delete filter")
	}
}

func TestShapeTwice(t *testing.T) {
	assert, tearDown, shaper := setUpShaperTest(t)
	defer tearDown()

	// Set up class + filters (ipv4/ipv6) using shape_on
	mark := int64(7)
	shaping := &atc_thrift.Shaping{
		Up: &atc_thrift.LinkShaping{
			Rate:  10,
			Delay: &atc_thrift.Delay{Delay: 10},
			Loss:  &atc_thrift.Loss{Percentage: 10},
		},
		Down: &atc_thrift.LinkShaping{
			Rate:  10,
			Delay: &atc_thrift.Delay{Delay: 10},
			Loss:  &atc_thrift.Loss{Percentage: 10},
		},
	}

	assert.NoError(shaper.Shape(mark, shaping), "could not shape")
	shaping.Up.Rate = 100
	shaping.Down.Rate = 100
	assert.NoError(shaper.Shape(mark, shaping), "could not reshape")

	if testing.Verbose() {
		for _, s := range []string{"wan", "lan"} {
			test_cmd(t, "tc", "qdisc", "show", "dev", s)
			test_cmd(t, "tc", "class", "show", "dev", s)
			test_cmd(t, "tc", "filter", "show", "dev", s)
		}
	}

	// FIXME Better asserts
}

func TestGroupCreateJoin(t *testing.T) {
	// Do this all in the same network namespace so all the groups exist at once.
	assert, tearDown, shaper := setUpShaperTest(t)
	defer tearDown()

	// Test most common group setups
	testGroupCreateJoin(assert, shaper, ip4)
	testGroupCreateJoin(assert, shaper, ip4, ip4)
	testGroupCreateJoin(assert, shaper, ip6)
	testGroupCreateJoin(assert, shaper, ip6, ip6)
	testGroupCreateJoin(assert, shaper, ip4, ip6)
	testGroupCreateJoin(assert, shaper, ip6, ip4)

	// Test some random setups.
	// 10 IPs in each group, between 0 and 10 IPv4 addresses per group
	for f := 0; f <= 10; f += 1 {
		testGroupCreateJoin(assert, shaper, randIPGens(10, f)...)
	}
}

/**
*** Testing Utilities
**/

func randIPGens(n, v4 int) []func() iptables.Target {
	v := mrand.Perm(n)
	gens := make([]func() iptables.Target, n)
	for i := 0; i < n; i++ {
		if v[i] < v4 {
			gens[i] = ip4
		} else {
			gens[i] = ip6
		}
	}
	return gens
}

func testGroupCreateJoin(assert *assertlib.Assertions, shaper *netlinkShaper, gens ...func() iptables.Target) {
	mark := int64(mrand.Int31())

	targets := make([]iptables.Target, 0, len(gens))
	for _, g := range gens {
		// hope there aren't collisions
		targets = append(targets, g())
	}

	for i, t := range targets {
		if i == 0 {
			err := shaper.CreateGroup(mark, t)
			assert.NoError(err, "could not create group with %v", t)
		} else {
			err := shaper.JoinGroup(mark, t)
			assert.NoError(err, "%v could not join group 0x%x", t, mark)
		}
	}

	for _, t := range targets {
		assertPacketsMarked(assert, shaper, mark, t)
	}
}

func ip6() iptables.Target {
	return randIP(net.IPv6len)
}

func ip4() iptables.Target {
	return randIP(net.IPv4len)
}

func randIP(l int) iptables.Target {
	b := make([]byte, l)
	_, err := crand.Read(b)
	if err != nil {
		panic(err)
	}
	return iptables.IPTarget(net.IP(b))
}

func assertPacketsMarked(assert *assertlib.Assertions, shaper *netlinkShaper, id int64, target iptables.Target) {
	t := shaper.ip4t
	world := "0.0.0.0/0"
	if target.V6() {
		t = shaper.ip6t
		world = "::/0"
	}

	markings, err := t.Table("mangle").Chain("FORWARD").GetRules(target)
	assert.NoError(err)
	assert.Len(markings, 2, "wrong number of iptables rules for %v", target)

	for _, mark := range markings {
		switch mark.In {
		case "wan":
			assert.Equal(world, mark.Source.String())
			assert.Equal(target, mark.Destination)
		case "lan":
			assert.Equal(target, mark.Source)
			assert.Equal(world, mark.Destination.String())
		default:
			assert.Fail("Mark has the wrong interface: %v", mark.In)
		}
		assert.Equal([]string{"MARK", "set", fmt.Sprintf("0x%x", id)}, mark.Args)
	}
}

func setUpShaperTest(t *testing.T) (*assertlib.Assertions, func(), *netlinkShaper) {
	if os.Getuid() != 0 {
		t.Skip("Skipped test because it requires root privileges")
	}
	assert := assertlib.New(t)

	// new temporary namespace so we don't pollute the host
	// lock thread since the namespace is thread local
	runtime.LockOSThread()
	ns, err := netns.New()
	if err != nil {
		runtime.UnlockOSThread()
		t.Fatalf("Failed to create new network namespace: %v", err)
	}

	link := setUpDummyInterface(t, "wan")
	assert.NoError(setupRootQdisc(link), "could not setup wan root qdisc")
	link = setUpDummyInterface(t, "lan")
	assert.NoError(setupRootQdisc(link), "could not setup lan root qdisc")
	LAN_INT = "lan"
	WAN_INT = "wan"

	shaper, err := GetShaper()
	assert.NoError(err, "could not get shaper")

	return assert, func() {
		ns.Close()
		runtime.UnlockOSThread()
	}, shaper.(*netlinkShaper)
}

func setUpNetlinkTest(t *testing.T) (func(), netlink.Link) {
	if os.Getuid() != 0 {
		t.Skip("Skipped test because it requires root privileges")
	}

	// new temporary namespace so we don't pollute the host
	// lock thread since the namespace is thread local
	runtime.LockOSThread()
	ns, err := netns.New()
	if err != nil {
		runtime.UnlockOSThread()
		t.Fatalf("Failed to create new network namespace: %v", err)
	}

	link := setUpDummyInterface(t, "foo")

	return func() {
		ns.Close()
		runtime.UnlockOSThread()
	}, link
}

func setUpDummyInterface(t *testing.T, name string) netlink.Link {
	// Use a Dummy interface for testing as boot2docker 1.8.1 does not suport
	// Ifb interfaces
	if err := netlink.LinkAdd(
		&netlink.Dummy{
			LinkAttrs: netlink.LinkAttrs{Name: name},
		}); err != nil {
		t.Fatal(err)
	}

	link, err := netlink.LinkByName(name)
	if err != nil {
		t.Fatal(err)
	}
	if err := netlink.LinkSetUp(link); err != nil {
		t.Fatal(err)
	}
	return link
}

func check(t *testing.T, err error, s string) {
	if err != nil {
		t.Fatalf(s+": %v", err)
	}
}

func test_cmd(t *testing.T, cmd string, args ...string) {
	if cmdOut, err := exec.Command(cmd, args...).CombinedOutput(); err != nil {
		t.Fatalf("There was an error running cmd %v:\n%v\n", err, string(cmdOut))
	} else {
		t.Logf("run_cmd %q\n%s\n", cmd+" "+strings.Join(args, " "), string(cmdOut))
	}
}
