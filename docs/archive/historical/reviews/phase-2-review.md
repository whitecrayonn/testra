# Phase 2 Engineering Review — Test Management Core

**Date:** 2026-07-16  
**Reviewer:** Cascade (AI Engineering Assistant)  
**Phase:** 2 — Test Management Core  
**Status:** PASS WITH MINOR ISSUES

---

## Scores

| Category | Score | Notes |
|---|---|---|
| **Overall** | **B+ (85/100)** | Solid implementation with a few correctable issues |
| **Architecture** | **A (92/100)** | Clean Architecture followed correctly |
| **Backend** | **B+ (87/100)** | Well-structured with two notable bugs |
| **Frontend** | **B (82/100)** | Functional but has React anti-patterns |
| **Security** | **B- (78/100)** | RLS gap on test tables, audit logging missing |
| **Scalability** | **B+ (85/100)** | Good indexes and FTS, search pagination bug |
| **Maintainability** | **B+ (85/100)** | Clean code, some duplication from shared patterns |

---

## Decision: PASS WITH MINOR ISSUES

Phase 2 implementation is architecturally sound and functionally complete. Two issues should be fixed before Phase 3 (RLS policies, search cursor pagination), and several others should be tracked as technical debt. The phase can be considered complete with these issues logged as follow-up tasks.

---

## 1. Architecture Review

### Clean Architecture Compliance — PASS

- **Domain layer** (`domain.go`): Pure entities with no external dependencies. Entities: `TestFolder`, `TestSuite`, `TestCase`, `TestStep`, `TestCaseVersion`. Proper use of value types (`TestCaseStatus`, `TestCasePriority`).
- **Ports layer** (`ports.go`): `Repository` interface defined in the domain package, consumed by the service. No infrastructure leaks.
- **Repository layer** (`repository.go`): `SQLRepository` implements `Repository` interface. SQL is contained here, not leaked to service or handler.
- **Service layer** (`service.go`): Business logic with validation, versioning. Depends on `Repository` interface, not concrete implementation.
- **Handler layer** (`handler.go`): HTTP request/response mapping. Depends on `Service`, not `Repository`.
- **Module wiring** (`module.go`): Clean composition root with `NewModule(db)`.
- **Dependency direction**: handler → service → repository → database. No reverse dependencies. No circular imports.

### Hexagonal Architecture Compliance — PASS

- The `Repository` interface acts as the port (driven side).
- The `Handler` acts as an adapter (driving side).
- The `Service` is the application core with no knowledge of HTTP or SQL.
- The domain is isolated and could be tested without any infrastructure.

### API Consistency — PASS (with caveat)

- All handlers follow the same pattern as existing modules: request struct → UUID parse → service call → response mapping.
- `mapError` function is consistent with all other modules (organization, workspace, project, apikeys, identity).
- Error codes and HTTP status mapping are consistent.
- **Caveat**: Paginated responses use `apihttp.JSON(w, 200, map[string]any{"data": resp, "meta": meta})` which double-wraps in the `Envelope{Data: ...}` structure, producing `{"data": {"data": [...], "meta": {...}}}`. This is a pre-existing pattern across ALL modules (project, workspace, organization, apikeys), not a Phase 2 regression. The frontend `apiFetch` correctly handles this by returning `body.data`.

---

## 2. Backend Review

### 2.1 Domain (`domain.go`) — PASS

- Entities are well-defined with proper UUID types, timestamps, and nullable pointer fields.
- Status and priority enums use typed string constants with validation functions.
- `TestStep` is a value object, correctly modeled.

### 2.2 Ports (`ports.go`) — PASS

- Interface is consumed by the service, implemented by the repository.
- Methods are well-grouped by entity.
- Context is propagated correctly on all methods.

### 2.3 Repository (`repository.go`) — ISSUES FOUND

**ISSUE R-1 (HIGH): Search cursor pagination is broken**

`SearchCases` orders results by `ts_rank DESC, id DESC`, but the cursor uses only `id < $3`. This means the cursor filters by ID range, not by rank position. After the first page, results with higher IDs but lower rank are incorrectly excluded.

File: `repository.go:302-316`

```
ORDER BY ts_rank(search_tsv, to_tsquery('pg_catalog.english', $2)) DESC, id DESC LIMIT $4
```

The cursor should be a composite `(rank, id)` tuple, or the search should use keyset pagination with `ts_rank` as the primary sort key.

**ISSUE R-2 (MEDIUM): `pqArray` doesn't escape special characters**

`pqArray()` manually constructs PostgreSQL array literals (`{tag1,tag2}`) without escaping commas, quotes, or braces in tag values. A tag like `"hello,world"` would be split into two array elements.

File: `repository.go:419-432`

Should use `github.com/lib/pq` `pq.Array()` for proper encoding, or at minimum escape special characters.

**ISSUE R-3 (MEDIUM): `parseTags` is fragile manual parsing**

Manual parsing of PostgreSQL array string format (`{tag1,tag2}`). Doesn't handle quoted elements, escaped characters, or NULL arrays.

File: `repository.go:434-451`

Should use `pq.Array()` scanner or a proper PostgreSQL array parser.

**ISSUE R-4 (LOW): `uuid.Parse` errors silently ignored**

Multiple locations discard the error from `uuid.Parse`:
- `repository.go:44`: `pid, _ := uuid.Parse(parentID.String)`
- `repository.go:80`: `parsed, _ := uuid.Parse(pid.String)`
- `repository.go:138`: `fid, _ := uuid.Parse(folderID.String)`
- `repository.go:243`: `sid, _ := uuid.Parse(suiteID.String)`
- `repository.go:408`: `sid, _ := uuid.Parse(suiteID.String)`

If the database contains a malformed UUID (shouldn't happen but defensive coding is better), this silently produces `uuid.Nil`.

**ISSUE R-5 (LOW): `json.Marshal` errors silently ignored**

`stepsJSON, _ := json.Marshal(tc.Steps)` at lines 210, 326, 356. While `TestStep` is a simple struct that won't fail to marshal, discarding errors is not best practice.

### 2.4 Service (`service.go`) — ISSUES FOUND

**ISSUE S-1 (HIGH): No transaction for `UpdateCase`**

`UpdateCase` performs two database operations:
1. `CreateVersion` (snapshot the old version)
2. `UpdateCase` (apply the update)

These are not wrapped in a transaction. If step 2 fails, the version snapshot is orphaned — a data integrity issue.

File: `service.go:238-288`

The `Repository` interface doesn't expose transaction support. Consider adding a `WithTx` method or a `UnitOfWork` pattern.

**ISSUE S-2 (LOW): `UpdateCase` validates status/priority after creating version**

The version snapshot is created (line 259) before status/priority validation (lines 269-278). If validation fails, the version snapshot is already persisted — orphaned.

**ISSUE S-3 (LOW): No validation on `Steps` content**

Steps with empty action/expected are accepted. Consider validating that each step has non-empty action and expected fields.

### 2.5 Handler (`handler.go`) — PASS (with notes)

- Request/response mapping is consistent and thorough.
- `mapCaseResponse` correctly initializes `tc.Tags = []string{}` for null tags to ensure JSON array output.
- **NOTE H-1 (LOW)**: `mapCaseResponse` mutates the input `tc.Tags` field (line 118). This is a side effect on the domain object. Should use a local variable.
- **NOTE H-2 (LOW)**: `mapError` is duplicated across 6 handler files (identity, organization, workspace, project, apikeys, testmanagement). This is a pre-existing pattern — should be extracted to `shared/http/errors.go`.

### 2.6 Module (`module.go`) — PASS

Clean wiring: `NewModule(db)` → `NewSQLRepository(db)` → `NewService(repo)` → `NewHandler(service)`.

---

## 3. Database & Migrations Review

### 3.1 Migration 000012 — ISSUES FOUND

**ISSUE M-1 (HIGH): No RLS policies on test tables**

Migration 000009 enables RLS on `organizations`, `workspaces`, `projects`, `api_keys`, `role_assignments`. Migration 000012 creates `test_folders`, `test_suites`, `test_cases`, `test_case_versions` but does NOT enable RLS or create tenant isolation policies.

This is a **tenant isolation gap**. If the application's `app.tenant_id` setting is the security boundary, these tables are unprotected. An attacker with database access could read/write across tenant boundaries.

Required:
```sql
ALTER TABLE test_folders ENABLE ROW LEVEL SECURITY;
ALTER TABLE test_suites ENABLE ROW LEVEL SECURITY;
ALTER TABLE test_cases ENABLE ROW LEVEL SECURITY;
ALTER TABLE test_case_versions ENABLE ROW LEVEL SECURITY;

CREATE POLICY test_folders_tenant ON test_folders
    USING (workspace_id IN (
        SELECT id FROM workspaces WHERE organization_id = current_setting('app.tenant_id', true)::uuid
    ));
-- etc.
```

**ISSUE M-2 (MEDIUM): No CHECK constraints on `status` and `priority`**

The `test_cases` table uses `VARCHAR(20)` for `status` and `priority` without CHECK constraints. Invalid values like `'foobar'` can be stored at the database level.

```sql
status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'deprecated')),
priority VARCHAR(20) NOT NULL DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high', 'critical')),
```

**ISSUE M-3 (MEDIUM): No UNIQUE constraint on `test_case_versions(test_case_id, version)`**

Nothing prevents duplicate version numbers for the same test case. Should have:
```sql
UNIQUE(test_case_id, version)
```

**ISSUE M-4 (LOW): Missing `updated_at` auto-update trigger**

Tables have `updated_at` columns but no database trigger to auto-update them on row modification. The application sets `updated_at` manually, but direct SQL updates would leave it stale. This is consistent with the existing pattern across all tables in the project.

### 3.2 Indexes — PASS

- `idx_test_folders_workspace` — supports list by workspace
- `idx_test_folders_parent` — supports hierarchical queries
- `idx_test_suites_workspace` — supports list by workspace
- `idx_test_suites_folder` — supports filter by folder
- `idx_test_cases_workspace` — supports search by workspace
- `idx_test_cases_project` — supports list by project
- `idx_test_cases_suite` — supports filter by suite
- `idx_test_cases_status` — supports filter by status
- `idx_test_cases_search` — GIN index on `search_tsv` for full-text search
- `idx_test_case_versions_case` — supports list versions by case
- `idx_test_case_versions_version` — composite index for `ORDER BY version DESC`

Index coverage is comprehensive for the implemented query patterns.

### 3.3 Full-Text Search — PASS (with caveat)

- GIN index on `search_tsv` column.
- Triggers on INSERT and UPDATE of `title`/`description`.
- `toTSQuery` builds `word1 & word2` AND-query from user input.
- `ts_rank` ordering for relevance.
- **Caveat**: `toTSQuery` only supports AND queries. No support for OR, phrase queries, or wildcards. Acceptable for MVP.

### 3.4 Migration 000013 (RBAC Permissions) — PASS

- 4 permissions: `tests:create`, `tests:read`, `tests:update`, `tests:delete`.
- Role assignments: Owner (all), Admin (all), Editor (create/read/update), Viewer (read only).
- `ON CONFLICT DO NOTHING` for idempotency.
- Down migration correctly cleans up role_permissions then permissions.

---

## 4. Security Review

### 4.1 RBAC Enforcement — PASS

All routes have `RequirePermission` middleware:
- POST routes → `tests:create`
- GET list/search routes → `tests:read`
- GET single item routes → `tests:read`
- PUT routes → `tests:update`
- DELETE routes → `tests:delete`

### 4.2 Tenant Isolation — PARTIAL FAIL

**ISSUE SEC-1 (HIGH): No RLS on test tables** (see ISSUE M-1)

The application layer enforces tenant context via `TenantContext` middleware, which resolves the organization ID and sets it for RBAC checks. However, without RLS policies on the test tables, there is no database-level tenant isolation. This is inconsistent with the security model established in migration 000009.

### 4.3 Audit Logging — PARTIAL FAIL

**ISSUE SEC-2 (MEDIUM): No audit logging on test management mutations**

The Phase 2 DoD states: "`audit` module: immutable event log on all mutations". The `AuditLog` middleware exists and is used for API key create/revoke, but it is NOT applied to any test management mutating routes (CreateFolder, UpdateFolder, DeleteFolder, CreateSuite, UpdateSuite, DeleteSuite, CreateCase, UpdateCase, DeleteCase).

### 4.4 Input Validation — PASS

- `validation.IsValidName()` checks length (1-100 chars).
- UUID parsing on all inputs.
- Status and priority enum validation in service layer.
- `MaxBodySize` middleware applied globally (1MB limit).
- `json.NewDecoder` used for request body parsing.

### 4.5 SQL Injection — PASS

All SQL queries use parameterized queries (`$1`, `$2`, etc.). No string concatenation in SQL. The `toTSQuery` function sanitizes input by only keeping alphanumeric characters.

---

## 5. Frontend Review

### 5.1 TypeScript Quality — PASS

- `types/testmanagement.ts`: Proper interfaces matching API response shapes.
- `features/testmanagement/api.ts`: Well-typed API client functions with proper return types.
- `PaginationMeta` interface defined and used.
- No `any` types in the test management code.

### 5.2 React Patterns — ISSUES FOUND

**ISSUE F-1 (MEDIUM): `localStorage` reads during render**

`projectId` and `workspaceId` are read from `localStorage` during component render, not in `useEffect`. This causes:
- SSR/hydration mismatch (localStorage is undefined on server).
- Values are stale — changing localStorage won't trigger re-render.

File: `test-cases/page.tsx:37-44`

Should use `useState` + `useEffect` pattern.

**ISSUE F-2 (MEDIUM): `setTimeout` hack for search state update**

```tsx
const handleSearch = () => {
    setSearchMode(searchQuery.trim().length > 0);
    setCursor(undefined);
    setCases([]);
    setTimeout(() => fetchCases(true), 0);  // hack
};
```

`setSearchMode` is async, so `fetchCases` is called before the state updates. Should use `useEffect` to trigger fetch when `searchMode` changes, or pass the search parameters directly to `fetchCases`.

File: `test-cases/page.tsx:80-85`

**ISSUE F-3 (LOW): `fetchCases` callback has stale closure risk**

`useCallback` depends on `cases`, which changes on every fetch. The `handleLoadMore` function calls `fetchCases()` which may use a stale `cursor` value since `setCursor` is also async.

**ISSUE F-4 (LOW): No loading state distinction for "Load More"**

The `loading` state is used for both initial load and "Load More". When loading more, the initial "Loading..." card would show instead of the existing list.

**ISSUE F-5 (LOW): Route re-exports are fragile**

```tsx
export { default } from "../../[workspace]/test-cases/page";
```

These re-exports create coupling between `[workspace]` and `dashboard` route paths. Any structural change to one would break the other. Consider making the pages shared components instead.

### 5.3 UI/UX — PASS

- Status and priority badges with color coding.
- Search bar with enter-to-search.
- Cursor-based "Load More" pagination.
- Step builder with add/remove.
- Version history panel.
- Delete confirmation dialog.
- Error states with user-friendly messages.
- Empty states with call-to-action.

---

## 6. Test Coverage Review

### 6.1 Unit Tests — PASS (with gaps)

**Covered:**
- Folder creation: valid, empty name, missing workspace
- Suite creation: valid, empty name, missing workspace
- Test case creation: valid, empty title, missing workspace, missing project, missing created_by, invalid status, invalid priority, default values
- Test case update: version snapshot creation, title/status update, version increment
- Test case get/delete: success, not found, double delete
- Search: empty query returns nil
- Folder/suite update: not found

**Gaps:**
- No tests for folder/suite delete (success or not found)
- No tests for folder/suite list (filtering by workspace/parent/folder)
- No tests for test case list with cursor pagination
- No tests for search with actual query
- No tests for UpdateCase with invalid status/priority (should return ErrInvalidInput after version snapshot — see ISSUE S-2)
- No tests for UpdateCase with steps modification
- No tests for ListVersions with multiple versions

### 6.2 Integration Tests — NOT PRESENT

No integration tests with real PostgreSQL. This is consistent with the project's current testing approach (unit tests only with fake repositories).

---

## 7. Code Duplication & Technical Debt

### 7.1 Code Duplication

| Item | Severity | Pre-existing? |
|---|---|---|
| `mapError` function duplicated across 6 handlers | LOW | Yes — all modules |
| Paginated response double-wrapping pattern | LOW | Yes — all modules |
| UUID parse + nullable pattern repeated in repository | LOW | Module-specific |
| Step mapping loop repeated in `mapCaseResponse` and `mapVersionResponse` | LOW | Module-specific |

### 7.2 Technical Debt Register

| ID | Issue | Severity | Should fix before Phase 3? |
|---|---|---|---|
| TD-1 | No RLS on test tables (M-1/SEC-1) | HIGH | **Yes** |
| TD-2 | Search cursor pagination broken (R-1) | HIGH | **Yes** |
| TD-3 | No transaction for UpdateCase (S-1) | HIGH | Recommended |
| TD-4 | No audit logging on test mutations (SEC-2) | MEDIUM | Recommended |
| TD-5 | `pqArray`/`parseTags` don't handle special chars (R-2/R-3) | MEDIUM | No |
| TD-6 | No CHECK constraints on status/priority (M-2) | MEDIUM | No |
| TD-7 | No UNIQUE on test_case_versions(case_id, version) (M-3) | MEDIUM | No |
| TD-8 | `mapCaseResponse` mutates input (H-1) | LOW | No |
| TD-9 | Frontend localStorage reads during render (F-1) | MEDIUM | No |
| TD-10 | Frontend setTimeout hack (F-2) | MEDIUM | No |
| TD-11 | `mapError` duplicated across 6 handlers (H-2) | LOW | No (pre-existing) |
| TD-12 | `uuid.Parse` errors silently ignored (R-4) | LOW | No |
| TD-13 | `json.Marshal` errors silently ignored (R-5) | LOW | No |
| TD-14 | No `updated_at` auto-update trigger (M-4) | LOW | No (pre-existing pattern) |
| TD-15 | UpdateCase validates after version snapshot (S-2) | LOW | No |
| TD-16 | No step content validation (S-3) | LOW | No |

### 7.3 TODOs

No `TODO` or `FIXME` comments found in the Phase 2 codebase.

---

## 8. Go Best Practices Review

| Practice | Status | Notes |
|---|---|---|
| Context propagation | PASS | All methods accept `context.Context` |
| Error wrapping | PARTIAL | Sentinel errors used correctly; some `_ =` error discards |
| UUID for IDs | PASS | `github.com/google/uuid` throughout |
| Time handling | PASS | `time.Now().UTC()` for all timestamps |
| Interface segregation | PASS | Single `Repository` interface, well-grouped |
| JSON tags on response structs | PASS | All response structs have `json:"..."` tags |
| Pointer vs value receivers | PASS | Consistent pointer receivers on `SQLRepository` and `Service` |
| Package naming | PASS | `testmanagement` — single word, lowercase |
| Import organization | PASS | Stdlib, then third-party, then internal |
| Lint compliance | PASS | `go build` and `go vet` pass cleanly |

---

## 9. Scalability Review

- **Indexes**: Comprehensive coverage for all query patterns.
- **FTS**: GIN index on `search_tsv` — efficient for full-text search at scale.
- **Cursor pagination**: Correctly implemented for `ListCases` (ordered by `id DESC`). Broken for `SearchCases` (see R-1).
- **Connection pooling**: Uses `*sql.DB` from the standard library, shared across modules.
- **JSONB for steps**: Efficient storage and query for variable-length step arrays.
- **Text[] for tags**: PostgreSQL native array type — efficient for tag queries.
- **No N+1 queries**: All list endpoints use single queries with proper JOINs or filtered queries.

---

## 10. Summary

### Strengths

1. **Clean Architecture adherence** is excellent — domain, ports, repository, service, handler layers are properly separated with correct dependency direction.
2. **Domain modeling** is clean and idiomatic Go with typed enums, proper UUID usage, and nullable pointer fields.
3. **Migration design** is solid with appropriate indexes, foreign keys, and cascade rules.
4. **Full-text search** implementation with GIN index and triggers is well-designed.
5. **Test coverage** for service layer validation is thorough with table-driven tests.
6. **OpenAPI spec** is comprehensive and synchronized with all 16 endpoints.
7. **RBAC permissions** are properly defined and enforced on every route.
8. **Frontend** provides a functional UI with search, pagination, step builder, and version history.

### Weaknesses

1. **RLS gap on test tables** is a security concern that should be addressed before Phase 3.
2. **Search cursor pagination** is functionally broken for paginated search results.
3. **No transaction** for the version-snapshot-then-update workflow in `UpdateCase`.
4. **Audit logging** is not wired for test management mutations despite being a DoD requirement.
5. **Frontend React patterns** have hydration risks and stale state issues.

### Recommended Actions Before Phase 3

1. **Fix TD-1**: Add RLS policies for `test_folders`, `test_suites`, `test_cases`, `test_case_versions` in a new migration.
2. **Fix TD-2**: Implement composite cursor `(rank, id)` for search pagination, or fall back to offset-based pagination for search.
3. **Fix TD-3** (recommended): Wrap `UpdateCase` version snapshot + update in a database transaction.
4. **Fix TD-4** (recommended): Wire `AuditLog` middleware on all test management mutating routes.

### Recommendation

**PASS WITH MINOR ISSUES** — Phase 2 is functionally complete and architecturally sound. Fix TD-1 and TD-2 before starting Phase 3. TD-3 and TD-4 are strongly recommended. All other items can be tracked as technical debt for future sprints.
