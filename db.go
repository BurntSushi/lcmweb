package main

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"

	_ "github.com/lib/pq"

	"github.com/gorilla/sessions"

	"github.com/BurntSushi/csql"
	"github.com/BurntSushi/migration"
	"github.com/BurntSushi/sqlsess"
)

const (
	apiVersion = 1
)

var schemaMigrations = []migration.Migrator{
	func(tx migration.LimitedTx) error {
		_, err := tx.Exec(`
			CREATE DOMAIN utctime AS timestamp with time zone
				CHECK (EXTRACT(TIMEZONE FROM VALUE) = '0');

			CREATE TABLE project (
				owner TEXT NOT NULL,
				name TEXT NOT NULL,
				display TEXT NOT NULL,
				created utctime NOT NULL,
				PRIMARY KEY (owner, name)
			);
			CREATE TABLE collaborator (
				project_owner TEXT NOT NULL,
				project_name TEXT NOT NULL,
				userid TEXT NOT NULL,
				PRIMARY KEY (project_owner, project_name, userid),
				FOREIGN KEY (project_owner, project_name)
					REFERENCES project (owner, name)
					ON DELETE CASCADE
					ON UPDATE CASCADE
			);
			CREATE TABLE document (
				project_owner TEXT NOT NULL,
				project_name TEXT NOT NULL,
				name TEXT NOT NULL,
				recorded DATE NOT NULL,
				categories TEXT NOT NULL,
				content TEXT NOT NULL,
				created_by TEXT NOT NULL,
				created utctime NOT NULL,
				modified utctime NOT NULL,
				PRIMARY KEY (name, recorded),
				FOREIGN KEY (project_owner, project_name)
					REFERENCES project (owner, name)
					ON DELETE CASCADE
					ON UPDATE CASCADE
			);
			CREATE TABLE score (
				project_owner TEXT NOT NULL,
				project_name TEXT NOT NULL,
				document_name TEXT NOT NULL,
				document_recorded DATE NOT NULL,
				word INTEGER NOT NULL,
				category TEXT NOT NULL,
				name TEXT NOT NULL,
				created_by TEXT NOT NULL,
				created utctime NOT NULL,
				PRIMARY KEY
					(project_owner, project_name,
					 document_name, document_recorded,
					 word, category),
				FOREIGN KEY (project_owner, project_name)
					REFERENCES project (owner, name)
					ON DELETE CASCADE
					ON UPDATE CASCADE,
				FOREIGN KEY (document_name, document_recorded)
					REFERENCES document (name, recorded)
					ON DELETE CASCADE
					ON UPDATE CASCADE
			);
			`)
		return err
	},
}

type lcmDB struct {
	*sql.DB
	conf configPgsql
}

func connect(conf configPgsql) *lcmDB {
	conns := fmt.Sprintf(
		"user=%s password=%s host=%s port=%d dbname=%s sslmode=disable",
		conf.User, conf.Password, conf.Host, conf.Port, conf.Database)

	pgsqlDB, err := migration.Open("postgres", conns, schemaMigrations)
	if err != nil {
		log.Fatalf("Could not connect to PostgreSQL (%s@%s/%s): %s",
			conf.User, conf.Host, conf.Database, err)
	}

	// All time operations in the database are done in UTC.
	// Times for the user (in their timezone) are mostly handled in template
	// helper functions.
	csql.Exec(pgsqlDB, "SET timezone = UTC")
	return &lcmDB{pgsqlDB, conf}
}

const (
	sessionName   = "void" // we only use one
	sessionUserId = "userid"
)

func newStore(db *lcmDB, conf configSecurity) *sqlsess.Store {
	decode64 := func(name, s string) []byte {
		dec := base64.StdEncoding
		bs, err := dec.DecodeString(s)
		if err != nil {
			log.Fatal("Could not decode %s key: %s", name, err)
		}
		return bs
	}
	hashKey := decode64("hash", conf.HashKey)
	blockKey := decode64("block", conf.BlockKey)

	var err error
	store, err := sqlsess.Open(db.DB)
	if err != nil {
		log.Fatalf("Could not open session storage: %s", err)
	}
	store.SetKeys(hashKey, blockKey)
	return store
}

func sessGet(sess *sessions.Session, key string) string {
	var val string
	var v interface{}
	var ok bool

	if v, ok = sess.Values[key]; !ok {
		return ""
	}
	if val, ok = v.(string); !ok {
		return ""
	}
	return val
}
