package main

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"

	_ "github.com/lib/pq"

	"github.com/gorilla/sessions"

	"github.com/BurntSushi/migration"
	"github.com/BurntSushi/sqlsess"
)

const (
	apiVersion = 1
)

var schemaMigrations = []migration.Migrator{
	func(tx migration.LimitedTx) error {
		_, err := tx.Exec(`
			CREATE TABLE project (
				name varchar (255) NOT NULL,
				userno smallint NOT NULL,
				display varchar (255) NOT NULL,
				timeline timestamp without time zone,
				PRIMARY KEY (name, userno)
			);
			CREATE TABLE collaborator (
				project_name varchar (255) NOT NULL,
				project_owner smallint NOT NULL,
				userno smallint NOT NULL,
				PRIMARY KEY (project_name, project_owner, userno),
				FOREIGN KEY (project_name, project_owner)
					REFERENCES project (name, userno)
					ON DELETE RESTRICT
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
		log.Fatalf("Could not connect to PostgreSQL (%s@s/%s): %s",
			conf.User, conf.Host, conf.Database, err)
	}
	return &lcmDB{pgsqlDB, conf}
}

const (
	sessionName   = "void" // we only use one
	sessionUserId = "userid"
)

func initSecureCookie(db *lcmDB, conf configSecurity) {
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
	store, err = sqlsess.Open(db.DB)
	if err != nil {
		log.Fatalf("Could not open session storage: %s", err)
	}
	store.SetKeys(hashKey, blockKey)
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
