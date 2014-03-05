package main

import (
	"bytes"
	"html/template"
	"log"
	"net/http"

	"github.com/BurntSushi/csql"

	"github.com/gorilla/schema"
	"github.com/gorilla/sessions"

	"github.com/codegangsta/martini"
	"github.com/martini-contrib/render"
)

// web is a convenience type for collecting commonly used state.
type web struct {
	lg     *log.Logger
	c      martini.Context
	routes martini.Routes
	params martini.Params
	r      *http.Request
	w      http.ResponseWriter
	s      *sessions.Session
	ren    render.Render
	decode formDecoder
	user   *lcmUser
}

func webGuest(
	lg *log.Logger,
	c martini.Context,
	routes martini.Routes,
	params martini.Params,
	r *http.Request,
	w http.ResponseWriter,
	s *sessions.Session,
	ren render.Render,
	dec formDecoder,
) {
	state := &web{
		lg: lg, c: c, routes: routes, params: params,
		r: r, w: w, s: s, ren: ren, decode: dec,
	}
	ren.Template().Funcs(template.FuncMap{
		"url": state.url,
	})
	c.Map(state)
}

func webAuth(
	lg *log.Logger,
	c martini.Context,
	routes martini.Routes,
	params martini.Params,
	r *http.Request,
	w http.ResponseWriter,
	s *sessions.Session,
	ren render.Render,
	dec formDecoder,
) {
	userId := sessGet(s, sessionUserId)
	if len(userId) == 0 {
		panic(ae(""))
	}
	state := &web{
		lg: lg, c: c, routes: routes, params: params,
		r: r, w: w, s: s, ren: ren, decode: dec,
		user: findUserById(userId),
	}
	ren.Template().Funcs(template.FuncMap{
		"url": state.url,
	})
	c.Map(state)
}

func (w *web) url(name string, params ...interface{}) string {
	return w.routes.URLFor(name, params...)
}

func (w *web) json(v interface{}) {
	w.ren.JSON(200, m{
		"status":  "success",
		"content": v,
	})
}

func (w *web) html(name string, data m) {
	w.ren.HTML(200, name, w.tplData(data))
}

func (w *web) renderBytes(name string, data m) []byte {
	buf := new(bytes.Buffer)
	assert(w.ren.Template().ExecuteTemplate(buf, name, w.tplData(data)))
	return buf.Bytes()
}

func (w *web) tplData(data m) m {
	if w.user.valid() {
		if data == nil {
			data = m{"User": w.user}
		} else {
			data["User"] = w.user
		}
	}
	w.lg.Printf("%#v", data)
	return data
}

func renderer() martini.Handler {
	return render.Renderer(render.Options{
		Directory:  "views",
		Layout:     "",
		Extensions: []string{".html"},
		Funcs:      []template.FuncMap{templateHelpers},
		IndentJSON: true,
	})
}

func session(store sessions.Store, name string) martini.Handler {
	return func(c martini.Context, r *http.Request, w http.ResponseWriter) {
		sess, err := store.Get(r, name)
		assert(err)
		assert(sess.Save(r, w))
		c.Map(sess)
	}
}

func recovery(
	c martini.Context,
	req *http.Request,
	ren render.Render,
	dec formDecoder,
) {
	defer func() {
		if r := recover(); r != nil {
			switch err := r.(type) {
			case jsonError:
				handleJsonError(err, ren)
			case authError:
				authenticate(err, dec, ren, req)
			case userError:
				ren.HTML(200, "error", m{
					"Message": formatMessage(err.Error()),
				})
			case csql.SQLError:
				ren.HTML(200, "error", m{
					"Message": formatMessage(err.Error()),
				})
			default:
				panic(r)
			}
		}
	}()
	c.Next()
}

func handleJsonError(err jsonError, ren render.Render) {
	switch e := err.error.(type) {
	case authError:
		ren.JSON(200, m{
			"status":  "noauth",
			"message": formatMessage(e.Error()),
		})
	case userError:
		ren.JSON(200, m{
			"status":  "fail",
			"message": formatMessage(e.Error()),
		})
	default:
		ren.JSON(200, m{
			"status":  "error",
			"message": formatMessage(e.Error()),
		})
	}
}

type jsonError struct {
	error
}

func jsonResp(c martini.Context) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				if _, ok := err.(jsonError); ok {
					panic(r)
				} else {
					panic(jsonError{err})
				}
			}
		}
	}()
	c.Next()
}

type formDecoder func(v interface{})
type multiDecoder func(v interface{})

func postDecoder() martini.Handler {
	dec := schema.NewDecoder()
	return func(c martini.Context, r *http.Request) {
		decode := func(v interface{}) {
			assert(r.ParseForm())
			assert(dec.Decode(v, r.PostForm))
		}
		c.Map(formDecoder(decode))
	}
}

func postMultiDecoder() martini.Handler {
	dec := schema.NewDecoder()
	return func(c martini.Context, r *http.Request) {
		decode := func(v interface{}) {
			assert(r.ParseMultipartForm(10737418240))
			assert(dec.Decode(v, r.MultipartForm.Value))
		}
		c.Map(multiDecoder(decode))
	}
}
