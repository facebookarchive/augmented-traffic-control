package api

import (
	"net"
	"net/http"
	"net/url"
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
	Addr             *net.TCPAddr
	ThriftUrl        *url.URL
	DBDriver, DBConn string
	V4, V6           string
	ProxyAddr        string
	AssetPath        string
}

type Server struct {
	AtcApiOptions
	Timeout   time.Duration
	Handler   http.Handler
	listener  net.Listener
	Atcd      AtcdCloser
	db        DbRunner
	bind_info *bindInfo
	assets    AssetManager
}

func ListenAndServe(options AtcApiOptions, ox ...OptionFunc) (*Server, error) {

	// apply functional arguments
	opts := defaults()
	for _, o := range ox {
		opts = o(opts)
	}

	db, err := opts.dbFactory(options.DBDriver, options.DBConn)
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
		assets: nil,
	}

	r := opts.muxFactory(srv)

	if options.AssetPath == "" {
		srv.assets = &BundleAssetManager{srv}
	} else {
		srv.assets = &LocalAssetManager{srv, options.AssetPath}
	}
	srv.setupHandlers(r)
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
	atcd := NewAtcdConn(srv.ThriftUrl)
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

func (srv *Server) setupHandlers(r *mux.Router) {
	apir := r.PathPrefix(ROOT_URL).Subrouter()
	for url, f := range APIURLMap {
		h := NewHandler(srv, f)
		apir.Handle(url, h)
		apir.Handle(url+"/", h)
	}
	r.HandleFunc("/", srv.assets.Index)
	r.HandleFunc("/static/{folder}/{name}", srv.assets.Asset)
	srv.Handler = r
}

func (srv *Server) GetAddr() string {
	return srv.listener.Addr().String()
}
