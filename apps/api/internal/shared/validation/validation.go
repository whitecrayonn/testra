package validation

import (
	"net/mail"
	"strings"
	"unicode"
)

const (
	MaxEmailLength = 254
	MaxNameLength  = 100
	MinNameLength  = 1
)

func IsValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	if len(email) == 0 || len(email) > MaxEmailLength {
		return false
	}
	_, err := mail.ParseAddress(email)
	return err == nil
}

func IsValidName(name string) bool {
	name = strings.TrimSpace(name)
	return len(name) >= MinNameLength && len(name) <= MaxNameLength
}

func Slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else if r == ' ' || r == '-' || r == '_' {
			if b.Len() > 0 && b.String()[b.Len()-1] != '-' {
				b.WriteRune('-')
			}
		}
	}
	result := strings.Trim(b.String(), "-")
	if result == "" {
		result = "org"
	}
	return result
}
