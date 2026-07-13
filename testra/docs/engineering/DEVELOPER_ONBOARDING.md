# Testra Developer Onboarding Guide

## Before Starting

Read, in order:

1. `docs/engineering/MASTER_DEVELOPMENT_GUIDE.md`
2. `docs/engineering/PHASES.md`
3. `docs/engineering/ENGINEERING_STANDARDS.md`
4. `docs/architecture/README.md`
5. The relevant API, database, and module documents

Confirm the task belongs to the current phase and one owning module. Do not implement work described only as an assumption.

## Repository Orientation

- `apps/api/` — Go modular-monolith API, worker, and migrator.
- `apps/web/` — Next.js web application.
- `apps/ml/` — Python ML service.
- `packages/` — shared tooling, UI, types, and SDK.
- `infra/` — Docker (optional), Kubernetes, and Terraform.
- `docs/` — architecture, contracts, governance, and operations.

## Local Setup

Install local services (PostgreSQL, Redis, Mailpit, MinIO) per the README instructions. Run `pnpm install` then `pnpm dev` to start all services. Docker is optional (ADR-009). Do not place real credentials in committed files.

## First Contribution

1. Create a short-lived branch using the naming rules.
2. Identify the owning module and source-of-truth documents.
3. Update OpenAPI/database/ADR documentation before implementation when applicable.
4. Implement the smallest coherent change within module boundaries.
5. Add unit, integration, or contract tests appropriate to the change.
6. Run formatting, linting, tests, and relevant builds.
7. Perform the self-review checklist.
8. Update phase/progress documentation when the work changes project status.

## Review Expectations

Reviewers check correctness, tenant isolation, authorization, error semantics, migration safety, tests, API compatibility, secret handling, and documentation. “It works locally” is not sufficient for production-facing changes.

## Escalation

Create or request an ADR for architecture deviations, new cross-module dependencies, storage changes, authentication changes, or decisions that affect future extraction/deployment. Apply ADR-003 through ADR-009 as mandatory architecture; implementation details that remain unspecified must be validated against those decisions and recorded in the owning progress report.
