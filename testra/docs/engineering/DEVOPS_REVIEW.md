# DevOps Review

**Scope:** Local development environment, CI/CD, and production deployment readiness.
**Status:** Review complete.

## 1. Local Development

Testra uses a native development environment (ADR-009). No Docker is used.

| Concern | Status | Notes |
|---|---|---|
| PostgreSQL 16+ | ✅ Good | Local service required |
| Redis 7+ | ✅ Good | Local service required |
| Mailpit | ✅ Good | Local binary for SMTP testing |
| MinIO | ✅ Good | Local S3-compatible object store |
| ClickHouse 24+ | ⚠️ Optional | Not wired to application code yet |
| `pnpm dev` | ✅ Good | Starts local services check, migrations, and Turbo dev |
| `.env.example` | ⚠️ Partial | `JWT_EXPIRY_HOURS` drift fixed; secrets still weak placeholders |

## 2. Build & Packaging

| Concern | Status | Notes |
|---|---|---|
| Go API build | ✅ Good | `go build ./...` in `apps/api` |
| Web build | ✅ Good | `next build` produces standalone output |
| ML service | ✅ Good | Python FastAPI with `uvicorn` |
| Worker | ❌ Stub | `apps/worker` and `apps/api/cmd/worker` are empty |
| Docker images | N/A | Docker is not used |

## 3. Production Deployment

| Concern | Status | Notes |
|---|---|---|
| Single Ubuntu VPS strategy | ✅ Approved | Documented in ADR-003, DEPLOYMENT_GUIDE |
| systemd unit files | ❌ Not present | Needed for api, web, worker, ml, nginx, postgres, redis, minio |
| nginx site config + TLS | ❌ Not present | Let's Encrypt (certbot) automation not written |
| CD pipeline | ❌ Not present | GitHub Actions only builds/tests |
| Secrets management | ❌ Not present | Production env-file scheme not documented |
| Backup / restore runbooks | ❌ Not present | `pg_dump`, `restic`, `logrotate` not documented |
| Observability | ❌ Not present | OpenTelemetry, Prometheus, Grafana, Loki not wired |

## 4. CI/CD

`.github/workflows/ci.yml` runs:

- Go `build`, `vet`, `test` in `apps/api`.
- Web `typecheck`, `build`.
- ML `ruff` lint.

Gaps:

- No compiled binary artifact or deployment job.
- No integration/E2E tests.
- No dependency/security scanning.

## 5. Recommendations

1. Create `docs/deployment/systemd/` with unit files and an nginx template.
2. Add a CD workflow that builds Linux binaries and deploys to the VPS.
3. Add `/health` endpoints and a monitoring stack.
4. Add `trivy`, `dependabot`, and `gosec` scanning.
5. Implement `apps/worker` or remove it.
