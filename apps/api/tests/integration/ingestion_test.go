//go:build integration

package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

type ingestResult struct {
	RunID      string `json:"run_id"`
	Total      int    `json:"total"`
	Passed     int    `json:"passed"`
	Failed     int    `json:"failed"`
	Skipped    int    `json:"skipped"`
	DurationMs int64  `json:"duration_ms"`
}

func TestIngestJUnit(t *testing.T) {
	db := openTestDB(t)
	handler := newTestServer(db)
	ten := newTenant(t, db, ownerRoleID)
	apiKey := createAPIKey(t, handler, ten, []string{"runs:ingest"})

	payload := `<?xml version="1.0" encoding="UTF-8"?>
<testsuites>
  <testsuite name="Suite 1" tests="3" failures="1" errors="0" skipped="1" time="2.5">
    <testcase name="Test A" classname="ClassA" time="0.5"/>
    <testcase name="Test B" classname="ClassB" time="1.0">
      <failure message="assertion failed" type="AssertionError">stack trace here</failure>
    </testcase>
    <testcase name="Test C" classname="ClassC" time="1.0" status="skipped"/>
  </testsuite>
</testsuites>`

	body := map[string]any{
		"workspace_id": ten.WorkspaceID.String(),
		"project_id":   ten.ProjectID.String(),
		"name":         "JUnit CI Build",
		"format":       "junit",
		"payload":      payload,
	}

	rr := makeAPIKeyRequest(t, handler, "POST", "/api/v1/ingest", apiKey, uuid.New().String(), body)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var res ingestResult
	env := parseResponse(t, rr)
	if err := json.Unmarshal(env.Data, &res); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}

	if res.Total != 3 || res.Passed != 1 || res.Failed != 1 || res.Skipped != 1 || res.DurationMs != 2500 {
		t.Fatalf("unexpected result: %+v", res)
	}

	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM test_runs WHERE project_id = $1`, ten.ProjectID).Scan(&count)
	if err != nil {
		t.Fatalf("count runs: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 run, got %d", count)
	}
}

func TestIngestPlaywright(t *testing.T) {
	db := openTestDB(t)
	handler := newTestServer(db)
	ten := newTenant(t, db, ownerRoleID)
	apiKey := createAPIKey(t, handler, ten, []string{"runs:ingest"})

	payload := `{
  "suites": [
    {
      "title": "Suite 1",
      "status": "completed",
      "tests": [
        {"title": "Test A", "status": "passed", "duration": 500},
        {"title": "Test B", "status": "failed", "duration": 1000, "error": "timeout"},
        {"title": "Test C", "status": "skipped", "duration": 0}
      ]
    }
  ]
}`

	body := map[string]any{
		"workspace_id": ten.WorkspaceID.String(),
		"project_id":   ten.ProjectID.String(),
		"name":         "Playwright CI Build",
		"format":       "playwright",
		"payload":      payload,
	}

	rr := makeAPIKeyRequest(t, handler, "POST", "/api/v1/ingest", apiKey, uuid.New().String(), body)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var res ingestResult
	env := parseResponse(t, rr)
	if err := json.Unmarshal(env.Data, &res); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}

	if res.Total != 3 || res.Passed != 1 || res.Failed != 1 || res.Skipped != 1 {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestIngestCypress(t *testing.T) {
	db := openTestDB(t)
	handler := newTestServer(db)
	ten := newTenant(t, db, ownerRoleID)
	apiKey := createAPIKey(t, handler, ten, []string{"runs:ingest"})

	payload := `{
  "suites": [
    {
      "title": "Suite 1",
      "status": "completed",
      "tests": [
        {"title": "Test A", "status": "passed", "duration": 300},
        {"title": "Test B", "status": "failed", "duration": 800, "error": "assertion failed"}
      ]
    }
  ]
}`

	body := map[string]any{
		"workspace_id": ten.WorkspaceID.String(),
		"project_id":   ten.ProjectID.String(),
		"name":         "Cypress CI Build",
		"format":       "cypress",
		"payload":      payload,
	}

	rr := makeAPIKeyRequest(t, handler, "POST", "/api/v1/ingest", apiKey, uuid.New().String(), body)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var res ingestResult
	env := parseResponse(t, rr)
	if err := json.Unmarshal(env.Data, &res); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}

	if res.Total != 2 || res.Passed != 1 || res.Failed != 1 {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestIngestDuplicateUpload(t *testing.T) {
	db := openTestDB(t)
	handler := newTestServer(db)
	ten := newTenant(t, db, ownerRoleID)
	apiKey := createAPIKey(t, handler, ten, []string{"runs:ingest"})
	idempotencyKey := uuid.New().String()

	payload := `<testsuites><testsuite name="S" tests="1" time="0.1"><testcase name="T" time="0.1"/></testsuite></testsuites>`
	body := map[string]any{
		"workspace_id": ten.WorkspaceID.String(),
		"project_id":   ten.ProjectID.String(),
		"name":         "Duplicate Build",
		"format":       "junit",
		"payload":      payload,
	}

	rr1 := makeAPIKeyRequest(t, handler, "POST", "/api/v1/ingest", apiKey, idempotencyKey, body)
	if rr1.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr1.Code, rr1.Body.String())
	}

	var first ingestResult
	env1 := parseResponse(t, rr1)
	if err := json.Unmarshal(env1.Data, &first); err != nil {
		t.Fatalf("unmarshal first: %v", err)
	}

	rr2 := makeAPIKeyRequest(t, handler, "POST", "/api/v1/ingest", apiKey, idempotencyKey, body)
	if rr2.Code != http.StatusCreated {
		t.Fatalf("expected replay 201, got %d: %s", rr2.Code, rr2.Body.String())
	}

	var second ingestResult
	env2 := parseResponse(t, rr2)
	if err := json.Unmarshal(env2.Data, &second); err != nil {
		t.Fatalf("unmarshal second: %v", err)
	}

	if first.RunID != second.RunID {
		t.Fatalf("duplicate request returned different run_id: %s vs %s", first.RunID, second.RunID)
	}

	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM test_runs WHERE project_id = $1`, ten.ProjectID).Scan(&count)
	if err != nil {
		t.Fatalf("count runs: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 run after duplicate request, got %d", count)
	}
}

func TestIngestDuplicateKeyDifferentPayload(t *testing.T) {
	db := openTestDB(t)
	handler := newTestServer(db)
	ten := newTenant(t, db, ownerRoleID)
	apiKey := createAPIKey(t, handler, ten, []string{"runs:ingest"})
	idempotencyKey := uuid.New().String()

	body1 := map[string]any{
		"workspace_id": ten.WorkspaceID.String(),
		"project_id":   ten.ProjectID.String(),
		"name":         "Build A",
		"format":       "junit",
		"payload":      `<testsuites><testsuite name="S" tests="1" time="0.1"><testcase name="T" time="0.1"/></testsuite></testsuites>`,
	}
	rr1 := makeAPIKeyRequest(t, handler, "POST", "/api/v1/ingest", apiKey, idempotencyKey, body1)
	if rr1.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr1.Code, rr1.Body.String())
	}

	body2 := map[string]any{
		"workspace_id": ten.WorkspaceID.String(),
		"project_id":   ten.ProjectID.String(),
		"name":         "Build B",
		"format":       "junit",
		"payload":      `<testsuites><testsuite name="S" tests="1" time="0.2"><testcase name="T2" time="0.2"/></testsuite></testsuites>`,
	}
	rr2 := makeAPIKeyRequest(t, handler, "POST", "/api/v1/ingest", apiKey, idempotencyKey, body2)
	if rr2.Code != http.StatusConflict {
		t.Fatalf("expected 409 conflict, got %d: %s", rr2.Code, rr2.Body.String())
	}
}

// TestIngestMissingKey verifies that omitting the Idempotency-Key header
// does not block ingestion: IdempotencyKey (see
// internal/shared/middleware/idempotency.go) treats the header as optional
// and simply skips replay/conflict detection when absent, rather than
// rejecting the request.
func TestIngestMissingKey(t *testing.T) {
	db := openTestDB(t)
	handler := newTestServer(db)
	ten := newTenant(t, db, ownerRoleID)
	apiKey := createAPIKey(t, handler, ten, []string{"runs:ingest"})

	body := map[string]any{
		"workspace_id": ten.WorkspaceID.String(),
		"project_id":   ten.ProjectID.String(),
		"name":         "Missing Key",
		"format":       "junit",
		"payload":      `<testsuites><testsuite name="S" tests="1" time="0.1"><testcase name="T" time="0.1"/></testsuite></testsuites>`,
	}

	rr := makeAPIKeyRequest(t, handler, "POST", "/api/v1/ingest", apiKey, "", body)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 without an Idempotency-Key, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestIngestInvalidPayload(t *testing.T) {
	db := openTestDB(t)
	handler := newTestServer(db)
	ten := newTenant(t, db, ownerRoleID)
	apiKey := createAPIKey(t, handler, ten, []string{"runs:ingest"})

	body := map[string]any{
		"workspace_id": ten.WorkspaceID.String(),
		"project_id":   ten.ProjectID.String(),
		"name":         "Invalid",
		"format":       "junit",
		"payload":      "not valid xml",
	}

	rr := makeAPIKeyRequest(t, handler, "POST", "/api/v1/ingest", apiKey, uuid.New().String(), body)
	if rr.Code != http.StatusInternalServerError && rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 or 500 for invalid payload, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestIngestUnsupportedFormat(t *testing.T) {
	db := openTestDB(t)
	handler := newTestServer(db)
	ten := newTenant(t, db, ownerRoleID)
	apiKey := createAPIKey(t, handler, ten, []string{"runs:ingest"})

	body := map[string]any{
		"workspace_id": ten.WorkspaceID.String(),
		"project_id":   ten.ProjectID.String(),
		"name":         "Unsupported",
		"format":       "testcomplete",
		"payload":      "{}",
	}

	rr := makeAPIKeyRequest(t, handler, "POST", "/api/v1/ingest", apiKey, uuid.New().String(), body)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unsupported format, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestIngestUnauthorized(t *testing.T) {
	db := openTestDB(t)
	handler := newTestServer(db)
	ten := newTenant(t, db, ownerRoleID)

	body := map[string]any{
		"workspace_id": ten.WorkspaceID.String(),
		"project_id":   ten.ProjectID.String(),
		"name":         "Unauthorized",
		"format":       "junit",
		"payload":      `<testsuites><testsuite name="S" tests="1" time="0.1"><testcase name="T" time="0.1"/></testsuite></testsuites>`,
	}

	rr := makeRequest(t, handler, "POST", "/api/v1/ingest", "", uuid.New().String(), body)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestIngestTenantIsolation(t *testing.T) {
	db := openTestDB(t)
	handler := newTestServer(db)
	tenA := newTenant(t, db, ownerRoleID)
	tenB := newTenant(t, db, ownerRoleID)
	apiKeyB := createAPIKey(t, handler, tenB, []string{"runs:ingest"})

	body := map[string]any{
		"workspace_id": tenA.WorkspaceID.String(),
		"project_id":   tenA.ProjectID.String(),
		"name":         "Cross Tenant",
		"format":       "junit",
		"payload":      `<testsuites><testsuite name="S" tests="1" time="0.1"><testcase name="T" time="0.1"/></testsuite></testsuites>`,
	}

	rr := makeAPIKeyRequest(t, handler, "POST", "/api/v1/ingest", apiKeyB, uuid.New().String(), body)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for cross-tenant request, got %d: %s", rr.Code, rr.Body.String())
	}
}

// TestIngestForeignProjectWithOwnWorkspace guards against pairing a caller's
// own (legitimate) workspace_id with another tenant's project_id to smuggle
// results into a foreign project. The workspace-ownership check alone would
// pass here since tenA really does own workspace A; only an explicit
// project->workspace check catches the mismatch.
func TestIngestForeignProjectWithOwnWorkspace(t *testing.T) {
	db := openTestDB(t)
	handler := newTestServer(db)
	tenA := newTenant(t, db, ownerRoleID)
	tenB := newTenant(t, db, ownerRoleID)
	apiKeyA := createAPIKey(t, handler, tenA, []string{"runs:ingest"})

	body := map[string]any{
		"workspace_id": tenA.WorkspaceID.String(),
		"project_id":   tenB.ProjectID.String(),
		"name":         "Foreign Project",
		"format":       "junit",
		"payload":      `<testsuites><testsuite name="S" tests="1" time="0.1"><testcase name="T" time="0.1"/></testsuite></testsuites>`,
	}

	rr := makeAPIKeyRequest(t, handler, "POST", "/api/v1/ingest", apiKeyA, uuid.New().String(), body)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for a project outside the claimed workspace, got %d: %s", rr.Code, rr.Body.String())
	}
}

// TestIngestAPIKeyWorkspaceScope verifies that an API key provisioned for one
// workspace cannot be used to ingest into a sibling workspace in the same
// organization: the key's own workspace scope must match the request, not
// just the organization.
func TestIngestAPIKeyWorkspaceScope(t *testing.T) {
	db := openTestDB(t)
	handler := newTestServer(db)
	ten := newTenant(t, db, ownerRoleID)
	apiKey := createAPIKey(t, handler, ten, []string{"runs:ingest"})

	otherWorkspaceID := uuid.New()
	otherProjectID := uuid.New()
	if _, err := db.Exec(
		`INSERT INTO workspaces (id, organization_id, name, slug, created_at, updated_at) VALUES ($1, $2, $3, $4, NOW(), NOW())`,
		otherWorkspaceID, ten.OrgID, "Sibling Workspace", "sibling-"+otherWorkspaceID.String()); err != nil {
		t.Fatalf("insert sibling workspace: %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO projects (id, workspace_id, name, key, description, created_at, updated_at) VALUES ($1, $2, $3, $4, '', NOW(), NOW())`,
		otherProjectID, otherWorkspaceID, "Sibling Project", "SIBLING01"); err != nil {
		t.Fatalf("insert sibling project: %v", err)
	}

	body := map[string]any{
		"workspace_id": otherWorkspaceID.String(),
		"project_id":   otherProjectID.String(),
		"name":         "Sibling Workspace Ingest",
		"format":       "junit",
		"payload":      `<testsuites><testsuite name="S" tests="1" time="0.1"><testcase name="T" time="0.1"/></testsuite></testsuites>`,
	}

	rr := makeAPIKeyRequest(t, handler, "POST", "/api/v1/ingest", apiKey, uuid.New().String(), body)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for a workspace outside the key's own scope, got %d: %s", rr.Code, rr.Body.String())
	}
}

// TestIngestInsufficientPermission verifies that an API key lacking the
// runs:ingest scope is rejected, exercising the scope-based authorization
// that guards /ingest (see sharedmiddleware.RequireScope in server.go).
func TestIngestInsufficientPermission(t *testing.T) {
	db := openTestDB(t)
	handler := newTestServer(db)
	target := newTenant(t, db, ownerRoleID)
	apiKey := createAPIKey(t, handler, target, []string{"results:read"})

	body := map[string]any{
		"workspace_id": target.WorkspaceID.String(),
		"project_id":   target.ProjectID.String(),
		"name":         "No Permission",
		"format":       "junit",
		"payload":      `<testsuites><testsuite name="S" tests="1" time="0.1"><testcase name="T" time="0.1"/></testsuite></testsuites>`,
	}

	rr := makeAPIKeyRequest(t, handler, "POST", "/api/v1/ingest", apiKey, uuid.New().String(), body)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for insufficient permission, got %d: %s", rr.Code, rr.Body.String())
	}
}
