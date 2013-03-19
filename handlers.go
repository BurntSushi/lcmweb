package main

import (
	"bytes"
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
	sess, err := store.New(c.req, sessionName)
	if err != nil {
		panic(err)
	}
	c.session = &session{sess}

	// If we're here, then we've been authenticated.
	c.user = findUser(c.session.Get(sessionUserId))

	// Always update the session "last updated" time.
	assert(c.session.Save(c.req, c.w))
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

func (c *controller) noop() {
	c.json(nil)
}

func (c *controller) testing() {
	data := m{
		"Title": "Testing 1 2 3",
		"Nav": c.mkNav(
			nav{"Wat 1", "/wat1"},
			nav{"Wat 2", "/wat2"},
			nav{"Wat 3", "/wat3"},
		),
	}
	c.render("test", data)
}

func (c *controller) render(name string, data interface{}) {
	bs := c.renderBytes(name, data)
	n, err := c.w.Write(bs)
	assert(err)
	if n != len(bs) {
		panic(e("Expected to write %d bytes but only wrote %d bytes.",
			len(bs), n))
	}
}

func (c *controller) renderString(name string, data interface{}) string {
	return string(c.renderBytes(name, data))
}

func (c *controller) renderBytes(name string, data interface{}) []byte {
	if c.user.valid() {
		if data == nil {
			data = m{"User": c.user}
		} else if m, ok := data.(m); ok {
			m["User"] = c.user
		}
	}

	buf := new(bytes.Buffer)
	assert(views.ExecuteTemplate(buf, name, data))
	return buf.Bytes()
}
