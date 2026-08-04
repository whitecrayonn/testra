// Package security provides basic Server-Side Request Forgery guards for outbound HTTP.
package security

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	// dnsCacheTTL bounds how long a resolved hostname is trusted before a
	// fresh lookup is required. Kept short (rather than a more typical
	// resolver TTL) because ValidateURL's result is used to gate an SSRF
	// check on a URL the caller then dials independently: caching widens the
	// window between "validated as public" and "actually connected to" that
	// a DNS-rebinding attacker could exploit by re-pointing the record. This
	// still gives repeat validations of the same host a cache hit without
	// leaving a long-lived blind spot.
	dnsCacheTTL = 30 * time.Second
	// dnsLookupTimeout bounds a single resolution so a slow or hanging
	// resolver cannot stall the caller beyond a fixed budget.
	dnsLookupTimeout = 2 * time.Second
	// dnsCacheMaxEntries bounds the cache's memory footprint. Entries are
	// user-influenced (any hostname passed to ValidateURL), so without a cap
	// the map would grow for the lifetime of the process.
	dnsCacheMaxEntries = 512
)

// lookupIPAddr resolves a hostname. It is a package-level variable so tests
// can substitute a fake resolver without depending on real DNS/network
// access.
var lookupIPAddr = net.DefaultResolver.LookupIPAddr

type dnsCacheEntry struct {
	addrs   []net.IPAddr
	expires time.Time
}

var (
	dnsCacheMu sync.Mutex
	dnsCache   = map[string]dnsCacheEntry{}
)

// resolveHost returns the resolved addresses for host, serving from an
// in-memory TTL cache when possible and otherwise performing a
// timeout-bounded lookup.
func resolveHost(ctx context.Context, host string) ([]net.IPAddr, error) {
	dnsCacheMu.Lock()
	if entry, ok := dnsCache[host]; ok && time.Now().Before(entry.expires) {
		dnsCacheMu.Unlock()
		return entry.addrs, nil
	}
	dnsCacheMu.Unlock()

	lookupCtx, cancel := context.WithTimeout(ctx, dnsLookupTimeout)
	defer cancel()

	addrs, err := lookupIPAddr(lookupCtx, host)
	if err != nil {
		return nil, err
	}

	dnsCacheMu.Lock()
	pruneExpiredDNSCacheLocked(time.Now())
	// If the cache is still full after pruning expired entries, skip caching
	// this result rather than growing further; the lookup above already
	// succeeded, so the caller is unaffected, and this bounds worst-case
	// memory even under a burst of many distinct hostnames within one TTL
	// window.
	if len(dnsCache) < dnsCacheMaxEntries {
		dnsCache[host] = dnsCacheEntry{addrs: addrs, expires: time.Now().Add(dnsCacheTTL)}
	}
	dnsCacheMu.Unlock()

	return addrs, nil
}

// pruneExpiredDNSCacheLocked removes expired entries from dnsCache. Callers
// must hold dnsCacheMu.
func pruneExpiredDNSCacheLocked(now time.Time) {
	for host, entry := range dnsCache {
		if now.After(entry.expires) {
			delete(dnsCache, host)
		}
	}
}

// ValidateURL blocks user-controlled URLs that point to local or private network
// destinations. It helps prevent server-side request forgery (SSRF) when the
// application is asked to call back to user-supplied endpoints.
//
// It rejects:
//   - non-HTTP(S) schemes
//   - empty or missing hostnames
//   - the literal host "localhost"
//   - IPv4/IPv6 loopback, link-local, multicast, and private-range addresses
//   - hostnames ending in .local, .localhost, or .internal
//   - DNS names that resolve only to blocked IP ranges
func ValidateURL(ctx context.Context, raw string) error {
	if raw == "" {
		return fmt.Errorf("empty URL")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("unsupported URL scheme %q", u.Scheme)
	}
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("missing URL host")
	}

	if strings.EqualFold(host, "localhost") {
		return fmt.Errorf("localhost is not allowed")
	}
	lowerHost := strings.ToLower(host)
	for _, suffix := range []string{".local", ".localhost", ".internal"} {
		if strings.HasSuffix(lowerHost, suffix) {
			return fmt.Errorf("internal hostname suffix %q is not allowed", suffix)
		}
	}

	if ip := net.ParseIP(host); ip != nil {
		if isBlockedIP(ip) {
			return fmt.Errorf("IP %s is not allowed", ip)
		}
		return nil
	}

	// For hostnames, resolve (via the cache, bounded by a timeout) and block
	// if any returned address is forbidden.
	addrs, err := resolveHost(ctx, host)
	if err != nil {
		// Treat resolution failure as a validation error rather than allowing
		// an uncontrolled outbound request.
		return fmt.Errorf("could not resolve host %q: %w", host, err)
	}
	if len(addrs) == 0 {
		return fmt.Errorf("host %q resolved to no addresses", host)
	}
	for _, addr := range addrs {
		if isBlockedIP(addr.IP) {
			return fmt.Errorf("host %q resolved to blocked IP %s", host, addr.IP)
		}
	}
	return nil
}

func isBlockedIP(ip net.IP) bool {
	return ip.IsLoopback() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsMulticast() ||
		ip.IsPrivate()
}
