package main

import (
	"database/sql"
	"log"
)

// migrations is a list of migration functions that take the database from
// version N to N + 1, where N is the index of the slice.
var migrations = []func(*sql.Tx) error{
	migrateWrap(migrate0to1, 1),
}

// version returns the current version of the database. If an error occurs,
// then version 0 is returned.
func (db *lcmDB) version() int {
	var ver int
	r := db.QueryRow("SELECT value FROM meta WHERE name = $1", "version")
	if err := r.Scan(&ver); err != nil {
		return 0
	}
	return ver
}

// migrate brings the current database we're connected to up to the current
// API version. All migrations steps are executed in a single transaction so
// that a failure will bring us back to last known working database.
func (db *lcmDB) migrate() {
	// If the current version of the database is zero, then it better be
	// empty.
	version := db.version()
	if version == 0 && !db.isEmpty() {
		log.Fatal("Database corrupted. Expected it to be empty but it's not.")
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Could not start DB transaction: %s", err)
	}
	for i := version; i < apiVersion; i++ {
		if err := migrations[i](tx); err != nil {
			if err2 := tx.Rollback(); err2 != nil {
				log.Fatalf(
					"When migrating from %d to %d, got error '%s' and "+
						"got error '%s' after trying to rollback.",
					i, i+1, err, err2)
			}
			log.Println(db.Close())
			log.Fatalf(
				"Tried migrating from %d to %d, got an error and successfully "+
					"rolled back: %s --- %T", i, i+1, err, err)
		}
	}
	if err := tx.Commit(); err != nil {
		log.Fatalf("Tried commiting migration from %d to %d, but errored: %s",
			version, apiVersion, err)
	}
	if newVersion := db.version(); newVersion != apiVersion {
		log.Fatalf("After successfully migrating, expected version number %d "+
			"but got %d.", apiVersion, newVersion)
	}
}

// isEmpty determines if the database is empty by returning true if and only
// if the number of tables in the database is zero.
// This is used upon initial setup of the database. Namely, if the version
// number is 0, then there should be zero tables. Otherwise, the database is
// corrupted.
func (db *lcmDB) isEmpty() bool {
	var count int

	row := db.QueryRow(`
		SELECT COUNT(*) AS count FROM information_schema.tables
		WHERE table_catalog = $1 AND table_schema = 'public'
	`, db.conf.Database)

	if err := row.Scan(&count); err != nil {
		log.Printf("SELECT ... TABLES: %s", err)
		return false
	}
	return count == 0
}

func migrate0to1(tx *sql.Tx) {
	mustExec(tx, `
		CREATE TABLE meta (
			name varchar (255) PRIMARY KEY,
			value varchar (1000) NOT NULL
		)
	`)

	mustExec(tx, `
		INSERT INTO meta (name, value) VALUES ('version', '1')
	`)

	mustExec(tx, `
		CREATE TABLE session (
			sessionid bytea NOT NULL,
			userid varchar (255) NOT NULL,
			session_name varchar (255) NOT NULL,
			key varchar (255) NOT NULL,
			value varchar (1000) NOT NULL,
			PRIMARY KEY (sessionid, userid, session_name, key)
		)
	`)

	mustExec(tx, `
		CREATE TABLE password (
			userno smallint PRIMARY KEY,
			password varchar (255) NOT NULL,
			salt1 varchar (255) NOT NULL,
			salt2 varchar (255) NOT NULL
		)
	`)
}

func updateVersion(tx *sql.Tx, newv int) error {
	_, err := tx.Exec("UPDATE meta SET value = $1 WHERE name = 'version'", newv)
	return err
}

func migrateWrap(migrate func(*sql.Tx), newVersion int) func(*sql.Tx) error {
	return func(tx *sql.Tx) (err error) {
		defer func() {
			if r := recover(); r != nil {
				var ok bool
				if err, ok = r.(error); ok {
					return
				}
			}
		}()
		migrate(tx)
		return updateVersion(tx, newVersion)
	}
}
