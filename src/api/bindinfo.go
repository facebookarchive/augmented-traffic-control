package api

import (
	"net"
	"net/http"
)

type bindInfo struct {
	ApiUrl string
	IP4    string
	IP6    string
	Port   string
}

func (info bindInfo) getPrimarySecondaryAddrs(r *http.Request) (primary, secondary string, err HttpError) {
	addr, err := GetClientAddr(r)
	if err != nil {
		return "", "", err
	}
	if info.IP6 != "" && info.IP4 != "" {
		// server is dual-stack
		if addr.To4() == nil {
			// client is ipv6
			primary = info.IP6
			secondary = info.IP4
		} else {
			// client is ipv4
			primary = info.IP4
			secondary = info.IP6
		}
	} else if info.IP6 == "" {
		// server is IPv4 single-stack
		primary = info.IP4
	} else if info.IP4 == "" {
		// server is IPv6 single-stack
		primary = info.IP6
	} else {
		// IPv6 and IPv4 are nil.
		// Should be prohibited by CLI argument validation.
		panic("Neither IPv6 nor IPv4 are set!")
	}
	return
}

type templateData struct {
	ApiUrl    string
	Primary   string
	Secondary string
}

func (info *bindInfo) templateFor(r *http.Request) (*templateData, HttpError) {
	data := &templateData{
		ApiUrl: info.ApiUrl,
	}
	var err HttpError
	data.Primary, data.Secondary, err = info.getPrimarySecondaryAddrs(r)
	if err != nil {
		return nil, err
	}
	// If the user didn't provide one of the two addresses, we pass the UI an
	// empty string.
	if data.Primary != "" {
		data.Primary = net.JoinHostPort(data.Primary, info.Port)
	}
	if data.Secondary != "" {
		data.Secondary = net.JoinHostPort(data.Secondary, info.Port)
	}
	return data, nil
}
