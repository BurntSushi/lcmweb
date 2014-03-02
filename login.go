package main

import (
	"net/http"
)

type formLogin struct {
	Email    string
	Password string
	BackTo   string
}

func (c *controller) logout() {
	if !c.user.valid() {
		panic(e("No user logged in."))
	}
	store.Delete(c.session)
	http.Redirect(c.w, c.req, "/", 302)
}

func (c *controller) authenticate(msg error) {
	var form formLogin
	c.decode(&form)

	if len(form.BackTo) == 0 {
		form.BackTo = c.req.URL.String()
	}

	c.render("login", m{
		"Message":      formatMessage(msg.Error()),
		"PrefillEmail": form.Email,
		"BackTo":       form.BackTo,
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
	newPassUrl := mkUrl("newpassword", "userid", user.Id)
	_, err := uauth.Get(user.Id)
	if err != nil {
		panic(ae("Account has no password. Please [set a new password]"+
			"(%s).", newPassUrl))
	}

	ok, err := uauth.Authenticate(user.Id, form.Password)
	if err != nil {
		panic(ae("Error checking password: %s", err))
	} else if !ok {
		panic(ae("Invalid password."))
	}

	c.session.Values[sessionUserId] = user.Id
	assert(c.session.Save(c.req, c.w))
	http.Redirect(c.w, c.req, form.BackTo, 302)
}

func findUserByEmail(email string) *lcmUser {
	for _, user := range conf.Users {
		if email == user.Email {
			return newLcmUser(user)
		}
	}
	return nil
}
