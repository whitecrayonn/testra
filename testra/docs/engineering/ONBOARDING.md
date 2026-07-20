# Testra Onboarding and Development Guide

**Purpose:** Orient new engineers and contributors to the repository, workflows, governance, and definition of done.
**Owner:** CTO / Engineering Lead
**Scope:** Onboarding, engineering principles, development workflow, architecture principles, DoD, and common pitfalls.
**Status:** Active
**Last Updated:** July 2026
**Classification:** Internal — Engineering
**Source of Truth:** ONBOARDING.md for contributor workflow and governance.
**Related documents:**
- [`BIBLICAL_TESTRA.md`](../BIBLICAL_TESTRA.md)
- [`ENGINEERING_STANDARDS.md`](ENGINEERING_STANDARDS.md)
- [`ROADMAP.md`](ROADMAP.md)

---

## 1. Purpose

This document is the **engineering source of truth** for the Testra platform. It governs how code is written, reviewed, tested, and shipped. All human and AI contributors must follow the governance, principles, and workflows defined here.

Companion documents:
- `docs/engineering/ROADMAP.md` — implementation roadmap and phase tracking
- `docs/engineering/ENGINEERING_STANDARDS.md` — detailed coding and process standards
- `docs/engineering/progress/` — chronological Engineering Progress Reports

---

## 2. Engineering Principles

1. **One module, one PRD, one owner.** Every feature belongs to exactly one domain module. Never duplicate ownership.
2. **Clean Architecture boundaries.** Domain → Application → Ports → Adapters. Business rules never depend on frameworks or infrastructure.
3. **API-first.** The OpenAPI spec is the contract. Endpoints are documented before implementation. The web UI is just another client.
4. **Privacy-first.** Zero customer source code retention. Zero API collection retention. No external LLM dependency. Tenant isolation enforced at every layer.
5. **Simple first.** Favor simplicity over feature count. Build reusable capabilities before specialized features.
6. **Test what matters.** Domain logic and use cases must have tests. Handlers and repositories are tested via integration tests. No test is deleted or weakened without explicit direction.
7. **Minimal changes.** Prefer focused, surgical edits. Avoid over-engineering. Use single-line changes when sufficient.
8. **No ungrounded assertions.** Every code change must be immediately runnable. Imports at the top. Dependencies declared. No placeholder code.

---

## 3. Development Workflow

### 3.1 Trunk-Based Development

- Short-lived feature branches (`feat/<scope>`, `fix/<scope>`) → PR → squash-merge to `main`.
- `main` is always deployable.
- Releases tagged `vX.Y.Z`.
- Feature flags for incomplete work (unleash or in-app config).

### 3.2 Local Development

Testra uses a **Native Development Environment** (ADR-009). Docker is not required.

1. Install local services: PostgreSQL 16+, Redis 7+, Mailpit, MinIO (see README.md for platform-specific instructions)
2. `pnpm install` — installs JS deps + auto-creates Python venv for ML
3. `pnpm dev` — checks local services, runs migrations, and starts API, web, worker, and ML simultaneously via Turborepo
4. `make test`, `make lint` before push

Docker and Docker Compose are not used. All services run natively on the local machine.

### 3.3 CI/CD

- **PR pipeline:** lint → unit tests → build → OpenAPI contract validation → security scan
- **Main pipeline:** all above + integration tests → build images (tagged by SHA) → deploy staging → manual promote to production
- Migrations applied by `cmd/migrator` in CI, never manually

### 3.4 Environments

| Environment | Purpose | Data |
|---|---|---|
| `local` | Native development (local PostgreSQL, Redis, Mailpit, MinIO) | Seeded fixtures |
| `staging` | Auto-deploy from `main` | Sanitized/synthetic |
| `production` | Manual promotion | Real, backed up, monitored |

Config via env vars (12-factor). `.env.example` is committed; local values use ignored files; MVP production values use environment files or a local secrets store on the Ubuntu VPS. Cloud-managed secrets managers may be adopted later if scale justifies it.

---

## 4. Architecture Principles

### 4.1 Modular Monolith

- Single deployable Go binary with internal domain modules.
- Each module: `domain.go`, `ports.go`, `repository.go`, `service.go`, `handler.go`, `module.go`.
- Modules communicate via interfaces (ports), never direct imports of another module's internals.
- Extraction to microservices is a future option — boundaries are already defined.

### 4.2 Multi-Tenancy

- `tenant_id` (organization_id) on every tenant-scoped table.
- PostgreSQL RLS is mandatory in staging and production; API roles do not bypass RLS.
- HTTP middleware authenticates and resolves candidate tenant scope; request context propagates it; service layers authorize resource relationships and permissions; repositories set transaction-local `app.tenant_id`.
- Tenant scope must propagate through queues, cache keys, exports, ClickHouse, and ML calls.

### 4.3 Data Layer

- **PostgreSQL** — OLTP, transactional data, audit trails.
- **ClickHouse** — OLAP, test results, time-series events.
- **Redis** — sessions, rate limiting, job queues (Asynq).
- **S3-compatible** — attachments, exports, model artifacts.

### 4.4 ML Boundary

- `apps/ml` (Python/FastAPI) is inference + training only.
- Called by Go API (sync) or worker (async) over internal network.
- Per-tenant models. Never receives source code or API payloads.

---

## 5. Definition of Done (DoD)

A task is **Done** when all of the following are true:

- [ ] Code implements the approved spec/PRD requirements
- [ ] OpenAPI spec updated if endpoints changed
- [ ] Unit tests written for domain/application logic
- [ ] All existing tests pass (`go test`, `pnpm test`)
- [ ] Linting passes (`go vet`, `eslint`, `ruff`)
- [ ] No secrets or sensitive data in code or commits
- [ ] Migration files created if schema changed (up + down)
- [ ] `.env.example` updated if new env vars introduced
- [ ] `ROADMAP.md` updated if phase status changed
- [ ] Engineering Progress Report saved if session produced meaningful work

---

## 6. Self-Review Process

Before requesting review or merging a PR:

1. **Read the diff.** Every line. Ask: "Would I approve this if someone else wrote it?"
2. **Check imports.** Are they at the top? Are they used? Are they the right packages?
3. **Check error handling.** Are errors wrapped with context? Are sentinel errors used correctly?
4. **Check tests.** Do they test behavior, not implementation? Do they cover edge cases?
5. **Check naming.** Are names clear and consistent with the codebase?
6. **Check boundaries.** Does the change respect module boundaries? No cross-module internal imports?
7. **Check security.** No hardcoded secrets. No SQL injection. Input validation on every endpoint.
8. **Check the OpenAPI spec.** If endpoints changed, is the spec updated?
9. **Run locally.** `make test && make lint` must pass.
10. **Update docs.** If governance or standards changed, update the relevant doc.

---

## 7. Engineering Progress Report Template

Every implementation session that produces meaningful work must save a progress report to `docs/engineering/progress/` using the filename format `YYYY-MM-DD-HHMM-session.md`.

```markdown
# Engineering Progress Report — YYYY-MM-DD HH:MM

## Session Summary
<one-sentence summary>

## Completed
- <item>

## In Progress
- <item>

## Blocked
- <item or "None">

## Next
- <item>

## Files Changed
- <path> — <description>

## Verification
- <command> — <result>
```

---

## 8. Rules for Future AI and Human Contributors

### 8.1 Source of Truth

- `ONBOARDING.md` (this file) — engineering governance. Update only when governance changes.
- `ROADMAP.md` — implementation roadmap. Update when phases start or complete.
- `ENGINEERING_STANDARDS.md` — coding standards. Update only when standards change.
- `docs/engineering/progress/` — progress reports. Append-only, never edit past reports.
- Approved product/architecture documents — **read-only**. Changes require a new ADR.

### 8.2 AI Contributor Rules

- Always read `ROADMAP.md` before starting work to understand current state.
- Always read `ENGINEERING_STANDARDS.md` before writing code.
- Never modify approved product or architecture documents without creating an ADR.
- Never skip tests. Never delete or weaken tests.
- Never commit secrets. Never hardcode API keys.
- Always update `ROADMAP.md` when a phase starts or completes.
- Always save a progress report at the end of a session.
- Prefer minimal, focused edits. Follow existing code patterns.
- Imports at the top of the file. Always.

### 8.3 Human Contributor Rules

- Review every AI-generated PR with the self-review process (§6).
- Approve PRs only when DoD is met (§5).
- Keep `ROADMAP.md` honest — if work is incomplete, mark it as such.
- Create an ADR for any architectural decision that deviates from approved documents.

---

## 9. Document Maintenance

| Document | When to Update | Who |
|---|---|---|
| `ONBOARDING.md` | Engineering governance changes | CTO / Lead |
| `ROADMAP.md` | Phase starts, completes, or scope changes | Any contributor |
| `ENGINEERING_STANDARDS.md` | Coding or process standards change | CTO / Lead |
| `docs/engineering/progress/*` | End of each implementation session | Session owner |
| ADRs (`docs/architecture/adrs/`) | Architectural deviation from approved docs | CTO / Lead |


## Engineer Onboarding

> Welcome. This guide assumes you are a senior engineer joining Testra with no prior conversation. Read it alongside [`README.md`](../README.md), [`PROJECT_OVERVIEW.md`](../PROJECT_OVERVIEW.md), [`BIBLICAL_TESTRA.md`](../BIBLICAL_TESTRA.md), and [`FEATURE_MATRIX.md`](../FEATURE_MATRIX.md).

### 1. Understand the repository

#### Top-level layout

```
testra/
├── apps/
│   ├── api/          # Go modular monolith backend
│   ├── web/          # Next.js 15 web application
│   ├── worker/       # Optional Go worker (currently stub)
│   └── ml/           # Python FastAPI ML service (placeholder)
├── packages/
│   ├── shared/       # Shared TypeScript types
│   ├── ui/           # Shared React components (shadcn/ui primitives)
│   ├── config/       # Shared tooling configs
│   └── sdk/           # Official SDK (not built yet)
├── docs/             # OpenAPI specs, ADRs, runbooks, deployment guides
│   ├── BIBLICAL_TESTRA.md     # Engineering handbook
│   ├── engineering/           # Roadmap, onboarding, standards
│   ├── architecture/          # ADRs and system design
│   ├── api/                   # OpenAPI spec and API guidelines
│   ├── operations/            # Runbooks and checklists
│   └── reports/               # Generated reports and reviews
├── scripts/          # Development and automation scripts
├── .github/          # CI/CD workflows
├── Makefile          # Common dev tasks
├── go.work           # Go workspace
├── pnpm-workspace.yaml
├── turbo.json
├── package.json
├── .env.example
└── README.md
```

#### How to read code

1. Start with `apps/api/internal/shared/server/server.go` to see how routes, middleware, and modules are wired.
2. Pick one complete module (e.g., `apps/api/internal/testmanagement/`) and read `domain.go`, `ports.go`, `repository.go`, `service.go`, `handler.go`, `module.go`.
3. Read the OpenAPI spec at `docs/api/openapi/openapi.yaml` for the public contract.
4. Explore the frontend from `apps/web/app/(dashboard)/layout.tsx` down through `[workspace]/test-cases/page.tsx`.

---

### 2. How modules interact

Testra is a **modular monolith**: all modules live in one Go binary but have strict internal boundaries.

- **No module calls another module's service directly.** If a module needs data from another scope, it looks it up through the database or a future port/interface.
- **Integration point is PostgreSQL.** Tenant context and shared entities (workspace, project) are resolved in middleware and passed on the request context.
- **Shared package is for cross-cutting concerns only:** config, errors, HTTP response helpers, pagination, JWT, password hashing, tenant resolution, middleware.
- **Frontend features mirror backend modules:** `apps/web/features/testmanagement/`, `features/results/`, `features/platform/`, etc.

---

### 3. How requests flow

#### Backend request

```
Client (Browser / CI runner)
  → Ingress / API Gateway
    → Go API router (chi)
      → Global middleware: Logger, Recoverer, RequestID, CORS, MaxBodySize
        → Auth middleware (parse JWT, set user_id in context)
          → TenantContext middleware (acquire *sql.Conn, SET app.tenant_id)
            → RequirePermission middleware (load permissions, check scope)
              → Handler (decode JSON, call service)
                → Service (business logic, validation)
                  → Repository (SQL queries on context-aware connection)
                    → PostgreSQL (RLS filters rows by app.tenant_id)
              → Handler (map to JSON envelope)
            → AuditLog / Idempotency post-processing
          → TenantContext reset/release connection
        → Response
```

#### Frontend request

```
Browser
  → Next.js App Router
    → Server Component (no client JS needed) or Client Component
      → `apiFetch` in `apps/web/lib/api.ts`
        → Reads `testra_token` from localStorage
        → Adds `Authorization: Bearer <token>` and `Content-Type: application/json`
        → Parses backend envelope `{ data, meta, error }`
        → Throws `ApiError` on error
```

> **Note:** The frontend currently does not refresh tokens automatically. Access tokens expire after 15 minutes. See `ROADMAP.md` §Technical Debt Register (H5).

---

### 4. Where to start debugging

| Symptom | Where to look |
|---------|---------------|
| 401/403 on API | `shared/middleware/auth.go`, `shared/middleware/tenant.go`, `shared/middleware/rbac.go` |
| Wrong tenant data | `shared/tenant/resolver.go`, `shared/db/db.go` connection context, RLS policies in the migration files |
| SQL error | `internal/<module>/repository.go`, `shared/db/db.go` |
| Handler panic | `shared/middleware/audit.go` (logging), chi `middleware.Recoverer` |
| Frontend 401 after 15 min | `apps/web/lib/api.ts`, no token refresh logic; also check `localStorage` keys |
| SSE not working | `apps/api/internal/results/handler.go` stream endpoint and `apps/web/app/(dashboard)/[workspace]/test-runs/[id]/page.tsx` `EventSource` usage |
| UI mismatch with API | `apps/web/features/<module>/api.ts` payload/response vs OpenAPI schema vs Go DTO |
| Build/lint failure | `pnpm typecheck` output, `go vet ./...`, GitHub Actions logs |

#### Useful commands

```bash
## Backend
go test ./...
go vet ./...
go run ./cmd/api

## Frontend
pnpm install
pnpm dev
pnpm typecheck
pnpm lint

## Database
cd apps/api && make migrate
## or
go run ./cmd/migrator

## Full stack
pnpm dev
```

---

### 5. How to add a new backend module

1. **Create the directory** under `apps/api/internal/<module>/`.
2. **Add standard files:**
   - `domain.go` — structs, value objects, validation helpers.
   - `ports.go` — repository interfaces.
   - `repository.go` — SQL implementation using `*shared.DB`.
   - `service.go` — business logic.
   - `handler.go` — HTTP handlers, DTOs, route helpers.
   - `module.go` — dependency wiring (`NewModule(db, config)`).
3. **Register routes** in `apps/api/internal/shared/server/server.go`.
4. **Add migrations** in `apps/api/migrations/` if new tables are needed.
5. **Add permissions** to the RBAC seed (migration `000006_add_rbac.up.sql` or a new migration) and `RequirePermission` calls.
6. **Add OpenAPI paths** in `docs/api/openapi/openapi.yaml`.
7. **Add a frontend API wrapper** in `apps/web/features/<module>/api.ts`.

#### Example module skeleton

```go
package mymodule

type Module struct {
    Handler *Handler
    Service *Service
    Repo    *Repository
}

func NewModule(db *db.DB, cfg *config.Config) *Module {
    repo := NewRepository(db)
    svc := NewService(repo)
    h := NewHandler(svc)
    return &Module{Handler: h, Service: svc, Repo: repo}
}
```

---

### 6. How to add a migration

1. **Name the migration:** `apps/api/migrations/000018_add_<description>.{up,down}.sql`.
2. **Write up and down files.** Every migration must have a down script.
3. **Run migrations locally:**

```bash
cd apps/api
make migrate
## or
go run ./cmd/migrator
```

4. **Verify RLS:** if the new table is tenant-scoped, add `ALTER TABLE ... ENABLE ROW LEVEL SECURITY` and create policies in the same migration or a follow-up RLS migration.
5. **Update [`DATABASE_GUIDE.md`](../architecture/DATABASE_GUIDE.md)** with the new table and RLS status.

---

### 7. How to add a new API endpoint

1. **Define the route** in `apps/api/internal/shared/server/server.go` using `r.Group()` or `r.With()`.
2. **Add the handler method** in the appropriate module's `handler.go`.
3. **Add the service method** in `service.go`.
4. **Add the repository method** in `repository.go` if it touches the database.
5. **Add middleware chain:** typically `Auth`, `TenantContext`, `RequirePermission(<permission>)`, and `AuditLog` for writes.
6. **Add the request/response DTOs** in `handler.go` or a dedicated `dto.go`.
7. **Update OpenAPI** at `docs/api/openapi/openapi.yaml`.
8. **Add a frontend wrapper** in `apps/web/features/<module>/api.ts`.
9. **Add tests** in the module package.

---

### 8. How to add a frontend page

1. **Choose the route location.**
   - Public/auth pages go in `apps/web/app/(auth)/<route>/page.tsx`.
   - Authenticated pages go in `apps/web/app/(dashboard)/[workspace]/<route>/page.tsx`.
2. **Create the page component.** Use Server Components for data fetching where possible; Client Components for interactivity (`"use client"`).
3. **Add the API wrapper** in `apps/web/features/<module>/api.ts` if one does not exist.
4. **Use shared UI components** from `packages/ui` or `apps/web/components/ui`.
5. **Use `react-hook-form` + `zod`** for forms.
6. **Handle loading and error states.** Add `loading.tsx` and `error.tsx` alongside the page if needed.
7. **Update navigation** in `apps/web/components/dashboard/sidebar.tsx` or `dashboard/header.tsx` if the page is top-level.

---

### 9. Coding conventions

#### Go

- **One module per directory** under `internal/<module>/`.
- **No cross-module service calls.** Use interfaces/ports or database lookups.
- **Return sentinel errors** from `shared/errors` and map to HTTP status in handlers.
- **Use `*shared.DB` wrapper** so tenant context and transactions flow through.
- **Never pass `context.Background()` for side effects** inside a request (e.g., audit logging).
- **Tests live next to code** as `*_test.go`.

#### TypeScript / Next.js

- **Use the App Router.** Server Components by default.
- **Put data fetching in Server Components** or in `api.ts` wrappers, not inline in pages.
- **Do not access `localStorage` during SSR.** Guard with `typeof window !== "undefined"` or use a client hook.
- **Use `apiFetch` for all backend calls.** It parses the backend envelope and throws `ApiError`.
- **Zod schemas** should mirror backend DTOs.
- **Run `pnpm typecheck` and `pnpm lint`** before pushing.

#### SQL / Migrations

- **All tenant tables must have RLS enabled.**
- **Each migration needs a down file.**
- **Do not alter columns in place to avoid data loss;** prefer new migrations.
- **Keep seed data (roles/permissions) in migrations** so every environment is consistent.

---

### 10. Testing workflow

| Layer | Command | When to run |
|-------|---------|-------------|
| Go unit tests | `go test ./...` | After backend changes |
| Go lint/vet | `go vet ./...` | Before commit |
| TypeScript type check | `pnpm typecheck` | After TS/JS changes |
| ESLint | `pnpm lint` | Before commit |
| Full local stack | `pnpm dev` | Before PR |
| Integration tests | *Not yet implemented* | Target for P1 |

#### Testing tips

- The backend uses `pgx/stdlib` with a wrapped `*sql.DB`. For integration tests, spin up a test PostgreSQL database and apply migrations.
- Frontend tests should focus on critical flows: login → onboarding → create project → create test case → run test.
- Use `make migrate` with a test database URL before running integration tests.

---

### 11. Review workflow

1. **Branch from `main`.**
2. **One logical change per PR.** If a change touches API + web + docs, keep it in one PR because it is a monorepo.
3. **Update related handover docs** if the change affects architecture, routes, database, or technical debt.
4. **Ensure CI passes:** Go build + vet + test, Next.js build, Python lint, TypeScript typecheck.
5. **Request review from the relevant squad lead** (Platform, Core, Testing, Enterprise, Ecosystem).
6. **Squash-merge** once approved.

---

### 12. Common pitfalls

| Pitfall | Why it happens | How to avoid |
|---------|----------------|--------------|
| **Forgetting to set `app.tenant_id`** | Direct `db.Query` bypasses `TenantContext` | Always use `db.WithContext(ctx)` or the context-aware repository methods |
| **Cross-tenant data leakage** | Using `organization_id` from body without verifying membership | Use `TenantContext` + `RequirePermission`; resolve tenant from the resource, not the request body |
| **Access tokens expiring every 15 minutes** | No refresh logic | Implement P1 token refresh or extend token lifetime only for local dev |
| **SSE auth in browser** | `EventSource` cannot send `Authorization` headers | Query-token auth is implemented for local/MVP; harden with a signed SSE token or cookie before public networks |
| **Frontend hydration mismatch** | `localStorage` accessed during SSR | Use `useEffect` or a client-only wrapper |
| **Permission checks silently pass/fail** | Name drift between RBAC seed and middleware | Centralize permission strings in a Go `const` package and in frontend shared constants |
| **Audit events lost** | `context.Background()` used | Pass request context or use a durable queue |
| **Running migrations manually in production** | No CD pipeline | Automate migrations via `cmd/migrator` in CI/CD; never run by hand |
| **Editing the wrong dashboard route tree** | Two route trees exist | After P2 consolidation, use only `[workspace]` routes |

---

### 13. Getting help

- **Architecture questions:** re-read [`BIBLICAL_TESTRA.md`](../BIBLICAL_TESTRA.md) and [`SYSTEM_FLOWS.md`](../architecture/SYSTEM_FLOWS.md).
- **Database questions:** see [`DATABASE_GUIDE.md`](../architecture/DATABASE_GUIDE.md).
- **Current status:** [`PROJECT_OVERVIEW.md`](../PROJECT_OVERVIEW.md) and [`FEATURE_MATRIX.md`](../FEATURE_MATRIX.md).
- **What to build next:** [`ROADMAP.md`](ROADMAP.md).
- **Known issues:** [`ROADMAP.md`](ROADMAP.md) §Technical Debt Register.

If you are still stuck, ask in the engineering channel with the **request ID**, **module**, and **expected vs actual behavior**.

## See Also

- [`BIBLICAL_TESTRA.md`](../BIBLICAL_TESTRA.md) — canonical engineering handbook
- [`ENGINEERING_STANDARDS.md`](ENGINEERING_STANDARDS.md) — coding and review standards
- [`ROADMAP.md`](ROADMAP.md) — implementation phases and technical debt
