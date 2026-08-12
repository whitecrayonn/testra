package validation

import (
	"strings"
	"testing"
)

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"user@example.com", true},
		{"user.name+tag@example.co.uk", true},
		{"a@b.co", true},
		{"", false},
		{"  ", false},
		{"invalid", false},
		{"@example.com", false},
		{"user@", false},
		{strings.Repeat("a", 250) + "@example.com", false},
	}

	for _, tc := range tests {
		t.Run(tc.email, func(t *testing.T) {
			if got := IsValidEmail(tc.email); got != tc.valid {
				t.Fatalf("IsValidEmail(%q) = %v, want %v", tc.email, got, tc.valid)
			}
		})
	}
}

func TestIsValidName(t *testing.T) {
	tests := []struct {
		name  string
		valid bool
	}{
		{"Alice", true},
		{"A", true},
		{strings.Repeat("a", 100), true},
		{"", false},
		{"  ", false},
		{strings.Repeat("a", 101), false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsValidName(tc.name); got != tc.valid {
				t.Fatalf("IsValidName(%q) = %v, want %v", tc.name, got, tc.valid)
			}
		})
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Hello World", "hello-world"},
		{"My Org Name", "my-org-name"},
		{"  spaces   here  ", "spaces-here"},
		{"---leading---", "leading"},
		{"special!@#chars", "specialchars"},
		{"UPPERCASE", "uppercase"},
		{"", "org"},
		{"!!!", "org"},
		{"a-b_c d", "a-b-c-d"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			if got := Slugify(tc.input); got != tc.want {
				t.Fatalf("Slugify(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
