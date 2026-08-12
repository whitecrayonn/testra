# Testra Sequence Diagrams

These diagrams describe approved/current boundaries and planned flows. A planned diagram is not evidence of implementation.

## Registration and Login — current contract

```mermaid
sequenceDiagram
    actor Client
    participant API as Go API
    participant Identity as Identity module
    participant DB as PostgreSQL
    Client->>API: POST /api/v1/auth/register or /login
    API->>Identity: Validate and execute use case
    Identity->>DB: Read/write user record
    DB-->>Identity: User or conflict
    Identity-->>API: Auth result or domain error
    API-->>Client: Envelope with token or safe error
```

## Tenant-Scoped Project Read — current contract

```mermaid
sequenceDiagram
    actor Client
    participant API as HTTP middleware/API
    participant Auth as Auth middleware
    participant Project as Project module
    participant DB as PostgreSQL
    Client->>API: GET /api/v1/projects/{id} + bearer token
    API->>Auth: Parse and validate JWT
    Auth-->>API: Authenticated principal
    API->>Project: Get project with principal context
    Project->>DB: Query project and authorized scope
    DB-->>Project: Project or not found
    Project-->>API: Result
    API-->>Client: Envelope response
```

## Planned Password Reset

```mermaid
sequenceDiagram
    actor User
    participant API
    participant Identity
    participant DB
    participant SMTP as SMTP/Mailpit or provider
    User->>API: Request password reset
    API->>Identity: Create single-use expiring token
    Identity->>DB: Store hashed token metadata
    Identity->>SMTP: Send reset message
    API-->>User: Generic success response
    User->>API: Submit token and new password
    API->>Identity: Validate token and password policy
    Identity->>DB: Consume token and update password hash
    API-->>User: Success or safe error
```

## Planned Automation Result Ingestion

```mermaid
sequenceDiagram
    actor CI
    participant API
    participant KeyAuth as API-key authentication
    participant Queue as Redis/Asynq
    participant Worker
    participant OLAP as ClickHouse
    CI->>API: Submit metadata/results with scoped key
    API->>KeyAuth: Verify hash, scope, revocation
    KeyAuth-->>API: Authorized project scope
    API->>Queue: Enqueue idempotent ingestion job
    API-->>CI: Accepted response
    Queue->>Worker: Deliver job
    Worker->>OLAP: Validate and write analytical facts
    Worker-->>Queue: Ack or retry/dead-letter
```

Ingestion endpoints require `Idempotency-Key`, acknowledge accepted batches within 500 ms for batches up to 1,000 result records, and deduplicate with stable domain event/result identifiers. The MVP processing target is 10,000 result records/minute. Accepted formats are documented in OpenAPI before implementation. Testra must not retain customer source code or raw API collection payloads.
