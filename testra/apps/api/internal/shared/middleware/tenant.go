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
//
// ADR-013: before the tenant is known the middleware sets app.lookup_user_id
// so the lookup policies can resolve a workspace/project/run/defect ID to an
// organization. The lookup session variable is reset as soon as the tenant
// is established, preventing accidental cross-tenant reads.
func TenantContext(pool *sql.DB, resolveOrg OrgResolverFunc, checker MembershipChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := UserIDFromContext(r.Context())
			if !ok {
				apihttp.ErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing user context")
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
				// Clear the tenant-scoped session variables before returning the
				// connection to the pool to avoid leaking context to later users.
				_, _ = conn.ExecContext(context.Background(), "RESET app.tenant_id")
				_, _ = conn.ExecContext(context.Background(), "RESET app.lookup_user_id")
				_ = conn.Close()
			}
			defer release()

			// 1. Allow the connection to resolve the tenant by the authenticated user.
			if _, err := conn.ExecContext(ctx, "SET app.lookup_user_id = $1", userID.String()); err != nil {
				apihttp.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to set lookup user context")
				return
			}

			// 2. Bind the dedicated connection to the context so resolveOrg uses it.
			ctx = db.WithConn(ctx, conn)
			ctx = db.WithLookupUserID(ctx, userID)

			orgID, err := resolveOrg(r.WithContext(ctx))
			if err != nil {
				apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "could not resolve organization from request")
				return
			}

			// 3. Tenant is known: narrow the connection to that tenant only.
			if _, err := conn.ExecContext(ctx, "RESET app.lookup_user_id"); err != nil {
				apihttp.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to reset lookup user context")
				return
			}
			if _, err := conn.ExecContext(ctx, "SET app.tenant_id = $1", orgID.String()); err != nil {
				apihttp.ErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to set tenant context")
				return
			}

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
	ResolveOrgFromTestPlan(ctx context.Context, planID uuid.UUID) (uuid.UUID, error)
	ResolveOrgFromDefect(ctx context.Context, defectID uuid.UUID) (uuid.UUID, error)
	ResolveOrgFromAPICollection(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
	ResolveOrgFromAPIFolder(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
	ResolveOrgFromAPIRequest(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
	ResolveOrgFromAPIEnvironment(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
	ResolveOrgFromAPIRequestHistory(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
	ResolveOrgFromAutomationProject(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
	ResolveOrgFromAutomationExecution(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
	ResolveOrgFromAutomationArtifact(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
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

func TestPlanToOrg(extractID OrgResolverFunc, resolver WorkspaceOrgResolver) OrgResolverFunc {
	return func(r *http.Request) (uuid.UUID, error) {
		planID, err := extractID(r)
		if err != nil {
			return uuid.Nil, err
		}
		return resolver.ResolveOrgFromTestPlan(r.Context(), planID)
	}
}

func DefectToOrg(extractID OrgResolverFunc, resolver WorkspaceOrgResolver) OrgResolverFunc {
	return func(r *http.Request) (uuid.UUID, error) {
		defectID, err := extractID(r)
		if err != nil {
			return uuid.Nil, err
		}
		return resolver.ResolveOrgFromDefect(r.Context(), defectID)
	}
}

func APICollectionToOrg(extractID OrgResolverFunc, resolver WorkspaceOrgResolver) OrgResolverFunc {
	return func(r *http.Request) (uuid.UUID, error) {
		id, err := extractID(r)
		if err != nil {
			return uuid.Nil, err
		}
		return resolver.ResolveOrgFromAPICollection(r.Context(), id)
	}
}

func APIFolderToOrg(extractID OrgResolverFunc, resolver WorkspaceOrgResolver) OrgResolverFunc {
	return func(r *http.Request) (uuid.UUID, error) {
		id, err := extractID(r)
		if err != nil {
			return uuid.Nil, err
		}
		return resolver.ResolveOrgFromAPIFolder(r.Context(), id)
	}
}

func APIRequestToOrg(extractID OrgResolverFunc, resolver WorkspaceOrgResolver) OrgResolverFunc {
	return func(r *http.Request) (uuid.UUID, error) {
		id, err := extractID(r)
		if err != nil {
			return uuid.Nil, err
		}
		return resolver.ResolveOrgFromAPIRequest(r.Context(), id)
	}
}

func APIEnvironmentToOrg(extractID OrgResolverFunc, resolver WorkspaceOrgResolver) OrgResolverFunc {
	return func(r *http.Request) (uuid.UUID, error) {
		id, err := extractID(r)
		if err != nil {
			return uuid.Nil, err
		}
		return resolver.ResolveOrgFromAPIEnvironment(r.Context(), id)
	}
}

func APIRequestHistoryToOrg(extractID OrgResolverFunc, resolver WorkspaceOrgResolver) OrgResolverFunc {
	return func(r *http.Request) (uuid.UUID, error) {
		id, err := extractID(r)
		if err != nil {
			return uuid.Nil, err
		}
		return resolver.ResolveOrgFromAPIRequestHistory(r.Context(), id)
	}
}

func AutomationProjectToOrg(extractID OrgResolverFunc, resolver WorkspaceOrgResolver) OrgResolverFunc {
	return func(r *http.Request) (uuid.UUID, error) {
		id, err := extractID(r)
		if err != nil {
			return uuid.Nil, err
		}
		return resolver.ResolveOrgFromAutomationProject(r.Context(), id)
	}
}

func AutomationExecutionToOrg(extractID OrgResolverFunc, resolver WorkspaceOrgResolver) OrgResolverFunc {
	return func(r *http.Request) (uuid.UUID, error) {
		id, err := extractID(r)
		if err != nil {
			return uuid.Nil, err
		}
		return resolver.ResolveOrgFromAutomationExecution(r.Context(), id)
	}
}

func AutomationArtifactToOrg(extractID OrgResolverFunc, resolver WorkspaceOrgResolver) OrgResolverFunc {
	return func(r *http.Request) (uuid.UUID, error) {
		id, err := extractID(r)
		if err != nil {
			return uuid.Nil, err
		}
		return resolver.ResolveOrgFromAutomationArtifact(r.Context(), id)
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
