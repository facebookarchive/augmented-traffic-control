package daemon

import (
	. "github.com/facebook/augmented-traffic-control/src/log"
)

var Log *LogMux

func init() {
	Log = NewMux(Syslog(), Stdlog())
}
