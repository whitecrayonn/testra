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
try {
    go run .\cmd\migrator
} finally {
    Pop-Location
}
