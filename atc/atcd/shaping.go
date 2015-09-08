package main

import (
	"fmt"

	"github.com/facebook/augmented-traffic-control/atc/atc_thrift/atc_thrift"
)

type Shaper interface {
	GetPlatform() atc_thrift.PlatformType
	Shape(addr string, shaping *atc_thrift.Setting) error
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

func (FakeShaper) Shape(addr string, shaping *atc_thrift.Setting) error {
	return nil
}

// *NetlinkShaper implements Shaper
type NetlinkShaper struct{}

func (*NetlinkShaper) GetPlatform() atc_thrift.PlatformType {
	return atc_thrift.PlatformType_LINUX
}

func (*NetlinkShaper) Shape(addr string, shaping *atc_thrift.Setting) error {
	return fmt.Errorf("Netlink not implemented")
}
