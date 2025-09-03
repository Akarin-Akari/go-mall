@echo off
echo ========================================
echo Go Environment Diagnostic Script
echo ========================================
echo.

echo [1] Checking Go installation...
where go >nul 2>&1
if %errorlevel% equ 0 (
    echo [OK] Go command found in PATH
    go version
) else (
    echo [ERROR] Go command NOT found in PATH
)
echo.

echo [2] Checking environment variables...
echo GOROOT: %GOROOT%
echo GOPATH: %GOPATH%
echo.

echo [3] Checking PATH variable...
echo PATH contains:
echo %PATH% | findstr /i "go"
if %errorlevel% equ 0 (
    echo [OK] Go paths found in PATH
) else (
    echo [ERROR] No Go paths found in PATH
)
echo.

echo [4] Common Go installation locations...
if exist "C:\Go\bin\go.exe" (
    echo [OK] Found Go at: C:\Go\bin\go.exe
    C:\Go\bin\go.exe version
) else (
    echo [ERROR] Go not found at: C:\Go\bin\go.exe
)

if exist "C:\Program Files\Go\bin\go.exe" (
    echo [OK] Found Go at: C:\Program Files\Go\bin\go.exe
    "C:\Program Files\Go\bin\go.exe" version
) else (
    echo [ERROR] Go not found at: C:\Program Files\Go\bin\go.exe
)
echo.

echo [5] Suggested solutions...
echo If Go is not found:
echo 1. Download Go from: https://golang.org/dl/
echo 2. Install Go to C:\Go or C:\Program Files\Go
echo 3. Add Go\bin to your PATH environment variable
echo 4. Restart your editor/terminal
echo.

echo [6] Manual PATH setup (if needed)...
echo Run these commands in PowerShell as Administrator:
echo [Environment]::SetEnvironmentVariable("GOROOT", "C:\Go", "Machine")
echo [Environment]::SetEnvironmentVariable("PATH", $env:PATH + ";C:\Go\bin", "Machine")
echo.

echo ========================================
echo Diagnostic completed!
echo ========================================
pause
