# Routes

**Purpose:** Inventory all public-facing frontend and backend routes with dynamic segments and known caveats.
**Owner:** Engineering Lead
**Scope:** Frontend and backend route inventory, workspace routing, dynamic segments, caveats.
**Source of Truth:** ROUTES.md for route inventory; `apps/api/internal/shared/server/server.go` for actual route wiring.
**Last Updated:** July 2026
**Related documents:**
- [`BIBLICAL_TESTRA.md`](BIBLICAL_TESTRA.md)
- [`API_DESIGN_GUIDELINES.md`](api/API_DESIGN_GUIDELINES.md)
- [`docs/api/openapi/openapi.yaml`](api/openapi/openapi.yaml)

> This document lists all public-facing routes in the Testra platform. For backend middleware and permission details, see [`BIBLICAL_TESTRA.md`](BIBLICAL_TESTRA.md). For frontend page analysis and workspace routing strategy, see [`PROJECT_OVERVIEW.md`](PROJECT_OVERVIEW.md) and [`ONBOARDING.md`](engineering/ONBOARDING.md).

## Frontend routes

All frontend routes are served by the Next.js 15 App Router in `apps/web/app/`.

### Auth group — `app/(auth)/`

Route group `(auth)` applies `app/(auth)/layout.tsx` (centered card layout). URLs do **not** include the group name.

| Route | File | Purpose |
|-------|------|---------|
| `/` | `apps/web/app/page.tsx` | Redirects to `/login` |
| `/login` | `apps/web/app/(auth)/login/page.tsx` | Email/password login, optional MFA code |
| `/register` | `apps/web/app/(auth)/register/page.tsx` | Account creation |
| `/forgot-password` | `apps/web/app/(auth)/forgot-password/page.tsx` | Request password reset email |
| `/reset-password` | `apps/web/app/(auth)/reset-password/page.tsx` | Confirm reset with token and new password |
| `/mfa-setup` | `apps/web/app/(auth)/mfa-setup/page.tsx` | TOTP enrollment and verification |
| `/onboarding` | `apps/web/app/(auth)/onboarding/page.tsx` | Create first organization + workspace |

### Dashboard group — `app/(dashboard)/`

Route group `(dashboard)` applies `app/(dashboard)/layout.tsx` (sidebar + main content area). The group name is not part of the URL.

#### Static `/dashboard` routes

| Route | File | Purpose |
|-------|------|---------|
| `/dashboard` | `apps/web/app/(dashboard)/dashboard/page.tsx` | Landing with quick links, workspace/project context |
| `/dashboard/settings` | `apps/web/app/(dashboard)/dashboard/settings/page.tsx` | Settings overview / shell |
| `/dashboard/settings/members` | `apps/web/app/(dashboard)/dashboard/settings/members/page.tsx` | Placeholder |
| `/dashboard/settings/roles` | `apps/web/app/(dashboard)/dashboard/settings/roles/page.tsx` | Placeholder |
| `/dashboard/settings/api-keys` | `apps/web/app/(dashboard)/dashboard/settings/api-keys/page.tsx` | Create and revoke API keys |
| `/dashboard/settings/audit-logs` | `apps/web/app/(dashboard)/dashboard/settings/audit-logs/page.tsx` | Placeholder |
| `/dashboard/settings/billing` | `apps/web/app/(dashboard)/dashboard/settings/billing/page.tsx` | Placeholder |
| `/dashboard/settings/notifications` | `apps/web/app/(dashboard)/dashboard/settings/notifications/page.tsx` | Notification preferences and channels |
| `/dashboard/notifications` | `apps/web/app/(dashboard)/dashboard/notifications/page.tsx` | In-app notification list, mark read/delete |
| `/dashboard/settings/profile` | `apps/web/app/(dashboard)/dashboard/settings/profile/page.tsx` | Placeholder |
| `/dashboard/settings/security` | `apps/web/app/(dashboard)/dashboard/settings/security/page.tsx` | Placeholder |
| `/dashboard/settings/organization` | `apps/web/app/(dashboard)/dashboard/settings/organization/page.tsx` | Placeholder |
| `/dashboard/settings/workspace` | `apps/web/app/(dashboard)/dashboard/settings/workspace/page.tsx` | Placeholder |

#### Dynamic workspace routes — `app/(dashboard)/[workspace]/`

`[workspace]` is a dynamic segment. The route is intended to be the workspace slug, but the frontend currently stores the workspace UUID in `localStorage` and the `/dashboard/*` tree re-exports the same pages from `[workspace]/...`.

| Route | File | Purpose |
|-------|------|---------|
| `/[workspace]` | `apps/web/app/(dashboard)/[workspace]/page.tsx` | Stores `workspaceId` param in `localStorage` |
| `/[workspace]/projects` | `apps/web/app/(dashboard)/[workspace]/projects/page.tsx` | List/select/create projects |
| `/[workspace]/test-cases` | `apps/web/app/(dashboard)/[workspace]/test-cases/page.tsx` | Search/list test cases |
| `/[workspace]/test-cases/new` | `apps/web/app/(dashboard)/[workspace]/test-cases/new/page.tsx` | Create test case |
| `/[workspace]/test-cases/[id]` | `apps/web/app/(dashboard)/[workspace]/test-cases/[id]/page.tsx` | Edit/view test case + version history |
| `/[workspace]/test-runs` | `apps/web/app/(dashboard)/[workspace]/test-runs/page.tsx` | List test runs |
| `/[workspace]/test-runs/new` | `apps/web/app/(dashboard)/[workspace]/test-runs/new/page.tsx` | Create manual test run |
| `/[workspace]/test-runs/[id]` | `apps/web/app/(dashboard)/[workspace]/test-runs/[id]/page.tsx` | Run detail + SSE progress |
| `/[workspace]/defects` | `apps/web/app/(dashboard)/[workspace]/defects/page.tsx` | Placeholder for Phase 4 |

### Workspace context persistence

The selected workspace/project IDs are stored in `localStorage` under:

| Key | Purpose |
|-----|---------|
| `testra_organization_id` | Selected organization UUID |
| `testra_workspace_id` | Selected workspace UUID |
| `testra_project_id` | Selected project UUID |
| `testra_project_name` | Selected project display name |

---

## Backend routes

All backend routes are mounted under `/api/v1` in `apps/api/internal/shared/server/server.go`.

### Health

```
GET /health
```

### Auth (no tenant context)

```
POST   /auth/register
POST   /auth/login
POST   /auth/refresh
POST   /auth/password-reset/request
POST   /auth/password-reset/confirm
GET    /auth/me          (Auth middleware)
POST   /auth/mfa/setup   (Auth middleware)
POST   /auth/mfa/verify  (Auth middleware)
POST   /auth/mfa/disable (Auth middleware)
```

### Organizations

```
POST   /organizations                         (Auth only, no tenant/permission gate)
GET    /organizations                         (Auth only)
GET    /organizations/{id}                      (TenantContext + orgs:read)
```

### Workspaces

```
POST   /workspaces                              (TenantContext via body org_id + workspaces:create)
GET    /workspaces?organization_id=...          (TenantContext via query org_id + workspaces:read)
GET    /workspaces/{id}                         (TenantContext via workspace_id + workspaces:read)
```

### Projects

```
POST   /projects                                (TenantContext via body workspace_id + projects:create)
GET    /projects?workspace_id=...               (TenantContext via query workspace_id + projects:read)
GET    /projects/{id}                           (TenantContext via project_id + projects:read)
```

### API Keys

```
POST   /api-keys                                (TenantContext via body workspace_id + apikeys:create + AuditLog)
GET    /api-keys?workspace_id=...               (TenantContext via query workspace_id + apikeys:read)
DELETE /api-keys/{id}                           (TenantContext via api_key_id + apikeys:delete + AuditLog)
```

### Test Management

```
POST   /test-folders   (tests:create + AuditLog)
POST   /test-suites    (tests:create + AuditLog)
POST   /test-cases     (tests:create + AuditLog)

GET    /test-folders?workspace_id=...
GET    /test-suites?workspace_id=...
GET    /test-cases/search?workspace_id=...    (full-text search)
GET    /test-cases?project_id=...
GET    /test-cases/{id}
GET    /test-cases/{id}/versions
PUT    /test-cases/{id}                         (tests:update + AuditLog)
DELETE /test-cases/{id}                         (tests:delete + AuditLog)

GET    /test-folders/{id}
GET    /test-suites/{id}
PUT    /test-folders/{id}                       (tests:update + AuditLog)
PUT    /test-suites/{id}                        (tests:update + AuditLog)
DELETE /test-folders/{id}                       (tests:delete + AuditLog)
DELETE /test-suites/{id}                        (tests:delete + AuditLog)
```

### Test Runs, Results & Ingestion

```
POST   /test-runs                              (runs:create + AuditLog)
GET    /test-runs?project_id=...               (runs:read)
GET    /test-runs/{id}                         (runs:read)
GET    /test-runs/{id}/items                   (runs:read)
GET    /test-runs/{id}/stream                  (runs:read, SSE)
PUT    /test-runs/{id}                         (runs:update + AuditLog)
DELETE /test-runs/{id}                         (runs:delete + AuditLog)
PUT    /test-run-items/{id}                    (runs:update + AuditLog)
POST   /ingest                                 (runs:ingest + AuditLog + IdempotencyKey)
```

---

## Dynamic routes

### Frontend dynamic segments

| Segment | Example URL | Resolver | Notes |
|---------|-------------|----------|-------|
| `[workspace]` | `/my-workspace/projects` | Slug or UUID from `localStorage` | The `[workspace]` tree re-exports `/dashboard/*` pages |
| `[id]` (test cases) | `/dashboard/test-cases/abc-123` | Test case UUID | Edit/view page |
| `[id]` (test runs) | `/dashboard/test-runs/abc-123` | Test run UUID | Detail + SSE page |

### Backend dynamic segments

| Segment | Example URL | Resolution | Notes |
|---------|-------------|------------|-------|
| `{id}` (organizations) | `/organizations/{id}` | `OrgIDFromURLParam` | Tenant context set to org ID |
| `{id}` (workspaces) | `/workspaces/{id}` | `WorkspaceToOrg` | Resolves workspace → organization |
| `{id}` (projects) | `/projects/{id}` | `ProjectToOrg` | Resolves project → workspace → organization |
| `{id}` (api-keys) | `/api-keys/{id}` | `APIKeyToOrg` | Resolves API key → workspace → organization |
| `{id}` (test cases) | `/test-cases/{id}` | Resolves via workspace/project | Tenant set by workspace context |
| `{id}` (test runs) | `/test-runs/{id}` | Resolves via project/workspace | SSE uses same route group |
| `{id}` (run items) | `/test-run-items/{id}` | `RunItemToOrg` | Resolves run item → run → workspace → organization |

---

## Dashboard routing strategy

1. **File-based routing:** Next.js App Router auto-generates URLs from `app/` directory structure.
2. **Route groups:** `(auth)` and `(dashboard)` wrap shared layouts without affecting the URL path.
3. **Context stored in `localStorage`:** The app does not use URL params for workspace/project selection on every page; instead it reads `localStorage`.
4. **Re-export pattern:** `/dashboard/*` pages are thin re-exports of `[workspace]/*` pages so the same UI is available under two URL schemes.

---

## Known routing caveats

| Caveat | Impact | Recommended fix |
|--------|--------|---------------|
| **Two dashboard route trees** | `/dashboard/projects` and `/:workspace/projects` serve the same page but store context differently. Confuses users and analytics. | Consolidate to one route tree; use `/:workspace` and redirect `/dashboard` to the selected workspace. |
| **No route guards** | Unauthenticated users can navigate to `/dashboard/*` and see empty/skeleton UI until API 401s. | Add client middleware or server-side auth check; redirect to `/login` when no token. |
| **`[workspace]` not validated** | Any slug in the URL is accepted; the page reads workspace UUID from `localStorage` and may ignore the URL. | Use the URL param as source of truth and validate it against the backend. |
| **No catch-all 404** | Unknown dashboard paths may fall through to Next.js default behavior. | Add a `not-found.tsx` in the dashboard group. |
| **Backend route permissions are inconsistent** | `POST /organizations` and `GET /organizations` only require a valid JWT, not `orgs:create` or membership. | Add `RequirePermission` gates if org creation should be restricted or list filtered by membership. |
| **SSE auth via query parameter (MVP)** | `GET /test-runs/{id}/stream` is behind `RequirePermission`; the `Auth` middleware reads the JWT from `Authorization: Bearer` or the `access_token` query parameter so `EventSource` can authenticate. This is acceptable for local/MVP but should be hardened before public networks. | Consider a short-lived signed SSE token or cookie-based session before exposing the stream over untrusted networks. |

---

## See Also

- [`BIBLICAL_TESTRA.md`](BIBLICAL_TESTRA.md) — engineering handbook and canonical sources map.
- [`API_DESIGN_GUIDELINES.md`](api/API_DESIGN_GUIDELINES.md) — REST, versioning, and idempotency conventions.
- [`docs/api/openapi/openapi.yaml`](api/openapi/openapi.yaml) — the authoritative HTTP contract.
- [`SYSTEM_FLOWS.md`](architecture/SYSTEM_FLOWS.md) — request lifecycle and sequence diagrams.
- [`ONBOARDING.md`](engineering/ONBOARDING.md) — frontend structure, development workflow, and common pitfalls.
