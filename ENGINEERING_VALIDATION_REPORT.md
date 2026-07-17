# Testra Engineering Validation Report

**Date:** 2025-11-25
**Scope:** Go API backend (`apps/api`), OpenAPI contract, Next.js/TypeScript frontend (`apps/web`), PostgreSQL schema/migrations.
**Objective:** Audit cross-layer consistency of the API contract, database schema, security, performance, and overall code quality.

---

## 1. Executive Summary

The Testra backend is built on a sensible layered architecture (domain → service → repository → handler) with PostgreSQL, Chi, JWT, RBAC, and row-level security (RLS). Unit tests pass and `go vet` is clean.

The **highest-risk finding** is an inconsistent JSON response envelope that causes runtime contract drift between the OpenAPI spec, the Go handlers, and the TypeScript frontend. Several handlers double-wrap the `data` field, which breaks frontend callers that expect a single envelope. Other issues include missing API-key authentication middleware, a broken cursor for one paginated endpoint, JWT stored in `localStorage`, and several ignored `go fmt` failures.

| Area | Rating | Notes |
| --- | --- | --- |
| API Contract | **Red** | Envelope double-wraps, OpenAPI/frontend mismatches, missing API-key auth. |
| Database Consistency | **Yellow-Green** | Migrations and domains mostly align; cursor decode bug and manual array formatting are risks. |
| Security | **Yellow** | RLS/JWT good, but token storage, query-param token fallback, and missing API-key auth need attention. |
| Performance | **Yellow** | Connection reuse and pagination present; no batch inserts and rate-limiter not wired. |
| Frontend | **Yellow** | `apiFetch` assumptions are coupled to backend envelope bugs. |
| Testing & Quality | **Yellow** | Unit tests pass; `gofmt` has failures, race tests not run. |

---

## 2. Scope and Methodology

Files reviewed include:

- `docs/api/openapi/openapi.yaml`
- `apps/api/internal/shared/server/server.go`
- Handlers: `identity`, `notification`, `testmanagement`, `results`, `automationhub`, `apikeys`
- Services and repositories for the modules above
- Middleware: `auth.go`, `rbac.go`, `idempotency.go`, `tenant.go`
- Frontend: `apps/web/lib/api.ts`, `features/testmanagement/api.ts`, `features/results/api.ts`, `features/notifications/api.ts`, `types/*.ts`
- Migrations: `000012_*`, `000014_*`, `000015_*`, `000017_*`, `000018_*`

Commands run (`apps/api`):

- `go test -count=1 ./...` → **passed**
- `go vet ./...` → **passed**
- `gofmt -l .` → **reported unformatted files** (see section 9)
- `go test -race -count=1 ./...` → **could not run** (`-race requires cgo`)

---

## 3. Architecture and Module Structure

### Strengths

- Clear separation of concerns: domain, service, repository, and handler packages per feature.
- Shared packages for cross-cutting concerns (`shared/middleware`, `shared/db`, `shared/http`, `shared/pagination`, `shared/errors`).
- `db.DB` wrapper supports per-request transactions and per-request `*sql.Conn` so RLS session variables are scoped to a connection.
- Tenant context middleware acquires a dedicated connection, sets `app.tenant_id`, and resets it before returning the connection to the pool.

### Findings

- Several feature modules are placeholders with no implementation or tests: `analytics`, `billing`, `defects`, `integrationhub`, `intelligence`.
- `rateLimitCfg` is created in `server.go` but immediately discarded with `_ = rateLimitCfg` at line 469, so **rate limiting is not wired**.
- `_ = tenantResolver` at line 470 is another indicator of unfinished wiring.
- `server.go` registers routes in a single large function; as the module count grows this will become hard to maintain.

---

## 4. API Contract Audit

This is the most important section. The source of truth for response formatting is `apps/api/internal/shared/http/response.go`:

```go
func JSON(w http.ResponseWriter, status int, data any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(Envelope{Data: data})
}
```

Every successful response is therefore wrapped as `{"data": <payload>}`. The bug is that many handlers already build a `map[string]any{"data": x, "meta": y}` and pass that as the payload, producing a **double envelope**:

```json
{"data": {"data": [...], "meta": {...}}}
```

### 4.1 Backend Response Envelope Bugs

**Handlers that use double envelopes:**

- `apps/api/internal/results/handler.go`
  - `CreateRun` (line 192): `apihttp.JSON(..., map[string]any{"data": mapRunResponse(run)})`
  - `GetRun` (line 208): same pattern
  - `UpdateRunStatus` (line 273): same pattern
  - `UpdateItemStatus` (line 338): same pattern
  - `ListRuns` (lines 244-247): `map[string]any{"data": resp, "meta": meta}`
  - `ListItems` (line 309): `map[string]any{"data": resp}`
- `apps/api/internal/notification/handler.go`
  - `List` (lines 133-136): `map[string]any{"data": resp, "meta": meta}`
  - `Create` (line 258): `map[string]any{"data": mapNotificationResponse(n)}`
  - `GetPreferences` (line 280): `map[string]any{"data": mapPreferencesResponse(p)}`
  - `UpdatePreferences` (line 323): same
  - `ListChannels` (line 344): `map[string]any{"data": resp}`
  - `CreateChannel` (line 391): `map[string]any{"data": mapChannelResponse(ch)}`
  - `UpdateChannel` (line 423): same
- `apps/api/internal/apikeys/handler.go`
  - `List` (lines 130-133): `map[string]any{"data": resp, "meta": meta}`

**Handlers that return a correct single envelope:**

- `apps/api/internal/testmanagement/handler.go` (create/update/get for folders/suites/cases, list versions)
- `apps/api/internal/identity/handler.go` (`Register`, `Login`, `Me`, `SetupMFA`)
- `apps/api/internal/automationhub/handler.go` (`Ingest` returns a struct directly)

### 4.2 Frontend Impact

`apps/web/lib/api.ts` returns `body.data`:

```ts
const body = await res.json();
return body.data as T;
```

Consequences:

- **Test management endpoints work:** backend returns `{"data": entity/array}` and `body.data` is the entity/array.
- **Paginated endpoints accidentally work when double-wrapped:** `ListRuns`, `ListItems`, `ListNotifications`, `ListChannels`, `ListAPIKeys` return `{"data": {"data": [...], "meta": {...}}}`. `apiFetch` returns the inner `{"data": [...], "meta": {...}}`, which the frontend `PaginatedResponse` types expect.
- **Single-resource results endpoints break the frontend:** `getTestRun`, `createTestRun`, `updateTestRunStatus`, `updateTestRunItemStatus` return `{"data": {"data": <entity>}}`. The frontend types expect `<entity>` directly, so callers receive `{data: entity}` instead of `entity`.
- `results/api.ts` `listTestRunItems` has a defensive `Array.isArray` check, but the backend always returns an object, so it falls through to `result.data`. This is fragile.
- `notifications/api.ts` frontend types (`Promise<{ data: NotificationPreferences }>`, `Promise<{ data: NotificationChannel }>`) were written to match the double envelope, so they work today but are inconsistent with test management types.

### 4.3 OpenAPI vs. Backend Mismatches

`docs/api/openapi/openapi.yaml` describes a single envelope for most endpoints (`allOf` with `#/components/schemas/Envelope`), but there are inconsistencies:

- **Test management endpoints** (`/test-folders`, `/test-suites`, `/test-cases`, `/test-cases/{id}/versions`) declare responses as the raw entity or raw array, **not** wrapped in the envelope. This contradicts the actual backend.
- **Notification endpoints** that declare `Envelope` with both `data` and `meta` expect `{"data": [...], "meta": {...}}`, but the backend double-wraps.
- **Notification `markNotificationRead`, `deleteNotification`, `deleteNotificationChannel`** declare raw `{"status": "updated"}` / `{"status": "deleted"}`. The backend wraps them to `{"data": {"status": "..."}}`.
- **API keys `GET /api-keys`** declares `Envelope` with `data` array but no `meta`; the backend returns double envelope with `meta`.

### 4.4 Pagination Cursor Bug

`pagination.EncodeCursor` base64-encodes `{"id": "<uuid>"}`.

- `testmanagement/repository.go` `ListCases` uses the cursor string directly in `WHERE id < $3` without decoding it. The second page of `/test-cases` will receive an invalid UUID string and fail.
- `results/repository.go` `ListRuns` correctly calls `pagination.DecodeCursor(cursor)` first.
- `testmanagement/handler.go` `SearchCases` uses `pagination.DecodeSearchCursor`. The plain `ListCases` path is the one that is broken.

### 4.5 Missing API-Key Authentication

`apps/api/internal/apikeys/service.go` has a `Validate` function, but `server.go` never mounts an API-key authentication middleware. The `/ingest` endpoint (and every other route) relies on the JWT `Auth` middleware. CI/automation clients that are expected to use API keys cannot authenticate with them.

---

## 5. Database Consistency Audit

### Strengths

- Migrations `000012` (test management), `000014` (test management RLS), `000015` (test runs), `000017` (idempotency), and `000018` (notifications) cover the reviewed tables, FKs, indexes, and RLS policies.
- Domain structs in `testmanagement/domain.go`, `results/domain.go`, and `notification/domain.go` align with the table columns.
- `app.tenant_id` is set per connection and reset on release, and RLS policies consistently use `current_setting('app.tenant_id', true)::uuid`.

### Findings

- `testmanagement/repository.go` and `apikeys/repository.go` implement manual PostgreSQL array formatting (`pqArray`) and parsing (`parseTags`). This does not escape commas or quotes in tag/key strings, which will corrupt array values or cause parse errors. Use `pq.Array`/`pgtype` instead.
- `results/service.go` `CreateRun` inserts the `TestRun` and then inserts each `TestRunItem` one at a time. There is no transaction wrapper, so a failure after the run is created leaves a partial run. This should run inside `repo.RunInTx`.
- `testmanagement/service.go` `UpdateCase` wraps the version snapshot and update in `repo.RunInTx`, which is correct.
- `results/repository.go` `DeleteRun` deletes the run; cascades handle `test_run_items` because of `ON DELETE CASCADE`.
- `test_case_versions` `created_at` column has `DEFAULT NOW()` and the code sets it explicitly, so it is safe.

---

## 6. Security Audit

### Authentication and Authorization

- JWT is signed with HS256 (`apps/api/internal/shared/jwt/jwt.go`) and validated in `shared/middleware/auth.go`.
- `Auth` middleware reads the token from the `Authorization: Bearer <token>` header **or** the `access_token` / `token` query parameters. The query-param fallback is applied to **all** routes, not just SSE. This leaks tokens into access logs and browser history and should be restricted to the SSE endpoint.
- RBAC middleware loads permissions once per request and stores them in context. The permission strings (`tests:read`, `runs:ingest`, etc.) match the route registrations.

### API Keys

- `apikeys/service.go` creates keys with `crypto/rand`, hashes them with SHA-256, and stores only the hash. It validates expiry and revocation.
- **Missing:** no middleware consumes API keys, so the `/ingest` automation endpoint cannot be used with an API key.
- **Missing:** API key scopes are stored but not enforced. RBAC uses user permissions, not key scopes.

### Frontend Token Storage

- `apps/web/lib/api.ts` reads the token from `localStorage.getItem("testra_token")`. Storing JWTs in `localStorage` makes them vulnerable to XSS exfiltration. Consider httpOnly cookies or at least `sessionStorage` with a short expiry and refresh-token rotation.

### Row-Level Security

- RLS is enabled on tenant-scoped tables and policies use `current_setting('app.tenant_id', true)::uuid`.
- `tenant.go` acquires a dedicated `*sql.Conn`, sets `app.tenant_id`, and resets it with `RESET app.tenant_id` before closing the connection. This prevents cross-request tenant leakage in a pooled environment.

### Other Risks

- `notification/service.go` `dispatchHTTP` POSTs to user-provided `url` from channel `config["url"]` without an allowlist or SSRF validation. An attacker with `notification_channels:create` permission could target internal services.
- `automationhub/service.go` parses XML/JSON from external CI systems. It was not fully audited for XML external entity (XXE) or malicious payload handling; ensure the XML parser is configured securely.

---

## 7. Performance and Concurrency Audit

- `db.DB` reuses a per-request `*sql.Conn`, which avoids repeatedly checking out connections and keeps the RLS session variable stable.
- Pagination is cursor-based for `test_cases` (search), `test_runs`, and `api_keys`, which is good for large tables.
- `gofmt` failures (section 9) are not a runtime issue but indicate code-quality drift.
- `go test -race` could not be run because CGO is disabled. Races in the SSE `progressHub` or the `db.DB` wrapper cannot be confirmed without enabling `CGO_ENABLED=1`.
- No bulk insert for test run items; large test runs will create many round-trips. Consider `COPY` or a multi-value `INSERT`.
- The local rate limiter is instantiated but never used, so there is no request throttling.

---

## 8. Frontend Audit

- `apps/web/lib/api.ts` is a thin wrapper around `fetch`. It always returns `body.data` and does not handle token refresh or automatic retry.
- `apps/web/features/testmanagement/api.ts` maps correctly to single-envelope backend responses.
- `apps/web/features/results/api.ts` expects `TestRun` / `TestRunItem` directly for `getTestRun`, `createTestRun`, `updateTestRunStatus`, and `updateTestRunItemStatus`, but the backend double-wraps. These calls will receive `{data: entity}` and are likely to fail at runtime.
- `apps/web/features/notifications/api.ts` types are written around the double envelope (e.g., `Promise<{ data: NotificationChannel }>`), so they work but are inconsistent with other features.
- `getWorkspaceId()` reads from `localStorage` on every call; minor, but repeated calls could be memoized.

---

## 9. Testing and Static Analysis

### Results

- `go test -count=1 ./...` passed for all packages with tests.
- `go vet ./...` passed with no output.
- `gofmt -l .` reported the following files need formatting:
  - `cmd/migrator/main.go`
  - `cmd/worker/main.go`
  - `internal/audit/domain.go`
  - `internal/notification/domain.go`
  - `internal/notification/service.go`
  - `internal/testmanagement/domain.go`
  - `tests/integration/sse_test.go`
- `go test -race -count=1 ./...` failed because `-race requires cgo`; rerun with `CGO_ENABLED=1`.

### Test Coverage Gaps

- Unit tests exist for `identity`, `results`, `notification`, `testmanagement`, `project`, and `middleware`, but they mostly use in-memory fake repositories.
- No integration tests exercising real HTTP routes, RLS policies, or the OpenAPI contract.
- No frontend tests or type-check runs were performed.
- Several packages (`analytics`, `billing`, `defects`, `integrationhub`, `intelligence`, `rbac`, `apikeys`) have `[no test files]`.

---

## 10. Code Quality and Technical Debt

- Repeated `DBTX` interface definitions in each repository package (`results/repository.go`, `testmanagement/repository.go`, etc.) could be centralized from `shared/db`.
- Manual PostgreSQL array string formatting in repositories should be replaced with `pq.Array` or the equivalent `pgtype` driver helper.
- Several errors are silently ignored (`_ = smtp.SendMail`, `_ = resp.Body.Close()`, `_ = json.Marshal(...)`, `_ = idempotencyStore.Save(...)`). This makes debugging failures difficult.
- Handler response patterns are inconsistent (some pass structs, some pass `map[string]any`). A shared response builder (`JSONWithMeta`, `JSONStatus`) would eliminate the double-envelope bugs.
- `_ = rateLimitCfg` and `_ = tenantResolver` in `server.go` show unfinished wiring.
- Placeholder modules should either be implemented or removed to reduce confusion.

---

## 11. Recommendations (Prioritized)

### Critical

1. **Fix the response envelope contract.**
   - Decide on a single contract: `{"data": <payload>, "meta": <meta>, "error": <error>}`.
   - Change `apihttp.JSON` to accept optional metadata, or add `apihttp.JSONWithMeta`.
   - Remove the `map[string]any{"data": ..., "meta": ...}` double wrapping from `results`, `notification`, and `apikeys` handlers.
   - Update `apps/web/lib/api.ts` to return `body` directly for single-envelope payloads (or keep `body.data` if `Envelope` becomes the top level).
2. **Implement and wire API-key authentication middleware**, and apply it to `/ingest` and other automation-facing routes. Enforce key scopes.
3. **Fix `/test-cases` pagination cursor decoding** in `testmanagement/repository.go` `ListCases`.
4. **Move JWT storage from `localStorage` to httpOnly cookies**, or implement short-lived access tokens with refresh-token rotation and XSS mitigations.
5. **Restrict query-parameter token extraction** to SSE/stream endpoints only.

### High

6. Run `gofmt -w .` and add a CI `gofmt` / `go vet` check.
7. Replace manual PostgreSQL array formatting with `pq.Array`/`pgtype`.
8. Run `CGO_ENABLED=1 go test -race ./...` and fix any races.
9. Wrap `results/service.go` `CreateRun` and item insertions in a transaction.
10. Add SSRF protection for notification webhook URLs (allowlist, DNS validation, private-IP block).
11. Wire the rate limiter to public/auth routes.
12. Add integration/contract tests that validate each handler response against the OpenAPI spec.

### Medium

13. Add refresh-token rotation and expiration to the identity service.
14. Centralize the `DBTX` interface in `shared/db`.
15. Remove or implement placeholder modules (`analytics`, `billing`, `defects`, `integrationhub`, `intelligence`).
16. Refactor `server.go` route registration into per-module route files or smaller route builders.

---

## 12. Appendix: Commands and Evidence

```text
$ go test -count=1 ./...
ok  	github.com/testra/testra/apps/api/internal/automationhub	0.619s
ok  	github.com/testra/testra/apps/api/internal/identity	1.618s
ok  	github.com/testra/testra/apps/api/internal/notification	0.569s
ok  	github.com/testra/testra/apps/api/internal/project	0.532s
ok  	github.com/testra/testra/apps/api/internal/results	0.552s
ok  	github.com/testra/testra/apps/api/internal/shared/middleware	0.636s
ok  	github.com/testra/testra/apps/api/internal/testmanagement	0.566s

$ go vet ./...
(no output)

$ gofmt -l .
cmd\migrator\main.go
cmd\worker\main.go
internal\audit\domain.go
internal\notification\domain.go
internal\notification\service.go
internal\testmanagement\domain.go
tests\integration\sse_test.go

$ go test -race -count=1 ./...
-error: -race requires cgo; enable cgo by setting CGO_ENABLED=1
```

### Key File References

- Response envelope helper: `apps/api/internal/shared/http/response.go`
- Double-wrap examples: `apps/api/internal/results/handler.go` (lines 192, 208, 244-247, 273, 309, 338), `apps/api/internal/notification/handler.go` (lines 133-136, 258, 280, 323, 344, 391, 423), `apps/api/internal/apikeys/handler.go` (lines 130-133)
- Correct single-wrap examples: `apps/api/internal/testmanagement/handler.go` (lines 212, 228, 264, 290, 349, 365, 401, 431, 531, 547, 707, 742)
- Frontend wrapper: `apps/web/lib/api.ts`
- Pagination encode/decode: `apps/api/internal/shared/pagination/pagination.go`
- Tenant/RLS connection handling: `apps/api/internal/shared/middleware/tenant.go`
- API-key service: `apps/api/internal/apikeys/service.go`
- OpenAPI contract: `docs/api/openapi/openapi.yaml`
