# Testra Local Development Guide

## Supported Local Shape

Local development uses a native development environment (ADR-009). PostgreSQL, Redis, Mailpit, and MinIO are installed and run as local services or binaries. ClickHouse is optional until Phase 3. Docker Compose remains available as an optional alternative in `infra/docker/` but is not required. Application services are started through the repository's documented development commands. Exact ports and credentials are defined by the current environment example file; do not copy production secrets locally.

## Standard Workflow

1. Install local services: PostgreSQL 16+, Redis 7+, Mailpit, MinIO (see README.md for platform-specific instructions).
2. Install repository dependencies: `pnpm install`.
3. Copy the environment example to a local, ignored environment file and set only local values.
4. Start local services (PostgreSQL, Redis, Mailpit, MinIO).
5. Run `pnpm dev` — this checks service availability, applies migrations, and starts API, web, worker, and ML services.
6. Exercise the API through the documented `v1` contract.
7. Run focused tests, then the full required checks before opening a PR.

## Data Safety

- Use synthetic fixtures only.
- Never import customer production data into a workstation without explicit approval and redaction.
- Treat local object storage, email previews, and logs as sensitive.
- Reset local state only with approved repository scripts; do not manually delete shared environments.

## Debugging Boundaries

API failures: inspect request ID, structured logs, auth context, and database connectivity. Queue failures: inspect Redis health and worker backlog. Frontend failures: inspect API base URL and browser network requests. ML failures: verify service availability and model artifact configuration.

## Common Development Checks

- Go formatting, vetting, unit tests, and race-sensitive tests as required by the standards.
- Web typechecking and tests through the repository's package manager.
- Python linting/tests for ML changes.
- OpenAPI validation for contract changes.
- Migration up/down review for schema changes.

Executable command names are authoritative in the repository README, Makefile, and development scripts. This guide defines workflow and safety requirements; it does not replace those command sources.
