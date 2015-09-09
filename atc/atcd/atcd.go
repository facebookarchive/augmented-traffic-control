package main

import (
	"fmt"
	//"time"

	"code.google.com/p/go-uuid/uuid"
	"github.com/facebook/augmented-traffic-control/atc/atc_thrift/atc_thrift"
	"github.com/hgfischer/go-otp"
	_ "github.com/mattn/go-sqlite3"
)

// Shared constant between ATCD and ATC_API
// Make sure to change in both places
var NoSuchItem error = fmt.Errorf("NO_SUCH_ITEM")

type Atcd struct {
	db *DbRunner
}

func NewAtcd(db *DbRunner) atc_thrift.Atcd {
	return &Atcd{
		db: db,
	}
}

func (atcd *Atcd) GetAtcdInfo() (*atc_thrift.AtcdInfo, error) {
	info := &atc_thrift.AtcdInfo{
		Platform: GetShaper().GetPlatform(),
		Version:  VERSION,
	}
	return info, nil
}

func (atcd *Atcd) CreateGroup(member string) (*atc_thrift.ShapingGroup, error) {
	grp := &atc_thrift.ShapingGroup{
		Members: []string{member},
		Shaping: nil,
	}
	dbgrp, err := atcd.db.updateGroup(DbGroup{
		secret: makeSecret(),
		tc:     nil,
	})
	if err != nil {
		return nil, err
	}
	_, err = atcd.db.updateMember(DbMember{
		addr:     member,
		group_id: dbgrp.id,
	})
	if err != nil {
		return nil, err
	}
	grp.Id = dbgrp.id
	return grp, nil
}

func (atcd *Atcd) GetGroup(id int64) (*atc_thrift.ShapingGroup, error) {
	group, err := atcd.db.getGroup(id)
	if err != nil {
		return nil, err
	}
	members, err := atcd.db.getMembersOf(id)
	if err != nil {
		return nil, err
	}
	grp := &atc_thrift.ShapingGroup{
		Id:      id,
		Members: members,
		Shaping: group.tc,
	}
	return grp, nil
}

func (atcd *Atcd) GetGroupWith(addr string) (*atc_thrift.ShapingGroup, error) {
	member, err := atcd.db.getMember(addr)
	if err != nil {
		return nil, err
	}
	return atcd.GetGroup(member.group_id)
}

func (atcd *Atcd) GetGroupToken(id int64) (string, error) {
	group, err := atcd.db.getGroup(id)
	if err != nil {
		return "", err
	}
	p := atcd.otp(group)
	return p.Get(), nil
}

func (atcd *Atcd) JoinGroup(id int64, to_add, token string) error {
	group, err := atcd.db.getGroup(id)
	if err != nil {
		return err
	}
	p := atcd.otp(group)
	if !p.Verify(token) {
		return fmt.Errorf("Unauthorized")
	}
	_, err = atcd.db.updateMember(DbMember{
		addr:     to_add,
		group_id: group.id,
	})
	return err
}

func (atcd *Atcd) LeaveGroup(id int64, to_remove, token string) error {
	member, err := atcd.db.getMember(to_remove)
	if err != nil {
		return err
	}
	if member.group_id != id {
		return fmt.Errorf("%q is not a member of group %d", to_remove, id)
	}
	group, err := atcd.db.getGroup(member.group_id)
	if err != nil {
		return err
	}
	p := atcd.otp(group)
	if !p.Verify(token) {
		return fmt.Errorf("Unauthorized")
	}
	return atcd.db.deleteMember(to_remove)
}

func (atcd *Atcd) ShapeGroup(id int64, settings *atc_thrift.Setting, token string) (*atc_thrift.Setting, error) {
	return nil, nil
}

func (atcd *Atcd) UnshapeGroup(id int64, token string) error {
	return nil
}

func (atcd *Atcd) otp(group *DbGroup) *otp.TOTP {
	return &otp.TOTP{
		Secret:         fmt.Sprintf("%s::%d", group.secret, group.id),
		IsBase32Secret: true,
	}
}

func makeSecret() string {
	// Can probably find a better source of random secrets than this...
	return uuid.New()
}
