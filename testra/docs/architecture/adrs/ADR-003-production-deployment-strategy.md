# ADR-003: Production Deployment Strategy

**Status:** Accepted (Amended by ADR-009)
**Date:** July 2026

## Context

Testra needs low operational cost for a solo developer while preserving privacy, managed reliability, enterprise networking, and a migration path for scale. Operating Kubernetes at MVP would add cost and operational burden without product value. ADR-009 replaces the local Docker Compose stage with a native development environment and updates the MVP deployment target from AWS ECS Fargate to a simpler Ubuntu VM with systemd.

## Decision

The deployment roadmap is:

| Stage | Compute | PostgreSQL | Redis | Object storage | ClickHouse | CDN/SSL | Secrets |
|---|---|---|---|---|---|---|---|
| Local | Native (Go, Node.js, Python binaries) — see ADR-009 | Local PostgreSQL service | Local Redis service | MinIO binary | Optional (not needed until Phase 3) | local HTTP; HTTPS optional | ignored local env file |
| MVP | Ubuntu VM with systemd + Nginx reverse proxy | PostgreSQL (local or managed) | Redis (local or managed) | MinIO (optional) or S3 | ClickHouse (optional until Phase 3) | Nginx TLS or Cloudflare | environment files or secrets manager |
| Beta | Ubuntu VMs across multiple AZs or AWS ECS Fargate behind ALB | RDS Multi-AZ | ElastiCache replication/failover | S3 versioning and lifecycle policies | ClickHouse Cloud production service | Cloudflare CDN/WAF plus ACM-managed TLS | AWS Secrets Manager with rotation |
| Enterprise | Private networking and dedicated capacity; EKS only if an explicit scale/team requirement justifies it | RDS Multi-AZ/read replicas or Aurora PostgreSQL after measured need | ElastiCache replication/failover or dedicated cluster | S3 replication, Object Lock where contracted | ClickHouse Cloud with private connectivity/dedicated resources | Cloudflare Enterprise controls plus ACM | AWS Secrets Manager and KMS, customer-specific secrets where required |

MVP runs Go API, Go worker, Next.js, and Python ML as systemd services on a single Ubuntu VM with Nginx as reverse proxy. This minimizes operational cost and complexity for a solo developer while preserving a migration path to AWS managed services.

Use S3-compatible interfaces in application ports so local MinIO and production S3 remain interchangeable. Keep ClickHouse behind a repository/port boundary to preserve future migration options. Docker files remain in the repository as optional deployment assets.

## Consequences

Ubuntu VM + systemd reduces MVP operational cost and eliminates container orchestration complexity. The native development environment (ADR-009) removes Docker Desktop as a development dependency. AWS managed services remain available as a future evolution path when scale justifies the investment. Kubernetes is deferred rather than required for enterprise readiness. Cloudflare provides edge protection without coupling application code to a CDN.
