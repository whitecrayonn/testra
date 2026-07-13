package middleware

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/google/uuid"
	"github.com/testra/testra/apps/api/internal/shared/errors"
	apihttp "github.com/testra/testra/apps/api/internal/shared/http"
)

const (
	permKey   contextKey = "permissions"
	tenantKey contextKey = "tenant_id"
)

func WithTenantID(ctx context.Context, tenantID uuid.UUID) context.Context {
	return context.WithValue(ctx, tenantKey, tenantID)
}

func TenantIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	v, ok := ctx.Value(tenantKey).(uuid.UUID)
	return v, ok
}

func WithPermissions(ctx context.Context, perms []string) context.Context {
	return context.WithValue(ctx, permKey, perms)
}

func PermissionsFromContext(ctx context.Context) []string {
	v, _ := ctx.Value(permKey).([]string)
	return v
}

type PermissionLoader interface {
	LoadPermissions(ctx context.Context, userID uuid.UUID, scopeType string, scopeID uuid.UUID) ([]string, error)
}

type RBACConfig struct {
	Loader PermissionLoader
}

func RequirePermission(cfg RBACConfig, permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := UserIDFromContext(r.Context())
			if !ok {
				apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", errors.ErrUnauthorized.Error())
				return
			}

			tenantID, ok := TenantIDFromContext(r.Context())
			if !ok {
				apihttp.ErrorJSON(w, http.StatusForbidden, "FORBIDDEN", "missing tenant context")
				return
			}

			perms := PermissionsFromContext(r.Context())
			if perms == nil {
				loaded, err := cfg.Loader.LoadPermissions(r.Context(), userID, "organization", tenantID)
				if err != nil {
					if err == sql.ErrNoRows {
						apihttp.ErrorJSON(w, http.StatusForbidden, "FORBIDDEN", errors.ErrForbidden.Error())
						return
					}
					apihttp.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
					return
				}
				perms = loaded
				r = r.WithContext(WithPermissions(r.Context(), perms))
			}

			for _, p := range perms {
				if p == permission {
					next.ServeHTTP(w, r)
					return
				}
			}

			apihttp.ErrorJSON(w, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
		})
	}
}
