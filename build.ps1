# TaaNOS Universal Build Script (v1.5)
# This script compiles TaaNOS and its Universal Installers.

$ErrorActionPreference = "Stop"

Write-Host "👾 Building TaaNOS v1.5 Epic Release with Installers..." -ForegroundColor Magenta

if (-not (Get-Command "go" -ErrorAction SilentlyContinue)) {
    Write-Host "Error: Go compiler is not installed or not in PATH." -ForegroundColor Red
    exit 1
}

if (-not (Test-Path "bin")) {
    New-Item -ItemType Directory -Path "bin" | Out-Null
}

$payloadPath = "cmd/installer/payload"

# 1. Build for Windows (amd64)
Write-Host "-> Compiling for Windows (amd64)..." -ForegroundColor Cyan
$env:GOOS="windows"
$env:GOARCH="amd64"
go build -o bin/taanos.exe ./cmd/taanos
if ($LASTEXITCODE -ne 0) { Write-Host "Failed to build Windows!" -ForegroundColor Red; exit 1 }

# Generate Windows Installer
Copy-Item "bin/taanos.exe" $payloadPath -Force
Write-Host "-> Packaging Windows Installer..." -ForegroundColor Yellow
go build -o bin/TaaNOS-Setup-Windows.exe ./cmd/installer
if ($LASTEXITCODE -ne 0) { Write-Host "Failed to package Windows Installer!" -ForegroundColor Red; exit 1 }


# 2. Build for Linux (amd64)
Write-Host "-> Compiling for Linux (amd64)..." -ForegroundColor Cyan
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -o bin/taanos-linux-amd64 ./cmd/taanos
if ($LASTEXITCODE -ne 0) { Write-Host "Failed to build Linux!" -ForegroundColor Red; exit 1 }

# Generate Linux Installer
Copy-Item "bin/taanos-linux-amd64" $payloadPath -Force
Write-Host "-> Packaging Linux Installer..." -ForegroundColor Yellow
go build -o bin/TaaNOS-Setup-Linux ./cmd/installer
if ($LASTEXITCODE -ne 0) { Write-Host "Failed to package Linux Installer!" -ForegroundColor Red; exit 1 }


# 3. Build for macOS (Apple Silicon / arm64)
Write-Host "-> Compiling for macOS (arm64)..." -ForegroundColor Cyan
$env:GOOS="darwin"
$env:GOARCH="arm64"
go build -o bin/taanos-darwin-arm64 ./cmd/taanos
if ($LASTEXITCODE -ne 0) { Write-Host "Failed to build macOS!" -ForegroundColor Red; exit 1 }

# Generate Mac ARM Installer
Copy-Item "bin/taanos-darwin-arm64" $payloadPath -Force
Write-Host "-> Packaging macOS (arm64) Installer..." -ForegroundColor Yellow
go build -o bin/TaaNOS-Setup-MacOS-arm64 ./cmd/installer

# 4. Build for macOS (Intel / amd64)
Write-Host "-> Compiling for macOS (amd64)..." -ForegroundColor Cyan
$env:GOOS="darwin"
$env:GOARCH="amd64"
go build -o bin/taanos-darwin-amd64 ./cmd/taanos
if ($LASTEXITCODE -ne 0) { Write-Host "Failed to build macOS amd64!" -ForegroundColor Red; exit 1 }

# Generate Mac Intel Installer
Copy-Item "bin/taanos-darwin-amd64" $payloadPath -Force
Write-Host "-> Packaging macOS (amd64) Installer..." -ForegroundColor Yellow
go build -o bin/TaaNOS-Setup-MacOS-amd64 ./cmd/installer


# Clean up
if (Test-Path $payloadPath) { Remove-Item $payloadPath }
$env:GOOS=""
$env:GOARCH=""

Write-Host "✅ All installers built successfully! Check the 'bin' folder." -ForegroundColor Green
