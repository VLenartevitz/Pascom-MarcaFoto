<#requires -Version 5.1
# Build do Watermark Studio (Wails v2) para Windows
# Requer: Go 1.22+, Node.js 18+, Wails v2 CLI instalado (go install github.com/wailsapp/wails/v2/cmd/wails@latest)
#>

$ErrorActionPreference = 'Stop'

$Repo = (Split-Path $PSScriptRoot -Parent)
$FrontendDir = Join-Path $Repo "frontend"
$BuildDir = Join-Path $Repo "build"
$BinDir = Join-Path $Repo "bin"

Write-Host "=== Watermark Studio - Build Wails v2 ===" -ForegroundColor Cyan

# Verificar pré-requisitos
$goVersion = go version 2>$null
if (-not $goVersion) {
    Write-Error "Go não encontrado no PATH. Instale Go 1.22+"
    exit 1
}
Write-Host "Go: $goVersion" -ForegroundColor Green

$nodeVersion = node --version 2>$null
if (-not $nodeVersion) {
    Write-Error "Node.js não encontrado no PATH. Instale Node.js 18+"
    exit 1
}
Write-Host "Node: $nodeVersion" -ForegroundColor Green

$wailsVersion = wails version 2>$null
if (-not $wailsVersion) {
    Write-Warning "Wails CLI não encontrado. Instalando..."
    go install github.com/wailsapp/wails/v2/cmd/wails@latest
    # Recarregar PATH
    $env:PATH = [Environment]::GetEnvironmentVariable("PATH","User") + ";" + [Environment]::GetEnvironmentVariable("PATH","Machine")
    $wailsVersion = wails version 2>$null
    if (-not $wailsVersion) {
        Write-Error "Falha ao instalar Wails. Verifique se $env:GOPATH\bin está no PATH."
        exit 1
    }
}
Write-Host "Wails: $wailsVersion" -ForegroundColor Green

# Instalar dependências do frontend
Write-Host "`n=== Instalando dependências do frontend ===" -ForegroundColor Cyan
Push-Location $FrontendDir
try {
    npm ci
} catch {
    Write-Host "npm ci falhou, tentando npm install..." -ForegroundColor Yellow
    npm install
}
Pop-Location

# Build do frontend (Vite)
Write-Host "`n=== Build do frontend (Vite) ===" -ForegroundColor Cyan
Push-Location $FrontendDir
npm run build
if (-not (Test-Path (Join-Path $Repo "dist"))) {
    Write-Error "Build do frontend falhou - pasta dist não criada"
    exit 1
}
Pop-Location

# Build do Wails
Write-Host "`n=== Build do Wails (Windows) ===" -ForegroundColor Cyan
Push-Location $Repo

# Limpar builds anteriores
if (Test-Path $BuildDir) { Remove-Item -Recurse -Force $BuildDir }
if (Test-Path $BinDir) { Remove-Item -Recurse -Force $BinDir }
New-Item -ItemType Directory -Path $BinDir -Force | Out-Null

# Build com Wails
wails build -platform windows/amd64 -clean -ldflags "-s -w"

if (-not (Test-Path (Join-Path $BuildDir "bin\watermark-studio.exe"))) {
    Write-Error "Build do Wails falhou - executável não encontrado"
    exit 1
}

# Copiar para bin/
Copy-Item (Join-Path $BuildDir "bin\watermark-studio.exe") (Join-Path $BinDir "watermark-studio.exe") -Force
Copy-Item (Join-Path $BuildDir "bin\watermark-studio.exe") (Join-Path $BinDir "Watermark Studio.exe") -Force

Pop-Location

Write-Host "`n=== Build Concluído! ===" -ForegroundColor Green
Write-Host "Executável: $BinDir\watermark-studio.exe" -ForegroundColor Cyan
Write-Host "Executável (com espaços): $BinDir\Watermark Studio.exe" -ForegroundColor Cyan
Write-Host "`nPara executar:" -ForegroundColor Yellow
Write-Host "  .\bin\watermark-studio.exe" -ForegroundColor White