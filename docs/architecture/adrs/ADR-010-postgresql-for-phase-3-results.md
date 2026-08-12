# ADR-010: PostgreSQL for Phase 3 Results (Deferred ClickHouse)

**Status:** Accepted  
**Date:** July 2026

## Context

DATABASE_DOCUMENTATION §1 and ENGINEERING_STANDARDS §4.2 assign test results, events, and time-series analytical data to ClickHouse. The planned ingestion sequence diagram (SEQUENCE_DIAGRAMS.md) shows Worker → ClickHouse for analytical writes.

Phase 3 implementation stores `test_runs` and `test_run_items` in PostgreSQL. This deviates from the approved data classification.

## Decision

Use PostgreSQL as the sole storage for Phase 3 test runs and results. ClickHouse integration is deferred to a future phase when result volume justifies analytical storage.

### Rationale

1. **MVP volume:** Early-stage test run volume does not justify ClickHouse operational complexity. PostgreSQL handles thousands of runs with standard indexes.
2. **Transactional consistency:** Test run status transitions (pending → running → passed/failed) require transactional updates. PostgreSQL provides ACID guarantees; ClickHouse is append-oriented and does not support row-level updates.
3. **Simpler deployment:** ClickHouse is listed as "Optional — not needed until Phase 3" in README.md. Requiring it for MVP adds operational burden without proportional value.
4. **Storage abstraction exists:** The `results.Repository` interface provides a clean swap point. A `ClickHouseRepository` can be implemented without changing service or handler code.
5. **Metadata constraint:** The `test_runs.metadata` JSONB column is constrained to an explicit allowlist of keys: `format` (string), `ci_build_id` (string), `ci_branch` (string), `ci_commit` (string). No arbitrary customer data may be stored in this field.

### When to Revisit

ClickHouse should be introduced when:
- Result volume exceeds 100,000 records/day
- Analytical queries (trend analysis, flaky test detection) become a performance concern on PostgreSQL
- The async ingestion pipeline (ADR-011) is implemented

## Consequences

- **Positive:** Simpler MVP deployment, transactional consistency for run status, no additional infrastructure required.
- **Negative:** Analytical queries on results will be slower at scale; migration to ClickHouse will require a data backfill.
- **Mitigation:** The `Repository` interface abstraction ensures the swap is isolated to a new repository implementation + module wiring change. The `RunInTx` method can be implemented as a no-op in a ClickHouse repository since analytical writes are append-only.

## Migration Path

1. Implement `ClickHouseRepository` satisfying `results.Repository`
2. Add ClickHouse migration scripts for `test_runs` and `test_run_items` equivalent tables
3. Wire `ClickHouseRepository` in `apps/api/internal/results/module.go` for analytical reads
4. Keep PostgreSQL for transactional status (run lifecycle) and use ClickHouse for analytical fact inserts
5. Backfill historical data from PostgreSQL to ClickHouse
