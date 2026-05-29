# 优雅关闭 - 快速参考

## 🚀 快速开始

### 启动服务

```bash
# 开发环境
ENV=dev go run cmd/server/main.go

# 生产环境
ENV=prod JWT_SECRET=xxx DB_PASSWORD=xxx go run cmd/server/main.go

# 编译后运行
go build -o mall-server.exe cmd/server/main.go
./mall-server.exe
```

---

## 🛑 关闭服务

### 方式1: SIGTERM（推荐）

```bash
# 获取进程ID
ps aux | grep mall-server

# 发送SIGTERM信号
kill -TERM <PID>

# 或使用pkill
pkill -TERM -f mall-server
```

### 方式2: SIGINT（Ctrl+C）

```bash
# 在终端中按 Ctrl+C
```

### 方式3: Docker

```bash
docker stop <container_id>
docker-compose down
```

---

## 📊 关闭流程

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
完全关闭 ✅
```

---

## ⏱️ 超时配置

| 参数 | 值 | 说明 |
|------|-----|------|
| ReadTimeout | 15秒 | 读取请求超时 |
| WriteTimeout | 15秒 | 写入响应超时 |
| IdleTimeout | 60秒 | 空闲连接超时 |
| 总关闭超时 | 30秒 | 整个关闭流程超时 |

---

## 📝 日志示例

```
INFO  收到关闭信号 signal=terminated
INFO  开始执行优雅关闭流程 timeout=30s
INFO  关闭HTTP服务器...
INFO  HTTP服务器已关闭
INFO  停止系统指标收集器...
INFO  系统指标收集器已停止
INFO  关闭Redis连接...
INFO  Redis连接已关闭
INFO  关闭数据库连接...
INFO  数据库连接已关闭
INFO  优雅关闭流程完成
INFO  服务器已完全关闭
```

---

## 🔍 监控关闭

```bash
# 实时查看关闭日志
tail -f logs/mall-go.log | grep -E "关闭|shutdown"

# 查看最后一次关闭
grep "优雅关闭流程完成" logs/mall-go.log | tail -1

# 统计关闭次数
grep "收到关闭信号" logs/mall-go.log | wc -l
```

---

## ✅ 关闭性能

| 场景 | 耗时 |
|------|------|
| 无请求 | < 100ms |
| 少量请求 | 100-500ms |
| 中等请求 | 500ms-2s |
| 大量请求 | 2-30s |

---

## 🐳 Docker 最佳实践

### Dockerfile

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o mall-server cmd/server/main.go

FROM alpine:latest
COPY --from=builder /app/mall-server /app/
ENTRYPOINT ["/app/mall-server"]
```

### docker-compose.yml

```yaml
version: '3.8'
services:
  mall-go:
    build: .
    ports:
      - "8081:8081"
    environment:
      ENV: prod
      JWT_SECRET: ${JWT_SECRET}
      DB_PASSWORD: ${DB_PASSWORD}
    stop_grace_period: 40s  # 大于30秒关闭超时
```

---

## ☸️ Kubernetes 最佳实践

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: mall-go
spec:
  terminationGracePeriodSeconds: 40  # 大于30秒
  containers:
  - name: mall-go
    image: mall-go:latest
    ports:
    - containerPort: 8081
    lifecycle:
      preStop:
        exec:
          command: ["/bin/sh", "-c", "sleep 5"]
```

---

## 🧪 测试优雅关闭

```bash
#!/bin/bash

# 启动服务
go run cmd/server/main.go &
SERVER_PID=$!

# 等待启动
sleep 2

# 发送请求
curl http://localhost:8081/health &
REQUEST_PID=$!

# 等待请求开始
sleep 1

# 发送关闭信号
kill -TERM $SERVER_PID

# 等待完成
wait $SERVER_PID
SHUTDOWN_STATUS=$?

wait $REQUEST_PID
REQUEST_STATUS=$?

# 检查结果
if [ $SHUTDOWN_STATUS -eq 0 ] && [ $REQUEST_STATUS -eq 0 ]; then
    echo "✅ 优雅关闭测试通过"
else
    echo "❌ 优雅关闭测试失败"
fi
```

---

## 🔧 故障排查

### 问题: 服务无法关闭

**检查**:
1. 查看日志中是否有错误
2. 检查是否有长时间运行的请求
3. 增加关闭超时时间

### 问题: 关闭时间过长

**检查**:
1. 是否有大量正在处理的请求
2. 数据库连接是否关闭缓慢
3. 是否有慢查询

### 问题: 数据丢失

**检查**:
1. 确保所有数据库操作都在事务中
2. 在关闭前刷新缓存
3. 增加关闭超时时间

---

## 📚 相关文件

- `pkg/shutdown/graceful.go` - 优雅关闭实现
- `cmd/server/main.go` - 服务启动和集成
- `docs/graceful-shutdown-guide.md` - 详细文档
- `docs/task5-graceful-shutdown-summary.md` - 完成总结

---

## 🔗 参考资源

- [Go http.Server.Shutdown()](https://pkg.go.dev/net/http#Server.Shutdown)
- [Go signal 包](https://pkg.go.dev/os/signal)
- [Kubernetes 优雅关闭](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#termination-of-pods)
- [Docker 信号处理](https://docs.docker.com/engine/reference/run/#signal-handling)

---

**最后更新**: 2025-01-31

