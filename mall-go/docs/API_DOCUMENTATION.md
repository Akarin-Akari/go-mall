# Mall-Go API 文档

## 概述

Mall-Go是一个基于Go语言开发的现代化电商后端API系统，提供完整的商城功能，包括用户管理、商品管理、购物车、订单处理和支付等核心业务功能。

- **版本**: 1.0.0
- **基础URL**: `http://localhost:8081`
- **API前缀**: `/api/v1`

## 统一响应格式

### 成功响应
```json
{
  "code": 200,
  "message": "操作成功",
  "data": {}, 
  "trace_id": "uuid",
  "timestamp": 1703123456
}
```

### 错误响应
```json
{
  "code": 400,
  "message": "参数验证失败",
  "details": [
    {
      "field": "email",
      "value": "invalid-email",
      "message": "邮箱格式不正确",
      "code": "VALIDATION_FAILED"
    }
  ],
  "trace_id": "uuid",
  "timestamp": 1703123456,
  "path": "/api/v1/users/register",
  "method": "POST"
}
```

## 认证

API使用JWT Bearer Token进行身份认证：

```
Authorization: Bearer <your-jwt-token>
```

## API 接口

### 1. 用户管理

#### 1.1 用户注册
```
POST /api/v1/users/register
```

**请求体**:
```json
{
  "username": "testuser",
  "email": "test@example.com",
  "password": "password123",
  "phone": "13800138000"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "注册成功",
  "data": {
    "user_id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

#### 1.2 用户登录
```
POST /api/v1/users/login
```

**请求体**:
```json
{
  "username": "testuser",
  "password": "password123"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "user_id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2025-01-01T00:00:00Z"
  }
}
```

#### 1.3 获取用户资料
```
GET /api/v1/users/profile
Authorization: Bearer <token>
```

**响应**:
```json
{
  "code": 200,
  "message": "获取用户资料成功",
  "data": {
    "user_id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "phone": "13800138000",
    "avatar": "https://example.com/avatar.jpg",
    "created_at": "2025-01-01T00:00:00Z"
  }
}
```

### 2. 商品管理

#### 2.1 获取商品列表
```
GET /api/v1/products?page=1&page_size=10&category_id=1&keyword=手机
```

**查询参数**:
- `page`: 页码 (默认: 1)
- `page_size`: 每页数量 (默认: 10, 最大: 100)
- `category_id`: 分类ID (可选)
- `keyword`: 搜索关键词 (可选)
- `min_price`: 最低价格 (可选)
- `max_price`: 最高价格 (可选)

**响应**:
```json
{
  "code": 200,
  "message": "获取商品列表成功",
  "data": {
    "list": [
      {
        "id": 1,
        "name": "iPhone 15 Pro",
        "description": "最新款iPhone",
        "price": "8999.00",
        "original_price": "9999.00",
        "stock": 100,
        "category_id": 1,
        "images": [
          "https://example.com/image1.jpg"
        ],
        "created_at": "2025-01-01T00:00:00Z"
      }
    ],
    "total": 50,
    "page": 1,
    "page_size": 10
  }
}
```

#### 2.2 获取商品详情
```
GET /api/v1/products/{id}
```

**响应**:
```json
{
  "code": 200,
  "message": "获取商品详情成功",
  "data": {
    "id": 1,
    "name": "iPhone 15 Pro",
    "description": "最新款iPhone，配备A17 Pro芯片...",
    "price": "8999.00",
    "original_price": "9999.00",
    "stock": 100,
    "category_id": 1,
    "category_name": "手机数码",
    "images": [
      "https://example.com/image1.jpg",
      "https://example.com/image2.jpg"
    ],
    "specifications": {
      "color": ["黑色", "白色", "金色"],
      "storage": ["128GB", "256GB", "512GB"]
    },
    "reviews_count": 1250,
    "rating": 4.8,
    "created_at": "2025-01-01T00:00:00Z"
  }
}
```

### 3. 购物车管理

#### 3.1 添加商品到购物车
```
POST /api/v1/cart/add
Authorization: Bearer <token>
```

**请求体**:
```json
{
  "product_id": 1,
  "sku_id": 0,
  "quantity": 2
}
```

**响应**:
```json
{
  "code": 200,
  "message": "添加商品到购物车成功",
  "data": {
    "id": 1,
    "cart_id": 1,
    "product_id": 1,
    "sku_id": 0,
    "quantity": 2,
    "price": "8999.00",
    "product_name": "iPhone 15 Pro",
    "product_image": "https://example.com/image1.jpg"
  }
}
```

#### 3.2 获取购物车
```
GET /api/v1/cart
Authorization: Bearer <token>
```

**响应**:
```json
{
  "code": 200,
  "message": "获取购物车成功",
  "data": {
    "cart": {
      "id": 1,
      "user_id": 1,
      "status": "active",
      "item_count": 2,
      "total_qty": 3,
      "total_amount": "17998.00",
      "items": [
        {
          "id": 1,
          "product_id": 1,
          "sku_id": 0,
          "quantity": 2,
          "price": "8999.00",
          "product_name": "iPhone 15 Pro",
          "product_image": "https://example.com/image1.jpg",
          "selected": true,
          "status": "normal"
        }
      ]
    },
    "summary": {
      "item_count": 2,
      "total_qty": 3,
      "selected_count": 2,
      "selected_qty": 3,
      "total_amount": "17998.00",
      "selected_amount": "17998.00",
      "discount_amount": "0.00",
      "shipping_fee": "0.00",
      "final_amount": "17998.00"
    }
  }
}
```

#### 3.3 更新购物车商品
```
PUT /api/v1/cart/items/{id}
Authorization: Bearer <token>
```

**请求体**:
```json
{
  "quantity": 3,
  "selected": true
}
```

### 4. 订单管理

#### 4.1 创建订单
```
POST /api/v1/orders
Authorization: Bearer <token>
```

**请求体**:
```json
{
  "receiver_name": "张三",
  "receiver_phone": "13800138000",
  "receiver_address": "北京市朝阳区xxx路xxx号",
  "province": "北京市",
  "city": "北京市",
  "district": "朝阳区",
  "payment_type": "alipay",
  "buyer_message": "请尽快发货",
  "items": [
    {
      "product_id": 1,
      "sku_id": 0,
      "quantity": 1,
      "price": "8999.00"
    }
  ]
}
```

**响应**:
```json
{
  "code": 200,
  "message": "创建订单成功",
  "data": {
    "id": 1,
    "order_no": "ORD202501010001",
    "user_id": 1,
    "status": "pending_payment",
    "payment_status": "unpaid",
    "total_amount": "8999.00",
    "payable_amount": "8999.00",
    "receiver_name": "张三",
    "receiver_phone": "13800138000",
    "receiver_address": "北京市朝阳区xxx路xxx号",
    "order_time": "2025-01-01T12:00:00Z",
    "pay_expire_time": "2025-01-01T13:00:00Z",
    "order_items": [
      {
        "id": 1,
        "product_id": 1,
        "sku_id": 0,
        "quantity": 1,
        "price": "8999.00",
        "product_name": "iPhone 15 Pro",
        "product_image": "https://example.com/image1.jpg"
      }
    ]
  }
}
```

#### 4.2 获取订单列表
```
GET /api/v1/orders?page=1&page_size=10&status=pending_payment
Authorization: Bearer <token>
```

**查询参数**:
- `page`: 页码 (默认: 1)
- `page_size`: 每页数量 (默认: 10)
- `status`: 订单状态 (可选)

**响应**:
```json
{
  "code": 200,
  "message": "获取订单列表成功",
  "data": {
    "list": [
      {
        "id": 1,
        "order_no": "ORD202501010001",
        "status": "pending_payment",
        "payment_status": "unpaid",
        "total_amount": "8999.00",
        "order_time": "2025-01-01T12:00:00Z",
        "receiver_name": "张三"
      }
    ],
    "total": 5,
    "page": 1,
    "page_size": 10
  }
}
```

#### 4.3 获取订单详情
```
GET /api/v1/orders/{id}
Authorization: Bearer <token>
```

### 5. 支付管理

#### 5.1 创建支付
```
POST /api/v1/payments
Authorization: Bearer <token>
```

**请求体**:
```json
{
  "order_id": 1,
  "payment_method": "alipay",
  "amount": "8999.00",
  "return_url": "https://yourapp.com/payment/return",
  "notify_url": "https://yourapi.com/payment/notify"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "创建支付成功",
  "data": {
    "payment_id": "PAY202501010001",
    "payment_url": "https://openapi.alipay.com/gateway.do?...",
    "qr_code": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA...",
    "expires_at": "2025-01-01T13:00:00Z"
  }
}
```

#### 5.2 查询支付状态
```
GET /api/v1/payments/{payment_id}
Authorization: Bearer <token>
```

**响应**:
```json
{
  "code": 200,
  "message": "查询支付状态成功",
  "data": {
    "payment_id": "PAY202501010001",
    "order_id": 1,
    "status": "paid",
    "amount": "8999.00",
    "payment_method": "alipay",
    "transaction_id": "2025010122001000000001",
    "paid_at": "2025-01-01T12:30:00Z"
  }
}
```

## 错误代码

| 错误代码 | 描述 | HTTP状态码 |
|---------|------|-----------|
| USER_NOT_FOUND | 用户不存在 | 404 |
| USER_EXISTS | 用户已存在 | 409 |
| INVALID_PASSWORD | 密码错误 | 400 |
| TOKEN_EXPIRED | Token已过期 | 401 |
| TOKEN_INVALID | Token无效 | 401 |
| PRODUCT_NOT_FOUND | 商品不存在 | 404 |
| INSUFFICIENT_STOCK | 库存不足 | 400 |
| PRODUCT_OFFLINE | 商品已下架 | 400 |
| ORDER_NOT_FOUND | 订单不存在 | 404 |
| ORDER_STATUS_ERROR | 订单状态错误 | 400 |
| PAYMENT_FAILED | 支付失败 | 400 |
| CART_ITEM_NOT_FOUND | 购物车商品不存在 | 404 |
| CART_EMPTY | 购物车为空 | 400 |
| VALIDATION_FAILED | 参数验证失败 | 400 |
| DATABASE_ERROR | 数据库错误 | 500 |
| NETWORK_ERROR | 网络错误 | 500 |
| SERVICE_UNAVAILABLE | 服务不可用 | 503 |

## 限流

API实施了限流机制：
- 每个IP每分钟最多60次请求
- 每个用户每分钟最多100次请求
- 超出限制时返回429状态码

## 环境配置

### 开发环境
- 基础URL: `http://localhost:8081`
- 数据库: SQLite (本地文件)
- Redis: `localhost:6379` (可选)

### 生产环境
- 基础URL: `https://api.yourdomain.com`
- 数据库: MySQL/PostgreSQL
- Redis: 集群模式

## SDK和示例

### JavaScript/TypeScript示例
```javascript
// 配置API客户端
const API_BASE_URL = 'http://localhost:8081/api/v1';

class MallAPI {
  constructor(token = null) {
    this.token = token;
    this.baseURL = API_BASE_URL;
  }

  async request(method, path, data = null) {
    const headers = {
      'Content-Type': 'application/json',
    };
    
    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    const config = {
      method,
      headers,
    };

    if (data) {
      config.body = JSON.stringify(data);
    }

    const response = await fetch(`${this.baseURL}${path}`, config);
    const result = await response.json();

    if (!response.ok) {
      throw new Error(result.message || 'API请求失败');
    }

    return result;
  }

  // 用户登录
  async login(username, password) {
    return this.request('POST', '/users/login', { username, password });
  }

  // 获取商品列表
  async getProducts(params = {}) {
    const query = new URLSearchParams(params).toString();
    return this.request('GET', `/products?${query}`);
  }

  // 添加到购物车
  async addToCart(productId, quantity = 1, skuId = 0) {
    return this.request('POST', '/cart/add', {
      product_id: productId,
      sku_id: skuId,
      quantity
    });
  }
}

// 使用示例
const api = new MallAPI();

// 登录
const loginResult = await api.login('testuser', 'password123');
api.token = loginResult.data.token;

// 获取商品
const products = await api.getProducts({ page: 1, page_size: 10 });

// 添加到购物车
await api.addToCart(1, 2);
```

### Go示例
```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type APIClient struct {
    BaseURL string
    Token   string
    Client  *http.Client
}

func NewAPIClient(baseURL string) *APIClient {
    return &APIClient{
        BaseURL: baseURL,
        Client:  &http.Client{},
    }
}

func (c *APIClient) Request(method, path string, data interface{}) (*http.Response, error) {
    var body []byte
    if data != nil {
        var err error
        body, err = json.Marshal(data)
        if err != nil {
            return nil, err
        }
    }

    req, err := http.NewRequest(method, c.BaseURL+path, bytes.NewBuffer(body))
    if err != nil {
        return nil, err
    }

    req.Header.Set("Content-Type", "application/json")
    if c.Token != "" {
        req.Header.Set("Authorization", "Bearer "+c.Token)
    }

    return c.Client.Do(req)
}

// 使用示例
func main() {
    client := NewAPIClient("http://localhost:8081/api/v1")
    
    // 登录
    loginData := map[string]string{
        "username": "testuser",
        "password": "password123",
    }
    
    resp, err := client.Request("POST", "/users/login", loginData)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    
    // 处理响应...
}
```

## 更新日志

### v1.0.0 (2025-01-01)
- 初始版本发布
- 完整的用户、商品、购物车、订单功能
- 支付宝、微信支付集成
- JWT认证系统
- 完善的错误处理和响应格式

## 联系我们

- 项目仓库: https://github.com/yourorg/mall-go
- 技术支持: support@yourdomain.com
- 文档反馈: docs@yourdomain.com