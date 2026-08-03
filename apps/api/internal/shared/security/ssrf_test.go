package security

import (
	"context"
	"net"
	"testing"
)

func TestValidateURL_Allowed(t *testing.T) {
	ctx := context.Background()
	cases := []string{
		"https://api.github.com",
		"https://example.com:443/path",
		"https://hooks.slack.com/services/T/B/123",
	}

	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			if err := ValidateURL(ctx, c); err != nil {
				t.Fatalf("ValidateURL(%q) returned unexpected error: %v", c, err)
			}
		})
	}
}

func TestValidateURL_Blocked(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		url     string
		wantErr string
	}{
		{"", "empty URL"},
		{"ftp://example.com", "unsupported URL scheme"},
		{"https://", "missing URL host"},
		{"https://localhost", "localhost is not allowed"},
		{"https://127.0.0.1", "is not allowed"},
		{"https://10.0.0.1", "is not allowed"},
		{"https://192.168.1.1", "is not allowed"},
		{"https://[::1]", "is not allowed"},
		{"https://[fe80::1]", "is not allowed"},
		{"https://my-service.local", "internal hostname suffix"},
		{"https://my-service.localhost", "internal hostname suffix"},
		{"https://my-service.internal", "internal hostname suffix"},
	}

	for _, tc := range cases {
		t.Run(tc.url, func(t *testing.T) {
			err := ValidateURL(ctx, tc.url)
			if err == nil {
				t.Fatalf("ValidateURL(%q) expected error containing %q, got nil", tc.url, tc.wantErr)
			}
			if !contains(err.Error(), tc.wantErr) {
				t.Fatalf("ValidateURL(%q) error %q does not contain %q", tc.url, err.Error(), tc.wantErr)
			}
		})
	}
}

func TestIsBlockedIP(t *testing.T) {
	tests := []struct {
		ip      string
		blocked bool
	}{
		{"127.0.0.1", true},
		{"::1", true},
		{"10.0.0.1", true},
		{"172.16.0.1", true},
		{"192.168.0.1", true},
		{"169.254.1.1", true},
		{"8.8.8.8", false},
		{"1.1.1.1", false},
		{"2001:4860:4860::8888", false},
	}

	for _, tc := range tests {
		t.Run(tc.ip, func(t *testing.T) {
			ip := net.ParseIP(tc.ip)
			if ip == nil {
				t.Fatalf("failed to parse IP %q", tc.ip)
			}
			if got := isBlockedIP(ip); got != tc.blocked {
				t.Fatalf("isBlockedIP(%q) = %v, want %v", tc.ip, got, tc.blocked)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && findSubstr(s, substr))
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
