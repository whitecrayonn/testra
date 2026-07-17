# Testra API

The **Testra API** is a Go modular-monolith backend that exposes the `/api/v1` REST API for the Testra platform.

## What lives here

- `cmd/api/main.go` — service entry point.
- `internal/shared/server/server.go` — route and middleware wiring.
- `internal/<module>/` — domain modules (identity, organization, workspace, project, apikeys, testmanagement, results, automationhub, notification, audit, rbac).
- `migrations/` — `golang-migrate` SQL migrations.
- `Dockerfile`, `go.mod`, `go.sum` — build and dependency files.

## Running locally

See the top-level `Makefile` and `docs/engineering/LOCAL_DEVELOPMENT_GUIDE.md` for full instructions. In short:

```bash
# from the repository root
make api-run
# or
cd apps/api && go run ./cmd/api
```

## Tests

```bash
cd apps/api && go test ./...
```

## Canonical documentation

- [Engineering Handbook](docs/BIBLICAL_TESTRA.md)
- [OpenAPI Contract](docs/api/openapi/openapi.yaml)
- [Local Development Guide](docs/engineering/LOCAL_DEVELOPMENT_GUIDE.md)
- [API Design Guidelines](docs/api/API_DESIGN_GUIDELINES.md)
