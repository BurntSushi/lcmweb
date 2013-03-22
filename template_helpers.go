package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	html "html/template"
	"log"
	"net/url"
	"reflect"
	"strings"
	"time"
)

var templateHelpers = map[string]interface{}{
	"escape":    htmlEscape,
	"join":      thJoin,
	"html":      thHTML,
	"split":     thSplit,
	"commafy":   thCommafy,
	"stringify": thStringify,
	"jsonify":   thJsonify,

	"url": mkUrl,

	"datetime": thDateTime,
	"date":     thDate,
	"time":     thTime,
}

func htmlEscape(s string) string {
	return html.HTMLEscapeString(s)
}

func mkUrl(name string, pairs ...string) *url.URL {
	obj := router.Get(name)
	if obj == nil {
		panic(e("URL page with name '%s' does not exist.", name))
	}
	u, err := obj.URL(pairs...)
	assert(err)
	return u
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

func thDateTime(user *lcmUser, t time.Time) string {
	return t.In(user.timeZone).Format(user.DateFmt + " at " + user.TimeFmt)
}

func thDate(user *lcmUser, t time.Time) string {
	return t.In(user.timeZone).Format(user.DateFmt)
}

func thTime(user *lcmUser, t time.Time) string {
	return t.In(user.timeZone).Format(user.TimeFmt)
}

func thStringify(values interface{}) []string {
	rval := reflect.ValueOf(values)
	switch rval.Kind() {
	case reflect.Slice:
		strs := make([]string, rval.Len())
		for i := 0; i < rval.Len(); i++ {
			strs[i] = rval.Index(i).Interface().(fmt.Stringer).String()
		}
		return strs
	}
	log.Printf("Could not extract list of strings from %T type.", values)
	return []string{"N/A"}
}

func thCommafy(strs []string) string {
	switch len(strs) {
	case 0:
		return ""
	case 1:
		return strs[0]
	case 2:
		return fmt.Sprintf("%s and %s", strs[0], strs[1])
	default:
		return fmt.Sprintf("%s and %s",
			strings.Join(strs[0:len(strs)-1], ", "), strs[len(strs)-1])
	}
}

func thJsonify(v interface{}) html.JS {
	bs, err := json.Marshal(v)
	assert(err)

	buf := new(bytes.Buffer)
	json.HTMLEscape(buf, bs)
	return html.JS(buf.String())
}
