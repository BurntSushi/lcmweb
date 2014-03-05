package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/csql"
	"github.com/BurntSushi/locker"
)

var (
	reProjectName = regexp.MustCompile("^[-a-zA-Z0-9_ ]+$")
)

func projects(w *web) {
	data := m{
		"js":    []string{"project"},
		"Title": "Projects",
		"Nav": w.mkNav(
			nav{"Projects", w.routes.URLFor("project-list")},
		),
		"MyProjects": w.user.projects(),
	}
	w.html("projects", data)
}

func bitMyProjects(w *web) {
	w.html("bit-myprojects", m{
		"MyProjects": w.user.projects(),
	})
}

func deleteProject(w *web) {
	proj := getProject(w.user, w.user.Id, w.params["project"])
	if w.user.Id != proj.Owner.Id {
		panic(ue("Only owners of projects can delete them."))
	}
	show := func(msg string) {
		w.html("projects_delete", m{
			"Nav": w.mkNav(
				nav{"Projects", w.routes.URLFor("project-list")},
			),
			"Project": proj,
			"Message": msg,
		})
	}
	if w.r.Method == "GET" {
		show("")
	} else if w.r.Method == "POST" {
		var form struct {
			DisplayName string
		}
		w.decode(&form)
		if form.DisplayName != proj.Display {
			show(fmt.Sprintf("The name %s does not match the project name.",
				form.DisplayName))
			return
		}
		proj.delete()
		http.Redirect(w.w, w.r, w.routes.URLFor("project-list"), 302)
	} else {
		panic(ef("Unrecognized request method: %s", w.r.Method))
	}
}

func addProject(w *web) {
	var form struct {
		DisplayName string
	}
	w.decode(&form)

	proj := project{
		Owner:   w.user,
		Name:    projDisplayToName(form.DisplayName),
		Display: form.DisplayName,
		Added:   time.Now().UTC(),
	}
	proj.validate()
	proj.add()

	w.json(htmlEscape(proj.Display))
}

func manageCollaborators(w *web) {
	var form struct {
		ProjectName   string
		Collaborators []string
	}
	w.decode(&form)
	proj := getProject(w.user, w.user.Id, form.ProjectName)

	// We need to do a delete followed by an insert, which means we need
	// exclusion for this project to prevent race conditions.
	lockKey := fmt.Sprintf("%s-%s", proj.Name, proj.Owner.Id)
	locker.Lock(lockKey)
	defer locker.Unlock(lockKey)

	csql.Panic(csql.Tx(db, func(tx *sql.Tx) {
		csql.Exec(tx, `
			DELETE FROM
				collaborator
			WHERE
				project_name = $1 AND project_owner = $2
		`, proj.Name, proj.Owner.Id)
		for _, collaborator := range form.Collaborators {
			u := findUserById(collaborator)
			csql.Exec(tx, `
				INSERT INTO collaborator
					(project_name, project_owner, userid)
				VALUES
					($1, $2, $3)
			`, proj.Name, proj.Owner.Id, u.Id)
		}
	}))

	w.json(w.r.PostForm)
}

func bitCollaborators(w *web) {
	proj := getProject(w.user, w.params["user"], w.params["project"])
	w.html("bit-collaborators", m{"Collabs": proj.Collaborators()})
}

type project struct {
	Owner         *lcmUser
	Name          string
	Display       string
	Added         time.Time
	collaborators []*lcmUser
}

// getProject finds a project given its primary key in the context of a
// request. Namely, it makes sure that the viewer has access to the project.
func getProject(user *lcmUser, projOwner, projName string) *project {
	owner := findUserById(projOwner)
	proj := &project{Owner: owner}

	assert(db.QueryRow(`
		SELECT
			name, display, timeline
		FROM
			project
		WHERE
			name = $1 AND userid = $2
	`, projName, owner.Id).
		Scan(&proj.Name, &proj.Display, &proj.Added))

	// If the owner of the project is the current user, then permission
	// is self evident.
	if user.Id == owner.Id {
		return proj
	}

	// Now the only way the viewing user has permission is if the user is
	// a collaborator on the project.
	for _, collab := range proj.Collaborators() {
		if user.Id == collab.Id {
			return proj
		}
	}

	// No permission!
	panic(ue("User '%s' does not have permission to see this project.",
		user.Id))
}

// projDisplayToName converts a project display name (seen by the user) to a
// project name used for identification purposes. The conversion is to simply
// replace space characters with underscore characters.
func projDisplayToName(display string) string {
	return strings.Replace(display, " ", "_", -1)
}

// projNameToDisplay converts a project name to a display name which is seen
// by the user. The conversion is to simply replace underscore characters with
// space characters.
func projNameToDisplay(name string) string {
	return strings.Replace(name, "_", " ", -1)
}

// add will add the project to the database.
func (proj *project) add() {
	csql.Exec(db, `
		INSERT INTO project 
			(name, userid, display, timeline)
		VALUES
			($1, $2, $3, $4)
	`, proj.Name, proj.Owner.Id, proj.Display, proj.Added)
}

// delete will delete the project from the database. This includes all
// attached collaborators and documents.
func (proj *project) delete() {
	csql.Exec(db, `
		DELETE FROM project
		WHERE name = $1 AND userid = $2
	`, proj.Name, proj.Owner.Id)
}

// validate will check to make sure a new project is valid and can be inserted
// into the DB.
func (proj *project) validate() {
	if len(proj.Name) < 1 {
		panic(ue("Projects names must be at least one character."))
	}
	if len(proj.Name) >= 100 {
		panic(ue("Project names must be fewer than 100 characters."))
	}
	if !reProjectName.MatchString(proj.Name) {
		panic(ue("Project names can only contain letters, numbers, " +
			"spaces, underscores and dashes."))
	}
	if proj.isDuplicate() {
		panic(ue("A project named **%s** already exists.", proj.Display))
	}
}

// isDuplicate returns true if the project already exists.
func (proj *project) isDuplicate() bool {
	n := csql.Count(db, `
		SELECT
			COUNT(*)
		FROM
			project
		WHERE
			name = $1 AND userid = $2
		`, proj.Name, proj.Owner.Id)
	return n > 0
}

func (proj *project) NumDocuments() int {
	return 0
}

func (proj *project) IsCollaborator(user *lcmUser) bool {
	for _, collaborator := range proj.Collaborators() {
		if user.Id == collaborator.Id {
			return true
		}
	}
	return false
}

func (proj *project) Collaborators() []*lcmUser {
	if proj.collaborators != nil {
		return proj.collaborators
	}

	proj.collaborators = make([]*lcmUser, 0)

	rows := csql.Query(db, `
		SELECT
			userid
		FROM
			collaborator
		WHERE
			project_name = $1 AND project_owner = $2
	`, proj.Name, proj.Owner.Id)
	csql.Panic(csql.ForRow(rows, func(s csql.RowScanner) {
		var userid string
		csql.Scan(rows, &userid)

		if user := findUserByNo(userid); user == nil {
			log.Printf("Could not find collaborator %s for "+
				"project '%s' owned by '%s'.", userid, proj.Name, proj.Owner.Id)
		} else {
			proj.collaborators = append(proj.collaborators, user)
		}
	}))
	sort.Sort(usersAlphabetical(proj.collaborators))
	return proj.collaborators
}

func (user *lcmUser) projects() []*project {
	projs := make([]*project, 0)
	rows := csql.Query(db, `
		SELECT
			name, display, timeline
		FROM
			project
		WHERE
			userid = $1
		ORDER BY
			display ASC
	`, user.Id)
	csql.Panic(csql.ForRow(rows, func(s csql.RowScanner) {
		proj := &project{Owner: user}
		csql.Scan(rows, &proj.Name, &proj.Display, &proj.Added)
		projs = append(projs, proj)
	}))
	return projs
}
