# Live Test Run Updates

This document describes the end-to-end implementation of live test run progress
updates using Server-Sent Events (SSE).

## What changed

### Backend

- `apps/api/internal/shared/middleware/auth.go`
  - `Auth` now extracts the JWT from the `Authorization: Bearer` header **or**
    an `access_token` query parameter. This lets browser `EventSource`
    connections authenticate, because `EventSource` cannot set custom headers.

- `apps/api/internal/shared/tenant/resolver.go`
  - Added `ResolveOrgFromRun` to look up the organization that owns a test run
    by its ID.

- `apps/api/internal/shared/middleware/tenant.go`
  - Added `ResolveOrgFromRun` to the `WorkspaceOrgResolver` interface.
  - Added `RunToOrg` helper so route groups can resolve tenants from a `run_id`
    path parameter.

- `apps/api/internal/shared/server/server.go`
  - The `/test-runs/{id}` route group now uses `RunToOrg` instead of
    `ProjectToOrg`. The previous resolver was treating the run ID as a project
    ID, which caused 400/404 errors for run detail, items, and stream requests.

- `apps/api/internal/results/service.go`
  - `UpdateRunStatus` now broadcasts a full `RunProgressEvent` including
    `skipped`, `blocked`, and `progress` so the UI does not lose count state
    when the run status changes.

### Frontend

- `apps/web/app/(dashboard)/[workspace]/test-runs/[id]/page.tsx`
  - Builds the EventSource URL against `NEXT_PUBLIC_API_URL` and passes the JWT
    in `access_token`.
  - Updates the run counts, status, and progress bar live.
  - Updates the matching test run item status when an item-level event is
    received.
  - Closes the stream when a terminal run-level event (`passed`, `failed`,
    `skipped`, `cancelled`) arrives and refreshes the full run/item snapshot.

### API documentation

- `docs/api/openapi/openapi.yaml`
  - `/test-runs/{id}/stream` now documents the `access_token` query parameter
    required by browser EventSource clients.

### Tests

- `apps/api/internal/shared/middleware/auth_test.go`
  - Unit tests for header-based, query-token, and missing-token authentication.

- `apps/api/tests/integration/sse_test.go`
  - Integration tests verifying that the SSE stream rejects unauthenticated
    requests and accepts a JWT via the `access_token` query parameter with the
    correct `text/event-stream` response type.

## Architecture

```text
Browser
  │
  ├─ GET /api/v1/test-runs/{id}              (REST, Bearer header)
  ├─ GET /api/v1/test-runs/{id}/items        (REST, Bearer header)
  └─ GET /api/v1/test-runs/{id}/stream?access_token=...  (SSE)
          │
          ▼
   Auth middleware ──► extracts JWT from header or query
          │
          ▼
   TenantContext (RunToOrg) ──► sets app.tenant_id for RLS
          │
          ▼
   results.StreamRunProgress ──► subscribes to progressHub
          │
          ▼
   progressHub broadcasts events emitted by
   UpdateRunStatus / UpdateItemStatus / automation ingestion
```

## Event payload

```json
{
  "run_id": "uuid",
  "item_id": "uuid-or-null",
  "status": "running",
  "total": 10,
  "passed": 3,
  "failed": 1,
  "skipped": 0,
  "blocked": 0,
  "progress": 0.4
}
```

`item_id` is the nil UUID (`00000000-0000-0000-0000-000000000000`) for
run-level events, and a real item UUID for item-level events.

## Verification commands

```bash
# From the repo root

# 1. Backend unit tests
cd apps/api
go test ./...

# 2. Integration tests (requires TEST_DATABASE_URL or DATABASE_URL)
go test -tags integration ./tests/integration

# 3. Frontend type check
cd apps/web
npm run typecheck

# 4. Manual end-to-end smoke test
#    a. Start the API and web dev servers.
#    b. Log in, create a workspace/project, and add a couple of test cases.
#    c. Go to Dashboard → Test Runs → New Test Run, select the test cases, and
#       create the run.
#    d. On the run detail page click "Start Run".
#    e. Use curl or the UI to update an item status to passed/failed:
#       curl -X PUT http://localhost:8080/api/v1/test-run-items/{item_id} \
#         -H "Authorization: Bearer $TOKEN" \
#         -H "Content-Type: application/json" \
#         -d '{"status":"passed","duration_ms":1000}'
#    f. The counts, progress bar, and item row should update without a page
#       reload.
```

## Known limitations / follow-up work

- `EventSource` uses the JWT in the query string. This is acceptable for the
  current MVP, but a dedicated short-lived SSE token or cookie-based session
  should be considered before exposing the stream over untrusted networks.
- The stream currently relies on in-memory subscriptions (`progressHub`). For
  horizontal deployments, progress events need to be published through a
  shared message bus (e.g. Redis pub/sub or NATS).
- `UpdateItemStatus` does not automatically mark a run terminal when all
  items are finished. A separate run-level status update is still required.
