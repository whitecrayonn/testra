// Package security provides basic Server-Side Request Forgery guards for outbound HTTP.
package security

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// dnsLookupTimeout bounds a single DNS resolution so a slow or hanging
	// resolver cannot stall the caller beyond a fixed budget.
	dnsLookupTimeout = 2 * time.Second
	// dialTimeout bounds the TCP connect step performed by SafeDialContext.
	// Kept equal to net/http.DefaultTransport's own dial timeout (30s) so
	// SafeHTTPClient callers see the same connect-latency budget they had
	// before — a slow or distant legitimate endpoint should not start
	// failing just because SSRF protection was added.
	dialTimeout = 30 * time.Second
)

// lookupIPAddr resolves a hostname. It is a package-level variable so tests
// can substitute a fake resolver without depending on real DNS/network
// access.
var lookupIPAddr = net.DefaultResolver.LookupIPAddr

// dialContext performs the actual network dial for SafeDialContext. It is a
// package-level variable so tests can substitute a fake dialer without
// depending on real network access.
var dialContext = (&net.Dialer{Timeout: dialTimeout}).DialContext

// resolveHost resolves host with a bounded timeout. It intentionally does
// not cache: it backs ValidateURL's pre-flight check only (e.g. rejecting an
// obviously-bad URL early with a friendly error when a user saves an
// integration/webhook config), not the actual security boundary. The real
// enforcement happens in SafeDialContext, which always performs its own
// fresh, uncached resolution immediately before connecting — see its doc
// comment for why an earlier check (cached or not) can never be sufficient
// on its own. A cache here would only add a window in which this pre-flight
// check itself gives a stale answer, for no benefit, since it isn't consulted
// at connection time anyway.
func resolveHost(ctx context.Context, host string) ([]net.IPAddr, error) {
	lookupCtx, cancel := context.WithTimeout(ctx, dnsLookupTimeout)
	defer cancel()
	return lookupIPAddr(lookupCtx, host)
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

	if err := checkHostnameAllowed(host); err != nil {
		return err
	}

	if ip := net.ParseIP(host); ip != nil {
		if isBlockedIP(ip) {
			return fmt.Errorf("IP %s is not allowed", ip)
		}
		return nil
	}

	// For hostnames, resolve (bounded by a timeout) and block if any
	// returned address is forbidden.
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

// checkHostnameAllowed rejects hostnames that are disallowed by name alone,
// before any DNS resolution is attempted.
func checkHostnameAllowed(host string) error {
	if strings.EqualFold(host, "localhost") {
		return fmt.Errorf("localhost is not allowed")
	}
	lowerHost := strings.ToLower(host)
	for _, suffix := range []string{".local", ".localhost", ".internal"} {
		if strings.HasSuffix(lowerHost, suffix) {
			return fmt.Errorf("internal hostname suffix %q is not allowed", suffix)
		}
	}
	return nil
}

// SafeDialContext is a DialContext function for http.Transport that
// re-validates the destination immediately before connecting and pins the
// connection to the specific address it just validated.
//
// ValidateURL alone has an inherent gap: it checks a hostname, and the
// caller then makes a *separate* connection which performs its own,
// independent DNS resolution — an attacker who controls the target's DNS
// record can have it answer safely for the check and answer with an
// internal address moments later for the real connection (DNS rebinding).
// Wiring SafeDialContext into the client that actually performs the request
// closes that gap: the address that gets validated is the exact address
// that gets dialed, with no second lookup in between.
func SafeDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid dial address %q: %w", addr, err)
	}

	if ip := net.ParseIP(host); ip != nil {
		if isBlockedIP(ip) {
			return nil, fmt.Errorf("dial to blocked IP %s is not allowed", ip)
		}
		return dialContext(ctx, network, addr)
	}

	if err := checkHostnameAllowed(host); err != nil {
		return nil, err
	}

	lookupCtx, cancel := context.WithTimeout(ctx, dnsLookupTimeout)
	defer cancel()
	addrs, err := lookupIPAddr(lookupCtx, host)
	if err != nil {
		return nil, fmt.Errorf("could not resolve host %q: %w", host, err)
	}
	if len(addrs) == 0 {
		return nil, fmt.Errorf("host %q resolved to no addresses", host)
	}

	var lastErr error
	for _, a := range addrs {
		if isBlockedIP(a.IP) {
			continue
		}
		conn, dialErr := dialContext(ctx, network, net.JoinHostPort(a.IP.String(), port))
		if dialErr == nil {
			return conn, nil
		}
		lastErr = dialErr
	}
	if lastErr != nil {
		return nil, fmt.Errorf("host %q had no allowed address reachable: %w", host, lastErr)
	}
	return nil, fmt.Errorf("host %q resolved only to blocked addresses", host)
}

// SafeHTTPClient returns an *http.Client whose Transport dials through
// SafeDialContext, so a request to a user-supplied URL is protected against
// SSRF (including DNS rebinding between check and use) at the moment of the
// actual outbound connection, not just by an earlier ValidateURL call.
//
// The transport is cloned from http.DefaultTransport rather than built from
// scratch, so callers keep everything they'd otherwise silently lose:
// HTTP_PROXY/HTTPS_PROXY/NO_PROXY support, connection pooling
// (MaxIdleConns/IdleConnTimeout), and HTTP/2 — only DialContext is replaced.
func SafeHTTPClient(timeout time.Duration) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = SafeDialContext
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}
