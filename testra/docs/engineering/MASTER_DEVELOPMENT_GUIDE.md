# Testra — Master Development Guide

**Status:** Active
**Last Updated:** July 2026
**Classification:** Internal — Engineering

---

## 1. Purpose

This document is the **engineering source of truth** for the Testra platform. It governs how code is written, reviewed, tested, and shipped. All human and AI contributors must follow the governance, principles, and workflows defined here.

 companion documents:
- `docs/engineering/PHASES.md` — implementation roadmap and phase tracking
- `docs/engineering/ENGINEERING_STANDARDS.md` — detailed coding and process standards
- `docs/engineering/progress/` — chronological Engineering Progress Reports

---

## 2. Engineering Principles

1. **One module, one PRD, one owner.** Every feature belongs to exactly one domain module. Never duplicate ownership.
2. **Clean Architecture boundaries.** Domain → Application → Ports → Adapters. Business rules never depend on frameworks or infrastructure.
3. **API-first.** The OpenAPI spec is the contract. Endpoints are documented before implementation. The web UI is just another client.
4. **Privacy-first.** Zero customer source code retention. Zero API collection retention. No external LLM dependency. Tenant isolation enforced at every layer.
5. **Simple first.** Favor simplicity over feature count. Build reusable capabilities before specialized features.
6. **Test what matters.** Domain logic and use cases must have tests. Handlers and repositories are tested via integration tests. No test is deleted or weakened without explicit direction.
7. **Minimal changes.** Prefer focused, surgical edits. Avoid over-engineering. Use single-line changes when sufficient.
8. **No ungrounded assertions.** Every code change must be immediately runnable. Imports at the top. Dependencies declared. No placeholder code.

---

## 3. Development Workflow

### 3.1 Trunk-Based Development

- Short-lived feature branches (`feat/<scope>`, `fix/<scope>`) → PR → squash-merge to `main`.
- `main` is always deployable.
- Releases tagged `vX.Y.Z`.
- Feature flags for incomplete work (unleash or in-app config).

### 3.2 Local Development

Testra uses a **Native Development Environment** (ADR-009). Docker is not required.

1. Install local services: PostgreSQL 16+, Redis 7+, Mailpit, MinIO (see README.md for platform-specific instructions)
2. `pnpm install` — installs JS deps + auto-creates Python venv for ML
3. `pnpm dev` — checks local services, runs migrations, and starts API, web, worker, and ML simultaneously via Turborepo
4. `make test`, `make lint` before push

Docker Compose remains available as an optional alternative in `infra/docker/`.

### 3.3 CI/CD

- **PR pipeline:** lint → unit tests → build → OpenAPI contract validation → security scan
- **Main pipeline:** all above + integration tests → build images (tagged by SHA) → deploy staging → manual promote to production
- Migrations applied by `cmd/migrator` in CI, never manually

### 3.4 Environments

| Environment | Purpose | Data |
|---|---|---|
| `local` | Native development (local PostgreSQL, Redis, Mailpit, MinIO) | Seeded fixtures |
| `staging` | Auto-deploy from `main` | Sanitized/synthetic |
| `production` | Manual promotion | Real, backed up, monitored |

Config via env vars (12-factor). `.env.example` is committed; local values use ignored files; MVP production values use environment files or a secrets manager on the Ubuntu VM; future AWS production uses AWS Secrets Manager encrypted by KMS and task roles.

---

## 4. Architecture Principles

### 4.1 Modular Monolith

- Single deployable Go binary with internal domain modules.
- Each module: `domain.go`, `ports.go`, `repository.go`, `service.go`, `handler.go`, `module.go`.
- Modules communicate via interfaces (ports), never direct imports of another module's internals.
- Extraction to microservices is a future option — boundaries are already defined.

### 4.2 Multi-Tenancy

- `tenant_id` (organization_id) on every tenant-scoped table.
- PostgreSQL RLS is mandatory in staging and production; API roles do not bypass RLS.
- HTTP middleware authenticates and resolves candidate tenant scope; request context propagates it; service layers authorize resource relationships and permissions; repositories set transaction-local `app.tenant_id`.
- Tenant scope must propagate through queues, cache keys, exports, ClickHouse, and ML calls.

### 4.3 Data Layer

- **PostgreSQL** — OLTP, transactional data, audit trails.
- **ClickHouse** — OLAP, test results, time-series events.
- **Redis** — sessions, rate limiting, job queues (Asynq).
- **S3-compatible** — attachments, exports, model artifacts.

### 4.4 ML Boundary

- `apps/ml` (Python/FastAPI) is inference + training only.
- Called by Go API (sync) or worker (async) over internal network.
- Per-tenant models. Never receives source code or API payloads.

---

## 5. Definition of Done (DoD)

A task is **Done** when all of the following are true:

- [ ] Code implements the approved spec/PRD requirements
- [ ] OpenAPI spec updated if endpoints changed
- [ ] Unit tests written for domain/application logic
- [ ] All existing tests pass (`go test`, `pnpm test`)
- [ ] Linting passes (`go vet`, `eslint`, `ruff`)
- [ ] No secrets or sensitive data in code or commits
- [ ] Migration files created if schema changed (up + down)
- [ ] `.env.example` updated if new env vars introduced
- [ ] `PHASES.md` updated if phase status changed
- [ ] Engineering Progress Report saved if session produced meaningful work

---

## 6. Self-Review Process

Before requesting review or merging a PR:

1. **Read the diff.** Every line. Ask: "Would I approve this if someone else wrote it?"
2. **Check imports.** Are they at the top? Are they used? Are they the right packages?
3. **Check error handling.** Are errors wrapped with context? Are sentinel errors used correctly?
4. **Check tests.** Do they test behavior, not implementation? Do they cover edge cases?
5. **Check naming.** Are names clear and consistent with the codebase?
6. **Check boundaries.** Does the change respect module boundaries? No cross-module internal imports?
7. **Check security.** No hardcoded secrets. No SQL injection. Input validation on every endpoint.
8. **Check the OpenAPI spec.** If endpoints changed, is the spec updated?
9. **Run locally.** `make test && make lint` must pass.
10. **Update docs.** If governance or standards changed, update the relevant doc.

---

## 7. Engineering Progress Report Template

Every implementation session that produces meaningful work must save a progress report to `docs/engineering/progress/` using the filename format `YYYY-MM-DD-HHMM-session.md`.

```markdown
# Engineering Progress Report — YYYY-MM-DD HH:MM

## Session Summary
<one-sentence summary>

## Completed
- <item>

## In Progress
- <item>

## Blocked
- <item or "None">

## Next
- <item>

## Files Changed
- <path> — <description>

## Verification
- <command> — <result>
```

---

## 8. Rules for Future AI and Human Contributors

### 8.1 Source of Truth

- `MASTER_DEVELOPMENT_GUIDE.md` (this file) — engineering governance. Update only when governance changes.
- `PHASES.md` — implementation roadmap. Update when phases start or complete.
- `ENGINEERING_STANDARDS.md` — coding standards. Update only when standards change.
- `docs/engineering/progress/` — progress reports. Append-only, never edit past reports.
- Approved product/architecture documents — **read-only**. Changes require a new ADR.

### 8.2 AI Contributor Rules

- Always read `PHASES.md` before starting work to understand current state.
- Always read `ENGINEERING_STANDARDS.md` before writing code.
- Never modify approved product or architecture documents without creating an ADR.
- Never skip tests. Never delete or weaken tests.
- Never commit secrets. Never hardcode API keys.
- Always update `PHASES.md` when a phase starts or completes.
- Always save a progress report at the end of a session.
- Prefer minimal, focused edits. Follow existing code patterns.
- Imports at the top of the file. Always.

### 8.3 Human Contributor Rules

- Review every AI-generated PR with the self-review process (§6).
- Approve PRs only when DoD is met (§5).
- Keep `PHASES.md` honest — if work is incomplete, mark it as such.
- Create an ADR for any architectural decision that deviates from approved documents.

---

## 9. Document Maintenance

| Document | When to Update | Who |
|---|---|---|
| `MASTER_DEVELOPMENT_GUIDE.md` | Engineering governance changes | CTO / Lead |
| `PHASES.md` | Phase starts, completes, or scope changes | Any contributor |
| `ENGINEERING_STANDARDS.md` | Coding or process standards change | CTO / Lead |
| `docs/engineering/progress/*` | End of each implementation session | Session owner |
| ADRs (`docs/architecture/adrs/`) | Architectural deviation from approved docs | CTO / Lead |
