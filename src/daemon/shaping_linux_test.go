package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/facebook/augmented-traffic-control/src/atc_thrift"
	"github.com/vishvananda/netlink"
)

func e(cmd string, args []string) {
	if cmdOut, err := exec.Command(cmd, args...).Output(); err != nil {
		fmt.Fprintln(os.Stderr, "There was an error running cmd ", err)
	} else {
		fmt.Fprint(os.Stdout, string(cmdOut))
	}
}

func TestShaping(t *testing.T) {
	tearDown := setUpNetlinkTest(t)
	defer tearDown()
	// Use a Dummy interface for testing as boot2docker 1.8.1 does not suport
	// Ifb interfaces
	if err := netlink.LinkAdd(
		&netlink.Dummy{
			LinkAttrs: netlink.LinkAttrs{Name: "foo"},
		}); err != nil {
		t.Fatal(err)
	}

	link, err := netlink.LinkByName("foo")
	if err != nil {
		t.Fatal(err)
	}
	if err := netlink.LinkSetUp(link); err != nil {
		t.Fatal(err)
	}

	// Setup root qdisc.
	if err := setupRootQdisc("foo"); err != nil {
		t.Fatal(err)
	}
	// Check that the root qdisc exists.
	qdiscs, err := netlink.QdiscList(link)
	if err != nil {
		t.Fatal(err)
	}
	if len(qdiscs) != 1 {
		t.Fatal("Failed to add qdisc")
	}
	_, ok := qdiscs[0].(*netlink.Htb)
	if !ok {
		t.Fatal("Qdisc is the wrong type")
	}

	// Set up class + filters (ipv4/ipv6) using shape_on
	mark := int64(5)
	shaping := &atc_thrift.LinkShaping{Rate: 10}
	shape_on(mark, shaping, link)

	classes, err := netlink.ClassList(
		link, netlink.MakeHandle(0x1, uint16(mark)),
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(classes) != 1 {
		t.Fatal("Failed to add class")
	}

	filters, err := netlink.FilterList(link, netlink.MakeHandle(0x1, 0))
	if err != nil {
		t.Fatal(err)
	}
	e("tc", []string{"qdisc", "show", "dev", "foo"})
	e("tc", []string{"class", "show", "dev", "foo"})
	e("tc", []string{"filter", "show", "dev", "foo"})

	// We expect 2 filters to be setup: 1 for ipv4, 1 for ipv6.
	if len(filters) != 2 {
		t.Fatal("Failed to add filter")
	}
}
