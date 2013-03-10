package main

import (
	"fmt"
	"go/build"
	thtml "html/template"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/gorilla/mux"

	"github.com/BurntSushi/toml"
)

var (
	pkg   = path.Join("github.com", "BurntSushi", "lcmweb")
	views *thtml.Template
	cwd   string
	conf  config
	db    *lcmDB
)

var e = fmt.Errorf

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

	db = connect(conf.MySQL)

	tables, err := db.Query("SHOW DATABASES")
	if err != nil {
		log.Fatal(err)
	}
	for tables.Next() {
		var dbName string
		if err := tables.Scan(&dbName); err != nil {
			log.Fatal(err)
		}
		log.Println(dbName)
	}
	if err := tables.Err(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", newHandler((*controller).index))
	r.PathPrefix("/static").HandlerFunc(newHandler((*controller).static))

	r.HandleFunc("/{location}", newHandler((*controller).notFound))

	http.Handle("/", r)
	if err := http.ListenAndServe(":8082", nil); err != nil {
		log.Fatalf("ListenAndServe: %s", err)
	}
}

func readable(fpath string) bool {
	_, err := os.Stat(fpath)
	return err == nil || !os.IsNotExist(err)
}
