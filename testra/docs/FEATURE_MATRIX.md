# Feature Matrix

**Purpose:** Track implementation status of every Testra feature across backend, frontend, OpenAPI, tests, and production readiness.
**Owner:** Engineering / Product Lead
**Scope:** Feature completion matrix and status legend.
**Source of Truth:** FEATURE_MATRIX.md for feature status; ROADMAP.md for phase plan; BIBLICAL_TESTRA.md for module list.
**Last Updated:** July 2026
**Related documents:**
- [`BIBLICAL_TESTRA.md`](BIBLICAL_TESTRA.md)
- [`ROADMAP.md`](engineering/ROADMAP.md)
- [`PROJECT_OVERVIEW.md`](PROJECT_OVERVIEW.md)

> This matrix covers every Testra module and its current implementation status across **Backend**, **Frontend**, **OpenAPI**, **Tests**, and **Production Ready**.

### Legend

| Symbol | Meaning |
|--------|---------|
| ✅ | Implemented / present |
| 🔄 | Partial / in progress / has known issues |
| ❌ | Not implemented / missing |
| N/A | Not applicable to that column |

---

## Platform layer

| Feature | Backend | Frontend | OpenAPI | Tests | Production Ready | Status |
|---------|---------|----------|---------|-------|------------------|--------|
| Identity / Auth (register, login, refresh, me) | ✅ | ✅ | ✅ | 🔄 | ❌ | Functional; needs rate limiting and token refresh on client |
| MFA TOTP (setup, verify, disable) | ✅ | ✅ | ✅ | 🔄 | ❌ | QR code rendered as `<img>` from backend data URL |
| Password reset (request, confirm) | ✅ | ✅ | ✅ | 🔄 | ❌ | Token emailed; no fallback if SMTP disabled |
| Organization management | ✅ | ✅ | ✅ | 🔄 | ❌ | POST/GET bypass tenant/permission gates |
| Workspace management | ✅ | ✅ | ✅ | 🔄 | ❌ | Functional |
| Project management | ✅ | ✅ | ✅ | 🔄 | ❌ | Frontend key generation now matches backend regex |
| API keys (CRUD) | ✅ | ✅ | ✅ | 🔄 | ❌ | Settings UI implemented; not used for `/ingest` auth yet |
| RBAC (roles, permissions, assignments) | 🔄 | ❌ | 🔄 | 🔄 | ❌ | Org-scoped only; permission-name drift |
| Audit logging | ✅ | ❌ | ❌ | 🔄 | ❌ | Fire-and-forget, no UI |
| Billing / subscriptions | ❌ | ❌ | ❌ | ❌ | ❌ | Not started |

## Testing layer

| Feature | Backend | Frontend | OpenAPI | Tests | Production Ready | Status |
|---------|---------|----------|---------|-------|------------------|--------|
| Test folders | ✅ | ❌ | ✅ | 🔄 | ❌ | Backend only |
| Test suites | ✅ | ❌ | ✅ | 🔄 | ❌ | Backend only |
| Test cases (CRUD, steps, tags, search) | ✅ | 🔄 | ✅ | 🔄 | ❌ | List/create/edit/versions UI exist; no suite/folder mgmt |
| Test case versioning | ✅ | 🔄 | ✅ | 🔄 | ❌ | History list UI exists |
| Test runs / results | ✅ | ✅ | ✅ | 🔄 | ❌ | Manual runs work; SSE uses query-token auth in browsers |
| Test run progress (SSE) | ✅ | ✅ | ✅ | 🔄 | ❌ | `Auth` middleware accepts `Authorization` header or `access_token` query param |
| Automation result ingestion (JUnit/Playwright/Cypress) | ✅ | ❌ | ✅ | 🔄 | ❌ | No UI; requires user JWT (API-key auth missing) |
| API testing engine | ❌ | ❌ | ❌ | ❌ | ❌ | Not started |
| Defects | ❌ | 🔄 | ❌ | ❌ | ❌ | Frontend placeholder only |
| Manual test execution tracker | 🔄 | 🔄 | 🔄 | 🔄 | ❌ | Runs can be created and started; no step-level UI |

## Core layer

| Feature | Backend | Frontend | OpenAPI | Tests | Production Ready | Status |
|---------|---------|----------|---------|-------|------------------|--------|
| Dashboard | 🔄 | 🔄 | ❌ | ❌ | ❌ | Skeleton with quick links; no real widgets |
| Settings shell + navigation | N/A | ✅ | ❌ | ❌ | ❌ | Shell exists; notifications and API keys pages implemented; most other tabs are placeholders |
| Settings — members | ❌ | ❌ | ❌ | ❌ | ❌ | Not started |
| Settings — roles | ❌ | ❌ | ❌ | ❌ | ❌ | Not started |
| Settings — API keys | ✅ | ✅ | ✅ | 🔄 | ❌ | Backend CRUD and `/dashboard/settings/api-keys` UI implemented |
| Settings — audit logs | ✅ | ❌ | ❌ | 🔄 | ❌ | Backend stores events; no UI |
| Settings — billing | ❌ | ❌ | ❌ | ❌ | ❌ | Not started |
| Settings — notifications | ✅ | ✅ | ✅ | 🔄 | ❌ | In-app feed, preferences, channels; needs production hardening |
| Settings — profile / security | 🔄 | 🔄 | ❌ | ❌ | ❌ | Placeholder pages |
| Settings — organization / workspace | 🔄 | 🔄 | 🔄 | ❌ | ❌ | Basic create only |

## Intelligence & analytics layer

| Feature | Backend | Frontend | OpenAPI | Tests | Production Ready | Status |
|---------|---------|----------|---------|-------|------------------|--------|
| Analytics / dashboards | ❌ | ❌ | ❌ | ❌ | ❌ | Not started |
| Flaky test detection | ❌ | ❌ | ❌ | ❌ | ❌ | V2 |
| Failure classification | ❌ | ❌ | ❌ | ❌ | ❌ | V2 |
| Risk scoring | ❌ | ❌ | ❌ | ❌ | ❌ | V2 |
| Coverage heatmap | ❌ | ❌ | ❌ | ❌ | ❌ | V2 |
| Release readiness | ❌ | ❌ | ❌ | ❌ | ❌ | V2 |
| ML inference service | 🔄 | ❌ | ❌ | ❌ | ❌ | `apps/ml/api/main.py` has a `/health` endpoint skeleton; no inference yet |

## Enterprise & ecosystem layer

| Feature | Backend | Frontend | OpenAPI | Tests | Production Ready | Status |
|---------|---------|----------|---------|-------|------------------|--------|
| SSO / SAML | ❌ | ❌ | ❌ | ❌ | ❌ | Not started |
| SCIM provisioning | ❌ | ❌ | ❌ | ❌ | ❌ | V2+ |
| Data residency | ❌ | ❌ | ❌ | ❌ | ❌ | Enterprise tier |
| Compliance modules / audit export | ❌ | ❌ | ❌ | ❌ | ❌ | Enterprise tier |
| Integration Hub (Jira, GitHub, GitLab, Slack) | ❌ | ❌ | ❌ | ❌ | ❌ | Not started |
| CI/CD integrations (GitHub Actions, GitLab CI, Jenkins) | 🔄 | ❌ | ❌ | ❌ | ❌ | Ingest endpoint exists; no native plugins |
| Public API / SDK | 🔄 | N/A | 🔄 | ❌ | ❌ | `/api/v1` exists; SDK not built |
| Marketplace / plugins | ❌ | ❌ | ❌ | ❌ | ❌ | V3 |

## Infrastructure & operations

| Feature | Backend | Frontend | OpenAPI | Tests | Production Ready | Status |
|---------|---------|----------|---------|-------|------------------|--------|
| Local development (pnpm dev native, Docker Compose for deps) | N/A | N/A | N/A | N/A | ✅ | Functional; native dev is default, Docker optional per ADR-009 |
| Dockerfiles | N/A | N/A | N/A | N/A | 🔄 | Multi-stage builds; `.dockerignore` not verified |
| Kubernetes manifests | N/A | N/A | N/A | N/A | ❌ | Base deployment + service only |
| Terraform modules | N/A | N/A | N/A | N/A | ❌ | Scaffold only; modules empty |
| CI pipeline (lint/build/test) | N/A | N/A | N/A | N/A | 🔄 | GitHub Actions builds Go/web/ML; no integration tests |
| CD pipeline (deploy) | N/A | N/A | N/A | N/A | ❌ | Not implemented |
| Observability (logs/metrics/traces) | N/A | N/A | N/A | N/A | ❌ | Not implemented |
| Secrets management | N/A | N/A | N/A | N/A | ❌ | Not implemented |

---

## Summary

- **Fully implemented end-to-end:** None. Even the most complete flows (Identity, Test Case Management) lack production hardening.
- **Backend functional, frontend missing:** API Keys, Test Folders, Test Suites, Automation Ingestion, Audit.
- **Frontend functional, backend partial:** Dashboard, Settings shell.
- **Not started:** Defects, API Testing, Billing, Analytics, Intelligence, Integration Hub, SSO, Marketplace, Public SDK. **Notifications are implemented.**
- **Blockers for production:** rate limiting, API-key auth, SSE auth, route guards, secrets management, K8s/Terraform completion, and deployment pipeline.


## See Also

- For the narrative current-state audit, P0/P1/P2 issues, and recommended next actions, see [`engineering/ROADMAP.md`](engineering/ROADMAP.md) and [`PROJECT_OVERVIEW.md`](PROJECT_OVERVIEW.md).
- For the canonical dependency graph and request lifecycle, see [`BIBLICAL_TESTRA.md`](BIBLICAL_TESTRA.md).
- For backend route and permission details, see [`ROUTES.md`](ROUTES.md).
