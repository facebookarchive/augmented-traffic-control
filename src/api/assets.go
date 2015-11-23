package api

import (
	"fmt"
	"html/template"
	"net"
	"net/http"

	"github.com/gorilla/mux"
)

type bindInfo struct {
	ApiUrl string
	IP4    string
	IP6    string
	Port   string
}

type templateData struct {
	ApiUrl    string
	Primary   string
	Secondary string
}

func (info *bindInfo) templateFor(r *http.Request) *templateData {
	data := &templateData{
		ApiUrl: info.ApiUrl,
	}
	addr, _, _ := net.SplitHostPort(r.RemoteAddr)
	if info.IP6 != "" && info.IP4 != "" {
		// server is dual-stack
		if p := net.ParseIP(addr); p.To4() == nil {
			// client is ipv6
			data.Primary = info.IP6
			data.Secondary = info.IP4
		} else {
			// client is ipv4
			data.Primary = info.IP4
			data.Secondary = info.IP6
		}
	} else if info.IP6 == "" {
		// server is IPv4 single-stack
		data.Primary = info.IP4
	} else if info.IP4 == "" {
		// server is IPv6 single-stack
		data.Primary = info.IP6
	} else {
		// IPv6 and IPv4 are nil.
		// Should be prohibited by CLI argument validation.
		panic("Neither IPv6 nor IPv4 are set!")
	}
	// If the user didn't provide one of the two addresses, we pass the UI an
	// empty string.
	if data.Primary != "" {
		data.Primary = net.JoinHostPort(data.Primary, info.Port)
	}
	if data.Secondary != "" {
		data.Secondary = net.JoinHostPort(data.Secondary, info.Port)
	}
	return data
}

func rootHandler(info *bindInfo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := Asset("static/index.htm")
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(404)
			return
		}
		tmpl, err := template.New("root").Parse(string(data))
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(200)
		tmpl.Execute(w, info.templateFor(r))
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
	data, err := Asset(fmt.Sprintf("static/%s/%s", folder, name))
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
