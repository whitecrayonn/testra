# Testra Production Readiness Checklist

**Purpose:** Provide a go-live checklist covering product, security, reliability, delivery, and approvals.
**Owner:** Platform / SRE Lead / CTO
**Scope:** Production readiness gates before launch.
**Source of Truth:** PRODUCTION_READINESS_CHECKLIST.md for launch gates; evidence must exist, not merely be planned.
**Last Updated:** July 2026
**Related documents:**
- [`BIBLICAL_TESTRA.md`](../BIBLICAL_TESTRA.md)
- [`RELEASE_CHECKLIST.md`](../release/RELEASE_CHECKLIST.md)
- [`SECURITY_CHECKLIST.md`](../security/SECURITY_CHECKLIST.md)
- [`DEPLOYMENT_GUIDE.md`](../deployment/DEPLOYMENT_GUIDE.md)

## Product and Architecture

- [ ] MVP scope and phase completion are explicitly approved.
- [ ] Current architecture, module dependencies, ERD, and system flows match implementation.
- [ ] ADR-003 through ADR-009 are reflected in implementation and operational evidence.
- [ ] No undocumented production endpoint or data store exists.

## Security and Privacy

- [ ] Authentication, MFA, password reset, RBAC, and API-key controls are implemented and reviewed as applicable.
- [ ] Tenant isolation is tested at HTTP, use-case, repository, queue, cache, and export boundaries.
- [ ] Security checklist is complete.
- [ ] environment files or a local secrets store/KMS, 90-day secret rotation, Let's Encrypt TLS, CORS, ADR-007 rate limits, and dependency scanning are operational.
- [ ] Data retention, deletion, and incident response are approved.

## Reliability and Operations

- [ ] Health checks, metrics, logs, traces, dashboards, and alerts are live.
- [ ] Runbooks cover common incidents and escalation contacts are current.
- [ ] PostgreSQL and object storage backups are encrypted and restore-tested.
- [ ] ADR-005 RPO/RTO, retention, backup schedules, and quarterly restore-drill evidence exist.
- [ ] Queue retry/dead-letter behavior, 24-hour PostgreSQL idempotency records, and domain-level ingestion deduplication are verified.

## Delivery

- [ ] CI gates pass on the release commit.
- [ ] Migrations are reviewed, reversible or have an approved forward-fix plan.
- [ ] Staging soak is complete and ADR-008 performance targets are met under representative load.
- [ ] Release and deployment checklists are signed off.
- [ ] Rollback or forward-fix path is rehearsed.

## Approval

- [ ] Engineering owner
- [ ] Security/privacy owner
- [ ] Operations/platform owner
- [ ] Product/business owner

A checked box means evidence exists, not merely that the feature is planned.

## See Also

- [`BIBLICAL_TESTRA.md`](../BIBLICAL_TESTRA.md) — canonical engineering handbook
- [`RELEASE_CHECKLIST.md`](../release/RELEASE_CHECKLIST.md) — release execution checklist
- [`SECURITY_CHECKLIST.md`](../security/SECURITY_CHECKLIST.md) — security review checklist
- [`DEPLOYMENT_GUIDE.md`](../deployment/DEPLOYMENT_GUIDE.md) — deployment strategy
