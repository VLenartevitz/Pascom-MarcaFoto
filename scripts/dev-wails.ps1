<#requires -Version 5.1
# Executa Watermark Studio em modo desenvolvimento (Wails v2)
# Hot-reload do frontend (Vite) + backend Go
#>

$ErrorActionPreference = 'Stop'

$Repo = (Split-Path $PSScriptRoot -Parent)
$FrontendDir = Join-Path $Repo "frontend"

Write-Host "=== Watermark Studio - Modo Desenvolvimento ===" -ForegroundColor Cyan

# Verificar Wails
$wailsVersion = wails version 2>$null
if (-not $wailsVersion) {
    Write-Error "Wails CLI não encontrado. Execute .\scripts\build-wails.ps1 primeiro para instalar."
    exit 1
}

# Instalar dependências se node_modules não existir
if (-not (Test-Path (Join-Path $FrontendDir "node_modules"))) {
    Write-Host "Instalando dependências do frontend..." -ForegroundColor Yellow
    Push-Location $FrontendDir
    npm install
    Pop-Location
}

# Executar em modo dev
Write-Host "Iniciando Wails dev server..." -ForegroundColor Green
Push-Location $Repo
wails dev
Pop-Location