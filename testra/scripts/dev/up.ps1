$ErrorActionPreference = "Stop"

# Optional: starts Docker Compose for developers who prefer containerized services.
# Native development (ADR-009) does not require Docker.
Write-Host "Optional: Starting Docker Compose for local services..."
Write-Host "Native development does not require Docker (see ADR-009)."
docker compose -f "$PSScriptRoot\..\..\infra\docker\docker-compose.yml" up -d
