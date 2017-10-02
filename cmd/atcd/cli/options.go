package cli

// OptionFunc is the functional argument abstraction for
// passing options to cli Execute
type OptionFunc func(cliOptions) cliOptions

// WithDatabaseFactory sets the database factory for the
// atcd system.
func WithDatabaseFactory(fact DatabaseFactory) OptionFunc {
	return func(o cliOptions) cliOptions {
		o.dbFactory = fact
		return o
	}
}

type cliOptions struct {
	dbFactory DatabaseFactory
}

func defaults() cliOptions {
	return cliOptions{
		dbFactory: standardDatabaseFactory,
	}
}
