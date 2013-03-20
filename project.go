package main

func (c *controller) projects() {
	data := m{
		"js":    []string{"project"},
		"Title": "Projects",
		"Nav": c.mkNav(
			nav{"Projects", "/projects"},
		),
	}
	c.render("projects", data)
}

type formProject struct {
	Name string
}

func (c *controller) addProject() {
	var form formProject
	c.decode(&form)

	if len(form.Name) < 1 {
		panic(jsonf("Projects names must be at least one character."))
	}

	panic(jsonf("No dice, mang."))
}
