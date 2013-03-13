package main

import (
	"crypto/rand"
	"log"
	"math/big"
	"time"
)

const cookieKeyName = "lcmweb_new-pass-key"

type formSecurity struct {
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

	var form formSecurity
	c.decode(&form)

	data := m{
		"js":         []string{"new-password"},
		"NoAuthUser": user,
	}

	// Form is only submitted when there's a password.
	if len(form.Password) > 0 {
		log.Println("Who who?")
		data["Message"] = "WATWAT"
	}

	// Make sure security key is valid (again).
	// Check password. Just look at length and whether they are equal.
	// Once everything is good, add a row to the `password` table and
	// redirect to the login page.

	c.render("new-password", data)
}

func (c *controller) newPasswordJson() {
	var form formSecurity
	c.decode(&form)
	_ = findResettableUser(form.UserId)

	// If there's no cookie key, then something has gone wrong.
	cookieKey := store.readCookie(c.req, cookieKeyName)
	if len(cookieKey) == 0 {
		panic(e("Could not determine your security code. Please try again."))
	}

	// Nothing matching is a user problem.
	if form.Key != cookieKey {
		panic(jsonf("The security key entered does not match the " +
			"key sent in the email. Please try entering it again."))
	}

	c.json(nil)
}

// sendEmail always sets a new cookie (overwriting an existing one)
// and sends an email to the user containing the security code.
//
func (c *controller) newPasswordSend() {
	var form formSecurity
	c.decode(&form)
	user := findResettableUser(form.UserId)

	// Wait to complete, since this is an asynchronous request already.
	setSecurityKey(c, user, false)

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
	log.Println(newKey)

	email := func() {
		time.Sleep(2 * time.Second)
		log.Println("Email sent!")
	}
	if async {
		go email()
	} else {
		email()
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
