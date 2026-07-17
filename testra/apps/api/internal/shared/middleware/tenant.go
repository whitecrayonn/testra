package middleware

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/testra/testra/apps/api/internal/shared/db"
	apihttp "github.com/testra/testra/apps/api/internal/shared/http"
)

type MembershipChecker interface {
	CheckMembership(ctx context.Context, userID, orgID uuid.UUID) error
}

type OrgResolverFunc func(*http.Request) (uuid.UUID, error)

// TenantContext resolves the tenant for the request, acquires a dedicated
// database connection, and sets the session-level app.tenant_id variable.
// All repository calls made while handling the request will use the same
// connection, ensuring PostgreSQL row-level security policies see the
// correct tenant.
func TenantContext(pool *sql.DB, resolveOrg OrgResolverFunc, checker MembershipChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := UserIDFromContext(r.Context())
			if !ok {
				apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user context")
				return
			}

			orgID, err := resolveOrg(r)
			if err != nil {
				apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "could not resolve organization from request")
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
				// Clear the tenant-scoped session variable before returning the
				// connection to the pool to avoid leaking context to later users.
				_, _ = conn.ExecContext(context.Background(), "RESET app.tenant_id")
				_ = conn.Close()
			}
			defer release()

			if _, err := conn.ExecContext(ctx, "SET app.tenant_id = $1", orgID.String()); err != nil {
				apihttp.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to set tenant context")
				return
			}

			ctx = db.WithConn(ctx, conn)
			ctx = db.WithTenantID(ctx, orgID)

			if err := checker.CheckMembership(ctx, userID, orgID); err != nil {
				apihttp.ErrorJSON(w, http.StatusForbidden, "FORBIDDEN", "not a member of this organization")
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type WorkspaceOrgResolver interface {
	ResolveOrgFromWorkspace(ctx context.Context, workspaceID uuid.UUID) (uuid.UUID, error)
	ResolveOrgFromProject(ctx context.Context, projectID uuid.UUID) (uuid.UUID, error)
	ResolveOrgFromAPIKey(ctx context.Context, apiKeyID uuid.UUID) (uuid.UUID, error)
	ResolveOrgFromRunItem(ctx context.Context, itemID uuid.UUID) (uuid.UUID, error)
	ResolveOrgFromRun(ctx context.Context, runID uuid.UUID) (uuid.UUID, error)
}

func WorkspaceToOrg(extractID OrgResolverFunc, resolver WorkspaceOrgResolver) OrgResolverFunc {
	return func(r *http.Request) (uuid.UUID, error) {
		wsID, err := extractID(r)
		if err != nil {
			return uuid.Nil, err
		}
		return resolver.ResolveOrgFromWorkspace(r.Context(), wsID)
	}
}

func ProjectToOrg(extractID OrgResolverFunc, resolver WorkspaceOrgResolver) OrgResolverFunc {
	return func(r *http.Request) (uuid.UUID, error) {
		projID, err := extractID(r)
		if err != nil {
			return uuid.Nil, err
		}
		return resolver.ResolveOrgFromProject(r.Context(), projID)
	}
}

func APIKeyToOrg(extractID OrgResolverFunc, resolver WorkspaceOrgResolver) OrgResolverFunc {
	return func(r *http.Request) (uuid.UUID, error) {
		keyID, err := extractID(r)
		if err != nil {
			return uuid.Nil, err
		}
		return resolver.ResolveOrgFromAPIKey(r.Context(), keyID)
	}
}

func RunItemToOrg(extractID OrgResolverFunc, resolver WorkspaceOrgResolver) OrgResolverFunc {
	return func(r *http.Request) (uuid.UUID, error) {
		itemID, err := extractID(r)
		if err != nil {
			return uuid.Nil, err
		}
		return resolver.ResolveOrgFromRunItem(r.Context(), itemID)
	}
}

func RunToOrg(extractID OrgResolverFunc, resolver WorkspaceOrgResolver) OrgResolverFunc {
	return func(r *http.Request) (uuid.UUID, error) {
		runID, err := extractID(r)
		if err != nil {
			return uuid.Nil, err
		}
		return resolver.ResolveOrgFromRun(r.Context(), runID)
	}
}

func OrgIDFromURLParam(param string) OrgResolverFunc {
	return func(r *http.Request) (uuid.UUID, error) {
		return uuid.Parse(chi.URLParam(r, param))
	}
}

func OrgIDFromQuery(param string) OrgResolverFunc {
	return func(r *http.Request) (uuid.UUID, error) {
		return uuid.Parse(r.URL.Query().Get(param))
	}
}

func OrgIDFromBody(field string) OrgResolverFunc {
	return func(r *http.Request) (uuid.UUID, error) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return uuid.Nil, err
		}
		r.Body = io.NopCloser(bytes.NewReader(body))

		var partial map[string]json.RawMessage
		if err := json.Unmarshal(body, &partial); err != nil {
			return uuid.Nil, err
		}
		raw, ok := partial[field]
		if !ok {
			return uuid.Nil, errMissingField
		}
		var idStr string
		if err := json.Unmarshal(raw, &idStr); err != nil {
			return uuid.Nil, err
		}
		return uuid.Parse(idStr)
	}
}

var errMissingField = &missingFieldError{}

type missingFieldError struct{}

func (e *missingFieldError) Error() string { return "required field missing from request body" }
