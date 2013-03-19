package main

import (
	html "html/template"
	"strings"
)

var templateHelpers = map[string]interface{}{
	"join":  thJoin,
	"html":  thHTML,
	"split": thSplit,
}

func thJoin(sep string, items []string) string {
	return strings.Join(items, sep)
}

func thHTML(s string) html.HTML {
	return html.HTML(s)
}

func thSplit(sep string, s string) []string {
	return strings.Split(s, sep)
}
