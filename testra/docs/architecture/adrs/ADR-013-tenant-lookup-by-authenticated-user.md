# ADR-013: Tenant Lookup by Authenticated User for RLS

**Status:** Accepted  
**Date:** 2026-07-31  
**Deciders:** Principal Software Engineer, Platform Lead  

## Context

ADR-004 established PostgreSQL Row-Level Security (RLS) as the final defense for tenant isolation. The middleware sets `app.tenant_id` on a dedicated database connection and all subsequent repository calls flow through that connection.

The unresolved problem is the **chicken-and-egg** lookup for JWT-authenticated requests:

1. The HTTP request carries a JWT containing `user_id` but **not** an organization ID.
2. The route provides a workspace, project, test run, or defect ID.
3. To set `app.tenant_id` the middleware must first discover the resource's `organization_id`.
4. The existing RLS policies on `workspaces`, `projects`, `test_runs`, `test_run_items`, and `defects` use `app.tenant_id`, which is not yet known, so the lookup returns zero rows.

Only the API-key authentication path had a bypass policy (`api_keys_lookup_by_hash`, migration 000019) because the key hash uniquely identifies the tenant before the tenant is known. JWT authentication has no equivalent lookup path, making the majority of the web UI routes fail under RLS.

## Decision

Add a second, tightly-scoped RLS lookup mode called `app.lookup_user_id`. The middleware sets this session variable to the authenticated user's UUID **before** resolving the tenant, and resets it immediately after.

New lookup policies will allow a connection to read tenant-scoped rows **only when the authenticated principal is a member of the owning organization**. Once the organization is known, the middleware sets `app.tenant_id` and resets `app.lookup_user_id`, restoring normal tenant-isolated access.

Key rules:

1. `app.lookup_user_id` is set **only** inside the dedicated request connection.
2. It is set **before** any tenant-resolution query.
3. It is **reset to an empty string** as soon as `app.tenant_id` is set.
4. Lookup policies are permissive OR policies that coexist with the existing tenant policies.
5. Lookup policies never grant broader access than the user's existing `organization_members` rows.

Lookup policies are:

- `organization_members_lookup_user` — `user_id = current_setting('app.lookup_user_id', true)::uuid`
- `workspaces_lookup_user` — `organization_id IN (SELECT organization_id FROM organization_members WHERE user_id = current_setting('app.lookup_user_id', true)::uuid)`
- `projects_lookup_user` — `workspace_id IN (SELECT id FROM workspaces WHERE organization_id IN (SELECT organization_id FROM organization_members WHERE user_id = current_setting('app.lookup_user_id', true)::uuid))`
- `test_runs_lookup_user` — `workspace_id IN (SELECT id FROM workspaces WHERE organization_id IN (SELECT organization_id FROM organization_members WHERE user_id = current_setting('app.lookup_user_id', true)::uuid))`
- `test_run_items_lookup_user` — `run_id IN (SELECT id FROM test_runs WHERE workspace_id IN (SELECT id FROM workspaces WHERE organization_id IN (SELECT organization_id FROM organization_members WHERE user_id = current_setting('app.lookup_user_id', true)::uuid)))`
- `defects_lookup_user` — `workspace_id IN (SELECT id FROM workspaces WHERE organization_id IN (SELECT organization_id FROM organization_members WHERE user_id = current_setting('app.lookup_user_id', true)::uuid))`
- `api_keys_lookup_user` — `workspace_id IN (SELECT id FROM workspaces WHERE organization_id IN (SELECT organization_id FROM organization_members WHERE user_id = current_setting('app.lookup_user_id', true)::uuid))`

## Consequences

- JWT tenant resolution now works under RLS without requiring a `SUPERUSER` or `BYPASSRLS` role.
- The blast radius of `app.lookup_user_id` is limited to resources the user is already a member of.
- Resetting `app.lookup_user_id` after tenant resolution prevents accidental cross-tenant reads for the rest of the request.
- All future tables that participate in JWT route resolution must add a `lookup_user` policy.
- Integration tests must verify that a user can resolve their own resources and cannot resolve resources belonging to other tenants.

## Migration

Migration `000027_add_user_lookup_rls_policies` creates the policies and indexes. Down migration removes them.

## References

- ADR-004 Tenant Isolation Strategy
- Migration `000009_add_rls_policies`
- Migration `000019_add_api_key_org_and_lookup_policy`
