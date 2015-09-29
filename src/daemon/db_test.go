package daemon

import (
	"testing"
	"time"

	"github.com/facebook/augmented-traffic-control/src/atc_thrift"
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

	group, err := db.updateGroup(DbGroup{
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
	if group, err = db.updateGroup(DbGroup{id: 1, tc: FakeShaping1}); err != nil {
		t.Fatal(err)
	}
	if group, err = db.updateGroup(DbGroup{id: 2, tc: FakeShaping2}); err != nil {
		t.Fatal(err)
	}

	err = db.deleteGroup(2)
	if err != nil {
		t.Fatal(err)
	}

	groups, err := db.getAllGroups()
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 1 {
		t.Fatalf("Wrong number of groups: 1 != %d", len(groups))
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

	if _, err := db.updateGroup(DbGroup{tc: FakeShaping1}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.updateGroup(DbGroup{tc: FakeShaping2}); err != nil {
		t.Fatal(err)
	}

	groups, err := db.getAllGroups()
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 2 {
		t.Fatalf("Wrong number of groups: 2 != %d", len(groups))
	}
}

func TestDBUpdatesGroup(t *testing.T) {
	db, err := NewDbRunner("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var group *DbGroup
	if group, err = db.updateGroup(DbGroup{secret: "asdf", tc: FakeShaping1}); err != nil {
		t.Fatal(err)
	}
	if _, err = db.updateGroup(DbGroup{id: group.id, secret: "qwer", tc: FakeShaping1}); err != nil {
		t.Fatal(err)
	}

	groups, err := db.getAllGroups()
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 1 {
		t.Fatalf("Wrong number of groups: 1 != %d", len(groups))
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
	if group, err = db.updateGroup(DbGroup{tc: FakeShaping1}); err != nil {
		t.Fatal(err)
	}
	if member, err = db.updateMember(DbMember{"1.2.3.4", group.id}); err != nil {
		t.Fatal(err)
	}

	if member.group_id != group.id {
		t.Fatalf("Wrong group id: %d != %d", group.id, member.group_id)
	}
	if member.addr != "1.2.3.4" {
		t.Fatalf(`Wrong member address: "1.2.3.4" != %q`, member.addr)
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
	if group, err = db.updateGroup(DbGroup{tc: FakeShaping1}); err != nil {
		t.Fatal(err)
	}
	if _, err = db.updateMember(DbMember{"1.2.3.4", group.id}); err != nil {
		t.Fatal(err)
	}

	if member, err = db.getMember("1.2.3.4"); err != nil {
		t.Fatal(err)
	}

	if member.group_id != group.id {
		t.Fatalf("Wrong group id: %d != %d", group.id, member.group_id)
	}
	if member.addr != "1.2.3.4" {
		t.Fatalf(`Wrong member address: "1.2.3.4" != %q`, member.addr)
	}
}

func TestDBGetsMembersForGroup(t *testing.T) {
	db, err := NewDbRunner("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var (
		group   *DbGroup
		members []string
	)
	if group, err = db.updateGroup(DbGroup{tc: FakeShaping1}); err != nil {
		t.Fatal(err)
	}
	if _, err = db.updateMember(DbMember{"1.2.3.4", group.id}); err != nil {
		t.Fatal(err)
	}
	if _, err = db.updateMember(DbMember{"2.3.4.5", group.id}); err != nil {
		t.Fatal(err)
	}

	if members, err = db.getMembersOf(group.id); err != nil {
		t.Fatal(err)
	}
	if len(members) != 2 {
		t.Fatalf("Wrong number of members: 2 != %d", len(members))
	}
}

func TestDBCleansEmptyGroups(t *testing.T) {
	db, err := NewDbRunner("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var group *DbGroup
	if _, err = db.updateGroup(DbGroup{secret: "qwer"}); err != nil {
		t.Fatal(err)
	}
	if group, err = db.updateGroup(DbGroup{secret: "asdf"}); err != nil {
		t.Fatal(err)
	}
	if _, err = db.updateMember(DbMember{"1.2.3.4", group.id}); err != nil {
		t.Fatal(err)
	}

	n, err := db.cleanupEmptyGroups()
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("Wrong number of groups deleted: 1 != %d", n)
	}

	groups, err := db.getAllGroups()
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 1 {
		t.Fatalf("Wrong number of groups: 1 != %d", len(groups))
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
	if _, err = db.updateGroup(DbGroup{secret: "qwer"}); err != nil {
		t.Fatal(err)
	}
	if group, err = db.updateGroup(DbGroup{secret: "asdf"}); err != nil {
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

	groups, err := db.getAllGroups()
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 1 {
		t.Fatalf("Wrong number of groups: 1 != %d", len(groups))
	}

	if groups[0].secret != "qwer" {
		t.Fatalf(`Wrong group: "qwer" != %q`, groups[0].secret)
	}
}
