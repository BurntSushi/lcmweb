package main

func documents(w *web) {
	proj := getProject(w.user, w.params["owner"], w.params["project"])
	w.html("documents", m{
		"P": proj,
	})
}

func addDocument(w *web) {
}
