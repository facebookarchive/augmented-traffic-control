package daemon

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/facebook/augmented-traffic-control/src/atc_thrift"
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
		"group":        `select id, secret, tc, timeout from shapinggroups where id=?`,
		"group update": `insert or replace into shapinggroups values (?, ?, ?, ?)`,
		"group delete": `delete from shapinggroups where id = ?`,
		"groups":       `select id, secret, tc, timeout from shapinggroups`,
		"group max id": `select max(id) from shapinggroups`,

		"member":           `select addr, group_id from groupmembers where addr = ?`,
		"member update":    `insert or replace into groupmembers values (?, ?)`,
		"member delete":    `delete from groupmembers where addr = ?`,
		"members in group": `select addr from groupmembers where group_id = ?`,

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
	addr     net.IP
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

func (runner *DbRunner) GetGroup(id int64) chan *DbGroup {
	result := make(chan *DbGroup)

	go func() {
		defer close(result)
		group, err := runner.getGroup(id)
		if err == nil {
			result <- group
		}
		runner.log(err)
	}()
	return result
}

func (runner *DbRunner) GetAllGroups() chan []DbGroup {
	result := make(chan []DbGroup)
	go func() {
		defer close(result)
		groups, err := runner.getAllGroups()
		if err == nil {
			result <- groups
		}
		runner.log(err)
	}()
	return result
}

func (runner *DbRunner) DeleteGroup(id int64) {
	go func() {
		err := runner.deleteGroup(id)
		runner.log(err)
	}()
}

func (runner *DbRunner) UpdateGroup(group DbGroup) chan *DbGroup {
	result := make(chan *DbGroup)
	go func() {
		defer close(result)
		group, err := runner.updateGroup(group)
		if err == nil {
			result <- group
		}
		runner.log(err)
	}()
	return result
}

func (runner *DbRunner) GetMember(addr net.IP) chan *DbMember {
	result := make(chan *DbMember)
	go func() {
		defer close(result)
		member, err := runner.getMember(addr)
		if err == nil {
			result <- member
		}
		runner.log(err)
	}()
	return result
}

func (runner *DbRunner) UpdateMember(member DbMember) chan *DbMember {
	result := make(chan *DbMember)
	go func() {
		defer close(result)
		member, err := runner.updateMember(member)
		if err == nil {
			result <- member
		}
		runner.log(err)
	}()
	return result
}

func (runner *DbRunner) DeleteMember(addr net.IP) {
	go func() {
		err := runner.deleteMember(addr)
		runner.log(err)
	}()
}

func (runner *DbRunner) GetMembersOf(id int64) chan []net.IP {
	result := make(chan []net.IP)
	go func() {
		defer close(result)
		members, err := runner.getMembersOf(id)
		if err == nil {
			result <- members
		}
		runner.log(err)
	}()
	return result
}

func (runner *DbRunner) Cleanup() {
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
}

/**
*** Plumbing (private...ish)
**/

func (runner *DbRunner) log(err error) {
	if err != nil {
		Log.Printf("DB: error: %v\n", err)
	}
}

func (runner *DbRunner) prep(name string) *sql.Stmt {
	return runner.prepared[name]
}

func (runner *DbRunner) getGroup(id int64) (*DbGroup, error) {
	runner.mutex.RLock()
	defer runner.mutex.RUnlock()
	row := runner.prep("group").QueryRow(id)
	grp, err := scanGroup(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return grp, err
}

func (runner *DbRunner) getAllGroups() ([]DbGroup, error) {
	runner.mutex.RLock()
	defer runner.mutex.RUnlock()
	rows, err := runner.prep("groups").Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	grps := make([]DbGroup, 0, 100)
	for rows.Next() {
		grp, err := scanGroup(rows)
		if err != nil {
			return nil, err
		}
		grps = append(grps, *grp)
	}
	return grps, nil
}

func (runner *DbRunner) deleteGroup(id int64) error {
	runner.mutex.RLock()
	defer runner.mutex.RUnlock()
	_, err := runner.prep("group delete").Exec(id)
	return err
}

func (runner *DbRunner) nextGroupId() (int64, error) {
	runner.mutex.RLock()
	defer runner.mutex.RUnlock()
	row := runner.prep("group max id").QueryRow()
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

func (runner *DbRunner) updateGroup(group DbGroup) (*DbGroup, error) {
	runner.mutex.RLock()
	defer runner.mutex.RUnlock()
	var err error
	if group.id == 0 {
		group.id, err = runner.nextGroupId()
		if err != nil {
			return nil, err
		}
	}
	var tc_bytes []byte = nil
	if group.tc != nil {
		buf := &bytes.Buffer{}
		err = json.NewEncoder(buf).Encode(group.tc)
		if err != nil {
			return nil, err
		}
		tc_bytes = buf.Bytes()
	}
	group.timeout = time.Now().Add(SHAPING_TIMEOUT_LENGTH)
	_, err = runner.prep("group update").Exec(group.id, group.secret, tc_bytes, group.timeout.Unix())
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (runner *DbRunner) getMember(addr net.IP) (*DbMember, error) {
	runner.mutex.RLock()
	defer runner.mutex.RUnlock()
	row := runner.prep("member").QueryRow(addr.String())
	member, err := scanMember(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return member, err
}

func (runner *DbRunner) updateMember(member DbMember) (*DbMember, error) {
	runner.mutex.RLock()
	defer runner.mutex.RUnlock()
	_, err := runner.prep("member update").Exec(member.addr.String(), member.group_id)
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (runner *DbRunner) deleteMember(addr net.IP) error {
	runner.mutex.RLock()
	defer runner.mutex.RUnlock()
	_, err := runner.prep("member delete").Exec(addr.String())
	return err
}

func (runner *DbRunner) getMembersOf(id int64) ([]net.IP, error) {
	runner.mutex.RLock()
	defer runner.mutex.RUnlock()
	rows, err := runner.prep("members in group").Query(id)
	if err != nil {
		return nil, err
	}
	members := make([]net.IP, 0, 10)
	for rows.Next() {
		var s string
		err := rows.Scan(&s)
		if err != nil {
			return nil, err
		}
		// ParseIP returns nil if addr isn't a valid IP.
		// This should never happen since atcd serializes IPs to the DB
		members = append(members, net.ParseIP(s))
	}
	return members, nil
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
		return nil, err
	}
	var shape *atc_thrift.Shaping
	if tc_bytes != nil {
		shape = &atc_thrift.Shaping{}
		err := json.Unmarshal(tc_bytes, shape)
		if err != nil {
			return nil, err
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
	return &DbMember{
		// ParseIP returns nil if addr isn't a valid IP.
		// This should never happen since atcd serializes IPs to the DB
		addr:     net.ParseIP(addr),
		group_id: gid,
	}, nil
}
