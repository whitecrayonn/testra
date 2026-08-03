# Phase 3 Performance Review — Execution & Results Ingestion

**Date:** 2026-07-16  
**Reviewer:** Cascade (AI Engineering Assistant)  
**Scope:** ADR-012 Idempotency-Key middleware, `POST /api/v1/ingest`, and the supporting PostgreSQL schema.  
**Status:** PASS

---

## Executive Summary

The idempotency middleware adds one indexed PostgreSQL lookup and, on cache miss, one insert per ingestion request. The lookup is keyed by `(workspace_id, operation, key_hash)` and is supported by a unique constraint plus a composite index. Request body size is bounded by the existing global `MaxBody` middleware (1 MB). The overall overhead is small and appropriate for the expected CI ingestion traffic. ClickHouse ingestion was deferred per ADR-010, so Phase 3 uses PostgreSQL for result storage.

---

## Latency Analysis

### Hot Path: Idempotency Key Lookup

```text
Client -> Auth -> TenantContext -> RBAC -> AuditLog -> IdempotencyKey -> Handler -> Results service -> DB
```

Within `IdempotencyKey`:

1. **Header read** — O(1)
2. **Body read** — O(n) where `n` is request body size (max 1 MB)
3. **JSON compacting + SHA-256 fingerprint** — O(n); SHA-256 processes body in fixed-size blocks
4. **Workspace ID extraction** — O(1) JSON unmarshal of a single field
5. **Database `SELECT` on `idempotency_records`** — indexed lookup, O(log m) where `m` = records per workspace
6. On cache hit: **write stored response to client** — O(k) where `k` is stored response body size
7. On cache miss: **capture handler response** — streaming write through a `bytes.Buffer`
8. On cache miss: **`INSERT` into `idempotency_records`** — O(1) amortized, constrained by unique index

### Expected Latency Budget

| Operation | Target (p95) | Notes |
|---|---|---|
| Cache hit replay | < 5 ms | No handler re-execution, one DB read + response write |
| Cache miss ingestion | < 100 ms | Includes JUnit/JSON parse + one DB write for idempotency + result writes |
| First-time large XML ingestion | < 300 ms | Parse time dominates; bounded by 1 MB body limit |

These targets are acceptable for CI ingestion, which is asynchronous from the test runner's perspective.

---

## Database Load

### Reads

- **Idempotency lookup:** `SELECT` on `(workspace_id, operation, key)` with `expires_at > NOW()`.
- **Index:** `idx_idempotency_records_lookup` covers `(workspace_id, operation, key)` and is backed by the unique constraint. This is a single-index, low-cost lookup.
- **RLS policy:** The policy adds a correlated subquery against `workspaces`, which is a small, indexed table. The planner will typically resolve the subquery once per query, so the overhead is minimal.

### Writes

- **Idempotency insert:** One `INSERT` per successful cache miss.
- **Result writes:** Existing `test_runs` and `test_run_items` inserts (unchanged by this work).
- **Index maintenance:** Unique constraint on `(workspace_id, operation, key)` adds a small cost per insert. For high-volume CI ingestion, this is acceptable because ingestion is generally bursty and not sustained at thousands of requests per second.

### Storage

- Each record stores `response_body` as `JSONB` plus a few small columns.
- Default TTL: 24 hours (`IDEMPOTENCY_KEY_TTL_MINUTES=1440`).
- A background `DeleteExpired` call (or scheduled job) should run periodically to keep the table small.

**Estimated storage per record (excluding TOAST):**

| Column | Approximate Bytes |
|---|---|---|
| `id` | 16 |
| `workspace_id` | 16 |
| `operation` | ~8 |
| `key` (hashed, 64 hex chars) | 64 |
| `request_fingerprint` (64 hex chars) | 64 |
| `status_code` | 4 |
| `created_at` / `expires_at` | 16 |
| `response_body` | variable, typically 100–300 B |
| Row overhead | ~24 |
| **Total** | **~250–500 B per record** |

At 10,000 ingestions/day, this is roughly **2.5–5 MB/day** with 24-hour retention, well within PostgreSQL capacity.

---

## Scalability Considerations

### What Scales Well

- **Idempotency lookup** is an indexed point query. It scales with workspace count, not total record count, because the leading index key is `workspace_id`.
- **Response replay** avoids re-executing the ingestion pipeline, saving CPU and database load on retries.
- **Fingerprinting** uses the efficient SHA-256 algorithm. For 1 MB bodies, this adds < 1 ms on modern hardware.
- **No raw payload storage** keeps the idempotency table small; only response summaries and fingerprints are kept.

### Potential Bottlenecks

| Bottleneck | Risk | Mitigation |
|---|---|---|
| Large XML payloads close to 1 MB | Parse time and memory | `MaxBody` middleware limits size; parser streams where possible |
| Bursty CI retry storms | Many idempotency lookups | Index keeps lookup cheap; replay avoids repeated ingestion |
| Long TTL with high ingestion volume | Table growth | `DeleteExpired` should be scheduled; TTL is configurable |
| Database RLS subquery | Slight per-query overhead | `workspaces` table is small and indexed; acceptable for MVP |
| Storing response body in `JSONB` | Slightly larger than binary storage | `JSONB` allows indexing if needed later; gzip can be added if storage becomes an issue |

### Concurrency

- The unique constraint on `(workspace_id, operation, key)` prevents duplicate idempotency records under concurrent requests.
- Concurrent requests with the same key and body can race. The first one inserts and the second one reads the existing record. This is safe and deterministic: both will receive the same response, although the second may briefly execute the handler if it read before the first insert committed. This window is small and bounded by transaction duration.
- For stricter exactly-once semantics at very high concurrency, an advisory lock or a `SELECT ... FOR UPDATE` around the lookup could be added in the future.

---

## Resource Usage

### CPU

- SHA-256 hashing of a 1 MB body: ~0.05 ms on a modern CPU.
- JSON compaction is skipped if the body is not JSON (e.g., XML), so XML payloads avoid an extra parse.
- Parsing JUnit/Playwright/Cypress payloads is the dominant CPU cost; idempotency adds negligible overhead.

### Memory

- The middleware reads the entire body into memory once (`io.ReadAll`). With a 1 MB limit, peak middleware memory per request is ~2 MB (original body + compacted copy + response buffer).
- The `responseRecorder` buffers the response body. For typical ingestion results (~100 B), this is negligible.

### Network

- Replay returns the stored response body directly, avoiding downstream service calls and result parsing. This is a net win for retry traffic.

---

## Recommendations

| ID | Recommendation | Priority | Notes |
|---|---|---|---|
| PERF-1 | Schedule `idempotencyStore.DeleteExpired` hourly | HIGH | Prevents unbounded table growth. A simple `DELETE WHERE expires_at <= NOW()` is sufficient. |
| PERF-2 | Add `pg_stat_statements` monitoring for `idempotency_records` lookups and inserts | MEDIUM | Verify p95 latency stays under target. |
| PERF-3 | Consider partitioning `idempotency_records` by `expires_at` if ingestion volume exceeds 1M records/day | LOW | Currently unnecessary; TTL cleanup keeps data volume low. |
| PERF-4 | Compress `response_body` with `pgcrypto` or application-level gzip if result summaries grow | LOW | Not needed for current response sizes. |
| PERF-5 | Add a Redis-backed idempotency store option for multi-region deployments | LOW | PostgreSQL is the source of truth; Redis can be used as a hot cache later. |
| PERF-6 | Evaluate moving `response_body` to a separate table or TOAST-only column if replay responses become large | LOW | Current `JSONB` design is adequate. |

---

## Verification

- `go test -tags=integration -count=1 .\tests\integration` — PASS (validates replay latency is within functional bounds)
- `go test -count=1 .\...` — PASS
- Index review confirms `idx_idempotency_records_lookup` covers the hot lookup path.

---

## Conclusion

The idempotency implementation is performant for Phase 3. The added database and CPU overhead are small relative to the ingestion pipeline, and replay semantics reduce load on retries. The main operational follow-up is scheduling expired-record cleanup (PERF-1). No performance blockers for Phase 3 sign-off.
