# Backend Audit & Migration Review

**Scope:** `testra/apps/api` Go backend — modules, routes, middleware, authentication, tenant isolation, RBAC, and the `testra/apps/api/migrations` schema.

**Status:** Audit complete. This document is a handover summary of the current implementation as of the latest review.

---

## 1. Executive Summary

The Testra backend is a Go HTTP API built on `go-chi/chi/v5`, PostgreSQL (via `pgx/stdlib`), and `golang-migrate`. It follows a layered package structure inside every domain module:

```
domain.go -> ports.go -> repository.go -> service.go -> handler.go -> module.go
```

Shared concerns live in `internal/shared/` and are wired together by `internal/shared/server/server.go`. The API implements identity, organizations, workspaces, projects, API keys, test management, test runs, CI result ingestion, audit logging, idempotency, and a role-based permission system backed by PostgreSQL Row Level Security (RLS).

---

## 2. Architecture & Entry Points

### 2.1 Top-level programs

| Program | Path | Purpose |
|--------|------|---------|
| API server | `apps/api/cmd/api/main.go` | Loads config, opens DB, builds router, listens on `PORT` |
| Migrator | `apps/api/cmd/migrator/main.go` | Runs `golang-migrate` up from `apps/api/migrations` |
| Worker | `apps/api/cmd/worker/main.go` | Stub (`fmt.Println`) — no background processing yet |

### 2.2 Server wiring

`apps/api/internal/shared/server/server.go` is the single place where all modules, middleware, and routes are assembled:

- Uses `chi.NewRouter()`.
- Applies global middleware: `Logger`, `Recoverer`, `RequestID`, `Content-Type`, CORS, `MaxBodySize(1MB)`.
- Instantiates modules and wires shared utilities (`tenant.Resolver`, `rbac.SQLPermissionLoader`, `idempotency.PostgresStore`, `audit.Service`).
- Mounts all routes under `/api/v1`.
- Exposes `/health`.

---

## 3. Domain Modules

Each module below uses the same layered pattern unless noted.

| Module | Path | Responsibility |
|--------|------|----------------|
| `identity` | `apps/api/internal/identity/` | Registration, login, JWT, refresh tokens, MFA TOTP, password reset |
| `organization` | `apps/api/internal/organization/` | Org CRUD and member add |
| `workspace` | `apps/api/internal/workspace/` | Workspace CRUD within org, member add |
| `project` | `apps/api/internal/project/` | Project CRUD within workspace |
| `apikeys` | `apps/api/internal/apikeys/` | Scoped API key generation, listing, revocation |
| `testmanagement` | `apps/api/internal/testmanagement/` | Test folders, suites, cases, versioning, search |
| `results` | `apps/api/internal/results/` | Test runs, run items, status updates, SSE progress stream |
| `automationhub` | `apps/api/internal/automationhub/` | Ingest JUnit / Playwright / Cypress payloads into runs |
| `audit` | `apps/api/internal/audit/` | Audit event persistence |
| `rbac` | `apps/api/internal/rbac/` | `SQLPermissionLoader` for permission lookup |

Several placeholder directories contain only `.gitkeep`: `analytics`, `apitesting`, `billing`, `defects`, `integrationhub`, `intelligence`, `notification`.

---

## 4. HTTP Route Map (`/api/v1`)

Routes are grouped by the `TenantContext` resolver they share. All routes except auth are behind `Auth` middleware.

### Auth (no tenant context)

```
POST   /auth/register
POST   /auth/login
POST   /auth/refresh
POST   /auth/password-reset/request
POST   /auth/password-reset/confirm
GET    /auth/me          (Auth)
POST   /auth/mfa/setup   (Auth)
POST   /auth/mfa/verify  (Auth)
POST   /auth/mfa/disable (Auth)
```

### Organizations

```
POST   /organizations                           (Auth only, no tenant/permission gate)
GET    /organizations                         (Auth only)
GET    /organizations/{id}                    (TenantContext + orgs:read)
```

### Workspaces

```
POST   /workspaces                              (TenantContext via body org_id + workspaces:create)
GET    /workspaces?organization_id=...        (TenantContext via query org_id + workspaces:read)
GET    /workspaces/{id}                         (TenantContext via workspace_id + workspaces:read)
```

### Projects & API Keys

```
POST   /projects                                (TenantContext via body workspace_id + projects:create)
GET    /projects?workspace_id=...               (TenantContext via query workspace_id + projects:read)
GET    /projects/{id}                           (TenantContext via project_id + projects:read)
POST   /api-keys                                (TenantContext via body workspace_id + apikeys:create, AuditLog)
GET    /api-keys?workspace_id=...               (TenantContext via query workspace_id + apikeys:read)
DELETE /api-keys/{id}                           (TenantContext via api_key_id + apikeys:delete, AuditLog)
```

### Test Management

```
POST   /test-folders   (tests:create, AuditLog)
POST   /test-suites    (tests:create, AuditLog)
POST   /test-cases     (tests:create, AuditLog)

GET    /test-folders?workspace_id=...
GET    /test-suites?workspace_id=...
GET    /test-cases/search?workspace_id=...   (full-text search)
GET    /test-cases?project_id=...
GET    /test-cases/{id}
GET    /test-cases/{id}/versions
PUT    /test-cases/{id}    (tests:update, AuditLog)
DELETE /test-cases/{id}    (tests:delete, AuditLog)

GET    /test-folders/{id}
GET    /test-suites/{id}
PUT    /test-folders/{id}  (tests:update, AuditLog)
PUT    /test-suites/{id}   (tests:update, AuditLog)
DELETE /test-folders/{id}  (tests:delete, AuditLog)
DELETE /test-suites/{id}   (tests:delete, AuditLog)
```

All test-management routes require `tests:*` permissions and resolve the tenant through `workspace_id` (body or query) or `project_id` (query/path), then `workspace_id`.

### Test Runs & Ingestion

```
POST   /test-runs                              (runs:create, AuditLog)
GET    /test-runs?project_id=...               (runs:read)
GET    /test-runs/{id}                         (runs:read)
GET    /test-runs/{id}/items                   (runs:read)
GET    /test-runs/{id}/stream                  (runs:read, SSE)
PUT    /test-runs/{id}                         (runs:update, AuditLog)
DELETE /test-runs/{id}                         (runs:delete, AuditLog)
PUT    /test-run-items/{id}                    (runs:update, AuditLog)
POST   /ingest                                   (runs:ingest, AuditLog, IdempotencyKey)
```

`/ingest` is the CI/CD ingestion endpoint handled by `automationhub.Handler.Ingest`.

### Notifications

```
GET    /notifications                              (notifications:read)
GET    /notifications/unread-count                 (notifications:read)
POST   /notifications                              (notifications:create, AuditLog)
PATCH  /notifications/{id}                         (notifications:update)
DELETE /notifications/{id}                         (notifications:delete)

GET    /notification-preferences                   (notification_preferences:read)
PUT    /notification-preferences                   (notification_preferences:update)

GET    /notification-channels                      (notification_channels:read)
POST   /notification-channels                      (notification_channels:create, AuditLog)
PUT    /notification-channels/{id}                 (notification_channels:update, AuditLog)
DELETE /notification-channels/{id}                 (notification_channels:delete, AuditLog)
```

---

## 5. Middleware Stack

All files in `apps/api/internal/shared/middleware/`.

| Middleware | Purpose | Usage Notes |
|-----------|---------|-------------|
| `Auth` | Validates `Authorization: Bearer <JWT>` or `?access_token=<JWT>` query parameter and stores `user_id` in context | Global for protected routes; query-token support required for browser `EventSource` |
| `TenantContext` | Acquires a dedicated DB conn, sets `app.tenant_id`, checks membership | Wrapped around all tenant-scoped routes |
| `RequirePermission` | Loads user permissions for the resolved tenant and checks a required permission | Per-route |
| `AuditLog` | Captures status `< 400` writes with user/action/resource/resource_id/IP | Applied to create/update/delete routes |
| `IdempotencyKey` | Replays or stores responses keyed by `Idempotency-Key` header + body fingerprint | Only on `POST /ingest` |
| `MaxBodySize` | Limits request body to 1 MB | Global |
| `RateLimit` / `LocalRateLimiter` | In-memory token bucket rate limiting | Created but **not wired** to any route (`rateLimitCfg` is unused) |
| CORS | Origin/method/header allow list | Global, configured via `CORS_ALLOWED_ORIGINS` |

### Tenant resolution helpers

`tenant.go` provides resolver factories that translate a request artifact into the organization ID used for RLS:

- `OrgIDFromURLParam`
- `OrgIDFromQuery`
- `OrgIDFromBody`
- `WorkspaceToOrg` — uses `tenant.Resolver.ResolveOrgFromWorkspace`
- `ProjectToOrg` — uses `tenant.Resolver.ResolveOrgFromProject`
- `APIKeyToOrg` — uses `tenant.Resolver.ResolveOrgFromAPIKey`
- `RunItemToOrg` — uses `tenant.Resolver.ResolveOrgFromRunItem`

`tenant.Resolver` in `apps/api/internal/shared/tenant/resolver.go` contains the SQL that joins workspaces/projects/api_keys/test_runs to determine the organization ID.

---

## 6. Authentication & Identity

`apps/api/internal/identity/service.go`:

- **JWT access tokens:** `HS256`, 15-minute expiry, claims `user_id` + `email`.
- **Refresh tokens:** Opaque `rt_` prefix, 32 random bytes, SHA-256 stored as `token_hash`; families support rotation. If a revoked token is used, the entire family is revoked (`ErrTokenRevoked`).
- **Passwords:** `bcrypt` via `internal/shared/password`.
- **MFA TOTP:** `pquerna/otp/totp`. `SetupMFA` returns secret + QR URL; `VerifyMFA` enables it; `DisableMFA` clears secret.
- **Password reset:** 30-minute tokens, SHA-256 hashed, plain token emailed over `net/smtp` when SMTP is configured.

### Password policy

`validatePassword` rejects passwords shorter than 12 characters. There is no complexity rule beyond length.

### Refresh token lifetime

`RefreshExpiryDays` (default 30) and `RefreshAbsoluteDays` (default 90) are both applied, with the earlier expiration winning.

---

## 7. Tenant Isolation & Row Level Security

### Mechanism

1. `TenantContext` middleware acquires a `*sql.Conn` from `cfg.DB`.
2. It executes `SET app.tenant_id = '<org_id>'` on that connection.
3. It calls `tenant.Resolver.CheckMembership` to confirm the user belongs to the org.
4. All repository queries for that request execute on the same connection/transaction.
5. Before returning the connection to the pool, `RESET app.tenant_id` is executed.

`internal/shared/db/db.go` (`DB` wrapper) transparently uses a transaction or connection stored in `context` so that middleware-set state applies to every repository call.

### Tables with RLS enabled

- `organizations`, `organization_members`
- `workspaces`, `workspace_members`
- `projects`
- `api_keys`
- `role_assignments`
- `test_folders`, `test_suites`, `test_cases`, `test_case_versions`
- `test_runs`, `test_run_items`
- `idempotency_records`

Each RLS policy compares the request's `app.tenant_id` (an organization UUID) against the row's `organization_id` or the `organization_id` of the workspace chain. See `apps/api/migrations/000009_add_rls_policies.up.sql` and `000014/000015` for test-management and run RLS.

---

## 8. RBAC

### Schema

- `roles` — system roles: `owner`, `admin`, `qa_engineer`, `viewer`.
- `permissions` — named permissions such as `projects:create`, `tests:read`, `runs:ingest`.
- `role_permissions` — many-to-many mapping.
- `role_assignments` — `(role_id, user_id, scope_type, scope_id)`; currently `scope_type` defaults to `'organization'` and `scope_id` is the organization UUID.

### Permission loading

`internal/rbac/loader.go` `SQLPermissionLoader.LoadPermissions` returns `DISTINCT p.name WHERE ra.user_id = $1 AND ra.scope_type = $2 AND ra.scope_id = $3`.

`RequirePermission` middleware loads permissions once per request and caches them in context under `permKey`.

### Permission naming used in routes

| Resource | Permissions |
|----------|-------------|
| Organizations | `orgs:read` |
| Workspaces | `workspaces:create`, `workspaces:read` |
| Projects | `projects:create`, `projects:read` |
| API Keys | `apikeys:create`, `apikeys:read`, `apikeys:delete` |
| Tests | `tests:create`, `tests:read`, `tests:update`, `tests:delete` |
| Runs | `runs:create`, `runs:read`, `runs:update`, `runs:delete`, `runs:ingest` |

---

## 9. Domain Logic Highlights

### Test Management (`testmanagement`)

- Test folders support nested `parent_id`.
- Test suites can be assigned to a folder.
- Test cases contain `steps` as JSONB and `tags` as `TEXT[]`.
- Full-text search via `search_tsv` `TSVECTOR` maintained by triggers on `title` and `description`.
- Versioning: every update creates a snapshot in `test_case_versions` and increments `test_cases.version`.
- Cursor-based pagination by `created_at DESC, id DESC`.

### Results (`results`)

- `TestRun` stores aggregate `total/passed/failed/skipped/blocked/duration_ms`.
- `TestRunItem` stores per-case status, duration, error message, stack trace, artifacts (JSONB), and `sort_order`.
- `Service.UpdateItemStatus` recalculates run aggregates and broadcasts `RunProgressEvent` to subscribers.
- `StreamRunProgress` is an SSE endpoint backed by an in-memory `progressHub` (map of `runID` to channel slices).

### Automation Hub (`automationhub`)

- Accepts JSON request `{workspace_id, project_id, suite_id?, name, format, payload}`.
- Supports `junit`, `playwright`, `cypress` (Playwright/Cypress share the same JSON parser).
- Parses JUnit XML or JSON report, creates a `TestRun` with status `running`, inserts `TestRunItem`s, then updates the run status to `passed`/`failed` and aggregates counts.

### API Keys (`apikeys`)

- Keys are generated as `tk_` prefix + 32 random bytes; the raw key is returned only at creation.
- Only SHA-256 `key_hash` and a 6-char `key_prefix` are stored.
- Scopes array, optional expiry, `revoked_at`, `last_used_at`.
- There is **no dedicated API-key authentication middleware** in the current route tree — the keys are stored but not used for request auth yet.

### Audit (`audit`)

- `audit_events` table stores `action`, `resource`, `resource_id`, `ip_address`, `metadata` JSONB, `created_at`.
- `AuditLog` middleware only logs when the response status is `< 400`.
- Calls are made in a background `context.Background()` from the audit log function in `server.go`, so audit writes are fire-and-forget.

### Idempotency

- Table `idempotency_records` keyed by `(workspace_id, operation, key)`.
- Middleware computes SHA-256 of `Idempotency-Key` header and a fingerprint of the JSON body (compacted, then hashed).
- Replays stored response if key + body match; returns `409` if key reused with different body.
- Currently only applied to `POST /ingest`.

---

## 10. Shared Utilities

| Package | Path | Purpose |
|--------|------|---------|
| `config` | `internal/shared/config/config.go` | Env-based config with defaults |
| `db` | `internal/shared/db/db.go` | `DBTX`/`BeginTxer` abstraction, context-aware tx/conn switching |
| `errors` | `internal/shared/errors/errors.go` | Sentinel errors |
| `http` | `internal/shared/http/response.go` | `JSON`/`ErrorJSON` envelope helpers |
| `jwt` | `internal/shared/jwt/jwt.go` | JWT sign/parse with `golang-jwt/jwt/v5` |
| `pagination` | `internal/shared/pagination/pagination.go` | Cursor parse/encode, `Params`/`Meta` |
| `password` | `internal/shared/password/password.go` | `bcrypt` hash/verify |
| `validation` | `internal/shared/validation/validation.go` | Email, name, slug helpers |
| `tenant` | `internal/shared/tenant/resolver.go` | Org resolution from workspace/project/key/run-item and membership check |
| `idempotency` | `internal/shared/idempotency/store.go` | Postgres-backed idempotency store |

---

## 11. Security & Correctness Observations

1. **API keys are not used for authentication.** `api_keys` is CRUD-only; no middleware validates a `X-API-Key` header. CI ingestion currently requires a JWT and `runs:ingest` permission.
2. **Rate limiting is unconfigured.** `LocalRateLimiter` is instantiated but never attached to routes; `rateLimitCfg` is assigned to `_`.
3. **SSE auth uses a query parameter for browsers.** `GET /test-runs/{id}/stream` is protected by `RequirePermission(runs:read)`. Browser `EventSource` connections pass the JWT in the `access_token` query parameter (the `Auth` middleware reads the header or the query param). This is an acceptable MVP workaround but should be hardened before exposing the stream over untrusted networks.
4. **`/organizations` POST/GET bypass tenant context and permission checks.** `POST /organizations` only requires a valid JWT. This is intentional for bootstrapping but may need guards if multi-tenancy rules change.
5. **Permission scope is organization-only.** `RequirePermission` always queries with `scope_type = 'organization'` and `scope_id = tenantID`. Fine-grained workspace/project RBAC is not yet implemented.
6. **Audit logging is fire-and-forget.** The log function in `server.go` calls `auditSvc.Log(context.Background(), ...)` with no error handling or retry. This is generally acceptable but can lose audit events under pressure.
7. **`RequestPasswordReset` never returns the token to the caller; the service emails it.** This is correct, but if SMTP is not configured the token is silently created and unusable (no fallback delivery mechanism).
8. **Idempotency middleware only looks for `workspace_id` in the JSON body.** Operations that resolve the workspace via query or path cannot use this middleware as-is.
9. **`RateLimitByEmail` ignores the `field` parameter and returns the IP-based key.** This is a bug if email-based rate limiting is intended.
10. **Refresh token rotation revokes the current token and issues a new one**, but the call `s.repo.RevokeRefreshToken(ctx, stored.ID, uuid.Nil)` stores `replaced_by` as `uuid.Nil` (`NULL`), which is harmless but loses the replacement chain.
11. **No request validation middleware** beyond per-handler JSON decoding and basic service checks. A shared request validation layer could be added later.
12. **Migration `000006` seeds `testcases:*`/`runs:execute`/`defects:*` permissions that are later duplicated/superseded by migrations `000008`, `000013`, `000016`.** Migration `000013` uses `tests:*` permissions and `000016` uses `runs:*`; the original `testcases:*`, `runs:execute`, `results:*` names may be stale. The route code uses `tests:*` and `runs:*`, so the older permission names are not referenced.

---

## 12. Recommended Next Steps

- Wire `RateLimit` middleware to sensitive unauthenticated endpoints (`/auth/login`, `/auth/register`, `/auth/password-reset/request`).
- Implement an `APIKeyAuth` middleware that looks up `api_keys.key_hash` from an `Authorization: ApiKey <key>` or `X-API-Key` header and injects a user/workspace context.
- Add permission gates to `POST /organizations` and `GET /organizations` if org creation should be restricted or filtered by membership.
- Remove or reconcile the unused permission names seeded in `000006` (`testcases:*`, `runs:execute`, `results:*`, `defects:*`, `settings:*`, `members:*`) with the route-level permissions (`tests:*`, `runs:*`, `orgs:*`, `workspaces:*`, `projects:*`, `apikeys:*`).
- Consider a retry queue or synchronous audit write for critical events.
- Expand `RequirePermission` to support workspace/project scope if fine-grained access is needed.
