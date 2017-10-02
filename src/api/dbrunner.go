package api

import "fmt"

type DbRunner interface {
	Close()

	UpdateProfile(profile Profile) chan *Profile
	GetProfiles() chan []Profile
	DeleteProfile(id int64)
}

// The DatabaseFactory is an abstraction that allows
// main entrypoints to override how databases backends
// are constructed.
type DatabaseFactory func(driver, conn string) (db DbRunner, err error)

func standardDatabaseFactory(driver, conn string) (db DbRunner, err error) {

	// Setup the database
	switch driver {
	case "mysql":
		fallthrough
	case "postgres":
		fallthrough
	case "sqlite3":
		db, err = NewSqlRunner(driver, conn)
	case "memory":
		db, err = NewMemoryRunner()
	default:
		err = fmt.Errorf("Unsupported db driver %s", driver)
	}

	return
}
