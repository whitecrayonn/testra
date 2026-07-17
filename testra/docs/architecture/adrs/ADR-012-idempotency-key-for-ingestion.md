# ADR-012: Idempotency-Key for Ingestion Endpoint

**Status:** Accepted  
**Date:** July 2026

## Context

ADR-006 states: "Use Idempotency-Key for create/command endpoints with external side effects, ingestion endpoints, exports, webhooks." ENGINEERING_STANDARDS §3.1 reiterates: "Idempotency-Key required for side-effecting commands, ingestion."

The current `POST /ingest` endpoint does not implement Idempotency-Key. Duplicate ingestion requests (from CI retries, network issues, or client bugs) create duplicate test runs.

## Decision

Implement Idempotency-Key on the `POST /ingest` endpoint per ADR-006.

### Implementation

1. **Header:** `Idempotency-Key` (UUID v4 recommended, any string accepted)
2. **Request body:** `IngestRequest` now includes a `payload` field carrying the raw JUnit XML, Playwright JSON, or Cypress JSON test results.
3. **Storage:** PostgreSQL `idempotency_records` table
   - `id` (UUID PK)
   - `key` (VARCHAR 255, indexed)
   - `workspace_id` (UUID FK)
   - `operation` (VARCHAR 50, e.g. "ingest")
   - `request_fingerprint` (VARCHAR 64, SHA-256 of normalized request body)
   - `response_body` (JSONB)
   - `status_code` (INTEGER)
   - `created_at` (TIMESTAMPTZ)
   - `expires_at` (TIMESTAMPTZ, 24 hours after creation)
4. **Logic:**
   - If `Idempotency-Key` header is missing, reject with 400 (required for ingestion)
   - If key exists with same fingerprint, return stored response
   - If key exists with different fingerprint, reject with 409 Conflict
   - If key does not exist, process request, store response, return
5. **Scope:** Key is scoped to `(workspace_id, operation, key)` — same key in different workspaces is independent
6. **Retention:** 24 hours per ADR-006; configurable via `IDEMPOTENCY_KEY_TTL_MINUTES` (default 1440).

### Migration

A new migration `000017_add_idempotency_records.up.sql` will create the `idempotency_records` table with appropriate indexes and RLS.

### Middleware

Implement as reusable middleware `IdempotencyKey` in `shared/middleware/` that wraps any POST handler. The middleware:
1. Reads `Idempotency-Key` header
2. Checks the idempotency record store
3. If replay, returns stored response
4. If new, wraps `http.ResponseWriter` to capture response, stores it, then returns

### Priority

**HIGH** — This is a correctness issue. Duplicate runs from CI retries corrupt analytics and user dashboards. Must be implemented before Phase 3 DoD sign-off.

## Consequences

- **Positive:** Duplicate ingestion requests are safely deduplicated; CI retry storms do not create duplicate runs.
- **Negative:** Additional database write per ingestion request; 24-hour storage cost; adds middleware complexity.
- **Mitigation:** The idempotency record is a single indexed lookup + insert. The 24-hour TTL with automatic cleanup keeps storage bounded.
