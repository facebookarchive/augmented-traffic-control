// Package shaping package the `Shaper` interface to which platform-specific
// shapers must conform. It also contains platform-specific shapers.
package shaping

import (
	"atc_thrift"

	"github.com/facebook/augmented-traffic-control/src/iptables"
	atc_log "github.com/facebook/augmented-traffic-control/src/log"
)

// Log is the package-wide logger
var Log *atc_log.LogMux

func init() {
	Log = atc_log.NewMux(atc_log.Syslog(), atc_log.Stdlog())
}

// Target is a target ipaddress for which to apply shaping rules
type Target iptables.Target

// A Shaper is a collection of platform-specific operations
// for shaping network traffic
type Shaper interface {
	// Get the platform type for this shaper.
	GetPlatform() atc_thrift.PlatformType

	// Initialize the shaper. Does any setup the platform requires
	// before shaping and return nil if successful or an error describing
	// what went wrong.
	Initialize() error

	// Create a new group with the given id number and initial member.
	CreateGroup(id int64, member Target) error

	// Add the provided member to the given group.
	JoinGroup(id int64, member Target) error

	// Remove the provided member from the given group.
	LeaveGroup(id int64, member Target) error

	// Delete a group. The group is assumed to be empty *before* DeleteGroup
	// is called.
	DeleteGroup(id int64) error

	// Apply the provided shaping settings to the given group.
	Shape(id int64, settings *atc_thrift.Shaping) error

	// Remove the current shaping settings for the given group.
	Unshape(id int64) error
}

// FakeShaper is a stub shaper used by tests.
// All functions except GetPlatform return nil.
// GetPlatform returns atc_thrift.PlatformType_LINUX
func FakeShaper() Shaper {
	return &fakeShaper{}
}

// you do not have to document if you do not export
// the struct

type fakeShaper struct{}

func (fakeShaper) GetPlatform() atc_thrift.PlatformType {
	return atc_thrift.PlatformType_LINUX
}

func (fakeShaper) Initialize() error                                 { return nil }
func (fakeShaper) CreateGroup(int64, Target) error                   { return nil }
func (fakeShaper) JoinGroup(int64, Target) error                     { return nil }
func (fakeShaper) LeaveGroup(int64, Target) error                    { return nil }
func (fakeShaper) DeleteGroup(int64) error                           { return nil }
func (fakeShaper) Shape(id int64, shaping *atc_thrift.Shaping) error { return nil }
func (fakeShaper) Unshape(int64) error                               { return nil }
