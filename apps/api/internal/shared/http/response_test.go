package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
)

func TestMapError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"unwrapped not found", sharederrors.ErrNotFound, http.StatusNotFound, "NOT_FOUND"},
		{"wrapped invalid input", fmt.Errorf("validation: %w", sharederrors.ErrInvalidInput), http.StatusBadRequest, "INVALID_INPUT"},
		{"wrapped conflict", fmt.Errorf("db: %w", sharederrors.ErrConflict), http.StatusConflict, "CONFLICT"},
		{"unauthorized", sharederrors.ErrUnauthorized, http.StatusUnauthorized, "UNAUTHORIZED"},
		{"forbidden", sharederrors.ErrForbidden, http.StatusForbidden, "FORBIDDEN"},
		{"invalid credentials", sharederrors.ErrInvalidCredential, http.StatusUnauthorized, "INVALID_CREDENTIALS"},
		{"mfa required", sharederrors.ErrMFARequired, http.StatusUnauthorized, "MFA_REQUIRED"},
		{"token expired", sharederrors.ErrTokenExpired, http.StatusUnauthorized, "TOKEN_EXPIRED"},
		{"token revoked", sharederrors.ErrTokenRevoked, http.StatusUnauthorized, "TOKEN_REVOKED"},
		{"too many requests", sharederrors.ErrTooManyRequests, http.StatusTooManyRequests, "TOO_MANY_REQUESTS"},
		{"unknown error", fmt.Errorf("boom"), http.StatusInternalServerError, "INTERNAL_ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			MapError(rec, tt.err)

			if rec.Code != tt.wantStatus {
				t.Errorf("MapError status = %d, want %d", rec.Code, tt.wantStatus)
			}

			var env Envelope
			if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}
			if env.Error == nil {
				t.Fatalf("expected error envelope, got nil")
			}
			if env.Error.Code != tt.wantCode {
				t.Errorf("MapError code = %q, want %q", env.Error.Code, tt.wantCode)
			}
		})
	}
}
