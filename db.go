package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

const apiVersion = 1

type lcmDB struct {
	*sql.DB
	conf configPgsql
}

func connect(conf configPgsql) *lcmDB {
	conns := fmt.Sprintf(
		"user=%s password=%s host=%s port=%d dbname=%s sslmode=disable",
		conf.User, conf.Password, conf.Host, conf.Port, conf.Database)

	pgsqlDB, err := sql.Open("postgres", conns)
	if err != nil {
		log.Fatalf("Could not connect to PostgreSQL (%s@s/%s): %s",
			conf.User, conf.Host, conf.Database, err)
	}

	db := &lcmDB{pgsqlDB, conf}
	db.migrate()

	log.Printf("Database version: %d (API: %d)", db.version(), apiVersion)
	return db
}
