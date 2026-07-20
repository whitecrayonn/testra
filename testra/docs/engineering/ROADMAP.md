# Testra Roadmap

**Purpose:** Track implementation phases, engineering priorities, technical debt, and documentation roadmap.
**Owner:** Engineering Lead / CTO
**Scope:** Phase plans, definition of done, completed work, dependencies, and technical debt register.
**Status:** Active
**Last Updated:** July 2026
**Source of Truth:** ROADMAP.md for implementation status and priorities.
**Related documents:**
- [`BIBLICAL_TESTRA.md`](../BIBLICAL_TESTRA.md)
- [`PROJECT_OVERVIEW.md`](../PROJECT_OVERVIEW.md)
- [`FEATURE_MATRIX.md`](../FEATURE_MATRIX.md)

---

## Phase Overview

| Phase | Name | Status | Target |
|---|---|---|---|
| 0 | Foundation | **Completed** | Scaffold hardening, CI, OpenAPI skeleton |
| 1 | Identity & Tenancy | **Completed** | Self-hosted auth, RBAC, API keys, web onboarding |
| 2 | Test Management Core | **Completed** | Test cases, suites, folders, search, audit |
| 3 | Execution & Results | **Completed** | Manual runs, CI ingestion, SSE (ClickHouse deferred per ADR-010) |
| 3.5 | Product UX Completion & Frontend Stabilization | **Completed** | UX polish, placeholders, settings, a11y, responsive, build/verification |
| 4 | API Testing & Defects | In Progress | Notifications complete; defects and API test engine in progress |
| 5 | Dashboard, Analytics & Launch | Pending | Analytics, SDK, deploy, MVP launch |
| 6 | V2 Intelligence | Pending | ML flaky detection, failure classification, risk scores |

---

## Phase 0 — Foundation (Completed)

### Objectives
- Harden the existing scaffold with proper CI, repo hygiene, and native development environment.
- Establish OpenAPI contract skeleton and ADR-001 (hybrid auth).

### Definition of Done
- [x] Native development environment with local PostgreSQL, Redis, Mailpit, MinIO (no Docker per ADR-009)
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
- [x] `ROADMAP.md` updated to mark Phase 1 complete

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
- [x] OpenAPI spec updated for all Phase 1 endpoints (refresh, password reset, API key expiry, MFA)
- [x] Migrations 000008–000011 (RLS policies, refresh tokens, audit events)

### Remaining Work
- None — all Phase 1 carryover items completed

---

## Phase 2 — Test Management Core (Approved)

### Objectives
- Implement test case management with folders, suites, and versioning.
- Enable full-text search across test cases.
- Wire audit trail into all mutating use cases.

### Definition of Done
- [x] `testmanagement` module: test cases, suites, folders, version history
- [x] PG full-text search on test cases (title, description)
- [x] `audit` module: immutable event log on all mutations
- [x] Web: test case CRUD, suite tree, rich editor
- [x] OpenAPI spec updated
- [x] Unit tests for testmanagement domain logic
- [x] Migrations for test_cases, test_suites, test_folders, audit_events
- [x] RLS policies on test tables (migration 000014)
- [x] Composite cursor pagination for search (rank, id)
- [x] Transactional version snapshot + update for test cases
- [x] Audit logging on all 9 mutating endpoints

### Engineering Review
- Review, resolution, and phase gate reports are archived under `docs/archive/historical/reviews/` and `docs/archive/historical/phase-gates/`.

### Dependencies
- Phase 1 (identity, tenancy, RBAC)

---

## Phase 3 — Execution & Results (Completed)

### Objectives
- Manual test run execution with live progress.
- CI/CD result ingestion (JUnit XML, Playwright/Cypress JSON).
- ClickHouse ingestion path for high-volume results.
- SSE for live test execution updates.

### Definition of Done
- [x] Manual test runs: plans, execution flow, statuses
- [x] `automationhub`: CI ingestion API (results/metadata only — zero code retention)
- [x] `results` module + PostgreSQL ingestion (ClickHouse deferred per ADR-010)
- [x] SSE endpoint for live run progress
- [x] Web: runs list, run detail, live execution view
- [x] OpenAPI spec updated (v0.4.0)
- [x] Idempotency-Key on ingestion endpoint (ADR-012 — HIGH priority)
- [x] Integration tests for ingestion pipeline
- [x] Phase 3 security and performance reviews are archived under `docs/archive/historical/reviews/`.

### Engineering Review
- Final gate review is archived under `docs/archive/historical/phase-gates/`.

### Dependencies
- Phase 2 (test management, audit)

---

## Phase 3.5 — Product UX Completion & Frontend Stabilization (Completed)

### Objectives
- Stabilize the frontend product experience before Phase 4.
- Ensure every sidebar menu item has a working page, route, and consistent layout.
- Implement all settings pages with consistent UX and placeholder pages for unimplemented Phase 4 features.
- Polish dashboard UX, empty/loading/error states, accessibility, and responsiveness.
- Resolve build, typecheck, and lint errors.

### Definition of Done
- [x] Settings layout and all settings subpages created.
- [x] Production-quality placeholder pages for unimplemented features.
- [x] Dashboard, projects, test cases, and test runs polished with `PageHeader`, `EmptyState`, `Skeleton`, `Badge`, and `LinkButton`.
- [x] Sidebar active states and accessible navigation.
- [x] `pnpm turbo run typecheck` passes.
- [x] `pnpm lint` passes with 0 warnings.
- [x] `pnpm turbo run build` passes.

### Deliverables
- UX review artifacts are archived under `docs/archive/historical/frontend-ux-review.md` and `docs/archive/historical/product-ux-completion.md`.

---

## Phase 4 — API Testing & Defects (In Progress)

### Objectives
- API test definitions and execution engine.
- Defect tracking with integration sync.
- Notifications and notification channels are already implemented; remaining integration sync and channels hardening continue in this phase.

### Definition of Done
- [ ] `apitesting` module: request definitions, environments, execution (zero collection retention)
- [x] `defects` module: CRUD, linking to runs/cases, status/severity/priority lifecycle
- [ ] `integrationhub`: Jira sync, CI webhooks
- [x] `notification` module: in-app feed, preferences, email/Slack/Teams/webhook channels (completed in Phase 3.5)
- [ ] Defects and integration hub notifications/channels hardening
- [x] Web: defect list/create UI
- [ ] Web: API test builder
- [ ] Web: notification center refinements
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
- single-VPS scaling or a future managed platform (not planned for MVP).

### Definition of Done
- [ ] `intelligence` module + `apps/ml`: flaky detection
- [ ] Failure classification (clustering + XGBoost)
- [ ] Risk/health scores with SHAP explainability
- [ ] Meilisearch integration
- [ ] Stripe billing
- [ ] WorkOS SSO (conditional)
- [ ] Document migration path to a managed platform (not MVP)

### Dependencies
- Phase 5 (MVP launched, production data available)


## Engineering Next Steps

> This is the prioritized engineering roadmap derived from [`PROJECT_OVERVIEW.md`](../PROJECT_OVERVIEW.md) and the Technical Debt Register below. The sequence is designed to make the platform safe, usable, and complete before new feature work begins.

### Priority overview

| Priority | Focus | Why it exists |
|----------|-------|---------------|
| **P0** | Production hardening + security | Nothing else matters if the platform is not secure, available, and compliant. |
| **P1** | Frontend foundation + authentication reliability | The current frontend loses sessions, has broken SSE, and exposes unauthenticated UI flashes. |
| **P2** | Core MVP completion: Defects + Notifications | Finishes the day-to-day QA workflow (test → defect → alert) and unlocks value. |
| **P3** | API Testing module | Displaces Postman for QA-led API testing; core product differentiator. |
| **P4** | Integration Hub + CI/CD integrations | Automates result ingestion and defect sync, driving retention and enterprise stickiness. |
| **P5** | Real dashboards + Reporting | Replaces the dashboard skeleton with actionable quality signals and reports. |
| **P6** | Analytics + Intelligence (V2) | Builds the data-driven quality moat: flaky tests, failure classification, risk scoring. |
| **P7** | Enterprise hardening (SSO, compliance, data residency) | Unlocks enterprise deals and regulated verticals. |
| **P8** | Public API + Marketplace (V3) | Turns Testra from a product into a platform and ecosystem. |

---

### P0 — Production hardening and security

**Goal:** Make the platform safe to deploy and operate.

1. ✅ **Wire rate limiting** on `/auth/*` and `/ingest` using Redis-backed token buckets.
2. ✅ **Implement API-key authentication middleware** so CI/CD runners can use scoped keys instead of user JWTs.
3. **Harden SSE authentication** for live test-run progress — a query-token workaround is already implemented; replace with a signed SSE token or cookie before public networks.
4. **Move secrets out of `.env.example`** and add startup validation; integrate with a secrets manager for production.
5. **Make audit logging durable** — write inside the request transaction or enqueue to a reliable worker queue.
6. ✅ **Add frontend route guards**; `localStorage` token storage remains and should move to `httpOnly` cookie for production hardening.

**Why this is P0:** These items are prerequisites for any production deployment, SOC 2 readiness, and enterprise sales. Without them, Testra is not a credible B2B SaaS product.

---

### P1 — Frontend foundation and auth reliability

**Goal:** Make the web app reliable and pleasant to use.

1. **Add a global auth state layer** (Zustand or React Context) with token refresh on 401. ✅ Implemented in `lib/api.ts` fetch wrapper.
2. ✅ **Implement a fetch interceptor** that calls `/auth/refresh` and retries once, then redirects to `/login` on failure.
3. **Consolidate dashboard route trees** to `/[workspace]` and redirect `/dashboard` to the selected workspace.
4. **Add global `loading.tsx`, `error.tsx`, and `not-found.tsx`** in the dashboard group.
5. **Add structured logging and health/readiness endpoints** to the Go API.

**Why this is P1:** The frontend still loses sessions every 15 minutes and exposes unauthenticated UI flashes. Fixing auth state, token refresh, and navigation unblocks every subsequent feature.

---

### P2 — Core MVP completion: Defects + Notifications

**Goal:** Complete the core QA execution loop.

1. **Notifications module (backend + frontend):** ✅ Completed.
   - `apps/api/internal/notification/` with in-app notifications, preferences, and email/Slack/Teams/webhook channels.
   - `apps/web/app/(dashboard)/dashboard/notifications/page.tsx` list page and `dashboard/settings/notifications/` preferences page.
   - Sidebar bell with unread count badge.
2. ✅ **Defects module (backend + frontend):**
   - `apps/api/internal/defects/` with CRUD, lifecycle, severity/priority fields, and tenant isolation.
   - `apps/web/app/(dashboard)/[workspace]/defects/` list/create page backed by `/api/v1/defects`.
3. **Jira/Linear/GitHub Issues integration design:** at minimum design the outbound webhook schema and queue job for defect sync.

**Why this is P2:** Test execution without defect tracking and alerting is incomplete. These two modules close the manual testing workflow and are table stakes for any QA platform.

---

### P3 — API Testing module

**Goal:** Build a native API testing experience that can replace Postman for QA teams.

1. **Backend (`apps/api/internal/apitesting/`):**
   - API request definitions (method, URL, headers, body, assertions).
   - Environment and variable scoping per project/workspace.
   - Execution engine using Go `net/http` with result capture.
   - Store request/response history in `test_runs`/`test_run_items` or a dedicated table.
2. **Frontend (`apps/web/features/api-testing/`):**
   - Collection/folder tree, request editor, environment selector, response viewer, and run history.

**Why this is P3:** API testing is a core differentiator and one of the top displacements (Postman). It depends on P0/P1 stability but can be built in parallel with P2 once the platform is reliable.

---

### P4 — Integration Hub + CI/CD integrations

**Goal:** Automate the flow of results and defects into Testra.

1. **Integration Hub backend (`apps/api/internal/integrationhub/`):**
   - Webhook receivers for GitHub Actions, GitLab CI, Jenkins, and CircleCI.
   - Outbound webhooks for Jira/Linear/GitHub issue sync with HMAC-SHA256 signatures.
   - Credential storage (OAuth tokens, API keys) per workspace.
2. **Frontend settings pages for integrations.**
3. **Document and publish `/ingest` contract** for CI plugins and update OpenAPI.

**Why this is P4:** CI/CD result ingestion is a major retention driver (automation-heavy teams see value immediately) and removes the biggest objection from enterprise buyers who already have Jira.

---

### P5 — Real dashboards + reporting

**Goal:** Replace the dashboard skeleton with actionable quality signals.

1. **Backend (`apps/api/internal/analytics/`):**
   - Aggregations: pass/fail rate, run count, open defects by severity, recent runs, top failing tests.
2. **Frontend (`apps/web/features/analytics/`):**
   - Role-based dashboard widgets (QA engineer, QA lead, engineering manager).
   - Run history, test case coverage, and project health summary.
3. **Reports:** PDF/CSV export for run summaries and traceability matrices.

**Why this is P5:** Dashboards and reports are the "aha moment" for leadership buyers. They also create the data foundation for the intelligence layer in P6.

---

### P6 — Analytics + Intelligence (Version 2.0)

**Goal:** Deliver data-driven quality insights that competitors cannot easily copy.

1. **Flaky test detection** using pass/fail variance over time.
2. **Failure classification** (environment, test data, product defect, infrastructure) via rules and clustering.
3. **Risk scoring** and **test suite health scores** with human-readable explanations.
4. **Release readiness report** aggregating coverage, flakiness, open defects, and recent failure trends.
5. **Wire `apps/ml` Python service** for model training/inference and `apps/worker` for background jobs.
6. **Adopt ClickHouse** for high-volume result analytics.

**Why this is P6:** Intelligence features require enough historical data to be credible. Building them after P2–P5 ensures the data pipeline and storage choices are ready.

---

### P7 — Enterprise hardening

**Goal:** Make Testra sellable to regulated, large organizations.

1. **SSO / SAML 2.0 and OIDC** integration.
2. **SCIM provisioning** for user lifecycle management.
3. **Workspace/project-level RBAC** and custom roles.
4. **Advanced audit export** and compliance report templates.
5. **Data residency** options (Singapore, Indonesia) and multi-region deployment.
6. **SOC 2 Type II evidence** collection and security documentation.

**Why this is P7:** These are enterprise hard requirements that unlock $50K+ ACV deals. They are unnecessary for the initial mid-market self-serve motion but critical for the long-term moat.

---

### P8 — Public API + Marketplace (Version 3.0)

**Goal:** Turn Testra into a platform.

1. **Public API (`/api/v1` stabilization + versioning strategy).**
2. **Official TypeScript SDK** generated from OpenAPI in `packages/sdk`.
3. **Partner Marketplace** for test integrations, custom report templates, and notification channels.
4. **Predictive analytics** and cross-project governance dashboards.

**Why this is P8:** Public API and marketplace create ecosystem lock-in and partner revenue. They require a mature, stable, and well-documented core product first.

---

### How to use this roadmap

- **Start at P0 and do not skip.** Production hardening is non-negotiable.
- **Within each priority, tackle foundational blockers first** (e.g., auth state before notifications UI).
- **Track progress in [`FEATURE_MATRIX.md`](../FEATURE_MATRIX.md).** When a feature becomes `Production Ready`, update the matrix and close the related item in the Technical Debt Register below.
- **Re-evaluate after each milestone.** Customer feedback from P1–P3 may shift the order of P4–P8, but P0 must remain complete before launch.


## Documentation Roadmap

> This roadmap schedules the documentation improvements identified in [`DOCUMENTATION_CONSOLIDATION_REPORT.md`](../archive/superseded/DOCUMENTATION_CONSOLIDATION_REPORT.md). It is aligned with the engineering phases in `docs/engineering/ROADMAP.md`.

### Status legend

- **Immediate:** Do within the current documentation audit pass (days).
- **Short-term:** Before or during the first sprint of Phase 4.
- **Medium-term:** During Phase 4 execution (API testing, defects, integrations).
- **Long-term:** Phase 5+ (analytics, launch, enterprise features).

### Immediate (this pass)

| # | Deliverable | Why | Success criteria | Owner |
|---|-------------|-----|------------------|-------|
| 1 | `docs/README.md` (Documentation Index) | One canonical map of every document and its status. | All major docs listed with status and cross-references. | ✅ Done |
| 2 | `docs/archive/superseded/DOCUMENTATION_CONSOLIDATION_REPORT.md` | Score every doc so stakeholders know what to trust. | 38+ docs scored with category averages and top issues. | ✅ Done |
| 3 | `docs/archive/superseded/DOCUMENTATION_CONSOLIDATION_REPORT.md` | Enumerate exactly what is stale, missing, duplicated, or conflicting. | Gaps numbered, prioritized, with remediation actions. | ✅ Done |
| 4 | `docs/engineering/ROADMAP.md` | This file. | Time-bound plan linked to engineering phases. | ✅ Done |
| 5 | Update `docs/BIBLICAL_TESTRA.md` | Keep the engineering handbook current with Phase 3 routes and auth. | Notification routes added, SSE query-token auth noted, doc index/audit reports referenced. | ✅ Done |
| 6 | Update `testra/README.md` | Point new users to the documentation index and handbook. | Doc section added at top of root README. | ✅ Done |
| 7 | Refresh `docs/PROJECT_OVERVIEW.md` and `FEATURE_MATRIX.md` | Reflect Phase 3.5 completions (API keys UI, SSE auth, MFA QR, project key, onboarding slug). | Stale rows updated; fixed issues no longer marked as broken. | ✅ Done |
| 8 | Update `docs/ROUTES.md` | SSE auth caveat is outdated. | Caveat now states query-token auth is supported as an MVP workaround. | ✅ Done |
| 9 | Refresh `docs/archive/merged-sources/frontend-audit.md` | Lists fixed issues (MFA QR, project key, SSE auth, onboarding slug) as current blockers. | Fixed items moved to a resolved section. | ✅ Done |
| 10 | Refresh `docs/archive/merged-sources/functional-audit.md` | P0 broken list is stale after Phase 3.5. | Re-triage and mark resolved issues. | ✅ Done |
| 11 | Add stale headers to superseded reports | Prevent engineers from using outdated reports. | `ENGINEERING_DOCUMENTATION_REPORT.md`, `TESTRA_ENGINEERING_HANDOVER_REPORT.md`, `DOCUMENTATION_HEALTH_REPORT.md` clearly marked. | ✅ Done |
| 12 | Expand app READMEs | One-line stubs in `apps/*` do not orient engineers. | `api`, `web`, `worker`, `ml` READMEs link to canonical docs and state current status. | Out of scope for docs-only pass |

### Short-term (Phase 4 kickoff)

| # | Deliverable | Why | Success criteria | Owner |
|---|-------------|-----|------------------|-------|
| 13 | Incident response runbook in `docs/operations/` | `TROUBLESHOOTING_GUIDE.md` covers triage but not declared incidents. | Document severity levels, escalation, comms, rollback, and post-mortem procedures. | Platform / SRE |
| 14 | Archive or redirect superseded documents | Reduce confusion and duplicate sources. | Pre-implementation draft and old engineering reports moved to `docs/archive/` or contain redirects. | Docs / Platform |
| 15 | Final pass on `archive/merged-sources/` audit files | Ensure `PROJECT_OVERVIEW.md`, `FEATURE_MATRIX.md`, `frontend-audit.md`, `functional-audit.md`, `backend-audit.md` are consistent with Phase 3.5. | No fixed issue listed as broken; all statuses match code/OpenAPI. | Docs / Engineering |
| 16 | Document Phase 4 scope in `docs/engineering/ROADMAP.md` and BIBLICAL | As Phase 4 starts, update canonical sources. | New modules (defects, API testing, integrations) and their dependencies recorded. | Engineering Lead |
| 17 | Expand API key CI/CD guide | API keys exist but no integration guide for CI. | Document `X-API-Key` or `Authorization: ApiKey` usage for `/ingest` once the middleware is wired. | API / Platform |

### Medium-term (Phase 4 execution)

| # | Deliverable | Why | Success criteria | Owner |
|---|-------------|-----|------------------|-------|
| 18 | Update OpenAPI for Phase 4 endpoints | OpenAPI is the source of truth for API behavior. | Defects, API testing, integrations, webhooks added as they are implemented. | API / Backend |
| 19 | Update `docs/architecture/DATABASE_GUIDE.md` and BIBLICAL data model | New Phase 4 entities (defects, test plans, API test cases). | ERD includes new entities; BIBLICAL schema groups table updated. | Data / Backend |
| 20 | Create module READMEs in `apps/api/internal/<module>/` | Each module should document its ports and dependencies. | New modules have `README.md` explaining domain, service, and handler entry points. | Backend Engineers |
| 21 | Frontend state architecture ADR | `localStorage`-only state is a known limitation; a decision should be recorded. | ADR compares Zustand / React Context / TanStack Query and picks one. | Frontend Lead |
| 22 | SSE authentication hardening ADR | Query-token auth for `EventSource` is an MVP workaround. | ADR records the long-term approach (signed SSE token, cookie, or `fetch` streaming). | Backend / Security |
| 23 | Document CI/CD integration plugins | Phase 4 adds GitHub Actions / GitLab / Jenkins plugins. | Runbook for each plugin with examples. | Integrations |

### Long-term (Phase 5 and beyond)

| # | Deliverable | Why | Success criteria | Owner |
|---|-------------|-----|------------------|-------|
| 24 | ClickHouse schema and analytics runbook | Analytics and dashboards require ClickHouse. | Schema, ingestion pipeline, query patterns, and operational guide documented before launch. | Data / Analytics |
| 25 | Single-Ubuntu-VPS systemd + nginx deployment runbooks | Production deployment needs concrete service units and nginx configs. | Step-by-step runbooks for systemd services, Nginx, PostgreSQL, Redis, MinIO on Ubuntu. | Infra / SRE |
| 26 | Generated TypeScript SDK + README | Public API and partner integrations depend on SDK. | `packages/sdk/` is generated from OpenAPI and has usage examples. | API / Platform |
| 27 | Security and compliance runbooks | SOC 2 readiness requires documented evidence. | Penetration test, vulnerability management, access review, and audit export runbooks. | Security |
| 28 | Enterprise features docs | SSO/SAML, SCIM, data residency require dedicated guides. | Create an `enterprise/` documentation directory with setup and configuration guides. | Enterprise |

### Process improvements

| # | Improvement | Why | Target | Owner |
|---|-------------|-----|--------|-------|
| 29 | OpenAPI validation in CI | Prevent drift between code and contract. | CI job runs `redocly lint` or `swagger-codegen validate` on every PR. | Platform |
| 30 | Markdown internal link checker | Broken relative links degrade discoverability. | CI job runs `markdown-link-check` or `lychee` on every PR. | Platform |
| 31 | Mermaid render test | Diagram syntax errors are not caught by text review. | CI job renders all `.md` Mermaid diagrams. | Platform |
| 32 | Documentation update gate in PR template | Engineers often forget to update docs. | PR template asks "Which docs were updated?" and links to `docs/README.md`. | Engineering |
| 33 | Quarterly doc health review | Docs decay as code changes. | Re-run scoring in `DOCUMENTATION_CONSOLIDATION_REPORT.md` each quarter and update `ROADMAP.md` documentation roadmap. | Docs / Engineering Lead |

### Definition of done for documentation

Documentation is "done" for a feature when:

1. OpenAPI paths and schemas are updated (`docs/api/openapi/openapi.yaml`).
2. BIBLICAL route table and data model sections are updated.
3. `docs/engineering/ROADMAP.md` is updated if the feature changes phase status.
4. `docs/README.md` (Documentation Index) is updated if new docs are added.
5. `docs/engineering/ROADMAP.md` and `docs/PROJECT_OVERVIEW.md` are updated if the change resolves or creates debt.
6. A new ADR is created if the change affects an accepted architectural decision.
7. `docs/archive/superseded/DOCUMENTATION_CONSOLIDATION_REPORT.md` is re-scored if a doc materially changes.

### Links

- [`DOCUMENTATION_CONSOLIDATION_REPORT.md`](../archive/superseded/DOCUMENTATION_CONSOLIDATION_REPORT.md) — consolidation summary, scores, and top issues.
- [`docs/README.md`](../README.md) — canonical documentation index.
- [`docs/BIBLICAL_TESTRA.md`](../BIBLICAL_TESTRA.md) — engineering handbook.
- [`docs/engineering/ROADMAP.md`](ROADMAP.md) — implementation phase status.


## Technical Debt Register

> This document merges every known issue from the handover audits into a single master register. For module-specific findings and recommendations, see [`backend-audit.md`](../archive/merged-sources/backend-audit.md), [`frontend-audit.md`](../archive/merged-sources/frontend-audit.md), and [`infra-audit.md`](../archive/merged-sources/infra-audit.md).

### Severity legend

| Severity | Meaning |
|----------|---------|
| **Critical** | Blocks production launch or exposes data/customers to serious risk |
| **High** | Significant functional or security gap; must fix before public beta |
| **Medium** | Real impact but can be deferred until critical/high items are resolved |
| **Low** | Polish, cleanup, or future optimization |

---

### Critical

| # | Description | Impact | Affected modules | Recommended fix | Priority | Current status |
|---|-------------|--------|------------------|-----------------|----------|----------------|
| C1 | **Rate limiting is now wired** on `/auth/*` and `/ingest`. Redis-backed `LocalRateLimiter` supports per-IP, per-email, and per-API-key hash keys. | Brute force / DoS risk reduced | `identity`, `shared/middleware` | Scale to a distributed Redis rate limiter and add more granular public endpoint rules | P0 | Closed |
| C2 | **API-key authentication middleware is implemented and wired to `/ingest`.** Scoped keys validate hash, expiry, scopes, and set tenant context. | CI/CD can use revocable keys instead of user JWTs | `apikeys`, `automationhub`, `shared/middleware` | Document `X-API-Key` / `Authorization: ApiKey` usage in CI guide and OpenAPI | P0 | Closed |
| C3 | **~~SSE progress stream inaccessible from browsers.~~** Query-token auth is implemented: `GET /test-runs/{id}/stream?access_token=${token}` works in browsers. Harden before public networks. | Manual test execution live progress works locally/MVP | `results`, `frontend` | Replace query-token with session cookie or short-lived signed SSE token before public networks | P0 | Closed |
| C4 | **Startup secret validation added.** `config.Validate()` fails fast in production when `JWT_SECRET` is weak/default/short or `DATABASE_URL` uses example credentials or disables TLS; `.env.example` values remain for local dev only. | Production deployments can no longer boot with known credentials | `infra`, `config` | Integrate with a secrets manager (Vault/environment files or a local secrets store) for rotation; still remove any real defaults from CI | P0 | Mitigated |
| C5 | **Client-side route guards added.** `DashboardLayout` and `AuthLayout` redirect unauthenticated users. Token refresh on 401 is implemented. | `localStorage` still exposes tokens to XSS; risk partially mitigated | `frontend` | Move access token to `httpOnly` cookie or secure wrapper before production launch | P0 | Mitigated |
| C6 | **Audit persistence failures are now logged, not silently dropped**, and run under a bounded 5s detached context. Still best-effort (no transactional write or durable queue). | Lost compliance events are observable in logs | `audit`, `shared/middleware` | Write audit events synchronously inside the request transaction or enqueue to a durable queue (Redis/Asynq) | P0 | Mitigated |

---

### High

| # | Description | Impact | Affected modules | Recommended fix | Priority | Current status |
|---|-------------|--------|------------------|-----------------|----------|----------------|
| H1 | **Permission-name drift between RBAC seed and middleware.** Middleware checks `orgs:read` but seed names include `organization:read`; `members:manage` vs `organization_members:manage`. | Authorization checks silently fail or succeed incorrectly | `rbac`, `shared/middleware`, all modules | Create a single source-of-truth `Permission` constants package; align migration seeds and middleware usage | P1 | Open |
| H2 | **`POST /organizations` and `GET /organizations` lack permission/membership gates.** | Any authenticated user can create or list organizations without restriction | `organization`, `shared/middleware` | Add `RequirePermission("orgs:create")` and filter `GET` by `organization_members.user_id` | P1 | Open |
| H3 | **Tenant resolution can pick the wrong tenant.** `TenantContext` checks `organization_id`, `workspace_id`, `project_id`, `api_key_id`, `run_id` in a fixed order from body/query; a malicious request can supply a valid member org ID plus a different `workspace_id` belonging to another org. | Cross-tenant data access if other checks are skipped | `shared/middleware`, `shared/tenant` | Resolve tenant from the **least-privileged, most-specific** resource and verify membership before every operation | P1 | Open |
| H4 | **~~Project key generation mismatch.~~** Frontend now generates uppercase alphanumeric keys matching the backend regex. | Users can create projects with auto-generated keys | `project`, `frontend` | Add unique constraint/index on `(organization_id, key)` | P1 | Closed |
| H5 | **Token refresh on 401 is implemented in `lib/api.ts`.** The client stores refresh tokens and retries once after a 401. | Users are no longer logged out every 15 minutes | `frontend`, `identity` | Move token storage to `httpOnly` cookie and add proactive refresh before expiry | P1 | Closed |
| H6 | **~~MFA QR code rendered as text.~~** Frontend now renders `qr_code` as an `<img src={qr_code} />` data URL. | Users can scan QR code with authenticator apps | `frontend`, `identity` | Add copy-to-clipboard fallback | P1 | Closed |
| H7 | **No production systemd service units, nginx config, or deployment runbooks.** | Cannot deploy to production safely; no health checks or secret management | `docs/deployment/` | Create systemd unit files, nginx site config, health checks, and environment file guidance | P1 | Open |
| H8 | **No VPS provisioning or repeatable deployment scripts.** | Infrastructure cannot be provisioned consistently | `docs/deployment/`, `scripts/` | Document server setup, firewall, PostgreSQL/Redis/MinIO install, and CI/CD artifact delivery | P1 | Open |
| H9 | **No integration or end-to-end test suite.** | Regression risk, low confidence in refactors | `all` | Add Go integration tests against test DB, Playwright E2E for critical flows | P1 | Open |
| H10 | **No observability stack.** No OpenTelemetry, Prometheus, Grafana, Loki, or structured request logging. | Cannot debug production incidents | `shared`, `docs/operations/` | Add OpenTelemetry traces, structured logs, and metrics endpoints | P1 | Open |

---

### Medium

| # | Description | Impact | Affected modules | Recommended fix | Priority | Current status |
|---|-------------|--------|------------------|-----------------|----------|----------------|
| M1 | **Duplicate dashboard route trees.** `/dashboard/*` and `/:workspace/*` re-export the same pages. | Confusing URLs, inconsistent context, broken analytics | `frontend` | Consolidate on `/:workspace` tree; redirect `/dashboard` to last selected workspace | P2 | Open |
| M2 | **Frontend relies entirely on `localStorage` for state.** | Hydration mismatches, no reactive cross-page state, repetitive `typeof window` guards | `frontend` | Introduce Zustand or React Context for auth/workspace state; keep `localStorage` only for persistence | P2 | Open |
| M3 | **OpenAPI spec is at version 0.4.0 and missing modules.** | Defects, analytics, billing, integrations still missing; notifications now documented | `docs/api/openapi` | Complete spec incrementally; generate SDK from it | P2 | Open |
| M4 | **`apps/worker` is a stub and `cmd/worker` does nothing.** | No background processing for email, reports, or ML jobs | `worker`, `apps/api/cmd/worker` | Implement Asynq worker or promote `cmd/worker` to real background service | P2 | Open |
| M5 | **Settings sub-pages are mostly placeholders.** | Cannot manage members, roles, API keys, audit logs, billing; notifications page implemented | `frontend` | Build settings pages using existing `platform/api.ts` and backend endpoints | P2 | Open |
| M6 | **Audit middleware does not capture request result/status before response.** | Audit metadata may be incomplete if handler errors | `audit`, `shared/middleware` | Use `responseWriter` wrapper to capture status code in middleware | P2 | Open |
| M7 | **`mailpit` health check is not wired into `pnpm dev`.** | Service startup order is not verified | `scripts/dev/` | Add a local health check for Mailpit in `start-infra.mjs` | P2 | Open |
| M8 | **No single `start` script for all backing services.** | Developers must start PostgreSQL, Redis, Mailpit, MinIO separately | `scripts/dev/` | Provide optional native start/stop helper scripts | P2 | Open |
| M9 | **No bulk import / migration tools from TestRail/Excel/CSV.** | High friction for customer onboarding | `testmanagement`, `frontend` | Build import endpoints and UI wizards | P3 | Open |
| M10 | **Environment variable name mismatch.** `apps/api/.env.example` has `JWT_EXPIRY_HOURS=168` while code reads `JWT_EXPIRY_MINUTES`. | Config silently falls to default 15 minutes or fails to parse | `config`, `scripts/` | Fix `.env.example` and add startup validation for required vars | P2 | Open |

---

### Low

| # | Description | Impact | Affected modules | Recommended fix | Priority | Current status |
|---|-------------|--------|------------------|-----------------|----------|----------------|
| L1 | **Frontend `apiFetch` does not retry or handle network errors gracefully.** | Transient errors show raw error text | `frontend` | Add exponential backoff for idempotent GETs and user-friendly error toasts | P4 | Open |
| L2 | **Project `key` field is limited to 10 uppercase characters.** | Some teams may want longer/more readable keys | `project` | Relax or configure key rules | P4 | Open |
| L3 | **`.dockerignore` remains in `apps/api`.** | Not needed because Docker is not used | `apps/api` | Remove `.dockerignore` | P4 | ✅ Closed |
| L4 | **No global error boundary or loading state.** | Each page duplicates loader/error handling | `frontend` | Add `error.tsx` and `loading.tsx` in route groups and feature wrappers | P4 | Open |
| L5 | **`localStorage` keys are not namespaced.** | Risk of collision with other apps or future keys | `frontend` | Prefix with `testra_` consistently (already partially true) | P5 | Open |
| L6 | **Stale or sample files remain in `.github/workflows` and documentation.** | Confusion for new engineers | `.github`, `docs/` | Remove sample placeholders or document them | P5 | Open |

---

### Consolidated priority view

| Priority | What to fix |
|----------|--------------|
| **P0 — Critical** | ✅ Rate limiting, ✅ API-key auth, SSE auth, secrets management, ✅ frontend route guards + token refresh, audit durability |
| **P1 — High** | Permission-name alignment, org permission gates, tenant resolver, ✅ token refresh (completed), systemd/nginx deployment runbooks, tests, observability |
| **P2 — Medium** | Route consolidation, global state, OpenAPI completion, worker implementation, settings pages, audit middleware, env vars, local service helper scripts |
| **P3+ — Lower** | Bulk import, advanced UX polish, error boundaries, key namespaces, cleanup |

For the exact implementation order, see this `ROADMAP.md` file.

---

## Production Launch Roadmap — 2026-08-02 Update

**Status:** Active planning  
**Source of truth for launch readiness:** `docs/engineering/LAUNCH_READINESS_PLAN.md`  
**Detailed task backlog:** `docs/engineering/SPRINT_BACKLOG.md`  

| Milestone | Target | Definition of Done | Status |
|-----------|--------|--------------------|--------|
| **M1 — Production Security & Trust** | 2026-09 | Cookie/session auth, CSRF, hardened password policy, audit read, PII redaction, host firewall rules, secrets store | Not started |
| **M2 — Production Infrastructure & Deploy** | 2026-10 | Ubuntu VPS provisioning runbook, systemd service units, nginx TLS, PostgreSQL/Redis/MinIO setup, GitHub Actions artifact delivery, secrets store | Not started |
| **M3 — Observability & Reliability** | 2026-10 | OpenTelemetry, Grafana/Tempo/Loki, SLO dashboards, alerting, incident runbooks | Not started |
| **M4 — Commercial SaaS Core** | 2026-11 | Stripe billing, entitlements, OpenAPI-generated SDK, member/role UI, audit UI, admin console | Not started |
| **M5 — Data & Performance at Scale** | 2026-12 | Missing DB indexes, pagination remaining lists, retention jobs, SSR/caching, load tests | Not started |
| **M6 — Enterprise & Phase 4+ Features** | 2027-Q1 | SSO/SAML/SCIM, custom roles, data residency, advanced intelligence, partner marketplace | Not started |

### Launch Gates

- **Alpha Ready:** Build/test gates pass; core auth, test management, runs, ingestion work end-to-end on a local developer machine.
- **Beta Ready:** Cookie auth, real single-Ubuntu-VPS staging with TLS, OpenAPI/SDK, observability, pagination/indexes in place; first design partners onboarded.
- **GA Ready:** Production single-Ubuntu-VPS deployment, DB backups, billing/entitlements, security audit, SLO monitoring, load tests passed.
- **Enterprise Ready:** SSO/SAML/SCIM, custom roles, audit export, data residency, SLA reporting.

### Phase Status Update

- **Phase 4 — API Testing & Defects:** Backend routes wired; UI and full business logic remain partial. Continue in parallel with M5/M6.
- **Phase 5 — Dashboard, Analytics & Launch:** Re-scoped into M1–M5 of the production launch roadmap.
- **Phase 6 — V2 Intelligence:** Moved to M6 enterprise/scale phase.

## See Also

- [`BIBLICAL_TESTRA.md`](../BIBLICAL_TESTRA.md) — canonical engineering handbook
- [`PROJECT_OVERVIEW.md`](../PROJECT_OVERVIEW.md) — product vision and current state
- [`FEATURE_MATRIX.md`](../FEATURE_MATRIX.md) — feature completion matrix
