# Testra Code Audit Report

**Date:** generated from the current audit session  
**Scope:** `apps/api`, `apps/web`, `apps/worker`, shared packages, middleware, queue and background worker  
**Repository:** `c:/Private/project/testra`  

---

## Executive Summary / Top-line Verdict

The Testra monorepo is **functionally well-structured and broadly works as documented** for the happy path. All Go unit tests pass, and the API test suites pass in Chromium, Firefox and Mobile Chrome. However the code base has a **material accumulation of silent-error paths, several unbounded data-access methods, one SSRF-sized hole in the outbound API-testing executor, and a WebKit rendering/test-stability problem** that blocks 50 of 230 Playwright tests.

Overall production readiness: **conditional — not ready for a high-trust, multi-tenant SaaS launch without remediation of the critical issues below.** The architecture and middleware are sound, but the current implementation tolerates too many ignored errors, has no hard guarantee that every transaction applies the tenant RLS context, and exposes internal services to potentially attacker-controlled URLs.

---

## Scope & Methodology

- Ran `go test ./... -count=1 -v` in `apps/api` and captured `apps/api/gotest_full.log`.
- Attempted `go test ./... -race`; failed because the toolchain reports `-race requires cgo` (Windows build without `CGO_ENABLED=1`).
- Ran Playwright E2E tests against Chromium, Firefox, WebKit and Mobile Chrome and captured logs in `tests/pw_*.log`.
- Ran `node scripts/check-openapi-drift.mjs` and captured output.
- Audited middleware (`auth`, `cookies`, `csrf`, `idempotency`, `ratelimit`, `rbac`, `redact`, `tenant`), server routing, per-module handlers/services/repositories, the queue/worker, and the Next.js frontend.
- Traced patterns for swallowed errors (`_ = ...`), unbounded queries, pagination, N+1 loops, race conditions, `TODO/FIXME/stub`, `localStorage` and `any` usage.

---

## Test Execution Summary

### Go unit tests

`go test ./... -count=1 -v` completed with **no failures**.

- Passing packages with test coverage: `apikeys`, `apitesting`, `automationhub`, `defects`, `identity`, `notification`, `project`, `results`, `shared/config`, `shared/http`, `shared/jwt`, `shared/middleware`, `shared/pagination`, `shared/password`, `shared/security`, `shared/tenant`, `shared/validation`, `testmanagement`.
- Packages with **no test files**: `cmd/*`, `analytics`, `audit`, `billing`, `integrationhub`, `intelligence`, `metrics`, `organization`, `queue`, `rbac`, `search`, `shared/db`, `shared/errors`, `shared/eventbus`, `shared/idempotency`, `shared/secrets`, `shared/server`, `workspace`.
- Notable warning in `shared/middleware` tests: Redis connection failure caused the rate limiter to fall back to the local in-memory limiter.
- **Race detector not executed** because the build environment does not have cgo enabled (`go: -race requires cgo; enable cgo by setting CGO_ENABLED=1`). This is a gap in verification.

### Playwright E2E

| Browser | Result |
|---------|--------|
| Chromium | **230 passed** in 23.7s |
| Firefox | **230 passed** in 38.4s |
| WebKit | **180 passed, 50 failed** |
| Mobile Chrome | **230 passed** in 16.2s |

- The 50 WebKit failures are concentrated in **accessibility, visual-regression and dashboard UI tests**. The common failure is `expect(locator).toBeVisible()` failing for the `Dashboard` heading, with `element(s) not found` / timeout 15s.
- All browsers show a **logged error** `secret "SMTP_PASSWORD" is not set` during password-reset tests; the code paths the error but the test still passes because the API endpoint succeeds even though the e-mail is not sent.
- No `.only` or `.skip` markers were found in test files.

### OpenAPI drift

`node scripts/check-openapi-drift.mjs` found the following **implemented-but-undocumented** routes:

- `GET /analytics/recent-activity`
- `PUT /projects/{id}`
- `DELETE /projects/{id}`

The script exits with code 1 and prints only these three. Other route groups may also have drift, but the current tooling does not surface them.

---

## Critical Findings

### 1. SSRF-sized hole in the API testing executor

File: `apps/api/internal/apitesting/service.go:640-740`

`executeRequest` parses a user-supplied URL, applies variables, sets headers and authentication, then calls `s.httpClient.Do`. It validates scheme and host presence but **never calls `security.ValidateURL`**. The repository has a capable SSRF validator (`apps/api/internal/shared/security/ssrf.go`) and it is used by integration providers and notification dispatch, but it is **not wired into the API testing executor**.

Implication: a workspace member can run an API request against `http://169.254.169.254/latest/meta-data/`, `http://localhost:5432`, or other internal metadata/services, depending on where the API server is deployed.

### 2. Tenant RLS context can be silently lost inside transactions

File: `apps/api/internal/shared/db/db.go:104-127`

```go
if tenantID, ok := TenantIDFromContext(ctx); ok {
    _ = SetLocalTenantID(ctx, tx, tenantID)
}
return tx, nil
```

`BeginTx` ignores the error from `SetLocalTenantID`. If the `SET LOCAL` statement fails (connection issue, syntax issue, policy mismatch), the transaction proceeds **without the tenant session variable set**. Because all repository methods then run on `tx` via the context, row-level security may see `app.tenant_id = NULL` and either deny all rows or, depending on policy, match a default tenant — a severe isolation risk.

The same pattern is repeated by repository `RunInTx` helpers that call `_ = db.SetLocalTenantID(...)`, and the error is not propagated back to the caller.

### 3. `SetSessionTenantID` / `SetLocalTenantID` / `SetLookupKeyHash` interpolate into SQL

File: `apps/api/internal/shared/db/db.go:52-68`

```go
func SetSessionTenantID(ctx context.Context, exec DBTX, tenantID uuid.UUID) error {
    _, err := exec.ExecContext(ctx, fmt.Sprintf("SET app.tenant_id = '%s'", tenantID.String()))
    return err
}
```

Tenant IDs are UUIDs so this is not currently exploitable, but `SetLookupKeyHash` does the same with an arbitrary string:

```go
func SetLookupKeyHash(ctx context.Context, exec DBTX, hash string) error {
    _, err := exec.ExecContext(ctx, fmt.Sprintf("SET app.lookup_key_hash = '%s'", hash))
    return err
}
```

The hash is a SHA-256 hex string in the code path, but the **pattern is brittle**: any future change that passes user-controlled data here is an immediate SQL injection. Use a placeholder mechanism or at minimum quote/escape.

### 4. Results repository `ListItems` is unbounded

File: `apps/api/internal/results/repository.go:150-159`

```go
func (r *SQLRepository) ListItems(ctx context.Context, runID uuid.UUID) ([]TestRunItem, error) {
    rows, err := r.db.QueryContext(ctx,
        `SELECT ... FROM test_run_items WHERE run_id = $1 ORDER BY sort_order ASC`, runID)
    ...
}
```

There is no `LIMIT`. The public `Service.ListItems` calls this method directly. The HTTP handler happens to override `ListItems` and call `ListItemsPaged` with pagination (`manual_handler.go:296-327`), so the UI is safe, but the service/repository API remains unbounded and can be called from workers, scripts or future endpoints.

### 5. Multiple ignored `RowsAffected` errors

File: `apps/api/internal/results/repository.go:111-177` and `manual_repository.go:145-284`

```go
rows, _ := result.RowsAffected()
if rows == 0 {
    return sharederrors.ErrNotFound
}
```

The error from `RowsAffected()` is discarded. If the driver returns an error (connection lost during exec, for example) the code may incorrectly return `ErrNotFound` instead of the real error. The same pattern appears in `results`, `defects`, `analytics`, `integrationhub`, `intelligence` and `apitesting` repositories, while some packages (`testmanagement`, `automationhub`, `apikeys`, `notification`) correctly check it.

### 6. Integration hub silently swallows repository update failures

File: `apps/api/internal/integrationhub/service.go:123-134` and `196-254`

```go
_ = s.repo.UpdateIntegration(ctx, i)
```

and

```go
_ = s.repo.UpdateEvent(ctx, e)
```

The service modifies the in-memory entity and then ignores the result of the persistence call. If the update fails, the audit log and in-memory state claim success, but the database is inconsistent.

### 7. API key validation silently ignores `UpdateLastUsed` failure

File: `apps/api/internal/apikeys/service.go:108-121`

```go
_ = s.repo.UpdateLastUsed(ctx, key.ID)
return key, nil
```

Failure to record key usage is silent; an unavailable table could cause the key to appear valid and usable while the audit trail is incomplete.

### 8. JSON unmarshal errors are widely ignored in repository scanning

File: `apps/api/internal/results/repository.go:217-262`, `apps/api/internal/analytics/repository.go:236`, `apps/api/internal/intelligence/repository.go:179`, `apps/api/internal/integrationhub/repository.go:196-245`, `apps/api/internal/integrationhub/providers.go:95-569`

```go
_ = json.Unmarshal([]byte(metaStr), &run.Metadata)
```

If a row contains malformed JSON, the repository silently returns a partially-initialised struct with a zero/default value and no error. This can lead to data loss on the next update (a corrupted JSON column gets overwritten with `null` or `{}`).

### 9. Analytics repository ignores `QueryRowContext` errors

File: `apps/api/internal/analytics/repository.go` (scanning `scanDashboard` and `GetMetrics` paths)

Several paths read JSON columns with:

```go
_ = json.Unmarshal([]byte(configStr), &d.Config)
```

Additionally, summary/metric paths call `r.db.QueryRowContext` and `rows.Scan` but the scan errors are not always returned correctly; in the dashboard scan the JSON unmarshal errors are dropped and the function continues, returning dashboards with `Config: nil` instead of an error.

### 10. Billing worker job is a no-op; invoice import is not idempotent

File: `apps/api/cmd/worker/main.go:389-401`

```go
case "billing:sync":
    ...
    _, err = r.billing.Service.GetSubscription(ctx, orgID)
    return err
```

The worker’s billing job merely **reads** the subscription. It does not persist any sync result, enqueue a state change, or write back to Stripe.

File: `apps/api/internal/billing/service.go:88-116`

`ListInvoices` fetches provider invoices and, for each one, calls `repo.CreateInvoice` with a **new random UUID** and the current timestamp. If the job is retried or runs again before the retention cleanup, the same Stripe invoice will be inserted multiple times with different IDs.

### 11. Playwright WebKit failures indicate real rendering/test instability

File: `tests/pw_webkit.log`

50 tests fail only in WebKit. They are not API failures; they are UI tests that cannot find the `Dashboard` heading. Possible root causes include:

- WebKit-specific cookie handling or `Secure` cookie semantics in local HTTP testing.
- A hydration/timing issue in the Next.js dashboard (the heading may not render within 15 s in WebKit).
- The visual-regression test expecting a fixed viewport layout that WebKit renders differently.

This is not merely a test flake; it suggests the WebKit user journey is broken and must be investigated before shipping to Safari users.

---

## Per-Module Audit

### Identity / Authentication

- **Strengths:** JWT access/refresh rotation, password policy, token refresh with family ID, CSRF double-submit cookies, secure cookie flags in `cookies.go`.
- **Weaknesses:**
  - Password-reset e-mail fails silently when `SMTP_PASSWORD` secret is missing (`gotest_full.log` and Playwright logs). The handler returns success even though the e-mail was never sent.
  - `login` and `register` routes skip CSRF, which is correct, but the `ForgotPassword` path should also be verified.

### Workspace / Organization / Project

- **Strengths:** Tenant resolver resolves org from workspace/project IDs.
- **Weaknesses:**
  - `PUT /projects/{id}` and `DELETE /projects/{id}` are implemented but not documented in OpenAPI.
  - No unit test files for `workspace`, `organization`, `project` packages.

### Test Management

- **Status:** strong.
- Correctly handles `RowsAffected` errors, validates steps, creates versions on update, and uses cursor pagination in list methods.

### Results / Test Runs

- **Strengths:** Handler override provides pagination for `ListItems`.
- **Weaknesses:**
  - `results/repository.go` `ListItems` is unbounded and `Service.ListItems` still calls it.
  - Update methods ignore `RowsAffected` errors, potentially misreporting not-found.
  - `UpdateItem` silently `json.Marshal` failures with `artifactsJSON, _ := json.Marshal(...)` (line 162-163) and does not return them; a malformed struct would be stored as `null`.

### Defects

- **Status:** standard CRUD with correct tenant resolution.
- **Weakness:** `RowsAffected` errors ignored in `repository.go`.

### Automation Hub

- **Strengths:** JUnit/Xray/Playwright artifact import, artifact upload, execution rerun.
- **Weakness:** `RowsAffected` correctly checked in `automationhub/repository.go`.

### API Testing

- **Strengths:** Cursor pagination, environment variables, request/folder/collection CRUD, execution history.
- **Weaknesses:**
  - **No SSRF validation on outbound URLs** (see Critical Finding #1).
  - `UpdateRequest`/`DeleteRequest` `RowsAffected` errors are checked, but the `Update` method in `apitesting/repository.go` does not.
  - Variable substitution on headers/auth could leak secrets into outbound headers if not carefully audited, but that is expected behaviour for an API client.

### Notifications

- **Strengths:**
  - Channel config secrets are masked in responses (`maskChannelSecrets` in `notification/handler.go:105-119`).
  - `RowsAffected` is correctly checked.
  - Pagination and unread count are implemented.
- **Weaknesses:**
  - `notification/service.go` dispatch path uses `urlValidator` (SSRF) for webhook URLs, which is good.
  - No unit test coverage for the `notification` package beyond integration.

### Analytics

- **Strengths:** Dashboard CRUD, trends, summary, CSV export, metrics aggregation.
- **Weaknesses:**
  - `GET /analytics/recent-activity` is implemented but **not in OpenAPI**.
  - Heavy queries (recent activity, export, top-failed lists) do not consistently limit result sets or paginate; `GetMetrics` can return very large JSON.
  - JSON config unmarshalling is ignored (`_ = json.Unmarshal` in `scanDashboard`).
  - No unit tests.

### Search

- **Strengths:** Global search across entities with `ILIKE` + full-text, scoped to workspace, permission `workspaces:read`.
- **Weaknesses:**
  - `Get` at `server.go:1264` does **not** wrap `IdempotencyKey` middleware. This is acceptable for a GET, but the route is also the only non-mutating global search; ensure this is intentional.
  - No unit tests.

### Intelligence

- **Status:** wired but effectively stubbed.
- `intelligence/mlclient.go` uses a real HTTP client when `MLServiceURL` is set, otherwise falls back to a deterministic heuristic (`localMLClient`).
- The heuristic is useful for tests but means **no real ML is used in production unless configured**.
- `RowsAffected` is ignored in `intelligence/repository.go:95`.
- No unit tests.

### Integration Hub

- **Strengths:** Provider abstraction, webhook signature verification for GitHub/GitLab, retry with dead-letter queue.
- **Weaknesses:**
  - Multiple `_ = s.repo.UpdateIntegration` / `_ = s.repo.UpdateEvent` calls (see Critical Finding #6).
  - JSON responses from provider APIs are unmarshalled with `_ = json.Unmarshal` in `providers.go` lines 95, 207, 306, 407, 506, 569. Failure to parse a provider response returns an empty string or partial data instead of an error.
  - `integrationhub/repository.go` also ignores JSON unmarshalling for `Config` and `Payload`.

### Billing

- **Status:** wired but stub-like.
- `billing/provider.go` calls Stripe with `io.ReadAll(resp.Body)` (no size limit) and returns raw Stripe error bodies.
- `billing:sync` worker job is a read-only no-op (see Critical Finding #10).
- `ListInvoices` creates duplicate rows on each sync.
- No unit tests.

### Queue / Worker

- **Strengths:**
  - `queue.DequeueOne` uses `FOR UPDATE SKIP LOCKED` for safe concurrent workers.
  - Worker applies exponential backoff, job retention cleanup, and `SetLocalTenantID` in `processOne`.
  - Handles commit/rollback and marks jobs as `dead_letter` after max attempts.
- **Weaknesses:**
  - `worker` app directory (`apps/worker`) is essentially empty; the real worker binary is in `apps/api/cmd/worker/main.go`.
  - `queue.Enqueue` correctly sets `app.tenant_id`, but `DequeueOne` returns a transaction without tenant context; the worker has to set it. This is acceptable because the worker does set it, but any other consumer that forgets to will bypass RLS.
  - `billing:sync` and `intelligence:predict` jobs are present but do not drive durable state changes beyond side effects.

### Middleware

#### Tenant

- **Strengths:** Acquires a dedicated DB connection per request, sets `app.tenant_id`, resets `app.lookup_user_id`, resolves org from many resource types.
- **Weaknesses:**
  - The `release` function ignores errors from `RESET` and `Close` (`_, _ = conn.ExecContext(...)` and `_ = conn.Close()`). These are unlikely to fail, but the pattern is inconsistent with the critical nature of tenant reset.

#### RBAC

- **Strengths:** Loads permissions per user and tenant, caches in request context, returns proper 403.
- **Weakness:** Uses `err == sql.ErrNoRows` instead of `errors.Is`; brittle if wrapped.

#### Idempotency

- **Strengths:** Stores request body fingerprint, tenant- and workspace-scoped, replays cached responses, uses TTL.
- **Weakness:** `IdempotencyKey` middleware is applied to **all** route groups including GETs in some groups, but the middleware itself only acts on mutating methods, so it is safe but adds overhead.

#### Rate Limit

- **Strengths:** Pluggable, supports IP/email/API-key, in-memory fallback, configurable fail-open/fail-closed.
- **Weakness:** Redis failure silently falls back to in-memory per-process limiter, which does not share state across pods.

#### CSRF

- **Strengths:** Double-submit cookie, constant-time comparison (`CSRFTokenEqual`), safe methods skipped, configurable skip list.
- **Weakness:** None found.

#### Cookies

- **Strengths:** HttpOnly, Secure, SameSite, constant-time token comparison.
- **Weakness:** Cookie `SameSite` defaults to `Lax` in the config loader; should be `Strict` for a SaaS admin tool unless cross-site POST flows are required.

#### Auth

- **Strengths:** Accepts `Authorization: Bearer` or access cookie, validates JWT, sets user context.
- **Weakness:** None found.

#### Redact

- **Strengths:** Masks tokens, API keys, passwords, emails, and IP addresses in logs.
- **Weakness:** None found.

---

## Frontend Audit (`apps/web`)

### Security headers and CSP

`apps/web/middleware.ts` sets HSTS, CSP, X-Frame-Options, Referrer-Policy and Permissions-Policy. CSP includes `connect-src 'self' ${apiOrigin}`, so the browser is only allowed to reach the configured API.

### API client

`apps/web/lib/api.ts`:
- Refreshes tokens on 401 and retries once.
- Redirects to `/login` on 401 after refresh failure.
- Caches `X-CSRF-Token` and includes it on mutating requests.
- Does **not** store access tokens in `localStorage`; tokens are kept in HttpOnly cookies.

### `localStorage` usage

Workspace, project and organization IDs are stored in `localStorage` in many pages and providers:

- `components/providers/workspace-provider.tsx`
- `components/workspace/create-workspace-form.tsx`
- `app/(dashboard)/[workspace]/*/page.tsx`
- `features/notifications/api.ts`

These are **not authentication tokens**, but if an XSS vector ever exists, `localStorage` is an easier target than `HttpOnly` cookies. This is a medium-risk defence-in-depth issue, not a blocker.

### `any` usage

A source-level `any` appears in `features/analytics/components/DashboardAnalytics.tsx:258` (`data: any[]`). All other matches were generated files under `.next/types` and can be ignored.

### XSS / dangerous APIs

No occurrences of `dangerouslySetInnerHTML` or `eval()` were found in `.ts/.tsx` source.

### Loading / error states

Most server/client pages set `loading` and `error` states. `workspace-provider.tsx` has a full loading/error flow and redirects to `/create-workspace` when no workspace exists. No obvious unhandled loading skeletons found.

### Links and routing

Next.js `Link` is used consistently. No plain `<a href>` to external sites without validation observed, but this audit did not inspect every component.

---

## Bug Hunt Details

### Swallowed errors

The following patterns were found by grep in the Go code base:

| Pattern | Count | Severity | Notes |
|---------|-------|----------|-------|
| `_ = json.Unmarshal` | 12+ | Medium | Repository scans, integration provider responses, analytics dashboard |
| `_ = s.repo.Update(...)` (integration/apikeys) | 6+ | High | Integration event/integration updates, API key last-used |
| `_ = SetLocalTenantID(...)` | Multiple | High | `db.BeginTx` and repository `RunInTx` helpers |
| `rows, _ := result.RowsAffected()` | 15+ | Medium | results, defects, analytics, integrationhub, intelligence, apitesting |
| `_ = tx.Rollback()` in queue | 2 | Low | Generally acceptable on dequeue error; error already returned |

### N+1 / missing pagination

| Location | Problem |
|----------|---------|
| `results/repository.go:150` | `ListItems` has no `LIMIT` |
| `results/manual_repository.go:21` | `ListItemsByRunPaged` builds query but does not append `LIMIT` in the snippet inspected; confirm in production |
| `analytics/repository.go` | Summary/recent-activity/top-failed queries can return large result sets |
| `workspace-provider.tsx:57` | Loads all workspaces for all organizations in a serial loop (`for (const org of orgs) { ... }`) — N+1 on the frontend |

### Race conditions

- **Not verified:** race detector could not run due to missing cgo.
- Potential races:
  - `shared/idempotency` in-memory store is not inspected; if it is a map without a mutex it will race under concurrent mutating requests.
  - Token refresh with rotation: two parallel 401s may race the refresh endpoint; the API tests cover this and pass, but a real distributed deployment could see window issues.

### Queue / worker

- Worker processes `notification:send`, `analytics:aggregate`, `intelligence:predict`, `integration:dispatch`, `integration:retry`, `billing:sync`.
- `billing:sync` is a no-op (see above).
- `integration:retry` calls `RetryEvent` with `uuid.Nil` user, but the service uses the event’s user from the DB; this is acceptable.
- Job `max_attempts` is read from the row; a default is not enforced on insert.

---

## Wired-but-Stub Modules

| Module | Evidence | Verdict |
|--------|----------|---------|
| `billing` | No-op `billing:sync` worker, duplicate invoice inserts, Stripe price ID loaded but not used | **Stub** — present but not production-grade |
| `intelligence` | Falls back to heuristic when no `MLServiceURL`; predictions are keyword-based | **Stub** unless external ML service is configured |
| `worker` | `apps/worker` directory empty; real worker in `apps/api/cmd/worker` | **Present but split confusingly** |
| `search` | Search works via SQL, but no dedicated tests and heavy ILIKE queries | **Operational but unproven at scale** |
| `analytics` | Dashboards/CSV/trends exist, but no unit tests and `recent-activity` undocumented | **Functional but needs hardening** |

---

## OpenAPI Drift

`scripts/check-openapi-drift.mjs` found:

- `GET /analytics/recent-activity`
- `PUT /projects/{id}`
- `DELETE /projects/{id}`

These routes are implemented in `server.go` but missing from `docs/api/openapi/openapi.yaml` (or wherever the script sources the OpenAPI spec). The script should also be reviewed — it stops after the first set of undocumented routes and may not report all drift.

---

## Production Readiness Assessment

| Area | Grade | Reason |
|------|-------|--------|
| Architecture | A | Clean modular monolith, chi router, middleware pipeline, RLS, event bus, queue |
| Authentication / session | A- | JWT + refresh rotation, CSRF, secure cookies; password-reset e-mail depends on secret config |
| Multi-tenancy | B+ | Tenant context middleware is comprehensive, but ignored `SetLocalTenantID` errors are a critical gap |
| RBAC / audit | B+ | Audit logging and RBAC on mutating routes; not all GET routes are audited |
| Input validation | B | Strong in testmanagement/identity; weak for SSRF in API testing and some JSON unmarshal paths |
| API surface consistency | C | OpenAPI drift and `recent-activity` undocumented |
| Testing | B | Unit tests pass; many packages have no tests; race detector could not run; WebKit E2E fails |
| Queue / worker | B | Solid dequeue locking; `billing:sync` and some jobs are stubs; no worker unit tests |
| Frontend security | B+ | CSP, CSRF, no tokens in localStorage; workspace IDs in localStorage are a minor concern |
| Observability | B | Metrics endpoint in worker, audit logs, structured logging; some silent error paths hide root cause |

---

## Recommendations (prioritised)

### P0 — fix before production

1. **Call `security.ValidateURL` in `apitesting/service.go` `executeRequest`** before issuing `httpClient.Do`. This closes the SSRF hole.
2. **Return / handle the error from `SetLocalTenantID` and `SetSessionTenantID` in every `BeginTx` and `RunInTx` helper.** If it fails, roll back and return 500. This is a multi-tenancy isolation guarantee.
3. **Fix the ignored `RowsAffected` and `json.Unmarshal` errors** in `results`, `analytics`, `integrationhub`, `intelligence`, `apitesting`, `defects` and `apikeys`. Use the same pattern as `testmanagement` and `notification`.
4. **Investigate the 50 WebKit Playwright failures.** Fix or quarantine the tests; if it is a real WebKit/Safari bug, fix the frontend.

### P1 — high value, low risk

5. **Add `LIMIT`/`OFFSET` or cursor pagination to unbounded list methods** (`results/repository.go:ListItems`, `analytics` summary/activity, `search`).
6. **Make `billing:sync` meaningful or remove it**; make `ListInvoices` idempotent by upserting on `provider_invoice_id`.
7. **Resolve the 3 OpenAPI drift items** and extend the drift checker to produce a full diff.
8. **Run race tests** by enabling `CGO_ENABLED=1` and a C compiler, or migrate to Linux CI where the race detector works.

### P2 — hardening

9. **Move workspace/project IDs out of `localStorage`** into a short-lived cookie or URL segment where XSS cannot easily access them.
10. **Replace `fmt.Sprintf` SQL for session variables** with a safe escaping/parameterised mechanism for `app.tenant_id` and `app.lookup_key_hash`.
11. **Add unit tests for untested packages:** `analytics`, `billing`, `integrationhub`, `intelligence`, `search`, `organization`, `workspace`, `queue`, `rbac`.
12. **Add alerts for Redis and SMTP secret misconfiguration** instead of relying on log lines.

---

## Commands for Reproduction

```bash
# Go unit tests
cd apps/api
go test ./... -count=1 -v 2>&1 | tee gotest_full.log

# Race tests (requires cgo)
CGO_ENABLED=1 go test ./... -race -count=1

# OpenAPI drift
cd c:\Private\project\testra
node scripts/check-openapi-drift.mjs

# Playwright
cd tests
npx playwright test --project=chromium
npx playwright test --project=firefox
npx playwright test --project=webkit
npx playwright test --project="Mobile Chrome"
```

---

## Conclusion

Testra has a strong architectural foundation and passes the majority of automated checks. The main blockers are **silent error handling in multi-tenant transaction setup, the SSRF-sized omission in the API testing executor, and the WebKit-specific UI failures.** Remediating the P0 items will materially improve reliability and security and bring the project much closer to production readiness.

*Report generated from the current audit session.*
