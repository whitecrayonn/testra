# Testra Codebase Review Backlog

**Generated:** 2026-08-02 (review pass)  
**Scope:** Product, engineering, developer experience, frontend, backend, database, observability, CI/CD, security, testing, production readiness, and documentation.  
**Source of truth:** `ENGINEERING_DEBT_REGISTER.md` for tracked debt and `docs/engineering/ROADMAP.md` for release phasing.

This backlog consolidates the gaps found during the end-to-end review. Each item includes the affected files, rationale, effort, dependencies, and expected impact.

---

## Backlog

| ID | Area | Title | Files | Rationale | Effort | Dependencies | Impact |
|----|------|-------|-------|-----------|--------|--------------|--------|
| **PM-01** | Product Maturity | Tenant onboarding & self-service billing | `apps/api/internal/billing/*`, `apps/web/app/(dashboard)/settings/billing/*`, `single VPS deployment runbooks/base/*` | No billing module is wired end-to-end; SaaS launches cannot collect revenue or enforce plan limits. | L | Stripe integration, plan seed data | Blocks commercial launch |
| **PM-02** | Product Maturity | Organization/workspace/project usage limits & entitlements | `apps/api/internal/organization/*`, `apps/api/internal/billing/service.go`, `docs/architecture/ADR-???` | No per-tenant seat/run/storage quotas; enterprise tier cannot be sold. | M | Billing module | High |
| **PM-03** | Product Maturity | Public SDK and typed API client | `packages/sdk/*`, `packages/shared/*`, `docs/api/openapi/openapi.yaml` | The SDK package is empty; external adoption and partner integrations require a generated client. | M | OpenAPI contract generation | High |
| **PM-04** | Product Maturity | SSO / SAML / SCIM enterprise provisioning | `apps/api/internal/identity/*`, `apps/web/app/(auth)/login/page.tsx` | Enterprise ICP requires SSO; current email/password only blocks large contracts. | L | Identity refactor, SAML library | Critical for enterprise |
| **PM-05** | Product Maturity | Admin/super-admin console & tenant impersonation | `apps/api/internal/rbac/*`, `apps/web/app/(admin)/*` | No first-line support or ops tooling to diagnose tenant issues safely. | M | Audit logging UI | Medium |
| **EM-01** | Engineering Maturity | Populate `packages/shared` with DTOs/validators and remove duplication | `packages/shared/src/*`, `apps/web/lib/api.ts`, `apps/api/internal/*/handler.go` | Shared package only exports a tagline; frontend/backend types drift and validation is duplicated. | M | OpenAPI or hand-written contracts | High |
| **EM-02** | Engineering Maturity | Generate and validate OpenAPI spec from `chi` routes | `docs/api/openapi/openapi.yaml`, `apps/api/internal/shared/server/server.go`, `.github/workflows/ci.yml` | 91 routes are manually wired; spec drift will break SDKs and consumers. | L | `swaggo/swag` or `ogen` integration | Critical |
| **EM-03** | Engineering Maturity | Dead-code and unused dependency sweep | `apps/api/internal/*`, `apps/web/app/*`, `package.json` | `apitesting`, `integration` modules are stubs; unused dependencies inflate attack surface. | S | Static analysis tooling | Low/Medium |
| **EM-04** | Engineering Maturity | Module dependency lint / import graph enforcement | `docs/architecture/MODULE_DEPENDENCIES.md`, `scripts/` | No automated guard against cross-module imports violating Clean Architecture. | S | `go` import-lint or custom script | Medium |
| **DX-01** | Developer Experience | Add root `.dockerignore` and per-service ignore files | Root `.dockerignore` | Build contexts currently include `.git`, `node_modules`, and build artifacts. | XS | — | Medium (merged) |
| **DX-02** | Developer Experience | Add database seed and fixture scripts for local/dev environments | `scripts/dev/seed.mjs`, `apps/api/cmd/migrator` | No seed data makes manual testing and demos tedious. | M | Migration stability | Medium |
| **DX-03** | Developer Experience | Hot-reload and unified `make dev-up`/`dev-down` targets | `Makefile`, `apps/api/.air.toml`, `native services/docker-compose.yml` | Makefile is thin; `air` exists but not documented/invoked. | S | Compose health checks | Medium |
| **FE-01** | Frontend Experience | Adopt SWR/React Query for caching, retries, and optimistic updates | `apps/web/lib/api.ts`, `apps/web/app/(dashboard)/*` | Current fetcher has no caching or stale-while-revalidate; dashboard polls naively. | M | — | High |
| **FE-02** | Frontend Experience | Standardize loading, error, and empty states across all routes | `apps/web/app/*/loading.tsx`, `apps/web/app/*/error.tsx`, `apps/web/components/ui/empty-state.tsx` | Skeletons and error boundaries are missing in many routes. | M | UI component library | Medium |
| **FE-03** | Frontend Experience | Accessibility audit: labels, focus management, color contrast | `apps/web/components/*`, `apps/web/app/*` | No ARIA landmarks or focus traps; blocks WCAG/SOC 2 readiness. | M | UX review | Medium |
| **FE-04** | Frontend Experience | Move auth tokens out of `localStorage` | `apps/web/lib/api.ts`, `apps/api/internal/identity/*`, `apps/web/middleware.ts` | `localStorage` tokens are XSS-exfiltrable; requires `httpOnly`/`Secure`/`SameSite=Strict` cookies or BFF session. | L | Backend cookie/PKCE session support | Critical |
| **BE-01** | Backend API | Enforce pagination and filtering on all list endpoints | `apps/api/internal/*/handler.go`, `apps/api/internal/*/repository.go` | Folders, suites, versions, run items, invoices still return unbounded sets. | M | Pagination utility | High |
| **BE-02** | Backend API | Add request/response DTO validation and normalize error envelopes | `apps/api/internal/shared/http/response.go`, `apps/api/internal/*/handler.go` | Some handlers return raw errors or inconsistent `meta` shapes. | M | `packages/shared` validators | High |
| **BE-03** | Backend API | Wire `Idempotency-Key` generation into frontend fetcher | `apps/web/lib/api.ts` | API supports keys but the web client never sends them; mutating retries may duplicate. | S | `Idempotency-Key` TTL config | Medium |
| **BE-04** | Backend API | OpenAPI-driven contract and SDK generation in CI | `docs/api/openapi/openapi.yaml`, `.github/workflows/ci.yml`, `packages/sdk/*` | Frontend and SDK currently hand-code DTOs and drift from backend. | L | BE-02, EM-02 | High |
| **DB-01** | Database | Review and add missing indexes for tenant-scoped lookups | `apps/api/migrations/*` | RLS queries on `organization_id`/`workspace_id`/`project_id` rely on sequential scans at scale. | M | Query plan analysis | High |
| **DB-02** | Database | Implement data retention / purge jobs for results, audit, notifications | `apps/api/internal/queue/*`, `apps/api/cmd/worker/main.go`, `apps/api/migrations/*` | No retention policy enforcement violates ADR-005 and drives storage cost. | M | Worker scheduler | Medium |
| **OBS-01** | Observability | Add OpenTelemetry traces and metrics export | `apps/api/internal/shared/server/server.go`, `apps/api/cmd/worker/main.go`, `infra/observability/*` | Only JSON logs and Prometheus-style metrics exist; no distributed tracing or SLO dashboards. | L | Collector/otel libraries | Critical |
| **OBS-02** | Observability | Define and instrument SLO/SLI dashboards (latency, error budget, queue depth) | `docs/operations/MONITORING_LOGGING_GUIDE.md`, `single VPS deployment runbooks/base/*` | No alertable SLOs or runbooks; incidents will be detected late. | M | OBS-01, Grafana/Loki stack | High |
| **CI-01** | CI/CD | Add SBOM, dependency vulnerability scan, and release automation | `.github/workflows/ci.yml`, `scripts/release/*` | No SBOM or automated semantic-version releases; audit/compliance gap. | M | CI secrets, signing | Medium |
| **CI-02** | CI/CD | Add integration test job and mock external services | `apps/api/tests/integration/*`, `.github/workflows/ci.yml` | Integration tests exist but are not run in CI. | M | Postgres/Redis services in CI | High |
| **SEC-01** | Security Posture | Add security headers and CSP enforcement to API and web | `apps/api/internal/shared/server/server.go`, `apps/web/middleware.ts`, `apps/web/next.config.ts` | `X-Frame-Options`, `Cache-Control`, and CSP were missing or incomplete. | S | — | High (partially merged) |
| **SEC-02** | Security Posture | Implement secrets manager integration (Vault/AWS SM/single-Ubuntu-VPS systemd) | `apps/api/internal/shared/config/config.go`, `apps/api/internal/shared/secrets/*` | `SecretProvider` is a thin env wrapper; production secrets live in single-Ubuntu-VPS systemd placeholders. | L | Infra provisioning | Critical |
| **SEC-03** | Security Posture | Add audit trail UI and immutable audit export | `apps/api/internal/audit/*`, `apps/web/app/(dashboard)/settings/audit-logs/*` | Audit events are fire-and-forget with no UI or export. | M | Role permissions | Medium |
| **SEC-04** | Security Posture | Webhook/ingestion signature verification and replay protection | `apps/api/internal/automationhub/*`, `apps/api/internal/integrationhub/*` | Ingestion endpoints accept API keys but payload signatures and nonces are not validated. | M | Webhook provider secrets | High |
| **TEST-01** | Testing Strategy | Add contract and property-based fuzz tests for API handlers | `apps/api/tests/contract/*`, `apps/api/internal/*/handler_test.go` | No contract/fuzz tests; invalid payloads can panic or leak errors. | L | OpenAPI/JSON schema | Medium |
| **TEST-02** | Testing Strategy | Add load and tenant-isolation chaos tests | `scripts/load/*`, `apps/api/tests/integration/*` | No evidence of load testing or cross-tenant denial under concurrency. | L | Staging environment | High |
| **PROD-01** | Production Readiness | Complete production single-Ubuntu-VPS systemd services overlays | `single VPS deployment runbooks/environments/production/*`, `single VPS deployment runbooks/overlays/production/*` | Base manifests are stubs; no managed DB/cache/storage/TLS. | L | Cloud account, domain | Critical |
| **PROD-02** | Production Readiness | Document and script backup/restore runbooks | `docs/operations/DISASTER_RECOVERY_GUIDE.md`, `scripts/ops/*` | DR guide is policy-only; no tested restore scripts. | M | Infra provisioning | High |
| **PROD-03** | Production Readiness | Add graceful shutdown, circuit breakers, and rate-limit alerts | `apps/api/cmd/api/main.go`, `apps/api/internal/shared/middleware/ratelimit.go` | API server does not drain in-flight requests; no circuit breaker for ML/Stripe. | M | OBS-01 | Medium |
| **DOC-01** | Documentation | Synchronize ADRs and retire stale / pre-implementation docs | `docs/architecture/adrs/*`, `docs/archive/*` | `04_Architecture` pre-implementation draft is archived but still referenced. | S | Doc owner review | Low |
| **DOC-02** | Documentation | Update `README.md`, `ONBOARDING.md`, and environment docs post-changes | `README.md`, `docs/engineering/ONBOARDING.md`, `.env.example` | New env vars and middleware are not reflected in onboarding. | S | DX changes | Low |

---

## Recommended Next Sprint (MVP Closing)

1. **Security & trust:** SEC-01 (headers/CSP), SEC-02 (secrets manager), FE-04 (token storage).  
2. **Contracts:** EM-02 / BE-04 (OpenAPI), PM-03 (SDK).  
3. **Completeness:** BE-01 (pagination), PM-01 (billing scaffolding).  
4. **Quality:** CI-02 (integration tests), DB-01 (indexes), TEST-02 (load/tenant isolation).  

---

## Scoring Summary (Pre-Review)

| Category | Score (1–10) | Notes |
|---|---|---|
| Product Maturity | 4 | Core test management works; billing, entitlements, SSO, and admin tooling missing. |
| Engineering Maturity | 5 | Clean structure, but empty packages, OpenAPI drift, and manual DTOs. |
| Developer Experience | 6 | Compose and air exist; Makefile and seeding are thin. |
| Frontend Experience | 5 | Auth, onboarding, dashboard skeleton; no caching or robust error states. |
| Backend API | 6 | Idempotency, RBAC, RLS present; pagination and filter consistency incomplete. |
| Database | 6 | RLS and migrations robust; index and retention strategy not verified. |
| Observability | 4 | JSON logs and request IDs; no traces or SLO dashboards. |
| CI/CD | 5 | Build/test gates green; no integration tests, SBOM, or release automation. |
| Security Posture | 5 | Auth/RLS strong; secrets management, CSP, audit UI, webhook signatures missing. |
| Testing Strategy | 4 | Unit tests only; no integration/contract/load tests in CI. |
| Production Readiness | 3 | single-Ubuntu-VPS systemd/single-Ubuntu-VPS systemd services are scaffolding; no tested DR or secrets manager. |
| Documentation | 6 | Canonical docs exist; some stale references and drift. |

**Overall readiness:** ~5/10 — functional monorepo with strong architecture, not yet production-grade SaaS.
