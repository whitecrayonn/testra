# Testra Database Documentation

## 1. Storage Responsibilities

| Store | Responsibility | Non-responsibilities |
|---|---|---|
| PostgreSQL 16 | transactional identity, tenancy, configuration, relational business data, audit records | high-volume analytical event querying |
| ClickHouse 24 | test results, events, and time-series analytical data | transactional authority, user/session state |
| Redis 7 | sessions, rate limiting, and Asynq job queues | durable business records |
| S3-compatible storage | attachments, exports, and model artifacts | relational metadata authority |

These responsibilities follow the engineering standards. Actual implementation status is tracked by phase documentation.

## 2. Relational Invariants

- UUID primary keys are used for entities.
- Tenant-scoped rows carry `organization_id` or an equivalent tenant identifier.
- Tables have UTC `created_at` and `updated_at` timestamps where mutable lifecycle data exists.
- Foreign keys and indexes are required for relationships and common filters.
- Parent-owned children use cascading deletion only where product retention and audit requirements permit it.
- Migrations are sequential `golang-migrate` files with both up and down directions.
- Merged migrations are immutable; corrections use a new migration.

## 3. Tenancy and Authorization

The organizational hierarchy is organization → workspace → project. A request must establish authenticated identity, membership, and resource relationship before reading or mutating tenant data. Client-supplied IDs are selectors, not proof of access.

PostgreSQL Row Level Security is mandatory for tenant-scoped tables in staging and production. Each transaction sets a transaction-local `app.tenant_id` after middleware and service authorization resolve the active scope. Application roles used by the API do not bypass RLS. Full responsibilities are defined in ADR-004.

## 4. Current and Planned Core Entities

As of Phase 3, the following are migrated and implemented: users, organizations, workspaces, projects, roles, permissions, role assignments, API keys, refresh tokens, password reset tokens, audit events, test folders, test suites, test cases, test case versions, test runs, test run items, idempotency records, notifications, notification preferences, and notification channels.

Later phases add API testing definitions, defects, analytics, integrations, and marketplace extensions.

The authoritative schema is the migrations directory (`apps/api/migrations/`). `docs/architecture/ERD.md` provides a logical relationship view but may lag the migrations.

## 5. ClickHouse Rules

ClickHouse is append-oriented analytical storage using the MergeTree family. Every event/result record includes tenant identity and event time. Partition by event month and order by tenant, project, and event time as the default analytical layout. Deduplicate ingestion using stable domain event/result identifiers. Retain analytical results for 13 months by default, with daily backups retained 14 days and weekly backups 8 weeks. Transactional status and authorization metadata remain authoritative in PostgreSQL. Full recovery rules are defined in ADR-005.

## 6. Redis Rules

Use the `testra:` namespace. Keys must encode purpose and scope, and have explicit TTLs where data is ephemeral. Redis loss must not destroy the only copy of business data. Queue jobs require retry policy, dead-letter handling, and idempotent consumers.

## 7. Migration Operations

1. Review migration SQL and its down migration.
2. Confirm lock, duration, index, and backfill impact.
3. Apply in staging against a representative backup.
4. Verify schema and application compatibility.
5. Promote through the deployment pipeline; migrations are not manually applied in production.
6. Record completion and rollback/forward-fix decision.

Destructive or long-running changes require an expand/contract plan and explicit production-readiness approval.
