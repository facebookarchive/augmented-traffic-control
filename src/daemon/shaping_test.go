package daemon

import (
	"testing"
)

func TestTypes(t *testing.T) {
	// Make sure concrete types implement common interface
	var _ Shaper = FakeShaper{}
	var _ Shaper = &NetlinkShaper{}
}

// Tests for netlink shaping should go here... eventually
