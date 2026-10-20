#requires -Version 5.1
# Cross-compila o CLI para Windows dentro do container (motor Pure-Go, sem CGO).
$ErrorActionPreference = 'Stop'
$Repo = (Split-Path $PSScriptRoot -Parent)
$ModCache = 'wm-gomod'
$exists = docker volume ls --format '{{.Name}}' | Select-String -Pattern "^$ModCache$" -Quiet
if (-not $exists) { docker volume create $ModCache | Out-Null }
New-Item -ItemType Directory -Path "$Repo\bin" -Force | Out-Null
docker run --rm -v "${Repo}:/app" -w /app -v "${ModCache}:/go/pkg/mod" -e CGO_ENABLED=0 -e GOOS=windows -e GOARCH=amd64 golang:1.27-bookworm go build -trimpath -o bin/watermark.exe ./cmd/watermark
Write-Host "OK: bin\watermark.exe"
