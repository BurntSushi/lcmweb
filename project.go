package main

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
	"time"

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
			nav{"Projects", "/projects"},
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

func addProject(w *web) {
	defer wrapErrorsJson()

	var form struct {
		DisplayName string
	}
	w.dec(&form)

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
	w.dec(&form)
	proj := getProject(w.user, w.user.Id, form.ProjectName)

	// We need to do a delete followed by an insert, which means we need
	// exclusion for this project to prevent race conditions.
	lockKey := fmt.Sprintf("%s-%s", proj.Name, proj.Owner.Id)
	locker.Lock(lockKey)
	defer locker.Unlock(lockKey)

	safeTransaction(db, func(tx *sql.Tx) {
		mustExec(tx, `
			DELETE FROM
				collaborator
			WHERE
				project_name = $1 AND project_owner = $2
		`, proj.Name, proj.Owner.No)
		for _, collaborator := range form.Collaborators {
			u := findUserById(collaborator)
			mustExec(tx, `
				INSERT INTO collaborator
					(project_name, project_owner, userno)
				VALUES
					($1, $2, $3)
			`, proj.Name, proj.Owner.No, u.No)
		}
	})

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
			name = $1 AND userno = $2
	`, projName, owner.No).
		Scan(&proj.Name, &proj.Display, &proj.Added))

	// If the owner of the project is the current user, then permission
	// is self evident.
	if user.No == owner.No {
		return proj
	}

	// Now the only way the viewing user has permission is if the user is
	// a collaborator on the project.
	for _, collab := range proj.Collaborators() {
		if user.No == collab.No {
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
	mustExec(db, `
		INSERT INTO project 
			(name, userno, display, timeline)
		VALUES
			($1, $2, $3, $4)
	`, proj.Name, proj.Owner.No, proj.Display, proj.Added)
}

// validate will check to make sure a new project is valid and can be inserted
// into the DB.
func (proj *project) validate() {
	defer wrapErrorsJson()

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
	find := db.QueryRow(`
		SELECT
			COUNT(*)
		FROM
			project
		WHERE
			name = $1 AND userno = $2
	`, proj.Name, proj.Owner.No)

	var count int
	mustScan(find, &count)
	return count > 0
}

func (proj *project) NumDocuments() int {
	return 0
}

func (proj *project) IsCollaborator(user *lcmUser) bool {
	for _, collaborator := range proj.Collaborators() {
		if user.No == collaborator.No {
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

	rows := mustQuery(db, `
		SELECT
			userno
		FROM
			collaborator
		WHERE
			project_name = $1 AND project_owner = $2
	`, proj.Name, proj.Owner.No)
	for rows.Next() {
		var userno int
		mustScan(rows, &userno)

		if user := findUserByNo(userno); user == nil {
			log.Printf("Could not find collaborator with number %d for "+
				"project '%s' owned by '%s'.", userno, proj.Name, proj.Owner.Id)
		} else {
			proj.collaborators = append(proj.collaborators, user)
		}
	}
	assert(rows.Err())

	sort.Sort(usersAlphabetical(proj.collaborators))
	return proj.collaborators
}

func (user *lcmUser) projects() []*project {
	projs := make([]*project, 0)
	rows := mustQuery(db, `
		SELECT
			name, display, timeline
		FROM
			project
		WHERE
			userno = $1
		ORDER BY
			display ASC
	`, user.No)
	for rows.Next() {
		proj := &project{Owner: user}
		mustScan(rows, &proj.Name, &proj.Display, &proj.Added)
		projs = append(projs, proj)
	}
	assert(rows.Err())
	return projs
}
