@echo off
chcp 65001 >nul
echo [INFO] Starting Mall-Go Quick Test...
echo.

REM Set project paths
set "PROJECT_ROOT=%cd%"
set "BACKEND_DIR=%PROJECT_ROOT%\mall-go"
set "FRONTEND_DIR=%PROJECT_ROOT%\mall-frontend"

REM Backend test
echo [INFO] Running backend tests...
cd /d "%BACKEND_DIR%"

REM Check Go environment
go version >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Go not installed
    exit /b 1
)

REM Run simple Go test
go test -v ./tests/simple_test.go
if errorlevel 1 (
    echo [ERROR] Backend test failed
    set "BACKEND_RESULT=FAILED"
) else (
    echo [SUCCESS] Backend test passed
    set "BACKEND_RESULT=PASSED"
)

echo.
cd /d "%PROJECT_ROOT%"

REM Frontend test
echo [INFO] Running frontend tests...
cd /d "%FRONTEND_DIR%"

REM Check Node.js environment
node --version >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Node.js not installed
    exit /b 1
)

REM Run frontend test
call npm test -- --watchAll=false --coverage=false --passWithNoTests
if errorlevel 1 (
    echo [ERROR] Frontend test failed
    set "FRONTEND_RESULT=FAILED"
) else (
    echo [SUCCESS] Frontend test passed
    set "FRONTEND_RESULT=PASSED"
)

cd /d "%PROJECT_ROOT%"

REM Output results
echo.
echo ========== Test Results ==========
echo Backend: %BACKEND_RESULT%
echo Frontend: %FRONTEND_RESULT%
echo ==================================

REM Check overall result
if "%BACKEND_RESULT%"=="PASSED" if "%FRONTEND_RESULT%"=="PASSED" (
    echo [SUCCESS] All tests passed!
    exit /b 0
) else (
    echo [WARNING] Some tests failed
    exit /b 1
)
