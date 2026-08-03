package middleware

import (
	"net/http"
	"strings"

	apihttp "github.com/testra/testra/apps/api/internal/shared/http"
)

// CSRFConfig configures the double-submit cookie CSRF protection middleware.
type CSRFConfig struct {
	// Skip receives the request and should return true for routes that should
	// not be protected by CSRF (e.g., login, registration, and password reset).
	Skip func(*http.Request) bool
}

// CSRF validates the double-submit cookie for state-changing requests. Safe
// methods (GET, HEAD, OPTIONS, TRACE) are always allowed through.
func CSRF(cfg CSRFConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !isMutatingMethod(r.Method) {
				next.ServeHTTP(w, r)
				return
			}

			if cfg.Skip != nil && cfg.Skip(r) {
				next.ServeHTTP(w, r)
				return
			}

			csrfCookie, err := r.Cookie(CSRFCookieName)
			if err != nil || csrfCookie.Value == "" {
				apihttp.ErrorJSON(w, http.StatusForbidden, "CSRF_TOKEN_MISSING", "CSRF token cookie missing")
				return
			}

			header := strings.TrimSpace(r.Header.Get("X-CSRF-Token"))
			if header == "" || !CSRFTokenEqual(header, csrfCookie.Value) {
				apihttp.ErrorJSON(w, http.StatusForbidden, "CSRF_TOKEN_INVALID", "CSRF token mismatch")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isMutatingMethod(method string) bool {
	switch strings.ToUpper(method) {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return false
	default:
		return true
	}
}
