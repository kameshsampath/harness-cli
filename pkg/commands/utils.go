package commands

import (
	"strings"
	"unicode"
)

func idFromName(s string) string {
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
