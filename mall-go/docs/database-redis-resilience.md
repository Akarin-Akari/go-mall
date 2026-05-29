# 数据库连接池与 Redis 容错指南

## 概述

本文档介绍 GoMall 项目中数据库连接池优化和 Redis 容错机制的实现。

## 数据库连接池优化

### 配置参数

在 `configs/config.yaml` 和 `configs/config-prod.yaml` 中配置：

```yaml
database:
  driver: mysql
  host: localhost
  port: 3306
  username: root
  password: "your-password"
  dbname: mall
  charset: utf8mb4
  max_idle_conns: 10      # 最大空闲连接数
  max_open_conns: 100     # 最大打开连接数
  conn_max_lifetime: 3600 # 连接最大生命周期（秒）
```

### 生产环境推荐配置

```yaml
database:
  max_idle_conns: 10      # 保持10个空闲连接，快速响应
  max_open_conns: 50      # 最多50个连接，避免数据库过载
  conn_max_lifetime: 3600 # 1小时后回收连接，避免长连接问题
```

### 连接池监控

系统会自动监控连接池状态（仅生产环境）：

- **监控间隔**: 60秒
- **监控指标**:
  - 最大连接数
  - 当前打开连接数
  - 正在使用的连接数
  - 空闲连接数
  - 等待连接的次数和时间
  - 因各种原因关闭的连接数

### 健康检查

连接池健康检查会自动检测以下问题：

1. **连接池饱和**: 当打开连接数 >= 90% 最大连接数时告警
2. **等待时间过长**: 平均等待时间 > 100ms 时告警
3. **无空闲连接**: 当有活跃连接但无空闲连接时告警

### 手动获取连接池状态

```go
import "mall-go/pkg/database"

// 获取连接池统计信息
stats, err := database.GetPoolStats(db)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("打开连接数: %d/%d\n", stats.OpenConnections, stats.MaxOpenConnections)
fmt.Printf("正在使用: %d\n", stats.InUse)
fmt.Printf("空闲连接: %d\n", stats.Idle)
```

## Redis 容错机制

### 核心特性

1. **自动降级**: Redis 连续失败超过阈值后自动禁用
2. **自动恢复**: 健康检查通过后自动重新启用
3. **优雅降级**: 降级时不影响业务逻辑，只是跳过缓存操作
4. **健康监控**: 定期检查 Redis 连接状态

### 配置参数

```yaml
redis:
  host: localhost
  port: 6379
  password: ""
  db: 0
  pool_size: 100          # 连接池大小
  min_idle_conns: 10      # 最小空闲连接数
  max_retries: 3          # 最大重试次数
  dial_timeout: 5         # 连接超时（秒）
  read_timeout: 3         # 读取超时（秒）
  write_timeout: 3        # 写入超时（秒）
  idle_timeout: 300       # 空闲连接超时（秒）
  max_conn_age: 3600      # 连接最大存活时间（秒）
```

### 生产环境推荐配置

```yaml
redis:
  pool_size: 100          # 支持高并发
  min_idle_conns: 10      # 快速响应
  max_retries: 5          # 提高容错性
  dial_timeout: 10        # 适应网络延迟
  read_timeout: 5         # 平衡性能和稳定性
  write_timeout: 5        # 平衡性能和稳定性
  idle_timeout: 600       # 10分钟空闲超时
  max_conn_age: 7200      # 2小时连接最大存活时间
```

### 使用容错客户端

```go
import "mall-go/pkg/cache"

// 获取容错客户端
client := cache.GetFallbackRedisClient()

// 检查是否可用
if client.IsEnabled() {
    // Redis 可用，执行缓存操作
    err := client.Set(ctx, "key", "value", time.Hour)
    if err != nil {
        // 处理错误，但不影响业务
        log.Warn("缓存写入失败", err)
    }
} else {
    // Redis 不可用，跳过缓存
    log.Info("Redis 已禁用，跳过缓存")
}

// 或者直接使用，内部会自动处理降级
err := client.Set(ctx, "key", "value", time.Hour)
// 降级时不会返回错误，避免影响业务
```

### 降级策略

#### 写操作降级

当 Redis 不可用时，写操作会被跳过，不返回错误：

```go
// 降级前
err := client.Set(ctx, "user:123", userData, time.Hour)
if err != nil {
    return err // 影响业务
}

// 降级后
err := client.Set(ctx, "user:123", userData, time.Hour)
// 降级时 err == nil，不影响业务
```

#### 读操作降级

当 Redis 不可用时，读操作返回错误，业务层需要从数据库读取：

```go
val, err := client.Get(ctx, "user:123")
if err != nil {
    // Redis 不可用或缓存未命中，从数据库读取
    user, err := db.GetUser(123)
    if err != nil {
        return err
    }
    
    // 尝试写入缓存（降级时会被跳过）
    client.Set(ctx, "user:123", user, time.Hour)
    
    return user
}
```

### 健康检查配置

容错客户端会自动进行健康检查：

- **检查间隔**: 30秒
- **失败阈值**: 5次连续失败
- **恢复阈值**: 3次连续成功

### 获取统计信息

```go
stats := client.GetStats()
fmt.Printf("Redis 状态: %+v\n", stats)

// 输出示例:
// {
//   "enabled": true,
//   "failure_count": 0,
//   "last_check": "2025-01-31T12:00:00Z",
//   "max_failures": 5,
//   "check_interval": "30s"
// }
```

## 健康检查端点

### 综合健康检查

```bash
curl http://localhost:8081/health
```

**响应示例**:

```json
{
  "status": "healthy",
  "timestamp": "2025-01-31T12:00:00Z",
  "services": {
    "database": {
      "status": "healthy",
      "pool": {
        "max_open": 100,
        "open": 15,
        "in_use": 5,
        "idle": 10,
        "wait_count": 0,
        "usage_percent": 15.0
      }
    },
    "redis": {
      "status": "healthy",
      "pool": {
        "total_conns": 20,
        "idle_conns": 15,
        "hits": 1000,
        "misses": 50,
        "timeouts": 0
      }
    }
  }
}
```

### Kubernetes 就绪检查

```bash
curl http://localhost:8081/readiness
```

### Kubernetes 存活检查

```bash
curl http://localhost:8081/liveness
```

## 最佳实践

### 1. 数据库连接池

- **开发环境**: `max_open_conns: 10`, `max_idle_conns: 5`
- **生产环境**: `max_open_conns: 50-100`, `max_idle_conns: 10-20`
- **高并发场景**: 根据实际负载调整，监控连接池使用率

### 2. Redis 容错

- **始终使用容错客户端**: 避免 Redis 故障影响业务
- **合理设置超时**: 避免长时间等待
- **监控降级状态**: 及时发现 Redis 问题

### 3. 缓存策略

- **Cache-Aside 模式**: 先查缓存，未命中再查数据库
- **写入策略**: 先更新数据库，再更新缓存
- **降级策略**: 缓存失败不影响业务，只影响性能

### 4. 监控告警

- **连接池使用率 > 80%**: 考虑扩容
- **Redis 降级**: 立即检查 Redis 服务
- **等待时间过长**: 优化查询或增加连接数

## 故障排查

### 问题1: 数据库连接池耗尽

**症状**: 日志显示 "connection pool is under pressure"

**解决方案**:
1. 检查是否有慢查询
2. 增加 `max_open_conns`
3. 优化数据库查询
4. 检查是否有连接泄漏

### 问题2: Redis 频繁降级

**症状**: 日志显示 "redis is temporarily disabled"

**解决方案**:
1. 检查 Redis 服务状态
2. 检查网络连接
3. 增加 `max_retries`
4. 调整超时配置

### 问题3: 健康检查失败

**症状**: `/health` 端点返回 503

**解决方案**:
1. 检查数据库连接
2. 检查 Redis 连接
3. 查看详细错误信息
4. 检查连接池配置

## 参考资料

- [GORM 连接池配置](https://gorm.io/docs/generic_interface.html)
- [go-redis 最佳实践](https://redis.uptrace.dev/guide/go-redis.html)
- [数据库连接池调优](https://www.alexedwards.net/blog/configuring-sqldb)

