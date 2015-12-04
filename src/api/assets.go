package api

import (
	"fmt"
	"html/template"
	"net"
	"net/http"

	"github.com/facebook/augmented-traffic-control/src/assets"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
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
		if p := net.ParseIP(addr); p.To4() == nil {
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

func rootHandler(srv *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := assets.Asset("static/index.htm")
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(404)
			return
		}
		context.Set(r, srv_context_key, srv)
		tmpl, err := template.New("root").Parse(string(data))
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(500)
			return
		}
		tmpl_data, err := srv.bind_info.templateFor(r)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(400)
			return
		}
		w.WriteHeader(200)
		tmpl.Execute(w, tmpl_data)
	}
}

func cachedAssetHandler(w http.ResponseWriter, r *http.Request) {
	name, ok := mux.Vars(r)["name"]
	if !ok {
		w.WriteHeader(404)
		return
	}
	folder, ok := mux.Vars(r)["folder"]
	if !ok {
		w.WriteHeader(404)
		return
	}
	data, err := assets.Asset(fmt.Sprintf("static/%s/%s", folder, name))
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(404)
		return
	}
	w.WriteHeader(200)
	_, err = w.Write(data)
	if err != nil {
		fmt.Println(err)
		return
	}
}
