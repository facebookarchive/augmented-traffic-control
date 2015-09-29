package daemon

import (
	"fmt"
)

func GetShaper() (Shaper, error) {
	return nil, fmt.Errorf("Darwin platform not supported")
}
