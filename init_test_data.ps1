# Mall-Go 测试数据初始化脚本
Write-Host "🚀 开始初始化Mall-Go测试数据..." -ForegroundColor Green

# 检查数据库文件是否存在
$dbPath = "mall-go\mall_go.db"
if (-not (Test-Path $dbPath)) {
    Write-Host "❌ 数据库文件不存在: $dbPath" -ForegroundColor Red
    Write-Host "请先启动Go服务器以创建数据库文件" -ForegroundColor Yellow
    exit 1
}

Write-Host "✅ 找到数据库文件: $dbPath" -ForegroundColor Green

# 使用Go服务器的API来创建测试数据
$baseUrl = "http://localhost:8081"

# 检查服务器是否运行
try {
    $healthCheck = Invoke-RestMethod -Uri "$baseUrl/health" -Method GET
    Write-Host "✅ 服务器运行正常" -ForegroundColor Green
} catch {
    Write-Host "❌ 服务器未运行，请先启动Go服务器" -ForegroundColor Red
    exit 1
}

# 创建管理员用户
Write-Host "📝 创建管理员用户..." -ForegroundColor Cyan
try {
    $adminData = @{
        username = "admin"
        email = "admin@mall-go.com"
        password = "password123"
        nickname = "系统管理员"
    } | ConvertTo-Json

    $adminResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/users/register" -Method POST -Body $adminData -ContentType "application/json"
    Write-Host "✅ 管理员用户创建成功" -ForegroundColor Green
} catch {
    if ($_.Exception.Response.StatusCode -eq 400) {
        Write-Host "⚠️ 管理员用户可能已存在" -ForegroundColor Yellow
    } else {
        Write-Host "❌ 创建管理员用户失败: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# 创建测试用户
Write-Host "📝 创建测试用户..." -ForegroundColor Cyan
try {
    $testUserData = @{
        username = "testuser"
        email = "test@example.com"
        password = "password123"
        nickname = "测试用户"
    } | ConvertTo-Json

    $testUserResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/users/register" -Method POST -Body $testUserData -ContentType "application/json"
    Write-Host "✅ 测试用户创建成功" -ForegroundColor Green
} catch {
    if ($_.Exception.Response.StatusCode -eq 400) {
        Write-Host "⚠️ 测试用户可能已存在" -ForegroundColor Yellow
    } else {
        Write-Host "❌ 创建测试用户失败: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# 获取管理员Token
Write-Host "🔑 获取管理员Token..." -ForegroundColor Cyan
try {
    $loginData = @{
        username = "admin"
        password = "password123"
    } | ConvertTo-Json

    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/users/login" -Method POST -Body $loginData -ContentType "application/json"
    $adminToken = $loginResponse.data.token
    Write-Host "✅ 管理员Token获取成功" -ForegroundColor Green
} catch {
    Write-Host "❌ 获取管理员Token失败，使用测试用户Token" -ForegroundColor Yellow
    try {
        $testLoginData = @{
            username = "testuser"
            password = "password123"
        } | ConvertTo-Json
        $testLoginResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/users/login" -Method POST -Body $testLoginData -ContentType "application/json"
        $adminToken = $testLoginResponse.data.token
        Write-Host "✅ 使用测试用户Token" -ForegroundColor Green
    } catch {
        Write-Host "❌ 无法获取任何Token，跳过需要认证的操作" -ForegroundColor Red
        $adminToken = $null
    }
}

# 创建测试商品（如果有Token）
if ($adminToken) {
    Write-Host "📦 创建测试商品..." -ForegroundColor Cyan
    
    $products = @(
        @{ name = "iPhone 15 Pro"; description = "苹果最新旗舰手机"; price = 7999.00; stock = 50; category_id = 1 },
        @{ name = "MacBook Pro"; description = "苹果笔记本电脑"; price = 12999.00; stock = 30; category_id = 1 },
        @{ name = "Nike运动鞋"; description = "舒适透气运动鞋"; price = 599.00; stock = 100; category_id = 2 },
        @{ name = "无线蓝牙耳机"; description = "高音质无线耳机"; price = 299.00; stock = 200; category_id = 1 },
        @{ name = "智能手表"; description = "多功能智能手表"; price = 1299.00; stock = 80; category_id = 1 }
    )

    $headers = @{
        "Authorization" = "Bearer $adminToken"
        "Content-Type" = "application/json"
    }

    foreach ($product in $products) {
        try {
            $productJson = $product | ConvertTo-Json
            $productResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/products" -Method POST -Body $productJson -Headers $headers
            Write-Host "✅ 创建商品: $($product.name)" -ForegroundColor Green
        } catch {
            Write-Host "⚠️ 商品创建失败或已存在: $($product.name)" -ForegroundColor Yellow
        }
    }
}

# 测试API可用性
Write-Host "🧪 测试API可用性..." -ForegroundColor Cyan

# 测试商品列表
try {
    $productsResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/products?page=1&page_size=10" -Method GET
    Write-Host "✅ 商品列表API正常" -ForegroundColor Green
} catch {
    Write-Host "❌ 商品列表API异常" -ForegroundColor Red
}

# 测试购物车（需要Token）
if ($adminToken) {
    try {
        $cartHeaders = @{ "Authorization" = "Bearer $adminToken" }
        $cartResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/cart" -Method GET -Headers $cartHeaders
        Write-Host "✅ 购物车API正常" -ForegroundColor Green
    } catch {
        Write-Host "❌ 购物车API异常" -ForegroundColor Red
    }
}

Write-Host "`n🎉 测试数据初始化完成！" -ForegroundColor Green
Write-Host "📊 创建的测试数据:" -ForegroundColor Cyan
Write-Host "   👤 管理员用户: admin / password123" -ForegroundColor White
Write-Host "   👤 测试用户: testuser / password123" -ForegroundColor White
Write-Host "   📦 测试商品: 5个商品" -ForegroundColor White
Write-Host "`n✅ 现在可以重新运行API测试了！" -ForegroundColor Green
Write-Host "运行命令: .\mall_api_fixed_tester.exe" -ForegroundColor Yellow
