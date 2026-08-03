# Architecture

> This document is the high-level architecture guide for Testra. For module-level details, see [`backend-audit.md`](backend-audit.md), [`frontend-audit.md`](frontend-audit.md), [`migration-review.md`](migration-review.md), and [`infra-audit.md`](infra-audit.md).

## System context

```mermaid
flowchart TB
    subgraph Client["Client Layer"]
        Web["Next.js Web App (localhost:3000)"]
        CI["CI/CD Runner"]
    end

    Web -->|HTTPS / WebSocket| CDN["CDN / Vercel Edge"]
    CI -->|HTTPS POST /ingest| GW["Ingress / API Gateway"]

    CDN --> GW
    GW --> API["Go API (localhost:8080 /api/v1)"]

    API --> PG[("PostgreSQL 16<br/>OLTP + RLS")]
    API --> CH[("ClickHouse 24<br/>Analytics (future)")]
    API --> Redis[("Redis 7<br/>Cache / Queue")]
    API --> S3[("S3-compatible Object Storage<br/>MinIO / AWS S3")]

    API --> Worker["Go Worker (stub)"]
    API --> ML["Python ML Service (FastAPI)"]
```

## Backend architecture

Testra uses a **modular monolith** in Go. Every backend module follows the same Clean Architecture layers:

```mermaid
flowchart LR
    subgraph Module["Domain Module"]
        Domain["domain.go<br/>Entities / value objects"]
        Ports["ports.go<br/>Repository interfaces"]
        Repo["repository.go<br/>Postgres adapters"]
        Service["service.go<br/>Use cases"]
        Handler["handler.go<br/>HTTP adapters"]
        ModuleFile["module.go<br/>Dependency wiring"]
    end

    Domain --> Ports
    Ports --> Repo
    Repo --> Service
    Service --> Handler
    Handler --> ModuleFile
```

### Module list

| Module | Backend path | Responsibility |
|--------|--------------|----------------|
| `identity` | `apps/api/internal/identity/` | Registration, login, JWT, refresh tokens, MFA, password reset |
| `organization` | `apps/api/internal/organization/` | Organization CRUD and membership |
| `workspace` | `apps/api/internal/workspace/` | Workspace CRUD within an organization |
| `project` | `apps/api/internal/project/` | Project CRUD within a workspace |
| `apikeys` | `apps/api/internal/apikeys/` | Scoped API key lifecycle |
| `testmanagement` | `apps/api/internal/testmanagement/` | Test folders, suites, cases, versions, search |
| `results` | `apps/api/internal/results/` | Test runs, run items, status updates, SSE progress |
| `automationhub` | `apps/api/internal/automationhub/` | Ingest JUnit/Playwright/Cypress reports into runs |
| `audit` | `apps/api/internal/audit/` | Audit event persistence |
| `rbac` | `apps/api/internal/rbac/` | Permission loading for middleware |
| `shared` | `apps/api/internal/shared/` | Config, DB wrapper, JWT, errors, pagination, tenant resolver, middleware |

### Entry points

| Program | Path | Purpose |
|---------|------|---------|
| API server | `apps/api/cmd/api/main.go` | Loads env, opens DB, builds router, listens on `PORT` |
| Migrator | `apps/api/cmd/migrator/main.go` | Runs `golang-migrate` from `apps/api/migrations` |
| Worker | `apps/api/cmd/worker/main.go` | Stub; no background processing yet |

## Frontend architecture

```mermaid
flowchart TB
    subgraph Next["Next.js 15 App Router"]
        Root["app/layout.tsx"]
        Auth["(auth)/<br/>login, register, mfa-setup, onboarding"]
        Dash["(dashboard)/<br/>dashboard, [workspace]/..."]
        Settings["dashboard/settings/<br/>placeholder tabs"]
    end

    Root --> Auth
    Root --> Dash
    Dash --> Settings

    subgraph Features["Feature modules"]
        Platform["features/platform/api.ts"]
        TM["features/testmanagement/api.ts"]
        Results["features/results/api.ts"]
    end

    Dash --> Platform
    Dash --> TM
    Dash --> Results

    subgraph Lib["Shared utilities"]
        Api["lib/api.ts — fetch wrapper"]
        UI["components/ui/ — Tailwind primitives"]
    end

    Dash --> Lib
    Auth --> Lib
```

### Frontend stack

| Layer | Technology |
|-------|------------|
| Framework | Next.js 15 App Router |
| Language | TypeScript 5 |
| Runtime | React 18 |
| Styling | TailwindCSS 3, tailwind-merge, clsx |
| Forms | react-hook-form + Zod + @hookform/resolvers |
| Icons | lucide-react |
| State | Local `useState/useEffect` + `localStorage` (no global state library yet) |
| Package manager | pnpm 9.5 workspace |

See [`frontend-audit.md`](frontend-audit.md) for page-level details and findings.

## Database architecture

| Store | Technology | Role |
|-------|------------|------|
| Primary OLTP | PostgreSQL 16 | Users, orgs, workspaces, projects, test cases, runs, audit events |
| Analytics OLAP | ClickHouse 24 | Time-series test results, events, telemetry (not yet used) |
| Cache / Queue | Redis 7 | Sessions, rate limits, future Asynq jobs (not yet used) |
| Object Storage | MinIO / AWS S3 | Attachments, artifacts, exports (not yet used) |
| Search | PostgreSQL Full-Text Search | Test case search via `search_tsv` (Meilisearch planned for V2) |

### Multi-tenancy model

- **Shared database, shared schema** with `tenant_id` injected per HTTP request.
- PostgreSQL Row-Level Security (RLS) policies compare `app.tenant_id` against the row's organization or workspace chain.
- Tenant isolation is defense-in-depth: application layer resolves the tenant and sets it on a dedicated DB connection before any query executes.

See [`DATABASE_OVERVIEW.md`](DATABASE_OVERVIEW.md) and [`migration-review.md`](migration-review.md) for the full model.

## Authentication flow

```mermaid
sequenceDiagram
    autonumber
    participant U as User / Browser
    participant W as Web App
    participant A as Go API
    participant PG as PostgreSQL

    U->>W: Email + password
    W->>A: POST /api/v1/auth/login
    A->>PG: Verify password hash, check MFA
    PG-->>A: User + MFA state
    alt MFA enabled
        A-->>W: { mfa_required: true }
        U->>W: TOTP code
        W->>A: POST /api/v1/auth/mfa/verify
        A->>PG: Validate TOTP
    end
    A-->>A: Sign JWT access token (15 min)
    A-->>A: Create refresh-token family
    A-->>W: { token, refresh_token, user }
    W->>W: Store token in localStorage
    W->>A: Subsequent calls with Authorization: Bearer <token>
```

### Token model

| Token | Type | Storage | Lifetime |
|-------|------|---------|----------|
| Access token | JWT HS256 (`user_id`, `email`) | Client `localStorage` | 15 minutes (configurable) |
| Refresh token | Opaque `rt_` prefix + 32 bytes, SHA-256 stored | Client `localStorage` | 30 days sliding, 90 days absolute |

## Authorization flow

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant GW as Ingress
    participant API as Go API
    participant Auth as AuthMiddleware
    participant Tenant as TenantContext
    participant RBAC as RequirePermission
    participant H as Handler

    C->>GW: Request + Bearer token
    GW->>API: /api/v1/...
    API->>Auth: Validate JWT, set user_id in context
    API->>Tenant: Resolve org_id from URL/query/body, set app.tenant_id
    API->>RBAC: Load permissions for user + tenant
    RBAC->>H: Call handler if permission granted
    H->>H: Call service → repository on tenant-scoped connection
    H-->>C: Envelope { data, meta, error }
```

### Permission model

- System roles: `owner`, `admin`, `qa_engineer`, `viewer`.
- Permissions are namespaced strings such as `tests:create`, `runs:read`, `projects:create`.
- `RequirePermission` loads permissions once per request and caches them in context.
- Current scope is organization-only; workspace/project-level roles are not yet implemented.

## Tenant isolation and RLS flow

```mermaid
sequenceDiagram
    autonumber
    participant M as Middleware
    participant DB as DB Pool
    participant C as Dedicated Connection
    participant PG as PostgreSQL

    M->>DB: Acquire *sql.Conn
    DB-->>C: conn
    M->>C: SET app.tenant_id = '<org_id>'
    M->>C: Check organization membership
    C->>PG: Membership query (with RLS off for users/global tables)
    PG-->>C: Member confirmed
    M->>H: Pass conn through context
    H->>C: Repository query
    C->>PG: SELECT ... FROM test_cases
    PG->>PG: RLS policy filters by app.tenant_id
    PG-->>C: Tenant-scoped rows
    H-->>M: Response
    M->>C: RESET app.tenant_id
    M->>DB: Return conn
```

### Tables with RLS enabled

`organizations`, `organization_members`, `workspaces`, `workspace_members`, `projects`, `api_keys`, `role_assignments`, `test_folders`, `test_suites`, `test_cases`, `test_case_versions`, `test_runs`, `test_run_items`, `idempotency_records`.

Tables without RLS: `users`, `password_reset_tokens`, `roles`, `permissions`, `role_permissions`.

## Request lifecycle

```
HTTP request
  → Logger / Recoverer / RequestID / CORS / MaxBodySize (global)
  → Auth middleware (if protected)
  → TenantContext middleware (if tenant-scoped)
  → RequirePermission middleware (if permission required)
  → AuditLog / Idempotency middleware (write/idempotent operations)
  → Chi router dispatches to handler
  → Handler decodes request, calls service
  → Service executes business logic, calls repository
  → Repository uses context-aware DB connection (with tenant set)
  → Service returns domain model
  → Handler maps to JSON response envelope
  → Middleware post-processing (audit, idempotency store)
  → Response
```

## Middleware pipeline

All middleware lives in `apps/api/internal/shared/middleware/`.

| Middleware | Order | Purpose |
|------------|-------|---------|
| `Logger` | First | Request logging |
| `Recoverer` | Early | Panic recovery |
| `RequestID` | Early | Inject request correlation ID |
| `Content-Type` | Early | Default JSON content type |
| `CORS` | Early | Origin/method/header allow list |
| `MaxBodySize` | Before handlers | Limit body to 1 MB |
| `Auth` | Protected routes | Validate Bearer JWT and set `user_id` |
| `TenantContext` | Tenant-scoped routes | Resolve tenant and set `app.tenant_id` on a dedicated DB connection |
| `RequirePermission` | Per-route | Load and check required permission |
| `AuditLog` | Write routes | Fire-and-forget audit event write |
| `IdempotencyKey` | `POST /ingest` | Replay or store response keyed by idempotency key + body fingerprint |

## Repository → Service → Handler flow

```mermaid
flowchart TB
    subgraph HTTP["HTTP Layer"]
        H["handler.go<br/>Decode/encode JSON, HTTP status, route params"]
    end

    subgraph App["Application Layer"]
        S["service.go<br/>Business logic, validation, orchestration"]
    end

    subgraph Infra["Infrastructure Layer"]
        R["repository.go<br/>SQL queries using pgx/stdlib"]
        DB[("PostgreSQL with RLS")]
    end

    H -->|Calls| S
    S -->|Calls| R
    R -->|Queries| DB

    style HTTP fill:#f9f,stroke:#333
    style App fill:#bbf,stroke:#333
    style Infra fill:#bfb,stroke:#333
```

### Shared DB abstraction

`internal/shared/db/db.go` provides a `DB` wrapper that transparently uses either a transaction or a dedicated connection stored in `context`. This ensures `TenantContext` sets `app.tenant_id` once per request and all repository calls reuse that same connection.

## Automation Hub ingestion flow

```mermaid
sequenceDiagram
    autonumber
    participant CI as CI/CD Runner
    participant A as automationhub.Handler
    participant S as IngestService
    participant R as ResultsService
    participant PG as PostgreSQL

    CI->>A: POST /api/v1/ingest<br/>{ workspace_id, project_id, format, payload }
    A->>S: Ingest(ctx, req)
    S->>S: Parse JUnit XML or Playwright/Cypress JSON
    S->>R: Create test run (status = running)
    R->>PG: INSERT test_runs
    loop For each parsed case
        S->>R: Add test run item
        R->>PG: INSERT test_run_items
    end
    S->>R: Finalize run status + aggregates
    R->>PG: UPDATE test_runs
    R-->>S: RunProgressEvent broadcast
    S-->>A: IngestResult
    A-->>CI: 201 Created
```

### Ingestion details

- Supported formats: `junit`, `playwright`, `cypress`.
- Playwright and Cypress share the same JSON parser.
- The endpoint is protected by `runs:ingest` and the `IdempotencyKey` middleware.
- The current implementation requires a user JWT; scoped API-key auth is planned but not wired.

## Test execution flow

```mermaid
sequenceDiagram
    autonumber
    participant U as QA Engineer
    participant W as Web App
    participant A as results.Handler
    participant S as ResultsService
    participant PG as PostgreSQL
    participant SSE as SSE Stream

    U->>W: Create manual run (name + case IDs)
    W->>A: POST /api/v1/test-runs
    A->>S: CreateRun(ctx, input)
    S->>PG: INSERT test_runs (status = pending)
    S-->>W: Run created

    U->>W: Click "Start Run"
    W->>A: PUT /api/v1/test-runs/{id}
    A->>S: Update status to running
    S->>PG: UPDATE test_runs

    U->>W: Update item status
    W->>A: PUT /api/v1/test-run-items/{id}
    A->>S: UpdateItemStatus
    S->>PG: UPDATE test_run_items
    S->>S: Recalculate aggregates
    S->>PG: UPDATE test_runs
    S->>SSE: Broadcast RunProgressEvent

    W->>A: GET /api/v1/test-runs/{id}/stream
    A->>SSE: Subscribe
    SSE-->>W: SSE: data: { event }
```

### SSE progress stream

- `GET /api/v1/test-runs/{id}/stream` returns `text/event-stream`.
- The backend uses an in-memory `progressHub` to broadcast `RunProgressEvent` structs.
- **Known limitation:** `EventSource` in browsers cannot send custom headers; if the endpoint requires the `Authorization` header (as configured), browser clients will fail. See [`TECHNICAL_DEBT.md`](TECHNICAL_DEBT.md) and [`frontend-audit.md`](frontend-audit.md).

## Cross-cutting concerns

| Concern | Implementation |
|---------|----------------|
| **Configuration** | `internal/shared/config/config.go` reads env vars with typed defaults |
| **Errors** | Sentinel errors in `internal/shared/errors/errors.go`; handlers map to HTTP status |
| **HTTP envelope** | `internal/shared/http/response.go` returns `{ data, meta, error }` |
| **Pagination** | Cursor-based by `created_at DESC, id DESC` |
| **Validation** | `internal/shared/validation` for email, name, slug helpers |
| **Password hashing** | `bcrypt` via `internal/shared/password` |
| **JWT** | `golang-jwt/jwt/v5` in `internal/shared/jwt/jwt.go` |
| **Tenant resolution** | `internal/shared/tenant/resolver.go` joins workspaces/projects/keys/runs to resolve `organization_id` |
| **Rate limiting** | `LocalRateLimiter` exists but is **not wired** to any route |
