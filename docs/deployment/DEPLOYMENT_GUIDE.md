# Testra Deployment Guide

**Purpose:** How to run Testra locally today, and how to deploy it to a single rented VM when the time comes.
**Scope:** Local development (current focus) and a single-VM MVP deployment (not yet built).
**Source of Truth:** This guide; architecture decisions live in [`ADR-003`](../architecture/adrs/ADR-003-production-deployment-strategy.md) and [`ADR-009`](../architecture/adrs/ADR-009-native-development-environment.md).
**Last Updated:** August 2026

## Deployment Model

Testra runs as **plain native processes** — a Go API binary, a Go worker binary, a Next.js server, and (optionally) a Python ML service. There is no container runtime anywhere in the stack.

**Docker, Kubernetes, Terraform, and cloud-managed services are not used.** This is a deliberate decision (ADR-003), not an outstanding gap. Local development and production run the same binaries; only the machine differs.

| Stage | Compute | Database | Cache | Storage | TLS |
|---|---|---|---|---|---|
| **Local** (current) | Native processes via `pnpm dev` | Local PostgreSQL | Local Redis (optional — auto-started) | Local filesystem | None (HTTP on localhost) |
| **MVP** (planned) | Single rented VM | PostgreSQL on the same VM | Redis on the same VM (optional) | Local filesystem | Reverse proxy with an ACME certificate |

> **Target OS for the VM is deliberately undecided.** Nothing in the application depends on it: Go cross-compiles to any target with `GOOS`/`GOARCH`, and Next.js and PostgreSQL run everywhere. Pick the OS when you rent the machine, then write the process-supervision config for whatever it runs (`systemd`, Windows Services, or similar).

---

## Local Development

This is the supported, verified path.

```bash
pnpm install
pnpm dev
```

`pnpm dev` will:

1. Verify PostgreSQL is reachable on `localhost:5432`
2. Apply database migrations via `apps/api/cmd/migrator`
3. Start Redis automatically (via `@testra/redis-dev`), then launch the API, web, and worker through Turborepo

### Prerequisites

| Tool | Required? | Notes |
|---|---|---|
| Go 1.24+ | **Yes** | Runs the API, worker, and migrations. `pnpm dev` fails fast if missing |
| Node.js 20+ / pnpm 9+ | **Yes** | Web app and tooling |
| PostgreSQL 16+ | **Yes** | Must be running with a `testra` database and user before `pnpm dev` |
| Redis | No | Auto-started when installed. Without it the API falls back to an in-memory rate limiter |
| Python 3.12+ | No | Only for the ML service (`pnpm dev:all`) |
| MinIO / Mailpit | No | Only for S3 and email flows |

### Service Ports

| Service | Port |
|---|---|
| Go API | 8080 |
| Next.js Web | 3000 |
| PostgreSQL | 5432 |
| Redis | 6379 |
| ML service | 8000 |

See [`README.md`](../../README.md) for per-platform installation instructions.

---

## Deploying to a VM

> **Not yet implemented.** No provisioning scripts, service definitions, or reverse-proxy config exist. This section describes the intended shape so the work is unambiguous when you start it.

### One-time server setup

1. Rent the VM and enable a firewall allowing only SSH/RDP and 80/443.
2. Install PostgreSQL 16+; create the `testra` database and a dedicated user with a strong password.
3. Install a reverse proxy (nginx, Caddy, or IIS) and obtain a TLS certificate.
4. Create a non-privileged user to own and run the application processes.
5. Create the environment file with production values — see *Configuration* below.

### Each deployment

1. Build the artifacts on your machine, targeting the VM's OS/architecture:
   ```bash
   GOOS=linux GOARCH=amd64 go build -o dist/api ./apps/api/cmd/api
   GOOS=linux GOARCH=amd64 go build -o dist/worker ./apps/api/cmd/worker
   GOOS=linux GOARCH=amd64 go build -o dist/migrator ./apps/api/cmd/migrator
   pnpm --filter @testra/web build
   ```
   (Adjust `GOOS`/`GOARCH` to match the VM; drop them entirely if it matches your machine.)
2. Copy the artifacts to the VM.
3. **Back up the database** before applying migrations.
4. Run migrations with the `migrator` binary. Never apply schema changes by hand.
5. Restart the API, worker, and web processes.
6. Verify: `curl https://your-domain/health` should return `{"data":{"status":"ok"}}`, then check that the worker is processing jobs and log in through the UI.

### Rollback

Rollback is only safe when the schema is still compatible with the older binary. Prefer backward-compatible (expand/contract) migrations so the previous version keeps working. For a destructive migration, the recovery path is a forward fix or a restore from the pre-deployment backup — decide which **before** deploying, not during an incident.

---

## Configuration

All configuration is environment variables; see [`.env.example`](../../.env.example) and [`apps/api/.env.example`](../../apps/api/.env.example) for the full list.

**Never commit real secrets.** `JWT_PRIVATE_KEY`, `DATABASE_URL`, SMTP credentials, S3 keys, and integration tokens belong in the VM's environment file (readable only by the app user) or a secrets store.

The API refuses to start in production when configuration is unsafe — it rejects example credentials, `sslmode=disable`, and a missing `DATABASE_URL`. See `apps/api/internal/shared/config/config.go`.

---

## Before Going Live

Work through these when you actually deploy — they are the parts that are genuinely hard to retrofit:

- [ ] Automated PostgreSQL backups, **with a restore actually tested at least once**
- [ ] TLS certificate auto-renewal verified
- [ ] Secrets outside the repo, file permissions locked down
- [ ] Tenant isolation (RLS) verified against the production database
- [ ] Log rotation configured so the disk cannot fill
- [ ] A known-good rollback path for the current release

[`PRODUCTION_READINESS_CHECKLIST.md`](../operations/PRODUCTION_READINESS_CHECKLIST.md) has the exhaustive version. The list above is the subset that matters for a first launch.

---

## Appendix: Scaling Beyond One VM

Deliberately deferred. Revisit only when a measured limit is hit, not preemptively:

| Trigger | Consideration |
|---|---|
| Single VM saturated | Split web/API onto separate VMs, or scale the VM up first (usually cheaper and simpler) |
| Database is the bottleneck | Add read replicas, or move to managed PostgreSQL |
| Analytics queries slow down the primary | Introduce ClickHouse behind the existing repository port (ADR-010) |
| Team grows past solo | Add a staging environment and a formal promotion process |
| Compliance requirements | Revisit managed services, audit tooling, and immutable backups |

Adding any of this before there is a measured need buys operational burden with no product value.

---

## See Also

- [`BIBLICAL_TESTRA.md`](../BIBLICAL_TESTRA.md) — engineering handbook
- [`ADR-003`](../architecture/adrs/ADR-003-production-deployment-strategy.md) — why no Docker/Kubernetes
- [`ADR-009`](../architecture/adrs/ADR-009-native-development-environment.md) — native development environment
- [`DISASTER_RECOVERY_GUIDE.md`](../operations/DISASTER_RECOVERY_GUIDE.md) — backup and recovery
