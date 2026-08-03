# Engineer Onboarding

> Welcome. This guide assumes you are a senior engineer joining Testra with no prior conversation. Read it alongside [`README.md`](README.md), [`PROJECT_OVERVIEW.md`](PROJECT_OVERVIEW.md), [`ARCHITECTURE.md`](ARCHITECTURE.md), and [`CURRENT_STATE.md`](CURRENT_STATE.md).

## 1. Understand the repository

### Top-level layout

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
├── infra/
│   ├── docker/       # Docker Compose and images
│   ├── k8s/          # Kubernetes base manifests
│   └── terraform/    # Terraform scaffold
├── docs/
│   ├── handover/     # This wiki
│   ├── api/openapi/  # OpenAPI 3.1 spec
│   ├── architecture/ # ADRs
│   └── ...
├── scripts/          # Dev scripts (pnpm, Go, Python)
├── Makefile
├── pnpm-workspace.yaml
└── go.work
```

### How to read code

1. Start with `apps/api/internal/shared/server/server.go` to see how routes, middleware, and modules are wired.
2. Pick one complete module (e.g., `apps/api/internal/testmanagement/`) and read `domain.go`, `ports.go`, `repository.go`, `service.go`, `handler.go`, `module.go`.
3. Read the OpenAPI spec at `docs/api/openapi/openapi.yaml` for the public contract.
4. Explore the frontend from `apps/web/app/(dashboard)/layout.tsx` down through `[workspace]/test-cases/page.tsx`.

---

## 2. How modules interact

Testra is a **modular monolith**: all modules live in one Go binary but have strict internal boundaries.

- **No module calls another module's service directly.** If a module needs data from another scope, it looks it up through the database or a future port/interface.
- **Integration point is PostgreSQL.** Tenant context and shared entities (workspace, project) are resolved in middleware and passed on the request context.
- **Shared package is for cross-cutting concerns only:** config, errors, HTTP response helpers, pagination, JWT, password hashing, tenant resolution, middleware.
- **Frontend features mirror backend modules:** `apps/web/features/testmanagement/`, `features/results/`, `features/platform/`, etc.

---

## 3. How requests flow

### Backend request

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

### Frontend request

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

> **Note:** The frontend currently does not refresh tokens automatically. Access tokens expire after 15 minutes. See [`TECHNICAL_DEBT.md`](TECHNICAL_DEBT.md) H5.

---

## 4. Where to start debugging

| Symptom | Where to look |
|---------|---------------|
| 401/403 on API | `shared/middleware/auth.go`, `shared/middleware/tenant.go`, `shared/middleware/permission.go` |
| Wrong tenant data | `shared/tenant/resolver.go`, `shared/db/db.go` connection context, RLS policies in `migration-review.md` |
| SQL error | `internal/<module>/repository.go`, `shared/db/db.go` |
| Handler panic | `shared/middleware/recoverer.go`, `shared/middleware/logger.go` |
| Frontend 401 after 15 min | `lib/api.ts`, no token refresh logic; also check `localStorage` keys |
| SSE not working | `apps/api/internal/results/handler.go` stream endpoint and `apps/web/app/(dashboard)/[workspace]/test-runs/[id]/page.tsx` `EventSource` usage |
| UI mismatch with API | `apps/web/features/<module>/api.ts` payload/response vs OpenAPI schema vs Go DTO |
| Build/lint failure | `pnpm typecheck` output, `go vet ./...`, GitHub Actions logs |

### Useful commands

```bash
# Backend
go test ./...
go vet ./...
go run ./cmd/api

# Frontend
pnpm install
pnpm dev
pnpm typecheck
pnpm lint

# Database
cd apps/api && make migrate
# or
go run ./cmd/migrator

# Full stack
pnpm dev
```

---

## 5. How to add a new backend module

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

### Example module skeleton

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

## 6. How to add a migration

1. **Name the migration:** `apps/api/migrations/000018_add_<description>.{up,down}.sql`.
2. **Write up and down files.** Every migration must have a down script.
3. **Run migrations locally:**

```bash
cd apps/api
make migrate
# or
go run ./cmd/migrator
```

4. **Verify RLS:** if the new table is tenant-scoped, add `ALTER TABLE ... ENABLE ROW LEVEL SECURITY` and create policies in the same migration or a follow-up RLS migration.
5. **Update [`DATABASE_OVERVIEW.md`](DATABASE_OVERVIEW.md)** with the new table and RLS status.

---

## 7. How to add a new API endpoint

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

## 8. How to add a frontend page

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

## 9. Coding conventions

### Go

- **One module per directory** under `internal/<module>/`.
- **No cross-module service calls.** Use interfaces/ports or database lookups.
- **Return sentinel errors** from `shared/errors` and map to HTTP status in handlers.
- **Use `*shared.DB` wrapper** so tenant context and transactions flow through.
- **Never pass `context.Background()` for side effects** inside a request (e.g., audit logging).
- **Tests live next to code** as `*_test.go`.

### TypeScript / Next.js

- **Use the App Router.** Server Components by default.
- **Put data fetching in Server Components** or in `api.ts` wrappers, not inline in pages.
- **Do not access `localStorage` during SSR.** Guard with `typeof window !== "undefined"` or use a client hook.
- **Use `apiFetch` for all backend calls.** It parses the backend envelope and throws `ApiError`.
- **Zod schemas** should mirror backend DTOs.
- **Run `pnpm typecheck` and `pnpm lint`** before pushing.

### SQL / Migrations

- **All tenant tables must have RLS enabled.**
- **Each migration needs a down file.**
- **Do not alter columns in place to avoid data loss;** prefer new migrations.
- **Keep seed data (roles/permissions) in migrations** so every environment is consistent.

---

## 10. Testing workflow

| Layer | Command | When to run |
|-------|---------|-------------|
| Go unit tests | `go test ./...` | After backend changes |
| Go lint/vet | `go vet ./...` | Before commit |
| TypeScript type check | `pnpm typecheck` | After TS/JS changes |
| ESLint | `pnpm lint` | Before commit |
| Full local stack | `pnpm dev` | Before PR |
| Integration tests | *Not yet implemented* | Target for P1 |

### Testing tips

- The backend uses `pgx/stdlib` with a wrapped `*sql.DB`. For integration tests, spin up a test PostgreSQL database and apply migrations.
- Frontend tests should focus on critical flows: login → onboarding → create project → create test case → run test.
- Use `make migrate` with a test database URL before running integration tests.

---

## 11. Review workflow

1. **Branch from `main`.**
2. **One logical change per PR.** If a change touches API + web + docs, keep it in one PR because it is a monorepo.
3. **Update related handover docs** if the change affects architecture, routes, database, or technical debt.
4. **Ensure CI passes:** Go build + vet + test, Next.js build, Python lint, TypeScript typecheck.
5. **Request review from the relevant squad lead** (Platform, Core, Testing, Enterprise, Ecosystem).
6. **Squash-merge** once approved.

---

## 12. Common pitfalls

| Pitfall | Why it happens | How to avoid |
|---------|----------------|--------------|
| **Forgetting to set `app.tenant_id`** | Direct `db.Query` bypasses `TenantContext` | Always use `db.WithContext(ctx)` or the context-aware repository methods |
| **Cross-tenant data leakage** | Using `organization_id` from body without verifying membership | Use `TenantContext` + `RequirePermission`; resolve tenant from the resource, not the request body |
| **Access tokens expiring every 15 minutes** | No refresh logic | Implement P1 token refresh or extend token lifetime only for local dev |
| **SSE fails in browser** | `EventSource` cannot send `Authorization` | Fix in P0 with cookie or query token; do not keep adding header workarounds |
| **Frontend hydration mismatch** | `localStorage` accessed during SSR | Use `useEffect` or a client-only wrapper |
| **Permission checks silently pass/fail** | Name drift between RBAC seed and middleware | Centralize permission strings in a Go `const` package and in frontend shared constants |
| **Audit events lost** | `context.Background()` used | Pass request context or use a durable queue |
| **Running migrations manually in production** | No CD pipeline | Automate migrations via `cmd/migrator` in CI/CD; never run by hand |
| **Editing the wrong dashboard route tree** | Two route trees exist | After P2 consolidation, use only `[workspace]` routes |

---

## 13. Getting help

- **Architecture questions:** re-read [`ARCHITECTURE.md`](ARCHITECTURE.md).
- **Database questions:** see [`DATABASE_OVERVIEW.md`](DATABASE_OVERVIEW.md) and [`migration-review.md`](migration-review.md).
- **Current status:** [`CURRENT_STATE.md`](CURRENT_STATE.md) and [`FEATURE_MATRIX.md`](FEATURE_MATRIX.md).
- **What to build next:** [`NEXT_STEPS.md`](NEXT_STEPS.md).
- **Known issues:** [`TECHNICAL_DEBT.md`](TECHNICAL_DEBT.md).

If you are still stuck, ask in the engineering channel with the **request ID**, **module**, and **expected vs actual behavior**.
