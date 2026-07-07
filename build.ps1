# TaaNOS Universal Build Script (v1.5)
# This script compiles TaaNOS for Windows, Linux, and macOS.

$ErrorActionPreference = "Stop"

Write-Host "👾 Building TaaNOS v1.5 Epic Release..." -ForegroundColor Magenta

# Check if Go is installed
if (-not (Get-Command "go" -ErrorAction SilentlyContinue)) {
    Write-Host "Error: Go compiler is not installed or not in PATH." -ForegroundColor Red
    exit 1
}

# Create bin directory if not exists
if (-not (Test-Path "bin")) {
    New-Item -ItemType Directory -Path "bin" | Out-Null
}

# 1. Build for Windows (amd64)
Write-Host "-> Compiling for Windows (amd64)..." -ForegroundColor Cyan
$env:GOOS="windows"
$env:GOARCH="amd64"
go build -o bin/taanos.exe ./cmd/taanos
if ($LASTEXITCODE -ne 0) { Write-Host "Failed to build Windows!" -ForegroundColor Red; exit 1 }

# 2. Build for Linux (amd64)
Write-Host "-> Compiling for Linux (amd64)..." -ForegroundColor Cyan
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -o bin/taanos-linux-amd64 ./cmd/taanos
if ($LASTEXITCODE -ne 0) { Write-Host "Failed to build Linux!" -ForegroundColor Red; exit 1 }

# 3. Build for macOS (Apple Silicon / arm64)
Write-Host "-> Compiling for macOS (Apple Silicon arm64)..." -ForegroundColor Cyan
$env:GOOS="darwin"
$env:GOARCH="arm64"
go build -o bin/taanos-darwin-arm64 ./cmd/taanos
if ($LASTEXITCODE -ne 0) { Write-Host "Failed to build macOS arm64!" -ForegroundColor Red; exit 1 }

# 4. Build for macOS (Intel / amd64)
Write-Host "-> Compiling for macOS (Intel amd64)..." -ForegroundColor Cyan
$env:GOOS="darwin"
$env:GOARCH="amd64"
go build -o bin/taanos-darwin-amd64 ./cmd/taanos
if ($LASTEXITCODE -ne 0) { Write-Host "Failed to build macOS amd64!" -ForegroundColor Red; exit 1 }

# Reset environment
$env:GOOS=""
$env:GOARCH=""

# 5. Package for Linux (App Bundle)
Write-Host "-> Packaging Linux Bundle..." -ForegroundColor Cyan
if (Test-Path "install-linux.sh") {
    $zipPath = "bin/TaaNOS-Linux.zip"
    if (Test-Path $zipPath) { Remove-Item $zipPath }
    
    # Create a temporary staging directory
    $staging = "bin/TaaNOS-Linux"
    if (Test-Path $staging) { Remove-Item -Recurse -Force $staging }
    New-Item -ItemType Directory -Path $staging | Out-Null
    
    # Copy files
    Copy-Item "bin/taanos-linux-amd64" "$staging/"
    Copy-Item "install-linux.sh" "$staging/"
    if (Test-Path "icon.png") { Copy-Item "icon.png" "$staging/" }
    
    # Compress
    Compress-Archive -Path "$staging/*" -DestinationPath $zipPath
    Remove-Item -Recurse -Force $staging
    Write-Host "✅ Created Linux Bundle: $zipPath" -ForegroundColor Green
}

Write-Host "✅ All builds completed successfully! Check the 'bin' folder." -ForegroundColor Green
