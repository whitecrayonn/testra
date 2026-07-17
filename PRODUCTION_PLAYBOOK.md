# TESTRA Production Playbook

## 1. Objective

This document is the canonical sign-off artifact for the **Phase 8 — Production Go-Live Orchestration & Observability** mandate. It defines:

- Secure production secret provisioning mapped to `testra/infra/k8s/base/secret.yaml`.
- Rotation and injection flows for high-risk credentials.
- Asynchronous worker telemetry and Python ML engine latency tracking.
- The post-deployment smoke test suite.
- The ordered go-live checklist.

## 2. Production Environment Secrets Provisioning

### 2.1 Secret schema

Production values are never stored in Git. `testra/infra/k8s/base/secret.yaml` is the **shape** of the Kubernetes Secret consumed by the API, worker, and migrator. Values are injected at deploy time by the CI/CD pipeline from a vault or managed secret store (e.g. HashiCorp Vault, AWS Secrets Manager, GCP Secret Manager, Azure Key Vault, or Bitnami Sealed Secrets).

| Key | Consumer(s) | Purpose |
| --- | --- | --- |
| `DATABASE_URL` | `api`, `worker`, `migrator` | PostgreSQL connection string for the RLS-aware pooled role. |
| `JWT_SECRET` | `api`, `worker` (if it validates tokens) | Symmetric signing key for access/refresh tokens. |
| `STRIPE_SECRET_KEY` | `api` | Live Stripe API key for billing operations. |
| `STRIPE_PRICE_ID` | `api` | Stripe price ID used when creating subscriptions. |

Example production `DATABASE_URL` shape:

```text
postgres://testra_api:<random-pass>@prod-pgbouncer.testra.svc:5432/testra?sslmode=require&pool_max_conns=20&application_name=testra-api&search_path=public
```

### 2.2 Injection flow

```
Secret Store (Vault / KMS / Sealed Secrets)
         │
         ▼
CI/CD pipeline (GitHub Actions / GitLab CI / Argo CD)
         │
         ▼
kubectl apply -n testra -f infra/k8s/base/secret.yaml
         │
         ▼
Pod envFrom: secretRef: name: testra-secrets
         │
         ▼
apps/api/internal/shared/config/config.go reads env vars
```

Notes:
- `infra/k8s/base/secret.yaml` is committed with empty placeholder values and annotations that name the secret source (`testra.io/secret-source: vault`) and rotation policy.
- Real secret values are written by the pipeline only. Manual `kubectl edit secret` is reserved for break-glass rotation and must be recorded in the incident log.

### 2.3 Credential rotation strategy

#### `STRIPE_SECRET_KEY` and `STRIPE_PRICE_ID`

- **Rotation owner:** Finance / Platform lead.
- **STRIPE_SECRET_KEY:**
  1. Generate a restricted live key in the Stripe Dashboard with only the required permissions (`read products/prices`, `read customers`, `create subscriptions`, `read subscriptions`).
  2. Update the secret in the secret store.
  3. Trigger a rolling restart of `testra-api` and `testra-worker` so `config.Load()` picks up the new value.
  4. Revoke the old key after observing successful payment/subscription operations for one full billing cycle (or 24 hours minimum).
- **STRIPE_PRICE_ID:**
  - Does not contain cryptographic material; rotate when the pricing plan changes.
  - Update the value in the secret store, then restart `testra-api`.

#### `JWT_SECRET`

- **Rotation owner:** Security / Platform lead.
- **Generation:** `openssl rand -base64 64`.
- **Rotation flow (zero-downtime):**
  1. Generate a new signing key (`JWT_SECRET_NEW`).
  2. Update the application to issue tokens with the new key while still accepting tokens signed by the previous key during a grace window equal to the longest token lifetime (`JWT_ABSOLUTE_DAYS`).
  3. Update `JWT_SECRET` in the secret store to the new key and restart API pods.
  4. After the grace window, remove the old key acceptance logic.
- **Emergency rotation:** If leakage is suspected, rotate immediately. Existing sessions will be invalidated unless dual-key validation has been pre-deployed.

#### `DATABASE_URL` (PostgreSQL RLS pooling)

- **Rotation owner:** DBA / Platform lead.
- **Rotation flow:**
  1. Rotate the database role password or create a new pooled role.
  2. Update the connection string in the secret store.
  3. Rolling restart the API and worker deployments.
  4. Monitor `testra_worker_queue_jobs_status{status="failed"}` and API health endpoints for connection errors.
- **RLS pooling notes:**
  - Connections must target a pooler (e.g. PgBouncer) or Postgres-native pool with `pool_max_conns` bounded per pod.
  - Each transaction sets `SET LOCAL app.tenant_id = $1` (worker) and the API uses per-request RLS enforcement. The `DATABASE_URL` must include `sslmode=require` in production.

## 3. Async Worker Telemetry

### 3.1 Metrics exporter

The worker (`apps/api/cmd/worker/main.go`) now exposes a Prometheus-compatible `/metrics` endpoint on `METRICS_PORT` (default `9090`) via `apps/api/internal/metrics/metrics.go`.

`infra/k8s/base/worker.yaml` exposes `containerPort: 9090` and the `testra-worker` Service advertises Prometheus scrape annotations:

```yaml
prometheus.io/scrape: "true"
prometheus.io/port: "9090"
prometheus.io/path: "/metrics"
```

### 3.2 Worker metrics reference

| Metric | Type | Labels | Meaning |
| --- | --- | --- | --- |
| `testra_worker_jobs_total` | counter | `job_type`, `status` | Total processed jobs. `status` ∈ `success`, `retry`, `dead_letter`. |
| `testra_worker_job_duration_seconds` | histogram | `job_type`, `status` | Job execution latency in seconds. |
| `testra_worker_queue_jobs_status` | gauge | `status` | Live count of `queue_jobs` rows by status (`pending`, `processing`, `completed`, `failed`). |
| `testra_ml_requests_total` | counter | `method`, `status` | Calls to the Python ML engine. `method` ∈ `predict_flaky`, `classify_failure`. `status` ∈ `success`, `error`. |
| `testra_ml_request_duration_seconds` | histogram | `method`, `status` | ML HTTP call latency in seconds. |

Instrumentation points:
- `apps/api/cmd/worker/main.go` records `testra_worker_jobs_total` and `testra_worker_job_duration_seconds` around `processJob`.
- `apps/api/internal/intelligence/mlclient.go` records `testra_ml_requests_total` and `testra_ml_request_duration_seconds` for every HTTP call to `testra-ml`.
- `apps/api/internal/metrics/metrics.go` exposes live queue status by querying `SELECT status, COUNT(*) FROM queue_jobs GROUP BY status` on each scrape.

### 3.3 Key PromQL examples

Job failure / retry exhaustion rate:

```promql
rate(testra_worker_jobs_total{status="dead_letter"}[5m])
```

Slow `intelligence:predict` jobs (p99):

```promql
histogram_quantile(0.99,
  sum(rate(testra_worker_job_duration_seconds_bucket{job_type="intelligence:predict"}[5m])) by (le)
)
```

ML engine error rate:

```promql
rate(testra_ml_requests_total{status="error"}[5m])
/
rate(testra_ml_requests_total[5m])
```

Queue backlog and dead letters:

```promql
testra_worker_queue_jobs_status{status="pending"}
testra_worker_queue_jobs_status{status="failed"}
```

## 4. Post-Deployment Smoke Test Suite

Script: `testra/scripts/prod-smoke-test.sh`

Usage:

```bash
GATEWAY_URL=https://app.testra.io bash testra/scripts/prod-smoke-test.sh
```

What it validates:

1. **Outer gateway ping** — `GET /` returns HTTP 2xx/3xx.
2. **API health** — `GET /health` returns JSON containing `"status":"ok"`.
3. **Frontend login route** — `GET /login` returns a non-empty response and does not contain known Next.js static-generation or runtime error markers (e.g. `Application error`, `Internal Server Error`, `__NEXT_DATA__` with an `err` payload).

## 5. Go-Live Orchestration Checklist

1. **Pre-flight**
   - [ ] All migrations in `testra/apps/api/migrations/` applied to the production database.
   - [ ] `infra/k8s/base/secret.yaml` injected with real values via the CI/CD pipeline.
   - [ ] Container images tagged with the release SHA (e.g. `testra-api:v1.0.0-sha`, `testra-worker:v1.0.0-sha`, `testra-ml:v1.0.0-sha`, `testra-web:v1.0.0-sha`).
   - [ ] `CORS_ALLOWED_ORIGINS` in `infra/k8s/base/configmap.yaml` updated to the production web origin.

2. **Apply infrastructure in order**
   - [ ] `kubectl apply -f infra/k8s/base/configmap.yaml`
   - [ ] `kubectl apply -f infra/k8s/base/secret.yaml` (or pipeline-injected equivalent)
   - [ ] `kubectl apply -f infra/k8s/base/deployment.yaml`
   - [ ] `kubectl apply -f infra/k8s/base/worker.yaml`
   - [ ] `kubectl apply -f infra/k8s/base/ml.yaml`
   - [ ] `kubectl apply -f infra/k8s/base/web.yaml`
   - [ ] `kubectl apply -f infra/k8s/base/service.yaml`
   - [ ] Apply production `Ingress` / `Gateway` manifests in `infra/k8s/overlays/production/`.

3. **Validation**
   - [ ] `kubectl rollout status deployment/testra-api`
   - [ ] `kubectl rollout status deployment/testra-worker`
   - [ ] `kubectl rollout status deployment/testra-ml`
   - [ ] Run `testra/scripts/prod-smoke-test.sh` against the outer gateway.
   - [ ] Confirm Prometheus is scraping `testra-worker:9090/metrics`.

4. **Traffic cutover**
   - [ ] Update public DNS / CDN to point to the production ingress.
   - [ ] Verify TLS certificate and HSTS headers.
   - [ ] Monitor `testra_worker_queue_jobs_status{status="pending"}` for backlog spikes.

5. **Rollback**
   - [ ] Roll back to the previous image tag via `kubectl rollout undo deployment/<name>`.
   - [ ] If a secret is compromised, rotate the affected credential immediately using the flows in Section 2.3.

## 6. Sign-Off

- **RLS layers:** Verified and in production-ready state.
- **Error sanitization:** No secret leakage in worker or API error paths.
- **Kubernetes manifests:** Probes and resource constraints validated in prior phases.
- **Observability:** Worker `/metrics` endpoint, ML latency histograms, and queue status gauges are live.
- **Smoke tests:** `scripts/prod-smoke-test.sh` is ready for post-deploy validation.

**Status:** TESTRA is signed off and ready for global traffic routing.
