package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/Go-SQL-Driver/MySQL"
)

const apiVersion = 1

type lcmDB struct {
	*sql.DB
}

func connect(conf configMysql) *lcmDB {
	conns := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&autocommit=false",
		conf.User, conf.Password, conf.Host, conf.Port, conf.Database)

	mysqlDB, err := sql.Open("mysql", conns)
	if err != nil {
		log.Fatalf("Could not connect to MySQL (%s@s/%s): %s",
			conf.User, conf.Host, conf.Database, err)
	}

	db := &lcmDB{mysqlDB}
	db.migrate()

	log.Printf("Database version: %d (API: %d)", db.version(), apiVersion)
	return db
}

// version returns the current version of the database. If an error occurs,
// then version 0 is returned.
func (db *lcmDB) version() int {
	var ver int
	r := db.QueryRow("SELECT value FROM meta WHERE name = ?", "version")
	if err := r.Scan(&ver); err != nil {
		return 0
	}
	return ver
}

// migrate brings the current database we're connected to up to the current
// API version. Each migration step should be executed as a single transaction.
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
					"rolled back: %s", i, i+1, err)
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
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		log.Printf("SHOW TABLES: %s", err)
		return false
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}
	if err := rows.Err(); err != nil {
		log.Printf("SHOW TABLES iteration: %s", err)
		return false
	}
	return count == 0
}
