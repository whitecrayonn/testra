# ADR-004: Tenant Isolation Strategy

**Status:** Accepted
**Date:** July 2026

## Context

Testra is multi-tenant and privacy-first. Application authorization alone is insufficient protection against query or module defects, while relying only on database policies makes local reasoning and domain authorization incomplete.

## Decision

Use defense-in-depth tenant isolation:

1. Every tenant-scoped PostgreSQL row carries `organization_id` directly or through an enforced ownership relationship.
2. PostgreSQL Row Level Security is mandatory for tenant-scoped tables in staging and production.
3. Each request establishes an authenticated principal and an active organization/workspace/project scope in a request context after membership resolution.
4. The database transaction sets a non-bypassable, transaction-local tenant setting such as `app.tenant_id`; RLS policies read that setting. Application connection roles must not bypass RLS.
5. HTTP middleware authenticates and resolves candidate scope; it does not grant business permissions.
6. Service/use-case layers enforce resource relationships and RBAC permissions before repository calls.
7. Repositories accept explicit scope/principal data and use parameterized queries; they never trust arbitrary client IDs.
8. Queue jobs, cache keys, exports, object paths, ClickHouse rows, and ML requests carry tenant scope and repeat authorization at job boundaries.
9. Cross-tenant administrative operations require a separately audited system role and explicit service port; ordinary tenant paths cannot use it.

## Consequences

Every future module must implement tenant context propagation and isolation tests. RLS reduces blast radius from application defects, while service authorization preserves domain correctness. Local development uses the same policy shape where practical; tests must include cross-tenant denial cases.
