package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/testra/testra/apps/api/internal/shared/jwt"
)

func TestAuthAcceptsBearerHeader(t *testing.T) {
	secret := "test-secret"
	userID := uuid.New()
	token, err := jwt.Sign(userID, "test@example.com", secret, time.Hour)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	handler := Auth(AuthConfig{JWTSecret: secret})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid, ok := UserIDFromContext(r.Context())
		if !ok || uid != userID {
			t.Fatalf("expected user id in context")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestAuthAcceptsAccessTokenQueryParam(t *testing.T) {
	secret := "test-secret"
	userID := uuid.New()
	token, err := jwt.Sign(userID, "test@example.com", secret, time.Hour)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	handler := Auth(AuthConfig{JWTSecret: secret})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid, ok := UserIDFromContext(r.Context())
		if !ok || uid != userID {
			t.Fatalf("expected user id in context")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/?access_token="+token, nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for query token, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestAuthRejectsMissingToken(t *testing.T) {
	handler := Auth(AuthConfig{JWTSecret: "secret"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "UNAUTHORIZED") {
		t.Fatalf("expected UNAUTHORIZED error body, got %s", rr.Body.String())
	}
}
