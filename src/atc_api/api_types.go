package main

import (
	"github.com/facebook/augmented-traffic-control/src/atc_thrift"
)

type ServerInfo struct {
	Api  APIInfo    `json:"atc_api"`
	Atcd DaemonInfo `json:"atc_daemon"`
}

type APIInfo struct {
	Version string `json:"version"`
}

type DaemonInfo struct {
	Platform string `json:"platform"`
	Version  string `json:"version"`
}

type GroupToken struct {
	Token string `json:"token"`
	Id    int64  `json:"id"`
}

// used for group creation request
type MemberRequest struct {
	Member string `json:"member"`
}

// used for join and leave requests
type MemberTokenRequest struct {
	Token  string `json:"token"`
	Member string `json:"member"`
}

// used for join and leave requests
type MemberResponse struct {
	Member string `json:"member"`
	Id     int64  `json:"id"`
}

// just the ...
// used by unshaping requests
type Token struct {
	Token string `json:"token"`
}

// used by shaping requests
type TokenShaping struct {
	Token   string              `json:"token"`
	Shaping *atc_thrift.Setting `json:"shaping"`
}

// used by shaping responses
type GroupShaping struct {
	Id      int64               `json:"id"`
	Shaping *atc_thrift.Setting `json:"shaping"`
}
