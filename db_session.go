package main

import (
	"database/sql"

	"github.com/gorilla/securecookie"
	// "github.com/gorilla/sessions"
)

var (
	hashKey  = securecookie.GenerateRandomKey(64)
	blockKey = securecookie.GenerateRandomKey(64)
)

type authError struct {
	msg string
}

func (ae authError) Error() string {
	return ae.msg
}

type dbStore struct {
	*sql.DB
}

func newDBStore(db *sql.DB) *dbStore {
	store := &dbStore{db}
	return store
}
