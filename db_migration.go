package main

import (
	"database/sql"
)

var migrations = []func(*sql.Tx) error{
	migrate0to1,
}

func updateVersion(tx *sql.Tx, newv int) error {
	_, err := tx.Exec("UPDATE meta SET value = ? WHERE name = 'version'", newv)
	return err
}

func migrate0to1(tx *sql.Tx) error {
	var err error

	_, err = tx.Exec(`
		CREATE TABLE meta (
			name VARCHAR (255) COLLATE utf8_bin NOT NULL PRIMARY KEY,
			value TEXT COLLATE utf8_general_ci
		) Engine = InnoDB
	`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO meta (name, value) VALUES ('version', '1', 2)
	`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO meta (name, value) VALUES ('version', '500')
	`)
	if err != nil {
		return err
	}

	return updateVersion(tx, 1)
}
