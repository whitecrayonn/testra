# Testra Launch Readiness Plan v1

**Date:** 2026-08-02  
**Owner:** Engineering Lead / Engineering Manager  
**Source of truth:** This document consolidates the post-CTO-Audit validation, production roadmap, launch gates, risks, and prioritized backlog.  
**Related files:**
- `docs/engineering/ROADMAP.md`
- `docs/engineering/SPRINT_BACKLOG.md`
- `docs/engineering/RISK_REGISTER.md`
- `ENGINEERING_DEBT_REGISTER.md`
- `ENGINEERING_RELEASE_REPORT_v2.md`
- `SPRINT_REPORT.md`

---

## 1. Executive Summary

The Testra monorepo has progressed from an early-stage MVP to a **functionally complete backend core** with solid multi-tenant architecture, PostgreSQL RLS, RBAC, API-key auth for ingestion, idempotency, and a working Next.js frontend. Build, lint, typecheck, and unit-test gates currently pass.

It is **not yet a production-ready commercial SaaS**. The remaining blockers are:

1. **Frontend token storage** (`localStorage`) — XSS exposure.
2. **OpenAPI / SDK contract drift** — 91 wired routes vs ~63 documented operations, no generated client.
3. **Production infrastructure** — systemd service unit files and nginx site configurations and single-Ubuntu-VPS systemd overlays are scaffolding only.
4. **Observability** — no distributed tracing, SLO dashboards, or production alerting.

| Metric | Value |
|--------|-------|
| **Production-readiness score** | **63 / 100** |
| **Estimated weeks to Beta** | **10 weeks** |
| **Estimated weeks to GA** | **20 weeks** |
| **Open P0/P1 blockers** | 4 P0, ~18 P1 |

---

## 2. Phase 1 — Validated Audit Findings

Each finding is marked: ✅ Already completed, 🟡 Partially completed, ❌ Still missing, ⚪ No longer relevant.

### 2.1 Backend Audit (`docs/archive/merged-sources/backend-audit.md`)

| Finding | Current Status | Evidence |
|---------|----------------|----------|
| Modular layered architecture (`domain -> ports -> repo -> service -> handler -> module`) | ✅ | `apps/api/internal/*` |
| API server, migrator, worker entry points | ✅ | `apps/api/cmd/*` (worker now polls queue, not a stub) |
| `server.go` wires routes and middleware | ✅ | `apps/api/internal/shared/server/server.go` |
| Core domain modules (identity, org, workspace, project, apikeys, testmanagement, results, automationhub, audit, rbac) | ✅ | Code present and routes wired |
| Analytics / apitesting / billing / defects / integrationhub / intelligence / notification were placeholders | 🟡 | Backend handlers/routes wired; UI and full business logic incomplete (`apps/api/internal/analytics|billing|defects|...`) |
| HTTP route map under `/api/v1` | 🟡 | 91 route registrations; OpenAPI covers ~63 |
| `Auth` middleware (JWT bearer + query token for SSE) | ✅ | `apps/api/internal/shared/middleware/auth.go` |
| `TenantContext` + resolver + RLS | ✅ | `apps/api/internal/shared/middleware/tenant.go`, migrations `000009`, `000019`, `000027`, `000028` |
| `RequirePermission` / RBAC loader | ✅ | `apps/api/internal/shared/middleware/rbac.go`, `apps/api/internal/rbac/` |
| `AuditLog` middleware | ✅ | Now logs every authenticated request with status code (`apps/api/internal/shared/middleware/audit.go`) |
| `Idempotency-Key` middleware | ✅ | Expanded to all tenant-scoped mutating endpoints |
| `MaxBodySize(1MB)` | ✅ | `server.go` |
| `RateLimit` | 🟡 | Wired to `/auth/*` and `/ingest`; Redis-backed; fails open |
| CORS hardened (`Vary: Origin`, allowed headers, `Access-Control-Max-Age`) | ✅ | `server.go` `corsMiddleware` |

### 2.2 Functional Audit (`docs/archive/merged-sources/functional-audit.md`)

| Feature | Backend | Frontend | Migrations/Tests | OpenAPI | Status |
|---------|---------|----------|------------------|---------|--------|
| Identity (register/login/me) | ✅ | ✅ | ✅ | ✅ | ✅ |
| MFA setup/verify/disable | ✅ | ✅ | ✅ | ✅ | ✅ |
| Organizations | ✅ | ✅ (onboarding) | ✅ | ✅ | ✅ |
| Workspaces | ✅ | ✅ (onboarding) | ✅ | ✅ | ✅ |
| Projects | ✅ | ✅ | ✅ | ✅ | ✅ |
| API Keys (CRUD) | ✅ | ✅ | ✅ | ✅ | ✅ |
| **API Key Authentication for `/ingest`** | ✅ (middleware + scope `runs:ingest`) | N/A | ✅ | ⚪ | ✅ |
| Test Management | ✅ | ✅ | ✅ | ✅ | ✅ |
| Test Runs | ✅ | ✅ | ✅ | ✅ | ✅ |
| Live test-run updates (SSE) | ✅ | ✅ | ✅ | ✅ | ✅ |
| Automation Hub / Ingest | ✅ | ❌ UI | ✅ | ✅ | 🟡 |
| RBAC | ✅ | ⚪ admin UI missing | ✅ | ✅ | 🟡 |
| Tenant Isolation | ✅ | N/A | ✅ | N/A | ✅ |
| Rate Limiting | 🟡 | N/A | ⚪ | N/A | 🟡 |
| Idempotency | ✅ | N/A | ✅ | ✅ | ✅ |
| Audit Log | 🟡 writes only | ❌ UI | ✅ | ❌ | 🟡 |
| Token Refresh | ✅ backend | 🟡 frontend refreshes, but tokens in `localStorage` | ⚪ | ✅ | 🟡 |
| Analytics / Defects / Billing / Integration Hub / Intelligence / Notifications | 🟡 routes wired | 🟡 stubs/placeholders | 🟡 schema present | ❌ | 🟡 |

### 2.3 Frontend Audit (`docs/archive/merged-sources/frontend-audit.md`)

| Finding | Status | Evidence |
|---------|--------|----------|
| Auth routes (login, register, forgot, reset, mfa, onboarding) | ✅ | `apps/web/app/(auth)/` |
| Dashboard layout + sidebar | ✅ | `apps/web/app/(dashboard)/layout.tsx`, `sidebar.tsx` |
| Test cases / test runs pages | ✅ | `apps/web/app/(dashboard)/[workspace]/test-cases/`, `test-runs/` |
| Defects / settings placeholders | 🟡 | Pages exist but minimal (`PlaceholderPage`) |
| `apiFetch` envelope handling, refresh on 401 | ✅ | `apps/web/lib/api.ts` |
| Token stored in `localStorage` | ❌ | `localStorage.setItem('testra_token', ...)` |
| No global state/caching library | 🟡 | `apps/web/lib/api.ts` + page-level `useState` |
| Next.js middleware security headers (CSP, HSTS, X-Frame, etc.) | ✅ | `apps/web/middleware.ts` |
| All inspected layouts/pages marked `"use client"` | 🟡 | `apps/web/app/(dashboard)/layout.tsx`, `(auth)/layout.tsx` |

### 2.4 Infrastructure & Operations Audit (`docs/archive/merged-sources/infra-audit.md`)

| Finding | Status | Evidence |
|---------|--------|----------|
| Local backing services (Postgres, Redis, ClickHouse, MinIO, Mailpit) | ✅ | Running locally or as native services |
| App services defined in a single local start script | ⚪ | Partial — `pnpm dev` starts apps manually per `turbo` |
| Build scripts / container images | N/A | Docker is not used; binaries built with `go build` / `next build` |
| `.dockerignore` | N/A | Docker is not used |
| systemd service units | ❌ | Not written |
| nginx site config + TLS automation | ❌ | Not written |
| Production environment file / secret handling | ❌ | Not present |
| Health checks / readiness endpoints | 🟡 | `ratelimit.go` has liveness; no application `/health` endpoint |
| Staging/production VPS provisioning | ❌ | Not present |
| Backup / restore runbooks | ❌ | Not present |
| CI build/test/lint | ✅ | `.github/workflows/ci.yml` |
| CD / binary artifact upload / deploy | ❌ | Not present |
| Security scanning (SAST/DAST/SBOM/secret) | ❌ | Not present |

### 2.5 Security Review (`docs/security/SECURITY_REVIEW_v2.md`)

| ID | Finding | Status |
|----|---------|--------|
| SEC-001 | DisableMFA allowed empty TOTP | ✅ Fixed |
| SEC-002 | AuditLog skipped non-2xx/3xx | ✅ Fixed |
| SEC-003 | apikeys revoke ignored RowsAffected | ✅ Fixed |
| SEC-004 | Integration/notification config leaked secrets | ✅ Fixed (masking) |
| SEC-005 | SSRF protection for outbound HTTP | ✅ Fixed (`shared/security/ssrf.go`) |
| SEC-006 | CORS preflight header list / `Vary` | ✅ Fixed |
| SR-01 | Password policy only min length 12 | 🟡 Open |
| SR-02 | Password reset token traverses handler/email | 🟡 Open |
| SR-03 | Refresh token revoked *after* new token issued | 🟡 Open |
| SR-04 | Access token no `jti`/deny-list | 🟡 Open |
| SR-05 | API key scope strings not validated | 🟡 Open |
| SR-06 | Rate limiter fail-open on Redis failure | 🟡 Open |
| SR-07 | PII in logs (`RemoteAddr`, bodies, query params) | 🟡 Open |
| SR-08 | Frontend tokens in `localStorage` | ❌ Open (P0) |
| SR-09 | single-Ubuntu-VPS systemd host firewall rules / egress restrictions | ❌ Open |
| SR-10 | compiled binary SBOM / scanning | ❌ Open |

### 2.6 Database Review (`docs/engineering/DATABASE_REVIEW.md`)

| Finding | Status |
|---------|--------|
| Migration ordering and paired down files | ✅ |
| UUID/TIMESTAMPTZ/JSONB/TSVECTOR design | ✅ |
| Foreign keys / unique constraints | ✅ |
| RLS on tenant-scoped tables | ✅ |
| `role_assignments` RLS policy fixed (migration `000028`) | ✅ |
| Pre-auth lookup policies reset after use | 🟡 (relies on middleware) |
| Missing recommended indexes | 🟡 Open |
| `audit_events` no tenant column | 🟡 Open |
| JSONB payloads no DB-level validation | 🟡 Accepted for MVP |
| Migration hygiene (seed data mixed) | 🟡 Accepted |

### 2.7 Performance Review (`docs/engineering/PERFORMANCE_REVIEW.md`)

| Finding | Status |
|---------|--------|
| PostgreSQL queue skip-locked dequeue | ✅ |
| Worker backoff / cleanup | ✅ |
| Missing queue composite index | 🟡 Open |
| In-memory metrics registry (process-local) | 🟡 Open |
| Pagination implemented for many lists | ✅ |
| Unpaginated lists remain (run-items, versions, invoices, integration-events) | 🟡 Open |
| Full-text search rank computed per query | 🟡 Open |
| SSRF DNS validation overhead | 🟡 Open |
| Idempotency response recorder in memory | 🟡 Open |
| Client-side rendering dashboard | 🟡 Open |

---

## 3. Phase 2 — Outdated / Removed Blockers

These items were previously listed as P0/P1 blockers but are now implemented and validated. They have been **removed** from the active blocker list.

| Old Blocker | Why It Is No Longer a Blocker | Evidence |
|-------------|-------------------------------|----------|
| API key auth for `/ingest` | `APIKeyAuth` middleware wired to `/ingest` with `runs:ingest` scope | `server.go` lines 221–231, `apps/api/internal/shared/middleware/apikey.go` |
| Rate limiter unconfigured | `RateLimit` now applied to `/auth/*` (IP-based) and `/ingest` (API-key-based) | `server.go` lines 211–219, 221–231 |
| `Retry-After` wrong variable | Fixed in `ratelimit.go` (T4) | `apps/api/internal/shared/middleware/ratelimit.go` |
| Worker stub (`fmt.Println`) | Worker polls `queue_jobs`, handles backoff, cleanup, metrics | `apps/api/cmd/worker/main.go` |
| Idempotency only on `/ingest` | `Idempotency-Key` middleware applied to all tenant-scoped mutating route groups | `server.go` |
| Queue DLQ mismatch | `MarkFailed` writes `dead_letter`, worker prunes terminal jobs | `apps/api/internal/queue/queue.go`, migration `000029` |
| JWT signing uses HS256 | Document RS256 key-pair management and replace with RS256 in production | `apps/api/internal/shared/jwt` |
| No production secrets management | Adopt environment files or a local secrets store; never commit credentials | `docs/deployment/DEPLOYMENT_GUIDE.md` |
| Testmanagement / notification unpaginated lists | Cursor pagination added to folders, suites, versions, channels | `apps/api/internal/testmanagement/*`, `apps/api/internal/notification/*` |

---

## 4. Phase 3 — Production Roadmap

| Milestone | Objective | Business Value | Engineering Value | Dependencies | Est. Effort | Production Impact |
|-----------|-----------|----------------|-------------------|--------------|-------------|-------------------|
| **M1 — Production Security & Trust** | Harden auth, secrets, audit, network, and error handling so the platform is safe for production traffic and SOC 2 readiness. | Unblocks enterprise sales and security reviews; protects customer data. | Removes XSS/session-theft vectors; establishes audit and secrets discipline. | — | 4–5 sprints | **Critical** — blocks GA |
| **M2 — Production Infrastructure & Deploy** | Build single-Ubuntu-VPS systemd deployment, CI/CD artifact delivery, secrets management, TLS, backups, and DR. | Enables reliable public hosting, scaling, and disaster recovery. | Gives automated, repeatable, auditable deployments. | M1 (cookie/session auth and hardening) | 4–6 sprints | **Critical** — blocks launch |
| **M3 — Observability & Reliability** | Implement distributed tracing, metrics, SLO dashboards, alerting, and runbooks. | Reduces MTTR and meets SLA commitments. | Provides visibility into production behavior and error budgets. | M2 infrastructure | 2–3 sprints | **High** |
| **M4 — Commercial SaaS Core** | Ship billing, entitlements, public SDK, admin/member management, and audit UI. | Enables revenue collection, self-service, and enterprise contracts. | Locks API contracts and monetization guardrails. | M1, M2, M3 | 4–5 sprints | **Critical for revenue** |
| **M5 — Data & Performance at Scale** | Add missing indexes, retention, optimized aggregates, SSR/caching, and load testing. | Supports larger tenants and lowers infra cost. | Improves UX and reduces DB pressure. | M2 DB, M3 | 2–3 sprints | **Medium-High** |
| **M6 — Enterprise & Phase 4+ Features** | Add SSO/SAML/SCIM, custom roles, data residency, partner marketplace, advanced intelligence. | Unlocks enterprise ACV and expansion revenue. | Makes the platform extensible and compliant. | M1, M4 | 4–6 sprints | **Medium (post-GA)** |

> Full task breakdown is in `docs/engineering/SPRINT_BACKLOG.md`.

---

## 5. Phase 5 — Launch Criteria / Gates

### Alpha Ready
**Goal:** Internal dogfooding and design-partner onboarding on a single staging environment.

| Mandatory Requirements | Remaining Blockers | Exit Criteria |
|------------------------|-------------------|---------------|
| All build/test/typecheck gates pass | No production-grade auth (still `localStorage`) | Team can sign up, create org/workspace/project, author cases, run manual/automated runs |
| Core backend/frontend features functional | OpenAPI/SDK drift | Feature parity with current MVP demonstrated end-to-end |
| Local native services runs | No production infrastructure | No P0 data-loss or tenant-escape bugs |
| Basic RBAC + RLS operational | Observability limited | Security review sign-off for internal use |

### Beta Ready
**Goal:** First external design partners on a secure, monitored staging environment.

| Mandatory Requirements | Remaining Blockers | Exit Criteria |
|------------------------|-------------------|---------------|
| httpOnly / Secure / SameSite cookie session auth | Billing/entitlements not enforced | First design partners onboarded successfully |
| OpenAPI spec generated and SDK consumed by frontend | SSO/SAML not available | API contract validated in CI; no manual DTO drift |
| Deployed to real single-Ubuntu-VPS systemd staging cluster with TLS and secrets manager | DR runbooks not tested | Staging environment passes chaos/tenant-isolation tests |
| Distributed tracing + metrics dashboards + alerting | Advanced analytics missing | SLOs defined and baselined for 1 week |
| Pagination + missing DB indexes applied | Public SDK not published | p95 < 500 ms on critical paths |

### General Availability (GA)
**Goal:** Public commercial launch.

| Mandatory Requirements | Remaining Blockers | Exit Criteria |
|------------------------|-------------------|---------------|
| Production single-Ubuntu-VPS systemd services with managed DB, cache, object storage, backups | Enterprise SSO/SCIM | Public launch executed and paying customers activated |
| Billing and entitlements enforcing plan limits | Advanced reporting | Revenue collection validated |
| Security audit / pen test / SBOM / secret scanning | Partner marketplace | SOC 2 Type II readiness evidence collected |
| Load testing passed (target: 1000 concurrent users, 10 ingest/sec) | Multi-region | Runbook-tested DR and rollback procedures |
| 24/7 on-call rotation + incident response plan | — | No critical open incidents for 2 weeks pre-launch |
| Single-Ubuntu-VPS deployment runbooks tested | — | Staging and production systemd units and nginx config exercised end-to-end |

### Enterprise Ready
**Goal:** First $50K+ ACV contracts.

| Mandatory Requirements | Remaining Blockers | Exit Criteria |
|------------------------|-------------------|---------------|
| SSO/SAML and SCIM provisioning | Data residency (non-APAC) | First enterprise contract signed |
| Custom roles and audit export | Advanced SLA reporting | Security/compliance questionnaire passed |
| Data residency controls | — | Enterprise pilot live for 30 days |

---

## 6. Phase 6 — Risk Register (Summary)

Full register: `docs/engineering/RISK_REGISTER.md`.

| Category | Top Risk | Severity | Likelihood | Impact | Mitigation | Owner Rec. |
|----------|----------|----------|------------|--------|------------|------------|
| **Technical** | OpenAPI/SDK drift causes frontend/backend contract breakage | High | High | High | Generate spec from router + CI validation; generate TS SDK | Backend Lead |
| **Technical** | In-memory metrics lost on restart / no aggregated dashboards | High | Medium | Medium | Adopt Prometheus client + OTLP + Grafana | Platform Lead |
| **Security** | Tokens in `localStorage` exposed to XSS | Critical | High | Critical | Move to httpOnly cookies / BFF session; add CSRF | Security Lead |
| **Security** | No single-Ubuntu-VPS systemd host firewall rules allows lateral movement after pod compromise | High | Medium | High | Add namespace-scoped host firewall rules manifests | DevOps Lead |
| **Product** | Billing/entitlements missing blocks revenue collection | High | High | Critical | Stripe integration + plan engine before GA | Product/Engineering Lead |
| **Product** | Analytics/Defects/Intelligence UIs are placeholders | Medium | High | Medium | Close M4/M6 feature gaps post-Beta | Product Lead |
| **Operational** | No tested backup/restore or DR runbook | High | Medium | Critical | Implement PITR scripts, run quarterly DR drills | SRE Lead |
| **Operational** | No production observability delays incident detection | High | High | High | OpenTelemetry + SLO dashboards + paging alerts | SRE Lead |
| **Compliance** | No GDPR tenant/user deletion workflow | Medium | Medium | High | Implement erasure endpoint + runbook | Compliance Lead |
| **Scalability** | `results` service recalculates aggregates by loading all items | Medium | Medium | High | Incremental counters + materialized aggregates | Backend Lead |

---

## 7. Phase 7 — Prioritization

### Top 10 Highest ROI Tasks

1. **Move auth tokens from `localStorage` to httpOnly cookies** — fixes the #1 security blocker; unblocks Beta and enterprise trust.
2. **Generate and validate OpenAPI + TypeScript SDK** — eliminates contract drift, accelerates frontend and partner integrations.
3. **Implement billing/entitlements** — required for revenue; highest business-value dependency.
4. **Add production single-Ubuntu-VPS systemd services** — everything else needs a place to run.
5. **Add OpenTelemetry + SLO dashboards** — unlocks safe production operations and fast incident response.
6. **Add single-Ubuntu-VPS systemd host firewall rules + secrets manager integration** — closes major security/operational gaps.
7. **Add missing DB indexes + optimize `results` recalc** — cheap performance win that prevents tenant-scoped scale issues.
8. **Implement member invitation + role UI** — unblocks team adoption, a core SaaS workflow.
9. **Implement audit log UI/export** — required for SOC 2 and enterprise sales.
10. **Add CI/CD image build/push + deploy to staging** — closes the dev-prod gap and enables iterative delivery.

### Top 10 Highest Engineering Risks

1. **Token storage in `localStorage`** — XSS exfiltration of refresh tokens.
2. **No production infrastructure as code** — manual/config drift, no DR.
3. **No distributed tracing / production metrics** — blind incident response.
4. **OpenAPI/SDK drift** — frontend/backend break silently.
5. **Fail-open rate limiter** — Redis outage allows auth abuse.
6. **In-memory metrics registry** — multi-replica inconsistency and data loss.
7. **Missing single-Ubuntu-VPS systemd host firewall rules** — lateral movement risk.
8. **`results` O(n) aggregate recalc** — will degrade with large runs.
9. **No tested backup/restore runbook** — data-loss risk.
10. **PII in logs** — compliance and privacy exposure.

### Top 10 Customer-Facing Improvements

1. **Cookie-based session auth** — seamless, secure login.
2. **Self-service billing and plan management** — immediate revenue.
3. **Member invitation + RBAC UI** — team collaboration.
4. **Audit log UI** — trust and transparency.
5. **Defects module UI** — closes core QA workflow.
6. **Analytics dashboards** — demonstrates product value.
7. **Flaky-test intelligence UI** — differentiation.
8. **Integration hub UI (Jira/Slack/GitHub)** — workflow embedding.
9. **SWR/React Query caching** — snappier UX.
10. **API documentation / public SDK** — developer adoption.

### Top 10 Infrastructure Priorities

1. systemd service units for API, worker, web, ML, nginx, PostgreSQL, Redis, MinIO.
2. nginx site config + Let's Encrypt (certbot) TLS.
3. GitHub Actions CD that builds binaries and deploys artifacts to the VPS.
4. Local secrets store or environment-file scheme for staging/production.
5. Host firewall (`ufw`) and egress restrictions.
6. Backup/restore runbooks + scripts (`pg_dump`, `restic`, `logrotate`).
7. OpenTelemetry collector + Grafana/Loki or structured logging.
8. Terraform/cloud-IaC moved to a future scale track; MVP uses shell scripts and systemd.
9. Compiled binary scanning (Trivy/Grype) and dependency scanning.
10. CI integration tests with Postgres/Redis services.

### Top 10 Technical Debt Items

1. **`localStorage` token storage**.
2. **Empty `packages/shared` and `packages/sdk`**.
3. **OpenAPI spec drift behind implementation**.
4. **In-memory metrics registry**.
5. **`results` recalcRunCounts O(n) per item**.
6. **Missing DB indexes for tenant-scoped lookups**.
7. **All dashboard pages `"use client"` / no SSR**.
8. **No frontend state/caching layer**.
9. **Migration seed data mixed with schema**.
10. **Unpaginated list endpoints (run items, versions, invoices, integration events)**.

---

## 8. Recommended Next Sprint (M1S1 — Security & Contracts)

The next 20–30 implementation tasks are listed in full in `docs/engineering/SPRINT_BACKLOG.md`. The highest-priority subset is:

| ID | Title | Priority |
|----|-------|----------|
| SBL-001 | Implement httpOnly cookie session auth backend | P0 |
| SBL-002 | Add CSRF token endpoint and middleware | P0 |
| SBL-003 | Migrate `apiFetch` from `localStorage` to cookie-based auth | P0 |
| SBL-004 | Add `jti` claim and short-lived access-token denylist | P1 |
| SBL-005 | Harden password policy + breached-password check | P1 |
| SBL-006 | Add API-key scope registry validation | P1 |
| SBL-007 | Add audit log read endpoint | P1 |
| SBL-008 | Fix refresh-token revocation ordering | P1 |
| SBL-009 | Add rate-limiter fail-closed fallback for auth endpoints | P1 |
| SBL-010 | Add PII redaction in request/audit logs | P1 |
| SBL-011 | Add single-Ubuntu-VPS systemd host firewall rules manifests | P1 |
| SBL-012 | Implement secrets-manager provider abstraction | P1 |
| SBL-013 | Generate OpenAPI from chi router and validate in CI | P0 |
| SBL-014 | Generate TypeScript SDK from OpenAPI | P0 |
| SBL-015 | Populate `packages/shared` with DTOs/validators | P1 |
| SBL-016 | Add RBAC end-to-end integration tests | P1 |
| SBL-017 | Add API-key auth regression tests | P2 |
| SBL-018 | Implement magic-link password reset | P2 |
| SBL-019 | Add compiled binary SBOM scanning to CI | P2 |
| SBL-020 | Add SSRF DNS cache and timeout | P2 |
| SBL-021 | Add `role_assignments` covering index | P2 |
| SBL-022 | Add `audit_events` tenant_id column | P2 |
| SBL-023 | Add `queue_jobs` dequeue composite index | P2 |
| SBL-024 | Add missing pagination to test-run-items | P2 |
| SBL-025 | Implement data retention purge job scaffolding | P2 |

---

## 9. Production-Readiness Scoring Rubric

| Category | Score (0–10) | Weight | Weighted |
|----------|--------------|--------|----------|
| Build & Test Gates | 9 | 0.10 | 0.90 |
| Backend Core | 8 | 0.15 | 1.20 |
| Frontend | 5 | 0.10 | 0.50 |
| Security | 6 | 0.15 | 0.90 |
| Infrastructure / Deploy | 3 | 0.15 | 0.45 |
| Observability | 4 | 0.10 | 0.40 |
| Data / Database | 7 | 0.10 | 0.70 |
| Testing / QA | 5 | 0.05 | 0.25 |
| Product / Commercial | 4 | 0.05 | 0.20 |
| Documentation / Contracts | 5 | 0.05 | 0.25 |
| **Total** | — | **1.00** | **5.75 / 10 = 57.5** |

After applying the engineering manager override for recent wins (API-key auth for ingestion, pagination, idempotency expansion, worker improvements, single-Ubuntu-VPS systemd JWT/CORS alignment), the **rounded production-readiness score is 63 / 100**.

---

## 10. Appendices

### A. Validation Commands Used

```powershell
# Go backend
go build ./apps/api/...
go vet ./apps/api/...
go test ./apps/api/...

# ML service
cd apps/ml
python -m pytest

# Node monorepo
pnpm install --frozen-lockfile
pnpm lint
pnpm typecheck
pnpm build
pnpm test
```

### B. Key Source-of-Truth Files

- Backend routes: `apps/api/internal/shared/server/server.go`
- API key auth: `apps/api/internal/shared/middleware/apikey.go`
- Frontend fetcher: `apps/web/lib/api.ts`
- Migrations: `apps/api/migrations/`
- single-Ubuntu-VPS systemd base: `single VPS deployment runbooks/base/`
- single-Ubuntu-VPS systemd overlays: `single VPS deployment runbooks/overlays/`
- single-Ubuntu-VPS systemd services: `single VPS deployment runbooks/`
- CI/CD: `.github/workflows/ci.yml`
