package api

import (
	"github.com/facebook/augmented-traffic-control/src/atc_thrift"
)

// Used for server info request
type ServerInfo struct {
	Api  APIInfo    `json:"atc_api"`
	Atcd DaemonInfo `json:"atc_daemon"`
}

type APIInfo struct {
	Version string `json:"version"`
	IPv4    string `json:"ipv4_addr"`
	IPv6    string `json:"ipv6_addr"`
}

type DaemonInfo struct {
	Platform string `json:"platform"`
	Version  string `json:"version"`
}

// used by group creation response
type CreatedGroup struct {
	*atc_thrift.ShapingGroup
	Token string `json:"token"`
}

type GroupToken struct {
	Token string `json:"token"`
	Id    int64  `json:"id"`
}

// used for join and leave responses
type MemberResponse struct {
	Member string `json:"member"`
	Id     int64  `json:"id"`
}

// just the ...
// used by unshaping, join, leave requests
type Token struct {
	Token string `json:"token"`
}

// used by shaping requests
type TokenShaping struct {
	Token   string              `json:"token"`
	Shaping *atc_thrift.Shaping `json:"shaping"`
}

// used by shaping responses
type GroupShaping struct {
	Id      int64               `json:"id"`
	Shaping *atc_thrift.Shaping `json:"shaping"`
}

// used by profile creation request
type ProfileRequest struct {
	Name    string              `json:"name"`
	Shaping *atc_thrift.Shaping `json:"shaping"`
}

// used by profile creation response
type Profile struct {
	Id      int64               `json:"id"`
	Name    string              `json:"name"`
	Shaping *atc_thrift.Shaping `json:"shaping"`
}

// used by profile index response
type Profiles struct {
	Profiles []Profile `json:"profiles"`
}
