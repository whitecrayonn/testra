package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"

	"github.com/testra/testra/apps/api/internal/shared/db"
	"github.com/testra/testra/apps/api/internal/shared/tenant"
)

func TestTenantContextResolvesWithLookupUser(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("creating sqlmock: %v", err)
	}
	defer sqlDB.Close()

	dbHandle := db.Wrap(sqlDB)
	resolver := tenant.NewResolver(dbHandle)
	checker := tenant.NewResolver(dbHandle)

	userID := uuid.MustParse("c2c2c2c2-c2c2-c2c2-c2c2-c2c2c2c2c2c2")
	wsID := uuid.MustParse("a0a0a0a0-a0a0-a0a0-a0a0-a0a0a0a0a0a0")
	orgID := uuid.MustParse("b1b1b1b1-b1b1-b1b1-b1b1-b1b1b1b1b1b1")

	// Expectations for the dedicated connection lifecycle.
	mock.ExpectExec("SET app.lookup_user_id = \\$1").WithArgs(userID.String()).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT organization_id FROM workspaces WHERE id = \\$1").WithArgs(wsID).WillReturnRows(sqlmock.NewRows([]string{"organization_id"}).AddRow(orgID))
	mock.ExpectExec("RESET app.lookup_user_id").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("SET app.tenant_id = \\$1").WithArgs(orgID.String()).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT EXISTS").WithArgs(orgID, userID).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec("RESET app.tenant_id").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("RESET app.lookup_user_id").WillReturnResult(sqlmock.NewResult(0, 0))

	var capturedOrg uuid.UUID
	handler := TenantContext(sqlDB,
		func(r *http.Request) (uuid.UUID, error) {
			return resolver.ResolveOrgFromWorkspace(r.Context(), wsID)
		},
		checker,
	)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		org, _ := db.TenantIDFromContext(r.Context())
		capturedOrg = org
		w.WriteHeader(http.StatusOK)
	}))

	ctx := WithUserID(context.Background(), userID)
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if capturedOrg != orgID {
		t.Fatalf("expected org %s in context, got %s", orgID, capturedOrg)
	}
}
