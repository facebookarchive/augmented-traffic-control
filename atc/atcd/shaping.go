package main

import (
	"fmt"

	"github.com/facebook/augmented-traffic-control/atc/atc_thrift/atc_thrift"
)

type Shaper interface {
	GetPlatform() atc_thrift.PlatformType

	// Needs to be extended to include group info
	Shape(shaping *atc_thrift.Setting) error

	// Needs to be extended to include group info
	Unshape() error
}

func GetShaper() Shaper {
	// FIXME: do switching on platform type or something...
	return FakeShaper{}
}

// FakeShaper implements Shaper
type FakeShaper struct{}

func (FakeShaper) GetPlatform() atc_thrift.PlatformType {
	return atc_thrift.PlatformType_LINUX
}

func (FakeShaper) Shape(shaping *atc_thrift.Setting) error { return nil }
func (FakeShaper) Unshape() error                          { return nil }

// *NetlinkShaper implements Shaper
type NetlinkShaper struct{}

func (*NetlinkShaper) GetPlatform() atc_thrift.PlatformType {
	return atc_thrift.PlatformType_LINUX
}

func (*NetlinkShaper) Shape(shaping *atc_thrift.Setting) error {
	return fmt.Errorf("Netlink not implemented")
}
func (*NetlinkShaper) Unshape() error {
	return fmt.Errorf("Netlink not implemented")
}
