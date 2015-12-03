package api

import (
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

var (
	// HTTP Connection timeouts for read/write
	TIMEOUT = time.Second * 30

	ROOT_URL = "/api/v1"
)

type AtcApiOptions struct {
	Addr, ThriftAddr *net.TCPAddr
	ThriftProto      string
	DBDriver, DBConn string
	V4, V6           string
	ProxyAddr        string
}

type Server struct {
	AtcApiOptions
	Timeout   time.Duration
	Handler   http.Handler
	listener  net.Listener
	Atcd      AtcdCloser
	db        *DbRunner
	bind_info *bindInfo
}

func ListenAndServe(options AtcApiOptions) (*Server, error) {
	db, err := NewDbRunner(options.DBDriver, options.DBConn)
	if err != nil {
		return nil, err
	}
	srv := &Server{
		AtcApiOptions: options,
		listener:      nil,
		Handler:       nil,
		Timeout:       TIMEOUT,
		Atcd:          nil,
		db:            db,
		bind_info: &bindInfo{
			ApiUrl: ROOT_URL,
			IP4:    options.V4,
			IP6:    options.V6,
			Port:   strconv.Itoa(options.Addr.Port),
		},
	}
	srv.setupHandlers()
	err = srv.ListenAndServe()
	if err != nil {
		return nil, err
	}
	return srv, nil
}

func (srv *Server) GetInfo(r *http.Request) APIInfo {
	return APIInfo{
		Version: VERSION,
		IPv4:    srv.bind_info.IP4,
		IPv6:    srv.bind_info.IP6,
	}
}

func (srv *Server) GetAtcd() (AtcdCloser, HttpError) {
	if srv.Atcd != nil {
		return srv.Atcd, nil
	}
	atcd := NewAtcdConn(srv.ThriftAddr, srv.ThriftProto)
	if err := atcd.Open(); err != nil {
		return nil, HttpErrorf(502, "Could not connect to atcd: %v", err)
	}
	return atcd, nil
}

func (srv *Server) ListenAndServe() error {
	var err error
	srv.listener, err = net.ListenTCP("tcp", srv.Addr)
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
		Addr:         srv.Addr.String(),
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
