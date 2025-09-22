Write-Host "🔍 检查前后端服务状态" -ForegroundColor Cyan
Write-Host "====================================================" -ForegroundColor Gray

# 检查后端服务 (Port 8080)
Write-Host "📡 后端服务检查 (Port 8080):" -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8080/health" -TimeoutSec 3 -ErrorAction Stop
    Write-Host "  ✅ 后端服务正在运行 (状态码: $($response.StatusCode))" -ForegroundColor Green
    $backendRunning = $true
} catch {
    Write-Host "  ❌ 后端服务未运行" -ForegroundColor Red
    Write-Host "  💡 启动命令: cd mall-go && go run cmd/server/main.go" -ForegroundColor Yellow
    $backendRunning = $false
}

Write-Host ""

# 检查前端服务
Write-Host "🌐 前端服务检查:" -ForegroundColor Yellow
$frontendPorts = @("3000", "3001", "5173", "8081")
$frontendRunning = $false

foreach ($port in $frontendPorts) {
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:$port" -TimeoutSec 2 -ErrorAction Stop
        Write-Host "  ✅ 前端服务正在端口 $port 运行 (状态码: $($response.StatusCode))" -ForegroundColor Green
        $frontendRunning = $true
        break
    } catch {
        # 继续检查下一个端口
    }
}

if (-not $frontendRunning) {
    Write-Host "  ❌ 前端服务未运行" -ForegroundColor Red
    Write-Host "  💡 启动命令: cd mall-frontend && npm run dev" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "====================================================" -ForegroundColor Gray

# 总结
Write-Host "📊 服务状态总结:" -ForegroundColor Cyan
if ($backendRunning -and $frontendRunning) {
    Write-Host "  🎉 前后端服务都在正常运行，可以进行联调测试！" -ForegroundColor Green
} elseif ($backendRunning) {
    Write-Host "  ⚠️  后端服务正常，但前端服务需要启动" -ForegroundColor Yellow
} elseif ($frontendRunning) {
    Write-Host "  ⚠️  前端服务正常，但后端服务需要启动" -ForegroundColor Yellow
} else {
    Write-Host "  ❌ 前后端服务都需要启动" -ForegroundColor Red
}

# 检查端口占用情况
Write-Host ""
Write-Host "🔍 端口占用情况:" -ForegroundColor Cyan
$ports = @("8080", "3000", "3001", "5173", "8081")
foreach ($port in $ports) {
    $netstat = netstat -an | Select-String ":$port.*LISTENING"
    if ($netstat) {
        Write-Host "  Port $port : 正在使用" -ForegroundColor Green
    } else {
        Write-Host "  Port $port : 空闲" -ForegroundColor Gray
    }
}
