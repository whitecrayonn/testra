# Functional Audit & Priority Matrix

This audit maps the current state of the `testra` monorepo after reading the
engineering docs, backend (Go) and frontend (Next.js) code, OpenAPI spec,
migrations and existing tests.

## 1. High-level dependency graph

```text
Web (Next.js)
  │
  ├─ /auth/*                  → identity API (JWT)
  ├─ /dashboard               → dashboard + workspace context (localStorage)
  ├─ /dashboard/projects      → project API
  ├─ /dashboard/test-cases/*  → testmanagement API
  ├─ /dashboard/test-runs/*   → results API + SSE stream
  ├─ /dashboard/settings/*    → platform API (api-keys, profile)
  └─ (auth)/onboarding         → org + workspace creation

API (Go / Chi)
  │
  ├─ shared/
  │   ├─ middleware/auth       JWT-only auth
  │   ├─ middleware/rbac       permission loader from role_assignments
  │   ├─ middleware/tenant     org resolution + app.tenant_id RLS
  │   ├─ http/response         {data,meta,error} envelope
  │   └─ db/tenant             tenant context helpers
  │
  ├─ identity/                 register/login/MFA/password reset
  ├─ organization/             org CRUD
  ├─ workspace/                workspace CRUD
  ├─ project/                  project CRUD
  ├─ apikeys/                  key CRUD + validation (but no auth middleware)
  ├─ testmanagement/           folders/suites/cases with versioning
  ├─ results/                  runs, items, SSE progress hub
  ├─ automationhub/            JUnit / Playwright / Cypress ingestion
  └─ audit/                    audit-log service (only write paths wired)

Postgres
  ├─ RLS policies using app.tenant_id
  ├─ migrations for auth, RBAC, orgs, workspaces, projects,
  │   test management, test runs, api keys, idempotency
  └─ analytics/integration/billing/intelligence modules are placeholders
```

## 2. Feature status matrix

| Feature / Module | Backend | Frontend | Migrations / Tests | OpenAPI | Status |
|---|---|---|---|---|---|
| **Identity (register/login/me)** | `identity` handler/service/repo exist, JWT signing | Login page, localStorage token | migrations + unit tests | Documented | **Functional** |
| **MFA setup/verify/disable** | `identity` service methods present | `mfa-setup` page renders QR code as `<img>` | migrations, basic tests | Documented | **Functional** |
| **Organizations** | CRUD, ownership on create | Onboarding creates org | migrations, integration tests | Documented | **Functional** |
| **Workspaces** | CRUD | Onboarding creates workspace | migrations | Documented | **Functional** |
| **Projects** | CRUD, key validation aligned with UI | Project list/create page | migrations | Documented | **Functional** |
| **API Keys (CRUD)** | Create/list/revoke + hash storage | Settings page creates/lists/revokes | migrations, unit tests | Documented | **Functional** |
| **API Key Authentication** | `apikeys.Service.Validate` exists but **no middleware consumes `X-API-Key`** | N/A (server-to-server) | None directly | Missing `apiKeyAuth` security scheme | **P0 Broken** |
| **Test Management** | Folders/suites/cases + versioning + search | List/create pages, new-case page | migrations | Documented | **Functional** |
| **Test Runs** | Create/get/list/delete, status update, item update | List/detail pages, start run | migrations + integration tests | Documented | **Functional** |
| **Live Test Run Updates (SSE)** | `StreamRunProgress` + `progressHub` exist | Detail page opens `EventSource` with `access_token` query param | backend unit test exists | Documented | **Functional (MVP query-token auth)** |
| **Automation Hub / Ingest** | JUnit, Playwright, Cypress ingestion | No UI; intended for CI | integration tests | Documented | **Functional (with JWT only)** |
| **RBAC** | `permissions`, `role_permissions`, `role_assignments` seeded; middleware loads perms | Not exposed | migrations | Documented | **Functional** |
| **Tenant Isolation** | `TenantContext` sets `app.tenant_id`, resolver maps resource→org | N/A | integration tests | N/A | **Functional** |
| **Rate Limiting** | `LocalRateLimiter` implemented, `Retry-After` uses wrong variable; not wired to routes | N/A | None | N/A | **P0 Broken – unconfigured** |
| **Idempotency** | Middleware reads body, stores replay | N/A | integration tests | Documented | **Functional** |
| **Audit Log** | Service exists, wired into route middleware for writes | No UI | migrations | N/A | **Partial** |
| **Token Refresh** | `RefreshToken` handler exists | Frontend stores only `testra_token`; no refresh flow | None | Documented | **P1 Missing** |
| **Analytics / Defects / Billing / API Testing / Integration Hub / Intelligence / Notification** | Placeholder `module.go` files only | Placeholder pages (Defects, Members) | None | Not present | **P2 Not built** |

## 3. Broken / P0 issues

These issues still block production or violate the intended design as of the latest review.

1. **API keys are not usable for ingestion.**
   - `apikeys.Service.Validate` is implemented but no middleware validates `X-API-Key`.
   - `POST /api/v1/ingest` only accepts a user JWT, which CI systems do not have.
   - This breaks the primary external integration entry point.

2. **Rate limiting is unconfigured.**
   - `rateLimitCfg` is created in `server.New` but never attached to routes.
   - `ratelimit.go` has a bug where `Retry-After` is set to `remaining` instead of `retryAfter`.

3. **Token refresh not implemented on the frontend.**
   - After JWT expiry the user is silently logged out / API calls fail.

## 3.5. Resolved in Phase 3.5

The following items were previously listed as P0 but have been fixed:

- **Live Test Run Updates (SSE).**
  - `Auth` middleware now extracts the JWT from `Authorization: Bearer` or the `access_token` query parameter.
  - The frontend `EventSource` passes the token in `GET /api/v1/test-runs/{id}/stream?access_token=${token}`.
  - See `docs/handover/live-test-run-updates.md` for the full design and verification steps.

- **Project key validation mismatch.**
  - Frontend project key generation now matches the backend regex `^[A-Z][A-Z0-9]{1,9}$`.

- **MFA QR code display.**
  - The `mfa-setup` page renders the backend `qr_code` data URL in an `<img>` tag.

- **Onboarding slug fields.**
  - Onboarding now sends explicit `slug` values for organization and workspace creation.

## 4. Priority list

### P0 – Must fix (broken or user-objective)

1. **API key authentication for `/ingest`** – required for the CI/automation value proposition.
2. **Rate limiting wiring** – security/stability.
3. **Token refresh on the frontend** – users are logged out every 15 minutes.

### P1 – Important gaps

- Audit log UI / read endpoints.
- Member invitation / role management UI.
- Automation Hub upload UI in the dashboard.
- Frontend state layer beyond `localStorage` (see `frontend-audit.md` §10.6).

### P2 – Future phases

- Analytics, Defects, Billing, API Testing, Integration Hub, Intelligence. Notifications module is implemented in Phase 3.

## 5. Recommended next action

Implement **API key authentication for the `/ingest` endpoint** and **rate limiting wiring**:

- Build or wire `APIKeyAuth` middleware to validate `X-API-Key` / `Authorization: ApiKey <key>` against `api_keys.key_hash`, scope, and expiry.
- Set the tenant context from the API key's workspace for `/ingest` and other automation routes.
- Add/update backend tests and update `docs/api/openapi/openapi.yaml` with the `apiKeyAuth` security scheme.
- Wire `LocalRateLimiter` to `/auth/login`, `/auth/register`, `/auth/password-reset/request`, and `/ingest`.
- Fix `Retry-After` in `shared/middleware/ratelimit.go`.
