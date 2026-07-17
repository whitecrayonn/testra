# Testra Documentation Index

**Purpose:** One-stop index of every canonical and archived document in `testra/docs/`.

**Owner:** Documentation Architect / Engineering Lead

**Scope:** Lists canonical entry points, product documents, archive locations, and how to find common information. Does not contain detailed architecture or coding rules.

**Source of truth:** This file for documentation navigation; architectural and engineering truth lives in `BIBLICAL_TESTRA.md` and the canonical docs listed here.

**Last updated:** July 2026

> **One source of truth per concern.** This index maps canonical documentation and the archive for historical or superseded material. For the consolidation rationale and change list, see [`archive/superseded/DOCUMENTATION_CONSOLIDATION_REPORT.md`](archive/superseded/DOCUMENTATION_CONSOLIDATION_REPORT.md).

## Status legend

- **Canonical** - current source of truth for the topic.
- **Archive** - historical or superseded; kept for reference but not current.

## Canonical entry points

| Document | Purpose |
|----------|---------|
| [`BIBLICAL_TESTRA.md`](BIBLICAL_TESTRA.md) | Consolidated engineering handbook: architecture, rules, do-not-break list, canonical sources |
| [`AI_CONTEXT.md`](AI_CONTEXT.md) | AI entry point: reading order, verification workflow, forbidden actions, canonical ownership |
| [`AI_MEMORY.md`](AI_MEMORY.md) | Permanent architectural facts for AI agents |
| [`AI_RULES.md`](AI_RULES.md) | Change-impact matrix and AI-specific documentation rules |
| [`PROJECT_OVERVIEW.md`](PROJECT_OVERVIEW.md) | Product vision, goals, target users, current MVP scope, and repository status |
| [`ROADMAP.md`](engineering/ROADMAP.md) | Implementation phases, engineering priorities, technical debt, documentation roadmap |
| [`ONBOARDING.md`](engineering/ONBOARDING.md) | New engineer onboarding, development workflow, governance, DoD, and self-review |
| [`ENGINEERING_STANDARDS.md`](engineering/ENGINEERING_STANDARDS.md) | Coding standards for Go, TypeScript, API, DB, security |
| [`API_DESIGN_GUIDELINES.md`](api/API_DESIGN_GUIDELINES.md) | REST conventions, versioning, OpenAPI maintenance, response envelopes |
| [`DATABASE_GUIDE.md`](architecture/DATABASE_GUIDE.md) | Migration catalog, schema, RLS, ERD, and storage responsibilities |
| [`SYSTEM_FLOWS.md`](architecture/SYSTEM_FLOWS.md) | Platform, request, trust-boundary, and sequence diagrams |
| [`MODULE_DEPENDENCIES.md`](architecture/MODULE_DEPENDENCIES.md) | Module ownership and dependency matrix |
| [`ROUTES.md`](ROUTES.md) | Frontend and backend route inventory |
| [`FEATURE_MATRIX.md`](FEATURE_MATRIX.md) | Feature completion and functional audit matrix |
| [`DEPLOYMENT_GUIDE.md`](deployment/DEPLOYMENT_GUIDE.md) | Environment promotion, MVP deployment, and infrastructure findings |
| [`architecture/adrs/`](architecture/adrs/) | Accepted Architecture Decision Records (ADR-001 through ADR-012) |
| [`operations/DISASTER_RECOVERY_GUIDE.md`](operations/DISASTER_RECOVERY_GUIDE.md) | Backup, restore, RPO/RTO |
| [`operations/MONITORING_LOGGING_GUIDE.md`](operations/MONITORING_LOGGING_GUIDE.md) | Observability requirements |
| [`operations/PRODUCTION_READINESS_CHECKLIST.md`](operations/PRODUCTION_READINESS_CHECKLIST.md) | Production go-live gate |
| [`operations/TROUBLESHOOTING_GUIDE.md`](operations/TROUBLESHOOTING_GUIDE.md) | Symptom-based triage |
| [`release/RELEASE_CHECKLIST.md`](release/RELEASE_CHECKLIST.md) | Release execution checklist |
| [`security/SECURITY_CHECKLIST.md`](security/SECURITY_CHECKLIST.md) | Security review checklist |
| [`api/openapi/openapi.yaml`](api/openapi/openapi.yaml) | Authoritative HTTP contract |
| [`apps/api/migrations/`](../apps/api/migrations/) | Authoritative database schema |

## Product & strategy (root-level)

| Document | Purpose | Notes |
|----------|---------|-------|
| `../testra-master-context.md` | Vision, mission, ideal customer profile | Root-level product context |
| `../testra-product-strategy.md` | North Star metrics, principles, release sequencing | Root-level product strategy |
| `../testra-product-architecture-strategy.md` | Product architecture, domain decomposition | Root-level product architecture |
| `../testra-product-discovery.md` | Problem statement, market opportunity, USP | Root-level discovery |
| `../testra-brd.md` | Business Requirements Document | Root-level business case |
| `../04_Architecture/testra-software-architecture-decisions.md` | Pre-implementation draft | Conflicts with accepted ADRs; archive candidate |

## Archive

Historical, superseded, and merged-source documents are in [`archive/`](archive/):

- `archive/historical/` - progress reports, phase gates, reviews, handover records, UX reports.
- `archive/superseded/` - previous versions of canonical docs and stale summary reports.
- `archive/merged-sources/` - source documents whose content was merged into canonical files.

## Where to find what

| I want to know ... | Start here | Fallback |
|--------------------|------------|----------|
| What does Testra do? | `../testra-master-context.md` | [`PROJECT_OVERVIEW.md`](PROJECT_OVERVIEW.md) |
| What is implemented right now? | [`PROJECT_OVERVIEW.md`](PROJECT_OVERVIEW.md) > Current State | [`FEATURE_MATRIX.md`](FEATURE_MATRIX.md) |
| What is the engineering plan? | [`ROADMAP.md`](engineering/ROADMAP.md) | [`BIBLICAL_TESTRA.md`](BIBLICAL_TESTRA.md) |
| How do I onboard? | [`ONBOARDING.md`](engineering/ONBOARDING.md) | [`BIBLICAL_TESTRA.md`](BIBLICAL_TESTRA.md) |
| What HTTP endpoints exist? | [`api/openapi/openapi.yaml`](api/openapi/openapi.yaml) | [`ROUTES.md`](ROUTES.md) |
| What is the database schema? | [`apps/api/migrations/`](../apps/api/migrations/) | [`DATABASE_GUIDE.md`](architecture/DATABASE_GUIDE.md) |
| How is tenant isolation enforced? | [`architecture/adrs/ADR-004-tenant-isolation-strategy.md`](architecture/adrs/ADR-004-tenant-isolation-strategy.md) | [`BIBLICAL_TESTRA.md`](BIBLICAL_TESTRA.md) > Multi-tenancy |
| How do I build locally? | `../README.md` | [`ONBOARDING.md`](engineering/ONBOARDING.md) > Local Development |
| What are the coding rules? | [`ENGINEERING_STANDARDS.md`](engineering/ENGINEERING_STANDARDS.md) | [`BIBLICAL_TESTRA.md`](BIBLICAL_TESTRA.md) > Engineering Rules |
| What must never break? | [`BIBLICAL_TESTRA.md`](BIBLICAL_TESTRA.md) > Do Not Break List | [`security/SECURITY_CHECKLIST.md`](security/SECURITY_CHECKLIST.md) |
| What are the open risks / debt? | [`ROADMAP.md`](engineering/ROADMAP.md) > Technical Debt Register | [`ROADMAP.md`](engineering/ROADMAP.md) |

## Findings summary

- The canonical engineering handbook is [`BIBLICAL_TESTRA.md`](BIBLICAL_TESTRA.md).
- Root-level product docs (`testra-*.md`) are the canonical product sources.
- `04_Architecture/testra-software-architecture-decisions.md` is a pre-implementation draft; refer to the accepted ADRs instead.
- Superseded handover and audit documents are archived under [`archive/`](archive/).
- This index, `BIBLICAL_TESTRA.md`, and `ROADMAP.md` are the primary navigation surfaces.
- QA and release reports live in [`reports/`](reports/); the older consolidation and final audit reports are archived under [`archive/superseded/`](archive/superseded/).

## See Also

- [`BIBLICAL_TESTRA.md`](BIBLICAL_TESTRA.md) — canonical engineering handbook
- [`AI_CONTEXT.md`](AI_CONTEXT.md) — AI entry point and verification workflow
- [`ROADMAP.md`](engineering/ROADMAP.md) — implementation phases and technical debt
