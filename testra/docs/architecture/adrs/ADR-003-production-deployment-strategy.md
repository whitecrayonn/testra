# ADR-003: Production Deployment Strategy

**Status:** Accepted (Amended by ADR-009)
**Date:** July 2026

## Context

Testra is built by a solo developer with zero infrastructure budget. The deployment strategy must minimize operational cost and complexity while preserving a migration path for future scale. Container orchestration and cloud-managed services (Docker, Kubernetes, Terraform, AWS/GCP/Azure) are explicitly out of scope for MVP and local development because they add cost and operational burden without product value. ADR-009 establishes a native development environment; this ADR establishes the production deployment target.

## Decision

The deployment roadmap is:

| Stage | Compute | PostgreSQL | Redis | Object storage | ClickHouse | TLS/SSL | Secrets |
|---|---|---|---|---|---|---|---|
| Local | Native binaries (Go, Node.js, Python) — see ADR-009 | Local PostgreSQL service | Local Redis service | Local MinIO binary | Optional (not needed until Phase 3) | HTTP only; HTTPS optional | local `.env` file |
| MVP | Single Ubuntu VPS with systemd + Nginx reverse proxy | PostgreSQL on the same VPS | Redis on the same VPS | Local MinIO or filesystem-backed S3-compatible store | Optional (not needed until Phase 3) | Let's Encrypt via certbot on Nginx | environment file or a local secrets store |
| Beta | Single Ubuntu VPS or a small fleet of VPS instances | PostgreSQL on the VPS with streaming backups | Redis on the VPS with persistence | Filesystem/MinIO backups | ClickHouse Cloud only if analytics volume justifies it | Let's Encrypt with optional CDN/WAF | environment file or a local secrets store |
| Enterprise | Single Ubuntu VPS fleet or a managed platform only if an explicit scale/team requirement justifies it | PostgreSQL read replicas or managed PostgreSQL only after measured need | Redis replication or managed Redis only after measured need | Object-store backups with immutability where required | ClickHouse Cloud only if analytics volume justifies it | Let's Encrypt with optional CDN/WAF | environment file or a secrets store; customer-specific secrets where required |

MVP runs the Go API, Go worker, Next.js web app, and Python ML service as systemd services on a single Ubuntu VPS. Nginx terminates TLS (Let's Encrypt) and reverse-proxies to the application services. Migrations run from CI via `apps/api/cmd/migrator` and are never applied manually in production.

Use S3-compatible interfaces in application ports so local MinIO and a future object store remain interchangeable. Keep ClickHouse behind a repository/port boundary so analytics can be deferred until result volume justifies it. Docker, Kubernetes, and Terraform are not used.

## Consequences

A single Ubuntu VPS with systemd minimizes MVP operational cost and eliminates container orchestration and cloud-IaC complexity. The native development environment (ADR-009) removes Docker as a development dependency. Cloud-managed services may be reconsidered only after product-market fit and measured scale justify the budget. Let's Encrypt on Nginx provides free TLS without coupling the application to a CDN or certificate authority.
