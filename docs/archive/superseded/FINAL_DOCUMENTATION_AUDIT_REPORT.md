# Final Documentation Audit Report

**Scope:** All active markdown files under `testra/docs/`, plus selected `apps/*` READMEs, `testra/docs/api/openapi/openapi.yaml`, `apps/api/internal/shared/server/server.go`, and `apps/api/migrations/`.

**Method:**
1. Read every active canonical doc and the prior `DOCUMENTATION_CONSOLIDATION_REPORT.md` / `DOCUMENTATION_QA_REPORT.md`.
2. Cross-checked route tables, module ownership, migration count, OpenAPI tags, ADR status, and feature matrices against the running code.
3. Ran an automated link and code-reference checker (`scripts/doc_audit_check.py`) over the full docs tree.

---

## Executive summary

- **Active canonical docs:** 33
- **Archived docs:** 48
- **Broken markdown links in active docs:** 0
- **Stale code-path references in active docs:** 49 (intentional historical/future references; see §Link health)
- **Broken markdown links in archive:** 42 (expected for superseded documents)
- **Stale code-path references in archive:** 171 (expected for superseded documents)

**Overall documentation health score: 89/100**

**Verdict:** **APPROVED WITH MINOR RECOMMENDATIONS.** The canonical documentation set is coherent, cross-referenced, and aligned with the codebase. Remaining issues are either intentional historical records or planned-deliverable mentions, not broken links or contradictions.

---

## Scoring

| Category | Weight | Score | Rationale |
|----------|--------|-------|-----------|
| Completeness | 25% | 88 | BIBLICAL, ROADMAP, FEATURE_MATRIX, and OpenAPI cover the MVP well; future phases (defects, analytics, enterprise) are documented at the right level of detail. A few stub entrypoints (`apps/api/cmd/worker`) were under-documented. |
| Accuracy vs. code | 25% | 90 | Route inventory, middleware order, RLS/tenant isolation, migrations, and module list match `server.go` and the filesystem. RBAC permission-name drift is documented as a known caveat. |
| Consistency / no contradictions | 20% | 92 | One contradiction was found and fixed: `FEATURE_MATRIX.md` claimed `apps/ml` was a placeholder, while the skeleton `/health` endpoint already exists. Minor duplication remains between `BIBLICAL_TESTRA.md` flow diagrams and `SYSTEM_FLOWS.md`. |
| Navigation & canonical ownership | 15% | 90 | `README.md`, `BIBLICAL_TESTRA.md`, and `ADR-002` clearly define canonical sources. `ROUTES.md` intro was corrected to point to canonical docs instead of archived audit files. |
| Link / stale-ref health | 15% | 82 | Zero broken markdown links in active docs. The 49 flagged code-path refs are either historical paths inside `DOCUMENTATION_CONSOLIDATION_REPORT.md` or planned deliverables in `ROADMAP.md` / `DOCUMENTATION_QA_REPORT.md`. They should be converted to plain prose or qualified with "planned" to reduce noise. |
| **Overall** | 100% | **89** | Strong documentation hygiene; small refinements recommended. |

---

## Canonical topic ownership map

| Concern | Canonical source |
|---------|------------------|
| Engineering handbook & navigation | `testra/docs/BIBLICAL_TESTRA.md` |
| Documentation index | `testra/docs/README.md` |
| Product vision, MVP scope, current state | `testra/docs/PROJECT_OVERVIEW.md` |
| Implementation phases, priorities, technical debt | `testra/docs/engineering/ROADMAP.md` |
| Onboarding, governance, dev workflow | `testra/docs/engineering/ONBOARDING.md` |
| Coding standards | `testra/docs/engineering/ENGINEERING_STANDARDS.md` |
| HTTP API contract | `testra/docs/api/openapi/openapi.yaml` + `apps/api/internal/shared/server/server.go` |
| API conventions | `testra/docs/api/API_DESIGN_GUIDELINES.md` |
| Database schema, RLS, ERD | `testra/apps/api/migrations/*.sql` + `testra/docs/architecture/DATABASE_GUIDE.md` |
| System/request/sequence flows | `testra/docs/architecture/SYSTEM_FLOWS.md` |
| Route inventory | `testra/docs/ROUTES.md` |
| Feature status | `testra/docs/FEATURE_MATRIX.md` |
| Security checklist | `testra/docs/security/SECURITY_CHECKLIST.md` |
| Deployment & operations | `testra/docs/deployment/DEPLOYMENT_GUIDE.md` + `testra/docs/operations/*.md` |
| ADRs | `testra/docs/architecture/adrs/ADR-*.md` |
| Consolidation history | `testra/docs/reports/DOCUMENTATION_CONSOLIDATION_REPORT.md` |
| QA / audit | `testra/docs/reports/DOCUMENTATION_QA_REPORT.md` + this report |

---

## Cross-check findings

### Backend
- `apps/api/internal/shared/server/server.go` route tree matches `testra/docs/ROUTES.md` and `BIBLICAL_TESTRA.md` route table.
- Middleware order (`Logger` → `Recoverer` → `RequestID` → `Content-Type` → CORS → `MaxBodySize` → `Auth` → `TenantContext` → `RequirePermission` → `AuditLog` / `Idempotency`) is documented consistently across `BIBLICAL_TESTRA.md`, `SYSTEM_FLOWS.md`, and `server.go`.
- Migrations are numbered `000001`–`000018` and include the Phase 3.5 `notification` module (migration `000018`).
- `rbac` is implemented as a `SQLPermissionLoader` plus `shared/middleware/rbac.go`; `FEATURE_MATRIX.md` notes the permission-name drift correctly.
- `apikeys.Service.Validate` exists, but no middleware consumes `X-API-Key` for `/ingest`. This is documented as a P0 gap in `FEATURE_MATRIX.md` and `ROADMAP.md`.

### Frontend
- Next.js 15 App Router routes in `apps/web/app/` line up with `ROUTES.md`.
- `/dashboard/settings/notifications` and `/dashboard/notifications` pages exist, matching the implemented `notification` backend module.
- `/[workspace]/defects` is a placeholder and is correctly marked as Phase 4.

### ML service
- `apps/ml/api/main.py` provides a `/health` endpoint skeleton.
- `BIBLICAL_TESTRA.md` and `PROJECT_OVERVIEW.md` already described this skeleton; `FEATURE_MATRIX.md` was corrected to `🔄` and a clarifying status note.

### OpenAPI
- `testra/docs/api/openapi/openapi.yaml` is version `0.4.0` and defines tags for all implemented domains, including the new `Notifications` group.
- `apiKeyAuth` security scheme is still missing; this aligns with the documented backend gap.

### Migrations
- 18 migration pairs in `apps/api/migrations/`.
- `BIBLICAL_TESTRA.md` states migrations `000001`–`000018`; counts match.

### ADRs
- 12 accepted ADRs in `testra/docs/architecture/adrs/`.
- `BIBLICAL_TESTRA.md` canonical-sources table references ADRs and `ADR-002` specifically declares documentation source-of-truth rules.

---

## Duplication & contradiction audit

### Contradictions found and resolved
- `FEATURE_MATRIX.md` listed `ML inference service` as `❌` / "`apps/ml` is a placeholder". The repo contains a `/health` skeleton, and `BIBLICAL_TESTRA.md` / `PROJECT_OVERVIEW.md` described it as implemented-at-skeleton level. Corrected `FEATURE_MATRIX.md` to `🔄` with the accurate status note.

### Duplication remaining (acceptable but noted)
- `BIBLICAL_TESTRA.md` and `testra/docs/architecture/SYSTEM_FLOWS.md` both contain Mermaid flow diagrams for request lifecycle, authentication, and tenancy. `BIBLICAL` could reference `SYSTEM_FLOWS.md` for the full diagrams and keep only summaries.
- `ROUTES.md` and `BIBLICAL_TESTRA.md` both maintain route inventories; this is acceptable because `ROUTES.md` is the detailed route reference and `BIBLICAL` is the high-level handbook.
- `PROJECT_OVERVIEW.md` and `ROADMAP.md` overlap on implementation status and technical debt; the former is product-state focused, the latter is engineering-plan focused, so the overlap is intentional.

---

## Link and stale-reference check

A Python checker (`scripts/doc_audit_check.py`) scanned every markdown file under `testra/docs/` for broken markdown links and missing code-path references.

| Metric | Active docs | Archive |
|--------|-------------|---------|
| Broken markdown links | 0 | 42 (expected in superseded docs) |
| Missing code-path refs | 49 | 171 (expected in superseded docs) |

The 49 active missing code-path refs fall into three buckets:
1. **Historical record (41 refs):** `DOCUMENTATION_CONSOLIDATION_REPORT.md` lists former paths such as `docs/handover/...`, `docs/engineering/PHASES.md`, etc. These are expected and useful as an audit trail.
2. **Planned deliverables (5 refs):** `ROADMAP.md` mentions `docs/operations/INCIDENT_RESPONSE_RUNBOOK.md`, `docs/reports/archive/`, and `docs/enterprise/` as future documentation deliverables.
3. **Known gaps / recommendations (3 refs):** `DEPLOYMENT_GUIDE.md` and `DOCUMENTATION_QA_REPORT.md` mention `apps/worker/go.sum` and `apps/web/.env.example` as missing artifacts.

These are not broken markdown links, but they are noise for automated reference checkers. Recommendation: qualify planned/known-missing paths with prose such as "add `apps/web/.env.example`" rather than bare backticks.

---

## Changes made during this audit

- `testra/docs/ROUTES.md` — intro now points to `BIBLICAL_TESTRA.md`, `PROJECT_OVERVIEW.md`, and `ONBOARDING.md` instead of `archive/merged-sources/backend-audit.md` and `frontend-audit.md`.
- `testra/docs/BIBLICAL_TESTRA.md` — canonical-sources table and key-files index now include `DOCUMENTATION_QA_REPORT.md`; `apps/api/cmd` purpose and key-files index now mention the stub `worker`.
- `testra/docs/README.md` — findings summary now mentions `reports/` as the home for consolidation and QA audit reports.
- `testra/docs/FEATURE_MATRIX.md` — `ML inference service` row corrected from `❌` to `🔄` and status clarified.
- `testra/apps/worker/README.md` — stale `docs/engineering/PHASES.md` link fixed to `docs/engineering/ROADMAP.md`.

---

## Open recommendations

1. **Reduce stale-ref noise in canonical reports.** Convert bare backticks around historical/future paths in `DOCUMENTATION_CONSOLIDATION_REPORT.md`, `ROADMAP.md`, `DOCUMENTATION_QA_REPORT.md`, and `DEPLOYMENT_GUIDE.md` to prose or footnotes.
2. **Deduplicate flow diagrams.** Move the canonical Mermaid diagrams to `SYSTEM_FLOWS.md` and have `BIBLICAL_TESTRA.md` reference them.
3. **Resolve worker entrypoint ambiguity.** Decide whether `apps/api/cmd/worker` or `apps/worker` is the canonical future worker location and update `BIBLICAL_TESTRA.md` accordingly.
4. **Add CI markdown link checker.** Integrate `markdown-link-check` or `lychee` and the `scripts/doc_audit_check.py` check into `.github/workflows/` to catch regressions.
5. **Complete app READMEs.** `apps/api`, `apps/web`, `apps/worker`, and `apps/ml` READMEs should link to canonical docs and state current status (several still have one-line stubs).
6. **Resolve documented P0 code gaps.** API-key auth for `/ingest`, rate-limiting wiring, and frontend token refresh are documented as missing. Fixing the code will let the corresponding docs move from `🔄`/`❌` to `✅`.

---

## Verdict

**Documentation is in good shape and safe to rely on.** Canonical sources are clearly owned, the archive is correctly separated from active docs, and the active set is consistent with the backend, frontend, migrations, OpenAPI, and ADRs. The remaining items are refinements and known code gaps rather than documentation blockers.

---

## Appendix: reproducing this audit

The audit script is `testra/scripts/doc_audit_check.py`. Run it from `testra/`:

```bash
python scripts/doc_audit_check.py > doc_audit_check.json
```

It reports:
- every active and archived markdown file,
- broken markdown links,
- missing inline code-path references (relative to `testra/` or repo root for `04_Architecture/` and `testra-*.md`).

