@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

:: Mall-Go电商系统快速启动脚本 (Windows版本)

echo.
echo ========================================
echo   Mall-Go电商系统快速启动脚本
echo ========================================
echo.

:: 检查Go环境
echo [INFO] 检查Go环境...
go version >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Go未安装或未配置到PATH，请先安装Go 1.19+
    pause
    exit /b 1
)

:: 检查Node.js环境
echo [INFO] 检查Node.js环境...
node --version >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Node.js未安装或未配置到PATH，请先安装Node.js 18+
    pause
    exit /b 1
)

:: 检查npm
echo [INFO] 检查npm环境...
npm --version >nul 2>&1
if errorlevel 1 (
    echo [ERROR] npm未安装，请检查Node.js安装
    pause
    exit /b 1
)

:: 显示版本信息
for /f "tokens=3" %%i in ('go version') do set GO_VERSION=%%i
for /f "tokens=*" %%i in ('node --version') do set NODE_VERSION=%%i
echo [INFO] Go版本: %GO_VERSION%
echo [INFO] Node.js版本: %NODE_VERSION%

:: 检查端口占用
echo [INFO] 检查端口占用情况...
netstat -ano | findstr :8080 >nul
if not errorlevel 1 (
    echo [WARNING] 端口8080已被占用，可能影响后端服务启动
)

netstat -ano | findstr :3001 >nul
if not errorlevel 1 (
    echo [WARNING] 端口3001已被占用，Next.js将自动使用其他端口
)

:: 启动后端服务
echo.
echo [INFO] 启动后端服务...
cd mall-go
if errorlevel 1 (
    echo [ERROR] mall-go目录不存在，请确认项目结构
    pause
    exit /b 1
)

echo [INFO] 安装Go依赖...
go mod tidy
if errorlevel 1 (
    echo [ERROR] Go依赖安装失败
    pause
    exit /b 1
)

echo [INFO] 启动后端API服务...
start "Mall-Go Backend" cmd /c "go run cmd/server/main.go > backend.log 2>&1"

:: 等待后端启动
echo [INFO] 等待后端服务启动...
set /a count=0
:wait_backend
timeout /t 1 /nobreak >nul
set /a count+=1
curl -s http://localhost:8080/health >nul 2>&1
if not errorlevel 1 (
    echo [SUCCESS] 后端服务启动成功！
    goto backend_ready
)
if !count! geq 30 (
    echo [ERROR] 后端服务启动超时
    pause
    exit /b 1
)
goto wait_backend

:backend_ready

:: 启动前端服务
echo.
echo [INFO] 启动前端服务...
cd ..\mall-frontend
if errorlevel 1 (
    echo [ERROR] mall-frontend目录不存在，请确认项目结构
    pause
    exit /b 1
)

echo [INFO] 安装npm依赖...
call npm install
if errorlevel 1 (
    echo [ERROR] npm依赖安装失败
    pause
    exit /b 1
)

echo [INFO] 配置环境变量...
(
echo NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
echo NEXT_PUBLIC_API_TIMEOUT=10000
echo NEXT_PUBLIC_APP_NAME=Mall-Go
echo NEXT_PUBLIC_APP_VERSION=1.0.0
echo NEXT_PUBLIC_DEBUG=true
) > .env.local

echo [INFO] 启动前端Web服务...
start "Mall-Go Frontend" cmd /c "npm run dev"

:: 等待前端启动
echo [INFO] 等待前端服务启动...
timeout /t 10 /nobreak >nul

:: 验证服务状态
echo.
echo [INFO] 验证服务状态...

:: 验证后端
curl -s http://localhost:8080/health | findstr "ok" >nul
if not errorlevel 1 (
    echo [SUCCESS] 后端API服务正常
) else (
    echo [WARNING] 后端API服务可能异常，请检查日志
)

:: 验证前端端口
netstat -ano | findstr :3001 >nul
if not errorlevel 1 (
    set FRONTEND_URL=http://localhost:3001
    echo [SUCCESS] 前端Web服务正常 ^(端口3001^)
) else (
    netstat -ano | findstr :3000 >nul
    if not errorlevel 1 (
        set FRONTEND_URL=http://localhost:3000
        echo [SUCCESS] 前端Web服务正常 ^(端口3000^)
    ) else (
        set FRONTEND_URL=http://localhost:3001
        echo [WARNING] 前端服务端口检测失败，请手动检查
    )
)

:: 显示启动信息
echo.
echo ========================================
echo   🎉 Mall-Go电商系统启动成功！
echo ========================================
echo.
echo 📋 服务信息:
echo   后端API: http://localhost:8080
echo   前端Web: %FRONTEND_URL%
echo   健康检查: http://localhost:8080/health
echo.
echo 👤 测试账号:
echo   用户名: newuser2024
echo   密码: 123456789
echo.
echo 🔧 管理命令:
echo   查看后端日志: type mall-go\backend.log
echo   停止服务: 关闭对应的命令行窗口
echo.
echo 🌐 立即体验: %FRONTEND_URL%
echo.

:: 询问是否打开浏览器
set /p open_browser="是否自动打开浏览器? (y/n): "
if /i "%open_browser%"=="y" (
    echo [INFO] 正在打开浏览器...
    start "" "%FRONTEND_URL%"
)

echo.
echo [INFO] 服务已启动，按任意键退出此脚本...
echo [INFO] 注意: 关闭此窗口不会停止服务，服务将继续在后台运行
pause >nul

:: 返回原目录
cd ..
