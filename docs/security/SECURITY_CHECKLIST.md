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

Use this checklist for every security-sensitive change and before production launch. Checking a box is a reviewer sign-off action for launch, not an implementation-status flag — see the Sprint 4 note below for what a Sprint 4 M1 cleanup pass actually implemented.

> **Sprint 4 status (2026-08-04):** Since Sprint 3, the following controls were added or hardened (see `SPRINT_BACKLOG.md` M1 status note and PR #15 for detail): account lockout after repeated failed logins with uniform-response account-enumeration resistance (identity, item below); refresh-token invalidation and login-lockout clearing on password reset; a magic-link (HTTPS URL, not a bare token) password-reset email; SSRF hostname validation now caches DNS results for a short (30s) window with a size-bounded, pruned cache and dial-time re-validation pinned to the validated address (closing the gap between validation and the actual outbound connection); a CSP `report-uri` endpoint with body-size and shape limits; and a first pass at CI dependency scanning (Trivy + SBOM, currently report-only pending remediation of pre-existing CVEs it surfaced — see the follow-up task). This does not itself constitute the sign-off below; a reviewer still needs to check the relevant boxes with evidence before launch.

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
