@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

:: Mall-Go电商系统服务停止脚本 (Windows版本)

echo.
echo ========================================
echo   Mall-Go电商系统服务停止脚本
echo ========================================
echo.

echo [INFO] 正在查找并停止Mall-Go相关服务...

:: 停止占用8080端口的进程（后端服务）
echo [INFO] 停止后端服务 (端口8080)...
for /f "tokens=5" %%i in ('netstat -ano ^| findstr :8080') do (
    set PID=%%i
    if defined PID (
        tasklist /fi "pid eq !PID!" | findstr "go.exe\|main.exe" >nul
        if not errorlevel 1 (
            echo [INFO] 停止进程 PID: !PID!
            taskkill /pid !PID! /f >nul 2>&1
            if not errorlevel 1 (
                echo [SUCCESS] 后端服务已停止
            ) else (
                echo [WARNING] 停止后端服务失败，PID: !PID!
            )
        )
    )
)

:: 停止占用3001端口的进程（前端服务）
echo [INFO] 停止前端服务 (端口3001)...
for /f "tokens=5" %%i in ('netstat -ano ^| findstr :3001') do (
    set PID=%%i
    if defined PID (
        tasklist /fi "pid eq !PID!" | findstr "node.exe" >nul
        if not errorlevel 1 (
            echo [INFO] 停止进程 PID: !PID!
            taskkill /pid !PID! /f >nul 2>&1
            if not errorlevel 1 (
                echo [SUCCESS] 前端服务已停止 (端口3001)
            ) else (
                echo [WARNING] 停止前端服务失败，PID: !PID!
            )
        )
    )
)

:: 停止占用3000端口的进程（备用前端端口）
echo [INFO] 检查前端服务 (端口3000)...
for /f "tokens=5" %%i in ('netstat -ano ^| findstr :3000') do (
    set PID=%%i
    if defined PID (
        tasklist /fi "pid eq !PID!" | findstr "node.exe" >nul
        if not errorlevel 1 (
            echo [INFO] 停止进程 PID: !PID!
            taskkill /pid !PID! /f >nul 2>&1
            if not errorlevel 1 (
                echo [SUCCESS] 前端服务已停止 (端口3000)
            ) else (
                echo [WARNING] 停止前端服务失败，PID: !PID!
            )
        )
    )
)

:: 停止所有相关的Node.js进程（包含mall-frontend关键字）
echo [INFO] 停止其他相关Node.js进程...
for /f "tokens=2" %%i in ('tasklist /fi "imagename eq node.exe" /fo table /nh') do (
    set PID=%%i
    if defined PID (
        for /f "tokens=*" %%j in ('wmic process where "processid=!PID!" get commandline /value ^| findstr "CommandLine"') do (
            echo %%j | findstr "mall-frontend\|next\|dev" >nul
            if not errorlevel 1 (
                echo [INFO] 停止Mall-Go相关Node.js进程 PID: !PID!
                taskkill /pid !PID! /f >nul 2>&1
            )
        )
    )
)

:: 验证服务是否已停止
echo.
echo [INFO] 验证服务停止状态...

:: 检查8080端口
netstat -ano | findstr :8080 >nul
if errorlevel 1 (
    echo [SUCCESS] 后端服务已完全停止 (端口8080空闲)
) else (
    echo [WARNING] 端口8080仍被占用，可能有其他程序使用此端口
)

:: 检查3001端口
netstat -ano | findstr :3001 >nul
if errorlevel 1 (
    echo [SUCCESS] 前端服务已完全停止 (端口3001空闲)
) else (
    echo [WARNING] 端口3001仍被占用，可能有其他程序使用此端口
)

:: 检查3000端口
netstat -ano | findstr :3000 >nul
if errorlevel 1 (
    echo [SUCCESS] 端口3000空闲
) else (
    echo [INFO] 端口3000仍被占用 (可能是其他应用)
)

:: 清理日志文件（可选）
echo.
set /p clean_logs="是否清理日志文件? (y/n): "
if /i "%clean_logs%"=="y" (
    if exist "mall-go\backend.log" (
        del "mall-go\backend.log"
        echo [INFO] 已清理后端日志文件
    )
    if exist "mall-frontend\.next" (
        rmdir /s /q "mall-frontend\.next" >nul 2>&1
        echo [INFO] 已清理前端构建缓存
    )
)

echo.
echo ========================================
echo   ✅ Mall-Go服务停止完成
echo ========================================
echo.
echo 📋 操作总结:
echo   - 后端API服务已停止
echo   - 前端Web服务已停止
echo   - 相关端口已释放
echo.
echo 🔄 重新启动服务:
echo   运行 quick-start.bat
echo.

pause
