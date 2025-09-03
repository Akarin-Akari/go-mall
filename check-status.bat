@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

:: Mall-Go电商系统状态检查脚本 (Windows版本)

echo.
echo ========================================
echo   Mall-Go电商系统状态检查
echo ========================================
echo.

:: 检查环境依赖
echo [INFO] 检查环境依赖...
echo.

:: 检查Go
go version >nul 2>&1
if not errorlevel 1 (
    for /f "tokens=3" %%i in ('go version') do (
        echo ✅ Go: %%i
    )
) else (
    echo ❌ Go: 未安装或未配置到PATH
)

:: 检查Node.js
node --version >nul 2>&1
if not errorlevel 1 (
    for /f "tokens=*" %%i in ('node --version') do (
        echo ✅ Node.js: %%i
    )
) else (
    echo ❌ Node.js: 未安装或未配置到PATH
)

:: 检查npm
npm --version >nul 2>&1
if not errorlevel 1 (
    for /f "tokens=*" %%i in ('npm --version') do (
        echo ✅ npm: v%%i
    )
) else (
    echo ❌ npm: 未安装
)

:: 检查Git
git --version >nul 2>&1
if not errorlevel 1 (
    for /f "tokens=3" %%i in ('git --version') do (
        echo ✅ Git: %%i
    )
) else (
    echo ❌ Git: 未安装或未配置到PATH
)

echo.
echo ========================================
echo   项目结构检查
echo ========================================
echo.

:: 检查项目目录
if exist "mall-go" (
    echo ✅ 后端目录: mall-go
    
    :: 检查关键文件
    if exist "mall-go\cmd\server\main.go" (
        echo   ✅ 主程序文件存在
    ) else (
        echo   ❌ 主程序文件缺失
    )
    
    if exist "mall-go\go.mod" (
        echo   ✅ Go模块文件存在
    ) else (
        echo   ❌ Go模块文件缺失
    )
    
    if exist "mall-go\mall_go.db" (
        echo   ✅ 数据库文件存在
    ) else (
        echo   ⚠️  数据库文件不存在 (首次启动时会自动创建)
    )
) else (
    echo ❌ 后端目录: mall-go 不存在
)

if exist "mall-frontend" (
    echo ✅ 前端目录: mall-frontend
    
    :: 检查关键文件
    if exist "mall-frontend\package.json" (
        echo   ✅ package.json存在
    ) else (
        echo   ❌ package.json缺失
    )
    
    if exist "mall-frontend\next.config.js" (
        echo   ✅ Next.js配置文件存在
    ) else (
        echo   ❌ Next.js配置文件缺失
    )
    
    if exist "mall-frontend\node_modules" (
        echo   ✅ 依赖包已安装
    ) else (
        echo   ⚠️  依赖包未安装 (需要运行 npm install)
    )
) else (
    echo ❌ 前端目录: mall-frontend 不存在
)

echo.
echo ========================================
echo   服务运行状态
echo ========================================
echo.

:: 检查后端服务 (端口8080)
echo [INFO] 检查后端服务状态...
netstat -ano | findstr :8080 >nul
if not errorlevel 1 (
    echo ✅ 后端服务: 正在运行 (端口8080)
    
    :: 测试健康检查API
    curl -s http://localhost:8080/health >nul 2>&1
    if not errorlevel 1 (
        echo   ✅ 健康检查API: 正常响应
        
        :: 获取API响应内容
        for /f "delims=" %%i in ('curl -s http://localhost:8080/health') do (
            echo %%i | findstr "ok" >nul
            if not errorlevel 1 (
                echo   ✅ API状态: 健康
            ) else (
                echo   ⚠️  API状态: 响应异常
            )
        )
    ) else (
        echo   ❌ 健康检查API: 无响应
    )
) else (
    echo ❌ 后端服务: 未运行 (端口8080空闲)
)

:: 检查前端服务
echo.
echo [INFO] 检查前端服务状态...
netstat -ano | findstr :3001 >nul
if not errorlevel 1 (
    echo ✅ 前端服务: 正在运行 (端口3001)
) else (
    netstat -ano | findstr :3000 >nul
    if not errorlevel 1 (
        echo ✅ 前端服务: 正在运行 (端口3000)
    ) else (
        echo ❌ 前端服务: 未运行
    )
)

echo.
echo ========================================
echo   API端点测试
echo ========================================
echo.

:: 测试核心API端点
echo [INFO] 测试核心API端点...

:: 健康检查
curl -s http://localhost:8080/health >nul 2>&1
if not errorlevel 1 (
    echo ✅ GET /health - 健康检查
) else (
    echo ❌ GET /health - 健康检查失败
)

:: 商品列表
curl -s http://localhost:8080/api/v1/products >nul 2>&1
if not errorlevel 1 (
    echo ✅ GET /api/v1/products - 商品列表
) else (
    echo ❌ GET /api/v1/products - 商品列表失败
)

:: 商品分类
curl -s http://localhost:8080/api/v1/categories >nul 2>&1
if not errorlevel 1 (
    echo ✅ GET /api/v1/categories - 商品分类
) else (
    echo ❌ GET /api/v1/categories - 商品分类失败
)

echo.
echo ========================================
echo   系统资源使用情况
echo ========================================
echo.

:: 显示内存使用情况
echo [INFO] 内存使用情况:
for /f "skip=1 tokens=4" %%i in ('wmic OS get TotalVisibleMemorySize /value') do (
    if not "%%i"=="" (
        set /a TOTAL_MEM=%%i/1024
        echo   总内存: !TOTAL_MEM! MB
    )
)

for /f "skip=1 tokens=4" %%i in ('wmic OS get FreePhysicalMemory /value') do (
    if not "%%i"=="" (
        set /a FREE_MEM=%%i/1024
        echo   可用内存: !FREE_MEM! MB
    )
)

:: 显示磁盘使用情况
echo.
echo [INFO] 磁盘使用情况:
for /f "skip=1 tokens=2,3" %%i in ('wmic logicaldisk where "DeviceID='C:'" get Size^,FreeSpace /value') do (
    if not "%%i"=="" (
        set /a FREE_DISK=%%i/1024/1024/1024
        echo   C盘可用空间: !FREE_DISK! GB
    )
    if not "%%j"=="" (
        set /a TOTAL_DISK=%%j/1024/1024/1024
        echo   C盘总空间: !TOTAL_DISK! GB
    )
)

echo.
echo ========================================
echo   状态检查完成
echo ========================================
echo.

:: 生成建议
echo [INFO] 系统建议:
echo.

:: 检查是否需要启动服务
netstat -ano | findstr :8080 >nul
if errorlevel 1 (
    echo 💡 后端服务未运行，建议执行: quick-start.bat
)

netstat -ano | findstr ":3001\|:3000" >nul
if errorlevel 1 (
    echo 💡 前端服务未运行，建议执行: quick-start.bat
)

:: 检查是否需要安装依赖
if not exist "mall-frontend\node_modules" (
    echo 💡 前端依赖未安装，建议在mall-frontend目录执行: npm install
)

echo.
echo 🔗 快速链接:
if not errorlevel 1 (
    netstat -ano | findstr :8080 >nul
    if not errorlevel 1 (
        echo   后端API: http://localhost:8080
        echo   健康检查: http://localhost:8080/health
    )
    
    netstat -ano | findstr :3001 >nul
    if not errorlevel 1 (
        echo   前端应用: http://localhost:3001
    ) else (
        netstat -ano | findstr :3000 >nul
        if not errorlevel 1 (
            echo   前端应用: http://localhost:3000
        )
    )
)

echo.
pause
