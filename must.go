package main

import (
	"database/sql"
)

type execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

func mustExec(exe execer, query string, args ...interface{}) sql.Result {
	r, err := exe.Exec(query, args...)
	assert(err)
	return r
}

type queryer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

func mustQuery(qer queryer, query string, args ...interface{}) *sql.Rows {
	rows, err := qer.Query(query, args...)
	assert(err)
	return rows
}

func mustScan(scanner sql.Scanner, value interface{}) {
	err := scanner.Scan(value)
	assert(err)
}

func assert(err error) {
	if err != nil {
		panic(err)
	}
}
