# 支付系统文档

## 概述

本支付系统是一个完整的支付解决方案，支持多种支付方式（支付宝、微信支付等），提供安全可靠的支付服务。

## 功能特性

- ✅ 多支付方式支持（支付宝、微信支付、银联等）
- ✅ 支付安全验证（签名校验、防重放攻击）
- ✅ 异步回调处理
- ✅ 支付状态同步
- ✅ 退款功能
- ✅ 支付统计
- ✅ 配置管理
- ✅ 完整的单元测试和集成测试

## 系统架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   前端应用      │    │   支付网关      │    │   第三方支付    │
│                 │    │                 │    │                 │
│  - 发起支付     │───▶│  - 创建支付     │───▶│  - 支付宝       │
│  - 查询状态     │    │  - 状态查询     │    │  - 微信支付     │
│  - 申请退款     │    │  - 回调处理     │    │  - 银联支付     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │   订单系统      │
                       │                 │
                       │  - 状态同步     │
                       │  - 库存管理     │
                       │  - 订单更新     │
                       └─────────────────┘
```

## 快速开始

### 1. 环境配置

```bash
# 设置支付宝配置
export ALIPAY_APP_ID="your_app_id"
export ALIPAY_PRIVATE_KEY="your_private_key"
export ALIPAY_PUBLIC_KEY="alipay_public_key"

# 设置微信支付配置
export WECHAT_APP_ID="your_app_id"
export WECHAT_MCH_ID="your_mch_id"
export WECHAT_API_KEY="your_api_key"
```

### 2. 初始化支付服务

```go
package main

import (
    "mall-go/pkg/payment"
    "gorm.io/gorm"
)

func main() {
    // 加载配置
    config := payment.LoadConfigFromEnv()
    
    // 创建支付服务
    paymentService, err := payment.NewService(db, config)
    if err != nil {
        panic(err)
    }
    
    // 注册路由
    payment.RegisterRoutes(router, db, paymentService, alipayClient, wechatClient)
}
```

### 3. 创建支付

```go
// 创建支付请求
req := &model.PaymentCreateRequest{
    OrderID:        123,
    PaymentMethod:  model.PaymentMethodAlipay,
    Amount:         decimal.NewFromFloat(99.99),
    Subject:        "商品购买",
    Description:    "购买商品描述",
    ExpiredMinutes: 30,
}

// 调用支付服务
resp, err := paymentService.CreatePayment(req)
if err != nil {
    // 处理错误
    return
}

// 返回支付信息给前端
// resp.PaymentData 包含二维码或支付链接
```

## API 接口文档

### 基础信息

- **Base URL**: `/api/v1/payments`
- **认证方式**: Bearer Token (JWT)
- **数据格式**: JSON

### 接口列表

#### 1. 创建支付

**POST** `/api/v1/payments`

创建新的支付订单。

**请求参数:**

```json
{
  "order_id": 123,
  "payment_method": "alipay",
  "amount": "99.99",
  "subject": "商品购买",
  "description": "购买商品描述",
  "expired_minutes": 30,
  "return_url": "https://example.com/return",
  "notify_url": "https://example.com/notify"
}
```

**响应示例:**

```json
{
  "code": 200,
  "message": "创建支付成功",
  "data": {
    "payment_id": 456,
    "payment_no": "PAY20231201123456",
    "payment_method": "alipay",
    "amount": "99.99",
    "payment_data": {
      "qr_code": "https://qr.alipay.com/bax08c25",
      "payment_url": "https://openapi.alipay.com/gateway.do?..."
    },
    "expired_at": "2023-12-01T15:30:00Z",
    "created_at": "2023-12-01T15:00:00Z"
  }
}
```

#### 2. 查询支付状态

**GET** `/api/v1/payments/{id}`

根据支付ID查询支付状态。

**路径参数:**
- `id`: 支付ID

**响应示例:**

```json
{
  "code": 200,
  "message": "查询支付成功",
  "data": {
    "payment_id": 456,
    "payment_no": "PAY20231201123456",
    "order_id": 123,
    "payment_method": "alipay",
    "payment_status": "success",
    "payment_type": "order",
    "amount": "99.99",
    "actual_amount": "99.99",
    "currency": "CNY",
    "subject": "商品购买",
    "third_party_id": "2023120122001234567890",
    "paid_at": "2023-12-01T15:05:00Z",
    "created_at": "2023-12-01T15:00:00Z"
  }
}
```

#### 3. 获取支付方式列表

**GET** `/api/v1/payments/methods`

获取所有可用的支付方式。

**响应示例:**

```json
{
  "code": 200,
  "message": "获取支付方式列表成功",
  "data": [
    {
      "id": 1,
      "payment_method": "alipay",
      "is_enabled": true,
      "display_name": "支付宝",
      "display_order": 1,
      "icon": "/static/icons/alipay.png",
      "min_amount": "0.01",
      "max_amount": "50000.00"
    },
    {
      "id": 2,
      "payment_method": "wechat",
      "is_enabled": true,
      "display_name": "微信支付",
      "display_order": 2,
      "icon": "/static/icons/wechat.png",
      "min_amount": "0.01",
      "max_amount": "50000.00"
    }
  ]
}
```

#### 4. 申请退款

**POST** `/api/v1/payments/refund`

对已支付的订单申请退款。

**请求参数:**

```json
{
  "payment_id": 456,
  "refund_amount": "50.00",
  "refund_reason": "用户申请退款"
}
```

**响应示例:**

```json
{
  "code": 200,
  "message": "申请退款成功",
  "data": {
    "refund_id": 789,
    "refund_no": "REF20231201123456",
    "payment_id": 456,
    "refund_amount": "50.00",
    "refund_status": "success",
    "refund_reason": "用户申请退款",
    "created_at": "2023-12-01T16:00:00Z"
  }
}
```

#### 5. 获取支付统计

**GET** `/api/v1/payments/statistics`

获取支付统计数据。

**查询参数:**
- `start_date`: 开始日期 (YYYY-MM-DD)
- `end_date`: 结束日期 (YYYY-MM-DD)
- `payment_method`: 支付方式 (可选)
- `group_by`: 分组方式 (day/month/year)

**响应示例:**

```json
{
  "code": 200,
  "message": "获取支付统计成功",
  "data": {
    "total_amount": "10000.00",
    "total_count": 100,
    "success_amount": "9500.00",
    "success_count": 95,
    "failed_count": 5,
    "refund_amount": "500.00",
    "refund_count": 10,
    "method_stats": {
      "alipay": {
        "method": "alipay",
        "amount": "6000.00",
        "count": 60,
        "success_rate": 0.95
      },
      "wechat": {
        "method": "wechat",
        "amount": "4000.00",
        "count": 40,
        "success_rate": 0.97
      }
    }
  }
}
```

## 回调处理

### 支付宝回调

**POST** `/api/v1/payments/callback/alipay`

处理支付宝异步通知。

**Content-Type**: `application/x-www-form-urlencoded`

**回调参数** (部分重要参数):
- `out_trade_no`: 商户订单号
- `trade_no`: 支付宝交易号
- `trade_status`: 交易状态
- `total_amount`: 订单金额
- `sign`: 签名

**响应**: 
- 成功: `success`
- 失败: `fail`

### 微信支付回调

**POST** `/api/v1/payments/callback/wechat`

处理微信支付异步通知。

**Content-Type**: `application/xml`

**回调数据** (XML格式):
```xml
<xml>
  <return_code><![CDATA[SUCCESS]]></return_code>
  <result_code><![CDATA[SUCCESS]]></result_code>
  <out_trade_no><![CDATA[PAY20231201123456]]></out_trade_no>
  <transaction_id><![CDATA[4200001234567890]]></transaction_id>
  <total_fee>9999</total_fee>
  <time_end><![CDATA[20231201150500]]></time_end>
  <sign><![CDATA[signature]]></sign>
</xml>
```

**响应** (XML格式):
```xml
<!-- 成功 -->
<xml>
  <return_code><![CDATA[SUCCESS]]></return_code>
  <return_msg><![CDATA[OK]]></return_msg>
</xml>

<!-- 失败 -->
<xml>
  <return_code><![CDATA[FAIL]]></return_code>
  <return_msg><![CDATA[错误信息]]></return_msg>
</xml>
```

## 错误码说明

| 错误码 | 错误信息 | 说明 |
|--------|----------|------|
| 400 | INVALID_PAYMENT_METHOD | 无效的支付方式 |
| 400 | INVALID_AMOUNT | 无效的金额 |
| 400 | PAYMENT_EXPIRED | 支付已过期 |
| 400 | PAYMENT_ALREADY_PAID | 支付已完成 |
| 400 | INSUFFICIENT_AMOUNT | 金额不足 |
| 404 | PAYMENT_NOT_FOUND | 支付记录不存在 |
| 404 | ORDER_NOT_FOUND | 订单不存在 |
| 500 | REFUND_FAILED | 退款失败 |
| 500 | PAYMENT_FAILED | 支付失败 |

## 支付状态说明

| 状态 | 说明 |
|------|------|
| pending | 待支付 |
| paying | 支付中 |
| success | 支付成功 |
| failed | 支付失败 |
| cancelled | 已取消 |
| refunded | 已退款 |

## 安全机制

### 1. 签名验证

所有第三方支付回调都会进行签名验证，确保数据来源的可靠性。

### 2. 防重放攻击

- 时间戳验证：请求时间戳不能超过5分钟
- 随机数验证：每个随机数只能使用一次

### 3. IP白名单

回调接口支持IP白名单限制，只允许指定IP访问。

### 4. 限流保护

实现了基于IP和接口的限流机制，防止恶意请求。

## 配置说明

### 环境变量配置

```bash
# 基础配置
PAYMENT_ENVIRONMENT=prod
PAYMENT_DEBUG=false

# 支付宝配置
ALIPAY_APP_ID=your_app_id
ALIPAY_PRIVATE_KEY=your_private_key
ALIPAY_PUBLIC_KEY=alipay_public_key

# 微信支付配置
WECHAT_APP_ID=your_app_id
WECHAT_MCH_ID=your_mch_id
WECHAT_API_KEY=your_api_key
```

### 配置文件

支持JSON格式的配置文件：

```json
{
  "environment": "prod",
  "debug": false,
  "default_currency": "CNY",
  "alipay": {
    "enabled": true,
    "app_id": "your_app_id",
    "private_key": "your_private_key",
    "public_key": "alipay_public_key",
    "gateway_url": "https://openapi.alipay.com/gateway.do"
  },
  "wechat": {
    "enabled": true,
    "app_id": "your_app_id",
    "mch_id": "your_mch_id",
    "api_key": "your_api_key",
    "gateway_url": "https://api.mch.weixin.qq.com"
  }
}
```

## 部署指南

### 1. 数据库初始化

```sql
-- 创建支付相关表
CREATE TABLE payments (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  payment_no VARCHAR(64) UNIQUE NOT NULL,
  order_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  payment_method VARCHAR(32) NOT NULL,
  payment_status VARCHAR(32) NOT NULL,
  amount DECIMAL(10,2) NOT NULL,
  -- 其他字段...
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 创建支付配置表
CREATE TABLE payment_configs (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  payment_method VARCHAR(32) UNIQUE NOT NULL,
  is_enabled BOOLEAN DEFAULT TRUE,
  display_name VARCHAR(100) NOT NULL,
  min_amount DECIMAL(10,2) DEFAULT 0.01,
  max_amount DECIMAL(10,2) DEFAULT 50000.00,
  -- 其他字段...
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

### 2. 服务启动

```bash
# 编译项目
go build -o mall-go cmd/main.go

# 启动服务
./mall-go
```

### 3. 健康检查

```bash
# 检查服务状态
curl http://localhost:8080/health

# 检查支付方式
curl http://localhost:8080/api/v1/payments/methods
```

## 常见问题

### Q: 支付创建后如何获取支付链接？

A: 调用创建支付接口后，响应中的 `payment_data` 字段包含支付相关信息：
- 支付宝：`qr_code` 字段包含二维码内容
- 微信支付：`code_url` 字段包含支付链接

### Q: 如何处理支付超时？

A: 系统会自动处理支付超时，超时的支付订单状态会变为 `cancelled`。可以通过查询接口检查支付状态。

### Q: 退款多久到账？

A: 退款到账时间取决于第三方支付平台：
- 支付宝：1-3个工作日
- 微信支付：1-3个工作日

### Q: 如何测试支付功能？

A: 
1. 使用沙箱环境进行测试
2. 配置测试环境的支付参数
3. 运行集成测试：`go test ./test/integration/...`

## 联系支持

如有问题，请联系开发团队或查看项目文档。

---

**版本**: v1.0.0  
**更新时间**: 2023-12-01  
**维护团队**: Mall-Go 开发团队
