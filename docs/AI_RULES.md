# AI Rules — What to Update When Something Changes

**Purpose:** Provide a precise change-impact matrix so any AI (or human) can update the right canonical document for every engineering change without guessing or duplicating content.

**Owner:** CTO / Engineering Lead

**Scope:** This document covers documentation, architecture, API, database, frontend, operations, and security updates. It does not cover runtime business logic.

**Related documents:**
- [`BIBLICAL_TESTRA.md`](BIBLICAL_TESTRA.md) — engineering handbook and canonical sources.
- [`AI_CONTEXT.md`](AI_CONTEXT.md) — how an AI starts and verifies work.
- [`AI_MEMORY.md`](AI_MEMORY.md) — permanent architectural facts.
- [`ONBOARDING.md`](engineering/ONBOARDING.md) — contributor workflow and DoD.

**Source of truth:** `BIBLICAL_TESTRA.md` §Documentation Maintenance Guide and §Canonical Sources and Document Health.

**Last updated:** July 2026

---

## 1. Universal AI rules

These apply to every change, regardless of type:

1. Read `AI_CONTEXT.md`, `AI_MEMORY.md`, and this file before starting.
2. Read `BIBLICAL_TESTRA.md` §Canonical Sources Map to find the canonical owner for every topic.
3. Read the ADRs that apply to the area you are changing.
4. Update the canonical owner for the topic you are changing; do not create a parallel copy.
5. Update `BIBLICAL_TESTRA.md` only when rules, modules, dependencies, or canonical sources change.
6. Do not modify merged migrations; create a new migration.
7. Do not change accepted ADRs; create a new ADR and escalate for human review.
8. Run `python testra/scripts/doc_audit_check.py` and verify no broken links or code refs in active docs.
9. Run `make test` and `make lint` (or the language equivalents) before finishing.
10. Verify the change does not violate `BIBLICAL_TESTRA.md` §Do Not Break List or `AI_MEMORY.md`.

---

## 2. Change-impact matrix

Find the row that matches your change and update every document in the **Update list** column.

### 2.1 New or changed API endpoint / HTTP contract

| Update | Why |
|---|---|
| `docs/api/openapi/openapi.yaml` | The contract is the source of truth for HTTP behavior. |
| `docs/ROUTES.md` | Public-facing backend route inventory must match the router. |
| `docs/BIBLICAL_TESTRA.md` route table and API-design sections | Keep the handbook summary current. |
| `docs/FEATURE_MATRIX.md` | Mark affected feature(s) as implemented or updated. |
| `docs/PROJECT_OVERVIEW.md` | Update current-state notes if the change affects MVP status. |
| `docs/engineering/ROADMAP.md` | Close or advance related items in the technical debt register. |

### 2.2 New or changed database schema

| Update | Why |
|---|---|
| `apps/api/migrations/000XXX_*.sql` | Schema is the authoritative source of truth. Add `up` and `down`. |
| `docs/architecture/DATABASE_GUIDE.md` | Migration catalog, ERD, column summary, and RLS policy matrix must reflect the schema. |
| `docs/BIBLICAL_TESTRA.md` data model section | Keep the handbook summary current. |
| `docs/api/openapi/openapi.yaml` | Update schemas if API representations changed. |
| `docs/FEATURE_MATRIX.md` | Update storage/test/roadmap rows as needed. |

### 2.3 New backend module

| Update | Why |
|---|---|
| `apps/api/internal/<module>/` | Follow the standard file layout (`domain.go`, `ports.go`, `repository.go`, `service.go`, `handler.go`, `module.go`). |
| `apps/api/internal/shared/server/server.go` | Register routes and middleware. |
| `docs/architecture/MODULE_DEPENDENCIES.md` | Add the module to the dependency graph and cycle check. |
| `docs/BIBLICAL_TESTRA.md` module list, dependency graph, route table | Keep the engineering handbook current. |
| `docs/engineering/ROADMAP.md` | Update phase/feature status. |
| `docs/FEATURE_MATRIX.md` | Add/update implementation status. |
| `docs/api/openapi/openapi.yaml` | Add new endpoints and schemas. |
| `docs/ROUTES.md` | Add new routes. |
| `docs/architecture/adrs/ADR-XXX-*.md` | Create an ADR if the module introduces a new cross-cutting dependency or architecture change. |

### 2.4 New frontend page or route

| Update | Why |
|---|---|
| `apps/web/app/(dashboard)/[workspace]/<route>/` | Follow Next.js 15 App Router conventions. |
| `apps/web/features/<module>/api.ts` | Add or update API wrapper. |
| `apps/web/components/dashboard/sidebar.tsx` | Update navigation if top-level. |
| `docs/ROUTES.md` | Add new frontend/backend route. |
| `docs/FEATURE_MATRIX.md` | Mark frontend feature status. |
| `docs/BIBLICAL_TESTRA.md` frontend section if architecture changes | Only if workspace/routing rules change. |

### 2.5 New or changed architecture decision

| Update | Why |
|---|---|
| `docs/architecture/adrs/ADR-XXX-*.md` | Create a new ADR; never edit an accepted one. |
| `docs/BIBLICAL_TESTRA.md` future architecture / do-not-break list | Update summary references and rules if they change. |
| `docs/AI_MEMORY.md` | Add/update permanent architectural facts if the decision establishes a long-lived constraint. |
| `docs/PROJECT_OVERVIEW.md` | Update architecture summary and current state. |
| `docs/engineering/ROADMAP.md` | Update technical debt or documentation roadmap if needed. |

### 2.6 New or changed deployment / infrastructure

| Update | Why |
|---|---|
| `docs/deployment/DEPLOYMENT_GUIDE.md` | Update deployment model, service architecture, gates, and findings. |
| `docs/operations/DISASTER_RECOVERY_GUIDE.md` | Update RPO/RTO, backup, and recovery implications. |
| `docs/operations/MONITORING_LOGGING_GUIDE.md` | Update observability requirements if new signals are needed. |
| `docs/BIBLICAL_TESTRA.md` deployment section | Keep the handbook summary current. |
| `docs/architecture/adrs/ADR-XXX-*.md` | Create an ADR if the change affects accepted deployment architecture. |

### 2.7 New or changed security / privacy control

| Update | Why |
|---|---|
| `docs/security/SECURITY_CHECKLIST.md` | Update identity, data, application, and operations checklists. |
| `docs/operations/PRODUCTION_READINESS_CHECKLIST.md` | Update go-live requirements if affected. |
| `docs/BIBLICAL_TESTRA.md` security and do-not-break list | Update rules and guarantees. |
| `docs/architecture/adrs/ADR-XXX-*.md` | Create an ADR if the change is a new security architecture decision. |
| `docs/AI_MEMORY.md` | Add a permanent fact if the rule is long-lived. |

### 2.8 New or changed product / MVP scope

| Update | Why |
|---|---|
| `PROJECT_OVERVIEW.md` | Update vision, goals, MVP scope, and current state. |
| `FEATURE_MATRIX.md` | Add/update feature implementation status. |
| `ROADMAP.md` | Move phases/items to completed or adjust priorities. |
| `BIBLICAL_TESTRA.md` | Update repository evolution timeline and feature dependency matrix. |
| Root product documents (`testra-master-context.md`, `testra-product-strategy.md`, `testra-brd.md`) | Update business/product strategy if required. |

### 2.9 New or changed test strategy / engineering process

| Update | Why |
|---|---|
| `docs/engineering/ENGINEERING_STANDARDS.md` | Update testing, linting, CI/CD, and workflow rules. |
| `docs/engineering/ONBOARDING.md` | Update onboarding and contributor rules. |
| `docs/BIBLICAL_TESTRA.md` engineering rules and AI rules | Update if process rules change. |
| `docs/reports/` | Record meaningful work in a dated report under `docs/reports/`. |

### 2.10 Documentation / repository hygiene

| Update | Why |
|---|---|
| `docs/README.md` | Update the documentation index when files are added, archived, or renamed. |
| `docs/BIBLICAL_TESTRA.md` canonical sources map | Update the map when canonical ownership changes. |
| `docs/reports/AI_DOCUMENTATION_CERTIFICATION.md` | Re-certify after major documentation changes. |
| `testra/scripts/doc_audit_check.py` | Update only if the audit rules themselves change. |

---

## 3. Forbidden shortcuts

- Do not add a new doc for a topic that already has a canonical owner.
- Do not copy a table or diagram from another doc; link to the canonical owner and, if necessary, summarize in one sentence.
- Do not update only the most visible doc (e.g., `README.md`) and skip the canonical source.
- Do not mark `FEATURE_MATRIX.md` or `ROADMAP.md` as complete before the corresponding OpenAPI/migration/tests are actually done.
- Do not update `BIBLICAL_TESTRA.md` with implementation detail that belongs in `DATABASE_GUIDE.md`, `SYSTEM_FLOWS.md`, `API_DESIGN_GUIDELINES.md`, or `ROUTES.md`.

---

## 4. Verification checklist per change type

After editing, confirm:

- [ ] The canonical owner contains the detail; other docs link to it.
- [ ] `docs/README.md` index is still accurate.
- [ ] `BIBLICAL_TESTRA.md` canonical sources map still points to the right files.
- [ ] No broken markdown links or missing code references in active docs (`python testra/scripts/doc_audit_check.py`).
- [ ] Mermaid diagrams render correctly (no syntax errors).
- [ ] New doc names follow the `UPPER_SNAKE_CASE.md` convention used in `testra/docs/`.
- [ ] No new top-level docs are created without adding them to `docs/README.md` and `BIBLICAL_TESTRA.md`.

## See Also

- [`BIBLICAL_TESTRA.md`](BIBLICAL_TESTRA.md) — canonical engineering handbook
- [`AI_CONTEXT.md`](AI_CONTEXT.md) — AI entry point and verification workflow
- [`AI_MEMORY.md`](AI_MEMORY.md) — permanent architectural facts
