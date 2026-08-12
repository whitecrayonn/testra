# Testra Production Quality Sweep — Report

## Executive Summary

This report documents the results of a final production-quality sweep for the Testra MVP. The sweep focused on real issues that reduce product quality across the Go backend, React/TypeScript frontend, API consistency, and tenant isolation/security. No new modules or features were added; only bugs, inconsistencies, and safety gaps were fixed.

**Verdict:** The core build, test, lint, and OpenAPI-drift validations now pass. The most significant functional/security gaps found in the notification center and API documentation have been fixed. There are remaining product-level caveats (noted below) that keep the MVP from a confident "ship today" recommendation without a short follow-up verification window.

**Readiness Score:** 7.5 / 10  
**Shipping Decision:** **Conditional NO** — do not ship until the remaining blockers/limitations are triaged. The product is much closer to ship-ready than at the start of the sweep.

---

## Validation Results

All command-line validations were run from `c:\Private\project\testra` and `c:\Private\project\testra\apps\api`.

| Validation | Command | Result |
|------------|---------|--------|
| Backend format | `go fmt ./...` (in `apps/api`) | **PASS** |
| Backend vet | `go vet ./...` (in `apps/api`) | **PASS** |
| Backend build | `go build ./...` (in `apps/api`) | **PASS** |
| Backend tests | `go test ./...` (in `apps/api`) | **PASS** |
| Frontend lint | `pnpm lint` | **PASS** |
| Frontend typecheck | `pnpm typecheck` | **PASS** |
| Frontend build | `pnpm build` | **PASS** |
| Frontend tests | `pnpm test` | **PASS** |
| OpenAPI drift | `node scripts/check-openapi-drift.mjs` | **PASS** — 168 routes synchronized |

`pnpm test` executed all monorepo test/build tasks successfully:
- `@testra/web:build`
- `@testra/api:build`
- `@testra/api:test` (`go test ./...`)

---

## Issues Found and Fixed

### 1. OpenAPI drift — undocumented `GET /search` route

**Finding:** The API implements `GET /search` in `apps/api/internal/search/handler.go` but the route was missing from `docs/api/openapi/openapi.yaml`. The `check-openapi-drift.mjs` script flagged it.

**Fix:** Added a complete `GET /search` operation to `openapi.yaml` including:
- `Search` tag
- Query parameters: `workspace_id` (required, uuid), `q` (required, string), `limit` (optional, integer)
- Response envelope (`Envelope`) referencing a new `SearchResult` schema
- `SearchResult` and `SearchItem` component schemas with the categorized result fields returned by the handler

**Files modified:**
- `docs/api/openapi/openapi.yaml`

---

### 2. Frontend lint/type errors in global search and command palette

**Finding:** `pnpm lint` failed with `react-hooks/exhaustive-deps` warnings and unescaped double quotes in JSX in `GlobalSearch.tsx` and `CommandPalette.tsx`.

**Fix:**
- Added `useCallback` import and memoized the `navigate` and `run` functions.
- Reordered function declarations to avoid temporal dead zone.
- Escaped unescaped quotes in JSX.

**Files modified:**
- `apps/web/components/global-search.tsx`
- `apps/web/components/command-palette.tsx`

---

### 3. Notification template validation and trimming gaps

**Finding:**
- `CreateTemplate` accepted any `ChannelType` string; it did not validate against the allowed `email|slack|teams|webhook|in_app` values.
- `UpdateTemplate` did not trim `Name`, `Subject`, or `Body` and did not validate `ChannelType` when supplied.

**Fix:**
- `CreateTemplate` now rejects invalid `ChannelType` via `IsValidChannelType(...)`.
- `UpdateTemplate` now trims the `Name`, `Subject`, and `Body` fields and validates `ChannelType` before updating.

**Files modified:**
- `apps/api/internal/notification/service.go`

---

### 4. SMTP header injection vulnerability

**Finding:** `dispatchEmail` constructed raw RFC-5322 headers by interpolating `to` (from user-supplied channel config) and `input.Title` (notification subject) without sanitizing line breaks. A crafted subject or `to` value could inject additional headers.

**Fix:**
- Added `containsHeaderInjection(s string)` helper in `service.go`.
- `dispatchEmail` now rejects `to`, `s.smtp.From`, and `input.Title` values that contain `\r` or `\n`, returning `ErrInvalidInput`.

**Files modified:**
- `apps/api/internal/notification/service.go`

---

### 5. Notification center tenant resolution broken for most routes

**Finding:** The entire `/api/v1` notification center block used a single `TenantContext` resolver (`WorkspaceToOrg(OrgIDFromQuery("workspace_id"))`). Many notification endpoints have no `workspace_id` query:
- `GET /notifications`, `GET /notifications/unread-count`
- `GET /notification-preferences`, `PUT /notification-preferences`
- `GET /notification-templates`, `GET /notification-templates/{id}`, `POST /notification-templates`, `PUT /notification-templates/{id}`, `DELETE /notification-templates/{id}`
- `PATCH /notifications/{id}`, `DELETE /notifications/{id}`
- `POST /notification-channels` (workspace_id is in the body, not query)
- `PUT /notification-channels/{id}`, `DELETE /notification-channels/{id}`
- `GET /notification-history` (uses `notification_id` query)

These endpoints would fail at `TenantContext` with `400 could not resolve organization from request` unless the caller happened to pass an undocumented `workspace_id` query.

**Fix:**
- Split the monolithic notification route group into sub-groups, each with the correct tenant resolver:
  - `WorkspaceToOrg(OrgIDFromQuery("workspace_id"))` for `GET /notification-channels` and `POST /notifications`.
  - `WorkspaceToOrg(OrgIDFromBody("workspace_id"))` for `POST /notification-channels`.
  - `OrgIDFromQuery("organization_id")` for `GET /notification-templates`.
  - `OrgIDFromBody("organization_id")` for `POST /notification-templates`.
  - `NotificationToOrg(OrgIDFromURLParam("id"))` for `PATCH/DELETE /notifications/{id}`.
  - `NotificationChannelToOrg(OrgIDFromURLParam("id"))` for `PUT/DELETE /notification-channels/{id}`.
  - `NotificationTemplateToOrg(OrgIDFromURLParam("id"))` for `GET/PUT/DELETE /notification-templates/{id}`.
  - `NotificationToOrg(OrgIDFromQuery("notification_id"))` for `GET /notification-history`.
  - `DefaultOrgForUser(...)` for `GET /notifications`, `GET /notifications/unread-count`, and `GET/PUT /notification-preferences`.
- Added `ResolveOrgFromNotification`, `ResolveOrgFromNotificationChannel`, `ResolveOrgFromNotificationTemplate`, and `ResolveDefaultOrgForUser` to `tenant.Resolver` and the `WorkspaceOrgResolver` interface.
- Added corresponding middleware wrappers in `middleware/tenant.go`: `NotificationToOrg`, `NotificationChannelToOrg`, `NotificationTemplateToOrg`, `DefaultOrgForUser`.
- Added migration `000038_add_notification_lookup_rls_policies.up.sql` (and `.down.sql`) to enable authenticated-user lookup on `notifications`, `notification_preferences`, `notification_channels`, `notification_templates`, and `notification_history` before `app.tenant_id` is known. This is required for the id-based and user-based resolvers to query the DB during tenant resolution.

**Files modified:**
- `apps/api/internal/shared/server/server.go`
- `apps/api/internal/shared/middleware/tenant.go`
- `apps/api/internal/shared/tenant/resolver.go`
- `apps/api/migrations/000038_add_notification_lookup_rls_policies.up.sql`
- `apps/api/migrations/000038_add_notification_lookup_rls_policies.down.sql`

---

### 6. Migration version collision

**Finding:** While adding the notification lookup RLS migration, an untracked pre-existing migration pair `000037_add_workspace_description.*` already existed in the migrations directory. Two `000037` files would conflict in the migration runner.

**Fix:** Renamed the new notification lookup RLS migration to `000038`.

**Files modified:**
- `apps/api/migrations/000038_add_notification_lookup_rls_policies.up.sql`
- `apps/api/migrations/000038_add_notification_lookup_rls_policies.down.sql`

---

## Files Modified During the Sweep

Backend:
- `apps/api/internal/notification/service.go`
- `apps/api/internal/shared/middleware/tenant.go`
- `apps/api/internal/shared/server/server.go`
- `apps/api/internal/shared/tenant/resolver.go`
- `apps/api/migrations/000038_add_notification_lookup_rls_policies.up.sql`
- `apps/api/migrations/000038_add_notification_lookup_rls_policies.down.sql`

API documentation:
- `docs/api/openapi/openapi.yaml`

Frontend:
- `apps/web/components/global-search.tsx`
- `apps/web/components/command-palette.tsx`

(Note: `git status` shows additional pre-existing modified and untracked files outside the above list. Those were not edited or reviewed as part of this sweep.)

---

## Remaining Blockers / Limitations

1. **Default organization selection for user-scoped notification endpoints**  
   `GET /notifications`, `GET /notifications/unread-count`, and `GET/PUT /notification-preferences` now resolve the tenant via `ResolveDefaultOrgForUser`, which selects `LIMIT 1` from `organization_members` for the authenticated user. If a user belongs to multiple organizations, the chosen organization is non-deterministic and may vary per request. This is functional but not product-grade multi-tenant UX. Recommended fix: require an explicit `organization_id` query/body parameter on these endpoints (or derive from a persisted "active organization" session).

2. **Pre-existing uncommitted changes not reviewed**  
   `git status` shows a large set of pre-existing modified and untracked files (`apps/api/internal/workspace/*`, many `apps/web/*` pages/components, `apps/web/features/search/`, `apps/web/lib/hooks/`, `000037_add_workspace_description.*`, etc.). These were not introduced or reviewed in this sweep. They should be reconciled with the branch and validated before shipping.

3. **No runtime/end-to-end verification of the notification tenant fix**  
   The notification route changes are validated by `go test`/`go build` and the route table compiles, but no integration or UI tests were executed against a live PostgreSQL instance to confirm the new lookup policies and resolvers behave correctly under RLS.

4. **Email channel security surface**  
   The SMTP header-injection guard is in place, but email `to` addresses are not fully validated as RFC-compliant mailboxes. Invalid `to` values may still be passed to `smtp.SendMail` and fail at send time rather than at configuration time.

---

## Readiness Assessment

| Area | Status | Notes |
|------|--------|-------|
| Build & tests | **Green** | `go` and `pnpm` build/test/lint/typecheck pass. |
| API documentation | **Green** | OpenAPI drift check passes. |
| API consistency | **Improved** | Notification tenant resolution aligned with rest of app. |
| Security | **Improved** | SMTP header injection mitigated; notification lookup RLS added. |
| Tenant isolation | **Improved** | Notification routes now resolve org correctly; RLS policies added. |
| UI/UX consistency | **Partial** | Frontend lint/build pass; deeper UI/UX review of all pages not completed. |
| Multi-org UX | **Yellow** | Default-org selection for notifications is a placeholder. |
| Uncommitted changes | **Red/Yellow** | Pre-existing modified/untracked files outside this sweep remain. |

**Overall readiness score:** 7.5 / 10

**Shipping decision:** **Conditional NO.** The fixes in this sweep are real and important, and the automated validation suite is green. However, because of the non-deterministic default-org behavior for user-scoped notification endpoints and the large set of unreviewed pre-existing changes, I would not ship the MVP today. A short follow-up (explicit org scoping for notifications + reconciliation of the pre-existing uncommitted files + a focused integration test pass) would bring the score to 8.5–9/10 and a clear ship recommendation.
