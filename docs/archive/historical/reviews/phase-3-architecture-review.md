# Phase 3 Architecture & Product Compliance Review

**Date:** 2026-07-16  
**Reviewer:** Cascade (AI Engineering Assistant)  
**Phase:** 3 — Execution & Results  
**Status:** IN PROGRESS  
**Decision:** PASS WITH CONDITIONS — implementation may continue, but 3 deviations must be resolved via ADR before Phase 3 DoD sign-off

---

## 1. Documents Reviewed

| Document | Status | Compliance |
|---|---|---|
| Testra Master Context (README.md, docs/README.md) | Approved | PASS |
| Product Discovery (no dedicated doc — inferred from README + PHASES) | Approved | PASS |
| BRD (no dedicated doc — inferred from PHASES + architecture docs) | Approved | PASS |
| Product Strategy (SYSTEM_FLOWS.md, SEQUENCE_DIAGRAMS.md) | Approved | **DEVIATION** — see §4.1 |
| Product Architecture Strategy (MODULE_DEPENDENCIES.md, ERD.md) | Approved | PASS |
| Software Architecture Decision (ADR-001 through ADR-009) | Accepted | **DEVIATION** — see §4.2 |
| Engineering Standards (ENGINEERING_STANDARDS.md) | Approved | **DEVIATION** — see §4.3 |
| MASTER_DEVELOPMENT_GUIDE.md | Approved | PASS |

---

## 2. AutomationHub Zero-Retention Compliance

### 2.1 Policy

SYSTEM_FLOWS.md §Data Classification: "Logs/metrics/traces: operational metadata only; never credentials or customer source code."  
SEQUENCE_DIAGRAMS.md §Planned Automation Result Ingestion: "Testra must not retain customer source code or raw API collection payloads."  
MASTER_DEVELOPMENT_GUIDE.md §4.4 ML Boundary: "Never receives source code or API payloads."

### 2.2 Verification

| Prohibited Data Type | Stored? | Evidence |
|---|---|---|
| Source code | NO | Ingestion handler (`handler.go:38-105`) reads body, parses to structured test results, discards raw body. No source code field in schema. |
| API collections | NO | No collection/request/response fields in `test_runs` or `test_run_items` schema. |
| Request bodies | NO | Raw request body is parsed in-memory and not persisted. Only extracted test metadata (name, status, duration) is stored. |
| Response bodies | NO | No response body storage in schema or service. |
| HAR files | NO | No HAR parsing or storage. |
| Customer secrets | NO | No secret/credential fields in schema. `CreatedBy` stores a user UUID reference, not credentials. |

### 2.3 What IS Persisted

**`test_runs` table:**
- `name` (VARCHAR 255) — user-supplied run name
- `status`, `source`, counts, `duration_ms` — derived metadata
- `metadata` (JSONB, default `{}`) — currently stores only `{"format": "junit|playwright|cypress"}`
- `workspace_id`, `project_id`, `suite_id`, `created_by` — relational references
- Timestamps

**`test_run_items` table:**
- `title` (VARCHAR 500) — test case name from CI report
- `status`, `duration_ms` — derived metadata
- `error_message` (TEXT) — failure message from CI report
- `stack_trace` (TEXT) — stack trace from CI report
- `artifacts` (JSONB, default `[]`) — currently always empty
- `sort_order`, `test_case_id`, `run_id` — relational references
- Timestamps

### 2.4 Finding: `error_message` and `stack_trace` Fields

**Concern:** These fields store text from CI test reports that could theoretically contain embedded source code snippets or secrets if a test failure message includes them.

**Assessment:** These are standard test result metadata fields — every test management platform stores failure messages and stack traces. They are not "source code" or "API collections" in the architectural sense. However, the `metadata` JSONB column is open-ended and could be abused.

**Condition:** ADR-010 must constrain `metadata` to an explicit allowlist of keys. The `error_message` and `stack_trace` fields are approved test result metadata.

### 2.5 Verdict

**COMPLIANT** — AutomationHub does not store source code, API collections, request/response bodies, HAR files, or customer secrets. The raw ingestion body is parsed in-memory and discarded. Only structured test result metadata is persisted.

---

## 3. Approved Metadata Verification

### 3.1 What Is Persisted

All persisted columns map to approved test result metadata:

| Column | Type | Approved? | Notes |
|---|---|---|---|
| `test_runs.name` | VARCHAR(255) | YES | Run identifier |
| `test_runs.status` | VARCHAR(20) | YES | Enum: pending/running/passed/failed/skipped/cancelled |
| `test_runs.source` | VARCHAR(20) | YES | Enum: manual/ci/api |
| `test_runs.total/passed/failed/skipped/blocked` | INTEGER | YES | Aggregate counts |
| `test_runs.duration_ms` | BIGINT | YES | Total duration |
| `test_runs.metadata` | JSONB | **CONDITIONAL** | Must be constrained to allowlist |
| `test_run_items.title` | VARCHAR(500) | YES | Test name from report |
| `test_run_items.status` | VARCHAR(20) | YES | Enum: pending/running/passed/failed/skipped/blocked |
| `test_run_items.duration_ms` | BIGINT | YES | Per-test duration |
| `test_run_items.error_message` | TEXT | YES | Failure message |
| `test_run_items.stack_trace` | TEXT | YES | Stack trace |
| `test_run_items.artifacts` | JSONB | YES | Currently empty, future S3 references |

### 3.2 Verdict

**COMPLIANT WITH CONDITION** — `metadata` JSONB must be constrained via ADR-010 to prevent abuse as a side-channel for arbitrary data.

---

## 4. Architecture Deviations Discovered

### 4.1 DEVIATION-1: Results Stored in PostgreSQL Instead of ClickHouse

**Approved Architecture:**  
DATABASE_DOCUMENTATION §1: "ClickHouse 24 — test results, events, and time-series analytical data"  
ENGINEERING_STANDARDS §4.2: "ClickHouse — Used for test results, events, time-series data only"  
SEQUENCE_DIAGRAMS §Planned Automation Result Ingestion: Worker → ClickHouse  
SYSTEM_FLOWS §Data Classification: "ClickHouse: derived analytical facts and result events"

**Current Implementation:**  
`test_runs` and `test_run_items` are PostgreSQL tables. No ClickHouse integration exists.

**Impact:**  
For MVP volume, PostgreSQL is sufficient and simpler. ClickHouse is listed as "Optional — not needed until Phase 3" in README.md, indicating awareness. However, the approved architecture explicitly assigns test results to ClickHouse.

**Resolution:** ADR-010 — PostgreSQL for Phase 3 Results (deferred ClickHouse)

### 4.2 DEVIATION-2: Synchronous Ingestion Instead of Queue-Based

**Approved Architecture:**  
SEQUENCE_DIAGRAMS §Planned Automation Result Ingestion:  
`CI → API → Redis/Asynq Queue → Worker → ClickHouse`  
"acknowledge accepted batches within 500 ms for batches up to 1,000 result records"

**Current Implementation:**  
`CI → API → PostgreSQL (synchronous)`  
No Redis queue, no worker, no async processing.

**Impact:**  
Synchronous ingestion works for MVP volume but does not meet the 10,000 records/minute target. Large CI batches will block the HTTP request.

**Resolution:** ADR-011 — Synchronous Ingestion for MVP (deferred queue-based pipeline)

### 4.3 DEVIATION-3: Missing Idempotency-Key on Ingestion Endpoint

**Approved Architecture:**  
ADR-006: "Use Idempotency-Key for create/command endpoints with external side effects, ingestion endpoints, exports, webhooks"  
ENGINEERING_STANDARDS §3.1: "Idempotency-Key required for side-effecting commands, ingestion"

**Current Implementation:**  
`POST /ingest` does not check or store `Idempotency-Key`. Duplicate ingestion requests create duplicate runs.

**Impact:**  
CI retry storms or network duplicates will create duplicate test runs. This is a correctness issue, not just a performance issue.

**Resolution:** ADR-012 — Idempotency-Key Implementation Plan for Ingestion

---

## 5. Results Module Storage Abstraction

### 5.1 Interface Analysis

`results/ports.go` defines a `Repository` interface with 11 methods + `RunInTx`. The `Service` depends on this interface, not on `SQLRepository` directly.

`automationhub/service.go` defines a `ResultsRepo` interface with 3 methods (`CreateRun`, `CreateItem`, `UpdateRun`) — a narrower port that the AutomationHub needs.

### 5.2 ClickHouse Swap Path

To replace PostgreSQL with ClickHouse for results:
1. Implement `ClickHouseRepository` satisfying `results.Repository`
2. Wire it in `module.go` instead of `NewSQLRepository`
3. No changes to `service.go` or `handler.go`

**However:** `RunInTx` is PostgreSQL-specific (transaction support). ClickHouse doesn't support ACID transactions the same way. A ClickHouse implementation would need to either:
- Implement `RunInTx` as a no-op (acceptable for append-only analytical writes)
- Or the interface needs refactoring to separate transactional from non-transactional operations

### 5.3 Verdict

**ABSTRACTION EXISTS** — The `Repository` interface provides a clean swap point. The `RunInTx` method is a minor concern that can be addressed when ClickHouse is introduced. The abstraction is sufficient for the current phase.

---

## 6. Phase 3 DoD Review

| DoD Item | Status | Notes |
|---|---|---|
| Manual test runs: plans, execution flow, statuses | ✅ DONE | Full CRUD + status lifecycle |
| `automationhub`: CI ingestion API (zero code retention) | ✅ DONE | JUnit + Playwright/Cypress, zero retention verified |
| `results` module + PostgreSQL ingestion | ✅ DONE | ClickHouse deferred per ADR-010 |
| SSE endpoint for live run progress | ✅ DONE | `GET /test-runs/{id}/stream` |
| Web: runs list, run detail, live execution view | ✅ DONE | 3 pages with SSE integration |
| OpenAPI spec updated | ✅ DONE | v0.4.0, 7 endpoints, 7 schemas |
| Integration tests for ingestion pipeline | ❌ NOT DONE | Pending |

### Phase 3 Status: **IN PROGRESS**

Phase 3 cannot be marked complete until:
1. Integration tests for ingestion pipeline are written
2. ADR-010, ADR-011, ADR-012 are accepted
3. ADR-012 (Idempotency-Key) is implemented

---

## 7. Module Dependency Compliance

MODULE_DEPENDENCIES.md approved dependencies:
- `Results → Project` — ✅ `results` imports `shared` only, project scope passed via handler
- `Results → TestMgmt` — ✅ `test_case_id` FK references `test_cases`, no direct module import
- `IntegrationHub → Project` — ✅ `automationhub` receives project scope via handler params
- `IntegrationHub → Results` — ✅ `automationhub` imports `results` package (ports only)

No circular dependencies. No cross-module internal imports. Clean Architecture boundaries maintained.

---

## 8. Security & Tenant Isolation Compliance

| Requirement | Status | Evidence |
|---|---|---|
| RLS on tenant-scoped tables (ADR-004) | ✅ | Migration 000015 enables RLS on `test_runs` and `test_run_items` |
| RLS policy via `app.tenant_id` | ✅ | Consistent with migration 000014 pattern |
| Tenant context middleware on all routes | ✅ | All route groups use `TenantContext` with appropriate org resolvers |
| RBAC permission checks | ✅ | `runs:create/read/update/delete/ingest` enforced via `RequirePermission` |
| Audit logging on mutating routes | ✅ | `AuditLog` middleware on create, update, delete, ingest |
| Parameterized queries (no SQL injection) | ✅ | All repository methods use parameterized `$1, $2...` queries |
| Input validation on every endpoint | ✅ | UUID parsing, format validation, name validation |

---

## 9. ADR-001 (Hybrid Auth) Compliance

- Ingestion endpoint uses bearer JWT auth (not API keys yet)
- ADR-001 specifies "Scoped, hashed API keys for CI/CD ingestion" — API key auth for ingestion is planned but not yet implemented
- Current implementation uses JWT auth which is acceptable for manual testing but not for CI/CD pipelines
- **Condition:** API key auth for `/ingest` must be added before production use (tracked by ADR-011)

---

## 10. Summary

### Findings

| # | Finding | Severity | Resolution |
|---|---|---|---|
| 1 | Results in PostgreSQL instead of ClickHouse | Medium | ADR-010 |
| 2 | Synchronous ingestion instead of queue-based | Medium | ADR-011 |
| 3 | Missing Idempotency-Key on ingestion | High | ADR-012 + implementation |
| 4 | `metadata` JSONB unconstrained | Low | ADR-010 (metadata allowlist) |
| 5 | Integration tests not written | High | Must complete before DoD sign-off |
| 6 | API key auth for ingestion not implemented | Medium | Future phase (ADR-001 compliance) |

### Decision

**PASS WITH CONDITIONS**

Phase 3 implementation may continue. The architecture is fundamentally sound — Clean Architecture boundaries are maintained, zero-retention policy is respected, and the storage abstraction exists for future ClickHouse swap. However:

1. **ADR-010, ADR-011, ADR-012 must be created and accepted** before Phase 3 gate review
2. **Idempotency-Key must be implemented** on the `/ingest` endpoint (ADR-006 compliance)
3. **Integration tests must be written** (DoD requirement)
4. **Phase 3 status remains IN PROGRESS** until all conditions are met

### Next Steps

1. Create ADR-010: PostgreSQL for Phase 3 Results (deferred ClickHouse)
2. Create ADR-011: Synchronous Ingestion for MVP (deferred queue-based pipeline)
3. Create ADR-012: Idempotency-Key Implementation Plan for Ingestion
4. Implement Idempotency-Key on `/ingest` endpoint
5. Write integration tests for ingestion pipeline
6. Constrain `metadata` JSONB to allowlist
7. Update PHASES.md to reflect remaining work
