# Frontend Audit

**Scope:** `testra/apps/web` Next.js application and shared UI/types.

**Status:** Audit complete.

---

## 1. Executive Summary

The Testra frontend is a Next.js 15 application using the App Router. It is built in a pnpm monorepo and styled with Tailwind CSS. State is local (React `useState`/`useEffect`) and context is persisted in `localStorage` (token, organization/workspace/project IDs). It consumes the Go backend at `NEXT_PUBLIC_API_URL` with a small `apiFetch` wrapper.

Implemented pages cover authentication, onboarding, workspace/project selection, test case management, test run execution, and settings placeholders. Defects, API testing, and several settings sub-pages are stubs or placeholder pages.

---

## 2. Technology Stack

| Layer | Technology |
|-------|------------|
| Framework | Next.js 15 (`next@15.0.0`) with App Router |
| Language | TypeScript 5.5 |
| Runtime | React 18.3 |
| Styling | Tailwind CSS 3, `tailwind-merge`, `clsx` |
| Forms | `react-hook-form` + `zod` + `@hookform/resolvers` |
| Icons | `lucide-react` |
| State | Local component state + `localStorage` |
| Package manager | pnpm 9.5.0 (workspace) |

Workspace dependencies:

- `@testra/shared`
- `@testra/ui`
- `@testra/config` (dev)

See `apps/web/package.json` for the full dependency list.

---

## 3. App Structure & Routing

Root layout and entry:

- `apps/web/app/layout.tsx` — minimal root layout with `html`/`body`.
- `apps/web/app/page.tsx` — root page, likely redirects (not reviewed in detail).

### 3.1 Auth routes — `app/(auth)/...`

| Route | File | Purpose |
|-------|------|---------|
| `/login` | `login/page.tsx` | Email/password login, optional MFA code input |
| `/register` | `register/page.tsx` | Account creation |
| `/forgot-password` | `forgot-password/page.tsx` | Request reset email |
| `/reset-password` | `reset-password/page.tsx` | Confirm reset with token and new password |
| `/mfa-setup` | `mfa-setup/page.tsx` | TOTP enrollment and verification |
| `/onboarding` | `onboarding/page.tsx` | Create organization + workspace after registration |

`app/(auth)/layout.tsx` wraps auth pages in a centered card layout.

### 3.2 Dashboard routes — `app/(dashboard)/...`

The dashboard is split between two route trees that duplicate some paths:

1. **`app/(dashboard)/dashboard/...`**
2. **`app/(dashboard)/[workspace]/...`** (dynamic workspace slug)

| Route | File | Purpose |
|-------|------|---------|
| `/dashboard` | `dashboard/page.tsx` | Landing with quick links |
| `/dashboard/projects` | `[workspace]/projects/page.tsx` | List/select/create projects |
| `/dashboard/test-cases` | `[workspace]/test-cases/page.tsx` | Search/list test cases |
| `/dashboard/test-cases/new` | `[workspace]/test-cases/new/page.tsx` | Create test case |
| `/dashboard/test-cases/[id]` | `[workspace]/test-cases/[id]/page.tsx` | Edit/view test case + version history |
| `/dashboard/test-runs` | `[workspace]/test-runs/page.tsx` | List test runs |
| `/dashboard/test-runs/new` | `[workspace]/test-runs/new/page.tsx` | Create manual test run |
| `/dashboard/test-runs/[id]` | `[workspace]/test-runs/[id]/page.tsx` | Run detail + SSE progress + start |
| `/dashboard/defects` | `[workspace]/defects/page.tsx` | Placeholder for Phase 4 |
| `/dashboard/settings` | `dashboard/settings/page.tsx` | Settings overview |
| `/dashboard/settings/*` | `dashboard/settings/*/page.tsx` | Placeholder/empty settings tabs |

`app/(dashboard)/layout.tsx` renders `Sidebar` and the main content area.

### 3.3 Dynamic workspace route

`app/(dashboard)/[workspace]/page.tsx` stores the `workspaceId` path param in `localStorage` as `testra_workspace_id`.

---

## 4. API Layer (`apps/web/lib/api.ts`)

`apiFetch` is a thin wrapper around `fetch`:

- Base URL from `process.env.NEXT_PUBLIC_API_URL` (defaults to `http://localhost:8080`). API wrappers prefix paths with `/api/v1`.
- Adds `Content-Type: application/json`.
- Reads `testra_token` from `localStorage` and sets `Authorization: Bearer <token>`.
- Parses the backend envelope `{ data, meta, error }` and throws `ApiError` when `error` is present.
- `setToken`/`clearToken`/`getToken` manage `localStorage` token key.

There is no refresh-token logic on the client; expired access tokens will surface as `ApiError` and the user must re-login.

---

## 5. State & Context Management

There is **no global state library** (no Redux, Zustand, React Context for auth). Each page uses local `useState`/`useEffect` and reads from `localStorage` directly.

Persisted `localStorage` keys:

| Key | Purpose |
|-----|---------|
| `testra_token` | JWT access token |
| `testra_organization_id` | Selected organization UUID |
| `testra_workspace_id` | Selected workspace UUID |
| `testra_project_id` | Selected project UUID |
| `testra_project_name` | Selected project name |

This is simple but leads to duplicated `typeof window !== "undefined"` checks and potential hydration mismatches.

---

## 6. Domain API Wrappers

### `features/platform/api.ts`

- `listOrganizations`, `getOrganization`
- `listWorkspaces`, `getWorkspace`
- `listProjects`, `createProject`, `getProject`
- `listAPIKeys`, `createAPIKey`, `revokeAPIKey`
- `getCurrentUser`

### `features/testmanagement/api.ts`

- `listTestCases`, `searchTestCases`, `getTestCase`, `createTestCase`, `updateTestCase`, `deleteTestCase`
- `listTestCaseVersions`
- `listTestFolders`, `listTestSuites`

### `features/results/api.ts`

- `listTestRuns`, `getTestRun`, `createTestRun`, `updateTestRunStatus`, `deleteTestRun`
- `listTestRunItems`
- `updateTestRunItemStatus`

`features/{analytics,api-testing,automation-hub,defects,settings,test-management}` are mostly empty (`.gitkeep`) except the three wrappers above.

---

## 7. Component Library

`apps/web/components/ui/` contains custom shadcn/ui-style primitives:

- `Button` — variants `primary|secondary|ghost|danger`, sizes `sm|md|lg`, `loading` prop
- `Input` — label, error message, Tailwind-styled
- `Card`, `CardHeader`, `CardTitle`, `CardContent`
- `Badge` — variants mapped to status/priority colors
- `PageHeader` — title, description, breadcrumbs, actions
- `EmptyState` — icon, title, description, primary/secondary actions
- `PlaceholderPage` — used for unimplemented Phase 4 pages (defects, API tests)
- `Skeleton` and `CardSkeleton` — loading placeholders
- `LinkButton` — Next.js `Link` styled as a button

`apps/web/components/dashboard/sidebar.tsx` is the main navigation sidebar.
`apps/web/components/dashboard/settings-nav.tsx` renders the settings tab bar.

---

## 8. Key Pages Deep Dive

### 8.1 Login (`login/page.tsx`)

- `zod` schema: email + password.
- On success:
  - If `mfa_required`, shows a 6-digit MFA code input and calls `/auth/mfa/verify`.
  - Otherwise stores token and redirects to `/dashboard`.
- No "remember me" or token refresh handling.

### 8.2 Register (`register/page.tsx`)

- `zod` schema: name, email, password (min 12).
- Calls `/auth/register`, stores token, redirects to `/onboarding`.

### 8.3 Onboarding (`onboarding/page.tsx`)

- Collects organization name and workspace name.
- Calls `POST /api/v1/organizations` with `{ name, slug }`.
- Calls `POST /api/v1/workspaces` with `{ organization_id, name, slug }`.
- Stores returned IDs in `localStorage`.
- **Status:** Onboarding now sends explicit `slug` values.

### 8.4 Projects (`[workspace]/projects/page.tsx`)

- Loads projects for the workspace stored in `localStorage`.
- Allows creating a new project inline.
- Auto-generates project key from name to match the backend regex `^[A-Z][A-Z0-9]{1,9}$` (uppercase, starts with a letter, no hyphens).
- **Status:** Project key generator is now aligned with the backend.
- Clicking "Select" stores `testra_project_id` and `testra_project_name` in `localStorage`.

### 8.5 Test Cases

**List page (`[workspace]/test-cases/page.tsx`)**

- Requires `testra_project_id`.
- Lists cases with cursor-based pagination ("Load More").
- Full-text search mode calls `searchTestCases` by workspace.
- Maps `status`/`priority` to `Badge` variants.

**New page (`[workspace]/test-cases/new/page.tsx`)**

- Form with title, status, priority, description, preconditions, tags.
- Dynamic test steps (action/expected/test_data) with add/remove.
- Calls `createTestCase`.

**Detail page (`[workspace]/test-cases/[id]/page.tsx`)**

- Loads case, populates form fields.
- Allows editing, saving (`updateTestCase`), deleting (`deleteTestCase`).
- "Version History" loads `listTestCaseVersions`.

### 8.6 Test Runs

**List page (`[workspace]/test-runs/page.tsx`)**

- Loads runs for selected project; cursor pagination.
- Shows status/source badges and counts.

**New page (`[workspace]/test-runs/new/page.tsx`)**

- Takes a run name and comma-separated test case IDs.
- Calls `createTestRun` with `source: "manual"`.

**Detail page (`[workspace]/test-runs/[id]/page.tsx`)**

- Loads run and items.
- If run is `pending`/`running`, opens an `EventSource` to `/api/v1/test-runs/${runId}/stream?access_token=${token}`.
- Displays a progress bar and per-item status icons.
- "Start Run" button sets status to `running`.
- **Status:** SSE endpoint now accepts the JWT via an `access_token` query parameter for `EventSource` connections (MVP workaround).

### 8.7 Defects (`[workspace]/defects/page.tsx`)

Uses `PlaceholderPage` with "Planned for Phase 4" status.

---

## 9. Type Definitions

- `apps/web/types/platform.ts` — `User`, `Organization`, `Workspace`, `Project`, `APIKey`.
- `apps/web/types/testmanagement.ts` — `TestStep`, `TestCase`, `TestCaseVersion`, `TestFolder`, `TestSuite`, `PaginationMeta`.
- `apps/web/types/results.ts` — `TestRun`, `TestRunItem`, `RunProgressEvent`, `PaginationMeta`.

`PaginatedResponse<T>` is duplicated in each feature API file and in the type files.

---

## 10. Findings & Recommendations

1. **Auth token refresh is missing.** The access token expires after 15 minutes; the client has no refresh-token exchange. Consider adding an axios/fetch interceptor that calls `/auth/refresh` on 401 and retries.
2. ~~**MFA setup QR code is displayed as text.**~~ **Resolved in Phase 3.5.** The backend returns `qr_code` as a data URL string; the frontend now renders it in an `<img src={qr_code} />` tag.
3. ~~**Onboarding omits `slug` fields.**~~ **Resolved in Phase 3.5.** Onboarding now sends explicit `slug` values.
4. ~~**Project key generation can produce invalid keys.**~~ **Resolved in Phase 3.5.** The frontend generates uppercase alphanumeric keys matching the backend regex.
5. ~~**SSE stream cannot authenticate.**~~ **Resolved in Phase 3.5.** `EventSource` passes the JWT in the `access_token` query parameter (acceptable for local/MVP; a more secure token mechanism is a future hardening item).
6. **LocalStorage is the only state layer.** This works for a prototype but causes:
   - Repetitive `typeof window !== "undefined"` guards.
   - No reactive cross-tab updates.
   - Potential Next.js hydration mismatches.
   Consider a lightweight context/provider or `zustand` for auth/workspace/project state.
7. **No route guards.** The dashboard routes are not protected server-side or client-side. An unauthenticated user will see pages briefly and receive 401s from API calls.
8. **Duplicate `/dashboard` and `/[workspace]` route trees.** Both exist and store context differently. Consolidate to a single dashboard routing scheme to avoid confusion.
9. **Settings sub-pages are partially implemented.** `/dashboard/settings/notifications` and `/dashboard/settings/api-keys` are implemented. Members, roles, audit logs, billing, profile, security, organization, and workspace settings pages are still placeholders or basic cards.
10. **No error boundary or global loading state.** Each page handles its own loading/error; a shared `ErrorBoundary` and `Loading` component would improve UX.
11. **Test suite/folder management UI is missing.** The API supports `test-folders` and `test-suites`, but the frontend only lists cases by project and uses folders/suites minimally.
12. **API base URL env var is not documented in `.env.example` for web.** Add `NEXT_PUBLIC_API_URL` documentation and a `.env.example` in `apps/web`.
