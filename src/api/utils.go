package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/textproto"
	"runtime/debug"

	"atc_thrift"
	"github.com/gorilla/context"
)

var (
	// PanicStack is provided to reduce noise during tests.
	// DO NOT SET THIS TO FALSE
	PanicStack = true
)

type contextKey int

const (
	dbContextKey contextKey = iota
	srvContextKey
	atcdContextKey
)

// GetAtcd gets the ATC thrift client from the request context
func GetAtcd(r *http.Request) atc_thrift.Atcd {
	if rv := context.Get(r, atcdContextKey); rv != nil {
		return rv.(atc_thrift.Atcd)
	}
	log.Printf("Error: could not GetAtcd for request %v\n", r)
	return nil
}

// GetServer gets the atcApi server object from the request context
func GetServer(r *http.Request) *Server {
	if rv := context.Get(r, srvContextKey); rv != nil {
		return rv.(*Server)
	}
	log.Printf("Error: could not GetServer for request %v\n", r)
	return nil
}

// GetDB gets the database runner object from the request context
func GetDB(r *http.Request) DbRunner {
	if rv := context.Get(r, dbContextKey); rv != nil {
		return rv.(DbRunner)
	}
	log.Printf("Error: could not GetDB for request %v\n", r)
	return nil
}

// HandlerFunc is our custom handler function type for supporting JSON
// based HTTP endpoints.
type HandlerFunc func(http.ResponseWriter, *http.Request) (interface{}, HttpError)

// ErrorFunc is an HTTP function which returns a special, HTTP speific error
type ErrorFunc func(http.ResponseWriter, *http.Request) HttpError

// NewHandler adapts our custom HandlerFunc to an net/http.Handler.
func NewHandler(srv *Server, f HandlerFunc) http.Handler {
	return context.ClearHandler(ErrorHandler(ContextHandler(srv, JSONHandler(f))))
}

// ContextHandler is the handler wrapper for populating the request context.
func ContextHandler(srv *Server, f ErrorFunc) ErrorFunc {
	return func(w http.ResponseWriter, r *http.Request) HttpError {
		atcd, err := srv.GetAtcd()
		if err != nil {
			return err
		}
		defer atcd.Close()
		context.Set(r, atcdContextKey, atcd)
		context.Set(r, dbContextKey, srv.db)
		context.Set(r, srvContextKey, srv)
		return f(w, r)
	}
}

// ErrorHandler wraps the  given function in error recovery, logging, and
// translation logic.
func ErrorHandler(f ErrorFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			e := recover()
			if e != nil {
				log.Printf("panic: %v\n", e)
				if PanicStack {
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

// JSONHandler wraps the handler function in a JSON-aware
// http handler.
func JSONHandler(f HandlerFunc) ErrorFunc {
	return func(w http.ResponseWriter, r *http.Request) HttpError {
		v, httpErr := f(w, r)
		if httpErr != nil {
			return httpErr
		}
		if v == nil {
			w.WriteHeader(204)
			return nil
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)

		err := json.NewEncoder(w).Encode(v)
		if err != nil {
			return HttpErrorf(500, "Bad JSON: %v", err)
		}
		return nil
	}
}

// CORS adds cross-site access control headers to the response.
func CORS(w http.ResponseWriter, methods ...string) {
	w.Header().Set("Accept", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Del("Access-Control-Allow-Methods")
	for _, s := range methods {
		w.Header().Add("Access-Control-Allow-Methods", s)
	}
}

// GetClientAddr gets the IP address of the client
func GetClientAddr(r *http.Request) (net.IP, HttpError) {
	srv := GetServer(r)
	return getProxiedClientAddr(srv, r)
}

func getProxiedClientAddr(srv *Server, r *http.Request) (net.IP, HttpError) {
	srvProxy := srv.ProxyAddr
	remoteAddr, _, _ := net.SplitHostPort(r.RemoteAddr)
	realIps, ok := r.Header[textproto.CanonicalMIMEHeaderKey("X_HTTP_REAL_IP")]
	proxyRequest := ok && len(realIps) == 1
	proxyServer := srvProxy != ""
	if proxyRequest && proxyServer {
		// Server and client were both proxied
		if srvProxy == remoteAddr {
			if ip := net.ParseIP(realIps[0]); ip != nil {
				return ip, nil
			}

			Log.Printf("Could not parse IP [X_HTTP_REAL_IP]: %q",
				realIps[0])
			return nil, ServerError
		}

		Log.Printf("Unauthorized proxied request from %s on behalf of %v",
			remoteAddr, realIps[0])
		return nil, HttpErrorf(http.StatusBadRequest,
			"Invalid proxy address")

	}

	if proxyRequest {
		// Client was proxied but the server wasn't
		Log.Printf("Unexpected proxied request from %s on behalf of %v",
			remoteAddr, realIps[0])
		return nil, HttpErrorf(http.StatusBadRequest, "Invalid proxy address")
	}

	if proxyServer {
		if len(realIps) > 1 {
			Log.Printf("Multiple X_HTTP_REAL_IP headers: %v", realIps)
			return nil, HttpErrorf(http.StatusBadRequest,
				"Multiple X_HTTP_REAL_IP headers")
		}

		// Server was proxied, but the client wasn't
		Log.Printf("Unexpected non-proxied request from %s", remoteAddr)
		return nil, HttpErrorf(http.StatusBadRequest,
			"Missing proxied address")
	}

	// Neither the server nor the client were proxied.
	if ip := net.ParseIP(remoteAddr); ip != nil {
		return ip, nil
	}

	Log.Printf("Could not parse IP [request.RemoteAddr]: %q",
		realIps[0])
	return nil, ServerError

}
