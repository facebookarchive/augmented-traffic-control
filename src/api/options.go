package api

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

// Options contains the functional option values changeable via OptionFunc
type Options struct {
	dbFactory DatabaseFactory
}

// default options
func defaults() Options {
	return Options{
		dbFactory: standardDatabaseFactory,
	}
}
