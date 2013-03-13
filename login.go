package main

import (
	"crypto/sha512"
	"database/sql"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/securecookie"
)

type formLogin struct {
	Email    string
	Password string
}

func (c *controller) authenticate(msg error) {
	var form formLogin
	c.decode(&form)

	c.render("login", m{
		"Message":      formatMessage(msg.Error()),
		"PrefillEmail": form.Email,
	})
}

func (c *controller) postLogin() {
	var form formLogin
	c.decode(&form)

	user := findUserByEmail(form.Email)
	if !user.valid() {
		panic(ae("User **%s** does not exist.", form.Email))
	}

	// If the user doesn't have a password in the database, then they need
	// to set a password.
	hi, ok := user.getHashInfo()
	if !ok {
		next := c.mkUrl("newpassword", "userid", user.Id)
		http.Redirect(c.w, c.req, next.String(), 302)
		return
	}

	user.authenticate(hi, form.Password)

	c.render("login", m{"Message": "WAT: " + form.Password})
}

func findUserByEmail(email string) configUser {
	for _, user := range conf.Users {
		if email == user.Email {
			return user
		}
	}
	return configUser{}
}

func (u configUser) getHashInfo() (hashInfo, bool) {
	var password, salt1, salt2 string
	row := db.QueryRow(`
		SELECT
			password, salt1, salt2
		FROM
			password
		WHERE
			userno = $1
	`, u.No)
	if err := row.Scan(&password, &salt1, &salt2); err != nil {
		if err == sql.ErrNoRows {
			return hashInfo{}, false
		}

		// An unexpected error!
		panic(err)
	}
	return hashInfo{password, salt1, salt2}, true
}

func (u configUser) authenticate(hi hashInfo, givenPassword string) {
}

type hashInfo struct {
	password, salt1, salt2 string
}

func newHashInfo(password string) hashInfo {
	passh := sha512.New()
	salth1 := sha512.New()
	salth2 := sha512.New()

	io.WriteString(passh, password)
	salth1.Write(securecookie.GenerateRandomKey(64))
	salth2.Write(securecookie.GenerateRandomKey(64))

	return hashInfo{
		fmt.Sprintf("%x", passh.Sum(nil)),
		fmt.Sprintf("%x", salth1.Sum(nil)),
		fmt.Sprintf("%x", salth2.Sum(nil)),
	}
}

func (hi hashInfo) hashGivenPassword(password string) string {
	h := sha512.New()
	io.WriteString(h, hi.salt1)
	io.WriteString(h, password)
	io.WriteString(h, hi.salt2)
	return fmt.Sprintf("%x", h.Sum(nil))
}
