# Engineering Progress Report — 2026-07-13 23:20

## Session Summary
Unified local development experience: a single `pnpm install` + `pnpm dev` now starts the entire Testra stack (infrastructure + all four applications).

## Completed
- Created `package.json` for `apps/api`, `apps/worker`, `apps/ml` so Turborepo can orchestrate them
- Created cross-platform Node.js dev scripts:
  - `scripts/dev/start-infra.mjs` — starts Docker, waits for PostgreSQL health, runs migrations
  - `scripts/dev/go-dev.mjs` — runs Go apps with `air` hot reload (auto-detected) or falls back to `go run`
  - `scripts/dev/ml-dev.mjs` — runs FastAPI with `uvicorn --reload` using the `.venv` Python
  - `scripts/dev/setup-python.mjs` — auto-creates Python venv and installs ML deps on `pnpm install`
  - `scripts/dev/clean.mjs` — stops Docker + removes all build artifacts
- Updated root `package.json`: `dev` now runs infra startup then `turbo run dev`; added `clean` and `postinstall` scripts
- Updated `turbo.json`: added `clean` task
- Created `apps/api/.air.toml` for Go API hot-reload configuration
- Updated `Makefile` to delegate to `pnpm` commands
- Updated `README.md` with full Local Development section: prerequisites, one-command setup, service ports table, env vars
- Updated `.env.example` with ML service URL
- Updated `.gitignore` with `apps/api/tmp/` (air build output)
- Simplified legacy PowerShell/shell scripts to delegate to unified `pnpm` commands

## In Progress
- Phase 1: Identity & Tenancy — RBAC route enforcement, OpenAPI spec update, SMTP email sending

## Blocked
- None

## Next
- Phase 2: Test Management Core

## Files Changed
- `apps/api/package.json` — created
- `apps/worker/package.json` — created
- `apps/ml/package.json` — created
- `apps/api/.air.toml` — created
- `scripts/dev/start-infra.mjs` — created
- `scripts/dev/go-dev.mjs` — created
- `scripts/dev/ml-dev.mjs` — created
- `scripts/dev/setup-python.mjs` — created
- `scripts/dev/clean.mjs` — created
- `package.json` — updated (dev, clean, postinstall scripts)
- `turbo.json` — updated (clean task)
- `Makefile` — updated (delegates to pnpm)
- `README.md` — updated (Local Development section with ports table)
- `.env.example` — updated (ML_SERVICE_URL)
- `.gitignore` — updated (apps/api/tmp/)
- `scripts/dev/dev.ps1` — simplified (delegates to pnpm dev)
- `scripts/dev/install.ps1` — simplified (delegates to pnpm install)
- `scripts/dev/setup.sh` — simplified (delegates to pnpm install)

## Verification
- `pnpm install` — pass (9 workspace projects, postinstall runs setup-python.mjs)
- `npx turbo run dev --dry-run` — all 4 target apps recognized:
  - `@testra/api#dev` → `node ../../scripts/dev/go-dev.mjs cmd/api`
  - `@testra/web#dev` → `next dev`
  - `@testra/worker#dev` → `node ../../scripts/dev/go-dev.mjs cmd/worker`
  - `@testra/ml#dev` → `node ../../scripts/dev/ml-dev.mjs`
- All Node.js scripts pass `node --check` syntax validation
