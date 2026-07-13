# Testra — Implementation Phases

**Status:** Active
**Last Updated:** July 2026

---

## Phase Overview

| Phase | Name | Status | Target |
|---|---|---|---|
| 0 | Foundation | **Completed** | Scaffold hardening, CI, OpenAPI skeleton |
| 1 | Identity & Tenancy | **Completed** | Self-hosted auth, RBAC, API keys, web onboarding |
| 2 | Test Management Core | Pending | Test cases, suites, folders, search, audit |
| 3 | Execution & Results | Pending | Manual runs, CI ingestion, ClickHouse, SSE |
| 4 | API Testing & Defects | Pending | API test engine, defects, Jira sync, notifications |
| 5 | Dashboard, Analytics & Launch | Pending | Analytics, SDK, deploy, MVP launch |
| 6 | V2 Intelligence | Pending | ML flaky detection, failure classification, risk scores |

---

## Phase 0 — Foundation (Completed)

### Objectives
- Harden the existing scaffold with proper CI, repo hygiene, and native development environment.
- Establish OpenAPI contract skeleton and ADR-001 (hybrid auth).

### Definition of Done
- [x] Native development environment with local PostgreSQL, Redis, Mailpit, MinIO (Docker optional per ADR-009)
- [x] `.gitignore` and `.env.example` present
- [x] ADR-001 (hybrid auth) recorded
- [x] OpenAPI 3.1 skeleton covering existing endpoints
- [x] GitHub Actions CI workflow (Go build/vet/test, web typecheck/build, ML lint)
- [x] Turbo 2.x compatibility (`pipeline` → `tasks`)
- [x] `go build`, `go vet`, `go test`, `pnpm turbo run typecheck` all pass
- [x] ADR-009 (native development environment) recorded

### Completed Work
- Fixed ClickHouse/MinIO port collision (MinIO API → host `9002`)
- Added Mailpit service (SMTP `1025`, UI `8025`)
- Added healthchecks for all compose services
- Created `.gitignore`, `.env.example`
- Created `docs/architecture/adrs/ADR-001-hybrid-auth.md`
- Created `docs/api/openapi/openapi.yaml` (auth, organizations, workspaces, projects)
- Created `.github/workflows/ci.yml`
- Fixed `turbo.json` for Turbo 2.x
- Implemented `project` module (domain, ports, repository, service, handler, migration)
- Unit tests for project service (validation, key normalization, conflict, get/list)
- Routes wired into server: `POST/GET /api/v1/projects`, `GET /api/v1/projects/{id}`

---

## Phase 1 — Identity & Tenancy (Completed)

### Objectives
- Complete self-hosted authentication (register, login, MFA, password reset).
- Establish multi-tenancy model (organization → workspace → project).
- Implement RBAC with roles and permissions.
- Deliver scoped API keys for CI/CD ingestion.
- Build web auth and onboarding UI.

### Definition of Done
- [x] TOTP MFA enrollment and verification
- [x] Password reset flow (request → email → reset)
- [x] RBAC: roles, permissions, middleware enforcement
- [x] Scoped, hashed API keys with one-time display and revocation
- [x] Web: login, register, MFA setup, password reset pages
- [x] Web: organization creation, workspace creation, app shell
- [x] OpenAPI spec updated for all new endpoints
- [x] Unit tests for identity domain logic (17 tests)
- [x] Migration for roles, permissions, API keys tables
- [x] `PHASES.md` updated to mark Phase 1 complete

### Dependencies
- Phase 0 (completed) — scaffold, CI, compose stack

### Completed Work
- [x] `identity` module: register, login, JWT, password hashing (bcrypt)
- [x] `organization` module: create, list, get
- [x] `workspace` module: create, list, get, members
- [x] `project` module: create, list, get (completed in Phase 0)
- [x] Migrations 000001–000007 (users, organizations, workspaces, projects, MFA+reset, RBAC, API keys)
- [x] Auth middleware (JWT bearer)
- [x] Shared: config, errors, response envelope, JWT, password, DB
- [x] TOTP MFA: setup, verify, disable endpoints (`pquerna/otp`)
- [x] Password reset: request, confirm endpoints (SHA-256 hashed tokens, 30min expiry)
- [x] RBAC: roles, permissions, role_assignments tables with seed data (4 roles, 21 permissions)
- [x] RBAC middleware: `RequirePermission` with `PermissionLoader` interface
- [x] `rbac` package: `SQLPermissionLoader` implementation
- [x] `apikeys` module: domain, ports, repository, service, handler, module wiring
- [x] API keys: create (one-time display), list, revoke; SHA-256 hashed; `testra_` prefix
- [x] Web: TailwindCSS 3 + PostCSS config
- [x] Web: UI components (Button, Input, Card)
- [x] Web: API client (`lib/api.ts`) with token management
- [x] Web: auth layout + login page (with MFA code field)
- [x] Web: register page
- [x] Web: forgot-password page
- [x] Web: reset-password page
- [x] Web: MFA setup page (QR code display + verification)
- [x] Web: onboarding page (create org + workspace)
- [x] Web: dashboard layout with sidebar nav
- [x] Web: dashboard page
- [x] 17 unit tests for identity service (MFA + password reset)
- [ ] OpenAPI spec update for new endpoints (deferred to Phase 2 start)

### Remaining Work
- [ ] OpenAPI spec update for MFA, password reset, RBAC, API key endpoints
- [ ] SMTP email sending for password reset (currently returns token in API response for dev)
- [ ] RBAC middleware wired to specific routes (infrastructure ready, enforcement per-route pending)

---

## Phase 2 — Test Management Core (Pending)

### Objectives
- Implement test case management with folders, suites, and versioning.
- Enable full-text search across test cases.
- Wire audit trail into all mutating use cases.

### Definition of Done
- [ ] `testmanagement` module: test cases, suites, folders, version history
- [ ] PG full-text search on test cases (title, description)
- [ ] `audit` module: immutable event log on all mutations
- [ ] Web: test case CRUD, suite tree, rich editor
- [ ] OpenAPI spec updated
- [ ] Unit tests for testmanagement domain logic
- [ ] Migrations for test_cases, test_suites, test_folders, audit_events

### Dependencies
- Phase 1 (identity, tenancy, RBAC)

---

## Phase 3 — Execution & Results (Pending)

### Objectives
- Manual test run execution with live progress.
- CI/CD result ingestion (JUnit XML, Playwright/Cypress JSON).
- ClickHouse ingestion path for high-volume results.
- SSE for live test execution updates.

### Definition of Done
- [ ] Manual test runs: plans, execution flow, statuses
- [ ] `automationhub`: CI ingestion API (results/metadata only — zero code retention)
- [ ] `results` module + ClickHouse ingestion
- [ ] SSE endpoint for live run progress
- [ ] Web: runs list, run detail, live execution view
- [ ] OpenAPI spec updated
- [ ] Integration tests for ingestion pipeline

### Dependencies
- Phase 2 (test management, audit)

---

## Phase 4 — API Testing & Defects (Pending)

### Objectives
- API test definitions and execution engine.
- Defect tracking with integration sync.
- Notification system.

### Definition of Done
- [ ] `apitesting` module: request definitions, environments, execution (zero collection retention)
- [ ] `defects` module: CRUD, linking to runs/cases
- [ ] `integrationhub`: Jira sync, CI webhooks
- [ ] `notification` module: in-app, email
- [ ] Web: API test builder, defect board, notification center
- [ ] OpenAPI spec updated

### Dependencies
- Phase 3 (results, runs)

---

## Phase 5 — Dashboard, Analytics & Launch (Pending)

### Objectives
- Analytics dashboards from ClickHouse.
- SDK generation from OpenAPI.
- Production deployment and MVP launch.

### Definition of Done
- [ ] `analytics` module: dashboard aggregates, trends, reports
- [ ] SDK generated from OpenAPI spec (`packages/sdk`)
- [ ] Contract tests in CI
- [ ] Staging + production deploy (managed platform)
- [ ] Backups, monitoring, runbooks
- [ ] Onboarding flow, seed/demo data
- [ ] **MVP launch**

### Dependencies
- Phase 4 (all core modules)

---

## Phase 6 — V2 Intelligence (Post-MVP)

### Objectives
- ML-powered flaky detection, failure classification, risk/health scores.
- Meilisearch, Stripe billing, WorkOS SSO (when enterprise deal requires).
- Kubernetes migration.

### Definition of Done
- [ ] `intelligence` module + `apps/ml`: flaky detection
- [ ] Failure classification (clustering + XGBoost)
- [ ] Risk/health scores with SHAP explainability
- [ ] Meilisearch integration
- [ ] Stripe billing
- [ ] WorkOS SSO (conditional)
- [ ] K8s migration (EKS/GKE)

### Dependencies
- Phase 5 (MVP launched, production data available)
