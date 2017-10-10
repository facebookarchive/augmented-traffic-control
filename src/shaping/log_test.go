package shaping

import (
	atc_log "github.com/facebook/augmented-traffic-control/src/log"
)

func init() {
	// Tests log to stdout, not syslog
	Log = atc_log.NewMux(atc_log.Stdlog())
}
