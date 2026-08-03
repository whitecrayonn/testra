$ErrorActionPreference = "Stop"

$goBin = "C:\Program Files\Go\bin"
if (Test-Path "$goBin\go.exe") {
    if ($env:Path -notlike "*$goBin*") {
        $env:Path += ";$goBin"
    }
} else {
    Write-Error "Go not found at $goBin. Install Go and try again."
    exit 1
}

$apiRoot = Join-Path $PSScriptRoot "..\..\apps\api" | Resolve-Path
Push-Location $apiRoot

function Invoke-Native {
    param([string]$Command)
    Invoke-Expression $Command
    if ($LASTEXITCODE -ne 0) {
        throw "Command failed with exit code ${LASTEXITCODE}: $Command"
    }
}

try {
    Write-Host "Running go mod tidy in $apiRoot ..."
    Invoke-Native "go mod tidy"

    if (-not (Test-Path .\bin)) {
        New-Item -ItemType Directory -Path .\bin | Out-Null
    }

    Write-Host "Building API server..."
    Invoke-Native "go build -o .\bin\api.exe .\cmd\api"

    Write-Host "Building migrator..."
    Invoke-Native "go build -o .\bin\migrator.exe .\cmd\migrator"

    Write-Host "Building worker..."
    Invoke-Native "go build -o .\bin\worker.exe .\cmd\worker"

    Write-Host "Backend build complete."
} finally {
    Pop-Location
}
