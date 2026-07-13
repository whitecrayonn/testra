# Testra API Versioning Guide

## Policy

Testra uses URL-based major API versions: `/api/v1`. The OpenAPI server URLs and route prefixes must agree. A version identifies a compatibility boundary, not an implementation deployment.

## Compatible Changes

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

## Deprecation

- Mark deprecated operations and fields in OpenAPI with `deprecated: true`.
- Document replacement, migration steps, and removal target.
- Emit a response header such as `Deprecation` only when standardized by the API owner.
- Preserve deprecated behavior for a minimum of 12 months or two major release cycles, whichever is longer.

## Change Workflow

1. Identify consumers: web, SDK, integrations, CI jobs, and external users.
2. Update OpenAPI and this change assessment before implementation.
3. Classify the change as compatible, deprecated, or breaking.
4. Add contract and regression tests.
5. Publish migration notes and update SDK generation inputs.
6. Roll out server changes before clients that depend on them.
7. Remove deprecated behavior only after the approved window and release communication.

## Versioning Responsibilities

The API owner maintains the OpenAPI contract. Module owners own endpoint semantics. Release owners verify compatibility and changelog coverage. Security-sensitive changes require security review even when technically backward compatible.

## Current State

The repository currently documents `v1` endpoints for authentication, organizations, workspaces, and projects. Future Phase 1 endpoints must extend `v1` only if compatible with the documented conventions; otherwise an ADR is required.
