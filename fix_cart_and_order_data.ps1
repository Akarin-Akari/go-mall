# Mall-Go 购物车和订单数据修复脚本
Write-Host "🔧 开始修复Mall-Go购物车和订单数据..." -ForegroundColor Green

$baseUrl = "http://localhost:8081"

# 检查服务器是否运行
try {
    $health = Invoke-RestMethod -Uri "$baseUrl/health" -Method GET
    Write-Host "✅ 服务器运行正常" -ForegroundColor Green
} catch {
    Write-Host "❌ 服务器未运行，请先启动Go服务器" -ForegroundColor Red
    exit 1
}

# 获取用户Token
Write-Host "🔑 获取用户Token..." -ForegroundColor Cyan
try {
    $loginData = @{
        username = "testuser"
        password = "password123"
    }
    $loginJson = $loginData | ConvertTo-Json
    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/users/login" -Method POST -Body $loginJson -ContentType "application/json"
    $token = $loginResponse.data.token
    Write-Host "✅ 用户Token获取成功" -ForegroundColor Green
} catch {
    Write-Host "❌ 获取用户Token失败: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

# 1. 先检查购物车状态
Write-Host "📦 检查购物车状态..." -ForegroundColor Cyan
try {
    $cartResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/cart" -Method GET -Headers $headers
    Write-Host "✅ 购物车API正常，商品数量: $($cartResponse.data.cart.item_count)" -ForegroundColor Green
} catch {
    Write-Host "⚠️ 购物车API异常: $($_.Exception.Message)" -ForegroundColor Yellow
}

# 2. 添加商品到购物车（为订单创建准备数据）
Write-Host "🛒 添加商品到购物车..." -ForegroundColor Cyan

$products = @(
    @{ product_id = 1; quantity = 2 },
    @{ product_id = 2; quantity = 1 },
    @{ product_id = 3; quantity = 3 }
)

$addedItems = @()
foreach ($product in $products) {
    try {
        $addToCartData = @{
            product_id = $product.product_id
            quantity = $product.quantity
        }
        $addToCartJson = $addToCartData | ConvertTo-Json
        $addResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/cart/add" -Method POST -Body $addToCartJson -Headers $headers
        Write-Host "✅ 添加商品 $($product.product_id) 到购物车成功" -ForegroundColor Green
        $addedItems += $addResponse.data.id
    } catch {
        Write-Host "⚠️ 添加商品 $($product.product_id) 失败: $($_.Exception.Message)" -ForegroundColor Yellow
    }
}

# 3. 再次检查购物车
Write-Host "🔍 再次检查购物车状态..." -ForegroundColor Cyan
try {
    $cartResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/cart" -Method GET -Headers $headers
    Write-Host "✅ 购物车更新成功，商品数量: $($cartResponse.data.cart.item_count)" -ForegroundColor Green
    
    if ($cartResponse.data.cart.items -and $cartResponse.data.cart.items.Count -gt 0) {
        Write-Host "📋 购物车商品列表:" -ForegroundColor Cyan
        foreach ($item in $cartResponse.data.cart.items) {
            Write-Host "   - 商品ID: $($item.product_id), 数量: $($item.quantity), 购物车项ID: $($item.id)" -ForegroundColor White
        }
        
        # 4. 尝试创建订单
        Write-Host "📝 尝试创建订单..." -ForegroundColor Cyan
        
        # 获取购物车商品项ID
        $cartItemIds = @()
        foreach ($item in $cartResponse.data.cart.items) {
            $cartItemIds += $item.id
        }
        
        $orderData = @{
            cart_item_ids = $cartItemIds
            receiver_name = "张三"
            receiver_phone = "13800138000"
            receiver_address = "某某街道123号"
            province = "北京市"
            city = "北京市"
            district = "朝阳区"
            remark = "测试订单"
        }
        
        try {
            $orderJson = $orderData | ConvertTo-Json
            $orderResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/orders" -Method POST -Body $orderJson -Headers $headers
            Write-Host "✅ 订单创建成功！订单ID: $($orderResponse.data.id)" -ForegroundColor Green
        } catch {
            Write-Host "❌ 订单创建失败: $($_.Exception.Message)" -ForegroundColor Red
            
            # 显示详细错误信息
            if ($_.Exception.Response) {
                $errorStream = $_.Exception.Response.GetResponseStream()
                $reader = New-Object System.IO.StreamReader($errorStream)
                $errorBody = $reader.ReadToEnd()
                Write-Host "错误详情: $errorBody" -ForegroundColor Red
            }
        }
    } else {
        Write-Host "⚠️ 购物车为空，无法创建订单" -ForegroundColor Yellow
    }
} catch {
    Write-Host "❌ 检查购物车失败: $($_.Exception.Message)" -ForegroundColor Red
}

# 5. 测试订单列表API
Write-Host "📋 测试订单列表API..." -ForegroundColor Cyan
try {
    $ordersResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/orders?page=1&page_size=10" -Method GET -Headers $headers
    Write-Host "✅ 订单列表API正常，订单数量: $($ordersResponse.data.total)" -ForegroundColor Green
} catch {
    Write-Host "❌ 订单列表API失败: $($_.Exception.Message)" -ForegroundColor Red
    
    # 显示详细错误信息
    if ($_.Exception.Response) {
        $errorStream = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($errorStream)
        $errorBody = $reader.ReadToEnd()
        Write-Host "错误详情: $errorBody" -ForegroundColor Red
    }
}

Write-Host "`n🎉 购物车和订单数据修复测试完成！" -ForegroundColor Green
Write-Host "📊 修复结果总结:" -ForegroundColor Cyan
Write-Host "   ✅ 购物车功能测试" -ForegroundColor White
Write-Host "   ✅ 商品添加到购物车" -ForegroundColor White
Write-Host "   🔄 订单创建功能测试" -ForegroundColor White
Write-Host "   🔄 订单列表功能测试" -ForegroundColor White
Write-Host "`n💡 如果订单功能仍有问题，请检查:" -ForegroundColor Yellow
Write-Host "   1. 数据库中是否有商品数据" -ForegroundColor White
Write-Host "   2. 购物车商品是否正确关联" -ForegroundColor White
Write-Host "   3. 订单服务的购物车查询逻辑" -ForegroundColor White
