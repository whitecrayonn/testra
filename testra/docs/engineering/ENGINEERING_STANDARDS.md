# Testra — Engineering Standards

**Purpose:** Define coding, testing, infrastructure, security, and review standards for backend, frontend, and operations.
**Owner:** CTO / Engineering Lead
**Scope:** Coding standards for Go, TypeScript, SQL, API, database, security, infrastructure, and review workflows.
**Status:** Active
**Last Updated:** July 2026
**Source of Truth:** ENGINEERING_STANDARDS.md for coding and process standards.
**Related documents:**
- [`BIBLICAL_TESTRA.md`](../BIBLICAL_TESTRA.md)
- [`ONBOARDING.md`](ONBOARDING.md)
- [`DEPLOYMENT_GUIDE.md`](../deployment/DEPLOYMENT_GUIDE.md)
- [`SECURITY_CHECKLIST.md`](../security/SECURITY_CHECKLIST.md)

---

## 1. Backend (Go)

### 1.1 Project Layout

Each domain module in `internal/<domain>/` contains:
- `domain.go` — entities, value objects, domain errors
- `ports.go` — repository and service interfaces
- `repository.go` — SQL repository implementation
- `service.go` — use case / application logic
- `handler.go` — HTTP handlers
- `module.go` — wiring (NewModule function)
- `*_test.go` — unit tests for service layer

### 1.2 Code Rules

- **Go version:** 1.23+
- **Router:** chi v5
- **Linting:** `go vet` mandatory; `golangci-lint` recommended
- **Errors:** use sentinel errors from `shared/errors`; wrap with context using `fmt.Errorf("...: %w", err)`
- **No `panic`** in request paths. Panics only in `init()` or impossible-state guards.
- **Context propagation:** every function that does I/O takes `context.Context` as first parameter.
- **UUIDs:** use `github.com/google/uuid` for all entity IDs.
- **Time:** always use `time.Now().UTC()` for stored timestamps.
- **Interfaces:** defined at consumer side (ports), not at implementation side.
- **SQL:** parameterized queries only. No string concatenation for SQL.
- **JSON tags:** always present on response structs. Use `snake_case`.

### 1.3 Testing

- Table-driven tests for service layer.
- Use fake/mock repositories implementing port interfaces.
- Test behavior, not implementation details.
- `go test -race -count=1 ./...` must pass.
- Integration tests (with real Postgres) in `tests/` package, tagged with `//go:build integration`.

---

## 2. Frontend (Next.js + TypeScript)

### 2.1 Project Layout

- `app/` — Next.js App Router (route groups: `(auth)`, `(dashboard)`)
- `features/` — feature-based modules mirroring backend domains
- `components/` — shared UI components
- `lib/` — utilities, API client, hooks
- `types/` — shared TypeScript types (generated from OpenAPI where possible)

### 2.2 Code Rules

- **TypeScript:** `strict: true`, no `any`, no `// @ts-ignore`
- **Framework:** Next.js 15 App Router
- **Styling:** TailwindCSS 4 + shadcn/ui
- **Server state:** Server Components + lightweight API client (TanStack Query planned)
- **Client state:** React `useState` / `useContext` (Zustand or TanStack Store under evaluation)
- **Forms:** React Hook Form + Zod
- **Tables:** TanStack Table (planned)
- **Charts:** Tremor / Recharts (planned)
- **Imports:** absolute paths via `@/` alias (tsconfig paths)
- **Server Components by default**; Client Components only for interactivity (`"use client"`)
- **No inline styles**; use Tailwind classes
- **Accessibility:** Radix UI primitives for interactive components

### 2.3 Testing

- Vitest for unit tests (planned)
- React Testing Library for component tests (planned)
- Playwright for E2E (future)
- `pnpm turbo run typecheck` must pass

---

## 3. API

### 3.1 Design Rules

- **RESTful JSON API** with OpenAPI 3.1 as source of truth
- **Versioning:** URL-based major versions (`/api/v1/...`); compatible changes remain in the major version.
- **Pagination:** cursor pagination for all new list endpoints; default 50, maximum 100.
- **Idempotency:** `Idempotency-Key` required for side-effecting commands, ingestion, exports, webhooks, and payment-like operations; PostgreSQL record retention is 24 hours.
- **Resource-oriented** endpoints aligned with domain modules
- **Consistent envelope:**
  ```json
  { "data": {}, "meta": {}, "error": null }
  ```
- **HTTP status codes:** 200 (OK), 201 (Created), 204 (No Content), 400 (Bad Request), 401 (Unauthorized), 403 (Forbidden), 404 (Not Found), 409 (Conflict), 422 (Unprocessable Entity), 500 (Internal Error)
- **Error format:**
  ```json
  { "error": { "code": "NOT_FOUND", "message": "resource not found" } }
  ```

### 3.2 Documentation

- OpenAPI spec at `docs/api/openapi/openapi.yaml`
- Every endpoint documented before implementation
- Schemas for all request and response bodies
- Scalar or Swagger UI for interactive docs (future endpoint)
- SDK generated from spec (`packages/sdk`)

### 3.3 Security

- Bearer JWT for session auth
- Scoped API keys for CI/CD (hashed in DB, one-time display)
- Rate limiting via Redis token bucket per tenant/API key
- Request ID on every request (chi middleware)
- CORS configured per environment

---

## 4. Database

### 4.1 PostgreSQL

- **Migrations:** `golang-migrate` with sequential numbered files (`000NNN_description.{up,down}.sql`)
- **Migrations path:** `apps/api/migrations/`
- **Every migration has up AND down SQL**
- **Never modify a merged migration** — create a new one instead
- **UUIDs** for all primary keys
- `tenant_id` (or `organization_id`) on every tenant-scoped table
- `created_at` and `updated_at` as `TIMESTAMPTZ NOT NULL DEFAULT NOW()` on every table
- Foreign keys with `ON DELETE CASCADE` for parent-owned children
- Indexes on all foreign key columns and frequently queried columns
- Row-level security policies are mandatory on tenant-scoped tables in staging and production; transaction-local `app.tenant_id` is set after authenticated scope resolution and API roles do not bypass RLS.

### 4.2 ClickHouse

- Used for test results, events, time-series data only
- No transactional data
- MergeTree engine family
- Tenant isolation via `tenant_id` column in every table

### 4.3 Redis

- Sessions, rate limiting, job queues (Asynq)
- No persistent business data
- Keys prefixed with `testra:` namespace

---

## 5. Security

- **Password policy:** minimum 12 characters; approved maintained password hashing; single-use reset tokens expire after 30 minutes.
- **JWT:** 15-minute access tokens signed by rotated secret-managed keys.
- **Refresh tokens:** opaque, rotating, hashed in PostgreSQL; 30-day inactivity expiry and 90-day absolute expiry; reuse revokes the session family.
- **MFA:** TOTP; required for organization administrators and enterprise users, enforceable organization-wide.
- **API keys:** SHA-256 hashed, shown once, scoped per organization/workspace/project, 90-day default expiry, 365-day maximum.
- **Session revocation:** per-session, user-wide, password-change, MFA-reset, and compromise revocation are mandatory.
- **No secrets in code or commits** — `.env` files gitignored
- **Input validation** on every endpoint — reject early, fail fast
- **SQL injection prevention** — parameterized queries only, no string concat
- **Rate limiting:** Redis token buckets; login 10/IP/15 minutes and 5/account/15 minutes, registration 5/IP/hour, password reset 5/account/hour, API keys 120 requests/minute by default.
- **CORS** — restrict origins per environment
- **TLS:** Let's Encrypt (certbot) on Nginx for the single Ubuntu VPS; a CDN/WAF may be added later if justified.

---

## 6. Infrastructure

### 6.1 Deployment Roadmap

- **Local:** Native development with locally installed PostgreSQL, Redis, Mailpit, and MinIO (no Docker, see ADR-009).
- **MVP:** Single Ubuntu VPS with systemd + Nginx, running Go API, Go worker, Next.js, and Python ML as systemd services. PostgreSQL and Redis on the same VPS; MinIO optional. Cloud-managed services are a future evolution path only after measured scale (see ADR-003, ADR-009).
- **Beta:** Single VPS or small VPS fleet, with backups, replication, and monitoring.
- **Enterprise:** Managed platform or dedicated capacity only after measured need.

### 6.2 Build & Packaging

- **No Docker.** Builds produce native binaries and a Next.js standalone output.
- **Go:** `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build` for the API, worker, migrator, and OpenAPI commands.
- **Web:** `next build` producing `.next/standalone` for serving behind Nginx or as a systemd service.
- **ML:** Python FastAPI served via `uvicorn` as a systemd service.
- **Artifacts tagged by commit SHA** in CI.

### 6.3 Deployment (MVP)

- systemd unit files for API, worker, web, ML, Nginx, PostgreSQL, Redis, and optional MinIO.
- Environment files per stage; secrets never committed.
- Nginx reverse proxy and TLS termination via Let's Encrypt.
- Backup/restore and log rotation runbooks.

### 6.4 Cloud/IaC (Future Evolution Only)

- Container orchestration, Terraform, and managed cloud services are **not part of MVP**.
- If scale justifies it, evaluate managed platforms or dedicated capacity after product-market fit.

### 6.5 Observability

- OpenTelemetry for traces
- Prometheus for metrics
- Grafana for dashboards
- Loki for logs
- Structured logging (JSON) in Go via `log/slog` or compatible
- Application logs: 30 days hot, 90 days archived
- Metrics: 15 months
- Traces: 14 days
- Audit records: minimum 2 years MVP/Beta and 7 years for Enterprise governance

---

## 7. Performance Targets

MVP targets are defined in ADR-008: read API p95 ≤ 300 ms, write API p95 ≤ 500 ms, indexed PostgreSQL query p95 ≤ 50 ms, 30-second maximum synchronous request timeout, 5-minute background job timeout, 25 MiB request uploads, 500 concurrent authenticated users, 50 sustained requests/second, and 10,000 ClickHouse result records/minute.

## 8. Testing

### 8.1 Test Pyramid

| Level | What | Tool | When |
|---|---|---|---|
| **Unit** | Service/domain logic | Go testing, Vitest | Every PR |
| **Integration** | Repository + DB | Go testing (+build tag) | Every PR (with compose) |
| **Contract** | API vs OpenAPI spec | Contract test tool | Every PR |
| **E2E** | Full user flows | Playwright | Pre-release (future) |

### 8.2 Rules

- Never delete or weaken tests without explicit direction
- Test names: `TestServiceCreate`, `TestServiceCreateDuplicateKey` — descriptive
- Table-driven tests for multiple scenarios
- Assert on behavior, not implementation
- Each test is independent — no shared mutable state between tests

---

## 9. Documentation

### 9.1 Engineering Docs

- `docs/engineering/ONBOARDING.md` — governance, onboarding, and contributor rules (update only when governance changes)
- `docs/engineering/ROADMAP.md` — roadmap, phases, and technical debt (update when phases change)
- `docs/engineering/ENGINEERING_STANDARDS.md` — standards (update only when standards change)
- `docs/engineering/progress/` — session reports (append-only)

### 9.2 Architecture Docs

- `docs/architecture/adrs/` — Architecture Decision Records
- New ADR required for any deviation from approved architecture documents
- ADR format: Context → Decision → Consequences

### 9.3 API Docs

- `docs/api/openapi/openapi.yaml` — OpenAPI 3.1 spec
- Updated before implementation of new endpoints

### 9.4 Code Documentation

- Go: doc comments on exported functions and types
- TypeScript: JSDoc on exported functions when non-obvious
- No inline comments explaining *what* — only *why* when non-obvious
- No commented-out code in committed files

---

## 10. Git Workflow

### 10.1 Branching

- **Trunk-based:** short-lived feature branches → PR → squash-merge to `main`
- Branch naming: `feat/<scope>`, `fix/<scope>`, `chore/<scope>`, `docs/<scope>`
- `main` is always deployable
- No `develop` or `release` branches (solo dev)

### 10.2 Commits

- **Conventional commits:** `type(scope): description`
- Types: `feat`, `fix`, `chore`, `docs`, `test`, `refactor`, `ci`, `build`
- Scope: domain module name (e.g., `identity`, `project`, `web`)
- Examples:
  - `feat(project): add create endpoint with key validation`
  - `fix(compose): resolve MinIO/ClickHouse port conflict`
  - `docs(engineering): add master development guide`

### 10.3 PR Rules

- One PR per feature/fix — atomic, reviewable
- PR title follows conventional commit format
- PR description includes: what changed, why, testing notes
- All CI checks must pass before merge
- Squash-merge only

---

## 11. Code Review Expectations

### 11.1 Reviewer Checklist

- [ ] Does the code follow Clean Architecture boundaries?
- [ ] Are imports at the top and used?
- [ ] Is error handling correct (wrapped, sentinel errors)?
- [ ] Are there tests for new domain logic?
- [ ] Is the OpenAPI spec updated if endpoints changed?
- [ ] Are there migrations if schema changed?
- [ ] No secrets, no hardcoded values?
- [ ] Naming is clear and consistent?
- [ ] No cross-module internal imports?
- [ ] Input validation on every endpoint?

### 11.2 Author Checklist

- [ ] Self-review completed (see ONBOARDING.md §6)
- [ ] DoD met (see ONBOARDING.md §5)
- [ ] `make test && make lint` pass locally
- [ ] `ROADMAP.md` updated if phase status changed
- [ ] Progress report saved if session produced meaningful work

---

## See Also

- [`BIBLICAL_TESTRA.md`](../BIBLICAL_TESTRA.md) — engineering handbook, canonical sources, and do-not-break list.
- [`ONBOARDING.md`](ONBOARDING.md) — contributor workflow, DoD, self-review, and getting started.
- [`API_DESIGN_GUIDELINES.md`](../api/API_DESIGN_GUIDELINES.md) — REST, OpenAPI, versioning, and idempotency conventions.
- [`SECURITY_CHECKLIST.md`](../security/SECURITY_CHECKLIST.md) — security review checklist.
- [`DEPLOYMENT_GUIDE.md`](../deployment/DEPLOYMENT_GUIDE.md) — deployment, infrastructure, and rollback guidance.
