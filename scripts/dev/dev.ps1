$ErrorActionPreference = "Stop"

$repoRoot = Join-Path $PSScriptRoot "..\.." | Resolve-Path
Push-Location $repoRoot
try {
    pnpm dev
} finally {
    Pop-Location
}
