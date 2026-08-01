# P0 Audit Fix & Verification Report

This report resolves the open contradictions and verifies the P0 fixes from the previous Testra monorepo audit.

---

## 1. RLS fail-mode when `app.tenant_id` is unset

**Status:** Confirmed  
**Severity:** High (tenant-isolation behaviour)

- Migration file `apps/api/migrations/000009_add_rls_policies.up.sql` uses `current_setting('app.tenant_id', true)::uuid`. The `true` argument means the setting returns `NULL` when unset, which fails all RLS predicates.
- Empirical reproduction was run directly against PostgreSQL:
  - `rls_select_no_tenant.log` — `SELECT` returned 0 rows when `app.tenant_id` was not set.
  - `rls_select_tenant1.log` / `rls_select_tenant2.log` — rows were visible only when `app.tenant_id` matched the row's `organization_id`.
  - `rls_select_tx_no_local.log` — even inside a transaction, `SET LOCAL app.tenant_id` is required; without it, zero rows are returned.

**Conclusion:** RLS enforces a fail-closed model. Missing or mismatched tenant context causes zero access, not a bypass.

---

## 2. SSRF protection in API testing outbound executor

**Status:** Fixed and tested  
**Severity:** Critical

### Fix
- `apps/api/internal/apitesting/service.go` now calls `security.ValidateURL` inside `runHTTPRequest` before `s.httpClient.Do`:

```go
if err := security.ValidateURL(ctx, u.String()); err != nil {
    return nil, fmt.Errorf("%w: %s", sharederrors.ErrInvalidInput, err)
}
```

### Tests
- New test `TestServiceExecuteRequestRejectsSSRF` in `apps/api/internal/apitesting/service_test.go` uses a `failIfUsed` `http.RoundTripper` to prove that a link-local metadata URL (`http://169.254.169.254/latest/meta-data/iam/security-credentials/`) is rejected **before** any network call.
- The test was verified to fail on the pre-fix code and pass on the post-fix code.
- Existing tests that use local handlers were updated to use a `handlerRoundTripper` with a public-looking URL, so they still pass while SSRF validation remains active.

### Full suite
```
go test ./... -count=1
```
Result: all packages pass.

---

## 3. Tenant context propagation in `BeginTx` / `RunInTx`

**Status:** Fixed and tested  
**Severity:** High

### Fix
Every transaction helper now checks the error returned by `db.SetLocalTenantID` and rolls back on failure:

- `apps/api/internal/shared/db/db.go` (`BeginTx`)
- `apps/api/internal/results/repository.go` (`RunInTx`)
- `apps/api/internal/testmanagement/repository.go` (`RunInTx`)
- `apps/api/internal/analytics/repository.go` (`RunInTx`)
- `apps/api/internal/apitesting/repository.go` (`RunInTx`)
- `apps/api/internal/automationhub/repository.go` (`RunInTx`)
- `apps/api/internal/billing/repository.go` (`RunInTx`)
- `apps/api/internal/integrationhub/repository.go` (`RunInTx`)
- `apps/api/internal/intelligence/repository.go` (`RunInTx`)
- `apps/api/internal/results/manual_repository.go` (`withTenantInTx`)

The pattern is:

```go
if tenantID, ok := db.TenantIDFromContext(ctx); ok {
    if err := db.SetLocalTenantID(ctx, tx, tenantID); err != nil {
        _ = tx.Rollback()
        return nil, fmt.Errorf("set tenant context: %w", err)
    }
}
```

### Tests
- Added `apps/api/internal/shared/db/db_test.go` with `TestBeginTx_RollsBackWhenSetLocalTenantIDFails`.
  - Pre-fix: the test failed because `BeginTx` returned a non-nil transaction and a nil error.
  - Post-fix: it passes, and `sqlmock` verifies `Rollback()` is called.

---

## 4. WebKit Playwright failure root cause

**Status:** Reconciled  
**Severity:** Medium (test-infrastructure / flakiness)

### Reproduction
Ran targeted WebKit tests with `--trace on` and `DEBUG=pw:api`:

- `e2e/accessibility/accessibility-expanded.spec.ts:72` — **real failure evidence**:
  ```
  Error: strict mode violation: locator('[role="alert"]') resolved to 2 elements:
    1) <p role="alert" class="text-sm text-red-600">invalid credentials</p>
    2) <div role="alert" aria-live="assertive" id="__next-route-announcer__">...</div>
  ```
- `e2e/testplans/testplans.spec.ts:6` (full run) — navigation to `/dashboard/test-plans/new` was interrupted by a redirect to `/create-workspace`, meaning workspace context was not established in time.
- `e2e/visual/visual-regression.spec.ts` — six `toHaveScreenshot` failures in the full suite, but the same test passes in isolation.

### Diagnosis
- **Not SSL/chunk loading.** Network traces show HTTP/1.1 200 responses, `load`/`domcontentloaded` events fire, and all page elements resolve correctly.
- The actual cause is a combination of:
  1. **Test selectors colliding with Next.js** — the Next.js route announcer adds a second `role="alert"` element in WebKit, breaking strict single-match assertions.
  2. **Workspace-context timing** — tests rely on `localStorage` being set via `setWorkspaceContext` before the application redirects. Under parallel full-suite load this is flaky, causing the `/create-workspace` redirect.
  3. **Visual flakiness** — screenshot comparison is affected by the same workspace-context race and parallel worker state.

### Classification
Real product behaviour (Next.js announcer) + test-infrastructure timing issue, not a rendering/SSL bug.

---

## 5. OpenAPI drift

**Status:** Fixed and synchronized  
**Severity:** Low

### Script fix
- `scripts/check-openapi-drift.mjs` previously exited after printing only the first batch of drift.
- Updated to compare **both** directions:
  - `missing` — implemented but not documented
  - `surplus` — documented but not implemented
- It now prints both lists and only exits `1` when all output is complete.

### Diff found
Original run showed:

```
Implemented but not documented:
  DELETE /projects/{id}
  GET /analytics/recent-activity
  PUT /projects/{id}

Documented but not implemented:
  GET /analytics/activity
```

### Resolution
- Added `PUT /projects/{id}` and `DELETE /projects/{id}` to `docs/api/openapi/openapi.yaml`.
- Renamed the documented path from `/analytics/activity` to `/analytics/recent-activity` to match the actual server route.

### Verification
```
node scripts/check-openapi-drift.mjs
# OpenAPI is synchronized with the chi router (168 routes checked).
```

---

## 6. Token-storage location

**Status:** Reconciled  
**Severity:** None (current design is correct)

### Findings
- **Auth tokens are stored in HttpOnly cookies**, not `localStorage`.
- `apps/api/internal/shared/middleware/cookies.go`:
  ```go
  func SetAccessTokenCookie(...)  { ... newCookie(..., true, ...) }  // httpOnly = true
  func SetRefreshTokenCookie(...) { ... newCookie(..., true, ...) }  // httpOnly = true
  func SetCSRFCookie(...)         { ... newCookie(..., false, ...) } // readable by JS
  ```
- `apps/web/lib/api.ts` sends `credentials: "include"` and does not read access/refresh tokens from `localStorage`.
- `apps/web/app/(auth)/login/page.tsx` and `apps/web/app/(auth)/register/page.tsx` receive `token`/`refresh_token` in the JSON body but do not store them; the browser persists the cookies sent by the server.
- `localStorage` is used only for non-secret UI state (`testra_workspace_id`, `testra_organization_id`, `testra_project_id`, etc.) in `apps/web/components/providers/workspace-provider.tsx`.

**Conclusion:** The contradiction is resolved. Token storage is server-side HttpOnly cookies; `localStorage` only holds workspace/organization/project selection state.

---

## 7. Swallowed-error sweep (`errcheck`)

**Status:** Linter run and output grouped by package  
**Severity:** Medium (code-quality / reliability)

### Method
```bash
go install github.com/kisielk/errcheck@latest
errcheck -blank -ignoretests ./...
```

### Summary
| Package | Unchecked calls |
|---|---|
| `cmd/worker` | 5 |
| `internal/analytics` | 46 |
| `internal/apikeys` | 4 |
| `internal/apitesting` | 11 |
| `internal/audit` | 1 |
| `internal/automationhub` | 15 |
| `internal/billing` | 6 |
| `internal/defects` | 4 |
| `internal/integrationhub` | 60 |
| `internal/intelligence` | 16 |
| `internal/metrics` | 2 |
| `internal/notification` | 9 |
| `internal/organization` | 3 |
| `internal/project` | 3 |
| `internal/queue` | 4 |
| `internal/rbac` | 1 |
| `internal/results` | 43 |
| `internal/search` | 7 |
| `internal/shared/db` | 1 |
| `internal/shared/http` | 3 |
| `internal/shared/jwt` | 1 |
| `internal/shared/middleware` | 9 |
| `internal/shared/server` | 1 |
| `internal/testmanagement` | 7 |
| `internal/workspace` | 3 |
| `cmd/api` | 1 |
| **Total** | **266** |

Full grouped output is in `apps/api/errcheck_grouped.txt`. High-priority patterns include `_ = tx.Rollback()`, `_, _ = w.Write(...)`, `rows, _ := result.RowsAffected()`, and `body, _ := json.Marshal(...)`.

---

## 8. Race detector in CI/Linux

**Status:** Confirmed  
**Severity:** N/A

- `.github/workflows/ci.yml` runs on `ubuntu-latest` and executes:
  ```yaml
  - name: Test
    run: go test -race -count=1 ./apps/api/...
  ```
- Local Windows run could not execute `-race` because no C compiler is installed (`cgo: C compiler "gcc" not found`), but this is an environment limitation, not a CI limitation. Linux CI provides `gcc` and therefore supports the race detector.

---

## 9. Test suite status

```bash
go test ./... -count=1
```

All Go test packages pass after the P0 fixes.

---

## 10. Summary of changes

| Area | Change | Key files |
|---|---|---|
| SSRF | Add `security.ValidateURL` call + test | `apps/api/internal/apitesting/service.go`, `service_test.go` |
| Tenant context | Propagate `SetLocalTenantID` errors and rollback | `apps/api/internal/shared/db/db.go`, `*/repository.go` |
| DB | Regression test for failed tenant context | `apps/api/internal/shared/db/db_test.go` |
| OpenAPI | Fix drift script, sync spec | `scripts/check-openapi-drift.mjs`, `docs/api/openapi/openapi.yaml` |
| WebKit | Diagnosis, trace/logs captured | `tests/pw_webkit*.log`, `tests/reports/test-results/*/trace.zip` |
| Tokens | Confirmed HttpOnly cookies vs `localStorage` | `apps/api/internal/shared/middleware/cookies.go`, `apps/web/lib/api.ts` |
| Swallowed errors | `errcheck -blank` grouped by package | `apps/api/errcheck_grouped.txt` |
| Race detector | Confirmed in CI | `.github/workflows/ci.yml` |
