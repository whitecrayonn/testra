# BIBLICAL_TESTRA — Testra Canonical Architecture Reference

**Purpose:** A single consolidated reference for the Testra project covering product context, technology stack, system architecture, database schema, APIs, authentication, tenancy/RLS, event flow, and roadmap.

**Audience:** Engineers, architects, product managers, and operators working on Testra.

**Status vocabulary used in this document:**

- `[Implemented]` — Code and migrations exist and are wired into the running application.
- `[Approved]` — Accepted by ADR or engineering review; may be partially implemented.
- `[Planned]` — Scheduled in a future phase.
- `[Assumption]` — Documented intent that has not yet been ratified.

**Last updated:** Generated from repository sources.

---

## Table of Contents

1. [Product Context](#1-product-context)
2. [Technology Stack](#2-technology-stack)
3. [Repository Layout](#3-repository-layout)
4. [System Architecture](#4-system-architecture)
5. [Backend Clean Architecture](#5-backend-clean-architecture)
6. [Domain Modules](#6-domain-modules)
7. [Data Architecture and Storage](#7-data-architecture-and-storage)
8. [Database Schema](#8-database-schema)
9. [Authentication and Authorization](#9-authentication-and-authorization)
10. [Multi-Tenancy and RLS](#10-multi-tenancy-and-rls)
11. [API Design and Routing](#11-api-design-and-routing)
12. [Event and Request Flow](#12-event-and-request-flow)
13. [Frontend Architecture](#13-frontend-architecture)
14. [Machine Learning Service](#14-machine-learning-service)
15. [Deployment and Infrastructure](#15-deployment-and-infrastructure)
16. [Security, Privacy, and Compliance](#16-security-privacy-and-compliance)
17. [Performance Targets](#17-performance-targets)
18. [Roadmap and Phase Status](#18-roadmap-and-phase-status)
19. [Canonical Sources and Document Health](#19-canonical-sources-and-document-health)
20. [Glossary](#20-glossary)

---

## 1. Product Context

Testra is a unified quality engineering platform for teams that want to manage test cases, execute and ingest automation results, and gain transparent analytics without tool sprawl.

- **Mission:** One platform for every test — manual, automated, and API.
- **Vision:** Become the APAC-first enterprise-ready test operations platform.
- **North star:** Reduce time spent switching between QA tools and provide trustworthy, explainable intelligence from the customer’s own data.
- **Core philosophies:** One Platform, Every Test, Enterprise Ready, Automation First, No External LLM, Customer Owns Data, Transparent ML, API First, Localization Ready.
- **Primary ICP:** Mid-market to enterprise SaaS companies in APAC that need governance, audit, and multi-tenancy.
- **Scope MVP (Phase 1-3):** identity and tenancy, test management, manual runs, CI result ingestion, notifications, and a stable web dashboard.
- **Out of scope for MVP:** built-in CI runner, source-code hosting, external LLM features, billing, and WorkOS SSO.

**Authoritative sources:** `testra-master-context.md`, `testra-product-strategy.md`, `testra-product-discovery.md`, `testra-brd.md`, `testra/docs/engineering/PHASES.md`.

---

## 2. Technology Stack

| Layer | Technology | Role | Status |
|-------|------------|------|--------|
| Backend runtime | Go 1.23+ | API, business logic, workers | [Implemented] |
| HTTP router | chi/v5 | REST route tree and middleware | [Implemented] |
| Frontend framework | Next.js 15+ (App Router) | Web application and dashboards | [Implemented] |
| Frontend language | TypeScript 5+ | Type safety across web and SDK | [Implemented] |
| Styling | TailwindCSS | Utility-first CSS | [Implemented] |
| UI components | shadcn/ui + Radix UI | Accessible component primitives | [Implemented] |
| Forms/state | React Hook Form + Zod | Validation and client forms | [Implemented] |
| Primary database | PostgreSQL 16+ | Transactional data, tenant isolation, audit | [Implemented] |
| Analytics database | ClickHouse 24+ | Time-series results and events | [Planned] |
| Cache / queue | Redis 7+ | Sessions, rate limits, job queues | [Implemented] skeleton, queues [Planned] |
| Object storage | S3-compatible (MinIO / AWS S3) | Attachments, exports, artifacts | [Implemented] config |
| Background jobs | Asynq over Redis | Async ingestion and ML pipelines | [Planned] |
| Real-time | Server-Sent Events (SSE) | Live test run progress | [Implemented] |
| ML runtime | Python 3.12+ with FastAPI | ML inference service | [Implemented] skeleton |
| ML libraries | scikit-learn, XGBoost, statsmodels, pandas, numpy | Transparent classical ML | [Planned] |
| API documentation | OpenAPI 3.1 + Scalar | Interactive API docs | [Approved] |
| Package management | pnpm + Go modules + go.work | Workspace and dependency management | [Implemented] |
| CI/CD | GitHub Actions | Lint, test, build, integration tests | [Implemented] |
| Local development | Native services (PostgreSQL, Redis, Mailpit, MinIO) | Local development per ADR-009 | [Implemented] |
| Containerization | Docker / Docker Compose (optional) | Local alternative and deployment artifacts | [Implemented] optional |
| Infrastructure as Code | Terraform / Kubernetes manifests | Cloud provisioning (future) | [Planned] |
| Observability | OpenTelemetry, Prometheus, Grafana, Loki | Metrics, logs, traces | [Planned] |

---

## 3. Repository Layout

The project is a monorepo rooted at `c:/Private/project` with the main application code under `testra/`.

```
c:/Private/project/
├── testra/
│   ├── apps/
│   │   ├── api/            # Go backend API
│   │   ├── web/            # Next.js frontend
│   │   ├── ml/             # Python ML inference service
│   │   └── worker/         # Go background workers (planned)
│   ├── packages/
│   │   ├── shared/         # Shared TypeScript types and utilities
│   │   ├── ui/             # Shared React component library
│   │   ├── config/         # Shared tooling configs
│   │   └── sdk/            # Official Testra client SDK (planned)
│   ├── infra/
│   │   ├── terraform/
│   │   ├── k8s/
│   │   └── docker/
│   ├── docs/
│   │   ├── api/
│   │   ├── architecture/
│   │   ├── engineering/
│   │   └── reports/
│   ├── scripts/
│   ├── Makefile
│   ├── pnpm-workspace.yaml
│   ├── go.work
│   └── .env.example
├── 04_Architecture/
│   └── testra-software-architecture-decisions.md
├── testra-master-context.md
├── testra-product-strategy.md
├── testra-product-architecture-strategy.md
├── testra-product-discovery.md
└── testra-brd.md
```

---

## 4. System Architecture

Testra is built as a **modular monolith** [Implemented/Approved] rather than distributed microservices. This preserves solo-developer velocity while keeping domain boundaries clean enough to extract later.

### High-level diagram

```
User/Browser ──> Next.js Web App
                    │
Customer CI/CD ────> Go API (chi) ──> PostgreSQL
                    │      │
                    │      ├── Redis (cache, rate limits)
                    │      ├── S3-compatible storage
                    │      └── Python ML service
                    │
                    └── SSE stream for live run progress
```

### Deployment shape (MVP)

- Local and MVP production use native services on an Ubuntu VM behind Nginx (per ADR-003 and ADR-009).
- Go API, Next.js web app, PostgreSQL, Redis, and optional MinIO run as systemd services or containers.
- Future stages move to managed Kubernetes (EKS/GKE) for scale and data residency.

### Why modular monolith

- One deployable unit reduces cross-service debugging and networking.
- Authentication, audit, RBAC, and tenant isolation are enforced consistently.
- Clean Architecture ports and adapters allow hot modules (e.g., ML, ingestion) to be extracted into services later.
- Lower operational cost for a solo developer.

---

## 5. Backend Clean Architecture

Every backend module follows **Clean / Hexagonal Architecture**:

| Layer | Responsibility | Example |
|-------|----------------|---------|
| Domain | Entities, value objects, domain rules | `testra/apps/api/internal/results/domain.go` |
| Application | Use cases and service orchestration | `testra/apps/api/internal/results/service.go` |
| Ports | Interfaces for repositories and external clients | `testra/apps/api/internal/results/ports.go` |
| Adapters | Concrete HTTP handlers, SQL repositories | `testra/apps/api/internal/results/handler.go`, `repository.go` |

### Shared cross-cutting packages

- `testra/apps/api/internal/shared/config` — environment configuration.
- `testra/apps/api/internal/shared/db` — database open/wrapper and tenant context propagation.
- `testra/apps/api/internal/shared/errors` — domain error constants.
- `testra/apps/api/internal/shared/http` — response envelope helpers.
- `testra/apps/api/internal/shared/jwt` — JWT signing and parsing (HS256).
- `testra/apps/api/internal/shared/middleware` — auth, tenant, RBAC, audit, idempotency, rate limit, max body.
- `testra/apps/api/internal/shared/pagination` — cursor pagination helpers.
- `testra/apps/api/internal/shared/password` — password hashing.
- `testra/apps/api/internal/shared/validation` — email/name validation.
- `testra/apps/api/internal/shared/server` — chi router wiring.
- `testra/apps/api/internal/shared/tenant` — tenant resolver.
- `testra/apps/api/internal/shared/idempotency` — PostgreSQL-backed idempotency store.

---

## 6. Domain Modules

| Module | Status | Capabilities | Key code |
|--------|--------|--------------|----------|
| `identity` | [Implemented] | Register, login, refresh, password reset, TOTP MFA, /me | `testra/apps/api/internal/identity/` |
| `organization` | [Implemented] | Create, list, get organizations | `testra/apps/api/internal/organization/` |
| `workspace` | [Implemented] | Create, list, get, membership | `testra/apps/api/internal/workspace/` |
| `project` | [Implemented] | Create, list, get projects | `testra/apps/api/internal/project/` |
| `apikeys` | [Implemented] | Create, list, revoke scoped API keys | `testra/apps/api/internal/apikeys/` |
| `rbac` | [Implemented] | Roles, permissions, role assignments, SQL loader | `testra/apps/api/internal/rbac/`, `shared/middleware/rbac.go` |
| `testmanagement` | [Implemented] | Folders, suites, cases, versioning, full-text search | `testra/apps/api/internal/testmanagement/` |
| `results` | [Implemented] | Test runs and items, status updates, SSE progress | `testra/apps/api/internal/results/` |
| `automationhub` | [Implemented] | Ingest JUnit XML and Playwright/Cypress JSON | `testra/apps/api/internal/automationhub/` |
| `notification` | [Implemented] | In-app notifications, preferences, channels | `testra/apps/api/internal/notification/` |
| `audit` | [Implemented] | Immutable audit events on mutating endpoints | `testra/apps/api/internal/audit/`, `shared/middleware/audit.go` |
| `apitesting` | [Planned] | API test definitions, environments, execution | — |
| `defects` | [Planned] | Bug tracking, Jira sync | — |
| `analytics` | [Planned] | Dashboards, trends, aggregations | — |
| `intelligence` | [Planned] | Flaky detection, failure classification, risk scores | — |
| `integrationhub` | [Planned] | Jira, GitHub, GitLab, CI/CD webhooks | — |
| `billing` | [Planned] | Subscriptions, usage, invoices | — |

---

## 7. Data Architecture and Storage

Storage responsibilities from `testra/docs/architecture/DATABASE_DOCUMENTATION.md`:

| Store | Responsibility | Non-responsibilities |
|-------|----------------|----------------------|
| PostgreSQL 16 | Identity, tenancy, relational business data, audit records | High-volume analytical querying |
| ClickHouse 24 | Test results, events, time-series analytics | Transactional authority and session state |
| Redis 7 | Sessions, rate limiting, job queues, ephemeral cache | Durable business records |
| S3-compatible | Attachments, exports, model artifacts | Relational metadata authority |

### Data retention rules

- Customer source code and test scripts are **never stored** (zero customer code retention).
- Customer API collections are **not retained** unless explicitly created as Testra test definitions.
- Test run results: tier-based retention (30 days Free, 90 days Pro, 1 year Enterprise).
- Audit logs: minimum 7 years.
- User-uploaded attachments: tier-based, deletable on account termination.

### ClickHouse rules

- Append-oriented MergeTree tables.
- Partition by event month, order by tenant/project/event time.
- Deduplicate by stable domain identifiers.
- Default retention 13 months; backups 14 days daily and 8 weeks weekly.
- Transactional status and authorization remain authoritative in PostgreSQL.

### Redis rules

- Use the `testra:` key namespace.
- Explicit TTLs for ephemeral data.
- Redis loss must not destroy the only copy of business data.
- Queue jobs need retry, dead-letter handling, and idempotent consumers.

---

## 8. Database Schema

The authoritative schema is the set of `golang-migrate` files under `testra/apps/api/migrations/`. Current migrations are `000001` through `000018`.

### Identity and session

**users** — one user can belong to many organizations.

- `id UUID PRIMARY KEY`
- `email VARCHAR(255) UNIQUE NOT NULL`
- `password VARCHAR(255) NOT NULL`
- `name VARCHAR(255) NOT NULL`
- `mfa_secret VARCHAR(255) NOT NULL DEFAULT ''`
- `mfa_enabled BOOLEAN NOT NULL DEFAULT false`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- Index: `idx_users_email`

**refresh_tokens** — rotating opaque refresh tokens with family and reuse detection.

- `id UUID PRIMARY KEY`
- `user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE`
- `token_hash VARCHAR(255) UNIQUE NOT NULL`
- `family_id UUID NOT NULL`
- `expires_at TIMESTAMPTZ NOT NULL`
- `absolute_expires_at TIMESTAMPTZ NOT NULL`
- `revoked_at TIMESTAMPTZ`
- `replaced_by UUID`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- Indexes: `idx_refresh_tokens_hash`, `idx_refresh_tokens_user`, `idx_refresh_tokens_family`

**password_reset_tokens** — one-time password reset tokens.

- `id UUID PRIMARY KEY`
- `user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE`
- `token_hash VARCHAR(255) UNIQUE NOT NULL`
- `expires_at TIMESTAMPTZ NOT NULL`
- `used_at TIMESTAMPTZ`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- Indexes: `idx_password_reset_tokens_hash`, `idx_password_reset_tokens_user`

### Tenancy

**organizations**

- `id UUID PRIMARY KEY`
- `name VARCHAR(255) NOT NULL`
- `slug VARCHAR(255) UNIQUE NOT NULL`
- `owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- Indexes: `idx_organizations_slug`, `idx_organizations_owner`

**organization_members** — membership of a user in an organization.

- `organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE`
- `user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE`
- `role VARCHAR(50) NOT NULL DEFAULT 'member'`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- Primary key: `(organization_id, user_id)`
- Index: `idx_organization_members_user`

**workspaces**

- `id UUID PRIMARY KEY`
- `organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE`
- `name VARCHAR(255) NOT NULL`
- `slug VARCHAR(255) NOT NULL`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- Unique: `(organization_id, slug)`
- Indexes: `idx_workspaces_organization`, `idx_workspaces_slug`

**workspace_members**

- `workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE`
- `user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE`
- `role VARCHAR(50) NOT NULL DEFAULT 'member'`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- Primary key: `(workspace_id, user_id)`
- Index: `idx_workspace_members_user`

**projects**

- `id UUID PRIMARY KEY`
- `workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE`
- `name VARCHAR(255) NOT NULL`
- `key VARCHAR(10) NOT NULL`
- `description TEXT NOT NULL DEFAULT ''`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- Unique: `(workspace_id, key)`
- Index: `idx_projects_workspace`

### RBAC

**roles**

- `id UUID PRIMARY KEY`
- `name VARCHAR(50) NOT NULL UNIQUE`
- `description VARCHAR(255) NOT NULL DEFAULT ''`
- `is_system BOOLEAN NOT NULL DEFAULT false`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

**permissions**

- `id UUID PRIMARY KEY`
- `name VARCHAR(100) NOT NULL UNIQUE`
- `description VARCHAR(255) NOT NULL DEFAULT ''`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

**role_permissions**

- `role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE`
- `permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE`
- Primary key: `(role_id, permission_id)`

**role_assignments**

- `id UUID PRIMARY KEY`
- `role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE`
- `user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE`
- `scope_type VARCHAR(50) NOT NULL DEFAULT 'organization'`
- `scope_id UUID NOT NULL`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- Unique: `(role_id, user_id, scope_type, scope_id)`
- Indexes: `idx_role_assignments_user`, `idx_role_assignments_scope`

### API keys

**api_keys**

- `id UUID PRIMARY KEY`
- `workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE`
- `name VARCHAR(255) NOT NULL`
- `key_hash VARCHAR(255) NOT NULL UNIQUE`
- `key_prefix VARCHAR(20) NOT NULL`
- `scopes TEXT[] NOT NULL DEFAULT '{}'`
- `last_used_at TIMESTAMPTZ`
- `expires_at TIMESTAMPTZ`
- `revoked_at TIMESTAMPTZ`
- `created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- Indexes: `idx_api_keys_hash`, `idx_api_keys_workspace`

### Test management

**test_folders**

- `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`
- `workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE`
- `parent_id UUID REFERENCES test_folders(id) ON DELETE CASCADE`
- `name VARCHAR(255) NOT NULL`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- Indexes: `idx_test_folders_workspace`, `idx_test_folders_parent`

**test_suites**

- `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`
- `workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE`
- `folder_id UUID REFERENCES test_folders(id) ON DELETE SET NULL`
- `name VARCHAR(255) NOT NULL`
- `description TEXT DEFAULT ''`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- Indexes: `idx_test_suites_workspace`, `idx_test_suites_folder`

**test_cases**

- `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`
- `workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE`
- `project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE`
- `suite_id UUID REFERENCES test_suites(id) ON DELETE SET NULL`
- `title VARCHAR(500) NOT NULL`
- `description TEXT DEFAULT ''`
- `preconditions TEXT DEFAULT ''`
- `steps JSONB DEFAULT '[]'`
- `status VARCHAR(20) NOT NULL DEFAULT 'draft'`
- `priority VARCHAR(20) NOT NULL DEFAULT 'medium'`
- `tags TEXT[] DEFAULT '{}'`
- `version INTEGER NOT NULL DEFAULT 1`
- `created_by UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `search_tsv TSVECTOR`
- Indexes: `idx_test_cases_workspace`, `idx_test_cases_project`, `idx_test_cases_suite`, `idx_test_cases_status`, `idx_test_cases_search` (GIN on `search_tsv`)
- Triggers: `test_cases_search_tsv_insert` and `test_cases_search_tsv_update` update `search_tsv` from `title` and `description`.

**test_case_versions** — immutable snapshots of test case edits.

- `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`
- `test_case_id UUID NOT NULL REFERENCES test_cases(id) ON DELETE CASCADE`
- `version INTEGER NOT NULL`
- `title VARCHAR(500) NOT NULL`
- `description TEXT DEFAULT ''`
- `preconditions TEXT DEFAULT ''`
- `steps JSONB DEFAULT '[]'`
- `changed_by UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- Indexes: `idx_test_case_versions_case`, `idx_test_case_versions_version`

### Execution and results

**test_runs**

- `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`
- `workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE`
- `project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE`
- `suite_id UUID REFERENCES test_suites(id) ON DELETE SET NULL`
- `name VARCHAR(255) NOT NULL`
- `status VARCHAR(20) NOT NULL DEFAULT 'pending'`
- `total INTEGER NOT NULL DEFAULT 0`
- `passed INTEGER NOT NULL DEFAULT 0`
- `failed INTEGER NOT NULL DEFAULT 0`
- `skipped INTEGER NOT NULL DEFAULT 0`
- `blocked INTEGER NOT NULL DEFAULT 0`
- `duration_ms BIGINT NOT NULL DEFAULT 0`
- `source VARCHAR(20) NOT NULL DEFAULT 'manual'`
- `metadata JSONB DEFAULT '{}'`
- `created_by UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL`
- `started_at TIMESTAMPTZ`
- `completed_at TIMESTAMPTZ`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- Indexes: `idx_test_runs_workspace`, `idx_test_runs_project`, `idx_test_runs_suite`, `idx_test_runs_status`, `idx_test_runs_created_by`, `idx_test_runs_created_at`

**test_run_items**

- `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`
- `run_id UUID NOT NULL REFERENCES test_runs(id) ON DELETE CASCADE`
- `test_case_id UUID REFERENCES test_cases(id) ON DELETE SET NULL`
- `title VARCHAR(500) NOT NULL`
- `status VARCHAR(20) NOT NULL DEFAULT 'pending'`
- `duration_ms BIGINT NOT NULL DEFAULT 0`
- `error_message TEXT DEFAULT ''`
- `stack_trace TEXT DEFAULT ''`
- `artifacts JSONB DEFAULT '[]'`
- `sort_order INTEGER NOT NULL DEFAULT 0`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- Indexes: `idx_test_run_items_run`, `idx_test_run_items_case`, `idx_test_run_items_status`, `idx_test_run_items_sort`

### Idempotency

**idempotency_records**

- `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`
- `workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE`
- `operation VARCHAR(50) NOT NULL`
- `key VARCHAR(255) NOT NULL`
- `request_fingerprint VARCHAR(64) NOT NULL`
- `response_body JSONB NOT NULL`
- `status_code INTEGER NOT NULL`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `expires_at TIMESTAMPTZ NOT NULL`
- Unique: `(workspace_id, operation, key)`
- Indexes: `idx_idempotency_records_lookup`, `idx_idempotency_records_expires`

### Notifications

**notifications**

- `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`
- `organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE`
- `user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE`
- `type VARCHAR(50) NOT NULL DEFAULT 'system'`
- `title VARCHAR(255) NOT NULL`
- `body TEXT DEFAULT ''`
- `link VARCHAR(500) DEFAULT ''`
- `read BOOLEAN NOT NULL DEFAULT false`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- Indexes: `idx_notifications_user_org`, `idx_notifications_read`

**notification_preferences**

- `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`
- `organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE`
- `user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE`
- `in_app_enabled BOOLEAN NOT NULL DEFAULT true`
- `email_enabled BOOLEAN NOT NULL DEFAULT false`
- `slack_enabled BOOLEAN NOT NULL DEFAULT false`
- `teams_enabled BOOLEAN NOT NULL DEFAULT false`
- `webhook_enabled BOOLEAN NOT NULL DEFAULT false`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- Unique: `(organization_id, user_id)`

**notification_channels**

- `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`
- `organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE`
- `workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE`
- `type VARCHAR(50) NOT NULL`
- `name VARCHAR(255) NOT NULL`
- `config JSONB NOT NULL DEFAULT '{}'`
- `created_by UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- Index: `idx_notification_channels_workspace`

### Audit

**audit_events**

- `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`
- `user_id UUID REFERENCES users(id) ON DELETE SET NULL`
- `action VARCHAR(100) NOT NULL`
- `resource VARCHAR(100) NOT NULL`
- `resource_id VARCHAR(255)`
- `ip_address VARCHAR(45)`
- `metadata JSONB DEFAULT '{}'`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- Indexes: `idx_audit_events_user_id`, `idx_audit_events_action`, `idx_audit_events_created_at`

---

## 9. Authentication and Authorization

### Authentication model [Implemented]

- **Primary auth:** email + password.
- **Password hashing:** bcrypt via `testra/apps/api/internal/shared/password`.
- **Access token:** JWT signed with HS256 (`golang-jwt/jwt/v5`), 15-minute expiry, claims include `user_id`, `email`, `sub`, `iat`, `exp`.
- **Refresh token:** opaque string (`rt_` + 32 random bytes hex), SHA-256 hashed in DB, belongs to a `family_id`, rotated on each use, 30-day sliding expiry and 90-day absolute expiry, reuse detection revokes the whole family.
- **MFA:** TOTP via `pquerna/otp`, setup returns secret and QR URL, verified before enabling, disabled with a valid code.
- **Password reset:** random token hashed with SHA-256, 30-minute expiry, one-time use, sent via SMTP.
- **API keys:** scoped, SHA-256 hashed with a `testra_` prefix, plaintext shown only at creation, default 90-day expiry, revocable. Validation checks expiry and updates `last_used_at`.

### Authorization model [Implemented]

- System roles seeded in migration `000006_add_rbac.up.sql`:
  - `owner` — full access.
  - `admin` — manage projects, members, settings.
  - `qa_engineer` — create and execute tests, report defects.
  - `viewer` — read-only.
- Permission strings are organized by resource (e.g., `orgs:read`, `workspaces:create`, `projects:read`, `tests:create`, `runs:ingest`, `notifications:read`).
- `shared/middleware/rbac.go` defines `RequirePermission(loader, permission)`. It loads the user’s permissions from `role_assignments` joined through `role_permissions` to `permissions` for the resolved `organization_id` scope and checks the required string.
- `shared/middleware/tenant.go` resolves the organization before `RequirePermission` runs, so permissions are always evaluated within the active tenant.

### Important current gap

The `POST /ingest` endpoint is currently protected by the JWT `Auth` middleware in `testra/apps/api/internal/shared/server/server.go`. ADR-001 and the API key domain both anticipate API-key-based CI authentication, but the middleware chain for `/ingest` does not yet verify an API key header. This is a known integration point.

---

## 10. Multi-Tenancy and RLS

### Tenant model

- Shared database, shared schema.
- Hierarchy: `organization > workspace > project`.
- Tenant-scoped tables carry `organization_id` or a `workspace_id` that resolves to an organization.
- A request must establish identity, organization membership, and permission before touching tenant data.

### Request-scoped tenant propagation

`shared/middleware/tenant.go` performs the following for every protected route:

1. Extract `user_id` from the JWT context.
2. Resolve the target `organization_id` from the URL parameter, query string, or request body (via `OrgIDFromURLParam`, `OrgIDFromQuery`, `OrgIDFromBody`, and workspace/project/run resolvers in `shared/tenant/resolver.go`).
3. Acquire a dedicated `*sql.Conn` from the pool.
4. Execute `SET app.tenant_id = '<org-id>'` on that connection.
5. Store the connection and tenant ID in the request context.
6. Verify membership in `organization_members`.
7. On handler completion, `RESET app.tenant_id` and close the connection.

`shared/db/db.go` wraps `*sql.DB` and transparently uses the context’s connection or transaction, so all repository calls on the request use the same connection with the tenant variable set.

### Row Level Security

RLS is enabled on tenant-scoped tables. Example policy from `testra/apps/api/migrations/000009_add_rls_policies.up.sql`:

```sql
CREATE POLICY organizations_tenant ON organizations
    USING (id = current_setting('app.tenant_id', true)::uuid);
```

Other tables use workspace/project indirection, e.g.:

```sql
CREATE POLICY projects_tenant ON projects
    USING (workspace_id IN (
        SELECT id FROM workspaces WHERE organization_id = current_setting('app.tenant_id', true)::uuid
    ));
```

Tables with RLS: `organizations`, `organization_members`, `workspaces`, `workspace_members`, `projects`, `api_keys`, `role_assignments`, `test_folders`, `test_suites`, `test_cases`, `test_case_versions`, `test_runs`, `test_run_items`, `idempotency_records`, `notifications`, `notification_preferences`, `notification_channels`.

Tables without RLS: `users` (cross-tenant), `roles`, `permissions`, `role_permissions` (system tables), `refresh_tokens`, `password_reset_tokens` (session tables tied to user ID).

### Transactions and RLS

`shared/db/db.go` `BeginTx` sets `SET LOCAL app.tenant_id` when a tenant ID is in the context, ensuring RLS policies see the correct scope inside a transaction.

---

## 11. API Design and Routing

### Conventions [Approved/Implemented]

- Base path: `/api/v1`.
- Major URL versioning. Breaking changes bump the version.
- RESTful resource-oriented routes.
- Response envelope: `data`, `meta`, `error` fields (most handlers return `map[string]any{"data": ...}`).
- Cursor pagination for list endpoints (`?cursor=` and `?limit=`).
- Idempotency-Key header required for `POST /ingest` and other mutating side-effect endpoints.
- Stable error codes: `INVALID_INPUT`, `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`, `CONFLICT`, `INTERNAL_ERROR`, etc.
- Input and output snake_case, RFC 3339 timestamps.

### Implemented routes

All routes are under `/api/v1` unless otherwise noted.

| Method | Path | Purpose | Permission / Notes |
|--------|------|---------|-------------------|
| POST | /auth/register | Register a new user | Public |
| POST | /auth/login | Login, returns JWT + refresh | Public |
| POST | /auth/refresh | Refresh access token | Public, refresh token |
| POST | /auth/password-reset/request | Request password reset | Public |
| POST | /auth/password-reset/confirm | Confirm password reset | Public |
| GET | /auth/me | Current user | Authenticated |
| POST | /auth/mfa/setup | Begin MFA enrollment | Authenticated |
| POST | /auth/mfa/verify | Complete MFA enrollment | Authenticated |
| POST | /auth/mfa/disable | Disable MFA | Authenticated |
| POST | /organizations | Create organization | Authenticated |
| GET | /organizations | List organizations | Authenticated |
| GET | /organizations/{id} | Get organization | `orgs:read` |
| POST | /workspaces | Create workspace | `workspaces:create` |
| GET | /workspaces | List workspaces | `workspaces:read` |
| GET | /workspaces/{id} | Get workspace | `workspaces:read` |
| POST | /projects | Create project | `projects:create` |
| GET | /projects | List projects | `projects:read` |
| GET | /projects/{id} | Get project | `projects:read` |
| POST | /api-keys | Create API key | `apikeys:create` |
| GET | /api-keys | List API keys | `apikeys:read` |
| DELETE | /api-keys/{id} | Revoke API key | `apikeys:delete` |
| POST | /test-folders | Create folder | `tests:create` |
| GET | /test-folders | List folders | `tests:read` |
| GET | /test-folders/{id} | Get folder | `tests:read` |
| PUT | /test-folders/{id} | Update folder | `tests:update` |
| DELETE | /test-folders/{id} | Delete folder | `tests:delete` |
| POST | /test-suites | Create suite | `tests:create` |
| GET | /test-suites | List suites | `tests:read` |
| GET | /test-suites/{id} | Get suite | `tests:read` |
| PUT | /test-suites/{id} | Update suite | `tests:update` |
| DELETE | /test-suites/{id} | Delete suite | `tests:delete` |
| POST | /test-cases | Create test case | `tests:create` |
| GET | /test-cases | List test cases | `tests:read` |
| GET | /test-cases/search | Search test cases | `tests:read` |
| GET | /test-cases/{id} | Get test case | `tests:read` |
| GET | /test-cases/{id}/versions | List versions | `tests:read` |
| PUT | /test-cases/{id} | Update test case | `tests:update` |
| DELETE | /test-cases/{id} | Delete test case | `tests:delete` |
| POST | /test-runs | Create run | `runs:create` |
| GET | /test-runs | List runs | `runs:read` |
| GET | /test-runs/{id} | Get run | `runs:read` |
| PUT | /test-runs/{id} | Update run status | `runs:update` |
| DELETE | /test-runs/{id} | Delete run | `runs:delete` |
| GET | /test-runs/{id}/items | List run items | `runs:read` |
| GET | /test-runs/{id}/stream | SSE run progress | `runs:read` |
| PUT | /test-run-items/{id} | Update item status | `runs:update` |
| POST | /ingest | Ingest CI results | `runs:ingest`, requires Idempotency-Key |
| GET | /notifications | List notifications | `notifications:read` |
| GET | /notifications/unread-count | Count unread | `notifications:read` |
| PATCH | /notifications/{id} | Mark read | `notifications:update` |
| DELETE | /notifications/{id} | Delete notification | `notifications:delete` |
| POST | /notifications | Create notification | `notifications:create` |
| GET | /notification-preferences | Get preferences | `notification_preferences:read` |
| PUT | /notification-preferences | Update preferences | `notification_preferences:update` |
| GET | /notification-channels | List channels | `notification_channels:read` |
| POST | /notification-channels | Create channel | `notification_channels:create` |
| PUT | /notification-channels/{id} | Update channel | `notification_channels:update` |
| DELETE | /notification-channels/{id} | Delete channel | `notification_channels:delete` |
| GET | /health | Health check | Public |

---

## 12. Event and Request Flow

### Request trust flow

Every protected request flows through:

```
TLS termination
    -> Request ID / structured context
    -> Authentication (JWT or API key)
    -> Tenant and membership resolution
    -> Permission check
    -> Input validation
    -> Module use case
    -> Authorized data access
    -> Safe response envelope
```

The middleware order in `shared/server/server.go` is:

1. `middleware.Logger`
2. `middleware.Recoverer`
3. `middleware.RequestID`
4. Content-Type header
5. CORS
6. `MaxBodySize`
7. `Auth` on protected groups
8. `TenantContext`
9. `RequirePermission`
10. `AuditLog` (on mutating endpoints)
11. Handler -> Service -> Repository

### Test run progress SSE flow

1. `POST /test-runs` creates a run in `pending` status with `TestRunItem` rows for each selected test case.
2. Clients open `GET /test-runs/{id}/stream`.
3. `results.Service.UpdateItemStatus` updates an item, recalculates run counts, and broadcasts a `RunProgressEvent` through an in-memory `progressHub`.
4. `StreamRunProgress` writes events as `text/event-stream`.
5. When the run reaches a terminal status, the hub closes all channels and removes the subscription list.

### Ingestion flow

1. `POST /ingest` with `Idempotency-Key` header and JSON body containing `workspace_id`, `project_id`, optional `suite_id`, `name`, `format` (`junit`, `playwright`, `cypress`), and `payload`.
2. The idempotency middleware checks `(workspace_id, operation, key_hash)`. If the same key and body fingerprint exist, the stored response is replayed. If the key exists with a different body, a 409 is returned.
3. `automationhub.Service.Ingest` parses the payload.
   - JUnit: XML unmarshals into `JUnitTestSuites`, maps cases to `test_run_items`.
   - Playwright/Cypress: JSON unmarshals into `PlaywrightReport`, maps tests to items.
4. A `test_runs` row is created with `source = ci`, status `running`, then counts are computed and the run is marked `passed` or `failed`.
5. Response returns `run_id`, `total`, `passed`, `failed`, `skipped`, `duration_ms`.

### Audit flow

`shared/middleware/audit.go` wraps mutating endpoints. After the handler returns, it calls `audit.Service.Log` with `user_id`, `action`, `resource`, `resource_id`, and `ip_address`. Events are appended to `audit_events`.

---

## 13. Frontend Architecture

- **Framework:** Next.js 15 App Router, React, TypeScript.
- **Styling:** TailwindCSS, shadcn/ui + Radix primitives.
- **State:** Server state is fetched directly through a lightweight API client; forms use React Hook Form + Zod.
- **API client:** `testra/apps/web/lib/api.ts` fetches `NEXT_PUBLIC_API_URL`, stores the JWT in `localStorage` as `testra_token`, and attaches `Authorization: Bearer ...`.
- **Route groups:**
  - `(auth)` — login, register, forgot-password, reset-password, MFA setup.
  - `(dashboard)` — dashboard, projects, test-cases, test-runs, settings.
  - Onboarding at `/onboarding` creates the first organization and workspace.
- **Workspace/project context:** The dashboard reads `testra_workspace_id`, `testra_workspace_name`, `testra_project_id`, `testra_project_name` from `localStorage` for the current context UI.
- **Shared packages:** `packages/shared`, `packages/ui`, `packages/config`.

---

## 14. Machine Learning Service

- **Runtime:** Python 3.12+ FastAPI service (`testra/apps/ml/api/main.py`).
- **Current state:** Skeleton with a `/health` endpoint [Implemented].
- **Planned capabilities (Phase 6):**
  - Flaky test detection using time-series variance scoring.
  - Failure classification with rule-based filtering + DBSCAN/HDBSCAN clustering and optional XGBoost.
  - Risk/health scores with logistic regression / XGBoost and SHAP explainability.
  - Release readiness thresholds and trends.
- **Principles:** No external LLM APIs; models trained per tenant on that tenant’s data only; inputs limited to test metadata and results, never source code or secrets.

---

## 15. Deployment and Infrastructure

### Local development [Implemented]

Per ADR-009, the official local workflow uses native services:

- PostgreSQL 16+, Redis 7+, Mailpit, MinIO.
- Go, Node.js, pnpm, Python installed locally.
- `pnpm dev` checks services, applies migrations, and starts API, web, worker, and ML services.
- Docker Compose remains available under `testra/infra/docker/` as an optional alternative.

### MVP production [Approved]

- Ubuntu VM with systemd and Nginx reverse proxy.
- PostgreSQL, Redis, and S3-compatible object store.
- Go API and Next.js web as systemd services.
- TLS terminated by Nginx or Cloudflare.
- Migrations applied via `testra/apps/api/cmd/migrator/main.go` in CI/CD, never manually in production.

### Future [Planned]

- AWS/GCP managed Kubernetes (EKS/GKE).
- Separate worker and ML pods.
- Terraform for repeatable infrastructure.
- Multi-region clusters for APAC data residency.

---

## 16. Security, Privacy, and Compliance

- **Authentication:** Short-lived 15-minute JWTs, rotating refresh tokens, TOTP MFA, API key expiry and revocation, minimum 12-character passwords.
- **Authorization:** RBAC, tenant membership, and RLS at the database layer.
- **Secrets:** Passwords, refresh tokens, password reset tokens, and API keys are all hashed (bcrypt/SHA-256) before storage.
- **Transport:** TLS in production, CORS restricted to configured origins, `MaxBodySize` middleware.
- **Rate limiting:** Local in-memory rate limiter implemented; Redis-backed token bucket planned.
- **Audit:** Immutable `audit_events` for mutating actions.
- **Privacy:** Zero customer code retention, zero API collection retention, customer-owned data, no external LLM processing, tenant-isolated ML models.
- **Compliance posture:** Audit logs (7 years), RBAC, MFA, encryption, and data-residency path support SOC 2 readiness.

---

## 17. Performance Targets

From `testra/docs/architecture/adrs/ADR-008-performance-targets.md`:

- API reads: p95 ≤ 300 ms, p99 ≤ 750 ms.
- API writes: p95 ≤ 500 ms, p99 ≤ 1000 ms.
- PostgreSQL queries: p95 ≤ 50 ms.
- Request timeout: 30 seconds.
- Background job timeout: 5 minutes.
- Upload size: max 50 MB.
- Capacity: 500 concurrent users, 50 req/s.
- ClickHouse ingestion: 10,000 records/minute (planned).

---

## 18. Roadmap and Phase Status

From `testra/docs/engineering/PHASES.md`:

| Phase | Name | Status | Highlights |
|-------|------|--------|------------|
| 0 | Foundation | Completed | CI, native dev env, OpenAPI skeleton, project module |
| 1 | Identity & Tenancy | Completed | Auth, MFA, RBAC, API keys, web onboarding |
| 2 | Test Management Core | Approved | Cases, suites, folders, versioning, search, audit |
| 3 | Execution & Results | Approved | Manual runs, CI ingestion, SSE, idempotency |
| 3.5 | Product UX Completion | Completed | Settings, placeholders, a11y, build/lint passes |
| 4 | API Testing & Defects | Pending | API tests, defects, Jira sync, notifications |
| 5 | Dashboard, Analytics & Launch | Pending | ClickHouse analytics, SDK, production deploy |
| 6 | V2 Intelligence | Pending | Flaky detection, failure classification, risk scoring |

---

## 19. Canonical Sources and Document Health

### Authoritative sources per ADR-002

| Concern | Source of truth |
|---------|-----------------|
| HTTP API contracts | `testra/docs/api/openapi/openapi.yaml` (and `server.go` for current wiring) |
| Database schema and runtime | `testra/apps/api/migrations/*.sql` |
| Implementation status | `testra/docs/engineering/PHASES.md` |
| Architectural decisions | `testra/docs/architecture/adrs/ADR-*.md` |
| Code behavior | `testra/apps/api/**/*.go` and `testra/apps/web/**/*.tsx` |

### Important distinction: draft architecture document

`04_Architecture/testra-software-architecture-decisions.md` is a **draft awaiting executive approval** and contains proposed alternatives (Clerk/WorkOS identity, Docker Compose local dev, managed-platform MVP) that differ from the accepted ADRs and implemented code. It should not be used as a source of truth until it is reconciled and approved.

### Known gaps and recommendations

1. **API key auth for ingestion:** The `/ingest` endpoint currently requires a JWT. ADR-001 and the API key domain expect API-key auth for CI/CD. Add an API key middleware or alternate auth path.
2. **OpenAPI synchronization:** Ensure `testra/docs/api/openapi/openapi.yaml` matches the route table in `server.go` after every endpoint change.
3. **Rate limiter wiring:** The local rate limiter is created but not attached to routes in `server.go`. Decide per-route limits and wire them.
4. **ClickHouse/async pipeline:** Deferred to Phase 5. Document the exact migration and contract when implemented.
5. **SDK generation:** `packages/sdk` is a placeholder; generate from OpenAPI after contracts are stable.
6. **Reconcile SADD with ADRs:** Resolve conflicts between `04_Architecture/...` and accepted ADRs, then either update the draft or archive it.

---

## 20. Glossary

- **ADR** — Architecture Decision Record.
- **API** — Application Programming Interface.
- **CI/CD** — Continuous Integration / Continuous Delivery.
- **JWT** — JSON Web Token.
- **MFA** — Multi-Factor Authentication.
- **RLS** — PostgreSQL Row Level Security.
- **SSE** — Server-Sent Events.
- **TOTP** — Time-based One-Time Password.
- **TSV** — Full-text search vector (`tsvector`).

---

## Index of Key Files

- `testra/apps/api/cmd/api/main.go` — API server entrypoint.
- `testra/apps/api/cmd/migrator/main.go` — Migration runner.
- `testra/apps/api/internal/shared/server/server.go` — Route tree and middleware wiring.
- `testra/apps/api/internal/shared/middleware/auth.go` — JWT authentication.
- `testra/apps/api/internal/shared/middleware/tenant.go` — Tenant context and RLS connection setup.
- `testra/apps/api/internal/shared/middleware/rbac.go` — Permission enforcement.
- `testra/apps/api/internal/shared/middleware/idempotency.go` — Idempotency-Key middleware.
- `testra/apps/api/internal/shared/db/db.go` — DB wrapper and transaction tenant handling.
- `testra/apps/api/internal/identity/service.go` — Registration, login, MFA, refresh, password reset.
- `testra/apps/api/internal/apikeys/service.go` — API key generation, hashing, validation.
- `testra/apps/api/internal/results/service.go` — Run lifecycle and SSE progress hub.
- `testra/apps/api/internal/automationhub/service.go` — JUnit / Playwright / Cypress ingestion.
- `testra/apps/web/lib/api.ts` — Web API client and token management.
- `testra/apps/ml/api/main.py` — ML service skeleton.
- `testra/docs/architecture/adrs/ADR-001-hybrid-auth.md` — Self-hosted auth decision.
- `testra/docs/architecture/adrs/ADR-004-tenant-isolation-strategy.md` — Tenant isolation decision.
- `testra/docs/architecture/adrs/ADR-009-native-development-environment.md` — Native dev environment decision.
- `testra/docs/engineering/PHASES.md` — Implementation phase status.
- `testra/docs/api/API_DESIGN_GUIDELINES.md` — API conventions.
- `testra/docs/architecture/DATABASE_DOCUMENTATION.md` — Storage responsibilities and invariants.
- `testra/docs/architecture/ERD.md` — Entity relationship overview.
- `testra/docs/architecture/SYSTEM_FLOWS.md` — System and request flows.
