# 高级篇第二章：Docker容器化部署 🐳

> *"容器化是现代应用部署的标准实践，它不仅解决了'在我机器上能跑'的问题，更是云原生架构的基石。掌握Docker和Kubernetes，就是掌握了现代软件交付的核心技能！"* 🚀

## 📚 本章学习目标

通过本章学习，你将掌握：

- 🐳 **Docker基础与Go应用容器化**：理解容器技术原理，掌握Go应用的容器化最佳实践
- 🏗️ **多阶段构建优化**：掌握Docker多阶段构建技术，大幅减小镜像体积并提高安全性
- ☸️ **Kubernetes部署策略**：深入学习K8s核心概念，掌握Deployment、Service、Ingress等资源
- 📦 **Helm图表管理**：学习Helm包管理器，实现应用的模板化部署和版本管理
- 🔒 **容器安全配置**：掌握容器安全最佳实践，包括非root用户、资源限制、安全扫描
- 🔄 **CI/CD流水线集成**：学习GitHub Actions、GitLab CI等CI/CD工具与容器化的集成
- 🏢 **企业级容器化方案**：结合mall-go项目，设计完整的企业级容器化部署方案

---

## 🐳 Docker基础与Go应用容器化

### Docker核心概念

Docker是一个开源的容器化平台，它使用Linux内核的特性（如cgroups和namespaces）来创建轻量级、可移植的容器。

```go
// Docker核心概念
package docker

import (
    "context"
    "time"
)

// 容器生命周期管理
type ContainerLifecycle struct {
    // 镜像构建阶段
    Build struct {
        Dockerfile   string `json:"dockerfile"`   // Dockerfile路径
        Context      string `json:"context"`      // 构建上下文
        BuildArgs    map[string]string `json:"build_args"` // 构建参数
        Tags         []string `json:"tags"`       // 镜像标签
        Platform     string `json:"platform"`    // 目标平台
    } `json:"build"`
    
    // 容器运行阶段
    Runtime struct {
        Image        string `json:"image"`        // 镜像名称
        Command      []string `json:"command"`    // 启动命令
        Environment  map[string]string `json:"environment"` // 环境变量
        Ports        []PortMapping `json:"ports"`  // 端口映射
        Volumes      []VolumeMount `json:"volumes"` // 卷挂载
        Networks     []string `json:"networks"`   // 网络配置
        Resources    ResourceLimits `json:"resources"` // 资源限制
    } `json:"runtime"`
    
    // 容器监控阶段
    Monitoring struct {
        HealthCheck  HealthCheckConfig `json:"health_check"` // 健康检查
        Logging      LoggingConfig `json:"logging"`      // 日志配置
        Metrics      MetricsConfig `json:"metrics"`      // 指标收集
    } `json:"monitoring"`
}

// 端口映射
type PortMapping struct {
    HostPort      int    `json:"host_port"`
    ContainerPort int    `json:"container_port"`
    Protocol      string `json:"protocol"` // tcp/udp
}

// 卷挂载
type VolumeMount struct {
    Source      string `json:"source"`      // 主机路径或卷名
    Target      string `json:"target"`      // 容器内路径
    Type        string `json:"type"`        // bind/volume/tmpfs
    ReadOnly    bool   `json:"read_only"`   // 只读挂载
}

// 资源限制
type ResourceLimits struct {
    CPULimit    string `json:"cpu_limit"`    // CPU限制 (如: "0.5")
    MemoryLimit string `json:"memory_limit"` // 内存限制 (如: "512m")
    CPURequest  string `json:"cpu_request"`  // CPU请求
    MemoryRequest string `json:"memory_request"` // 内存请求
}

// 健康检查配置
type HealthCheckConfig struct {
    Test        []string      `json:"test"`         // 检查命令
    Interval    time.Duration `json:"interval"`     // 检查间隔
    Timeout     time.Duration `json:"timeout"`      // 超时时间
    Retries     int           `json:"retries"`      // 重试次数
    StartPeriod time.Duration `json:"start_period"` // 启动等待期
}
```

### Go应用容器化最佳实践

```dockerfile
# Go应用基础Dockerfile
FROM golang:1.21-alpine AS base

# 设置工作目录
WORKDIR /app

# 安装必要的系统依赖
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 复制go.mod和go.sum文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o main ./cmd/server

# 运行阶段
FROM alpine:3.18

# 安装运行时依赖
RUN apk add --no-cache \
    ca-certificates \
    tzdata

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 复制二进制文件
COPY --from=base /app/main .

# 复制配置文件
COPY --from=base /app/configs ./configs

# 设置文件权限
RUN chown -R appuser:appgroup /app

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 启动应用
CMD ["./main"]
```

### Docker Compose配置

```yaml
# docker-compose.yml - Mall-Go完整环境
version: '3.8'

services:
  # Go后端服务
  mall-go:
    build:
      context: ./mall-go
      dockerfile: Dockerfile
      target: production
    container_name: mall-go-backend
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      - ENV=production
      - DB_HOST=mysql
      - DB_PORT=3306
      - DB_NAME=mall_go
      - DB_USER=root
      - DB_PASSWORD=123456
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - JWT_SECRET=your-jwt-secret
    volumes:
      - ./logs:/app/logs
      - ./uploads:/app/uploads
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - mall-network
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  # MySQL数据库
  mysql:
    image: mysql:8.0
    container_name: mall-mysql
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: 123456
      MYSQL_DATABASE: mall_go
      MYSQL_USER: mall_user
      MYSQL_PASSWORD: mall_password
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
      - ./mall-go/init_database.sql:/docker-entrypoint-initdb.d/init.sql
    command: --default-authentication-plugin=mysql_native_password
    networks:
      - mall-network
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Redis缓存
  redis:
    image: redis:7-alpine
    container_name: mall-redis
    restart: unless-stopped
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
      - ./redis.conf:/usr/local/etc/redis/redis.conf
    command: redis-server /usr/local/etc/redis/redis.conf
    networks:
      - mall-network
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 5

  # Nginx反向代理
  nginx:
    image: nginx:alpine
    container_name: mall-nginx
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf
      - ./nginx/conf.d:/etc/nginx/conf.d
      - ./ssl:/etc/nginx/ssl
      - ./logs/nginx:/var/log/nginx
    depends_on:
      - mall-go
    networks:
      - mall-network

  # 前端服务
  mall-frontend:
    build:
      context: ./mall-frontend
      dockerfile: Dockerfile
      target: production
    container_name: mall-frontend
    restart: unless-stopped
    ports:
      - "3000:3000"
    environment:
      - NODE_ENV=production
      - NEXT_PUBLIC_API_URL=http://mall-go:8080
    depends_on:
      - mall-go
    networks:
      - mall-network

volumes:
  mysql_data:
    driver: local
  redis_data:
    driver: local

networks:
  mall-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16
```

---

## 🏗️ 多阶段构建优化

### 多阶段构建原理

多阶段构建是Docker的一个强大特性，允许在单个Dockerfile中使用多个FROM语句，每个FROM指令开始一个新的构建阶段。这种技术的主要优势：

1. **减小镜像体积**：只保留运行时必需的文件
2. **提高安全性**：排除构建工具和源代码
3. **简化构建流程**：在一个文件中管理整个构建过程
4. **并行构建**：不同阶段可以并行执行

```go
// 多阶段构建策略
package build

import (
    "fmt"
    "time"
)

// 构建阶段定义
type BuildStage struct {
    Name        string            `json:"name"`         // 阶段名称
    BaseImage   string            `json:"base_image"`   // 基础镜像
    Purpose     string            `json:"purpose"`      // 阶段目的
    Operations  []BuildOperation  `json:"operations"`   // 构建操作
    Artifacts   []Artifact        `json:"artifacts"`    // 产出物
    Size        string            `json:"size"`         // 阶段大小
    Duration    time.Duration     `json:"duration"`     // 构建时长
}

// 构建操作
type BuildOperation struct {
    Type        string            `json:"type"`         // 操作类型
    Command     string            `json:"command"`      // 执行命令
    Description string            `json:"description"`  // 操作描述
    Cache       bool              `json:"cache"`        // 是否缓存
}

// 构建产出物
type Artifact struct {
    Source      string `json:"source"`      // 源路径
    Target      string `json:"target"`      // 目标路径
    Type        string `json:"type"`        // 文件类型
    Size        string `json:"size"`        // 文件大小
    Required    bool   `json:"required"`    // 是否必需
}

// Mall-Go多阶段构建策略
func GetMallGoMultiStageBuild() []BuildStage {
    return []BuildStage{
        {
            Name:      "dependencies",
            BaseImage: "golang:1.21-alpine",
            Purpose:   "下载和缓存Go模块依赖",
            Operations: []BuildOperation{
                {Type: "COPY", Command: "COPY go.mod go.sum ./", Description: "复制依赖文件"},
                {Type: "RUN", Command: "go mod download", Description: "下载依赖", Cache: true},
            },
            Artifacts: []Artifact{
                {Source: "/go/pkg/mod", Target: "go_modules", Type: "directory", Required: true},
            },
        },
        {
            Name:      "builder",
            BaseImage: "golang:1.21-alpine",
            Purpose:   "编译Go应用程序",
            Operations: []BuildOperation{
                {Type: "COPY", Command: "COPY . .", Description: "复制源代码"},
                {Type: "RUN", Command: "go build -ldflags='-w -s' -o main ./cmd/server", Description: "编译应用"},
            },
            Artifacts: []Artifact{
                {Source: "/app/main", Target: "binary", Type: "executable", Size: "~15MB", Required: true},
            },
        },
        {
            Name:      "runtime",
            BaseImage: "alpine:3.18",
            Purpose:   "创建最小运行时环境",
            Operations: []BuildOperation{
                {Type: "RUN", Command: "apk add --no-cache ca-certificates tzdata", Description: "安装运行时依赖"},
                {Type: "COPY", Command: "COPY --from=builder /app/main .", Description: "复制二进制文件"},
            },
            Artifacts: []Artifact{
                {Source: "final_image", Target: "production", Type: "image", Size: "~20MB", Required: true},
            },
        },
    }
}
```

### 优化的多阶段Dockerfile

```dockerfile
# syntax=docker/dockerfile:1
# Mall-Go优化的多阶段构建

# ================================
# 阶段1: 依赖缓存层
# ================================
FROM golang:1.21-alpine AS dependencies

# 设置工作目录
WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache \
    git \
    ca-certificates \
    gcc \
    musl-dev

# 设置Go环境变量
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY=https://goproxy.cn,direct

# 复制依赖文件（利用Docker层缓存）
COPY go.mod go.sum ./

# 下载依赖（这一层会被缓存，除非go.mod/go.sum改变）
RUN go mod download && \
    go mod verify

# ================================
# 阶段2: 构建层
# ================================
FROM dependencies AS builder

# 复制源代码
COPY . .

# 构建应用（使用编译优化参数）
RUN go build \
    -ldflags="-w -s -X main.version=$(git describe --tags --always --dirty) -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -a -installsuffix cgo \
    -o main ./cmd/server

# 验证二进制文件
RUN ./main --version

# ================================
# 阶段3: 测试层（可选）
# ================================
FROM builder AS tester

# 运行测试
RUN go test -v ./...

# 运行静态分析
RUN go vet ./...

# ================================
# 阶段4: 安全扫描层（可选）
# ================================
FROM builder AS security-scan

# 安装安全扫描工具
RUN go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# 运行安全扫描
RUN gosec ./...

# ================================
# 阶段5: 生产运行时
# ================================
FROM gcr.io/distroless/static-debian11:nonroot AS production

# 设置标签
LABEL maintainer="mall-go-team@example.com" \
      version="1.0.0" \
      description="Mall-Go E-commerce Backend Service" \
      org.opencontainers.image.source="https://github.com/your-org/mall-go"

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/main ./main

# 复制配置文件
COPY --from=builder /app/configs ./configs

# 复制静态资源（如果有）
COPY --from=builder /app/static ./static

# 设置非root用户（distroless已经包含）
USER nonroot:nonroot

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app/main", "healthcheck"]

# 启动应用
ENTRYPOINT ["/app/main"]
CMD ["server"]

# ================================
# 阶段6: 开发环境（可选）
# ================================
FROM golang:1.21-alpine AS development

WORKDIR /app

# 安装开发工具
RUN apk add --no-cache \
    git \
    make \
    curl \
    vim

# 安装Go开发工具
RUN go install github.com/cosmtrek/air@latest && \
    go install github.com/go-delve/delve/cmd/dlv@latest

# 复制源代码
COPY . .

# 下载依赖
RUN go mod download

# 暴露调试端口
EXPOSE 8080 2345

# 使用air进行热重载
CMD ["air"]
```

### 构建脚本优化

```bash
#!/bin/bash
# build.sh - Mall-Go优化构建脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置变量
IMAGE_NAME="mall-go"
REGISTRY="your-registry.com"
VERSION=${VERSION:-$(git describe --tags --always --dirty)}
BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT=$(git rev-parse --short HEAD)

# 函数：打印带颜色的消息
print_message() {
    echo -e "${2}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

# 函数：构建镜像
build_image() {
    local target=$1
    local tag_suffix=$2

    print_message "构建 $target 镜像..." $BLUE

    docker build \
        --target $target \
        --build-arg VERSION=$VERSION \
        --build-arg BUILD_DATE=$BUILD_DATE \
        --build-arg GIT_COMMIT=$GIT_COMMIT \
        --tag $IMAGE_NAME:$VERSION$tag_suffix \
        --tag $IMAGE_NAME:latest$tag_suffix \
        --cache-from $IMAGE_NAME:latest$tag_suffix \
        .

    print_message "$target 镜像构建完成" $GREEN
}

# 函数：运行测试
run_tests() {
    print_message "运行测试..." $BLUE

    docker build \
        --target tester \
        --tag $IMAGE_NAME:test \
        .

    print_message "测试完成" $GREEN
}

# 函数：安全扫描
security_scan() {
    print_message "运行安全扫描..." $BLUE

    docker build \
        --target security-scan \
        --tag $IMAGE_NAME:security \
        .

    print_message "安全扫描完成" $GREEN
}

# 函数：镜像大小分析
analyze_image_size() {
    print_message "分析镜像大小..." $BLUE

    echo "镜像大小对比："
    docker images $IMAGE_NAME --format "table {{.Repository}}:{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}"

    # 使用dive分析镜像层（如果安装了dive）
    if command -v dive &> /dev/null; then
        print_message "使用dive分析镜像层..." $YELLOW
        dive $IMAGE_NAME:$VERSION
    fi
}

# 函数：推送镜像
push_image() {
    local target=$1
    local tag_suffix=$2

    print_message "推送 $target 镜像到注册表..." $BLUE

    # 标记镜像
    docker tag $IMAGE_NAME:$VERSION$tag_suffix $REGISTRY/$IMAGE_NAME:$VERSION$tag_suffix
    docker tag $IMAGE_NAME:latest$tag_suffix $REGISTRY/$IMAGE_NAME:latest$tag_suffix

    # 推送镜像
    docker push $REGISTRY/$IMAGE_NAME:$VERSION$tag_suffix
    docker push $REGISTRY/$IMAGE_NAME:latest$tag_suffix

    print_message "$target 镜像推送完成" $GREEN
}

# 主函数
main() {
    print_message "开始Mall-Go多阶段构建..." $GREEN

    case "${1:-all}" in
        "dev")
            build_image "development" "-dev"
            ;;
        "test")
            run_tests
            ;;
        "security")
            security_scan
            ;;
        "prod")
            build_image "production" ""
            analyze_image_size
            ;;
        "push")
            build_image "production" ""
            push_image "production" ""
            ;;
        "all")
            run_tests
            security_scan
            build_image "production" ""
            build_image "development" "-dev"
            analyze_image_size
            ;;
        *)
            echo "用法: $0 {dev|test|security|prod|push|all}"
            exit 1
            ;;
    esac

    print_message "构建完成！" $GREEN
}

# 执行主函数
main "$@"
```

### 镜像优化对比

| 优化策略 | 镜像大小 | 构建时间 | 安全性 | 维护性 |
|----------|----------|----------|--------|--------|
| **单阶段构建** | ~1.2GB | 快 | 低 | 简单 |
| **基础多阶段** | ~50MB | 中等 | 中等 | 中等 |
| **优化多阶段** | ~20MB | 慢 | 高 | 复杂 |
| **Distroless** | ~15MB | 慢 | 很高 | 复杂 |

```go
// 镜像优化效果对比
type ImageOptimizationComparison struct {
    Strategy    string  `json:"strategy"`     // 优化策略
    Size        string  `json:"size"`         // 镜像大小
    BuildTime   string  `json:"build_time"`   // 构建时间
    Security    string  `json:"security"`     // 安全等级
    Layers      int     `json:"layers"`       // 镜像层数
    Vulnerabilities int `json:"vulnerabilities"` // 漏洞数量
}

func GetOptimizationComparison() []ImageOptimizationComparison {
    return []ImageOptimizationComparison{
        {
            Strategy: "单阶段构建",
            Size: "1.2GB",
            BuildTime: "2分钟",
            Security: "低",
            Layers: 15,
            Vulnerabilities: 50,
        },
        {
            Strategy: "基础多阶段",
            Size: "50MB",
            BuildTime: "3分钟",
            Security: "中等",
            Layers: 8,
            Vulnerabilities: 20,
        },
        {
            Strategy: "优化多阶段",
            Size: "20MB",
            BuildTime: "4分钟",
            Security: "高",
            Layers: 5,
            Vulnerabilities: 5,
        },
        {
            Strategy: "Distroless",
            Size: "15MB",
            BuildTime: "5分钟",
            Security: "很高",
            Layers: 3,
            Vulnerabilities: 0,
        },
    }
}
```

---

## ☸️ Kubernetes部署策略

### Kubernetes核心概念

Kubernetes是一个开源的容器编排平台，用于自动化容器化应用的部署、扩展和管理。

```go
// Kubernetes资源定义
package k8s

import (
    "time"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    networkingv1 "k8s.io/api/networking/v1"
)

// Mall-Go Kubernetes部署配置
type MallGoK8sConfig struct {
    Namespace   string                 `json:"namespace"`
    Deployment  *appsv1.Deployment     `json:"deployment"`
    Service     *corev1.Service        `json:"service"`
    Ingress     *networkingv1.Ingress  `json:"ingress"`
    ConfigMap   *corev1.ConfigMap      `json:"config_map"`
    Secret      *corev1.Secret         `json:"secret"`
    HPA         *HorizontalPodAutoscaler `json:"hpa"`
}

// 水平Pod自动扩缩器
type HorizontalPodAutoscaler struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`
    Spec              HPASpec   `json:"spec"`
    Status            HPAStatus `json:"status,omitempty"`
}

type HPASpec struct {
    ScaleTargetRef CrossVersionObjectReference `json:"scaleTargetRef"`
    MinReplicas    *int32                      `json:"minReplicas,omitempty"`
    MaxReplicas    int32                       `json:"maxReplicas"`
    Metrics        []MetricSpec                `json:"metrics,omitempty"`
}

type MetricSpec struct {
    Type     string            `json:"type"`
    Resource *ResourceMetricSource `json:"resource,omitempty"`
}

type ResourceMetricSource struct {
    Name   string             `json:"name"`
    Target MetricTarget       `json:"target"`
}

type MetricTarget struct {
    Type               string `json:"type"`
    AverageUtilization *int32 `json:"averageUtilization,omitempty"`
}
```

### Kubernetes YAML配置

#### 1. Namespace配置

```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: mall-go
  labels:
    name: mall-go
    environment: production
    team: backend
---
# ResourceQuota - 资源配额
apiVersion: v1
kind: ResourceQuota
metadata:
  name: mall-go-quota
  namespace: mall-go
spec:
  hard:
    requests.cpu: "4"
    requests.memory: 8Gi
    limits.cpu: "8"
    limits.memory: 16Gi
    persistentvolumeclaims: "10"
    services: "10"
    secrets: "10"
    configmaps: "10"
---
# NetworkPolicy - 网络策略
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: mall-go-network-policy
  namespace: mall-go
spec:
  podSelector:
    matchLabels:
      app: mall-go
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    - podSelector:
        matchLabels:
          app: mall-frontend
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: mysql
    ports:
    - protocol: TCP
      port: 3306
  - to:
    - podSelector:
        matchLabels:
          app: redis
    ports:
    - protocol: TCP
      port: 6379
```

#### 2. ConfigMap配置

```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: mall-go-config
  namespace: mall-go
  labels:
    app: mall-go
    component: config
data:
  app.yaml: |
    server:
      port: 8080
      mode: release
      read_timeout: 60s
      write_timeout: 60s

    database:
      driver: mysql
      host: mysql-service
      port: 3306
      name: mall_go
      charset: utf8mb4
      parse_time: true
      loc: Asia/Shanghai
      max_idle_conns: 10
      max_open_conns: 100
      conn_max_lifetime: 3600s

    redis:
      host: redis-service
      port: 6379
      password: ""
      db: 0
      pool_size: 10
      min_idle_conns: 5

    jwt:
      secret: ${JWT_SECRET}
      expire: 7200s

    log:
      level: info
      format: json
      output: stdout

    cors:
      allow_origins:
        - "http://localhost:3000"
        - "https://mall.example.com"
      allow_methods:
        - GET
        - POST
        - PUT
        - DELETE
        - OPTIONS
      allow_headers:
        - Origin
        - Content-Type
        - Authorization

  nginx.conf: |
    upstream mall-go-backend {
        server mall-go-service:8080;
    }

    server {
        listen 80;
        server_name mall.example.com;

        location /api/ {
            proxy_pass http://mall-go-backend/;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;

            # 超时设置
            proxy_connect_timeout 60s;
            proxy_send_timeout 60s;
            proxy_read_timeout 60s;

            # 缓冲设置
            proxy_buffering on;
            proxy_buffer_size 4k;
            proxy_buffers 8 4k;
        }

        location /health {
            access_log off;
            return 200 "healthy\n";
            add_header Content-Type text/plain;
        }
    }
```

#### 3. Secret配置

```yaml
# secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: mall-go-secret
  namespace: mall-go
  labels:
    app: mall-go
    component: secret
type: Opaque
data:
  # Base64编码的敏感信息
  jwt-secret: eW91ci1qd3Qtc2VjcmV0LWtleQ==  # your-jwt-secret-key
  db-password: MTIzNDU2  # 123456
  redis-password: ""

---
# TLS Secret for HTTPS
apiVersion: v1
kind: Secret
metadata:
  name: mall-go-tls
  namespace: mall-go
type: kubernetes.io/tls
data:
  tls.crt: LS0tLS1CRUdJTi... # Base64编码的证书
  tls.key: LS0tLS1CRUdJTi... # Base64编码的私钥
```

#### 4. Deployment配置

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mall-go
  namespace: mall-go
  labels:
    app: mall-go
    version: v1.0.0
    component: backend
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
  selector:
    matchLabels:
      app: mall-go
  template:
    metadata:
      labels:
        app: mall-go
        version: v1.0.0
        component: backend
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
    spec:
      # 安全上下文
      securityContext:
        runAsNonRoot: true
        runAsUser: 65534
        fsGroup: 65534

      # 服务账户
      serviceAccountName: mall-go-sa

      # 初始化容器
      initContainers:
      - name: wait-for-mysql
        image: busybox:1.35
        command: ['sh', '-c']
        args:
        - |
          until nc -z mysql-service 3306; do
            echo "等待MySQL启动..."
            sleep 2
          done
          echo "MySQL已就绪"

      - name: wait-for-redis
        image: busybox:1.35
        command: ['sh', '-c']
        args:
        - |
          until nc -z redis-service 6379; do
            echo "等待Redis启动..."
            sleep 2
          done
          echo "Redis已就绪"

      containers:
      - name: mall-go
        image: your-registry.com/mall-go:v1.0.0
        imagePullPolicy: IfNotPresent

        ports:
        - name: http
          containerPort: 8080
          protocol: TCP

        # 环境变量
        env:
        - name: ENV
          value: "production"
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: mall-go-secret
              key: jwt-secret
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: mall-go-secret
              key: db-password

        # 资源限制
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi

        # 健康检查
        livenessProbe:
          httpGet:
            path: /health
            port: http
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3

        readinessProbe:
          httpGet:
            path: /ready
            port: http
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3

        # 启动探针
        startupProbe:
          httpGet:
            path: /health
            port: http
          initialDelaySeconds: 10
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 30

        # 卷挂载
        volumeMounts:
        - name: config
          mountPath: /app/configs
          readOnly: true
        - name: logs
          mountPath: /app/logs
        - name: uploads
          mountPath: /app/uploads

        # 安全上下文
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          capabilities:
            drop:
            - ALL

      # 卷定义
      volumes:
      - name: config
        configMap:
          name: mall-go-config
      - name: logs
        emptyDir: {}
      - name: uploads
        persistentVolumeClaim:
          claimName: mall-go-uploads-pvc

      # 节点选择器
      nodeSelector:
        kubernetes.io/os: linux

      # 容忍度
      tolerations:
      - key: "node.kubernetes.io/not-ready"
        operator: "Exists"
        effect: "NoExecute"
        tolerationSeconds: 300
      - key: "node.kubernetes.io/unreachable"
        operator: "Exists"
        effect: "NoExecute"
        tolerationSeconds: 300

      # Pod反亲和性
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: app
                  operator: In
                  values:
                  - mall-go
              topologyKey: kubernetes.io/hostname
```

#### 5. Service配置

```yaml
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: mall-go-service
  namespace: mall-go
  labels:
    app: mall-go
    component: backend
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-type: nlb
    prometheus.io/scrape: "true"
    prometheus.io/port: "8080"
spec:
  type: ClusterIP
  ports:
  - name: http
    port: 8080
    targetPort: http
    protocol: TCP
  selector:
    app: mall-go

---
# Headless Service for StatefulSet (如果需要)
apiVersion: v1
kind: Service
metadata:
  name: mall-go-headless
  namespace: mall-go
  labels:
    app: mall-go
    component: backend
spec:
  clusterIP: None
  ports:
  - name: http
    port: 8080
    targetPort: http
  selector:
    app: mall-go
```

#### 6. Ingress配置

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: mall-go-ingress
  namespace: mall-go
  labels:
    app: mall-go
  annotations:
    # Nginx Ingress Controller配置
    kubernetes.io/ingress.class: nginx
    nginx.ingress.kubernetes.io/rewrite-target: /$2
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"

    # 速率限制
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/rate-limit-window: "1m"

    # 超时设置
    nginx.ingress.kubernetes.io/proxy-connect-timeout: "60"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "60"
    nginx.ingress.kubernetes.io/proxy-read-timeout: "60"

    # 缓冲设置
    nginx.ingress.kubernetes.io/proxy-buffering: "on"
    nginx.ingress.kubernetes.io/proxy-buffer-size: "4k"

    # CORS设置
    nginx.ingress.kubernetes.io/enable-cors: "true"
    nginx.ingress.kubernetes.io/cors-allow-origin: "https://mall.example.com"
    nginx.ingress.kubernetes.io/cors-allow-methods: "GET, POST, PUT, DELETE, OPTIONS"
    nginx.ingress.kubernetes.io/cors-allow-headers: "DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Authorization"

    # 证书管理器
    cert-manager.io/cluster-issuer: "letsencrypt-prod"

spec:
  tls:
  - hosts:
    - api.mall.example.com
    secretName: mall-go-tls

  rules:
  - host: api.mall.example.com
    http:
      paths:
      - path: /api(/|$)(.*)
        pathType: Prefix
        backend:
          service:
            name: mall-go-service
            port:
              number: 8080
      - path: /health
        pathType: Exact
        backend:
          service:
            name: mall-go-service
            port:
              number: 8080
```

#### 7. HPA配置

```yaml
# hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: mall-go-hpa
  namespace: mall-go
  labels:
    app: mall-go
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: mall-go

  minReplicas: 2
  maxReplicas: 10

  metrics:
  # CPU使用率
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70

  # 内存使用率
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80

  # 自定义指标：请求QPS
  - type: Pods
    pods:
      metric:
        name: http_requests_per_second
      target:
        type: AverageValue
        averageValue: "100"

  # 扩缩容行为
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 10
        periodSeconds: 60
      - type: Pods
        value: 2
        periodSeconds: 60
      selectPolicy: Min

    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
      - type: Pods
        value: 4
        periodSeconds: 60
      selectPolicy: Max
```

#### 8. PersistentVolumeClaim配置

```yaml
# pvc.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: mall-go-uploads-pvc
  namespace: mall-go
  labels:
    app: mall-go
    component: storage
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 10Gi
  storageClassName: fast-ssd

---
# 日志存储PVC
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: mall-go-logs-pvc
  namespace: mall-go
  labels:
    app: mall-go
    component: logs
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 5Gi
  storageClassName: standard
```

#### 9. ServiceAccount和RBAC配置

```yaml
# rbac.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: mall-go-sa
  namespace: mall-go
  labels:
    app: mall-go

---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: mall-go
  name: mall-go-role
rules:
- apiGroups: [""]
  resources: ["configmaps", "secrets"]
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list", "watch"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: mall-go-rolebinding
  namespace: mall-go
subjects:
- kind: ServiceAccount
  name: mall-go-sa
  namespace: mall-go
roleRef:
  kind: Role
  name: mall-go-role
  apiGroup: rbac.authorization.k8s.io
```

---

## 📦 Helm图表管理

### Helm基础概念

Helm是Kubernetes的包管理器，它使用称为"Charts"的包格式来定义、安装和升级复杂的Kubernetes应用。

```go
// Helm Chart结构定义
package helm

import (
    "time"
)

// Helm Chart配置
type HelmChart struct {
    Metadata    ChartMetadata         `json:"metadata"`
    Values      map[string]interface{} `json:"values"`
    Templates   []Template            `json:"templates"`
    Dependencies []Dependency         `json:"dependencies"`
    Hooks       []Hook               `json:"hooks"`
}

// Chart元数据
type ChartMetadata struct {
    Name        string   `json:"name"`
    Version     string   `json:"version"`
    AppVersion  string   `json:"app_version"`
    Description string   `json:"description"`
    Keywords    []string `json:"keywords"`
    Maintainers []Maintainer `json:"maintainers"`
    Sources     []string `json:"sources"`
    Dependencies []Dependency `json:"dependencies"`
}

// 维护者信息
type Maintainer struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    URL   string `json:"url"`
}

// Chart依赖
type Dependency struct {
    Name       string `json:"name"`
    Version    string `json:"version"`
    Repository string `json:"repository"`
    Condition  string `json:"condition,omitempty"`
    Tags       []string `json:"tags,omitempty"`
    Enabled    bool   `json:"enabled"`
}

// 模板文件
type Template struct {
    Name     string `json:"name"`
    Path     string `json:"path"`
    Content  string `json:"content"`
    Type     string `json:"type"` // deployment, service, ingress, etc.
}

// Helm Hook
type Hook struct {
    Name     string   `json:"name"`
    Weight   int      `json:"weight"`
    Events   []string `json:"events"` // pre-install, post-install, etc.
    Template Template `json:"template"`
}

// Release信息
type Release struct {
    Name      string    `json:"name"`
    Namespace string    `json:"namespace"`
    Version   int       `json:"version"`
    Status    string    `json:"status"`
    Chart     string    `json:"chart"`
    Updated   time.Time `json:"updated"`
    Values    map[string]interface{} `json:"values"`
}
```

### Mall-Go Helm Chart结构

```
mall-go-chart/
├── Chart.yaml              # Chart元数据
├── values.yaml             # 默认配置值
├── values-dev.yaml         # 开发环境配置
├── values-staging.yaml     # 测试环境配置
├── values-prod.yaml        # 生产环境配置
├── requirements.yaml       # Chart依赖
├── .helmignore            # 忽略文件
├── README.md              # 使用说明
├── templates/             # 模板文件目录
│   ├── _helpers.tpl       # 模板助手函数
│   ├── deployment.yaml    # Deployment模板
│   ├── service.yaml       # Service模板
│   ├── ingress.yaml       # Ingress模板
│   ├── configmap.yaml     # ConfigMap模板
│   ├── secret.yaml        # Secret模板
│   ├── hpa.yaml          # HPA模板
│   ├── pvc.yaml          # PVC模板
│   ├── rbac.yaml         # RBAC模板
│   ├── serviceaccount.yaml # ServiceAccount模板
│   ├── networkpolicy.yaml # NetworkPolicy模板
│   └── tests/            # 测试模板
│       └── test-connection.yaml
├── charts/               # 子Chart目录
│   ├── mysql/           # MySQL子Chart
│   └── redis/           # Redis子Chart
└── crds/                # 自定义资源定义
    └── mall-crd.yaml
```

### Chart.yaml配置

```yaml
# Chart.yaml
apiVersion: v2
name: mall-go
description: A Helm chart for Mall-Go E-commerce Backend Service
type: application
version: 1.0.0
appVersion: "1.0.0"

keywords:
  - ecommerce
  - golang
  - microservice
  - mall

home: https://github.com/your-org/mall-go
sources:
  - https://github.com/your-org/mall-go

maintainers:
  - name: Mall-Go Team
    email: team@mall-go.com
    url: https://mall-go.com

dependencies:
  - name: mysql
    version: 9.4.6
    repository: https://charts.bitnami.com/bitnami
    condition: mysql.enabled
    tags:
      - database

  - name: redis
    version: 17.3.7
    repository: https://charts.bitnami.com/bitnami
    condition: redis.enabled
    tags:
      - cache

  - name: nginx-ingress
    version: 4.4.0
    repository: https://kubernetes.github.io/ingress-nginx
    condition: ingress.enabled
    tags:
      - ingress

annotations:
  category: E-commerce
  licenses: MIT
```

### values.yaml配置

```yaml
# values.yaml - 默认配置值
# 全局配置
global:
  imageRegistry: ""
  imagePullSecrets: []
  storageClass: ""

# 应用配置
app:
  name: mall-go
  version: "1.0.0"

# 镜像配置
image:
  registry: your-registry.com
  repository: mall-go
  tag: "v1.0.0"
  pullPolicy: IfNotPresent
  pullSecrets: []

# 副本配置
replicaCount: 3

# 服务配置
service:
  type: ClusterIP
  port: 8080
  targetPort: http
  annotations: {}

# Ingress配置
ingress:
  enabled: true
  className: nginx
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /$2
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
  hosts:
    - host: api.mall.example.com
      paths:
        - path: /api(/|$)(.*)
          pathType: Prefix
  tls:
    - secretName: mall-go-tls
      hosts:
        - api.mall.example.com

# 资源限制
resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 100m
    memory: 128Mi

# 自动扩缩容
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80

# 健康检查
healthcheck:
  enabled: true
  livenessProbe:
    httpGet:
      path: /health
      port: http
    initialDelaySeconds: 30
    periodSeconds: 10
    timeoutSeconds: 5
    failureThreshold: 3
  readinessProbe:
    httpGet:
      path: /ready
      port: http
    initialDelaySeconds: 5
    periodSeconds: 5
    timeoutSeconds: 3
    failureThreshold: 3

# 配置文件
config:
  server:
    port: 8080
    mode: release
    read_timeout: 60s
    write_timeout: 60s
  database:
    driver: mysql
    host: "{{ .Release.Name }}-mysql"
    port: 3306
    name: mall_go
    charset: utf8mb4
  redis:
    host: "{{ .Release.Name }}-redis-master"
    port: 6379
    db: 0
  jwt:
    expire: 7200s
  log:
    level: info
    format: json

# 密钥配置
secrets:
  jwtSecret: "your-jwt-secret-key"
  dbPassword: "123456"
  redisPassword: ""

# 存储配置
persistence:
  enabled: true
  storageClass: ""
  accessMode: ReadWriteMany
  size: 10Gi
  annotations: {}

# 安全配置
security:
  runAsNonRoot: true
  runAsUser: 65534
  fsGroup: 65534
  readOnlyRootFilesystem: true

# 网络策略
networkPolicy:
  enabled: true
  ingress:
    - from:
      - namespaceSelector:
          matchLabels:
            name: ingress-nginx
      ports:
      - protocol: TCP
        port: 8080

# 服务账户
serviceAccount:
  create: true
  annotations: {}
  name: ""

# Pod安全策略
podSecurityPolicy:
  enabled: false

# 节点选择器
nodeSelector: {}

# 容忍度
tolerations: []

# 亲和性
affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 100
      podAffinityTerm:
        labelSelector:
          matchExpressions:
          - key: app.kubernetes.io/name
            operator: In
            values:
            - mall-go
        topologyKey: kubernetes.io/hostname

# MySQL依赖配置
mysql:
  enabled: true
  auth:
    rootPassword: "123456"
    database: "mall_go"
    username: "mall_user"
    password: "mall_password"
  primary:
    persistence:
      enabled: true
      size: 20Gi

# Redis依赖配置
redis:
  enabled: true
  auth:
    enabled: false
  master:
    persistence:
      enabled: true
      size: 5Gi

# 监控配置
monitoring:
  enabled: true
  serviceMonitor:
    enabled: true
    interval: 30s
    path: /metrics
    labels: {}

# 日志配置
logging:
  enabled: true
  level: info
  format: json
```

### Helm模板示例

#### _helpers.tpl模板助手

```yaml
{{/*
Expand the name of the chart.
*/}}
{{- define "mall-go.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "mall-go.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "mall-go.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "mall-go.labels" -}}
helm.sh/chart: {{ include "mall-go.chart" . }}
{{ include "mall-go.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "mall-go.selectorLabels" -}}
app.kubernetes.io/name: {{ include "mall-go.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "mall-go.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "mall-go.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Generate certificates secret name
*/}}
{{- define "mall-go.tlsSecretName" -}}
{{- printf "%s-tls" (include "mall-go.fullname" .) }}
{{- end }}

{{/*
Generate database connection string
*/}}
{{- define "mall-go.databaseURL" -}}
{{- if .Values.mysql.enabled }}
{{- printf "%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local" .Values.config.database.username .Values.secrets.dbPassword .Values.config.database.host (.Values.config.database.port | int) .Values.config.database.name .Values.config.database.charset }}
{{- else }}
{{- printf "%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local" .Values.config.database.username .Values.secrets.dbPassword .Values.config.database.host (.Values.config.database.port | int) .Values.config.database.name .Values.config.database.charset }}
{{- end }}
{{- end }}
```

### Helm部署命令

```bash
# 创建Chart
helm create mall-go-chart

# 验证Chart语法
helm lint mall-go-chart/

# 渲染模板（调试用）
helm template mall-go mall-go-chart/ --values mall-go-chart/values-dev.yaml

# 安装Chart
helm install mall-go mall-go-chart/ \
  --namespace mall-go \
  --create-namespace \
  --values mall-go-chart/values-prod.yaml

# 升级Release
helm upgrade mall-go mall-go-chart/ \
  --namespace mall-go \
  --values mall-go-chart/values-prod.yaml \
  --atomic \
  --timeout 10m

# 回滚Release
helm rollback mall-go 1 --namespace mall-go

# 查看Release状态
helm status mall-go --namespace mall-go

# 查看Release历史
helm history mall-go --namespace mall-go

# 卸载Release
helm uninstall mall-go --namespace mall-go

# 打包Chart
helm package mall-go-chart/

# 推送到Chart仓库
helm push mall-go-1.0.0.tgz oci://your-registry.com/helm-charts
```

---

## 🔒 容器安全配置

### 容器安全最佳实践

容器安全是企业级部署的重要考虑因素，需要从多个层面进行防护。

```go
// 容器安全配置
package security

import (
    "time"
)

// 安全配置策略
type SecurityConfig struct {
    // 镜像安全
    ImageSecurity ImageSecurityConfig `json:"image_security"`

    // 运行时安全
    RuntimeSecurity RuntimeSecurityConfig `json:"runtime_security"`

    // 网络安全
    NetworkSecurity NetworkSecurityConfig `json:"network_security"`

    // 数据安全
    DataSecurity DataSecurityConfig `json:"data_security"`

    // 访问控制
    AccessControl AccessControlConfig `json:"access_control"`

    // 审计日志
    AuditLogging AuditLoggingConfig `json:"audit_logging"`
}

// 镜像安全配置
type ImageSecurityConfig struct {
    // 镜像扫描
    VulnerabilityScanning struct {
        Enabled     bool     `json:"enabled"`
        Scanners    []string `json:"scanners"`    // trivy, clair, snyk
        Threshold   string   `json:"threshold"`   // low, medium, high, critical
        BlockDeploy bool     `json:"block_deploy"` // 阻止有漏洞的镜像部署
    } `json:"vulnerability_scanning"`

    // 镜像签名
    ImageSigning struct {
        Enabled    bool   `json:"enabled"`
        Provider   string `json:"provider"`   // cosign, notary
        PublicKey  string `json:"public_key"`
        Required   bool   `json:"required"`
    } `json:"image_signing"`

    // 可信镜像仓库
    TrustedRegistries []string `json:"trusted_registries"`

    // 基础镜像策略
    BaseImagePolicy struct {
        AllowedImages []string `json:"allowed_images"`
        BlockedImages []string `json:"blocked_images"`
        RequireMinimal bool    `json:"require_minimal"` // 要求使用最小化镜像
    } `json:"base_image_policy"`
}

// 运行时安全配置
type RuntimeSecurityConfig struct {
    // 用户权限
    UserSecurity struct {
        RunAsNonRoot           bool  `json:"run_as_non_root"`
        RunAsUser              int64 `json:"run_as_user"`
        RunAsGroup             int64 `json:"run_as_group"`
        FSGroup                int64 `json:"fs_group"`
        AllowPrivilegeEscalation bool `json:"allow_privilege_escalation"`
    } `json:"user_security"`

    // 文件系统安全
    FilesystemSecurity struct {
        ReadOnlyRootFilesystem bool     `json:"read_only_root_filesystem"`
        AllowedVolumes         []string `json:"allowed_volumes"`
        ForbiddenPaths         []string `json:"forbidden_paths"`
    } `json:"filesystem_security"`

    // 能力控制
    Capabilities struct {
        Drop []string `json:"drop"` // 删除的能力
        Add  []string `json:"add"`  // 添加的能力
    } `json:"capabilities"`

    // 资源限制
    ResourceLimits struct {
        CPU    string `json:"cpu"`
        Memory string `json:"memory"`
        PID    int    `json:"pid"`
    } `json:"resource_limits"`

    // 安全上下文
    SecurityContext struct {
        SELinuxOptions struct {
            Level string `json:"level"`
            Role  string `json:"role"`
            Type  string `json:"type"`
            User  string `json:"user"`
        } `json:"selinux_options"`

        AppArmorProfile string `json:"apparmor_profile"`
        SeccompProfile  string `json:"seccomp_profile"`
    } `json:"security_context"`
}

// 网络安全配置
type NetworkSecurityConfig struct {
    // 网络策略
    NetworkPolicies []NetworkPolicy `json:"network_policies"`

    // TLS配置
    TLSConfig struct {
        Enabled     bool   `json:"enabled"`
        MinVersion  string `json:"min_version"`  // TLS1.2, TLS1.3
        CipherSuites []string `json:"cipher_suites"`
        CertManager struct {
            Enabled bool   `json:"enabled"`
            Issuer  string `json:"issuer"`
        } `json:"cert_manager"`
    } `json:"tls_config"`

    // 服务网格
    ServiceMesh struct {
        Enabled  bool   `json:"enabled"`
        Provider string `json:"provider"` // istio, linkerd
        mTLS     bool   `json:"mtls"`
    } `json:"service_mesh"`
}

// 网络策略定义
type NetworkPolicy struct {
    Name      string              `json:"name"`
    Namespace string              `json:"namespace"`
    Selector  map[string]string   `json:"selector"`
    Ingress   []NetworkPolicyRule `json:"ingress"`
    Egress    []NetworkPolicyRule `json:"egress"`
}

type NetworkPolicyRule struct {
    From  []NetworkPolicyPeer `json:"from"`
    To    []NetworkPolicyPeer `json:"to"`
    Ports []NetworkPolicyPort `json:"ports"`
}

type NetworkPolicyPeer struct {
    PodSelector       map[string]string `json:"pod_selector"`
    NamespaceSelector map[string]string `json:"namespace_selector"`
    IPBlock           struct {
        CIDR   string   `json:"cidr"`
        Except []string `json:"except"`
    } `json:"ip_block"`
}

type NetworkPolicyPort struct {
    Protocol string `json:"protocol"`
    Port     int    `json:"port"`
}
```

### 安全扫描配置

```yaml
# security-scan.yaml - 安全扫描配置
apiVersion: v1
kind: ConfigMap
metadata:
  name: security-scan-config
  namespace: mall-go
data:
  trivy-config.yaml: |
    # Trivy配置
    vulnerability:
      type:
        - os
        - library
      severity:
        - UNKNOWN
        - LOW
        - MEDIUM
        - HIGH
        - CRITICAL
    secret:
      config: trivy-secret.yaml
    format: json
    output: /tmp/trivy-report.json

  falco-rules.yaml: |
    # Falco运行时安全规则
    - rule: Detect shell in container
      desc: Detect shell execution in container
      condition: >
        spawned_process and container and
        (proc.name in (shell_binaries) or
         proc.name in (bash, sh, zsh, fish))
      output: >
        Shell spawned in container (user=%user.name container=%container.name
        image=%container.image.repository:%container.image.tag shell=%proc.name)
      priority: WARNING

    - rule: Detect privilege escalation
      desc: Detect privilege escalation attempts
      condition: >
        spawned_process and container and
        proc.name in (sudo, su, doas)
      output: >
        Privilege escalation attempt (user=%user.name container=%container.name
        image=%container.image.repository:%container.image.tag command=%proc.cmdline)
      priority: HIGH

  opa-policies.rego: |
    # Open Policy Agent安全策略
    package kubernetes.admission

    deny[msg] {
        input.request.kind.kind == "Pod"
        input.request.object.spec.securityContext.runAsUser == 0
        msg := "Container must not run as root user"
    }

    deny[msg] {
        input.request.kind.kind == "Pod"
        input.request.object.spec.containers[_].securityContext.privileged == true
        msg := "Privileged containers are not allowed"
    }

    deny[msg] {
        input.request.kind.kind == "Pod"
        input.request.object.spec.containers[_].securityContext.allowPrivilegeEscalation == true
        msg := "Privilege escalation is not allowed"
    }
```

### Pod安全标准配置

```yaml
# pod-security-policy.yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: mall-go-psp
  namespace: mall-go
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
    - ALL
  volumes:
    - 'configMap'
    - 'emptyDir'
    - 'projected'
    - 'secret'
    - 'downwardAPI'
    - 'persistentVolumeClaim'
  runAsUser:
    rule: 'MustRunAsNonRoot'
  seLinux:
    rule: 'RunAsAny'
  fsGroup:
    rule: 'RunAsAny'
  readOnlyRootFilesystem: true

---
# 安全上下文约束 (OpenShift)
apiVersion: security.openshift.io/v1
kind: SecurityContextConstraints
metadata:
  name: mall-go-scc
allowHostDirVolumePlugin: false
allowHostIPC: false
allowHostNetwork: false
allowHostPID: false
allowHostPorts: false
allowPrivilegedContainer: false
allowedCapabilities: null
defaultAddCapabilities: null
requiredDropCapabilities:
- KILL
- MKNOD
- SETUID
- SETGID
fsGroup:
  type: MustRunAs
  ranges:
  - min: 1
    max: 65535
runAsUser:
  type: MustRunAsNonRoot
seLinuxContext:
  type: MustRunAs
volumes:
- configMap
- downwardAPI
- emptyDir
- persistentVolumeClaim
- projected
- secret
```

---

## 🔄 CI/CD流水线集成

### GitHub Actions工作流

```yaml
# .github/workflows/docker-build-deploy.yml
name: Docker Build and Deploy

on:
  push:
    branches: [ main, develop ]
    tags: [ 'v*' ]
  pull_request:
    branches: [ main ]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  # 代码质量检查
  code-quality:
    runs-on: ubuntu-latest
    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'

    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-

    - name: Download dependencies
      run: go mod download

    - name: Run tests
      run: go test -v -race -coverprofile=coverage.out ./...

    - name: Run linter
      uses: golangci/golangci-lint-action@v3
      with:
        version: latest

    - name: Security scan
      uses: securecodewarrior/github-action-gosec@master
      with:
        args: './...'

    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out

  # 安全扫描
  security-scan:
    runs-on: ubuntu-latest
    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Run Trivy vulnerability scanner
      uses: aquasecurity/trivy-action@master
      with:
        scan-type: 'fs'
        scan-ref: '.'
        format: 'sarif'
        output: 'trivy-results.sarif'

    - name: Upload Trivy scan results to GitHub Security tab
      uses: github/codeql-action/upload-sarif@v2
      with:
        sarif_file: 'trivy-results.sarif'

  # 构建和推送镜像
  build-and-push:
    needs: [code-quality, security-scan]
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v3

    - name: Log in to Container Registry
      uses: docker/login-action@v3
      with:
        registry: ${{ env.REGISTRY }}
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}

    - name: Extract metadata
      id: meta
      uses: docker/metadata-action@v5
      with:
        images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
        tags: |
          type=ref,event=branch
          type=ref,event=pr
          type=semver,pattern={{version}}
          type=semver,pattern={{major}}.{{minor}}
          type=sha,prefix={{branch}}-

    - name: Build and push Docker image
      uses: docker/build-push-action@v5
      with:
        context: .
        platforms: linux/amd64,linux/arm64
        push: true
        tags: ${{ steps.meta.outputs.tags }}
        labels: ${{ steps.meta.outputs.labels }}
        cache-from: type=gha
        cache-to: type=gha,mode=max
        build-args: |
          VERSION=${{ steps.meta.outputs.version }}
          BUILD_DATE=${{ steps.meta.outputs.created }}
          GIT_COMMIT=${{ github.sha }}

    - name: Scan Docker image
      uses: aquasecurity/trivy-action@master
      with:
        image-ref: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ steps.meta.outputs.version }}
        format: 'sarif'
        output: 'trivy-image-results.sarif'

    - name: Upload image scan results
      uses: github/codeql-action/upload-sarif@v2
      with:
        sarif_file: 'trivy-image-results.sarif'

  # 部署到Kubernetes
  deploy:
    needs: build-and-push
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up kubectl
      uses: azure/setup-kubectl@v3
      with:
        version: 'latest'

    - name: Set up Helm
      uses: azure/setup-helm@v3
      with:
        version: 'latest'

    - name: Configure kubectl
      run: |
        echo "${{ secrets.KUBECONFIG }}" | base64 -d > kubeconfig
        export KUBECONFIG=kubeconfig

    - name: Deploy to staging
      run: |
        helm upgrade --install mall-go-staging ./helm/mall-go \
          --namespace mall-go-staging \
          --create-namespace \
          --values ./helm/mall-go/values-staging.yaml \
          --set image.tag=${{ github.sha }} \
          --wait \
          --timeout 10m

    - name: Run smoke tests
      run: |
        kubectl wait --for=condition=ready pod -l app=mall-go -n mall-go-staging --timeout=300s
        kubectl run smoke-test --rm -i --restart=Never --image=curlimages/curl -- \
          curl -f http://mall-go-service.mall-go-staging.svc.cluster.local:8080/health

    - name: Deploy to production
      if: startsWith(github.ref, 'refs/tags/v')
      run: |
        helm upgrade --install mall-go-prod ./helm/mall-go \
          --namespace mall-go-prod \
          --create-namespace \
          --values ./helm/mall-go/values-prod.yaml \
          --set image.tag=${{ github.ref_name }} \
          --wait \
          --timeout 15m

  # 通知
  notify:
    needs: [deploy]
    runs-on: ubuntu-latest
    if: always()

    steps:
    - name: Notify Slack
      uses: 8398a7/action-slack@v3
      with:
        status: ${{ job.status }}
        channel: '#deployments'
        webhook_url: ${{ secrets.SLACK_WEBHOOK }}
        fields: repo,message,commit,author,action,eventName,ref,workflow
```

### GitLab CI配置

```yaml
# .gitlab-ci.yml
stages:
  - test
  - security
  - build
  - deploy

variables:
  DOCKER_DRIVER: overlay2
  DOCKER_TLS_CERTDIR: "/certs"
  REGISTRY: $CI_REGISTRY
  IMAGE_NAME: $CI_PROJECT_PATH
  KUBECONFIG: /tmp/kubeconfig

# 代码测试
test:
  stage: test
  image: golang:1.21-alpine
  services:
    - mysql:8.0
    - redis:7-alpine
  variables:
    MYSQL_ROOT_PASSWORD: "123456"
    MYSQL_DATABASE: "mall_go_test"
    REDIS_URL: "redis://redis:6379"
  before_script:
    - apk add --no-cache git make
    - go mod download
  script:
    - go test -v -race -coverprofile=coverage.out ./...
    - go tool cover -html=coverage.out -o coverage.html
  coverage: '/coverage: \d+\.\d+% of statements/'
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage.xml
    paths:
      - coverage.html
    expire_in: 1 week
  only:
    - merge_requests
    - main
    - develop

# 安全扫描
security:
  stage: security
  image: aquasec/trivy:latest
  script:
    - trivy fs --format json --output trivy-report.json .
    - trivy fs --format template --template "@contrib/sarif.tpl" --output trivy-report.sarif .
  artifacts:
    reports:
      sast: trivy-report.sarif
    paths:
      - trivy-report.json
    expire_in: 1 week
  only:
    - merge_requests
    - main
    - develop

# 构建镜像
build:
  stage: build
  image: docker:24-dind
  services:
    - docker:24-dind
  before_script:
    - echo $CI_REGISTRY_PASSWORD | docker login -u $CI_REGISTRY_USER --password-stdin $CI_REGISTRY
  script:
    - |
      if [[ "$CI_COMMIT_BRANCH" == "$CI_DEFAULT_BRANCH" ]]; then
        tag=""
        echo "Running on default branch '$CI_DEFAULT_BRANCH': tag = 'latest'"
      else
        tag=":$CI_COMMIT_REF_SLUG"
        echo "Running on branch '$CI_COMMIT_BRANCH': tag = $tag"
      fi
    - docker build --pull -t "$CI_REGISTRY_IMAGE${tag}" .
    - docker push "$CI_REGISTRY_IMAGE${tag}"
    # 镜像安全扫描
    - docker run --rm -v /var/run/docker.sock:/var/run/docker.sock aquasec/trivy image "$CI_REGISTRY_IMAGE${tag}"
  rules:
    - if: $CI_COMMIT_BRANCH
      exists:
        - Dockerfile

# 部署到测试环境
deploy_staging:
  stage: deploy
  image: alpine/helm:latest
  before_script:
    - apk add --no-cache kubectl
    - echo "$KUBECONFIG_STAGING" | base64 -d > $KUBECONFIG
    - kubectl config use-context staging
  script:
    - |
      helm upgrade --install mall-go-staging ./helm/mall-go \
        --namespace mall-go-staging \
        --create-namespace \
        --values ./helm/mall-go/values-staging.yaml \
        --set image.tag=$CI_COMMIT_SHA \
        --wait \
        --timeout 10m
  environment:
    name: staging
    url: https://staging-api.mall.example.com
  only:
    - develop

# 部署到生产环境
deploy_production:
  stage: deploy
  image: alpine/helm:latest
  before_script:
    - apk add --no-cache kubectl
    - echo "$KUBECONFIG_PRODUCTION" | base64 -d > $KUBECONFIG
    - kubectl config use-context production
  script:
    - |
      helm upgrade --install mall-go-prod ./helm/mall-go \
        --namespace mall-go-prod \
        --create-namespace \
        --values ./helm/mall-go/values-prod.yaml \
        --set image.tag=$CI_COMMIT_TAG \
        --wait \
        --timeout 15m
  environment:
    name: production
    url: https://api.mall.example.com
  when: manual
  only:
    - tags
```

---

## 🏢 企业级容器化方案

### Mall-Go完整容器化架构

```go
// 企业级容器化架构设计
package enterprise

import (
    "context"
    "time"
)

// 企业级容器化方案
type EnterpriseContainerization struct {
    // 基础设施层
    Infrastructure InfrastructureLayer `json:"infrastructure"`

    // 平台层
    Platform PlatformLayer `json:"platform"`

    // 应用层
    Application ApplicationLayer `json:"application"`

    // 运维层
    Operations OperationsLayer `json:"operations"`

    // 安全层
    Security SecurityLayer `json:"security"`

    // 监控层
    Monitoring MonitoringLayer `json:"monitoring"`
}

// 基础设施层
type InfrastructureLayer struct {
    // 容器运行时
    ContainerRuntime struct {
        Engine  string `json:"engine"`   // docker, containerd, cri-o
        Version string `json:"version"`
        Config  map[string]interface{} `json:"config"`
    } `json:"container_runtime"`

    // 容器编排
    Orchestration struct {
        Platform string `json:"platform"` // kubernetes, docker-swarm, nomad
        Version  string `json:"version"`
        Cluster  ClusterConfig `json:"cluster"`
    } `json:"orchestration"`

    // 存储
    Storage struct {
        CSI        []CSIDriver `json:"csi"`
        VolumeTypes []string   `json:"volume_types"`
        BackupStrategy BackupStrategy `json:"backup_strategy"`
    } `json:"storage"`

    // 网络
    Network struct {
        CNI     string `json:"cni"`      // calico, flannel, weave
        Ingress string `json:"ingress"`  // nginx, traefik, istio
        LoadBalancer string `json:"load_balancer"`
        ServiceMesh  string `json:"service_mesh"`
    } `json:"network"`
}

// 集群配置
type ClusterConfig struct {
    Masters []NodeConfig `json:"masters"`
    Workers []NodeConfig `json:"workers"`
    ETCD    ETCDConfig   `json:"etcd"`
    HA      HAConfig     `json:"ha"`
}

type NodeConfig struct {
    Name     string            `json:"name"`
    IP       string            `json:"ip"`
    Role     string            `json:"role"`
    Labels   map[string]string `json:"labels"`
    Taints   []Taint          `json:"taints"`
    Resources ResourceSpec     `json:"resources"`
}

type Taint struct {
    Key    string `json:"key"`
    Value  string `json:"value"`
    Effect string `json:"effect"`
}

type ResourceSpec struct {
    CPU    string `json:"cpu"`
    Memory string `json:"memory"`
    Disk   string `json:"disk"`
}

// 平台层
type PlatformLayer struct {
    // 镜像仓库
    Registry struct {
        Type     string   `json:"type"`     // harbor, nexus, artifactory
        URL      string   `json:"url"`
        Auth     AuthConfig `json:"auth"`
        Replication []ReplicationConfig `json:"replication"`
        Scanning ScanningConfig `json:"scanning"`
    } `json:"registry"`

    // 包管理
    PackageManager struct {
        Helm struct {
            Version    string   `json:"version"`
            Repos      []string `json:"repos"`
            ChartMuseum string  `json:"chart_museum"`
        } `json:"helm"`

        Kustomize struct {
            Version string `json:"version"`
            Bases   []string `json:"bases"`
        } `json:"kustomize"`
    } `json:"package_manager"`

    // 配置管理
    ConfigManagement struct {
        External struct {
            Vault    VaultConfig    `json:"vault"`
            Consul   ConsulConfig   `json:"consul"`
            ETCD     ETCDConfig     `json:"etcd"`
        } `json:"external"`

        Native struct {
            ConfigMaps []string `json:"config_maps"`
            Secrets    []string `json:"secrets"`
        } `json:"native"`
    } `json:"config_management"`
}

// 应用层
type ApplicationLayer struct {
    // 微服务架构
    Microservices []MicroserviceConfig `json:"microservices"`

    // 数据层
    DataLayer struct {
        Databases []DatabaseConfig `json:"databases"`
        Caches    []CacheConfig    `json:"caches"`
        MessageQueues []MQConfig   `json:"message_queues"`
    } `json:"data_layer"`

    // 网关层
    Gateway struct {
        APIGateway   GatewayConfig `json:"api_gateway"`
        LoadBalancer LBConfig      `json:"load_balancer"`
        CDN          CDNConfig     `json:"cdn"`
    } `json:"gateway"`
}

// 微服务配置
type MicroserviceConfig struct {
    Name        string            `json:"name"`
    Image       string            `json:"image"`
    Replicas    int               `json:"replicas"`
    Resources   ResourceSpec      `json:"resources"`
    Environment map[string]string `json:"environment"`
    Volumes     []VolumeMount     `json:"volumes"`
    Probes      ProbesConfig      `json:"probes"`
    Autoscaling AutoscalingConfig `json:"autoscaling"`
}

type ProbesConfig struct {
    Liveness  ProbeConfig `json:"liveness"`
    Readiness ProbeConfig `json:"readiness"`
    Startup   ProbeConfig `json:"startup"`
}

type ProbeConfig struct {
    HTTPGet             HTTPGetAction `json:"http_get"`
    InitialDelaySeconds int           `json:"initial_delay_seconds"`
    PeriodSeconds       int           `json:"period_seconds"`
    TimeoutSeconds      int           `json:"timeout_seconds"`
    FailureThreshold    int           `json:"failure_threshold"`
}

type HTTPGetAction struct {
    Path   string `json:"path"`
    Port   int    `json:"port"`
    Scheme string `json:"scheme"`
}

type AutoscalingConfig struct {
    MinReplicas int      `json:"min_replicas"`
    MaxReplicas int      `json:"max_replicas"`
    Metrics     []string `json:"metrics"`
    Behavior    ScalingBehavior `json:"behavior"`
}

type ScalingBehavior struct {
    ScaleUp   ScalingPolicy `json:"scale_up"`
    ScaleDown ScalingPolicy `json:"scale_down"`
}

type ScalingPolicy struct {
    StabilizationWindowSeconds int `json:"stabilization_window_seconds"`
    Policies                   []ScalingPolicyRule `json:"policies"`
}

type ScalingPolicyRule struct {
    Type          string `json:"type"`
    Value         int    `json:"value"`
    PeriodSeconds int    `json:"period_seconds"`
}
```

### 部署策略对比

| 部署策略 | 优点 | 缺点 | 适用场景 |
|----------|------|------|----------|
| **蓝绿部署** | 零停机、快速回滚 | 资源消耗大 | 关键业务系统 |
| **滚动更新** | 资源利用率高 | 更新时间长 | 一般业务系统 |
| **金丝雀部署** | 风险可控、渐进式 | 复杂度高 | 新功能发布 |
| **A/B测试** | 数据驱动决策 | 需要流量分割 | 功能验证 |

```go
// 部署策略实现
type DeploymentStrategy struct {
    Type       string                 `json:"type"`
    Config     map[string]interface{} `json:"config"`
    Validation ValidationConfig       `json:"validation"`
    Rollback   RollbackConfig         `json:"rollback"`
}

// 蓝绿部署配置
type BlueGreenConfig struct {
    BlueEnvironment  EnvironmentConfig `json:"blue_environment"`
    GreenEnvironment EnvironmentConfig `json:"green_environment"`
    SwitchStrategy   SwitchStrategy    `json:"switch_strategy"`
    HealthCheck      HealthCheckConfig `json:"health_check"`
}

// 金丝雀部署配置
type CanaryConfig struct {
    Steps []CanaryStep `json:"steps"`
    Analysis AnalysisConfig `json:"analysis"`
    TrafficSplit TrafficSplitConfig `json:"traffic_split"`
}

type CanaryStep struct {
    Weight      int           `json:"weight"`
    Duration    time.Duration `json:"duration"`
    Metrics     []string      `json:"metrics"`
    Thresholds  map[string]float64 `json:"thresholds"`
}
```

---

## 🎯 面试常考知识点

### 核心概念面试题

**Q1: 什么是Docker多阶段构建？有什么优势？**

**标准答案：**
Docker多阶段构建允许在单个Dockerfile中使用多个FROM语句，每个FROM开始一个新的构建阶段。

**主要优势：**
1. **减小镜像体积**：只保留运行时必需的文件，排除构建工具和源代码
2. **提高安全性**：最终镜像不包含编译器、构建工具等潜在安全风险
3. **简化构建流程**：在一个文件中管理整个构建过程
4. **并行构建**：不同阶段可以并行执行，提高构建效率
5. **缓存优化**：每个阶段都可以利用Docker层缓存

**实际应用：**
```dockerfile
# 构建阶段
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o main .

# 运行阶段
FROM alpine:3.18
COPY --from=builder /app/main .
CMD ["./main"]
```

**Q2: Kubernetes中Deployment、Service、Ingress的作用和关系？**

**标准答案：**

| 组件 | 作用 | 关系 |
|------|------|------|
| **Deployment** | 管理Pod的创建、更新、扩缩容 | 创建和管理Pod |
| **Service** | 为Pod提供稳定的网络访问入口 | 选择和代理Pod |
| **Ingress** | 提供HTTP/HTTPS路由和负载均衡 | 路由到Service |

**工作流程：**
1. Deployment创建和管理Pod副本
2. Service通过标签选择器关联Pod，提供稳定的ClusterIP
3. Ingress根据域名和路径规则路由到对应的Service
4. 外部流量：Internet → Ingress → Service → Pod

**Q3: 容器化应用如何实现高可用？**

**标准答案：**

**1. 应用层高可用：**
- 多副本部署（replicas > 1）
- Pod反亲和性（避免单点故障）
- 健康检查（liveness/readiness probe）
- 优雅关闭（graceful shutdown）

**2. 基础设施高可用：**
- 多可用区部署
- 节点故障自动恢复
- 存储冗余备份
- 网络多路径

**3. 数据层高可用：**
- 数据库主从复制
- 分布式存储
- 定期备份和恢复测试

**Q4: 如何优化Docker镜像大小？**

**标准答案：**

**1. 选择合适的基础镜像：**
```dockerfile
# 不推荐：使用完整的Ubuntu镜像
FROM ubuntu:20.04

# 推荐：使用Alpine或Distroless镜像
FROM alpine:3.18
FROM gcr.io/distroless/static-debian11
```

**2. 多阶段构建：**
```dockerfile
# 构建阶段
FROM golang:1.21-alpine AS builder
RUN go build -o app .

# 运行阶段
FROM alpine:3.18
COPY --from=builder /app/app .
```

**3. 减少层数和清理缓存：**
```dockerfile
# 不推荐：多个RUN指令
RUN apt-get update
RUN apt-get install -y package1
RUN apt-get install -y package2

# 推荐：合并RUN指令并清理缓存
RUN apt-get update && \
    apt-get install -y package1 package2 && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*
```

**4. 使用.dockerignore：**
```
node_modules
.git
*.md
.env
```

**Q5: Kubernetes中如何实现配置管理？**

**标准答案：**

**1. ConfigMap（非敏感配置）：**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  database.host: "mysql.example.com"
  database.port: "3306"
```

**2. Secret（敏感配置）：**
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: app-secret
type: Opaque
data:
  password: cGFzc3dvcmQ=  # base64编码
```

**3. 使用方式：**
- 环境变量注入
- 文件挂载
- 命令行参数

**4. 外部配置管理：**
- Vault（密钥管理）
- Consul（配置中心）
- External Secrets Operator

### 技术实现面试题

**Q6: 如何设计一个企业级的容器化CI/CD流水线？**

**标准答案：**

**流水线阶段设计：**

1. **代码质量阶段**
   - 单元测试
   - 代码覆盖率检查
   - 静态代码分析
   - 安全漏洞扫描

2. **构建阶段**
   - 多阶段Docker构建
   - 镜像安全扫描
   - 镜像签名验证
   - 推送到镜像仓库

3. **部署阶段**
   - 部署到测试环境
   - 自动化测试
   - 部署到生产环境
   - 健康检查

4. **监控阶段**
   - 部署状态监控
   - 应用性能监控
   - 告警通知

**关键技术点：**
- GitOps工作流
- 蓝绿/金丝雀部署
- 自动回滚机制
- 环境一致性保证

**Q7: 容器化环境下如何处理日志收集？**

**标准答案：**

**1. 日志收集架构：**
```
应用容器 → 日志代理 → 日志聚合 → 存储 → 可视化
```

**2. 实现方案：**

**方案一：Sidecar模式**
```yaml
containers:
- name: app
  image: mall-go:latest
- name: log-agent
  image: fluent/fluent-bit:latest
  volumeMounts:
  - name: logs
    mountPath: /var/log
```

**方案二：DaemonSet模式**
```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: fluentd
spec:
  template:
    spec:
      containers:
      - name: fluentd
        image: fluent/fluentd:latest
        volumeMounts:
        - name: varlog
          mountPath: /var/log
```

**3. 技术栈选择：**
- **ELK Stack**: Elasticsearch + Logstash + Kibana
- **EFK Stack**: Elasticsearch + Fluentd + Kibana
- **Loki Stack**: Grafana Loki + Promtail + Grafana
- **云原生方案**: AWS CloudWatch, GCP Stackdriver

**Q8: 如何在Kubernetes中实现服务间通信安全？**

**标准答案：**

**1. 网络策略（Network Policy）：**
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: deny-all
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
```

**2. 服务网格（Service Mesh）：**
- **Istio**: 提供mTLS、流量管理、安全策略
- **Linkerd**: 轻量级服务网格
- **Consul Connect**: HashiCorp的服务网格解决方案

**3. TLS加密：**
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: tls-secret
type: kubernetes.io/tls
data:
  tls.crt: <base64-encoded-cert>
  tls.key: <base64-encoded-key>
```

**4. 身份认证：**
- JWT Token认证
- OAuth 2.0/OIDC
- 服务账户（ServiceAccount）
- Pod身份验证

**Q9: 容器化应用如何实现数据持久化？**

**标准答案：**

**1. 存储类型：**

| 存储类型 | 特点 | 适用场景 |
|----------|------|----------|
| **EmptyDir** | 临时存储，Pod删除时数据丢失 | 缓存、临时文件 |
| **HostPath** | 挂载主机目录 | 日志收集、监控 |
| **PV/PVC** | 持久化存储，生命周期独立 | 数据库、文件存储 |
| **ConfigMap/Secret** | 配置文件存储 | 应用配置 |

**2. 存储类（StorageClass）：**
```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: fast-ssd
provisioner: kubernetes.io/aws-ebs
parameters:
  type: gp3
  iops: "3000"
  throughput: "125"
reclaimPolicy: Retain
allowVolumeExpansion: true
```

**3. 数据备份策略：**
- 定期快照
- 跨区域复制
- 增量备份
- 恢复测试

**Q10: 如何监控容器化应用的性能？**

**标准答案：**

**1. 监控层次：**
- **基础设施监控**: 节点CPU、内存、磁盘、网络
- **容器监控**: 容器资源使用、状态
- **应用监控**: 业务指标、错误率、响应时间
- **用户体验监控**: 页面加载时间、用户行为

**2. 监控技术栈：**
```yaml
# Prometheus配置
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
data:
  prometheus.yml: |
    global:
      scrape_interval: 15s
    scrape_configs:
    - job_name: 'kubernetes-pods'
      kubernetes_sd_configs:
      - role: pod
```

**3. 关键指标：**
- **RED指标**: Rate(请求率)、Errors(错误率)、Duration(响应时间)
- **USE指标**: Utilization(使用率)、Saturation(饱和度)、Errors(错误)
- **四个黄金信号**: 延迟、流量、错误、饱和度

**4. 告警策略：**
```yaml
groups:
- name: mall-go-alerts
  rules:
  - alert: HighErrorRate
    expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
    for: 5m
    annotations:
      summary: "High error rate detected"
```

---

## 🏋️ 练习题

### 练习1：设计Mall-Go完整容器化方案

**题目描述：**
为Mall-Go电商系统设计一套完整的企业级容器化部署方案，包括开发、测试、生产环境。

**要求：**
1. 设计多阶段Dockerfile，优化镜像大小
2. 编写完整的Kubernetes YAML配置
3. 设计Helm Chart，支持多环境部署
4. 配置CI/CD流水线，实现自动化部署
5. 实现监控、日志、安全等运维功能
6. 设计灾难恢复和备份策略

**参考架构：**
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   开发环境       │    │   测试环境       │    │   生产环境       │
│                │    │                │    │                │
│ • 单节点K8s     │    │ • 多节点K8s     │    │ • 高可用K8s     │
│ • 本地存储      │    │ • 网络存储      │    │ • 分布式存储     │
│ • 基础监控      │    │ • 完整监控      │    │ • 企业级监控     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### 练习2：实现蓝绿部署策略

**题目描述：**
为Mall-Go实现蓝绿部署策略，确保零停机更新。

**要求：**
1. 设计蓝绿环境的Kubernetes配置
2. 实现流量切换机制
3. 配置健康检查和回滚策略
4. 编写自动化部署脚本
5. 设计监控和告警机制

**实现框架：**
```go
type BlueGreenDeployment struct {
    BlueEnvironment  Environment
    GreenEnvironment Environment
    TrafficManager   TrafficManager
    HealthChecker    HealthChecker
    RollbackManager  RollbackManager
}

func (bg *BlueGreenDeployment) Deploy(version string) error {
    // 实现蓝绿部署逻辑
}

func (bg *BlueGreenDeployment) SwitchTraffic() error {
    // 实现流量切换逻辑
}

func (bg *BlueGreenDeployment) Rollback() error {
    // 实现回滚逻辑
}
```

### 练习3：构建企业级监控系统

**题目描述：**
为容器化的Mall-Go系统构建完整的监控体系。

**要求：**
1. 部署Prometheus + Grafana监控栈
2. 配置应用指标收集
3. 设计监控大盘和告警规则
4. 实现日志聚合和分析
5. 配置链路追踪
6. 设计性能基线和SLA

**监控指标设计：**
```yaml
# 应用指标
- http_requests_total
- http_request_duration_seconds
- http_requests_in_flight
- database_connections_active
- cache_hit_ratio
- queue_messages_pending

# 基础设施指标
- node_cpu_usage
- node_memory_usage
- node_disk_usage
- pod_cpu_usage
- pod_memory_usage
- container_restarts_total
```

---

## 📚 章节总结

### 🎯 核心知识点回顾

通过本章学习，我们全面掌握了Docker容器化部署的核心技术：

1. **Docker基础与Go应用容器化**
   - 理解了容器技术原理和Docker核心概念
   - 掌握了Go应用的容器化最佳实践
   - 学会了编写高质量的Dockerfile和docker-compose配置

2. **多阶段构建优化**
   - 深入理解了多阶段构建的原理和优势
   - 掌握了镜像大小优化的各种技术
   - 学会了构建安全、高效的生产镜像

3. **Kubernetes部署策略**
   - 全面掌握了K8s核心资源对象的使用
   - 理解了Deployment、Service、Ingress的工作原理
   - 学会了设计高可用、可扩展的K8s应用架构

4. **Helm图表管理**
   - 掌握了Helm包管理器的使用方法
   - 学会了编写可复用的Helm Chart
   - 理解了模板化部署和版本管理的重要性

5. **容器安全配置**
   - 深入了解了容器安全的各个层面
   - 掌握了安全扫描、镜像签名等安全实践
   - 学会了配置Pod安全策略和网络策略

6. **CI/CD流水线集成**
   - 掌握了GitHub Actions和GitLab CI的使用
   - 学会了设计完整的容器化CI/CD流水线
   - 理解了GitOps和自动化部署的最佳实践

7. **企业级容器化方案**
   - 通过Mall-Go项目实践，掌握了企业级容器化架构设计
   - 学会了处理复杂的微服务部署场景
   - 理解了生产环境的运维和监控要求

### 🚀 实践应用价值

1. **技术架构能力**：能够设计和实现企业级的容器化架构
2. **运维自动化能力**：掌握了CI/CD和GitOps的实践方法
3. **安全防护能力**：具备了容器安全的全面防护技能
4. **问题解决能力**：掌握了容器化环境的故障诊断和处理
5. **成本优化能力**：理解了资源优化和成本控制的方法

### 🎓 下一步学习建议

1. **深入学习监控系统**：学习下一章的监控与日志系统
2. **实践项目部署**：在实际项目中应用容器化技术
3. **云原生技术栈**：学习Service Mesh、Serverless等技术
4. **安全深化学习**：深入学习容器安全和零信任架构
5. **性能优化实践**：通过实际项目验证性能优化效果

### 💡 关键技术要点

- **容器化不仅是技术选择，更是架构思维的转变**
- **多阶段构建是生产环境的必备技能**，能显著提升安全性和效率
- **Kubernetes是容器编排的事实标准**，掌握其核心概念至关重要
- **Helm简化了复杂应用的部署管理**，是企业级部署的重要工具
- **安全是容器化的重中之重**，需要从多个层面进行防护
- **CI/CD是容器化价值实现的关键**，自动化程度决定了效率
- **监控和可观测性是生产环境的基础**，有助于快速定位问题

### 🌟 技术发展趋势

1. **云原生安全**：零信任架构、运行时安全防护
2. **边缘计算容器化**：轻量级容器运行时、边缘K8s
3. **AI/ML容器化**：GPU容器调度、模型服务化部署
4. **绿色计算**：能耗优化、碳中和的容器化实践
5. **WebAssembly容器**：更轻量、更安全的容器技术

### 🔗 与其他章节的联系

- **分布式系统**：为容器化应用提供理论基础
- **监控与日志**：为容器化环境提供可观测性支持
- **性能优化**：容器化应用的性能调优实践
- **生产实践**：容器化在生产环境的综合应用

通过本章的学习，你已经具备了设计和实现企业级容器化解决方案的能力。容器化技术是现代软件交付的核心，掌握这些技能将为你的职业发展提供强有力的支撑！ 🚀

---

*"容器化不仅改变了我们部署应用的方式，更重要的是改变了我们思考软件架构的方式。它让我们能够构建更加灵活、可扩展、可维护的系统。掌握容器化技术，就是掌握了现代软件工程的核心竞争力！"* 🐳✨
