$ErrorActionPreference = "Stop"

Write-Host "╔══════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║                Installing TaaNOS CLI                     ║" -ForegroundColor Cyan
Write-Host "╚══════════════════════════════════════════════════════════╝" -ForegroundColor Cyan

# Define URLs
$Repo = "taasezer/TaaNOS"
$DownloadUrl = "https://github.com/$Repo/releases/latest/download/taanos-windows-amd64.exe"

# Define Target Directory
$InstallDir = "$env:USERPROFILE\.taanos\bin"
if (!(Test-Path -Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

$TargetExe = "$InstallDir\taanos.exe"

Write-Host "📥 Downloading TaaNOS from GitHub Releases..." -ForegroundColor Yellow
Invoke-WebRequest -Uri $DownloadUrl -OutFile $TargetExe

Write-Host "⚙️ Adding TaaNOS to your PATH..." -ForegroundColor Yellow
$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($UserPath -notmatch [regex]::Escape($InstallDir)) {
    $NewPath = "$UserPath;$InstallDir"
    [Environment]::SetEnvironmentVariable("PATH", $NewPath, "User")
    Write-Host "✅ PATH updated. You may need to restart your terminal for changes to take effect." -ForegroundColor Green
} else {
    Write-Host "✅ PATH already configured." -ForegroundColor Green
}

Write-Host ""
Write-Host "✅ Installation complete!" -ForegroundColor Green
Write-Host "🚀 Run 'taanos init' in a new terminal to set up your system." -ForegroundColor Cyan
