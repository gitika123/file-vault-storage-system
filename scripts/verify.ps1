$ErrorActionPreference = 'Stop'
$projectRoot = Split-Path -Parent $PSScriptRoot
Push-Location $projectRoot
try {
  $go = 'C:\Program Files\Go\bin\go.exe'
  if (-not (Test-Path $go)) { throw 'Go installation not found.' }
  $env:GOCACHE = Join-Path $projectRoot 'backend\.gocache'
  $env:GOMODCACHE = Join-Path $projectRoot 'backend\.gomodcache'
  & $go -C backend test ./...
  if ($LASTEXITCODE -ne 0) { throw "Backend verification failed with exit code $LASTEXITCODE." }
  npm run build --prefix frontend
  if ($LASTEXITCODE -ne 0) { throw "Frontend verification failed with exit code $LASTEXITCODE." }
  Write-Output 'Native verification passed.'
  Write-Output 'Docker verification requires Docker Desktop: docker compose config && docker compose up --build.'
} finally { Pop-Location }
