package main

import (
	"net/http"

	"github.com/codegangsta/martini"
	"github.com/martini-contrib/render"
)

type formLogin struct {
	Email    string
	Password string
	BackTo   string
}

func logout(w *web) {
	if !w.user.valid() {
		panic(ue("No user logged in."))
	}
	store.Delete(w.s)
	http.Redirect(w.w, w.r, "/", 302)
}

// We can't use *web here since this is called from within the middleware.
// (The *web type isn't registered until the route is executed.)
func authenticate(
	msg error,
	dec formDecoder,
	ren render.Render,
	req *http.Request,
) {
	var form formLogin
	dec(&form)

	if len(form.BackTo) == 0 {
		form.BackTo = req.URL.String()
	}
	ren.HTML(200, "login", m{
		"Message":      formatMessage(msg.Error()),
		"PrefillEmail": form.Email,
		"BackTo":       form.BackTo,
	})
}

func postLogin(w *web, routes martini.Routes) {
	var form formLogin
	w.dec(&form)

	user := findUserByEmail(form.Email)
	if !user.valid() {
		panic(ae("User **%s** does not exist.", form.Email))
	}

	// If the user doesn't have a password in the database, then they need
	// to set a password.
	newPassUrl := routes.URLFor("newpassword", "userid", user.Id)
	_, err := uauth.Get(user.Id)
	if err != nil {
		panic(ae("Account has no password. Please [set a new password]"+
			"(%s).", newPassUrl))
	}

	ok, err := uauth.Authenticate(user.Id, form.Password)
	if err != nil || !ok {
		panic(ae("Invalid password."))
	}

	w.s.Values[sessionUserId] = user.Id
	assert(w.s.Save(w.r, w.w))
	http.Redirect(w.w, w.r, form.BackTo, 302)
}

func findUserByEmail(email string) *lcmUser {
	for _, user := range conf.Users {
		if email == user.Email {
			return newLcmUser(user)
		}
	}
	return nil
}
