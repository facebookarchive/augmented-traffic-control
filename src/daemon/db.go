package daemon

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/facebook/augmented-traffic-control/src/atc_thrift"
	"github.com/facebook/augmented-traffic-control/src/iptables"

	_ "github.com/mattn/go-sqlite3"
)

const (
	SHAPING_TIMEOUT_LENGTH = 24 * time.Hour
)

var (
	schema = `
	create table if not exists shapinggroups(
		id integer primary key not null,
		secret varchar,
		tc blob,
		timeout integer
	);

	create table if not exists groupmembers(
		addr varchar primary key not null,
		group_id integer not null,
		foreign key(group_id) references shapinggroups(id) on delete cascade
	);
	`

	queries = map[string]string{
		"health": `select 1`,

		"group":        `select id, secret, tc, timeout from shapinggroups where id=?`,
		"group update": `insert or replace into shapinggroups values (?, ?, ?, ?)`,
		"group delete": `delete from shapinggroups where id = ?`,
		"groups":       `select id, secret, tc, timeout from shapinggroups`,
		"group max id": `select max(id) from shapinggroups`,

		"member":           `select addr, group_id from groupmembers where addr = ?`,
		"member update":    `insert or replace into groupmembers values (?, ?)`,
		"member delete":    `delete from groupmembers where addr = ?`,
		"members in group": `select addr from groupmembers where group_id = ?`,
		"all members":      `select addr, group_id from groupmembers order by group_id`,

		"empty group cleanup": `delete from shapinggroups where id not in (select distinct(group_id) from groupmembers)`,
		"old group cleanup":   `delete from shapinggroups where timeout < ?`,
	}
)

type DbGroup struct {
	id      int64
	secret  string
	tc      *atc_thrift.Shaping
	timeout time.Time
}

type DbMember struct {
	addr     iptables.Target
	group_id int64
}

type DbRunner struct {
	db              *sql.DB
	mutex           *sync.RWMutex
	prepared        map[string]*sql.Stmt
	shaping_timeout time.Duration
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
		runner.Close()
		return nil, fmt.Errorf("Could not open database connection: %v", err)
	}

	if _, err := runner.db.Exec(schema); err != nil {
		runner.Close()
		return nil, fmt.Errorf("Could not create database schema: %v", err)
	}

	for name, query := range queries {
		if runner.prepared[name], err = db.Prepare(query); err != nil {
			runner.Close()
			return nil, fmt.Errorf("Could not prepare database query %q: %v", name, err)
		}
	}
	return runner, nil
}

func (runner *DbRunner) Close() {
	runner.close(true)
}

func (runner *DbRunner) close(lock bool) {
	if lock {
		runner.mutex.Lock()
	}
	// Don't unlock the mutex again
	for _, stmt := range runner.prepared {
		stmt.Close()
	}
	runner.db.Close()
}

/**
*** Porcelain (public)
**/

func (runner *DbRunner) GetGroup(id int64) (*DbGroup, error) {
	runner.mutex.RLock()
	defer runner.mutex.RUnlock()
	row := runner.prep("group").QueryRow(id)
	grp, err := scanGroup(row)
	if errors.Cause(err) == sql.ErrNoRows {
		return nil, nil
	}
	runner.log(err)
	return grp, err
}

func (runner *DbRunner) GetAllGroups() (chan *DbGroup, error) {
	if err := runner.healthCheck(); err != nil {
		return nil, err
	}

	result := make(chan *DbGroup)
	go func() {
		defer close(result)
		runner.log(runner.getAllGroups(result))
	}()
	return result, nil
}

func (runner *DbRunner) DeleteGroup(id int64) error {
	runner.mutex.RLock()
	defer runner.mutex.RUnlock()
	_, err := runner.prep("group delete").Exec(id)
	runner.log(err)
	return err
}

func (runner *DbRunner) UpdateGroup(group DbGroup) (*DbGroup, error) {
	runner.mutex.RLock()
	var err error
	defer runner.mutex.RUnlock()
	if group.id == 0 {
		group.id, err = runner.nextGroupId()
		if err != nil {
			runner.log(err)
			return nil, err
		}
	}
	var tc_bytes []byte = nil
	if group.tc != nil {
		buf := &bytes.Buffer{}
		err = json.NewEncoder(buf).Encode(group.tc)
		if err != nil {
			runner.log(err)
			return nil, err
		}
		tc_bytes = buf.Bytes()
	}
	group.timeout = time.Now().Add(SHAPING_TIMEOUT_LENGTH)
	//fmt.Printf("%T %T %T %T", group.id, group.secret, tc_bytes, group.timeout.Unix())
	_, err = runner.prep("group update").Exec(group.id, group.secret, tc_bytes, group.timeout.Unix())
	if err != nil {
		runner.log(err)
		return nil, err
	}
	return &group, nil
}

func (runner *DbRunner) GetMember(addr iptables.Target) (*DbMember, error) {
	runner.mutex.RLock()
	defer runner.mutex.RUnlock()
	row := runner.prep("member").QueryRow(addr.String())
	member, err := scanMember(row)
	if errors.Cause(err) == sql.ErrNoRows {
		return nil, nil
	}
	runner.log(err)
	return member, err
}

func (runner *DbRunner) UpdateMember(member DbMember) (*DbMember, error) {
	runner.mutex.RLock()
	defer runner.mutex.RUnlock()
	_, err := runner.prep("member update").Exec(member.addr.String(), member.group_id)
	if err != nil {
		runner.log(err)
		return nil, err
	}
	return &member, nil
}

func (runner *DbRunner) DeleteMember(addr iptables.Target) error {
	runner.mutex.RLock()
	defer runner.mutex.RUnlock()
	_, err := runner.prep("member delete").Exec(addr.String())
	runner.log(err)
	return err
}

func (runner *DbRunner) GetMembersOf(id int64) (chan iptables.Target, error) {
	if err := runner.healthCheck(); err != nil {
		return nil, err
	}

	result := make(chan iptables.Target)
	go func() {
		defer close(result)
		runner.log(runner.getMembersOf(id, result))
	}()
	return result, nil
}

func (runner *DbRunner) GetAllMembers() (chan *DbMember, error) {
	if err := runner.healthCheck(); err != nil {
		return nil, err
	}

	results := make(chan *DbMember)
	go func() {
		defer close(results)
		runner.log(runner.getAllMembers(results))
	}()
	return results, nil
}

func (runner *DbRunner) Cleanup() error {
	if err := runner.healthCheck(); err != nil {
		return err
	}

	go func() {
		n, err := runner.cleanupEmptyGroups()
		runner.log(err)
		if n > 0 {
			Log.Printf("DB: Cleaned %d empty groups\n", n)
		}
		n, err = runner.cleanupOldGroups()
		runner.log(err)
		if n > 0 {
			Log.Printf("DB: Cleaned %d expired groups\n", n)
		}
	}()
	return nil
}

/**
*** Plumbing (private...ish)
**/

func (runner *DbRunner) healthCheck() error {
	runner.mutex.RLock()
	defer runner.mutex.RUnlock()
	row := runner.prep("health").QueryRow()
	var i int64
	if err := row.Scan(&i); err != nil {
		return fmt.Errorf("Unhealthy database: %v", err)
	}
	if i != 1 {
		return fmt.Errorf("Unhealthy database")
	}
	return nil
}

func (runner *DbRunner) log(err error) {
	if err != nil {
		Log.Printf("DB: error: %v\n", err)
	}
}

func (runner *DbRunner) prep(name string) *sql.Stmt {
	return runner.prepared[name]
}

func (runner *DbRunner) getAllGroups(grps chan *DbGroup) error {
	runner.mutex.RLock()
	defer runner.mutex.RUnlock()
	rows, err := runner.prep("groups").Query()
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		grp, err := scanGroup(rows)
		if err != nil {
			return errors.Wrap(err, "error scanning group")
		}
		grps <- grp
	}
	return nil
}

func (runner *DbRunner) nextGroupId() (int64, error) {
	runner.mutex.RLock()
	defer runner.mutex.RUnlock()
	row := runner.prep("group max id").QueryRow()
	var id *int64
	// max(id) returns nil if the table is empty instead of an error
	// hence the double pointer...
	err := row.Scan(&id)
	if errors.Cause(err) == sql.ErrNoRows || id == nil {
		// No groups yet
		return 1, nil
	}
	if err != nil {
		return 0, err
	}
	// Next = Highest + 1
	return *id + 1, nil
}

func (runner *DbRunner) getMembersOf(id int64, members chan iptables.Target) error {
	runner.mutex.RLock()
	defer runner.mutex.RUnlock()
	rows, err := runner.prep("members in group").Query(id)
	if err != nil {
		return err
	}
	for rows.Next() {
		var s string
		err := rows.Scan(&s)
		if err != nil {
			return err
		}
		tgt, err := iptables.ParseTarget(s)
		if err != nil {
			return fmt.Errorf("Could not load target from db: %v", err)
		}
		members <- tgt
	}
	return nil
}

func (runner *DbRunner) getAllMembers(members chan *DbMember) error {
	runner.mutex.RLock()
	defer runner.mutex.RUnlock()
	rows, err := runner.prep("all members").Query()
	if err != nil {
		return err
	}
	for rows.Next() {
		var addr string
		var id int64
		err := rows.Scan(&addr, &id)
		if err != nil {
			return err
		}
		tgt, err := iptables.ParseTarget(addr)
		if err != nil {
			return fmt.Errorf("Could not load target from db: %v", err)
		}
		members <- &DbMember{tgt, id}
	}
	return nil
}

func (runner *DbRunner) cleanupOldGroups() (int64, error) {
	runner.mutex.RLock()
	defer runner.mutex.RUnlock()
	res, err := runner.prep("old group cleanup").Exec(time.Now().Unix())
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (runner *DbRunner) cleanupEmptyGroups() (int64, error) {
	runner.mutex.RLock()
	defer runner.mutex.RUnlock()
	res, err := runner.prep("empty group cleanup").Exec()
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return n, nil
}

/**
*** Helpers
**/

type scanner interface {
	Scan(...interface{}) error
}

func scanGroup(sc scanner) (*DbGroup, error) {
	var (
		id       int64
		tc_bytes []byte
		timeout  int64
		secret   string
	)
	if err := sc.Scan(&id, &secret, &tc_bytes, &timeout); err != nil {
		return nil, errors.Wrap(err, "error scanning columns of group")
	}
	var shape *atc_thrift.Shaping
	if tc_bytes != nil && len(tc_bytes) > 2 { // anything under {} is invalid json
		shape = &atc_thrift.Shaping{}
		err := json.Unmarshal(tc_bytes, shape)
		if err != nil {
			return nil, errors.Wrap(err, "error unmarshalling json")
		}
	}
	return &DbGroup{
		id:      id,
		secret:  secret,
		tc:      shape,
		timeout: time.Unix(timeout, 0),
	}, nil
}

func scanMember(sc scanner) (*DbMember, error) {
	var (
		addr string
		gid  int64
	)
	if err := sc.Scan(&addr, &gid); err != nil {
		return nil, err
	}
	tgt, err := iptables.ParseTarget(addr)
	if err != nil {
		return nil, fmt.Errorf("Could not load target from db: %v", err)
	}
	return &DbMember{
		addr:     tgt,
		group_id: gid,
	}, nil
}
