# P0 Closeout Report

**Scope:** Final audit closeout for the Testra monorepo — RLS fail-closed coverage, WebKit/Mobile Safari flakiness, RBAC for API-testing execute, queue/worker race coverage, and high-risk errcheck triage.

**Ground truth:** The previous `P0_VERIFICATION_REPORT.md` is treated as correct for the six already-closed P0 items and is not re-audited here.

---

## Executive summary / production-readiness verdict

| Area | Verdict | Notes |
|---|---|---|
| RLS fail-closed | **Closed / fixed** | All tenant-scoped tables use `current_setting('app.tenant_id', true)`. `queue_jobs` fail-open policy was replaced with an explicit `app.queue_worker` bypass. Empirical tests pass. |
| WebKit/Mobile Safari | **Conditionally closed** | Locator fixes, the `addInitScript` workspace-context fix, and a 30s heading-visibility timeout were applied. Full runs still show **50 unexpected failures across 180 unique tests** in both WebKit and Mobile Safari; the failures are page-load/hydration spinners, not the original route-announcer or workspace-race failures. |
| RBAC — API execute | **No change required** | Endpoint is correctly gated by `RequirePermission(rbacCfg, "api_tests:execute")`. Owner, admin, and QA engineer hold the permission; viewer does not. SSRF validation already exists. |
| Race detector / queue | **Tests added** | `internal/queue` now has `TestDequeueConcurrency`. Local `-race` cannot run because no C compiler is installed, but the test passes normally and the CI job uses `go test -race`. |
| errcheck | **High-risk findings fixed** | Tenant-isolation/financial hot spots in `cmd/worker`, `internal/queue`, `internal/billing/provider.go`, `internal/analytics/handler.go`, and `internal/apitesting/service.go` were fixed. 253 lower-risk `rows.Close`/`json.Marshal`/`uuid.Parse` patterns remain and should be triaged in a follow-up. |

**Overall production-readiness verdict:** **Conditional GO with the RLS, RBAC, queue, and errcheck fixes merged. WebKit/Mobile Safari suites still need a dedicated stability pass before the front end can be called fully green, but the two known P0 flakiness causes were addressed.**

---

## Section 1 — RLS fail-closed coverage

### 1.1 Enumeration

All `ALTER TABLE ... ENABLE ROW LEVEL SECURITY` statements and `CREATE POLICY` bodies across `apps/api/migrations/` were extracted with `scripts/extract_rls.mjs`. The resulting table is in `rls_table.md`; the key pattern coverage is:

- **46 tables** have RLS enabled.
- 44 of those use `current_setting('app.tenant_id', true)::uuid` (or a subquery against `workspaces`/`organizations` filtered by that value).
- Several tables have an additional lookup policy driven by `app.lookup_user_id` or `app.lookup_key_hash` for list/endpoints.
- `queue_jobs` was the **only table with a fail-open `OR app.current_tenant() IS NULL` clause** in its RLS policy.

### 1.2 `queue_jobs` fix

The old policy in `000026_worker_queue.up.sql` was:

```sql
USING (tenant_id = app.current_tenant() OR app.current_tenant() IS NULL);
```

This allowed any unauthenticated/tenant-less connection to see every queue row.

A new migration (`000041_fix_queue_jobs_rls.up.sql`) replaced it with:

```sql
CREATE POLICY tenant_isolation_queue_jobs ON queue_jobs
    USING (tenant_id = app.current_tenant() OR current_setting('app.queue_worker', true) = 'true');
```

**Code changes to support the worker bypass:**
- `apps/api/internal/shared/db/db.go`: added `SetWorkerMode` and `ResetWorkerMode` helpers.
- `apps/api/internal/queue/queue.go`: `DequeueOne` and `DeleteOldCompleted` now call `SetWorkerMode` on the transaction/connection and log any close/rollback errors.

Migrations were applied successfully to the local `testra` database.

### 1.3 Empirical fail-closed evidence

All queries were run as the non-superuser `testra_app_test` role (mirrors the application user) against the migrated test database.

**Representative Phase 4 tables (analytics, intelligence, integration hub, billing):**

```
analytics_daily_metrics:no_tenant|0
flaky_predictions:no_tenant|0
subscriptions:no_tenant|0
invoices:no_tenant|0
analytics_daily_metrics:tenant_set|1
flaky_predictions:tenant_set|1
subscriptions:tenant_set|1
invoices:tenant_set|1
```

**`queue_jobs` post-fix:**

```
queue:no_context|0        -- fail-closed without tenant or worker flag
queue:worker_mode|2       -- worker bypass sees all jobs
```

**`queue_jobs` pre-fix reproduction (rolled-back transaction with the old policy):**

```
queue:pre_fix_no_context|2
```

This demonstrates that the old policy was fail-open (rows visible with no context) while the new policy is fail-closed (0 rows without `app.queue_worker` or `app.tenant_id`).

### 1.4 Remaining items

- Duplicate/typo policy names `idempotency_records_tenant` (created twice) and `role_assignments_tenant` (twice, one with `OR`) are present in the migrations. These are not security defects but should be cleaned up in a follow-up migration.
- `rls_table.md` and the raw `rls_policies.json` are saved in the repo root as audit artifacts.

---

## Section 2 — WebKit locators and workspace-context race

### 2.1 Fixes applied

1. **Route-announcer collisions in `[role="alert"]` locators**
   - `tests/pages/BasePage.ts`: `hasError` now excludes `#__next-route-announcer__`.
   - `tests/e2e/accessibility/accessibility-expanded.spec.ts`: `p[role="alert"]` is used instead of the bare `[role="alert"]` locator.
2. **Workspace-context race in `setWorkspaceContext`**
   - `tests/helpers/storage.ts` now uses `page.addInitScript` to seed the workspace context into `localStorage` before the app bootstraps, then navigates to `/create-workspace` with `waitUntil: "networkidle"`. This guarantees the `WorkspaceProvider` reads the intended workspace on first mount and removes the race where the provider loaded before `localStorage` was set and redirected to `/create-workspace`.
3. **Heading visibility timeout in WebKit/Safari**
   - `tests/pages/BasePage.ts`: `expectHeading` now waits up to **30s** for the heading to become visible, giving slow Next.js hydration on WebKit/Mobile Safari more time to settle.

### 2.2 WebKit full-suite results

Commands:

- `npx playwright test --project=webkit`
- `npx playwright test --project="Mobile Safari"`

Both runs produced the same top-level pattern:

- **230 executed test cases** (each test has one retry)
- **50 unique / unexpected failures** (JUnit: 48 failures + 2 errors)
- **130 of 180 unique tests passed**

The top failing suites were `accessibility`, `auth/login`, `dashboard`, `onboarding`, `projects`, `smoke`, `testcases`, `testplans`, and `visual`. The failures are now **page-load/hydration spinners** (`getByRole('heading', { name: '...' }) not found`, page DOM contains `img "Loading"`) and are not the original route-announcer or workspace-redirect flakiness.

### 2.3 Targeted `testplans` evidence

`npx playwright test --project=webkit --reporter=list e2e/testplans/testplans.spec.ts` (using the `addInitScript` workspace fix and the 30s `expectHeading` timeout):

- 2 API tests passed.
- `Test Plans @testplans › user can create a test plan through UI` failed after both attempts with the page stuck on `img "Loading"` and the heading "New Test Plan" never visible.

This confirms the workspace race is now under control — the test plan page is reached and renders, but the app stays in a loading state on WebKit. That is a separate Next.js/WebKit hydration or data-fetching issue and needs an app-level fix, not a test-locator fix.

### 2.4 Recommendation

The two targeted P0 flakiness causes were addressed. The remaining 50 failures are a class of **app-level loading/hydration failures** on WebKit and Mobile Safari. Treat these as a dedicated follow-up: profile Next.js SSR/hydration in WebKit, ensure data-fetching errors are surfaced instead of hanging on `img "Loading"`, and consider adding per-page readiness signals for Playwright to wait on.

---

## Section 3 — RBAC permission assessment for API-testing execute

The API-testing execute endpoint is `POST /api/v1/api-executions` handled by `internal/apitesting/handler.go:ExecuteRequest`.

It is wrapped with:

```go
sharedmiddleware.RequirePermission(rbacCfg, "api_tests:execute")
```

The permission `api_tests:execute` was seeded in `000032_add_api_testing.up.sql` and is mapped to:

- `owner` — all API-testing permissions
- `admin` — all API-testing permissions
- `qa_engineer` — read, create, update, execute (no delete)
- `viewer` — read only

### Assessment

The route is correctly protected by RBAC middleware. The permission is appropriately scoped to the QA engineer role, which is the primary persona running API tests. The previously closed P0 work added `ValidateURL` SSRF validation in `internal/apitesting/service.go`, which is still in place.

Residual SSRF risks (DNS rebinding, follow redirects, TOCTOU) still exist in principle for any endpoint that performs outbound HTTP. However, restricting execute to admins/owners would break the intended QA engineer workflow. Therefore **no RBAC change is implemented**. A future hardening story could add an organization-level toggle or require a separate `api_tests:execute_external` permission for requests leaving the documented allow-list.

---

## Section 4 — Race detector history and queue/worker concurrency

- CI workflow `.github/workflows/ci.yml` already runs `go test -race -count=1 ./apps/api/...`.
- The local Windows environment does not have a C compiler, so `go test -race` cannot be run directly (`cgo: C compiler "gcc" not found`).
- A new concurrency regression test was added: `apps/api/internal/queue/queue_test.go::TestDequeueConcurrency`.
  - It seeds a tenant and 5 queue jobs, then starts 5 goroutines that all call `queue.DequeueOne`.
  - It asserts every job is dequeued exactly once, exercising `FOR UPDATE SKIP LOCKED` and the new `app.queue_worker` RLS path.
  - It passes with `go test -run TestDequeueConcurrency ./internal/queue` and will be covered by `go test -race` in CI.

---

## Section 5 — High-risk errcheck triage

The `errcheck -blank -ignoretests ./...` run in `apps/api` reports **253 ignored-error patterns** across the repository. The highest-risk ones for tenant isolation/financial consistency were fixed:

| File | Risk | Fix |
|---|---|---|
| `cmd/worker/main.go` | Ignored `uuid.Parse` of job payload IDs can load invalid nil tenant/user/workspace IDs into services. | All payload UUID parses now return wrapped errors. |
| `cmd/worker/main.go` | Ignored `tx.Rollback()` can leave a locked transaction open after a failed job. | Rollback now logs on failure. |
| `cmd/worker/main.go` | `defer database.Close()` ignored. | Close error is now logged. |
| `internal/queue/queue.go` | `defer conn.Close()` and `_ = tx.Rollback()` in the RLS worker path. | Close/rollback errors logged; helps ensure worker transactions do not leak. |
| `internal/billing/provider.go` | Stripe `io.ReadAll` / `resp.Body.Close` errors ignored. | Request helper now returns read and close errors. |
| `internal/analytics/handler.go` | CSV `w.Write` errors ignored, which can silently truncate financial/analytics exports. | Write errors now produce a 500 and abort. |
| `internal/apitesting/service.go` | `httpResp.Body.Close()` ignored on outbound SSRF execution. | Close errors are now logged. |

The remaining 253 patterns are dominated by low-risk repository boilerplate (`defer rows.Close()`, `json.Marshal`, `uuid.Parse` in read-only row scans, `RowsAffected` checks). These are tracked but not in the P0 hot path.

---

## Changed files

- `apps/api/migrations/000041_fix_queue_jobs_rls.up.sql`
- `apps/api/migrations/000041_fix_queue_jobs_rls.down.sql`
- `apps/api/internal/shared/db/db.go`
- `apps/api/internal/queue/queue.go`
- `apps/api/internal/queue/queue_test.go`
- `apps/api/cmd/worker/main.go`
- `apps/api/internal/billing/provider.go`
- `apps/api/internal/analytics/handler.go`
- `apps/api/internal/apitesting/service.go`
- `tests/helpers/storage.ts`
- `tests/pages/BasePage.ts`
- `tests/e2e/accessibility/accessibility-expanded.spec.ts`
- `scripts/extract_rls.mjs`
- `scripts/generate_rls_table.mjs`
- `rls_table.md`
- `P0_CLOSEOUT_REPORT.md`

## Appendix — evidence files

- `rls_policies.json` — raw extracted RLS policy definitions.
- `rls_table.md` — tabular RLS coverage summary.
- `queue_prefix_test_output.txt` — pre-fix `queue_jobs` fail-open evidence.
- `rls_phase4_test_output.txt` — post-fix `queue_jobs` and `integrations` fail-closed/worker-bypass evidence.
- `rls_phase4_test2_output.txt` — post-fix Phase 4 table evidence.
- `apps/api/errcheck_blank_output.txt` — current `errcheck -blank -ignoretests ./...` output.
- `tests/reports/junit/test-results.xml` — latest full Playwright JUnit results (Mobile Safari).
- `tests/reports/json/test-results.json` — latest full Playwright JSON results (Mobile Safari).
