package main

import "io/ioutil"

func uploadDocument(w *web) {
	var form struct {
		UploadDoc string
	}
	w.multiDecode(&form)
	file, header, err := w.r.FormFile("UploadDoc")
	if err != nil {
		panic(ue("There was a problem uploading your document: %s", err))
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		panic(ue("There was a problem reading your uploaded document: %s", err))
	}
	w.lg.Printf("File size: %d", len(data))
	w.lg.Printf("%#v", header)
	w.json(m{"document": string(data)})
}
