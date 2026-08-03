package middleware

import (
	"net"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

var (
	// emailPattern matches most common email addresses in text.
	emailPattern = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	// tokenPattern matches common bearer/api token values in query strings.
	tokenPattern = regexp.MustCompile(`(?i)(token|api[_-]?key|apikey|auth|password|passwd|pwd|secret|refresh_token|access_token)=([^&]*)`)
)

// RedactQuery removes sensitive values from a raw query string. It redacts
// values for known sensitive keys and masks any email addresses that appear.
func RedactQuery(query string) string {
	if query == "" {
		return ""
	}

	values, err := url.ParseQuery(query)
	if err != nil {
		// Fall back to regex-based redaction if the query is malformed.
		return tokenPattern.ReplaceAllString(query, `${1}=[REDACTED]`)
	}

	sensitiveKeys := map[string]bool{
		"token":         true,
		"api_key":       true,
		"apikey":        true,
		"api-key":       true,
		"auth":          true,
		"password":      true,
		"passwd":        true,
		"pwd":           true,
		"secret":        true,
		"refresh_token": true,
		"access_token":  true,
	}

	parts := make([]string, 0, len(values))
	for key, vals := range values {
		lower := strings.ToLower(key)
		for _, v := range vals {
			if sensitiveKeys[lower] {
				v = "[REDACTED]"
			} else {
				v = RedactPII(v)
			}
			parts = append(parts, key+"="+v)
		}
	}
	// Preserve a deterministic output order by sorting the key=value pairs.
	sort.Strings(parts)
	return strings.Join(parts, "&")
}

// RedactPII masks email addresses and token-like values in arbitrary text.
func RedactPII(text string) string {
	text = emailPattern.ReplaceAllString(text, "[EMAIL]")
	text = tokenPattern.ReplaceAllString(text, `${1}=[REDACTED]`)
	return text
}

// MaskIP returns a privacy-preserving representation of a client address. For
// IPv4 it replaces the last octet with zero; for IPv6 it replaces the last
// group with zeros. If parsing fails, the original value is returned.
func MaskIP(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
		port = ""
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return addr
	}

	if ip4 := ip.To4(); ip4 != nil {
		ip4[3] = 0
		host = ip4.String()
	} else {
		ip6 := ip.To16()
		ip6[15] = 0
		ip6[14] = 0
		host = ip6.String()
	}

	if port != "" {
		return net.JoinHostPort(host, port)
	}
	return host
}
