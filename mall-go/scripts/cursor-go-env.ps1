# Cursor Terminal Go Environment Setup (Temporary)
# Run this script in Cursor terminal to temporarily enable Go commands

Write-Host "Setting up Go environment for Cursor terminal..." -ForegroundColor Cyan

# Set Go environment variables for current session
$env:GOROOT = "C:\Program Files\Go"
$env:GOPATH = "C:\Users\$env:USERNAME\go"
$env:PATH += ";C:\Program Files\Go\bin"

Write-Host "✅ Go environment configured:" -ForegroundColor Green
Write-Host "   GOROOT: $env:GOROOT" -ForegroundColor White
Write-Host "   GOPATH: $env:GOPATH" -ForegroundColor White
Write-Host ""

# Test Go installation
try {
    $version = go version
    Write-Host "✅ Go is working: $version" -ForegroundColor Green
} catch {
    Write-Host "❌ Go command failed" -ForegroundColor Red
    Write-Host "Please check if Go is installed at: C:\Program Files\Go" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "You can now use Go commands in this terminal session!" -ForegroundColor Cyan
Write-Host "Try: go version, go mod tidy, go run, etc." -ForegroundColor White
