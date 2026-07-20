# Testra

**One Platform. Every Test.**

This monorepo contains the complete Testra platform: a modular monolith Go backend, Next.js web application, Python ML inference service, shared libraries, and deployment runbooks.

## Documentation

- [Documentation Index](docs/README.md) — where every document lives and its canonical status.
- [Engineering Handbook](docs/BIBLICAL_TESTRA.md) — architecture, rules, do-not-break list, and canonical sources.
- [Roadmap](docs/engineering/ROADMAP.md) — what is complete and what is planned.
- [OpenAPI Contract](docs/api/openapi/openapi.yaml) — the authoritative HTTP API surface.
- [Documentation Reports](docs/reports/) — audit, consolidation, and release reports.

## Repository Layout

```
testra/
├── apps/
│   ├── api/        # Go modular monolith backend (API + worker + migrator)
│   ├── web/        # Next.js 15 web application
│   ├── worker/     # Optional standalone Go worker
│   └── ml/         # Python FastAPI ML inference service
├── packages/
│   ├── config/     # Shared tooling configs
│   ├── shared/     # Shared TypeScript types and utilities
│   ├── ui/         # Shared React component library
│   └── sdk/        # Official Testra TypeScript SDK
├── docs/           # OpenAPI specs, ADRs, runbooks, deployment guides
└── scripts/        # Development and automation scripts
```

## Local Development

Testra uses a **Native Development Environment** — no Docker is required. See [ADR-009](docs/architecture/adrs/ADR-009-native-development-environment.md) for the full rationale.

### Prerequisites

| Tool | Version | Notes |
|------|---------|-------|
| Node.js | 20+ | comes with corepack for pnpm |
| pnpm | 9.5+ | `corepack enable && corepack prepare pnpm@9.5.0 --activate` |
| Go | 1.23+ | [go.dev/dl](https://go.dev/dl/) |
| Python | 3.12+ | [python.org](https://www.python.org/) — required for ML service |
| PostgreSQL | 16+ | local service or binary |
| Redis | 7+ | local service or binary |
| Mailpit | latest | [binary](https://github.com/axllent/mailpit/releases) — SMTP testing |
| MinIO | latest | [binary](https://min.io/download) — S3-compatible local storage |
| ClickHouse | 24+ | **Optional** — not needed until Phase 3 |

Docker is **not used**. All services run natively on the local machine or as systemd services on a single Ubuntu VPS.

### One-command setup

```bash
pnpm install      # installs JS deps + auto-creates Python venv for ML
pnpm dev          # starts everything
```

That's it. `pnpm dev` will:

1. Check that local services (PostgreSQL, Redis) are reachable
2. Run database migrations automatically
3. Launch all four applications simultaneously via Turborepo:
   - **Go API** — `go run ./cmd/api` (or `air` for hot reload if installed)
   - **Next.js Web** — `next dev`
   - **Go Worker** — `go run ./cmd/worker` (or `air` for hot reload if installed)
   - **Python ML** — `uvicorn api.main:app --reload`

Database services (PostgreSQL, Redis, Mailpit, MinIO) must be installed and running locally before starting development. See the installation guides below.

### Installing Local Services

#### Windows

- **PostgreSQL:** Download from [postgresql.org](https://www.postgresql.org/download/windows/) or use `choco install postgresql16`
- **Redis:** Use [Memurai](https://www.memurai.com/) or WSL2 Redis
- **Mailpit:** Download binary from [mailpit releases](https://github.com/axllent/mailpit/releases)
- **MinIO:** Download binary from [min.io/download](https://min.io/download)

#### macOS

```bash
brew install postgresql@16 redis mailpit minio/stable/minio
brew services start postgresql@16 redis mailpit
```

#### Linux (Ubuntu/Debian)

```bash
sudo apt install postgresql-16 redis-server
# Mailpit and MinIO: download binaries from their respective release pages
```

### Optional: Go hot reload

Install [air](https://github.com/air-verse/air) for automatic rebuilds on file change:

```bash
go install github.com/air-verse/air@latest
```

The dev script auto-detects `air` and uses it when available. Otherwise it falls back to `go run`.

### Other commands

| Command | Description |
|---------|-------------|
| `pnpm build` | Build all apps |
| `pnpm test` | Run all tests |
| `pnpm lint` | Lint all apps |
| `pnpm typecheck` | Type-check TypeScript packages |
| `pnpm clean` | Remove build artifacts |

### Service Ports

| Service | Port | URL |
|---------|------|-----|
| Go API | 8080 | http://localhost:8080 |
| Next.js Web | 3000 | http://localhost:3000 |
| ML Service (FastAPI) | 8000 | http://localhost:8000 |
| PostgreSQL | 5432 | localhost:5432 |
| Redis | 6379 | localhost:6379 |
| ClickHouse HTTP | 8123 | http://localhost:8123 |
| ClickHouse Native | 9000 | localhost:9000 |
| MinIO S3 | 9002 | http://localhost:9002 |
| MinIO Console | 9001 | http://localhost:9001 |
| Mailpit SMTP | 1025 | localhost:1025 |
| Mailpit UI | 8025 | http://localhost:8025 |

### Environment variables

Copy `.env.example` to `.env` and adjust as needed. The Go API auto-loads `.env` via `godotenv`.

```bash
cp .env.example .env
```

## Technology Stack

- **Backend**: Go 1.24, PostgreSQL 16, Redis 7, ClickHouse 24
- **Frontend**: Next.js 15, React 18, TypeScript 5, TailwindCSS 3
- **ML**: Python 3.12, FastAPI, scikit-learn, XGBoost
- **Infrastructure**: Native local development; production target is a single Ubuntu VPS with systemd, nginx, PostgreSQL, Redis, and Let's Encrypt. Cloud-managed services and container orchestration are not planned for MVP.

## Architecture Principles

- Modular monolith with Clean Architecture boundaries
- API-first design with OpenAPI 3.1 contracts
- Multi-tenant, enterprise-ready from day one
- No external LLM dependency; transparent classical ML only
