package daemon

import (
	"github.com/facebook/augmented-traffic-control/src/atc_thrift"
)

type Shaper interface {
	GetPlatform() atc_thrift.PlatformType

	CreateGroup(member string) (int64, error)
	JoinGroup(id int64, member string) error
	LeaveGroup(id int64, member string) error
	DeleteGroup(id int64) error
	Shape(id int64, settings *atc_thrift.Shaping) error
	Unshape(id int64) error
}

// FakeShaper implements Shaper
type FakeShaper struct{}

func (FakeShaper) GetPlatform() atc_thrift.PlatformType {
	return atc_thrift.PlatformType_LINUX
}

func (FakeShaper) CreateGroup(string) (int64, error)                 { return 0, nil }
func (FakeShaper) JoinGroup(int64, string) error                     { return nil }
func (FakeShaper) LeaveGroup(int64, string) error                    { return nil }
func (FakeShaper) DeleteGroup(int64) error                           { return nil }
func (FakeShaper) Shape(id int64, shaping *atc_thrift.Shaping) error { return nil }
func (FakeShaper) Unshape(int64) error                               { return nil }
