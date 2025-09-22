# Debug Authentication Test with detailed error reporting

Write-Host "🔍 Debug Mall-Go Authentication Test" -ForegroundColor Green

# Generate unique username
$timestamp = Get-Date -Format "yyyyMMddHHmmss"
$uniqueUser = "user$timestamp"

Write-Host "Using unique username: $uniqueUser" -ForegroundColor Cyan

# Test Registration with detailed error handling
Write-Host "`n🔐 Testing User Registration" -ForegroundColor Yellow

$registerBody = @{
    username = $uniqueUser
    email    = "$uniqueUser@example.com"
    password = "12345678"
    nickname = "Test User $timestamp"
}

$jsonBody = $registerBody | ConvertTo-Json -Depth 10
Write-Host "Request Body:" -ForegroundColor Cyan
Write-Host $jsonBody -ForegroundColor Gray

try {
    $response = Invoke-WebRequest -Uri "http://localhost:8081/api/v1/users/register" -Method POST -Body $jsonBody -ContentType "application/json" -UseBasicParsing
    Write-Host "✅ Registration successful!" -ForegroundColor Green
    Write-Host "Status Code: $($response.StatusCode)" -ForegroundColor Cyan
    Write-Host "Response Body: $($response.Content)" -ForegroundColor Cyan
    
    $responseData = $response.Content | ConvertFrom-Json
    $token = $responseData.data.token
    Write-Host "Token received: $($token.Substring(0,30))..." -ForegroundColor Green
}
catch {
    Write-Host "❌ Registration failed!" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    
    if ($_.Exception.Response) {
        $statusCode = $_.Exception.Response.StatusCode
        Write-Host "Status Code: $statusCode" -ForegroundColor Red
        
        try {
            $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
            $errorBody = $reader.ReadToEnd()
            Write-Host "Error Response Body: $errorBody" -ForegroundColor Red
        }
        catch {
            Write-Host "Could not read error response body" -ForegroundColor Red
        }
    }
}

# Test Login with detailed error handling
Write-Host "`n🔑 Testing User Login" -ForegroundColor Yellow

$loginBody = @{
    username = $uniqueUser
    password = "12345678"
}

$loginJsonBody = $loginBody | ConvertTo-Json -Depth 10
Write-Host "Login Request Body:" -ForegroundColor Cyan
Write-Host $loginJsonBody -ForegroundColor Gray

try {
    $loginResponse = Invoke-WebRequest -Uri "http://localhost:8081/api/v1/users/login" -Method POST -Body $loginJsonBody -ContentType "application/json" -UseBasicParsing
    Write-Host "✅ Login successful!" -ForegroundColor Green
    Write-Host "Status Code: $($loginResponse.StatusCode)" -ForegroundColor Cyan
    Write-Host "Response Body: $($loginResponse.Content)" -ForegroundColor Cyan
    
    $loginData = $loginResponse.Content | ConvertFrom-Json
    $loginToken = $loginData.data.token
    Write-Host "Login Token: $($loginToken.Substring(0,30))..." -ForegroundColor Green
}
catch {
    Write-Host "❌ Login failed!" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    
    if ($_.Exception.Response) {
        $statusCode = $_.Exception.Response.StatusCode
        Write-Host "Status Code: $statusCode" -ForegroundColor Red
        
        try {
            $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
            $errorBody = $reader.ReadToEnd()
            Write-Host "Error Response Body: $errorBody" -ForegroundColor Red
        }
        catch {
            Write-Host "Could not read error response body" -ForegroundColor Red
        }
    }
}

Write-Host "`n🎯 Debug test completed!" -ForegroundColor Green
