# Testra Engineering Wiki

> **One Platform. Every Test.**  
> This directory is the single source of truth for the current Testra codebase, architecture, and engineering handover. It is written for a new senior engineer or another AI picking up the project with no prior conversation.

## What Testra is

Testra is an **Intelligent Quality Engineering Platform** — a B2B SaaS product that unifies test management, API testing, automation result ingestion, defect tracking, reporting, and analytics into one modern platform. It is built APAC-first for mid-market SaaS and regulated enterprise verticals (fintech, banking, insurance, healthcare, government).

The long-term vision is to become the single platform where software teams **manage, execute, and understand** software quality.

## Repository purpose

This monorepo contains the complete Testra platform implementation:

- **`apps/api`** — Go modular-monolith backend (API, worker stub, migrator)
- **`apps/web`** — Next.js 15 web application
- **`apps/worker`** — Standalone Go worker (currently a stub)
- **`apps/ml`** — Python FastAPI ML inference service (placeholder)
- **`packages/*`** — Shared TypeScript types, UI components, config, SDK
- **`infra/*`** — Terraform, Kubernetes, Docker, and CI/CD manifests
- **`docs/handover/`** — This engineering wiki

## Current implementation status

The codebase is in a **pre-MVP hardening phase**. The Platform Layer (Identity, Organization, Workspace, Project) and the core Testing Layer (Test Management, Test Runs, Automation Hub) are implemented end-to-end on the backend. The frontend supports authentication, onboarding, dashboard navigation, test cases, and manual test runs. Several MVP modules are still placeholders and infrastructure is scaffolded but not production-ready.

See [`CURRENT_STATE.md`](CURRENT_STATE.md) and [`FEATURE_MATRIX.md`](FEATURE_MATRIX.md) for exact completion.

## Current engineering phase

- **Phase complete:** Core platform and test-management flows are functional locally.
- **Phase in progress:** Production hardening, security fixes, infrastructure completion, and onboarding a new engineering team.
- **Next major phase:** Defects, Notifications, and CI/CD integration (see [`NEXT_STEPS.md`](NEXT_STEPS.md)).

## Recommended reading order

1. [`PROJECT_OVERVIEW.md`](PROJECT_OVERVIEW.md) — Product context, goals, users, and roadmap.
2. [`ARCHITECTURE.md`](ARCHITECTURE.md) — High-level system, backend/frontend/database architecture, and request flows.
3. [`DATABASE_OVERVIEW.md`](DATABASE_OVERVIEW.md) — Simplified data model, tenancy, and RLS.
4. [`backend-audit.md`](backend-audit.md) — Backend modules, routes, middleware, auth, RBAC, and security findings.
5. [`frontend-audit.md`](frontend-audit.md) — Frontend stack, routing, API layer, state, and UI findings.
6. [`infra-audit.md`](infra-audit.md) — Docker, Kubernetes, Terraform, CI/CD, and environment gaps.
7. [`ROUTES.md`](ROUTES.md) — Frontend and backend route inventory.
8. [`CURRENT_STATE.md`](CURRENT_STATE.md) — What is finished, partial, placeholder, broken, and missing.
9. [`TECHNICAL_DEBT.md`](TECHNICAL_DEBT.md) — Master issue register, categorized by severity.
10. [`NEXT_STEPS.md`](NEXT_STEPS.md) — Prioritized engineering roadmap with rationale.
11. [`ENGINEER_ONBOARDING.md`](ENGINEER_ONBOARDING.md) — How to build, debug, extend, and contribute.
12. [`migration-review.md`](migration-review.md) — Full migration catalog, schema, permissions, and RLS details.

## Document index

| Document | Purpose |
|----------|---------|
| `PROJECT_OVERVIEW.md` | Product vision, target users, MVP scope, roadmap, completion estimate |
| `ARCHITECTURE.md` | System architecture with diagrams, auth/authz flows, request lifecycle |
| `DATABASE_OVERVIEW.md` | Simplified database model, ERD, RLS, table ownership |
| `backend-audit.md` | Backend audit findings and module reference |
| `frontend-audit.md` | Frontend audit findings and UI reference |
| `infra-audit.md` | Infrastructure and DevOps audit |
| `ROUTES.md` | Complete route map (frontend and `/api/v1`) |
| `CURRENT_STATE.md` | Exact repository maturity and module status |
| `FEATURE_MATRIX.md` | Feature-by-feature completion matrix |
| `TECHNICAL_DEBT.md` | Consolidated issues with impact and recommended fixes |
| `NEXT_STEPS.md` | Prioritized roadmap |
| `ENGINEER_ONBOARDING.md` | New engineer guide |
| `migration-review.md` | Database migration deep-dive |
