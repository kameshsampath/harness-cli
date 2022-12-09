package utils

import (
	"strings"
	"unicode"
)

// IDFromName sanitizes the name and returns the identifier for the resources
func IDFromName(s string) string {
	var name string
	// ensure the total length is not more than 64 chars
	if len(s) > 64 {
		for _, w := range s {
			if len(name) > 64 {
				break
			}
			name += string(w)
		}
	} else {
		name = s
	}
	// remove leading digit
	if r := rune(s[0]); unicode.IsDigit(r) {
		name = s[1:]
	}
	// remove leading $
	name = strings.TrimPrefix(name, "$")
	// remove all dashes
	name = strings.ReplaceAll(name, "-", "")
	// replace all spaces with underscore
	name = strings.ReplaceAll(name, " ", "_")

	return strings.ToLower(name)
}

// TagMapFromStringArray splits the key:value and returns as map[key]=value
func TagMapFromStringArray(tags []string) map[string]string {
	tm := make(map[string]string, len(tags))
	if len(tags) > 0 {
		for _, t := range tags {
			if t != "" {
				s := strings.Split(t, ":")
				tm[s[0]] = s[1]
			}
		}
	}
	return tm
}
