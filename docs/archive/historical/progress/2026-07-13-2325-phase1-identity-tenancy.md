# Engineering Progress Report — 2026-07-13 23:25

## Session Summary
Completed Phase 1 (Identity & Tenancy) — implemented TOTP MFA, password reset, RBAC middleware, scoped API keys, and full web auth/onboarding UI.

## Completed
- TOTP MFA: setup (generate secret + QR URL), verify (enable), disable endpoints in identity module
- Password reset: request (generate SHA-256 hashed token, 30min expiry), confirm (validate + update password) endpoints
- Migration 000005: added `mfa_secret`, `mfa_enabled` columns to users; created `password_reset_tokens` table
- RBAC: migration 000006 with roles, permissions, role_permissions, role_assignments tables + seed data (4 roles, 21 permissions)
- RBAC middleware: `RequirePermission` with `PermissionLoader` interface in shared/middleware
- `rbac` package: `SQLPermissionLoader` implementation
- API keys: full module (domain, ports, repository, service, handler, module) with create/list/revoke
- API keys: SHA-256 hashed storage, `testra_` prefix, one-time raw key display, scope support
- Migration 000007: `api_keys` table with workspace FK, scopes array, expiry, revocation
- Identity service tests: 17 tests covering MFA login/setup/verify/disable + password reset request/reset/edge cases
- Web: TailwindCSS 3 + PostCSS config, brand color theme
- Web: UI components (Button with variants/sizes/loading, Input with label/error, Card suite)
- Web: API client (`lib/api.ts`) with token management, error handling, envelope parsing
- Web: auth layout, login page (with MFA code field), register page
- Web: forgot-password page, reset-password page (with token from URL params)
- Web: MFA setup page (QR display + 6-digit verification)
- Web: onboarding page (create org + workspace in sequence)
- Web: dashboard layout with sidebar navigation (6 nav items + sign out)
- Web: dashboard page
- All routes wired: `/auth/mfa/*`, `/auth/password-reset/*`, `/api-keys`
- `ErrMFARequired` sentinel error added to shared errors
- `pquerna/otp` v1.5.0 dependency added
- PHASES.md updated: Phase 1 marked as Completed

## In Progress
- None (Phase 1 complete)

## Blocked
- None

## Next
- Phase 2: Test Management Core — test cases, suites, folders, versioning, full-text search, audit trail

## Files Changed
- `apps/api/migrations/000005_add_mfa_and_password_reset.{up,down}.sql` — created
- `apps/api/migrations/000006_add_rbac.{up,down}.sql` — created
- `apps/api/migrations/000007_add_api_keys.{up,down}.sql` — created
- `apps/api/internal/identity/domain.go` — added MFA fields, PasswordResetToken entity
- `apps/api/internal/identity/ports.go` — added MFA/reset repository methods
- `apps/api/internal/identity/repository.go` — implemented MFA/reset methods, updated queries for MFA columns
- `apps/api/internal/identity/service.go` — added MFA setup/verify/disable, password reset request/reset, token helpers
- `apps/api/internal/identity/handler.go` — added MFA setup/verify/disable, password reset request/confirm handlers
- `apps/api/internal/identity/service_test.go` — created (17 tests)
- `apps/api/internal/shared/errors/errors.go` — added ErrMFARequired
- `apps/api/internal/shared/middleware/rbac.go` — created (RequirePermission, tenant/permission context)
- `apps/api/internal/rbac/loader.go` — created (SQLPermissionLoader)
- `apps/api/internal/apikeys/{domain,ports,repository,service,handler,module}.go` — created
- `apps/api/internal/shared/server/server.go` — wired MFA, password reset, API key routes
- `apps/web/tailwind.config.ts` — created
- `apps/web/postcss.config.js` — created
- `apps/web/lib/api.ts` — created
- `apps/web/lib/utils.ts` — created
- `apps/web/components/ui/{button,input,card}.tsx` — created
- `apps/web/components/dashboard/sidebar.tsx` — created
- `apps/web/app/(auth)/layout.tsx` — created
- `apps/web/app/(auth)/login/page.tsx` — created
- `apps/web/app/(auth)/register/page.tsx` — created
- `apps/web/app/(auth)/forgot-password/page.tsx` — created
- `apps/web/app/(auth)/reset-password/page.tsx` — created
- `apps/web/app/(auth)/mfa-setup/page.tsx` — created
- `apps/web/app/(auth)/onboarding/page.tsx` — created
- `apps/web/app/(dashboard)/layout.tsx` — created
- `apps/web/app/(dashboard)/dashboard/page.tsx` — created
- `apps/web/app/page.tsx` — updated (redirect to /login)
- `docs/engineering/PHASES.md` — updated (Phase 1 completed)

## Verification
- `go build ./...` — pass
- `go vet ./...` — pass
- `go test -count=1 ./...` — pass (identity: 17 tests, project: ok)
- `pnpm turbo run typecheck` — 4/4 tasks successful
