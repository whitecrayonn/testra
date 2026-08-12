package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/testra/testra/apps/api/internal/shared/jwt"
)

func TestAuthAcceptsBearerHeader(t *testing.T) {
	tm, err := jwt.NewTestManager("test-issuer", "test-audience")
	if err != nil {
		t.Fatalf("create token manager: %v", err)
	}
	userID := uuid.New()
	token, err := tm.Sign(userID, "test@example.com", time.Hour)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	handler := Auth(AuthConfig{TokenManager: tm})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

func TestAuthRejectsAccessTokenQueryParam(t *testing.T) {
	tm, err := jwt.NewTestManager("test-issuer", "test-audience")
	if err != nil {
		t.Fatalf("create token manager: %v", err)
	}
	userID := uuid.New()
	token, err := tm.Sign(userID, "test@example.com", time.Hour)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	handler := Auth(AuthConfig{TokenManager: tm})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/?access_token="+token, nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestAuthAcceptsAccessTokenCookie(t *testing.T) {
	tm, err := jwt.NewTestManager("test-issuer", "test-audience")
	if err != nil {
		t.Fatalf("create token manager: %v", err)
	}
	userID := uuid.New()
	token, err := tm.Sign(userID, "test@example.com", time.Hour)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	handler := Auth(AuthConfig{TokenManager: tm})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid, ok := UserIDFromContext(r.Context())
		if !ok || uid != userID {
			t.Fatalf("expected user id in context")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: AccessTokenCookieName, Value: token})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestAuthRejectsMissingToken(t *testing.T) {
	tm, _ := jwt.NewTestManager("test-issuer", "test-audience")
	handler := Auth(AuthConfig{TokenManager: tm})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

func TestAuthRejectsRevokedToken(t *testing.T) {
	tm, err := jwt.NewTestManager("test-issuer", "test-audience")
	if err != nil {
		t.Fatalf("create token manager: %v", err)
	}
	userID := uuid.New()
	token, err := tm.Sign(userID, "test@example.com", time.Hour)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	handler := Auth(AuthConfig{
		TokenManager: tm,
		IsRevoked: func(ctx context.Context, jti string) (bool, error) {
			return true, nil
		},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called for a revoked token")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for revoked token, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestAuthAllowsNonRevokedTokenWhenIsRevokedConfigured(t *testing.T) {
	tm, err := jwt.NewTestManager("test-issuer", "test-audience")
	if err != nil {
		t.Fatalf("create token manager: %v", err)
	}
	userID := uuid.New()
	token, err := tm.Sign(userID, "test@example.com", time.Hour)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	var checkedJTI string
	handler := Auth(AuthConfig{
		TokenManager: tm,
		IsRevoked: func(ctx context.Context, jti string) (bool, error) {
			checkedJTI = jti
			return false, nil
		},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for non-revoked token, got %d: %s", rr.Code, rr.Body.String())
	}
	if checkedJTI == "" {
		t.Fatal("expected IsRevoked to be called with a non-empty jti")
	}
}

func TestAuthFailsClosedWhenIsRevokedErrors(t *testing.T) {
	tm, err := jwt.NewTestManager("test-issuer", "test-audience")
	if err != nil {
		t.Fatalf("create token manager: %v", err)
	}
	userID := uuid.New()
	token, err := tm.Sign(userID, "test@example.com", time.Hour)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	handler := Auth(AuthConfig{
		TokenManager: tm,
		IsRevoked: func(ctx context.Context, jti string) (bool, error) {
			return false, errors.New("denylist lookup failed")
		},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called when the denylist check errors")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 when denylist check errors, got %d: %s", rr.Code, rr.Body.String())
	}
}
