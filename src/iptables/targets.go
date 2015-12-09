package iptables

import (
	"fmt"
	"net"
)

type Target interface {
	V6() bool
	String() string
}

type IPTarget net.IP

func (t IPTarget) V6() bool {
	return net.IP(t).To4() == nil
}

func (t IPTarget) String() string {
	return net.IP(t).String()
}

type CIDRTarget struct {
	Net *net.IPNet
}

func (t CIDRTarget) V6() bool {
	return t.Net.IP.To4() == nil
}

func (t CIDRTarget) String() string {
	return t.Net.String()
}

func parseTarget(s string) (Target, error) {
	if ip := net.ParseIP(s); ip != nil {
		return IPTarget(ip), nil
	}
	if _, net, err := net.ParseCIDR(s); err == nil {
		return &CIDRTarget{net}, nil
	} else {
		return nil, fmt.Errorf("Could not parse iptables target: %q", s)
	}
}
