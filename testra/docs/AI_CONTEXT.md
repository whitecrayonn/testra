# AI Context — How to Work Inside the Testra Repository

**Purpose:** Orient any automated agent (Claude, GPT, Gemini, Devin, Kimi, GLM, Codex, etc.) before it reads or modifies this repository.

**Owner:** Engineering Lead / Documentation Architect

**Scope:** This document is a thin wrapper around the canonical engineering handbook. It does not duplicate detailed architecture, coding standards, or product context.

**Related documents:**
- [`BIBLICAL_TESTRA.md`](BIBLICAL_TESTRA.md) — canonical engineering handbook, canonical sources map, do-not-break list.
- [`AI_MEMORY.md`](AI_MEMORY.md) — permanent architectural facts that must remain true.
- [`AI_RULES.md`](AI_RULES.md) — what to update when something changes.
- [`ONBOARDING.md`](engineering/ONBOARDING.md) — human/AI contributor workflow and DoD.
- [`ENGINEERING_STANDARDS.md`](engineering/ENGINEERING_STANDARDS.md) — coding standards.
- [`ROADMAP.md`](engineering/ROADMAP.md) — current implementation status and technical debt.
- [`docs/api/openapi/openapi.yaml`](api/openapi/openapi.yaml) — authoritative HTTP contract.
- [`docs/architecture/adrs/`](architecture/adrs/) — accepted architecture decisions.

**Source of truth:** `BIBLICAL_TESTRA.md` is the single source of truth for engineering rules and architecture; this document is an AI-specific entry point.

**Last updated:** July 2026

---

## Reading order for an AI agent

1. Read `BIBLICAL_TESTRA.md` → `Canonical Sources Map` and `AI Contributor Reference`.
2. Read `AI_MEMORY.md` for the permanent facts that shape every decision.
3. Read `AI_RULES.md` for the change-impact matrix.
4. Read the ADRs that govern the area you are touching (`ADR-001` through `ADR-012`).
5. Read `ROADMAP.md` to confirm the current implementation status and phase priorities.
6. Read the relevant OpenAPI operations and the module code under `apps/api/internal/<module>/`.
7. Make the smallest change that satisfies the request; update the canonical docs listed in `AI_RULES.md`.

---

## Architecture rules

- Testra is a **modular monolith** with Clean/Hexagonal Architecture in the Go backend (`apps/api/internal/<module>/`).
- One module owns one domain. Do not import another module's internal packages.
- Dependency rule points inward: handlers/repositories depend on domain and ports.
- Every protected request passes `Auth` → `TenantContext` → `RequirePermission` in that order.
- PostgreSQL Row-Level Security is mandatory for every tenant-scoped table in staging and production.
- Multi-tenancy is enforced by `organization_id` propagation and `app.tenant_id` on the database connection.
- The frontend (`apps/web/`) is a Next.js 15 App Router client; it never touches the database.
- The ML service (`apps/ml/`) is a separate FastAPI runtime; it never stores customer source code or API collections.

For the full architecture, see `BIBLICAL_TESTRA.md` §System Architecture, §Backend Clean Architecture, and `SYSTEM_FLOWS.md`.

---

## Coding rules

- Follow `ENGINEERING_STANDARDS.md` for Go, TypeScript, SQL, and infrastructure conventions.
- Never call the database from a handler or UI component; use services and repositories.
- Never bypass tenant isolation, RLS, RBAC, or audit logging.
- Never expose internal error details to clients; return the canonical `{ data, meta, error }` envelope.
- Never change a merged migration; create a new numbered `up`/`down` pair.
- Use existing patterns in the same module or similar modules before introducing new abstractions.
- Add or update tests for changed domain logic. Run `make test && make lint` before declaring work done.

---

## Documentation rules

- One topic = one owner. If detailed content already exists, link to it; do not copy it.
- Update `BIBLICAL_TESTRA.md` only when rules, modules, dependencies, or canonical sources change.
- Update `ROADMAP.md` when a phase starts or completes.
- Update `FEATURE_MATRIX.md` when a feature changes backend/frontend/OpenAPI/test/production-ready status.
- Update `DATABASE_GUIDE.md` and `BIBLICAL_TESTRA.md` data model section when schema changes.
- Update `ROUTES.md` and `api/openapi/openapi.yaml` when HTTP routes change.
- Create an ADR for any architectural decision that deviates from accepted ADRs.
- Do not create new markdown files unless the topic has no canonical owner.

---

## Update rules (summary)

For the full impact matrix, see [`AI_RULES.md`](AI_RULES.md).

| If you change ... | Then also update ... |
|---|---|
| API contract | `api/openapi/openapi.yaml`, `ROUTES.md`, `BIBLICAL_TESTRA.md` route table, `FEATURE_MATRIX.md` |
| Database schema | `apps/api/migrations/*.sql`, `DATABASE_GUIDE.md`, `BIBLICAL_TESTRA.md` data model |
| Module or dependency | `BIBLICAL_TESTRA.md` modules/dependency graph, `MODULE_DEPENDENCIES.md`, `ROADMAP.md` |
| Architecture decision | New ADR, `BIBLICAL_TESTRA.md` canonical sources, `AI_MEMORY.md` if fact changes |
| Feature implementation | `FEATURE_MATRIX.md`, `ROADMAP.md`, `PROJECT_OVERVIEW.md`, `BIBLICAL_TESTRA.md` status notes |
| Deployment/infra | `DEPLOYMENT_GUIDE.md`, ADR if architecture changes, `BIBLICAL_TESTRA.md` deployment section |
| Security control | `SECURITY_CHECKLIST.md`, ADR if decision changes, `BIBLICAL_TESTRA.md` security section |

---

## Verification workflow

Before claiming a task is done:

1. Run `make test` and `make lint` (or the equivalent Go/TypeScript commands).
2. Run `python testra/scripts/doc_audit_check.py` and confirm no new broken links or code refs in active docs.
3. Confirm Mermaid syntax in edited markdown is valid (visual check or CI render test).
4. Confirm the change matches the OpenAPI spec if it touches HTTP behavior.
5. Confirm the change does not violate `AI_MEMORY.md` or `BIBLICAL_TESTRA.md` `Do Not Break List`.
6. Update the canonical documents listed in `AI_RULES.md` for your change type.
7. Write or update tests for new domain logic.

---

## Forbidden actions

- Do not modify production source code, business logic, database schema, APIs, or tests unless explicitly asked.
- Do not rewrite `BIBLICAL_TESTRA.md` from scratch; improve it with focused additions or corrections.
- Do not change accepted ADRs; create a new ADR and escalate for human review.
- Do not duplicate content between active documents; update the canonical owner and link to it.
- Do not invent architecture, libraries, or modules that are not already aligned with `BIBLICAL_TESTRA.md`, the ADRs, and existing code patterns.
- Do not bypass tenant isolation, RLS, RBAC, or audit logging.
- Do not hardcode tenant IDs, secrets, environment names, or paths.
- Do not modify merged migrations; always create a new migration.
- Do not skip tests or delete/weaken existing tests.
- Do not expose secrets, source code, or raw API collection payloads in logs or responses.

---

## Canonical ownership

Every concern has exactly one canonical document. When in doubt, consult `BIBLICAL_TESTRA.md` §Canonical Sources and Document Health and §Canonical Sources Map.

Quick map:

| Question | Canonical owner |
|---|---|
| What is Testra and where is it now? | `testra-master-context.md` → `PROJECT_OVERVIEW.md` |
| What is the architecture? | `BIBLICAL_TESTRA.md` → `SYSTEM_FLOWS.md`, `MODULE_DEPENDENCIES.md`, ADRs |
| What are the coding rules? | `ENGINEERING_STANDARDS.md` |
| How do I set up and contribute? | `ONBOARDING.md` |
| What is the API contract? | `api/openapi/openapi.yaml` |
| What are the API conventions? | `API_DESIGN_GUIDELINES.md` |
| What is the database schema? | `DATABASE_GUIDE.md` + `apps/api/migrations/` |
| What are the system flows? | `SYSTEM_FLOWS.md` |
| What is implemented? | `FEATURE_MATRIX.md` |
| What is the roadmap? | `ROADMAP.md` |
| How do I deploy? | `DEPLOYMENT_GUIDE.md` |
| What are the security controls? | `SECURITY_CHECKLIST.md` + `ADR-007-security-standards.md` |
| What must never change? | `AI_MEMORY.md` |
| What must I update when X changes? | `AI_RULES.md` |

---

## Definition of Done

A task is done when:

- Code implements the approved requirement using existing patterns.
- OpenAPI, migrations, and canonical docs are updated if the change touches them.
- Unit tests cover new domain logic; all existing tests pass.
- `make test`, `make lint`, and `python testra/scripts/doc_audit_check.py` pass.
- No secrets or hardcoded assumptions are introduced.
- `BIBLICAL_TESTRA.md` and the relevant canonical doc(s) reflect the change.

For the full DoD, see [`ONBOARDING.md`](engineering/ONBOARDING.md) §5.

## See Also

- [`BIBLICAL_TESTRA.md`](BIBLICAL_TESTRA.md) — canonical engineering handbook
- [`AI_MEMORY.md`](AI_MEMORY.md) — permanent architectural facts
- [`AI_RULES.md`](AI_RULES.md) — change-impact matrix
