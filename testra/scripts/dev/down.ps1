$ErrorActionPreference = "Stop"

# Optional: stops Docker Compose containers if using the optional Docker workflow.
# Native development (ADR-009) does not require Docker.
Write-Host "Optional: Stopping Docker Compose containers..."
docker compose -f "$PSScriptRoot\..\..\infra\docker\docker-compose.yml" down
