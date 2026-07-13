# Testra API Design Guidelines

**Status:** Approved guidance
**Scope:** Public and internal HTTP APIs exposed by the Testra API service

## 1. Contract Authority

`docs/api/openapi/openapi.yaml` is the contract for implemented/documented endpoints. New endpoints must be documented before implementation and reviewed with the owning module.

The current contract is intentionally incomplete: MFA, password reset, RBAC, API keys, test management, execution, defects, analytics, and integrations are roadmap areas unless added to OpenAPI.

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
- CI/CD integrations use scoped, hashed API keys once the Phase 1 feature is implemented.
- Authentication is not authorization. Every tenant-scoped operation must resolve organization/workspace/project scope and check permissions.
- Do not infer tenant scope solely from client-provided IDs; verify membership and relationship server-side.

## 8. Idempotency and Retries

Mutating endpoints with external side effects, ingestion, exports, webhooks, or payment-like operations require an `Idempotency-Key`. The key is scoped by tenant, principal, and operation and is stored in PostgreSQL for 24 hours with a request fingerprint and final response. Reuse with a different fingerprint is rejected. Pure reads and deterministic metadata updates do not require it.

Clients should retry only documented transient failures (`408`, `429`, selected `5xx`) with exponential backoff and jitter.

## 9. Documentation and Review

Every operation must include summary, security, parameters, request/response schemas, error responses, and examples where behavior is non-obvious. Contract changes require compatibility review, OpenAPI validation, tests, and a release-note entry.
