package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/textproto"
	"runtime/debug"

	"github.com/facebook/augmented-traffic-control/src/atc_thrift"
	"github.com/gorilla/context"
)

var (
	// DO NOT SET THIS TO FALSE
	// Provided to reduce noise in tests.
	PANIC_STACK = true
)

type context_key int

const (
	db_context_key context_key = iota
	srv_context_key
	atcd_context_key
)

// Context functions to retrieve context from requests
func GetAtcd(r *http.Request) atc_thrift.Atcd {
	if rv := context.Get(r, atcd_context_key); rv != nil {
		return rv.(atc_thrift.Atcd)
	}
	log.Printf("Error: could not GetAtcd for request %v\n", r)
	return nil
}

func GetServer(r *http.Request) *Server {
	if rv := context.Get(r, srv_context_key); rv != nil {
		return rv.(*Server)
	}
	log.Printf("Error: could not GetServer for request %v\n", r)
	return nil
}

func GetDB(r *http.Request) *DbRunner {
	if rv := context.Get(r, db_context_key); rv != nil {
		return rv.(*DbRunner)
	}
	log.Printf("Error: could not GetDB for request %v\n", r)
	return nil
}

// Write new HTTP handlers using this type.
type HandlerFunc func(http.ResponseWriter, *http.Request) (interface{}, HttpError)

// internal to this file:
type errorFunc func(http.ResponseWriter, *http.Request) HttpError

func NewHandler(srv *Server, f HandlerFunc) http.Handler {
	return context.ClearHandler(ErrorHandler(ContextHandler(srv, JsonHandler(f))))
}

// Http handler to set context data for requests.
func ContextHandler(srv *Server, f errorFunc) errorFunc {
	return func(w http.ResponseWriter, r *http.Request) HttpError {
		atcd, err := srv.GetAtcd()
		if err != nil {
			return err
		}
		defer atcd.Close()
		context.Set(r, atcd_context_key, atcd)
		context.Set(r, db_context_key, srv.db)
		context.Set(r, srv_context_key, srv)
		return f(w, r)
	}
}

/*
Http handler that adds better error handling.
*/
func ErrorHandler(f errorFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			e := recover()
			if e != nil {
				log.Printf("panic: %v\n", e)
				if PANIC_STACK {
					debug.PrintStack()
				}
				writeError(w, ServerError)
			}
		}()
		err := f(w, r)
		if err != nil && err != NoStatus {
			if err.Status() != http.StatusNotFound {
				log.Printf("Error: %v\n", err)
			}
			writeError(w, err)
		}
	}
}

func writeError(w http.ResponseWriter, e HttpError) {
	w.WriteHeader(e.Status())
	fmt.Fprintln(w, e.Error())
}

/*
Http handler that adds JSON serialization of returned data.
*/
func JsonHandler(f HandlerFunc) errorFunc {
	return func(w http.ResponseWriter, r *http.Request) HttpError {
		v, err := f(w, r)
		if err != nil {
			return err
		}
		if v == nil {
			w.WriteHeader(204)
			return nil
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
		}
		json_err := json.NewEncoder(w).Encode(v)
		if json_err != nil {
			return HttpErrorf(500, "Bad JSON: %v", json_err)
		}
		return nil
	}
}

// Adds CORS headers to a response.
func CORS(w http.ResponseWriter, methods ...string) {
	w.Header().Set("Accept", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Del("Access-Control-Allow-Methods")
	for _, s := range methods {
		w.Header().Add("Access-Control-Allow-Methods", s)
	}
}

/*
Gets the IP address of the client
*/
func GetClientAddr(r *http.Request) (net.IP, HttpError) {
	srv := GetServer(r)
	return getProxiedClientAddr(srv, r)
}

func getProxiedClientAddr(srv *Server, r *http.Request) (net.IP, HttpError) {
	srv_proxy := srv.ProxyAddr
	remote_addr, _, _ := net.SplitHostPort(r.RemoteAddr)
	real_ips, ok := r.Header[textproto.CanonicalMIMEHeaderKey("X_HTTP_REAL_IP")]
	proxy_request := ok && len(real_ips) == 1
	proxy_server := srv_proxy != ""
	if proxy_request && proxy_server {
		// Server and client were both proxied
		if srv_proxy == remote_addr {
			if ip := net.ParseIP(real_ips[0]); ip != nil {
				return ip, nil
			} else {
				Log.Printf("Could not parse IP [X_HTTP_REAL_IP]: %q",
					real_ips[0])
				return nil, ServerError
			}
		} else {
			Log.Printf("Unauthorized proxied request from %s on behalf of %v",
				remote_addr, real_ips[0])
			return nil, HttpErrorf(http.StatusBadRequest,
				"Invalid proxy address")
		}
	} else if proxy_request {
		// Client was proxied but the server wasn't
		Log.Printf("Unexpected proxied request from %s on behalf of %v",
			remote_addr, real_ips[0])
		return nil, HttpErrorf(http.StatusBadRequest, "Invalid proxy address")
	} else if proxy_server {
		if len(real_ips) > 1 {
			Log.Printf("Multiple X_HTTP_REAL_IP headers: %v", real_ips)
			return nil, HttpErrorf(http.StatusBadRequest,
				"Multiple X_HTTP_REAL_IP headers")
		} else {

			// Server was proxied, but the client wasn't
			Log.Printf("Unexpected non-proxied request from %s", remote_addr)
			return nil, HttpErrorf(http.StatusBadRequest,
				"Missing proxied address")
		}
	} else {
		// Neither the server nor the client were proxied.
		if ip := net.ParseIP(remote_addr); ip != nil {
			return ip, nil
		} else {
			Log.Printf("Could not parse IP [request.RemoteAddr]: %q",
				real_ips[0])
			return nil, ServerError
		}
	}
}
