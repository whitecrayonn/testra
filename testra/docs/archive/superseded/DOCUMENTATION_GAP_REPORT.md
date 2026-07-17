# Documentation Gap Report

> This report lists the documentation gaps, inconsistencies, duplicates, and stale references found during the audit. For remediation owners and dates see [`DOCUMENTATION_ROADMAP.md`](DOCUMENTATION_ROADMAP.md).

## 1. Conflicting / non-canonical documents

| # | Document | Problem | Recommended action | Priority |
|---|----------|---------|--------------------|----------|
| G1 | `04_Architecture/testra-software-architecture-decisions.md` | Pre-implementation draft containing proposed alternatives (microservices, managed identity, different DB choices) that conflict with accepted ADRs (ADR-001, ADR-003, ADR-004, ADR-009, ADR-010). | Move to `docs/reports/archive/` as a historical artifact and add a header explaining it was superseded by ADR-001 through ADR-012. | P0 |
| G2 | `docs/engineering/ENGINEERING_DOCUMENTATION_REPORT.md` | Framed as a documentation pass report but states Phase 1 is the latest completed work and Phase 2 is active. Superseded by `PHASES.md` and `HANDOVER_PHASE3_TO_PHASE4.md`. | Add a prominent `STALE` header at the top pointing to current canonical docs, then archive. | P1 |
| G3 | `docs/engineering/TESTRA_ENGINEERING_HANDOVER_REPORT.md` | States Phase 2 is in progress; superseded by `PHASES.md` and `HANDOVER_PHASE3_TO_PHASE4.md`. | Add a prominent `STALE` header and archive or delete after reconciliation. | P1 |
| G4 | `docs/reports/DOCUMENTATION_HEALTH_REPORT.md` | Superseded by `DOCUMENTATION_HEALTH_AUDIT.md` in this pass. | Replace body with a redirect to `DOCUMENTATION_HEALTH_AUDIT.md` or archive. | P1 |

## 2. Stale findings in handover audits

Several `handover/` audit files were written before Phase 3.5 completion and still describe fixed issues as current blockers.

| # | Document | Stale finding | Actual state | Recommended fix | Priority |
|---|----------|---------------|--------------|-----------------|----------|
| G5 | `docs/handover/frontend-audit.md` §8.6 / §10.5 | SSE stream cannot authenticate because `EventSource` cannot send `Authorization` header. | `Auth` middleware now accepts `Authorization: Bearer` or `access_token` query parameter; SSE works with query-token auth. | Move fixed issue to a "Resolved in Phase 3.5" section and update the active finding list. | P1 |
| G6 | `docs/handover/frontend-audit.md` §10.2 | MFA QR code rendered as text. | `mfa-setup` page renders `qr_code` as an `<img src={qr_code} />`. | Mark as resolved. | P1 |
| G7 | `docs/handover/frontend-audit.md` §10.4 | Project key generation can contain hyphens, rejected by backend. | Frontend generates uppercase alphanumeric keys matching `^[A-Z][A-Z0-9]{1,9}$`. | Mark as resolved. | P1 |
| G8 | `docs/handover/frontend-audit.md` §10.3 | Onboarding omits `slug` fields. | Onboarding sends explicit `slug` for org and workspace. | Mark as resolved. | P1 |
| G9 | `docs/handover/frontend-audit.md` §10.9 | Settings sub-pages are placeholders; API keys settings page does not exist. | `/dashboard/settings/api-keys` and `/dashboard/settings/notifications` are implemented. | Update status; list which settings pages remain placeholders. | P1 |
| G10 | `docs/handover/functional-audit.md` §3 / §4 | P0 list: SSE broken, API key auth for `/ingest` broken, rate limiting broken, project key mismatch, MFA QR text. | SSE auth, MFA QR, project key, and onboarding slug are fixed. API-key auth for `/ingest` and rate-limit wiring remain open. | Re-triage the P0/P1 list; move fixed items to resolved. | P1 |
| G11 | `docs/handover/backend-audit.md` §11 | Mentions SSE requires Bearer header without noting query-token support. | `Auth` middleware now supports `access_token` query param. | Add a one-line note. | P2 |
| G12 | `docs/handover/CURRENT_STATE.md` | Partially updated in this pass, but still has minor stale rows (e.g., settings pages status). | API keys and notifications settings pages are implemented; SSE auth is fixed. | Final consistency pass. | P2 |
| G13 | `docs/handover/FEATURE_MATRIX.md` | Partially updated; still marks some settings pages as placeholders and API-key UI as missing. | `/dashboard/settings/api-keys` is implemented. | Final consistency pass. | P2 |

## 3. Missing documentation

| # | Gap | Why it matters | Owner | Priority |
|---|-----|--------------|-------|----------|
| G14 | Incident response runbook | `TROUBLESHOOTING_GUIDE.md` covers triage but not declared incidents, severity, comms, and rollback decisions. | Platform / SRE | P1 |
| G15 | SDK usage guide and generated README | `packages/sdk/` is a placeholder; no guide for consumers of the public API. | API / Platform | P2 |
| G16 | API key CI/CD integration guide | The API key module exists and the settings UI is implemented, but no guide explains how CI pipelines send `X-API-Key` or `Authorization: ApiKey` once the middleware is wired. | Platform / Core | P2 |
| G17 | ClickHouse schema and operational guide | Deferred to Phase 5 but should exist before analytics implementation starts. | Data / Platform | P3 |
| G18 | Terraform / Kubernetes deployment runbooks | `DEPLOYMENT_GUIDE.md` covers stages conceptually; concrete runbooks for EKS/RDS/S3 provisioning are missing. | Infra / SRE | P3 |
| G19 | Frontend state architecture decision | `localStorage`-only state is documented as a limitation, but no ADR or design doc records the planned state layer (Zustand/Context/TanStack Query). | Frontend | P2 |
| G20 | SSE authentication hardening decision | Query-token auth for `EventSource` is a temporary MVP workaround; an ADR or runbook should record the long-term approach (signed SSE token, cookie, or `fetch` streaming). | Backend / Security | P2 |

## 4. Duplication and overlap

| # | Overlap | Documents affected | Recommended action | Priority |
|---|---------|--------------------|--------------------|----------|
| G21 | Product overview | `testra-master-context.md`, `testra-product-strategy.md`, `testra-product-architecture-strategy.md`, `docs/handover/PROJECT_OVERVIEW.md`, `docs/handover/ARCHITECTURE.md` | Root-level product docs are canonical; `handover/` overviews should reference them rather than duplicate. | P2 |
| G22 | Database documentation | `docs/architecture/DATABASE_DOCUMENTATION.md`, `docs/handover/DATABASE_OVERVIEW.md`, `docs/handover/migration-review.md` | Keep `DATABASE_DOCUMENTATION.md` canonical; ensure `DATABASE_OVERVIEW.md` and `migration-review.md` only supplement it. | P2 |
| G23 | Route/reference information | `docs/BIBLICAL_TESTRA.md`, `docs/handover/ROUTES.md`, `docs/api/openapi/openapi.yaml` | Keep OpenAPI canonical for contracts; BIBLICAL for rules; `ROUTES.md` for inventory. Update all three together when routes change. | P2 |
| G24 | Engineering handover reports | `docs/engineering/TESTRA_ENGINEERING_HANDOVER_REPORT.md`, `docs/engineering/HANDOVER_PHASE3_TO_PHASE4.md` | Archive the older report and keep `HANDOVER_PHASE3_TO_PHASE4.md` as the current Phase 3→4 source. | P1 |
| G25 | Documentation health reports | `docs/reports/DOCUMENTATION_HEALTH_REPORT.md`, `docs/reports/DOCUMENTATION_HEALTH_AUDIT.md` (this pass) | Archive `DOCUMENTATION_HEALTH_REPORT.md` and redirect to the new audit. | P1 |

## 5. Application README gaps

| # | README | Issue | Fix | Priority |
|---|--------|-------|-----|----------|
| G26 | `apps/api/README.md` | One-line stub; no link to onboarding, BIBLICAL, or OpenAPI. | Expand to include purpose, key commands, and links to canonical docs. | P2 |
| G27 | `apps/web/README.md` | One-line stub. | Expand or point to `docs/engineering/LOCAL_DEVELOPMENT_GUIDE.md` and frontend audit. | P2 |
| G28 | `apps/worker/README.md` | One-line stub; worker is not implemented. | State current stub status and link to roadmap. | P3 |
| G29 | `apps/ml/README.md` | One-line stub; ML service is a skeleton. | State skeleton status and Phase 6 plan. | P3 |

## 6. Automation and process gaps

| # | Gap | Problem | Recommended action | Priority |
|---|-----|---------|--------------------|----------|
| G30 | No CI check for OpenAPI validity | OpenAPI can drift from `server.go` without warning. | Add `swagger-codegen` or `redocly lint` to CI. | P2 |
| G31 | No internal link checker | Broken relative links between docs are possible. | Add `markdown-link-check` or `lychee` to CI. | P2 |
| G32 | No Mermaid diagram render check | Diagrams in docs may have syntax errors that are not caught until publication. | Add a Mermaid render test in docs CI. | P3 |
| G33 | No documentation update gate in PR template | Engineers may forget to update BIBLICAL/OpenAPI/PHASES when changing modules. | Add a PR checklist item and a docs-review bot. | P2 |

## 7. Summary of recommended moves

1. Create `docs/reports/archive/` and move the following files there (or delete if the history is already captured):
   - `04_Architecture/testra-software-architecture-decisions.md`
   - `docs/engineering/ENGINEERING_DOCUMENTATION_REPORT.md`
   - `docs/engineering/TESTRA_ENGINEERING_HANDOVER_REPORT.md`
   - `docs/reports/DOCUMENTATION_HEALTH_REPORT.md`
2. Refresh the `handover/` audit files (`frontend-audit.md`, `functional-audit.md`, `CURRENT_STATE.md`, `FEATURE_MATRIX.md`) to mark Phase 3.5 fixes as resolved.
3. Add small application READMEs.
4. Create the missing runbooks and guides per [`DOCUMENTATION_ROADMAP.md`](DOCUMENTATION_ROADMAP.md).
5. Add documentation validation to CI (OpenAPI, links, Mermaid).

## 8. Verification checklist

- [ ] No document claims a feature exists that is not in `PHASES.md` or code.
- [ ] All broken/outdated findings in `frontend-audit.md` and `functional-audit.md` are re-triaged.
- [ ] All superseded reports are archived or redirect to current canonical docs.
- [ ] `docs/README.md` (Documentation Index) remains the single source of truth for where each document belongs.
- [ ] Every new or modified route is reflected in `docs/api/openapi/openapi.yaml` and `docs/BIBLICAL_TESTRA.md`.
