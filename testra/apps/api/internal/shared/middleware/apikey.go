package middleware

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/testra/testra/apps/api/internal/shared/db"
	apihttp "github.com/testra/testra/apps/api/internal/shared/http"
)

// APIKeyInfo is the subset of API key metadata required by middleware.  The
// apikeys.APIKey domain type satisfies this interface without importing this
// package, avoiding a circular dependency.
type APIKeyInfo interface {
	GetWorkspaceID() uuid.UUID
	GetOrganizationID() uuid.UUID
	GetCreatedBy() uuid.UUID
	GetScopes() []string
}

// APIKeyValidator validates a raw API key.  The apikeys.Service satisfies this
// when wrapped by an adapter in the server package.
type APIKeyValidator interface {
	Validate(ctx context.Context, rawKey string) (APIKeyInfo, error)
}

// nilTenantID is used while the real tenant is still unknown so that existing
// RLS policies see a valid UUID expression without matching any real tenant.
var nilTenantID = uuid.Nil.String()

type apiKeyContextKey string

const (
	apiKeyIDKey     apiKeyContextKey = "api_key_id"
	apiKeyScopesKey apiKeyContextKey = "api_key_scopes"
)

// APIKeyAuth authenticates requests using an API key, establishes the tenant
// context, and makes the key scopes available to downstream middleware.
func APIKeyAuth(pool *sql.DB, validator APIKeyValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawKey := extractAPIKey(r)
			if rawKey == "" {
				apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "API key required")
				return
			}

			ctx := r.Context()
			conn, err := pool.Conn(ctx)
			if err != nil {
				apihttp.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "database connection failed")
				return
			}

			released := false
			release := func() {
				if released {
					return
				}
				released = true
				_, _ = conn.ExecContext(context.Background(), "RESET app.tenant_id")
				_, _ = conn.ExecContext(context.Background(), "RESET app.lookup_key_hash")
				_ = conn.Close()
			}
			defer release()

			hash := hashAPIKey(rawKey)
			if _, err := conn.ExecContext(ctx, "SET app.tenant_id = $1", nilTenantID); err != nil {
				apihttp.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to initialize tenant context")
				return
			}
			if _, err := conn.ExecContext(ctx, "SET app.lookup_key_hash = $1", hash); err != nil {
				apihttp.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to initialize API key lookup")
				return
			}

			ctx = db.WithConn(ctx, conn)
			key, err := validator.Validate(ctx, rawKey)
			if err != nil {
				apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or revoked API key")
				return
			}

			if _, err := conn.ExecContext(ctx, "SET app.tenant_id = $1", key.GetOrganizationID().String()); err != nil {
				apihttp.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to set tenant context")
				return
			}

			ctx = db.WithTenantID(ctx, key.GetOrganizationID())
			ctx = WithUserID(ctx, key.GetCreatedBy())
			ctx = WithAPIKeyScopes(ctx, key.GetScopes())

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireScope enforces that the authenticated API key has the given scope.
func RequireScope(scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			scopes, ok := APIKeyScopesFromContext(r.Context())
			if !ok {
				apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "API key context required")
				return
			}
			for _, s := range scopes {
				if s == scope {
					next.ServeHTTP(w, r)
					return
				}
			}
			apihttp.ErrorJSON(w, http.StatusForbidden, "FORBIDDEN", "insufficient API key scope")
		})
	}
}

func extractAPIKey(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if h != "" {
		parts := strings.SplitN(h, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "ApiKey") {
			return parts[1]
		}
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return parts[1]
		}
	}
	return r.Header.Get("X-API-Key")
}

func hashAPIKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

func WithAPIKeyScopes(ctx context.Context, scopes []string) context.Context {
	return context.WithValue(ctx, apiKeyScopesKey, scopes)
}

func APIKeyScopesFromContext(ctx context.Context) ([]string, bool) {
	v, ok := ctx.Value(apiKeyScopesKey).([]string)
	return v, ok
}
