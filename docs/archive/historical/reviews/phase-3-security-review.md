# Phase 3 Security Review — Execution & Results Ingestion

**Date:** 2026-07-16  
**Reviewer:** Cascade (AI Engineering Assistant)  
**Scope:** ADR-012 Idempotency-Key middleware, `POST /api/v1/ingest`, supporting infrastructure (migrations, config, OpenAPI), and related integration tests.  
**Status:** PASS WITH REMEDIATIONS APPLIED

---

## Summary

The idempotency implementation for `POST /api/v1/ingest` correctly enforces header presence, tenant scoping, request fingerprinting, and replay semantics. All identified high-severity concerns were remediated before the review was finalized. The remaining observations are low-severity operational or architectural recommendations.

---

## Positive Findings

- **Idempotency-Key presence is enforced.** Requests without the header receive HTTP 400 with code `IDEMPOTENCY_KEY_REQUIRED`.
- **Keys and fingerprints are SHA-256 digests.** Raw keys are never persisted; only `SHA-256(key)` and `SHA-256(compacted body)` are stored.
- **Conflict detection works.** Reusing a key with a different request fingerprint returns HTTP 409 with code `IDEMPOTENCY_KEY_CONFLICT`.
- **Replay returns the stored response body and status.** Verified by integration tests: duplicate uploads return the same `run_id`, response body, and HTTP 201.
- **Tenant scoping is enforced.** Records are partitioned by `workspace_id`; RLS on `idempotency_records` restricts reads/updates to the caller's tenant.
- **No raw source code retention.** The request body is fingerprinted, not stored. The response body is the result summary envelope, not the uploaded test artifact.
- **TTL and cleanup are configurable.** `IDEMPOTENCY_KEY_TTL_MINUTES` defaults to 1440 (24 hours); `DeleteExpired` supports scheduled purging.
- **RBAC is enforced before idempotency.** The `runs:ingest` permission gate runs before the middleware, so unauthorized requests never create records.
- **Audit logging now captures retried attempts.** `AuditLog` middleware is ordered before `IdempotencyKey` so replayed requests are still recorded in the audit trail.
- **OpenAPI contract updated.** `POST /ingest` documents the required `Idempotency-Key` header and the 409 Conflict response.

---

## Remediated Issues

### ISSUE SEC-1 (HIGH): Double-wrapped response envelope

**File:** `apps/api/internal/automationhub/handler.go`  
**Finding:** The handler wrapped the result inside `map[string]any{"data": ...}` and then called `apihttp.JSON`, which already produces an envelope. This caused `{"data": {"data": ...}}` and broke response parsing for clients.  
**Fix:** Return `ingestResponse` directly to `apihttp.JSON`.

### ISSUE SEC-2 (HIGH): Idempotency replay status was always HTTP 200

**File:** `apps/api/internal/shared/middleware/idempotency.go`  
**Finding:** `responseRecorder` was initialized with `statusCode: http.StatusOK` (200), so even successful `201 Created` ingestions were stored and replayed as `200 OK`.  
**Fix:** Initialize `responseRecorder` with status code `0` and rely on the downstream handler's `WriteHeader` call.

### ISSUE SEC-3 (HIGH): Execution-permissions migration conflicted with RBAC seed

**File:** `apps/api/migrations/000016_add_execution_permissions.up.sql`  
**Finding:** Migration 000016 attempted to re-insert `runs:create` and `runs:read` permissions (already seeded in 000006 with different IDs). The `ON CONFLICT (name) DO NOTHING` clause skipped the insert, but the role-permission assignments referenced the new IDs, violating the foreign-key constraint.  
**Fix:** Migration 000016 now only inserts the new permissions (`runs:update`, `runs:delete`, `runs:ingest`) and assigns them by their fixed IDs.

### ISSUE SEC-4 (MEDIUM): No `payload` field in request contract

**File:** `docs/api/openapi/openapi.yaml`, `apps/api/internal/automationhub/handler.go`  
**Finding:** The OpenAPI spec and the handler did not expose a `payload` field; the handler read the raw request body as the test artifact, which conflicts with JSON metadata envelopes.  
**Fix:** Added `payload` to `IngestRequest`, made it required, and updated the handler to ingest `[]byte(meta.Payload)`.

### ISSUE SEC-5 (MEDIUM): Idempotency records could be created without audit trail

**File:** `apps/api/internal/shared/server/server.go`  
**Finding:** `IdempotencyKey` was ordered before `AuditLog`, so a replayed request short-circuited before audit logging.  
**Fix:** Reordered the middleware chain so `AuditLog` executes first.

---

## Observations and Recommendations

| ID | Observation | Severity | Recommendation |
|---|---|---|---|
| OBS-1 | `Idempotency-Key` accepts any string; recommended format is UUID but not validated. | LOW | Add optional format validation or document that any opaque string is accepted. |
| OBS-2 | Failed ingestion responses (5xx, 4xx) are not stored, so a retry after a transient 500 could create duplicate test runs. | LOW | For true exactly-once semantics, store 4xx/5xx responses or require clients to use a new key after failure. ADR-012 intentionally scopes storage to successful responses for MVP. |
| OBS-3 | The `idempotency_records` table stores the entire JSON response body (result summary). Although this is the intended contract, large result summaries could consume more storage than fingerprint-only schemes. | LOW | Consider gzip-compressing `response_body` if result payloads grow. |
| OBS-4 | `request_fingerprint` is computed on the compacted JSON body; field ordering matters. Clients must send canonical JSON or use the same serialization for retries. | LOW | Document in API spec that requests must be byte-identical for replay. |
| OBS-5 | Integration tests require a running PostgreSQL instance with `TEST_DATABASE_URL` or `DATABASE_URL` set. | LOW | Document in `docs/engineering/integrations.md` and optionally add a `docker-compose` test profile. |

---

## Verification

- `go build .\...` — PASS
- `go vet .\...` — PASS
- `go test -count=1 .\...` — PASS
- `go test -tags=integration -count=1 .\tests\integration` — PASS
- `openapi.yaml` validates the required `Idempotency-Key` header and 409 response.

---

## Conclusion

The idempotency implementation is secure for Phase 3 MVP. All high-severity issues found during implementation were fixed. Tenant isolation, RBAC, audit logging, and replay semantics are functioning as intended. Address OBS-1 through OBS-5 as part of Phase 4 hardening.
