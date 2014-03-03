package main

import (
	"fmt"
	html "html/template"
	"strings"

	"github.com/russross/blackfriday"
)

var ef = fmt.Errorf

func assert(err error) {
	if err != nil {
		panic(err)
	}
}

type jsonError struct {
	error
}

func wrapErrorsJson() {
	if r := recover(); r != nil {
		if err, ok := r.(error); ok {
			if _, ok := err.(jsonError); ok {
				panic(err)
			} else {
				panic(jsonError{err})
			}
		}
	}
}

type userError struct {
	error
}

func ue(format string, v ...interface{}) userError {
	return userError{ef(format, v...)}
}

func (ue userError) Error() string {
	return ue.error.Error()
}

type systemError struct {
	error
}

func se(format string, v ...interface{}) systemError {
	return systemError{ef(format, v...)}
}

func (se systemError) Error() string {
	return se.error.Error()
}

type authError struct {
	error
}

func ae(format string, v ...interface{}) authError {
	return authError{ef(format, v...)}
}

func (ae authError) Error() string {
	return ae.error.Error()
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
