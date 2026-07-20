# Testra Deployment Guide

**Purpose:** Describe the production deployment model, promotion sequence, configuration, rollback, and operational runbooks for a single Ubuntu VPS.
**Owner:** Platform / SRE Lead
**Scope:** Local development, MVP production on a single Ubuntu VPS, and future migration options.
**Source of Truth:** `DEPLOYMENT_GUIDE.md`; canonical architecture decisions are in `docs/architecture/adrs/ADR-003-production-deployment-strategy.md` and `docs/architecture/adrs/ADR-009-native-development-environment.md`.
**Last Updated:** July 2026
**Related documents:**
- [`BIBLICAL_TESTRA.md`](../BIBLICAL_TESTRA.md)
- [`ADR-003-production-deployment-strategy.md`](../architecture/adrs/ADR-003-production-deployment-strategy.md)
- [`ADR-009-native-development-environment.md`](../architecture/adrs/ADR-009-native-development-environment.md)
- [`DISASTER_RECOVERY_GUIDE.md`](../operations/DISASTER_RECOVERY_GUIDE.md)
- [`PRODUCTION_READINESS_CHECKLIST.md`](../operations/PRODUCTION_READINESS_CHECKLIST.md)

## Deployment Model

Testra's production target is a **single Ubuntu VPS** managed with `systemd` and `nginx`. Local development runs the same binaries natively on a developer machine. Container orchestration, Kubernetes, Terraform, Docker, and cloud-managed services are **not used** for MVP.

| Stage | Compute | Database | Cache | Storage | Analytics | Reverse Proxy / TLS |
|---|---|---|---|---|---|---|
| Local | Native binaries (Go, Node.js, Python) | Local PostgreSQL | Local Redis | Local MinIO or filesystem | ClickHouse (optional) | HTTP on localhost |
| MVP | Single Ubuntu VPS + systemd | PostgreSQL on the same VPS | Redis on the same VPS | MinIO on the same VPS or filesystem-backed S3-compatible store | ClickHouse (optional) | Nginx + Let's Encrypt |
| Beta | Single VPS or small VPS fleet | PostgreSQL with backups | Redis with persistence | Filesystem / MinIO backups | ClickHouse Cloud only if justified | Let's Encrypt + optional CDN/WAF |
| Enterprise | Single VPS fleet or managed platform only if justified | Managed PostgreSQL only after measured need | Managed Redis only after measured need | Object-store backups with immutability | ClickHouse Cloud only if justified | Let's Encrypt + optional CDN/WAF |

MVP runs the Go API, Go worker, Next.js web app, and Python ML service as `systemd` services. Nginx terminates TLS and reverse-proxies to the application services. Migrations run from CI/CD via `apps/api/cmd/migrator` and are never applied manually in production.

## MVP Service Architecture

| Service | Process Manager | Port | Notes |
|---|---|---|---|
| Go API | `systemd` unit | 8080 | Compiled `linux/amd64` Go binary |
| Go Worker | `systemd` unit | — | Background job processor (currently a stub) |
| Next.js Web | `systemd` unit | 3000 | Standalone `next` build |
| Python ML | `systemd` unit | 8000 | `uvicorn` behind `systemd` |
| Nginx | `systemd` | 80 / 443 | TLS termination, reverse proxy, static file serving |
| PostgreSQL | `systemd` | 5432 | Application database |
| Redis | `systemd` | 6379 | Sessions, rate limits, future queues |
| MinIO | `systemd` | 9002 / 9001 | S3-compatible object storage for artifacts |

## Promotion Sequence

1. Merge only reviewed, passing changes to `main`.
2. Build immutable artifacts (compiled binaries, web standalone build) and record the commit SHA.
3. Validate OpenAPI, tests, security scans, and the migration plan.
4. Deploy the same artifacts to a staging VPS.
5. Apply migrations through `apps/api/cmd/migrator` in the deployment pipeline.
6. Run smoke tests for health, authentication, tenancy, and core API paths.
7. Observe staging metrics/logs for the agreed soak period.
8. Obtain release approval and promote the same artifacts to the production VPS.
9. Verify deployment, migrations, background workers, queues, and critical user journeys.
10. Record outcome and rollback / forward-fix decision.

## Configuration and Secrets

MVP configuration is injected through environment files on the Ubuntu VPS. Secrets (`JWT_SECRET`, `DATABASE_URL`, SMTP credentials, S3 keys, integration credentials) are never committed. Local development uses ignored `.env` files and non-production credentials.

- TLS is terminated by Nginx with a Let's Encrypt certificate.
- All services communicate over localhost or Unix sockets; no public exposure except Nginx ports 80/443.
- Backups and log rotation are handled by standard Linux tools (`cron`, `logrotate`, `pg_dump`, `restic`, etc.).

## Rollback

Application rollback is safe only when schema compatibility is preserved. Prefer backward-compatible expand/contract migrations. If a migration is destructive or irreversible, rollback must be a forward fix or a restore plan approved before release.

## Deployment Gates

Use `PRODUCTION_READINESS_CHECKLIST.md`, `RELEASE_CHECKLIST.md`, and `SECURITY_CHECKLIST.md`. No production deployment is approved when tenant isolation, backup verification, observability, migration recovery, or critical security controls are unverified.

## Local Development

Local development uses a native environment. See `README.md` and `ADR-009` for prerequisites. `pnpm dev` checks services, runs migrations, and starts the API, web, worker, and ML services via `turbo`.

## Current Infrastructure Status

- **No systemd service unit files** for API, web, worker, or ML.
- **No nginx site configuration** or TLS automation (certbot).
- **No VPS provisioning or deployment runbooks** exist yet.
- **No CD pipeline**; GitHub Actions only builds/tests code.
- `scripts/dev/` contains local development helpers only.

## Findings & Recommendations

1. **No production systemd unit files or nginx config.** Create `docs/deployment/systemd/` with service units and an nginx site template.
2. **No deployment runbook.** Document server setup, package installation, firewall (`ufw`), PostgreSQL/Redis/MinIO install, artifact delivery, and migration steps.
3. **No CD pipeline.** Add a GitHub Actions workflow that builds binaries, uploads artifacts, and triggers a deployment script on the VPS.
4. **No observability stack.** Add OpenTelemetry, Prometheus, Grafana, or Loki; start with structured logs and basic `systemd` status checks.
5. **No dependency/security scanning.** Add `dependabot`, `trivy`, `gosec`, or GitHub Advanced Security scans.
6. **Environment variable drift.** Align `apps/api/.env.example` with the code: rename `JWT_EXPIRY_HOURS` to `JWT_EXPIRY_MINUTES` and remove unused variables.
7. **No staging/production secrets management.** Document how secrets are injected via environment files or a local secrets store.
8. **No web environment example.** Add `apps/web/.env.example` with `NEXT_PUBLIC_API_URL` and any public runtime config.
9. **CI does not test the web build with a running API.** Add integration tests that spin up the API and exercise critical frontend flows.
10. **Worker is a stub.** Decide whether `apps/worker` is needed; if so, implement it and add a `systemd` unit.

## See Also

- [`BIBLICAL_TESTRA.md`](../BIBLICAL_TESTRA.md) — canonical engineering handbook
- [`DISASTER_RECOVERY_GUIDE.md`](../operations/DISASTER_RECOVERY_GUIDE.md) — backup and recovery
- [`PRODUCTION_READINESS_CHECKLIST.md`](../operations/PRODUCTION_READINESS_CHECKLIST.md) — go-live gates
