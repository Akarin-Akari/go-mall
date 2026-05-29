# 多阶段构建 - 第一阶段：编译
FROM golang:1.25-alpine AS builder

# 安装必要的构建工具
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /build

# 复制go.mod和go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 编译应用
# 禁用CGO以支持alpine镜像
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o mall-server \
    ./cmd/server/main.go

# 多阶段构建 - 第二阶段：运行时
FROM alpine:latest

# 安装运行时依赖（包括SQLite支持）
RUN apk add --no-cache ca-certificates tzdata sqlite-libs

# 创建非root用户
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# 设置工作目录
WORKDIR /app

# 从builder阶段复制编译好的二进制文件
COPY --from=builder /build/mall-server .

# 复制配置文件
COPY --from=builder /build/configs ./configs

# 复制迁移文件
COPY --from=builder /build/migrations ./migrations

# 复制脚本文件
COPY --from=builder /build/scripts ./scripts

# 设置文件权限
RUN chown -R appuser:appuser /app

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8081

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

# 启动应用
CMD ["./mall-server"]

