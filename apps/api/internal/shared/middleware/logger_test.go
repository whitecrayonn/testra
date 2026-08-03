package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func TestStructuredLoggerEmitsJSON(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if reqID := middleware.GetReqID(r.Context()); reqID != "" {
				w.Header().Set(middleware.RequestIDHeader, reqID)
			}
			next.ServeHTTP(w, r)
		})
	})
	r.Use(RequestLogger(NewStructuredLogFormatter(logger)))
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("short"))
	})

	req := httptest.NewRequest(http.MethodGet, "/test?foo=bar", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusTeapot {
		t.Fatalf("expected 418, got %d", rr.Code)
	}

	// The response should carry a request ID.
	requestID := rr.Header().Get(middleware.RequestIDHeader)
	if requestID == "" {
		t.Fatalf("expected %s header", middleware.RequestIDHeader)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	var last map[string]interface{}
	for _, line := range lines {
		if line == "" {
			continue
		}
		var entry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("invalid JSON log line %q: %v", line, err)
		}
		last = entry
	}
	if last == nil {
		t.Fatal("no log entry emitted")
	}

	if last["method"] != "GET" {
		t.Fatalf("expected method GET, got %v", last["method"])
	}
	if last["path"] != "/test" {
		t.Fatalf("expected path /test, got %v", last["path"])
	}
	if last["query"] != "foo=bar" {
		t.Fatalf("expected query foo=bar, got %v", last["query"])
	}
	if int(last["status"].(float64)) != http.StatusTeapot {
		t.Fatalf("expected status 418, got %v", last["status"])
	}
	if last["request_id"] != requestID {
		t.Fatalf("expected request_id %s, got %v", requestID, last["request_id"])
	}
}
