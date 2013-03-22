package main

import (
	"go/build"
	html "html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

var (
	pkg       = path.Join("github.com", "BurntSushi", "lcmweb")
	views     *html.Template
	cwd       string
	conf      config
	db        *lcmDB
	store     *dbStore
	schemaDec *schema.Decoder
	router    *mux.Router
)

func init() {
	for _, dir := range build.Default.SrcDirs() {
		if readable(path.Join(dir, pkg)) {
			cwd = path.Join(dir, pkg)
			break
		}
	}

	views = html.New("views").Funcs(templateHelpers)
	views = html.Must(views.ParseGlob(path.Join(cwd, "views", "*.html")))

	initConfig()

	initSecureCookie(conf.Security)
	db = connect(conf.PgSQL)
	store = newDBStore(db.DB)
	schemaDec = schema.NewDecoder()
}

func main() {
	// Remove stale sessions periodically.
	go func() {
		ticker := time.Tick(time.Minute)
		for _ = range ticker {
			store.deleteStale(conf.Options.sessionTimeout)
		}
	}()

	router = mux.NewRouter()

	r := router
	r.HandleFunc("/favicon.ico",
		http.NotFound)
	r.PathPrefix("/static").
		HandlerFunc(htmlHandler((*controller).static))

	r.HandleFunc("/login",
		htmlHandler((*controller).postLogin)).Methods("POST")
	r.HandleFunc("/newpassword/{userid}",
		htmlHandler((*controller).newPassword)).
		Name("newpassword")
	r.HandleFunc("/newpassword-save",
		jsonHandler((*controller).newPasswordSave)).Methods("POST")
	r.HandleFunc("/newpassword-send",
		jsonHandler((*controller).newPasswordSend)).Methods("POST")
	r.HandleFunc("/logout",
		htmlHandler(auth((*controller).logout)))

	r.HandleFunc("/",
		htmlHandler(auth((*controller).projects))).Methods("GET")
	r.HandleFunc("/projects",
		htmlHandler(auth((*controller).projects))).Methods("GET").
		Name("projects")
	r.HandleFunc("/bit/myprojects",
		htmlHandler(auth((*controller).bitMyProjects))).Methods("GET").
		Name("bit-myprojects")
	r.HandleFunc("/add-project",
		jsonHandler(auth((*controller).addProject))).Methods("POST").
		Name("add-project")
	r.HandleFunc("/manage-collaborators",
		jsonHandler(auth((*controller).manageCollaborators))).Methods("POST").
		Name("manage-collaborators")
	r.HandleFunc("/bit/{user}/{project}/collaborators",
		htmlHandler(auth((*controller).bitCollaborators))).Methods("GET").
		Name("bit-collaborators")

	r.HandleFunc("/{user}/{project}",
		htmlHandler(auth((*controller).documents))).Methods("GET").
		Name("documents")

	r.HandleFunc("/noop",
		jsonHandler(auth((*controller).noop))).Name("noop")
	r.HandleFunc("/test",
		htmlHandler(auth((*controller).testing))).Name("test")

	// Catch-alls for pages that don't match a route.
	r.HandleFunc("/{a}",
		htmlHandler(auth((*controller).notFound)))
	r.HandleFunc("/{a}/",
		htmlHandler(auth((*controller).notFound)))
	r.HandleFunc("/{a}/{b}",
		htmlHandler(auth((*controller).notFound)))
	r.HandleFunc("/{a}/{b}/",
		htmlHandler(auth((*controller).notFound)))

	srv := &http.Server{
		Addr:        ":8082",
		Handler:     r,
		ReadTimeout: 15 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("ListenAndServe: %s", err)
	}
}

func readable(fpath string) bool {
	_, err := os.Stat(fpath)
	return err == nil || !os.IsNotExist(err)
}

func (c *controller) mkHttpUrl(name string, pairs ...string) *url.URL {
	u := mkUrl(name, pairs...)
	u.Host = c.req.Host
	u.Scheme = "http"
	return u
}
