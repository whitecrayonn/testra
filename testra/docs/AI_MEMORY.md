# AI Memory — Permanent Architectural Facts for Testra

**Purpose:** Record immutable architectural truths that every future AI must treat as constraints, not suggestions.

**Owner:** CTO / Engineering Lead

**Scope:** This file lists permanent or long-lived facts about how Testra is built. It does not repeat dynamic status; for that, see `ROADMAP.md` and `FEATURE_MATRIX.md`.

**Related documents:**
- [`BIBLICAL_TESTRA.md`](BIBLICAL_TESTRA.md) — engineering handbook and do-not-break list.
- [`AI_CONTEXT.md`](AI_CONTEXT.md) — how an AI should start and verify work.
- [`AI_RULES.md`](AI_RULES.md) — what to update when something changes.
- [`docs/architecture/adrs/`](architecture/adrs/) — accepted architecture decisions.

**Source of truth:** `BIBLICAL_TESTRA.md` §Do Not Break List, §Engineering Rules, and accepted ADRs.

**Last updated:** July 2026

---

## Identity & authentication

- The canonical authentication mechanism is **self-hosted JWT access tokens + rotating opaque refresh tokens** (ADR-001).
- Access JWTs expire after **15 minutes** (ADR-007, `BIBLICAL_TESTRA.md`).
- Refresh tokens have a **30-day inactivity expiry** and **90-day absolute expiry**; reuse detection revokes the entire family (ADR-007).
- TOTP MFA is implemented and required for organization administrators and enterprise users (ADR-001, ADR-007).
- API keys are scoped, hashed (SHA-256), displayed once at creation, and expire after **90 days by default** with a **365-day maximum** (ADR-007, `BIBLICAL_TESTRA.md`).
- Passwords require **at least 12 characters**; reset tokens are single-use and expire after **30 minutes** (ADR-007).
- Secrets (JWT signing keys, database credentials, Redis, SMTP, S3, integration tokens) are never committed and must be managed outside source control (SECURITY_CHECKLIST, DEPLOYMENT_GUIDE).

## Tenancy & isolation

- The tenant boundary is the **organization**. A user can belong to many organizations (BIBLICAL, DATABASE_GUIDE).
- Hierarchy: **Organization → Workspace → Project** (BIBLICAL, ADR-004).
- **PostgreSQL Row-Level Security (RLS)** is mandatory for tenant-scoped tables in staging and production (ADR-004).
- The application sets `app.tenant_id` per request/database connection; RLS policies read this setting (BIBLICAL, SYSTEM_FLOWS, ADR-004).
- Cross-tenant access is prohibited without an explicit, audited system role and service port (ADR-004).
- Queue jobs, cache keys, exports, object paths, ClickHouse rows, and ML requests must carry tenant scope and re-check authorization at job boundaries (ADR-004).

## Authorization

- RBAC uses permission strings assigned to roles and roles assigned to users at organization scope (BIBLICAL, DATABASE_GUIDE).
- The route tree checks specific permissions (e.g., `orgs:read`, `workspaces:create`, `tests:read`) (BIBLICAL, ROUTES.md, `apps/api/internal/shared/server/server.go`).
- Permission-name drift exists between older seed migrations and current route checks; new work should use the names the middleware actually checks (DATABASE_GUIDE §5.5).
- The protected middleware chain is `Auth` → `TenantContext` → `RequirePermission` (BIBLICAL, ONBOARDING.md).

## Data & schema

- The authoritative database schema is `apps/api/migrations/*.sql` (currently `000001` through `000028`) (BIBLICAL, DATABASE_GUIDE).
- Merged migrations are immutable. Never edit them; create a new `up`/`down` pair (BIBLICAL, ENGINEERING_STANDARDS, ONBOARDING).
- Every migration must have both `up` and `down` files (BIBLICAL, ENGINEERING_STANDARDS).
- UUID primary keys are used for all entities (BIBLICAL, DATABASE_GUIDE).
- `audit_events` is append-only and tenant-scoped; mutating actions must produce immutable audit records (BIBLICAL, ADR-004).
- `idempotency_records` are scoped to `(workspace_id, operation, key)` and expire after 24 hours by default (ADR-006, ADR-012).
- Phase 3 test runs and results live in PostgreSQL, not ClickHouse; ClickHouse is deferred to a future phase (ADR-010).
- Customer source code, test scripts, and API collection payloads are never stored (BIBLICAL Do Not Break List, ADR-011, SECURITY_CHECKLIST).

## API & contracts

- The API contract is defined in `docs/api/openapi/openapi.yaml` (OpenAPI 3.1) and must be updated before implementation (BIBLICAL, API_DESIGN_GUIDELINES, ADR-006).
- Base path is `/api/v1` (BIBLICAL, ROUTES.md, OpenAPI).
- All API responses use the canonical envelope `{ data, meta, error }` (BIBLICAL, API_DESIGN_GUIDELINES, ADR-006).
- List endpoints use **cursor pagination**, not offset pagination (BIBLICAL, API_DESIGN_GUIDELINES, ADR-006).
- Side-effecting endpoints must respect `Idempotency-Key` (BIBLICAL, API_DESIGN_GUIDELINES, ADR-006).
- Ingestion uses synchronous processing for MVP; async queue processing is deferred (ADR-011).

## Architecture & runtime

- The backend is a **modular monolith** in Go with Clean/Hexagonal Architecture (BIBLICAL, MODULE_DEPENDENCIES, ENGINEERING_STANDARDS).
- Modules must not import another module's internal packages; dependencies go through ports (BIBLICAL, MODULE_DEPENDENCIES).
- The Go router is **chi** with global middleware: Logger, Recoverer, RequestID, CORS, MaxBodySize (BIBLICAL, `server.go`).
- The frontend is a Next.js 15 App Router + TypeScript application using TailwindCSS and shadcn/ui (BIBLICAL, ENGINEERING_STANDARDS).
- The ML service is a Python FastAPI skeleton (`apps/ml/api/main.py`) planned for Phase 6 (BIBLICAL, ROADMAP).
- The worker (`apps/worker/`) is currently a stub; background processing is planned for Phase 5+ (BIBLICAL, ROADMAP).

## Deployment & infrastructure

- Local development uses native services; no Docker is used (ADR-009).
- MVP production is an Ubuntu VM with systemd and Nginx reverse proxy (ADR-003, DEPLOYMENT_GUIDE).
- TLS is terminated at Nginx with Let's Encrypt on the single Ubuntu VPS (MVP).
- Migrations are applied through CI/CD via `apps/api/cmd/migrator`; never manually in production (BIBLICAL, DEPLOYMENT_GUIDE).
- Cloud-managed services and container orchestration may be considered for future scale, but they are not planned for MVP (ADR-003, DEPLOYMENT_GUIDE).

## Observability & operations

- Required observability stack: OpenTelemetry, Prometheus, Grafana, Loki, structured Go logs (MONITORING_LOGGING_GUIDE, ROADMAP H10).
- Retention is fixed by ADR-005: audit 7 years enterprise / 2 years MVP/Beta, logs 30 days hot / 90 days archived, metrics 15 months, traces 14 days (BIBLICAL, ADR-005, DISASTER_RECOVERY_GUIDE).
- PostgreSQL target: RPO ≤ 5 minutes MVP, RTO ≤ 4 hours MVP (ADR-005, DISASTER_RECOVERY_GUIDE).

## Performance

- API read target: p95 ≤ 300 ms, p99 ≤ 750 ms (ADR-008).
- API write target: p95 ≤ 500 ms, p99 ≤ 1000 ms (ADR-008).
- PostgreSQL query target: p95 ≤ 50 ms (ADR-008).
- Request timeout: 30 seconds; background job timeout: 5 minutes (ADR-008).
- Capacity target: 500 concurrent users, 50 req/s sustained, 2x burst for 5 minutes (ADR-008).

## Security & privacy (do not break)

- No external LLM processing; ML is transparent, explainable, and tenant-isolated (BIBLICAL, ADR-001).
- Secrets are hashed at rest (passwords, tokens, API keys) (BIBLICAL, ADR-007).
- Customer data is customer-owned; zero source code or API collection retention (BIBLICAL Do Not Break List, ADR-011).
- Rate limiting and abuse controls are required on auth endpoints (BIBLICAL, ADR-007, ROADMAP C1).

## Documentation ownership

- `BIBLICAL_TESTRA.md` is the single entry point for engineering knowledge.
- `AI_CONTEXT.md`, `AI_MEMORY.md`, and `AI_RULES.md` are AI-specific overlays; they do not replace `BIBLICAL_TESTRA.md`.
- `ROADMAP.md` is authoritative for implementation status and debt.
- `FEATURE_MATRIX.md` is authoritative for feature completion.
- `docs/api/openapi/openapi.yaml` is authoritative for HTTP contracts.
- `apps/api/migrations/*.sql` is authoritative for schema.
- ADRs are authoritative for accepted architecture decisions.

---

## How this file is maintained

- Add a fact when a new accepted ADR or engineering rule establishes a long-lived constraint.
- Do not record speculative or planned work here; use `ROADMAP.md` or `FEATURE_MATRIX.md` instead.
- If a fact becomes false, update this file and add a note with the new ADR or decision that superseded it.

## See Also

- [`BIBLICAL_TESTRA.md`](BIBLICAL_TESTRA.md) — canonical engineering handbook
- [`AI_CONTEXT.md`](AI_CONTEXT.md) — AI entry point and verification workflow
- [`AI_RULES.md`](AI_RULES.md) — change-impact matrix
