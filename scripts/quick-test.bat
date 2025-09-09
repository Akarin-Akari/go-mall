@echo off
REM Mall-Go快速测试脚本 - 仅运行核心测试
REM 适用于开发过程中的快速验证

echo [INFO] 开始Mall-Go快速测试...
echo.

REM 设置项目路径
set "PROJECT_ROOT=%cd%"
set "BACKEND_DIR=%PROJECT_ROOT%\mall-go"
set "FRONTEND_DIR=%PROJECT_ROOT%\mall-frontend"

REM 检查项目目录
if not exist "%BACKEND_DIR%" (
    echo [ERROR] 后端项目目录不存在: %BACKEND_DIR%
    pause
    exit /b 1
)

if not exist "%FRONTEND_DIR%" (
    echo [ERROR] 前端项目目录不存在: %FRONTEND_DIR%
    pause
    exit /b 1
)

REM 后端快速测试
echo [INFO] 运行后端单元测试...
cd /d "%BACKEND_DIR%"

REM 检查Go环境
go version >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Go环境未安装，请先安装Go
    pause
    exit /b 1
)

REM 运行Go测试
go test -v ./internal/handler/... ./pkg/...
if errorlevel 1 (
    echo [ERROR] 后端测试失败
    set "BACKEND_RESULT=FAILED"
) else (
    echo [SUCCESS] 后端测试通过
    set "BACKEND_RESULT=PASSED"
)

echo.
cd /d "%PROJECT_ROOT%"

REM 前端快速测试
echo [INFO] 运行前端单元测试...
cd /d "%FRONTEND_DIR%"

REM 检查Node.js环境
node --version >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Node.js环境未安装，请先安装Node.js
    pause
    exit /b 1
)

REM 检查是否已安装依赖
if not exist "node_modules" (
    echo [INFO] 安装前端依赖...
    call npm install
)

REM 运行前端测试
call npm test -- --watchAll=false --coverage=false
if errorlevel 1 (
    echo [ERROR] 前端测试失败
    set "FRONTEND_RESULT=FAILED"
) else (
    echo [SUCCESS] 前端测试通过
    set "FRONTEND_RESULT=PASSED"
)

cd /d "%PROJECT_ROOT%"

REM 输出测试结果
echo.
echo ========== 快速测试结果 ==========
echo 后端测试: %BACKEND_RESULT%
echo 前端测试: %FRONTEND_RESULT%
echo =====================================

REM 检查整体结果
if "%BACKEND_RESULT%"=="PASSED" if "%FRONTEND_RESULT%"=="PASSED" (
    echo [SUCCESS] 🎉 所有快速测试通过！
    exit /b 0
) else (
    echo [WARNING] ⚠️  部分测试失败，请检查详细日志
    pause
    exit /b 1
)
