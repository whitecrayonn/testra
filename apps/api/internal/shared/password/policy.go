package password

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

// ErrPolicyViolation is returned when a password does not meet the policy.
var ErrPolicyViolation = errors.New("password does not meet security policy")

// Policy defines the password complexity and blocklist requirements.
type Policy struct {
	MinLength          int
	RequireUppercase   bool
	RequireLowercase   bool
	RequireDigit       bool
	RequireSpecialChar bool
	BlockedPasswords   map[string]bool
}

// DefaultPolicy returns the production-grade password policy used by Testra.
func DefaultPolicy() Policy {
	return Policy{
		MinLength:          12,
		RequireUppercase:   true,
		RequireLowercase:   true,
		RequireDigit:       true,
		RequireSpecialChar: true,
		BlockedPasswords:   commonPasswords(),
	}
}

// Validate checks a password against the policy and returns a descriptive
// error explaining which requirement was not met.
func (p Policy) Validate(password string) error {
	if len(password) < p.MinLength {
		return fmt.Errorf("password must be at least %d characters", p.MinLength)
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	)
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	if p.RequireUppercase && !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if p.RequireLowercase && !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if p.RequireDigit && !hasDigit {
		return errors.New("password must contain at least one digit")
	}
	if p.RequireSpecialChar && !hasSpecial {
		return errors.New("password must contain at least one special character")
	}

	if p.BlockedPasswords != nil && p.BlockedPasswords[strings.ToLower(password)] {
		return errors.New("password is too common or has been compromised; please choose a different one")
	}

	return nil
}

// commonPasswords returns a small, high-impact blocklist of breached or
// trivially guessable passwords. This is a pragmatic local substitute for an
// external breached-password API and avoids runtime network dependencies.
func commonPasswords() map[string]bool {
	list := []string{
		"password", "password123", "password123!", "password!123",
		"123456", "123456789", "qwerty", "qwerty123", "qwerty!123",
		"letmein", "letmein123", "welcome", "welcome123", "welcome!123",
		"admin", "admin123", "root", "root123",
		"testra", "testra123", "testra!123",
		"monkey", "monkey123", "dragon", "dragon123",
		"football", "football123", "baseball", "baseball123",
		"iloveyou", "iloveyou123", "trustno1", "trustno1!",
		"sunshine", "sunshine123", "princess", "princess123",
		"abc123", "abc123!", "abcd1234", "abcd1234!",
		"p@ssw0rd", "p@ssw0rd123", "passw0rd", "passw0rd!",
	}
	blocked := make(map[string]bool, len(list))
	for _, pw := range list {
		blocked[pw] = true
	}
	return blocked
}
