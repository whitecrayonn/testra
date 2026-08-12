# ADR-011: Synchronous Ingestion for MVP (Deferred Queue-Based Pipeline)

**Status:** Accepted  
**Date:** July 2026

## Context

SEQUENCE_DIAGRAMS.md §Planned Automation Result Ingestion specifies an async pipeline: `CI → API → Redis/Asynq Queue → Worker → ClickHouse`. The API should "acknowledge accepted batches within 500 ms for batches up to 1,000 result records" with an MVP processing target of 10,000 result records/minute.

Phase 3 implementation uses synchronous ingestion: `CI → API → PostgreSQL`. No Redis queue, no worker, no async processing.

## Decision

Use synchronous ingestion for the MVP. The async queue-based pipeline is deferred to a future phase.

### Rationale

1. **MVP volume:** Early-stage ingestion volume is low (manual runs, small CI batches). Synchronous processing completes well within HTTP timeout for batches up to ~500 test cases.
2. **Operational simplicity:** The async pipeline requires Redis Asynq, a standalone worker process, dead-letter handling, and idempotent consumers. This is unnecessary complexity for MVP.
3. **Error transparency:** Synchronous ingestion returns immediate success/failure to the CI caller. Async ingestion requires callback or polling mechanisms for error reporting.
4. **ADR-001 compliance gap:** The approved ingestion flow uses API key auth (`KeyAuth` in the sequence diagram). Current implementation uses JWT auth. API key auth for CI/CD ingestion must be added before production use, but this is independent of sync vs async.

### When to Revisit

The async pipeline should be implemented when:
- Ingestion batches regularly exceed 500 test cases
- The 500ms acknowledgment SLA becomes a requirement
- Result volume approaches the 10,000 records/minute target
- ClickHouse is introduced (ADR-010) — the worker would write to ClickHouse instead of PostgreSQL

### API Key Auth

API key authentication for the `/ingest` endpoint is required by ADR-001 but not yet implemented. This is tracked as a separate concern and must be completed before production deployment. The current JWT-based auth is acceptable for development and manual testing.

## Consequences

- **Positive:** Simpler deployment (no worker process), immediate error feedback, no queue infrastructure.
- **Negative:** Large batches block the HTTP request; no retry/dead-letter handling; does not meet the 10,000 records/minute target.
- **Mitigation:** The `automationhub.Service.Ingest` method is synchronous but stateless. Moving to async requires wrapping the same logic in an Asynq job handler — the service code does not change.

## Migration Path

1. Add Redis Asynq dependency
2. Create `IngestJob` Asynq task type
3. Move `Service.Ingest` call into worker handler
4. API endpoint enqueues job and returns 202 Accepted
5. Worker processes job and writes to PostgreSQL (or ClickHouse per ADR-010)
6. Add dead-letter queue and retry policy
