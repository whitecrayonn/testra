# Testra Sprint Backlog — Launch Readiness

**Date:** 2026-08-02  
**Owner:** Engineering Lead / Engineering Manager  
**Source of truth:** `docs/engineering/SPRINT_BACKLOG.md`  
**Scope:** 112 actionable implementation tasks derived from the validated CTO Audit findings and the production roadmap.  
**Legend:** Effort = `XS` / `S` / `M` / `L`. Priority = `P0` / `P1` / `P2` / `P3`.

---

## M1 — Production Security & Trust

> **Sprint 1–3 status (2026-08-03 audit + implementation pass):** SBL-001, 002, 003,
> 005, 006, 008, 009, 010, 012 were found already implemented and verified in code
> (cookie/CSRF auth, apiFetch migration, password policy incl. breach blocklist,
> API-key scope registry, refresh-token revocation ordering, rate-limiter
> fail-closed config, log/audit redaction, secrets provider abstraction) — no
> changes needed. SBL-004 (jti + access-token denylist), SBL-007 (audit-events
> read endpoint, scoped to the current user — see the note below), and SBL-013
> (RBAC integration tests: cross-tenant denial + permission grant/revoke) were
> genuine gaps and have now been implemented, with unit/integration test
> coverage. SBL-011 stays open (blocked on SBL-025/VM in M2). SBL-014–024 are
> untouched and are the real remaining Sprint 4 work — see `SPRINT_PLAN.md`.
>
> SBL-007 note: `audit_events` has no `organization_id` column yet (SBL-080),
> so the read endpoint is scoped to the requesting user's own events rather
> than an org-wide view, to avoid a cross-tenant data leak. An org-wide view
> for owners/admins is future work once SBL-080 lands.

| ID | Title | Priority | Effort | Affected Modules | Expected Deliverables | Validation | Rollback Considerations | Dependencies |
|----|-------|----------|--------|------------------|---------------------|------------|-------------------------|--------------|
| SBL-001 | Implement httpOnly cookie/session auth backend | P0 | L | `apps/api/internal/identity`, `apps/api/internal/shared/server`, `apps/api/internal/shared/middleware` | Cookie auth endpoints, session store schema, middleware attaching user context | `go test ./apps/api/internal/identity/...`; integration tests; login/logout E2E | Revert middleware order; keep JWT bearer fallback | — |
| SBL-002 | Add CSRF token endpoint and middleware | P0 | M | `apps/api/internal/shared/middleware`, `apps/web/lib/api.ts` | `GET /auth/csrf`, double-submit cookie validation | Unit tests; 403 on missing token for mutating requests | Disable CSRF via feature flag if cookie-only | SBL-001 |
| SBL-003 | Migrate `apiFetch` from `localStorage` to cookie-based auth | P0 | M | `apps/web/lib/api.ts`, `apps/web/app/(auth)/*`, `apps/web/components/auth/route-guard.tsx` | Remove `localStorage` token reads; rely on cookies; refresh flow preserved | `pnpm build`; `pnpm test`; manual login/logout flows | Revert to `localStorage` fallback commit | SBL-001, SBL-002 |
| SBL-004 | Add `jti` claim and short-lived access-token denylist | P1 | M | `apps/api/internal/shared/jwt`, `apps/api/internal/identity` | JWT manager emits `jti`; denylist table + check | Unit tests; revoked token returns 401 | Clear denylist table; old tokens without `jti` remain valid during grace period | — |
| SBL-005 | Harden password policy + breached-password check | P1 | M | `apps/api/internal/identity/service.go`, `apps/api/internal/shared/validation` | Complexity rules; Have I Been Pwned / common-word check | Unit tests; registration rejects weak/breached passwords | Relax validation constant; no schema changes | — |
| SBL-006 | Add API-key scope registry validation | P1 | S | `apps/api/internal/apikeys/service.go` | Allowed-scope registry; reject unknown scopes at create time | `go test ./apps/api/internal/apikeys/...` | Remove registry check | — |
| SBL-007 | Add audit log read endpoint and UI | P1 | M | `apps/api/internal/audit`, `apps/web/app/(dashboard)/settings/audit-logs` | `GET /audit-events` paginated; admin UI page | Unit/integration tests; UI renders events | Disable route; revert UI | SBL-001 |
| SBL-008 | Fix refresh-token revocation ordering | P1 | S | `apps/api/internal/identity/service.go` | Revoke old refresh token before issuing new | `go test`; no regression in refresh flow | Swap ordering back | — |
| SBL-009 | Add rate-limiter fail-closed fallback for auth endpoints | P1 | M | `apps/api/internal/shared/middleware/ratelimit.go` | In-memory fallback when Redis unavailable; fail-closed config | Tests simulate Redis down; auth endpoints still protected | Switch to fail-open mode via env | — |
| SBL-010 | Add PII redaction in request/audit logs | P1 | M | `apps/api/internal/shared/server/server.go`, `apps/api/internal/shared/middleware/audit.go` | Sanitize `RemoteAddr`, bodies, query params for email/token/password | Log inspection; unit tests | Disable redaction via log level | — |
| SBL-011 | Add production VM host firewall rules | P1 | M | VM host firewall config (not in repo yet; part of `scripts/ops/provision.sh` once written) | Default-deny inbound with allow-listed ports (SSH/RDP, 80, 443) and restricted egress | Manual firewall audit; network connectivity tests | Reset to allow-all (not recommended) | SBL-025 (VM provisioning) |
| SBL-012 | Implement secrets-manager provider abstraction | P1 | M | `apps/api/internal/shared/secrets`, `apps/api/internal/shared/config` | A Vault (or similar) implementation of `secrets.Provider`, pluggable alongside the existing `EnvProvider` | Unit tests; config loads secret from provider | Revert to env-var provider | — |
| SBL-013 | Add RBAC end-to-end integration tests | P1 | M | `apps/api/tests/integration` | Tests for cross-tenant denial, permission grant/revoke | `go test ./apps/api/tests/integration/...` | Remove tests if env missing | SBL-001 |
| SBL-014 | Add API-key auth regression tests | P2 | S | `apps/api/internal/shared/middleware`, `apps/api/tests/integration` | Tests for `X-API-Key`, `Authorization: ApiKey`, invalid/revoked keys | `go test` | Remove new test files | SBL-006 |
| SBL-015 | Implement magic-link password reset | P2 | M | `apps/api/internal/identity`, `apps/web/app/(auth)/reset-password` | Generate HTTPS magic link; replace raw token in email | E2E password reset | Keep legacy raw-token endpoint behind flag | — |
| SBL-016 | Add compiled binary scanning and SBOM to CI | P2 | M | `.github/workflows/ci.yml` | Trivy/Grype scan of compiled Go binaries and dependency manifests (`go.sum`, `pnpm-lock.yaml`); SBOM artifact | CI passes with no critical CVEs | Remove CI step | — |
| SBL-017 | Add SSRF DNS cache and timeout | P2 | S | `apps/api/internal/shared/security/ssrf.go` | Cache resolved IPs per host; context timeout for lookups | Unit tests | Disable cache via env | — |
| SBL-018 | Add security headers validation tests | P2 | S | `apps/api/internal/shared/server/server_test.go`, `apps/web/__tests__` | Assert `Cache-Control`, CSP, HSTS, etc. | Tests pass | Remove test cases | SBL-003 |
| SBL-019 | Add audit export endpoint | P2 | M | `apps/api/internal/audit` | `GET /audit-events/export` (CSV/JSON) | Integration tests | Disable route | SBL-007 |
| SBL-020 | Add `X-Content-Type-Options`, `Referrer-Policy`, `Permissions-Policy` to API responses | P2 | XS | `apps/api/internal/shared/server/server.go` | Response header additions | Header check tests | Revert header set | — |
| SBL-021 | Add webhook/ingestion signature verification | P2 | M | `apps/api/internal/automationhub`, `apps/api/internal/integrationhub` | HMAC signature validation for incoming webhooks | Unit/integration tests | Allow unsigned via config | — |
| SBL-022 | Implement account lockout on repeated failed logins | P2 | M | `apps/api/internal/identity` | Lock account after N failed attempts within window | Unit tests | Disable lockout via config | SBL-001 |
| SBL-023 | Add session termination on password change | P2 | S | `apps/api/internal/identity` | Invalidate refresh-token family on password reset | Unit tests | Keep existing behavior | SBL-008 |
| SBL-024 | Add CSP report-uri and monitor violations | P3 | S | `apps/web/middleware.ts` | `report-uri` pointing to internal endpoint | Receiving CSP reports | Remove directive | SBL-003 |

## M2 — Production Infrastructure & Deploy

> **Rewritten 2026-08-xx.** This milestone previously described building Terraform-managed
> AWS infrastructure (VPC/EKS/RDS/S3/CloudFront/Route53) and Kubernetes manifests. That
> content was left over from before ADR-003 closed the decision against Docker, Kubernetes,
> and Terraform for MVP. The items below match the actual plan: a single rented VM (target
> OS undecided until rental time) running supervised native processes. See
> `docs/deployment/DEPLOYMENT_GUIDE.md`, which is the source of truth these items implement.

| ID | Title | Priority | Effort | Affected Modules | Expected Deliverables | Validation | Rollback Considerations | Dependencies |
|----|-------|----------|--------|------------------|---------------------|------------|-------------------------|--------------|
| SBL-025 | Rent the production VM and choose its OS | P0 | S | (hosting provider; no repo path) | VM rented; SSH/RDP access configured; firewall allows only SSH/RDP and 80/443 | VM reachable; matches `DEPLOYMENT_GUIDE.md` "One-time server setup" step 1 | Terminate VM, re-provision | — |
| SBL-026 | Install and configure PostgreSQL on the VM | P0 | S | (VM setup; no repo path yet) | `testra` database and a dedicated user with a strong password | `psql` connects as the app user; RLS-enforced queries succeed | Drop database, re-provision | SBL-025 |
| SBL-027 | Install a reverse proxy and obtain a TLS certificate | P0 | M | (VM setup; no repo path yet) | nginx/Caddy/IIS installed; Let's Encrypt or platform-equivalent ACME certificate, auto-renewing | `curl https://<domain>/health` returns 200 over TLS | Remove reverse-proxy config | SBL-025 |
| SBL-028 | Write and commit VM provisioning scripts | P0 | M | `scripts/ops/provision.*` (new) | Scripted, repeatable setup (PostgreSQL, reverse proxy, app user, directories) so VM state is reproducible from source instead of manual steps | Fresh VM provisioned from script matches the manual runbook | Revert to the manual runbook | SBL-025, SBL-026, SBL-027 |
| SBL-029 | Write process-supervisor unit files for API, worker, and web | P0 | M | `scripts/ops/` (new; e.g. systemd units or the chosen OS's equivalent) | Supervised-process definitions with auto-restart for each binary | Services survive a reboot and a process crash | Remove unit files, run manually | SBL-025 |
| SBL-030 | Write the build-and-deploy script | P0 | M | `scripts/ops/deploy.*` (new) | Cross-compiles binaries (`GOOS`/`GOARCH` for the VM), copies artifacts, runs the migrator, restarts services — per `DEPLOYMENT_GUIDE.md` "Each deployment" | Dry run succeeds against a test VM | Revert to a fully manual deployment | SBL-028, SBL-029 |
| SBL-031 | Add automatic pre-deploy database backup | P0 | S | `scripts/ops/deploy.*` | Deploy script backs up the database before running migrations, per `DEPLOYMENT_GUIDE.md` step 3 | Backup produced and verified restorable before every deploy | Remove the backup step (not recommended) | SBL-030, SBL-036 |
| SBL-032 | Add a CD workflow that ships built artifacts to the VM | P0 | L | `.github/workflows/cd.yml` | On merge to `main`, CI builds binaries and the web app, then runs `scripts/ops/deploy.*` against the VM (e.g., over SSH) | Successful automated deploy to staging | Disable workflow, deploy manually | SBL-030 |
| SBL-033 | Add Trivy/Grype image scan in CI | P2 | S | `.github/workflows/ci.yml` | Scan build outputs and dependency manifests in PRs | CI passes with no critical CVEs | Remove step | SBL-016 |
| SBL-034 | Add environment files for staging/production with real origins | P0 | M | VM environment file (not in the repo) | Per-environment `CORS_ALLOWED_ORIGINS`, `ML_SERVICE_URL`, `NEXT_PUBLIC_API_URL`, and other values, in a file readable only by the app user | Config validation passes (`shared/config/config.go`); app starts with production values | Revert to local `.env` | SBL-029 |
| SBL-035 | Add a pluggable secrets-store backend (optional, post-MVP) | P2 | M | `apps/api/internal/shared/secrets` | A Vault (or similar) implementation of `secrets.Provider`, for when a plain environment file is no longer enough | Unit tests; config can load a secret from the new provider | Revert to `EnvProvider` | SBL-012 |
| SBL-036 | Implement backup/restore runbook and scripts | P1 | L | `docs/operations/DISASTER_RECOVERY_GUIDE.md`, `scripts/ops/` | PITR scripts, restore verification, off-VM copy | Tested restore to a temporary instance | Document rollback to a snapshot | SBL-026 |
| SBL-037 | Configure DNS for the production domain | P1 | S | (DNS provider; no repo path) | A/AAAA (or CNAME) record pointing the domain at the VM, matching the TLS certificate | DNS resolves; TLS valid | Revert DNS records | SBL-027 |
| SBL-038 | Add basic rate-limiting/filtering at the reverse-proxy layer (optional) | P2 | M | reverse-proxy config (not in the repo yet) | Rate-based rules and common SQLi/XSS request filtering at the proxy, or a CDN/WAF in front of it if load-tested need justifies it | Load test; malicious payloads blocked | Disable rules | SBL-027 |
| SBL-039 | Add CI integration test job with Postgres/Redis services | P1 | M | `.github/workflows/ci.yml`, `apps/api/tests/integration` | Integration tests run on every PR | CI job green | Remove services job | SBL-013 |
| SBL-040 | Add Dependabot/Snyk config | P2 | S | `.github/dependabot.yml` | Weekly dependency update PRs | Config valid | Remove file | — |
| SBL-041 | Add environment promotion workflow | P1 | M | `.github/workflows/promote.yml` | staging → production promotion with an approval gate | Dry-run promotion | Disable workflow | SBL-032 |
| SBL-042 | Add health-check based restart for worker and web processes | P1 | S | `scripts/ops/` | Process supervisor restarts worker/web on `/health` failure or crash, with a startup grace period | Simulated crash triggers restart within N seconds | Remove the restart policy | SBL-029 |
| SBL-043 | ~~Add pod topology spread constraints~~ — **retired.** No pod scheduler exists on a single VM; there is nothing to spread across. See `DEPLOYMENT_GUIDE.md` "Scaling Beyond One VM" for what actually gets revisited if a second VM is ever added. | — | — | — | — | — | — | SBL-030 |

## M3 — Observability & Reliability

| ID | Title | Priority | Effort | Affected Modules | Expected Deliverables | Validation | Rollback Considerations | Dependencies |
|----|-------|----------|--------|------------------|---------------------|------------|-------------------------|--------------|
| SBL-044 | Instrument API with OpenTelemetry traces | P0 | L | `apps/api/internal/shared/server/server.go`, `apps/api/cmd/api/main.go` | OTLP exporter, trace spans for handlers/DB | Traces visible in Grafana/Tempo | Disable exporter via env | SBL-025 |
| SBL-045 | Instrument worker with OpenTelemetry | P1 | M | `apps/api/cmd/worker/main.go` | Trace context propagation for jobs | Worker traces in collector | Disable exporter | SBL-044 |
| SBL-046 | Deploy self-hosted OTLP collector, Grafana, Tempo, and Prometheus on the VM | P0 | L | `scripts/ops/` (new; process-supervisor units for each) | Collector, Grafana, Tempo, Prometheus running as supervised processes alongside the app | End-to-end trace visible | Stop the observability processes | SBL-029 |
| SBL-047 | Define SLO/SLI dashboards | P1 | M | `scripts/ops/observability/grafana-dashboards/` (new) | Latency, error rate, availability dashboards | Review with SRE | Delete dashboard JSON | SBL-046 |
| SBL-048 | Add alerting rules for error budget and queue depth | P1 | M | `scripts/ops/observability/alerting/` (new), `.github/workflows/deploy-monitoring.yml` | PagerDuty/Slack alertmanager rules | Trigger test alert | Silence alerts | SBL-047 |
| SBL-049 | Replace hand-rolled metrics with Prometheus Go client | P1 | M | `apps/api/internal/metrics` | `github.com/prometheus/client_golang` metrics endpoint | `/metrics` exposes standard buckets | Revert to custom registry | SBL-046 |
| SBL-050 | Add `http_request_duration_seconds` middleware | P1 | S | `apps/api/internal/shared/server/server.go` | Histogram by route/method/status | Metrics query returns data | Remove middleware | SBL-049 |
| SBL-051 | Add structured JSON logs to worker | P1 | S | `apps/api/cmd/worker/main.go` | JSON log format with job metadata | Log output valid JSON | Revert to text logs | SBL-045 |
| SBL-052 | Add log aggregation (Loki) | P1 | M | `scripts/ops/observability/` (new) | Logs shipped to a self-hosted Loki instance and queryable | Query logs in Grafana | Disable agent | SBL-046 |
| SBL-053 | Add synthetic health checks | P2 | M | `scripts/ops/synthetic-checks/` | Cron job hitting `/health` and key flows from outside the VM | Alert on failure | Disable cron | SBL-034 |
| SBL-054 | Add pprof endpoints | P2 | S | `apps/api/internal/shared/server/server.go` | `/debug/pprof` behind auth | Profiles retrievable | Disable in production | SBL-001 |
| SBL-055 | Define incident response runbook | P2 | M | `docs/operations/INCIDENT_RESPONSE.md` | Escalation, rollback, communication steps | Tabletop exercise | Update runbook | SBL-048 |
| SBL-056 | Add distributed trace sampling in production | P2 | S | `apps/api/cmd/api/main.go` | Head-based sampling config (e.g., 10%) | Trace volume within budget | Set sampling to 100% or 0% | SBL-044 |
| SBL-057 | Add queue-depth metrics and alerts | P2 | S | `apps/api/cmd/worker/main.go`, `apps/api/internal/queue` | Gauge for pending/in-progress jobs | Alert fires on backlog | Disable alert | SBL-048 |
| SBL-058 | Add DB slow query monitoring | P2 | S | PostgreSQL config on the VM (`postgresql.conf`) | Enable `log_min_duration_statement`, ship slow logs | Slow query log visible | Disable logging | SBL-026 |
| SBL-059 | Add on-call rotation doc | P2 | XS | `docs/operations/ON_CALL.md` | Primary/secondary rotation, runbook links | Team acknowledges | Update doc | SBL-055 |

## M4 — Commercial SaaS Core

| ID | Title | Priority | Effort | Affected Modules | Expected Deliverables | Validation | Rollback Considerations | Dependencies |
|----|-------|----------|--------|------------------|---------------------|------------|-------------------------|--------------|
| SBL-060 | Generate OpenAPI from chi router and validate in CI | P0 | L | `docs/api/openapi/openapi.yaml`, `.github/workflows/ci.yml` | Generated/validated spec covering all 91 routes | CI gate fails on drift | Revert to manually maintained spec | — |
| SBL-061 | Generate TypeScript SDK from OpenAPI | P0 | L | `packages/sdk/`, `.github/workflows/ci.yml` | Typed `api.ts` SDK published in monorepo | `pnpm build`, `pnpm typecheck` | Delete generated SDK | SBL-060 |
| SBL-062 | Populate `packages/shared` with DTOs/validators | P1 | M | `packages/shared/src/`, `apps/api/internal/*/handler.go` | Shared TS types + Zod schemas | `pnpm typecheck` across web/sdk | Keep separate copies if revert | SBL-061 |
| SBL-063 | Implement billing subscription service with Stripe | P0 | L | `apps/api/internal/billing`, VM environment file (Stripe secret key) | Stripe webhook handler, subscription CRUD | Unit + webhook tests | Disable billing enforcement | — |
| SBL-064 | Implement plan/entitlement engine | P0 | M | `apps/api/internal/billing`, `apps/api/internal/organization` | Seat/run/storage limits, feature flags | Tests enforce limits | Remove limit checks | SBL-063 |
| SBL-065 | Implement self-service billing UI | P1 | M | `apps/web/app/(dashboard)/settings/billing` | Plan selection, card update, invoice list | UI tests; Stripe test mode | Hide billing page | SBL-063 |
| SBL-066 | Implement usage tracking (seats, runs, storage) | P1 | M | `apps/api/internal/billing`, `apps/api/internal/results`, `apps/api/internal/organization` | Usage counters, periodic rollup | Queries return usage | Stop writing counters | SBL-064 |
| SBL-067 | Implement member invitation and role UI | P1 | M | `apps/web/app/(dashboard)/settings/members`, `apps/api/internal/organization` | Invite by email, accept, role assignment | E2E tests | Hide members page | SBL-001 |
| SBL-068 | Implement admin console and safe impersonation | P2 | M | `apps/web/app/(admin)/`, `apps/api/internal/rbac` | Support impersonation with audit | Admin tests | Disable admin routes | SBL-001, SBL-007 |
| SBL-069 | Implement audit log UI | P1 | M | `apps/web/app/(dashboard)/settings/audit-logs` | Paginated audit view, filters | UI tests | Hide page | SBL-007 |
| SBL-070 | Implement notification channels UI | P2 | M | `apps/web/app/(dashboard)/settings/notifications` | Slack/Teams/email channel config | UI tests; masked secrets | Hide settings tab | — |
| SBL-071 | Implement notification preferences UI | P2 | S | `apps/web/app/(dashboard)/settings/notifications` | Toggle per channel type | UI tests | Revert to defaults | SBL-070 |
| SBL-072 | Implement settings sub-pages (org, workspace, profile) | P2 | M | `apps/web/app/(dashboard)/settings/*` | Functional forms for org/workspace/profile | UI tests | Keep placeholder pages | SBL-001 |
| SBL-073 | Implement public API documentation | P2 | M | `docs/api/README.md`, `apps/web/app/docs` | Render OpenAPI as docs site | Docs build passes | Remove docs route | SBL-060 |
| SBL-074 | Implement trial/tier management | P1 | M | `apps/api/internal/billing` | Trial flags, downgrade/upgrades | Unit tests | Disable trial logic | SBL-063 |
| SBL-075 | Implement invoice listing UI | P2 | S | `apps/web/app/(dashboard)/settings/billing` | Invoice list with PDF link | UI tests | Hide invoices | SBL-063 |
| SBL-076 | Add public API rate limits and quotas | P1 | M | `apps/api/internal/shared/middleware/ratelimit.go`, `apps/api/internal/billing` | Quota enforcement per plan | Quota tests | Disable quotas | SBL-064 |
| SBL-077 | Add feature flags for commercial tiers | P2 | M | `apps/api/internal/shared/config`, `apps/web` | LaunchDarkly/simple flag service | Feature gated in UI | Remove flag checks | SBL-064 |
| SBL-078 | Implement signed webhooks for integrations | P2 | M | `apps/api/internal/integrationhub` | Webhook secret verification | Signature tests | Allow unsigned webhooks | SBL-021 |

## M5 — Data & Performance at Scale

| ID | Title | Priority | Effort | Affected Modules | Expected Deliverables | Validation | Rollback Considerations | Dependencies |
|----|-------|----------|--------|------------------|---------------------|------------|-------------------------|--------------|
| SBL-079 | Add missing DB indexes | P1 | M | `apps/api/migrations/` | Indexes for `role_assignments`, `audit_events`, `refresh_tokens`, `notification_channels`, `test_cases`, `queue_jobs` | `go test`; query plans show index usage | Run down migrations | — |
| SBL-080 | Add `organization_id` to `audit_events` | P2 | M | `apps/api/migrations/`, `apps/api/internal/audit` | New column, backfill, tenant-scoped queries | Migration tests; audit UI still works | Down migration | SBL-007 |
| SBL-081 | Optimize `results` recalcRunCounts | P1 | M | `apps/api/internal/results/service.go` | Incremental counter updates | Benchmark large run; no regression | Revert to full recalc | — |
| SBL-082 | Implement data retention purge jobs | P1 | M | `apps/api/internal/queue`, `apps/api/cmd/worker/main.go`, `apps/api/migrations/` | Jobs to purge old results/audit/notifications | Unit + integration tests | Stop scheduling job | — |
| SBL-083 | Add `queue_jobs` dequeue composite index | P1 | S | `apps/api/migrations/` | `(queue_name, status, scheduled_at, created_at)` | Explain plan | Down migration | — |
| SBL-084 | Add pagination to test-run-items endpoint | P2 | M | `apps/api/internal/results/handler.go`, `repository.go`, `ports.go` | Cursor pagination + `nextCursor` | `go test`; large run UI loads | Revert repository signature | — |
| SBL-085 | Add pagination to test-case-versions endpoint | P2 | M | `apps/api/internal/testmanagement/*` | Cursor pagination for versions | `go test` | Revert repository | — |
| SBL-086 | Add pagination to billing/invoices endpoint | P2 | M | `apps/api/internal/billing/*` | Cursor pagination | `go test` | Revert | SBL-063 |
| SBL-087 | Add pagination to integration-events endpoint | P2 | M | `apps/api/internal/integrationhub/*` | Cursor pagination | `go test` | Revert | — |
| SBL-088 | Add materialized search rank for test_cases | P2 | L | `apps/api/migrations/`, `apps/api/internal/testmanagement` | Stored/generated rank or functional index | Search latency benchmark | Remove index | — |
| SBL-089 | Convert dashboard layout to Server Component | P2 | M | `apps/web/app/(dashboard)/layout.tsx` | SSR auth check, reduced JS bundle | Lighthouse TTFB improves | Revert to `"use client"` | SBL-003 |
| SBL-090 | Adopt SWR/React Query for frontend caching | P2 | M | `apps/web/lib/api.ts`, `apps/web/features/**/api.ts` | Stale-while-revalidate, retries, optimistic updates | `pnpm test`; reduced network calls | Revert to raw `fetch` wrapper | SBL-003 |
| SBL-091 | Add `next/image` and CDN loader | P2 | S | `apps/web/next.config.ts` | Image optimization config | Build passes | Disable loader | SBL-037 |
| SBL-092 | Implement bundle analyzer and lazy load | P2 | S | `apps/web/next.config.ts`, `apps/web/app/(dashboard)/*` | Split heavy feature chunks | Bundle size reduced | Remove analyzer | — |
| SBL-093 | Add load testing suite (k6) | P2 | M | `scripts/load/` | k6 scripts for auth, test-runs, ingest, search | Run against staging | Disable load test | SBL-034 |
| SBL-094 | Add DB query plan tracking doc | P3 | S | `docs/engineering/DATABASE_PERFORMANCE.md` | Top queries and plan snapshots | Doc reviewed | Archive doc | SBL-079 |
| SBL-095 | Implement `app.current_tenant()` fail-closed behavior | P2 | S | `apps/api/migrations/` | Raise error on malformed `app.tenant_id` | Migration tests | Down migration | — |

## M6 — Feature Completeness & Scale

| ID | Title | Priority | Effort | Affected Modules | Expected Deliverables | Validation | Rollback Considerations | Dependencies |
|----|-------|----------|--------|------------------|---------------------|------------|-------------------------|--------------|
| SBL-096 | Implement defects module UI | P2 | L | `apps/web/app/(dashboard)/[workspace]/defects`, `apps/api/internal/defects` | List/create/edit defects page | UI + API tests | Hide page | — |
| SBL-097 | Implement analytics dashboards UI | P2 | L | `apps/web/app/(dashboard)/dashboard/analytics`, `apps/api/internal/analytics` | Dashboard widgets, summary, trends | UI tests | Hide analytics | SBL-060 |
| SBL-098 | Implement intelligence/flaky-tests UI | P2 | L | `apps/web/app/(dashboard)/flaky-tests`, `apps/api/internal/intelligence` | Flaky predictions, failure clusters | UI tests | Hide page | SBL-060 |
| SBL-099 | Implement integration hub UI | P2 | L | `apps/web/app/(dashboard)/settings/integrations`, `apps/api/internal/integrationhub` | Jira/Slack/GitHub config UI | UI tests | Hide integrations tab | — |
| SBL-100 | Implement API testing engine scaffolding | P2 | L | `apps/api/internal/apitesting/`, `apps/web/app/(dashboard)/api-tests` | Test collection/execution endpoints and UI | Unit tests | Remove module | — |
| SBL-101 | Add GitHub Actions CI/CD plugin | P3 | L | `plugins/github-action/` | Action that posts results to `/ingest` | Test in sample repo | Remove plugin | SBL-099 |
| SBL-102 | Add GitLab/Jenkins plugin scaffolding | P3 | L | `plugins/gitlab/`, `plugins/jenkins/` | Plugin examples and docs | Docs build | Remove plugins | SBL-101 |
| SBL-103 | Implement requirements traceability matrix | P3 | L | `apps/api/internal/testmanagement`, `apps/web` | Link requirements to cases/runs | UI tests | Hide feature | — |
| SBL-104 | Implement bulk import for test cases | P2 | M | `apps/api/internal/testmanagement`, `apps/web` | CSV/XLSX import endpoint + UI | Import tests | Disable endpoint | SBL-096 |
| SBL-105 | Implement advanced reporting PDF export | P3 | M | `apps/api/internal/reports/`, `apps/web` | PDF report generation | PDF output valid | Remove feature | SBL-097 |
| SBL-106 | Implement SSO/SAML | P3 | L | `apps/api/internal/identity`, `apps/web` | SAML SSO configuration and login | SAML test (e.g., Okta) | Disable SSO | SBL-001 |
| SBL-107 | Implement SCIM provisioning | P3 | L | `apps/api/internal/identity`, `apps/api/internal/organization` | User/group provisioning endpoints | SCIM tests | Disable endpoints | SBL-106 |
| SBL-108 | Implement custom roles | P3 | M | `apps/api/internal/rbac`, `apps/web` | Admin-defined roles and permissions | RBAC tests | Revert to default roles | SBL-067 |
| SBL-109 | Implement data residency controls | P3 | L | VM provisioning scripts (once written), `apps/api/internal/shared/config` | Region selection, data-locality enforcement — requires a per-region VM, deferred until scale beyond one VM is justified | Config + region tests | Disable region flag | SBL-025 |
| SBL-110 | Implement disaster recovery runbooks | P1 | M | `docs/deployment/DISASTER_RECOVERY_GUIDE.md`, `scripts/ops/` | Step-by-step runbooks and scripts | Tabletop DR exercise | Update runbook | SBL-036 |
| SBL-111 | Implement automated rollback scripts | P2 | M | `scripts/ops/rollback.sh`, `.github/workflows/rollback.yml` | Blue/green rollback triggered via GitHub | Dry-run rollback | Manual rollback process | SBL-032 |
| SBL-112 | Implement security incident response plan | P2 | M | `docs/security/INCIDENT_RESPONSE.md` | IR plan with communication and forensics | Review and sign-off | Update doc | SBL-055 |
