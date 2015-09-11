package main

import (
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

var (
	// HTTP Connection timeouts for read/write
	TIMEOUT = time.Second * 30
)

type Server struct {
	Addr     string
	Timeout  time.Duration
	Handler  http.Handler
	listener net.Listener
	Atcd     *AtcdConn
}

func ListenAndServe(addr string) (*Server, error) {
	srv := &Server{
		Addr:     addr,
		listener: nil,
		Handler:  nil,
		Timeout:  TIMEOUT,
	}
	srv.setupHandlers()
	err := srv.ListenAndServe()
	if err != nil {
		return nil, err
	}
	return srv, nil
}

func (srv *Server) GetAtcd() (*AtcdConn, HttpError) {
	if srv.Atcd != nil {
		return srv.Atcd, nil
	}
	atcd := NewAtcdConn()
	if err := atcd.Open(); err != nil {
		return nil, Errorf(502, "Could not connect to atcd: %v", err)
	}
	return atcd, nil
}

func (srv *Server) ListenAndServe() error {
	var err error
	srv.listener, err = net.Listen("tcp", srv.Addr)
	if err != nil {
		return err
	}
	go srv.Serve()
	return nil
}

func (srv *Server) Kill() {
	if srv.listener != nil {
		// Ignore error
		srv.listener.Close()
		srv.listener = nil
	}
}

func (srv *Server) Serve() {
	_srv := &http.Server{
		Addr:         srv.Addr,
		Handler:      srv.Handler,
		ReadTimeout:  TIMEOUT,
		WriteTimeout: TIMEOUT,
	}
	_srv.Serve(srv.listener)
}

func (srv *Server) setupHandlers() {
	r := mux.NewRouter()
	for prefix, urls := range URL_MAP {
		s := r.PathPrefix(prefix).Subrouter()
		for url, f := range urls {
			h := NewHandler(srv, f)
			s.HandleFunc(url, h)
			s.HandleFunc(url+"/", h)
		}
	}
	srv.Handler = r
}
