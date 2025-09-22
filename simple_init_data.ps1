# Mall-Go Simple Test Data Initialization
Write-Host "Starting Mall-Go test data initialization..." -ForegroundColor Green

$baseUrl = "http://localhost:8081"

# Check server status
try {
    $health = Invoke-RestMethod -Uri "$baseUrl/health" -Method GET
    Write-Host "Server is running" -ForegroundColor Green
} catch {
    Write-Host "Server is not running. Please start the Go server first." -ForegroundColor Red
    exit 1
}

# Create admin user
Write-Host "Creating admin user..." -ForegroundColor Cyan
try {
    $adminData = @{
        username = "admin"
        email = "admin@mall-go.com"
        password = "password123"
        nickname = "Admin User"
    }
    $adminJson = $adminData | ConvertTo-Json
    $adminResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/users/register" -Method POST -Body $adminJson -ContentType "application/json"
    Write-Host "Admin user created successfully" -ForegroundColor Green
} catch {
    Write-Host "Admin user may already exist" -ForegroundColor Yellow
}

# Create test user
Write-Host "Creating test user..." -ForegroundColor Cyan
try {
    $testData = @{
        username = "testuser"
        email = "test@example.com"
        password = "password123"
        nickname = "Test User"
    }
    $testJson = $testData | ConvertTo-Json
    $testResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/users/register" -Method POST -Body $testJson -ContentType "application/json"
    Write-Host "Test user created successfully" -ForegroundColor Green
} catch {
    Write-Host "Test user may already exist" -ForegroundColor Yellow
}

# Login as admin to get token
Write-Host "Getting admin token..." -ForegroundColor Cyan
try {
    $loginData = @{
        username = "admin"
        password = "password123"
    }
    $loginJson = $loginData | ConvertTo-Json
    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/users/login" -Method POST -Body $loginJson -ContentType "application/json"
    $token = $loginResponse.data.token
    Write-Host "Admin token obtained" -ForegroundColor Green
} catch {
    Write-Host "Failed to get admin token, trying test user..." -ForegroundColor Yellow
    try {
        $testLoginData = @{
            username = "testuser"
            password = "password123"
        }
        $testLoginJson = $testLoginData | ConvertTo-Json
        $testLoginResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/users/login" -Method POST -Body $testLoginJson -ContentType "application/json"
        $token = $testLoginResponse.data.token
        Write-Host "Test user token obtained" -ForegroundColor Green
    } catch {
        Write-Host "Failed to get any token" -ForegroundColor Red
        $token = $null
    }
}

# Create test products if we have a token
if ($token) {
    Write-Host "Creating test products..." -ForegroundColor Cyan
    
    $headers = @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }

    $products = @(
        @{ name = "iPhone 15"; description = "Latest iPhone"; price = 999.99; stock = 50; category_id = 1 },
        @{ name = "MacBook"; description = "Apple Laptop"; price = 1299.99; stock = 30; category_id = 1 },
        @{ name = "Nike Shoes"; description = "Sports Shoes"; price = 99.99; stock = 100; category_id = 2 }
    )

    foreach ($product in $products) {
        try {
            $productJson = $product | ConvertTo-Json
            $productResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/products" -Method POST -Body $productJson -Headers $headers
            Write-Host "Created product: $($product.name)" -ForegroundColor Green
        } catch {
            Write-Host "Failed to create product: $($product.name)" -ForegroundColor Yellow
        }
    }
}

# Test API endpoints
Write-Host "Testing API endpoints..." -ForegroundColor Cyan

# Test product list
try {
    $products = Invoke-RestMethod -Uri "$baseUrl/api/v1/products?page=1&page_size=10" -Method GET
    Write-Host "Product list API working" -ForegroundColor Green
} catch {
    Write-Host "Product list API failed" -ForegroundColor Red
}

# Test cart if we have token
if ($token) {
    try {
        $cartHeaders = @{ "Authorization" = "Bearer $token" }
        $cart = Invoke-RestMethod -Uri "$baseUrl/api/v1/cart" -Method GET -Headers $cartHeaders
        Write-Host "Cart API working" -ForegroundColor Green
    } catch {
        Write-Host "Cart API failed" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "Test data initialization completed!" -ForegroundColor Green
Write-Host "Test accounts:" -ForegroundColor Cyan
Write-Host "  Admin: admin / password123" -ForegroundColor White
Write-Host "  User: testuser / password123" -ForegroundColor White
Write-Host ""
Write-Host "You can now run the API tests again!" -ForegroundColor Green
