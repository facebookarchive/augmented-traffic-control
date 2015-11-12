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

	ROOT_URL = "/api/v1"
)

type Server struct {
	Addr         string
	Timeout      time.Duration
	Handler      http.Handler
	listener     net.Listener
	Atcd         AtcdCloser
	db           *DbRunner
	thrift_proto string
	thrift_addr  string
	bind_info    *bindInfo
}

func ListenAndServe(addr, thrift_addr, thrift_proto, dbdriver, dbconn, v4, v6 string) (*Server, error) {
	db, err := NewDbRunner(dbdriver, dbconn)
	if err != nil {
		return nil, err
	}
	_, port, _ := net.SplitHostPort(addr)
	srv := &Server{
		Addr:         addr,
		listener:     nil,
		Handler:      nil,
		Timeout:      TIMEOUT,
		thrift_addr:  thrift_addr,
		thrift_proto: thrift_proto,
		Atcd:         nil,
		db:           db,
		bind_info: &bindInfo{
			ApiUrl: ROOT_URL,
			IP4:    v4,
			IP6:    v6,
			Port:   port,
		},
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
	apir := r.PathPrefix(ROOT_URL).Subrouter()
	for url, f := range API_URL_MAP {
		h := NewHandler(srv, f)
		apir.HandleFunc(url, h)
		apir.HandleFunc(url+"/", h)
	}
	r.HandleFunc("/", rootHandler(srv.bind_info))
	r.HandleFunc("/static/{folder}/{name}", cachedAssetHandler)
	srv.Handler = r
}

func (srv *Server) GetAddr() string {
	return srv.listener.Addr().String()
}
