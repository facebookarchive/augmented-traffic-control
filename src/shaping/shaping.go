/*
The `shaping` package contains the `Shaper` interface to which platform-specific
shapers must conform. It also contains platform-specific shapers.
*/
package shaping

import (
	"net"

	"github.com/facebook/augmented-traffic-control/src/atc_thrift"
	. "github.com/facebook/augmented-traffic-control/src/log"
)

var Log *LogMux

func init() {
	Log = NewMux(Syslog(), Stdlog())
}

type Shaper interface {
	// Get the platform type for this shaper.
	GetPlatform() atc_thrift.PlatformType

	// Initialize the shaper. Does any setup the platform requires
	// before shaping and return nil if successful or an error describing
	// what went wrong.
	Initialize() error

	// Create a new group with the given id number and initial member.
	CreateGroup(id int64, member net.IP) error

	// Add the provided member to the given group.
	JoinGroup(id int64, member net.IP) error

	// Remove the provided member from the given group.
	LeaveGroup(id int64, member net.IP) error

	// Delete a group. The group is assumed to be empty *before* DeleteGroup
	// is called.
	DeleteGroup(id int64) error

	// Apply the provided shaping settings to the given group.
	Shape(id int64, settings *atc_thrift.Shaping) error

	// Remove the current shaping settings for the given group.
	Unshape(id int64) error
}

/*
A stub shaper used by tests.

All functions except GetPlatform return nil.

GetPlatform returns atc_thrift.PlatformType_LINUX
*/
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
