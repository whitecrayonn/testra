# Performance Review v1 — Testra Platform

**Date:** 2026-07-19  
**Scope:** Backend (`apps/api`) and frontend (`apps/web`) performance hotspots, including the PostgreSQL-backed queue, in-memory metrics registry, pagination, full-text search, SSRF validation, and build/render architecture.  
**Goal:** Identify bottlenecks and define pragmatic performance improvements before scaling to larger tenants and higher ingestion throughput.

## Executive Summary

The current architecture is appropriate for an early-stage multi-tenant SaaS. The main performance concerns are:
1. **PostgreSQL-backed worker queue** (`queue_jobs`) using `FOR UPDATE SKIP LOCKED` with a single worker per queue.
2. **Unbounded/unpaginated list endpoints** that can return large result sets.
3. **In-memory metrics registry** that is process-local and not persisted or aggregated.
4. **Client-side rendering** of the entire dashboard shell via `"use client"` pages and layouts.
5. **SSR URL validation** on every outbound HTTP call adds network resolution latency.

Most issues are moderate or low severity at current scale but should be addressed before high-volume ingestion or large workspaces.

## 1. Backend Performance

### 1.1 PostgreSQL Job Queue

- Implemented in `apps/api/internal/queue/queue.go`.
- `DequeueOne` uses `FOR UPDATE SKIP LOCKED` to prevent multiple workers from claiming the same job.
- Jobs are selected by `queue_name = $1 AND status = 'pending' AND scheduled_at <= NOW()` ordered by `created_at ASC`.

**Strengths:**
- Skip-locked dequeue avoids worker contention.
- Each job is processed inside a transaction returned to the caller.
- Failed jobs are requeued with exponential backoff (5 minutes × attempt count).
- Old completed/failed jobs can be deleted via `DeleteOldCompleted`.

**Concerns:**
- There is **no composite index** on `(queue_name, status, scheduled_at, created_at)` for the dequeue query. As the queue grows, the scan cost rises.
- `ORDER BY created_at ASC` combined with `FOR UPDATE` can create serialization pressure.
- **Single-threaded per queue** by default; adding more workers requires scaling the worker process, not adding queue partitions.
- `queue_jobs` does not have an explicit `tenant_id` index for tenant-scoped monitoring, and the RLS policy will require a seq-scan if not indexed.

**Recommendations:**
1. Add a composite index:
   ```sql
   CREATE INDEX idx_queue_jobs_dequeue ON queue_jobs (queue_name, status, scheduled_at, created_at ASC);
   ```
2. Add `tenant_id` to the index or create a separate index for `tenant_id` + `status`.
3. Evaluate a dedicated message broker (e.g., NATS, RabbitMQ, SQS) if queue throughput exceeds a few thousand jobs/minute.
4. Add queue-depth metrics and alerting.

### 1.2 Metrics Registry

- `apps/api/internal/metrics/metrics.go` implements a process-local, Prometheus-compatible metrics store.
- Exposes `/metrics` with job counters, duration histograms, ML call metrics, and queue status.

**Strengths:**
- Lightweight, no external dependency.
- Correct histogram implementation with pre-defined buckets.

**Concerns:**
- **Process-local:** multi-replica deployments each expose different metrics.
- **No persistence:** metrics are lost on restart.
- **`defaultRegistry` is a global mutable variable**, making it hard to test and reason about.
- No `pprof` endpoints or built-in request latency histograms.

**Recommendations:**
1. Replace the hand-rolled registry with `github.com/prometheus/client_golang` for standard metrics and exposition.
2. Add `http_request_duration_seconds` and `http_requests_total` middleware.
3. Persist key metrics in Postgres or ship them to Prometheus/Grafana.

### 1.3 Database Queries and Indexes

**Pagination:**
- Cursor pagination is implemented for `organizations`, `workspaces`, `projects`, `test_cases`, `test_cases/search`, `test_runs`, and `notifications`.
- `pagination.ParseParams` now clamps `limit` to `MaxLimit` (100) — fixed in this pass.

**Missing indexes identified in `DATABASE_REVIEW.md`:**
- `role_assignments (user_id, scope_type, scope_id)`
- `audit_events (created_at DESC)`
- `refresh_tokens (family_id, revoked_at)`
- `notification_channels (organization_id, type)`
- `test_cases (workspace_id, status, priority)`
- `queue_jobs` dequeue and tenant indexes

**Full-text search:**
- `test_cases.search_tsv` uses a GIN index and `to_tsquery('pg_catalog.english', ...)`.
- Search ranking is computed twice per row (`ts_rank` in `WHERE`/`ORDER BY` and `SELECT`). Add a subquery/materialized expression or functional index if query volume grows.
- `to_tsquery` does not normalize special characters; `tsQuery` construction should be reviewed for query-bloat attacks.

### 1.4 SSRF Validation Overhead

- `shared/security/ssrf.go` resolves and validates every outbound URL.
- `net.LookupIP` and `net.ParseIP` are fast, but DNS resolution adds latency and can fail for legitimate high-availability endpoints.

**Recommendations:**
1. Cache resolved IPs per URL/host for a short TTL (e.g., 1 minute) to avoid repeated DNS lookups.
2. Allow-list known integrations by hostname via an admin configuration instead of validating every request at runtime.
3. Add a timeout context to DNS/network checks.

### 1.5 API Idempotency and Retries

- Only `/ingest` currently applies `Idempotency-Key` middleware.
- The middleware reads and buffers the entire request body (`io.ReadAll`) and recomputes a SHA-256 fingerprint. For multi-MB payloads this is acceptable, but memory pressure can grow under bulk ingestion.
- Response bodies are captured in memory by a `responseRecorder`; large responses could OOM the process.

**Recommendations:**
1. Apply idempotency to all mutating endpoints (see `API_REVIEW.md`).
2. Limit request body size for idempotent endpoints.
3. Consider streaming or capping response size recorded for replay.

## 2. Frontend Performance

### 2.1 Rendering Model

- Every inspected layout and page in `apps/web/app/` is marked `"use client"`.
- The dashboard layout (`(dashboard)/layout.tsx`) renders `Sidebar` and `RouteGuard` on the client.

**Concerns:**
- No Server Component usage for initial render, resulting in:
  - Larger JavaScript bundles.
  - No server-side data fetching for first paint.
  - Flash of unauthenticated content and client-only redirects.
- `Sidebar` polls `/notifications/unread-count` every 30 seconds indefinitely; this is fine for low-user counts but should be batched or moved to a WebSocket/push model at scale.

**Recommendations:**
1. Convert layouts and read-heavy pages to Server Components.
2. Use Server Components to verify the session and redirect before sending HTML.
3. Defer non-essential client hydration (e.g., notification polling) until after interactive content loads.

### 2.2 Bundle and Asset Delivery

- Next.js `output: 'standalone'` is enabled in production builds.
- `web.build script` copies standalone output correctly.
- No image optimization or CDN configuration is visible in the repo.
- Tailwind content paths include `app`, `features`, and `components`.

**Recommendations:**
1. Add `next/image` and a CDN or loader configuration for user-uploaded attachments/screenshots.
2. Add `next/head` metadata and preload critical fonts/icons.
3. Review bundle size with `@next/bundle-analyzer` and split heavy feature pages.

## 3. Scalability Boundaries

| Component | Current Limit | Scaling Trigger | Mitigation |
|-----------|--------------|-----------------|------------|
| PostgreSQL queue | Single worker per queue, DB connection per dequeue | > 1 job/sec sustained | Add dequeue index, multiple worker replicas, or external broker |
| In-memory metrics | Per-process, lost on restart | > 1 replica or need persistence | Adopt Prometheus client + remote write |
| Unpaginated lists | Full result set returned | Large workspaces | Add cursor pagination to remaining endpoints (see `API_REVIEW.md`) |
| Full-text search | GIN index, rank computed per query | High search volume | Materialize search rank or dedicated search service |
| LocalStorage auth | Tokens held in browser, no SSR session | Any production security requirement | Move to `HttpOnly` cookies/BFF |

## 4. Load Testing Recommendations

Before the next milestone:
1. Define SLOs for p50/p95/p99 response time and error rate.
2. Run `k6` or `artillery` against `/api/v1/auth/login`, `/api/v1/test-runs`, `/api/v1/ingest`, and `/api/v1/test-cases/search`.
3. Monitor PostgreSQL `pg_stat_statements` for slow queries.
4. Verify queue throughput by enqueueing 10,000+ test-run processing jobs.
5. Measure frontend Lighthouse scores and Time-to-First-Byte (TTFB).

## 5. Conclusion

The platform performs well for an MVP but has clear scaling boundaries: the PostgreSQL queue, the in-memory metrics registry, and the client-only rendering model. The cheapest high-impact wins are adding queue indexes, completing cursor pagination, adopting Prometheus metrics, and converting the dashboard shell to Server Components. Dedicated message brokers and external metrics systems can be deferred until measured load justifies them.
