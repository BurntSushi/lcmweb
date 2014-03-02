package main

import (
	"fmt"
	html "html/template"
	"log"
	"strings"

	"github.com/russross/blackfriday"
)

var e = fmt.Errorf

type authError struct {
	msg string
}

func ae(format string, v ...interface{}) authError {
	return authError{fmt.Sprintf(format, v...)}
}

func (ae authError) Error() string {
	return ae.msg
}

func (c *controller) error(msg error) {
	log.Printf("ERROR: %s", msg)

	data := m{"Message": formatMessage(msg.Error())}
	err := views.ExecuteTemplate(c.w, "error", data)
	if err != nil {
		// Something is seriously wrong.
		log.Fatal(err)
	}
}

func (c *controller) notFound() {
	c.w.WriteHeader(404)
	c.render("404", m{"Location": c.req.URL.String()})
}

func formatMessage(s string) html.HTML {
	fmtd := strings.TrimSpace(toMarkdown(html.HTMLEscapeString(s)))
	if strings.HasPrefix(fmtd, "<p>") {
		fmtd = fmtd[3:]
	}
	if strings.HasSuffix(fmtd, "</p>") {
		fmtd = fmtd[:len(fmtd)-4]
	}
	return html.HTML(fmtd)
}

func toMarkdown(s string) string {
	return string(blackfriday.MarkdownBasic([]byte(s)))
}
