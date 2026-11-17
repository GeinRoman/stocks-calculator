$ErrorActionPreference = 'Stop'
Set-Location -LiteralPath $PSScriptRoot

if (!(Test-Path -LiteralPath 'release.env' -PathType Leaf) -or !(Test-Path -LiteralPath 'release.config.yaml' -PathType Leaf)) {
    throw 'Create release.env and release.config.yaml first; see docs/release.en.md.'
}
$config = Get-Content -LiteralPath 'release.config.yaml' -Raw
if ($config -notmatch '(?m)^[ \t]*jwt_key:[ \t]*(?:"[0-9a-fA-F]{64,}"|[0-9a-fA-F]{64,})[ \t]*\r?$') {
    throw 'Set jwt_key to at least 32 random bytes encoded as hex in release.config.yaml; see docs/release.en.md.'
}
if (!(Get-Command docker -ErrorAction SilentlyContinue)) {
    throw 'Docker and the Docker Compose plugin are required.'
}
docker compose version *> $null
if ($LASTEXITCODE -ne 0) { throw 'Docker Compose is required.' }
docker compose --env-file release.env -p stocks-calculator-release -f compose.release.yaml up -d --build
if ($LASTEXITCODE -ne 0) { throw 'Deployment failed. Check the Docker Compose output.' }
