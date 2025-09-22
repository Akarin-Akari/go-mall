# Mall-Go 认证API测试脚本

Write-Host "🚀 开始测试Mall-Go认证API..." -ForegroundColor Green

# 1. 健康检查
Write-Host "`n📋 1. 健康检查测试" -ForegroundColor Yellow
$healthResponse = Invoke-RestMethod -Uri "http://localhost:8081/health" -Method GET
Write-Host "健康检查响应: $($healthResponse | ConvertTo-Json)" -ForegroundColor Cyan

# 2. 用户注册测试
Write-Host "`n🔐 2. 用户注册测试" -ForegroundColor Yellow
$registerData = @{
    username = "testuser"
    email = "test@example.com"
    password = "123456"
    nickname = "测试用户"
} | ConvertTo-Json

try {
    $registerResponse = Invoke-RestMethod -Uri "http://localhost:8081/api/v1/users/register" -Method POST -Body $registerData -ContentType "application/json"
    Write-Host "注册成功响应: $($registerResponse | ConvertTo-Json)" -ForegroundColor Green
    $token = $registerResponse.data.token
    Write-Host "获得Token: $token" -ForegroundColor Cyan
} catch {
    Write-Host "注册失败: $($_.Exception.Message)" -ForegroundColor Red
    $registerResponse = $_.Exception.Response
}

# 3. 用户登录测试
Write-Host "`n🔑 3. 用户登录测试" -ForegroundColor Yellow
$loginData = @{
    email = "test@example.com"
    password = "123456"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "http://localhost:8081/api/v1/users/login" -Method POST -Body $loginData -ContentType "application/json"
    Write-Host "登录成功响应: $($loginResponse | ConvertTo-Json)" -ForegroundColor Green
    $loginToken = $loginResponse.data.token
    Write-Host "登录Token: $loginToken" -ForegroundColor Cyan
} catch {
    Write-Host "登录失败: $($_.Exception.Message)" -ForegroundColor Red
}

# 4. 受保护接口测试（如果有token）
if ($loginToken) {
    Write-Host "`n🛡️ 4. 受保护接口测试" -ForegroundColor Yellow
    $headers = @{
        "Authorization" = "Bearer $loginToken"
        "Content-Type" = "application/json"
    }
    
    try {
        $profileResponse = Invoke-RestMethod -Uri "http://localhost:8081/api/v1/users/profile" -Method GET -Headers $headers
        Write-Host "用户信息获取成功: $($profileResponse | ConvertTo-Json)" -ForegroundColor Green
    } catch {
        Write-Host "用户信息获取失败: $($_.Exception.Message)" -ForegroundColor Red
    }
}

Write-Host "`n✅ 认证API测试完成！" -ForegroundColor Green
