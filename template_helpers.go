package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	html "html/template"
	"log"
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
	"combine":   thCombine,

	"datetime": thDateTime,
	"date":     thDate,
	"time":     thTime,

	// This is filled in when the routes are resolved.
	// Seems like a blemish in Martini.
	"url": func(name string, pairs ...interface{}) string { return "" },
}

func htmlEscape(s string) string {
	return html.HTMLEscapeString(s)
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

// combine provides a way to compose values during template execution.
// This is particularly useful when executing sub-templates. For example,
// say you've defined two variables `$a` and `$b` that you want to pass to
// a sub-template. But templates can only take a single pipeline. Combine will
// let you bind any number of values. For example:
//
//	{{ template "tpl_name" (Combine "a" $a "b" $b) }}
//
// The template "tpl_name" can then access `$a` and `$b` with `.a` and `.b`.
//
// Note that the first and every other subsequent value must be strings. The
// second and every other subsequent value may be anything. There must be an
// even number of arguments given. If any part of this contract is violated,
// the function panics.
func thCombine(keyvals ...interface{}) map[string]interface{} {
	if len(keyvals)%2 != 0 {
		log.Printf("Combine must have even number of parameters but %d isn't.",
			len(keyvals))
		return nil
	}
	m := make(map[string]interface{})
	for i := 0; i < len(keyvals); i += 2 {
		key, ok := keyvals[i].(string)
		if !ok {
			log.Printf("Parameter %d to Combine must be a string but it is "+
				"a %T.", i, keyvals[i])
			return nil
		}
		m[key] = keyvals[i+1]
	}
	return m
}
