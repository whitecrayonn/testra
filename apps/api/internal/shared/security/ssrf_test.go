package security

import (
	"context"
	"errors"
	"fmt"
	"net"
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

// withFakeResolver swaps lookupIPAddr for the duration of the test and clears
// the shared DNS cache before and after, so this test cannot see entries left
// by other tests (real hostnames like api.github.com) and vice versa.
func withFakeResolver(t *testing.T, fn func(ctx context.Context, host string) ([]net.IPAddr, error)) {
	t.Helper()
	original := lookupIPAddr
	lookupIPAddr = fn
	dnsCacheMu.Lock()
	dnsCache = map[string]dnsCacheEntry{}
	dnsCacheMu.Unlock()
	t.Cleanup(func() {
		lookupIPAddr = original
		dnsCacheMu.Lock()
		dnsCache = map[string]dnsCacheEntry{}
		dnsCacheMu.Unlock()
	})
}

func TestResolveHost_CachesResult(t *testing.T) {
	var calls int32
	withFakeResolver(t, func(ctx context.Context, host string) ([]net.IPAddr, error) {
		atomic.AddInt32(&calls, 1)
		return []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}}, nil
	})

	for i := 0; i < 3; i++ {
		addrs, err := resolveHost(context.Background(), "cached.example.com")
		if err != nil {
			t.Fatalf("resolveHost call %d: unexpected error: %v", i, err)
		}
		if len(addrs) != 1 || addrs[0].IP.String() != "8.8.8.8" {
			t.Fatalf("resolveHost call %d: unexpected addrs %v", i, addrs)
		}
	}

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected the underlying resolver to be called once (cached thereafter), got %d calls", got)
	}
}

func TestResolveHost_ReResolvesAfterTTLExpiry(t *testing.T) {
	var calls int32
	withFakeResolver(t, func(ctx context.Context, host string) ([]net.IPAddr, error) {
		atomic.AddInt32(&calls, 1)
		return []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}}, nil
	})

	if _, err := resolveHost(context.Background(), "expiring.example.com"); err != nil {
		t.Fatalf("first resolveHost call: unexpected error: %v", err)
	}

	// Force the cached entry to look expired instead of waiting out the real TTL.
	dnsCacheMu.Lock()
	entry := dnsCache["expiring.example.com"]
	entry.expires = time.Now().Add(-time.Second)
	dnsCache["expiring.example.com"] = entry
	dnsCacheMu.Unlock()

	if _, err := resolveHost(context.Background(), "expiring.example.com"); err != nil {
		t.Fatalf("second resolveHost call: unexpected error: %v", err)
	}

	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("expected a fresh lookup after TTL expiry, got %d total calls", got)
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

// TestResolveHost_PrunesExpiredEntries is a regression test: the DNS cache
// must not retain entries for hostnames that are no longer being queried,
// since ValidateURL runs against user-supplied endpoints (webhook/integration
// URLs) with an unbounded set of distinct hostnames.
func TestResolveHost_PrunesExpiredEntries(t *testing.T) {
	withFakeResolver(t, func(ctx context.Context, host string) ([]net.IPAddr, error) {
		return []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}}, nil
	})

	if _, err := resolveHost(context.Background(), "stale-host.example.com"); err != nil {
		t.Fatalf("resolveHost: unexpected error: %v", err)
	}

	dnsCacheMu.Lock()
	if len(dnsCache) != 1 {
		dnsCacheMu.Unlock()
		t.Fatalf("expected 1 cache entry, got %d", len(dnsCache))
	}
	// Force the entry to look expired, as if its TTL had already elapsed.
	entry := dnsCache["stale-host.example.com"]
	entry.expires = time.Now().Add(-time.Second)
	dnsCache["stale-host.example.com"] = entry
	dnsCacheMu.Unlock()

	// A lookup for a *different* host triggers the write-path prune, which
	// should evict the now-expired entry above rather than leaving it to
	// occupy memory indefinitely.
	if _, err := resolveHost(context.Background(), "other-host.example.com"); err != nil {
		t.Fatalf("resolveHost: unexpected error: %v", err)
	}

	dnsCacheMu.Lock()
	defer dnsCacheMu.Unlock()
	if _, ok := dnsCache["stale-host.example.com"]; ok {
		t.Fatal("expected the expired entry to have been pruned")
	}
	if len(dnsCache) != 1 {
		t.Fatalf("expected only the fresh entry to remain, got %d entries", len(dnsCache))
	}
}

// TestResolveHost_CapsCacheSize is a regression test: even within a single
// TTL window, the number of distinct cached hostnames must not grow without
// bound.
func TestResolveHost_CapsCacheSize(t *testing.T) {
	withFakeResolver(t, func(ctx context.Context, host string) ([]net.IPAddr, error) {
		return []net.IPAddr{{IP: net.ParseIP("8.8.8.8")}}, nil
	})

	for i := 0; i < dnsCacheMaxEntries+50; i++ {
		host := fmt.Sprintf("host-%d.example.com", i)
		if _, err := resolveHost(context.Background(), host); err != nil {
			t.Fatalf("resolveHost(%s): unexpected error: %v", host, err)
		}
	}

	dnsCacheMu.Lock()
	defer dnsCacheMu.Unlock()
	if len(dnsCache) > dnsCacheMaxEntries {
		t.Fatalf("expected cache size to stay capped at %d, got %d", dnsCacheMaxEntries, len(dnsCache))
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
