# Security Review v2 — Testra Platform

**Date:** 2026-07-19  
**Scope:** Backend API (`apps/api`), web frontend (`apps/web`), deployment configuration (no service unit files or reverse-proxy config exist yet — see `docs/deployment/DEPLOYMENT_GUIDE.md`), and runtime configuration.  
**Goal:** Identify security gaps that affect authentication, authorization, session handling, injection/SSRF, secrets management, and deployment posture, then remediate the critical/high items that can be fixed safely in this pass.

## Executive Summary

This review found the platform has a solid baseline (bcrypt password hashing, parameterized SQL, RS256 JWTs with key rotation, refresh-token rotation and revocation, and PostgreSQL RLS). However, several high-impact gaps were present in the current code. This pass fixed five critical/high severity issues and documented the remaining remediation work.

### Fixed in this review

| ID | Finding | Severity | Fix |
|----|---------|----------|-----|
| SEC-001 | `identity.Service.DisableMFA` allowed disabling MFA with an empty TOTP code, bypassing second-factor authentication. | **Critical** | `apps/api/internal/identity/service.go` now rejects an empty code with `ErrMFARequired` before the database is updated. Tests updated. |
| SEC-002 | `AuditLog` middleware only logged requests with `status < 400`, so failed authentication and authorization attempts were not audited. | **High** | `apps/api/internal/shared/middleware/audit.go` now logs every authenticated request and includes the HTTP status code in the audit metadata. `server.go` maps the new `StatusCode` field into `audit.LogInput.Metadata`. |
| SEC-003 | `apikeys.SQLRepository.Revoke` ignored the error returned by `result.RowsAffected()`. | **Medium** | `apps/api/internal/apikeys/repository.go` now returns `RowsAffected` errors instead of swallowing them. |
| SEC-004 | Integration and notification channel API responses returned raw `token`, `secret`, `password`, and `api_key` values to clients. | **High** | `apps/api/internal/integrationhub/handler.go` and `apps/api/internal/notification/handler.go` now mask sensitive config keys in responses. |
| SEC-005 | `integrationhub` adapters and `notification.dispatchHTTP` made outbound HTTP calls to user-controlled URLs without SSRF protection. | **High** | New `apps/api/internal/shared/security/ssrf.go` validates URLs, blocking `localhost`, private IP ranges, link-local/multicast addresses, and internal host suffixes. Integrated into all integration adapters and notification dispatch. |
| SEC-006 | CORS middleware did not set `Vary: Origin` and did not allow the `Idempotency-Key`, `X-API-Key`, or `X-CSRF-Token` headers for preflight. | **Medium** | `apps/api/internal/shared/server/server.go` `corsMiddleware` now sets `Vary: Origin`, `Access-Control-Max-Age`, and the additional allowed headers. |

### Previously fixed (CORS configuration)

The P0-5 finding from the code-review pass was resolved in the prior session:

- `CORS_ALLOWED_ORIGINS`, `ML_SERVICE_URL`, and `NEXT_PUBLIC_API_URL` are no longer hardcoded anywhere in the codebase; they are read from environment variables (`apps/api/internal/shared/config/config.go`), with `http://localhost:3000` as the local-dev default only.
- Production and staging values are set per environment via each deployment's environment file — see `docs/deployment/DEPLOYMENT_GUIDE.md`. (The original finding referenced Kubernetes ConfigMap overlays; that infrastructure directory has since been deleted along with the Docker/Kubernetes/Terraform tooling per ADR-003.)

## Methodology

1. **Static code review** of authentication (`identity`), authorization (`rbac`, `middleware`), API keys, tenant isolation, audit, notifications, integrations, and outbound HTTP clients.
2. **Dependency and configuration review** of `config.go`, systemd service unit files and nginx site configurations, native services, and environment examples.
3. **Automated verification** with `go build ./...` and `go test -count=1 ./...` after each change.
4. **Threat-modeling** against OWASP Top 10 categories relevant to the codebase.

## Detailed Findings

### 1. Authentication and Session Management

#### 1.1 Password Policy (Medium — Open)
`identity/service.go:validatePassword` only enforces a minimum length of 12 characters. It does not require mixed case, digits, or symbols, and does not reject breached/common passwords.

**Recommendation:**
- Add complexity checks (upper, lower, digit, symbol) or integrate a password-strength library.
- Add a breached-password check against a service such as Have I Been Pwned (rate-limited) before accepting registration/reset.

#### 1.2 Password Reset Token Leakage Surface (Low — Open)
`identity/service.go:RequestPasswordReset` returns the raw reset token to the handler. The handler discards it and emails it, but the value still traverses the call stack. `sendPasswordResetEmail` also includes the raw token in the email body with no magic-link wrapper, increasing phishing/copy-paste risk.

**Recommendation:**
- Return only a success boolean from `RequestPasswordReset`.
- Send a time-bound HTTPS magic link (`https://app.testra.example.com/reset?token=<token>`) instead of a naked token.

#### 1.3 Refresh Token Family Revocation Timing (Low — Open)
`RefreshToken` revokes the old token *after* issuing a new one (line 381). If the request crashes between new-token creation and old-token revocation, an attacker could replay the old token.

**Recommendation:**
- Revoke the presented refresh token *before* issuing the new token, or wrap both operations in a single database transaction.

#### 1.4 Access Token Binding (Low — Open)
JWT access tokens contain `UserID` and `Email` but no `jti`/token ID and no session binding. A stolen access token can be replayed anywhere until it expires (default 15 minutes).

**Recommendation:**
- Add a `jti` claim and maintain a small deny-list for revoked access tokens, or move to short-lived opaque access tokens behind a session store.

### 2. Authorization and RBAC

#### 2.1 Tenant Context Enforcement (Mostly Good, One Gap)
`shared/db/db.go` sets `app.tenant_id` on every transaction, and repositories use parameterized queries. This provides good row-level isolation.

**Gap:** `shared/middleware/tenant.go:OrgIDFromBody` rewrites `r.Body` after reading it for tenant resolution. Handlers later in the chain re-decode the body. This is currently safe but fragile; a future change that forgets to restore the body will break downstream handlers. Added a small comment and test coverage would reduce this risk.

#### 2.2 API Key Scope Validation (Medium — Open)
`apikeys/service.go:Create` accepts `Scopes []string` but does not validate them against a known set. A bug or malicious caller could create a key with arbitrary scope strings that are silently ignored by `rbac.RequirePermission`.

**Recommendation:**
- Define an allowed scope registry and reject unknown scopes in `apikeys.Service.Create`.

### 3. JWT and Cryptography

#### 3.1 JWT Implementation (Good)
`shared/jwt/manager.go` uses RS256, validates `alg`/`kid`, issuer, audience, and expiration, and exposes JWKS. Key rotation keeps historical verification keys available. This is a strong implementation.

#### 3.2 Dev Ephemeral Keys (Low — Open)
`config.go` generates an ephemeral RSA key when `JWT_PRIVATE_KEY_FILE` is not set. `Config.Validate()` blocks this in production, but a misconfigured environment variable could still allow it. This is acceptable given the fail-fast validation.

### 4. SQL Injection

#### 4.1 Parameterized Queries (Good)
All reviewed repositories use `sql.DB`/`DBTX` with `$n` placeholders. No string concatenation was found in SQL statements. The only dynamic SQL-like strings are in `notification/service.go` and `identity/service.go` for SMTP message bodies, which are not SQL.

**Status:** No SQLi findings remain for the backend.

### 5. SSRF and Outbound HTTP

#### 5.1 New SSRF Guard (Fixed)
`shared/security/ssrf.go` blocks:
- Empty or missing URLs
- Non-HTTP(S) schemes
- `localhost` and `127.0.0.1`/`::1`
- Private, link-local, multicast, and loopback IP ranges
- Hostnames ending in `.local`, `.localhost`, or `.internal`
- DNS names that resolve to any blocked IP

The guard is called in `integrationhub/adapters.go` (`jiraAdapter`, `githubAdapter`, `gitlabAdapter`, `slackAdapter`, `webhookAdapter`) and `notification/service.go:dispatchHTTP`.

**Residual Risk:** DNS rebinding can bypass hostname-based checks if the attacker controls both DNS and a TTL. For high-assurance environments, use an egress proxy or dedicated outbound network namespace and a separate URL allowlist.

#### 5.2 ML Service URL (Not User-Controlled)
`intelligence/mlclient.go` uses `ML_SERVICE_URL` from environment. This is not user-controlled, so the SSRF guard was not applied there. Ensure `ML_SERVICE_URL` is restricted to an internal service endpoint (e.g., `localhost` or a private network address) on the production VM, not a publicly reachable address.

### 6. CORS and Web Security

#### 6.1 CORS Hardening (Fixed)
`shared/server/server.go:corsMiddleware` now:
- Adds `Vary: Origin` to every response.
- Matches `Origin` only against explicitly configured origins.
- Allows `Authorization`, `Content-Type`, `Idempotency-Key`, `X-API-Key`, and `X-CSRF-Token` headers.
- Sets `Access-Control-Allow-Credentials: true` and `Access-Control-Max-Age: 600`.

#### 6.2 Production Origin Configuration (Fixed)
`CORS_ALLOWED_ORIGINS` is read from the environment (`apps/api/internal/shared/config/config.go`), with no hardcoded default beyond `http://localhost:3000` for local dev. Each deployment target (staging, production) sets its own value via its environment file. (The overlay-based Kubernetes ConfigMap approach originally described here no longer applies — that infrastructure was removed per ADR-003.)

#### 6.3 Frontend Token Storage (Fixed since this review)
This finding is stale. As of the 2026-07-31 P0 verification pass (`docs/audits/P0_VERIFICATION_REPORT.md` §6), tokens are issued as `httpOnly` cookies (`apps/api/internal/shared/middleware/cookies.go`), and `apps/web/lib/api.ts` sends `credentials: "include"` rather than reading tokens from `localStorage`. `localStorage` is used only for non-secret UI state (workspace/organization/project selection).

### 7. Secrets Management

#### 7.1 Integration/Notification Secrets Masking (Fixed)
Both `integrationhub` and `notification` handlers now strip values for keys containing `token`, `secret`, `password`, `private_key`, or `api_key` before serializing `config` to JSON.

#### 7.2 SMTP Secret Provider (Good)
`identity/service.go` and `notification/service.go` support a `secrets.Provider` for the SMTP password, defaulting to an environment-variable provider. `config.go` `SecretProvider()` is a thin wrapper that can be swapped for a Vault or cloud-secret-manager implementation.

#### 7.3 Deployment Secret Management (Open)
No deployment scripts or environment-file templates exist yet (see `docs/deployment/DEPLOYMENT_GUIDE.md`). `CORS_ALLOWED_ORIGINS` and `ML_SERVICE_URL` are not secrets, but when the production environment file is created it must be readable only by the app user. If genuinely sensitive values need externalizing beyond a plain environment file, migrate them to a secrets store (e.g., Vault) via the existing `secrets.Provider` abstraction (ADR-016).

### 8. Audit and Logging

#### 8.1 Audit Coverage (Fixed)
`AuditLog` middleware previously skipped non-2xx/3xx responses. It now logs every authenticated request and records the HTTP status. This is essential for detecting brute force and privilege escalation attempts.

#### 8.2 PII in Logs (Open)
`RequestLogger` and audit logs currently store `r.RemoteAddr` as the client IP. In a container behind a reverse proxy, `RemoteAddr` may be the internal proxy address unless `middleware.RealIP` is trusted and configured. Also, query parameters and request bodies are not sanitized before logging.

**Recommendation:**
- Explicitly trust `X-Forwarded-For` only from known proxies.
- Redact email, tokens, and `password` fields before logging.

### 9. Rate Limiting

#### 9.1 Fail-Open Behavior (Medium — Open)
`shared/middleware/ratelimit.go` allows the request if the Redis check fails. This is a fail-open design; if Redis is unavailable the rate limiter is disabled.

**Recommendation:**
- Add a circuit-breaker that falls back to a per-instance in-memory limiter, or provide a `fail-closed` mode for the public auth endpoints.

### 10. Deployment and Infrastructure

#### 10.1 Binary Build and Dependencies (Open)
This review did not inspect build scripts/base images. Ensure images are minimal, non-root, and updated regularly.

#### 10.2 Network Policies (Open)
The systemd service unit files and nginx site configurations do not include `host firewall rules` resources. Without them, a compromised pod can reach any other internal service.

**Recommendation:**
- Add `host firewall rules` objects that restrict ingress/egress per service.
- Place ML and database services in isolated namespaces with explicit allow rules.

## Remediation Backlog

The following items are **not** addressed in this pass but should be prioritized before production launch:

1. **Frontend cookie-based auth** (High) — move tokens out of `localStorage`.
2. **Password policy hardening** (Medium) — complexity + breached-password checks.
3. **API key scope validation** (Medium) — reject unknown scopes.
4. **Refresh-token revocation ordering** (Low-Medium) — revoke before issuing.
5. **Access-token `jti`/deny-list** (Low-Medium) — limit replay window.
6. **Rate limiter fail-closed fallback** (Medium) — protect auth endpoints when Redis is down.
7. **Network policies and egress restrictions** (Medium) — contain SSRF blast radius.
8. **PII redaction in logs** (Low) — sanitize `RemoteAddr`, query params, and bodies.
9. **Magic-link password reset** (Low) — reduce token exposure.
10. **compiled binary hardening and SBOM scanning** (Low) — standard DevSecOps hygiene.

## Verification

After the fixes:

```powershell
go build ./...
go test -count=1 ./...
```

Both commands pass with exit code 0 for `apps/api`.

## Files Changed

- `apps/api/internal/identity/service.go`
- `apps/api/internal/identity/service_test.go`
- `apps/api/internal/apikeys/repository.go`
- `apps/api/internal/shared/middleware/audit.go`
- `apps/api/internal/shared/server/server.go`
- `apps/api/internal/integrationhub/adapters.go`
- `apps/api/internal/integrationhub/handler.go`
- `apps/api/internal/notification/service.go`
- `apps/api/internal/notification/handler.go`
- `apps/api/internal/notification/service_test.go`
- `apps/api/internal/shared/security/ssrf.go` (new)

## Conclusion

The platform's security posture improved materially in this pass. The most dangerous gaps (MFA bypass, missing audit of failures, secret leakage, and unprotected outbound HTTP) are now closed. The remaining work is mostly policy, frontend architecture, and infrastructure hardening. The next highest-priority item is moving browser authentication from `localStorage` tokens to `httpOnly` cookies with CSRF protection.
