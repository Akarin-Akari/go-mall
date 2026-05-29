# 安全加固指南

## 概述

GoMall 项目实现了完整的安全加固机制，包括：
- ✅ JWT认证增强（黑名单、刷新机制）
- ✅ 密码加密验证（bcrypt cost=12）
- ✅ CORS安全配置（白名单机制）
- ✅ 输入验证（SQL注入、XSS防护）
- ✅ 速率限制（基于IP和用户）
- ✅ 安全响应头

---

## 1. JWT认证增强

### JWT黑名单

用于实现Token登出功能：

```go
import "mall-go/pkg/security"

// 获取黑名单管理器
blacklist := security.GetJWTBlacklist()

// 将token添加到黑名单（登出时调用）
blacklist.AddToBlacklist(ctx, token, 24*time.Hour)

// 检查token是否在黑名单中
if blacklist.IsBlacklisted(ctx, token) {
    // Token已失效
}
```

### Token刷新机制

```go
import "mall-go/pkg/auth"

// 生成访问令牌（1小时有效期）
accessToken, _, err := auth.GenerateTokenWithExpiry(userID, username, role, time.Hour)

// 生成刷新令牌（7天有效期）
refreshToken, _, err := auth.GenerateRefreshToken(userID, username, role)

// 验证刷新令牌
claims, err := auth.ValidateRefreshToken(refreshToken)
```

### 配置

在 `configs/config-prod.yaml` 中配置：

```yaml
jwt:
  secret: ${JWT_SECRET}  # 必须通过环境变量设置
  expire: 24h            # Token过期时间
```

**安全要求**:
- ✅ JWT密钥长度 ≥ 32字符
- ✅ 生产环境必须通过环境变量设置
- ✅ 定期轮换密钥
- ✅ 使用HS256算法

---

## 2. 密码加密验证

### 密码加密

```go
import "mall-go/pkg/auth"

// 加密密码（bcrypt cost=12）
hashedPassword, err := auth.HashPassword(password)

// 验证密码
isValid := auth.VerifyPassword(password, hashedPassword)
```

### 密码强度验证

```go
import "mall-go/pkg/security"

validator := security.GetInputValidator()

// 验证密码强度
valid, msg := validator.ValidatePassword(password)
if !valid {
    // msg: "密码必须包含大写字母、小写字母、数字和特殊字符中的至少3种"
}
```

**密码要求**:
- ✅ 长度: 8-128字符
- ✅ 必须包含大写字母、小写字母、数字、特殊字符中的至少3种
- ✅ 使用bcrypt加密（cost=12）
- ✅ 不能使用常见密码

---

## 3. CORS安全配置

### 配置白名单

在 `configs/config-prod.yaml` 中配置：

```yaml
cors:
  allow_origins:
    - "https://www.example.com"
    - "https://admin.example.com"
  allow_methods:
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
  allow_headers:
    - "Content-Type"
    - "Authorization"
    - "X-Requested-With"
```

### 生产环境要求

- ✅ **禁止使用通配符 "*"**
- ✅ 明确指定允许的源
- ✅ 只允许必要的HTTP方法
- ✅ 只允许必要的请求头

### 验证CORS配置

```bash
# 测试CORS预检请求
curl -X OPTIONS http://localhost:8081/api/v1/users \
  -H "Origin: https://www.example.com" \
  -H "Access-Control-Request-Method: POST" \
  -v
```

---

## 4. 输入验证

### SQL注入防护

```go
import "mall-go/pkg/security"

validator := security.GetInputValidator()

// 验证SQL注入
if !validator.ValidateSQL(userInput) {
    // 检测到可能的SQL注入
}
```

### XSS防护

```go
// 验证XSS攻击
if !validator.ValidateXSS(userInput) {
    // 检测到可能的XSS攻击
}

// 清理输入
cleanInput := validator.SanitizeInput(userInput)
```

### 其他验证

```go
// 验证邮箱
if !validator.ValidateEmail(email) {
    // 邮箱格式无效
}

// 验证用户名
if !validator.ValidateUsername(username) {
    // 用户名格式无效（3-32字符，只能包含字母、数字、下划线、连字符）
}
```

---

## 5. 速率限制

### 基于IP的速率限制

```go
import "mall-go/pkg/security"

limiter := security.GetRateLimiter()

// 检查是否超过限制
allowed, remaining, err := limiter.CheckLimit(ctx, clientIP)
if !allowed {
    // 返回429 Too Many Requests
}

// 获取剩余请求数
remaining, err := limiter.GetRemaining(ctx, clientIP)
```

### 配置

在 `configs/config-prod.yaml` 中配置：

```yaml
rate_limit:
  requests_per_minute: 100
  burst_size: 10
  window_duration: 1m
```

### 不同端点的限流策略

```go
// 登录端点：更严格的限制
loginLimiter := security.NewRateLimiter(rdb, security.RateLimitConfig{
    RequestsPerMinute: 5,
    BurstSize:         1,
    WindowDuration:    time.Minute,
})

// API端点：标准限制
apiLimiter := security.NewRateLimiter(rdb, security.RateLimitConfig{
    RequestsPerMinute: 100,
    BurstSize:         10,
    WindowDuration:    time.Minute,
})
```

---

## 6. 安全响应头

### 自动添加的安全头

| 头名称 | 值 | 说明 |
|--------|-----|------|
| X-Content-Type-Options | nosniff | 防止MIME类型嗅探 |
| X-Frame-Options | DENY | 防止点击劫持 |
| X-XSS-Protection | 1; mode=block | 启用浏览器XSS防护 |
| Content-Security-Policy | default-src 'self' | 内容安全策略 |
| Strict-Transport-Security | max-age=31536000 | 强制HTTPS（生产环境） |
| Referrer-Policy | strict-origin-when-cross-origin | 控制Referrer信息 |
| Permissions-Policy | geolocation=(), microphone=(), camera=() | 控制浏览器功能 |

---

## 7. 中间件集成

### 在路由中使用安全中间件

```go
import "mall-go/internal/middleware"

// 注册安全中间件
r.Use(middleware.SecurityHeadersMiddleware())      // 安全响应头
r.Use(middleware.CORSSecureMiddleware())           // 安全CORS
r.Use(middleware.RateLimitMiddleware())            // 速率限制
r.Use(middleware.InputValidationMiddleware())      // 输入验证
r.Use(middleware.RequestSizeLimitMiddleware(10*1024*1024))  // 请求体大小限制
r.Use(middleware.JWTBlacklistMiddleware())         // JWT黑名单检查
```

---

## 8. 最佳实践

### 1. 环境变量管理

```bash
# 生产环境必须设置
export JWT_SECRET="your-secret-key-at-least-32-chars"
export DB_PASSWORD="your-db-password"
export REDIS_PASSWORD="your-redis-password"
```

### 2. HTTPS配置

```nginx
# Nginx配置示例
server {
    listen 443 ssl http2;
    server_name www.example.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    # 重定向HTTP到HTTPS
    if ($scheme != "https") {
        return 301 https://$server_name$request_uri;
    }
}
```

### 3. 日志监控

```bash
# 监控安全事件
tail -f logs/mall-go.log | grep -E "SQL注入|XSS|速率限制|黑名单"
```

### 4. 定期安全审计

- 检查JWT密钥是否定期轮换
- 验证CORS配置是否正确
- 检查密码加密是否使用bcrypt
- 监控速率限制是否有效

---

## 9. 故障排查

### 问题1: CORS请求被拒绝

**症状**: 浏览器报错 "Access to XMLHttpRequest blocked by CORS policy"

**解决方案**:
1. 检查 `configs/config-prod.yaml` 中的 `cors.allow_origins`
2. 确保请求的Origin在白名单中
3. 检查是否使用了通配符"*"（生产环境禁止）

### 问题2: JWT验证失败

**症状**: 返回401 Unauthorized

**解决方案**:
1. 检查JWT密钥是否正确设置
2. 检查token是否过期
3. 检查token是否在黑名单中
4. 检查Authorization头格式是否正确（"Bearer <token>"）

### 问题3: 请求被速率限制

**症状**: 返回429 Too Many Requests

**解决方案**:
1. 检查请求频率是否过高
2. 调整 `rate_limit.requests_per_minute` 配置
3. 检查是否有异常流量

### 问题4: 密码验证失败

**症状**: 密码强度验证不通过

**解决方案**:
1. 确保密码长度 ≥ 8字符
2. 确保密码包含大写字母、小写字母、数字、特殊字符中的至少3种
3. 检查密码是否包含非法字符

---

## 10. 参考资源

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)
- [CORS Specification](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS)
- [Content Security Policy](https://developer.mozilla.org/en-US/docs/Web/HTTP/CSP)
- [bcrypt Documentation](https://pkg.go.dev/golang.org/x/crypto/bcrypt)

---

**最后更新**: 2025-01-31

