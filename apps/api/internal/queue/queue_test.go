package queue

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// openTestDB returns a connection to the integration database. The test is
// skipped if no database is available, but it should be run under -race in CI.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://testra:testra@localhost:5432/testra?sslmode=disable"
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Skipf("unable to open test database: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("test database unavailable: %v", err)
	}
	return db
}

func seedTenant(t *testing.T, db *sql.DB) uuid.UUID {
	t.Helper()
	orgID := uuid.New()
	userID := uuid.New()
	if _, err := db.Exec(
		"INSERT INTO users (id, email, password, name, created_at, updated_at) VALUES ($1, $2, $3, $4, NOW(), NOW())",
		userID, fmt.Sprintf("%s@testra.local", orgID), "hash", "test user"); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if _, err := db.Exec(
		"INSERT INTO organizations (id, name, slug, owner_id, created_at, updated_at) VALUES ($1, $2, $3, $4, NOW(), NOW())",
		orgID, orgID.String(), orgID.String(), userID); err != nil {
		t.Fatalf("insert org: %v", err)
	}
	return orgID
}

func insertJobs(t *testing.T, db *sql.DB, tenantID uuid.UUID, n int) []uuid.UUID {
	t.Helper()
	ids := make([]uuid.UUID, n)
	for i := 0; i < n; i++ {
		ids[i] = uuid.New()
		if _, err := db.Exec(
			`INSERT INTO queue_jobs (id, queue_name, job_type, payload, status, tenant_id, scheduled_at, created_at, updated_at)
			 VALUES ($1, 'test', 'concurrency', '{}'::jsonb, 'pending', $2, NOW(), NOW(), NOW())`,
			ids[i], tenantID); err != nil {
			t.Fatalf("insert job: %v", err)
		}
	}
	return ids
}

// TestDequeueConcurrency runs multiple workers that all call DequeueOne at the
// same time. Only one worker should receive each pending job because the query
// uses FOR UPDATE SKIP LOCKED. This also exercises the worker-mode RLS path
// under the race detector.
func TestDequeueConcurrency(t *testing.T) {
	if os.Getenv("CI") == "" && os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("concurrency test requires a database; set TEST_DATABASE_URL or run in CI")
	}

	db := openTestDB(t)
	defer db.Close()

	orgID := seedTenant(t, db)
	jobIDs := insertJobs(t, db, orgID, 5)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var mu sync.Mutex
	var got []uuid.UUID
	var errs []error

	var wg sync.WaitGroup
	for i := 0; i < len(jobIDs); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tx, job, err := DequeueOne(ctx, db, "test")
			if err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				return
			}
			if job == nil {
				return
			}
			if err := MarkDone(ctx, tx, job.ID); err != nil {
				_ = tx.Rollback()
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				return
			}
			if err := tx.Commit(); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				return
			}
			mu.Lock()
			got = append(got, job.ID)
			mu.Unlock()
		}()
	}
	wg.Wait()

	if len(errs) > 0 {
		t.Fatalf("worker errors: %v", errs)
	}
	if len(got) != len(jobIDs) {
		t.Fatalf("expected %d jobs, got %d", len(jobIDs), len(got))
	}

	seen := make(map[uuid.UUID]bool)
	for _, id := range got {
		if seen[id] {
			t.Fatalf("job %s was dequeued more than once", id)
		}
		seen[id] = true
	}

	for _, id := range jobIDs {
		if !seen[id] {
			t.Fatalf("job %s was not dequeued", id)
		}
	}
}
