package shaping

import (
	"fmt"
)

// GetShaper returns a shaper suitable for the current platform.
func GetShaper() (Shaper, error) {
	return nil, fmt.Errorf("Darwin platform not supported")
}
