# ADR-009: Native Development Environment

**Status:** Accepted
**Date:** July 2026

## Context

The primary development machine is Windows without virtualization support. Docker Desktop cannot be used reliably and should never become a blocker for development. ADR-003 previously specified Docker Compose as the local development environment. This created a hard dependency on Docker Desktop for every developer on the team.

## Decision

Replace Docker Compose with a **Native Development Environment** as the official local development workflow.

### Local Development Stack

Developers install and run the following services natively:

| Service | Requirement Level | Notes |
|---|---|---|
| Go 1.23+ | Required | API, worker, migrator |
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

- Docker and Docker Compose are **optional** deployment and development assets.
- Docker files remain in the repository under `infra/docker/` for optional use.
- Docker Desktop is **not required** for any development workflow.
- Docker may be used by developers who prefer it, but it is not the official path.

### Production Strategy Update

The MVP deployment target is updated from AWS ECS Fargate to:

| Component | Technology |
|---|---|
| Compute | Ubuntu VM |
| Process manager | systemd |
| Reverse proxy | Nginx |
| API | Go binary (systemd service) |
| Worker | Go binary (systemd service) |
| Web | Next.js standalone (systemd service or PM2) |
| ML | Python FastAPI (systemd service) |
| Database | PostgreSQL (local or managed) |
| Cache | Redis (local or managed) |
| Object storage | MinIO (optional) or S3 |

Kubernetes remains a future enterprise deployment target after product-market fit. AWS managed services remain a future evolution path when scale justifies the operational investment.

## Consequences

- **Positive:** Development is unblocked on machines without virtualization support. No Docker Desktop dependency. Simpler onboarding for new developers.
- **Positive:** Production MVP deployment is simpler and lower-cost on a single Ubuntu VM with systemd.
- **Negative:** Developers must install and manage local PostgreSQL, Redis, Mailpit, and MinIO instances.
- **Mitigation:** Document clear installation steps for each platform. Provide helper scripts for starting/stopping local services.
- **Negative:** Environment consistency across developers is less guaranteed than with Docker Compose.
- **Mitigation:** Document required versions and provide configuration validation in dev scripts.
- **Supersedes:** ADR-003's local development stage (Docker Compose) is replaced by native development. ADR-003's MVP/Beta/Enterprise AWS stages remain as future evolution paths.
