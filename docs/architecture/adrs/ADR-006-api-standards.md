# ADR-006: API Standards

**Status:** Accepted
**Date:** July 2026

## Context

Testra needs predictable APIs for the web client, future SDK, CI/CD integrations, and enterprise consumers without premature complexity.

## Decision

- Use cursor pagination for all new list endpoints. Parameters are `cursor` and `page_size`; default page size is 50 and maximum is 100. Stable ordering uses a unique tie-breaker such as `(created_at, id)`. Existing low-volume endpoints may remain unpaginated only until their next contract revision.
- Use `Idempotency-Key` for create/command endpoints with external side effects, ingestion endpoints, exports, webhooks, and payment-like operations if introduced. Pure reads and deterministic metadata updates do not require it.
- Store an idempotency record in PostgreSQL, scoped by tenant, principal, operation, and key, for 24 hours. Store request fingerprint and final response; reject a reused key with a different fingerprint. Ingestion consumers remain idempotent beyond that window using domain event/result identifiers.
- Continue URL major versioning at `/api/v1`. A major version changes only for breaking contract, authentication, or tenant-visibility changes.
- Compatible additions are allowed in `v1`. Deprecations require OpenAPI marking, replacement guidance, migration notes, and a minimum 12-month compatibility window or two major release cycles, whichever is longer.
- Error codes and response envelopes are stable contract elements. OpenAPI is updated before implementation.

## Consequences

Cursor pagination scales without deep-offset degradation. PostgreSQL idempotency records are durable and auditable. A 24-hour replay window limits storage cost while domain-level ingestion identifiers protect longer-running retries.
