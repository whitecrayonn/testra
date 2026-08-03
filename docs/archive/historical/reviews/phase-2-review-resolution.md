# Engineering Progress Report — Phase 2 Review Resolution

**Date:** 2026-07-16  
**Phase:** 2 — Test Management Core (Review Resolution)  
**Status:** All HIGH severity issues resolved and verified

---

## Summary

Resolved all 4 HIGH severity issues identified in the Phase 2 Engineering Review. Build, tests, and typecheck all pass.

## Issues Fixed

### TD-1: RLS Policies on Test Tables (HIGH)

**Problem:** `test_folders`, `test_suites`, `test_cases`, and `test_case_versions` tables were created without Row Level Security policies, creating a tenant isolation gap inconsistent with the security model established in migration 000009.

**Fix:** Created migration `000014_add_test_management_rls` that:
- Enables RLS on all 4 test management tables
- Creates policies that scope access through `workspace_id → workspaces.organization_id → app.tenant_id`
- `test_case_versions` scoped through `test_case_id → test_cases.workspace_id → workspaces.organization_id`

**Files:**
- `apps/api/migrations/000014_add_test_management_rls.up.sql` (created)
- `apps/api/migrations/000014_add_test_management_rls.down.sql` (created)

### TD-2: Search Cursor Pagination (HIGH)

**Problem:** `SearchCases` ordered results by `ts_rank DESC, id DESC` but the cursor only filtered by `id < $cursor`, causing incorrect pagination — records with higher IDs but lower rank were incorrectly excluded on subsequent pages.

**Fix:** Implemented composite `(rank, id)` cursor pagination:
- `SearchCases` now returns `([]TestCase, string, error)` — the string is the next cursor
- Cursor encodes both `rank` (float64) and `id` (UUID string) as base64-encoded JSON
- WHERE clause uses `(rank < $cursorRank OR (rank = $cursorRank AND id < $cursorID))` for stable keyset pagination
- Added `scanCasesWithRank` to capture the `ts_rank` value from each row
- Added `encodeSearchCursor`/`decodeSearchCursor` helpers
- Handler updated to use the repository-returned cursor instead of encoding from last item ID
- `Repository` interface, `Service`, `Handler`, and test fake repository all updated for the new signature

**Files:**
- `apps/api/internal/testmanagement/ports.go` (modified — SearchCases signature)
- `apps/api/internal/testmanagement/repository.go` (modified — composite cursor, scanCasesWithRank, encode/decode helpers)
- `apps/api/internal/testmanagement/service.go` (modified — SearchCases signature)
- `apps/api/internal/testmanagement/handler.go` (modified — use returned cursor)
- `apps/api/internal/testmanagement/service_test.go` (modified — fake repo + test)

### TD-3: Transactional UpdateCase (HIGH)

**Problem:** `UpdateCase` performed `CreateVersion` (snapshot) then `UpdateCase` (apply update) as two separate database operations. If the update failed, the version snapshot was orphaned — a data integrity issue.

**Fix:** Wrapped both operations in a single database transaction:
- Added `RunInTx(ctx, func(Repository) error)` to the `Repository` interface
- Implemented `RunInTx` in `SQLRepository` using `BeginTx`/`Commit`/`Rollback`
- Added `DBTX` interface to allow `SQLRepository` to work with both `*sql.DB` and `*sql.Tx`
- `UpdateCase` now calls `s.repo.RunInTx` — if either CreateVersion or UpdateCase fails, the entire transaction rolls back
- Validation (status, priority) moved before the transaction to prevent orphaned snapshots from invalid input
- Fake repository implements `RunInTx` as a passthrough for unit tests

**Files:**
- `apps/api/internal/testmanagement/ports.go` (modified — RunInTx method)
- `apps/api/internal/testmanagement/repository.go` (modified — DBTX, RunInTx)
- `apps/api/internal/testmanagement/service.go` (modified — transactional UpdateCase)
- `apps/api/internal/testmanagement/service_test.go` (modified — fake RunInTx)

### TD-4: Audit Logging on Test Management Mutations (HIGH)

**Problem:** No `AuditLog` middleware was applied to any of the 9 test management mutating routes (3 create, 3 update, 3 delete), despite the Phase 2 DoD requiring "immutable event log on all mutations".

**Fix:** Wired `AuditLog` middleware on all 9 mutating routes in `server.go`:
- `test_folder.create` — POST `/test-folders`
- `test_suite.create` — POST `/test-suites`
- `test_case.create` — POST `/test-cases`
- `test_case.update` — PUT `/test-cases/{id}`
- `test_case.delete` — DELETE `/test-cases/{id}`
- `test_folder.update` — PUT `/test-folders/{id}`
- `test_folder.delete` — DELETE `/test-folders/{id}`
- `test_suite.update` — PUT `/test-suites/{id}`
- `test_suite.delete` — DELETE `/test-suites/{id}`

Each route uses the existing `auditLogFn` and `sharedmiddleware.UserIDFromContext` for user extraction. Create routes pass empty string for resource ID (ID is generated server-side). Update/delete routes extract the ID from the URL parameter.

**Files:**
- `apps/api/internal/shared/server/server.go` (modified — 9 routes updated)

---

## Verification

| Check | Result |
|---|---|
| `go build ./...` | PASS (exit code 0) |
| `go test -count=1 ./...` | PASS (identity: ok, project: ok, testmanagement: ok) |
| `npx tsc --noEmit` (web) | PASS (exit code 0) |
| All existing unit tests | PASS (no regressions) |
| Lint errors | None |

---

## Remaining Technical Debt

| ID | Issue | Severity | Status |
|---|---|---|---|
| TD-5 | `pqArray`/`parseTags` don't handle special chars | MEDIUM | Deferred |
| TD-6 | No CHECK constraints on status/priority columns | MEDIUM | Deferred |
| TD-7 | No UNIQUE on test_case_versions(case_id, version) | MEDIUM | Deferred |
| TD-8 | `mapCaseResponse` mutates input domain object | LOW | Deferred |
| TD-9 | Frontend localStorage reads during render | MEDIUM | Deferred |
| TD-10 | Frontend setTimeout hack for search state | MEDIUM | Deferred |
| TD-11 | `mapError` duplicated across 6 handlers | LOW | Pre-existing |
| TD-12 | `uuid.Parse` errors silently ignored in repository | LOW | Deferred |
| TD-13 | `json.Marshal` errors silently ignored in repository | LOW | Deferred |
| TD-14 | No `updated_at` auto-update trigger | LOW | Pre-existing |
| TD-15 | No step content validation (empty action/expected) | LOW | Deferred |
| TD-16 | Paginated response double-wrapping in Envelope | LOW | Pre-existing |

---

## Files Changed

**Created:**
- `apps/api/migrations/000014_add_test_management_rls.up.sql`
- `apps/api/migrations/000014_add_test_management_rls.down.sql`

**Modified:**
- `apps/api/internal/testmanagement/ports.go` — SearchCases signature, RunInTx method
- `apps/api/internal/testmanagement/repository.go` — DBTX interface, RunInTx, composite cursor pagination, scanCasesWithRank, encode/decode search cursor
- `apps/api/internal/testmanagement/service.go` — SearchCases signature, transactional UpdateCase with validation-before-snapshot
- `apps/api/internal/testmanagement/handler.go` — SearchCases handler uses returned cursor
- `apps/api/internal/testmanagement/service_test.go` — fake repo SearchCases + RunInTx, test fix
- `apps/api/internal/shared/server/server.go` — AuditLog middleware on 9 mutating routes
- `docs/engineering/PHASES.md` — DoD updated with review resolution items

---

## Conclusion

All 4 HIGH severity issues from the Phase 2 Engineering Review have been resolved and verified. Phase 2 is now fully complete with no blocking issues remaining. Phase 3 may begin.
