package main

import (
	"github.com/BurntSushi/csql"

	"regexp"
	"time"
)

var (
	reDocumentName = regexp.MustCompile("^[-a-zA-Z0-9 ]+$")
)

func documentNav(w *web, proj *project, d *document, label string) []string {
	doclist := w.routes.URLFor("document-list", proj.Owner.Id, proj.Name)
	navs := []nav{
		{"Projects", w.routes.URLFor("project-list")},
		{proj.Display, doclist},
	}
	if d != nil {
		docmain := w.routes.URLFor(
			"document", proj.Owner.Id, proj.Name, d.Name, d.Recorded)
		navs = append(navs, nav{d.Name, docmain})
	}
	if len(label) > 0 {
		navs = append(navs, nav{label, ""})
	}
	return w.mkNav(navs...)
}

func documents(w *web) {
	proj := getProject(w.user, w.params["owner"], w.params["project"])
	w.html("document-list", m{
		"Nav": documentNav(w, proj, nil, ""),
		"P":   proj,
	})
}

func addDocument(w *web) {
	proj := getProject(w.user, w.params["owner"], w.params["project"])
	show := func(msg string) {
		w.html("document-add", m{
			"js":      []string{"document-upload"},
			"Nav":     documentNav(w, proj, nil, "Add Document"),
			"Message": msg,
			"P":       proj,
			"Conf":    conf,
		})
	}
	if w.r.Method == "GET" {
		show("")
	} else if w.r.Method == "POST" {
	} else {
		panic(ef("Unrecognized request method: %s", w.r.Method))
	}
}

type document struct {
	Project    *project
	Display    string
	Name       string
	Recorded   time.Time
	Categories []string
	Content    string
	CreatedBy  *lcmUser
	Created    time.Time
	Modified   time.Time
}

func insertDocument(
	creator *lcmUser,
	proj *project,
	display string,
	recorded time.Time,
	categories []string,
	content string,
) (*document, error) {
	d := &document{
		Project:    proj,
		Display:    display,
		Name:       displayToName(display),
		Recorded:   recorded,
		Categories: categories,
		Content:    content,
		CreatedBy:  creator,
		Created:    time.Now().UTC(),
		Modified:   time.Now().UTC(),
	}
	if err := d.validate(); err != nil {
		return nil, err
	}
	csql.Exec(db, `
		INSERT INTO document (
			project_owner, project_name, name, recorded, categories,
			content, created_by, created, modified
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`,
		d.Project.Owner.Id, d.Project.Name, d.Name, d.Recorded,
		d.Categories, d.Content, d.CreatedBy.Id, d.Created, d.Modified)
	return d, nil
}

// validate will check to make sure a document is valid and can be inserted
// into the DB. If there is a problem with the document, an error is returned.
func (d *document) validate() error {
	if len(d.Name) < 1 {
		return ue("Document names must be at least one character.")
	}
	if len(d.Name) >= 100 {
		return ue("Document names must be fewer than 100 characters.")
	}
	if !reDocumentName.MatchString(d.Name) {
		return ue("Document names can only contain letters, numbers, " +
			"spaces and dashes.")
	}
	if d.isDuplicate() {
		return ue("A document named **%s** and recorded on **%s** "+
			"already exists.", d.Display, thDate(d.CreatedBy, d.Recorded))
	}
	return nil
}

func (d *document) isDuplicate() bool {
	n := csql.Count(db, `
		SELECT COUNT(*)
		FROM document
		WHERE project_owner = $1 AND project_name = $2
			AND name = $3 AND recorded = $4
		`, d.Project.Owner.Id, d.Project.Name, d.Name, d.Recorded)
	return n > 0
}
