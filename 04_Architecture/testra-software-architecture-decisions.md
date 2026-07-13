# Testra — Software Architecture Decision Document

**Version:** 1.0  
**Status:** Draft — Awaiting Engineering Approval  
**Prepared by:** Chief Software Architect / Engineering Lead  
**Date:** July 2026  
**Classification:** Internal — Confidential

---

## Purpose

This document records the complete software architecture decisions for Testra before implementation begins. It covers the technology stack, system architecture, repository strategy, backend, frontend, database, authentication, API, machine learning, deployment, and folder structure. Each decision is justified against Testra's constraints: solo developer, budget-conscious, scalable, enterprise-ready, API-first, privacy-first, zero customer code retention, zero API collection retention, and no external LLM dependency.

---

## 1. Complete Technology Stack

### 1.1 Core Stack

| Layer | Technology | Role |
|---|---|---|
| **Backend Runtime** | Go (Golang) 1.23+ | Primary API, business logic, workers |
| **Frontend Framework** | Next.js 15+ (App Router) | Web application, dashboards, marketing site |
| **Frontend Language** | TypeScript 5+ | Type safety across web and SDK |
| **Styling** | TailwindCSS 4+ | Utility-first styling |
| **UI Components** | shadcn/ui + Radix UI | Accessible, composable component base |
| **Primary Database** | PostgreSQL 16+ | Transactional data, tenant isolation |
| **Analytics Database** | ClickHouse 24+ | Time-series test results, events, telemetry |
| **Cache / Queue** | Redis 7+ | Sessions, caching, job queues, rate limiting |
| **Search** | PostgreSQL Full-Text Search (MVP) → Meilisearch (V2+) | Test case and defect search |
| **Object Storage** | S3-compatible (MinIO self-hosted / AWS S3) | Attachments, exports, CI artifacts metadata |
| **Background Jobs** | Asynq (Go) over Redis | Task queues, scheduled jobs, ML inference |
| **Real-Time** | Server-Sent Events (SSE) + WebSockets | Live test execution, notifications |
| **ML Runtime** | Python 3.12+ with scikit-learn / XGBoost / statsmodels | Transparent, non-LLM intelligence |
| **Containerization** | Docker + Docker Compose (local) / Kubernetes (production) | Consistent environments |
| **Infrastructure as Code** | Terraform | Cloud provisioning |
| **Observability** | OpenTelemetry → Prometheus + Grafana + Loki | Metrics, logs, traces |
| **API Documentation** | OpenAPI 3.1 + Scalar | Interactive API docs |
| **Package Management** | pnpm (JS) + Go modules | Workspace and dependency management |

### 1.2 Why This Stack

- **Go** is compiled, memory-efficient, and concurrency-native. A single developer can build, reason about, and deploy one cohesive backend without runtime complexity. Its static binary simplifies container images and CI/CD.
- **Next.js + TypeScript** provides a modern React stack with server-side rendering for fast dashboards, excellent type safety, and a large hiring market for future team expansion.
- **PostgreSQL** is the proven default for relational, multi-tenant SaaS. It supports row-level security, JSONB for flexible metadata, and robust ACID guarantees for audit trails.
- **ClickHouse** handles high-ingest, time-series test results and events far more cost-effectively than relational scaling. It aligns with Testra's automation result ingestion volume.
- **Redis + Asynq** keeps job processing in-language (Go) and avoids operational sprawl from separate queue systems.
- **Python for ML only** isolates the statistical intelligence layer. Python's ecosystem dominates transparent, explainable ML without external AI APIs.
- **Terraform + Kubernetes** create an enterprise-ready, cloud-agnostic, audit-friendly deployment path, while Docker Compose supports local solo development.

---

## 2. High-Level System Architecture

### 2.1 Pattern: Modular Monolith

Testra will ship as a **modular monolith** rather than microservices.

### 2.2 Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Client Layer                                │
│   Next.js Web App   │   Browser Extensions (future)   │   CI/CD    │
└──────────────────────────────────┬──────────────────────────────────┘
                                   │
                              ┌────┴────┐
                              │   CDN   │
                              └────┬────┘
                                   │
┌──────────────────────────────────┴──────────────────────────────────┐
│                         API Gateway / Ingress                         │
│   TLS termination │ Rate limiting │ WAF │ Request ID / tracing         │
└──────────────────────────────────┬──────────────────────────────────┘
                                   │
┌──────────────────────────────────┴──────────────────────────────────┐
│                         API Application (Go)                        │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌─────────────┐  │
│  │   Identity   │ │   Workspace  │ │   Projects   │ │   Billing   │  │
│  └──────────────┘ └──────────────┘ └──────────────┘ └─────────────┘  │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌─────────────┐  │
│  │ Test Mgmt    │ │ API Testing  │ │ Automation   │ │   Defects   │  │
│  └──────────────┘ └──────────────┘ └──────────────┘ └─────────────┘  │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌─────────────┐  │
│  │   Results    │ │  Analytics   │ │ Intelligence │ │   Audit     │  │
│  └──────────────┘ └──────────────┘ └──────────────┘ └─────────────┘  │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐                │
│  │ Notification │ │ Integration  │ │  Marketplace │                │
│  └──────────────┘ └──────────────┘ └──────────────┘                │
└──────────────────────────────────┬──────────────────────────────────┘
                                   │
┌──────────────────────────────────┴──────────────────────────────────┐
│                         Worker Layer (Go + Python)                  │
│   Asynq Workers  │  ML Inference Service  │  Report Generators       │
└──────────────────────────────────┬──────────────────────────────────┘
                                   │
┌──────────────────────────────────┴──────────────────────────────────┐
│                         Data Layer                                  │
│   PostgreSQL (OLTP)   │   ClickHouse (OLAP)   │   Redis              │
│   Object Storage      │   Search (PG FTS → Meilisearch)            │
└─────────────────────────────────────────────────────────────────────┘
```

### 2.3 Why Modular Monolith

- **Solo developer velocity**: A single deployable unit eliminates distributed-system debugging, cross-service versioning, and inter-service networking.
- **Enterprise-ready from MVP**: Authentication, audit trails, RBAC, and tenant isolation can be enforced consistently in one codebase.
- **Scalable exit path**: Domains are internally isolated via Clean Architecture ports and adapters. If a module later needs to become a standalone service, its boundary is already defined.
- **Cost control**: One container fleet, one database cluster (with read replicas), one CI/CD pipeline.

---

## 3. Repository Structure: Monorepo

### 3.1 Decision

Use a **single monorepo** for the entire Testra platform.

### 3.2 Repository Layout

```
testra/
├── apps/
│   ├── api/                  # Go backend API
│   ├── web/                  # Next.js frontend
│   ├── worker/               # Go background job workers
│   └── ml/                   # Python ML inference service
├── packages/
│   ├── shared/               # Shared TypeScript types and utilities
│   ├── ui/                   # Shared React component library
│   ├── config/               # Shared tooling configs (eslint, tsconfig)
│   └── sdk/                  # Official Testra client SDK (TypeScript)
├── infra/
│   ├── terraform/            # Cloud infrastructure
│   ├── k8s/                  # Kubernetes manifests / Helm charts
│   └── docker/               # Docker Compose and images
├── docs/
│   ├── api/                  # OpenAPI specifications
│   ├── architecture/         # ADRs and diagrams
│   └── runbooks/             # Operational playbooks
├── scripts/                  # Automation scripts
├── Makefile                  # Common development tasks
├── pnpm-workspace.yaml       # JS workspace definition
├── go.work                   # Go workspace definition
└── README.md
```

### 3.3 Why Monorepo

- **Atomic changes**: A feature touching API, web, and SDK can be committed, reviewed, and deployed together.
- **Single CI/CD pipeline**: Reduces DevOps overhead for a solo developer.
- **Shared contracts**: OpenAPI specs and TypeScript types live in one place, preventing API/client drift.
- **No cross-repo dependency hell**: Especially important when iterating rapidly toward MVP.
- **Future team scaling**: Turborepo / pnpm workspaces and Go workspaces support future parallelization without splitting repositories.

---

## 4. Backend Architecture

### 4.1 Pattern: Clean / Hexagonal Architecture

Each domain module has four internal layers:

| Layer | Responsibility |
|---|---|
| **Domain** | Entities, value objects, domain rules, invariants |
| **Application** | Use cases, command/query handlers, service orchestration |
| **Ports** | Interfaces for repositories, message buses, external clients |
| **Adapters** | Concrete implementations: SQL repositories, HTTP handlers, queue consumers, external API clients |

### 4.2 Domain Modules

Mapped directly from approved Testra modules:

1. `identity` — users, sessions, SSO, MFA, API keys
2. `organization` — tenants, subscriptions, billing seats
3. `workspace` — workspaces, projects, settings
4. `project` — project metadata, environments, variables
5. `testmanagement` — test cases, suites, plans, folders, version history
6. `apitesting` — API test definitions, collections, variables, execution engine
7. `automationhub` — automation framework ingestion (Playwright, Cypress, JUnit, Pytest)
8. `results` — test run results, logs, artifacts metadata
9. `defects` — bug tracking, Jira/Linear/GitHub sync
10. `analytics` — dashboards, trends, aggregations
11. `intelligence` — flaky detection, failure classification, risk/health scores (ports only; inference in ML service)
12. `notification` — in-app, email, Slack, Teams alerts
13. `integrationhub` — Jira, GitHub, GitLab, CI/CD webhooks
14. `audit` — immutable audit trail and evidence export
15. `billing` — subscriptions, usage, invoices (initially integrated; later Stripe)
16. `marketplace` — plugin/extension hooks (V3)

### 4.3 Cross-Cutting Concerns

| Concern | Implementation |
|---|---|
| **Multi-tenancy** | Row-level `tenant_id` columns + middleware injection |
| **RBAC** | Permission definitions + middleware enforcement |
| **Audit logging** | Event stream written to PostgreSQL and archived to object storage |
| **Rate limiting** | Redis token bucket per tenant / API key |
| **Request tracing** | OpenTelemetry trace IDs through all layers |
| **Error handling** | Structured errors, consistent API response envelope |

### 4.4 Why This Backend Architecture

- **Clean Architecture** keeps business rules independent of framework and database, enabling testing and future service extraction.
- **Domain boundaries** mirror the approved product modules, so one PRD maps to one code module.
- **Go's simplicity** allows one engineer to own the entire backend without language or framework surprises.

---

## 5. Frontend Architecture

### 5.1 Framework and Patterns

| Decision | Choice |
|---|---|
| Framework | Next.js 15 App Router |
| Rendering | Server Components by default; Client Components for interactivity |
| Data fetching | TanStack Query (React Query) for server state |
| Client state | Zustand for global UI state |
| Forms | React Hook Form + Zod |
| Tables | TanStack Table |
| Charts | Tremor / Recharts |
| Routing | Next.js file-based App Router |
| Build output | Standalone output for container deployment |

### 5.2 Frontend Structure

```
apps/web/
├── app/                        # Next.js App Router
│   ├── (auth)/                 # Auth routes group
│   ├── (dashboard)/            # Authenticated dashboard routes
│   │   ├── [workspace]/
│   │   │   ├── projects/
│   │   │   ├── test-cases/
│   │   │   ├── api-tests/
│   │   │   ├── runs/
│   │   │   ├── defects/
│   │   │   └── settings/
│   └── api/                    # Next.js API routes for edge hooks
├── features/                   # Feature-based modules
│   ├── test-management/
│   ├── api-testing/
│   ├── automation-hub/
│   ├── defects/
│   ├── analytics/
│   └── settings/
├── components/                 # Shared UI components
├── lib/                        # Utilities, API clients, hooks
├── styles/
└── types/
```

### 5.3 Why This Frontend Architecture

- **Next.js App Router** provides server-side rendering for fast initial dashboard loads and SEO-friendly public pages without a separate marketing site.
- **Server Components** reduce client-side JavaScript and simplify data access patterns.
- **TanStack Query** handles caching, background refetching, and optimistic updates — critical for live test execution dashboards.
- **Feature-based organization** keeps the codebase aligned with product domains, making it easier to add modules (Analytics, Intelligence) incrementally.

---

## 6. Database Architecture

### 6.1 Database Roles

| Store | Purpose | Technology |
|---|---|---|
| **Primary OLTP** | Users, orgs, workspaces, projects, test cases, suites, defects, permissions, audit events | PostgreSQL |
| **Analytics OLAP** | Test run results, execution metrics, time-series events, failure patterns | ClickHouse |
| **Cache / Session** | Sessions, rate limit counters, job queues, ephemeral cache | Redis |
| **Object Storage** | File attachments, export packages, artifact metadata | S3-compatible |
| **Search Index** | Test case and defect search (MVP uses PG FTS) | PostgreSQL → Meilisearch |

### 6.2 Multi-Tenancy Model

- **Shared database, shared schema** with `tenant_id` column on every tenant-scoped table.
- Row-level security (RLS) policies on PostgreSQL as a defense-in-depth layer.
- Tenant context injected at the HTTP middleware and verified on every query.
- Dedicated schema isolation reserved for enterprise data-residency deployments.

### 6.3 Data Retention Rules

| Data Category | Retention |
|---|---|
| Customer source code / test scripts | **Never stored** (zero customer code retention) |
| API collection definitions (customer API specs) | **Never retained** beyond explicit user-created test definitions inside Testra (zero API collection retention) |
| Test run results | Tier-based retention: 30 days (Free), 90 days (Pro), 1 year (Enterprise), configurable |
| Audit logs | Minimum 7 years for compliance |
| User-uploaded attachments | Tier-based, deletable on account termination |

### 6.4 Why This Database Architecture

- **PostgreSQL + ClickHouse split** separates transactional consistency from high-volume analytics, preventing OLTP performance degradation as test result volume grows.
- **Shared-tenant model** is the simplest operational choice for a solo developer while still supporting enterprise features like audit logs, RBAC, and later schema-isolated data residency.
- **Explicit zero-retention rules** implement Testra's privacy-first positioning directly in data architecture.

---

## 7. Authentication Architecture

### 7.1 Identity Strategy

| Capability | Approach |
|---|---|
| **Primary authentication** | Email + password with bcrypt/Argon2, mandatory MFA for enterprise seats |
| **Enterprise SSO** | SAML 2.0 and OIDC via an identity platform |
| **SCIM provisioning** | Supported via identity platform |
| **CI/CD / API access** | Scoped API keys with tenant-scoped permissions |
| **Passwordless** | Magic links for Free/Pro tiers (optional) |

### 7.2 Identity Provider Choice

**Decision**: Use **Clerk** for core identity (users, sessions, passwordless, basic SSO) and **WorkOS** when enterprise SAML/SCIM deals require it.

**Alternative considered**: Self-hosted Keycloak or FusionAuth. Rejected for MVP because operational overhead (security patching, HA setup, SAML debugging) exceeds a solo developer's bandwidth and delays enterprise-ready launch.

### 7.3 API Key Model

- API keys are issued per workspace or project.
- Keys carry scopes: `results:write`, `results:read`, `tests:read`, `defects:write`, etc.
- Keys are hashed in the database; only the plaintext is shown once at creation.
- Keys are rotated with one-click revocation.

### 7.4 Why This Authentication Architecture

- **Clerk + WorkOS** keeps the solo developer focused on product value instead of identity infrastructure while satisfying MVP enterprise requirements (SSO, MFA, audit).
- **Scoped API keys** allow CI/CD ingestion without exposing user sessions, supporting zero code retention (CI only sends results, never source code).
- **MFA from day one** supports the enterprise-ready principle without custom implementation.

---

## 8. API Architecture

### 8.1 API Style

- **RESTful JSON API** with OpenAPI 3.1 specification as the source of truth.
- **Versioning**: URL-based (`/api/v1/...`). Major breaking changes bump version.
- **Resource-oriented** endpoints aligned with domain modules.
- **Consistent envelope**:
  ```json
  {
    "data": {},
    "meta": {},
    "error": null
  }
  ```

### 8.2 API Categories

| Category | Consumers | Authentication |
|---|---|---|
| **Internal API** | Next.js web app | Session cookie + CSRF |
| **CI/CD Ingestion API** | GitHub Actions, GitLab CI, Jenkins, CircleCI | Scoped API key |
| **Integration API** | Jira, Linear, Slack, Teams | OAuth2 / API key / webhook signatures |
| **Public API** (V3+) | Customers, partners, SDK | OAuth2 client credentials + API key |

### 8.3 Real-Time and Webhooks

- **Server-Sent Events (SSE)** for live test execution progress and notifications.
- **WebSockets** reserved for collaborative editing (future).
- **Outbound webhooks** for CI/CD status, Jira sync, and third-party integrations with HMAC-SHA256 signature verification.

### 8.4 API Governance

- OpenAPI spec is generated from Go handler annotations or hand-maintained as code contract.
- SDKs generated from OpenAPI.
- Automated contract tests run in CI.
- Rate limits enforced per tenant and per API key.

### 8.5 Why This API Architecture

- **REST + OpenAPI** is the most familiar pattern for QA engineers, DevOps, and enterprise integrators.
- **API-first design** ensures the web UI is just another client; CI/CD and future public API use the same endpoints.
- **SSE over WebSockets** for live updates is simpler to scale horizontally and fits test-run progress semantics.

---

## 9. Machine Learning Architecture (No LLMs)

### 9.1 Principle: Transparent, Customer-Owned Intelligence

Testra's intelligence is built from the customer's own historical test data using explainable statistical and classical ML models. No external LLM APIs are used.

### 9.2 ML Service

| Attribute | Decision |
|---|---|
| Runtime | Python 3.12 FastAPI service |
| Libraries | scikit-learn, XGBoost, statsmodels, pandas, numpy |
| Serving | REST API called by Go backend (synchronous) or Asynq worker (asynchronous) |
| Training | Periodic batch jobs per tenant; no centralized training across tenants |
| Feature store | PostgreSQL / ClickHouse tables per tenant |
| Model artifacts | Stored in object storage, versioned, tenant-scoped |

### 9.3 Model Use Cases

| Feature | Technique | Why |
|---|---|---|
| **Flaky Test Detection** | Time-series statistical tests + pass/fail variance scoring | Identifies non-deterministic tests without needing deep learning |
| **Failure Classification** | Rule-based filtering + clustering (DBSCAN/HDBSCAN) + optional XGBoost classifier | Routes failures to code, infra, or test issues; explainable |
| **Risk Scoring** | Feature-based logistic regression / XGBoost with SHAP values | Predicts regression risk per test case; outputs human-readable drivers |
| **Health Score** | Weighted scoring model with configurable weights | Aggregates pass rate, flakiness, coverage, recent change load |
| **Release Readiness** | Threshold + trend model on health and risk metrics | Data-driven go/no-go signal |

### 9.4 Data Isolation and Privacy

- Models are trained per tenant on that tenant's data only.
- No cross-tenant model sharing.
- Model inputs are limited to test metadata and results — never source code, customer API payloads, or secrets.
- All feature pipelines are auditable and documented for enterprise transparency.

### 9.5 Why This ML Architecture

- **Classical ML is sufficient** for Testra's signals and is far cheaper, faster, and more explainable than LLMs.
- **No external AI dependency** satisfies a core product principle and removes procurement risk.
- **Tenant-isolated models** reinforce privacy-first positioning and create data gravity (models improve only on that customer's history).

---

## 10. Deployment Architecture

### 10.1 Deployment Strategy

| Phase | Target | Rationale |
|---|---|---|
| **MVP / Alpha** | Managed platform (Render or Railway) with Docker | Fastest path to design partners |
| **Public Beta / V1.0** | AWS/GCP with managed Kubernetes (EKS/GKE) | Enterprise-ready, SOC 2 path, data residency |
| **Enterprise / V2+** | Multi-region Kubernetes clusters | Data residency (Singapore, Indonesia, etc.) |

### 10.2 Production Topology

```
Internet
   │
Cloudflare / Cloud Provider Load Balancer
   │
Ingress Controller (NGINX / Traefik)
   ├── Next.js Web Pods (standalone)
   ├── Go API Pods (horizontal)
   ├── Worker Pods (separate deployment, scales independently)
   └── ML Service Pods (Python)
   │
PostgreSQL Primary + Read Replicas
ClickHouse Cluster
Redis Cluster
Object Storage
```

### 10.3 Environments

| Environment | Purpose |
|---|---|
| `local` | Docker Compose on developer machine |
| `staging` | Mirror production with sanitized data |
| `production` | Multi-AZ, backed up, monitored |

### 10.4 CI/CD

- GitHub Actions: lint, test, build, integration tests, security scan, deploy to staging, deploy to production.
- Trunk-based development with feature flags.
- Immutable container images tagged per commit SHA.
- Database migrations applied via Flyway or golang-migrate in CI, never manually.

### 10.5 Why This Deployment Architecture

- **Managed platform for MVP** preserves cash and engineering focus while validating product-market fit.
- **Kubernetes for production** provides the observability, scaling, security, and audit posture enterprises require.
- **Separate worker and ML pods** let compute scale independently: API pods handle user traffic while workers handle result ingestion and model inference.
- **Terraform** enables repeatable, reviewable infrastructure and supports future multi-region expansion.

---

## 11. Recommended Folder Structure

The final repository folder structure is designed for the monorepo decision and Clean Architecture backend:

```
testra/
├── apps/
│   ├── api/
│   │   ├── cmd/                    # Application entrypoints
│   │   │   ├── api/
│   │   │   ├── worker/
│   │   │   └── migrator/
│   │   ├── internal/
│   │   │   ├── identity/
│   │   │   ├── organization/
│   │   │   ├── workspace/
│   │   │   ├── project/
│   │   │   ├── testmanagement/
│   │   │   ├── apitesting/
│   │   │   ├── automationhub/
│   │   │   ├── results/
│   │   │   ├── defects/
│   │   │   ├── analytics/
│   │   │   ├── intelligence/
│   │   │   ├── notification/
│   │   │   ├── integrationhub/
│   │   │   ├── audit/
│   │   │   ├── billing/
│   │   │   └── shared/             # middleware, errors, tracing, config
│   │   ├── pkg/                    # public-ish shared packages
│   │   ├── migrations/
│   │   ├── configs/
│   │   └── tests/
│   ├── web/
│   │   ├── app/
│   │   ├── features/
│   │   ├── components/
│   │   ├── lib/
│   │   └── types/
│   ├── worker/
│   │   └── ...                     # Go worker entrypoint if separated
│   └── ml/
│       ├── api/
│       ├── models/
│       ├── features/
│       ├── training/
│       └── tests/
├── packages/
│   ├── shared/
│   ├── ui/
│   ├── config/
│   └── sdk/
├── infra/
│   ├── terraform/
│   ├── k8s/
│   └── docker/
├── docs/
│   ├── api/
│   ├── architecture/
│   └── runbooks/
├── scripts/
├── Makefile
├── pnpm-workspace.yaml
├── go.work
└── README.md
```

---

## 12. Why These Decisions Are Best for Testra

### 12.1 Alignment with Constraints

| Constraint | Decision Response |
|---|---|
| **Solo developer** | Monorepo, Go backend, Clean Architecture, managed identity (Clerk/WorkOS), Docker Compose local dev |
| **Budget-conscious** | Open-source core stack (Go, PostgreSQL, Redis, ClickHouse, scikit-learn), managed platform for MVP, right-sized Kubernetes later |
| **Scalable** | Modular monolith with clear extraction boundaries, ClickHouse for high ingest, horizontal pod scaling, read replicas |
| **Enterprise-ready** | SSO/SAML/SCIM, RBAC, audit logs, SOC 2-aligned deployment, encryption, data residency path |
| **API-first** | OpenAPI-driven REST API, CI ingestion and web UI share endpoints, SDK generated from spec |
| **Privacy-first** | Tenant isolation, per-tenant ML models, scoped API keys, explicit data retention rules |
| **Zero customer code retention** | CI ingestion accepts only results/metadata; no repository or source-code storage |
| **Zero API collection retention** | Customer API collections are not retained unless explicitly created as Testra test definitions by users |
| **No external LLM dependency** | Classical ML only (scikit-learn/XGBoost), no OpenAI or third-party LLM APIs |

### 12.2 Strategic Fit

- **APAC-first, global-ready**: PostgreSQL/ClickHouse/Redis are deployable in every major region; Terraform supports multi-region expansion.
- **QA Engineer first**: Next.js delivers a fast, modern UX; SSE gives live test progress; TanStack Query keeps dashboards fresh.
- **Enterprise moat**: Audit trail, RBAC, SSO, data residency, and transparent ML create defensible differentiation against point solutions.
- **Modular future**: Clean Architecture boundaries make it possible to extract hot modules (e.g., ML service, ingestion pipeline) into independent services when team and revenue justify it.

### 12.3 Trade-offs Acknowledged

| Decision | Trade-off |
|---|---|
| Modular monolith vs microservices | Slightly harder to scale ingestion independently; mitigated by separate worker/ML pods |
| Clerk/WorkOS vs self-hosted IdP | Adds vendor cost; avoids identity engineering risk and accelerates enterprise readiness |
| ClickHouse added complexity | Requires operational learning; relational scaling for test results would become cost-prohibitive |
| Go for backend | Smaller ML ecosystem; addressed by Python microservice for ML only |

---

## Approval Status

| Role | Name | Status | Date |
|---|---|---|---|
| Chief Software Architect / Engineering Lead | | Draft | July 2026 |
| CTO / Engineering Lead | | Pending Approval | |
| Product Lead | | Pending Approval | |

**Next Step**: Await executive approval. Upon approval, implementation begins with repository setup, local development environment, and the Identity + Organization domains.
