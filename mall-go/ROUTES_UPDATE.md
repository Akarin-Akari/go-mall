# Go后端路由更新说明

## 概述
已成功在 `mall-go/internal/handler/routes.go` 文件中添加了购物车（cart）和支付（payment）路由的注册。

## 更新内容

### 1. 导入包更新
```go
import (
    "mall-go/internal/handler/cart"      // 新增
    "mall-go/internal/handler/payment"   // 新增
    "mall-go/pkg/payment"               // 新增
    "github.com/go-redis/redis/v8"      // 新增
    // ... 其他导入
)
```

### 2. 函数签名更新
```go
// 原来
func RegisterRoutes(r *gin.Engine, db *gorm.DB)

// 更新后
func RegisterRoutes(r *gin.Engine, db *gorm.DB, rdb *redis.Client, paymentService *payment.Service)
```

### 3. 购物车路由注册
路径前缀: `/api/v1/cart`
认证要求: 所有路由都需要 `middleware.AuthMiddleware()`

| 方法 | 路径 | 处理函数 | 说明 |
|------|------|----------|------|
| GET | `/api/v1/cart` | `GetCart` | 获取购物车 |
| POST | `/api/v1/cart/add` | `AddToCart` | 添加商品到购物车 |
| PUT | `/api/v1/cart/:id` | `UpdateCartItem` | 更新购物车商品 |
| DELETE | `/api/v1/cart/:id` | `RemoveFromCart` | 从购物车移除商品 |
| DELETE | `/api/v1/cart/clear` | `ClearCart` | 清空购物车 |
| POST | `/api/v1/cart/batch` | `BatchUpdateCart` | 批量更新购物车 |
| POST | `/api/v1/cart/select-all` | `SelectAllItems` | 全选/取消全选 |
| GET | `/api/v1/cart/count` | `GetCartItemCount` | 获取购物车商品数量 |
| POST | `/api/v1/cart/sync` | `SyncCartItems` | 同步购物车商品信息 |

### 4. 支付路由注册
路径前缀: `/api/v1/payments`
认证要求: 所有路由都需要 `middleware.AuthMiddleware()`

| 方法 | 路径 | 处理函数 | 说明 |
|------|------|----------|------|
| POST | `/api/v1/payments` | `CreatePayment` | 创建支付 |
| GET | `/api/v1/payments` | `ListPayments` | 获取支付列表 |
| GET | `/api/v1/payments/:id` | `GetPaymentByID` | 根据ID获取支付详情 |
| GET | `/api/v1/payments/query` | `QueryPayment` | 查询支付状态 |

### 5. 支付回调路由（预留）
路径前缀: `/api/v1/payments/callback`
认证要求: 无需认证（第三方支付平台回调）

```go
// 预留的回调路由，需要根据实际回调handler实现
callbackGroup := v1.Group("/payments/callback")
{
    // callbackGroup.POST("/alipay", callbackHandler.AlipayCallback)
    // callbackGroup.POST("/wechat", callbackHandler.WechatCallback)
}
```

## 前后端API对接

### 前端API常量更新
已同步更新 `mall-frontend/src/constants/index.ts` 中的API端点：

```typescript
// 购物车相关 ✅ 已在Go后端routes.go中添加路由注册
CART: {
  LIST: '/api/v1/cart',
  ADD: '/api/v1/cart/add',
  UPDATE: (id: number) => `/api/v1/cart/${id}`,
  REMOVE: (id: number) => `/api/v1/cart/${id}`,
  CLEAR: '/api/v1/cart/clear',
  SYNC: '/api/v1/cart/sync',
  BATCH_UPDATE: '/api/v1/cart/batch',
  SELECT_ALL: '/api/v1/cart/select-all',
  COUNT: '/api/v1/cart/count',
},

// 支付相关 ✅ 已在Go后端routes.go中添加路由注册
PAYMENT: {
  CREATE: '/api/v1/payments',
  LIST: '/api/v1/payments',
  DETAIL: (id: number) => `/api/v1/payments/${id}`,
  QUERY: '/api/v1/payments/query',
  CALLBACK: '/api/v1/payments/callback',
  REFUND: '/api/v1/payments/refund', // 需要后端添加此路由
},
```

## 使用说明

### 1. 启动服务器时的参数更新
由于 `RegisterRoutes` 函数签名已更新，启动服务器的代码需要相应调整：

```go
// 需要传入Redis客户端和支付服务
handler.RegisterRoutes(r, db, rdb, paymentService)
```

### 2. 测试路由
可以使用提供的测试文件 `test_routes.go` 来验证路由注册：

```bash
cd mall-go
go run test_routes.go
```

### 3. API文档
访问 `http://localhost:8080/` 可以查看更新后的API端点列表。

## 注意事项

1. **Redis依赖**: 购物车功能依赖Redis缓存，如果Redis不可用，购物车handler会自动降级到纯数据库模式。

2. **支付服务**: 支付功能需要正确初始化 `payment.Service`。

3. **认证中间件**: 所有购物车和支付路由都需要用户认证。

4. **回调路由**: 支付回调路由已预留，但具体实现需要根据实际的回调handler来完成。

5. **错误处理**: 所有路由都包含完整的错误处理和响应格式化。

## 完成状态

- ✅ 购物车路由注册完成
- ✅ 支付路由注册完成  
- ✅ 前端API常量同步更新
- ✅ 路由测试文件创建
- ⚠️ 支付回调路由需要进一步实现
- ⚠️ 地址管理路由仍需实现（前端已标记）
