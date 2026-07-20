# Engineering Risk Register

**Date:** 2026-08-02  
**Owner:** Engineering Lead / Engineering Manager  
**Source of truth:** `docs/engineering/RISK_REGISTER.md`  
**Scope:** Engineering, security, product, operational, and scalability risks blocking or threatening production launch.

---

## 1. Risk Scoring Legend

| Severity | Definition |
|----------|------------|
| **Critical** | Would block launch, cause data breach, major outage, or legal/compliance failure. |
| **High** | Significant customer impact, revenue impact, or operational burden if realized. |
| **Medium** | Manageable impact; requires mitigation before GA. |
| **Low** | Minor impact; can be addressed post-GA. |

**Likelihood:** Rare / Unlikely / Possible / Likely / Almost Certain.

---

## 2. Technical & Architecture Risks

| ID | Risk | Severity | Likelihood | Impact | Detection | Mitigation | Owner Rec. | Target Resolution |
|----|------|----------|------------|--------|-----------|------------|------------|-------------------|
| T-01 | OpenAPI spec and implementation drift; frontend silently breaks as backend evolves. | High | Likely | High | CI diff, manual code review | Generate spec from `chi` router; validate in CI; generate TypeScript SDK and consume it in `apps/web` | Backend Lead | SBL-060, SBL-061 |
| T-02 | In-memory metrics registry is per-pod, loses data on restart and cannot aggregate across replicas. | High | Almost Certain | Medium | Missing centralized metrics; observability gaps | Replace with Prometheus Go client and OTLP export; deploy collector + Grafana | SRE Lead | SBL-049, SBL-046 |
| T-03 | `results.Service.recalcRunCounts` loads all run items into memory — O(n) and scales poorly. | High | Likely | High | Load tests; slow queries on large runs | Incremental counters + materialized aggregates; benchmark before/after | Backend Lead | SBL-081 |
| T-04 | `localStorage` token storage exposes refresh token to XSS and `document.cookie` read for clientside scripts. | Critical | Likely | Critical | Security audit; pen test | Move to httpOnly/Secure/SameSite cookies + BFF/session; add CSRF | Security Lead | SBL-001, SBL-002, SBL-003 |
| T-05 | API key scope strings are accepted without validation, allowing arbitrary/invalid scopes. | Medium | Possible | Medium | Code review; misuse in UI | Implement allowed-scope registry in `apikeys.Service` | Backend Lead | SBL-006 |
| T-06 | Rate limiter falls back to in-memory store and fails open if Redis disappears. | High | Possible | High | Redis outage simulation | Fail-closed on Redis auth endpoints; keep local in-memory fallback for non-auth routes | SRE Lead | SBL-009 |
| T-07 | Client-side rendering of authenticated pages causes hydration flicker and "Flash of Unauthenticated Content". | Medium | Likely | Medium | Lighthouse, UX review | Convert layouts to Server Components; move auth check to middleware/server | Frontend Lead | SBL-089 |
| T-08 | Idempotency cache lives only in-process; duplicated requests may hit different pods. | Medium | Possible | Medium | Load tests; duplicate request tests | Move idempotency store to Redis/Postgres with TTL | Backend Lead | SBL-046 |
| T-09 | SSRF guard performs DNS resolution on every outbound request and has no cache/timeout, adding latency. | Low | Likely | Low | Latency profiling; security test | Add TTL-bounded DNS cache and context timeout | Security Lead | SBL-017 |
| T-10 | Placeholder pages and modules (`analytics`, `apitesting`, `integrationhub`, `billing`, `intelligence`) create user-facing dead ends. | Medium | High | Medium | UX audit; customer feedback | Hide incomplete pages behind feature flags; complete or remove before GA | Product Lead | SBL-096–SBL-100 |

## 3. Security & Compliance Risks

| ID | Risk | Severity | Likelihood | Impact | Detection | Mitigation | Owner Rec. | Target Resolution |
|----|------|----------|------------|--------|-----------|------------|------------|-------------------|
| S-01 | Access tokens lack `jti`/deny-list; stolen tokens cannot be revoked until expiry. | High | Likely | High | Security review; pen test | Add `jti` claim and deny-list table; validate on every request | Security Lead | SBL-004 |
| S-02 | Refresh tokens are revoked *after* a new token is issued, allowing replay race. | High | Unlikely | High | Token-replay tests | Revoke old refresh token before issuing new; test concurrent refresh | Backend Lead | SBL-008 |
| S-03 | Password reset token is passed raw through handler and email, increasing leak surface. | Medium | Possible | Medium | Code review; email inspection | Switch to HTTPS magic-link with short-lived signed token | Security Lead | SBL-015 |
| S-04 | PII (emails, IPs, query params) may be written to logs or audit metadata. | High | Possible | High | Log review; compliance audit | Redact `RemoteAddr`, body, and query params in audit and log output | Security Lead | SBL-010 |
| S-05 | No host firewall rules on the VPS allow unrestricted ingress/egress. | High | Possible | High | Infrastructure audit; pen test | Add default-deny `ufw`/`nftables` rules with allow-listed ports and egress | DevOps Lead | SBL-011 |
| S-06 | No compiled binary scanning or SBOM generation; vulnerable dependencies can ship. | Medium | Possible | High | CI audit | Add Trivy/Grype to CI; fail builds on critical CVEs | DevOps Lead | SBL-016 |
| S-07 | Secrets are sourced only from env vars / plain text files; no secret rotation or audit. | High | Likely | High | Security audit | Adopt a local secrets store or secrets manager; implement rotation policy | Security Lead | SBL-012, SBL-035 |
| S-08 | No GDPR/data-erasure workflow for tenant/user deletion. | Medium | Possible | High | Compliance review | Implement tenant/user deletion API and purge jobs | Compliance Lead | SBL-082, SBL-109 |
| S-09 | Weak password policy (min 12 only) permits common/weak passwords. | Medium | Likely | Medium | Credential-stuffing test | Add complexity + breached-password check | Security Lead | SBL-005 |
| S-10 | No Web Application Firewall or DDoS protection at ingress. | Medium | Possible | High | Load test; pen test | Add Let's Encrypt TLS, nginx hardening, and optional CDN/WAF rules on the single Ubuntu VPS | DevOps Lead | SBL-038 |

## 4. Product & Commercial Risks

| ID | Risk | Severity | Likelihood | Impact | Detection | Mitigation | Owner Rec. | Target Resolution |
|----|------|----------|------------|--------|-----------|------------|------------|-------------------|
| P-01 | Billing/entitlements missing; cannot monetize or enforce plan limits. | Critical | High | Critical | Revenue readiness review | Implement Stripe subscription + plan engine; enforce limits in API | Product/Eng Lead | SBL-063, SBL-064 |
| P-02 | Public SDK and OpenAPI contract missing; developer/customer integrations stall. | High | Likely | High | Partner feedback | Generate and publish OpenAPI + TypeScript SDK | Product Lead | SBL-060, SBL-061 |
| P-03 | No member invitation / RBAC UI; teams cannot self-manage access. | High | High | High | UX review; sales feedback | Build members settings page with invite flow | Product Lead | SBL-067 |
| P-04 | Defects, analytics, integration hub, API testing are placeholders; product is incomplete for QA teams. | High | High | High | Customer demos | Prioritize core modules before GA; hide or scaffold behind flags | Product Lead | SBL-096–SBL-100 |
| P-05 | No feature flags or tier gating; enterprise/commercial features cannot be controlled. | Medium | Possible | Medium | Go-to-market review | Add feature-flag service and plan-aware checks | Product Lead | SBL-077 |
| P-06 | No public API documentation; developer adoption blocked. | Medium | Likely | Medium | Docs audit | Render OpenAPI as docs site | Product Lead | SBL-073 |

## 5. Operational & Infrastructure Risks

| ID | Risk | Severity | Likelihood | Impact | Detection | Mitigation | Owner Rec. | Target Resolution |
|----|------|----------|------------|--------|-----------|------------|------------|-------------------|
| O-01 | No production-ready single-Ubuntu-VPS systemd services/single-Ubuntu-VPS systemd; deployment is manual and non-repeatable. | Critical | High | Critical | Infra audit | Implement full modules + overlays + CI/CD | DevOps Lead | SBL-025–SBL-034 |
| O-02 | No distributed tracing or production dashboards; incident response is blind. | Critical | High | Critical | On-call drills | OpenTelemetry + Grafana + alerting | SRE Lead | SBL-044–SBL-048 |
| O-03 | No tested backup/restore or DR runbook; data loss possible. | Critical | Possible | Critical | DR tabletop | PITR backups, restore scripts, quarterly drills | SRE Lead | SBL-036, SBL-110 |
| O-04 | CI does not run integration tests against real Postgres/Redis; regressions slip through. | High | Likely | High | Test failures in staging | Add services job in GitHub Actions | QA Lead | SBL-039 |
| O-05 | No binary artifact upload / deployment pipeline; releases are ad-hoc. | High | High | High | Release process audit | Build CD workflow to ECR and staging single-Ubuntu-VPS systemd | DevOps Lead | SBL-032 |
| O-06 | No autoscaling / PDB; single-node failures take service down. | Medium | Possible | High | single-Ubuntu-VPS systemd review | Add HPA, PDB, topology spread | DevOps Lead | SBL-030, SBL-043 |
| O-07 | `apps/worker` is not defined as a single-Ubuntu-VPS systemd deployment; background jobs do not run in cluster. | High | High | High | Missing worker manifests | Add `worker-deployment.yaml` and CI build | DevOps Lead | SBL-042 |
| O-08 | Remote single-Ubuntu-VPS systemd services state lacks DynamoDB locking; state corruption risk. | Medium | Possible | Medium | single-Ubuntu-VPS systemd services init review | Configure S3 backend with DynamoDB lock | DevOps Lead | SBL-028 |
| O-09 | No on-call rotation or incident response plan; outages extend MTTR. | Medium | Possible | High | Ops review | Define rotation, runbooks, paging | SRE Lead | SBL-055, SBL-059 |
| O-10 | No synthetic/health probes from outside the cluster; issues detected by users. | Medium | Likely | Medium | Monitoring gaps | Add external synthetic checks | SRE Lead | SBL-053 |

## 6. Scalability & Performance Risks

| ID | Risk | Severity | Likelihood | Impact | Detection | Mitigation | Owner Rec. | Target Resolution |
|----|------|----------|------------|--------|-----------|------------|------------|-------------------|
| SP-01 | Missing indexes on high-churn tenant-scoped tables cause sequential scans. | High | Likely | High | Slow query logs; load tests | Add composite indexes per `DATABASE_REVIEW.md` | Backend Lead | SBL-079 |
| SP-02 | `queue_jobs` dequeue lacks composite index; worker throughput degrades. | High | Likely | High | Worker latency metrics | Add `(queue_name, status, scheduled_at, created_at)` index | Backend Lead | SBL-083 |
| SP-03 | Full-text search recomputes rank per query; large workspaces slow down. | Medium | Possible | Medium | Search latency metrics | Add materialized/generated rank or caching | Backend Lead | SBL-088 |
| SP-04 | Unpaginated list endpoints return entire tables (`test-run-items`, `versions`, `invoices`, `integration-events`). | Medium | Likely | Medium | Large-tenant tests; API timeouts | Add cursor pagination to remaining lists | Backend Lead | SBL-084–SBL-087 |
| SP-05 | Frontend bundle and client-side rendering slow initial page loads. | Medium | Likely | Medium | Lighthouse / TTFB | SSR, SWR, image optimization, lazy loading | Frontend Lead | SBL-089–SBL-092 |
| SP-06 | No load testing; unknown behavior at target concurrency. | Medium | Possible | High | Launch capacity review | Implement k6 suite; run against staging | SRE Lead | SBL-093 |
| SP-07 | Data retention not enforced; storage grows unbounded. | Medium | Likely | Medium | Storage cost alerts | Implement purge jobs and retention policy | Backend Lead | SBL-082 |

## 7. Risk Priority Matrix

| Risk | Severity × Likelihood | Phase / Sprint Backlog |
|------|----------------------|------------------------|
| T-04 Token storage in `localStorage` | **Critical × Likely** | SBL-001–SBL-003 |
| P-01 Missing billing/entitlements | **Critical × High** | SBL-063–SBL-064 |
| O-01 No production IaC | **Critical × High** | SBL-025–SBL-034 |
| O-02 No observability | **Critical × High** | SBL-044–SBL-048 |
| T-03 O(n) run aggregate recalc | **High × Likely** | SBL-081 |
| T-01 OpenAPI/SDK drift | **High × Likely** | SBL-060–SBL-061 |
| S-05 No single-Ubuntu-VPS systemd host firewall rules | **High × Possible** | SBL-011 |
| S-01 No token `jti`/deny-list | **High × Likely** | SBL-004 |
| O-03 No tested DR | **Critical × Possible** | SBL-036, SBL-110 |
| SP-01 Missing DB indexes | **High × Likely** | SBL-079 |

## 8. Mitigation Tracking

| Risk | Status | Next Review |
|------|--------|-------------|
| All P0 risks | Open | 2026-08-16 |
| P1 risks | Open | 2026-08-30 |
| P2/P3 risks | Open | 2026-09-15 |

---

**Prepared by:** Cascade (pair programming assistant)  
**Review cadence:** Weekly during launch-readiness sprint; bi-weekly post-Beta.
