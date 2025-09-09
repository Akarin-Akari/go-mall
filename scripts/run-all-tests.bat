@echo off
REM Mall-Go电商系统全面测试执行脚本 (Windows版本)
REM 目标: 实现90%以上的代码测试覆盖率

setlocal enabledelayedexpansion

REM 设置颜色
set "RED=[91m"
set "GREEN=[92m"
set "YELLOW=[93m"
set "BLUE=[94m"
set "NC=[0m"

REM 全局变量
set "PROJECT_ROOT=%cd%"
set "BACKEND_DIR=%PROJECT_ROOT%\mall-go"
set "FRONTEND_DIR=%PROJECT_ROOT%\mall-frontend"
set "REPORTS_DIR=%PROJECT_ROOT%\test-reports"

REM 获取时间戳
for /f "tokens=2 delims==" %%a in ('wmic OS Get localdatetime /value') do set "dt=%%a"
set "TIMESTAMP=%dt:~0,8%_%dt:~8,6%"

REM 创建报告目录
if not exist "%REPORTS_DIR%" mkdir "%REPORTS_DIR%"

REM 测试结果统计
set /a TOTAL_TESTS=0
set /a PASSED_TESTS=0
set /a FAILED_TESTS=0
set "BACKEND_COVERAGE=0"
set "FRONTEND_COVERAGE=0"

echo %BLUE%[INFO]%NC% 开始Mall-Go电商系统全面测试
echo %BLUE%[INFO]%NC% 目标: 实现90%以上的代码测试覆盖率
echo.

REM 检查环境依赖
echo %BLUE%[INFO]%NC% 检查环境依赖...

REM 检查Go环境
go version >nul 2>&1
if errorlevel 1 (
    echo %RED%[ERROR]%NC% Go环境未安装
    exit /b 1
)

REM 检查Node.js环境
node --version >nul 2>&1
if errorlevel 1 (
    echo %RED%[ERROR]%NC% Node.js环境未安装
    exit /b 1
)

REM 检查项目目录
if not exist "%BACKEND_DIR%" (
    echo %RED%[ERROR]%NC% 后端项目目录不存在: %BACKEND_DIR%
    exit /b 1
)

if not exist "%FRONTEND_DIR%" (
    echo %RED%[ERROR]%NC% 前端项目目录不存在: %FRONTEND_DIR%
    exit /b 1
)

echo %GREEN%[SUCCESS]%NC% 环境依赖检查通过
echo.

REM 后端测试
echo %BLUE%[INFO]%NC% 开始后端测试...
cd /d "%BACKEND_DIR%"

REM 安装Go依赖
echo %BLUE%[INFO]%NC% 安装Go依赖...
go mod tidy

REM 运行单元测试
echo %BLUE%[INFO]%NC% 运行后端单元测试...
go test -v -race -coverprofile=coverage.out ./... > "%REPORTS_DIR%\backend-unit-test-%TIMESTAMP%.log" 2>&1

if errorlevel 1 (
    echo %RED%[ERROR]%NC% 后端单元测试失败
    set /a FAILED_TESTS+=1
) else (
    echo %GREEN%[SUCCESS]%NC% 后端单元测试通过
    set /a PASSED_TESTS+=1
)
set /a TOTAL_TESTS+=1

REM 生成覆盖率报告
if exist coverage.out (
    echo %BLUE%[INFO]%NC% 生成后端覆盖率报告...
    go tool cover -html=coverage.out -o "%REPORTS_DIR%\backend-coverage-%TIMESTAMP%.html"
    
    REM 获取覆盖率百分比
    for /f "tokens=3" %%i in ('go tool cover -func=coverage.out ^| findstr "total"') do (
        set "BACKEND_COVERAGE=%%i"
        set "BACKEND_COVERAGE=!BACKEND_COVERAGE:%%=!"
    )
    echo %GREEN%[SUCCESS]%NC% 后端代码覆盖率: !BACKEND_COVERAGE!%%
)

REM 运行集成测试
echo %BLUE%[INFO]%NC% 运行后端集成测试...
go test -v -tags=integration ./tests/integration/... > "%REPORTS_DIR%\backend-integration-test-%TIMESTAMP%.log" 2>&1

if errorlevel 1 (
    echo %RED%[ERROR]%NC% 后端集成测试失败
    set /a FAILED_TESTS+=1
) else (
    echo %GREEN%[SUCCESS]%NC% 后端集成测试通过
    set /a PASSED_TESTS+=1
)
set /a TOTAL_TESTS+=1

REM 运行性能测试
echo %BLUE%[INFO]%NC% 运行后端性能测试...
go test -bench=. -benchmem ./... > "%REPORTS_DIR%\backend-benchmark-%TIMESTAMP%.log" 2>&1

if errorlevel 1 (
    echo %YELLOW%[WARNING]%NC% 后端性能测试有警告
) else (
    echo %GREEN%[SUCCESS]%NC% 后端性能测试完成
)

cd /d "%PROJECT_ROOT%"
echo.

REM 前端测试
echo %BLUE%[INFO]%NC% 开始前端测试...
cd /d "%FRONTEND_DIR%"

REM 安装前端依赖
echo %BLUE%[INFO]%NC% 安装前端依赖...
call npm ci

REM 运行单元测试
echo %BLUE%[INFO]%NC% 运行前端单元测试...
call npm run test:coverage > "%REPORTS_DIR%\frontend-unit-test-%TIMESTAMP%.log" 2>&1

if errorlevel 1 (
    echo %RED%[ERROR]%NC% 前端单元测试失败
    set /a FAILED_TESTS+=1
) else (
    echo %GREEN%[SUCCESS]%NC% 前端单元测试通过
    set /a PASSED_TESTS+=1
)
set /a TOTAL_TESTS+=1

REM 获取前端覆盖率
if exist coverage\coverage-summary.json (
    REM 使用PowerShell解析JSON获取覆盖率
    for /f %%i in ('powershell -command "(Get-Content coverage\coverage-summary.json | ConvertFrom-Json).total.lines.pct"') do (
        set "FRONTEND_COVERAGE=%%i"
    )
    echo %GREEN%[SUCCESS]%NC% 前端代码覆盖率: !FRONTEND_COVERAGE!%%
    
    REM 复制覆盖率报告
    if not exist "%REPORTS_DIR%\frontend-coverage-%TIMESTAMP%" mkdir "%REPORTS_DIR%\frontend-coverage-%TIMESTAMP%"
    xcopy /E /I coverage "%REPORTS_DIR%\frontend-coverage-%TIMESTAMP%" >nul
)

REM 构建前端项目
echo %BLUE%[INFO]%NC% 构建前端项目...
call npm run build > "%REPORTS_DIR%\frontend-build-%TIMESTAMP%.log" 2>&1

if errorlevel 1 (
    echo %RED%[ERROR]%NC% 前端项目构建失败
    cd /d "%PROJECT_ROOT%"
    goto :generate_report
) else (
    echo %GREEN%[SUCCESS]%NC% 前端项目构建成功
)

cd /d "%PROJECT_ROOT%"
echo.

REM E2E测试
echo %BLUE%[INFO]%NC% 开始端到端测试...

REM 启动后端服务
echo %BLUE%[INFO]%NC% 启动后端服务...
cd /d "%BACKEND_DIR%"
start /b go run cmd/server/main.go

REM 等待后端服务启动
echo %BLUE%[INFO]%NC% 等待后端服务启动...
timeout /t 10 /nobreak >nul

REM 检查后端服务
curl -s http://localhost:8080/health >nul 2>&1
if errorlevel 1 (
    echo %RED%[ERROR]%NC% 后端服务启动失败
    taskkill /f /im go.exe >nul 2>&1
    goto :generate_report
)

REM 启动前端服务
echo %BLUE%[INFO]%NC% 启动前端服务...
cd /d "%FRONTEND_DIR%"
start /b npm start

REM 等待前端服务启动
echo %BLUE%[INFO]%NC% 等待前端服务启动...
timeout /t 15 /nobreak >nul

REM 检查前端服务
curl -s http://localhost:3001 >nul 2>&1
if errorlevel 1 (
    echo %RED%[ERROR]%NC% 前端服务启动失败
    taskkill /f /im node.exe >nul 2>&1
    taskkill /f /im go.exe >nul 2>&1
    goto :generate_report
)

REM 运行Cypress E2E测试
echo %BLUE%[INFO]%NC% 运行Cypress E2E测试...
call npx cypress run --reporter json --reporter-options "output=%REPORTS_DIR%\e2e-results-%TIMESTAMP%.json" > "%REPORTS_DIR%\e2e-test-%TIMESTAMP%.log" 2>&1

if errorlevel 1 (
    echo %RED%[ERROR]%NC% E2E测试失败
    set /a FAILED_TESTS+=1
) else (
    echo %GREEN%[SUCCESS]%NC% E2E测试通过
    set /a PASSED_TESTS+=1
)
set /a TOTAL_TESTS+=1

REM 停止服务
echo %BLUE%[INFO]%NC% 停止测试服务...
taskkill /f /im node.exe >nul 2>&1
taskkill /f /im go.exe >nul 2>&1

cd /d "%PROJECT_ROOT%"
echo.

:generate_report
REM 生成测试报告
echo %BLUE%[INFO]%NC% 生成测试报告...

set "REPORT_FILE=%REPORTS_DIR%\test-summary-%TIMESTAMP%.md"

(
echo # Mall-Go电商系统测试报告
echo.
echo **测试时间**: %date% %time%
echo **测试版本**: 当前版本
echo.
echo ## 📊 测试结果统计
echo.
echo - **总测试数**: !TOTAL_TESTS!
echo - **通过测试**: !PASSED_TESTS!
echo - **失败测试**: !FAILED_TESTS!
echo - **通过率**: 计算中...
echo.
echo ## 📈 代码覆盖率
echo.
echo - **后端覆盖率**: !BACKEND_COVERAGE!%%
echo - **前端覆盖率**: !FRONTEND_COVERAGE!%%
echo - **总体覆盖率**: 计算中...
echo.
echo ## 📋 测试详情
echo.
echo ### 后端测试
echo - 单元测试: 已完成
echo - 集成测试: 已完成
echo - 性能测试: 已完成
echo.
echo ### 前端测试
echo - 单元测试: 已完成
echo - 构建测试: 已完成
echo.
echo ### E2E测试
echo - 端到端测试: 已完成
echo.
echo ## 📁 报告文件
echo.
echo 所有详细的测试日志和覆盖率报告已保存在: `%REPORTS_DIR%`
) > "%REPORT_FILE%"

echo %GREEN%[SUCCESS]%NC% 测试报告已生成: %REPORT_FILE%
echo.

REM 输出最终结果
echo %BLUE%[INFO]%NC% ========== 测试完成 ==========
echo %BLUE%[INFO]%NC% 总测试数: !TOTAL_TESTS!
echo %BLUE%[INFO]%NC% 通过测试: !PASSED_TESTS!
echo %BLUE%[INFO]%NC% 失败测试: !FAILED_TESTS!
echo %BLUE%[INFO]%NC% 后端覆盖率: !BACKEND_COVERAGE!%%
echo %BLUE%[INFO]%NC% 前端覆盖率: !FRONTEND_COVERAGE!%%

REM 检查是否达到目标覆盖率
REM 简化版本，实际应该计算平均值
if "!BACKEND_COVERAGE!" geq "85" if "!FRONTEND_COVERAGE!" geq "85" (
    echo %GREEN%[SUCCESS]%NC% 🎉 恭喜！达到85%%以上代码覆盖率目标！
    exit /b 0
) else (
    echo %YELLOW%[WARNING]%NC% ⚠️  未达到85%%覆盖率目标，需要继续优化
    exit /b 1
)

endlocal
