# Testra API Design Guidelines

**Purpose:** Define REST API conventions: versioning, pagination, idempotency, response envelopes, security, and OpenAPI maintenance.
**Owner:** API Architect / Engineering Lead
**Status:** Approved guidance
**Scope:** Public and internal HTTP APIs exposed by the Testra API service
**Source of Truth:** API_DESIGN_GUIDELINES.md for conventions; `docs/api/openapi/openapi.yaml` for actual contracts.
**Last Updated:** July 2026
**Related documents:**
- [`BIBLICAL_TESTRA.md`](../BIBLICAL_TESTRA.md)
- [`ROUTES.md`](../ROUTES.md)
- [`docs/api/openapi/openapi.yaml`](openapi/openapi.yaml)
- [`ADR-006-api-standards.md`](../architecture/adrs/ADR-006-api-standards.md)

## 1. Contract Authority

`docs/api/openapi/openapi.yaml` is the contract for implemented/documented endpoints. New endpoints must be documented before implementation and reviewed with the owning module.

As of Phase 3, the contract includes authentication (register, login, MFA, password reset, refresh), organizations, workspaces, projects, API keys, RBAC-authorized resources, test management (folders, suites, cases, versions), test runs, CI ingestion, and notifications. Defects, analytics, integrations, and future modules remain roadmap areas until added to OpenAPI.

## 2. Resource Design

- Use plural nouns: `/organizations`, `/workspaces`, `/projects`.
- Nest only when the parent is required to identify scope; otherwise use a filter such as `workspace_id`.
- Use UUID path identifiers.
- Keep endpoint behavior aligned with one domain module.
- Prefer explicit actions only for non-CRUD state transitions, such as `/runs/{id}/cancel`.
- Do not expose database table names, internal package names, or implementation-specific fields.

## 3. HTTP Semantics

- `GET` reads and is safe; `POST` creates or triggers a command; `PATCH` partially updates; `DELETE` removes or revokes.
- `201 Created` is used for successful creation.
- `200 OK` is used for successful reads and commands returning a representation.
- `204 No Content` is used when no response representation is needed.
- `400` means malformed or invalid request syntax; `401` means missing/invalid authentication; `403` means authenticated but unauthorized; `404` means the resource is not visible or does not exist; `409` means a state or uniqueness conflict; `422` is reserved for semantically invalid input where distinguished from `400`.
- `429` must be used for rate limiting; `5xx` responses must not disclose internals.

## 4. Response Envelope

Successful responses follow:

```json
{ "data": {}, "meta": {}, "error": null }
```

Errors follow:

```json
{ "data": null, "meta": {}, "error": { "code": "NOT_FOUND", "message": "resource not found" } }
```

Error codes are stable machine-readable identifiers. Messages are safe for clients and must not contain secrets, SQL, stack traces, or tenant data from another request.

## 5. Pagination, Filtering, and Sorting

All new list endpoints use cursor pagination with `cursor` and `page_size`. The default page size is 50 and the maximum is 100. Ordering must be stable and include a unique tie-breaker. Existing low-volume endpoints may remain unpaginated until their next contract revision. Responses expose pagination state in `meta`, never in an undocumented top-level field.

## 6. Input and Output Rules

- Validate request bodies, query parameters, and path parameters at the boundary.
- Use `snake_case` JSON fields.
- Use RFC 3339 UTC timestamps.
- Do not return password hashes, raw API keys, MFA secrets, access tokens beyond their intended issuance response, or internal authorization metadata.
- Define required fields, formats, length limits, patterns, and nullable behavior in OpenAPI.
- Reject unknown request fields by default; exceptions require an API owner decision recorded in the endpoint contract.

## 7. Authentication and Authorization

- Browser/user sessions use bearer JWT as recorded in ADR-001.
- Scoped, hashed API keys are implemented for workspace management; the ingestion endpoint is still authenticated by JWT and API-key authentication is the next integration step.
- Authentication is not authorization. Every tenant-scoped operation must resolve organization/workspace/project scope and check permissions.
- Do not infer tenant scope solely from client-provided IDs; verify membership and relationship server-side.

## 8. Idempotency and Retries

Mutating endpoints with external side effects, ingestion, exports, webhooks, or payment-like operations require an `Idempotency-Key`. The key is scoped by tenant, principal, and operation and is stored in PostgreSQL for 24 hours with a request fingerprint and final response. Reuse with a different fingerprint is rejected. Pure reads and deterministic metadata updates do not require it.

Clients should retry only documented transient failures (`408`, `429`, selected `5xx`) with exponential backoff and jitter.

## 9. Documentation and Review

Every operation must include summary, security, parameters, request/response schemas, error responses, and examples where behavior is non-obvious. Contract changes require compatibility review, OpenAPI validation, tests, and a release-note entry.


## API Versioning

### Policy

Testra uses URL-based major API versions: `/api/v1`. The OpenAPI server URLs and route prefixes must agree. A version identifies a compatibility boundary, not an implementation deployment.

### Compatible Changes

The following are normally backward compatible in an existing major version:

- Adding an optional request field.
- Adding a response field that clients must ignore.
- Adding a new endpoint or resource.
- Adding a documented error response for a previously unspecified failure, provided clients already handle non-2xx responses.

The following require a new major version or an explicitly approved migration plan:

- Removing or renaming an endpoint, field, enum value, or error code.
- Changing a field type, requiredness, meaning, or identifier format.
- Changing authentication semantics or tenant visibility.
- Changing pagination or ordering guarantees in a breaking way.

### Deprecation

- Mark deprecated operations and fields in OpenAPI with `deprecated: true`.
- Document replacement, migration steps, and removal target.
- Emit a response header such as `Deprecation` only when standardized by the API owner.
- Preserve deprecated behavior for a minimum of 12 months or two major release cycles, whichever is longer.

### Change Workflow

1. Identify consumers: web, SDK, integrations, CI jobs, and external users.
2. Update OpenAPI and this change assessment before implementation.
3. Classify the change as compatible, deprecated, or breaking.
4. Add contract and regression tests.
5. Publish migration notes and update SDK generation inputs.
6. Roll out server changes before clients that depend on them.
7. Remove deprecated behavior only after the approved window and release communication.

### Versioning Responsibilities

The API owner maintains the OpenAPI contract. Module owners own endpoint semantics. Release owners verify compatibility and changelog coverage. Security-sensitive changes require security review even when technically backward compatible.

### Current State

The repository currently documents `v1` endpoints for authentication, organizations, workspaces, projects, API keys, test management, test runs, ingestion, and notifications. Future endpoints must extend `v1` only if compatible with the documented conventions; otherwise an ADR is required.


## OpenAPI Maintenance

### Source of Truth

`openapi.yaml` is the versioned OpenAPI 3.1 contract for Testra's documented REST API. It currently covers authentication, organizations, workspaces, projects, API keys, test management (folders, suites, cases, versions), test runs and run items, CI ingestion, notifications, notification preferences, and notification channels. The version is `0.4.0` as of Phase 3.

### Update Rules

- Update the contract before implementing a new endpoint.
- Keep operation IDs unique and stable.
- Define request and response schemas explicitly.
- Include successful and expected failure responses.
- Add security requirements to every protected operation.
- Keep server URLs and API version aligned with the deployed gateway.
- Do not document an endpoint as implemented merely because it is planned in `ROADMAP.md`.

### Review Checklist

- [ ] Path and method follow `API_DESIGN_GUIDELINES.md`.
- [ ] Tenant and authorization behavior is described.
- [ ] Request validation is represented in schemas.
- [ ] Response envelope and error schema are consistent.
- [ ] No secret, source code, or sensitive payload is returned.
- [ ] Compatibility impact is classified using the API Versioning section in this document.
- [ ] Contract validation and endpoint tests pass.

### Finalized API Constraints

- Bearer JWT remains the public session mechanism under ADR-001 and ADR-007.
- New list endpoints use cursor pagination with default page size 50 and maximum 100.
- Side-effecting commands, ingestion, exports, webhooks, and payment-like operations require `Idempotency-Key` with a 24-hour PostgreSQL replay record.
- URL major versioning uses `/api/v1`; deprecation support is at least 12 months or two major release cycles, whichever is longer.
- Interactive documentation publication is a delivery concern; it must not be exposed as an undocumented production endpoint.

---

## See Also

- [`BIBLICAL_TESTRA.md`](../BIBLICAL_TESTRA.md) — engineering handbook and canonical sources map.
- [`ROUTES.md`](../ROUTES.md) — frontend and backend route inventory.
- [`docs/api/openapi/openapi.yaml`](openapi/openapi.yaml) — the authoritative HTTP contract.
- [`ENGINEERING_STANDARDS.md`](../engineering/ENGINEERING_STANDARDS.md) — coding, security, and review standards.
- Accepted ADRs in [`docs/architecture/adrs/`](../architecture/adrs/) — especially ADR-001, ADR-006, ADR-007, and ADR-012.
