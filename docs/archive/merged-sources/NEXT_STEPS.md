# Next Steps

> This is the prioritized engineering roadmap derived from [`CURRENT_STATE.md`](CURRENT_STATE.md) and [`TECHNICAL_DEBT.md`](TECHNICAL_DEBT.md). The sequence is designed to make the platform safe, usable, and complete before new feature work begins.

## Priority overview

| Priority | Focus | Why it exists |
|----------|-------|---------------|
| **P0** | Production hardening + security | Nothing else matters if the platform is not secure, available, and compliant. |
| **P1** | Frontend foundation + authentication reliability | The current frontend loses sessions, has broken SSE, and exposes unauthenticated UI flashes. |
| **P2** | Core MVP completion: Defects + Notifications | Finishes the day-to-day QA workflow (test → defect → alert) and unlocks value. |
| **P3** | API Testing module | Displaces Postman for QA-led API testing; core product differentiator. |
| **P4** | Integration Hub + CI/CD integrations | Automates result ingestion and defect sync, driving retention and enterprise stickiness. |
| **P5** | Real dashboards + Reporting | Replaces the dashboard skeleton with actionable quality signals and reports. |
| **P6** | Analytics + Intelligence (V2) | Builds the data-driven quality moat: flaky tests, failure classification, risk scoring. |
| **P7** | Enterprise hardening (SSO, compliance, data residency) | Unlocks enterprise deals and regulated verticals. |
| **P8** | Public API + Marketplace (V3) | Turns Testra from a product into a platform and ecosystem. |

---

## P0 — Production hardening and security

**Goal:** Make the platform safe to deploy and operate.

1. **Wire rate limiting** on `/auth/*` and `/ingest` using Redis-backed token buckets.
2. **Implement API-key authentication middleware** so CI/CD runners can use scoped keys instead of user JWTs.
3. **Fix SSE authentication** for live test-run progress (session cookie or signed query token).
4. **Move secrets out of `.env.example`** and add startup validation; integrate with a secrets manager for production.
5. **Make audit logging durable** — write inside the request transaction or enqueue to a reliable worker queue.
6. **Add frontend route guards** and stop storing the access token in `localStorage` (move to `httpOnly` cookie or a secure wrapper).

**Why this is P0:** These items are prerequisites for any production deployment, SOC 2 readiness, and enterprise sales. Without them, Testra is not a credible B2B SaaS product.

---

## P1 — Frontend foundation and auth reliability

**Goal:** Make the web app reliable and pleasant to use.

1. **Add a global auth state layer** (Zustand or React Context) with token refresh on 401.
2. **Implement a fetch interceptor** that calls `/auth/refresh` and retries once, then redirects to `/login` on failure.
3. **Fix MFA QR rendering** (`<img src={qr_code} />`) and add copy-to-clipboard fallback.
4. **Fix project key generation** to match backend validation and add a unique constraint.
5. **Consolidate dashboard route trees** to `/[workspace]` and redirect `/dashboard` to the selected workspace.
6. **Add global `loading.tsx`, `error.tsx`, and `not-found.tsx`** in the dashboard group.
7. **Add structured logging and health/readiness endpoints** to the Go API.

**Why this is P1:** The current frontend has broken flows (SSE, 15-minute logout, unauthenticated UI flashes) that prevent daily use. Fixing auth and navigation unblocks every subsequent feature.

---

## P2 — Core MVP completion: Defects + Notifications

**Goal:** Complete the core QA execution loop.

1. **Notifications module (backend + frontend):** ✅ Completed.
   - `apps/api/internal/notification/` with in-app notifications, preferences, and email/Slack/Teams/webhook channels.
   - `apps/web/dashboard/notifications/` list page and `dashboard/settings/notifications/` preferences page.
   - Sidebar bell with unread count badge.
2. **Defects module (backend + frontend):** Next priority.
   - `apps/api/internal/defects/` with CRUD, lifecycle, links to test run items, and severity/priority fields.
   - `apps/web/[workspace]/defects/` list, create, detail, and edit pages.
3. **Jira/Linear/GitHub Issues integration design:** at minimum design the outbound webhook schema and queue job for defect sync.

**Why this is P2:** Test execution without defect tracking and alerting is incomplete. These two modules close the manual testing workflow and are table stakes for any QA platform.

---

## P3 — API Testing module

**Goal:** Build a native API testing experience that can replace Postman for QA teams.

1. **Backend (`apps/api/internal/apitesting/`):**
   - API request definitions (method, URL, headers, body, assertions).
   - Environment and variable scoping per project/workspace.
   - Execution engine using Go `net/http` with result capture.
   - Store request/response history in `test_runs`/`test_run_items` or a dedicated table.
2. **Frontend (`apps/web/features/api-testing/`):**
   - Collection/folder tree, request editor, environment selector, response viewer, and run history.

**Why this is P3:** API testing is a core differentiator and one of the top displacements (Postman). It depends on P0/P1 stability but can be built in parallel with P2 once the platform is reliable.

---

## P4 — Integration Hub + CI/CD integrations

**Goal:** Automate the flow of results and defects into Testra.

1. **Integration Hub backend (`apps/api/internal/integrationhub/`):**
   - Webhook receivers for GitHub Actions, GitLab CI, Jenkins, and CircleCI.
   - Outbound webhooks for Jira/Linear/GitHub issue sync with HMAC-SHA256 signatures.
   - Credential storage (OAuth tokens, API keys) per workspace.
2. **Frontend settings pages for integrations.**
3. **Document and publish `/ingest` contract** for CI plugins and update OpenAPI.

**Why this is P4:** CI/CD result ingestion is a major retention driver (automation-heavy teams see value immediately) and removes the biggest objection from enterprise buyers who already have Jira.

---

## P5 — Real dashboards + reporting

**Goal:** Replace the dashboard skeleton with actionable quality signals.

1. **Backend (`apps/api/internal/analytics/`):**
   - Aggregations: pass/fail rate, run count, open defects by severity, recent runs, top failing tests.
2. **Frontend (`apps/web/features/analytics/`):**
   - Role-based dashboard widgets (QA engineer, QA lead, engineering manager).
   - Run history, test case coverage, and project health summary.
3. **Reports:** PDF/CSV export for run summaries and traceability matrices.

**Why this is P5:** Dashboards and reports are the "aha moment" for leadership buyers. They also create the data foundation for the intelligence layer in P6.

---

## P6 — Analytics + Intelligence (Version 2.0)

**Goal:** Deliver data-driven quality insights that competitors cannot easily copy.

1. **Flaky test detection** using pass/fail variance over time.
2. **Failure classification** (environment, test data, product defect, infrastructure) via rules and clustering.
3. **Risk scoring** and **test suite health scores** with human-readable explanations.
4. **Release readiness report** aggregating coverage, flakiness, open defects, and recent failure trends.
5. **Wire `apps/ml` Python service** for model training/inference and `apps/worker` for background jobs.
6. **Adopt ClickHouse** for high-volume result analytics.

**Why this is P6:** Intelligence features require enough historical data to be credible. Building them after P2–P5 ensures the data pipeline and storage choices are ready.

---

## P7 — Enterprise hardening

**Goal:** Make Testra sellable to regulated, large organizations.

1. **SSO / SAML 2.0 and OIDC** integration.
2. **SCIM provisioning** for user lifecycle management.
3. **Workspace/project-level RBAC** and custom roles.
4. **Advanced audit export** and compliance report templates.
5. **Data residency** options (Singapore, Indonesia) and multi-region deployment.
6. **SOC 2 Type II evidence** collection and security documentation.

**Why this is P7:** These are enterprise hard requirements that unlock $50K+ ACV deals. They are unnecessary for the initial mid-market self-serve motion but critical for the long-term moat.

---

## P8 — Public API + Marketplace (Version 3.0)

**Goal:** Turn Testra into a platform.

1. **Public API (`/api/v1` stabilization + versioning strategy).**
2. **Official TypeScript SDK** generated from OpenAPI in `packages/sdk`.
3. **Partner Marketplace** for test integrations, custom report templates, and notification channels.
4. **Predictive analytics** and cross-project governance dashboards.

**Why this is P8:** Public API and marketplace create ecosystem lock-in and partner revenue. They require a mature, stable, and well-documented core product first.

---

## How to use this roadmap

- **Start at P0 and do not skip.** Production hardening is non-negotiable.
- **Within each priority, tackle foundational blockers first** (e.g., auth state before notifications UI).
- **Track progress in [`FEATURE_MATRIX.md`](FEATURE_MATRIX.md).** When a feature becomes `Production Ready`, update the matrix and close the related item in [`TECHNICAL_DEBT.md`](TECHNICAL_DEBT.md).
- **Re-evaluate after each milestone.** Customer feedback from P1–P3 may shift the order of P4–P8, but P0 must remain complete before launch.
