package daemon

import (
	"fmt"
	"net"

	"github.com/facebook/augmented-traffic-control/src/atc_thrift"
	"github.com/facebook/augmented-traffic-control/src/shaping"
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

func ReshapeFromDb(shaper shaping.Shaper, db *DbRunner) error {
	groups := <-db.GetAllGroups()
	if groups == nil || len(groups) == 0 {
		return nil
	}

	Log.Println("Reshaping from database")
	// Setup all the groups' shaping again
	for _, group := range groups {
		// First make sure the group has all the members in the DB
		members := <-db.GetMembersOf(group.id)
		if members == nil || len(members) == 0 {
			// If there aren't any members, don't bother doing anything.
			// Empty groups are cleaned regularly so this is unlikely
			continue
		}
		first := false
		for _, member := range members {
			mem_ip := net.ParseIP(member)
			var err error
			if first {
				err = shaper.CreateGroup(group.id, mem_ip)
				first = false
			} else {
				err = shaper.JoinGroup(group.id, mem_ip)
			}
			if err != nil {
				return err
			}
		}

		// Second shape the group using the settings from the DB
		// but only if the group has shaping settings to begin with
		if group.tc != nil {
			if err := shaper.Shape(group.id, group.tc); err != nil {
				return err
			}
		}
	}
	return nil
}

type Atcd struct {
	db      *DbRunner
	shaper  shaping.Shaper
	options AtcdOptions
}

func NewAtcd(db *DbRunner, shaper shaping.Shaper, options *AtcdOptions) atc_thrift.Atcd {
	if options == nil {
		options = &DefaultAtcdOptions
	}
	return &Atcd{
		db:      db,
		shaper:  shaper,
		options: *options,
	}
}

func (atcd *Atcd) GetAtcdInfo() (*atc_thrift.AtcdInfo, error) {
	info := &atc_thrift.AtcdInfo{
		Platform: atcd.shaper.GetPlatform(),
		Version:  VERSION,
	}
	return info, nil
}

func (atcd *Atcd) CreateGroup(member string) (*atc_thrift.ShapingGroup, error) {
	ip := net.ParseIP(member)
	if ip == nil {
		return nil, fmt.Errorf("Malformed IP address: %q", member)
	}
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
	// Have to create group in database before creating the shaper since
	// the database gives us the unique ID of the group, which the shaper
	// needs for the mark.
	if err := atcd.shaper.CreateGroup(dbgrp.id, ip); err != nil {
		return nil, err
	}
	dbmem := <-atcd.db.UpdateMember(DbMember{
		addr:     member,
		group_id: dbgrp.id,
	})
	if dbmem == nil {
		return nil, DbError
	}
	grp.ID = dbgrp.id
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
		ID:      id,
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
	group, err := atcd.db.getGroup(id)
	if err != nil {
		return "", err
	}
	return atcd.token(group), nil
}

func (atcd *Atcd) JoinGroup(id int64, to_add, token string) error {
	ip := net.ParseIP(to_add)
	if ip == nil {
		return fmt.Errorf("Malformed IP address: %q", to_add)
	}
	group, err := atcd.db.getGroup(id)
	if err != nil {
		return err
	}
	if !atcd.verify(group, token) {
		return fmt.Errorf("Unauthorized")
	}
	if err := atcd.shaper.JoinGroup(id, ip); err != nil {
		return err
	}
	_, err = atcd.db.updateMember(DbMember{
		addr:     to_add,
		group_id: group.id,
	})
	return err
}

func (atcd *Atcd) LeaveGroup(id int64, to_remove, token string) error {
	ip := net.ParseIP(to_remove)
	if ip == nil {
		return fmt.Errorf("Malformed IP address: %q", to_remove)
	}
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
	if err := atcd.shaper.LeaveGroup(id, ip); err != nil {
		return err
	}
	// FIXME: clean shaper's group too!
	defer atcd.db.Cleanup()
	return atcd.db.deleteMember(to_remove)
}

func (atcd *Atcd) ShapeGroup(id int64, settings *atc_thrift.Shaping, token string) (*atc_thrift.Shaping, error) {
	group, err := atcd.db.getGroup(id)
	if err != nil {
		return nil, err
	}
	if !atcd.verify(group, token) {
		return nil, fmt.Errorf("Unauthorized")
	}
	group.tc = settings
	Log.Println("Shaping group", group.id)
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
	Log.Println("Unshaping group", group.id)
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
	if token == group.secret || !atcd.options.Secure {
		return true
	}
	return atcd.otp(group).Verify(token)
}

func (atcd *Atcd) token(group *DbGroup) string {
	if !atcd.options.Secure {
		return ""
	}
	return atcd.otp(group).Get()
}

func (atcd *Atcd) otp(group *DbGroup) *otp.TOTP {
	return &otp.TOTP{
		Secret: fmt.Sprintf("%s::%d", group.secret, group.id),
		Period: atcd.options.OtpTimeout,
	}
}

func makeSecret() string {
	// Can probably find a better source of random secrets than this...
	return uuid.New()
}
