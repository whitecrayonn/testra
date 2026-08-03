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

	rr := makeRequest(t, handler, "POST", "/api/v1/ingest", ten.Token, uuid.New().String(), body)
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

	rr := makeRequest(t, handler, "POST", "/api/v1/ingest", ten.Token, uuid.New().String(), body)
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

	rr := makeRequest(t, handler, "POST", "/api/v1/ingest", ten.Token, uuid.New().String(), body)
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
	idempotencyKey := uuid.New().String()

	payload := `<testsuites><testsuite name="S" tests="1" time="0.1"><testcase name="T" time="0.1"/></testsuite></testsuites>`
	body := map[string]any{
		"workspace_id": ten.WorkspaceID.String(),
		"project_id":   ten.ProjectID.String(),
		"name":         "Duplicate Build",
		"format":       "junit",
		"payload":      payload,
	}

	rr1 := makeRequest(t, handler, "POST", "/api/v1/ingest", ten.Token, idempotencyKey, body)
	if rr1.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr1.Code, rr1.Body.String())
	}

	var first ingestResult
	env1 := parseResponse(t, rr1)
	if err := json.Unmarshal(env1.Data, &first); err != nil {
		t.Fatalf("unmarshal first: %v", err)
	}

	rr2 := makeRequest(t, handler, "POST", "/api/v1/ingest", ten.Token, idempotencyKey, body)
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
	idempotencyKey := uuid.New().String()

	body1 := map[string]any{
		"workspace_id": ten.WorkspaceID.String(),
		"project_id":   ten.ProjectID.String(),
		"name":         "Build A",
		"format":       "junit",
		"payload":      `<testsuites><testsuite name="S" tests="1" time="0.1"><testcase name="T" time="0.1"/></testsuite></testsuites>`,
	}
	rr1 := makeRequest(t, handler, "POST", "/api/v1/ingest", ten.Token, idempotencyKey, body1)
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
	rr2 := makeRequest(t, handler, "POST", "/api/v1/ingest", ten.Token, idempotencyKey, body2)
	if rr2.Code != http.StatusConflict {
		t.Fatalf("expected 409 conflict, got %d: %s", rr2.Code, rr2.Body.String())
	}
}

func TestIngestMissingKey(t *testing.T) {
	db := openTestDB(t)
	handler := newTestServer(db)
	ten := newTenant(t, db, ownerRoleID)

	body := map[string]any{
		"workspace_id": ten.WorkspaceID.String(),
		"project_id":   ten.ProjectID.String(),
		"name":         "Missing Key",
		"format":       "junit",
		"payload":      `<testsuites><testsuite name="S" tests="1" time="0.1"><testcase name="T" time="0.1"/></testsuite></testsuites>`,
	}

	rr := makeRequest(t, handler, "POST", "/api/v1/ingest", ten.Token, "", body)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing Idempotency-Key, got %d: %s", rr.Code, rr.Body.String())
	}
	env := parseResponse(t, rr)
	if env.Error == nil || env.Error.Code != "IDEMPOTENCY_KEY_REQUIRED" {
		t.Fatalf("expected IDEMPOTENCY_KEY_REQUIRED, got %+v", env.Error)
	}
}

func TestIngestInvalidPayload(t *testing.T) {
	db := openTestDB(t)
	handler := newTestServer(db)
	ten := newTenant(t, db, ownerRoleID)

	body := map[string]any{
		"workspace_id": ten.WorkspaceID.String(),
		"project_id":   ten.ProjectID.String(),
		"name":         "Invalid",
		"format":       "junit",
		"payload":      "not valid xml",
	}

	rr := makeRequest(t, handler, "POST", "/api/v1/ingest", ten.Token, uuid.New().String(), body)
	if rr.Code != http.StatusInternalServerError && rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 or 500 for invalid payload, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestIngestUnsupportedFormat(t *testing.T) {
	db := openTestDB(t)
	handler := newTestServer(db)
	ten := newTenant(t, db, ownerRoleID)

	body := map[string]any{
		"workspace_id": ten.WorkspaceID.String(),
		"project_id":   ten.ProjectID.String(),
		"name":         "Unsupported",
		"format":       "robot",
		"payload":      "{}",
	}

	rr := makeRequest(t, handler, "POST", "/api/v1/ingest", ten.Token, uuid.New().String(), body)
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

	body := map[string]any{
		"workspace_id": tenA.WorkspaceID.String(),
		"project_id":   tenA.ProjectID.String(),
		"name":         "Cross Tenant",
		"format":       "junit",
		"payload":      `<testsuites><testsuite name="S" tests="1" time="0.1"><testcase name="T" time="0.1"/></testsuite></testsuites>`,
	}

	rr := makeRequest(t, handler, "POST", "/api/v1/ingest", tenB.Token, uuid.New().String(), body)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for cross-tenant request, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestIngestInsufficientPermission(t *testing.T) {
	db := openTestDB(t)
	handler := newTestServer(db)
	target := newTenant(t, db, ownerRoleID)
	viewer := newTenant(t, db, viewerRoleID)

	// Grant the viewer membership in the target organization by adding them to the target org.
	_, err := db.Exec(`
		INSERT INTO organization_members (organization_id, user_id, role, created_at)
		VALUES ($1, $2, 'member', NOW())
		ON CONFLICT DO NOTHING`,
		target.OrgID, viewer.UserID)
	if err != nil {
		t.Fatalf("grant membership: %v", err)
	}
	_, err = db.Exec(`
		INSERT INTO role_assignments (id, role_id, user_id, scope_type, scope_id, created_at)
		VALUES ($1, $2, $3, 'organization', $4, NOW())`,
		uuid.New(), viewerRoleID, viewer.UserID, target.OrgID)
	if err != nil {
		t.Fatalf("assign viewer role: %v", err)
	}

	body := map[string]any{
		"workspace_id": target.WorkspaceID.String(),
		"project_id":   target.ProjectID.String(),
		"name":         "No Permission",
		"format":       "junit",
		"payload":      `<testsuites><testsuite name="S" tests="1" time="0.1"><testcase name="T" time="0.1"/></testsuite></testsuites>`,
	}

	rr := makeRequest(t, handler, "POST", "/api/v1/ingest", viewer.Token, uuid.New().String(), body)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for insufficient permission, got %d: %s", rr.Code, rr.Body.String())
	}
}
