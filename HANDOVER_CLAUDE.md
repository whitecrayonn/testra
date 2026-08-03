# Testra — Full Handover Document for Claude (Reviewer)

> **Purpose:** Complete knowledge transfer for an AI reviewer (Claude) to understand the Testra codebase, business context, architecture, current phase, and ongoing work. This document consolidates information from `BIBLICAL_TESTRA.md`, `ROADMAP.md`, `FEATURE_MATRIX.md`, `ENGINEERING_DEBT_REGISTER.md`, and the current debugging session.

---

## 1. Product & Business Context

### What is Testra?

**Testra** is a unified quality engineering platform — *"One Platform. Every Test."*

- **Mission:** One platform for manual, automated, and API testing.
- **Vision:** Become the APAC-first enterprise-ready test operations platform.
- **North Star:** Reduce time spent switching between QA tools and provide trustworthy, explainable intelligence from the customer's own data.

### Target Market

- **Primary ICP:** Mid-market to enterprise SaaS companies in APAC that need governance, audit, and multi-tenancy.
- **Business model:** B2B SaaS (billing planned for Phase 5+, not yet implemented).

### Core Philosophies

1. **One Platform** — manual, automated, and API testing in one place
2. **Enterprise Ready** — multi-tenancy, RBAC, audit, RLS from day one
3. **Automation First** — CI/CD ingestion, API keys, webhook integrations
4. **No External LLM** — transparent classical ML only (scikit-learn, XGBoost)
5. **Customer Owns Data** — zero code retention, tenant-isolated ML models
6. **API First** — OpenAPI 3.1 contracts drive all consumers
7. **Localization Ready** — APAC-first design

### MVP Scope (Phases 0–3.5, completed)

Self-hosted identity and tenancy, RBAC, API keys, test management, manual and CI result runs, in-app notifications, polished web dashboard, stable local developer workflow.

### Out of Scope for MVP

Built-in CI runner, source-code hosting, external LLM features, billing, WorkOS SSO.

---

## 2. Technology Stack

| Layer | Technology | Status |
|-------|------------|--------|
| Backend runtime | Go 1.24+ | Implemented |
| HTTP router | chi/v5 | Implemented |
| Frontend framework | Next.js 15+ (App Router) | Implemented |
| Frontend language | TypeScript 5+ | Implemented |
| Styling | TailwindCSS + shadcn/ui + Radix UI | Implemented |
| Forms | React Hook Form + Zod | Implemented |
| Primary database | PostgreSQL 16+ | Implemented |
| Analytics database | ClickHouse 24+ | Planned (deferred per ADR-010) |
| Cache / queue | Redis 7+ | Implemented (skeleton), queues Planned |
| Object storage | S3-compatible (local MinIO) | Implemented (config) |
| Background jobs | Asynq over Redis | Planned |
| Real-time | Server-Sent Events (SSE) | Implemented |
| ML runtime | Python 3.12+ FastAPI | Skeleton implemented |
| ML libraries | scikit-learn, XGBoost | Planned |
| API documentation | OpenAPI 3.1 + Scalar | Approved |
| Package management | pnpm + Go modules + go.work | Implemented |
| CI/CD | GitHub Actions | Implemented |
| Local development | Native services (no Docker) | Implemented (ADR-009) |
| Containerization | Not used | Rejected for MVP |
| Production deployment | Single Ubuntu VPS + systemd + nginx | Planned |

---

## 3. Repository Structure

```
c:/Private/project/
├── testra/                           # Application monorepo
│   ├── apps/
│   │   ├── api/                      # Go backend (modular monolith)
│   │   │   ├── cmd/                  # Entrypoints: api, migrator, worker
│   │   │   ├── internal/             # Private implementation
│   │   │   │   ├── shared/           # Cross-cutting: config, db, errors, eventbus, http, idempotency, jwt (RS256), middleware (auth, tenant, rbac, audit, idempotency, ratelimit, csrf, cookies, redact, logger, maxbody), pagination, password, secrets, security (SSRF), server, tenant, validation
│   │   │   │   ├── identity/         # Auth, MFA, password reset, JWT
│   │   │   │   ├── organization/     # Org CRUD
│   │   │   │   ├── workspace/        # Workspace CRUD
│   │   │   │   ├── project/          # Project CRUD
│   │   │   │   ├── apikeys/          # Scoped API keys
│   │   │   │   ├── rbac/             # Roles, permissions, SQL loader
│   │   │   │   ├── testmanagement/   # Folders, suites, cases, versioning, search
│   │   │   │   ├── results/          # Test runs, run items, SSE progress
│   │   │   │   ├── automationhub/    # JUnit/Playwright/Cypress ingestion
│   │   │   │   ├── apitesting/       # API test definitions, execution
│   │   │   │   ├── defects/          # Defect tracking CRUD
│   │   │   │   ├── notification/     # In-app notifications, channels, preferences
│   │   │   │   ├── audit/            # Immutable audit events
│   │   │   │   ├── analytics/        # Dashboard aggregates (routes wired)
│   │   │   │   ├── intelligence/     # ML flaky detection, risk scores (routes wired, rule-based)
│   │   │   │   ├── integrationhub/   # Jira/GitHub/GitLab sync (routes wired)
│   │   │   │   ├── billing/          # Subscriptions (routes wired, Stripe stub)
│   │   │   │   ├── integration/      # Tenant isolation integration tests
│   │   │   │   ├── search/           # Global search
│   │   │   │   ├── queue/            # Background job queue (dead-letter, visibility timeout)
│   │   │   │   └── metrics/          # Prometheus metrics
│   │   │   ├── migrations/           # 40 migration pairs (golang-migrate)
│   │   │   └── storage/              # File storage (automation artifacts, runtime created)
│   │   ├── web/                      # Next.js 15 App Router frontend
│   │   │   ├── app/
│   │   │   │   ├── (auth)/           # login, register, forgot-password, reset-password, mfa-setup
│   │   │   │   ├── (dashboard)/      # dashboard, [workspace]/*, settings/*
│   │   │   │   ├── onboarding/       # First-run org + workspace creation
│   │   │   │   ├── layout.tsx        # Root layout
│   │   │   │   ├── error.tsx         # Global error boundary
│   │   │   │   ├── loading.tsx       # Global loading state
│   │   │   │   └── not-found.tsx     # 404 page
│   │   │   ├── components/           # UI components, global-search, command-palette, providers
│   │   │   ├── lib/                  # api.ts (fetch client), hooks, utils
│   │   │   └── middleware.ts         # Security headers, CSP
│   │   ├── ml/                       # Python FastAPI ML service (skeleton)
│   │   └── worker/                   # Reserved (worker now runs from apps/api/cmd/worker)
│   ├── packages/
│   │   ├── shared/                   # Shared TS types (placeholder)
│   │   ├── ui/                       # Shared React components (placeholder)
│   │   ├── config/                   # Shared tooling configs
│   │   └── sdk/                      # Generated TS SDK (placeholder)
│   ├── tests/                        # Playwright E2E test suite
│   │   ├── e2e/                      # Test specs (23 test directories)
│   │   │   ├── accessibility/        # a11y tests
│   │   │   ├── analytics/            # Analytics dashboard tests
│   │   │   ├── api-testing/          # API testing module tests
│   │   │   ├── auth/                 # Authentication tests
│   │   │   ├── automation/           # Automation hub tests
│   │   │   ├── crud/                 # CRUD operations tests
│   │   │   ├── dashboard/            # Dashboard tests
│   │   │   ├── defects/              # Defects tests
│   │   │   ├── file-upload/          # File upload tests
│   │   │   ├── integrations/         # Integration hub tests
│   │   │   ├── manual-run/           # Manual test execution tests
│   │   │   ├── notifications/        # Notification tests
│   │   │   ├── onboarding/           # Onboarding flow tests
│   │   │   ├── pagination/           # Pagination tests
│   │   │   ├── permissions/          # RBAC/permissions tests
│   │   │   ├── projects/             # Project management tests
│   │   │   ├── search/               # Global search tests
│   │   │   ├── smoke/                # Smoke tests
│   │   │   ├── tenant/               # Tenant isolation tests
│   │   │   ├── testcases/            # Test case tests
│   │   │   ├── testplans/            # Test plan tests
│   │   │   ├── validation/           # Input validation tests
│   │   │   └── visual/               # Visual regression tests
│   │   ├── pages/                    # Page Object Models (LoginPage, DashboardPage, etc.)
│   │   ├── factories/                # Test data factories (User, Project, TestCase, etc.)
│   │   ├── fixtures/                 # Playwright fixtures (auth, api, page)
│   │   ├── helpers/                  # API helper, assertions, CSV, pagination, random, storage, wait
│   │   ├── constants/                # Routes, roles, selectors
│   │   ├── config/                   # Test data, test config
│   │   └── playwright.config.ts      # Playwright configuration
│   ├── docs/                         # OpenAPI, ADRs, runbooks, architecture docs
│   ├── scripts/                      # Dev automation scripts
│   ├── .github/workflows/            # CI pipeline (ci.yml)
│   ├── Makefile                      # Common dev tasks
│   ├── pnpm-workspace.yaml           # JS workspace definition
│   ├── go.work                       # Go workspace definition
│   ├── .env.example                  # Local environment template
│   ├── ENGINEERING_DEBT_REGISTER.md  # Technical debt tracking
│   ├── ENGINEERING_RELEASE_REPORT_v2.md
│   ├── EXECUTIVE_ENGINEERING_REPORT.md
│   ├── FINAL_LOCAL_ENGINEERING_REPORT.md
│   ├── PRODUCTION_QUALITY_REPORT.md
│   ├── SPRINT_REPORT.md
│   └── README.md                     # Project README
├── testra-master-context.md          # Master product context (outside repo)
├── testra-product-strategy.md        # Product strategy
├── testra-product-architecture-strategy.md
├── testra-product-discovery.md       # Market and persona research
└── testra-brd.md                     # Business requirements
```

---

## 4. Architecture

### System Architecture

Testra is a **modular monolith**. A single Go process serves the API with Clean Architecture boundaries. It communicates with a Next.js frontend, Python ML service, PostgreSQL, Redis, and S3-compatible storage.

```
Browser → Next.js → Nginx → Go API → PostgreSQL (RLS-enforced)
                                    → Redis (cache/queue)
                                    → S3 (storage)
                                    → ML Service (optional, future)
CI/CD → Go API (via API key) → AutomationHub → Results
```

### Backend Clean Architecture

Every backend module follows Clean/Hexagonal Architecture:

| Layer | Responsibility | Example |
|-------|----------------|---------|
| Domain | Entities, value objects, invariants | `internal/results/domain.go` |
| Application | Use cases, service orchestration | `internal/results/service.go` |
| Ports | Interfaces for repositories | `internal/results/ports.go` |
| Adapters | HTTP handlers, SQL repositories | `internal/results/handler.go`, `repository.go` |

### Shared Cross-Cutting Packages

- `shared/config` — environment configuration + production validation (JWT keys, DB SSL, credentials)
- `shared/db` — database wrapper, tenant context propagation
- `shared/errors` — domain error constants
- `shared/eventbus` — in-process event bus for cross-module communication (256 buffer)
- `shared/http` — response envelope helpers
- `shared/idempotency` — PostgreSQL-backed idempotency store (org-scoped)
- `shared/jwt` — RS256 JWT signing/parsing, RSA key pairs, key rotation, JWKS marshaling, `kid` header support
- `shared/middleware` — auth, API key auth, tenant context, RBAC, audit, idempotency, rate limit (Redis+local), CSRF, max body, cookies, request logger (structured slog), redact, context helpers
- `shared/pagination` — cursor pagination helpers
- `shared/password` — password hashing (bcrypt)
- `shared/secrets` — secrets provider abstraction (env-based, swappable)
- `shared/security` — SSRF protection (URL validation)
- `shared/server` — chi router wiring (1392 lines, all route groups)
- `shared/tenant` — tenant resolver (resolves org from workspace, project, API key, run, run item, defect, automation project, integration, notification, etc.)
- `shared/validation` — email/name validation

### Domain Modules

| Module | Status | Capabilities |
|--------|--------|--------------|
| `identity` | Implemented | Register, login, refresh, logout, logout-all-devices, password reset, TOTP MFA (setup/verify/disable), /me, CSRF token |
| `organization` | Implemented | Create, list, get |
| `workspace` | Implemented | Create, list, get, membership |
| `project` | Implemented | Create, list, get, update, delete |
| `apikeys` | Implemented | Create, list, revoke scoped API keys |
| `rbac` | Implemented | Roles, permissions, role assignments, SQL loader |
| `testmanagement` | Implemented | Folders CRUD, suites CRUD, cases CRUD, versioning, full-text search (tsvector) |
| `results` | Implemented | Test runs, run items, status updates, SSE progress, bulk update, clone, rerun, evidence attach/list/delete, defect link/unlink, test plans, run-from-plan, item execution, item history |
| `automationhub` | Implemented | JUnit XML and Playwright/Cypress JSON ingestion, automation projects, executions, artifacts, logs |
| `apitesting` | Implemented | API test collections, folders, requests, environments, execution engine, history |
| `defects` | Implemented | CRUD, lifecycle, severity/priority, tenant isolation |
| `notification` | Implemented | In-app notifications, preferences, email/Slack/Teams/webhook channels, templates, delivery history |
| `audit` | Implemented | Immutable audit events on mutating endpoints |
| `analytics` | Wired (routes active) | Dashboards CRUD, summary, trends, metrics, recent-activity, CSV export |
| `intelligence` | Wired (routes active, ML client returns rule-based) | Predict flaky, classify failure, list predictions/clusters |
| `integrationhub` | Wired (routes active) | Integrations CRUD, enable/disable, events, dead-letter, incoming webhooks, health checks |
| `billing` | Wired (routes active, Stripe stub) | Subscription update/get, invoices list |
| `queue` | Implemented | Background job queue with dead-letter, visibility timeout, poison pill handling |
| `metrics` | Implemented | Prometheus metrics instrumentation for API and worker |
| `search` | Implemented | Global search across test cases, runs, defects |
| `integration` | Test-only | Tenant isolation integration tests |

### Module Dependency Matrix

```
Identity → Shared
Organization → Identity
Workspace → Organization
Project → Workspace
RBAC → Identity, Organization
APIKeys → Workspace
TestManagement → Project
Results → Project, TestManagement
AutomationHub → Results
APITesting → Project
Defects → Results
Analytics → Results
Intelligence → Analytics
Notification → Identity, Workspace
Audit → cross-cutting (Identity, Organization, Workspace)
IntegrationHub → Project, Results
```

### Request Lifecycle (Middleware Order)

1. `RequestID` (chi middleware, echoed in response header)
2. `RealIP` (chi middleware)
3. `Recoverer` (chi middleware)
4. `RequestLogger` (structured `slog` JSON handler, not chi text logger)
5. `Content-Type: application/json` (chi SetHeader)
6. CORS middleware (origin-allowlisted, `Vary: Origin`)
7. `MaxBodySize` (default limit)
8. `apiSecurityHeaders` (Cache-Control, X-Content-Type-Options, X-Frame-Options, Referrer-Policy, Permissions-Policy)
9. Per-route-group middleware:
   - `RateLimit` (Redis-backed with local fallback; auth: 20/min per IP, API key: 100/min per key)
   - `CSRF` (skips login/register/refresh/password-reset; token from `/api/v1/auth/csrf`)
   - `Auth` (JWT RS256) or `APIKeyAuth` on protected routes
   - `TenantContext` (per-route group) — resolves org from URL param, body, or query; sets `app.tenant_id`; verifies membership
   - `RequirePermission` (per handler, via RBAC SQL loader)
   - `AuditLog` (on mutating handlers, fire-and-forget with 5s timeout)
   - `IdempotencyKey` (on all mutating tenant-scoped endpoints)
10. Handler → Service → Repository

### Multi-Tenancy

- **Hierarchy:** Organization → Workspace → Project
- **Isolation:** PostgreSQL Row Level Security (RLS) using `current_setting('app.tenant_id')`
- **Tenant resolver:** Derives `organization_id` from workspace_id, project_id, api_key_id, run_id, run_item_id, test_case_id, test_folder_id, test_suite_id, defect_id, automation_project_id, automation_execution_id, automation_artifact_id, api_collection_id, api_folder_id, api_environment_id, api_request_id, api_execution_id, notification_id, notification_channel_id, notification_template_id, integration_id, integration_event_id, or default org for user
- **Critical rule:** No repository call may run without `app.tenant_id` set. No handler may trust client-supplied tenant IDs.

### Authentication

| Method | Use case | Status |
|--------|----------|--------|
| Email + password + TOTP MFA | Human users | Implemented |
| JWT access token (RS256, 15 min, JWKS + key rotation) | API requests | Implemented |
| Rotating refresh token | Long-lived sessions | Implemented |
| API key (SHA-256 hashed) | CI/CD | Implemented |
| OAuth / SSO | Enterprise | Planned (Phase 6) |

### Frontend Architecture

- **Framework:** Next.js 15 App Router, React, TypeScript
- **Styling:** TailwindCSS, shadcn/ui + Radix primitives
- **State:** Server state via `lib/api.ts` fetch client; forms via React Hook Form + Zod
- **Token storage:** HttpOnly cookies set by the API (`shared/middleware/cookies.go`); `apps/web/lib/api.ts` sends `credentials: "include"`. `localStorage` holds only non-secret UI state (`testra_workspace_id`, `testra_organization_id`, `testra_project_id`). See `docs/audits/P0_VERIFICATION_REPORT.md` §6.
- **Route groups:**
  - `(auth)` — login, register, forgot-password, reset-password, MFA setup
  - `(dashboard)` — dashboard, projects, test-cases, test-runs, settings, api-tests, defects, automation
  - `onboarding` — first-run org + workspace creation

### API Design Conventions

- Base path: `/api/v1`
- JWKS endpoint: `/.well-known/jwks.json` (public, exposes RSA public keys for JWT verification)
- Health endpoint: `/health` (public, checks DB connectivity)
- CSRF: `/api/v1/auth/csrf` returns CSRF token; required on all mutating auth endpoints except login/register/refresh/password-reset
- Public webhook: `/api/v1/webhooks/{provider}/{integration_id}` (no auth, signature verified by provider)
- Response envelope: `{ "data": ..., "meta": ..., "error": null }`
- Error shape: `{ "data": null, "meta": {}, "error": { "code": "...", "message": "..." } }`
- Cursor pagination (`?cursor=` and `?limit=`)
- `Idempotency-Key` header required for all mutating tenant-scoped endpoints (org-scoped, PostgreSQL-backed)
- `snake_case` JSON fields, RFC 3339 UTC timestamps
- Stable error codes: `INVALID_INPUT`, `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`, `CONFLICT`, `INTERNAL_ERROR`
- Security headers on all `/api/v1` responses: `Cache-Control: no-store`, `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: strict-origin-when-cross-origin`, `Permissions-Policy: camera=(), microphone=(), geolocation=()`
- Rate limiting: Redis-backed with local fallback; auth endpoints 20/min per IP, API key ingestion 100/min per key

### Database Schema

40 migration pairs (000001–000040) covering:

| Group | Tables |
|-------|--------|
| Identity | `users` |
| Session | `refresh_tokens`, `password_reset_tokens` |
| Tenancy | `organizations`, `organization_members`, `workspaces`, `workspace_members`, `projects` |
| RBAC | `roles`, `permissions`, `role_permissions`, `role_assignments` |
| API Keys | `api_keys` |
| Test Management | `test_folders`, `test_suites`, `test_cases`, `test_case_versions` |
| Execution | `test_runs`, `test_run_items` |
| Idempotency | `idempotency_records` |
| Notifications | `notifications`, `notification_preferences`, `notification_channels` |
| Audit | `audit_events` |
| Defects | `defects` |
| API Testing | `api_collections`, `api_folders`, `api_environments`, `api_requests`, `api_executions` |
| Automation Hub | `automation_projects`, `automation_runs` |
| Worker Queue | `queue_jobs` |
| Analytics | `analytics_snapshots` |
| Intelligence | `intelligence_scores` |
| Integration Hub | `integrations`, `integration_sync_logs` |
| Billing | `subscriptions`, `invoices` |

RLS enabled on all tenant-scoped tables. Tables without RLS: `users`, `roles`, `permissions`, `role_permissions`, `refresh_tokens`, `password_reset_tokens`.

---

## 5. Development Workflow

### Local Development (Native, no Docker)

```bash
pnpm install      # installs JS deps + auto-creates Python venv for ML
pnpm dev          # starts everything (API, Web, Worker, ML)
```

`pnpm dev` will:
1. Check local services (PostgreSQL, Redis)
2. Run database migrations
3. Launch all apps via Turborepo

### Service Ports

| Service | Port | URL |
|---------|------|-----|
| Go API | 8080 | http://localhost:8080 |
| Next.js Web | 3000 | http://localhost:3000 |
| ML Service | 8000 | http://localhost:8000 |
| PostgreSQL | 5432 | localhost:5432 |
| Redis | 6379 | localhost:6379 |
| MinIO S3 | 9002 | http://localhost:9002 |
| MinIO Console | 9001 | http://localhost:9001 |
| Mailpit SMTP | 1025 | localhost:1025 |
| Mailpit UI | 8025 | http://localhost:8025 |

### Common Commands

| Command | Description |
|---------|-------------|
| `pnpm dev` | Start all apps |
| `pnpm build` | Build all apps |
| `pnpm test` | Run all tests |
| `pnpm lint` | Lint all apps |
| `pnpm typecheck` | Type-check TypeScript |

### Testing

#### Go API Tests

```bash
cd apps/api
go test ./... -count=1
```

#### Playwright E2E Tests

```bash
cd tests
npx playwright test --project=chromium          # Chrome desktop
npx playwright test --project=firefox           # Firefox desktop
npx playwright test --project=webkit            # Safari desktop (WebKit)
npx playwright test --project="Mobile Chrome"   # Pixel 5
npx playwright test --project="Mobile Safari"   # iPhone 12
```

Playwright config: `tests/playwright.config.ts`
- Timeout: 60000ms, expect timeout: 15000ms
- Retries: 1 (local), 2 (CI)
- Fully parallel (local), 1 worker (CI)
- Web servers auto-started: Go API (`bin\api.exe`) and Next.js (`next start`)
- Reporters: list, HTML, JSON, JUnit
- Output dirs: `./reports/test-results`, `./reports/html`, `./reports/json`, `./reports/junit`

#### Test Structure

- **Page Objects:** `tests/pages/` — LoginPage, RegisterPage, DashboardPage, ProjectPage, TestCasePage, etc.
- **Factories:** `tests/factories/` — UserFactory, ProjectFactory, TestCaseFactory, etc.
- **Fixtures:** `tests/fixtures/` — auth, api, page fixtures
- **Helpers:** `tests/helpers/` — ApiHelper (HTTP client), assertions, CSV, pagination, random data, storage, wait
- **Constants:** `tests/constants/` — routes, roles, selectors

### CI Pipeline (`.github/workflows/ci.yml`)

| Job | What it does |
|-----|-------------|
| `go` | `go build`, `go vet`, `go test -race -count=1` |
| `web` | `pnpm install`, `pnpm turbo run typecheck`, `pnpm turbo run build` |
| `openapi` | Check OpenAPI drift, regenerate SDK, typecheck SDK |
| `ml` | `pip install -e ".[dev]"`, `ruff check`, `pytest` |

**Note:** CI does NOT run Playwright E2E tests. Only Go unit tests, web typecheck/build, OpenAPI drift, and ML tests run in CI.

---

## 6. Phase Status & Roadmap

### Phase Overview

| Phase | Name | Status | Target |
|-------|------|--------|--------|
| 0 | Foundation | **Completed** | CI, native dev env, OpenAPI skeleton |
| 1 | Identity & Tenancy | **Completed** | Auth, RBAC, API keys, web onboarding |
| 2 | Test Management Core | **Completed** | Test cases, suites, folders, search, audit |
| 3 | Execution & Results | **Completed** | Manual runs, CI ingestion, SSE |
| 3.5 | Product UX Completion | **Completed** | UX polish, placeholders, settings, a11y, responsive |
| 4 | API Testing & Defects | **In Progress** | API testing engine, defects, integration hub |
| 5 | Dashboard, Analytics & Launch | **Pending** | Analytics, SDK, deploy, MVP launch |
| 6 | V2 Intelligence | **Pending** | ML flaky detection, failure classification, risk scores |

### Engineering Priority Order

| Priority | Focus |
|----------|-------|
| P0 | Production hardening + security |
| P1 | Frontend foundation + auth reliability |
| P2 | Core MVP: Defects + Notifications |
| P3 | API Testing module |
| P4 | Integration Hub + CI/CD integrations |
| P5 | Real dashboards + reporting |
| P6 | Analytics + Intelligence (V2) |
| P7 | Enterprise hardening (SSO, compliance) |
| P8 | Public API + Marketplace (V3) |

### Phase 4 — Current Phase (In Progress)

**Completed in Phase 4:**
- [x] `defects` module: CRUD, linking to runs/cases, status/severity/priority lifecycle
- [x] `notification` module: in-app feed, preferences, email/Slack/Teams/webhook channels, templates, history
- [x] Web: defect list/create UI
- [x] API testing: backend module with collections, folders, requests, environments, execution
- [x] API testing: frontend Studio page
- [x] `integrationhub` module: routes wired (integrations CRUD, enable/disable, events, dead-letter, incoming webhooks, health checks)
- [x] `analytics` module: routes wired (dashboards CRUD, summary, trends, metrics, recent-activity, CSV export)
- [x] `intelligence` module: routes wired (predict flaky, classify failure, list predictions/clusters)
- [x] `billing` module: routes wired (subscription update/get, invoices list, Stripe stub)
- [x] Background worker with job queue (dead-letter, visibility timeout, poison pill handling)
- [x] Prometheus metrics instrumentation
- [x] CSRF middleware on auth endpoints
- [x] Redis-backed rate limiting with local fallback
- [x] Structured JSON logging (slog)
- [x] SSRF protection in shared/security
- [x] In-process event bus for cross-module communication

**Remaining in Phase 4:**
- [ ] Web: notification center refinements
- [ ] Web: API test builder (frontend refinement)
- [ ] OpenAPI spec updated for all Phase 4 endpoints (63 documented vs 91+ route registrations)
- [ ] End-to-end testing of analytics, intelligence, integrationhub, billing routes
- [ ] Production hardening (secrets management, deployment pipeline)

---

## 7. Current Work — Code Audit and P0 Remediation

> **Read the reports in this order.** Each supersedes the previous where they conflict:
> 1. `docs/audits/TESTRA_CODE_AUDIT_REPORT.md` — the audit that found the issues
> 2. `docs/audits/P0_VERIFICATION_REPORT.md` — first remediation round
> 3. `docs/audits/P0_CLOSEOUT_REPORT.md` — second remediation round (newest)

### What was done

An E2E test debugging session (described below under *E2E history*) was followed by a full code audit of `apps/api`, `apps/web` and the worker, then two rounds of P0 remediation.

### P0 fixes in the working tree

| Area | Change |
|------|--------|
| RLS fail-open | `queue_jobs` policy had `OR app.current_tenant() IS NULL`, letting any tenant-less connection read every queue row. Migration `000041` replaces it with an explicit `app.queue_worker` bypass |
| Tenant context | 10 `RunInTx`/`BeginTx` helpers ignored `SetLocalTenantID` errors; they now roll back and return a wrapped error |
| SSRF | `security.ValidateURL` added to the outbound API-test executor, with a link-local metadata regression test |
| errcheck | 7 high-risk swallowed errors fixed (worker UUID parses, rollbacks, Stripe body reads, CSV writes) |
| OpenAPI | Drift script only compared one direction; fixed, and the spec is now synchronized (168 routes) |
| Tests added | `internal/queue/queue_test.go`, `internal/shared/db/db_test.go` |

### Follow-up fixes applied after the closeout report

These were found while verifying the closeout report's claims:

| Issue | Detail |
|-------|--------|
| Worker-mode RLS bypass leaked across pooled connections | `SetWorkerMode` used `SET`, which survives `COMMIT` and stays on the connection after it returns to the pool. Renamed to `SetLocalWorkerMode` and changed to `SET LOCAL`; `DeleteOldCompleted` now runs in a transaction so the flag is scoped. Verified empirically: the flag no longer persists |
| `DeleteOldCompleted` never worked | Binding a `time.Duration` made PostgreSQL infer the parameter as a timestamp, so `NOW() - $1` produced an interval and the query failed with `operator does not exist: timestamp with time zone < interval`. **Pre-existing bug**, identical in `fa6440c` — worker cleanup had been logging an error every 5 minutes since it shipped. Now uses `make_interval(secs => $1)` |
| Dead code | `ResetWorkerMode` was defined but never called; removed |

### Verified state (re-run independently, not taken from the reports)

| Check | Result |
|-------|--------|
| `go build ./...` | Clean |
| `go vet ./...` | Clean |
| `go test ./... -count=1` | All packages pass |
| `node scripts/check-openapi-drift.mjs` | Synchronized, 168 routes |
| Migration `000041` | Up and down files both present (41 pairs) |

⚠️ **`TestDequeueConcurrency` does not run in CI.** It requires PostgreSQL and skips when unavailable. The `go` job in `.github/workflows/ci.yml` has no `services: postgres`, so it skips there as well as locally. It does pass when given a database (`TEST_DATABASE_URL=... go test ./internal/queue/...`). The closeout report's claim that it "will be covered by `go test -race` in CI" is not currently true.

### E2E history (earlier session, superseded in part)


### Previous Changes (already committed in HEAD)

These changes were made in prior sessions and are part of the current HEAD commit (`fa6440c`):

| File | Change |
|------|--------|
| `apps/api/internal/shared/jwt/manager.go` | Added unique `jti` (JWT ID) claim to prevent identical tokens issued in the same second (JWT already uses RS256 with `kid` header) |
| `apps/api/internal/identity/service.go` | `RequestPasswordReset` logs SMTP errors instead of returning them |
| `apps/api/internal/results/handler.go` | Added validation to reject empty `test_case_ids` in CreateRun handler |
| `tests/e2e/auth/session.spec.ts` | Fixed register test to use factory instead of re-registering existing user |
| `tests/pages/ProjectPage.ts` | Fixed labels to match UI: "Project name" and "Project key" |
| `tests/e2e/permissions/permissions.spec.ts` | Fixed to use unique slug to avoid CONFLICT errors |
| `tests/pages/ApiTestingPage.ts` | Fixed `createCollection` to press Enter instead of looking for "Add" button |
| `tests/e2e/file-upload/file-upload.spec.ts` | Fixed Import button selector with `exact: true` |
| `tests/pages/SearchPage.ts` | Fixed to dispatch `open-global-search` event and wait for sidebar |

### Final Test Results

| Suite | Total | Passed | Failed | Flaky |
|-------|-------|--------|--------|-------|
| Go API tests | All | All | 0 | 0 |
| Playwright Chromium | 230 | 229 | 0 | 1 |
| Playwright Firefox | 230 | 230 | 0 | 0 |
| Playwright Mobile Chrome | 230 | 230 | 0 | 0 |
| Playwright WebKit | 230 | 180 | 50 | 0 |
| Playwright Mobile Safari | 230 | 180 | 50 | 0 |

### Known Issues (Not Fixed — Environment)

**WebKit / Mobile Safari on Windows: 50 tests fail**

> ❌ **The "SSL connect error" diagnosis below was wrong.** `docs/audits/P0_VERIFICATION_REPORT.md` §4 disproved it: network traces show HTTP/1.1 200 responses with `load`/`domcontentloaded` firing normally. It is kept here only to stop the theory being re-derived.
>
> ~~Root cause: WebKit on Windows fails to load Next.js JavaScript chunks over HTTP with `SSL connect error`.~~

- **Actual causes identified:**
  1. Test selectors colliding with the Next.js route announcer, which adds a second `role="alert"` element (fixed in `BasePage.ts` and `accessibility-expanded.spec.ts`)
  2. Workspace-context timing — `setWorkspaceContext` raced the app's redirect to `/create-workspace` (fixed via `page.addInitScript` in `tests/helpers/storage.ts`)
  3. A remaining class of app-level load/hydration failures where the page stays on `img "Loading"`
- **Status:** Causes 1 and 2 are fixed. Cause 3 still accounts for ~50 unexpected failures out of 180 unique tests in both WebKit and Mobile Safari, and needs an app-level fix, not a test-locator fix.
- **Note:** WebKit does render correctly at least some of the time — the committed `*-webkit-win32.png` visual baselines show fully rendered pages, not spinners.

### Known Issues (Flaky Tests)

**2 flaky tests under parallel load:**

1. `e2e/auth/login.spec.ts:5` — "user can register and log in" — passes on retry, flaky under parallel load due to API registration timing
2. `e2e/accessibility/accessibility-expanded.spec.ts:72` — "login page error messages have role alert" — passes on retry, flaky due to API response latency under parallel load

Both are handled by existing `retries: 1` config.

---

## 8. Engineering Rules (Must Follow)

### Mandatory Rules

1. **Never bypass middleware** — Auth → TenantContext → RequirePermission in order
2. **Never query SQL from handlers** — all persistence through repository ports
3. **Never skip tenant validation** — client IDs are selectors, not proof of access
4. **Never ignore RLS** — every tenant-scoped query needs `app.tenant_id` set
5. **Never return raw secrets** — no password hashes, API key plaintext, MFA secrets outside issuance
6. **Never duplicate response envelopes** — use `shared/http` helpers
7. **Never mutate merged migrations** — add new up/down pairs
8. **Never skip migration down files** — every up needs a down
9. **Never introduce libraries without justification** — prefer existing stack
10. **Never create cross-module imports of internals** — depend on ports or shared primitives
11. **Always update OpenAPI before implementation**
12. **Always update ADRs for architecture changes**
13. **Always update BIBLICAL when adding modules**
14. **Always write tests for domain logic**
15. **Always run migrations up and down locally before committing**

### Do Not Break List

- Tenant isolation (RLS)
- RBAC enforcement
- Audit logs (immutable)
- Idempotency on side-effecting endpoints
- Response envelope shape
- Cursor pagination (never offset)
- Migration ordering (monotonic, immutable)
- API compatibility (breaking changes need new version or ADR)
- Zero customer code retention
- Zero API collection retention (unless created by Testra)
- No external LLM processing
- Secrets hashed at rest

---

## 9. Technical Debt Summary

### P0 — Production Blockers

| ID | Title | Status |
|----|-------|--------|
| P0-5 | Production systemd ConfigMap hard-codes localhost CORS | Accepted |

### P1 — High Risk

| ID | Title | Status |
|----|-------|--------|
| P1-11 | Production systemd manifests store HTTP internal URLs | Open |

### P2 — Medium Risk (selected)

| ID | Title | Status |
|----|-------|--------|
| P2-1 | `packages/sdk` and `packages/ui` are empty placeholders | Open |
| P2-2 | `packages/shared` only exports a tagline constant | Open |
| P2-3 | No generated API contract / SDK | Open |
| P2-7 | Analytics/intelligence modules may not be fully wired | Open |
| P2-12 | `results.Service` recalculates aggregates by loading all items | Open |
| P2-13 | Several list endpoints remain unpaginated | Open |
| P2-15 | OpenAPI contract drifts behind implementation (63 vs 91 routes) | Open |

### Key Open Backlog Items

| ID | Title | Priority | Status |
|----|-------|----------|--------|
| T9 | ~~Harden frontend auth token storage (localStorage → httpOnly cookies)~~ | High | **Closed** — already HttpOnly cookies; see `docs/audits/P0_VERIFICATION_REPORT.md` §6 |
| T17 | `TestDequeueConcurrency` never executes in CI (no PostgreSQL service in the `go` job) | High | Open |
| T11 | Regression tests for queue and idempotency | High | Open |
| T14 | Generate and validate OpenAPI spec against routes | Medium | Open |
| T15 | Implement package shared runtime/types and SDK | Low | Open |
| T16 | Complete production systemd/nginx overlays | High | Open |

---

## 10. Key Files Reference

| File | Purpose |
|------|---------|
| `apps/api/cmd/api/main.go` | API server entrypoint |
| `apps/api/cmd/migrator/main.go` | Migration runner |
| `apps/api/cmd/worker/main.go` | Background worker (job queue processing with exponential backoff, cleanup) |
| `apps/api/internal/shared/server/server.go` | Route tree and middleware wiring |
| `apps/api/internal/shared/middleware/auth.go` | JWT authentication |
| `apps/api/internal/shared/middleware/tenant.go` | Tenant context and RLS |
| `apps/api/internal/shared/middleware/rbac.go` | Permission enforcement |
| `apps/api/internal/shared/middleware/idempotency.go` | Idempotency-Key middleware |
| `apps/api/internal/identity/service.go` | Registration, login, MFA, refresh, password reset |
| `apps/api/internal/results/service.go` | Run lifecycle and SSE progress hub |
| `apps/api/internal/automationhub/service.go` | JUnit/Playwright/Cypress ingestion |
| `apps/api/internal/apitesting/handler.go` | API testing collections, requests, environments, execution |
| `apps/api/internal/defects/handler.go` | Defect CRUD and lifecycle |
| `apps/api/internal/notification/handler.go` | Notifications, preferences, channels, templates, history |
| `apps/api/internal/integrationhub/handler.go` | Integrations CRUD, events, dead-letter, webhooks |
| `apps/api/internal/analytics/handler.go` | Analytics dashboards, summary, trends, metrics, CSV export |
| `apps/api/internal/intelligence/handler.go` | ML flaky prediction, failure classification |
| `apps/api/internal/billing/handler.go` | Subscription and invoice management (Stripe stub) |
| `apps/api/internal/queue/queue.go` | Background job queue with dead-letter and visibility timeout |
| `apps/api/internal/metrics/metrics.go` | Prometheus metrics instrumentation |
| `apps/api/internal/search/handler.go` | Global search across test cases, runs, defects |
| `apps/api/internal/shared/eventbus/eventbus.go` | In-process event bus |
| `apps/api/internal/shared/secrets/secrets.go` | Secrets provider abstraction |
| `apps/api/internal/shared/security/ssrf.go` | SSRF protection (URL validation) |
| `apps/api/internal/shared/middleware/csrf.go` | CSRF token middleware |
| `apps/api/internal/shared/middleware/cookies.go` | Cookie helpers for auth tokens |
| `apps/api/internal/shared/middleware/redact.go` | Request body redaction for logging |
| `apps/api/internal/shared/middleware/redis_ratelimiter.go` | Redis-backed rate limiter |
| `apps/web/lib/api.ts` | Web API client and token management |
| `apps/web/middleware.ts` | Security headers, CSP |
| `tests/playwright.config.ts` | Playwright configuration |
| `tests/helpers/api.ts` | E2E API helper (HTTP client with CSRF) |
| `tests/factories/user.ts` | User test data factory |
| `tests/pages/BasePage.ts` | Base page object with common methods |
| `docs/BIBLICAL_TESTRA.md` | Engineering handbook (canonical) |
| `docs/engineering/ROADMAP.md` | Implementation phases and priorities |
| `docs/FEATURE_MATRIX.md` | Feature completion matrix |
| `docs/api/openapi/openapi.yaml` | OpenAPI 3.1 contract |
| `docs/architecture/DATABASE_GUIDE.md` | Database schema, RLS, ERD |
| `docs/AI_CONTEXT.md` | AI entry point and verification workflow |
| `docs/AI_MEMORY.md` | Permanent architectural facts |
| `docs/AI_RULES.md` | Change-impact matrix for AI agents |

---

## 11. How to Review This Codebase

### If reviewing the current changes:

1. Check `git diff HEAD` for uncommitted changes (~29 modified files), then read the three reports in the order given in §7
2. Verify Go tests pass: `cd apps/api && go test ./... -count=1`
   - For the database-backed tests, also run with `TEST_DATABASE_URL` set, or they silently skip
3. Verify Chromium tests pass: `cd tests && npx playwright test --project=chromium`
4. Verify Firefox tests pass: `cd tests && npx playwright test --project=firefox`
5. Verify Mobile Chrome tests pass: `cd tests && npx playwright test --project="Mobile Chrome"`
6. WebKit/Mobile Safari failures are expected on Windows (SSL connect error — environment issue, not code bug)

### If reviewing the overall codebase:

1. Start with `docs/BIBLICAL_TESTRA.md` — the engineering handbook
2. Check `docs/engineering/ROADMAP.md` — what's done and what's planned
3. Check `docs/FEATURE_MATRIX.md` — feature completion status
4. Check `ENGINEERING_DEBT_REGISTER.md` — known technical debt
5. Review `docs/api/openapi/openapi.yaml` — API contract
6. Review `apps/api/internal/shared/server/server.go` — route wiring
7. Review test structure under `tests/`

### If continuing the E2E debugging work:

1. The uncommitted changes in the working tree are the P0 fixes plus the follow-ups listed in §7
2. WebKit/Mobile Safari: ~50 failures are app-level loading/hydration issues, **not** an SSL/environment problem — that theory was disproved
3. 2 flaky tests (login + accessibility alert) are handled by retry config
4. Chromium, Firefox and Mobile Chrome pass: 690 passed, 0 failed, 1 flaky
5. Highest-value next steps:
   - Add a `services: postgres` block to the `go` job in CI so `TestDequeueConcurrency` actually runs (currently skips everywhere)
   - Diagnose the WebKit loading/hydration failures at the app level
   - Triage the remaining ~253 lower-risk `errcheck` findings
   - Phase 4 remainder: E2E coverage for analytics/intelligence/integrationhub/billing, notification centre refinements, API test builder frontend, production hardening

---

## 12. Environment Notes

- **OS:** Windows (development machine)
- **Database:** PostgreSQL 16+ running locally
- **Redis:** Running locally (Memurai or WSL2)
- **SMTP:** Mailpit on port 1025 (SMTP_PASSWORD secret not set in test env — causes warning but not failure)
- **MinIO:** Running locally for S3-compatible storage
- **Go:** 1.24+
- **Node.js:** 20+
- **pnpm:** 9.5+
- **Python:** 3.12+ (for ML service)

### Known Environment Quirks

- WebKit / Mobile Safari on Windows: ~50 E2E tests fail on pages stuck loading (app-level hydration, **not** an SSL/chunk issue — see §7)
- `go test -race` cannot run locally: no C compiler installed (`cgo: C compiler "gcc" not found`). CI on Ubuntu does run it
- Database-backed Go tests skip unless `TEST_DATABASE_URL` is set (default DSN: `postgres://testra:testra@localhost:5432/testra?sslmode=disable`)
- `SMTP_PASSWORD` secret not set in test environment — password reset email sending logs a warning but doesn't fail tests (error is logged, not returned)
- `dns.setDefaultResultOrder("ipv4first")` in Playwright config to handle Windows DNS resolution issues
- Playwright web servers reuse existing servers in local mode (not CI)
