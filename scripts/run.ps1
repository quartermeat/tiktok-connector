$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
Push-Location $root
try {
    go run ./cmd/tiktok-connector -addr 127.0.0.1:8787
}
finally {
    Pop-Location
}
