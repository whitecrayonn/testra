# Current State

> This document describes the repository **exactly as it exists today**. It is the handover snapshot a new engineer reads before writing any code.

## Maturity summary

- **Phase:** Pre-MVP hardening / engineering handover.
- **Backend:** Core platform and test-management flows are functional locally.
- **Frontend:** Auth, onboarding, dashboard skeleton, test cases, and test runs are partially implemented.
- **Infrastructure:** Scaffolding exists; not production-ready.
- **Production readiness:** **Not ready** for customer-facing deployment.

---

## Completed modules

These modules have working backend code, database migrations, and at least partial frontend support.

| Module | Backend | Frontend | Notes |
|--------|---------|----------|-------|
| **Identity** | ✅ | ✅ | Register, login, JWT, refresh tokens, MFA TOTP, password reset |
| **Organization** | ✅ | ✅ | Create/list organizations; onboarding flow |
| **Workspace** | ✅ | ✅ | Create/list workspaces within an organization |
| **Project** | ✅ | ✅ | Create/list projects; project key validation |
| **Test Management** | ✅ | ✅ | Folders, suites, test cases, versions, full-text search; frontend case list/create/edit/versions |
| **Test Runs / Results** | ✅ | ✅ | Manual runs, run items, SSE progress stream (query-token auth); UI list/detail/start |
| **Automation Hub** | ✅ | ❌ | Ingest JUnit/Playwright/Cypress payloads into runs; no UI |
| **API Keys** | ✅ | ✅ | Backend CRUD and `/dashboard/settings/api-keys` UI for create/revoke; not yet used for `/ingest` auth |
| **RBAC** | ✅ (seed/loader) | ❌ | Roles/permissions seeded; org-scoped enforcement only |
| **Audit** | ✅ | ❌ | Fire-and-forget audit events; no UI |
| **Notifications** | ✅ | ✅ | In-app feed, read/unread, preferences, email/Slack/Teams/webhook channels, RLS |
| **OpenAPI Spec** | 📝 | N/A | Covers auth, orgs, workspaces, projects, test mgmt, runs, ingest, notifications |

---

## Partially completed modules

These have scaffolding or partial implementation but are not shippable.

| Module | What works | What is missing / broken |
|--------|------------|--------------------------|
| **Dashboard** | Landing page, sidebar, workspace/project context from `localStorage` | Real widgets, role-based views, recent activity feed, intelligence teasers |
| **Settings** | Settings shell and navigation | Most sub-pages are still placeholders; **notifications** and **API keys** settings pages are implemented |
| **OpenAPI** | Core routes documented | Defects, analytics, billing, integrations, webhooks still missing; notifications documented |
| **Test Suites / Folders UI** | Backend list endpoints and API wrappers | No user interface to create/manage suites or folders |
| **API Keys UI** | API wrappers in `features/platform/api.ts` | Settings page at `/dashboard/settings/api-keys` implemented |
| **Test Run SSE** | Backend stream endpoint exists | `Auth` middleware accepts `Authorization: Bearer` or `access_token` query parameter; works in browsers with query token |
| **MFA Setup UI** | Form and TOTP code input | QR code rendered as an `<img>` from the backend data URL |
| **Project creation UI** | Inline create | Auto-generated project key matches backend `^[A-Z][A-Z0-9]{1,9}$` regex |
| **Onboarding slug** | Creates org + workspace | Sends explicit `slug` for organization and workspace creation |
| **CI/CD** | GitHub Actions lint/build | No image push, no deployment, no integration tests |
| **Kubernetes** | Base deployment + service manifests | No ConfigMap/Secret, probes, resources, ingress, web/worker/migrator |
| **Terraform** | Provider/backend scaffold | No modules or resources defined |

---

## Placeholder pages

The frontend contains pages or directories that render empty cards or placeholder text.

| Route / Directory | File | Note |
|-------------------|------|------|
| `/dashboard/defects` | `[workspace]/defects/page.tsx` | `PlaceholderPage` with "Planned for Phase 4" |
| `/dashboard/settings/*` | `dashboard/settings/*/page.tsx` | API keys and notifications pages implemented; most other tabs are placeholders |
| `/dashboard/api-tests` | `(dashboard)/[workspace]/api-tests/` or similar | Not implemented; directory may be empty |
| `apps/web/features/api-testing/` | `.gitkeep` or empty | No API testing frontend |
| `apps/web/features/analytics/` | Empty | No analytics frontend |
| `apps/web/features/defects/` | Empty | No defects frontend |

---

## Unimplemented modules

These modules exist as directories with only a `.gitkeep` or no code at all.

| Module | Backend directory | Frontend directory | Why it matters |
|--------|-------------------|--------------------|----------------|
| **Defects** | `apps/api/internal/defects/` | `apps/web/features/defects/` | Core MVP feature; required for issue tracking |
| **API Testing** | `apps/api/internal/apitesting/` | `apps/web/features/api-testing/` | Displaces Postman; key differentiator |
| **Billing / Subscriptions** | `apps/api/internal/billing/` | `apps/web/features/settings/` | Required for commercial launch |
| **Integration Hub** | `apps/api/internal/integrationhub/` | N/A | Jira, GitHub, GitLab, CI/CD webhooks |
| **Analytics** | `apps/api/internal/analytics/` | `apps/web/features/analytics/` | Reporting and dashboards beyond skeleton |
| **Intelligence** | `apps/api/internal/intelligence/` | `apps/web/features/analytics/` | Flaky detection, risk scoring, failure classification (V2) |
| **Marketplace** | N/A | N/A | V3 ecosystem play |
| **Public API / SDK** | N/A (uses `/api/v1`) | `packages/sdk/` | Expose stable API for customers/partners |
| **SSO / SAML** | N/A | N/A | Enterprise hard requirement |
| **Bulk import / migration** | N/A | N/A | TestRail/Excel/CSV migration |

---

## Broken features

These features compile and run but have known runtime defects that block production use.

| Feature | Problem | Evidence / location |
|---------|---------|---------------------|
| **SSE progress stream** | `EventSource` cannot send `Authorization: Bearer` header. The endpoint at `/api/v1/test-runs/{id}/stream` is protected by `RequirePermission("runs:read")`, so browsers cannot authenticate and the stream fails. | `frontend-audit.md` §8.6, `backend-audit.md` §4 |
| **MFA QR code display** | ~~Fixed~~: frontend now renders `qr_code` as an `<img src={qr_code} />`. | `frontend-audit.md` §8.1/§10 |
| **Project key generation** | ~~Fixed~~: frontend now generates uppercase alphanumeric keys (starting with a letter, 2–10 chars) matching the backend regex. | `frontend-audit.md` §8.4 |
| **Onboarding slug** | ~~Fixed~~: onboarding now explicitly sends `slug` for organization and workspace creation. | `frontend-audit.md` §8.3 |
| **Auth token refresh (client)** | Access tokens expire in 15 minutes; the frontend does not call `/auth/refresh` on 401. Users must re-login. | `frontend-audit.md` §10 |

---

## Known limitations

| Limitation | Impact | Where to fix |
|------------|--------|--------------|
| **No API-key authentication middleware** | CI/CD ingestion currently requires a user JWT; scoped API keys are stored but not used for auth. | `backend-audit.md` §11.1, `TECHNICAL_DEBT.md` |
| **Rate limiting is unconfigured** | `LocalRateLimiter` is instantiated but never wired to routes. Auth endpoints are vulnerable to brute force. | `backend-audit.md` §11.2 |
| **Audit logging is fire-and-forget** | `auditSvc.Log(context.Background(), ...)` runs in a background context with no retry; audit events can be lost. | `backend-audit.md` §11.5 |
| **Permission scope is organization-only** | Workspace/project-level roles are not supported; `RequirePermission` always uses `scope_type = 'organization'`. | `backend-audit.md` §11.4 |
| **No route guards in frontend** | Unauthenticated users can briefly see dashboard pages and only fail on API calls. | `frontend-audit.md` §10.7 |
| **No global state management** | `localStorage` is the only state layer; causes hydration mismatches and repetitive `typeof window` guards. | `frontend-audit.md` §5, §10.6 |
| **Duplicate dashboard route trees** | `/dashboard/*` and `/:workspace/*` both exist; same components served from different URLs with different context logic. | `frontend-audit.md` §3.2, §10.8 |
| **No error boundary / global loading** | Each page handles loading/error independently. | `frontend-audit.md` §10.10 |
| **No production-ready infrastructure** | Kubernetes/Terraform are scaffolds; no secrets, probes, ingress, or CD. | `infra-audit.md` |
| **Worker is a stub** | `apps/worker` and `apps/api/cmd/worker` are empty. No background processing exists. | `infra-audit.md` finding #12 |
| **ClickHouse / Redis / MinIO not used by app** | Provisioned in Docker Compose but no application code connects to them. | `infra-audit.md` |
| **Environment variable drift** | `apps/api/.env.example` has `JWT_EXPIRY_HOURS=168` while code reads `JWT_EXPIRY_MINUTES`. | `infra-audit.md` §6.3 |

---

## Open architecture decisions

| Decision | Options | Current lean | Why it matters |
|----------|---------|--------------|--------------|
| **Identity provider** | Self-hosted Keycloak/FusionAuth vs Clerk/WorkOS | Undecided; current code is self-built JWT + MFA | SSO/SAML/SCIM is a must-have for enterprise; vendor choice affects ops overhead |
| **Frontend state layer** | Keep `localStorage` vs add Zustand/React Context vs TanStack Query | Undecided; currently local state | Required for route guards, refresh-token flow, and hydration safety |
| **SSE authentication** | Session cookie vs query-param token vs fetch-based streaming | Undecided; currently broken with header auth | Must fix before production test-run progress |
| **API key auth header** | `Authorization: ApiKey <key>` vs `X-API-Key` | Undecided; currently no middleware | Required for CI/CD ingestion without user JWT |
| **ClickHouse adoption** | Start storing test results in ClickHouse now vs defer to V2 | Deferred; results live in Postgres | Analytics and intelligence depend on this |
| **Worker framework** | Asynq over Redis vs custom Go worker | Undecided; `apps/worker` is a stub | Background jobs needed for notifications, reports, ML inference |
| **Microservices extraction trigger** | Keep modular monolith vs extract ML service/ingestion pipeline first | Keep monolith for now | Decision needed once scale or team size justifies it |
| **Multi-region / data residency** | Schema-per-tenant vs row-level tenant with regional DBs | Row-level tenant in shared schema | Enterprise Edition requirement |
| **SDK generation** | Generate TypeScript SDK from OpenAPI vs hand-write | Undecided; `packages/sdk` exists | Public API and partner integrations depend on this |

---

## Current repository maturity

| Capability | Maturity | Evidence |
|------------|----------|----------|
| **Local development** | ✅ Functional | `pnpm dev`, `make migrate`, Docker Compose for deps |
| **Core backend flows** | ✅ Functional | Unit tests build; manual API calls work |
| **Frontend happy path** | 🔄 Partial | Login → onboarding → create project → test cases works with caveats |
| **Test automation** | ❌ Minimal | Go unit tests only; no integration/E2E suite |
| **Security posture** | ❌ Not production | No rate limiting, no API-key auth, weak `.env.example` |
| **Observability** | ❌ Not present | No OpenTelemetry, Prometheus, Grafana, or centralized logs |
| **Deployment pipeline** | ❌ Not present | CI builds only; no image push or deploy |
| **Documentation** | 🔄 Partial | Product docs complete; engineering wiki being finalized |

### One-sentence verdict

The Testra repository is a **functional alpha prototype** with a solid Clean Architecture backend and a partial Next.js frontend. It demonstrates the core value loop (auth → org → workspace → project → test cases → test runs → ingestion), but it requires production hardening, missing MVP modules, and infrastructure completion before any customer-facing launch.
