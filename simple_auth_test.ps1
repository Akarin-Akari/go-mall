# 简化的认证API测试

Write-Host "Testing Mall-Go Authentication APIs..." -ForegroundColor Green

# 1. Health Check
Write-Host "`n1. Health Check" -ForegroundColor Yellow
try {
    $health = Invoke-RestMethod -Uri "http://localhost:8081/health" -Method GET
    Write-Host "Health: $($health.status) - $($health.message)" -ForegroundColor Green
}
catch {
    Write-Host "Health check failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# 2. User Registration
Write-Host "`n2. User Registration" -ForegroundColor Yellow
$registerBody = @{
    username = "testuser2"
    email    = "test2@example.com"
    password = "123456"
    nickname = "Test User 2"
}

try {
    $registerResponse = Invoke-RestMethod -Uri "http://localhost:8081/api/v1/users/register" -Method POST -Body ($registerBody | ConvertTo-Json) -ContentType "application/json"
    Write-Host "Registration successful!" -ForegroundColor Green
    Write-Host "User ID: $($registerResponse.data.user.id)" -ForegroundColor Cyan
    Write-Host "Token received: $($registerResponse.data.token.Length) characters" -ForegroundColor Cyan
    $token = $registerResponse.data.token
}
catch {
    $errorResponse = $_.Exception.Response
    if ($errorResponse.StatusCode -eq 400) {
        Write-Host "Registration failed - User might already exist" -ForegroundColor Yellow
    }
    else {
        Write-Host "Registration failed: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# 3. User Login
Write-Host "`n3. User Login" -ForegroundColor Yellow
$loginBody = @{
    username = "testuser2"
    password = "123456"
}

try {
    $loginResponse = Invoke-RestMethod -Uri "http://localhost:8081/api/v1/users/login" -Method POST -Body ($loginBody | ConvertTo-Json) -ContentType "application/json"
    Write-Host "Login successful!" -ForegroundColor Green
    Write-Host "User: $($loginResponse.data.user.username)" -ForegroundColor Cyan
    Write-Host "Token: $($loginResponse.data.token.Substring(0,20))..." -ForegroundColor Cyan
    $loginToken = $loginResponse.data.token
}
catch {
    Write-Host "Login failed: $($_.Exception.Message)" -ForegroundColor Red
}

# 4. Protected Route Test
if ($loginToken) {
    Write-Host "`n4. Protected Route Test" -ForegroundColor Yellow
    $headers = @{
        "Authorization" = "Bearer $loginToken"
    }
    
    try {
        $profileResponse = Invoke-RestMethod -Uri "http://localhost:8081/api/v1/users/profile" -Method GET -Headers $headers
        Write-Host "Profile access successful!" -ForegroundColor Green
        Write-Host "Profile: $($profileResponse.data.username)" -ForegroundColor Cyan
    }
    catch {
        Write-Host "Profile access failed: $($_.Exception.Message)" -ForegroundColor Red
    }
}

Write-Host "`nAuthentication API testing completed!" -ForegroundColor Green
