# Testra Entity Relationship Diagram

## Status

This is a documentation-level model. Solid entities are recorded as current in project progress; planned entities are approved roadmap items and must be reconciled with migrations before implementation.

## Core ERD

```mermaid
erDiagram
    USER ||--o{ ORGANIZATION_MEMBER : belongs_to
    ORGANIZATION ||--o{ ORGANIZATION_MEMBER : contains
    ORGANIZATION ||--o{ WORKSPACE : owns
    WORKSPACE ||--o{ PROJECT : contains

    ORGANIZATION ||--o{ ROLE : defines
    ROLE ||--o{ ROLE_PERMISSION : grants
    PERMISSION ||--o{ ROLE_PERMISSION : included_in
    USER ||--o{ ROLE_ASSIGNMENT : receives
    ROLE ||--o{ ROLE_ASSIGNMENT : assigned
    ORGANIZATION ||--o{ API_KEY : scopes
    WORKSPACE o|--o{ API_KEY : narrows
    PROJECT o|--o{ API_KEY : narrows

    USER {
        uuid id PK
        string email
        string name
        timestamp created_at
    }
    ORGANIZATION {
        uuid id PK
        string name
        string slug UK
        timestamp created_at
    }
    ORGANIZATION_MEMBER {
        uuid organization_id FK
        uuid user_id FK
        string status
    }
    WORKSPACE {
        uuid id PK
        uuid organization_id FK
        string name
        string slug
    }
    PROJECT {
        uuid id PK
        uuid workspace_id FK
        string name
        string key
    }
    ROLE {
        uuid id PK
        uuid organization_id FK
        string name
    }
    PERMISSION {
        string code PK
        string description
    }
    ROLE_PERMISSION {
        uuid role_id FK
        string permission_code FK
    }
    ROLE_ASSIGNMENT {
        uuid user_id FK
        uuid role_id FK
        uuid organization_id FK
    }
    API_KEY {
        uuid id PK
        uuid organization_id FK
        string key_hash
        timestamp revoked_at
    }
```

## Relationship Rules

- A user may belong to multiple organizations.
- An organization owns workspaces; a workspace owns projects.
- Membership is represented by an organization membership relationship keyed by organization and user; membership is the prerequisite for all tenant access.
- Roles and assignments are organization-scoped. Permission narrowing is enforced by resource scope in the service layer.
- API keys are stored as hashes and may be narrowed to workspace/project scope; plaintext is displayed only at creation.
- PostgreSQL RLS and request-scoped tenant propagation are mandatory as defined in ADR-004.

## Planned Domain Extensions

Phase 2 adds test cases, suites, folders, and immutable audit events. Phase 3 adds runs and analytical results. Phase 4 adds defects, notifications, and integration records. Those entities are intentionally omitted from the current ERD until ownership and retention rules are approved.

## Review Requirement

Any schema change must update this document, the relevant migration plan, and OpenAPI when the change affects an API representation. The diagram is not a substitute for migration SQL.
