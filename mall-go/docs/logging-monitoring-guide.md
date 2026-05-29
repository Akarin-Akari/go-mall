# 日志与监控指南

## 概述

GoMall 项目使用 Zap 进行结构化日志记录，使用 Prometheus 进行指标监控。

## 日志系统

### 日志配置

在 `configs/config.yaml` 和 `configs/config-prod.yaml` 中配置：

```yaml
log:
  level: info              # 日志级别: debug, info, warn, error
  format: json             # 日志格式: json, console
  output: stdout           # 输出目标: stdout, stderr, file
  max_size: 100            # 单个日志文件最大大小(MB)
  max_age: 30              # 日志文件保留天数
  max_backups: 3           # 保留的旧日志文件数量
  compress: true           # 是否压缩旧日志文件
```

### 开发环境配置

```yaml
log:
  level: debug
  format: console          # 人类可读格式
  output: stdout
```

### 生产环境配置

```yaml
log:
  level: info
  format: json             # JSON格式，便于日志收集
  output: stdout           # 输出到标准输出，由容器收集
```

### 使用日志

```go
import "mall-go/pkg/logger"

// 基本日志
logger.Info("用户登录成功", zap.String("username", "admin"))
logger.Warn("Redis连接失败", zap.Error(err))
logger.Error("数据库查询失败", zap.Error(err), zap.String("query", sql))

// 带上下文的日志
logger.Debug("处理订单", 
    zap.Int64("order_id", 123),
    zap.String("status", "pending"),
)

// 性能日志
logger.LogPerformance(ctx, "create_order", duration, map[string]interface{}{
    "order_id": 123,
    "items": 5,
})
```

### 日志级别

- **Debug**: 详细的调试信息（仅开发环境）
- **Info**: 一般信息（默认级别）
- **Warn**: 警告信息（不影响功能但需要关注）
- **Error**: 错误信息（影响功能但不致命）
- **Fatal**: 致命错误（程序无法继续运行）

### JSON日志格式示例

```json
{
  "level": "info",
  "ts": "2025-01-31T12:00:00.000Z",
  "caller": "handler/user_handler.go:45",
  "msg": "用户登录成功",
  "username": "admin",
  "ip": "192.168.1.100"
}
```

## Prometheus 监控

### 指标端点

访问 `http://localhost:8081/metrics` 获取 Prometheus 格式的指标。

### 指标分类

#### 1. HTTP 请求指标

- `mall_go_http_requests_total`: HTTP请求总数
  - Labels: `method`, `endpoint`, `status`
- `mall_go_http_request_duration_seconds`: HTTP请求耗时
  - Labels: `method`, `endpoint`
- `mall_go_http_request_size_bytes`: HTTP请求大小
  - Labels: `method`, `endpoint`
- `mall_go_http_response_size_bytes`: HTTP响应大小
  - Labels: `method`, `endpoint`

#### 2. 业务指标

- `mall_go_business_user_registrations_total`: 用户注册总数
  - Labels: `source`
- `mall_go_business_user_logins_total`: 用户登录总数
  - Labels: `status`
- `mall_go_business_orders_created_total`: 订单创建总数
  - Labels: `type`
- `mall_go_business_orders_completed_total`: 订单完成总数
  - Labels: `type`
- `mall_go_business_payments_processed_total`: 支付处理总数
  - Labels: `method`, `status`
- `mall_go_business_product_views_total`: 商品浏览总数
  - Labels: `category`
- `mall_go_business_cart_operations_total`: 购物车操作总数
  - Labels: `operation`

#### 3. 数据库指标

- `mall_go_database_connections_open`: 打开的数据库连接数
- `mall_go_database_connections_in_use`: 正在使用的数据库连接数
- `mall_go_database_connections_idle`: 空闲的数据库连接数
- `mall_go_database_query_duration_seconds`: 数据库查询耗时
  - Labels: `operation`
- `mall_go_database_query_total`: 数据库查询总数
  - Labels: `operation`, `status`

#### 4. Redis 指标

- `mall_go_redis_connections_total`: Redis连接总数
- `mall_go_redis_connections_idle`: Redis空闲连接数
- `mall_go_redis_command_duration_seconds`: Redis命令耗时
  - Labels: `command`
- `mall_go_redis_command_total`: Redis命令总数
  - Labels: `command`, `status`
- `mall_go_redis_errors_total`: Redis错误总数
  - Labels: `type`

#### 5. 系统指标

- `mall_go_runtime_goroutines`: Goroutine数量
- `mall_go_runtime_memory_alloc_bytes`: 已分配的堆内存
- `mall_go_runtime_memory_sys_bytes`: 从OS获取的总内存
- `mall_go_runtime_gc_pause_total`: GC暂停总次数
- `mall_go_runtime_gc_pause_duration_seconds`: GC暂停时间

### 使用业务指标

```go
import "mall-go/pkg/metrics"

// 获取Prometheus指标
pm := metrics.GetPrometheusMetrics()

// 记录用户注册
pm.UserRegistrations.WithLabelValues("web").Inc()

// 记录用户登录
pm.UserLogins.WithLabelValues("success").Inc()

// 记录订单创建
pm.OrdersCreated.WithLabelValues("normal").Inc()

// 记录支付处理
pm.PaymentsProcessed.WithLabelValues("alipay", "success").Inc()

// 记录商品浏览
pm.ProductViews.WithLabelValues("electronics").Inc()

// 记录购物车操作
pm.CartOperations.WithLabelValues("add_item").Inc()
```

### Prometheus 配置

创建 `prometheus.yml` 配置文件：

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'mall-go'
    static_configs:
      - targets: ['localhost:8081']
    metrics_path: '/metrics'
```

### 启动 Prometheus

```bash
docker run -d \
  --name prometheus \
  -p 9090:9090 \
  -v $(pwd)/prometheus.yml:/etc/prometheus/prometheus.yml \
  prom/prometheus
```

访问 `http://localhost:9090` 查看 Prometheus UI。

### 常用 PromQL 查询

#### HTTP 请求速率

```promql
# 每秒HTTP请求数
rate(mall_go_http_requests_total[5m])

# 按状态码分组
sum(rate(mall_go_http_requests_total[5m])) by (status)

# 按端点分组
sum(rate(mall_go_http_requests_total[5m])) by (endpoint)
```

#### HTTP 请求延迟

```promql
# P50延迟
histogram_quantile(0.5, rate(mall_go_http_request_duration_seconds_bucket[5m]))

# P95延迟
histogram_quantile(0.95, rate(mall_go_http_request_duration_seconds_bucket[5m]))

# P99延迟
histogram_quantile(0.99, rate(mall_go_http_request_duration_seconds_bucket[5m]))
```

#### 业务指标

```promql
# 每分钟用户注册数
rate(mall_go_business_user_registrations_total[1m]) * 60

# 每分钟订单创建数
rate(mall_go_business_orders_created_total[1m]) * 60

# 支付成功率
sum(rate(mall_go_business_payments_processed_total{status="success"}[5m])) 
/ 
sum(rate(mall_go_business_payments_processed_total[5m]))
```

#### 数据库指标

```promql
# 数据库连接池使用率
mall_go_database_connections_in_use / mall_go_database_connections_open

# 数据库查询延迟P95
histogram_quantile(0.95, rate(mall_go_database_query_duration_seconds_bucket[5m]))
```

#### 系统指标

```promql
# Goroutine数量
mall_go_runtime_goroutines

# 内存使用
mall_go_runtime_memory_alloc_bytes / 1024 / 1024  # MB

# GC频率
rate(mall_go_runtime_gc_pause_total[5m])
```

## Grafana 可视化

### 安装 Grafana

```bash
docker run -d \
  --name grafana \
  -p 3000:3000 \
  grafana/grafana
```

访问 `http://localhost:3000`（默认用户名/密码: admin/admin）

### 添加 Prometheus 数据源

1. 进入 Configuration → Data Sources
2. 点击 "Add data source"
3. 选择 "Prometheus"
4. URL: `http://prometheus:9090`
5. 点击 "Save & Test"

### 导入仪表板

推荐使用以下 Grafana 仪表板模板：

- **Go Metrics**: Dashboard ID `10826`
- **Prometheus Stats**: Dashboard ID `2`
- **Node Exporter Full**: Dashboard ID `1860`

## 告警规则

创建 `alerts.yml` 文件：

```yaml
groups:
  - name: mall-go-alerts
    interval: 30s
    rules:
      # HTTP错误率告警
      - alert: HighHTTPErrorRate
        expr: |
          sum(rate(mall_go_http_requests_total{status=~"5.."}[5m])) 
          / 
          sum(rate(mall_go_http_requests_total[5m])) > 0.05
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "HTTP错误率过高"
          description: "HTTP 5xx错误率超过5%"

      # 数据库连接池告警
      - alert: DatabaseConnectionPoolNearFull
        expr: |
          mall_go_database_connections_in_use 
          / 
          mall_go_database_connections_open > 0.9
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "数据库连接池接近饱和"
          description: "数据库连接池使用率超过90%"

      # 内存使用告警
      - alert: HighMemoryUsage
        expr: mall_go_runtime_memory_alloc_bytes > 1073741824  # 1GB
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "内存使用过高"
          description: "内存使用超过1GB"
```

## 最佳实践

### 1. 日志记录

- **结构化日志**: 使用字段而不是字符串拼接
- **合适的级别**: Debug用于调试，Info用于正常流程，Warn用于异常但可恢复，Error用于错误
- **敏感信息**: 不要记录密码、密钥等敏感信息
- **上下文信息**: 记录足够的上下文（用户ID、订单ID等）

### 2. 指标监控

- **命名规范**: 使用 `namespace_subsystem_name_unit` 格式
- **标签使用**: 合理使用标签，避免高基数标签
- **采样频率**: 根据业务需求调整采样频率
- **告警设置**: 设置合理的告警阈值和持续时间

### 3. 性能优化

- **异步日志**: 使用缓冲区减少I/O开销
- **采样**: 对高频事件进行采样
- **批量处理**: 批量发送指标数据

## 参考资料

- [Zap 文档](https://pkg.go.dev/go.uber.org/zap)
- [Prometheus 文档](https://prometheus.io/docs/)
- [Grafana 文档](https://grafana.com/docs/)
- [PromQL 教程](https://prometheus.io/docs/prometheus/latest/querying/basics/)

