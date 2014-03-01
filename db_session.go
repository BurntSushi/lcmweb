package main

import (
	"database/sql"
	"encoding/base64"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

const (
	sessionIdCookieName = "lcmweb_sessionid"
	sessionName         = "void" // we only use one
	sessionLastUpdated  = "__lastupdated"
	sessionUserId       = "__userid"
)

var scookie *securecookie.SecureCookie

func initSecureCookie(conf configSecurity) {
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
	scookie = securecookie.New(hashKey, blockKey)
}

type dbStore struct {
	*sql.DB
}

func newDBStore(db *sql.DB) *dbStore {
	store := &dbStore{db}
	return store
}

// InitClient must be called on a client after they have been authorized.
// It will set the appropriate cookies needed to track the user's session.
// It will also initialize the session in the database.
func (s *dbStore) InitSession(
	r *http.Request, w http.ResponseWriter, userId string) (err error) {

	sessionId := string(securecookie.GenerateRandomKey(64))
	writeCookie(r, w, sessionIdCookieName, sessionId)

	// Now we must create a session with at least one key in the database.
	// The idea here is that a session ID in a cookie is only valid if it
	// already exists in the database, so that New will fail if the session
	// doesn't already exist.
	_, err = s.Exec(`
		INSERT INTO session
			(sessionid, session_name, key, value)
		VALUES
			($1, $2, $3, $4),
			($1, $2, $5, $6)
	`, sessionId, sessionName,
		sessionLastUpdated, time.Now().UTC(),
		sessionUserId, userId)
	return err
}

func (s *dbStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	// Haven't figured out caching yet.
	// Don't really need it for this app, everything goes into the controller.
	return s.New(r, name)
}

func (s *dbStore) New(r *http.Request, name string) (*sessions.Session, error) {
	sessid, ok := s.getValidSession(r)
	if !ok {
		return nil, authError{}
	}

	sess := sessions.NewSession(s, name)
	sess.Values = make(map[interface{}]interface{})

	locker.rlock(sessid)
	defer locker.runlock(sessid)

	rows, err := s.Query(`
		SELECT
			key, value
		FROM
			session
		WHERE
			sessionid = $1 AND session_name = $2
	`, sessid, sessionName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		sess.Values[key] = value
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return sess, nil
}

func (s *dbStore) Save(r *http.Request, w http.ResponseWriter,
	sess *sessions.Session) (err error) {

	sessid, ok := s.getValidSession(r)
	if !ok {
		return authError{}
	}

	locker.lock(sessid)
	defer locker.unlock(sessid)

	sess.Values[sessionLastUpdated] = time.Now().UTC()

	tx, err := s.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		DELETE FROM session WHERE sessionid = $1
	`, sessid)
	if err != nil {
		assert(tx.Rollback())
		return err
	}

	prep, err := tx.Prepare(`
		INSERT INTO session
			(sessionid, session_name, key, value)
		VALUES
			($1, $2, $3, $4)
	`)
	defer prep.Close()

	if err != nil {
		assert(tx.Rollback())
		return err
	}
	for k, v := range sess.Values {
		_, err = prep.Exec(sessid, sess.Name(), k, v)
		if err != nil {
			assert(tx.Rollback())
			return err
		}
	}

	return tx.Commit()
}

// getValidSession returns the session and user ids of the current HTTP
// request and validates them against the database.
// It returns true if the user's session is valid by checking
// that the session data in the database matches the session data in the
// user's cookie. Returns false if there is any mismatch.
func (s *dbStore) getValidSession(r *http.Request) (string, bool) {
	sessid := s.sessionId(r)
	if len(sessid) == 0 {
		return "", false
	}

	// Now make sure that at least the void session exists.
	var count int
	row := db.QueryRow(`
		SELECT
			COUNT(*) AS count
		FROM
			session
		WHERE
			sessionid = $1 AND session_name = $2
	`, sessid, sessionName)
	if err := row.Scan(&count); err != nil {
		log.Printf("[getValidSession]: %s", err)
		return "", false
	}
	return sessid, count >= 1
}

// sessionId gets the value of the session ID cookie.
func (s *dbStore) sessionId(r *http.Request) string {
	return readCookie(r, sessionIdCookieName)
}

// wipe will delete all information in the DB attached to the session ID.
// We don't bother with deleting the cookie, since this automatically
// invalidates the user session.
func (s *dbStore) wipe(sessid string) {
	mustExec(db, `
		DELETE FROM
			session
		WHERE
			sessionid = $1
	`, sessid)
}

// deleteStale removes rows from the session table that haven't been updated
// in the duration specified.
func (s *dbStore) deleteStale(timeout time.Duration) {
	// A session itself has no concept of when it was last updated. Instead,
	// we maintain this information as a special session key, which is stored
	// as a normal string.
	//
	// In the future, we may want to make "last updated" part of a session to
	// make this operation faster.

	locker.lock("deleteStale")
	defer locker.unlock("deleteStale")

	rows, err := db.Query(`
		SELECT
			sessionid, value
		FROM
			session
		WHERE
			key = $1
	`, sessionLastUpdated)
	if err != nil {
		log.Printf("[deleteStale.query] %s", err)
		return
	}
	defer rows.Close()

	cutoff := time.Now().UTC().Add(-timeout)
	for rows.Next() {
		var sessid, last string
		if err := rows.Scan(&sessid, &last); err != nil {
			log.Printf("[deleteStale.scan] %s", err)
			return
		}

		lastUpdated, err := time.Parse(time.RFC3339Nano, last)
		if err != nil {
			log.Printf("[deleteStale.time] time '%s': %s", last, err)
			return
		}

		if lastUpdated.Before(cutoff) {
			log.Printf("Deleting session '%s' which is before '%s'. Now: '%s'.",
				lastUpdated, cutoff, time.Now().UTC())

			_, err = db.Exec(`
				DELETE FROM
					session
				WHERE
					sessionid = $1
			`, sessid)
			if err != nil {
				log.Printf("[deleteStale.delete] %s", err)
				return
			}
		}
	}
	if err := rows.Err(); err != nil {
		log.Printf("[deleteStale.end] %s", err)
		return
	}
}

// session wraps gorilla.sessions.Session and makes accessing values in the
// session more convenient.
type session struct {
	*sessions.Session
}

func (sess *session) Get(key string) string {
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

// Returns an empty string if the cookie doesn't exist or if there was
// a problem decoding it.
func readCookie(r *http.Request, cname string) string {
	if cook, err := r.Cookie(cname); err == nil {
		var v string
		if err = scookie.Decode(cname, cook.Value, &v); err == nil {
			return v
		} else {
			log.Printf("[readCookie]: %s", err)
		}
	}
	return ""
}

// Writes the value to the named cookie. Logs errors but doesn't report
// them to the user.
func writeCookie(
	r *http.Request, w http.ResponseWriter, cname, cvalue string) {

	if encoded, err := scookie.Encode(cname, cvalue); err == nil {
		cook := &http.Cookie{
			Name:     cname,
			Value:    encoded,
			Path:     "/",
			HttpOnly: true,
		}
		http.SetCookie(w, cook)
	} else {
		log.Printf("[writeCookie]: %s", err)
	}
}
