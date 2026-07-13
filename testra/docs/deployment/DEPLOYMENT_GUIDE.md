# Testra Deployment Guide

## Deployment Model

The deployment roadmap is defined by ADR-003 (amended by ADR-009). Local uses native development with locally installed services. MVP uses an Ubuntu VM with systemd and Nginx. Beta adds multi-AZ compute and managed data services. Enterprise uses private AWS networking with dedicated capacity and optional EKS.

| Stage | Compute | Database | Cache | Storage | Analytics | Reverse Proxy/CDN |
|---|---|---|---|---|---|---|
| Local | Native (Go, Node.js, Python) | Local PostgreSQL | Local Redis | MinIO binary | ClickHouse (optional) | local HTTP |
| MVP | Ubuntu VM + systemd | PostgreSQL (local or managed) | Redis (local or managed) | MinIO (optional) or S3 | ClickHouse (optional until Phase 3) | Nginx (TLS) |
| Beta | Ubuntu VMs multi-AZ or AWS ECS Fargate | RDS Multi-AZ | ElastiCache replication | S3 with versioning | ClickHouse Cloud | Cloudflare CDN/WAF + ACM |
| Enterprise | Private AWS networking; EKS optional | RDS Multi-AZ/read replicas | ElastiCache dedicated | S3 replication, Object Lock | ClickHouse Cloud dedicated | Cloudflare Enterprise + ACM |

MVP runs Go API, Go worker, Next.js, and Python ML as systemd services on a single Ubuntu VM. Nginx terminates TLS and reverse-proxies to the application services. Docker files remain in the repository as optional deployment assets but are not required.

## MVP Service Architecture

| Service | Process Manager | Port | Notes |
|---|---|---|---|
| Go API | systemd unit | 8080 | Compiled Go binary |
| Go Worker | systemd unit | — | Background job processor |
| Next.js Web | systemd unit or PM2 | 3000 | Standalone build |
| Python ML | systemd unit | 8000 | uvicorn behind systemd |
| Nginx | systemd | 80/443 | TLS termination, reverse proxy |

## Promotion Sequence

1. Merge only a reviewed, passing change to `main`.
2. Build immutable artifacts and record commit SHA.
3. Validate OpenAPI, tests, security scans, and migration plan.
4. Deploy to staging.
5. Apply migrations through the migrator in the deployment pipeline.
6. Run smoke tests for health, authentication, tenancy, and core API paths.
7. Observe staging metrics/logs for the agreed soak period.
8. Obtain release approval and promote the same artifacts to production.
9. Verify deployment, migrations, background workers, queues, and critical user journeys.
10. Record outcome and rollback/forward-fix decision.

## Configuration and Secrets

MVP configuration is injected through environment files on the Ubuntu VM. Future AWS production uses AWS Secrets Manager, encrypted with AWS KMS, and accessed through task roles. Cloudflare, RDS, ElastiCache, S3, ClickHouse Cloud, SMTP, JWT signing, and integration credentials are never committed. Local development uses ignored environment files and non-production credentials. TLS is terminated at Nginx (MVP) or Cloudflare/CloudFront and ACM (future AWS stages); private service traffic remains inside private networking.

## Rollback

Application rollback is safe only when schema compatibility is preserved. Prefer backward-compatible expand/contract migrations. If a migration is destructive or irreversible, rollback must be a forward fix or restore plan approved before release.

## Deployment Gates

Use `PRODUCTION_READINESS_CHECKLIST.md`, `RELEASE_CHECKLIST.md`, and `SECURITY_CHECKLIST.md`. No production deployment is approved when tenant isolation, backup verification, observability, migration recovery, or critical security controls are unverified.

Kubernetes (EKS) remains an optional enterprise evolution target only when measured scale, scheduling needs, or organizational capability justify its additional operational burden.
