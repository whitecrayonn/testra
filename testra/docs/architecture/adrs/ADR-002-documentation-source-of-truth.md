# ADR-002: Documentation Source-of-Truth Boundaries

**Status:** Accepted
**Date:** July 2026

## Context

Testra has product, architecture, engineering, API, operational, and progress documents. Future implementation work needs a predictable way to distinguish implemented behavior, approved architecture, planned work, and assumptions without treating diagrams or guides as proof of code behavior.

## Decision

Adopt the following documentation boundaries:

- OpenAPI is authoritative for documented HTTP contracts.
- Migrations and implementation are authoritative for actual schema/runtime behavior.
- `ROADMAP.md` and progress reports are authoritative for implementation status.
- ADRs are authoritative for accepted architectural decisions and deviations.
- Architecture diagrams describe approved logical relationships and must label planned/assumed elements.
- Operational guides define required controls, not evidence that controls are deployed.

Use the status vocabulary **Implemented**, **Approved**, **Planned**, and **Assumption** in new documentation.

## Consequences

- Documentation can support parallel engineering without claiming nonexistent implementation.
- Engineers must reconcile docs with code and migrations at phase boundaries.
- Future implementation work may still be planned, but finalized architectural constraints are recorded in accepted ADRs.
- Automated contract/link/render checks remain recommended engineering controls.
