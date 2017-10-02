package cli

import (
	"fmt"

	"github.com/facebook/augmented-traffic-control/src/daemon"
)

// The DatabaseFactory is an abstraction that allows
// main entrypoints to override how databases backends
// are constructed.
type DatabaseFactory func(driver, conn string) (db daemon.DbRunner, err error)

func standardDatabaseFactory(
	driver string,
	conn string) (db daemon.DbRunner, err error) {

	// Setup the database
	switch driver {
	case "mysql":
		fallthrough
	case "postgres":
		fallthrough
	case "sqlite3":
		db, err = daemon.NewSqlRunner(driver, conn)
	case "memory":
		db, err = daemon.NewMemoryRunner()
	default:
		err = fmt.Errorf("Unsupported db driver %s", driver)
	}

	return
}
