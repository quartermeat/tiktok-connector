$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
$wellfieldWeb = Join-Path (Split-Path -Parent $root) "wellfield\web"
$docs = Join-Path $root "docs"

New-Item -ItemType Directory -Force $docs | Out-Null
Copy-Item -Force (Join-Path $wellfieldWeb "index.html") (Join-Path $docs "index.html")
Copy-Item -Force (Join-Path $wellfieldWeb "game.wasm") (Join-Path $docs "game.wasm")
Copy-Item -Force (Join-Path $wellfieldWeb "wasm_exec.js") (Join-Path $docs "wasm_exec.js")
New-Item -ItemType File -Force (Join-Path $docs ".nojekyll") | Out-Null

$index = Join-Path $docs "index.html"
$html = Get-Content -Raw $index
$html = $html -replace '(?s)\s+const reload = new EventSource\("/livereload"\);.*?console\.warn\("Wellfield hot reload build failed\. Check the dev server output\."\);\s+\}\);', ""
Set-Content -NoNewline -Path $index -Value $html

Write-Host "Synced Wellfield static site into docs/"
