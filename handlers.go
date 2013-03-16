package main

import (
	"log"
	"net/http"
	"path"

	"github.com/gorilla/mux"
)

type controller struct {
	w       http.ResponseWriter
	req     *http.Request
	params  map[string]string
	session *session
	user    configUser
}

func newController(w http.ResponseWriter, req *http.Request) *controller {
	return &controller{
		w:      w,
		req:    req,
		params: mux.Vars(req),
	}
}

type handler func(*controller)

func auth(h handler) handler {
	return func(c *controller) {
		c.auth(h)
	}
}

func (c *controller) auth(h handler) {
	c.loadSession()
	h(c)
}

func (c *controller) loadSession() {
	// _, userid, _ := store.getValidSession(c.req)
	// if !ok {
	// if err = store.InitClient(c.req, c.w, "andrew"); err != nil {
	// panic(err)
	// }
	// }

	sess, err := store.New(c.req, sessionName)
	if err != nil {
		panic(err)
	}
	c.session = &session{sess}

	// If we're here, then we've been authenticated.
	c.user = findUser(c.session.Get(sessionUserId))
}

func htmlHandler(h handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		c := newController(w, req)
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
		h(c)
	}
}

type m map[string]interface{}

func (c *controller) decode(v interface{}) {
	assert(c.req.ParseForm())
	assert(schemaDec.Decode(v, c.req.PostForm))
}

func (c *controller) decodeMultipart(formVals interface{}) {
	assert(c.req.ParseMultipartForm(10737418240))
	assert(schemaDec.Decode(formVals, c.req.MultipartForm.Value))
}

func (c *controller) static() {
	fileServer := http.FileServer(http.Dir(path.Join(cwd, "static")))
	http.StripPrefix("/static", fileServer).ServeHTTP(c.w, c.req)
}

func (c *controller) index() {
	c.render("index", nil)
}

func (c *controller) testing() {
	c.render("test", nil)
}

func (c *controller) render(name string, data interface{}) {
	if c.user.valid() {
		if data == nil {
			data = m{"User": c.user}
		} else if m, ok := data.(map[string]interface{}); ok {
			m["User"] = c.user
			log.Println("wat", m["User"])
		}
	}
	if err := views.ExecuteTemplate(c.w, name, data); err != nil {
		c.error(err)
	}
}
