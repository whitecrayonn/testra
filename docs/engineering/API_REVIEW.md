# API Review v1 — Testra Platform

**Date:** 2026-07-19  
**Scope:** `apps/api/internal/shared/server/server.go`, all module handlers, `docs/api/openapi/openapi.yaml`, `docs/api/API_DESIGN_GUIDELINES.md`, and the shared `pagination` / `idempotency` packages.  
**Goal:** Verify that HTTP endpoints follow the documented REST conventions for status codes, pagination, idempotency, response envelopes, and OpenAPI coverage.

## Verification

```powershell
go build ./...
go test -count=1 ./...
```

Both pass for `apps/api`.

## 1. Endpoint Inventory

`apps/api/internal/shared/server/server.go` registers `91` route declarations under `/api/v1` plus `/.well-known/jwks.json`, `/health`, and the CORS/preflight handling. `docs/api/openapi/openapi.yaml` documents `63` operations. The difference is primarily Phase 3/4 modules that are implemented but still marked as roadmap in the contract.

### Module Coverage

| Module | Implemented Routes | OpenAPI Coverage | Notes |
|--------|-------------------|------------------|-------|
| Auth (register, login, refresh, MFA, password reset, logout) | ✅ | ✅ | |
| Organizations | ✅ | ✅ | |
| Workspaces | ✅ | ✅ | |
| Projects | ✅ | ✅ | |
| API Keys | ✅ | ✅ | |
| Test Management (folders, suites, cases, versions) | ✅ | ✅ | |
| Test Runs & Run Items | ✅ | Mostly | |
| CI Ingestion | ✅ | ✅ | |
| Defects | ✅ | ⚠️ Partial / missing | |
| Notifications (in-app, preferences, channels) | ✅ | ⚠️ Partial / missing | |
| Analytics (dashboards, summary, trends) | ✅ | ❌ Not in OpenAPI | |
| Intelligence (flaky prediction, failure classification, clusters) | ✅ | ❌ Not in OpenAPI | |
| Integration Hub (integrations, events, test) | ✅ | ❌ Not in OpenAPI | |
| Billing (subscription, invoices) | ✅ | ❌ Not in OpenAPI | |

## 2. HTTP Status Codes and Response Envelope

### 2.1 What is working

- All handlers use `apihttp.JSON` / `apihttp.ErrorJSON`, which produces a consistent envelope:
  ```json
  { "data": {}, "meta": {}, "error": { "code": "...", "message": "..." } }
  ```
- `201 Created` is returned for creation endpoints (`POST /organizations`, `POST /workspaces`, `POST /projects`, `POST /test-folders`, `POST /test-suites`, `POST /test-cases`, etc.).
- `200 OK` is used for reads and successful updates.
- `204 No Content` is used for deletes where the handler does not return a body.
- Common errors are mapped in handler `mapError` helpers:
  - `ErrConflict` → `409 Conflict`
  - `ErrNotFound` → `404 Not Found`
  - `ErrInvalidInput` → `400 Bad Request`
  - `ErrForbidden` → `403 Forbidden`
  - Default → `500 Internal Server Error` with a safe generic message.

### 2.2 Gaps

- Not all handlers have a `mapError` helper; some call `apihttp.ErrorJSON` inline with inconsistent code strings (e.g., `INVALID_INPUT` vs `NOT_FOUND`). The codebase is mostly consistent but a shared `HTTPErrorMapper` would reduce drift.
- Some endpoints return `200 OK` for mutating operations that trigger side effects (e.g., `POST /integrations/{id}/test`, `POST /intelligence/predict-flaky`) because they are synchronous commands. This is acceptable, but the contract should document the returned representation.
- `DELETE` endpoints that return `204` do not return an error envelope on failure in some handlers; confirm that `mapError` is invoked before `WriteHeader`.

## 3. Pagination

### 3.1 What is working

The `pagination` package is clean:
- `ParseParams` reads `?cursor` and `?limit`.
- Default `limit` is `20`, max is `100`.
- `EncodeCursor` / `DecodeCursor` use base64-URL-encoded JSON containing the last ID.

Paginated endpoints (cursor-based, returning `data` + `meta`):
- `GET /organizations`
- `GET /workspaces`
- `GET /projects`
- `GET /test-cases`
- `GET /test-cases/search`
- `GET /test-runs`
- `GET /notifications`

### 3.2 Gaps

The following list endpoints are still unpaginated and return the full result set as a top-level array. This is acceptable for low-volume MVP data but must be addressed before large workspaces:

| Endpoint | Handler | Current behavior |
|----------|---------|------------------|
| `GET /test-folders` | `testmanagement.ListFolders` | Returns `[]folderResponse` (no `meta`) |
| `GET /test-suites` | `testmanagement.ListSuites` | Returns `[]suiteResponse` (no `meta`) |
| `GET /test-cases/{id}/versions` | `testmanagement.ListVersions` | Returns `[]versionResponse` (no `meta`) |
| `GET /test-runs/{id}/items` | `results.ListItems` | Returns `[]itemResponse` (no `meta`) |
| `GET /notification-channels` | `notification.ListChannels` | Returns `[]channelResponse` (no `meta`) |
| `GET /billing/invoices` | `billing.ListInvoices` | Returns `[]invoiceResponse` (no `meta`) |

**Recommendation:** Extend `pagination.ParseParams` usage to these handlers, update service/repository signatures to accept `cursor` and `limit`, and return the envelope `{"data": [...], "meta": {...}}`.

## 4. Idempotency

### 4.1 What is working

- `shared/middleware/idempotency.go` implements an `Idempotency-Key` middleware that:
  - Requires an `Idempotency-Key` header.
  - Hashes the key and computes a fingerprint of the request body.
  - Looks up a prior record scoped to `(workspace_id, operation, key_hash)`.
  - Replays a stored `2xx`/`4xx` response when the key and fingerprint match.
  - Returns `409 Conflict` if the same key is reused with a different body.
  - Stores responses in PostgreSQL with a configurable TTL (default 24 hours).
- `POST /ingest` uses this middleware, which is correct for CI ingestion that may retry.

### 4.2 Gaps

- The middleware extracts `workspace_id` from the request body (`extractWorkspaceID`) only. It cannot be applied to endpoints that take `workspace_id` as a query parameter or path parameter without modification.
- Only `/ingest` is protected. Per `API_DESIGN_GUIDELINES.md`, side-effecting commands such as `POST /test-runs`, `POST /test-cases`, `PUT /billing/subscription`, `POST /integrations/dispatch`, and `POST /notification-channels` should also accept `Idempotency-Key`.
- The middleware currently stores records for `200 <= status < 500`. It should not store `5xx` responses because they are transient; retries should be allowed to reach the server again.

**Recommendation:**
1. Generalize the middleware to obtain `workspace_id` from request context (set by `TenantContext`) instead of parsing the body, then apply it to all mutating `/api/v1` routes inside the authenticated group.
2. Change the persistence condition to `status >= 200 && status < 500 && status != 500` (i.e., store `2xx`, `3xx`, `4xx`; skip `5xx`).

## 5. OpenAPI Contract Drift

- `docs/api/openapi/openapi.yaml` is the contract authority per `API_DESIGN_GUIDELINES.md`, but it is behind implementation.
- Endpoints for **Defects, Analytics, Intelligence, Integration Hub, and Billing** are implemented in Go but not documented in the contract.
- `ROUTES.md` lists backend routes but does not enumerate all of them; it also still describes the frontend workspace slug routing as UUID-based in `localStorage`.

**Recommendation:**
1. Add the missing modules to `openapi.yaml` in a single pass.
2. Introduce an OpenAPI validation step in CI (`swagger-codegen validate` or `redocly lint`) so the contract cannot drift again.

## 6. Input Validation and Error Handling

### 6.1 Working

- Handlers validate UUIDs and required query/body fields at the boundary.
- `json.NewDecoder(r.Body).Decode(&req)` is used consistently.
- Strongly typed request structs keep validation close to the HTTP boundary.

### 6.2 Gaps

- Most handlers do **not** use `DisallowUnknownFields()` on the decoder. This allows clients to send extra fields that are silently ignored, contrary to `API_DESIGN_GUIDELINES.md` §6.
  ```go
  dec := json.NewDecoder(r.Body)
  dec.DisallowUnknownFields()
  if err := dec.Decode(&req); err != nil { ... }
  ```
- Pagination `limit` clamping happens inside `pagination.ParseParams` (max 100) but some list handlers call repository methods with a raw `limit` parsed elsewhere. Audit every handler to ensure it uses the package.
- Some endpoints rely on `mapError` `default` branch for `ErrUnauthorized`/`ErrUnauthenticated` rather than mapping them to `401` explicitly.

## 7. Security and CORS

- `corsMiddleware` now sets `Vary: Origin`, `Access-Control-Allow-Headers` (including `Idempotency-Key`, `X-API-Key`, `X-CSRF-Token`), and `Access-Control-Max-Age: 600`. (See `SECURITY_REVIEW_v2.md` for details.)
- API key authentication is mounted for `/ingest`; other endpoints still use JWT. The contract currently documents JWT and API-key support should be expanded once enabled.

## 8. Recommended Next Steps

1. **Paginate** `GET /test-folders`, `GET /test-suites`, `GET /test-cases/{id}/versions`, `GET /test-runs/{id}/items`, `GET /notification-channels`, and `GET /billing/invoices`.
2. **Apply `Idempotency-Key`** middleware to all authenticated `POST`, `PUT`, `PATCH`, and `DELETE` endpoints. Refactor it to read `workspace_id` from context rather than the request body.
3. **Update OpenAPI** to cover Defects, Analytics, Intelligence, Integration Hub, and Billing.
4. **Enable `DisallowUnknownFields()`** on all JSON decoders or document where unknown fields are intentionally tolerated.
5. **Add a shared `HTTPErrorMapper`** in `shared/http` that converts every `sharederrors` sentinel to a stable status/code pair; remove per-handler `mapError` duplication.

## 9. Conclusion

The API already follows most of the documented conventions: versioning, resource naming, response envelopes, cursor pagination for high-volume lists, and status-code mapping. The largest gaps are missing OpenAPI coverage for newer modules and incomplete idempotency/pagination adoption. These are not security blockers but are prerequisites for SDK generation and a stable public API.
