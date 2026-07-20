# Architecture Review v2 — Testra Platform

**Date:** 2026-07-19  
**Scope:** Backend monolith (`apps/api`), module boundaries, dependency graph, data flow, and deployment architecture.  
**Goal:** Evaluate the current architecture against the documented Clean Architecture / modular-monolith design in `MODULE_DEPENDENCIES.md` and identify consolidation, coupling, and production-readiness gaps.

## 1. Executive Summary

Testra is implemented as a **Go modular monolith** with clear package ownership per bounded context (`identity`, `workspace`, `project`, `testmanagement`, `results`, `integrationhub`, etc.). Cross-cutting concerns (`shared`) provide primitives for config, HTTP, JWT, middleware, pagination, tenant resolution, idempotency, and now SSRF protection.

The architecture is **fit for an MVP and early production** but has three categories of issues:

1. **Dependency graph drift** — the documented module map does not match the current implementation; some modules couple directly to others instead of through ports.
2. **Scalability bottlenecks** — the job queue is implemented inside PostgreSQL, and `metrics` is an in-memory global registry that is not persisted in production.
3. **Partial / placeholder modules** — `apitesting` is essentially empty and `integration` only contains a tenant-isolation test, creating confusion.

## 2. Current Module Inventory

| Module | Responsibility | Status |
|--------|----------------|--------|
| `identity` | Registration, login, MFA, password reset, JWT issuance, refresh-token families | Solid |
| `organization` | Organization CRUD and membership | Okay |
| `workspace` | Workspace CRUD and members | Okay |
| `project` | Project CRUD under workspace | Okay |
| `testmanagement` | Test cases, suites, folders, steps | Recently hardened (pq.Array, validation) |
| `results` | Test runs, run items, ingestion facts | Okay |
| `automationhub` | JUnit / Playwright / Cypress ingestion | **Coupling issue** (see §3.3) |
| `integrationhub` | External adapters (Jira, GitHub, GitLab, Slack, webhook) | Recently hardened (SSRF, secret masking) |
| `notification` | In-app, email, Slack/Teams/webhook channels | Recently hardened (SSRF) |
| `apikeys` | API key creation, revocation, validation | Okay (RowsAffected fix applied) |
| `audit` | Audit event logging | Recently hardened |
| `rbac` | SQL permission loader | Minimal; only `loader.go` exists |
| `intelligence` | ML client / local heuristics | In-memory metrics coupling (see §3.4) |
| `analytics` | Analytics data access | Not reviewed in detail |
| `billing` | Billing domain | Not reviewed in detail |
| `defects` | Defect tracking | Not reviewed in detail |
| `shared` | Cross-cutting platform primitives | Good shape |
| `queue` | PostgreSQL-backed job queue | Works for MVP, limited scalability |
| `metrics` | In-memory Prometheus-style registry | Not persisted; `db` field is unused in production |
| `apitesting` | (placeholder) | Empty package except `.gitkeep` |
| `integration` | (misplaced) | Contains only `tenant_isolation_test.go` |

## 3. Findings

### 3.1 Dependency Graph Is Out of Date (High)

`docs/architecture/MODULE_DEPENDENCIES.md` only lists 10 modules in its mermaid diagram and omits `automationhub`, `analytics`, `billing`, `defects`, `intelligence`, `metrics`, `queue`, `apitesting`, `integration`, and `apikeys`. The ADR directory is more complete but the primary dependency map is now misleading.

**Impact:** New engineers may import the wrong packages or create cycles because the canonical diagram no longer reflects reality.

**Recommendation:**
- Update `MODULE_DEPENDENCIES.md` to include every module in `apps/api/internal`.
- Add arrows discovered by `go list`:
  - `automationhub -> results` (currently direct)
  - `intelligence -> metrics` (currently direct)
  - `shared/server -> all modules` (composition root, expected)
  - `cmd/worker -> analytics, billing, integrationhub, intelligence, metrics, notification, queue` (expected)

### 3.2 AutomationHub Directly Depends on Results (High)

`automationhub/service.go` and `automationhub/module.go` import `github.com/testra/testra/apps/api/internal/results` directly and use `results.TestRun`, `results.TestRunItem`, and `results.Repository`. This violates the Clean Architecture rule that a module may only depend on abstractions (ports) owned by the consumer.

**Evidence:**
```go
import "github.com/testra/testra/apps/api/internal/results"

type ResultsRepo interface {
    CreateRun(ctx context.Context, run *results.TestRun) error
    ...
}
```

**Impact:**
- `results` cannot evolve without risking `automationhub` compile failures.
- `automationhub` service is building `results` domain objects directly, duplicating the responsibility of the `results` module.

**Recommendation:**
- Define `automationhub`’s own ingestion port, e.g.:
  ```go
  type IngestionRepo interface {
      CreateRun(ctx context.Context, run IngestionRun) error
      CreateItem(ctx context.Context, runID uuid.UUID, item IngestionItem) error
      UpdateRun(ctx context.Context, runID uuid.UUID, totals RunTotals) error
  }
  ```
- Provide an adapter in `results` (or in `automationhub` as `resultsadapter`) that maps `IngestionRun` to `results.TestRun`.
- Keep `automationhub` free of any `results` import.

### 3.3 Intelligence Directly Depends on Metrics (Low)

`intelligence/mlclient.go` calls `metrics.RecordMLCall`. `metrics` is a cross-cutting package and the dependency is tolerable, but it introduces a global side effect inside an adapter.

**Recommendation:**
- Accept a `MetricsRecorder` interface in `intelligence.Service`/`MLClient` constructor and inject the recorder from `server.go`.
- This lets unit tests use a no-op recorder and removes the global dependency.

### 3.4 Composition Root Is Centralized in `shared/server` (Medium)

`shared/server/server.go` imports every module and wires handlers, services, repositories, and middleware. This is the natural composition root for a monolith, but the file is large (650+ lines) and mixes routing, middleware ordering, dependency injection, and module lifecycle.

**Recommendation:**
- Keep `server.go` as the wiring file but split it into `module wiring`, `middleware chain`, and `route registration` sections, or extract a `cmd/api/wire.go`/`providers.go` if a DI tool is adopted.
- Ensure middleware order is documented and enforced by tests (e.g. `TestMiddlewareOrder`).

### 3.5 Job Queue Is PostgreSQL-Based (Medium)

`queue/queue.go` uses `queue_jobs` table with `FOR UPDATE SKIP LOCKED` for polling. This is a robust single-node queue but has limits:
- Throughput is bounded by database connection pool and polling frequency.
- Job ordering is only by `created_at`; no priority, delay, or scheduled-at precision beyond `scheduled_at <= NOW()`.
- The worker `cmd/worker/main.go` sleeps for 5s on empty queues.

**Recommendation for production:**
- Adopt a purpose-built queue (Asynq, temporal, or cloud-native SQS/SNS) before scaling past tens of jobs/second.
- Keep the PostgreSQL queue as a fallback for self-hosted deployments but make the backend pluggable.

### 3.6 Metrics Registry Is Not Production-Persistent (Medium)

`metrics/metrics.go` defines `defaultRegistry = newRegistry(nil)`. In production the registry collects counts/histograms in process memory and is never flushed to a time-series store. If the process restarts, metrics are lost.

**Recommendation:**
- Replace or augment the in-memory registry with a real metrics exporter (Prometheus `/metrics` endpoint, OTel, or StatsD).
- Until then, document that `metrics` is for local observability only.

### 3.7 `apitesting` and `integration` Packages Are Placeholders (Low)

- `apitesting/module.go` is an empty package.
- `integration/` contains only `tenant_isolation_test.go` — a shared-style test living in a module-named folder, which is confusing.

**Recommendation:**
- Delete or populate `apitesting`.
- Move `integration/tenant_isolation_test.go` to `shared/tenant` (where the resolver it tests lives) or rename the package to `tenant_test`.

### 3.8 Tenant Resolution in `shared/tenant` (Low)

`shared/tenant/resolver.go` resolves an organization ID from workspace/project/API-key/test-run/defect IDs by querying tables that logically belong to other modules. It does this via raw SQL against `shared/db`, which avoids an import cycle but creates a **schema-coupling** in `shared`.

**Recommendation:**
- Keep the current resolver for now, but require each new domain module to register a `TenantResolver` function in `server.go` rather than extending `resolver.go`.
- This prevents `shared` from accumulating knowledge of every domain table.

### 3.9 `rbac` Module Is Under-Developed (Medium)

`rbac` only contains `loader.go` with a SQL permission loader. There are no domain models, service, or handler tests. The `shared/middleware/rbac.go` middleware uses a `PermissionLoader` interface, which is correct, but the `rbac` module itself is not a full bounded context.

**Recommendation:**
- Decide whether `rbac` should own role/permission definitions and membership checks, or remain a read-only loader.
- If it remains a loader, consider renaming it to `permissions` and moving it under `shared`.

### 3.10 Data Store Roles Are Well Separated (Good)

The platform uses:
- PostgreSQL for transactional state.
- Redis for rate limiting, idempotency, and queue coordination.
- ClickHouse for analytical result events.
- Object storage for artifacts.

This separation is consistent with `SYSTEM_FLOWS.md` and avoids analytic workloads affecting the transactional path.

## 4. Production-Readiness Matrix

| Capability | Status | Notes |
|------------|--------|-------|
| Bounded contexts / module ownership | ✅ Good | Clear package per domain |
| Clean Architecture dependency direction | ⚠️ Fair | `automationhub -> results` violation; `intelligence -> metrics` global |
| Tenant isolation | ✅ Good | RLS + `app.tenant_id` + parameterized queries |
| Idempotency | ✅ Good | Middleware + Redis |
| Authentication / JWT | ✅ Good | RS256, JWKS, rotation |
| Authorization / RBAC | ⚠️ Fair | Loader only; no module-level service |
| Audit logging | ✅ Good | Recently fixed to log all attempts |
| Outbound SSRF protection | ✅ Good | New `shared/security` package |
| Queue scalability | ⚠️ Fair | PostgreSQL queue; upgrade path needed |
| Metrics/observability | ⚠️ Fair | In-memory registry; needs real exporter |
| Module documentation accuracy | ❌ Needs update | `MODULE_DEPENDENCIES.md` is stale |
| Placeholder modules | ❌ Needs cleanup | `apitesting`, `integration` |

## 5. Remediation Backlog (Architecture)

In priority order:

1. **Update `MODULE_DEPENDENCIES.md`** to match the current module set and dependencies.
2. **Decouple `automationhub` from `results`** by introducing an ingestion port and adapter.
3. **Refactor `intelligence` metrics dependency** to be constructor-injected.
4. **Document or adopt a real queue backend** for production scale.
5. **Add a Prometheus/OTel metrics exporter** and deprecate the in-memory-only default registry.
6. **Clarify `rbac` scope** — either grow it to a full module or relocate to `shared/permissions`.
7. **Clean up `apitesting` and `integration`** placeholder packages.
8. **Add middleware-order regression tests** in `shared/server`.

## 6. Conclusion

The modular-monolith design is sound and the platform has made strong security and resilience improvements in this review pass. The highest-priority architectural work now is aligning the documented dependency graph with reality and removing the `automationhub -> results` direct dependency. Once those are done, the next strategic investments should be production-grade queuing and observability exporters.
