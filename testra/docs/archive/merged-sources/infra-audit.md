# Infrastructure & Operations Audit

**Scope:** `testra/infra` (Docker, Kubernetes, Terraform), `testra/.github/workflows`, `testra/Makefile`, `testra/package.json`, and environment configuration for `apps/api` and `apps/web`.

**Status:** Audit complete.

---

## 1. Executive Summary

The project ships infrastructure as code for local development (Docker Compose) and a Kubernetes base layer with Kustomize overlays for staging and production. Terraform is configured for AWS but the actual module directory is empty. CI is handled by a single GitHub Actions workflow that builds and tests the Go API, type-checks and builds the Next.js frontend, and lints the Python ML app. There is no CD workflow, no image registry push, and no automated deployment.

Local development uses `pnpm` + `turbo` and starts backing services via `scripts/dev/start-infra.mjs`.

---

## 2. Local Development Infrastructure

### 2.1 Docker Compose (`infra/docker/docker-compose.yml`)

Defines the following services:

| Service | Image | Ports | Purpose |
|---------|-------|-------|---------|
| `postgres` | `postgres:16-alpine` | `5432:5432` | Primary application database |
| `clickhouse` | `clickhouse/clickhouse-server:24` | `8123:8123`, `9000:9000` | Analytics store (not currently used by the app) |
| `redis` | `redis:7-alpine` | `6379:6379` | Caching / future worker queue |
| `minio` | `minio/minio:latest` | `9002:9000`, `9001:9001` | S3-compatible object storage for artifacts |
| `mailpit` | `axllent/mailpit:latest` | `1025:1025`, `8025:8025` | SMTP capture and web UI |

**Observations:**

- The application services (`api`, `web`, `worker`, `migrator`) are **not** defined in the compose file. Developers must run them manually or via `pnpm dev` / `turbo`.
- Default credentials are hard-coded (`testra`/`testra`, `testratestra`) — acceptable for local dev but must never reach production.
- Postgres, ClickHouse, Redis, and MinIO define `healthcheck`s; `mailpit` does not.

### 2.2 Dockerfile inventory (`infra/docker/`)

| File | Purpose |
|------|---------|
| `api.Dockerfile` | Multi-stage Go build for `apps/api/cmd/api` |
| `web.Dockerfile` | Multi-stage Next.js build producing a standalone output image |
| `migrator.Dockerfile` | Go build for `apps/api/cmd/migrator` |
| `worker.Dockerfile` | Go build for `apps/worker` (currently a stub) |
| `ml.Dockerfile` | Python image for `apps/ml` |

**Details:**

- `api.Dockerfile`: builder stage uses `golang:1.23-alpine`, runs `CGO_ENABLED=0 GOOS=linux go build`. Final stage is `alpine:latest` with `ca-certificates`.
- `web.Dockerfile`: builder uses `node:20-alpine` with `pnpm@9.5.0`. Copies full repo, runs `pnpm install --frozen-lockfile`, then `pnpm --filter @testra/web build`. Final stage copies `.next/standalone`, `.next/static`, and `public`.

**Observations:**

- Dockerfiles use `COPY . .` in the web builder, which can be slow and may include unnecessary files if `.dockerignore` is missing.
- No `.dockerignore` reviewed; verify one exists to keep build context small.
- The worker build is currently building an empty stub.

---

## 3. Kubernetes

### 3.1 Base manifests (`infra/k8s/base/`)

- `deployment.yaml` — `testra-api` Deployment with 2 replicas, image `testra-api:latest`, container port `8080`.
- `service.yaml` — `testra-api` Service on port `80` targeting `8080`.

**Observations:**

- No `ConfigMap` or `Secret` is mounted; environment variables (database URL, JWT secret, etc.) are not defined.
- No `livenessProbe`/`readinessProbe`.
- No `resources` (CPU/memory limits/requests).
- No `Ingress`, `HPA`, `PodDisruptionBudget`, or `ServiceAccount`.
- Only the API deployment is represented; no web, worker, or migrator Jobs.

### 3.2 Overlays (`infra/k8s/overlays/{staging,production}/`)

Each overlay contains a `kustomization.yaml`:

```yaml
namespace: testra-staging  # or testra-production
resources:
  - ../../base
images:
  - name: testra-api
    newTag: staging          # or production
```

**Observations:**

- Overlays only change the image tag and namespace. No environment-specific config, replica counts, resource limits, or secrets are applied.
- No Kustomize components for ingress, monitoring, or cert-manager.

---

## 4. Terraform

### 4.1 Root module (`infra/terraform/main.tf`)

- Defines AWS provider `~> 5.0`.
- Default region: `ap-southeast-1`.
- Required Terraform version: `>= 1.8`.

### 4.2 Environments (`infra/terraform/environments/{staging,production}/main.tf`)

- Each environment uses an `s3` backend with no explicit bucket/key block (to be supplied at init time).
- Each calls a module from `../../modules` with `environment` and `region` variables.

**Observations:**

- `infra/terraform/modules/` appears empty (0 items listed in directory scan).
- There are no resources defined yet; Terraform is scaffolded but not implemented.
- No remote state locking (e.g., DynamoDB) configured.

---

## 5. CI/CD

### 5.1 GitHub Actions workflow (`.github/workflows/ci.yml`)

Triggers: `pull_request` and `push` to `main`.

**Jobs:**

1. **Go (`api, worker`)**
   - `setup-go` v1.23
   - `go build ./...` in `apps/api`
   - `go vet ./...` in `apps/api`
   - `go test -race -count=1 ./...` in `apps/api`
   - Note: worker build/test not explicitly executed; cache path references `apps/worker/go.sum`.

2. **Web (`typecheck, lint, build`)**
   - `pnpm/action-setup@v4` with pnpm 9
   - `setup-node` v20 with pnpm cache
   - `pnpm install --frozen-lockfile`
   - `pnpm turbo run typecheck`
   - `pnpm turbo run build`

3. **ML (`lint, test`)**
   - Python 3.12
   - Installs `ruff`
   - `ruff check .` in `apps/ml`

**Observations:**

- No Docker image build or push.
- No deployment job to staging/production.
- No dependency vulnerability scanning, static analysis (SAST/DAST), or secret scanning.
- `go test` runs against no database; tests likely unit-only or may fail if integration tests require Postgres.
- No workflow for `apps/worker` (Go or Python).

---

## 6. Monorepo Scripts & Tooling

### 6.1 Root `package.json`

```json
{
  "scripts": {
    "dev": "node scripts/dev/start-infra.mjs && turbo run dev --filter=\"!@testra/ml\"",
    "dev:all": "node scripts/dev/start-infra.mjs && turbo run dev",
    "build": "turbo run build",
    "test": "turbo run test",
    "lint": "turbo run lint",
    "typecheck": "turbo run typecheck",
    "clean": "node scripts/dev/clean.mjs",
    "format": "prettier --write ...",
    "postinstall": "node scripts/dev/setup-python.mjs"
  },
  "devDependencies": {
    "prettier": "^3.3.3",
    "turbo": "^2.0.9"
  },
  "packageManager": "pnpm@9.5.0"
}
```

`pnpm-workspace.yaml` includes `apps/*` and `packages/*`.

### 6.2 Makefile

```
dev     -> pnpm dev
build   -> pnpm build
test    -> pnpm test
lint    -> pnpm lint
typecheck -> pnpm typecheck
clean   -> pnpm clean
migrate -> go run ./apps/api/cmd/migrator
```

`make migrate` runs the Go migrator directly against `DATABASE_URL` in the environment.

### 6.3 Environment configuration

`apps/api/.env.example`:

```
ENV=development
PORT=8080
DATABASE_URL=postgres://testra:testra@localhost:5432/testra?sslmode=disable
MIGRATIONS_PATH=apps/api/migrations
JWT_SECRET=dev-jwt-secret-change-in-production
JWT_EXPIRY_HOURS=168
REDIS_ADDR=localhost:6379
IDEMPOTENCY_KEY_TTL_MINUTES=1440
```

**Observations:**

- `JWT_EXPIRY_HOURS=168` contradicts the default `JWTExpiryMinutes=15` in the code; the env example is not used if the variable is not read. The actual env key in code is `JWT_EXPIRY_MINUTES`.
- No `.env.example` in `apps/web` for `NEXT_PUBLIC_API_URL`.
- No documentation on required secrets for production (SMTP, S3/MinIO, ClickHouse, Redis).
- `JWT_SECRET` is a weak placeholder and flagged correctly.

---

## 7. Findings & Recommendations

1. **No application services in Docker Compose.** Add `api`, `web`, and `migrator` services so a single `docker compose up` runs the full stack locally.
2. **Kubernetes is incomplete.** Add `ConfigMap`, `Secret`, `Ingress`, `livenessProbe`/`readinessProbe`, resource requests/limits, and a `Job` for database migrations.
3. **Terraform is not implemented.** Populate `infra/terraform/modules/` with VPC, EKS/RDS/S3, and IAM resources, or replace with the intended cloud architecture.
4. **No CD pipeline.** Add a workflow that builds and pushes Docker images to a registry, runs migrations, and deploys to staging/production via Kustomize/Helm/ArgoCD.
5. **No container image registry.** Define where images are pushed (ECR, GHCR, Docker Hub) and how tags are managed.
6. **No observability stack.** Consider adding `loki`, `prometheus`, `grafana`, or cloud-native equivalents for logs/metrics.
7. **No dependency/security scanning.** Add `dependabot`, `snyk`, `trivy`, or GitHub Advanced Security scans.
8. **Environment variable drift.** Align `.env.example` with the code: rename `JWT_EXPIPY_HOURS` to `JWT_EXPIRY_MINUTES` and remove unused variables.
9. **No staging/production secrets management.** Document how `JWT_SECRET`, `DATABASE_URL`, SMTP credentials, and S3 keys are injected in each environment.
10. **Web `.env.example` missing.** Add `apps/web/.env.example` with `NEXT_PUBLIC_API_URL` and any public keys.
11. **CI does not test the web build with API dependency.** The web build may pass type-check but integration tests against a real backend are absent.
12. **Worker is a stub.** Decide whether `apps/worker` is needed; if so, implement and add a CI/CD job.
