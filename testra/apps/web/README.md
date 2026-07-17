# Testra Web

The **Testra Web** application is a Next.js 15 + React + TypeScript frontend for the Testra platform.

## What lives here

- `app/` — Next.js App Router pages and layouts.
  - `(auth)/` — login, register, MFA setup, password reset.
  - `(dashboard)/` — dashboard, projects, test-cases, test-runs, settings.
  - `onboarding/` — first organization/workspace setup.
- `components/` — shared UI components.
- `features/` — domain API wrappers and hooks.
- `lib/api.ts` — API client that talks to `NEXT_PUBLIC_API_URL`.
- `types/` — TypeScript type definitions.

## Environment

The app expects `NEXT_PUBLIC_API_URL` (default `http://localhost:8080`) pointing to the Go API.

## Running locally

See `docs/engineering/LOCAL_DEVELOPMENT_GUIDE.md`. In short:

```bash
# from the repository root
pnpm dev
# or
cd apps/web && pnpm dev
```

## Canonical documentation

- [Engineering Handbook](docs/BIBLICAL_TESTRA.md)
- [Frontend Audit](docs/handover/frontend-audit.md)
- [Local Development Guide](docs/engineering/LOCAL_DEVELOPMENT_GUIDE.md)
- [OpenAPI Contract](docs/api/openapi/openapi.yaml)
