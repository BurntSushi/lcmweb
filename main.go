package main

import (
	"go/build"
	thtml "html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"

	"github.com/BurntSushi/toml"
)

var (
	pkg       = path.Join("github.com", "BurntSushi", "lcmweb")
	views     *thtml.Template
	cwd       string
	conf      config
	db        *lcmDB
	store     *dbStore
	schemaDec *schema.Decoder
	router    *mux.Router
)

func init() {
	var err error

	for _, dir := range build.Default.SrcDirs() {
		if readable(path.Join(dir, pkg)) {
			cwd = path.Join(dir, pkg)
			break
		}
	}

	views = thtml.Must(thtml.ParseGlob(path.Join(cwd, "views", "*.html")))

	confFile := path.Join(cwd, "config.toml")
	if _, err = toml.DecodeFile(confFile, &conf); err != nil {
		log.Fatalf("Error loading config.toml: %s", err)
	}

	initSecureCookie(conf.Security)
	db = connect(conf.PgSQL)
	store = newDBStore(db.DB)
	schemaDec = schema.NewDecoder()
}

func main() {
	router = mux.NewRouter()

	r := router
	r.HandleFunc("/",
		htmlHandler(auth((*controller).index)))
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

	r.HandleFunc("/test",
		htmlHandler(auth((*controller).testing))).Name("test")

	// Catch-alls for pages that don't match a route.
	r.HandleFunc("/{a}",
		htmlHandler(auth((*controller).notFound)))
	r.HandleFunc("/{a}/",
		htmlHandler(auth((*controller).notFound)))
	r.HandleFunc("/{a}/{b}",
		htmlHandler(auth((*controller).notFound)))

	http.Handle("/", r)
	if err := http.ListenAndServe(":8082", nil); err != nil {
		log.Fatalf("ListenAndServe: %s", err)
	}
}

func readable(fpath string) bool {
	_, err := os.Stat(fpath)
	return err == nil || !os.IsNotExist(err)
}

func (c *controller) mkUrl(name string, pairs ...string) *url.URL {
	u, err := router.Get(name).URL(pairs...)
	assert(err)
	return u
}

func (c *controller) mkHttpUrl(name string, pairs ...string) *url.URL {
	u := c.mkUrl(name, pairs...)
	u.Host = c.req.Host
	u.Scheme = "http"
	return u
}
