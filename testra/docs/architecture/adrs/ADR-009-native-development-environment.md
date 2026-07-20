# ADR-009: Native Development Environment

**Status:** Accepted
**Date:** July 2026

## Context

The primary development machine is Windows without virtualization support. Docker cannot be used reliably and should never become a blocker for development. ADR-003 previously specified Docker Compose as the local development environment. This created a hard dependency on Docker for every developer on the team and did not match the zero-budget, single-Ubuntu-VPS production target.

## Decision

Replace Docker Compose with a **Native Development Environment** as the official local development workflow.

### Local Development Stack

Developers install and run the following services natively:

| Service | Requirement Level | Notes |
|---|---|---|
| Go 1.24+ | Required | API, worker, migrator |
| Node.js 20+ | Required | Web app, dev scripts |
| pnpm 9.5+ | Required | Monorepo package manager |
| Python 3.12+ | Required for ML | Optional if ML service is not needed |
| PostgreSQL 16 | Required | Local service or binary |
| Redis 7 | Required | Local service or binary |
| Mailpit | Required | Binary; SMTP testing |
| MinIO | Required | Binary; S3-compatible local storage |
| ClickHouse 24 | Optional | Not needed until Phase 3 |

### Development Workflow

The official workflow is:

```
pnpm install
pnpm dev
```

This starts API, worker, Next.js web, and ML service using locally installed dependencies and locally running database services. No Docker is required.

### Docker Status

- Docker and Docker Compose are **not used** for local development or production.
- No Docker files or images remain in the repository.
- Native binaries and locally installed services are the only supported local workflow.

### Production Strategy Update

The MVP deployment target is a single Ubuntu VPS managed with systemd and Nginx:

| Component | Technology |
|---|---|
| Compute | Single Ubuntu VPS |
| Process manager | systemd |
| Reverse proxy | Nginx |
| API | Go binary (systemd service) |
| Worker | Go binary (systemd service) |
| Web | Next.js standalone (systemd service) |
| ML | Python FastAPI (systemd service) |
| Database | PostgreSQL on the same VPS |
| Cache | Redis on the same VPS |
| Object storage | Local MinIO or filesystem-backed S3-compatible store |

Cloud-managed services and container orchestration (Kubernetes, Terraform, AWS/GCP/Azure) are out of scope for MVP. They may be reconsidered only after product-market fit and measured scale justify the budget.

## Consequences

- **Positive:** Development is unblocked on machines without virtualization support. No Docker dependency. Simpler onboarding for new developers.
- **Positive:** Production MVP deployment is simpler and lower-cost on a single Ubuntu VPS with systemd.
- **Negative:** Developers must install and manage local PostgreSQL, Redis, Mailpit, and MinIO instances.
- **Mitigation:** Document clear installation steps for each platform. Provide helper scripts for starting/stopping local services.
- **Negative:** Environment consistency across developers is less guaranteed than with Docker Compose.
- **Mitigation:** Document required versions and provide configuration validation in dev scripts.
- **Supersedes:** ADR-003's local development stage (Docker Compose) is replaced by native development. ADR-003's cloud-managed stages are replaced by the single-Ubuntu-VPS target with an optional future managed-platform migration.
