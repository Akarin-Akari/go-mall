# Go Environment Setup Script for Windows
# Run this script as Administrator to permanently configure Go environment

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Go Environment Setup Script" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check if running as Administrator
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")

if (-not $isAdmin) {
    Write-Host "❌ This script must be run as Administrator!" -ForegroundColor Red
    Write-Host "Please right-click PowerShell and select 'Run as Administrator'" -ForegroundColor Yellow
    Read-Host "Press Enter to exit"
    exit 1
}

# Go installation path
$goRoot = "C:\Program Files\Go"
$goPath = "C:\Users\$env:USERNAME\go"

Write-Host "[1] Checking Go installation..." -ForegroundColor Yellow

if (Test-Path "$goRoot\bin\go.exe") {
    Write-Host "✅ Go found at: $goRoot" -ForegroundColor Green
    $version = & "$goRoot\bin\go.exe" version
    Write-Host "   Version: $version" -ForegroundColor Green
} else {
    Write-Host "❌ Go not found at: $goRoot" -ForegroundColor Red
    Write-Host "Please install Go from: https://golang.org/dl/" -ForegroundColor Yellow
    Read-Host "Press Enter to exit"
    exit 1
}

Write-Host ""
Write-Host "[2] Setting up environment variables..." -ForegroundColor Yellow

# Set GOROOT
try {
    [Environment]::SetEnvironmentVariable("GOROOT", $goRoot, "Machine")
    Write-Host "✅ GOROOT set to: $goRoot" -ForegroundColor Green
} catch {
    Write-Host "❌ Failed to set GOROOT: $($_.Exception.Message)" -ForegroundColor Red
}

# Set GOPATH
try {
    [Environment]::SetEnvironmentVariable("GOPATH", $goPath, "Machine")
    Write-Host "✅ GOPATH set to: $goPath" -ForegroundColor Green
} catch {
    Write-Host "❌ Failed to set GOPATH: $($_.Exception.Message)" -ForegroundColor Red
}

# Update PATH
Write-Host ""
Write-Host "[3] Updating PATH variable..." -ForegroundColor Yellow

try {
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", "Machine")
    $goBinPath = "$goRoot\bin"
    
    if ($currentPath -notlike "*$goBinPath*") {
        $newPath = $currentPath + ";$goBinPath"
        [Environment]::SetEnvironmentVariable("PATH", $newPath, "Machine")
        Write-Host "✅ Added Go bin to PATH: $goBinPath" -ForegroundColor Green
    } else {
        Write-Host "✅ Go bin already in PATH" -ForegroundColor Green
    }
} catch {
    Write-Host "❌ Failed to update PATH: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "[4] Verification..." -ForegroundColor Yellow

# Refresh environment variables for current session
$env:GOROOT = [Environment]::GetEnvironmentVariable("GOROOT", "Machine")
$env:GOPATH = [Environment]::GetEnvironmentVariable("GOPATH", "Machine")
$env:PATH = [Environment]::GetEnvironmentVariable("PATH", "Machine")

Write-Host "GOROOT: $env:GOROOT" -ForegroundColor Cyan
Write-Host "GOPATH: $env:GOPATH" -ForegroundColor Cyan

# Test Go command
try {
    $goVersion = & go version 2>$null
    Write-Host "✅ Go command working: $goVersion" -ForegroundColor Green
} catch {
    Write-Host "❌ Go command not working yet" -ForegroundColor Red
    Write-Host "Please restart your terminal/editor" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Setup completed!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "1. Restart Cursor editor completely" -ForegroundColor White
Write-Host "2. Open a new terminal in Cursor" -ForegroundColor White
Write-Host "3. Test with: go version" -ForegroundColor White
Write-Host "4. Navigate to your project and run: go mod tidy" -ForegroundColor White
Write-Host ""

Read-Host "Press Enter to exit"
