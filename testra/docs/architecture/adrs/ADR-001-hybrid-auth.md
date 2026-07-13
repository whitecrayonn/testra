# ADR-001: Hybrid Self-Hosted Authentication

**Status:** Accepted
**Date:** July 2026

## Context

The Software Architecture Decision Document (§7) originally specified Clerk for core identity and WorkOS for enterprise SAML/SCIM. During the CTO implementation review, this was revisited against Testra's constraints: privacy-first positioning, zero vendor data exposure, solo-developer budget, and portfolio quality.

## Decision

Replace Clerk with **self-hosted core authentication** implemented in the Go `identity` module:

- Password hashing using the approved maintained password-hashing policy, with migration from the current bcrypt baseline to Argon2id when implementation readiness permits
- 15-minute access JWTs issued by the API
- Rotating opaque refresh tokens stored as hashes, with 30-day inactivity and 90-day absolute expiry
- TOTP MFA required for organization administrators and enterprise users, enforceable organization-wide
- Scoped, hashed API keys for CI/CD ingestion, 90-day default expiry and 365-day maximum
- Password reset via SMTP (Mailpit locally)

**WorkOS is deferred** and will be added behind a clean port in the `identity` module only when the first enterprise deal requires SAML/SCIM.

## Consequences

- **Positive:** zero identity vendor cost; no customer PII shared with third parties; full control over session and API-key semantics; stronger privacy story for enterprise buyers.
- **Negative:** we own security-sensitive code (hashing, session revocation, MFA); enterprise SSO is not available until WorkOS integration lands.
- **Mitigation:** use well-audited libraries only, mandatory security review checklist before production, and rate limiting on all auth endpoints.
