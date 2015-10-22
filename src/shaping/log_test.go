package shaping

import (
	. "github.com/facebook/augmented-traffic-control/src/log"
)

func init() {
	// Tests log to stdout, not syslog
	Log = NewMux(Stdlog())
}
