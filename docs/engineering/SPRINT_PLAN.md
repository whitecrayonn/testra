# Testra Sprint Plan — Next 90 Days

**Date:** 2026-08-03
**Owner:** Engineering Lead
**Purpose:** Turn the 112-item `SPRINT_BACKLOG.md` into an actual sequence you can execute sprint by sprint, so "what do I do this week" always has one answer.
**Source data:** `docs/engineering/SPRINT_BACKLOG.md` (task IDs, effort, dependencies), `docs/engineering/LAUNCH_READINESS_PLAN.md` (production-readiness score: 63/100; ~10 weeks to Beta), `docs/engineering/ROADMAP.md` (M1–M6 milestones).
**Cadence assumed:** 2-week sprints, solo/small-team pace. Adjust down if you have more hands.

---

## Where you are right now

- Test Management core is done, including the deterministic test-case-generation feature (v1.1.0) shipped this session.
- Production-readiness score: **63/100**. The gap is entirely M1 (security) and M2 (real infrastructure) — not features.
- The single biggest unblock is **SBL-001 (cookie/session auth)**. Six other tasks in the backlog are stuck behind it (`SBL-002, 003, 007, 013, 018, 022, 024`). Nothing else in M1 fully unblocks until this lands.

**Rule for the next 3 months: don't start new features. Finish M1, then M2.** Everything below is sequencing `SPRINT_BACKLOG.md`, not replacing it — task IDs match that file exactly.

---

## Status: Sprints 1–3 complete (2026-08-03)

An audit found most of Sprint 1–2's backend work already implemented and
verified in code from an earlier pass — no changes needed there. The three
genuine gaps (SBL-004 jti/denylist, SBL-007 audit read endpoint + UI,
SBL-013 RBAC integration tests) have been implemented with test coverage.
Details and the one deliberate scoping decision (SBL-007 is self-scoped
pending SBL-080) are in the `SPRINT_BACKLOG.md` M1 status note. **Sprint 4 is
next** — see the table below, all of it is still untouched.

---

## Sprint 1 — Auth backbone (2 weeks)

Goal: kill the `localStorage` XSS exposure, which is the #1 flagged risk in `LAUNCH_READINESS_PLAN.md`.

| Task | Effort | Why first |
|---|---|---|
| **SBL-001** — httpOnly cookie/session auth backend | L | Foundational; unblocks 6 other tasks |
| SBL-006 — API-key scope registry validation | S | No dependencies, quick win to run in parallel |
| SBL-008 — Fix refresh-token revocation ordering | S | No dependencies, small, unblocks SBL-023 later |

**Exit criteria:** login/logout works end-to-end via cookies in a local test; `go test ./apps/api/internal/identity/...` green.

---

## Sprint 2 — Finish the auth migration (2 weeks)

Goal: actually cut the frontend over, so SBL-001 pays off immediately instead of running two auth paths.

| Task | Effort | Depends on |
|---|---|---|
| **SBL-002** — CSRF token endpoint + middleware | M | SBL-001 |
| **SBL-003** — Migrate `apiFetch` off `localStorage` to cookies | M | SBL-001, SBL-002 |
| SBL-004 — `jti` claim + access-token denylist | M | — |
| SBL-005 — Harden password policy + breached-password check | M | — |

**Exit criteria:** no `localStorage.getItem('testra_token')` left in `apps/web`; `pnpm build && pnpm test` green; manual login/logout/refresh verified.

---

## Sprint 3 — Audit, rate limiting, secrets (2 weeks)

Goal: close out the remaining P0/P1 security items that don't touch the frontend.

| Task | Effort | Depends on |
|---|---|---|
| SBL-007 — Audit log read endpoint + UI | M | SBL-001 |
| SBL-009 — Rate-limiter fail-closed fallback | M | — |
| SBL-010 — PII redaction in request/audit logs | M | — |
| SBL-012 — Secrets-manager provider abstraction | M | — |
| SBL-013 — RBAC end-to-end integration tests | M | SBL-001 |

**Exit criteria:** audit events are viewable in-app; auth endpoints stay protected when Redis is down (test it); no raw email/token/password strings show up in logs.

---

## Sprint 4 — M1 cleanup (2 weeks)

Goal: sweep the rest of M1 so the milestone is fully closed before moving to infrastructure.

| Task | Effort | Depends on |
|---|---|---|
| SBL-014 — API-key auth regression tests | S | SBL-006 |
| SBL-015 — Magic-link password reset | M | — |
| SBL-016 — Binary/dependency scanning (Trivy/Grype) in CI | M | — |
| SBL-017 — SSRF DNS cache + timeout | S | — |
| SBL-018 — Security headers validation tests | S | SBL-003 |
| SBL-020 — Add `X-Content-Type-Options`, `Referrer-Policy`, `Permissions-Policy` | XS | — |
| SBL-022 — Account lockout on repeated failed logins | M | SBL-001 |
| SBL-023 — Session termination on password change | S | SBL-008 |
| SBL-024 — CSP report-uri | S | SBL-003 |

**Exit criteria:** every M1 row in `SPRINT_BACKLOG.md` is checked off except deliberately-deferred infra-dependent items (see below).

> **Note — SBL-011 (production VM firewall rules) is filed under M1 but actually depends on SBL-025 (rent the VM) in M2.** Don't try to do it in Sprint 1–4; it belongs in Sprint 5 once a VM exists. `SPRINT_BACKLOG.md`'s milestone grouping is by *topic*, not strict execution order — this is the one place that matters.

---

## Sprint 5 — Get a real server running (2 weeks)

Goal: stop developing only on your laptop. This is the first infrastructure sprint (M2).

| Task | Effort | Depends on |
|---|---|---|
| SBL-025 — Rent the production VM, choose OS | S | — |
| SBL-026 — Install/configure PostgreSQL on the VM | S | SBL-025 |
| SBL-027 — Reverse proxy + TLS cert | M | SBL-025 |
| SBL-011 — VM host firewall rules (deferred from M1) | M | SBL-025 |
| SBL-037 — Configure production DNS | S | SBL-027 |

**Exit criteria:** `curl https://<your-domain>/health` returns 200 over TLS from outside your network.

---

## Sprint 6 — Make deploys repeatable (2 weeks)

Goal: turn "I SSH'd in and ran commands" into a script, so deploys aren't a one-off ritual you'll forget.

| Task | Effort | Depends on |
|---|---|---|
| SBL-028 — VM provisioning scripts | M | SBL-025–027 |
| SBL-029 — Process-supervisor unit files (API, worker, web) | M | SBL-025 |
| SBL-030 — Build-and-deploy script | M | SBL-028, SBL-029 |
| SBL-031 — Automatic pre-deploy DB backup | S | SBL-030 |
| SBL-034 — Staging/production environment files | M | SBL-029 |

**Exit criteria:** a fresh VM can be stood up from scripts, not memory; one command deploys a new build and backs up the DB first.

---

## After Sprint 6 — where this hands off

At this point M1 is closed and you have a real, repeatable deployment. What's next depends on what you actually need:

- **If you want a design partner / real user on it soon:** jump to CD automation (SBL-032) and basic observability (`SBL-044, 046` — you need to see errors before a stranger hits them), then loop back for the rest of M3.
- **If you're not charging money yet:** M4 (billing, SDK, member/role UI) can wait — don't build a billing system for zero customers.
- **If load/scale is a concern:** M5 is mostly indexes and pagination; cheap, do it opportunistically rather than as a dedicated sprint.

Full detail for M3–M6 is already in `SPRINT_BACKLOG.md` (SBL-044 onward) — sequence it the same way (pull P0 first, respect the `Dependencies` column) once you get there. I can build out Sprint 7+ in the same format when you're closer to it; planning six sprints ahead in detail right now would mostly just go stale.

---

## See Also

- [`SPRINT_BACKLOG.md`](SPRINT_BACKLOG.md) — full 112-task backlog this plan sequences
- [`LAUNCH_READINESS_PLAN.md`](LAUNCH_READINESS_PLAN.md) — audit findings and readiness score
- [`ROADMAP.md`](ROADMAP.md) — milestone definitions and phase history
