package middleware

import "testing"

func TestRedactQuery(t *testing.T) {
	cases := map[string]string{
		"email=admin@example.com&name=test": "email=[EMAIL]&name=test",
		"api_key=supersecret&other=value":   "api_key=[REDACTED]&other=value",
		"password=12345&token=abc":          "password=[REDACTED]&token=[REDACTED]",
		"query=admin@example.com":           "query=[EMAIL]",
		"":                                  "",
		"foo=bar":                           "foo=bar",
	}

	for input, expected := range cases {
		got := RedactQuery(input)
		if got != expected {
			t.Fatalf("RedactQuery(%q) = %q, want %q", input, got, expected)
		}
	}
}

func TestMaskIP(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"192.168.1.100:54321", "192.168.1.0:54321"},
		{"192.168.1.100", "192.168.1.0"},
		{"[::1]:12345", "[::]:12345"},
		{"10.0.0.5", "10.0.0.0"},
		{"not-an-ip", "not-an-ip"},
	}

	for _, c := range cases {
		got := MaskIP(c.input)
		if got != c.expected {
			t.Fatalf("MaskIP(%q) = %q, want %q", c.input, got, c.expected)
		}
	}
}
