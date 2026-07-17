# Project Overview

**Purpose:** Engineering-facing product overview: vision, goals, target users, MVP scope, and current state.
**Owner:** Product / Engineering Lead
**Scope:** Product vision, target users, MVP scope, repository status, and architecture summary.
**Source of Truth:** PROJECT_OVERVIEW.md for engineering overview; root product docs (`testra-master-context.md`, `testra-product-strategy.md`, `testra-brd.md`) for product strategy.
**Last Updated:** July 2026
**Related documents:**
- [`BIBLICAL_TESTRA.md`](BIBLICAL_TESTRA.md)
- [`ROADMAP.md`](engineering/ROADMAP.md)
- [`FEATURE_MATRIX.md`](FEATURE_MATRIX.md)
- [`testra-master-context.md`](../../testra-master-context.md)

> Source documents: `testra-master-context.md`, `testra-product-discovery.md`, `testra-product-strategy.md`, `testra-brd.md`. For accepted architecture decisions, see [`architecture/adrs/`](architecture/adrs/); `04_Architecture/testra-software-architecture-decisions.md` is a pre-implementation draft and should not be treated as current.

## Product vision

> **To become the single platform where every software team manages, executes, and understands their quality engineering — replacing the fragmented tools of the past with one intelligent, unified experience.**

Testra unifies test management, API testing, automation result ingestion, defect tracking, reporting, and analytics into one modern SaaS platform. The ultimate goal is to replace the 4–7 disconnected tools that engineering teams currently maintain.

## Product mission

> **Unify software testing — eliminating the tool sprawl that slows teams down — by delivering a modern, intelligent platform that makes quality engineering faster, clearer, and more impactful for every team.**

## Product goals

### Year 1 — Foundation / MVP

- Launch a publicly available product with core test management, API testing, automation ingestion, and reporting.
- Acquire 50–100 paying customers and reach **$500K–$1M ARR**.
- Establish 3–5 design partners in Southeast Asia.
- Achieve SOC 2 Type II readiness.

### Year 2 — Growth / Version 2

- Reach **$3–5M ARR** and first $50K+ ACV enterprise contracts.
- Launch quality intelligence: flaky-test detection, failure classification, risk scoring, coverage heatmaps, and release-readiness reports.
- Launch team/enterprise tier pricing.

### Year 3 — Scale / Series A

- Reach **$15–25M ARR**.
- Expand to 3+ geographic markets.
- Build public API and partner marketplace.

## Target users

### Primary ICP

- **Mid-market SaaS, fintech, e-commerce, and logistics companies in Southeast Asia.**
- 50–200 employees, 20–200 engineers, 3–30 QA engineers.
- Currently using TestRail/Zephyr + Postman + Playwright/Cypress + Jira + spreadsheets.

### Secondary ICP

- **Enterprise engineering organizations** in banking, insurance, government, healthcare, and large SaaS across APAC.
- 2,000–100,000 employees, 200–5,000 engineers, 30–500 QA engineers.

### Personas

| Persona | Role | Top pain | What success looks like |
|-----------|------|----------|------------------------|
| **QA Lead — "Alex"** | Leads 3–15 QA engineers | 30%+ of time spent on reporting and coordination; no single quality dashboard | Live quality health dashboard, automated CI/CD runs, clear defect trends |
| **Automation Engineer — "Sam"** | Maintains test automation | Flaky tests, results buried in CI logs, no link to test cases | All automated results in one place with automatic flaky-test identification |
| **Manual QA — "Jordan"** | Executes test plans | Test cases scattered across Excel/TestRail, repetitive defect logging | Organized cases with traceability and fast defect creation |
| **Engineering Manager — "Morgan"** | Manages engineering | Late quality data, gut-feel release decisions, no QA ROI | Real-time release-readiness dashboard and trend data |
| **Backend Developer — "Taylor"** | Builds APIs | Test results not surfaced in dev workflow, unclear ownership | Test results in PRs and clear failure ownership |

## Current MVP scope

The MVP is defined as the set of capabilities needed for the first commercial launch. The repository currently implements **a subset** of the full MVP:

### Implemented in code

- **Identity** — email/password registration and login, JWT access/refresh tokens, TOTP MFA enrollment/verify/disable, password reset.
- **Organization** — create and list organizations, add members.
- **Workspace** — create and list workspaces within an organization.
- **Project** — create and list projects within a workspace.
- **Test Management** — test folders, test suites, test cases (with steps, tags, versions), full-text search.
- **Test Runs / Results** — create manual test runs, update run item status, view run progress, list run history.
- **Automation Hub** — ingest JUnit XML and Playwright/Cypress JSON payloads into test runs.
- **RBAC** — system roles (owner/admin/qa_engineer/viewer), permission catalog, role assignments scoped to organization.
- **Audit** — fire-and-forget audit event persistence.

### MVP scope not yet implemented

- **Defect Management** — placeholder directory only.
- **API Testing** — engine and UI not built.
- **CI/CD pipeline integrations** — ingestion endpoint exists, but no native GitHub Actions/GitLab/Jenkins plugins.
- **Requirements traceability matrix** — not implemented.
- **Advanced reporting** — dashboard is a skeleton.
- **SSO / SAML** — not implemented.
- **Billing and entitlements** — placeholder directory only.
- **Bulk import / migration tools** — not implemented.

## Completed phases

| Phase | Status | Notes |
|-------|--------|-------|
| Product Discovery | ✅ Complete | Personas, pain points, market opportunity, competitive positioning documented |
| Product Strategy / BRD | ✅ Complete | Roadmap, monetization, go-to-market, success metrics documented |
| Product Architecture Strategy | ✅ Complete | Module decomposition, principles, layer map, squad assignments documented |
| Software Architecture Decisions | ✅ Approved | Stack and architecture recorded as ADR-001 through ADR-012; pre-implementation draft archived |
| Repository bootstrap | ✅ Complete | Monorepo structure, tooling, dev environment, CI baseline in place |
| Platform backend | ✅ Functional | Identity, organization, workspace, project, RBAC, audit |
| Testing backend | ✅ Functional | Test management, results, automation ingestion |
| Web frontend | 🔄 Partial | Auth, onboarding, dashboard skeleton, test cases, test runs |
| Infrastructure | 📝 Scaffolding | Docker, Kubernetes, Terraform scaffold; not production-ready |

## Repository status

- **Monorepo:** `apps/api`, `apps/web`, `apps/worker`, `apps/ml`, `packages/*`, `infra/*`.
- **Backend:** Go 1.23 modular monolith with Clean Architecture boundaries.
- **Frontend:** Next.js 15 App Router, TypeScript 5, TailwindCSS, react-hook-form + Zod.
- **Database:** PostgreSQL 16 with Row-Level Security; ClickHouse and Redis present but not yet used by application code.
- **Tests:** Go unit tests present; integration and frontend test coverage minimal.
- **CI/CD:** GitHub Actions builds and lints Go, builds/type-checks web, lints Python ML; no deployment pipeline.

## Roadmap summary

```
MVP (Year 1)        → Version 2.0 (Year 2)      → Enterprise Edition (Year 2+)   → Version 3.0 (Year 3)
Identity/Org/Workspace   Flaky-test detection        Data residency                    Marketplace
Test Management        Failure classification        Advanced compliance               Public API / SDK
API Testing (basic)    Risk scoring / health score   SSO / SAML enhancements           Predictive analytics
Automation Ingestion   Coverage heatmaps             SLA / dedicated support           Cross-project governance
Defects               Release readiness             Advanced audit export             Multi-region scale
Reporting / Dashboards Custom report builder                                             Regional expansion
Notifications          Custom fields / tags
RBAC / Audit / SSO
```

## Current completion estimate

These estimates are directional and based on functional coverage and production readiness, not line count.

| Area | Estimate | Status |
|------|----------|--------|
| Backend MVP features | ~70% | Core flows work; missing defects, billing, CI/CD integrations (notifications now implemented) |
| Frontend MVP features | ~50% | Auth, onboarding, test cases, test runs are partial; settings/dashboard placeholders reduced with notifications UI |
| Infrastructure / DevOps | ~25% | Local dev works; K8s/Terraform/CD are scaffolded but incomplete |
| OpenAPI / SDK | ~55% | Spec covers auth, orgs, workspaces, projects, test management, test runs, ingestion, notifications; missing defects, analytics, billing, integrations |
| Test coverage | ~20% | Unit tests only; no integration or E2E test suite |
| Production readiness | ~15% | No rate limiting on auth, no API-key auth, no secrets management, no monitoring, no CD |
| **Overall MVP** | **~40% complete** | Not ready for customer-facing production deployment |

## Major engineering decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Architecture pattern** | Modular monolith with Clean Architecture | Solo-developer velocity, enterprise-ready from MVP, clear extraction boundaries |
| **Backend language** | Go 1.23 | Compiled, memory-efficient, concurrency-native, simple deployment |
| **Frontend framework** | Next.js 15 App Router + TypeScript | Modern React, SSR for dashboards, large hiring market |
| **Primary database** | PostgreSQL 16 | ACID, RLS, JSONB, proven multi-tenant SaaS choice |
| **Analytics store** | ClickHouse 24 | High-ingest time-series result data (not yet wired) |
| **Cache / queue** | Redis 7 | Sessions, rate limiting, future Asynq job queue |
| **ML runtime** | Python 3.12 + FastAPI + scikit-learn/XGBoost | Transparent, explainable, no external LLM dependency |
| **Multi-tenancy** | Shared DB, shared schema, `tenant_id` column + PostgreSQL RLS | Simplest operational model that still supports enterprise isolation |
| **Authentication** | JWT access tokens + opaque refresh-token families + TOTP MFA | Supports session security and enterprise MFA requirements |
| **API style** | RESTful JSON, URL versioning `/api/v1`, OpenAPI 3.1 spec | Familiar to QA engineers and enterprise integrators |
| **Real-time** | Server-Sent Events (SSE) for test-run progress | Simpler horizontal scaling than WebSockets for one-way progress streams |
| **Deployment target** | Kubernetes on AWS (EKS) for production; native/local dev first | Enterprise-ready path; local dev skips Docker for speed |

## Major ADR / architecture summary

Key decisions recorded in `04_Architecture/testra-software-architecture-decisions.md`:

1. **Modular monolith over microservices** — avoid distributed-system overhead while keeping module boundaries clean for future extraction.
2. **Monorepo** — atomic changes across API, web, SDK, and infrastructure; one CI pipeline.
3. **Clean / Hexagonal backend** — every module has `domain → ports → repository/service → handler` layers.
4. **Native development environment** — `pnpm dev` starts local services directly; Docker optional for local development.
5. **API-first** — the web UI is one client; CI/CD ingestion and future public API use the same endpoints.
6. **Zero customer code retention** — Testra ingests test results and metadata, never customer source code or API collections.
7. **No external LLM dependency** — intelligence uses classical ML/statistical models trained per tenant on tenant-owned data.
8. **Customer data ownership** — per-tenant models, tenant isolation by RLS, data portability first-class.
9. **Enterprise-ready from day one** — RBAC, audit trails, SSO/SAML path, SOC 2-aligned deployment.
10. **APAC-first, global-ready** — localization, multi-currency, data residency, and regional integrations built into the architecture.


## Current State

> This document describes the repository **exactly as it exists today**. It is the handover snapshot a new engineer reads before writing any code.

### Maturity summary

- **Phase:** Pre-MVP hardening / engineering handover.
- **Backend:** Core platform and test-management flows are functional locally.
- **Frontend:** Auth, onboarding, dashboard skeleton, test cases, and test runs are partially implemented.
- **Infrastructure:** Scaffolding exists; not production-ready.
- **Production readiness:** **Not ready** for customer-facing deployment.

---

### Completed modules

These modules have working backend code, database migrations, and at least partial frontend support.

| Module | Backend | Frontend | Notes |
|--------|---------|----------|-------|
| **Identity** | ✅ | ✅ | Register, login, JWT, refresh tokens, MFA TOTP, password reset |
| **Organization** | ✅ | ✅ | Create/list organizations; onboarding flow |
| **Workspace** | ✅ | ✅ | Create/list workspaces within an organization |
| **Project** | ✅ | ✅ | Create/list projects; project key validation |
| **Test Management** | ✅ | ✅ | Folders, suites, test cases, versions, full-text search; frontend case list/create/edit/versions |
| **Test Runs / Results** | ✅ | ✅ | Manual runs, run items, SSE progress stream (query-token auth); UI list/detail/start |
| **Automation Hub** | ✅ | ❌ | Ingest JUnit/Playwright/Cypress payloads into runs; no UI |
| **API Keys** | ✅ | ✅ | Backend CRUD and `/dashboard/settings/api-keys` UI for create/revoke; not yet used for `/ingest` auth |
| **RBAC** | ✅ (seed/loader) | ❌ | Roles/permissions seeded; org-scoped enforcement only |
| **Audit** | ✅ | ❌ | Fire-and-forget audit events; no UI |
| **Notifications** | ✅ | ✅ | In-app feed, unread count, mark read/unread, preferences, channels, RLS |
| **OpenAPI Spec** | 📝 | N/A | Covers auth, orgs, workspaces, projects, test mgmt, runs, ingest, notifications |

---

### Partially completed modules

These have scaffolding or partial implementation but are not shippable.

| Module | What works | What is missing / broken |
|--------|------------|--------------------------|
| **Dashboard** | Landing page, sidebar, workspace/project context from `localStorage` | Real widgets, role-based views, recent activity feed, intelligence teasers |
| **Settings** | Settings shell and navigation | Most sub-pages are still placeholders; **notifications** and **API keys** settings pages are implemented |
| **OpenAPI** | Core routes documented | Defects, analytics, billing, integrations, webhooks still missing; notifications documented |
| **Test Suites / Folders UI** | Backend list endpoints and API wrappers | No user interface to create/manage suites or folders |
| **CI/CD** | GitHub Actions lint/build | No image push, no deployment, no integration tests |
| **Kubernetes** | Base deployment + service manifests | No ConfigMap/Secret, probes, resources, ingress, web/worker/migrator |
| **Terraform** | Provider/backend scaffold | No modules or resources defined |

---

### Placeholder pages

The frontend contains pages or directories that render empty cards or placeholder text.

| Route / Directory | File | Note |
|-------------------|------|------|
| `/dashboard/defects` | `app/(dashboard)/[workspace]/defects/page.tsx` | `PlaceholderPage` with "Planned for Phase 4" |
| `/dashboard/settings/*` | `app/(dashboard)/dashboard/settings/*/page.tsx` | API keys and notifications pages implemented; most other tabs are placeholders |
| `/dashboard/api-tests` | `app/(dashboard)/[workspace]/api-tests/page.tsx` | Not implemented; may be empty |
| `apps/web/features/api-testing/` | `.gitkeep` or empty | No API testing frontend |
| `apps/web/features/analytics/` | Empty | No analytics frontend |
| `apps/web/features/defects/` | Empty | No defects frontend |

---

### Unimplemented modules

These modules exist as directories with only a `.gitkeep` or no code at all.

| Module | Backend directory | Frontend directory | Why it matters |
|--------|-------------------|--------------------|----------------|
| **Defects** | `apps/api/internal/defects/` | `apps/web/features/defects/` | Core MVP feature; required for issue tracking |
| **API Testing** | `apps/api/internal/apitesting/` | `apps/web/features/api-testing/` | Displaces Postman; key differentiator |
| **Billing / Subscriptions** | `apps/api/internal/billing/` | `apps/web/features/settings/` | Required for commercial launch |
| **Integration Hub** | `apps/api/internal/integrationhub/` | N/A | Jira, GitHub, GitLab, CI/CD webhooks |
| **Analytics** | `apps/api/internal/analytics/` | `apps/web/features/analytics/` | Reporting and dashboards beyond skeleton |
| **Intelligence** | `apps/api/internal/intelligence/` | `apps/web/features/analytics/` | Flaky detection, risk scoring, failure classification (V2) |
| **Marketplace** | N/A | N/A | V3 ecosystem play |
| **Public API / SDK** | N/A (uses `/api/v1`) | `packages/sdk/` | Expose stable API for customers/partners |
| **SSO / SAML** | N/A | N/A | Enterprise hard requirement |
| **Bulk import / migration** | N/A | N/A | TestRail/Excel/CSV migration |

---

### Broken features

These features compile and run but have known runtime defects that block production use.

| Feature | Problem | Evidence / location |
|---------|---------|---------------------|
| **Auth token refresh (client)** | Access tokens expire in 15 minutes; the frontend does not call `/auth/refresh` on 401. Users must re-login. | `frontend-audit.md` §10, `ROADMAP.md` §P1 |

> Previously reported broken items (SSE browser auth, MFA QR display, project key generation, onboarding slug) are resolved as of Phase 3.5.

---

### Known limitations

| Limitation | Impact | Where to fix |
|------------|--------|--------------|
| **No API-key authentication middleware** | CI/CD ingestion currently requires a user JWT; scoped API keys are stored but not used for auth. | `backend-audit.md` §11.1, `ROADMAP.md` §Technical Debt Register |
| **Rate limiting is unconfigured** | `LocalRateLimiter` is instantiated but never wired to routes. Auth endpoints are vulnerable to brute force. | `backend-audit.md` §11.2 |
| **Audit logging is fire-and-forget** | `auditSvc.Log(context.Background(), ...)` runs in a background context with no retry; audit events can be lost. | `backend-audit.md` §11.5 |
| **Permission scope is organization-only** | Workspace/project-level roles are not supported; `RequirePermission` always uses `scope_type = 'organization'`. | `backend-audit.md` §11.4 |
| **No route guards in frontend** | Unauthenticated users can briefly see dashboard pages and only fail on API calls. | `frontend-audit.md` §10.7 |
| **No global state management** | `localStorage` is the only state layer; causes hydration mismatches and repetitive `typeof window` guards. | `frontend-audit.md` §5, §10.6 |
| **Duplicate dashboard route trees** | `/dashboard/*` and `/:workspace/*` both exist; same components served from different URLs with different context logic. | `frontend-audit.md` §3.2, §10.8 |
| **No error boundary / global loading** | Each page handles loading/error independently. | `frontend-audit.md` §10.10 |
| **No production-ready infrastructure** | Kubernetes/Terraform are scaffolds; no secrets, probes, ingress, or CD. | `infra-audit.md` |
| **Worker is a stub** | `apps/worker` and `apps/api/cmd/worker` are empty. No background processing exists. | `infra-audit.md` finding #12 |
| **ClickHouse / Redis / MinIO not used by app** | Provisioned in Docker Compose but no application code connects to them. | `infra-audit.md` |
| **Environment variable drift** | `apps/api/.env.example` has `JWT_EXPIRY_HOURS=168` while code reads `JWT_EXPIRY_MINUTES`. | `infra-audit.md` §6.3 |

---

### Open architecture decisions

| Decision | Options | Current lean | Why it matters |
|----------|---------|--------------|--------------|
| **Identity provider** | Self-hosted Keycloak/FusionAuth vs Clerk/WorkOS | Undecided; current code is self-built JWT + MFA | SSO/SAML/SCIM is a must-have for enterprise; vendor choice affects ops overhead |
| **Frontend state layer** | Keep `localStorage` vs add Zustand/React Context vs TanStack Query | Undecided; currently local state | Required for route guards, refresh-token flow, and hydration safety |
| **SSE authentication** | Session cookie vs signed query-token vs fetch-based streaming | Query-token auth implemented for MVP; harden before production | Must harden before public-network test-run progress |
| **API key auth header** | `Authorization: ApiKey <key>` vs `X-API-Key` | Undecided; currently no middleware | Required for CI/CD ingestion without user JWT |
| **ClickHouse adoption** | Start storing test results in ClickHouse now vs defer to V2 | Deferred; results live in Postgres | Analytics and intelligence depend on this |
| **Worker framework** | Asynq over Redis vs custom Go worker | Undecided; `apps/worker` is a stub | Background jobs needed for notifications, reports, ML inference |
| **Microservices extraction trigger** | Keep modular monolith vs extract ML service/ingestion pipeline first | Keep monolith for now | Decision needed once scale or team size justifies it |
| **Multi-region / data residency** | Schema-per-tenant vs row-level tenant with regional DBs | Row-level tenant in shared schema | Enterprise Edition requirement |
| **SDK generation** | Generate TypeScript SDK from OpenAPI vs hand-write | Undecided; `packages/sdk` exists | Public API and partner integrations depend on this |

---

### Current repository maturity

| Capability | Maturity | Evidence |
|------------|----------|----------|
| **Local development** | ✅ Functional | `pnpm dev`, `make migrate`, Docker Compose for deps |
| **Core backend flows** | ✅ Functional | Unit tests build; manual API calls work |
| **Frontend happy path** | 🔄 Partial | Login → onboarding → create project → test cases works with caveats |
| **Test automation** | ❌ Minimal | Go unit tests only; no integration/E2E suite |
| **Security posture** | ❌ Not production | No rate limiting, no API-key auth, weak `.env.example` |
| **Observability** | ❌ Not present | No OpenTelemetry, Prometheus, Grafana, or centralized logs |
| **Deployment pipeline** | ❌ Not present | CI builds only; no image push or deploy |
| **Documentation** | 🔄 Partial | Product docs complete; engineering wiki being finalized |

#### One-sentence verdict

The Testra repository is a **functional alpha prototype** with a solid Clean Architecture backend and a partial Next.js frontend. It demonstrates the core value loop (auth → org → workspace → project → test cases → test runs → ingestion → notifications), but it requires production hardening, missing MVP modules, and infrastructure completion before any customer-facing launch.

## See Also

- [`BIBLICAL_TESTRA.md`](BIBLICAL_TESTRA.md) — canonical engineering handbook
- [`ROADMAP.md`](engineering/ROADMAP.md) — implementation phases and technical debt
- [`FEATURE_MATRIX.md`](FEATURE_MATRIX.md) — feature completion matrix
- [`ENGINEERING_STANDARDS.md`](engineering/ENGINEERING_STANDARDS.md) — coding and review standards
