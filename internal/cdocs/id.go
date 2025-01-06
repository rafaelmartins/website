package cdocs

import (
	"regexp"
	"strings"
)

var (
	reIdFilter = regexp.MustCompile(`[^a-zA-Z0-9_]+`)
)

func id(name string) string {
	rv := reIdFilter.ReplaceAllString(name, " ")
	rv = strings.TrimSpace(rv)
	rv = strings.ReplaceAll(rv, " ", "-")
	return strings.ToLower(rv)
}
