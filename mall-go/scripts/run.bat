@echo off
chcp 65001 >nul

echo 🚀 启动 Mall Go 商城后端项目...

REM 检查Go环境
go version >nul 2>&1
if errorlevel 1 (
    echo ❌ 错误: 未找到Go环境，请先安装Go
    pause
    exit /b 1
)

echo ✅ Go环境检查通过

REM 检查配置文件
if not exist "configs\config.yaml" (
    echo ❌ 错误: 配置文件 configs\config.yaml 不存在
    pause
    exit /b 1
)

echo ✅ 配置文件检查通过

REM 安装依赖
echo 📦 安装项目依赖...
go mod tidy

if errorlevel 1 (
    echo ❌ 依赖安装失败
    pause
    exit /b 1
)

echo ✅ 依赖安装完成

REM 检查数据库连接
echo 🔍 检查数据库连接...
REM 这里可以添加数据库连接检查逻辑

REM 运行项目
echo 🎯 启动服务器...
go run cmd/server/main.go

pause
