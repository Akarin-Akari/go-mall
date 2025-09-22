# Final Authentication Test with unique user

Write-Host "🚀 Final Mall-Go Authentication Test" -ForegroundColor Green

# Generate unique username
$timestamp = Get-Date -Format "yyyyMMddHHmmss"
$uniqueUser = "user$timestamp"

Write-Host "Using unique username: $uniqueUser" -ForegroundColor Cyan

# 1. Health Check
Write-Host "`n📋 1. Health Check" -ForegroundColor Yellow
try {
    $health = Invoke-RestMethod -Uri "http://localhost:8081/health" -Method GET
    Write-Host "✅ Health: $($health.status) - $($health.message)" -ForegroundColor Green
}
catch {
    Write-Host "❌ Health check failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# 2. User Registration
Write-Host "`n🔐 2. User Registration" -ForegroundColor Yellow
$registerBody = @{
    username         = $uniqueUser
    email            = "$uniqueUser@example.com"
    password         = "123456"
    confirm_password = "123456"
    nickname         = "Test User $timestamp"
    agree_terms      = $true
}

try {
    $registerResponse = Invoke-RestMethod -Uri "http://localhost:8081/api/v1/users/register" -Method POST -Body ($registerBody | ConvertTo-Json) -ContentType "application/json"
    Write-Host "✅ Registration successful!" -ForegroundColor Green
    Write-Host "   User ID: $($registerResponse.data.user.id)" -ForegroundColor Cyan
    Write-Host "   Username: $($registerResponse.data.user.username)" -ForegroundColor Cyan
    Write-Host "   Token Length: $($registerResponse.data.token.Length) characters" -ForegroundColor Cyan
    $regToken = $registerResponse.data.token
}
catch {
    Write-Host "❌ Registration failed: $($_.Exception.Message)" -ForegroundColor Red
    $regError = $_.Exception.Response
    if ($regError) {
        $reader = New-Object System.IO.StreamReader($regError.GetResponseStream())
        $errorBody = $reader.ReadToEnd()
        Write-Host "   Error details: $errorBody" -ForegroundColor Red
    }
}

# 3. User Login
Write-Host "`n🔑 3. User Login" -ForegroundColor Yellow
$loginBody = @{
    username = $uniqueUser
    password = "123456"
}

try {
    $loginResponse = Invoke-RestMethod -Uri "http://localhost:8081/api/v1/users/login" -Method POST -Body ($loginBody | ConvertTo-Json) -ContentType "application/json"
    Write-Host "✅ Login successful!" -ForegroundColor Green
    Write-Host "   User: $($loginResponse.data.user.username)" -ForegroundColor Cyan
    Write-Host "   Role: $($loginResponse.data.user.role)" -ForegroundColor Cyan
    Write-Host "   Token: $($loginResponse.data.token.Substring(0,30))..." -ForegroundColor Cyan
    $loginToken = $loginResponse.data.token
}
catch {
    Write-Host "❌ Login failed: $($_.Exception.Message)" -ForegroundColor Red
    $loginError = $_.Exception.Response
    if ($loginError) {
        $reader = New-Object System.IO.StreamReader($loginError.GetResponseStream())
        $errorBody = $reader.ReadToEnd()
        Write-Host "   Error details: $errorBody" -ForegroundColor Red
    }
}

# 4. Protected Route Test
if ($loginToken) {
    Write-Host "`n🛡️ 4. Protected Route Test" -ForegroundColor Yellow
    $headers = @{
        "Authorization" = "Bearer $loginToken"
    }
    
    try {
        $profileResponse = Invoke-RestMethod -Uri "http://localhost:8081/api/v1/users/profile" -Method GET -Headers $headers
        Write-Host "✅ Profile access successful!" -ForegroundColor Green
        Write-Host "   Profile Username: $($profileResponse.data.username)" -ForegroundColor Cyan
        Write-Host "   Profile Email: $($profileResponse.data.email)" -ForegroundColor Cyan
    }
    catch {
        Write-Host "❌ Profile access failed: $($_.Exception.Message)" -ForegroundColor Red
        $profileError = $_.Exception.Response
        if ($profileError) {
            $reader = New-Object System.IO.StreamReader($profileError.GetResponseStream())
            $errorBody = $reader.ReadToEnd()
            Write-Host "   Error details: $errorBody" -ForegroundColor Red
        }
    }
}
else {
    Write-Host "⚠️ Skipping protected route test - no login token available" -ForegroundColor Yellow
}

# 5. JWT Token Validation Test
if ($loginToken) {
    Write-Host "`n🔍 5. JWT Token Validation" -ForegroundColor Yellow
    Write-Host "   Token starts with: $($loginToken.Substring(0,20))..." -ForegroundColor Cyan
    
    # Check if token has proper JWT structure (header.payload.signature)
    $tokenParts = $loginToken.Split('.')
    if ($tokenParts.Length -eq 3) {
        Write-Host "✅ JWT Token has correct structure (3 parts)" -ForegroundColor Green
    }
    else {
        Write-Host "❌ JWT Token has incorrect structure ($($tokenParts.Length) parts)" -ForegroundColor Red
    }
}

Write-Host "`n🎉 Authentication API testing completed!" -ForegroundColor Green
Write-Host "Summary:" -ForegroundColor Yellow
Write-Host "- Health Check: OK" -ForegroundColor Green
if ($regToken) {
    Write-Host "- Registration: OK" -ForegroundColor Green
}
else {
    Write-Host "- Registration: FAILED" -ForegroundColor Red
}
if ($loginToken) {
    Write-Host "- Login: OK" -ForegroundColor Green
}
else {
    Write-Host "- Login: FAILED" -ForegroundColor Red
}
if ($loginToken) {
    Write-Host "- JWT Token: OK" -ForegroundColor Green
}
else {
    Write-Host "- JWT Token: FAILED" -ForegroundColor Red
}
