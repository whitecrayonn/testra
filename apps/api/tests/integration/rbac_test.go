//go:build integration

package integration

import (
	"encoding/json"
	"net/http"
	"testing"
)

// TestRBAC_CrossTenantProjectAccessDenied verifies that a user in one
// organization cannot read a project that belongs to a different
// organization, even with a valid, otherwise-unrestricted (owner-role)
// token. TenantContext resolves the project's real owning org via a lookup
// connection (not scoped to the requester), then CheckMembership must reject
// the requester since they don't belong to that org — this is the core
// tenant-isolation guarantee the whole RBAC/RLS design depends on.
func TestRBAC_CrossTenantProjectAccessDenied(t *testing.T) {
	db := openTestDB(t)
	handler := newTestServer(db)

	tenantA := newTenant(t, db, ownerRoleID)
	tenantB := newTenant(t, db, ownerRoleID)

	// Sanity check: tenant A can read its own project.
	rr := makeRequest(t, handler, http.MethodGet, "/api/v1/projects/"+tenantA.ProjectID.String(), tenantA.Token, "", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected tenant A to read its own project (200), got %d: %s", rr.Code, rr.Body.String())
	}

	// Tenant B, a different organization entirely, must not be able to read
	// tenant A's project by guessing/reusing its ID.
	rr = makeRequest(t, handler, http.MethodGet, "/api/v1/projects/"+tenantA.ProjectID.String(), tenantB.Token, "", nil)
	if rr.Code == http.StatusOK {
		t.Fatalf("SECURITY: tenant B was able to read tenant A's project; expected denial, got 200: %s", rr.Body.String())
	}
	if rr.Code != http.StatusForbidden && rr.Code != http.StatusNotFound {
		t.Fatalf("expected 403 or 404 for cross-tenant project access, got %d: %s", rr.Code, rr.Body.String())
	}

	env := parseResponse(t, rr)
	if env.Error == nil {
		t.Fatal("expected an error envelope for cross-tenant access denial")
	}
}

// TestRBAC_ViewerCannotCreateProject verifies that the `viewer` system role
// (which is granted `projects:read` but not `projects:create`) is denied
// write access, while a user with the `owner` role in the same kind of
// organization succeeds under identical conditions. This exercises both
// halves of "permission grant/revoke": a role that lacks a permission is
// denied, and a role that has it is allowed.
func TestRBAC_ViewerCannotCreateProject(t *testing.T) {
	db := openTestDB(t)
	handler := newTestServer(db)

	viewer := newTenant(t, db, viewerRoleID)
	owner := newTenant(t, db, ownerRoleID)

	viewerBody := map[string]any{
		"workspace_id": viewer.WorkspaceID.String(),
		"name":         "Viewer Attempted Project",
		"key":          "VIEWDENY01",
	}
	rr := makeRequest(t, handler, http.MethodPost, "/api/v1/projects", viewer.Token, "", viewerBody)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected viewer role to be denied project creation (403), got %d: %s", rr.Code, rr.Body.String())
	}
	env := parseResponse(t, rr)
	if env.Error == nil || env.Error.Code != "FORBIDDEN" {
		t.Fatalf("expected FORBIDDEN error code, got %+v", env.Error)
	}

	ownerBody := map[string]any{
		"workspace_id": owner.WorkspaceID.String(),
		"name":         "Owner Allowed Project",
		"key":          "OWNERGRANT1",
	}
	rr = makeRequest(t, handler, http.MethodPost, "/api/v1/projects", owner.Token, "", ownerBody)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected owner role to be allowed to create a project (201), got %d: %s", rr.Code, rr.Body.String())
	}

	var created struct {
		ID string `json:"id"`
	}
	env = parseResponse(t, rr)
	if err := json.Unmarshal(env.Data, &created); err != nil {
		t.Fatalf("parse created project: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected created project to have an id")
	}
}

// TestRBAC_ViewerCanReadOwnProject verifies the positive counterpart of the
// permission checks above: a role that does have the required permission
// (viewer has projects:read) is allowed through, so the tests above aren't
// just measuring "everything gets denied".
func TestRBAC_ViewerCanReadOwnProject(t *testing.T) {
	db := openTestDB(t)
	handler := newTestServer(db)

	viewer := newTenant(t, db, viewerRoleID)

	rr := makeRequest(t, handler, http.MethodGet, "/api/v1/projects/"+viewer.ProjectID.String(), viewer.Token, "", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected viewer to read its own project (200), got %d: %s", rr.Code, rr.Body.String())
	}
}
