package main

import (
	"strings"
)

var templateHelpers = map[string]interface{}{
	"join":  thJoin,
	"split": thSplit,
}

func thJoin(sep string, items []string) string {
	return strings.Join(items, sep)
}

func thSplit(sep string, s string) []string {
	return strings.Split(s, sep)
}
