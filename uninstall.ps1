$ErrorActionPreference = "Stop"

Write-Host "╔══════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║                Uninstalling TaaNOS CLI                   ║" -ForegroundColor Cyan
Write-Host "╚══════════════════════════════════════════════════════════╝" -ForegroundColor Cyan

$InstallDir = "$env:USERPROFILE\.taanos\bin"
$ConfigDir = "$env:USERPROFILE\.taanos"

if (Test-Path -Path $ConfigDir) {
    Write-Host "🗑️ Removing TaaNOS configuration and history ($ConfigDir)..." -ForegroundColor Yellow
    Remove-Item -Path $ConfigDir -Recurse -Force
} else {
    Write-Host "✅ TaaNOS configuration directory not found." -ForegroundColor Green
}

Write-Host "⚙️ Removing TaaNOS from your PATH..." -ForegroundColor Yellow
$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($UserPath -match [regex]::Escape($InstallDir)) {
    # Remove the path and any trailing/leading semicolons
    $NewPath = $UserPath -replace [regex]::Escape($InstallDir) + ";?", ""
    $NewPath = $NewPath -replace ";$", ""
    
    [Environment]::SetEnvironmentVariable("PATH", $NewPath, "User")
    Write-Host "✅ PATH updated. You may need to restart your terminal for changes to take effect." -ForegroundColor Green
} else {
    Write-Host "✅ TaaNOS was not in your PATH." -ForegroundColor Green
}

Write-Host ""
Write-Host "✅ Uninstallation complete!" -ForegroundColor Green
