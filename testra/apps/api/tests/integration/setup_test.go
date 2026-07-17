//go:build integration

package integration

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testra/testra/apps/api/internal/shared/jwt"
	"github.com/testra/testra/apps/api/internal/shared/server"
	"golang.org/x/crypto/bcrypt"
)

const testJWTSecret = "integration-test-jwt-secret-do-not-use"

var ownerRoleID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
var viewerRoleID = uuid.MustParse("00000000-0000-0000-0000-000000000004")

type testTenant struct {
	UserID      uuid.UUID
	OrgID       uuid.UUID
	WorkspaceID uuid.UUID
	ProjectID   uuid.UUID
	Token       string
}

func databaseURL(t *testing.T) string {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL or DATABASE_URL not set; skipping integration test")
	}
	return dsn
}

func apiRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			panic("could not locate apps/api module root")
		}
		dir = parent
	}
}

func fileSource(path string) string {
	path = filepath.ToSlash(path)
	// file source driver strips "file://" and opens the remainder, so on Windows
	// an absolute path like C:/path must become file://C:/path (not file:///C:/path).
	return "file://" + path
}

func openTestDB(t *testing.T) *sql.DB {
	dsn := databaseURL(t)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("ping database: %v", err)
	}

	migrateDSN := dsn
	if strings.HasPrefix(migrateDSN, "postgres://") {
		migrateDSN = "pgx://" + migrateDSN[len("postgres://"):]
	}

	migrationsPath := filepath.Join(apiRoot(), "migrations")
	sourceURL := fileSource(migrationsPath)

	m, err := migrate.New(sourceURL, migrateDSN)
	if err != nil {
		t.Fatalf("create migrator: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		dirtyVersion := -1
		if de, ok := err.(*migrate.ErrDirty); ok {
			dirtyVersion = de.Version
		} else {
			var deAs *migrate.ErrDirty
			if errors.As(err, &deAs) {
				dirtyVersion = deAs.Version
			}
		}
		if dirtyVersion == -1 && strings.Contains(err.Error(), "Dirty database version") {
			parts := strings.SplitN(err.Error(), "Dirty database version ", 2)
			if len(parts) == 2 {
				vStr := strings.TrimSuffix(strings.Fields(parts[1])[0], ".")
				if v, convErr := strconv.Atoi(vStr); convErr == nil {
					dirtyVersion = v
				}
			}
		}
		if dirtyVersion > 1 {
			if forceErr := m.Force(dirtyVersion - 1); forceErr != nil {
				t.Fatalf("force clean migration state: %v", forceErr)
			}
			if upErr := m.Up(); upErr != nil && upErr != migrate.ErrNoChange {
				t.Fatalf("run migrations after force: %v", upErr)
			}
		} else {
			t.Fatalf("run migrations: %v", err)
		}
	}

	return db
}

func newTestServer(db *sql.DB) http.Handler {
	return server.New(server.Config{
		DB:                  db,
		JWTSecret:           testJWTSecret,
		JWTExpiry:           time.Hour,
		RefreshExpiryDays:   30,
		RefreshAbsoluteDays: 90,
		CORSAllowedOrigins:  "http://localhost:3000",
		IdempotencyKeyTTL:   time.Hour,
	})
}

func newTenant(t *testing.T, db *sql.DB, roleID uuid.UUID) *testTenant {
	t.Helper()

	userID := uuid.New()
	email := fmt.Sprintf("user-%s@example.com", userID.String())
	pwHash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO users (id, email, password, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())`,
		userID, email, string(pwHash), "Test User")
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	orgID := uuid.New()
	orgSlug := fmt.Sprintf("org-%s", orgID.String())
	_, err = db.Exec(`
		INSERT INTO organizations (id, name, slug, owner_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())`,
		orgID, fmt.Sprintf("Org %s", orgID.String()), orgSlug, userID)
	if err != nil {
		t.Fatalf("insert organization: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO organization_members (organization_id, user_id, role, created_at)
		VALUES ($1, $2, 'member', NOW())`,
		orgID, userID)
	if err != nil {
		t.Fatalf("insert organization_members: %v", err)
	}

	workspaceID := uuid.New()
	workspaceSlug := fmt.Sprintf("ws-%s", workspaceID.String())
	_, err = db.Exec(`
		INSERT INTO workspaces (id, organization_id, name, slug, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())`,
		workspaceID, orgID, fmt.Sprintf("Workspace %s", workspaceID.String()), workspaceSlug)
	if err != nil {
		t.Fatalf("insert workspace: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO workspace_members (workspace_id, user_id, role, created_at)
		VALUES ($1, $2, 'member', NOW())`,
		workspaceID, userID)
	if err != nil {
		t.Fatalf("insert workspace_members: %v", err)
	}

	projectID := uuid.New()
	projectKey := strings.ReplaceAll(projectID.String(), "-", "")[:10]
	_, err = db.Exec(`
		INSERT INTO projects (id, workspace_id, name, key, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, '', NOW(), NOW())`,
		projectID, workspaceID, fmt.Sprintf("Project %s", projectID.String()), projectKey)
	if err != nil {
		t.Fatalf("insert project: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO role_assignments (id, role_id, user_id, scope_type, scope_id, created_at)
		VALUES ($1, $2, $3, 'organization', $4, NOW())`,
		uuid.New(), roleID, userID, orgID)
	if err != nil {
		t.Fatalf("insert role_assignment: %v", err)
	}

	token, err := jwt.Sign(userID, email, testJWTSecret, time.Hour)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	return &testTenant{
		UserID:      userID,
		OrgID:       orgID,
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		Token:       token,
	}
}

func makeRequest(t *testing.T, handler http.Handler, method, path, token, idempotencyKey string, body map[string]any) *httptest.ResponseRecorder {
	t.Helper()

	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

type responseEnvelope struct {
	Data  json.RawMessage `json:"data"`
	Error *responseError  `json:"error"`
}

type responseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func parseResponse(t *testing.T, rr *httptest.ResponseRecorder) responseEnvelope {
	t.Helper()
	var env responseEnvelope
	if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
		t.Fatalf("parse response %s: %v", rr.Body.String(), err)
	}
	return env
}
