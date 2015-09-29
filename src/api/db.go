package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/facebook/augmented-traffic-control/src/atc_thrift"
	_ "github.com/mattn/go-sqlite3"
)

var (
	schema = `
	create table if not exists profiles(
		id integer primary key not null,
		name varchar,
		settings blob
	);
	`

	queries = map[string]string{
		"profiles":        `select id, name, settings from profiles order by name`,
		"profile update":  `insert or replace into profiles values (?, ?, ?)`,
		"profile delete":  `delete from profiles where id = ?`,
		"profiles max id": `select max(id) from profiles`,
	}
)

type DbRunner struct {
	db              *sql.DB
	mutex           *sync.RWMutex
	prepared        map[string]*sql.Stmt
	driver, connstr string
}

func NewDbRunner(driver, connstr string) (*DbRunner, error) {
	db, err := sql.Open(driver, connstr)
	if err != nil {
		return nil, fmt.Errorf("Could not open database connection: %v", err)
	}
	mutex := &sync.RWMutex{}
	mutex.Lock()
	defer mutex.Unlock()
	runner := &DbRunner{
		db:       db,
		mutex:    mutex,
		prepared: make(map[string]*sql.Stmt),
		driver:   driver,
		connstr:  connstr,
	}

	if err := runner.db.Ping(); err != nil {
		runner.close(false)
		return nil, fmt.Errorf("Could not open database connection: %v", err)
	}

	if _, err := runner.db.Exec(schema); err != nil {
		runner.close(false)
		return nil, fmt.Errorf("Could not create database schema: %v", err)
	}

	for name, query := range queries {
		if runner.prepared[name], err = db.Prepare(query); err != nil {
			runner.close(false)
			return nil, fmt.Errorf("Could not prepare database query %q: %v", name, err)
		}
	}
	return runner, nil
}

func (runner *DbRunner) Close() {
	runner.mutex.Lock()
	// Don't unlock the mutex again
	for _, stmt := range runner.prepared {
		stmt.Close()
	}
	runner.db.Close()
}

func (runner *DbRunner) close(lock bool) {
	if lock {
		runner.mutex.Lock()
	}
	// Don't unlock the mutex again
	for _, stmt := range runner.prepared {
		if stmt != nil {
			stmt.Close()
		}
	}
	runner.db.Close()
}

/**
*** Porcelain (public)
**/

func (runner *DbRunner) UpdateProfile(profile Profile) chan *Profile {
	result := make(chan *Profile)
	go func() {
		defer close(result)
		profile, err := runner.updateProfile(profile)
		if err == nil {
			result <- profile
		}
		runner.log(err)
	}()
	return result
}

func (runner *DbRunner) GetProfiles() chan []Profile {
	result := make(chan []Profile)
	go func() {
		defer close(result)
		profiles, err := runner.getProfiles()
		if err == nil {
			result <- profiles
		}
		runner.log(err)
	}()
	return result
}

func (runner *DbRunner) DeleteProfile(id int64) {
	go func() {
		runner.log(runner.deleteProfile(id))
	}()
}

/**
*** Plumbing (private...ish)
**/

func (runner *DbRunner) log(err error) {
	if err != nil {
		log.Printf("DB: error: %v\n", err)
	}
}

func (runner *DbRunner) prep(name string) *sql.Stmt {
	return runner.prepared[name]
}

func (runner *DbRunner) nextProfileId() (int64, error) {
	runner.mutex.RLock()
	defer runner.mutex.RUnlock()
	row := runner.prep("profiles max id").QueryRow()
	var id *int64
	// max(id) returns nil if the table is empty instead of an error
	// hence the double pointer...
	err := row.Scan(&id)
	if err == sql.ErrNoRows || id == nil {
		// No groups yet
		return 1, nil
	}
	if err != nil {
		return 0, err
	}
	// Next = Highest + 1
	return *id + 1, nil
}

func (runner *DbRunner) updateProfile(profile Profile) (*Profile, error) {
	runner.mutex.RLock()
	defer runner.mutex.RUnlock()
	var err error
	if profile.Id <= 0 {
		profile.Id, err = runner.nextProfileId()
		if err != nil {
			return nil, err
		}
	}
	var settings_bytes []byte = nil
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(profile.Shaping); err != nil {
		return nil, err
	}
	settings_bytes = buf.Bytes()
	_, err = runner.prep("profile update").Exec(profile.Id, profile.Name, settings_bytes)
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (runner *DbRunner) getProfiles() ([]Profile, error) {
	runner.mutex.RLock()
	defer runner.mutex.RUnlock()
	rows, err := runner.prep("profiles").Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	profiles := make([]Profile, 0, 100)
	for rows.Next() {
		profile, err := scanProfile(rows)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, *profile)
	}
	return profiles, nil
}

func (runner *DbRunner) deleteProfile(id int64) error {
	runner.mutex.RLock()
	defer runner.mutex.RUnlock()
	_, err := runner.prep("profile delete").Exec(id)
	return err
}

/**
*** Helpers
**/

type scanner interface {
	Scan(...interface{}) error
}

func scanProfile(sc scanner) (*Profile, error) {
	var (
		id             int64
		name           string
		settings_bytes []byte
	)
	if err := sc.Scan(&id, &name, &settings_bytes); err != nil {
		return nil, err
	}
	var shape *atc_thrift.Shaping
	if settings_bytes != nil {
		shape = new(atc_thrift.Shaping)
		err := json.Unmarshal(settings_bytes, shape)
		if err != nil {
			return nil, err
		}
	}
	return &Profile{
		Id:      id,
		Name:    name,
		Shaping: shape,
	}, nil
}
