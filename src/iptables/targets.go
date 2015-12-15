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
		// ParseIP returns IPv4 addresses as 0::ff:ff:w:x:y:z for some reason
		// this causes issues with unit tests since they compare byte arrays
		// work around by truncating when the address is a v4 address.
		if v4 := ip.To4(); v4 == nil {
			return IPTarget(ip), nil
		} else {
			return IPTarget(v4), nil
		}
	}
	if _, net, err := net.ParseCIDR(s); err == nil {
		return &CIDRTarget{net}, nil
	} else {
		return nil, fmt.Errorf("Could not parse iptables target: %q", s)
	}
}
