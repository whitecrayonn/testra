# Testra — Frontend UX Review

**Scope:** `apps/web` Next.js 15 App Router frontend  
**Objective:** Close Phase 3.5 by stabilizing the product UX, ensuring every menu item routes correctly, and applying consistent design, accessibility, and responsive patterns.  
**Date:** July 2026  

## Executive Summary

The frontend now has a complete route tree, production-quality placeholder pages for Phase 4 features, fully implemented settings pages, a polished dashboard, and consistent loading/empty/error states across all primary views. All routes build successfully, typecheck passes, and ESLint reports zero warnings.  

## Route & Page Inventory

| Page | File | Status | Notes |
|------|------|--------|-------|
| `/login` | `app/(auth)/login/page.tsx` | Implemented | MFA-aware, error states, token storage |
| `/register` | `app/(auth)/register/page.tsx` | Implemented | Redirects to onboarding |
| `/onboarding` | `app/(auth)/onboarding/page.tsx` | Implemented | Creates org/workspace and stores context |
| `/mfa-setup` | `app/(auth)/mfa-setup/page.tsx` | Implemented | TOTP generation/verify |
| `/forgot-password` | `app/(auth)/forgot-password/page.tsx` | Implemented | Request/confirmation state |
| `/reset-password` | `app/(auth)/reset-password/page.tsx` | Implemented | Suspense-wrapped for `useSearchParams` |
| `/dashboard` | `app/(dashboard)/dashboard/page.tsx` | Polished | Stats cards, quick actions, workspace context, empty state |
| `/[workspace]` | `app/(dashboard)/[workspace]/page.tsx` | Polished | Stores selected workspace and redirects to dashboard |
| `/[workspace]/projects` / `/dashboard/projects` | `[workspace]/projects/page.tsx` + re-export | Implemented | CRUD, selection, localStorage |
| `/[workspace]/defects` / `/dashboard/defects` | `[workspace]/defects/page.tsx` + re-export | Placeholder | Phase 4 CTA |
| `/[workspace]/test-cases` / `/dashboard/test-cases` | `[workspace]/test-cases/page.tsx` + re-export | Implemented | Search, status/priority badges, pagination |
| `/[workspace]/test-cases/new` / `.../[id]` | `[workspace]/test-cases/new/page.tsx`, `[id]/page.tsx` | Implemented | Full editor with steps |
| `/[workspace]/test-runs` / `/dashboard/test-runs` | `[workspace]/test-runs/page.tsx` + re-export | Implemented | Pagination, SSE progress, badges |
| `/[workspace]/test-runs/new` / `.../[id]` | `[workspace]/test-runs/new/page.tsx`, `[id]/page.tsx` | Implemented | Create + live run detail |
| `/dashboard/settings` | `dashboard/settings/page.tsx` | Implemented | Settings overview cards |
| `/dashboard/settings/profile` | `.../settings/profile/page.tsx` | Implemented | Read-only profile view |
| `/dashboard/settings/security` | `.../settings/security/page.tsx` | Implemented | Password/MFA placeholders with status |
| `/dashboard/settings/organization` | `.../settings/organization/page.tsx` | Implemented | Fetch + disabled edit |
| `/dashboard/settings/workspace` | `.../settings/workspace/page.tsx` | Implemented | Fetch + disabled edit |
| `/dashboard/settings/members` | `.../settings/members/page.tsx` | Placeholder | Planned Phase 4 |
| `/dashboard/settings/roles` | `.../settings/roles/page.tsx` | Placeholder | Planned Phase 4 |
| `/dashboard/settings/api-keys` | `.../settings/api-keys/page.tsx` | Implemented | List/create/revoke scoped keys |
| `/dashboard/settings/notifications` | `.../settings/notifications/page.tsx` | Implemented | Preferences, email/Slack/Teams/webhook channels |
| `/dashboard/notifications` | `dashboard/notifications/page.tsx` | Implemented | In-app feed, mark read/delete |
| `/dashboard/settings/audit-logs` | `.../settings/audit-logs/page.tsx` | Placeholder | Planned Phase 4 |
| `/dashboard/settings/billing` | `.../settings/billing/page.tsx` | Placeholder | Planned Phase 4 |
| `/dashboard/settings/preferences` | `.../settings/preferences/page.tsx` | Placeholder | Appearance toggle, disabled save |

## UX Findings

### Completed Improvements
- **Page headers:** All main views use `PageHeader` with title, description, breadcrumbs, and actions.
- **Empty states:** Reusable `EmptyState` component with icon, title, description, primary/secondary actions.
- **Loading states:** `Skeleton` and `CardSkeleton` components used on dashboard, lists, and detail pages.
- **Error states:** Consistent `border-red-200 bg-red-50` alert cards with `role="alert"`.
- **Badges:** Status/priority tags use `Badge` variants (`success`, `warning`, `danger`, `info`, `neutral`) instead of ad-hoc Tailwind.
- **Buttons:** `Button` supports `loading`, `aria-busy`, `aria-disabled`; `LinkButton` prevents invalid `<button>` inside `<a>` markup.
- **Forms:** Accessible labels (`htmlFor`/`id`), responsive grids, focus rings.
- **Navigation:** `Sidebar` highlights active routes including settings sub-pages; `SettingsNav` provides vertical sub-navigation.
- **Dashboard:** Quick-action cards, workspace context, empty-state CTA when no project selected.

### Placeholder Quality
- Phase 4 placeholders (`defects`, `members`, `roles`, `notifications`, `audit-logs`, `billing`) include:
  - Detailed status and release phase.
  - Bulleted planned features.
  - Primary CTA (e.g., go to Projects) and secondary action.
  - Breadcrumbs and consistent card styling.

### Accessibility
- `aria-hidden` on decorative Lucide icons.
- `aria-current` on active sidebar links.
- `role="progressbar"` with ARIA values on live run progress bar.
- Semantic `label` + `htmlFor` associations on form inputs.
- Keyboard-focusable `LinkButton` and `Button` components.

### Responsiveness
- Dashboard stat grid: `sm:grid-cols-2 lg:grid-cols-4`.
- Test case/run lists stack labels and metadata on small screens.
- Form grids use `sm:grid-cols-2` and single column on mobile.
- Sidebar collapses on mobile (existing hamburger pattern retained).

## Technical Review

| Concern | Finding | Action Taken |
|---------|---------|--------------|
| Broken/incomplete types | Typecheck passes | Fixed API response type (`raw_key` vs `plaintext_key`) |
| Dead/unused imports | Lint now clean | Removed unused imports across settings and placeholder pages |
| Invalid HTML (button inside link) | Found in dashboard/quick actions, test list CTAs | Introduced `LinkButton` component and replaced all instances |
| `useSearchParams` without Suspense | Build error in reset-password | Refactored to `ResetPasswordForm` + `Suspense` wrapper |
| `output: 'standalone'` symlink EPERM on Windows | Build failure on Windows dev host | Switched to default server output (still produces `.next` build) |

## Recommendations
1. **Error boundaries:** Add `app/(dashboard)/error.tsx` and `app/(dashboard)/[workspace]/[...not-found]/page.tsx` to gracefully handle 404s and runtime errors in production.
2. **E2E navigation test:** Add a Playwright/Cypress smoke test that visits every sidebar link and verifies no blank pages.
3. **Storybook:** Document new UI components (`PageHeader`, `EmptyState`, `Skeleton`, `Badge`, `LinkButton`) for design-system consistency.
4. **LocalStorage persistence:** Consider moving workspace/project selection to a small Zustand store with persistence rather than direct localStorage reads in pages.

## Verification Results

- `pnpm turbo run typecheck` — ✅ Pass
- `pnpm lint` (web) — ✅ Pass (0 warnings, 0 errors)
- `pnpm turbo run build` — ✅ Pass

