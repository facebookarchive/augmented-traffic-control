package daemon

import (
	"net"

	"github.com/facebook/augmented-traffic-control/src/atc_thrift"
)

type Shaper interface {
	GetPlatform() atc_thrift.PlatformType

	Initialize() error
	CreateGroup(id int64, member net.IP) error
	JoinGroup(id int64, member net.IP) error
	LeaveGroup(id int64, member net.IP) error
	DeleteGroup(id int64) error
	Shape(id int64, settings *atc_thrift.Shaping) error
	Unshape(id int64) error
}

// FakeShaper implements Shaper
type FakeShaper struct{}

func (FakeShaper) GetPlatform() atc_thrift.PlatformType {
	return atc_thrift.PlatformType_LINUX
}

func (FakeShaper) Initialize() error                                 { return nil }
func (FakeShaper) CreateGroup(int64, net.IP) error                   { return nil }
func (FakeShaper) JoinGroup(int64, net.IP) error                     { return nil }
func (FakeShaper) LeaveGroup(int64, net.IP) error                    { return nil }
func (FakeShaper) DeleteGroup(int64) error                           { return nil }
func (FakeShaper) Shape(id int64, shaping *atc_thrift.Shaping) error { return nil }
func (FakeShaper) Unshape(int64) error                               { return nil }
