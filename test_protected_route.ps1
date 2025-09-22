# Test Protected Route with JWT Token

Write-Host "🛡️ Testing Protected Route Access" -ForegroundColor Green

# Use the token from previous test (you can replace this with actual token)
$token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyLCJ1c2VybmFtZSI6InVzZXIyMDI1MDkxOTA1MTcyNiIsInJvbGUiOiJ1c2VyIiwiaXNzIjoibWFsbC1nbyIsInN1YiI6InVzZXIyMDI1MDkxOTA1MTcyNiIsImV4cCI6MTc1ODMxNjY0NywibmJmIjoxNzU4MjMwMjQ3LCJpYXQiOjE3NTgyMzAyNDd9.Hqbcl7d0sF91lb0cXQTyePVIQtkFG5XEqx6s3jyKJtk"

Write-Host "Using JWT Token: $($token.Substring(0,50))..." -ForegroundColor Cyan

# Test protected profile endpoint
Write-Host "`n📋 Testing GET /api/v1/users/profile" -ForegroundColor Yellow

$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

try {
    $profileResponse = Invoke-WebRequest -Uri "http://localhost:8081/api/v1/users/profile" -Method GET -Headers $headers -UseBasicParsing
    Write-Host "✅ Profile access successful!" -ForegroundColor Green
    Write-Host "Status Code: $($profileResponse.StatusCode)" -ForegroundColor Cyan
    Write-Host "Response Body: $($profileResponse.Content)" -ForegroundColor Cyan
    
    $profileData = $profileResponse.Content | ConvertFrom-Json
    Write-Host "`nProfile Details:" -ForegroundColor Yellow
    Write-Host "- Username: $($profileData.data.username)" -ForegroundColor Cyan
    Write-Host "- Email: $($profileData.data.email)" -ForegroundColor Cyan
    Write-Host "- Role: $($profileData.data.role)" -ForegroundColor Cyan
    Write-Host "- Status: $($profileData.data.status)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Profile access failed!" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    
    if ($_.Exception.Response) {
        $statusCode = $_.Exception.Response.StatusCode
        Write-Host "Status Code: $statusCode" -ForegroundColor Red
        
        try {
            $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
            $errorBody = $reader.ReadToEnd()
            Write-Host "Error Response Body: $errorBody" -ForegroundColor Red
        } catch {
            Write-Host "Could not read error response body" -ForegroundColor Red
        }
    }
}

# Test without token (should fail)
Write-Host "`n🚫 Testing without Authorization header (should fail)" -ForegroundColor Yellow

try {
    $noAuthResponse = Invoke-WebRequest -Uri "http://localhost:8081/api/v1/users/profile" -Method GET -UseBasicParsing
    Write-Host "❌ Unexpected success - security issue!" -ForegroundColor Red
    Write-Host "Response: $($noAuthResponse.Content)" -ForegroundColor Red
} catch {
    Write-Host "✅ Correctly rejected unauthorized access!" -ForegroundColor Green
    Write-Host "Status Code: $($_.Exception.Response.StatusCode)" -ForegroundColor Cyan
}

Write-Host "`n🎯 Protected route test completed!" -ForegroundColor Green
