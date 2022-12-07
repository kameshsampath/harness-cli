package utils

import (
	"strings"
	"unicode"
)

// IDFromName sanitizes the name and returns the identifier for the resources
func IDFromName(s string) string {
	var name string
	//max length 64 chars
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
	if r := rune(s[0]); unicode.IsDigit(r) {
		name = s[1:]
	}
	name = strings.TrimPrefix(name, "$")
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, " ", "")

	return strings.ToLower(name)
}

// TagMapFromStringArray splits the key:value and returns a map with key and value
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
