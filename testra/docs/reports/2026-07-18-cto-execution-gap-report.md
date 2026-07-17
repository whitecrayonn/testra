# TESTRA CTO Execution — Architectural Gap Report

**Date:** 2026-07-18  
**Scope:** Full monorepo audit against BRD, strategy, AI_CONTEXT, AI_MEMORY, AI_RULES, ADR-001–012, and technical-debt audits.  
**Prepared by:** Principal Full-Stack & Lead DevOps Engineer (AI execution agent)

---

## Executive summary

The repository is a functional Phase 1–3 modular monolith (Go API, Next.js web, Python ML skeleton, worker stub). The feature matrix and roadmaps estimate **overall MVP readiness at ~40%**. The biggest remaining gaps are:

1. **Missing domain modules** — `analytics`, `intelligence`, `integrationhub`, and `billing` are empty package directories.
2. **Missing background worker** — `apps/worker` contains only a `fmt.Println` stub.
3. **Missing ML inference** — `apps/ml/api/main.py` only exposes `/health`.
4. **Placeholder frontend pages** — settings members, roles, billing, audit-logs, organization/workspace/profile/security are `PlaceholderPage` components.
5. **Infrastructure is scaffold-only** — K8s base has only API deployment+service; Terraform modules directory is empty; Docker Compose has no application services.
6. **OpenAPI / ROUTES / canonical docs** lag behind code; defects is implemented in backend but not fully reflected in ROUTES.md / OpenAPI.
7. **Several P0 production blockers remain open** from `TECHNICAL_DEBT.md`: rate-limiting not wired, API-key auth for `/ingest` incomplete, SSE auth uses query token, default secrets in `.env.example`, no client route guards, audit durability issues.

This report is the baseline for the execution work below.

---

## Backend gaps

### Implemented modules
| Module | Files | Status |
|--------|-------|--------|
| identity | 7 files | Functional |
| organization | 6 files | Functional |
| workspace | 6 files | Functional |
| project | 7 files | Functional |
| apikeys | 7 files | Functional |
| rbac | 1 file | Functional |
| testmanagement | 7 files | Functional |
| results | 7 files | Functional |
| automationhub | 5 files | Functional |
| defects | 7 files | Functional |
| notification | 7 files | Functional |
| audit | 4 files | Functional |

### Missing / placeholder modules
| Module | Current state | Required per directive |
|--------|---------------|------------------------|
| `analytics` | `module.go` with empty package line | Dashboard aggregations, trends, release-readiness |
| `intelligence` | `module.go` with empty package line | Flaky detection, failure classification, risk scoring, ML client |
| `integrationhub` | `module.go` with empty package line | Jira/GitHub/GitLab/Slack integrations, CI webhooks |
| `billing` | `module.go` with empty package line | Subscriptions, usage, invoices, Stripe adapter |
| `apitesting` | `.gitkeep` only | API test definitions/execution (Phase 4) |

### Route wiring
`server.go` registers all current modules; the four modules above are **not imported** and have **no routes**.

### Migrations
Existing migrations `000001`–`000020` cover users, orgs, workspaces, projects, RBAC, API keys, RLS, test management, test runs, idempotency, notifications, and defects. No migrations exist for analytics, intelligence, integration, or billing tables.

### RLS / tenant isolation
RLS is enabled for existing tenant-scoped tables via `000009`, `000014`, `000015`, `000017`, `000018`, `000020`. New modules must follow the same pattern: table `organization_id` or resolvable `workspace_id`/`project_id`, RLS policy using `app.tenant_id`, and an `ALTER TABLE … FORCE ROW LEVEL SECURITY` line.

### Idempotency
`POST /ingest` already uses `idempotency.PostgresStore` with operation `ingest`. Other new side-effecting modules should respect `Idempotency-Key` via the same middleware.

### Worker
`apps/worker/go.mod` is an empty module. `apps/worker/cmd/worker/main.go` only prints a message. No Asynq/Redis queue, no processors.

---

## ML service gaps

`apps/ml/api/main.py` is a 9-line FastAPI app with only `/health`. Requirements:
- `/predict-flaky` — accepts test history, returns flaky-probability and explanation.
- `/classify-failure` — accepts failure log/stack trace, returns cluster label and confidence.
- Feature/model scaffolding in `features/` and `models/` is empty.
- No Go client in `intelligence` module to call ML service.

---

## Frontend gaps

### Placeholder settings pages
All pages below use `PlaceholderPage`:
- `dashboard/settings/members/page.tsx`
- `dashboard/settings/roles/page.tsx`
- `dashboard/settings/billing/page.tsx`
- `dashboard/settings/audit-logs/page.tsx`
- `dashboard/settings/organization/page.tsx`
- `dashboard/settings/workspace/page.tsx`
- `dashboard/settings/profile/page.tsx`
- `dashboard/settings/security/page.tsx`

### Workspace route tree
`[workspace]/api-tests/` and `[workspace]/runs/` directories are empty. `[workspace]/settings/` is empty.

### API wrappers
`apps/web/features/{analytics,api-testing,automation-hub,defects,settings}` are mostly `.gitkeep`. `settings` and `analytics` wrappers do not exist.

### Cross-workspace routing
`[workspace]/page.tsx` stores the workspace UUID in `localStorage` but does not validate the URL slug against the backend.

---

## Infrastructure gaps

### Docker Compose
- No `api`, `web`, `worker`, or `ml` services.
- `mailpit` has no `healthcheck`.
- `.dockerignore` not verified.

### Kubernetes
- Base only has API deployment and service.
- Missing ConfigMap, Secret, probes, Ingress, HPA, PodDisruptionBudget, ServiceAccount.
- Overlays only change namespace and image tag.
- No web/worker/migrator/ML manifests.

### Terraform
- `infra/terraform/main.tf` only configures AWS provider.
- `infra/terraform/modules/` is empty.
- No VPC, EKS, RDS, ElastiCache, S3, Route53, WAF modules.

---

## Verification gaps

- `apps/api/tests/integration/` has only 3 files; integration coverage is minimal.
- Worker has no tests.
- ML has no tests.
- Frontend has no E2E tests.
- `make test` / `pnpm turbo run typecheck build` will fail if new modules are not fully wired.

---

## Execution plan

1. Implement `apps/api/internal/analytics` (domain, ports, repository, service, handler, module, migration).
2. Implement `apps/api/internal/intelligence` with ML service client (domain, ports, repo, service, handler, module, migration).
3. Implement `apps/api/internal/integrationhub` (domain, ports, repo, service, handler, module, migration).
4. Implement `apps/api/internal/billing` (domain, ports, repo, service, handler, module, migration).
5. Implement `apps/ml/api/main.py` with `/predict-flaky` and `/classify-failure`.
6. Implement `apps/worker` with Asynq/Redis processors for notifications and aggregations.
7. Wire new modules into `server.go` and add required permissions/seeds.
8. Add frontend settings pages and API wrappers for members, roles, billing, audit-logs, organization/workspace, profile/security.
9. Complete Docker Compose app services, K8s manifests, and Terraform modules.
10. Run `go test ./...`, `pnpm turbo run typecheck build`, and integration tests.

---

## References

- `testra/docs/FEATURE_MATRIX.md`
- `testra/docs/engineering/ROADMAP.md`
- `testra/docs/archive/merged-sources/TECHNICAL_DEBT.md`
- `testra/docs/archive/merged-sources/backend-audit.md`
- `testra/docs/archive/merged-sources/frontend-audit.md`
- `testra/docs/archive/merged-sources/infra-audit.md`
- `testra/docs/BIBLICAL_TESTRA.md` §Domain Modules, §Do Not Break List, §Data Model
- `testra/docs/AI_CONTEXT.md`, `AI_MEMORY.md`, `AI_RULES.md`
