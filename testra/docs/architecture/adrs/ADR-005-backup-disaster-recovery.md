# ADR-005: Backup, Disaster Recovery, and Retention

**Status:** Accepted
**Date:** July 2026

## Context

Testra requires privacy-preserving recovery with low operating cost. PostgreSQL is authoritative; ClickHouse is derived analytical data; Redis is ephemeral coordination; object storage may contain customer-owned artifacts.

## Decision

### PostgreSQL

- Automated daily snapshots retained 35 days.
- Continuous WAL/PITR retained 35 days.
- Weekly snapshot retained 12 weeks.
- MVP target: RPO ≤ 5 minutes and RTO ≤ 4 hours for a regional service failure.
- Beta target: RPO ≤ 1 minute and RTO ≤ 2 hours using Multi-AZ and tested restore automation.
- Enterprise contracts may tighten targets but may not weaken the baseline without an ADR.

### ClickHouse

- Daily backup retained 14 days; weekly backup retained 8 weeks.
- Analytical results retained 13 months by default, with shorter tenant-configured retention where legally/product-approved.
- Re-ingestion from durable source is preferred over treating ClickHouse as transactional authority.

### Object Storage

- S3 versioning enabled for all production buckets.
- Lifecycle transitions to lower-cost storage after 30 days where appropriate.
- Cross-region replication for enterprise production; same-region redundancy for MVP/Beta.
- Customer deletion requests remove current and non-required historical versions according to legal/audit policy.

### Audit and Operational Data

- Immutable audit records retained 7 years by default for enterprise governance; MVP/Beta minimum 2 years unless contract or law requires longer.
- Application logs retained 30 days hot and 90 days archived.
- Metrics retained 15 months for capacity and release trend analysis.
- Traces retained 14 days, with sampled incident traces exported separately when necessary.

All backups are encrypted with KMS-managed keys, access-controlled, monitored, and restore-tested quarterly. Redis is not backed up as business data; durable job sources and idempotency records are used to recover work.

## Consequences

The policy provides concrete engineering targets and bounded storage costs. Enterprise retention and residency requirements may require separate AWS regions/accounts and ClickHouse resources, but the baseline remains maintainable for a solo operator.
