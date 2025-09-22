# 后端Token管理API接口要求

## 概述

为了实现真正安全的Token存储策略，后端需要实现以下API接口来支持httpOnly Cookie的设置和管理。

## 必需的API接口

### 1. 设置刷新令牌Cookie

**接口**: `POST /api/auth/set-refresh-token`

**描述**: 设置httpOnly的刷新令牌Cookie

**请求体**:

```json
{
  "refreshToken": "string"
}
```

**响应**:

```json
{
  "success": true,
  "message": "Refresh token cookie set successfully"
}
```

**Cookie设置**:

```
Set-Cookie: refresh_token=<token>; HttpOnly; Secure; SameSite=Strict; Path=/; Max-Age=2592000
```

**安全要求**:

- 必须设置 `HttpOnly=true`
- 生产环境必须设置 `Secure=true`
- 设置 `SameSite=Strict`
- 设置适当的过期时间（建议30天）

### 2. 刷新访问令牌

**接口**: `POST /api/auth/refresh`

**描述**: 使用httpOnly Cookie中的刷新令牌获取新的访问令牌

**请求头**:

```
Cookie: refresh_token=<token>
Content-Type: application/json
```

**响应**:

```json
{
  "accessToken": "string",
  "expiresIn": 3600,
  "tokenType": "Bearer"
}
```

**错误响应**:

```json
{
  "error": "invalid_token",
  "message": "Refresh token is invalid or expired"
}
```

**安全要求**:

- 验证刷新令牌的有效性
- 检查令牌是否过期
- 可选：实现令牌轮换（返回新的刷新令牌）
- 记录刷新操作的日志

### 3. 登出接口

**接口**: `POST /api/auth/logout`

**描述**: 清除httpOnly Cookie并使令牌失效

**请求头**:

```
Cookie: refresh_token=<token>
```

**响应**:

```json
{
  "success": true,
  "message": "Logged out successfully"
}
```

**Cookie清除**:

```
Set-Cookie: refresh_token=; HttpOnly; Secure; SameSite=Strict; Path=/; Max-Age=0
```

**安全要求**:

- 将刷新令牌加入黑名单
- 清除相关的会话数据
- 记录登出操作的日志

### 4. 登录接口（更新）

**接口**: `POST /api/auth/login`

**描述**: 用户登录，返回访问令牌并设置刷新令牌Cookie

**请求体**:

```json
{
  "username": "string",
  "password": "string"
}
```

**响应**:

```json
{
  "accessToken": "string",
  "expiresIn": 3600,
  "tokenType": "Bearer",
  "user": {
    "id": "string",
    "username": "string",
    "email": "string"
  }
}
```

**Cookie设置**:

```
Set-Cookie: refresh_token=<refresh_token>; HttpOnly; Secure; SameSite=Strict; Path=/; Max-Age=2592000
```

## 安全最佳实践

### 1. Cookie安全配置

```go
// Go示例
http.SetCookie(w, &http.Cookie{
    Name:     "refresh_token",
    Value:    refreshToken,
    HttpOnly: true,
    Secure:   true, // 生产环境必须为true
    SameSite: http.SameSiteStrictMode,
    Path:     "/",
    MaxAge:   30 * 24 * 60 * 60, // 30天
})
```

### 2. 令牌验证

- 使用JWT并验证签名
- 检查令牌的过期时间
- 验证令牌的发行者和受众
- 实现令牌黑名单机制

### 3. 安全头设置

```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000; includeSubDomains
```

### 4. CORS配置

```go
// 允许凭据传递
w.Header().Set("Access-Control-Allow-Credentials", "true")
w.Header().Set("Access-Control-Allow-Origin", "https://yourdomain.com")
```

## 错误处理

### 常见错误码

- `400` - 请求参数错误
- `401` - 未授权（令牌无效或过期）
- `403` - 禁止访问
- `429` - 请求过于频繁
- `500` - 服务器内部错误

### 错误响应格式

```json
{
  "error": "error_code",
  "message": "Human readable error message",
  "details": {
    "field": "Additional error details"
  }
}
```

## 监控和日志

### 需要记录的事件

1. 令牌刷新操作
2. 登录/登出操作
3. 令牌验证失败
4. 异常的令牌使用模式

### 日志格式示例

```json
{
  "timestamp": "2025-01-12T10:30:00Z",
  "event": "token_refresh",
  "user_id": "user123",
  "ip_address": "192.168.1.1",
  "user_agent": "Mozilla/5.0...",
  "success": true
}
```

## 测试要求

### 单元测试

- 令牌生成和验证
- Cookie设置和读取
- 错误处理逻辑

### 集成测试

- 完整的登录/刷新/登出流程
- 跨域请求处理
- 安全头验证

### 安全测试

- XSS攻击防护
- CSRF攻击防护
- 令牌泄露场景测试

## 部署注意事项

1. **HTTPS要求**: 生产环境必须使用HTTPS
2. **域名配置**: 确保Cookie域名配置正确
3. **负载均衡**: 确保会话粘性或使用共享存储
4. **监控告警**: 设置异常登录和令牌使用的告警

## 迁移计划

1. **阶段1**: 实现新的API接口
2. **阶段2**: 前端集成新的Token管理器
3. **阶段3**: 逐步迁移现有用户
4. **阶段4**: 移除旧的不安全存储方式
