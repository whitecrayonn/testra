//go:build integration

package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRunProgressStreamRequiresAuth(t *testing.T) {
	db := openTestDB(t)
	handler := newTestServer(db)
	ten := newTenant(t, db, ownerRoleID)

	body := map[string]any{
		"workspace_id": ten.WorkspaceID.String(),
		"project_id":   ten.ProjectID.String(),
		"name":         "SSE Auth Test",
		"source":       "manual",
	}
	rr := makeRequest(t, handler, "POST", "/api/v1/test-runs", ten.Token, "", body)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
	env := parseResponse(t, rr)
	var run struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(env.Data, &run); err != nil {
		t.Fatalf("unmarshal run: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/v1/test-runs/"+run.ID+"/stream", nil)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req)
	if rr2.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d: %s", rr2.Code, rr2.Body.String())
	}
}

func TestRunProgressStreamRejectsQueryToken(t *testing.T) {
	db := openTestDB(t)
	handler := newTestServer(db)
	ten := newTenant(t, db, ownerRoleID)

	body := map[string]any{
		"workspace_id": ten.WorkspaceID.String(),
		"project_id":   ten.ProjectID.String(),
		"name":         "SSE Stream Test",
		"source":       "manual",
	}
	rr := makeRequest(t, handler, "POST", "/api/v1/test-runs", ten.Token, "", body)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
	env := parseResponse(t, rr)
	var run struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(env.Data, &run); err != nil {
		t.Fatalf("unmarshal run: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/v1/test-runs/"+run.ID+"/stream?access_token="+ten.Token, nil)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req)

	if rr2.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for query token, got %d: %s", rr2.Code, rr2.Body.String())
	}
}
