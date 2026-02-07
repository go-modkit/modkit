package cmd

import (
	"strings"
	"unicode"
)

func sanitizePackageName(name string) string {
	lower := strings.ToLower(name)
	b := strings.Builder{}
	for _, r := range lower {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		}
	}

	s := b.String()
	if s == "" {
		return "pkg"
	}
	if s[0] >= '0' && s[0] <= '9' {
		return "pkg" + s
	}
	return s
}

func exportedIdentifier(name string) string {
	parts := strings.FieldsFunc(name, func(r rune) bool {
		return r == '-' || r == '_' || unicode.IsSpace(r)
	})

	if len(parts) == 0 {
		return "Generated"
	}

	b := strings.Builder{}
	for _, part := range parts {
		if part == "" {
			continue
		}
		runes := []rune(strings.ToLower(part))
		runes[0] = unicode.ToUpper(runes[0])
		for _, r := range runes {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				b.WriteRune(r)
			}
		}
	}

	s := b.String()
	if s == "" {
		return "Generated"
	}
	if s[0] >= '0' && s[0] <= '9' {
		return "X" + s
	}
	return s
}
