package daemon

import (
	"fmt"

	"github.com/facebook/augmented-traffic-control/src/atc_thrift"
)

func GetShaper() (Shaper, error) {
	return nil, fmt.Errorf("Linux platform not supported")
}

// *netlinkShaper implements Shaper
type netlinkShaper struct{}

func (*netlinkShaper) GetPlatform() atc_thrift.PlatformType {
	return atc_thrift.PlatformType_LINUX
}

func (*netlinkShaper) CreateGroup(string) (int64, error) {
	return 0, fmt.Errorf("netlink is not implemented")
}
func (*netlinkShaper) JoinGroup(int64, string) error {
	return fmt.Errorf("netlink is not implemented")
}
func (*netlinkShaper) LeaveGroup(int64, string) error {
	return fmt.Errorf("netlink is not implemented")
}
func (*netlinkShaper) DeleteGroup(int64) error {
	return fmt.Errorf("netlink is not implemented")
}
func (*netlinkShaper) Shape(id int64, shaping *atc_thrift.Shaping) error {
	return fmt.Errorf("netlink is not implemented")
}
func (*netlinkShaper) Unshape(int64) error {
	return fmt.Errorf("netlink is not implemented")
}
