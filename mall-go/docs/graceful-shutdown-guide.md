# 优雅关闭指南

## 概述

GoMall 项目实现了完整的优雅关闭（Graceful Shutdown）机制，确保在服务器关闭时：
- 停止接收新请求
- 等待正在处理的请求完成
- 正确清理所有资源
- 记录详细的关闭日志

## 工作原理

### 关闭流程

```
收到信号 (SIGTERM/SIGINT)
    ↓
停止接收新请求
    ↓
等待现有请求完成（最多15秒）
    ↓
停止系统指标收集器
    ↓
关闭Redis连接
    ↓
关闭数据库连接
    ↓
完全关闭
```

### 关键特性

1. **信号捕获**: 监听 SIGTERM 和 SIGINT 信号
2. **HTTP优雅关闭**: 使用 `http.Server.Shutdown()` 等待现有请求完成
3. **资源清理**: 按顺序关闭所有资源
4. **超时保护**: 30秒总超时，防止无限等待
5. **详细日志**: 记录每个关闭步骤

## 配置参数

### HTTP服务器超时

在 `cmd/server/main.go` 中配置：

```go
server := &http.Server{
    Addr:         ":8081",
    Handler:      r,
    ReadTimeout:  15 * time.Second,   // 读取请求超时
    WriteTimeout: 15 * time.Second,   // 写入响应超时
    IdleTimeout:  60 * time.Second,   // 空闲连接超时
}
```

### 关闭超时

```go
gracefulShutdown := shutdown.NewGracefulShutdown(
    server, 
    db, 
    rdb, 
    30*time.Second,  // 总关闭超时
)
```

## 使用方法

### 启动服务

```bash
# 开发环境
ENV=dev go run cmd/server/main.go

# 生产环境
ENV=prod JWT_SECRET=xxx DB_PASSWORD=xxx go run cmd/server/main.go
```

### 优雅关闭

#### 方式1: 发送SIGTERM信号（推荐用于容器）

```bash
# 获取进程ID
ps aux | grep mall-server

# 发送SIGTERM信号
kill -TERM <PID>

# 或使用pkill
pkill -TERM -f mall-server
```

#### 方式2: 发送SIGINT信号（Ctrl+C）

```bash
# 在终端中按 Ctrl+C
# 或使用kill命令
kill -INT <PID>
```

#### 方式3: Docker容器

```bash
# Docker会自动发送SIGTERM信号
docker stop <container_id>

# 或使用docker-compose
docker-compose down
```

### 关闭日志示例

```json
{
  "level": "info",
  "ts": "2025-01-31T12:00:00.000Z",
  "msg": "收到关闭信号",
  "signal": "terminated"
}
{
  "level": "info",
  "ts": "2025-01-31T12:00:00.001Z",
  "msg": "开始执行优雅关闭流程",
  "timeout": "30s"
}
{
  "level": "info",
  "ts": "2025-01-31T12:00:00.002Z",
  "msg": "关闭HTTP服务器..."
}
{
  "level": "info",
  "ts": "2025-01-31T12:00:00.500Z",
  "msg": "HTTP服务器已关闭"
}
{
  "level": "info",
  "ts": "2025-01-31T12:00:00.501Z",
  "msg": "停止系统指标收集器..."
}
{
  "level": "info",
  "ts": "2025-01-31T12:00:00.502Z",
  "msg": "系统指标收集器已停止"
}
{
  "level": "info",
  "ts": "2025-01-31T12:00:00.503Z",
  "msg": "关闭Redis连接..."
}
{
  "level": "info",
  "ts": "2025-01-31T12:00:00.504Z",
  "msg": "Redis连接已关闭"
}
{
  "level": "info",
  "ts": "2025-01-31T12:00:00.505Z",
  "msg": "关闭数据库连接..."
}
{
  "level": "info",
  "ts": "2025-01-31T12:00:00.506Z",
  "msg": "数据库连接已关闭"
}
{
  "level": "info",
  "ts": "2025-01-31T12:00:00.507Z",
  "msg": "优雅关闭流程完成"
}
{
  "level": "info",
  "ts": "2025-01-31T12:00:00.508Z",
  "msg": "服务器已完全关闭"
}
```

## 最佳实践

### 1. 容器环境

在 Dockerfile 中使用 `exec` 形式的 ENTRYPOINT：

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o mall-server cmd/server/main.go

FROM alpine:latest
COPY --from=builder /app/mall-server /app/
ENTRYPOINT ["/app/mall-server"]
```

这样可以确保 Go 进程接收到信号。

### 2. Kubernetes 部署

配置 Pod 的 terminationGracePeriodSeconds：

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: mall-go
spec:
  terminationGracePeriodSeconds: 40  # 大于30秒关闭超时
  containers:
  - name: mall-go
    image: mall-go:latest
    ports:
    - containerPort: 8081
    lifecycle:
      preStop:
        exec:
          command: ["/bin/sh", "-c", "sleep 5"]  # 给负载均衡器时间移除该Pod
```

### 3. 监控关闭过程

使用日志监控关闭过程：

```bash
# 实时查看日志
tail -f logs/mall-go.log | grep -E "关闭|shutdown"

# 统计关闭耗时
grep "优雅关闭流程完成" logs/mall-go.log | tail -1
```

### 4. 测试优雅关闭

创建测试脚本 `test-graceful-shutdown.sh`：

```bash
#!/bin/bash

# 启动服务
go run cmd/server/main.go &
SERVER_PID=$!

# 等待服务启动
sleep 2

# 发送请求
curl http://localhost:8081/health &
REQUEST_PID=$!

# 等待请求开始处理
sleep 1

# 发送关闭信号
kill -TERM $SERVER_PID

# 等待服务关闭
wait $SERVER_PID
SHUTDOWN_STATUS=$?

# 检查请求是否完成
wait $REQUEST_PID
REQUEST_STATUS=$?

if [ $SHUTDOWN_STATUS -eq 0 ] && [ $REQUEST_STATUS -eq 0 ]; then
    echo "✅ 优雅关闭测试通过"
else
    echo "❌ 优雅关闭测试失败"
fi
```

运行测试：

```bash
chmod +x test-graceful-shutdown.sh
./test-graceful-shutdown.sh
```

## 故障排查

### 问题1: 服务无法关闭

**症状**: 发送信号后服务仍在运行

**原因**: 
- 某个请求处理时间过长
- 数据库连接未正确关闭

**解决方案**:
1. 检查日志中是否有错误
2. 增加关闭超时时间
3. 检查是否有长时间运行的后台任务

### 问题2: 关闭时间过长

**症状**: 关闭耗时超过30秒

**原因**:
- 有大量正在处理的请求
- 数据库连接关闭缓慢

**解决方案**:
1. 检查是否有慢查询
2. 优化数据库连接池配置
3. 增加关闭超时时间

### 问题3: 数据丢失

**症状**: 关闭时有未保存的数据

**原因**:
- 请求处理中的数据未提交
- 缓存未同步到数据库

**解决方案**:
1. 确保所有数据库操作都在事务中
2. 在关闭前刷新缓存
3. 增加关闭超时时间

## 性能指标

### 典型关闭时间

| 场景 | 耗时 |
|------|------|
| 无请求处理 | < 100ms |
| 少量请求（<10） | 100-500ms |
| 中等请求（10-100） | 500ms-2s |
| 大量请求（>100） | 2-30s |

### 资源清理顺序

1. HTTP服务器关闭: ~100ms
2. 指标收集器停止: ~10ms
3. Redis连接关闭: ~50ms
4. 数据库连接关闭: ~100ms

**总计**: 通常 < 500ms（无长时间请求）

## 参考资料

- [Go http.Server.Shutdown()](https://pkg.go.dev/net/http#Server.Shutdown)
- [Go signal 包](https://pkg.go.dev/os/signal)
- [Kubernetes 优雅关闭](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#termination-of-pods)
- [Docker 信号处理](https://docs.docker.com/engine/reference/run/#signal-handling)

