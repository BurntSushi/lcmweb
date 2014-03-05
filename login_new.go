package main

import (
	"crypto/rand"
	"math/big"
)

const sessionSecurityKey = "newpass-key"

type formNewPass struct {
	UserId          string
	Key             string
	Password        string
	PasswordConfirm string
}

func newPassword(w *web) {
	user := findResettableUser(w.params["userid"])

	// If there's no security key set, then let's add one.
	skey := sessGet(w.s, sessionSecurityKey)
	if len(skey) == 0 {
		// Don't wait for the email to finish sending.
		setSecurityKey(w, user, true)
	}

	w.ren.HTML(200, "new-password", m{
		"js":         []string{"new-password"},
		"NoAuthUser": user,
	})
}

func newPasswordSave(w *web) {
	var form formNewPass
	w.decode(&form)
	user := findResettableUser(form.UserId)

	// If there's no security key, then something has gone wrong.
	skey := sessGet(w.s, sessionSecurityKey)
	if len(skey) == 0 {
		panic(ue("Could not determine your security code. " +
			"Try re-sending the email."))
	}

	// Nothing matching is a user problem.
	if form.Key != skey {
		panic(ue("The security key entered does not match the " +
			"key sent in the email. Please try entering it again."))
	}

	if len(form.Password) < 8 {
		panic(ue("Passwords must contain at least 8 characters."))
	}
	if form.Password != form.PasswordConfirm {
		panic(ue("Passwords do not match."))
	}

	// Okay, user's data has been validated. Save the new password for the
	// user.
	if err := uauth.Set(user.Id, form.Password); err != nil {
		panic(ue("Error setting password: %s", err))
	}

	// Clear the security key for good measure.
	delete(w.s.Values, sessionSecurityKey)
	assert(w.s.Save(w.r, w.w))
	w.json(nil)
}

// sendEmail always sets a new cookie (overwriting an existing one)
// and sends an email to the user containing the security code.
func newPasswordSend(w *web) {
	var form formNewPass
	w.decode(&form)
	user := findResettableUser(form.UserId)

	// Wait to complete, since this is an asynchronous request already.
	setSecurityKey(w, user, false)
	w.json(nil)
}

// setSecurityKey sets an encrypted cookie with the security key and sends
// an email.
//
// Sending an email can take a while, so if this needs to be done in a
// synchronous request, pass true to `async` and the email will be done in
// its own goroutine.
func setSecurityKey(w *web, user *lcmUser, async bool) {
	newKey := genKey()
	w.s.Values[sessionSecurityKey] = newKey
	assert(w.s.Save(w.r, w.w))

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

func findResettableUser(userid string) *lcmUser {
	user := findUserById(userid)
	if hash, err := uauth.Get(user.Id); err == nil && len(hash) > 0 {
		panic(ue("User **%s** already has a password. "+
			"Please contact the administrator if you'd like to reset your "+
			"password.", user.Id))
	}
	return user
}
