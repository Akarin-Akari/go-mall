# 安全加固快速参考卡片

## 🔐 核心安全组件

### JWT黑名单
```go
import "mall-go/pkg/security"

blacklist := security.GetJWTBlacklist()
blacklist.AddToBlacklist(ctx, token, 24*time.Hour)
if blacklist.IsBlacklisted(ctx, token) { /* 已失效 */ }
```

### 速率限制
```go
limiter := security.GetRateLimiter()
allowed, remaining, err := limiter.CheckLimit(ctx, clientIP)
if !allowed { /* 返回429 */ }
```

### 输入验证
```go
validator := security.GetInputValidator()
if !validator.ValidateSQL(input) { /* SQL注入 */ }
if !validator.ValidateXSS(input) { /* XSS */ }
valid, msg := validator.ValidatePassword(password)
```

---

## 🛡️ 安全中间件

```go
r.Use(middleware.SecurityHeadersMiddleware())      // 安全头
r.Use(middleware.CORSSecureMiddleware())           // CORS
r.Use(middleware.RateLimitMiddleware())            // 限流
r.Use(middleware.InputValidationMiddleware())      // 输入验证
r.Use(middleware.RequestSizeLimitMiddleware(10*1024*1024))
r.Use(middleware.JWTBlacklistMiddleware())         // JWT黑名单
```

---

## 📋 配置清单

### 生产环境必需

- [ ] JWT_SECRET ≥ 32字符
- [ ] CORS白名单已配置（禁止"*"）
- [ ] HTTPS已启用
- [ ] 数据库密码已设置
- [ ] Redis密码已设置
- [ ] 日志输出到stdout（JSON格式）

### 密码要求

- [ ] 长度: 8-128字符
- [ ] 包含大写字母、小写字母、数字、特殊字符中的至少3种
- [ ] 使用bcrypt加密（cost=12）

### CORS配置

```yaml
cors:
  allow_origins:
    - "https://www.example.com"
  allow_methods:
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
  allow_headers:
    - "Content-Type"
    - "Authorization"
```

---

## 🚀 启动命令

```bash
# 生产环境
ENV=prod \
JWT_SECRET="your-secret-key-at-least-32-chars" \
DB_PASSWORD="your-db-password" \
REDIS_PASSWORD="your-redis-password" \
go run cmd/server/main.go
```

---

## 🧪 测试命令

```bash
# 测试CORS
curl -X OPTIONS http://localhost:8081/api/v1/users \
  -H "Origin: https://www.example.com" \
  -H "Access-Control-Request-Method: POST" \
  -v

# 测试速率限制
for i in {1..150}; do curl http://localhost:8081/api/v1/users; done

# 测试SQL注入检测
curl "http://localhost:8081/api/v1/users?name='; DROP TABLE users; --"

# 测试XSS检测
curl "http://localhost:8081/api/v1/users?name=<script>alert('xss')</script>"

# 测试JWT黑名单
curl -X POST http://localhost:8081/api/v1/auth/logout \
  -H "Authorization: Bearer <token>"
```

---

## 📊 安全响应头

| 头名称 | 值 |
|--------|-----|
| X-Content-Type-Options | nosniff |
| X-Frame-Options | DENY |
| X-XSS-Protection | 1; mode=block |
| Content-Security-Policy | default-src 'self' |
| Strict-Transport-Security | max-age=31536000 |
| Referrer-Policy | strict-origin-when-cross-origin |
| Permissions-Policy | geolocation=(), microphone=(), camera=() |

---

## ⚠️ 常见错误

| 错误 | 原因 | 解决方案 |
|------|------|---------|
| 401 Unauthorized | JWT无效或过期 | 检查token、刷新token |
| 429 Too Many Requests | 请求过于频繁 | 等待或调整限流配置 |
| 400 Bad Request | 输入包含SQL/XSS | 清理输入 |
| CORS错误 | Origin不在白名单 | 添加到allow_origins |

---

## 🔍 监控指标

```bash
# 查看安全日志
tail -f logs/mall-go.log | grep -E "SQL注入|XSS|速率限制|黑名单"

# 检查Redis黑名单
redis-cli KEYS "jwt:blacklist:*"

# 检查限流计数
redis-cli KEYS "ratelimit:*"
```

---

## 📚 相关文档

- [详细安全指南](./security-hardening-guide.md)
- [配置管理指南](./config-management-guide.md)
- [日志监控指南](./logging-monitoring-guide.md)

---

**最后更新**: 2025-01-31

