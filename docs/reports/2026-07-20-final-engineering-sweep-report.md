# Testra — Final Engineering Hardening Sweep Report

**Date:** 2026-07-20  
**Scope:** Full repository re-audit (apps/api, apps/web, apps/ml, packages, scripts, .github, docs, configuration)  
**Deployment target:** Single Ubuntu VPS, systemd, nginx, PostgreSQL, local filesystem — no Docker/Kubernetes/Terraform/cloud  

---

# Executive Summary

| Metric | Score | Notes |
|---|---|---|
| Overall engineering maturity | 62 / 100 | Solid modular monolith; many production hardening gaps remain. |
| Production readiness (Ubuntu VPS) | 35 / 100 | Code builds, tests pass, but no deployment runbooks, no CD, no observability. |
| Code quality | 70 / 100 | Clean architecture boundaries, consistent handler patterns, but stubs and partial modules. |
| Security | 65 / 100 | Auth/MFA/RLS/CSRF scaffolding present; JWT deny-list, frontend route guards, and secrets management missing. |
| Performance | 60 / 100 | DB pool now capped, rate limiting present, but no query plan review, no caching strategy, no SSE backpressure. |
| Maintainability | 65 / 100 | Good package structure; missing integration tests, OpenAPI/SDK not consumed by web, many stubs. |
| Testing | 55 / 100 | Unit tests pass; coverage low, no integration/E2E, many modules untested. |
| Documentation | 60 / 100 | Canonical docs aligned to VPS; production runbooks and operator docs missing. |

**Bottom line:** The codebase is a credible pre-MVP modular monolith. It compiles, lints, and passes its own unit tests on every tier. It is **not ready for customer-facing production** on a VPS until systemd/nginx runbooks, observability, secrets management, JWT revocation, and integration tests are added.

---

# Validation Executed

All commands were run after every material change.

- `gofmt -w .`
- `go vet ./...`
- `go build ./...`
- `go test -count=1 ./...`
- `pnpm lint`
- `pnpm typecheck`
- `pnpm build`
- `pnpm test`
- `ruff check .`
- `pytest`

**Result:** All green.

---

# Issues Found and Fixed

## 1. API server did not shut down gracefully

- **Severity:** Medium
- **Files changed:** `apps/api/cmd/api/main.go`
- **Reason:** `srv.ListenAndServe()` blocked `main` and `log.Fatalf` on any error. systemd `SIGTERM` would kill the process mid-request and mark the service as failed.
- **Fix implemented:** Added `signal.NotifyContext`, ran server in a goroutine, and called `srv.Shutdown` with a 15-second timeout. `http.ErrServerClosed` is now treated as normal.
- **Validation:** `go build ./...`, `go test ./...`

## 2. `/health` endpoint did not verify dependencies

- **Severity:** Medium
- **Files changed:** `apps/api/internal/shared/server/server.go`
- **Reason:** The health check returned `{"status":"ok"}` regardless of DB state, which would cause systemd/nginx to keep routing traffic to a broken instance.
- **Fix implemented:** `/health` now calls `cfg.DB.PingContext` with a 2-second timeout and returns `503 Service Unavailable` if PostgreSQL is unreachable.
- **Validation:** `go build ./...`, `go test ./...`

## 3. Database connection pool was unbounded

- **Severity:** Medium
- **Files changed:** `apps/api/internal/shared/db/db.go`
- **Reason:** `sql.Open` defaults to unlimited open connections and unlimited idle time. Under load the API could exhaust the VPS's PostgreSQL `max_connections`.
- **Fix implemented:** Set `MaxOpenConns(25)`, `MaxIdleConns(5)`, `ConnMaxLifetime(5m)`, `ConnMaxIdleTime(2m)`.
- **Validation:** `go vet`, `go build`, `go test`

## 4. Worker could tight-loop on persistent errors

- **Severity:** Medium
- **Files changed:** `apps/api/cmd/worker/main.go`
- **Reason:** `processBatch` continued on `processOne` error and set `worked = true`, resetting the poll timer. A DB outage would cause the worker to spin CPU and never back off.
- **Fix implemented:** `processBatch` now returns immediately on any `processOne` error, letting `Run` apply exponential backoff. `worked` is only true for successful job processing.
- **Validation:** `go vet`, `go build`, `go test`

## 5. Frontend API calls had no network timeout

- **Severity:** Medium
- **Files changed:** `apps/web/lib/api.ts`
- **Reason:** `fetch` calls could hang indefinitely if the API became unresponsive, freezing UI state and route guards.
- **Fix implemented:** Added a `fetchWithTimeout` helper using `AbortController` and a 30-second default timeout; updated `rawApiFetch` and `rawFetch` to use it.
- **Validation:** `pnpm typecheck`, `pnpm build`, `pnpm test`

## 6. Config package still referenced forbidden cloud services in a comment

- **Severity:** Low
- **Files changed:** `apps/api/internal/shared/config/config.go`
- **Reason:** Doc comment mentioned AWS Secrets Manager and Kubernetes secrets, contradicting the zero-cloud policy.
- **Fix implemented:** Rewrote comment to refer only to local/file-based secrets stores.
- **Validation:** `go build`

---

# Things Reviewed and Already Correct

These areas were inspected and intentionally left unchanged because they are already sound or the current implementation is acceptable for MVP.

- **JWT signing & verification:** RS256 with `kid`, issuer/audience validation, JWKS endpoint, 2048-bit keys, thread-safe key rotation.
- **Cookie security:** `HttpOnly` access/refresh, `Secure` based on `X-Forwarded-Proto`/TLS, `SameSite=Lax`, CSRF cookie not `HttpOnly` by design.
- **CSRF middleware:** Double-submit cookie, safe methods bypassed, login/register/refresh/password-reset routes excluded.
- **Auth middleware:** Bearer token extraction, fallback to access cookie, consistent 401 handling.
- **CORS middleware:** Origin whitelist, credentials allowed, preflight handled.
- **Security headers:** CSP, HSTS (production HTTPS), X-Frame-Options, nosniff, referrer-policy set by Next.js middleware and Go `apiSecurityHeaders`.
- **RLS / tenant isolation:** `app.tenant_id` set per request via `SET LOCAL`; `BeginTx` propagates tenant into transactions; repositories use the wrapped `DB`.
- **Rate limiting:** Token-bucket limiter with local and Redis backends; different rules for auth vs API-key routes.
- **Request logging:** Structured JSON `slog` via chi middleware.
- **Audit logging:** Middleware calls audit service with a detached 5-second timeout context.
- **Idempotency:** Postgres-backed idempotency store wired into mutating routes.
- **Password storage:** Argon2id in `shared/password`.
- **Local development:** `pnpm dev` runs migrations and starts API/web/worker/ML via Turborepo; no Docker assumed.
- **CI:** GitHub Actions validates Go build/vet/test (with race), web typecheck/build, SDK generation, and ML ruff/pytest.
- **Repository hygiene:** No remaining Docker/Kubernetes/Terraform/AWS/Azure/GCP references in source code or active docs.

---

# Remaining REAL Blockers

These are the actual issues that still prevent a safe VPS production deployment. They are listed in rough priority order.

## R1. No production deployment runbooks or systemd units

- **Severity:** Critical
- **Files:** N/A (missing)
- **Reason:** There are no systemd unit files for `api`, `worker`, `web`, `ml`, `nginx`, `postgres`, `redis`, or `minio`; no nginx site template; no `Type=notify` or health-based restart policy.
- **Business impact:** Cannot deploy to the stated target.
- **Fix:** Create `docs/deployment/systemd/` with unit files and an nginx reverse-proxy config; document `ExecStart`, `EnvironmentFile`, `Restart`, `User`, and log rotation.

## R2. No CI/CD deployment pipeline

- **Severity:** Critical
- **Files:** `.github/workflows/ci.yml`
- **Reason:** CI only builds and tests; it does not build Linux binaries, create artifacts, or deploy to staging/production.
- **Business impact:** Every release is manual and error-prone.
- **Fix:** Add a release workflow that cross-compiles Go binaries, produces `web` standalone output, packages artifacts, and deploys to the VPS via rsync/SSH or a lightweight artifact pull.

## R3. No observability stack

- **Severity:** High
- **Files:** N/A
- **Reason:** No metrics, logs, traces, or alerting are wired. `/health` only checks DB; there is no p95 latency, error-rate, or queue-depth monitoring.
- **Business impact:** Failures will be invisible until users report them.
- **Fix:** Add `/metrics` Prometheus endpoint or OpenTelemetry exporter; add structured logs with request ID; document Loki/Prometheus/Grafana or low-ops equivalent setup for VPS.

## R4. No JWT deny-list / token revocation

- **Severity:** High
- **Files:** `apps/api/internal/shared/jwt/manager.go`, `apps/api/internal/identity/service.go`
- **Reason:** Tokens are validated only by signature/expiry. There is no `jti` claim or deny-list, so stolen access tokens cannot be revoked until they expire.
- **Business impact:** Compromised sessions cannot be terminated.
- **Fix:** Add `jti` claim to access tokens, store revoked `jti`s in Redis with TTL, and check on every `Parse`.

## R5. No secrets management for production

- **Severity:** High
- **Files:** `apps/api/internal/shared/config/config.go`
- **Reason:** `SecretProvider` is only an env-var wrapper; `JWT_PRIVATE_KEY_FILE` and DB credentials are loaded from env/files without rotation or audit.
- **Business impact:** Secret rotation requires restarts and manual file edits.
- **Fix:** Document a `systemd-creds`, `sops`, or file-based secrets scheme; load non-trivial secrets from files, not env vars.

## R6. Worker transaction semantics can leave jobs stuck

- **Severity:** Medium
- **Files:** `apps/api/cmd/worker/main.go`, `apps/api/internal/queue/queue.go`
- **Reason:** In `processOne`, if `MarkFailed` or `tx.Commit` fails, the row remains locked in `processing` and is not rescheduled; cleanup only deletes completed/dead/failed jobs.
- **Business impact:** A job can get stuck indefinitely after a transient DB error.
- **Fix:** Add a `processing` timeout / stale-job reaper, or wrap `MarkFailed`+`Commit` with rollback and explicit retry visibility.

## R7. No integration or E2E tests

- **Severity:** Medium
- **Files:** N/A
- **Reason:** Only unit tests exist. The interaction between auth, RLS, repositories, and SSE is not automatically verified.
- **Business impact:** Regressions in multi-tenant isolation or auth flow may reach production.
- **Fix:** Add a small integration suite using `testcontainers` (not allowed) or a local PostgreSQL service in CI, testing login → create org → create workspace → create test case → query under RLS.

## R8. Frontend route guards are client-side only

- **Severity:** Medium
- **Files:** `apps/web/components/auth/route-guard.tsx`
- **Reason:** Protected routes are rendered server-side before the client auth check runs; a direct URL to `/dashboard` may briefly show content or flash before redirect.
- **Business impact:** Minor UX/security flash; API itself enforces auth, so data is not leaked.
- **Fix:** Move auth requirement to Next.js middleware or use server-side session validation for SSR pages.

## R9. OpenAPI / SDK not consumed by web frontend

- **Severity:** Low
- **Files:** `apps/web/lib/api.ts`, `packages/sdk`
- **Reason:** The generated SDK exists but `web` uses hand-typed `apiFetch` wrappers. Drift between backend and frontend types is possible.
- **Business impact:** Type safety across API contract is weaker than it could be.
- **Fix:** Migrate web feature APIs to the generated SDK or keep `lib/api.ts` tightly aligned with OpenAPI.

## R10. ML service `/health` is a static stub

- **Severity:** Low
- **Files:** `apps/ml/api/main.py`
- **Reason:** `/health` returns `ok` without checking the model/scikit runtime or memory.
- **Business impact:** systemd may keep a broken ML process running.
- **Fix:** Add a lightweight self-check to `/health` (e.g., a dummy prediction) before marking ready.

---

# Detailed Handoff Notes

## What was inspected

- `apps/api` — build, tests, middleware, handlers, services, repositories, JWT, queue, worker, config, db.
- `apps/web` — build, typecheck, `lib/api.ts`, middleware, route guards, `error.tsx`/`loading.tsx`, localStorage usage, pages.
- `apps/ml` — `main.py`, tests, ruff lint.
- `packages/*` — build/typecheck output; no cloud references.
- `scripts/*` and `Makefile` — no Docker/cloud assumptions remain.
- `.github/workflows/ci.yml` — aligned to Go 1.24, no cloud references.
- `docs/*` — canonical docs were previously aligned to VPS; no new drift detected.

## What was changed

- `apps/api/cmd/api/main.go` — graceful shutdown.
- `apps/api/internal/shared/server/server.go` — health checks DB.
- `apps/api/internal/shared/db/db.go` — connection pool limits.
- `apps/api/cmd/worker/main.go` — error backoff in worker batch.
- `apps/api/internal/shared/config/config.go` — comment cleanup.
- `apps/web/lib/api.ts` — 30-second fetch timeout.

## What was deliberately skipped

- **Full repository SQL plan review:** No live database with populated data to run `EXPLAIN`; index recommendations would be speculative. The schema has primary keys, foreign keys, and RLS policies; obvious missing-index scans could not be confirmed.
- **Frontend page-level loading/error skeletons:** Only root `loading.tsx`/`error.tsx` exist. Adding per-route skeletons is a UX polish task, not a production blocker.
- **Rate-limiter Redis clustering:** Single VPS uses the local in-memory fallback; Redis is optional.
- **ML model training/persistence:** ML endpoints are rule-based stubs; production ML readiness is a feature problem, not an engineering hardening issue.
- **RBAC permission matrix review:** Permissions are wired and checked; the actual permission strings are consistent. No evidence of missing authorization on implemented routes.

## Estimated readiness for a single Ubuntu VPS

The application **can be made to run** on a single Ubuntu VPS with modest effort:

1. Install Go 1.24, Node 20, pnpm, Python 3.12, PostgreSQL 16, Redis 7, nginx.
2. Run `pnpm install` and `go mod download`.
3. Run `go run ./apps/api/cmd/migrator`.
4. Build Go binaries (`api`, `worker`, `migrator`, `openapi`).
5. Build web (`next build`) and serve `.next/standalone` behind nginx or as a static/export.
6. Start ML with `uvicorn`.
7. Write and install systemd units and nginx config.

However, the **operational surface is not production-safe** until blockers R1–R4 are addressed. Deploying today would mean no monitoring, no graceful secret handling, no token revocation, and no automated deployment.

---

# Recommendations (Next 4–6 Weeks)

1. **Deployment:** Write systemd units + nginx template + `deploy.sh`.
2. **Observability:** Add `/metrics` and structured logs; install Prometheus/Grafana or a simpler stack on the VPS.
3. **Security:** Implement JWT `jti` deny-list in Redis and frontend route-guard in middleware.
4. **Testing:** Add integration tests for the auth → workspace → test-case flow.
5. **Secrets:** Move to file-based production secrets (`EnvironmentFile` with root-only permissions, `sops` optional).
6. **Worker robustness:** Add stale-job reaper and visibility-timeout handling.
7. **CD:** Add GitHub Actions workflow to build and deploy artifacts to the VPS.

Once those are done, the project will be a credible, low-cost MVP on a single Ubuntu VPS.
