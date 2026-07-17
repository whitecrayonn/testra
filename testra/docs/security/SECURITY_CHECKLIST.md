# Testra Security Checklist

**Purpose:** Provide a security checklist for every security-sensitive change and before production launch.
**Owner:** Security / CTO
**Scope:** Identity, data, application, and operations security controls.
**Source of Truth:** SECURITY_CHECKLIST.md for security review; ADR-007 for security standards.
**Last Updated:** July 2026
**Related documents:**
- [`BIBLICAL_TESTRA.md`](../BIBLICAL_TESTRA.md)
- [`ADR-007-security-standards.md`](../architecture/adrs/ADR-007-security-standards.md)
- [`PRODUCTION_READINESS_CHECKLIST.md`](../operations/PRODUCTION_READINESS_CHECKLIST.md)

Use this checklist for every security-sensitive change and before production launch.

## Identity and Access

- [ ] Passwords require at least 12 characters and use the approved maintained hashing policy.
- [ ] Access JWTs expire after 15 minutes; rotating opaque refresh tokens use 30-day inactivity and 90-day absolute expiry.
- [ ] JWT signing keys are secret-managed, rotated at least every 90 days, and never logged.
- [ ] MFA enrollment, verification, recovery, and reset flows are threat-reviewed; administrators and enterprise users are enforced.
- [ ] API keys expire by default after 90 days, never exceed 365 days, are scoped, hashed, displayed once, revocable, and auditable.
- [ ] Rate limits and lockout/abuse controls cover authentication endpoints using the ADR-007 thresholds.
- [ ] Authorization checks enforce organization/workspace/project scope.
- [ ] Default roles and permissions follow least privilege.

## Data and Privacy

- [ ] No customer source code or raw API collection payloads are retained contrary to product policy.
- [ ] Sensitive fields are encrypted in transit and protected at rest.
- [ ] Logs, traces, analytics, exports, and backups are reviewed for data leakage.
- [ ] PostgreSQL RLS, middleware scope resolution, request context, service authorization, queue propagation, cache keys, exports, and ClickHouse tenant columns are tested for cross-tenant denial.
- [ ] Data retention and deletion behavior is documented.

## Application and API

- [ ] Inputs are validated and bounded.
- [ ] SQL is parameterized.
- [ ] Error responses do not disclose internals.
- [ ] CORS, CSRF posture, headers, and TLS are environment-appropriate.
- [ ] OpenAPI security requirements and schemas match behavior.
- [ ] Dependencies are scanned and vulnerabilities triaged.
- [ ] Webhooks/integrations authenticate, authorize, and prevent replay.

## Operations

- [ ] Secrets are managed outside source control.
- [ ] Audit/security events are observable and alertable.
- [ ] Backups are encrypted with KMS, access-controlled, monitored, and restore-tested quarterly.
- [ ] Incident contacts and revocation procedures are current.
- [ ] Production readiness and release checklists are complete.

**Note:** ADR-001 records the accepted hybrid authentication direction and explicitly makes security ownership a Testra responsibility.

## See Also

- [`BIBLICAL_TESTRA.md`](../BIBLICAL_TESTRA.md) — canonical engineering handbook
- [`ADR-007-security-standards.md`](../architecture/adrs/ADR-007-security-standards.md) — security standards ADR
- [`PRODUCTION_READINESS_CHECKLIST.md`](../operations/PRODUCTION_READINESS_CHECKLIST.md) — go-live gates
