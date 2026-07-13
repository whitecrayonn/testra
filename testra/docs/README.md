# Testra Documentation

## Purpose

This directory is the documentation source of truth for Testra architecture, API contracts, engineering workflows, operations, and security.

## Documentation Map

- `api/openapi/openapi.yaml` — OpenAPI 3.1 contract for currently documented API surface.
- `api/openapi/README.md` — OpenAPI ownership, validation, and publication rules.
- `api/API_DESIGN_GUIDELINES.md` — REST, resource, response, error, and security conventions.
- `api/API_VERSIONING_GUIDE.md` — API compatibility and release policy.
- `architecture/DATABASE_DOCUMENTATION.md` — storage responsibilities, tenancy, migration, and integrity rules.
- `architecture/ERD.md` — current and planned entity relationships.
- `architecture/SEQUENCE_DIAGRAMS.md` — approved and planned request/data flows.
- `architecture/SYSTEM_FLOWS.md` — platform-level flows and trust boundaries.
- `architecture/MODULE_DEPENDENCIES.md` — modular-monolith ownership and dependency rules.
- `architecture/adrs/` — accepted architectural decisions, including ADR-001 through ADR-009.
- `engineering/DEVELOPER_ONBOARDING.md` — onboarding path for engineers.
- `engineering/LOCAL_DEVELOPMENT_GUIDE.md` — local setup and daily workflow.
- `deployment/DEPLOYMENT_GUIDE.md` — environment promotion and deployment procedure.
- `operations/TROUBLESHOOTING_GUIDE.md` — symptom-based diagnosis.
- `operations/MONITORING_LOGGING_GUIDE.md` — observability requirements and response.
- `operations/DISASTER_RECOVERY_GUIDE.md` — backup, restore, and recovery objectives.
- `operations/PRODUCTION_READINESS_CHECKLIST.md` — go-live gate.
- `release/RELEASE_CHECKLIST.md` — release execution checklist.
- `security/SECURITY_CHECKLIST.md` — security review checklist.
- `engineering/ENGINEERING_DOCUMENTATION_REPORT.md` — documentation completion report.

## Status Vocabulary

- **Implemented** — supported by the current repository or explicitly recorded progress.
- **Approved** — architectural direction accepted but implementation may be pending.
- **Planned** — roadmap item not yet implemented.
- **Assumption** — non-architectural implementation detail that must be validated during delivery; it cannot override an accepted ADR.

Documentation must not be interpreted as evidence that a planned feature exists in code. Implementation status remains governed by `docs/engineering/PHASES.md` and progress reports. No architectural decision is pending in the current ADR set (ADR-001 through ADR-009).
