# Frontend Review v1 — Testra Web Application

**Date:** 2026-07-19  
**Scope:** `apps/web` Next.js application (`app/`, `components/`, `features/`, `lib/`)  
**Goal:** Assess architecture, state management, authentication storage, security posture, accessibility, and readiness for production.

## Executive Summary

The frontend is a modern Next.js 15 + React 18 + TypeScript application using Tailwind CSS and a small internal UI component library. It follows the App Router convention but currently uses `"use client"` in most layouts and pages, shifting rendering almost entirely to the browser. The largest risk is **JWT and refresh-token storage in `localStorage`**, which exposes the application to XSS theft and prevents `HttpOnly`/`SameSite` cookie protections. Several route guards and state-hydration patterns also depend on `localStorage` at runtime, creating flicker and security concerns.

## 1. Architecture & Rendering

### 1.1 Stack

- **Framework:** Next.js 15.0.0 (`next.config.ts` with `reactStrictMode: true` and optional `output: 'standalone'`)
- **Runtime:** React 18.3.1, React DOM 18.3.1
- **Language:** TypeScript 5.5.3
- **Styling:** Tailwind CSS 3, `clsx` + `tailwind-merge` (`cn` utility)
- **Form/validation:** `react-hook-form` 7.81.0, `zod` 4.4.3, `@hookform/resolvers`
- **Icons:** `lucide-react`
- **Internal packages:** `@testra/ui`, `@testra/shared`, `@testra/config` (workspace)

### 1.2 App Router vs Client Rendering

- All inspected layouts (`(auth)/layout.tsx`, `(dashboard)/layout.tsx`) and pages (`login/page.tsx`, etc.) are marked `"use client"`.
- This means **Server Components are not being leveraged** for:
  - Pre-rendered shell / navigation
  - Server-side session verification
  - Sensitive token handling via `HttpOnly` cookies
  - Reduced JS bundle size

**Recommendation:** Move authentication, layout shells, and data-fetching pages toward Server Components where possible. Keep Client Components only for interactive sub-trees.

## 2. Authentication and Token Storage (Critical)

### 2.1 Current implementation

`apps/web/lib/api.ts` stores tokens in `localStorage`:

```ts
const TOKEN_KEY = "testra_token";
const REFRESH_TOKEN_KEY = "testra_refresh_token";

export function setAuth(token: string, refreshToken: string) {
  if (typeof window !== "undefined") {
    localStorage.setItem(TOKEN_KEY, token);
    localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken);
  }
}
```

Access token and refresh token are read from `localStorage` on every `apiFetch` call and appended as a `Bearer` `Authorization` header.

### 2.2 Risk

- **XSS:** Any injected script can exfiltrate tokens from `localStorage`.
- **No `HttpOnly` / `Secure` / `SameSite` cookie attributes:** Tokens are accessible from JavaScript and may leak to third-party scripts or through malicious browser extensions.
- **CSRF:** While `Authorization: Bearer <token>` is not automatically sent by the browser, storing tokens in `localStorage` still increases the attack surface compared to short-lived cookies with `SameSite=Lax|Strict`.
- **Token replay after logout:** `localStorage` values can persist across sessions/tabs until explicitly cleared.

### 2.3 Recommended remediation

1. **Migrate to a backend-for-frontend (BFF) pattern or an auth proxy** (e.g., Next.js Route Handlers in `app/api/*`) that sets `HttpOnly`, `Secure`, `SameSite=Lax` cookies for access/refresh tokens.
2. Keep access tokens short-lived (5–15 minutes) and refresh tokens in `HttpOnly` cookies.
3. If a pure SPA token-in-`localStorage` model is intentionally retained, add:
   - strict Content-Security-Policy (`default-src 'self'`, no `unsafe-inline`/`unsafe-eval`)
   - Subresource Integrity for third-party scripts
   - `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY` headers
   - frontend XSS output encoding (React already does basic escaping, but URL and `dangerouslySetInnerHTML` usage must be audited)

This is already captured as a **high-risk open issue** in `SECURITY_REVIEW_v2.md`.

## 3. API Client and State

### 3.1 `lib/api.ts`

- Uses a singleton `fetch` wrapper.
- Automatically refreshes access tokens on `401` (except on `/api/v1/auth/refresh`) using a global `refreshing` promise to prevent thundering herd.
- Clears tokens and redirects to `/login` when refresh fails.
- Returns `body.data` from the API envelope.

### 3.2 Concerns

- **`refreshing` is a module-level mutable variable.** In concurrent tabs or after hot reload, state can drift.
- **`token` is captured once at the top of `apiFetch`**. If a refresh occurs during the call, the retried request still uses the old `token` because `apiFetch` is called recursively with `...options.headers` but `token` is read from closure at the outer call.
- **No request timeout or abort signal support.** Long-running `fetch` calls may hang.
- **No retry/back-off for transient `5xx` / network errors** outside the `401` refresh flow.
- **Type casting:** `return body.data as T` bypasses runtime validation; a Zod schema per endpoint would be safer.

## 4. Route Guards and Authorization

### 4.1 `components/auth/route-guard.tsx`

- Uses `useEffect` to read `isAuthenticated()` from `localStorage` and redirect.
- Has a hydration check via `ready` state.
- Redirect logic uses `window.location.search` to read `returnUrl`, which is safe because it is only used for same-origin redirect.

### 4.2 Concerns

- The guard is **client-side only**. Direct deep links to `/dashboard/...` are not protected until JavaScript runs, causing a flash of unauthenticated content or route leakage.
- `ready` defaults to `false`, so all protected routes render `fallback` (default `null`) on first paint; this may cause layout shift.

**Recommendation:** Add server-side session verification, e.g., via a middleware (`middleware.ts`) that reads the cookie/JWT before rendering the page.

## 5. Workspace and Tenant Context

- `features/notifications/api.ts` and other modules read `testra_workspace_id` from `localStorage`.
- There is no global workspace context/provider to ensure consistent scope, and the workspace ID is not validated against the current user’s memberships before use.
- Requests append `workspace_id` to query strings manually; inconsistencies could lead to wrong-tenant reads if the value is stale.

## 6. UI Components and Accessibility

### 6.1 `components/ui`

- `Button` uses `aria-busy` for loading state and `aria-disabled` for disabled state. Good.
- `Input` generates an `inputId` from the label and associates `<label htmlFor={inputId}>` with the input. Good.
- `Badge`, `Card`, `Switch`, etc., are simple presentational components.
- Tailwind tokens are centrally themed (`brand` palette).

### 6.2 Concerns

- No `aria-live` region for `serverError` in `login/page.tsx`; screen-reader users may not hear login errors.
- Many interactive components are not covered by unit tests (`@testing-library/react` is not in `devDependencies`).
- `Input` type prop may be `password` but there is no show/hide toggle or password-manager hints (`autoComplete` defaults are not set).

## 7. Build and Tooling

- `next.config.ts` supports `output: 'standalone'` for containerized deployments.
- `package.json` has `build`, `lint`, and `typecheck` scripts.
- `tsconfig.json` is standard for Next.js.
- **No visible test runner** (Jest/Vitest not configured). `apps/web` relies entirely on manual or `tsc` checks.

## 8. Security Headers and CSP

- No `next.config.ts` `headers` configuration for `Content-Security-Policy`, `X-Frame-Options`, `Strict-Transport-Security`, `Referrer-Policy`, or `Permissions-Policy`.
- `next/headers` is not used to inject security headers.

**Recommendation:** Add a default security headers map in `next.config.ts` or use a middleware to append them per request. If tokens remain in `localStorage`, a strong CSP is mandatory.

## 9. Findings Summary

| # | Finding | Severity | Recommendation |
|---|---------|----------|----------------|
| 1 | Tokens stored in `localStorage` | **Critical** | Move to `HttpOnly` cookies/BFF, or add CSP + harden XSS defenses |
| 2 | All layouts/pages are Client Components | High | Leverage Server Components for auth shells and data fetching |
| 3 | No server-side route protection / middleware | High | Add `middleware.ts` for auth checks and security headers |
| 4 | `apiFetch` closure-captures stale token; no timeout/retries | Medium | Re-read token after refresh; add `AbortSignal` and retry logic |
| 5 | Workspace ID pulled from `localStorage` with no context validation | Medium | Add workspace context and validate membership server-side |
| 6 | Missing security headers / CSP | Medium | Configure headers in `next.config.ts` or middleware |
| 7 | No frontend unit tests | Medium | Add `@testing-library/react` + Vitest/Jest and cover auth flow |
| 8 | No input type coercion / runtime response validation | Low | Add Zod schemas per API response shape |

## 10. Conclusion

The frontend is visually clean and uses a sensible component structure, but the **token storage and client-only auth model are the dominant production risks**. Fixing token storage should be paired with Server Component adoption and a Next.js middleware for route protection. Once those are in place, the remaining work (tests, headers, response validation, timeout/retry) can be addressed incrementally.
