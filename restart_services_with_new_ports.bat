@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

:: Mall-Go电商系统端口配置更新后的重启脚本

echo.
echo ========================================
echo   Mall-Go端口配置更新重启脚本
echo ========================================
echo.
echo 📋 端口配置更新说明:
echo   - 后端服务: 8080 → 8081
echo   - 前端服务: 3001 → 3000
echo.

:: 停止现有服务
echo [INFO] 正在停止现有服务...

:: 停止占用8080和8081端口的进程
echo [INFO] 停止后端服务...
for /f "tokens=5" %%i in ('netstat -ano ^| findstr ":8080\|:8081"') do (
    set PID=%%i
    if defined PID (
        tasklist /fi "pid eq !PID!" | findstr "go.exe\|main.exe" >nul
        if not errorlevel 1 (
            echo [INFO] 停止后端进程 PID: !PID!
            taskkill /pid !PID! /f >nul 2>&1
        )
    )
)

:: 停止占用3000和3001端口的进程
echo [INFO] 停止前端服务...
for /f "tokens=5" %%i in ('netstat -ano ^| findstr ":3000\|:3001"') do (
    set PID=%%i
    if defined PID (
        tasklist /fi "pid eq !PID!" | findstr "node.exe" >nul
        if not errorlevel 1 (
            echo [INFO] 停止前端进程 PID: !PID!
            taskkill /pid !PID! /f >nul 2>&1
        )
    )
)

:: 等待进程完全停止
echo [INFO] 等待进程完全停止...
timeout /t 3 /nobreak >nul

:: 验证端口是否已释放
echo [INFO] 验证端口释放状态...
netstat -ano | findstr ":8080\|:8081\|:3000\|:3001" >nul
if errorlevel 1 (
    echo [SUCCESS] 所有相关端口已释放
) else (
    echo [WARNING] 部分端口仍被占用，可能影响服务启动
)

echo.
echo ========================================
echo   启动更新后的服务
echo ========================================
echo.

:: 启动后端服务（新端口8081）
echo [INFO] 启动后端服务 (端口8081)...
cd /d "%~dp0mall-go"
if not exist "cmd\server\main.go" (
    echo [ERROR] 后端项目文件不存在，请检查目录结构
    pause
    exit /b 1
)

start "Mall-Go Backend (8081)" cmd /k "go run cmd/server/main.go"
echo [SUCCESS] 后端服务启动命令已执行

:: 等待后端服务启动
echo [INFO] 等待后端服务启动...
timeout /t 5 /nobreak >nul

:: 验证后端服务
curl -s http://localhost:8081/health >nul 2>&1
if not errorlevel 1 (
    echo [SUCCESS] 后端服务8081端口启动成功
) else (
    echo [WARNING] 后端服务可能仍在启动中，请稍后验证
)

:: 启动前端服务（新端口3000）
echo.
echo [INFO] 启动前端服务 (端口3000)...
cd /d "%~dp0mall-frontend"
if not exist "package.json" (
    echo [ERROR] 前端项目文件不存在，请检查目录结构
    pause
    exit /b 1
)

start "Mall-Go Frontend (3000)" cmd /k "npm run dev"
echo [SUCCESS] 前端服务启动命令已执行

echo.
echo ========================================
echo   服务启动完成
echo ========================================
echo.
echo 📊 服务访问地址:
echo   - 后端API: http://localhost:8081
echo   - 前端Web: http://localhost:3000
echo   - 健康检查: http://localhost:8081/health
echo.
echo 💡 验证建议:
echo   1. 等待30秒让服务完全启动
echo   2. 运行 test_port_configuration.exe 验证配置
echo   3. 在浏览器中访问 http://localhost:3000
echo.
echo [INFO] 脚本执行完成，请检查服务启动状态
pause
