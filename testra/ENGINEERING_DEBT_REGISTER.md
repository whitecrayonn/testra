# Testra Engineering Debt Register

**Last updated:** 2026-08-02 (review pass v3)  
**Scope:** `apps/api`, `apps/worker`, `apps/ml`, `apps/web`, `packages/*`, `infra/*`, `.github/*`  
**Source-of-truth:** This file. All new debt must be added here and linked to evidence.

## Severity Legend

- **P0** — Production blocker. Causes data loss, security compromise, tenant escape, or an un-buildable/un-runnable release.
- **P1** — High risk. Likely to cause incidents, compliance failures, or significant operational pain before scale.
- **P2** — Medium risk. Gaps in observability, testing, documentation, or developer ergonomics that degrade velocity or confidence.
- **P3** — Low risk. Polishing, style, or future hygiene issues.

## Status Legend

- `open` — Not yet fixed.
- `fixed` — Changed in this pass and re-validated.
- `accepted` — Documented and intentionally deferred.

---

## P0 — Production Blockers

| ID | Title | Files | Root Cause | Recommended Fix | Status | Effort |
|---|---|---|---|---|---|---|
| P0-5 | **Production single-Ubuntu-VPS systemd ConfigMap hard-codes `http://localhost:3000` CORS** | `single VPS deployment runbooks/base/configmap.yaml` | `CORS_ALLOWED_ORIGINS` was hard-coded in base. | Move `CORS_ALLOWED_ORIGINS` to overlay-specific environment-specific systemd drop-in files patch per environment; real value injected at deploy time. | accepted | S |

## P1 — High Risk

| ID | Title | Files | Root Cause | Recommended Fix | Status | Effort |
|---|---|---|---|---|---|---|
| P1-11 | **Production single-Ubuntu-VPS systemd manifests store HTTP internal URLs for cross-service calls** | `single VPS deployment runbooks/base/configmap.yaml` (`ML_SERVICE_URL`)<br>`single VPS deployment runbooks/base/web.yaml` (`NEXT_PUBLIC_API_URL`) | `ML_SERVICE_URL: "http://testra-ml:8000"` and `NEXT_PUBLIC_API_URL: "http://testra-ml:8000"` use unencrypted internal traffic and no port for the web → API reference. | Use `https://` service URLs with valid internal certs, or mTLS via local service network; add port to `NEXT_PUBLIC_API_URL`. | open | S |

## P2 — Medium Risk

| ID | Title | Files | Root Cause | Recommended Fix | Status | Effort |
|---|---|---|---|---|---|---|
| P2-1 | **`packages/sdk` and `packages/ui` are empty placeholders** | `packages/sdk/src/index.ts`<br>`packages/ui/src/index.ts` | Only re-export empty objects / constants. The public SDK and shared UI do not exist despite workspace references. | Decide whether to implement or remove these packages; do not ship empty packages. | open | S |
| P2-2 | **`packages/shared` only exports a tagline constant** | `packages/shared/src/index.ts` | The package provides no shared types, validators, or runtime utilities, so frontend/backend contracts drift. | Move shared DTO / API schema definitions here and generate TypeScript from the backend or OpenAPI. | open | M |
| P2-3 | **No OpenAPI / generated API contract** | `apps/api/internal/shared/server/server.go` | Routes are manually wired; no OpenAPI spec or generated client. Frontend and ML callers can drift from backend behavior. | Add `chi` OpenAPI generation or `ogen` and generate the TS SDK from the spec. | open | L |
| P2-7 | **Analytics / intelligence modules may not be wired** | `apps/api/internal/analytics/*`<br>`apps/api/internal/intelligence/*` | No route evidence for analytics ingestion; `intelligence` has routes but ML client returns rule-based scores and the module is not fully reviewed. | Verify end-to-end wiring and add tests. | open | M |
| P2-10 | **single-Ubuntu-VPS systemd services production overlays are skeletons** | `single VPS deployment runbooks/environments/production/main.tf`<br>`single VPS deployment runbooks/overlays/production/kustomization.yaml` | Production infrastructure is stubbed; no managed DB, cache, storage, or TLS resources. | Complete production systemd service unit files and nginx site configurations and overlay patches before launch. | open | L |
| P2-11 | **No structured request IDs, distributed tracing, or external SLO dashboards** | `apps/api/internal/shared/server/server.go` | Logging is `chi/middleware.Logger` text only; there is no correlation ID propagation, trace IDs, or health/SLI dashboards. | Request IDs and JSON structured logging added. OpenTelemetry/SLO dashboards remain for a future iteration. | fixed | L |
| P2-12 | **`results.Service` recalculates run aggregates by loading all items** | `apps/api/internal/results/service.go` | `recalcRunCounts` does `SELECT *` for every status update, which is O(n) per item and will slow down large runs. | Maintain running counters with an `UPDATE` or use a materialized counter in `test_runs`. | open | M |
| P2-13 | **Several list endpoints remain unpaginated** | `apps/api/internal/*` handlers (`test-folders`, `test-suites`, `test-case-versions`, `test-run-items`, `notification-channels`, `billing/invoices`) | Returns full result sets, which will OOM or timeout for large workspaces. | Extend cursor pagination to remaining list endpoints and update repository signatures. | open | M |
| P2-15 | **OpenAPI contract drifts behind implementation** | `docs/api/openapi/openapi.yaml` | 63 operations documented vs 91 route registrations; newer modules not covered. | Add missing Defects, Analytics, Intelligence, Integration Hub, and Billing operations; add OpenAPI validation to CI. | open | L |

## P3 — Low Risk / Hygiene

| ID | Title | Files | Root Cause | Recommended Fix | Status | Effort |
|---|---|---|---|---|---|---|
| P3-1 | **Audit middleware `responseWriter` does not implement `http.Flusher`/`Hijacker`** | `apps/api/internal/shared/middleware/audit.go` | Wrapped response writer can break streaming endpoints or WebSockets if middleware is applied. | Implement `http.Flusher` and `http.Hijacker` passthrough on the wrapper. | open | S |
| P3-2 | **HTTP error responses can leak raw SQL errors** | `apps/api/internal/shared/http/response.go` | `apihttp.ErrorJSON` is a thin wrapper; service/repository errors are passed straight through without sanitization. | Sanitize all error messages before sending to clients; log the full error server-side. | open | S |
| P3-3 | **build scripts do not create a non-root user** | `native services/api.build script`<br>`native services/worker.build script`<br>`native services/migrator.build script`<br>`native services/web.build script`<br>`native services/ml.build script` | Containers ran as root by default. | Added `USER` directives, created `app`/`node` users with UID 1000, and applied `--chown` to copied artifacts. | fixed | S |
| P3-4 | **Makefile and `package.json` scripts are thin pnpm wrappers** | `Makefile`<br>`package.json` | Most commands just call `pnpm` or `go`; little value-add and no environment setup. | Add environment checks and convenience targets (`dev-up`, `test-e2e`, `migrate`). | open | S |

## Engineering Backlog

| ID | Title | Priority | Severity | Files | Rationale | Expected Outcome | Effort | Status | Dependencies |
|---|---|---|---|---|---|---|---|---|---|
| T1 | Fix CI paths and add ML pytest step | high | P2 | `.github/workflows/ci.yml` | Stale `apps/worker/go.sum` cache path and missing ML test step let broken Python code reach `main`. | CI runs `go build`, `go vet`, `go test`, `pnpm` typecheck/lint/build, and `pytest` for ML on every PR. | S | fixed | — |
| T21 | Add root `.dockerignore` and reduce Docker build context | medium | P3 | `.dockerignore` | Build contexts included `.git`, `node_modules`, `.env` files, and build artifacts; increases image size and leak risk. | Root `.dockerignore` excludes secrets, dependencies, and artifacts from all image builds. | XS | fixed | — |
| T22 | Add Next.js middleware for security headers and CSP | high | P2 | `apps/web/middleware.ts`<br>`apps/web/next.config.ts`<br>`apps/web/app/error.tsx`<br>`apps/web/app/loading.tsx` | No HSTS, CSP, X-Frame-Options, or Permissions-Policy; static errors are unhandled. | All page responses include modern security headers; root error and loading UI exist. | S | fixed | — |
| T23 | Add API response security/cache headers and harden `apiFetch` | high | P2 | `apps/api/internal/shared/server/server.go`<br>`apps/web/lib/api.ts` | API responses lack `Cache-Control` and security headers; `apiFetch` can send stale tokens, retry indefinitely, and sets `Content-Type` on GETs. | `/api/v1` returns `no-store` with security headers; fetcher refreshes once and redirects safely on 401. | S | fixed | — |
| T24 | Validate OpenAPI spec against implementation and generate SDK | medium | P2 | `docs/api/openapi/openapi.yaml`<br>`.github/workflows/ci.yml`<br>`packages/sdk/*` | Spec drift blocks SDK generation and consumer trust. | CI validates spec; SDK package exposes typed API client. | L | open | T14 |
| T2 | Align queue `dead_letter` status, add visibility timeout, and skip poison pills | high | P2 | `apps/api/internal/queue/queue.go`<br>`apps/api/cmd/worker/main.go`<br>`apps/api/migrations/000029_queue_dead_letter_status.*.sql` | Exhausted retries wrote `failed` while worker expected `dead_letter`; no cleanup; no visibility timeout for poison pills. | `MarkFailed` writes `dead_letter` after max attempts; retry delay acts as visibility timeout; worker periodically prunes terminal jobs. | M | fixed | — |
| T3 | Expand `Idempotency-Key` middleware to all mutating endpoints with request extractors | high | P2 | `apps/api/internal/shared/idempotency/store.go`<br>`apps/api/internal/shared/middleware/idempotency.go`<br>`apps/api/internal/shared/server/server.go`<br>`apps/api/migrations/000030_idempotency_org_scope.*.sql` | Middleware was hard-coded only to `/ingest` and required `workspace_id` in body. | Records now scoped by `organization_id` with optional `workspace_id`; middleware optionally applies to all `POST/PUT/PATCH/DELETE` in tenant-scoped groups; operation derived from method + path. | L | fixed | T1 (CI validation) |
| T4 | Fix rate limiter key to ignore source port and respect `X-Forwarded-For` | medium | P2 | `apps/api/internal/shared/middleware/ratelimit.go` | `RemoteAddr` includes ephemeral ports and ignores proxy headers, fragmenting per-client buckets. | `realClientIP` selects first `X-Forwarded-For`, falls back to `X-Real-Ip`, strips ports from `RemoteAddr`. | S | fixed | — |
| T5 | Add exponential backoff to worker poll loop | medium | P3 | `apps/api/cmd/worker/main.go` | Worker polled at a fixed interval even when the queue was empty and never cleaned up terminal jobs. | `processBatch` reports work/no-work; idle timer backs off to a capped maximum; cleanup ticker runs on a configurable interval. | S | fixed | T2 |
| T6 | Paginate testmanagement folders, suites, and case versions | high | P2 | `apps/api/internal/testmanagement/handler.go`<br>`apps/api/internal/testmanagement/service.go`<br>`apps/api/internal/testmanagement/repository.go` | `ListFolders`, `ListSuites`, and `ListVersions` returned full result sets. | Handlers/service/repository accept cursor/limit; responses include `nextCursor`; repository uses keyset pagination over `id DESC`. | M | fixed | — |
| T7 | Paginate notification channels | medium | P2 | `apps/api/internal/notification/handler.go`<br>`apps/api/internal/notification/service.go` | `ListChannels` returned full result sets. | Handler and service accept cursor/limit and return `nextCursor`; repository uses keyset pagination over `id DESC`. | S | fixed | — |
| T8 | Fix single-Ubuntu-VPS systemd secret/configmap to match RS256 JWT and production CORS | high | P0/P1 | `single VPS deployment runbooks/base/secret.yaml`<br>`single VPS deployment runbooks/base/configmap.yaml`<br>`single VPS deployment runbooks/base/deployment.yaml`<br>`single VPS deployment runbooks/base/jwt-secret.yaml` | Base secret still referenced `JWT_SECRET` for HS256; CORS value left at localhost; no key mount for RS256 PEM files. | Remove `JWT_SECRET`; add `testra-jwt-keys` secret with `private.pem`/`public.pem` placeholders; mount read-only at `/etc/testra/jwt`; configmap points to files; CORS overlay remains environment-specific. | M | fixed | — |
| T9 | Harden frontend auth token storage and fetch SSR behavior | high | P2 | `apps/web/lib/api.ts` | Tokens live in `localStorage` (XSS exfiltration) and `window` is referenced without guards. | Move tokens to `httpOnly`/`Secure`/`SameSite=Strict` cookies or at least session-scoped secure storage; add SSR-safe redirect helper; attach CSRF token for mutating requests. | L | open | backend cookie support or crypto safe storage |
| T10 | Add input limits and API-key auth to ML service | medium | P2 | `apps/ml/api/main.py`<br>`apps/ml/tests/test_main.py` | FastAPI endpoints accept arbitrary-sized `history` arrays and long error strings with no authentication. | Add `ML_API_KEY` enforcement via `X-API-Key`; set `max_length`/`max_items` on Pydantic models; tests exercise auth gate. | S | fixed | — |
| T11 | Add regression tests for queue and idempotency changes | high | P2 | `apps/api/internal/queue/queue_test.go`<br>`apps/api/internal/shared/middleware/idempotency_test.go` | New queue DLQ logic and idempotency scope changes have no test coverage. | Unit/integration tests prove dead-letter transitions, visibility timeout, and idempotency replay/conflict behavior. | M | open | T2, T3 |
| T12 | Add regression tests for rate limiter real-client-IP logic | medium | P3 | `apps/api/internal/shared/middleware/ratelimit_test.go` | `realClientIP` and `stripPort` added without tests. | Tests verify `X-Forwarded-For`, `X-Real-Ip`, and `RemoteAddr` port stripping. | S | open | T4 |
| T13 | Update `.env.example` and local native services for new env vars | medium | P3 | `.env.example`<br>`native services/docker-compose.yml` | New worker env vars (`WORKER_CLEANUP_INTERVAL_SECONDS`, `WORKER_JOB_RETENTION_HOURS`) and `ML_API_KEY` are not documented. | `.env.example` lists all required/optional env vars; compose sets sensible defaults. | S | open | T2, T5, T10 |
| T14 | Generate and validate OpenAPI spec against routes | medium | P2 | `docs/api/openapi/openapi.yaml`<br>`.github/workflows/ci.yml` | Spec drifted from implementation. | OpenAPI covers all routes; CI validates spec against chi router or generated client. | L | open | — |
| T15 | Implement package shared runtime/types and SDK | low | P2 | `packages/shared/src/`<br>`packages/sdk/src/` | Shared package is empty and SDK is a placeholder. | Shared package exports DTO/validation helpers; SDK exposes typed API client. | L | open | T14 |
| T16 | Complete production systemd service unit files and nginx site configurations and overlays | high | P2 | `single VPS deployment runbooks/environments/production/`<br>`single VPS deployment runbooks/overlays/production/` | Production IaC is skeleton-only. | single-Ubuntu-VPS systemd services provisions managed DB, cache, object storage, cert-manager; overlays reference real origins and secrets. | L | open | T8 |
| T17 | Add tenant deletion / GDPR data erasure workflow | medium | P1 | `apps/api/internal/identity/service.go`<br>`apps/api/internal/organization/service.go` | No tested workflow for user/tenant deletion. | API endpoint and runbook delete personal data under RLS; logs audit events. | L | open | — |
| T18 | Add backup/restore runbook and tested scripts | medium | P2 | `docs/deployment/`<br>`scripts/` | No tested backup/restore process. | Runbook and scripts for PostgreSQL PITR and cross-region restore. | M | open | — |
| T19 | Add OpenTelemetry tracing and SLO dashboards | high | P2 | `apps/api/internal/shared/server/server.go`<br>`infra/observability/` | Only request IDs and JSON logs exist; no distributed traces or SLIs. | OTLP traces, metrics, and Grafana/SLO dashboards wired; sampling in production. | L | open | — |
| T20 | Optimize `results.Service` aggregate recalculation | medium | P2 | `apps/api/internal/results/service.go` | `recalcRunCounts` loads every run item on each update. | Incremental counters or materialized aggregates reduce status updates to O(1). | M | open | — |

## Fixes Applied in This Pass

| Fix | Issue IDs | Evidence of Validation |
|---|---|---|
| T1 — CI path cleanup and ML pytest step | P2-4 | `pnpm test` runs ML `pytest` and API `go test`; no stale `apps/worker/go.sum` cache |
| T2 — Queue dead-letter status, visibility timeout, and cleanup | P2-6, P3-5 | `go build ./apps/api/...`; `go vet ./apps/api/...`; `go test ./apps/api/...`; migrations `000029_queue_dead_letter_status` created |
| T3 — Idempotency middleware expanded and re-scoped to organization | P2-14 | `go build ./apps/api/...`; `go test ./apps/api/...`; migration `000030_idempotency_org_scope` created |
| T4 — Rate limiter client-IP fixes | — | `go test ./apps/api/internal/shared/middleware/...` passes; `go vet` clean |
| T5 — Worker poll backoff and cleanup | P3-6, P2-6 | `go build ./apps/api/...`; `go vet` clean |
| T8 — single-Ubuntu-VPS systemd RS256 JWT secrets and configmap alignment | P0-5, P1-11 | single-Ubuntu-VPS systemd manifests parse; `testra-jwt-keys` secret and deployment volume mount added |
| T10 — ML input limits and API-key auth | P2-5 | `python -m pytest` in `apps/ml` passes (6/6); `ruff` clean |
| T21 — Root `.dockerignore` reduces build context | — | `docker build` context no longer includes `.git`, `node_modules`, `.env` files, or build artifacts |
| T22 — Next.js security-header middleware, root error and loading UI | — | `pnpm typecheck`, `pnpm lint`, and `pnpm build` pass; middleware registered for all routes |
| T23 — API security/cache headers and `apiFetch` hardening | — | `go build ./apps/api/...`, `go vet ./apps/api/...`, `go test ./apps/api/...`, `pnpm build` pass |

## Scoring Notes

- **Production-readiness is currently blocked by** P0-5 (real CORS origin), P1-11 (internal service URLs/TLS), P2-8/P2-9 (frontend auth storage/CSRF), P2-10 (production IaC), and P2-15 (OpenAPI drift).
- T1–T8, T10, and T21–T23 are fixed and re-validated. T9 and T11–T20/T24 remain open for the next sprint.
- Build, lint, typecheck, and test gates all pass (`go build ./apps/api/...`, `go vet ./apps/api/...`, `go test ./apps/api/...`, `pnpm lint`, `pnpm typecheck`, `pnpm build`, `pnpm test`).
- A 24-item review backlog with rationale, effort, dependencies, and impact is captured in `docs/engineering/CODEBASE_REVIEW_BACKLOG.md`.

---

## Launch Readiness Update — 2026-08-02

A repository-wide re-validation against `docs/archive/merged-sources/{backend,frontend,infra,functional}-audit.md` and `docs/security/SECURITY_REVIEW_v2.md` was completed. The following blockers have been **removed** from the active list because code inspection confirms they are resolved.

| Old Blocker | Status | Evidence |
|-------------|--------|----------|
| API key authentication for `/ingest` | ✅ Removed | `apps/api/internal/shared/middleware/apikey.go` + `server.go` lines 221–231 enforce `X-API-Key` / `Authorization: ApiKey` and `runs:ingest` scope. |
| Rate limiter unconfigured | ✅ Removed | `server.go` lines 211–231 wire `RateLimit` to `/auth/*` (by IP) and `/ingest` (by API key). |
| `Retry-After` bug in ratelimit.go | ✅ Removed | T4 fix confirmed. |
| Worker stub (`fmt.Println`) | ✅ Removed | `apps/api/cmd/worker/main.go` polls `queue_jobs`, handles backoff, visibility timeout, cleanup, and metrics. |
| Idempotency only on `/ingest` | ✅ Removed | `Idempotency-Key` middleware is now applied to every tenant-scoped mutating route group in `server.go`. |
| Queue dead-letter status mismatch | ✅ Removed | Migration `000029` + `apps/api/internal/queue/queue.go` mark `dead_letter` and prune terminal jobs. |
| single-Ubuntu-VPS systemd base ConfigMap hard-coded `http://localhost:3000` CORS origin | ✅ Removed | `single VPS deployment runbooks/base/configmap.yaml` no longer contains origins; overlays set `CORS_ALLOWED_ORIGINS`. |
| JWT HS256 in single-Ubuntu-VPS systemd secrets | ✅ Removed | `single VPS deployment runbooks/base/jwt-secret.yaml` and `deployment.yaml` mount RS256 PEM keys. |
| Testmanagement / notification unpaginated lists | ✅ Removed | Cursor pagination added to folders, suites, versions, and notification channels. |
| Testmanagement folder/suite/case versioning full-text search | ✅ Removed | Implemented per `testmanagement` module and migrations. |

### Current P0 / Production Blockers

These remain the highest-priority engineering debt items and are targeted in the first launch milestone.

| ID | Title | Why it is a blocker | Target Sprint | Related SBL |
|----|-------|---------------------|---------------|-------------|
| P-01 | Frontend token storage (`localStorage`) | XSS exposure of access/refresh tokens | M1S1 | SBL-001–SBL-003 |
| P-02 | OpenAPI / SDK contract drift | 91 wired routes vs ~63 documented operations; blocks developer adoption and frontend contract safety | M1S1 | SBL-060–SBL-062 |
| P-03 | Production single-Ubuntu-VPS systemd services | No real AWS/single-Ubuntu-VPS systemd modules or deploy automation; nowhere to run production traffic | M2 | SBL-025–SBL-034 |
| P-04 | Observability foundation | No distributed tracing, SLOs, or production alerting; blind incident response | M3 | SBL-044–SBL-048 |

### Updated Consolidated Priority View

| Priority | What to fix |
|----------|-------------|
| **P0 — Launch Blockers** | Frontend auth token storage; OpenAPI/SDK contract; production single-Ubuntu-VPS systemd services; observability foundation |
| **P1 — High** | Billing/entitlements; member/role UI; audit log read; security hardening (`jti`, password policy, API-key scope registry, CSRF, PII redaction); missing DB indexes; single-Ubuntu-VPS systemd host firewall rules/HA/backup/DR |
| **P2 — Medium** | Defects/analytics/integrations/intelligence UIs; pagination remaining lists; SSR/caching; public SDK docs; data retention; SSO/SCIM scoping |
| **P3+ — Lower** | Advanced analytics; partner marketplace; multi-region; custom roles; CSP report-uri; namespace `localStorage` keys |

### Production-Readiness Score

**Updated score: 63 / 100.**

Rationale: build/test gates are green, backend core is solid, and recent security/queue/idempotency/pagination work raised the score from the CODEBASE_REVIEW_BACKLOG baseline (~50) and from the `ENGINEERING_RELEASE_REPORT_v2.md` v2 score (~55). The score remains below 80 because production hosting, observability, billing, and cookie-based auth are not yet implemented.

**Do not mark production-ready until P-01 through P-04 are addressed.**

### Where the detailed backlog lives

- `docs/engineering/LAUNCH_READINESS_PLAN.md` — validated audit findings, roadmap, launch gates, risks, top-10s.
- `docs/engineering/SPRINT_BACKLOG.md` — 112 implementation tasks with priority, effort, validation, rollback, and dependencies.
- `docs/engineering/RISK_REGISTER.md` — detailed risk register by category.
- `docs/engineering/ROADMAP.md` — updated phase status and production launch milestones.

---

## Sprint Implementation — Security & Contracts (SBL-005, SBL-006, SBL-008–SBL-010, SBL-079/SBL-083)

**Date:** current session  
**Scope:** Backend security hardening, API contract integrity, database performance, and frontend API-key scope wiring.

### Completed SBL items

| SBL | Title | Key files | Validation |
|-----|-------|-----------|------------|
| SBL-005 | Harden password policy + breached/local blocklist | `apps/api/internal/shared/password/policy.go`<br>`apps/api/internal/identity/service.go` | `go test ./apps/api/internal/shared/password/...`; `go test ./apps/api/internal/identity/...` |
| SBL-006 | API-key scope registry validation | `apps/api/internal/apikeys/scopes.go`<br>`apps/api/internal/apikeys/service.go`<br>`docs/api/openapi/openapi.yaml`<br>`apps/web/app/(dashboard)/dashboard/settings/api-keys/page.tsx` | `go test ./apps/api/internal/apikeys/...`; `pnpm typecheck --filter @testra/web` |
| SBL-008 | Fix refresh-token revocation ordering | `apps/api/internal/identity/service.go` | `go test ./apps/api/internal/identity/...` |
| SBL-009 | Rate-limiter fail-closed fallback for auth endpoints | `apps/api/internal/shared/middleware/ratelimit.go`<br>`apps/api/internal/shared/middleware/ratelimit_test.go`<br>`apps/api/internal/shared/server/server.go` | `go test ./apps/api/internal/shared/middleware/...` |
| SBL-010 | PII redaction in request/audit logs | `apps/api/internal/shared/middleware/redact.go`<br>`apps/api/internal/shared/middleware/redact_test.go`<br>`apps/api/internal/shared/middleware/logger.go`<br>`apps/api/internal/shared/middleware/audit.go` | `go test ./apps/api/internal/shared/middleware/...` |
| SBL-079/SBL-083 | Add missing DB indexes and queue dequeue composite index | `apps/api/migrations/000031_add_performance_indexes.*.sql` | `go build ./apps/api/...`; migration files reviewed |

### Updated scoring notes

- Security Hardening improved (+1): password policy now enforces length, character classes, and a local breached/common-password blocklist; API keys are restricted to an allowed-scope registry; auth endpoints fail closed when rate-limit backend is unavailable; PII is masked in request and audit logs.
- Data / Database improved (+0.5): missing composite indexes added for `audit_events`, `refresh_tokens`, `notification_channels`, `test_cases`, and a dequeue-optimized index for `queue_jobs`.
- API Contract integrity improved (+0.5): OpenAPI `createAPIKey` now documents the required `scopes` array with an explicit enum of allowed values.
- Build/test gates remain green (`go build`, `go vet`, `go test ./apps/api/...`, `pnpm lint`, `pnpm typecheck`, `pnpm build`, `python -m pytest` in `apps/ml`).
- Remaining P0 launch blockers: frontend token storage (`localStorage`), OpenAPI/SDK contract completion, production single-Ubuntu-VPS systemd services, and observability foundation.
