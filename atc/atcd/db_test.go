package main

import (
	"testing"

	"github.com/facebook/augmented-traffic-control/atc/atc_thrift/atc_thrift"
)

var (
	FakeShaping1 *atc_thrift.Setting = &atc_thrift.Setting{
		Up:   &atc_thrift.Shaping{},
		Down: nil,
	}
	FakeShaping2 *atc_thrift.Setting = &atc_thrift.Setting{
		Up:   nil,
		Down: &atc_thrift.Shaping{},
	}
)

func TestDBCreatesSchema(t *testing.T) {
	err := initDB("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer shutdownDB()

	db := <-db_chan
	defer func() { db_chan <- db }()
	_, err = db.Query("SELECT * FROM ShapingGroups")
	if err != nil {
		t.Fatal(err)
	}
}

func TestDBInsertsGroup(t *testing.T) {
	err := initDB("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer shutdownDB()

	group, err := dbUpdateGroup(DbGroup{
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
	err := initDB("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer shutdownDB()

	var group *DbGroup
	if group, err = dbUpdateGroup(DbGroup{id: 1, tc: FakeShaping1}); err != nil {
		t.Fatal(err)
	}
	if group, err = dbUpdateGroup(DbGroup{id: 2, tc: FakeShaping2}); err != nil {
		t.Fatal(err)
	}

	err = dbDeleteGroup(2)
	if err != nil {
		t.Fatal(err)
	}

	groups, err := dbGetAllGroups()
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
	err := initDB("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer shutdownDB()

	if _, err := dbUpdateGroup(DbGroup{tc: FakeShaping1}); err != nil {
		t.Fatal(err)
	}
	if _, err := dbUpdateGroup(DbGroup{tc: FakeShaping2}); err != nil {
		t.Fatal(err)
	}

	groups, err := dbGetAllGroups()
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 2 {
		t.Fatalf("Wrong number of groups: 2 != %d", len(groups))
	}
}

func TestDBUpdatesGroup(t *testing.T) {
	err := initDB("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer shutdownDB()

	var group *DbGroup
	if group, err = dbUpdateGroup(DbGroup{secret: "asdf", tc: FakeShaping1}); err != nil {
		t.Fatal(err)
	}
	if _, err = dbUpdateGroup(DbGroup{id: group.id, secret: "qwer", tc: FakeShaping1}); err != nil {
		t.Fatal(err)
	}

	groups, err := dbGetAllGroups()
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
	err := initDB("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer shutdownDB()

	var (
		group  *DbGroup
		member *DbMember
	)
	if group, err = dbUpdateGroup(DbGroup{tc: FakeShaping1}); err != nil {
		t.Fatal(err)
	}
	if member, err = dbUpdateMember(DbMember{"1.2.3.4", group.id}); err != nil {
		t.Fatal(err)
	}

	if member.group_id != group.id {
		t.Fatal("Wrong group id: %d != %d", group.id, member.group_id)
	}
	if member.addr != "1.2.3.4" {
		t.Fatal(`Wrong member address: "1.2.3.4" != %q`, member.addr)
	}
}

func TestDBGetsMember(t *testing.T) {
	err := initDB("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer shutdownDB()

	var (
		group  *DbGroup
		member *DbMember
	)
	if group, err = dbUpdateGroup(DbGroup{tc: FakeShaping1}); err != nil {
		t.Fatal(err)
	}
	if _, err = dbUpdateMember(DbMember{"1.2.3.4", group.id}); err != nil {
		t.Fatal(err)
	}

	if member, err = dbGetMember("1.2.3.4"); err != nil {
		t.Fatal(err)
	}

	if member.group_id != group.id {
		t.Fatal("Wrong group id: %d != %d", group.id, member.group_id)
	}
	if member.addr != "1.2.3.4" {
		t.Fatal(`Wrong member address: "1.2.3.4" != %q`, member.addr)
	}
}

func TestDBGetsMembersForGroup(t *testing.T) {
	err := initDB("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer shutdownDB()

	var (
		group   *DbGroup
		members []string
	)
	if group, err = dbUpdateGroup(DbGroup{tc: FakeShaping1}); err != nil {
		t.Fatal(err)
	}
	if _, err = dbUpdateMember(DbMember{"1.2.3.4", group.id}); err != nil {
		t.Fatal(err)
	}
	if _, err = dbUpdateMember(DbMember{"2.3.4.5", group.id}); err != nil {
		t.Fatal(err)
	}

	if members, err = dbGetMembers(group.id); err != nil {
		t.Fatal(err)
	}
	if len(members) != 2 {
		t.Fatalf("Wrong number of members: 2 != %d", len(members))
	}
}

// FIXME: testing
func TestDBUpdatesMember(t *testing.T) { t.Skip("test not implemented") }
func TestDBDeletesMember(t *testing.T) { t.Skip("test not implemented") }
