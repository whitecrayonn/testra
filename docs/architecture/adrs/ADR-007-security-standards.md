# ADR-007: Security Standards

**Status:** Accepted
**Date:** July 2026

## Context

ADR-001 selects self-hosted core authentication. Testra therefore owns session security, credential protection, abuse prevention, and auditability while preserving a future enterprise SSO port.

## Decision

- Access JWT lifetime is 15 minutes. Access and refresh tokens are issued as httpOnly, Secure, SameSite=Lax cookies for browser clients; non-browser clients may continue to supply the access JWT in the `Authorization: Bearer` header.
- Use rotating, opaque refresh tokens with 30-day inactivity expiry and 90-day absolute expiry. Store only a hash and session family metadata in PostgreSQL. Reuse detection revokes the entire family.
- Support explicit per-session revocation, user-wide revocation, password-change revocation, MFA-reset revocation, and suspected-compromise revocation.
- API keys expire after 90 days by default, require explicit rotation, are scoped to organization/workspace/project, are stored as SHA-256 hashes, and are displayed once. Maximum expiry is 365 days; non-expiring keys are prohibited.
- Passwords require at least 12 characters, breached-password rejection when a maintained offline corpus is available, and no arbitrary composition rule that encourages unsafe reuse. Password reset tokens are single-use and expire after 30 minutes.
- TOTP MFA is optional for MVP users, required for organization administrators and all enterprise users, and enforceable organization-wide. Recovery codes are single-use and hashed.
- Mutating requests authenticated with cookies require a valid double-submit CSRF token in the `X-CSRF-Token` header matching the `testra_csrf_token` cookie. The `GET /auth/csrf` endpoint issues the CSRF cookie.
- Rate limits use Redis token buckets: login 10 attempts per IP per 15 minutes and 5 per account per 15 minutes; registration 5 per IP per hour; password reset 5 per account per hour; API keys 120 requests/minute per key by default, with endpoint-specific limits.
- Rotate JWT signing keys at least every 90 days and immediately on compromise, retaining verification keys for the maximum access-token lifetime. Rotate database, Redis, SMTP, storage, and integration credentials at least every 90 days or through provider-managed rotation.
- Audit authentication, authorization changes, membership/role changes, API-key lifecycle, MFA lifecycle, password/reset events, exports, deletion, administrative access, security configuration, and suspected abuse. Audit records are immutable and tenant-scoped.

## Consequences

Short JWTs reduce token theft impact; refresh rotation adds session state but enables revocation. Mandatory admin/enterprise MFA aligns with enterprise expectations without blocking MVP adoption. Rate limits and audit events increase operational requirements but are essential under ADR-001.
