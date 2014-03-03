package main

import (
	"log"
	"net/http"
	"path"
	"time"

	"github.com/codegangsta/martini"

	"github.com/BurntSushi/sqlauth"
	"github.com/BurntSushi/sqlsess"
)

type m map[string]interface{}

var (
	pkg   = path.Join("github.com", "BurntSushi", "lcmweb")
	cwd   string
	conf  config
	db    *lcmDB
	store *sqlsess.Store
	uauth *sqlauth.Store
)

func main() {
	var err error

	conf = newConfig()
	db = connect(conf.PgSQL)
	store = newStore(db, conf.Security)
	if uauth, err = sqlauth.Open(db.DB); err != nil {
		log.Fatalf("Could not open authenticator: %s", err)
	}

	// Remove stale sessions periodically.
	go func() {
		ticker := time.Tick(time.Minute)
		for _ = range ticker {
			store.Clean(conf.Options.sessionTimeout)
		}
	}()

	m := martini.Classic()
	m.Use(martini.Static("static", martini.StaticOptions{
		Prefix:      "/static",
		SkipLogging: true,
	}))
	m.Use(postDecoder())
	m.Use(postMultiDecoder())
	m.Use(renderer())
	m.Use(session(store, sessionName))
	m.Use(errors())

	m.Any("/favicon.ico", http.NotFound)

	m.Post("/login", webGuest, postLogin)
	m.Get("/newpassword/:userid", webGuest, newPassword).Name("newpassword")
	m.Post("/newpassword-save", webGuest, newPasswordSave)
	m.Post("/newpassword-send", webGuest, newPasswordSend)

	m.Get("/logout", webAuth, logout)
	m.Post("/noop", webAuth, func(w *web) { w.json(nil) }).Name("noop")

	m.Get("/", webAuth, projects)
	m.Get("/projects", webAuth, projects).Name("projects")
	m.Get("/bit/myprojects", webAuth, bitMyProjects).Name("bit-myprojects")
	m.Post("/add-project", webAuth, addProject).Name("add-project")
	m.Post("/manage-collaborators", webAuth, manageCollaborators).
		Name("manage-collaborators")
	m.Get("/bit/:user/:project/collaborators", webAuth, bitCollaborators).
		Name("bit-collaborators")

	m.Get("/:user/:project", webAuth, documents).Name("documents")

	m.Run()
}
