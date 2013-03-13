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
	userIdCookieName    = "lcmweb_userid"
	sessionName         = "void" // we only use one
	lastUpdated         = "__lastupdated"
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
func (s *dbStore) InitClient(
	r *http.Request, w http.ResponseWriter, userId string) error {

	var err error

	sessionId := string(securecookie.GenerateRandomKey(64))
	s.writeCookie(r, w, sessionIdCookieName, sessionId)
	s.writeCookie(r, w, userIdCookieName, userId)

	// Now we must create a session with at least one key in the database.
	// The idea here is that a session ID in a cookie is only valid if it
	// already exists in the database, so that New will fail if the session
	// doesn't already exist.
	// But first, delete any other session with this user id.
	tx, err := s.Begin()
	if err != nil {
		return err
	}
	defer tx.Commit()

	_, err = tx.Exec(`
		DELETE FROM
			session
		WHERE
			userid = $1
	`, userId)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO session
			(sessionid, userid, session_name, key, value)
		VALUES
			($1, $2, $3, $4, $5)
	`, sessionId, userId, sessionName, lastUpdated, time.Now().UTC())
	return err
}

func (s *dbStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	// Haven't figured out caching yet.
	// Don't really need it for this app, everything goes into the controller.
	return s.New(r, name)
}

func (s *dbStore) New(r *http.Request, name string) (*sessions.Session, error) {
	sessid, userid, ok := s.getValidSession(r)
	if !ok {
		return nil, authError{}
	}

	sess := sessions.NewSession(s, name)
	sess.Values = make(map[interface{}]interface{})

	rows, err := s.Query(`
		SELECT
			key, value
		FROM
			session
		WHERE
			sessionid = $1 AND userid = $2 AND session_name = $3
	`, sessid, userid, sessionName)
	if err != nil {
		return nil, err
	}
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

func (s *dbStore) Save(
	r *http.Request, w http.ResponseWriter, sess *sessions.Session) error {

	var err error

	sessid, userid, ok := s.getValidSession(r)
	if !ok {
		return authError{}
	}

	// Saving requires dumping all of the rows corresponding to this user
	// in the session table and then re-adding the keys. This has two major
	// pitfalls:
	//
	// 1) Dumping all of the rows results in an intermediate state where the
	//    user does not have a valid session, therefore, we lock it in a
	//    transaction. (To avoid race conditions, we don't care so much about
	//    rolling back---just let the session invalidate.)
	// 2) If the updated session has no keys, we must add one to keep a valid
	//    session.
	if _, ok = sess.Values[lastUpdated]; !ok {
		sess.Values[lastUpdated] = time.Now().UTC()
	}

	tx, err := s.Begin()
	if err != nil {
		return err
	}
	defer tx.Commit()

	_, err = tx.Exec(`
		DELETE FROM session WHERE userid = $1
	`, userid)
	if err != nil {
		return err
	}

	prep, err := tx.Prepare(`
		INSERT INTO session
			(sessionid, userid, session_name, key, value)
		VALUES
			($1, $2, $3, $4, $5)
	`)
	if err != nil {
		return err
	}
	for k, v := range sess.Values {
		_, err = prep.Exec(sessid, userid, sess.Name(), k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

// getValidSession returns the session and user ids of the current HTTP
// request and validates them against the database.
// It returns true if the user's session is valid by checking
// that the session data in the database matches the session data in the
// user's cookie. Returns false if there is any mismatch.
func (s *dbStore) getValidSession(r *http.Request) (string, string, bool) {
	sessid := s.sessionId(r)
	userid := s.userId(r)

	// If either are empty, then the user isn't authorized.
	if len(sessid) == 0 || len(userid) == 0 {
		return "", "", false
	}

	// Now make sure that at least the void session exists.
	var count int
	row := db.QueryRow(`
		SELECT
			COUNT(*) AS count
		FROM
			session
		WHERE
			sessionid = $1 AND userid = $2 AND session_name = $3
	`, sessid, userid, sessionName)
	if err := row.Scan(&count); err != nil {
		log.Printf("[hasValidSession]: %s", err)
		return "", "", false
	}
	return sessid, userid, count >= 1
}

// sessionId gets the value of the session ID cookie.
func (s *dbStore) sessionId(r *http.Request) string {
	return s.readCookie(r, sessionIdCookieName)
}

// userId gets the value of the user ID cookie.
func (s *dbStore) userId(r *http.Request) string {
	return s.readCookie(r, userIdCookieName)
}

// Returns an empty string if the cookie doesn't exist or if there was
// a problem decoding it.
func (s *dbStore) readCookie(r *http.Request, cname string) string {
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
func (s *dbStore) writeCookie(
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
