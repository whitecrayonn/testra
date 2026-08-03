package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"
	"time"
)

// Cookie names used for browser session authentication and CSRF protection.
const (
	AccessTokenCookieName  = "testra_access_token"
	RefreshTokenCookieName = "testra_refresh_token"
	CSRFCookieName         = "testra_csrf_token"
)

// cookieMaxAgeRefreshToken is a sensible default for the CSRF cookie. The
// CSRF cookie is rotated with the session and should live at least as long as
// the refresh token.
const csrfCookieMaxAgeSeconds = 30 * 24 * 60 * 60

// newCookie creates a cookie with the shared security attributes used by the
// application. All auth/session cookies are HttpOnly and use SameSite=Lax by
// default; Secure is set based on whether the request was made over HTTPS so
// local development continues to work without TLS.
func newCookie(name, value string, maxAge int, httpOnly bool, r *http.Request) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: httpOnly,
		Secure:   IsRequestSecure(r),
		SameSite: http.SameSiteLaxMode,
	}
}

// IsRequestSecure reports whether the request is being served over HTTPS,
// either directly or via a trusted TLS-terminating proxy.
func IsRequestSecure(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	if strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		return true
	}
	return false
}

// SetAccessTokenCookie writes the short-lived access token cookie.
func SetAccessTokenCookie(w http.ResponseWriter, r *http.Request, token string, maxAge int) {
	http.SetCookie(w, newCookie(AccessTokenCookieName, token, maxAge, true, r))
}

// SetRefreshTokenCookie writes the long-lived refresh token cookie.
func SetRefreshTokenCookie(w http.ResponseWriter, r *http.Request, token string, maxAge int) {
	http.SetCookie(w, newCookie(RefreshTokenCookieName, token, maxAge, true, r))
}

// SetCSRFCookie writes the double-submit CSRF cookie. It is intentionally not
// HttpOnly so JavaScript can read it and send it back in the X-CSRF-Token header.
func SetCSRFCookie(w http.ResponseWriter, r *http.Request, token string, maxAge int) {
	http.SetCookie(w, newCookie(CSRFCookieName, token, maxAge, false, r))
}

// ClearAuthCookies clears all authentication and CSRF cookies.
func ClearAuthCookies(w http.ResponseWriter, r *http.Request) {
	clearCookie(w, r, AccessTokenCookieName, true)
	clearCookie(w, r, RefreshTokenCookieName, true)
	clearCookie(w, r, CSRFCookieName, false)
}

func clearCookie(w http.ResponseWriter, r *http.Request, name string, httpOnly bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: httpOnly,
		Secure:   IsRequestSecure(r),
		SameSite: http.SameSiteLaxMode,
	})
}

// AccessTokenFromCookie returns the access token value if present.
func AccessTokenFromCookie(r *http.Request) (string, bool) {
	c, err := r.Cookie(AccessTokenCookieName)
	if err != nil || c.Value == "" {
		return "", false
	}
	return c.Value, true
}

// RefreshTokenFromCookie returns the refresh token value if present.
func RefreshTokenFromCookie(r *http.Request) (string, bool) {
	c, err := r.Cookie(RefreshTokenCookieName)
	if err != nil || c.Value == "" {
		return "", false
	}
	return c.Value, true
}

// CSRFTokenFromCookie returns the CSRF token value if present.
func CSRFTokenFromCookie(r *http.Request) (string, bool) {
	c, err := r.Cookie(CSRFCookieName)
	if err != nil || c.Value == "" {
		return "", false
	}
	return c.Value, true
}

// GenerateCSRFToken creates a random token suitable for the double-submit cookie.
func GenerateCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// CSRFTokenEqual compares two CSRF tokens in constant time to avoid timing attacks.
func CSRFTokenEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// DefaultCSRFCookieMaxAge returns the default max-age for the CSRF cookie.
func DefaultCSRFCookieMaxAge() int { return csrfCookieMaxAgeSeconds }
