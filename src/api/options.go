package api

import (
  "github.com/gorilla/mux"
)

// OptionFunc is the functional argument abstraction for
// passing options to the api server
type OptionFunc func(Options) Options

// WithDatabaseFactory sets the database factory for the
// atc_api system.
func WithDatabaseFactory(fact DatabaseFactory) OptionFunc {
	return func(o Options) Options {
		o.dbFactory = fact
		return o
	}
}

// A MuxFactory is a function which builds a router
type MuxFactory func(s *Server) *mux.Router

// WithMuxFactory allows overriding the initial HTTP router
func WithMuxFactory(fact MuxFactory) OptionFunc {
  return func(o Options) Options {
    o.muxFactory = fact
    return o
  }
}

// Options contains the functional option values changeable via OptionFunc
type Options struct {
  dbFactory  DatabaseFactory
  muxFactory MuxFactory
}

// default options
func defaults() Options {
	return Options{
    dbFactory:  standardDatabaseFactory,
    muxFactory: newMuxRouter,
	}
}

// default mux impl
func newMuxRouter(s *Server) *mux.Router {
  return mux.NewRouter()
}
