package main

import (
	"crypto/rand"
	"math/big"
)

const cookieKeyName = "lcmweb_new-pass-key"

type formNewPass struct {
	UserId          string
	Key             string
	Password        string
	PasswordConfirm string
}

func (c *controller) newPassword() {
	user := findResettableUser(c.params["userid"])

	// If there's no cookie key set, then let's add one.
	cookieKey := store.readCookie(c.req, cookieKeyName)
	if len(cookieKey) == 0 {
		// Don't wait for the email to finish sending.
		setSecurityKey(c, user, true)
	}

	c.render("new-password", m{
		"js":         []string{"new-password"},
		"NoAuthUser": user,
	})
}

func (c *controller) newPasswordSave() {
	var form formNewPass
	c.decode(&form)
	user := findResettableUser(form.UserId)

	// If there's no cookie key, then something has gone wrong.
	cookieKey := store.readCookie(c.req, cookieKeyName)
	if len(cookieKey) == 0 {
		panic(e("Could not determine your security code. " +
			"Try re-sending the email."))
	}

	// Nothing matching is a user problem.
	if form.Key != cookieKey {
		panic(jsonf("The security key entered does not match the " +
			"key sent in the email. Please try entering it again."))
	}

	if len(form.Password) < 8 {
		panic(jsonf("Passwords must contain at least 8 characters."))
	}
	if form.Password != form.PasswordConfirm {
		panic(jsonf("Passwords do not match."))
	}

	// Okay, user's data has been validated. Save the new password for the
	// user.
	hi := newHashInfo(form.Password)
	mustExec(db, `
		INSERT INTO password
			(userno, password, salt1, salt2)
		VALUES
			($1, $2, $3, $4)
	`, user.No, hi.password, hi.salt1, hi.salt2)

	c.json(nil)
}

// sendEmail always sets a new cookie (overwriting an existing one)
// and sends an email to the user containing the security code.
//
func (c *controller) newPasswordSend() {
	var form formNewPass
	c.decode(&form)
	user := findResettableUser(form.UserId)

	// Wait to complete, since this is an asynchronous request already.
	setSecurityKey(c, user, false)

	// blank success
	c.json(nil)
}

// setSecurityKey sets an encrypted cookie with the security key and sends
// an email.
//
// Sending an email can take a while, so if this needs to be done in a
// synchronous request, pass true to `async` and the email will be done in
// its own goroutine.
func setSecurityKey(c *controller, user configUser, async bool) {
	newKey := genKey()
	store.writeCookie(c.req, c.w, cookieKeyName, newKey)

	email := func() error {
		return user.email("security key", "Security key: "+newKey)
	}
	if async {
		go email()
	} else {
		assert(email())
	}
}

func genKey() string {
	asciis := make([]byte, 8)
	for i := 0; i < 8; i++ {
		bint, err := rand.Int(rand.Reader, big.NewInt(52))
		assert(err)
		b := byte(bint.Int64())

		if b < 26 {
			asciis[i] = 'A' + b
		} else {
			asciis[i] = 'a' + b - 26
		}
	}
	return string(asciis)
}

func findResettableUser(userid string) configUser {
	user := findUser(userid)
	if _, ok := user.getHashInfo(); ok {
		panic(e("User **%s** already has a password. "+
			"Please contact the administrator if you'd like to reset your "+
			"password.", user.Id))
	}
	return user
}
