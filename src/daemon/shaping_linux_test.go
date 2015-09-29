package daemon

import (
	"testing"
)

func TestNetlinkInterface(t *testing.T) {
	var _ Shaper = &netlinkShaper{}
}

// Tests for netlink shaping should go here... eventually
