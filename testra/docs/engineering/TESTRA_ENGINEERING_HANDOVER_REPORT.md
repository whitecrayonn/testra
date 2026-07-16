# Testra Engineering Handover Report

## 1. Current Phase — Phase 2 (Test Management Core) + why

Phase 1 Identity & Tenancy is complete according to the latest engineering progress report. MFA, password reset, RBAC middleware, scoped API keys, and web authentication/onboarding are implemented and verified. Phase 2 is now the active roadmap phase for test cases, suites, folders, versioning, search, and audit trail.

## 2. Current Progress

Completed documentation work includes the API design and versioning policies, OpenAPI ownership guide, database documentation, ERD, sequence diagrams, system flow diagrams, module dependency model, developer onboarding, local development, deployment, troubleshooting, monitoring/logging, disaster recovery, security, release, and production readiness guides. ADR-002 through ADR-008 now record the finalized engineering decisions.

The latest implementation progress records Phase 1 completion: TOTP MFA, password reset, RBAC, scoped API keys, web authentication/onboarding, migrations, tests, and route wiring. Approved/current architectural references remain the modular-monolith and Clean Architecture governance, OpenAPI-first API, organization → workspace → project tenancy hierarchy, PostgreSQL/Redis/ClickHouse/S3-compatible storage responsibilities, and ADR-001 hybrid authentication.

Pending capabilities are Phase 2 test management, followed by execution/results, API testing/defects, analytics/launch, and post-MVP intelligence.

## 3. Current Repository Structure

The documentation set is organized under `docs/api`, `docs/architecture`, `docs/engineering`, `docs/deployment`, `docs/operations`, `docs/release`, and `docs/security`. The existing OpenAPI contract and ADR-001 were retained. New diagrams and guides document logical architecture without changing application or infrastructure structure.

## 4. Decisions Made

- Documentation source-of-truth boundaries are accepted in ADR-002.
- ADR-001 hybrid authentication is accepted and reconciled with ADR-007: self-hosted core auth, 15-minute JWTs, rotating refresh tokens, administrator/enterprise MFA, and scoped expiring API keys.
- ADR-003 selects the production deployment roadmap (amended by ADR-009): Ubuntu VM with systemd + Nginx for MVP; AWS ECS Fargate, RDS PostgreSQL, ElastiCache Redis, S3, ClickHouse Cloud, Cloudflare CDN/WAF, ACM TLS, and AWS Secrets Manager/KMS for Beta through Enterprise. Kubernetes is optional only after measured need.
- ADR-004 mandates defense-in-depth tenancy: middleware identity/scope resolution, request context propagation, service-layer authorization, PostgreSQL RLS, and tenant propagation through jobs, caches, exports, ClickHouse, and ML boundaries.
- ADR-005 fixes backup schedules, retention, RPO, and RTO: PostgreSQL PITR 35 days with 5-minute MVP RPO/4-hour MVP RTO; stronger Beta targets; bounded ClickHouse, audit, log, metric, trace, and object-storage retention.
- ADR-006 selects cursor pagination, 24-hour PostgreSQL idempotency records for side-effecting operations, `/api/v1` major versioning, and a minimum 12-month/two-major-release deprecation window.
- ADR-008 fixes MVP performance targets: p95 read 300 ms, p95 write 500 ms, 30-second synchronous timeout, 5-minute job timeout, 25 MiB request uploads, 500 concurrent users, 50 sustained requests/second, and 10,000 ClickHouse records/minute.
- ADR-009 replaces Docker Compose with a native development environment: local PostgreSQL, Redis, Mailpit, MinIO. Docker is optional. MVP deployment is Ubuntu VM + systemd + Nginx.

## 5. Technology Stack

Backend: Go modular monolith with Clean Architecture and chi routing.

Frontend: Next.js, React, TypeScript, and TailwindCSS as recorded in the repository baseline.

Database: PostgreSQL 16 for OLTP; ClickHouse 24 for OLAP/results; Redis 7 for ephemeral state and queues.

Cache: Redis.

Queue: Redis-backed Asynq direction recorded in engineering standards.

Infrastructure: Native development locally (no Docker required, see ADR-009); Ubuntu VM with systemd + Nginx for MVP; AWS ECS Fargate/RDS/ElastiCache/S3/ClickHouse Cloud/Cloudflare/ACM/Secrets Manager-KMS for Beta through Enterprise, with optional EKS only after measured need.

CI/CD: GitHub Actions baseline with staged promotion and immutable commit-SHA artifacts as documented guidance.

ML: Python FastAPI service with classical ML direction; no external LLM dependency.

Authentication: Self-hosted core authentication under ADR-001 and ADR-007 with 15-minute JWTs, rotating refresh tokens, enforced administrator/enterprise TOTP MFA, scoped expiring API keys, and deferred WorkOS SSO.

Storage: S3-compatible object storage for attachments, exports, and model artifacts.

Observability: OpenTelemetry, Prometheus, Grafana, Loki, and structured logging, with finalized retention of 30/90-day application logs, 15-month metrics, 14-day traces, and 2-year/7-year audit records.

Testing: Go unit/integration tests, web typechecking/tests, ML lint/tests, OpenAPI contract validation, and future E2E coverage.

## 6. Remaining Work

Priority 1: Implement Phase 2 test cases, suites, folders, versioning, full-text search, audit trail, OpenAPI extensions, migrations, and tests.

Priority 2: Implement and evidence the accepted ADR-003 through ADR-008 controls as Phase 2 and later modules are delivered.

Priority 3: Reconcile the ERD and API documentation with the completed Phase 1 implementation and migrations; this is implementation alignment, not an open architectural decision.

Priority 4: Implement later approved phases: execution/results; API testing/defects; dashboard, analytics, SDK, deployment, and MVP launch; then V2 intelligence.

## 7. Blockers

No architectural blockers remain. Production readiness remains conditional on implementing and evidencing the accepted controls, tests, backups, restore drills, and performance gates.

## 8. Risks

- Security risk remains while the newly completed Phase 1 controls undergo production hardening and tenant-isolation verification.
- Documentation drift if Phase 2 implementation does not reconcile diagrams, ERD, and contracts.
- Data-loss risk remains until the accepted backup policy is implemented and quarterly restore drills produce evidence.
- Privacy risk remains if implementation violates the accepted no-source/no-raw-payload logging and retention rules.
- Performance risk remains until ADR-008 targets are measured under representative load.

## 9. Files Modified

Documentation under `docs/` now includes the documentation index, API guides, architecture references and diagrams, onboarding/local/deployment guides, operations runbooks, security/release/readiness checklists, ADR-002 through ADR-009, and the engineering documentation report. The OpenAPI contract was extended for the implemented Phase 1 MFA, password-reset, and API-key routes. No application or infrastructure code was modified.

## 10. Important Notes

The OpenAPI contract covers auth, MFA, password reset, organizations, workspaces, projects, and API keys. Future endpoints must be added when approved and implemented. Planned product modules remain governed by `PHASES.md`, while architecture and engineering constraints are finalized in ADR-001 through ADR-009. The existing progress report remains append-only and authoritative for implementation status.

## Architectural Decisions Finalized

- **Production platform:** Ubuntu VM with systemd + Nginx for MVP (ADR-003 amended by ADR-009). AWS managed services with ECS Fargate for Beta through Enterprise: RDS PostgreSQL, ElastiCache Redis, S3, ClickHouse Cloud, Cloudflare CDN/WAF, ACM TLS, and AWS Secrets Manager/KMS. Local, MVP, Beta, and Enterprise stages are defined in ADR-003 and ADR-009. Local development is native (no Docker required).
- **Recovery and retention:** PostgreSQL daily snapshots/WAL PITR for 35 days, 5-minute MVP RPO, 4-hour MVP RTO, 13-month ClickHouse result retention, versioned S3, 2-year minimum audit retention for MVP/Beta, 7-year Enterprise audit retention, 30/90-day application log retention, 15-month metrics, and 14-day traces.
- **Tenant isolation:** PostgreSQL RLS is mandatory in staging/production; middleware resolves identity and candidate scope, request context propagates it, services authorize it, repositories set `app.tenant_id`, and all asynchronous/analytical paths carry tenant scope.
- **API standards:** Cursor pagination with 50 default/100 maximum, `Idempotency-Key` for side-effecting operations with 24-hour PostgreSQL records, `/api/v1` major versioning, and deprecation support for at least 12 months or two major release cycles.
- **Security:** 15-minute JWTs, rotating hashed refresh tokens with 30-day inactivity/90-day absolute expiry, revocation and reuse detection, 90-day default/365-day maximum API keys, 12-character passwords, administrator/enterprise MFA, defined Redis rate limits, 90-day secret rotation, and immutable security audit events.
- **MVP performance:** read p95 ≤ 300 ms, write p95 ≤ 500 ms, indexed query p95 ≤ 50 ms, 30-second maximum synchronous request, 5-minute job timeout, 25 MiB request upload limit, 500 concurrent authenticated users, 50 sustained requests/second, and 10,000 ClickHouse records/minute.
