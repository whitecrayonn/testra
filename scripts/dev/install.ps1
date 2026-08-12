$ErrorActionPreference = "Stop"

$hasPnpm = Get-Command pnpm -ErrorAction SilentlyContinue
if (-not $hasPnpm) {
    $hasCorepack = Get-Command corepack -ErrorAction SilentlyContinue
    if ($hasCorepack) {
        Write-Host "Enabling pnpm via corepack..."
        corepack enable
        corepack prepare pnpm@9.5.0 --activate
    } else {
        Write-Host "pnpm not found and corepack unavailable. Install Node.js 20+ or run: npm install -g pnpm"
        exit 1
    }
}

$repoRoot = Join-Path $PSScriptRoot "..\.." | Resolve-Path
Push-Location $repoRoot
try {
    Write-Host "Installing dependencies (JS + Python venv) in $repoRoot ..."
    pnpm install
    Write-Host "Done."
} finally {
    Pop-Location
}
