# Documentation Roadmap

> This roadmap schedules the documentation improvements identified in [`DOCUMENTATION_GAP_REPORT.md`](DOCUMENTATION_GAP_REPORT.md). It is aligned with the engineering phases in `docs/engineering/PHASES.md`.

## Status legend

- **Immediate:** Do within the current documentation audit pass (days).
- **Short-term:** Before or during the first sprint of Phase 4.
- **Medium-term:** During Phase 4 execution (API testing, defects, integrations).
- **Long-term:** Phase 5+ (analytics, launch, enterprise features).

## Immediate (this pass)

| # | Deliverable | Why | Success criteria | Owner |
|---|-------------|-----|------------------|-------|
| 1 | `docs/README.md` (Documentation Index) | One canonical map of every document and its status. | All major docs listed with status and cross-references. | ✅ Done |
| 2 | `docs/reports/DOCUMENTATION_HEALTH_AUDIT.md` | Score every doc so stakeholders know what to trust. | 38+ docs scored with category averages and top issues. | ✅ Done |
| 3 | `docs/reports/DOCUMENTATION_GAP_REPORT.md` | Enumerate exactly what is stale, missing, duplicated, or conflicting. | Gaps numbered, prioritized, with remediation actions. | ✅ Done |
| 4 | `docs/reports/DOCUMENTATION_ROADMAP.md` | This file. | Time-bound plan linked to engineering phases. | ✅ Done |
| 5 | Update `docs/BIBLICAL_TESTRA.md` | Keep the engineering handbook current with Phase 3 routes and auth. | Notification routes added, SSE query-token auth noted, doc index/audit reports referenced. | ✅ Done |
| 6 | Update `testra/README.md` | Point new users to the documentation index and handbook. | Doc section added at top of root README. | ✅ Done |
| 7 | Refresh `docs/handover/CURRENT_STATE.md` and `FEATURE_MATRIX.md` | Reflect Phase 3.5 completions (API keys UI, SSE auth, MFA QR, project key, onboarding slug). | Stale rows updated; fixed issues no longer marked as broken. | ✅ Done |
| 8 | Update `docs/handover/ROUTES.md` | SSE auth caveat is outdated. | Caveat now states query-token auth is supported as an MVP workaround. | ✅ Done |
| 9 | Refresh `docs/handover/frontend-audit.md` | Lists fixed issues (MFA QR, project key, SSE auth, onboarding slug) as current blockers. | Fixed items moved to a resolved section. | Pending |
| 10 | Refresh `docs/handover/functional-audit.md` | P0 broken list is stale after Phase 3.5. | Re-triage and mark resolved issues. | Pending |
| 11 | Add stale headers to superseded reports | Prevent engineers from using outdated reports. | `ENGINEERING_DOCUMENTATION_REPORT.md`, `TESTRA_ENGINEERING_HANDOVER_REPORT.md`, `DOCUMENTATION_HEALTH_REPORT.md` clearly marked. | Pending |
| 12 | Expand app READMEs | One-line stubs in `apps/*` do not orient engineers. | `api`, `web`, `worker`, `ml` READMEs link to canonical docs and state current status. | Pending |

## Short-term (Phase 4 kickoff)

| # | Deliverable | Why | Success criteria | Owner |
|---|-------------|-----|------------------|-------|
| 13 | `docs/operations/INCIDENT_RESPONSE_RUNBOOK.md` | `TROUBLESHOOTING_GUIDE.md` covers triage but not declared incidents. | Document severity levels, escalation, comms, rollback, and post-mortem procedures. | Platform / SRE |
| 14 | Archive or redirect superseded documents | Reduce confusion and duplicate sources. | `04_Architecture/testra-software-architecture-decisions.md`, `ENGINEERING_DOCUMENTATION_REPORT.md`, `TESTRA_ENGINEERING_HANDOVER_REPORT.md`, `DOCUMENTATION_HEALTH_REPORT.md` moved to `docs/reports/archive/` or contain redirects. | Docs / Platform |
| 15 | Final pass on `handover/` audit files | Ensure `CURRENT_STATE.md`, `FEATURE_MATRIX.md`, `frontend-audit.md`, `functional-audit.md`, `backend-audit.md` are consistent with Phase 3.5. | No fixed issue listed as broken; all statuses match code/OpenAPI. | Docs / Engineering |
| 16 | Document Phase 4 scope in `docs/engineering/PHASES.md` and BIBLICAL | As Phase 4 starts, update canonical sources. | New modules (defects, API testing, integrations) and their dependencies recorded. | Engineering Lead |
| 17 | Expand API key CI/CD guide | API keys exist but no integration guide for CI. | Document `X-API-Key` or `Authorization: ApiKey` usage for `/ingest` once the middleware is wired. | API / Platform |

## Medium-term (Phase 4 execution)

| # | Deliverable | Why | Success criteria | Owner |
|---|-------------|-----|------------------|-------|
| 18 | Update OpenAPI for Phase 4 endpoints | OpenAPI is the source of truth for API behavior. | Defects, API testing, integrations, webhooks added as they are implemented. | API / Backend |
| 19 | Update `docs/architecture/ERD.md` and BIBLICAL data model | New Phase 4 entities (defects, test plans, API test cases). | ERD includes new entities; BIBLICAL schema groups table updated. | Data / Backend |
| 20 | Create module READMEs in `apps/api/internal/<module>/` | Each module should document its ports and dependencies. | New modules have `README.md` explaining domain, service, and handler entry points. | Backend Engineers |
| 21 | Frontend state architecture ADR | `localStorage`-only state is a known limitation; a decision should be recorded. | ADR compares Zustand / React Context / TanStack Query and picks one. | Frontend Lead |
| 22 | SSE authentication hardening ADR | Query-token auth for `EventSource` is an MVP workaround. | ADR records the long-term approach (signed SSE token, cookie, or `fetch` streaming). | Backend / Security |
| 23 | Document CI/CD integration plugins | Phase 4 adds GitHub Actions / GitLab / Jenkins plugins. | Runbook for each plugin with examples. | Integrations |

## Long-term (Phase 5 and beyond)

| # | Deliverable | Why | Success criteria | Owner |
|---|-------------|-----|------------------|-------|
| 24 | ClickHouse schema and analytics runbook | Analytics and dashboards require ClickHouse. | Schema, ingestion pipeline, query patterns, and operational guide documented before launch. | Data / Analytics |
| 25 | Terraform / Kubernetes deployment runbooks | Production deployment needs concrete manifests. | Step-by-step runbooks for EKS/RDS/S3/Redis provisioning and deployment. | Infra / SRE |
| 26 | Generated TypeScript SDK + README | Public API and partner integrations depend on SDK. | `packages/sdk/` is generated from OpenAPI and has usage examples. | API / Platform |
| 27 | Security and compliance runbooks | SOC 2 readiness requires documented evidence. | Penetration test, vulnerability management, access review, and audit export runbooks. | Security |
| 28 | Enterprise features docs | SSO/SAML, SCIM, data residency require dedicated guides. | `docs/enterprise/` with setup and configuration guides. | Enterprise |

## Process improvements

| # | Improvement | Why | Target | Owner |
|---|-------------|-----|--------|-------|
| 29 | OpenAPI validation in CI | Prevent drift between code and contract. | CI job runs `redocly lint` or `swagger-codegen validate` on every PR. | Platform |
| 30 | Markdown internal link checker | Broken relative links degrade discoverability. | CI job runs `markdown-link-check` or `lychee` on every PR. | Platform |
| 31 | Mermaid render test | Diagram syntax errors are not caught by text review. | CI job renders all `.md` Mermaid diagrams. | Platform |
| 32 | Documentation update gate in PR template | Engineers often forget to update docs. | PR template asks "Which docs were updated?" and links to `docs/README.md`. | Engineering |
| 33 | Quarterly doc health review | Docs decay as code changes. | Re-run `DOCUMENTATION_HEALTH_AUDIT.md` each quarter and update `DOCUMENTATION_GAP_REPORT.md`. | Docs / Engineering Lead |

## Definition of done for documentation

Documentation is "done" for a feature when:

1. OpenAPI paths and schemas are updated (`docs/api/openapi/openapi.yaml`).
2. BIBLICAL route table and data model sections are updated.
3. `docs/engineering/PHASES.md` is updated if the feature changes phase status.
4. `docs/README.md` (Documentation Index) is updated if new docs are added.
5. `docs/handover/TECHNICAL_DEBT.md` and `CURRENT_STATE.md` are updated if the change resolves or creates debt.
6. A new ADR is created if the change affects an accepted architectural decision.
7. `docs/reports/DOCUMENTATION_HEALTH_AUDIT.md` is re-scored if a doc materially changes.

## Links

- [`DOCUMENTATION_GAP_REPORT.md`](DOCUMENTATION_GAP_REPORT.md) — detailed gaps and remediation items.
- [`DOCUMENTATION_HEALTH_AUDIT.md`](DOCUMENTATION_HEALTH_AUDIT.md) — current scores and top issues.
- [`docs/README.md`](../README.md) — canonical documentation index.
- [`docs/BIBLICAL_TESTRA.md`](../BIBLICAL_TESTRA.md) — engineering handbook.
- [`docs/engineering/PHASES.md`](../engineering/PHASES.md) — implementation phase status.
