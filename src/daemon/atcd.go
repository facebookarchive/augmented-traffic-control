package daemon

import (
	"fmt"
	"net"

	"github.com/facebook/augmented-traffic-control/src/atc_thrift"
	"github.com/facebook/augmented-traffic-control/src/iptables"
	"github.com/hgfischer/go-otp"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
)

var (
	// Database errors are caught and logged internally
	// This is returned to the thrift client.
	DbError = fmt.Errorf("Database Error")

	// Shared constant between ATCD and ATC_API
	// Make sure to change in both places
	NoSuchItem = fmt.Errorf("NO_SUCH_ITEM")
)

func ReshapeFromDb(shaper *ShapingEngine, db *DbRunner) error {
	groups, err := db.GetAllGroups()
	if err != nil {
		return err
	}

	Log.Debugln("Reshaping from database")
	// Setup all the groups' shaping again
	for group := range groups {
		// First make sure the group has all the members in the DB
		members, err := db.GetMembersOf(group.id)
		if err != nil {
			continue
		}
		first := false
		for member := range members {
			var err error
			if first {
				err = shaper.CreateGroup(group.id, member)
				first = false
			} else {
				err = shaper.JoinGroup(group.id, member)
			}
			if err != nil {
				return errors.Wrapf(err, "failed to evaluate members of group %d", group.id)
			}
		}

		// Second shape the group using the settings from the DB
		// but only if the group has shaping settings to begin with
		if group.tc != nil {
			if err := shaper.Shape(group.id, group.tc); err != nil {
				return errors.Wrapf(err, "failed to shape for group %d", group.id)
			}
		}
	}
	return nil
}

type Atcd struct {
	db      *DbRunner
	shaper  *ShapingEngine
	options AtcdOptions
}

func NewAtcd(db *DbRunner, shaper *ShapingEngine, options *AtcdOptions) atc_thrift.Atcd {
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

func (atcd *Atcd) ListGroups() ([]*atc_thrift.ShapingGroup, error) {
	groups, err := atcd.db.GetAllGroups()
	if err != nil {
		return nil, DbError
	}
	results := make([]*atc_thrift.ShapingGroup, 0, 10)
	for grp := range groups {
		members, err := atcd.db.GetMembersOf(grp.id)
		if err != nil {
			return nil, DbError
		}
		results = append(results, &atc_thrift.ShapingGroup{
			ID:      grp.id,
			Shaping: grp.tc,
			Members: TargetsToStrings(members),
		})
	}
	return results, nil
}

func (atcd *Atcd) CreateGroup(member string) (*atc_thrift.ShapingGroup, error) {
	tgt, err := iptables.ParseTarget(member)
	if err != nil {
		return nil, err
	}
	grp := &atc_thrift.ShapingGroup{
		Members: []string{member},
		Shaping: nil,
	}
	dbgrp, err := atcd.db.UpdateGroup(DbGroup{
		secret: makeSecret(),
		tc:     nil,
	})
	if err != nil {
		return nil, DbError
	}
	// Have to insert the addr into the db before shaping
	// because hooks might depend on it.
	_, err = atcd.db.UpdateMember(DbMember{
		addr:     tgt,
		group_id: dbgrp.id,
	})
	if err != nil {
		return nil, DbError
	}
	// Have to create group in database before creating the shaper since
	// the database gives us the unique ID of the group, which the shaper
	// needs for the mark.
	if err := atcd.shaper.CreateGroup(dbgrp.id, tgt); err != nil {
		return nil, err
	}
	grp.ID = dbgrp.id
	defer atcd.Cleanup()
	return grp, nil
}

func (atcd *Atcd) GetGroup(id int64) (*atc_thrift.ShapingGroup, error) {
	group, err := atcd.db.GetGroup(id)
	if err != nil {
		return nil, DbError
	}
	if group == nil {
		return nil, NoSuchItem
	}

	members, err := atcd.db.GetMembersOf(id)
	if err != nil {
		return nil, DbError
	}
	grp := &atc_thrift.ShapingGroup{
		ID:      id,
		Members: TargetsToStrings(members),
		Shaping: group.tc,
	}
	return grp, nil
}

func (atcd *Atcd) GetGroupWith(addr string) (*atc_thrift.ShapingGroup, error) {
	tgt, err := iptables.ParseTarget(addr)
	if err != nil {
		return nil, err
	}

	member, err := atcd.db.GetMember(tgt)
	if err != nil {
		return nil, DbError
	}

	if member == nil {
		if ip, ok := tgt.(iptables.IPTarget); ok {
			// Search for network targets that contain the ip
			members, err := atcd.db.GetAllMembers()
			if err != nil {
				return nil, DbError
			}
			for member := range members {
				if cidr, ok := member.addr.(*iptables.CIDRTarget); ok {
					if cidr.Net.Contains((net.IP)(ip)) {
						return atcd.GetGroup(member.group_id)
					}
				}
			}
		}
		return nil, NoSuchItem
	}
	return atcd.GetGroup(member.group_id)
}

func (atcd *Atcd) GetGroupToken(id int64) (string, error) {
	group, err := atcd.db.GetGroup(id)
	if err != nil {
		return "", err
	}
	return atcd.token(group), nil
}

func (atcd *Atcd) JoinGroup(id int64, to_add, token string) error {
	tgt, err := iptables.ParseTarget(to_add)
	if err != nil {
		return err
	}
	group, err := atcd.db.GetGroup(id)
	if err != nil {
		return err
	}
	if group == nil {
		return fmt.Errorf("Group not found: %d", id)
	}
	if !atcd.verify(group, token) {
		return fmt.Errorf("Unauthorized")
	}
	// Insert into the db first since hooks might rely on the API.
	_, err = atcd.db.UpdateMember(DbMember{
		addr:     tgt,
		group_id: group.id,
	})
	if err := atcd.shaper.JoinGroup(id, tgt); err != nil {
		return err
	}
	defer atcd.Cleanup()
	return err
}

func (atcd *Atcd) LeaveGroup(id int64, to_remove, token string) error {
	tgt, err := iptables.ParseTarget(to_remove)
	if err != nil {
		return err
	}
	member, err := atcd.db.GetMember(tgt)
	if err != nil {
		return err
	}
	if member == nil {
		return NoSuchItem
	}
	if member.group_id != id {
		return fmt.Errorf("%q is not a member of group %d", to_remove, id)
	}
	group, err := atcd.db.GetGroup(member.group_id)
	if err != nil {
		return err
	}
	if !atcd.verify(group, token) {
		return fmt.Errorf("Unauthorized")
	}
	if err := atcd.shaper.LeaveGroup(id, tgt); err != nil {
		return err
	}
	defer atcd.Cleanup()
	return atcd.db.DeleteMember(tgt)
}

func (atcd *Atcd) ShapeGroup(id int64, settings *atc_thrift.Shaping, token string) (*atc_thrift.Shaping, error) {
	group, err := atcd.db.GetGroup(id)
	if err != nil {
		return nil, err
	}
	if !atcd.verify(group, token) {
		return nil, fmt.Errorf("Unauthorized")
	}
	group.tc = settings
	Log.Debugf("Shaping group %d with %+v\n", group.id, settings)
	err = atcd.shaper.Shape(group.id, group.tc)
	if err != nil {
		return nil, err
	}
	group, err = atcd.db.UpdateGroup(*group)
	if err != nil {
		return nil, err
	}
	return group.tc, nil
}

func (atcd *Atcd) UnshapeGroup(id int64, token string) error {
	group, err := atcd.db.GetGroup(id)
	if err != nil {
		return err
	}
	if !atcd.verify(group, token) {
		return fmt.Errorf("Unauthorized")
	}
	group.tc = nil
	Log.Debugf("Unshaping group %d\n", group.id)
	err = atcd.shaper.Unshape(group.id)
	if err != nil {
		return err
	}
	_, err = atcd.db.UpdateGroup(*group)
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

func (atcd *Atcd) Cleanup() {
	atcd.db.Cleanup()
}

func makeSecret() string {
	// Can probably find a better source of random secrets than this...
	return uuid.New()
}

func TargetsToStrings(ips chan iptables.Target) []string {
	s := make([]string, 0, 10)
	for ip := range ips {
		s = append(s, ip.String())
	}
	return s
}
