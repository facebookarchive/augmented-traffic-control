package daemon

import "github.com/facebook/augmented-traffic-control/src/iptables"

type DbRunner interface {
	Close()
	GetGroup(id int64) (*DbGroup, error)
	GetAllGroups() (chan *DbGroup, error)
	DeleteGroup(id int64) error
	UpdateGroup(group DbGroup) (*DbGroup, error)
	GetMember(addr iptables.Target) (*DbMember, error)
	UpdateMember(member DbMember) (*DbMember, error)
	DeleteMember(addr iptables.Target) error
	GetMembersOf(id int64) (chan iptables.Target, error)
	GetAllMembers() (chan *DbMember, error)
	Cleanup() error
}
