# Technical Debt

> This document merges every known issue from the handover audits into a single master register. For module-specific findings and recommendations, see [`backend-audit.md`](backend-audit.md), [`frontend-audit.md`](frontend-audit.md), and [`infra-audit.md`](infra-audit.md).

## Severity legend

| Severity | Meaning |
|----------|---------|
| **Critical** | Blocks production launch or exposes data/customers to serious risk |
| **High** | Significant functional or security gap; must fix before public beta |
| **Medium** | Real impact but can be deferred until critical/high items are resolved |
| **Low** | Polish, cleanup, or future optimization |

---

## Critical

| # | Description | Impact | Affected modules | Recommended fix | Priority | Current status |
|---|-------------|--------|------------------|-----------------|----------|----------------|
| C1 | **No rate limiting on auth endpoints.** `LocalRateLimiter` is created but never wired to routes. | Account takeover via brute force; DoS risk | `identity`, `shared/middleware` | Wire `LocalRateLimiter` to `/auth/*` and ingest routes; prefer Redis-backed rate limiter for production | P0 | Open |
| C2 | **No API-key authentication middleware.** CI/CD ingestion requires a user JWT. | Scoped keys are unused, CI pipelines must embed user tokens, keys leaked to CI logs cannot be revoked without disabling the user | `apikeys`, `automationhub`, `shared/middleware` | Build `APIKeyAuth` middleware that looks up `api_keys.key_hash`, validates scopes, and sets tenant context | P0 | Open |
| C3 | **SSE progress stream is inaccessible from browsers.** `GET /test-runs/{id}/stream` requires `Authorization: Bearer` but `EventSource` cannot send custom headers. | Manual test execution live progress does not work in the browser | `results`, `frontend` | Move auth to session cookie, short-lived signed query token, or replace `EventSource` with `fetch` + `ReadableStream` | P0 | Open |
| C4 | **Default/weak secrets in `.env.example` and no secret management.** Example contains `JWT_SECRET=testratestra` and `DATABASE_URL` with hard-coded password. | Production deployments will use known credentials | `infra`, `config` | Generate secrets in CI/CD or secrets manager; remove defaults; fail fast if secrets are unchanged | P0 | Open |
| C5 | **No client-side route guards and token stored in `localStorage`.** Unauthenticated users can navigate to `/dashboard/*` and tokens are exposed to XSS. | Data leakage, UX flash, security exposure | `frontend` | Add server-side or client middleware that validates token before rendering; move to `httpOnly` cookie or secure wrapper | P0 | Open |
| C6 | **Audit logging uses `context.Background()` and no retry/queue.** Events are fired asynchronously with no guarantee of persistence. | Compliance evidence can be silently lost | `audit`, `shared/middleware` | Write audit events synchronously inside the request transaction or enqueue to a durable queue (Redis/Asynq) | P0 | Open |

---

## High

| # | Description | Impact | Affected modules | Recommended fix | Priority | Current status |
|---|-------------|--------|------------------|-----------------|----------|----------------|
| H1 | **Permission-name drift between RBAC seed and middleware.** Middleware checks `orgs:read` but seed names include `organization:read`; `members:manage` vs `organization_members:manage`. | Authorization checks silently fail or succeed incorrectly | `rbac`, `shared/middleware`, all modules | Create a single source-of-truth `Permission` constants package; align migration seeds and middleware usage | P1 | Open |
| H2 | **`POST /organizations` and `GET /organizations` lack permission/membership gates.** | Any authenticated user can create or list organizations without restriction | `organization`, `shared/middleware` | Add `RequirePermission("orgs:create")` and filter `GET` by `organization_members.user_id` | P1 | Open |
| H3 | **Tenant resolution can pick the wrong tenant.** `TenantContext` checks `organization_id`, `workspace_id`, `project_id`, `api_key_id`, `run_id` in a fixed order from body/query; a malicious request can supply a valid member org ID plus a different `workspace_id` belonging to another org. | Cross-tenant data access if other checks are skipped | `shared/middleware`, `shared/tenant` | Resolve tenant from the **least-privileged, most-specific** resource and verify membership before every operation | P1 | Open |
| H4 | **Project key generation mismatch.** Frontend uses `name.toUpperCase().replace(/\s+/g, "-").slice(0, 10)`; backend regex `^[A-Z][A-Z0-9]{1,9}$` rejects hyphens. | Users cannot create projects with auto-generated keys | `project`, `frontend` | Align generator to backend regex and add unique constraint/index on `(organization_id, key)` | P1 | Open |
| H5 | **No token refresh on 401 in the frontend.** Access tokens expire after 15 minutes. | Users are logged out every 15 minutes | `frontend`, `identity` | Add a fetch interceptor that calls `/auth/refresh` on 401 and retries the original request once | P1 | Open |
| H6 | **MFA QR code rendered as text.** Backend returns data URL; frontend displays it in a `<p>` tag. | Users cannot scan QR code with authenticator apps | `frontend`, `identity` | Render `qr_code` in an `<img src={qr_code} />` tag | P1 | Open |
| H7 | **No Kubernetes ConfigMap, Secret, probes, or ingress definitions.** | Cannot deploy to production safely; no health checks or secret management | `infra/k8s`, `infra/terraform` | Add Kustomize/Helm charts, `ConfigMap`, `Secret`, liveness/readiness probes, `Ingress`, and resource limits | P1 | Open |
| H8 | **Terraform scaffold has no resources or modules.** | Infrastructure cannot be provisioned | `infra/terraform` | Implement VPC, EKS, RDS, S3, Redis, Route53, WAF modules | P1 | Open |
| H9 | **No integration or end-to-end test suite.** | Regression risk, low confidence in refactors | `all` | Add Go integration tests against test DB, Playwright E2E for critical flows | P1 | Open |
| H10 | **No observability stack.** No OpenTelemetry, Prometheus, Grafana, Loki, or structured request logging. | Cannot debug production incidents | `shared`, `infra` | Add OpenTelemetry traces, structured logs, and metrics endpoints | P1 | Open |

---

## Medium

| # | Description | Impact | Affected modules | Recommended fix | Priority | Current status |
|---|-------------|--------|------------------|-----------------|----------|----------------|
| M1 | **Duplicate dashboard route trees.** `/dashboard/*` and `/:workspace/*` re-export the same pages. | Confusing URLs, inconsistent context, broken analytics | `frontend` | Consolidate on `/:workspace` tree; redirect `/dashboard` to last selected workspace | P2 | Open |
| M2 | **Frontend relies entirely on `localStorage` for state.** | Hydration mismatches, no reactive cross-page state, repetitive `typeof window` guards | `frontend` | Introduce Zustand or React Context for auth/workspace state; keep `localStorage` only for persistence | P2 | Open |
| M3 | **OpenAPI spec is at version 0.4.0 and missing modules.** | Defects, analytics, billing, integrations still missing; notifications now documented | `docs/api/openapi` | Complete spec incrementally; generate SDK from it | P2 | Open |
| M4 | **`apps/worker` is a stub and `cmd/worker` does nothing.** | No background processing for email, reports, or ML jobs | `worker`, `apps/api/cmd/worker` | Implement Asynq worker or promote `cmd/worker` to real background service | P2 | Open |
| M5 | **Settings sub-pages are mostly placeholders.** | Cannot manage members, roles, API keys, audit logs, billing; notifications page implemented | `frontend` | Build settings pages using existing `platform/api.ts` and backend endpoints | P2 | Open |
| M6 | **Audit middleware does not capture request result/status before response.** | Audit metadata may be incomplete if handler errors | `audit`, `shared/middleware` | Use `responseWriter` wrapper to capture status code in middleware | P2 | Open |
| M7 | **`mailpit` has no Docker Compose healthcheck.** | `depends_on` ordering is weaker without health status | `infra/docker` | Add `healthcheck` block to `mailpit` service | P2 | Open |
| M8 | **Docker Compose does not define application services.** | Developers must run `api`, `web`, `worker`, `ml` manually | `infra/docker` | Add optional `api`, `web`, `worker`, `ml` services for container-first workflows | P2 | Open |
| M9 | **No bulk import / migration tools from TestRail/Excel/CSV.** | High friction for customer onboarding | `testmanagement`, `frontend` | Build import endpoints and UI wizards | P3 | Open |
| M10 | **Environment variable name mismatch.** `apps/api/.env.example` has `JWT_EXPIRY_HOURS=168` while code reads `JWT_EXPIRY_MINUTES`. | Config silently falls to default 15 minutes or fails to parse | `config`, `infra` | Fix `.env.example` and add startup validation for required vars | P2 | Open |

---

## Low

| # | Description | Impact | Affected modules | Recommended fix | Priority | Current status |
|---|-------------|--------|------------------|-----------------|----------|----------------|
| L1 | **Frontend `apiFetch` does not retry or handle network errors gracefully.** | Transient errors show raw error text | `frontend` | Add exponential backoff for idempotent GETs and user-friendly error toasts | P4 | Open |
| L2 | **Project `key` field is limited to 10 uppercase characters.** | Some teams may want longer/more readable keys | `project` | Relax or configure key rules | P4 | Open |
| L3 | **Removed `.dockerignore` in `apps/api`** | Larger build context, slower builds | `apps/api` | Restore `.dockerignore` | P4 | Open |
| L4 | **No global error boundary or loading state.** | Each page duplicates loader/error handling | `frontend` | Add `error.tsx` and `loading.tsx` in route groups and feature wrappers | P4 | Open |
| L5 | **`localStorage` keys are not namespaced.** | Risk of collision with other apps or future keys | `frontend` | Prefix with `testra_` consistently (already partially true) | P5 | Open |
| L6 | **Stale or sample files remain in `.github/workflows` and `infra` directories.** | Confusion for new engineers | `infra`, `.github` | Remove sample placeholders or document them | P5 | Open |

---

## Consolidated priority view

| Priority | What to fix |
|----------|--------------|
| **P0 — Critical** | Rate limiting, API-key auth, SSE auth, secrets management, frontend route guards, audit durability |
| **P1 — High** | Permission-name alignment, org permission gates, tenant resolver, token refresh, project key fix, K8s/Terraform completion, tests, observability |
| **P2 — Medium** | Route consolidation, global state, OpenAPI completion, worker implementation, settings pages, audit middleware, env vars, Docker app services |
| **P3+ — Lower** | Bulk import, advanced UX polish, error boundaries, key namespaces, cleanup |

For the exact implementation order, see [`NEXT_STEPS.md`](NEXT_STEPS.md).
