# Database Engineering Review v1 — Testra Platform

**Date:** 2026-07-19  
**Scope:** PostgreSQL schema, migrations (`apps/api/migrations/000001_*.sql` through `000028_*.sql`), RLS policies, indexes, and data-integrity concerns.  
**Goal:** Verify that the schema is correct, performant, and properly isolates tenants.

## Executive Summary

The PostgreSQL schema is well-normalized, uses UUID primary keys, TIMESTAMPTZ for temporal data, and JSONB for flexible metadata. RLS is enabled on almost all tenant-scoped tables, and the migration sequence is ordered correctly. The review found one incorrect RLS policy (`role_assignments_tenant`) and several missing/recommended indexes and documentation updates; these are addressed or documented below.

## 1. Migration Quality

| Aspect | Status | Notes |
|--------|--------|-------|
| Sequential ordering | ✅ Good | `000001` through `000028` are numbered and have paired `.down.sql` files |
| Down migrations | ✅ Good | Every `.up.sql` has a `.down.sql` |
| `IF NOT EXISTS` / `DROP IF EXISTS` | ✅ Good | Migrations are idempotent where appropriate |
| Schema documentation | ⚠️ Fair | `DATABASE_GUIDE.md` covered only migrations `000001`–`000018`; updated in this pass to `000028` |
| Mixed concerns in migrations | ⚠️ Fair | Several migrations seed permissions/roles; acceptable for an MVP but consider separating seed data |

## 2. Schema Design

### 2.1 Core Entities

- `users` — stores credentials, MFA state, and name. `email` is `UNIQUE` and indexed.
- `organizations` / `organization_members` / `workspaces` / `workspace_members` — standard multi-tenant hierarchy.
- `projects` — child of workspace, `UNIQUE(workspace_id, key)`.
- `test_folders` / `test_suites` / `test_cases` / `test_case_versions` — test management tree, with `parent_id` and `folder_id` self-references.
- `test_runs` / `test_run_items` — execution results, with aggregated totals on `test_runs`.
- `integrations` / `integration_events`, `notifications` / `notification_channels` / `notification_preferences`, `subscriptions` / `invoices`, `analytics` dashboards, `defects`, `intelligence` predictions — domain modules following the same pattern.
- `queue_jobs` — PostgreSQL-backed job queue.

### 2.2 Data Types

- `UUID` primary keys with `gen_random_uuid()` defaults.
- `TIMESTAMPTZ` for all timestamps.
- `JSONB` for `metadata`, `payload`, `config`, `artifacts`, `steps`, `features`, and dashboard `config`.
- `TSVECTOR` on `test_cases.search_tsv` for full-text search, maintained by triggers.
- `TEXT[]` for tags and API-key scopes.

### 2.3 Referential Integrity

- Foreign keys use `ON DELETE CASCADE` or `SET NULL` consistently.
- `NOT NULL` constraints are applied to required columns.
- `UNIQUE` constraints exist on natural keys (`email`, `organization_id+slug`, `workspace_id+key`, `role_id+user_id+scope_type+scope_id`, `key_hash`).

## 3. Tenant Isolation

### 3.1 RLS Strategy

The design uses `current_setting('app.tenant_id', true)::uuid` per transaction, set by the application layer. Every tenant-scoped table has a `USING` policy that restricts rows to the current tenant.

**Good:**
- RLS is enabled on `organizations`, `organization_members`, `workspaces`, `workspace_members`, `projects`, `api_keys`, `role_assignments`, test management tables, `test_runs`, `test_run_items`, `idempotency_records`, `notifications`, `notification_preferences`, `notification_channels`, `integrations`, `integration_events`, `subscriptions`, `invoices`, `queue_jobs`, and others.

### 3.2 Pre-authentication Lookup Policies

Migrations `000019` (API-key lookup) and `000027` (user-lookup policies) add permissive `USING` policies that allow a connection to resolve the tenant before `app.tenant_id` is known. The application is responsible for resetting `app.lookup_user_id` and `app.lookup_key_hash` after use.

**Risk:**
- If the application forgets to reset `app.lookup_user_id`, subsequent queries in the same connection may leak rows to the lookup user. This is mitigated by middleware but should be covered by integration tests.
- `app.current_tenant()` in `000026` swallows all exceptions and returns `NULL`. A malformed `app.tenant_id` becomes invisible rather than failing closed. Consider changing the exception block to raise an error after logging.

### 3.3 `role_assignments` RLS Policy Was Incorrect

**File:** `apps/api/migrations/000009_add_rls_policies.up.sql`

The original policy was:
```sql
CREATE POLICY role_assignments_tenant ON role_assignments
    USING (scope_id = current_setting('app.tenant_id', true)::uuid);
```

`role_assignments` stores `scope_type` (`organization`, `workspace`, `project`) and `scope_id`. The policy above only works for `scope_type = 'organization'`; for workspace or project scope IDs it compares the wrong value to the tenant organization ID.

**Fix applied in this review:**
- New migration `apps/api/migrations/000028_fix_role_assignments_rls_policy.up.sql` replaces the policy with a version that checks `scope_type` and resolves workspace/project IDs through the tenant hierarchy.
- A paired `.down.sql` reverts to the original simple policy.

## 4. Indexing

### 4.1 Existing Indexes

The migrations generally include sensible indexes on foreign keys, search vectors, and lookup columns. Examples:
- `idx_users_email`
- `idx_api_keys_hash`, `idx_api_keys_workspace`, `idx_api_keys_organization`
- `idx_test_cases_workspace`, `idx_test_cases_project`, `idx_test_cases_suite`, `idx_test_cases_search` (GIN on `search_tsv`)
- `idx_test_runs_*` family of indexes
- `idx_idempotency_records_lookup` and `idx_idempotency_records_expires`
- `idx_queue_jobs_status_scheduled` and `idx_queue_jobs_queue_status`

### 4.2 Missing / Recommended Indexes

| Table / Column | Recommendation | Rationale |
|----------------|----------------|-----------|
| `role_assignments` | `(user_id, scope_type, scope_id)` covering for permission loader | The loader filters on `user_id`, `scope_type`, `scope_id`; the existing `(scope_type, scope_id)` index is not ordered for `user_id` first. |
| `audit_events` | `(created_at DESC)` with `action` | Audit queries are likely filtered by time range and action. |
| `refresh_tokens` | `(family_id, revoked_at)` | Family revocation queries currently scan the `(family_id)` index. |
| `notification_channels` | `(organization_id, type)` | Likely filtered by org and channel type. |
| `test_cases` | `(workspace_id, status, priority)` | List/filter queries on status/priority within a workspace. |
| `queue_jobs` | `(tenant_id, status, scheduled_at)` | The worker dequeues globally by status and scheduled_at, but tenant-scoped monitoring would benefit from `tenant_id`. |

**Note:** The RLS `USING` clause on `queue_jobs` does not currently have a direct index on `tenant_id`; the worker uses `FOR UPDATE SKIP LOCKED` with `(queue_name, status, scheduled_at)`.

## 5. Security and Consistency Findings

### 5.1 `audit_events` Has No Tenant Column

`audit_events` is not scoped by `organization_id` or `workspace_id`. This is acceptable if audit is a global, append-only log, but tenant administrators cannot query their own audit stream without a join. Consider adding `organization_id` for tenant-specific audit views.

### 5.2 `password_reset_tokens` and `refresh_tokens` Are Not Tenant-Scoped

These are not tenant-scoped by design; they reference `users` and are used before the tenant is resolved. This is correct.

### 5.3 `api_keys.organization_id` Backfill

Migration `000019` adds `organization_id` to `api_keys`, backfills it from `workspaces`, and creates a trigger for future inserts. This is a clean denormalization that solves the chicken-and-egg API-key authentication problem.

### 5.4 JSONB Payloads Without Validation

`test_cases.steps`, `test_run_items.artifacts`, `integration_events.payload`, and `notification_channels.config` are `JSONB` with no database-level validation. The application validates these, but malformed data can still be inserted through direct DB access. This is acceptable for an MVP but should be revisited if direct DB write paths are opened.

## 6. Migration Hygiene

- `000006_add_rbac.up.sql` seeds roles, permissions, and role-permission mappings. This is fine, but consider moving seed data to a separate `seeds/` directory for clarity.
- `000008`, `000013`, `000016`, `000018`, `000025` all insert permissions/role-permissions. The growing permission catalog is becoming harder to maintain in separate migration files. A single permissions registry file loaded by a seed task may be easier to manage.

## 7. Recommended Next Steps

1. Apply migration `000028` to fix the `role_assignments_tenant` RLS policy.
2. Add the missing indexes listed in §4.2 through new migrations.
3. Add a `NOT NULL` or `CHECK` constraint on `role_assignments.scope_type` for the valid set of scope types (`organization`, `workspace`, `project`, `api_key` if used).
4. Consider adding `organization_id` to `audit_events` for tenant-scoped audit queries.
5. Create a `DATABASE_PERFORMANCE.md` doc that tracks the most-executed queries and their plans once the application is under load.
6. Continue updating `DATABASE_GUIDE.md` as new migrations are added.

## 8. Verification

```powershell
go build ./...
go test -count=1 ./...
```

Both pass. The new `000028` migration is SQL-only and was syntax-checked by reading and comparing against PostgreSQL policy syntax.

## Conclusion

The database design is solid for a multi-tenant SaaS MVP. The most important issue found was the `role_assignments_tenant` RLS policy, which has been corrected with migration `000028`. Indexing and seed-data organization are the next areas to improve.
