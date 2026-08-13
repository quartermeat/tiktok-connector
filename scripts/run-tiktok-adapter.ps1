$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Push-Location $root
try {
    if (-not $env:TIKTOK_USERNAME) {
        throw "Set TIKTOK_USERNAME first, for example: `$env:TIKTOK_USERNAME='your_handle'"
    }
    $adapterPython = "C:\Users\jerem\scratch\venvs\tiktok-live-adapter\Scripts\python.exe"
    if (-not (Test-Path -LiteralPath $adapterPython)) { $adapterPython = "python" }
    & $adapterPython .\adapters\tiktok_live_adapter.py
}
finally {
    Pop-Location
}
