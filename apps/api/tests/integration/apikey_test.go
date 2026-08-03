//go:build integration

package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

type apiKeyCreateResult struct {
	APIKey struct {
		ID string `json:"id"`
	} `json:"api_key"`
	RawKey string `json:"raw_key"`
}

func createAPIKey(t *testing.T, handler http.Handler, ten *testTenant) apiKeyCreateResult {
	t.Helper()

	body := map[string]any{
		"workspace_id": ten.WorkspaceID.String(),
		"name":         "Regression Key",
		"scopes":       []string{"runs:ingest"},
	}
	rr := makeRequest(t, handler, "POST", "/api/v1/api-keys", ten.Token, uuid.New().String(), body)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 creating api key, got %d: %s", rr.Code, rr.Body.String())
	}

	var res apiKeyCreateResult
	env := parseResponse(t, rr)
	if err := json.Unmarshal(env.Data, &res); err != nil {
		t.Fatalf("unmarshal api key: %v", err)
	}
	return res
}

func ingestWithAPIKey(handler http.Handler, ten *testTenant, header, scheme, key string) *httptest.ResponseRecorder {
	body := map[string]any{
		"workspace_id": ten.WorkspaceID.String(),
		"project_id":   ten.ProjectID.String(),
		"name":         "API Key Regression Build",
		"format":       "junit",
		"payload":      `<testsuites><testsuite name="S" tests="1" time="0.1"><testcase name="T" time="0.1"/></testsuite></testsuites>`,
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/ingest", strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", uuid.New().String())
	if header == "X-API-Key" {
		req.Header.Set("X-API-Key", key)
	} else {
		req.Header.Set("Authorization", scheme+" "+key)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

// TestAPIKeyAuthXAPIKeyHeader is a regression test for SBL-014: the /ingest
// endpoint must authenticate a valid key sent via the X-API-Key header.
func TestAPIKeyAuthXAPIKeyHeader(t *testing.T) {
	db := openTestDB(t)
	handler := newTestServer(db)
	ten := newTenant(t, db, ownerRoleID)

	created := createAPIKey(t, handler, ten)

	rr := ingestWithAPIKey(handler, ten, "X-API-Key", "", created.RawKey)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 with valid X-API-Key, got %d: %s", rr.Code, rr.Body.String())
	}
}

// TestAPIKeyAuthAuthorizationApiKeyScheme is a regression test for SBL-014:
// the /ingest endpoint must also accept the key via
// "Authorization: ApiKey <key>".
func TestAPIKeyAuthAuthorizationApiKeyScheme(t *testing.T) {
	db := openTestDB(t)
	handler := newTestServer(db)
	ten := newTenant(t, db, ownerRoleID)

	created := createAPIKey(t, handler, ten)

	rr := ingestWithAPIKey(handler, ten, "Authorization", "ApiKey", created.RawKey)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 with Authorization: ApiKey, got %d: %s", rr.Code, rr.Body.String())
	}
}

// TestAPIKeyAuthRejectsInvalidKey is a regression test for SBL-014: a
// well-formed but unknown key must be rejected with 401.
func TestAPIKeyAuthRejectsInvalidKey(t *testing.T) {
	db := openTestDB(t)
	handler := newTestServer(db)
	ten := newTenant(t, db, ownerRoleID)

	rr := ingestWithAPIKey(handler, ten, "X-API-Key", "", "testra_"+uuid.New().String())
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid key, got %d: %s", rr.Code, rr.Body.String())
	}
}

// TestAPIKeyAuthRejectsRevokedKey is a regression test for SBL-014: once an
// API key is revoked via DELETE /api-keys/{id}, it must stop authenticating
// immediately rather than remaining valid until its natural expiry.
func TestAPIKeyAuthRejectsRevokedKey(t *testing.T) {
	db := openTestDB(t)
	handler := newTestServer(db)
	ten := newTenant(t, db, ownerRoleID)

	created := createAPIKey(t, handler, ten)

	rr := ingestWithAPIKey(handler, ten, "X-API-Key", "", created.RawKey)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 before revocation, got %d: %s", rr.Code, rr.Body.String())
	}

	revokeRR := makeRequest(t, handler, "DELETE", "/api/v1/api-keys/"+created.APIKey.ID, ten.Token, uuid.New().String(), nil)
	if revokeRR.Code != http.StatusOK {
		t.Fatalf("expected 200 revoking api key, got %d: %s", revokeRR.Code, revokeRR.Body.String())
	}

	rr = ingestWithAPIKey(handler, ten, "X-API-Key", "", created.RawKey)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 after revocation, got %d: %s", rr.Code, rr.Body.String())
	}
	env := parseResponse(t, rr)
	if env.Error == nil || env.Error.Code != "UNAUTHORIZED" {
		t.Fatalf("expected UNAUTHORIZED error code, got %+v", env.Error)
	}
}
