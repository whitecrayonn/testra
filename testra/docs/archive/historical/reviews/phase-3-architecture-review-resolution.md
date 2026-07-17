# Phase 3 Architecture Review — Resolution

**Date:** 2026-07-16  
**Phase:** 3 — Execution & Results  
**Status:** PASS — All Phase 3 architecture conditions resolved

---

## Summary

This document resolves the three architecture deviations identified in `docs/engineering/reviews/phase-3-architecture-review.md`. ADR-010, ADR-011, and ADR-012 have been accepted and implemented; integration tests have been written and pass; the ingestion endpoint now enforces `Idempotency-Key` per ADR-006. The `metadata` JSONB column is constrained by the allowlist specified in ADR-010.

---

## Deviations Resolved

### DEVIATION-1: Results Stored in PostgreSQL Instead of ClickHouse

**Resolution:** ADR-010 accepted — `PostgreSQL for Phase 3 Results (Deferred ClickHouse)`.  
**Evidence:**
- Migrations `000015_add_test_runs` and `000016_add_execution_permissions` create `test_runs` and `test_run_items` in PostgreSQL.
- The `results.Repository` interface remains storage-agnostic; a `ClickHouseRepository` can be substituted later without changing service/handler code.
- `RunInTx` is implemented by `SQLRepository`; a ClickHouse implementation can no-op it for append-only analytical writes.

### DEVIATION-2: Synchronous Ingestion Instead of Queue-Based

**Resolution:** ADR-011 accepted — `Synchronous Ingestion for MVP (Deferred Queue-Based Pipeline)`.  
**Evidence:**
- `automationhub.Service.Ingest` is invoked synchronously by `POST /api/v1/ingest`.
- The service is stateless; moving it to an Asynq worker in the future requires only wrapping `Service.Ingest` in a task handler and returning `202 Accepted` from the API.
- Current volume fits within the synchronous HTTP timeout for batches up to ~500 test cases.

### DEVIATION-3: Missing Idempotency-Key on Ingestion Endpoint

**Resolution:** ADR-012 accepted and implemented — `Idempotency-Key for Ingestion Endpoint`.  
**Evidence:**
- Middleware `sharedmiddleware.IdempotencyKey` enforces header presence, stores request fingerprint, and replays stored responses.
- `POST /api/v1/ingest` returns `400` for missing header, `409` for key reuse with different payload, and `201` with replay on duplicate identical requests.
- Migration `000017_add_idempotency_records` creates the `idempotency_records` table with RLS and indexes.
- Integration tests cover all idempotency scenarios.

### ADDITIONAL CONDITION: Metadata JSONB Allowlist

**Resolution:** ADR-010 restricts `test_runs.metadata` to an explicit allowlist: `format`, `ci_build_id`, `ci_branch`, `ci_commit`.  
**Evidence:**
- `automationhub/service.go` only writes `{"format": "..."}` into metadata.
- Tests use only allowed keys.

---

## Integration Tests

The ingestion pipeline integration tests are complete and passing:

- `TestIngestJUnit`
- `TestIngestPlaywright`
- `TestIngestCypress`
- `TestIngestDuplicateUpload`
- `TestIngestDuplicateKeyDifferentPayload`
- `TestIngestMissingKey`
- `TestIngestInvalidPayload`
- `TestIngestUnsupportedFormat`
- `TestIngestUnauthorized`
- `TestIngestTenantIsolation`
- `TestIngestInsufficientPermission`

Run them with:

```powershell
$env:TEST_DATABASE_URL='postgres://testra:testra@localhost:5432/testra?sslmode=disable'
go test -tags=integration -count=1 .\tests\integration
```

---

## Verification

| Check | Command | Result |
|---|---|---|
| Go build | `go build .\...` | PASS |
| Go vet | `go vet .\...` | PASS |
| Unit tests | `go test -count=1 .\...` | PASS |
| Integration tests | `go test -tags=integration -count=1 .\tests\integration` | PASS |
| Frontend typecheck | `pnpm turbo run typecheck` | PASS |

---

## Conclusion

All Phase 3 architecture review conditions are resolved. The deviations are formally accepted through ADR-010, ADR-011, and ADR-012. The implementation matches the accepted architecture and the OpenAPI contract. Phase 3 is ready for final gate approval.
