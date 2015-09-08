package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/facebook/augmented-traffic-control/atc/atc_thrift/atc_thrift"
	_ "github.com/mattn/go-sqlite3"
)

const (
	TIMEOUT_LENGTH = 24 * time.Hour
)

var (
	db_chan chan *sql.DB

	// Not a prepared statement!
	SHAPING_CREATE_QUERY = `
	CREATE TABLE IF NOT EXISTS ShapedAddrs(
		addr VARCHAR PRIMARY KEY NOT NULL,
		group_id INTEGER NOT NULL,
		FOREIGN KEY(group_id) REFERENCES ShapingGroups(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS ShapingGroups(
		id INTEGER PRIMARY KEY NOT NULL,
		secret VARCHAR,
		tc BLOB,
		timeout INTEGER
	);
	`

	GROUP_MAX_ID_STMT  *sql.Stmt
	GROUP_MAX_ID_QUERY = `
	SELECT MAX(id) FROM ShapingGroups
	`

	GROUP_INSERT_STMT  *sql.Stmt
	GROUP_INSERT_QUERY = `
	INSERT OR REPLACE INTO ShapingGroups values (?, ?, ?, ?)
	`

	GROUP_DELETE_STMT  *sql.Stmt
	GROUP_DELETE_QUERY = `
	DELETE FROM ShapingGroups WHERE id = ?
	`

	GROUP_SELECT_ALL_STMT  *sql.Stmt
	GROUP_SELECT_ALL_QUERY = `
	SELECT id, secret, tc, timeout FROM ShapingGroups
	`

	GROUP_SELECT_ONE_STMT  *sql.Stmt
	GROUP_SELECT_ONE_QUERY = `
	SELECT id, secret, tc, timeout FROM ShapingGroups WHERE id = ?
	`

	MEMBER_INSERT_STMT  *sql.Stmt
	MEMBER_INSERT_QUERY = `
	INSERT OR REPLACE INTO ShapedAddrs values (?, ?)
	`

	MEMBER_DELETE_STMT  *sql.Stmt
	MEMBER_DELETE_QUERY = `
	DELETE FROM ShapedAddrs WHERE addr = ?
	`

	MEMBER_SELECT_ONE_STMT  *sql.Stmt
	MEMBER_SELECT_ONE_QUERY = `
	SELECT addr, group_id FROM ShapedAddrs where addr = ?
	`

	MEMBER_SELECT_GROUP_STMT  *sql.Stmt
	MEMBER_SELECT_GROUP_QUERY = `
	SELECT addr FROM ShapedAddrs where group_id = ?
	`
)

func initDB(driver, connstr string) error {
	db_chan = make(chan *sql.DB, 1)

	db, err := sql.Open(driver, connstr)
	if err != nil {
		return err
	}

	// Double check it's working...
	if err = db.Ping(); err != nil {
		return fmt.Errorf("Could not open database connection: %v", err)
	}

	// Create the schema
	if _, err := db.Exec(SHAPING_CREATE_QUERY); err != nil {
		return fmt.Errorf("Could not create database schema: %v", err)
	}

	// Prepare statements...
	if GROUP_MAX_ID_STMT, err = db.Prepare(GROUP_MAX_ID_QUERY); err != nil {
		return fmt.Errorf("Could not prepare query: %v", err)
	}

	if GROUP_INSERT_STMT, err = db.Prepare(GROUP_INSERT_QUERY); err != nil {
		return fmt.Errorf("Could not prepare query: %v", err)
	}

	if GROUP_DELETE_STMT, err = db.Prepare(GROUP_DELETE_QUERY); err != nil {
		return fmt.Errorf("Could not prepare query: %v", err)
	}

	if GROUP_SELECT_ALL_STMT, err = db.Prepare(GROUP_SELECT_ALL_QUERY); err != nil {
		return fmt.Errorf("Could not prepare query: %v", err)
	}

	if GROUP_SELECT_ONE_STMT, err = db.Prepare(GROUP_SELECT_ONE_QUERY); err != nil {
		return fmt.Errorf("Could not prepare query: %v", err)
	}

	if MEMBER_INSERT_STMT, err = db.Prepare(MEMBER_INSERT_QUERY); err != nil {
		return fmt.Errorf("Could not prepare query: %v", err)
	}

	if MEMBER_DELETE_STMT, err = db.Prepare(MEMBER_DELETE_QUERY); err != nil {
		return fmt.Errorf("Could not prepare query: %v", err)
	}

	if MEMBER_SELECT_ONE_STMT, err = db.Prepare(MEMBER_SELECT_ONE_QUERY); err != nil {
		return fmt.Errorf("Could not prepare query: %v", err)
	}

	if MEMBER_SELECT_GROUP_STMT, err = db.Prepare(MEMBER_SELECT_GROUP_QUERY); err != nil {
		return fmt.Errorf("Could not prepare query: %v", err)
	}

	// Make the db available
	db_chan <- db
	return nil
}

func shutdownDB() error {
	// Wait for db to be available
	// Don't write it back into the chan.
	db := <-db_chan

	GROUP_MAX_ID_STMT.Close()
	GROUP_INSERT_STMT.Close()
	GROUP_DELETE_STMT.Close()
	GROUP_SELECT_ALL_STMT.Close()
	GROUP_SELECT_ONE_STMT.Close()
	close(db_chan)
	return db.Close()
}

type DbGroup struct {
	id      int64
	secret  string
	tc      *atc_thrift.Setting
	timeout time.Time
}

type DbMember struct {
	addr     string
	group_id int64
}

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
	var shape *atc_thrift.Setting
	if tc_bytes != nil {
		shape = new(atc_thrift.Setting)
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
		addr:     addr,
		group_id: gid,
	}, nil
}

func dbGetGroup(id int64) (*DbGroup, error) {
	row := GROUP_SELECT_ONE_STMT.QueryRow(id)
	grp, err := scanGroup(row)
	if err == sql.ErrNoRows {
		return nil, NoSuchItem
	}
	return grp, err
}

func dbGetAllGroups() ([]DbGroup, error) {
	rows, err := GROUP_SELECT_ALL_STMT.Query()
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

func dbDeleteGroup(id int64) error {
	_, err := GROUP_DELETE_STMT.Exec(id)
	return err
}

func dbNextId() (int64, error) {
	row := GROUP_MAX_ID_STMT.QueryRow()
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

// Update/insert
func dbUpdateGroup(grp DbGroup) (*DbGroup, error) {
	var err error
	if grp.id == 0 {
		grp.id, err = dbNextId()
		if err != nil {
			return nil, err
		}
	}
	var tc_bytes []byte = nil
	if grp.tc != nil {
		buf := &bytes.Buffer{}
		err = json.NewEncoder(buf).Encode(grp.tc)
		if err != nil {
			return nil, err
		}
		tc_bytes = buf.Bytes()
	}
	grp.timeout = time.Now().Add(TIMEOUT_LENGTH)
	_, err = GROUP_INSERT_STMT.Exec(grp.id, grp.secret, tc_bytes, grp.timeout.Unix())
	if err != nil {
		return nil, err
	}
	return &grp, nil
}

func dbGetMember(addr string) (*DbMember, error) {
	row := MEMBER_SELECT_ONE_STMT.QueryRow(addr)
	member, err := scanMember(row)
	if err == sql.ErrNoRows {
		return nil, NoSuchItem
	}
	return member, err
}

// Update/insert
func dbUpdateMember(member DbMember) (*DbMember, error) {
	_, err := MEMBER_INSERT_STMT.Exec(member.addr, member.group_id)
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func dbDeleteMember(addr string) error {
	_, err := MEMBER_DELETE_STMT.Exec(addr)
	return err
}

func dbGetMembers(id int64) ([]string, error) {
	rows, err := MEMBER_SELECT_GROUP_STMT.Query(id)
	if err != nil {
		return nil, err
	}
	members := make([]string, 0, 10)
	for rows.Next() {
		var s string
		err := rows.Scan(&s)
		if err != nil {
			return nil, err
		}
		members = append(members, s)
	}
	return members, nil
}
