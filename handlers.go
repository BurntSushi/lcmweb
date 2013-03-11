package main

import (
	"log"
	"net/http"
	"path"

	"github.com/gorilla/mux"
)

type controller struct {
	w      http.ResponseWriter
	req    *http.Request
	params map[string]string
}

type handler func(*controller)

func newHandler(h handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		c := &controller{w, req, mux.Vars(req)}
		defer func() {
			if r := recover(); r != nil {
				switch err := r.(type) {
				case authError:
					c.authenticate(err)
				case error:
					c.error(err)
				default:
					panic(r)
				}
			}
		}()
		h(c)
	}
}

type m map[string]interface{}

func (c *controller) static() {
	fileServer := http.FileServer(http.Dir(path.Join(cwd, "static")))
	http.StripPrefix("/static", fileServer).ServeHTTP(c.w, c.req)
}

func (c *controller) index() {
	panic(authError{})
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
