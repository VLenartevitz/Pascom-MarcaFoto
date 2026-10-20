#requires -Version 5.1
# Wrapper: executa o toolchain Go dentro do container (nenhum Go nativo exigido).
param([Parameter(ValueFromRemainingArguments = $true)][string[]]$GoArgs)
$ErrorActionPreference = 'Stop'
$Repo = (Split-Path $PSScriptRoot -Parent)
$ModCache = 'wm-gomod'
$exists = docker volume ls --format '{{.Name}}' | Select-String -Pattern "^$ModCache$" -Quiet
if (-not $exists) { docker volume create $ModCache | Out-Null }
docker run --rm -v "${Repo}:/app" -w /app -v "${ModCache}:/go/pkg/mod" -e CGO_ENABLED=0 golang:1.27-bookworm go @GoArgs
