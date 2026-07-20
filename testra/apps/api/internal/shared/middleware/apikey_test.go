package middleware

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

type fakeAPIKeyInfo struct {
	workspaceID    uuid.UUID
	organizationID uuid.UUID
	createdBy      uuid.UUID
	scopes         []string
}

func (f fakeAPIKeyInfo) GetWorkspaceID() uuid.UUID    { return f.workspaceID }
func (f fakeAPIKeyInfo) GetOrganizationID() uuid.UUID { return f.organizationID }
func (f fakeAPIKeyInfo) GetCreatedBy() uuid.UUID      { return f.createdBy }
func (f fakeAPIKeyInfo) GetScopes() []string          { return f.scopes }

type fakeAPIKeyValidator struct {
	key APIKeyInfo
	err error
}

func (f *fakeAPIKeyValidator) Validate(_ context.Context, _ string) (APIKeyInfo, error) {
	return f.key, f.err
}

func TestExtractAPIKey(t *testing.T) {
	tests := []struct {
		name   string
		header string
		xkey   string
		want   string
	}{
		{"X-API-Key header", "", "secret", "secret"},
		{"Authorization ApiKey", "ApiKey secret", "", "secret"},
		{"Authorization Bearer fallback", "Bearer secret", "", "secret"},
		{"no key", "", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			if tt.xkey != "" {
				req.Header.Set("X-API-Key", tt.xkey)
			}
			if got := extractAPIKey(req); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestHashAPIKey(t *testing.T) {
	h1 := hashAPIKey("testra_abc")
	h2 := hashAPIKey("testra_abc")
	if h1 != h2 {
		t.Fatal("hash not deterministic")
	}
	if h1 == hashAPIKey("testra_xyz") {
		t.Fatal("hash collision")
	}
}

func TestAPIKeyScopesContext(t *testing.T) {
	scopes := []string{"runs:ingest"}
	ctx := WithAPIKeyScopes(context.Background(), scopes)
	got, ok := APIKeyScopesFromContext(ctx)
	if !ok || len(got) != 1 || got[0] != "runs:ingest" {
		t.Fatal("scopes not preserved in context")
	}
}

func TestRequireScope(t *testing.T) {
	handler := RequireScope("runs:ingest")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without scopes, got %d", rr.Code)
	}

	ctx := WithAPIKeyScopes(context.Background(), []string{"other:scope"})
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for missing scope, got %d", rr.Code)
	}

	ctx = WithAPIKeyScopes(context.Background(), []string{"runs:ingest"})
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 with scope, got %d", rr.Code)
	}
}

func TestAPIKeyAuthSuccess(t *testing.T) {
	registerFakeDriver()
	db, err := sql.Open("fakeapikey", "")
	if err != nil {
		t.Fatalf("open fake db: %v", err)
	}
	defer db.Close()

	orgID := uuid.New()
	wsID := uuid.New()
	userID := uuid.New()
	validator := &fakeAPIKeyValidator{
		key: fakeAPIKeyInfo{
			organizationID: orgID,
			workspaceID:    wsID,
			createdBy:      userID,
			scopes:         []string{"runs:ingest"},
		},
	}

	var ctxUser uuid.UUID
	handler := APIKeyAuth(db, validator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid, ok := UserIDFromContext(r.Context())
		if !ok {
			t.Fatal("expected user id in context")
		}
		ctxUser = uid
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "testra_validkey")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if ctxUser != userID {
		t.Fatalf("expected user %s, got %s", userID, ctxUser)
	}
}

func TestAPIKeyAuthRejectsMissingKey(t *testing.T) {
	registerFakeDriver()
	db, err := sql.Open("fakeapikey", "")
	if err != nil {
		t.Fatalf("open fake db: %v", err)
	}
	defer db.Close()

	handler := APIKeyAuth(db, &fakeAPIKeyValidator{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

func TestAPIKeyAuthRejectsInvalidKey(t *testing.T) {
	registerFakeDriver()
	db, err := sql.Open("fakeapikey", "")
	if err != nil {
		t.Fatalf("open fake db: %v", err)
	}
	defer db.Close()

	validator := &fakeAPIKeyValidator{err: http.ErrNoCookie}
	handler := APIKeyAuth(db, validator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "testra_invalid")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

// --- fake sql driver for middleware tests ---

func registerFakeDriver() {
	if !fakeDriverRegistered {
		sql.Register("fakeapikey", &fakeSQLDriver{})
		fakeDriverRegistered = true
	}
}

var fakeDriverRegistered bool

type fakeSQLDriver struct{}

func (d *fakeSQLDriver) Open(_ string) (driver.Conn, error) {
	return &fakeSQLConn{}, nil
}

type fakeSQLConn struct{}

func (c *fakeSQLConn) Prepare(_ string) (driver.Stmt, error) { return &fakeSQLStmt{}, nil }
func (c *fakeSQLConn) Close() error                          { return nil }
func (c *fakeSQLConn) Begin() (driver.Tx, error)             { return &fakeSQLTx{}, nil }
func (c *fakeSQLConn) ExecContext(_ context.Context, _ string, _ []driver.NamedValue) (driver.Result, error) {
	return &fakeSQLResult{}, nil
}
func (c *fakeSQLConn) QueryContext(_ context.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
	return &fakeSQLRows{}, nil
}

type fakeSQLStmt struct{}

func (s *fakeSQLStmt) Close() error                                 { return nil }
func (s *fakeSQLStmt) NumInput() int                                { return -1 }
func (s *fakeSQLStmt) Exec(_ []driver.Value) (driver.Result, error) { return &fakeSQLResult{}, nil }
func (s *fakeSQLStmt) Query(_ []driver.Value) (driver.Rows, error)  { return &fakeSQLRows{}, nil }

type fakeSQLTx struct{}

func (t *fakeSQLTx) Commit() error   { return nil }
func (t *fakeSQLTx) Rollback() error { return nil }

type fakeSQLResult struct{}

func (r *fakeSQLResult) LastInsertId() (int64, error) { return 0, nil }
func (r *fakeSQLResult) RowsAffected() (int64, error) { return 1, nil }

type fakeSQLRows struct{}

func (r *fakeSQLRows) Columns() []string           { return nil }
func (r *fakeSQLRows) Close() error                { return nil }
func (r *fakeSQLRows) Next(_ []driver.Value) error { return nil }
