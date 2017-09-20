package daemon

import (
	"net"
	"testing"
	"time"

	"atc_thrift"
	"github.com/facebook/augmented-traffic-control/src/iptables"
)

var (
	FakeShaping1 *atc_thrift.Shaping = &atc_thrift.Shaping{
		Up:   &atc_thrift.LinkShaping{},
		Down: nil,
	}
	FakeShaping2 *atc_thrift.Shaping = &atc_thrift.Shaping{
		Up:   nil,
		Down: &atc_thrift.LinkShaping{},
	}

	IP1 = iptables.IPTarget(net.IPv4(1, 2, 3, 4))
	IP2 = iptables.IPTarget(net.IPv4(2, 3, 4, 5))
)

func TestDBCreatesSchema(t *testing.T) {
	db, err := NewDbRunner("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.db.Query("SELECT * FROM ShapingGroups")
	if err != nil {
		t.Fatal(err)
	}
}

func TestDBInsertsGroup(t *testing.T) {
	db, err := NewDbRunner("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	group, err := db.UpdateGroup(DbGroup{
		tc: FakeShaping1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if group.id < 0 {
		t.Fatalf("Id should be >= 0: %d", group.id)
	}
	if group.tc.Up == nil {
		t.Fatalf(`Mismatched tc settings: %+v`, group.tc)
	}
}

func TestDBDeletesGroup(t *testing.T) {
	db, err := NewDbRunner("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var group *DbGroup
	if group, err = db.UpdateGroup(DbGroup{id: 1, tc: FakeShaping1}); err != nil {
		t.Fatal(err)
	}
	if group, err = db.UpdateGroup(DbGroup{id: 2, tc: FakeShaping2}); err != nil {
		t.Fatal(err)
	}

	err = db.DeleteGroup(2)
	if err != nil {
		t.Fatal(err)
	}

	groups_c, err := db.GetAllGroups()
	if err != nil {
		t.Fatal(err)
	}
	groups := accumulate_groups(groups_c)

	if len(groups) != 1 {
		t.Fatalf("Wrong number of groups (expecting 1 group): %#v", groups)
	}
	if groups[0].tc.Up == nil {
		t.Fatalf(`Mismatched tc settings: %+v`, group.tc)
	}
}

func TestDBGetsAllGroups(t *testing.T) {
	db, err := NewDbRunner("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.UpdateGroup(DbGroup{tc: FakeShaping1}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.UpdateGroup(DbGroup{tc: FakeShaping2}); err != nil {
		t.Fatal(err)
	}

	groups_c, err := db.GetAllGroups()
	if err != nil {
		t.Fatal(err)
	}
	groups := accumulate_groups(groups_c)

	if len(groups) != 2 {
		t.Fatalf("Wrong number of groups (expecting 2 groups): %#v", groups)
	}
}

func TestDBUpdatesGroup(t *testing.T) {
	db, err := NewDbRunner("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var group *DbGroup
	if group, err = db.UpdateGroup(DbGroup{secret: "asdf", tc: FakeShaping1}); err != nil {
		t.Fatal(err)
	}
	if _, err = db.UpdateGroup(DbGroup{id: group.id, secret: "qwer", tc: FakeShaping1}); err != nil {
		t.Fatal(err)
	}

	groups_c, err := db.GetAllGroups()
	if err != nil {
		t.Fatal(err)
	}
	groups := accumulate_groups(groups_c)

	if len(groups) != 1 {
		t.Fatalf("Wrong number of groups (expecting 1 group): %#v", groups)
	}
	if groups[0].id != group.id {
		t.Fatalf("Wrong group id: %d != %d", group.id, groups[0].id)
	}
	if groups[0].secret != "qwer" {
		t.Fatalf(`Wrong group secret: "qwer" != %q`, groups[0].secret)
	}
}

func TestDBInsertsMember(t *testing.T) {
	db, err := NewDbRunner("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var (
		group  *DbGroup
		member *DbMember
	)
	if group, err = db.UpdateGroup(DbGroup{tc: FakeShaping1}); err != nil {
		t.Fatal(err)
	}
	if member, err = db.UpdateMember(DbMember{IP1, group.id}); err != nil {
		t.Fatal(err)
	}

	if member.group_id != group.id {
		t.Fatalf("Wrong group id: %d != %d", group.id, member.group_id)
	}
	if member.addr.String() != IP1.String() {
		t.Fatalf(`Wrong member address: %q != %q`, IP1, member.addr)
	}
}

func TestDBGetsMember(t *testing.T) {
	db, err := NewDbRunner("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var (
		group  *DbGroup
		member *DbMember
	)
	if group, err = db.UpdateGroup(DbGroup{tc: FakeShaping1}); err != nil {
		t.Fatal(err)
	}
	if _, err = db.UpdateMember(DbMember{IP1, group.id}); err != nil {
		t.Fatal(err)
	}

	if member, err = db.GetMember(IP1); err != nil {
		t.Fatal(err)
	}

	if member.group_id != group.id {
		t.Fatalf("Wrong group id: %d != %d", group.id, member.group_id)
	}
	if member.addr.String() != IP1.String() {
		t.Fatalf(`Wrong member address: %q != %q`, IP1, member.addr)
	}
}

func TestDBGetsMembersForGroup(t *testing.T) {
	db, err := NewDbRunner("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var (
		group     *DbGroup
		members_c chan iptables.Target
	)
	if group, err = db.UpdateGroup(DbGroup{tc: FakeShaping1}); err != nil {
		t.Fatal(err)
	}
	if _, err = db.UpdateMember(DbMember{IP1, group.id}); err != nil {
		t.Fatal(err)
	}
	if _, err = db.UpdateMember(DbMember{IP2, group.id}); err != nil {
		t.Fatal(err)
	}

	if members_c, err = db.GetMembersOf(group.id); err != nil {
		t.Fatal(err)
	}
	members := accumulate_targets(members_c)

	if len(members) != 2 {
		t.Fatalf("Wrong number of members (expecting 2 members): %#v", members)
	}
}

func TestDBCleansEmptyGroups(t *testing.T) {
	db, err := NewDbRunner("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var group *DbGroup
	if _, err = db.UpdateGroup(DbGroup{secret: "qwer"}); err != nil {
		t.Fatal(err)
	}
	if group, err = db.UpdateGroup(DbGroup{secret: "asdf"}); err != nil {
		t.Fatal(err)
	}
	if _, err = db.UpdateMember(DbMember{IP1, group.id}); err != nil {
		t.Fatal(err)
	}

	n, err := db.cleanupEmptyGroups()
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("Wrong number of groups deleted: 1 != %d", n)
	}

	groups_c, err := db.GetAllGroups()
	if err != nil {
		t.Fatal(err)
	}
	groups := accumulate_groups(groups_c)

	if len(groups) != 1 {
		t.Fatalf("Wrong number of groups (expecting 1 group): %#v", groups)
	}

	if groups[0].secret != "asdf" {
		t.Fatalf(`Wrong group: "asdf" != %q`, groups[0].secret)
	}
}

func TestDBCleansOldGroups(t *testing.T) {
	db, err := NewDbRunner("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var group *DbGroup
	if _, err = db.UpdateGroup(DbGroup{secret: "qwer"}); err != nil {
		t.Fatal(err)
	}
	if group, err = db.UpdateGroup(DbGroup{secret: "asdf"}); err != nil {
		t.Fatal(err)
	}

	_, err = db.db.Exec(`update shapinggroups set timeout=? where id=?`, time.Now().Add(-24*time.Hour).Unix(), group.id)
	if err != nil {
		t.Fatal(err)
	}

	n, err := db.cleanupOldGroups()
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("Wrong number of groups deleted: 1 != %d", n)
	}

	groups_c, err := db.GetAllGroups()
	if err != nil {
		t.Fatal(err)
	}
	groups := accumulate_groups(groups_c)

	if len(groups) != 1 {
		t.Fatalf("Wrong number of groups: 1 != %d", len(groups))
	}

	if groups[0].secret != "qwer" {
		t.Fatalf(`Wrong group: "qwer" != %q`, groups[0].secret)
	}
}

func accumulate_groups(groups chan *DbGroup) []*DbGroup {
	results := make([]*DbGroup, 0, 10)
	for group := range groups {
		results = append(results, group)
	}
	return results
}

func accumulate_targets(members chan iptables.Target) []iptables.Target {
	results := make([]iptables.Target, 0, 10)
	for member := range members {
		results = append(results, member)
	}
	return results
}
