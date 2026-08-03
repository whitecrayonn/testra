# Handover — Phase 3 (Execution & Results) to Phase 4 (API Testing & Defects)

**Date:** 2026-07-16  
**Prepared by:** Cascade (AI Engineering Assistant)  
**Status:** Phase 3 approved; Phase 4 ready to begin

---

## Current Repository Status

| Area | State |
|---|---|
| Build | `go build .\...` passes (apps/api) |
| Lint / vet | `go vet .\...` passes |
| Unit tests | `go test -count=1 .\...` passes |
| Integration tests | `go test -tags=integration .\tests\integration` passes (requires PostgreSQL + `TEST_DATABASE_URL` or `DATABASE_URL`) |
| Frontend typecheck | `pnpm turbo run typecheck` passes |
| OpenAPI | `docs/api/openapi/openapi.yaml` v0.4.0 documents all Phase 3 endpoints |
| Migrations | Up to `000017_add_idempotency_records` applied cleanly |
| Docs | Final gate, security review, performance review, architecture resolution, and this handover are committed |

---

## Completed Architecture

### Core platform
- Modular Go monolith with Clean Architecture boundaries.
- PostgreSQL 16 + Redis 7 native development stack (ADR-009).
- Self-hosted authentication: JWT access + rotating opaque refresh tokens, TOTP MFA, password reset (ADR-001 / ADR-007).
- Tenant isolation via organization → workspace → project hierarchy and PostgreSQL RLS (ADR-004).
- RBAC with seeded roles/permissions and middleware enforcement.
- API keys domain implemented (hashed, scoped); auth middleware is JWT-only today (see risks).
- Audit event logging on all mutating routes.
- Response envelope, error envelope, and input validation conventions established.

### Phase 3 modules
- `results` — test runs and test run items with status lifecycle, cursor pagination, SSE progress hub, and transactional updates.
- `automationhub` — CI ingestion for JUnit XML, Playwright JSON, Cypress JSON; zero source-code retention; idempotent ingestion.
- `idempotency` — reusable PostgreSQL-backed idempotency store and middleware with SHA-256 fingerprints and configurable TTL.
- Web pages for runs list, run detail (with SSE live progress), and new run creation.

---

## Completed Modules

| Module | Status | Notes |
|---|---|---|
| `identity` | ✅ Complete | Login, register, MFA, password reset, refresh |
| `organization` | ✅ Complete | CR/L |
| `workspace` | ✅ Complete | CR/L, members |
| `project` | ✅ Complete | CR/L |
| `apikeys` | ✅ Complete | Domain/service/repo; JWT-only auth still used for ingestion |
| `rbac` | ✅ Complete | Roles, permissions, middleware |
| `audit` | ✅ Complete | Immutable event logging |
| `testmanagement` | ✅ Complete | Folders, suites, cases, versions, full-text search |
| `results` | ✅ Complete | Runs, items, SSE, pagination |
| `automationhub` | ✅ Complete | Ingestion + idempotency |
| `idempotency` (shared) | ✅ Complete | Store + middleware |
| Web dashboard (runs) | ✅ Complete | List, detail with SSE, new run |

---

## Remaining Modules (Phase 4 Scope)

| Module | Purpose | Dependencies |
|---|---|---|
| `apitesting` | API test definitions, environments, execution (zero collection retention) | `project` |
| `defects` | Defect CRUD, linking to runs/cases | `results`, `testmanagement` |
| `integrationhub` | Jira sync, CI webhooks | `project`, `results` |
| `notification` | In-app + email notifications | `identity` |

---

## Current Risks

| Risk | Impact | Likelihood | Notes |
|---|---|---|---|
| CI ingestion uses JWT, not API keys | Medium | Medium | ADR-001 requires scoped API keys for CI/CD; the `apikeys` domain exists but ingestion is not wired to `KeyAuth` |
| Synchronous ingestion for large batches | Medium | Low | Works for MVP batches; large payloads block the HTTP request (ADR-011 deferred) |
| Idempotency table grows without scheduled cleanup | Low | Low | `DeleteExpired` exists; no cron/worker schedules it yet |
| Idempotency key not validated as UUID | Low | Low | Any opaque string works; clients should use UUIDs but enforcement is optional |
| `test_runs`/`test_run_items` in PostgreSQL | Low | Low | ADR-010 accepted; ClickHouse deferred to future phase |

---

## Recommended Implementation Order for Phase 4

1. **Notifications (`notification`)**
   - Low dependency surface; start with in-app notifications and an event publisher interface.
   - Foundation for defects ("issue assigned") and integrations ("CI failed").

2. **Defects (`defects`)**
   - Core Phase 4 deliverable; depends on `results` and `testmanagement`.
   - Links defects to test cases (`testmanagement`) and test runs (`results`).
   - Provides natural notifications targets.

3. **Integration Hub (`integrationhub`)**
   - Starts with CI webhooks (ingest trigger) and then Jira sync.
   - Depends on `project` and `results` via ports; should not import internals.

4. **API Testing (`apitesting`)**
   - Largest Phase 4 module: request definitions, environments, runner/executor.
   - Must honor zero collection retention (no raw API collection payloads stored).
   - Can reuse the existing `results` run model for execution output.

Each module should be delivered with:
- Domain, ports, repository, service, handler, module wiring.
- Migration `up` and `down` SQL.
- Unit tests for service layer.
- Integration tests for handlers (as appropriate).
- OpenAPI spec updates before or alongside code.
- Web UI pages + API client + TypeScript types.

---

## Implementation Constraints

### Architecture
- Clean Architecture: domain → application → ports → adapters. No cross-module internal imports.
- Every tenant-scoped table must have `organization_id` or equivalent and RLS policies.
- Every mutating endpoint must have RBAC permission check and audit logging.
- `Idempotency-Key` is required for side-effecting commands, ingestion, exports, webhooks (ADR-006).

### API
- URL major version `/api/v1`.
- Cursor pagination for new list endpoints (default 50, max 100).
- Response envelope: `{ data, meta, error }`.
- Error codes stable; messages safe.

### Data and Privacy
- Zero customer source code retention.
- Zero raw API collection payload retention in `apitesting`.
- Sensitive data encrypted in transit and at rest.

### Testing
- `go test -count=1 ./...` must pass.
- `go test -tags=integration ./tests/integration` must pass (requires PostgreSQL).
- `pnpm turbo run typecheck` must pass.

### Deployment
- Native development environment (ADR-009); Docker optional.
- MVP target: Ubuntu VM + systemd + Nginx.
- Migrations applied by `cmd/migrator`.

---

## ADR Summary

| ADR | Topic | Status | Relevance to Phase 4 |
|---|---|---|---|
| ADR-001 | Hybrid self-hosted auth | Accepted | API key auth for CI/webhooks; MFA requirements |
| ADR-002 | Documentation source-of-truth | Accepted | Update OpenAPI before implementation |
| ADR-003 | Production deployment strategy | Accepted (amended by ADR-009) | Ubuntu VM/systemd MVP |
| ADR-004 | Tenant isolation | Accepted | RLS, scope propagation |
| ADR-005 | Backup / disaster recovery | Accepted | Retention, restore testing |
| ADR-006 | API standards | Accepted | Idempotency, pagination, versioning |
| ADR-007 | Security standards | Accepted | Rate limits, sessions, secrets |
| ADR-008 | Performance targets | Accepted | p95 targets, capacity |
| ADR-009 | Native development environment | Accepted | Local dev stack |
| ADR-010 | PostgreSQL for Phase 3 results | Accepted | `results` remains in PostgreSQL until volume justifies ClickHouse |
| ADR-011 | Synchronous ingestion for MVP | Accepted | Async queue deferred; keep `Service.Ingest` stateless |
| ADR-012 | Idempotency-Key for ingestion | Accepted | Reuse/adapt middleware for Phase 4 command endpoints |

---

## Definition of Done for Phase 4

- [ ] `apitesting` module: request definitions, environments, execution engine (zero collection retention).
- [ ] `defects` module: CRUD, linking to runs/cases, status lifecycle.
- [ ] `integrationhub` module: CI webhooks + Jira sync.
- [ ] `notification` module: in-app + email notifications.
- [ ] Web: API test builder, defect board, notification center.
- [ ] OpenAPI spec updated for all Phase 4 endpoints.
- [ ] Unit tests for each new service layer.
- [ ] Integration tests for critical new flows (webhooks, defect CRUD, API test execution).
- [ ] RBAC permissions added and enforced for new modules.
- [ ] RLS policies added for new tenant-scoped tables.
- [ ] Audit logging on all mutating endpoints.
- [ ] `go build`, `go vet`, `go test`, `go test -tags=integration`, and `pnpm turbo run typecheck` pass.
- [ ] Phase 4 security and performance reviews completed.
- [ ] `PHASES.md` updated to mark Phase 4 in progress/approved.
- [ ] `docs/engineering/progress/` report saved.

---

## Immediate Next Steps

1. Open `PHASES.md` and mark Phase 4 as **In Progress** when work begins.
2. Create the `notification` module scaffolding first (domain, ports, repository, service, handler, module).
3. Update OpenAPI for notification endpoints before implementation.
4. Create migration with `up` and `down` SQL for notifications tables.
5. Add unit tests and a minimal handler integration test.
6. Add corresponding web UI pages and TypeScript types.

---

## Sign-off

Phase 3 is complete and the repository is ready for Phase 4 implementation. No Phase 4 code should be merged until the OpenAPI spec and migration plan for the first module are in place.
