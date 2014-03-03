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

type multiScanner interface {
	Scan(dest ...interface{}) error
}

func mustScan(scanner multiScanner, dest ...interface{}) {
	assert(scanner.Scan(dest...))
}

type transactor interface {
	Begin() (*sql.Tx, error)
}

// safeTransaction will execute `run` within a transaction.
// If `run` panics, then the transaction will be rolled back and the
// panic will be allowed to bubble up. Otherwise, the transaction
// is committed.
func safeTransaction(db transactor, run func(*sql.Tx)) {
	tx, err := db.Begin()
	assert(err) // no need to rollback yet

	defer func() {
		if r := recover(); r != nil {
			assert(tx.Rollback())
			panic(r)
		}
	}()

	run(tx)
	assert(tx.Commit())
}
