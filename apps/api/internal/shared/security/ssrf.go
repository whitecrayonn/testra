// Package security provides basic Server-Side Request Forgery guards for outbound HTTP.
package security

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
)

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

	// For hostnames, resolve and block if any returned address is forbidden.
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
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
