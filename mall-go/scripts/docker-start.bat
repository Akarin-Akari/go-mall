@echo off
REM GoMall Docker启动脚本 (Windows)

setlocal enabledelayedexpansion

echo ==========================================
echo GoMall Docker启动脚本
echo ==========================================
echo.

REM 检查Docker是否安装
docker --version >nul 2>&1
if errorlevel 1 (
    echo ❌ Docker未安装，请先安装Docker Desktop
    pause
    exit /b 1
)

REM 检查Docker Compose是否安装
docker-compose --version >nul 2>&1
if errorlevel 1 (
    echo ❌ Docker Compose未安装，请先安装Docker Desktop
    pause
    exit /b 1
)

REM 检查Docker守护进程是否运行
docker ps >nul 2>&1
if errorlevel 1 (
    echo ❌ Docker守护进程未运行，请启动Docker Desktop
    pause
    exit /b 1
)

echo ✅ Docker和Docker Compose已安装
echo.

REM 获取环境参数
set ENV=%1
if "%ENV%"=="" set ENV=dev

if "%ENV%"=="dev" (
    echo 🚀 启动开发环境...
    echo.
    
    docker-compose build
    if errorlevel 1 (
        echo ❌ 镜像构建失败
        pause
        exit /b 1
    )
    
    docker-compose up -d
    if errorlevel 1 (
        echo ❌ 服务启动失败
        pause
        exit /b 1
    )
    
    echo.
    echo ✅ 开发环境启动成功！
    echo.
    echo 📋 服务信息：
    docker-compose ps
    
    echo.
    echo 🌐 访问地址：
    echo   API: http://localhost:8081
    echo   健康检查: http://localhost:8081/health
    echo   Metrics: http://localhost:8081/metrics
    echo.
    echo 💾 数据库信息：
    echo   MySQL: localhost:3306 (mall_user / mall_password)
    echo   Redis: localhost:6379 (password: redis123456)
    
) else if "%ENV%"=="prod" (
    echo 🚀 启动生产环境...
    echo.
    
    if not exist ".env.prod" (
        echo ❌ .env.prod文件不存在，请创建该文件
        echo 参考: docs/docker-usage-guide.md
        pause
        exit /b 1
    )
    
    docker-compose -f docker-compose.prod.yml --env-file .env.prod build
    if errorlevel 1 (
        echo ❌ 镜像构建失败
        pause
        exit /b 1
    )
    
    docker-compose -f docker-compose.prod.yml --env-file .env.prod up -d
    if errorlevel 1 (
        echo ❌ 服务启动失败
        pause
        exit /b 1
    )
    
    echo.
    echo ✅ 生产环境启动成功！
    echo.
    echo 📋 服务信息：
    docker-compose -f docker-compose.prod.yml ps
    
) else (
    echo ❌ 无效的环境参数: %ENV%
    echo 用法: docker-start.bat [dev^|prod]
    pause
    exit /b 1
)

echo.
echo 📖 查看日志: docker-compose logs -f mall-go
echo 🛑 停止服务: docker-compose down
echo.
pause

