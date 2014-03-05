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
	m.Use(recovery)

	m.Any("/favicon.ico", http.NotFound)

	m.Post("/login", webGuest, postLogin).Name("login")
	m.Get("/newpassword/:userid", webGuest, newPassword).Name("newpassword")
	m.Post("/newpassword-save", jsonResp, webGuest, newPasswordSave)
	m.Post("/newpassword-send", jsonResp, webGuest, newPasswordSend)

	m.Get("/logout", webAuth, logout)
	m.Post("/noop", webGuest, func(w *web) { w.json(nil) }).Name("noop")

	m.Get("/", webAuth, projects)
	m.Get("/project/list", webAuth, projects).Name("project-list")
	m.Get("/project/bit/my", webAuth, bitMyProjects).Name("project-bit-my")
	m.Post("/project/add", jsonResp, webAuth, addProject).Name("project-add")
	m.Get("/project/delete/:project", webAuth, deleteProject).
		Name("project-delete")
	m.Post("/project/delete/:project", webAuth, deleteProject)
	m.Post("/project/collab/manage", jsonResp, webAuth, manageCollaborators).
		Name("project-collab-manage")
	m.Get("/project/collab/list/:user/:project", webAuth, bitCollaborators).
		Name("project-bit-collab")

	m.Get("/:user/:project", webAuth, documents).Name("documents")

	m.Run()
}
