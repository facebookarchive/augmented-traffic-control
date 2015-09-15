package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"runtime/debug"

	"github.com/facebook/augmented-traffic-control/src/atc_thrift"
)

var (
	// DO NOT SET THIS TO FALSE
	// Provided to reduce noise in tests.
	PANIC_STACK = true
)

// Write new HTTP handlers using this type.
type HandlerFunc func(atc_thrift.Atcd, http.ResponseWriter, *http.Request) (interface{}, HttpError)

// internal to this file:
type jsonFunc func(http.ResponseWriter, *http.Request) (interface{}, HttpError)
type errorFunc func(http.ResponseWriter, *http.Request) HttpError

func NewHandler(srv *Server, f HandlerFunc) http.HandlerFunc {
	return ErrorHandler(JsonHandler(AtcdHandler(f, srv)))
}

/*
Http handler that adds atcd connection management
*/
func AtcdHandler(f HandlerFunc, srv *Server) jsonFunc {
	return func(w http.ResponseWriter, r *http.Request) (interface{}, HttpError) {
		atcd, err := srv.GetAtcd()
		if err != nil {
			return nil, err
		}
		defer atcd.Close()
		return f(atcd, w, r)
	}
}

/*
Http handler that adds better error handling.
*/
func ErrorHandler(f errorFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// FIXME: error logging
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
func JsonHandler(f jsonFunc) errorFunc {
	return func(w http.ResponseWriter, r *http.Request) HttpError {
		v, err := f(w, r)
		if err != nil {
			return err
		}
		if v == nil {
			w.WriteHeader(204)
			return nil
		} else {
			w.WriteHeader(200)
		}
		json_err := json.NewEncoder(w).Encode(v)
		if json_err != nil {
			return HttpErrorf(500, "Bad JSON: %v", json_err)
		}
		return nil
	}
}

/*
Gets the IP address of the client
*/
func GetClientAddr(r *http.Request) string {
	// FIXME: check headers for X_HTTP_CLIENT_IP or something
	// FIXME: error handling (third return value)
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}
