//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/testra/testra/apps/api/internal/queue"
)

// TestIngestionQueueFlow simulates the full runtime path:
//   1. Ingest a test result (with an idempotency key).
//   2. Queue an asynchronous analytics job bound to the tenant.
//   3. Dequeue, prepare an execution payload, and complete the job.
//   4. Replay the ingest with the same idempotency key to confirm idempotency.
//
// The goal is to confirm the ingestion → queue → execution sequence does not
// deadlock and respects tenant context.
func TestIngestionQueueFlow(t *testing.T) {
	db := openTestDB(t)
	handler := newTestServer(db)
	ten := newTenant(t, db, ownerRoleID)

	payload := `<testsuites><testsuite name="S" tests="1" time="0.1"><testcase name="T" time="0.1"/></testsuite></testsuites>`
	body := map[string]any{
		"workspace_id": ten.WorkspaceID.String(),
		"project_id":   ten.ProjectID.String(),
		"name":         "Queue Flow Build",
		"format":       "junit",
		"payload":      payload,
	}
	idempotencyKey := uuid.New().String()

	rr := makeRequest(t, handler, "POST", "/api/v1/ingest", ten.Token, idempotencyKey, body)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var res ingestResult
	env := parseResponse(t, rr)
	if err := json.Unmarshal(env.Data, &res); err != nil {
		t.Fatalf("unmarshal ingest result: %v", err)
	}

	jobPayload := map[string]any{
		"run_id":      res.RunID,
		"workspace_id": ten.WorkspaceID.String(),
		"project_id":   ten.ProjectID.String(),
	}
	if err := queue.Enqueue(context.Background(), db, ten.OrgID, "default", "analytics:aggregate", jobPayload); err != nil {
		t.Fatalf("enqueue analytics job: %v", err)
	}

	ctx := context.Background()
	tx, job, err := queue.DequeueOne(ctx, db, "default")
	if err != nil {
		t.Fatalf("dequeue job: %v", err)
	}
	if job == nil {
		t.Fatalf("expected a pending job after enqueue")
	}
	if job.TenantID != ten.OrgID {
		t.Fatalf("job tenant mismatch: got %s want %s", job.TenantID, ten.OrgID)
	}
	if job.JobType != "analytics:aggregate" {
		t.Fatalf("unexpected job type: %s", job.JobType)
	}

	// Bind the transaction to the tenant so downstream data access obeys RLS.
	if _, err := tx.ExecContext(ctx, "SET LOCAL app.tenant_id = $1", ten.OrgID.String()); err != nil {
		t.Fatalf("set tenant context on worker tx: %v", err)
	}

	var prepared map[string]any
	if err := job.ParsePayload(&prepared); err != nil {
		t.Fatalf("parse execution payload: %v", err)
	}
	if prepared["run_id"] != res.RunID {
		t.Fatalf("execution payload missing run_id: %+v", prepared)
	}

	if err := queue.MarkDone(ctx, tx, job.ID); err != nil {
		t.Fatalf("mark job done: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit worker tx: %v", err)
	}

	var status string
	if err := db.QueryRow("SELECT status FROM queue_jobs WHERE id = $1", job.ID).Scan(&status); err != nil {
		t.Fatalf("lookup job status: %v", err)
	}
	if status != "completed" {
		t.Fatalf("expected job status completed, got %s", status)
	}

	// Idempotency replay must return the same run and not create a new one.
	rr2 := makeRequest(t, handler, "POST", "/api/v1/ingest", ten.Token, idempotencyKey, body)
	if rr2.Code != http.StatusCreated {
		t.Fatalf("expected replay 201, got %d: %s", rr2.Code, rr2.Body.String())
	}

	var replay ingestResult
	env2 := parseResponse(t, rr2)
	if err := json.Unmarshal(env2.Data, &replay); err != nil {
		t.Fatalf("unmarshal replay result: %v", err)
	}
	if replay.RunID != res.RunID {
		t.Fatalf("idempotency replay returned different run_id: %s vs %s", replay.RunID, res.RunID)
	}

	var runCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM test_runs WHERE project_id = $1", ten.ProjectID).Scan(&runCount); err != nil {
		t.Fatalf("count test runs: %v", err)
	}
	if runCount != 1 {
		t.Fatalf("idempotency replay should produce exactly one run, got %d", runCount)
	}
}
