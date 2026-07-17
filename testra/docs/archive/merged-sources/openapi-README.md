# OpenAPI Documentation

## Source of Truth

`openapi.yaml` is the versioned OpenAPI 3.1 contract for Testra's documented REST API. It currently covers authentication, organizations, workspaces, projects, API keys, test management (folders, suites, cases, versions), test runs and run items, CI ingestion, notifications, notification preferences, and notification channels. The version is `0.4.0` as of Phase 3.

## Update Rules

- Update the contract before implementing a new endpoint.
- Keep operation IDs unique and stable.
- Define request and response schemas explicitly.
- Include successful and expected failure responses.
- Add security requirements to every protected operation.
- Keep server URLs and API version aligned with the deployed gateway.
- Do not document an endpoint as implemented merely because it is planned in `PHASES.md`.

## Review Checklist

- [ ] Path and method follow `API_DESIGN_GUIDELINES.md`.
- [ ] Tenant and authorization behavior is described.
- [ ] Request validation is represented in schemas.
- [ ] Response envelope and error schema are consistent.
- [ ] No secret, source code, or sensitive payload is returned.
- [ ] Compatibility impact is classified using `API_VERSIONING_GUIDE.md`.
- [ ] Contract validation and endpoint tests pass.

## Finalized API Constraints

- Bearer JWT remains the public session mechanism under ADR-001 and ADR-007.
- New list endpoints use cursor pagination with default page size 50 and maximum 100.
- Side-effecting commands, ingestion, exports, webhooks, and payment-like operations require `Idempotency-Key` with a 24-hour PostgreSQL replay record.
- URL major versioning uses `/api/v1`; deprecation support is at least 12 months or two major release cycles, whichever is longer.
- Interactive documentation publication is a delivery concern; it must not be exposed as an undocumented production endpoint.
