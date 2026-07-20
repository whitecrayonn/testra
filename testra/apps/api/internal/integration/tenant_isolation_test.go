//go:build integration

package integration

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/testra/testra/apps/api/internal/shared/db"
	"github.com/testra/testra/apps/api/internal/shared/tenant"
)

const appRole = "testra_app_test"

// TestTenantIsolation verifies that the ADR-013 RLS lookup policies correctly
// resolve tenant-scoped resources for an authenticated user and deny access to
// resources belonging to other tenants.
func TestTenantIsolation(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}

	adminDB, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open admin db: %v", err)
	}
	defer adminDB.Close()

	if err := resetSchema(adminDB); err != nil {
		t.Fatalf("reset schema: %v", err)
	}

	if err := runMigrations(dsn); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	if err := createAppRole(adminDB); err != nil {
		t.Fatalf("create app role: %v", err)
	}

	// Seed two isolated tenants with one user and workspace each.
	userA, orgA, wsA, err := seedTenant(adminDB, "tenant-a")
	if err != nil {
		t.Fatalf("seed tenant A: %v", err)
	}
	_, _, wsB, err := seedTenant(adminDB, "tenant-b")
	if err != nil {
		t.Fatalf("seed tenant B: %v", err)
	}

	// Connect as the non-superuser application role. This is the role that
	// RLS policies are enforced against in production.
	appDSN, err := rewriteUser(dsn, appRole, "test")
	if err != nil {
		t.Fatalf("rewrite app dsn: %v", err)
	}
	appDB, err := sql.Open("pgx", appDSN)
	if err != nil {
		t.Fatalf("open app db: %v", err)
	}
	defer appDB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := appDB.Conn(ctx)
	if err != nil {
		t.Fatalf("acquire conn: %v", err)
	}
	defer conn.Close()

	dbHandle := db.Wrap(appDB)
	resolver := tenant.NewResolver(dbHandle)

	// --- lookup mode: user A resolving their own workspace ---
	if _, err := conn.ExecContext(ctx, "SET app.lookup_user_id = $1", userA.String()); err != nil {
		t.Fatalf("set lookup user id: %v", err)
	}
	lookupCtx := db.WithConn(ctx, conn)
	lookupCtx = db.WithLookupUserID(lookupCtx, userA)

	orgID, err := resolver.ResolveOrgFromWorkspace(lookupCtx, wsA)
	if err != nil {
		t.Fatalf("expected user A to resolve workspace A: %v", err)
	}
	if orgID != orgA {
		t.Fatalf("expected org %s, got %s", orgA, orgID)
	}

	// --- lookup mode: user A resolving workspace B (cross-tenant) ---
	if _, err := conn.ExecContext(ctx, "SET app.lookup_user_id = $1", userA.String()); err != nil {
		t.Fatalf("set lookup user id for cross-tenant: %v", err)
	}
	_, err = resolver.ResolveOrgFromWorkspace(lookupCtx, wsB)
	if err == nil {
		t.Fatalf("expected cross-tenant workspace resolution to be denied")
	}

	// --- tenant mode: membership check works within the resolved tenant ---
	if _, err := conn.ExecContext(ctx, "RESET app.lookup_user_id"); err != nil {
		t.Fatalf("reset lookup user id: %v", err)
	}
	if _, err := conn.ExecContext(ctx, "SET app.tenant_id = $1", orgA.String()); err != nil {
		t.Fatalf("set tenant id: %v", err)
	}
	tenantCtx := db.WithTenantID(lookupCtx, orgA)

	if err := resolver.CheckMembership(tenantCtx, userA, orgA); err != nil {
		t.Fatalf("expected user A to be a member of org A: %v", err)
	}
}

func resetSchema(adminDB *sql.DB) error {
	if _, err := adminDB.Exec("DROP SCHEMA IF EXISTS public CASCADE"); err != nil {
		return fmt.Errorf("drop schema: %w", err)
	}
	if _, err := adminDB.Exec("CREATE SCHEMA public"); err != nil {
		return fmt.Errorf("create schema: %w", err)
	}
	return nil
}

func runMigrations(dsn string) error {
	migDSN := dsn
	if idx := strings.Index(dsn, "://"); idx != -1 {
		migDSN = "pgx://" + dsn[idx+3:]
	}

	m, err := migrate.New("file://../../migrations", migDSN)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer m.Close()
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}

func createAppRole(adminDB *sql.DB) error {
	if _, err := adminDB.Exec(fmt.Sprintf(`
		DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = '%s') THEN
				CREATE ROLE %s LOGIN PASSWORD 'test';
			END IF;
		END
		$$`, appRole, appRole)); err != nil {
		return fmt.Errorf("create role: %w", err)
	}
	if _, err := adminDB.Exec("GRANT USAGE, CREATE ON SCHEMA public TO " + appRole); err != nil {
		return fmt.Errorf("grant schema: %w", err)
	}
	if _, err := adminDB.Exec("GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO " + appRole); err != nil {
		return fmt.Errorf("grant tables: %w", err)
	}
	if _, err := adminDB.Exec("GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO " + appRole); err != nil {
		return fmt.Errorf("grant sequences: %w", err)
	}
	return nil
}

func seedTenant(adminDB *sql.DB, slug string) (userID, orgID, workspaceID uuid.UUID, err error) {
	userID = uuid.New()
	orgID = uuid.New()
	workspaceID = uuid.New()

	if _, err := adminDB.Exec(
		"INSERT INTO users (id, email, password, name, created_at, updated_at) VALUES ($1, $2, $3, $4, NOW(), NOW())",
		userID, slug+"@testra.local", "hash", slug+" user"); err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("insert user: %w", err)
	}
	if _, err := adminDB.Exec(
		"INSERT INTO organizations (id, name, created_at, updated_at) VALUES ($1, $2, NOW(), NOW())",
		orgID, slug); err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("insert org: %w", err)
	}
	if _, err := adminDB.Exec(
		"INSERT INTO organization_members (organization_id, user_id, role, created_at, updated_at) VALUES ($1, $2, $3, NOW(), NOW())",
		orgID, userID, "owner"); err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("insert member: %w", err)
	}
	if _, err := adminDB.Exec(
		"INSERT INTO workspaces (id, organization_id, name, slug, created_at, updated_at) VALUES ($1, $2, $3, $4, NOW(), NOW())",
		workspaceID, orgID, slug, slug); err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, fmt.Errorf("insert workspace: %w", err)
	}
	return userID, orgID, workspaceID, nil
}

func rewriteUser(dsn, user, password string) (string, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return "", err
	}
	u.User = url.UserPassword(user, password)
	return u.String(), nil
}
