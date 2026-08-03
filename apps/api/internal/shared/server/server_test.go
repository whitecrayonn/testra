package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestAPISecurityHeaders is a regression test for SBL-018/SBL-020: every
// /api/v1 response must carry the baseline hardening headers, regardless of
// which handler ultimately serves the request.
func TestAPISecurityHeaders(t *testing.T) {
	handler := apiSecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/anything", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	cases := []struct {
		header string
		want   string
	}{
		{"Cache-Control", "no-store, private, must-revalidate"},
		{"Pragma", "no-cache"},
		{"Expires", "0"},
		{"X-Content-Type-Options", "nosniff"},
		{"X-Frame-Options", "DENY"},
		{"Referrer-Policy", "strict-origin-when-cross-origin"},
		{"X-XSS-Protection", "0"},
		{"Permissions-Policy", "camera=(), microphone=(), geolocation=()"},
	}

	for _, tc := range cases {
		if got := rr.Header().Get(tc.header); got != tc.want {
			t.Errorf("header %s = %q, want %q", tc.header, got, tc.want)
		}
	}

	if got := rr.Header().Values("Vary"); len(got) == 0 || got[0] != "Authorization, Origin, Cookie" {
		t.Errorf("header Vary = %v, want [\"Authorization, Origin, Cookie\"]", got)
	}
}

func TestCORSMiddleware_AllowsConfiguredOrigin(t *testing.T) {
	handler := corsMiddleware("https://app.testra.example, https://other.example")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/anything", nil)
	req.Header.Set("Origin", "https://app.testra.example")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "https://app.testra.example" {
		t.Errorf("Access-Control-Allow-Origin = %q, want the matched configured origin", got)
	}
	if got := rr.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("Access-Control-Allow-Credentials = %q, want %q", got, "true")
	}
}

func TestCORSMiddleware_RejectsUnknownOrigin(t *testing.T) {
	handler := corsMiddleware("https://app.testra.example")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/anything", nil)
	req.Header.Set("Origin", "https://evil.example")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Access-Control-Allow-Origin = %q, want empty for an unrecognized origin", got)
	}
}
