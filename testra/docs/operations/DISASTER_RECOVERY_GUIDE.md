# Testra Disaster Recovery Guide

**Purpose:** Define recovery objectives, backup requirements, recovery procedure, and testing for Testra services.
**Owner:** Platform / SRE Lead
**Scope:** Backup, restore, RPO/RTO, recovery procedure, and testing.
**Source of Truth:** DISASTER_RECOVERY_GUIDE.md and ADR-005 for recovery policy.
**Last Updated:** July 2026
**Related documents:**
- [`BIBLICAL_TESTRA.md`](../BIBLICAL_TESTRA.md)
- [`ADR-005-backup-disaster-recovery.md`](../architecture/adrs/ADR-005-backup-disaster-recovery.md)
- [`DEPLOYMENT_GUIDE.md`](../deployment/DEPLOYMENT_GUIDE.md)
- [`MONITORING_LOGGING_GUIDE.md`](MONITORING_LOGGING_GUIDE.md)

## Recovery Scope

Recover the authoritative PostgreSQL data first, then Redis coordination, ClickHouse analytical data, object storage artifacts, application artifacts, and external integrations. PostgreSQL is the transactional source of truth; ClickHouse and Redis are not interchangeable backups.

## Recovery Objectives

The following are the mandatory baseline objectives. Enterprise contracts may tighten them but may not weaken them without an ADR.

| Service/data | Backup/recovery policy | RPO | RTO |
|---|---|---:|---:|
| PostgreSQL | Daily snapshots 35 days, continuous WAL/PITR 35 days, weekly snapshots 12 weeks | ≤ 5 minutes MVP; ≤ 1 minute Beta | ≤ 4 hours MVP; ≤ 2 hours Beta |
| Object storage | S3 versioning, lifecycle policies, same-region redundancy MVP/Beta, cross-region replication Enterprise | ≤ 24 hours MVP; replication target Beta/Enterprise | ≤ 8 hours MVP; ≤ 4 hours Beta/Enterprise |
| ClickHouse | Daily backups 14 days, weekly backups 8 weeks, re-ingestion from durable source | ≤ 24 hours | ≤ 12 hours |
| Redis queues | No business-data backup; replay durable idempotent jobs | ≤ 1 hour for recoverable queued work | ≤ 4 hours |

Audit records are retained 7 years for enterprise governance and at least 2 years for MVP/Beta. Application logs are retained 30 days hot and 90 days archived. Metrics are retained 15 months. Traces are retained 14 days.

## Backup Requirements

- Encrypted, access-controlled, versioned backups.
- PostgreSQL backups plus restore verification.
- Object storage versioning is mandatory; cross-region replication is mandatory for Enterprise.
- ClickHouse backups and reproducible re-ingestion from durable identifiers are mandatory.
- Secrets and key recovery procedure stored separately from application data.
- Backup success and age monitored with alerts.

## Recovery Procedure

1. Declare incident and freeze risky deployments.
2. Identify failure scope and last known good state.
3. Provision clean infrastructure and verify network/secret controls.
4. Restore PostgreSQL and validate migrations/integrity.
5. Restore object artifacts and configure application services.
6. Rebuild or restore ClickHouse according to the approved strategy.
7. Recreate Redis and replay only idempotent jobs from durable source.
8. Deploy the exact approved application artifact.
9. Run smoke tests for auth, tenant isolation, core reads/writes, queues, and analytics.
10. Resume traffic gradually and monitor.
11. Document data loss, skipped work, customer impact, and follow-ups.

## Testing

Run restore drills at least before production launch and periodically afterward. A backup is untrusted until a restore has been demonstrated and measured.

## See Also

- [`BIBLICAL_TESTRA.md`](../BIBLICAL_TESTRA.md) — canonical engineering handbook
- [`ADR-005-backup-disaster-recovery.md`](../architecture/adrs/ADR-005-backup-disaster-recovery.md) — backup and recovery ADR
- [`DEPLOYMENT_GUIDE.md`](../deployment/DEPLOYMENT_GUIDE.md) — deployment strategy
- [`MONITORING_LOGGING_GUIDE.md`](MONITORING_LOGGING_GUIDE.md) — observability requirements
