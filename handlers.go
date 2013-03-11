package main

import (
	"log"
	"net/http"
	"path"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

type controller struct {
	w       http.ResponseWriter
	req     *http.Request
	params  map[string]string
	session *sessions.Session
}

type handler func(*controller)

func authHandler(h handler) http.HandlerFunc {
	return newHandler(h, true)
}

func noAuthHandler(h handler) http.HandlerFunc {
	return newHandler(h, false)
}

func newHandler(h handler, sessions bool) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		c := &controller{
			w:      w,
			req:    req,
			params: mux.Vars(req),
		}
		defer func() {
			if r := recover(); r != nil {
				switch e := r.(type) {
				case authError:
					c.authenticate(e)
				case error:
					c.error(e)
				default:
					panic(r)
				}
			}
		}()

		if sessions {
			var err error
			sessid, userid, ok := store.getValidSession(c.req)
			log.Println(len(sessid), userid)
			if !ok {
				if err = store.InitClient(c.req, c.w, 1); err != nil {
					panic(err)
				}
			}

			c.session, err = store.New(c.req, sessionName)
			if err != nil {
				panic(err)
			}
		}
		h(c)
	}
}

type m map[string]interface{}

func (c *controller) static() {
	fileServer := http.FileServer(http.Dir(path.Join(cwd, "static")))
	http.StripPrefix("/static", fileServer).ServeHTTP(c.w, c.req)
}

func (c *controller) index() {
	c.render("index", nil)
}

func (c *controller) notFound() {
	c.render("404", m{"Location": c.params["location"]})
}

func (c *controller) render(name string, data interface{}) {
	if err := views.ExecuteTemplate(c.w, name, data); err != nil {
		c.error(err)
	}
}

func (c *controller) authenticate(msg error) {
	c.render("login", m{"Message": msg.Error()})
}

func (c *controller) error(msg error) {
	log.Printf("ERROR: %s", msg)
	err := views.ExecuteTemplate(c.w, "error", m{"Message": msg.Error()})
	if err != nil {
		// Something is seriously wrong.
		log.Fatal(err)
	}
}
