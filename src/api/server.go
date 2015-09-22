package api

import (
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

var (
	// HTTP Connection timeouts for read/write
	TIMEOUT = time.Second * 30

	ServerData = serverData{
		ApiUrl: "/api/v1/",
	}
)

type serverData struct {
	ApiUrl string
}

type Server struct {
	Addr         string
	Timeout      time.Duration
	Handler      http.Handler
	listener     net.Listener
	Atcd         AtcdCloser
	db           *DbRunner
	thrift_proto string
	thrift_addr  string
}

func ListenAndServe(addr, thrift_addr, thrift_proto, dbdriver, dbconn string) (*Server, error) {
	db, err := NewDbRunner(dbdriver, dbconn)
	if err != nil {
		return nil, err
	}
	srv := &Server{
		Addr:         addr,
		listener:     nil,
		Handler:      nil,
		Timeout:      TIMEOUT,
		thrift_addr:  thrift_addr,
		thrift_proto: thrift_proto,
		Atcd:         nil,
		db:           db,
	}
	srv.setupHandlers()
	err = srv.ListenAndServe()
	if err != nil {
		return nil, err
	}
	return srv, nil
}

func (srv *Server) GetAtcd() (AtcdCloser, HttpError) {
	if srv.Atcd != nil {
		return srv.Atcd, nil
	}
	atcd := NewAtcdConn(srv.thrift_addr, srv.thrift_proto)
	if err := atcd.Open(); err != nil {
		return nil, HttpErrorf(502, "Could not connect to atcd: %v", err)
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
		srv.db.Close()
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
	apir := r.PathPrefix(ServerData.ApiUrl).Subrouter()
	for url, f := range API_URL_MAP {
		h := NewHandler(srv, f)
		apir.HandleFunc(url, h)
		apir.HandleFunc(url+"/", h)
	}
	r.HandleFunc("/", rootHandler)
	r.HandleFunc("/static/{folder}/{name}", diskAssetHandler)
	srv.Handler = r
}

func (srv *Server) GetAddr() string {
	return srv.listener.Addr().String()
}
