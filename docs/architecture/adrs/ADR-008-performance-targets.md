# ADR-008: MVP Performance Targets

**Status:** Accepted
**Date:** July 2026

## Context

Testra needs measurable targets that are achievable for a modular monolith and managed-service MVP, without prematurely optimizing for enterprise-scale traffic.

## Decision

For MVP production in the primary region:

- Read API target: p95 ≤ 300 ms and p99 ≤ 750 ms for ordinary authenticated reads.
- Write API target: p95 ≤ 500 ms and p99 ≤ 1,000 ms for ordinary transactional writes.
- Health endpoint target: p95 ≤ 100 ms excluding dependency checks.
- PostgreSQL query target: p95 ≤ 50 ms for indexed OLTP queries; any query over 200 ms in normal traffic requires review.
- Maximum synchronous request timeout: 30 seconds at the edge; application handlers should use shorter operation-specific deadlines.
- Background job timeout: 5 minutes per attempt; long workflows must be checkpointed and resumable. Default retry count is 5 with exponential backoff and dead-letter handling.
- Maximum upload size: 25 MiB per request for MVP. Larger artifacts must use presigned S3 multipart upload with a 1 GiB per-object limit unless a later ADR changes it.
- Capacity target: 500 concurrent authenticated users and 50 requests/second sustained API traffic, with 2x burst for 5 minutes, while meeting p95 targets.
- ClickHouse ingestion target: acknowledge accepted ingestion within 500 ms for batches up to 1,000 result records; process at least 10,000 result records/minute in the MVP environment with retry and deduplication.

Capacity tests must measure these targets using representative tenant distributions and must not use customer data. Targets are engineering gates, not product promises or a reason to retain prohibited payloads.

## Consequences

The targets are concrete enough for release testing and modest enough for a solo-maintained MVP. Scaling decisions are triggered by measured saturation, not assumptions; Beta increases capacity through additional ECS tasks, Multi-AZ data services, queue workers, and ClickHouse resources.
