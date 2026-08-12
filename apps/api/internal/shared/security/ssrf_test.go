package security

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
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

// withFakeResolver swaps lookupIPAddr for the duration of the test and
// restores it afterward.
func withFakeResolver(t *testing.T, fn func(ctx context.Context, host string) ([]net.IPAddr, error)) {
	t.Helper()
	original := lookupIPAddr
	lookupIPAddr = fn
	t.Cleanup(func() {
		lookupIPAddr = original
	})
}

// TestResolveHost_NeverCaches is a regression test: resolveHost must perform
// a fresh lookup on every call. It intentionally has no cache — see its doc
// comment — so a stale pre-flight answer can never be a factor, however
// briefly, in what ValidateURL approves.
func TestResolveHost_NeverCaches(t *testing.T) {
	var calls int32
	withFakeResolver(t, func(ctx context.Context, host string) ([]net.IPAddr, error) {
		atomic.AddInt32(&calls, 1)
		return []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}}, nil
	})

	for i := 0; i < 3; i++ {
		if _, err := resolveHost(context.Background(), "repeated.example.com"); err != nil {
			t.Fatalf("resolveHost call %d: unexpected error: %v", i, err)
		}
	}

	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("expected the underlying resolver to be called once per resolveHost call (no caching), got %d calls", got)
	}
}

func TestResolveHost_BoundsLookupWithTimeout(t *testing.T) {
	var sawDeadline bool
	var deadlineWithinBudget bool
	withFakeResolver(t, func(ctx context.Context, host string) ([]net.IPAddr, error) {
		deadline, ok := ctx.Deadline()
		sawDeadline = ok
		if ok {
			deadlineWithinBudget = time.Until(deadline) <= dnsLookupTimeout+time.Second
		}
		<-ctx.Done()
		return nil, ctx.Err()
	})

	_, err := resolveHost(context.Background(), "slow.example.com")
	if err == nil || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected a timeout error for a hanging resolver, got %v", err)
	}
	if !sawDeadline {
		t.Fatal("expected the resolver to receive a context with a deadline")
	}
	if !deadlineWithinBudget {
		t.Fatal("expected the deadline to be bounded by dnsLookupTimeout")
	}
}

type fakeConn struct {
	net.Conn
}

// withFakeDialer swaps dialContext for the duration of the test, so it never
// needs real network access, and restores it afterward.
func withFakeDialer(t *testing.T, fn func(ctx context.Context, network, addr string) (net.Conn, error)) {
	t.Helper()
	original := dialContext
	dialContext = fn
	t.Cleanup(func() {
		dialContext = original
	})
}

// TestSafeDialContext_RejectsBlockedLiteralIP is a regression test: dialing
// a raw loopback/private IP address must be rejected without ever reaching
// the real dialer.
func TestSafeDialContext_RejectsBlockedLiteralIP(t *testing.T) {
	var dialed bool
	withFakeDialer(t, func(ctx context.Context, network, addr string) (net.Conn, error) {
		dialed = true
		return nil, nil
	})

	_, err := SafeDialContext(context.Background(), "tcp", "127.0.0.1:8080")
	if err == nil {
		t.Fatal("expected error dialing a loopback IP")
	}
	if dialed {
		t.Fatal("expected the real dialer never to be invoked for a blocked IP")
	}
}

// TestSafeDialContext_RejectsInternalHostnameSuffix is a regression test:
// hostnames disallowed by name (e.g. .internal) must be rejected before any
// DNS resolution or dial is attempted.
func TestSafeDialContext_RejectsInternalHostnameSuffix(t *testing.T) {
	var resolved, dialed bool
	withFakeResolver(t, func(ctx context.Context, host string) ([]net.IPAddr, error) {
		resolved = true
		return []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}}, nil
	})
	withFakeDialer(t, func(ctx context.Context, network, addr string) (net.Conn, error) {
		dialed = true
		return nil, nil
	})

	_, err := SafeDialContext(context.Background(), "tcp", "service.internal:443")
	if err == nil {
		t.Fatal("expected error dialing an .internal hostname")
	}
	if resolved || dialed {
		t.Fatal("expected neither resolution nor dial to occur for a disallowed hostname")
	}
}

// TestSafeDialContext_PinsToResolvedAddress is a regression test for the
// DNS-rebinding gap: the address actually dialed must be the exact address
// that was just validated, not a second, independent resolution.
func TestSafeDialContext_PinsToResolvedAddress(t *testing.T) {
	withFakeResolver(t, func(ctx context.Context, host string) ([]net.IPAddr, error) {
		return []net.IPAddr{{IP: net.ParseIP("203.0.113.10")}}, nil
	})

	var dialedAddr string
	withFakeDialer(t, func(ctx context.Context, network, addr string) (net.Conn, error) {
		dialedAddr = addr
		return &fakeConn{}, nil
	})

	conn, err := SafeDialContext(context.Background(), "tcp", "api.example.com:443")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conn == nil {
		t.Fatal("expected a non-nil connection")
	}
	if dialedAddr != "203.0.113.10:443" {
		t.Fatalf("expected dial pinned to the resolved IP:port, got %q", dialedAddr)
	}
}

// TestSafeDialContext_SkipsBlockedAddressesInResolvedList is a regression
// test: if a hostname resolves to multiple addresses, a blocked one among
// them must not be dialed even if it's returned first.
func TestSafeDialContext_SkipsBlockedAddressesInResolvedList(t *testing.T) {
	withFakeResolver(t, func(ctx context.Context, host string) ([]net.IPAddr, error) {
		return []net.IPAddr{
			{IP: net.ParseIP("127.0.0.1")},
			{IP: net.ParseIP("203.0.113.20")},
		}, nil
	})

	var dialedAddr string
	withFakeDialer(t, func(ctx context.Context, network, addr string) (net.Conn, error) {
		dialedAddr = addr
		return &fakeConn{}, nil
	})

	_, err := SafeDialContext(context.Background(), "tcp", "multi.example.com:443")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dialedAddr != "203.0.113.20:443" {
		t.Fatalf("expected dial to skip the blocked address and use the allowed one, got %q", dialedAddr)
	}
}

// TestSafeDialContext_RejectsWhenAllResolvedAddressesBlocked is a regression
// test: a hostname that resolves only to disallowed addresses (e.g. after
// DNS rebinding) must be rejected outright.
func TestSafeDialContext_RejectsWhenAllResolvedAddressesBlocked(t *testing.T) {
	withFakeResolver(t, func(ctx context.Context, host string) ([]net.IPAddr, error) {
		return []net.IPAddr{{IP: net.ParseIP("127.0.0.1")}, {IP: net.ParseIP("169.254.169.254")}}, nil
	})
	withFakeDialer(t, func(ctx context.Context, network, addr string) (net.Conn, error) {
		t.Fatal("expected the dialer never to be invoked when every resolved address is blocked")
		return nil, nil
	})

	_, err := SafeDialContext(context.Background(), "tcp", "rebound.example.com:443")
	if err == nil {
		t.Fatal("expected error when every resolved address is blocked")
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

// TestSafeHTTPClient_PreservesDefaultTransportSettings is a regression test:
// SafeHTTPClient must keep proxy support, connection pooling, and HTTP/2
// from http.DefaultTransport, only swapping DialContext. A hand-built
// &http.Transport{DialContext: ...} would silently drop all of that.
func TestSafeHTTPClient_PreservesDefaultTransportSettings(t *testing.T) {
	client := SafeHTTPClient(10 * time.Second)

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("expected *http.Transport, got %T", client.Transport)
	}

	defaultTransport := http.DefaultTransport.(*http.Transport)

	if transport.Proxy == nil {
		t.Fatal("expected Proxy support (e.g. HTTP_PROXY/HTTPS_PROXY) to be preserved")
	}
	if transport.MaxIdleConns != defaultTransport.MaxIdleConns {
		t.Fatalf("expected MaxIdleConns %d, got %d", defaultTransport.MaxIdleConns, transport.MaxIdleConns)
	}
	if transport.IdleConnTimeout != defaultTransport.IdleConnTimeout {
		t.Fatalf("expected IdleConnTimeout %v, got %v", defaultTransport.IdleConnTimeout, transport.IdleConnTimeout)
	}
	if transport.ForceAttemptHTTP2 != defaultTransport.ForceAttemptHTTP2 {
		t.Fatalf("expected ForceAttemptHTTP2 %v, got %v", defaultTransport.ForceAttemptHTTP2, transport.ForceAttemptHTTP2)
	}
	if transport.DialContext == nil {
		t.Fatal("expected DialContext to be set to SafeDialContext")
	}
}
