# Backend Quality Cleanup Report v1

**Date:** 2026-07-19  
**Scope:** `apps/api` Go backend — panics, `TODO`/`FIXME`, dead code, ignored errors, resource leaks, and `go vet` findings.  
**Goal:** Fix the critical and high-impact quality issues that could mask failures or cause state inconsistency, and document remaining cleanup work.

## Verification

```powershell
go build ./...
go vet ./...
go test -count=1 ./...
```

All three commands pass with exit code 0 for `apps/api`.

## Fixes Applied

### 1. Identity — refresh-token revocation errors were silently dropped

**File:** `apps/api/internal/identity/service.go`

`Refresh`, `Logout`, and `LogoutAllDevices` ignored errors from `RevokeRefreshToken`, `RevokeRefreshTokenFamily`, and `RevokeAllUserRefreshTokens`. If the database failed to write the revocation, the old token could remain valid while the caller believed the session was terminated.

**Fix:**
- Propagate each revocation error as `sharederrors.ErrInternal`.
- Do not issue a new refresh token until the old one is successfully revoked.

### 2. Notification — removed dead `dbHandle` assignment

**File:** `apps/api/internal/notification/module.go`

`NewModule` created `db.Wrap(sqlDB)` into `dbHandle` and then discarded it. The repository already wraps the database, so the line was dead code that also pulled in an unused `shared/db` import.

**Fix:**
- Removed `dbHandle` and the `shared/db` import.
- Constructor now passes `sqlDB` directly to `NewSQLRepository`.

### 3. Results — run-count recalculation ignored update failures

**File:** `apps/api/internal/results/service.go`

`recalcRunCounts` computed totals from run items and then silently discarded the error from `repo.UpdateRun`. If the update failed, the cached totals (`Total`, `Passed`, `Failed`, `Skipped`, `Blocked`, `DurationMs`) were not persisted, yet the progress broadcast still reported the recomputed values.

**Fix:**
- Changed `recalcRunCounts` to return `error`.
- `UpdateItemStatus` now propagates that error instead of broadcasting stale/partial state.

### 4. IntegrationHub — event persistence and payload marshal errors ignored

**Files:** `apps/api/internal/integrationhub/repository.go`, `apps/api/internal/integrationhub/service.go`

- `repository.go` used `payloadJSON, _ := json.Marshal(e.Payload)`, so malformed payloads could be written as empty bytes.
- `service.go` ignored the second `CreateEvent` error used to persist the final event status after an external dispatch.

**Fix:**
- `CreateEvent` now returns an error if `json.Marshal` fails.
- `DispatchEvent` returns an error if the final event-status write fails.

### 5. Billing — subscription and invoice writes silently failed

**File:** `apps/api/internal/billing/service.go`

`createDefaultSubscription`, `GetSubscription` provider-sync, and `ListInvoices` provider-sync all ignored repository write errors. This could return a success response while the database contained stale or missing records.

**Fix:**
- `createDefaultSubscription` now returns `(*Subscription, error)`.
- `GetSubscription` propagates `UpsertSubscription` errors during provider refresh.
- `UpdateSubscription` handles errors from default-subscription creation.
- `ListInvoices` propagates `CreateInvoice` errors during provider invoice import.

## Remaining Quality Debt

The following items were identified but not addressed in this pass to keep the change set focused. They are candidates for the next cleanup sweep.

### A. Ignored JSON response decode errors in integration adapters

**Files:** `apps/api/internal/integrationhub/adapters.go`

The Jira, GitHub, and GitLab adapters call `_ = json.Unmarshal(respBody, &result)` when parsing response bodies. If the external API returns unexpected JSON, the adapter returns a zero-value and no error, leading to confusing downstream behavior.

**Recommended fix:**
```go
if err := json.Unmarshal(respBody, &result); err != nil {
    return "", fmt.Errorf("decode %s response: %w", adapterName, err)
}
```

### B. `shared/http/response.go` discards `Encode` errors

**File:** `apps/api/internal/shared/http/response.go`

`JSON` and `ErrorJSON` call `_ = json.NewEncoder(w).Encode(...)`. If the client has closed the connection, the write fails silently. There is no logger in this package.

**Recommended fix:**
- Add a package-level logger or accept an `slog.Logger` in `JSON`/`ErrorJSON`.
- Log encode/write failures at `slog.LevelWarn`.

### C. `apikeys/service.go` ignores `UpdateLastUsed` errors

**File:** `apps/api/internal/apikeys/service.go`

`Validate` ignores the result of `repo.UpdateLastUsed(ctx, key.ID)`. If this update fails, API-key usage tracking becomes inaccurate.

**Recommended fix:**
- Log a warning or record a metric; do not fail the request because the credential is still valid.

### D. `analytics/repository.go` ignores JSON unmarshal errors for config

**File:** `apps/api/internal/analytics/repository.go`

Config deserialization uses `_ = json.Unmarshal([]byte(configStr), &d.Config)`. Invalid JSON produces an empty config instead of an error.

**Recommended fix:**
```go
if err := json.Unmarshal([]byte(configStr), &d.Config); err != nil {
    return nil, fmt.Errorf("decode dashboard config: %w", err)
}
```

### E. Several repositories ignore rollback errors

**Files:** `apps/api/internal/testmanagement/repository.go`, `apps/api/internal/results/repository.go`, `apps/api/internal/billing/repository.go`, `apps/api/internal/integrationhub/repository.go`, `apps/api/internal/intelligence/repository.go`

Pattern:
```go
_ = tx.Rollback()
```

The application is already returning an error, but ignoring the rollback error can mask connection-pool exhaustion or database-side failures. Go best practice is to use `defer tx.Rollback()` or log the rollback error.

### F. Panics in test helpers

**Files:** `apps/api/internal/identity/service_test.go`, `tests/integration/setup_test.go`

Test setup helpers use `panic(err)` if `jwt.NewTestManager` or `os.Getwd` fail. While acceptable for tests, changing these to `t.Fatal` or returning an error would make test failures easier to diagnose.

### G. Empty / placeholder packages

**Files:** `apps/api/internal/apitesting/`, `apps/api/internal/integration/`

- `apitesting` contains only `.gitkeep` and an empty `module.go`.
- `integration` contains only `tenant_isolation_test.go`, which tests `shared/tenant` behavior.

**Recommended fix:**
- Populate `apitesting` or remove it.
- Move `integration/tenant_isolation_test.go` to `shared/tenant` or rename it to `tenant_test.go`.

### H. `cmd/worker` imports many modules but has limited error reporting

**File:** `apps/api/cmd/worker/main.go`

The worker polls `queue_jobs` and processes jobs. If a job panics or repeatedly fails, it relies on `MarkFailed` retries. There is no dead-letter queue or structured logging of retry exhaustion.

**Recommended fix:**
- Add panic recovery per job.
- Emit structured logs/metrics on permanent failure.
- Document the retry policy.

## Conclusion

The highest-risk quality issues — ignored revocation, dead code, stale run totals, and ignored repository writes — are now fixed. The remaining debt is mostly observability and defensive error handling. A follow-up pass should focus on adapter JSON decode failures, response encoder logging, and the empty placeholder packages.
