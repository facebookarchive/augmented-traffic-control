package daemon

import (
	"fmt"
	"log"

	"github.com/facebook/augmented-traffic-control/src/atc_thrift"
	"github.com/hgfischer/go-otp"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pborman/uuid"
)

var (
	// Database errors are caught and logged internally
	// This is returned to the thrift client.
	DbError = fmt.Errorf("Database Error")

	// Shared constant between ATCD and ATC_API
	// Make sure to change in both places
	NoSuchItem = fmt.Errorf("NO_SUCH_ITEM")
)

type Atcd struct {
	db     *DbRunner
	shaper Shaper
	secure bool
}

func NewAtcd(db *DbRunner, shaper Shaper, secure bool) atc_thrift.Atcd {
	return &Atcd{
		db:     db,
		shaper: shaper,
		secure: secure,
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
	dbgrp := <-atcd.db.UpdateGroup(DbGroup{
		secret: makeSecret(),
		tc:     nil,
	})
	if dbgrp == nil {
		return nil, DbError
	}
	dbmem := <-atcd.db.UpdateMember(DbMember{
		addr:     member,
		group_id: dbgrp.id,
	})
	if dbmem == nil {
		return nil, DbError
	}
	grp.Id = dbgrp.id
	return grp, nil
}

func (atcd *Atcd) GetGroup(id int64) (*atc_thrift.ShapingGroup, error) {
	group := <-atcd.db.GetGroup(id)
	if group == nil {
		return nil, NoSuchItem
	}
	members := <-atcd.db.GetMembersOf(id)
	if members == nil {
		return nil, NoSuchItem
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
	if member == nil {
		return nil, NoSuchItem
	}
	return atcd.GetGroup(member.group_id)
}

func (atcd *Atcd) GetGroupToken(id int64) (string, error) {
	if !atcd.secure {
		return "", nil
	}
	group, err := atcd.db.getGroup(id)
	if err != nil {
		return "", err
	}
	return atcd.token(group), nil
}

func (atcd *Atcd) JoinGroup(id int64, to_add, token string) error {
	group, err := atcd.db.getGroup(id)
	if err != nil {
		return err
	}
	if !atcd.verify(group, token) {
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
	if !atcd.verify(group, token) {
		return fmt.Errorf("Unauthorized")
	}
	defer atcd.db.Cleanup()
	return atcd.db.deleteMember(to_remove)
}

func (atcd *Atcd) ShapeGroup(id int64, settings *atc_thrift.Setting, token string) (*atc_thrift.Setting, error) {
	group, err := atcd.db.getGroup(id)
	if err != nil {
		return nil, err
	}
	if !atcd.verify(group, token) {
		return nil, fmt.Errorf("Unauthorized")
	}
	group.tc = settings
	log.Println("Shaping group", group.id)
	err = atcd.shaper.Shape(group.id, group.tc)
	if err != nil {
		return nil, err
	}
	group, err = atcd.db.updateGroup(*group)
	if err != nil {
		return nil, err
	}
	return group.tc, nil
}

func (atcd *Atcd) UnshapeGroup(id int64, token string) error {
	group, err := atcd.db.getGroup(id)
	if err != nil {
		return err
	}
	if !atcd.verify(group, token) {
		return fmt.Errorf("Unauthorized")
	}
	group.tc = nil
	log.Println("Unshaping group", group.id)
	err = atcd.shaper.Unshape(group.id)
	if err != nil {
		return err
	}
	_, err = atcd.db.updateGroup(*group)
	if err != nil {
		return err
	}
	return nil
}

func (atcd *Atcd) verify(group *DbGroup, token string) bool {
	if token == group.secret {
		return true
	}
	t := &otp.TOTP{
		Secret:         fmt.Sprintf("%s::%d", group.secret, group.id),
		IsBase32Secret: true,
	}
	return t.Verify(token)
}

func (atcd *Atcd) token(group *DbGroup) string {
	t := &otp.TOTP{
		Secret:         fmt.Sprintf("%s::%d", group.secret, group.id),
		IsBase32Secret: true,
	}
	return t.Get()
}

func makeSecret() string {
	// Can probably find a better source of random secrets than this...
	return uuid.New()
}
