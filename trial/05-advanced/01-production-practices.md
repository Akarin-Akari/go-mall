# 高级篇第一章：生产实践与运维 🚀

> *"代码写得好不好，生产环境说了算。掌握生产实践，就掌握了从开发到运维的完整技能链！"* 💪

## 📚 本章学习目标

通过本章学习，你将掌握：

- 🐳 **容器化部署**：Docker容器化、Kubernetes编排、云原生部署
- 📊 **监控与可观测性**：指标监控、链路追踪、日志聚合、告警系统
- 🔧 **性能优化**：Go程序性能分析、内存优化、并发优化
- 🛡️ **安全实践**：认证授权、数据加密、安全扫描、漏洞防护
- 📈 **扩展性设计**：水平扩展、负载均衡、缓存策略、数据库优化
- 🔄 **CI/CD流水线**：自动化构建、测试、部署、回滚
- 🚨 **故障处理**：故障诊断、应急响应、灾难恢复
- 🏢 **企业级实践**：结合mall-go项目的生产环境最佳实践

---

## 🐳 容器化部署

### Docker容器化实践

```go
// Docker容器化配置
package deployment

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
)

// Docker配置
type DockerConfig struct {
    BaseImage    string            `json:"base_image"`
    WorkDir      string            `json:"work_dir"`
    ExposedPorts []int             `json:"exposed_ports"`
    Environment  map[string]string `json:"environment"`
    Volumes      []VolumeMount     `json:"volumes"`
    HealthCheck  HealthCheckConfig `json:"health_check"`
}

type VolumeMount struct {
    HostPath      string `json:"host_path"`
    ContainerPath string `json:"container_path"`
    ReadOnly      bool   `json:"read_only"`
}

type HealthCheckConfig struct {
    Command     []string `json:"command"`
    Interval    string   `json:"interval"`
    Timeout     string   `json:"timeout"`
    Retries     int      `json:"retries"`
    StartPeriod string   `json:"start_period"`
}

// 生成Dockerfile
func GenerateDockerfile(config DockerConfig, appName string) string {
    dockerfile := fmt.Sprintf(`# 多阶段构建 - 构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git ca-certificates tzdata

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o %s .

# 运行阶段
FROM %s

# 安装ca证书和时区数据
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR %s

# 从构建阶段复制二进制文件
COPY --from=builder /app/%s .

# 复制配置文件
COPY --from=builder /app/configs ./configs

# 更改文件所有者
RUN chown -R appuser:appgroup .

# 切换到非root用户
USER appuser

# 暴露端口
`, appName, config.BaseImage, config.WorkDir, appName)

    // 添加暴露端口
    for _, port := range config.ExposedPorts {
        dockerfile += fmt.Sprintf("EXPOSE %d\n", port)
    }

    // 添加环境变量
    for key, value := range config.Environment {
        dockerfile += fmt.Sprintf("ENV %s=%s\n", key, value)
    }

    // 添加健康检查
    if len(config.HealthCheck.Command) > 0 {
        dockerfile += fmt.Sprintf(`
# 健康检查
HEALTHCHECK --interval=%s --timeout=%s --start-period=%s --retries=%d \
    CMD %s
`,
            config.HealthCheck.Interval,
            config.HealthCheck.Timeout,
            config.HealthCheck.StartPeriod,
            config.HealthCheck.Retries,
            config.HealthCheck.Command[0])
    }

    // 添加启动命令
    dockerfile += fmt.Sprintf("\n# 启动应用\nCMD [\"./%s\"]\n", appName)

    return dockerfile
}

// 生成docker-compose.yml
func GenerateDockerCompose(services []ServiceConfig) string {
    compose := `version: '3.8'

services:
`

    for _, service := range services {
        compose += fmt.Sprintf(`  %s:
    build:
      context: .
      dockerfile: Dockerfile
    image: %s:latest
    container_name: %s
    restart: unless-stopped
    ports:
`, service.Name, service.Image, service.ContainerName)

        for _, port := range service.Ports {
            compose += fmt.Sprintf("      - \"%d:%d\"\n", port.Host, port.Container)
        }

        if len(service.Environment) > 0 {
            compose += "    environment:\n"
            for key, value := range service.Environment {
                compose += fmt.Sprintf("      - %s=%s\n", key, value)
            }
        }

        if len(service.Volumes) > 0 {
            compose += "    volumes:\n"
            for _, volume := range service.Volumes {
                compose += fmt.Sprintf("      - %s:%s", volume.HostPath, volume.ContainerPath)
                if volume.ReadOnly {
                    compose += ":ro"
                }
                compose += "\n"
            }
        }

        if len(service.DependsOn) > 0 {
            compose += "    depends_on:\n"
            for _, dep := range service.DependsOn {
                compose += fmt.Sprintf("      - %s\n", dep)
            }
        }

        compose += "\n"
    }

    // 添加网络配置
    compose += `networks:
  default:
    driver: bridge

volumes:
  postgres_data:
  redis_data:
`

    return compose
}

type ServiceConfig struct {
    Name          string            `json:"name"`
    Image         string            `json:"image"`
    ContainerName string            `json:"container_name"`
    Ports         []PortMapping     `json:"ports"`
    Environment   map[string]string `json:"environment"`
    Volumes       []VolumeMount     `json:"volumes"`
    DependsOn     []string          `json:"depends_on"`
}

type PortMapping struct {
    Host      int `json:"host"`
    Container int `json:"container"`
}

// Mall-Go项目的Docker配置示例
func GetMallGoDockerConfig() []ServiceConfig {
    return []ServiceConfig{
        {
            Name:          "mall-go-api",
            Image:         "mall-go/api",
            ContainerName: "mall-go-api",
            Ports: []PortMapping{
                {Host: 8080, Container: 8080},
            },
            Environment: map[string]string{
                "GIN_MODE":     "release",
                "DB_HOST":      "postgres",
                "DB_PORT":      "5432",
                "REDIS_HOST":   "redis",
                "REDIS_PORT":   "6379",
                "LOG_LEVEL":    "info",
            },
            Volumes: []VolumeMount{
                {
                    HostPath:      "./logs",
                    ContainerPath: "/app/logs",
                    ReadOnly:      false,
                },
                {
                    HostPath:      "./configs",
                    ContainerPath: "/app/configs",
                    ReadOnly:      true,
                },
            },
            DependsOn: []string{"postgres", "redis"},
        },
        {
            Name:          "postgres",
            Image:         "postgres:15-alpine",
            ContainerName: "mall-go-postgres",
            Ports: []PortMapping{
                {Host: 5432, Container: 5432},
            },
            Environment: map[string]string{
                "POSTGRES_DB":       "mall_go",
                "POSTGRES_USER":     "mall_user",
                "POSTGRES_PASSWORD": "mall_password",
                "PGDATA":           "/var/lib/postgresql/data/pgdata",
            },
            Volumes: []VolumeMount{
                {
                    HostPath:      "postgres_data",
                    ContainerPath: "/var/lib/postgresql/data",
                    ReadOnly:      false,
                },
            },
        },
        {
            Name:          "redis",
            Image:         "redis:7-alpine",
            ContainerName: "mall-go-redis",
            Ports: []PortMapping{
                {Host: 6379, Container: 6379},
            },
            Volumes: []VolumeMount{
                {
                    HostPath:      "redis_data",
                    ContainerPath: "/data",
                    ReadOnly:      false,
                },
            },
        },
    }
}
```

### Kubernetes部署实践

```go
// Kubernetes部署配置
package k8s

import (
    "fmt"
    "strings"
)

// Kubernetes资源配置
type K8sConfig struct {
    Namespace   string                 `json:"namespace"`
    Deployment  DeploymentConfig       `json:"deployment"`
    Service     ServiceConfig          `json:"service"`
    Ingress     IngressConfig          `json:"ingress"`
    ConfigMap   map[string]string      `json:"config_map"`
    Secret      map[string]string      `json:"secret"`
    HPA         HPAConfig              `json:"hpa"`
}

type DeploymentConfig struct {
    Name         string            `json:"name"`
    Image        string            `json:"image"`
    Replicas     int32             `json:"replicas"`
    Port         int32             `json:"port"`
    Resources    ResourceConfig    `json:"resources"`
    Environment  map[string]string `json:"environment"`
    HealthCheck  K8sHealthCheck    `json:"health_check"`
}

type ResourceConfig struct {
    Requests ResourceRequirements `json:"requests"`
    Limits   ResourceRequirements `json:"limits"`
}

type ResourceRequirements struct {
    CPU    string `json:"cpu"`
    Memory string `json:"memory"`
}

type K8sHealthCheck struct {
    LivenessProbe  ProbeConfig `json:"liveness_probe"`
    ReadinessProbe ProbeConfig `json:"readiness_probe"`
}

type ProbeConfig struct {
    Path                string `json:"path"`
    Port                int32  `json:"port"`
    InitialDelaySeconds int32  `json:"initial_delay_seconds"`
    PeriodSeconds       int32  `json:"period_seconds"`
    TimeoutSeconds      int32  `json:"timeout_seconds"`
    FailureThreshold    int32  `json:"failure_threshold"`
}

type ServiceConfig struct {
    Name     string        `json:"name"`
    Type     string        `json:"type"`
    Ports    []ServicePort `json:"ports"`
    Selector map[string]string `json:"selector"`
}

type ServicePort struct {
    Name       string `json:"name"`
    Port       int32  `json:"port"`
    TargetPort int32  `json:"target_port"`
    Protocol   string `json:"protocol"`
}

type IngressConfig struct {
    Name        string              `json:"name"`
    Host        string              `json:"host"`
    Paths       []IngressPath       `json:"paths"`
    TLS         []IngressTLS        `json:"tls"`
    Annotations map[string]string   `json:"annotations"`
}

type IngressPath struct {
    Path        string `json:"path"`
    PathType    string `json:"path_type"`
    ServiceName string `json:"service_name"`
    ServicePort int32  `json:"service_port"`
}

type IngressTLS struct {
    Hosts      []string `json:"hosts"`
    SecretName string   `json:"secret_name"`
}

type HPAConfig struct {
    Name                     string `json:"name"`
    MinReplicas              int32  `json:"min_replicas"`
    MaxReplicas              int32  `json:"max_replicas"`
    TargetCPUUtilization     int32  `json:"target_cpu_utilization"`
    TargetMemoryUtilization  int32  `json:"target_memory_utilization"`
}

// 生成Kubernetes YAML
func GenerateK8sYAML(config K8sConfig) string {
    var yaml strings.Builder

    // Namespace
    yaml.WriteString(generateNamespace(config.Namespace))
    yaml.WriteString("---\n")

    // ConfigMap
    if len(config.ConfigMap) > 0 {
        yaml.WriteString(generateConfigMap(config.Namespace, config.Deployment.Name, config.ConfigMap))
        yaml.WriteString("---\n")
    }

    // Secret
    if len(config.Secret) > 0 {
        yaml.WriteString(generateSecret(config.Namespace, config.Deployment.Name, config.Secret))
        yaml.WriteString("---\n")
    }

    // Deployment
    yaml.WriteString(generateDeployment(config.Namespace, config.Deployment))
    yaml.WriteString("---\n")

    // Service
    yaml.WriteString(generateService(config.Namespace, config.Service))
    yaml.WriteString("---\n")

    // Ingress
    if config.Ingress.Name != "" {
        yaml.WriteString(generateIngress(config.Namespace, config.Ingress))
        yaml.WriteString("---\n")
    }

    // HPA
    if config.HPA.Name != "" {
        yaml.WriteString(generateHPA(config.Namespace, config.HPA))
    }

    return yaml.String()
}

func generateNamespace(name string) string {
    return fmt.Sprintf(`apiVersion: v1
kind: Namespace
metadata:
  name: %s
`, name)
}

func generateDeployment(namespace string, config DeploymentConfig) string {
    deployment := fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: %s
  namespace: %s
  labels:
    app: %s
spec:
  replicas: %d
  selector:
    matchLabels:
      app: %s
  template:
    metadata:
      labels:
        app: %s
    spec:
      containers:
      - name: %s
        image: %s
        ports:
        - containerPort: %d
        resources:
          requests:
            cpu: %s
            memory: %s
          limits:
            cpu: %s
            memory: %s
`, config.Name, namespace, config.Name, config.Replicas, config.Name, config.Name,
        config.Name, config.Image, config.Port,
        config.Resources.Requests.CPU, config.Resources.Requests.Memory,
        config.Resources.Limits.CPU, config.Resources.Limits.Memory)

    // 添加环境变量
    if len(config.Environment) > 0 {
        deployment += "        env:\n"
        for key, value := range config.Environment {
            deployment += fmt.Sprintf("        - name: %s\n          value: \"%s\"\n", key, value)
        }
    }

    // 添加健康检查
    if config.HealthCheck.LivenessProbe.Path != "" {
        deployment += fmt.Sprintf(`        livenessProbe:
          httpGet:
            path: %s
            port: %d
          initialDelaySeconds: %d
          periodSeconds: %d
          timeoutSeconds: %d
          failureThreshold: %d
`, config.HealthCheck.LivenessProbe.Path, config.HealthCheck.LivenessProbe.Port,
            config.HealthCheck.LivenessProbe.InitialDelaySeconds,
            config.HealthCheck.LivenessProbe.PeriodSeconds,
            config.HealthCheck.LivenessProbe.TimeoutSeconds,
            config.HealthCheck.LivenessProbe.FailureThreshold)
    }

    if config.HealthCheck.ReadinessProbe.Path != "" {
        deployment += fmt.Sprintf(`        readinessProbe:
          httpGet:
            path: %s
            port: %d
          initialDelaySeconds: %d
          periodSeconds: %d
          timeoutSeconds: %d
          failureThreshold: %d
`, config.HealthCheck.ReadinessProbe.Path, config.HealthCheck.ReadinessProbe.Port,
            config.HealthCheck.ReadinessProbe.InitialDelaySeconds,
            config.HealthCheck.ReadinessProbe.PeriodSeconds,
            config.HealthCheck.ReadinessProbe.TimeoutSeconds,
            config.HealthCheck.ReadinessProbe.FailureThreshold)
    }

    return deployment
}

func generateService(namespace string, config ServiceConfig) string {
    service := fmt.Sprintf(`apiVersion: v1
kind: Service
metadata:
  name: %s
  namespace: %s
  labels:
    app: %s
spec:
  type: %s
  ports:
`, config.Name, namespace, config.Name, config.Type)

    for _, port := range config.Ports {
        service += fmt.Sprintf(`  - name: %s
    port: %d
    targetPort: %d
    protocol: %s
`, port.Name, port.Port, port.TargetPort, port.Protocol)
    }

    service += "  selector:\n"
    for key, value := range config.Selector {
        service += fmt.Sprintf("    %s: %s\n", key, value)
    }

    return service
}

func generateConfigMap(namespace, name string, data map[string]string) string {
    configMap := fmt.Sprintf(`apiVersion: v1
kind: ConfigMap
metadata:
  name: %s-config
  namespace: %s
data:
`, name, namespace)

    for key, value := range data {
        configMap += fmt.Sprintf("  %s: \"%s\"\n", key, value)
    }

    return configMap
}

func generateSecret(namespace, name string, data map[string]string) string {
    secret := fmt.Sprintf(`apiVersion: v1
kind: Secret
metadata:
  name: %s-secret
  namespace: %s
type: Opaque
data:
`, name, namespace)

    for key, value := range data {
        // 注意：实际使用时需要base64编码
        secret += fmt.Sprintf("  %s: %s\n", key, value)
    }

    return secret
}

func generateHPA(namespace string, config HPAConfig) string {
    return fmt.Sprintf(`apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: %s
  namespace: %s
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: %s
  minReplicas: %d
  maxReplicas: %d
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: %d
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: %d
`, config.Name, namespace, config.Name, config.MinReplicas, config.MaxReplicas,
        config.TargetCPUUtilization, config.TargetMemoryUtilization)
}
```

---

## 📊 监控与可观测性

### Prometheus监控集成

```go
// Prometheus监控集成
package monitoring

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// 监控指标定义
var (
    // HTTP请求计数器
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status_code"},
    )

    // HTTP请求延迟直方图
    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )

    // 数据库连接池指标
    dbConnectionsInUse = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "db_connections_in_use",
            Help: "Number of database connections currently in use",
        },
    )

    dbConnectionsIdle = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "db_connections_idle",
            Help: "Number of idle database connections",
        },
    )

    // 业务指标
    ordersTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "orders_total",
            Help: "Total number of orders",
        },
        []string{"status"},
    )

    orderAmount = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "order_amount",
            Help:    "Order amount distribution",
            Buckets: []float64{10, 50, 100, 500, 1000, 5000, 10000},
        },
        []string{"currency"},
    )

    // 系统指标
    goroutinesCount = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "goroutines_count",
            Help: "Number of goroutines",
        },
    )

    memoryUsage = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "memory_usage_bytes",
            Help: "Memory usage in bytes",
        },
    )
)

// Prometheus中间件
func PrometheusMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()

        // 处理请求
        c.Next()

        // 记录指标
        duration := time.Since(start).Seconds()
        statusCode := c.Writer.Status()

        httpRequestsTotal.WithLabelValues(
            c.Request.Method,
            c.FullPath(),
            http.StatusText(statusCode),
        ).Inc()

        httpRequestDuration.WithLabelValues(
            c.Request.Method,
            c.FullPath(),
        ).Observe(duration)
    }
}

// 监控服务
type MonitoringService struct {
    registry *prometheus.Registry
}

func NewMonitoringService() *MonitoringService {
    return &MonitoringService{
        registry: prometheus.NewRegistry(),
    }
}

// 启动监控服务
func (ms *MonitoringService) StartMetricsServer(port string) {
    http.Handle("/metrics", promhttp.Handler())
    go http.ListenAndServe(":"+port, nil)
}

// 更新系统指标
func (ms *MonitoringService) UpdateSystemMetrics() {
    go func() {
        ticker := time.NewTicker(30 * time.Second)
        defer ticker.Stop()

        for range ticker.C {
            // 更新goroutine数量
            goroutinesCount.Set(float64(runtime.NumGoroutine()))

            // 更新内存使用
            var m runtime.MemStats
            runtime.ReadMemStats(&m)
            memoryUsage.Set(float64(m.Alloc))
        }
    }()
}

// 业务指标记录
func RecordOrderCreated(status string, amount float64, currency string) {
    ordersTotal.WithLabelValues(status).Inc()
    orderAmount.WithLabelValues(currency).Observe(amount)
}

func UpdateDBConnectionMetrics(inUse, idle int) {
    dbConnectionsInUse.Set(float64(inUse))
    dbConnectionsIdle.Set(float64(idle))
}
```

### 分布式链路追踪

```go
// 分布式链路追踪
package tracing

import (
    "context"
    "fmt"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/exporters/jaeger"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/sdk/resource"
    "go.opentelemetry.io/otel/sdk/trace"
    "go.opentelemetry.io/otel/semconv/v1.4.0"
    oteltrace "go.opentelemetry.io/otel/trace"
)

// 链路追踪配置
type TracingConfig struct {
    ServiceName     string `json:"service_name"`
    ServiceVersion  string `json:"service_version"`
    JaegerEndpoint  string `json:"jaeger_endpoint"`
    SamplingRate    float64 `json:"sampling_rate"`
    Environment     string `json:"environment"`
}

// 初始化链路追踪
func InitTracing(config TracingConfig) (func(), error) {
    // 创建Jaeger导出器
    exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(config.JaegerEndpoint)))
    if err != nil {
        return nil, fmt.Errorf("failed to create jaeger exporter: %w", err)
    }

    // 创建资源
    res, err := resource.New(context.Background(),
        resource.WithAttributes(
            semconv.ServiceNameKey.String(config.ServiceName),
            semconv.ServiceVersionKey.String(config.ServiceVersion),
            attribute.String("environment", config.Environment),
        ),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create resource: %w", err)
    }

    // 创建追踪提供者
    tp := trace.NewTracerProvider(
        trace.WithBatcher(exp),
        trace.WithResource(res),
        trace.WithSampler(trace.TraceIDRatioBased(config.SamplingRate)),
    )

    // 设置全局追踪提供者
    otel.SetTracerProvider(tp)
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},
        propagation.Baggage{},
    ))

    // 返回清理函数
    return func() {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        tp.Shutdown(ctx)
    }, nil
}

// Gin链路追踪中间件
func TracingMiddleware(serviceName string) gin.HandlerFunc {
    tracer := otel.Tracer(serviceName)

    return func(c *gin.Context) {
        // 从请求头中提取追踪上下文
        ctx := otel.GetTextMapPropagator().Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

        // 创建span
        ctx, span := tracer.Start(ctx, fmt.Sprintf("%s %s", c.Request.Method, c.FullPath()),
            oteltrace.WithAttributes(
                attribute.String("http.method", c.Request.Method),
                attribute.String("http.url", c.Request.URL.String()),
                attribute.String("http.scheme", c.Request.URL.Scheme),
                attribute.String("http.host", c.Request.Host),
                attribute.String("http.user_agent", c.Request.UserAgent()),
            ),
        )
        defer span.End()

        // 将上下文传递给请求
        c.Request = c.Request.WithContext(ctx)

        // 处理请求
        c.Next()

        // 记录响应信息
        span.SetAttributes(
            attribute.Int("http.status_code", c.Writer.Status()),
            attribute.Int("http.response_size", c.Writer.Size()),
        )

        // 如果有错误，记录错误信息
        if len(c.Errors) > 0 {
            span.SetAttributes(attribute.String("error", c.Errors.String()))
        }
    }
}

// 数据库追踪装饰器
type TracedDB struct {
    db     *sql.DB
    tracer oteltrace.Tracer
}

func NewTracedDB(db *sql.DB, serviceName string) *TracedDB {
    return &TracedDB{
        db:     db,
        tracer: otel.Tracer(serviceName + "-db"),
    }
}

func (tdb *TracedDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
    ctx, span := tdb.tracer.Start(ctx, "db.query",
        oteltrace.WithAttributes(
            attribute.String("db.statement", query),
            attribute.String("db.operation", "query"),
        ),
    )
    defer span.End()

    rows, err := tdb.db.QueryContext(ctx, query, args...)
    if err != nil {
        span.SetAttributes(attribute.String("error", err.Error()))
    }

    return rows, err
}

func (tdb *TracedDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
    ctx, span := tdb.tracer.Start(ctx, "db.exec",
        oteltrace.WithAttributes(
            attribute.String("db.statement", query),
            attribute.String("db.operation", "exec"),
        ),
    )
    defer span.End()

    result, err := tdb.db.ExecContext(ctx, query, args...)
    if err != nil {
        span.SetAttributes(attribute.String("error", err.Error()))
    }

    return result, err
}

// HTTP客户端追踪
type TracedHTTPClient struct {
    client *http.Client
    tracer oteltrace.Tracer
}

func NewTracedHTTPClient(client *http.Client, serviceName string) *TracedHTTPClient {
    return &TracedHTTPClient{
        client: client,
        tracer: otel.Tracer(serviceName + "-http"),
    }
}

func (thc *TracedHTTPClient) Do(req *http.Request) (*http.Response, error) {
    ctx, span := thc.tracer.Start(req.Context(), fmt.Sprintf("HTTP %s", req.Method),
        oteltrace.WithAttributes(
            attribute.String("http.method", req.Method),
            attribute.String("http.url", req.URL.String()),
            attribute.String("http.scheme", req.URL.Scheme),
            attribute.String("http.host", req.URL.Host),
        ),
    )
    defer span.End()

    // 注入追踪上下文到请求头
    otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

    // 执行请求
    resp, err := thc.client.Do(req.WithContext(ctx))
    if err != nil {
        span.SetAttributes(attribute.String("error", err.Error()))
        return nil, err
    }

    // 记录响应信息
    span.SetAttributes(
        attribute.Int("http.status_code", resp.StatusCode),
        attribute.Int64("http.response_content_length", resp.ContentLength),
    )

    return resp, nil
}
```

---

## 🔧 性能优化

### Go程序性能分析

```go
// Go程序性能分析
package profiling

import (
    "context"
    "fmt"
    "net/http"
    _ "net/http/pprof"
    "runtime"
    "runtime/pprof"
    "time"
)

// 性能分析配置
type ProfilingConfig struct {
    EnablePprof     bool   `json:"enable_pprof"`
    PprofPort       string `json:"pprof_port"`
    CPUProfilePath  string `json:"cpu_profile_path"`
    MemProfilePath  string `json:"mem_profile_path"`
    ProfileDuration time.Duration `json:"profile_duration"`
}

// 性能分析服务
type ProfilingService struct {
    config ProfilingConfig
}

func NewProfilingService(config ProfilingConfig) *ProfilingService {
    return &ProfilingService{config: config}
}

// 启动pprof服务
func (ps *ProfilingService) StartPprofServer() {
    if !ps.config.EnablePprof {
        return
    }

    go func() {
        fmt.Printf("Starting pprof server on port %s\n", ps.config.PprofPort)
        if err := http.ListenAndServe(":"+ps.config.PprofPort, nil); err != nil {
            fmt.Printf("pprof server error: %v\n", err)
        }
    }()
}

// CPU性能分析
func (ps *ProfilingService) StartCPUProfile() error {
    if ps.config.CPUProfilePath == "" {
        return fmt.Errorf("CPU profile path not configured")
    }

    f, err := os.Create(ps.config.CPUProfilePath)
    if err != nil {
        return fmt.Errorf("create CPU profile file: %w", err)
    }

    if err := pprof.StartCPUProfile(f); err != nil {
        f.Close()
        return fmt.Errorf("start CPU profile: %w", err)
    }

    // 定时停止
    go func() {
        time.Sleep(ps.config.ProfileDuration)
        pprof.StopCPUProfile()
        f.Close()
        fmt.Printf("CPU profile saved to %s\n", ps.config.CPUProfilePath)
    }()

    return nil
}

// 内存性能分析
func (ps *ProfilingService) WriteMemProfile() error {
    if ps.config.MemProfilePath == "" {
        return fmt.Errorf("memory profile path not configured")
    }

    f, err := os.Create(ps.config.MemProfilePath)
    if err != nil {
        return fmt.Errorf("create memory profile file: %w", err)
    }
    defer f.Close()

    runtime.GC() // 强制垃圾回收
    if err := pprof.WriteHeapProfile(f); err != nil {
        return fmt.Errorf("write heap profile: %w", err)
    }

    fmt.Printf("Memory profile saved to %s\n", ps.config.MemProfilePath)
    return nil
}

// 性能监控指标
type PerformanceMetrics struct {
    Goroutines     int           `json:"goroutines"`
    MemoryAlloc    uint64        `json:"memory_alloc"`
    MemoryTotal    uint64        `json:"memory_total"`
    MemorySys      uint64        `json:"memory_sys"`
    GCCount        uint32        `json:"gc_count"`
    GCPauseTotal   time.Duration `json:"gc_pause_total"`
    LastGCTime     time.Time     `json:"last_gc_time"`
}

// 获取性能指标
func GetPerformanceMetrics() PerformanceMetrics {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    return PerformanceMetrics{
        Goroutines:     runtime.NumGoroutine(),
        MemoryAlloc:    m.Alloc,
        MemoryTotal:    m.TotalAlloc,
        MemorySys:      m.Sys,
        GCCount:        m.NumGC,
        GCPauseTotal:   time.Duration(m.PauseTotalNs),
        LastGCTime:     time.Unix(0, int64(m.LastGC)),
    }
}

// 性能优化建议
type OptimizationSuggestion struct {
    Category    string `json:"category"`
    Issue       string `json:"issue"`
    Suggestion  string `json:"suggestion"`
    Priority    string `json:"priority"`
}

// 分析性能并给出建议
func AnalyzePerformance(metrics PerformanceMetrics) []OptimizationSuggestion {
    var suggestions []OptimizationSuggestion

    // 检查goroutine数量
    if metrics.Goroutines > 10000 {
        suggestions = append(suggestions, OptimizationSuggestion{
            Category:   "Concurrency",
            Issue:      fmt.Sprintf("High goroutine count: %d", metrics.Goroutines),
            Suggestion: "Check for goroutine leaks, use worker pools to limit concurrency",
            Priority:   "High",
        })
    }

    // 检查内存使用
    if metrics.MemoryAlloc > 1024*1024*1024 { // 1GB
        suggestions = append(suggestions, OptimizationSuggestion{
            Category:   "Memory",
            Issue:      fmt.Sprintf("High memory usage: %d MB", metrics.MemoryAlloc/(1024*1024)),
            Suggestion: "Review memory allocations, use object pools, optimize data structures",
            Priority:   "High",
        })
    }

    // 检查GC频率
    if metrics.GCCount > 1000 {
        suggestions = append(suggestions, OptimizationSuggestion{
            Category:   "GC",
            Issue:      fmt.Sprintf("Frequent GC: %d collections", metrics.GCCount),
            Suggestion: "Reduce allocations, reuse objects, tune GOGC parameter",
            Priority:   "Medium",
        })
    }

    return suggestions
}
```

### 内存优化实践

```go
// 内存优化实践
package optimization

import (
    "sync"
    "time"
)

// 对象池优化
type ObjectPool struct {
    pool sync.Pool
}

func NewObjectPool(newFunc func() interface{}) *ObjectPool {
    return &ObjectPool{
        pool: sync.Pool{
            New: newFunc,
        },
    }
}

func (op *ObjectPool) Get() interface{} {
    return op.pool.Get()
}

func (op *ObjectPool) Put(obj interface{}) {
    op.pool.Put(obj)
}

// 字符串构建器池
var stringBuilderPool = sync.Pool{
    New: func() interface{} {
        return &strings.Builder{}
    },
}

// 优化的字符串拼接
func OptimizedStringConcat(parts []string) string {
    builder := stringBuilderPool.Get().(*strings.Builder)
    defer func() {
        builder.Reset()
        stringBuilderPool.Put(builder)
    }()

    for _, part := range parts {
        builder.WriteString(part)
    }

    return builder.String()
}

// 字节缓冲池
var byteBufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 0, 1024)
    },
}

// 优化的字节操作
func OptimizedByteOperation(data []byte) []byte {
    buffer := byteBufferPool.Get().([]byte)
    defer byteBufferPool.Put(buffer[:0])

    // 执行字节操作
    buffer = append(buffer, data...)
    // 处理逻辑...

    // 返回副本
    result := make([]byte, len(buffer))
    copy(result, buffer)
    return result
}

// 内存友好的大数据处理
type StreamProcessor struct {
    batchSize int
    processor func([]interface{}) error
}

func NewStreamProcessor(batchSize int, processor func([]interface{}) error) *StreamProcessor {
    return &StreamProcessor{
        batchSize: batchSize,
        processor: processor,
    }
}

func (sp *StreamProcessor) Process(data <-chan interface{}) error {
    batch := make([]interface{}, 0, sp.batchSize)

    for item := range data {
        batch = append(batch, item)

        if len(batch) >= sp.batchSize {
            if err := sp.processor(batch); err != nil {
                return err
            }
            batch = batch[:0] // 重置切片但保留容量
        }
    }

    // 处理剩余数据
    if len(batch) > 0 {
        return sp.processor(batch)
    }

    return nil
}

// 缓存优化
type LRUCache struct {
    capacity int
    cache    map[string]*CacheNode
    head     *CacheNode
    tail     *CacheNode
    mutex    sync.RWMutex
}

type CacheNode struct {
    key   string
    value interface{}
    prev  *CacheNode
    next  *CacheNode
    expiry time.Time
}

func NewLRUCache(capacity int) *LRUCache {
    head := &CacheNode{}
    tail := &CacheNode{}
    head.next = tail
    tail.prev = head

    return &LRUCache{
        capacity: capacity,
        cache:    make(map[string]*CacheNode),
        head:     head,
        tail:     tail,
    }
}

func (lru *LRUCache) Get(key string) (interface{}, bool) {
    lru.mutex.Lock()
    defer lru.mutex.Unlock()

    if node, exists := lru.cache[key]; exists {
        // 检查过期
        if time.Now().After(node.expiry) {
            lru.removeNode(node)
            delete(lru.cache, key)
            return nil, false
        }

        // 移动到头部
        lru.moveToHead(node)
        return node.value, true
    }

    return nil, false
}

func (lru *LRUCache) Put(key string, value interface{}, ttl time.Duration) {
    lru.mutex.Lock()
    defer lru.mutex.Unlock()

    if node, exists := lru.cache[key]; exists {
        // 更新现有节点
        node.value = value
        node.expiry = time.Now().Add(ttl)
        lru.moveToHead(node)
    } else {
        // 创建新节点
        newNode := &CacheNode{
            key:    key,
            value:  value,
            expiry: time.Now().Add(ttl),
        }

        lru.cache[key] = newNode
        lru.addToHead(newNode)

        // 检查容量
        if len(lru.cache) > lru.capacity {
            tail := lru.removeTail()
            delete(lru.cache, tail.key)
        }
    }
}

func (lru *LRUCache) moveToHead(node *CacheNode) {
    lru.removeNode(node)
    lru.addToHead(node)
}

func (lru *LRUCache) removeNode(node *CacheNode) {
    node.prev.next = node.next
    node.next.prev = node.prev
}

func (lru *LRUCache) addToHead(node *CacheNode) {
    node.prev = lru.head
    node.next = lru.head.next
    lru.head.next.prev = node
    lru.head.next = node
}

func (lru *LRUCache) removeTail() *CacheNode {
    lastNode := lru.tail.prev
    lru.removeNode(lastNode)
    return lastNode
}
```

---

## 🛡️ 安全实践

### 认证授权系统

```go
// 认证授权系统
package security

import (
    "crypto/rand"
    "crypto/sha256"
    "encoding/base64"
    "fmt"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
)

// JWT配置
type JWTConfig struct {
    SecretKey       string        `json:"secret_key"`
    AccessTokenTTL  time.Duration `json:"access_token_ttl"`
    RefreshTokenTTL time.Duration `json:"refresh_token_ttl"`
    Issuer          string        `json:"issuer"`
    Audience        string        `json:"audience"`
}

// JWT声明
type Claims struct {
    UserID   string   `json:"user_id"`
    Username string   `json:"username"`
    Roles    []string `json:"roles"`
    jwt.RegisteredClaims
}

// JWT服务
type JWTService struct {
    config JWTConfig
}

func NewJWTService(config JWTConfig) *JWTService {
    return &JWTService{config: config}
}

// 生成访问令牌
func (js *JWTService) GenerateAccessToken(userID, username string, roles []string) (string, error) {
    claims := Claims{
        UserID:   userID,
        Username: username,
        Roles:    roles,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(js.config.AccessTokenTTL)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
            Issuer:    js.config.Issuer,
            Audience:  []string{js.config.Audience},
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(js.config.SecretKey))
}

// 生成刷新令牌
func (js *JWTService) GenerateRefreshToken(userID string) (string, error) {
    claims := jwt.RegisteredClaims{
        Subject:   userID,
        ExpiresAt: jwt.NewNumericDate(time.Now().Add(js.config.RefreshTokenTTL)),
        IssuedAt:  jwt.NewNumericDate(time.Now()),
        NotBefore: jwt.NewNumericDate(time.Now()),
        Issuer:    js.config.Issuer,
        Audience:  []string{js.config.Audience},
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(js.config.SecretKey))
}

// 验证令牌
func (js *JWTService) ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(js.config.SecretKey), nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }

    return nil, fmt.Errorf("invalid token")
}

// 密码加密服务
type PasswordService struct {
    cost int
}

func NewPasswordService(cost int) *PasswordService {
    if cost < bcrypt.MinCost {
        cost = bcrypt.DefaultCost
    }
    return &PasswordService{cost: cost}
}

// 加密密码
func (ps *PasswordService) HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), ps.cost)
    return string(bytes), err
}

// 验证密码
func (ps *PasswordService) CheckPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

// 生成安全随机字符串
func GenerateSecureRandomString(length int) (string, error) {
    bytes := make([]byte, length)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// API密钥管理
type APIKeyService struct {
    keys map[string]APIKeyInfo
    mutex sync.RWMutex
}

type APIKeyInfo struct {
    UserID      string    `json:"user_id"`
    Permissions []string  `json:"permissions"`
    CreatedAt   time.Time `json:"created_at"`
    ExpiresAt   time.Time `json:"expires_at"`
    LastUsed    time.Time `json:"last_used"`
    IsActive    bool      `json:"is_active"`
}

func NewAPIKeyService() *APIKeyService {
    return &APIKeyService{
        keys: make(map[string]APIKeyInfo),
    }
}

// 生成API密钥
func (aks *APIKeyService) GenerateAPIKey(userID string, permissions []string, ttl time.Duration) (string, error) {
    // 生成密钥
    keyBytes := make([]byte, 32)
    if _, err := rand.Read(keyBytes); err != nil {
        return "", err
    }

    // 创建密钥哈希
    hash := sha256.Sum256(keyBytes)
    keyString := base64.URLEncoding.EncodeToString(hash[:])

    // 存储密钥信息
    aks.mutex.Lock()
    defer aks.mutex.Unlock()

    aks.keys[keyString] = APIKeyInfo{
        UserID:      userID,
        Permissions: permissions,
        CreatedAt:   time.Now(),
        ExpiresAt:   time.Now().Add(ttl),
        IsActive:    true,
    }

    return keyString, nil
}

// 验证API密钥
func (aks *APIKeyService) ValidateAPIKey(key string) (*APIKeyInfo, error) {
    aks.mutex.RLock()
    defer aks.mutex.RUnlock()

    keyInfo, exists := aks.keys[key]
    if !exists {
        return nil, fmt.Errorf("invalid API key")
    }

    if !keyInfo.IsActive {
        return nil, fmt.Errorf("API key is inactive")
    }

    if time.Now().After(keyInfo.ExpiresAt) {
        return nil, fmt.Errorf("API key has expired")
    }

    // 更新最后使用时间
    keyInfo.LastUsed = time.Now()
    aks.keys[key] = keyInfo

    return &keyInfo, nil
}

// 权限检查
func (aks *APIKeyService) CheckPermission(key, permission string) bool {
    keyInfo, err := aks.ValidateAPIKey(key)
    if err != nil {
        return false
    }

    for _, perm := range keyInfo.Permissions {
        if perm == permission || perm == "*" {
            return true
        }
    }

    return false
}
```

---

## 🔄 CI/CD流水线

### GitHub Actions配置

```yaml
# .github/workflows/ci-cd.yml
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

env:
  GO_VERSION: '1.21'
  DOCKER_REGISTRY: 'your-registry.com'
  IMAGE_NAME: 'mall-go'

jobs:
  # 代码质量检查
  lint:
    name: Code Linting
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ env.GO_VERSION }}

    - name: golangci-lint
      uses: golangci/golangci-lint-action@v3
      with:
        version: latest
        args: --timeout=5m

    - name: Go vet
      run: go vet ./...

    - name: Go fmt
      run: |
        if [ "$(gofmt -s -l . | wc -l)" -gt 0 ]; then
          echo "Code is not formatted properly"
          gofmt -s -l .
          exit 1
        fi

  # 安全扫描
  security:
    name: Security Scan
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ env.GO_VERSION }}

    - name: Run Gosec Security Scanner
      uses: securecodewarrior/github-action-gosec@master
      with:
        args: '-fmt sarif -out gosec.sarif ./...'

    - name: Upload SARIF file
      uses: github/codeql-action/upload-sarif@v2
      with:
        sarif_file: gosec.sarif

  # 单元测试
  test:
    name: Unit Tests
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: test_db
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

      redis:
        image: redis:7
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 6379:6379

    steps:
    - uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ env.GO_VERSION }}

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
      run: |
        go test -v -race -coverprofile=coverage.out ./...
        go tool cover -html=coverage.out -o coverage.html
      env:
        DB_HOST: localhost
        DB_PORT: 5432
        DB_USER: postgres
        DB_PASSWORD: postgres
        DB_NAME: test_db
        REDIS_HOST: localhost
        REDIS_PORT: 6379

    - name: Upload coverage reports
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
        flags: unittests
        name: codecov-umbrella

  # 构建和推送Docker镜像
  build:
    name: Build and Push
    runs-on: ubuntu-latest
    needs: [lint, security, test]
    if: github.ref == 'refs/heads/main'

    steps:
    - uses: actions/checkout@v4

    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v3

    - name: Login to Container Registry
      uses: docker/login-action@v3
      with:
        registry: ${{ env.DOCKER_REGISTRY }}
        username: ${{ secrets.DOCKER_USERNAME }}
        password: ${{ secrets.DOCKER_PASSWORD }}

    - name: Extract metadata
      id: meta
      uses: docker/metadata-action@v5
      with:
        images: ${{ env.DOCKER_REGISTRY }}/${{ env.IMAGE_NAME }}
        tags: |
          type=ref,event=branch
          type=ref,event=pr
          type=sha,prefix={{branch}}-
          type=raw,value=latest,enable={{is_default_branch}}

    - name: Build and push
      uses: docker/build-push-action@v5
      with:
        context: .
        platforms: linux/amd64,linux/arm64
        push: true
        tags: ${{ steps.meta.outputs.tags }}
        labels: ${{ steps.meta.outputs.labels }}
        cache-from: type=gha
        cache-to: type=gha,mode=max

  # 部署到测试环境
  deploy-staging:
    name: Deploy to Staging
    runs-on: ubuntu-latest
    needs: [build]
    if: github.ref == 'refs/heads/develop'
    environment: staging

    steps:
    - uses: actions/checkout@v4

    - name: Deploy to Kubernetes
      uses: azure/k8s-deploy@v1
      with:
        manifests: |
          k8s/staging/deployment.yaml
          k8s/staging/service.yaml
          k8s/staging/ingress.yaml
        images: |
          ${{ env.DOCKER_REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.sha }}
        kubectl-version: 'latest'

  # 部署到生产环境
  deploy-production:
    name: Deploy to Production
    runs-on: ubuntu-latest
    needs: [build]
    if: github.ref == 'refs/heads/main'
    environment: production

    steps:
    - uses: actions/checkout@v4

    - name: Deploy to Kubernetes
      uses: azure/k8s-deploy@v1
      with:
        manifests: |
          k8s/production/deployment.yaml
          k8s/production/service.yaml
          k8s/production/ingress.yaml
        images: |
          ${{ env.DOCKER_REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.sha }}
        kubectl-version: 'latest'

    - name: Run smoke tests
      run: |
        # 等待部署完成
        sleep 60
        # 执行冒烟测试
        curl -f https://api.mall-go.com/health || exit 1
        curl -f https://api.mall-go.com/metrics || exit 1
```

### 部署脚本

```go
// 部署脚本生成器
package deployment

import (
    "fmt"
    "os"
    "text/template"
)

// 部署配置
type DeploymentScript struct {
    AppName        string `json:"app_name"`
    Environment    string `json:"environment"`
    DockerImage    string `json:"docker_image"`
    Port           int    `json:"port"`
    HealthCheckURL string `json:"health_check_url"`
    BackupCount    int    `json:"backup_count"`
}

// 生成部署脚本
func GenerateDeployScript(config DeploymentScript) string {
    scriptTemplate := `#!/bin/bash

# 部署脚本 - {{.AppName}} ({{.Environment}})
set -e

APP_NAME="{{.AppName}}"
ENVIRONMENT="{{.Environment}}"
DOCKER_IMAGE="{{.DockerImage}}"
PORT={{.Port}}
HEALTH_CHECK_URL="{{.HealthCheckURL}}"
BACKUP_COUNT={{.BackupCount}}

echo "开始部署 $APP_NAME 到 $ENVIRONMENT 环境..."

# 1. 拉取最新镜像
echo "拉取Docker镜像..."
docker pull $DOCKER_IMAGE

# 2. 备份当前版本
echo "备份当前版本..."
if docker ps -q -f name=$APP_NAME > /dev/null; then
    BACKUP_NAME="${APP_NAME}_backup_$(date +%Y%m%d_%H%M%S)"
    docker commit $APP_NAME $BACKUP_NAME
    echo "备份完成: $BACKUP_NAME"

    # 清理旧备份
    OLD_BACKUPS=$(docker images --format "table {{.Repository}}" | grep "${APP_NAME}_backup" | tail -n +$((BACKUP_COUNT + 1)))
    if [ ! -z "$OLD_BACKUPS" ]; then
        echo "清理旧备份..."
        echo "$OLD_BACKUPS" | xargs -r docker rmi
    fi
fi

# 3. 停止旧容器
echo "停止旧容器..."
if docker ps -q -f name=$APP_NAME > /dev/null; then
    docker stop $APP_NAME
    docker rm $APP_NAME
fi

# 4. 启动新容器
echo "启动新容器..."
docker run -d \
    --name $APP_NAME \
    --restart unless-stopped \
    -p $PORT:$PORT \
    -e ENVIRONMENT=$ENVIRONMENT \
    --network mall-go-network \
    $DOCKER_IMAGE

# 5. 健康检查
echo "执行健康检查..."
RETRY_COUNT=0
MAX_RETRIES=30

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -f $HEALTH_CHECK_URL > /dev/null 2>&1; then
        echo "健康检查通过!"
        break
    fi

    echo "等待应用启动... ($((RETRY_COUNT + 1))/$MAX_RETRIES)"
    sleep 10
    RETRY_COUNT=$((RETRY_COUNT + 1))
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo "健康检查失败，开始回滚..."

    # 停止失败的容器
    docker stop $APP_NAME
    docker rm $APP_NAME

    # 恢复备份
    if [ ! -z "$BACKUP_NAME" ]; then
        docker run -d \
            --name $APP_NAME \
            --restart unless-stopped \
            -p $PORT:$PORT \
            -e ENVIRONMENT=$ENVIRONMENT \
            --network mall-go-network \
            $BACKUP_NAME
        echo "回滚完成"
    fi

    exit 1
fi

# 6. 清理无用镜像
echo "清理无用镜像..."
docker image prune -f

echo "部署完成!"
echo "应用状态: $(docker ps --format 'table {{.Names}}\t{{.Status}}' | grep $APP_NAME)"
echo "访问地址: $HEALTH_CHECK_URL"
`

    tmpl, _ := template.New("deploy").Parse(scriptTemplate)
    var result strings.Builder
    tmpl.Execute(&result, config)
    return result.String()
}

// 生成Kubernetes部署脚本
func GenerateK8sDeployScript(namespace, appName, image string) string {
    return fmt.Sprintf(`#!/bin/bash

# Kubernetes部署脚本
set -e

NAMESPACE="%s"
APP_NAME="%s"
IMAGE="%s"

echo "开始部署到Kubernetes..."

# 1. 创建命名空间（如果不存在）
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

# 2. 应用配置
echo "应用ConfigMap和Secret..."
kubectl apply -f k8s/configmap.yaml -n $NAMESPACE
kubectl apply -f k8s/secret.yaml -n $NAMESPACE

# 3. 更新部署
echo "更新Deployment..."
kubectl set image deployment/$APP_NAME $APP_NAME=$IMAGE -n $NAMESPACE

# 4. 等待部署完成
echo "等待部署完成..."
kubectl rollout status deployment/$APP_NAME -n $NAMESPACE --timeout=300s

# 5. 验证部署
echo "验证部署..."
kubectl get pods -n $NAMESPACE -l app=$APP_NAME

# 6. 检查服务状态
echo "检查服务状态..."
kubectl get svc -n $NAMESPACE -l app=$APP_NAME

echo "部署完成!"
`, namespace, appName, image)
}
```

---

## 🎯 面试常考点

### 1. 生产环境部署和运维

**问题：** 如何保证Go应用在生产环境的稳定性和可靠性？

**答案：**
```go
/*
生产环境稳定性保证措施：

1. 容器化部署
   - 使用Docker容器化，保证环境一致性
   - 多阶段构建，减小镜像体积
   - 健康检查和自动重启
   - 资源限制和配额管理

2. 监控和可观测性
   - 指标监控：CPU、内存、网络、业务指标
   - 日志聚合：结构化日志、集中收集
   - 链路追踪：分布式请求追踪
   - 告警系统：及时发现和响应问题

3. 高可用架构
   - 负载均衡：多实例部署
   - 故障转移：自动切换和恢复
   - 数据备份：定期备份和恢复测试
   - 灾难恢复：跨区域部署

4. 性能优化
   - 内存管理：对象池、内存复用
   - 并发控制：goroutine池、限流
   - 缓存策略：多级缓存、缓存预热
   - 数据库优化：连接池、查询优化

5. 安全防护
   - 认证授权：JWT、RBAC
   - 数据加密：传输加密、存储加密
   - 安全扫描：代码扫描、依赖检查
   - 访问控制：网络隔离、防火墙
*/

// 生产环境配置管理
type ProductionConfig struct {
    // 应用配置
    App struct {
        Name        string `json:"name"`
        Version     string `json:"version"`
        Environment string `json:"environment"`
        LogLevel    string `json:"log_level"`
    } `json:"app"`

    // 服务器配置
    Server struct {
        Host            string        `json:"host"`
        Port            int           `json:"port"`
        ReadTimeout     time.Duration `json:"read_timeout"`
        WriteTimeout    time.Duration `json:"write_timeout"`
        IdleTimeout     time.Duration `json:"idle_timeout"`
        MaxHeaderBytes  int           `json:"max_header_bytes"`
        GracefulTimeout time.Duration `json:"graceful_timeout"`
    } `json:"server"`

    // 数据库配置
    Database struct {
        Host            string        `json:"host"`
        Port            int           `json:"port"`
        Username        string        `json:"username"`
        Password        string        `json:"password"`
        Database        string        `json:"database"`
        MaxOpenConns    int           `json:"max_open_conns"`
        MaxIdleConns    int           `json:"max_idle_conns"`
        ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
        SSLMode         string        `json:"ssl_mode"`
    } `json:"database"`

    // Redis配置
    Redis struct {
        Host         string        `json:"host"`
        Port         int           `json:"port"`
        Password     string        `json:"password"`
        DB           int           `json:"db"`
        PoolSize     int           `json:"pool_size"`
        MinIdleConns int           `json:"min_idle_conns"`
        DialTimeout  time.Duration `json:"dial_timeout"`
        ReadTimeout  time.Duration `json:"read_timeout"`
        WriteTimeout time.Duration `json:"write_timeout"`
    } `json:"redis"`

    // 监控配置
    Monitoring struct {
        EnableMetrics bool   `json:"enable_metrics"`
        MetricsPort   string `json:"metrics_port"`
        EnableTracing bool   `json:"enable_tracing"`
        JaegerURL     string `json:"jaeger_url"`
        SampleRate    float64 `json:"sample_rate"`
    } `json:"monitoring"`
}

// 优雅关闭
func GracefulShutdown(server *http.Server, timeout time.Duration) {
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    <-quit
    log.Println("Shutting down server...")

    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }

    log.Println("Server exiting")
}
```

### 2. 性能优化和故障排查

**问题：** Go应用出现性能问题时，如何进行排查和优化？

**答案：**
```go
/*
Go应用性能排查步骤：

1. 性能分析工具
   - pprof：CPU、内存、goroutine分析
   - trace：程序执行轨迹分析
   - benchmark：基准测试
   - 监控指标：Prometheus + Grafana

2. 常见性能问题
   - 内存泄漏：goroutine泄漏、内存未释放
   - CPU占用高：算法复杂度、无限循环
   - GC压力大：频繁分配、大对象
   - 并发问题：锁竞争、channel阻塞

3. 优化策略
   - 内存优化：对象池、内存复用
   - 并发优化：worker pool、限流
   - 算法优化：数据结构、缓存
   - I/O优化：连接池、批处理
*/

// 性能监控和分析
type PerformanceAnalyzer struct {
    metrics map[string]interface{}
    mutex   sync.RWMutex
}

func NewPerformanceAnalyzer() *PerformanceAnalyzer {
    return &PerformanceAnalyzer{
        metrics: make(map[string]interface{}),
    }
}

// 记录性能指标
func (pa *PerformanceAnalyzer) RecordMetric(name string, value interface{}) {
    pa.mutex.Lock()
    defer pa.mutex.Unlock()
    pa.metrics[name] = value
}

// 获取性能报告
func (pa *PerformanceAnalyzer) GetReport() map[string]interface{} {
    pa.mutex.RLock()
    defer pa.mutex.RUnlock()

    report := make(map[string]interface{})
    for k, v := range pa.metrics {
        report[k] = v
    }

    // 添加系统指标
    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    report["goroutines"] = runtime.NumGoroutine()
    report["memory_alloc"] = m.Alloc
    report["memory_total"] = m.TotalAlloc
    report["memory_sys"] = m.Sys
    report["gc_count"] = m.NumGC
    report["gc_pause_total"] = time.Duration(m.PauseTotalNs)

    return report
}

// 性能测试辅助函数
func BenchmarkFunction(name string, fn func()) time.Duration {
    start := time.Now()
    fn()
    duration := time.Since(start)

    fmt.Printf("Benchmark %s: %v\n", name, duration)
    return duration
}

// 内存使用分析
func AnalyzeMemoryUsage() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    fmt.Printf("Memory Usage Analysis:\n")
    fmt.Printf("  Alloc: %d KB\n", m.Alloc/1024)
    fmt.Printf("  TotalAlloc: %d KB\n", m.TotalAlloc/1024)
    fmt.Printf("  Sys: %d KB\n", m.Sys/1024)
    fmt.Printf("  NumGC: %d\n", m.NumGC)
    fmt.Printf("  GCCPUFraction: %f\n", m.GCCPUFraction)
    fmt.Printf("  Goroutines: %d\n", runtime.NumGoroutine())
}
```

---

## ⚠️ 踩坑提醒

### 1. 容器化部署陷阱

```go
// ❌ 错误：不安全的Docker配置
/*
常见问题：
1. 使用root用户运行应用
2. 暴露不必要的端口
3. 没有设置资源限制
4. 忽略健康检查
5. 硬编码配置信息
*/

// 错误的Dockerfile示例
dockerfile_bad := `
FROM golang:1.21
WORKDIR /app
COPY . .
RUN go build -o app .
EXPOSE 8080
CMD ["./app"]
`

// ✅ 正确：安全的Docker配置
dockerfile_good := `
# 多阶段构建
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

# 运行阶段
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 复制二进制文件
COPY --from=builder /app/app .
RUN chown appuser:appgroup ./app

# 切换到非root用户
USER appuser

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

EXPOSE 8080
CMD ["./app"]
`

/*
最佳实践：
1. 使用多阶段构建减小镜像体积
2. 创建非root用户运行应用
3. 添加健康检查
4. 设置合适的资源限制
5. 使用环境变量配置
*/
```

### 2. 监控配置陷阱

```go
// ❌ 错误：监控配置不当
type BadMonitoringConfig struct {
    // 监控所有请求，导致性能问题
    TraceAllRequests bool

    // 指标过多，存储压力大
    DetailedMetrics map[string]interface{}

    // 没有采样，链路追踪开销大
    TracingSampleRate float64 // 1.0 = 100%

    // 日志级别过低，日志量巨大
    LogLevel string // "debug"
}

// ✅ 正确：合理的监控配置
type GoodMonitoringConfig struct {
    // 核心指标监控
    CoreMetrics struct {
        RequestCount    bool `json:"request_count"`
        ResponseTime    bool `json:"response_time"`
        ErrorRate       bool `json:"error_rate"`
        SystemMetrics   bool `json:"system_metrics"`
    } `json:"core_metrics"`

    // 合理的采样率
    TracingSampleRate float64 `json:"tracing_sample_rate"` // 0.1 = 10%

    // 分级日志
    LogConfig struct {
        Level      string `json:"level"`      // "info"
        MaxSize    int    `json:"max_size"`   // MB
        MaxBackups int    `json:"max_backups"`
        MaxAge     int    `json:"max_age"`    // days
        Compress   bool   `json:"compress"`
    } `json:"log_config"`

    // 告警阈值
    AlertThresholds struct {
        CPUUsage     float64 `json:"cpu_usage"`     // 80%
        MemoryUsage  float64 `json:"memory_usage"`  // 85%
        ErrorRate    float64 `json:"error_rate"`    // 5%
        ResponseTime float64 `json:"response_time"` // 1000ms
    } `json:"alert_thresholds"`
}

/*
监控最佳实践：
1. 监控核心指标，避免指标爆炸
2. 合理设置采样率，平衡性能和可观测性
3. 分级日志，控制日志量
4. 设置合理的告警阈值
5. 定期清理历史数据
*/
```

### 3. 性能优化陷阱

```go
// ❌ 错误：过度优化
func BadOptimization() {
    // 过早优化，代码复杂度增加
    type ComplexCache struct {
        l1Cache map[string]interface{}
        l2Cache map[string]interface{}
        l3Cache map[string]interface{}
        // ... 多级缓存，但实际不需要
    }

    // 过度使用goroutine
    for i := 0; i < 1000000; i++ {
        go func(i int) {
            // 简单操作，不需要goroutine
            fmt.Println(i)
        }(i)
    }

    // 不必要的内存池
    var stringPool = sync.Pool{
        New: func() interface{} {
            return make([]string, 0, 10)
        },
    }
    // 用于简单的字符串操作，开销大于收益
}

// ✅ 正确：合理优化
func GoodOptimization() {
    // 1. 先测量，再优化
    start := time.Now()
    defer func() {
        fmt.Printf("Operation took: %v\n", time.Since(start))
    }()

    // 2. 使用worker pool控制并发
    const workerCount = 10
    jobs := make(chan int, 100)
    results := make(chan int, 100)

    // 启动workers
    for w := 0; w < workerCount; w++ {
        go worker(jobs, results)
    }

    // 发送任务
    for i := 0; i < 100; i++ {
        jobs <- i
    }
    close(jobs)

    // 收集结果
    for i := 0; i < 100; i++ {
        <-results
    }

    // 3. 合理使用缓存
    cache := NewLRUCache(1000) // 适当的缓存大小

    // 4. 批量处理
    batchSize := 100
    batch := make([]Item, 0, batchSize)

    for item := range items {
        batch = append(batch, item)
        if len(batch) >= batchSize {
            processBatch(batch)
            batch = batch[:0] // 重置但保留容量
        }
    }

    // 处理剩余项目
    if len(batch) > 0 {
        processBatch(batch)
    }
}

func worker(jobs <-chan int, results chan<- int) {
    for job := range jobs {
        // 处理任务
        result := processJob(job)
        results <- result
    }
}

/*
性能优化原则：
1. 先测量，后优化
2. 关注热点代码
3. 避免过早优化
4. 平衡复杂度和性能
5. 持续监控和调整
*/
```

---

## 📝 练习题

### 练习题1：设计完整的监控系统（⭐⭐⭐）

**题目描述：**
为mall-go项目设计一个完整的监控系统，包括指标收集、链路追踪、日志聚合、告警系统等。

```go
// 练习题1：完整监控系统设计
package main

import (
    "context"
    "fmt"
    "sync"
    "time"
)

// 解答：
// 1. 监控系统架构
type MonitoringSystem struct {
    MetricsCollector  *MetricsCollector  `json:"metrics_collector"`
    TracingSystem     *TracingSystem     `json:"tracing_system"`
    LoggingSystem     *LoggingSystem     `json:"logging_system"`
    AlertManager      *AlertManager      `json:"alert_manager"`
    Dashboard         *Dashboard         `json:"dashboard"`
}

// 指标收集器
type MetricsCollector struct {
    metrics map[string]Metric
    mutex   sync.RWMutex
}

type Metric struct {
    Name      string                 `json:"name"`
    Type      string                 `json:"type"` // counter, gauge, histogram
    Value     float64                `json:"value"`
    Labels    map[string]string      `json:"labels"`
    Timestamp time.Time              `json:"timestamp"`
}

func NewMetricsCollector() *MetricsCollector {
    return &MetricsCollector{
        metrics: make(map[string]Metric),
    }
}

// 记录指标
func (mc *MetricsCollector) RecordCounter(name string, value float64, labels map[string]string) {
    mc.mutex.Lock()
    defer mc.mutex.Unlock()

    key := fmt.Sprintf("%s_%s", name, labelsToString(labels))
    metric := mc.metrics[key]
    metric.Name = name
    metric.Type = "counter"
    metric.Value += value
    metric.Labels = labels
    metric.Timestamp = time.Now()
    mc.metrics[key] = metric
}

func (mc *MetricsCollector) RecordGauge(name string, value float64, labels map[string]string) {
    mc.mutex.Lock()
    defer mc.mutex.Unlock()

    key := fmt.Sprintf("%s_%s", name, labelsToString(labels))
    mc.metrics[key] = Metric{
        Name:      name,
        Type:      "gauge",
        Value:     value,
        Labels:    labels,
        Timestamp: time.Now(),
    }
}

// 获取所有指标
func (mc *MetricsCollector) GetMetrics() []Metric {
    mc.mutex.RLock()
    defer mc.mutex.RUnlock()

    metrics := make([]Metric, 0, len(mc.metrics))
    for _, metric := range mc.metrics {
        metrics = append(metrics, metric)
    }
    return metrics
}

// 2. 链路追踪系统
type TracingSystem struct {
    spans map[string]*Span
    mutex sync.RWMutex
}

type Span struct {
    TraceID    string            `json:"trace_id"`
    SpanID     string            `json:"span_id"`
    ParentID   string            `json:"parent_id"`
    Operation  string            `json:"operation"`
    StartTime  time.Time         `json:"start_time"`
    EndTime    time.Time         `json:"end_time"`
    Duration   time.Duration     `json:"duration"`
    Tags       map[string]string `json:"tags"`
    Logs       []LogEntry        `json:"logs"`
}

type LogEntry struct {
    Timestamp time.Time         `json:"timestamp"`
    Fields    map[string]string `json:"fields"`
}

func NewTracingSystem() *TracingSystem {
    return &TracingSystem{
        spans: make(map[string]*Span),
    }
}

// 开始span
func (ts *TracingSystem) StartSpan(traceID, spanID, parentID, operation string) *Span {
    span := &Span{
        TraceID:   traceID,
        SpanID:    spanID,
        ParentID:  parentID,
        Operation: operation,
        StartTime: time.Now(),
        Tags:      make(map[string]string),
        Logs:      make([]LogEntry, 0),
    }

    ts.mutex.Lock()
    ts.spans[spanID] = span
    ts.mutex.Unlock()

    return span
}

// 结束span
func (ts *TracingSystem) FinishSpan(spanID string) {
    ts.mutex.Lock()
    defer ts.mutex.Unlock()

    if span, exists := ts.spans[spanID]; exists {
        span.EndTime = time.Now()
        span.Duration = span.EndTime.Sub(span.StartTime)
    }
}

// 3. 日志系统
type LoggingSystem struct {
    logs   []LogMessage
    mutex  sync.RWMutex
    config LogConfig
}

type LogMessage struct {
    Level     string                 `json:"level"`
    Message   string                 `json:"message"`
    Timestamp time.Time              `json:"timestamp"`
    Fields    map[string]interface{} `json:"fields"`
    TraceID   string                 `json:"trace_id"`
    SpanID    string                 `json:"span_id"`
}

type LogConfig struct {
    Level      string `json:"level"`
    MaxSize    int    `json:"max_size"`
    MaxBackups int    `json:"max_backups"`
    MaxAge     int    `json:"max_age"`
}

func NewLoggingSystem(config LogConfig) *LoggingSystem {
    return &LoggingSystem{
        logs:   make([]LogMessage, 0),
        config: config,
    }
}

// 记录日志
func (ls *LoggingSystem) Log(level, message string, fields map[string]interface{}, traceID, spanID string) {
    if !ls.shouldLog(level) {
        return
    }

    logMsg := LogMessage{
        Level:     level,
        Message:   message,
        Timestamp: time.Now(),
        Fields:    fields,
        TraceID:   traceID,
        SpanID:    spanID,
    }

    ls.mutex.Lock()
    ls.logs = append(ls.logs, logMsg)

    // 限制日志数量
    if len(ls.logs) > ls.config.MaxSize {
        ls.logs = ls.logs[len(ls.logs)-ls.config.MaxSize:]
    }
    ls.mutex.Unlock()
}

func (ls *LoggingSystem) shouldLog(level string) bool {
    levels := map[string]int{
        "debug": 0,
        "info":  1,
        "warn":  2,
        "error": 3,
    }

    configLevel := levels[ls.config.Level]
    msgLevel := levels[level]

    return msgLevel >= configLevel
}

// 4. 告警管理器
type AlertManager struct {
    rules   []AlertRule
    alerts  []Alert
    mutex   sync.RWMutex
}

type AlertRule struct {
    Name        string  `json:"name"`
    Metric      string  `json:"metric"`
    Condition   string  `json:"condition"` // >, <, ==
    Threshold   float64 `json:"threshold"`
    Duration    time.Duration `json:"duration"`
    Severity    string  `json:"severity"`
    Actions     []AlertAction `json:"actions"`
}

type Alert struct {
    Rule        AlertRule `json:"rule"`
    Value       float64   `json:"value"`
    Status      string    `json:"status"` // firing, resolved
    StartTime   time.Time `json:"start_time"`
    ResolveTime time.Time `json:"resolve_time"`
}

type AlertAction struct {
    Type   string                 `json:"type"` // email, webhook, slack
    Config map[string]interface{} `json:"config"`
}

func NewAlertManager() *AlertManager {
    return &AlertManager{
        rules:  make([]AlertRule, 0),
        alerts: make([]Alert, 0),
    }
}

// 添加告警规则
func (am *AlertManager) AddRule(rule AlertRule) {
    am.mutex.Lock()
    defer am.mutex.Unlock()
    am.rules = append(am.rules, rule)
}

// 检查告警
func (am *AlertManager) CheckAlerts(metrics []Metric) {
    am.mutex.Lock()
    defer am.mutex.Unlock()

    for _, rule := range am.rules {
        for _, metric := range metrics {
            if metric.Name == rule.Metric {
                if am.evaluateCondition(metric.Value, rule.Condition, rule.Threshold) {
                    am.fireAlert(rule, metric.Value)
                }
            }
        }
    }
}

func (am *AlertManager) evaluateCondition(value float64, condition string, threshold float64) bool {
    switch condition {
    case ">":
        return value > threshold
    case "<":
        return value < threshold
    case "==":
        return value == threshold
    default:
        return false
    }
}

func (am *AlertManager) fireAlert(rule AlertRule, value float64) {
    alert := Alert{
        Rule:      rule,
        Value:     value,
        Status:    "firing",
        StartTime: time.Now(),
    }

    am.alerts = append(am.alerts, alert)

    // 执行告警动作
    for _, action := range rule.Actions {
        go am.executeAction(action, alert)
    }
}

func (am *AlertManager) executeAction(action AlertAction, alert Alert) {
    switch action.Type {
    case "email":
        fmt.Printf("Sending email alert: %s\n", alert.Rule.Name)
    case "webhook":
        fmt.Printf("Sending webhook alert: %s\n", alert.Rule.Name)
    case "slack":
        fmt.Printf("Sending slack alert: %s\n", alert.Rule.Name)
    }
}

// 5. 仪表板
type Dashboard struct {
    widgets []Widget
}

type Widget struct {
    Type   string                 `json:"type"` // chart, table, gauge
    Title  string                 `json:"title"`
    Query  string                 `json:"query"`
    Config map[string]interface{} `json:"config"`
}

// 使用示例
func ExampleMonitoringSystem() {
    // 创建监控系统
    monitoring := &MonitoringSystem{
        MetricsCollector: NewMetricsCollector(),
        TracingSystem:    NewTracingSystem(),
        LoggingSystem:    NewLoggingSystem(LogConfig{Level: "info", MaxSize: 10000}),
        AlertManager:     NewAlertManager(),
    }

    // 添加告警规则
    monitoring.AlertManager.AddRule(AlertRule{
        Name:      "High CPU Usage",
        Metric:    "cpu_usage",
        Condition: ">",
        Threshold: 80.0,
        Duration:  5 * time.Minute,
        Severity:  "warning",
        Actions: []AlertAction{
            {Type: "email", Config: map[string]interface{}{"to": "admin@example.com"}},
        },
    })

    // 模拟指标收集
    monitoring.MetricsCollector.RecordGauge("cpu_usage", 85.0, map[string]string{"host": "server1"})
    monitoring.MetricsCollector.RecordCounter("http_requests_total", 1, map[string]string{"method": "GET", "status": "200"})

    // 模拟链路追踪
    span := monitoring.TracingSystem.StartSpan("trace1", "span1", "", "http_request")
    time.Sleep(100 * time.Millisecond)
    monitoring.TracingSystem.FinishSpan("span1")

    // 模拟日志记录
    monitoring.LoggingSystem.Log("info", "Request processed", map[string]interface{}{
        "method": "GET",
        "path":   "/api/users",
        "status": 200,
    }, "trace1", "span1")

    // 检查告警
    metrics := monitoring.MetricsCollector.GetMetrics()
    monitoring.AlertManager.CheckAlerts(metrics)
}

// 辅助函数
func labelsToString(labels map[string]string) string {
    var result string
    for k, v := range labels {
        result += fmt.Sprintf("%s=%s,", k, v)
    }
    return result
}

/*
监控系统设计要点：
1. 多维度监控：指标、链路、日志、告警
2. 高性能：异步处理、批量操作
3. 可扩展：插件化架构、配置驱动
4. 易用性：统一接口、可视化展示
5. 可靠性：容错处理、数据持久化

扩展思考：
- 如何实现监控数据的持久化存储？
- 如何处理大量监控数据的性能问题？
- 如何实现分布式监控数据的聚合？
- 如何设计灵活的告警规则引擎？
*/
```

### 练习题2：设计高可用部署架构（⭐⭐⭐⭐）

**题目描述：**
为mall-go项目设计一个高可用的部署架构，包括负载均衡、故障转移、数据备份、灾难恢复等。

```go
// 练习题2：高可用部署架构设计
package main

import (
    "context"
    "fmt"
    "sync"
    "time"
)

// 解答：
// 1. 高可用架构组件
type HighAvailabilityArchitecture struct {
    LoadBalancer    *LoadBalancer    `json:"load_balancer"`
    ServiceRegistry *ServiceRegistry `json:"service_registry"`
    HealthChecker   *HealthChecker   `json:"health_checker"`
    BackupManager   *BackupManager   `json:"backup_manager"`
    FailoverManager *FailoverManager `json:"failover_manager"`
    CircuitBreaker  *CircuitBreaker  `json:"circuit_breaker"`
}

// 负载均衡器
type LoadBalancer struct {
    algorithm string // round_robin, weighted, least_connections
    backends  []Backend
    mutex     sync.RWMutex
}

type Backend struct {
    ID       string    `json:"id"`
    Address  string    `json:"address"`
    Weight   int       `json:"weight"`
    Healthy  bool      `json:"healthy"`
    LastSeen time.Time `json:"last_seen"`
    Metrics  BackendMetrics `json:"metrics"`
}

type BackendMetrics struct {
    ActiveConnections int           `json:"active_connections"`
    ResponseTime      time.Duration `json:"response_time"`
    ErrorRate         float64       `json:"error_rate"`
    RequestCount      int64         `json:"request_count"`
}

func NewLoadBalancer(algorithm string) *LoadBalancer {
    return &LoadBalancer{
        algorithm: algorithm,
        backends:  make([]Backend, 0),
    }
}

// 添加后端服务
func (lb *LoadBalancer) AddBackend(backend Backend) {
    lb.mutex.Lock()
    defer lb.mutex.Unlock()
    lb.backends = append(lb.backends, backend)
}

// 选择后端服务
func (lb *LoadBalancer) SelectBackend() (*Backend, error) {
    lb.mutex.RLock()
    defer lb.mutex.RUnlock()

    healthyBackends := make([]Backend, 0)
    for _, backend := range lb.backends {
        if backend.Healthy {
            healthyBackends = append(healthyBackends, backend)
        }
    }

    if len(healthyBackends) == 0 {
        return nil, fmt.Errorf("no healthy backends available")
    }

    switch lb.algorithm {
    case "round_robin":
        return lb.roundRobin(healthyBackends), nil
    case "weighted":
        return lb.weighted(healthyBackends), nil
    case "least_connections":
        return lb.leastConnections(healthyBackends), nil
    default:
        return &healthyBackends[0], nil
    }
}

func (lb *LoadBalancer) roundRobin(backends []Backend) *Backend {
    // 简化实现，实际需要维护计数器
    return &backends[time.Now().Unix()%int64(len(backends))]
}

func (lb *LoadBalancer) weighted(backends []Backend) *Backend {
    totalWeight := 0
    for _, backend := range backends {
        totalWeight += backend.Weight
    }

    if totalWeight == 0 {
        return &backends[0]
    }

    random := int(time.Now().Unix()) % totalWeight
    currentWeight := 0

    for _, backend := range backends {
        currentWeight += backend.Weight
        if random < currentWeight {
            return &backend
        }
    }

    return &backends[0]
}

func (lb *LoadBalancer) leastConnections(backends []Backend) *Backend {
    minConnections := backends[0].Metrics.ActiveConnections
    selectedBackend := &backends[0]

    for _, backend := range backends {
        if backend.Metrics.ActiveConnections < minConnections {
            minConnections = backend.Metrics.ActiveConnections
            selectedBackend = &backend
        }
    }

    return selectedBackend
}

// 2. 服务注册与发现
type ServiceRegistry struct {
    services map[string][]ServiceInstance
    mutex    sync.RWMutex
}

type ServiceInstance struct {
    ID       string            `json:"id"`
    Name     string            `json:"name"`
    Address  string            `json:"address"`
    Port     int               `json:"port"`
    Tags     []string          `json:"tags"`
    Metadata map[string]string `json:"metadata"`
    Health   HealthStatus      `json:"health"`
    TTL      time.Duration     `json:"ttl"`
    LastHeartbeat time.Time    `json:"last_heartbeat"`
}

type HealthStatus struct {
    Status      string    `json:"status"` // healthy, unhealthy, unknown
    LastCheck   time.Time `json:"last_check"`
    CheckCount  int       `json:"check_count"`
    FailureCount int      `json:"failure_count"`
}

func NewServiceRegistry() *ServiceRegistry {
    return &ServiceRegistry{
        services: make(map[string][]ServiceInstance),
    }
}

// 注册服务
func (sr *ServiceRegistry) RegisterService(instance ServiceInstance) error {
    sr.mutex.Lock()
    defer sr.mutex.Unlock()

    instance.LastHeartbeat = time.Now()
    instance.Health.Status = "healthy"
    instance.Health.LastCheck = time.Now()

    if _, exists := sr.services[instance.Name]; !exists {
        sr.services[instance.Name] = make([]ServiceInstance, 0)
    }

    // 检查是否已存在
    for i, existing := range sr.services[instance.Name] {
        if existing.ID == instance.ID {
            sr.services[instance.Name][i] = instance
            return nil
        }
    }

    sr.services[instance.Name] = append(sr.services[instance.Name], instance)
    return nil
}

// 发现服务
func (sr *ServiceRegistry) DiscoverService(serviceName string) ([]ServiceInstance, error) {
    sr.mutex.RLock()
    defer sr.mutex.RUnlock()

    instances, exists := sr.services[serviceName]
    if !exists {
        return nil, fmt.Errorf("service %s not found", serviceName)
    }

    // 过滤健康的实例
    healthyInstances := make([]ServiceInstance, 0)
    for _, instance := range instances {
        if instance.Health.Status == "healthy" {
            healthyInstances = append(healthyInstances, instance)
        }
    }

    return healthyInstances, nil
}

// 心跳检查
func (sr *ServiceRegistry) Heartbeat(serviceID string) error {
    sr.mutex.Lock()
    defer sr.mutex.Unlock()

    for serviceName, instances := range sr.services {
        for i, instance := range instances {
            if instance.ID == serviceID {
                sr.services[serviceName][i].LastHeartbeat = time.Now()
                sr.services[serviceName][i].Health.Status = "healthy"
                return nil
            }
        }
    }

    return fmt.Errorf("service %s not found", serviceID)
}

// 3. 健康检查器
type HealthChecker struct {
    checks   map[string]HealthCheck
    results  map[string]HealthResult
    mutex    sync.RWMutex
    interval time.Duration
}

type HealthCheck struct {
    ID       string        `json:"id"`
    Type     string        `json:"type"` // http, tcp, script
    Target   string        `json:"target"`
    Interval time.Duration `json:"interval"`
    Timeout  time.Duration `json:"timeout"`
    Config   map[string]interface{} `json:"config"`
}

type HealthResult struct {
    CheckID   string    `json:"check_id"`
    Status    string    `json:"status"` // pass, fail, warn
    Message   string    `json:"message"`
    Timestamp time.Time `json:"timestamp"`
    Duration  time.Duration `json:"duration"`
}

func NewHealthChecker(interval time.Duration) *HealthChecker {
    return &HealthChecker{
        checks:   make(map[string]HealthCheck),
        results:  make(map[string]HealthResult),
        interval: interval,
    }
}

// 添加健康检查
func (hc *HealthChecker) AddCheck(check HealthCheck) {
    hc.mutex.Lock()
    defer hc.mutex.Unlock()
    hc.checks[check.ID] = check
}

// 执行健康检查
func (hc *HealthChecker) RunChecks(ctx context.Context) {
    ticker := time.NewTicker(hc.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            hc.executeChecks()
        }
    }
}

func (hc *HealthChecker) executeChecks() {
    hc.mutex.RLock()
    checks := make([]HealthCheck, 0, len(hc.checks))
    for _, check := range hc.checks {
        checks = append(checks, check)
    }
    hc.mutex.RUnlock()

    for _, check := range checks {
        go hc.executeCheck(check)
    }
}

func (hc *HealthChecker) executeCheck(check HealthCheck) {
    start := time.Now()
    var result HealthResult

    switch check.Type {
    case "http":
        result = hc.httpCheck(check)
    case "tcp":
        result = hc.tcpCheck(check)
    case "script":
        result = hc.scriptCheck(check)
    default:
        result = HealthResult{
            CheckID:   check.ID,
            Status:    "fail",
            Message:   "unknown check type",
            Timestamp: time.Now(),
            Duration:  time.Since(start),
        }
    }

    hc.mutex.Lock()
    hc.results[check.ID] = result
    hc.mutex.Unlock()
}

func (hc *HealthChecker) httpCheck(check HealthCheck) HealthResult {
    // HTTP健康检查实现
    return HealthResult{
        CheckID:   check.ID,
        Status:    "pass",
        Message:   "HTTP check passed",
        Timestamp: time.Now(),
        Duration:  50 * time.Millisecond,
    }
}

func (hc *HealthChecker) tcpCheck(check HealthCheck) HealthResult {
    // TCP健康检查实现
    return HealthResult{
        CheckID:   check.ID,
        Status:    "pass",
        Message:   "TCP check passed",
        Timestamp: time.Now(),
        Duration:  10 * time.Millisecond,
    }
}

func (hc *HealthChecker) scriptCheck(check HealthCheck) HealthResult {
    // 脚本健康检查实现
    return HealthResult{
        CheckID:   check.ID,
        Status:    "pass",
        Message:   "Script check passed",
        Timestamp: time.Now(),
        Duration:  100 * time.Millisecond,
    }
}

// 4. 故障转移管理器
type FailoverManager struct {
    policies []FailoverPolicy
    mutex    sync.RWMutex
}

type FailoverPolicy struct {
    ServiceName     string        `json:"service_name"`
    FailureThreshold int          `json:"failure_threshold"`
    RecoveryTime    time.Duration `json:"recovery_time"`
    Actions         []FailoverAction `json:"actions"`
}

type FailoverAction struct {
    Type   string                 `json:"type"` // switch, scale, notify
    Config map[string]interface{} `json:"config"`
}

func NewFailoverManager() *FailoverManager {
    return &FailoverManager{
        policies: make([]FailoverPolicy, 0),
    }
}

// 添加故障转移策略
func (fm *FailoverManager) AddPolicy(policy FailoverPolicy) {
    fm.mutex.Lock()
    defer fm.mutex.Unlock()
    fm.policies = append(fm.policies, policy)
}

// 执行故障转移
func (fm *FailoverManager) ExecuteFailover(serviceName string, failureCount int) {
    fm.mutex.RLock()
    defer fm.mutex.RUnlock()

    for _, policy := range fm.policies {
        if policy.ServiceName == serviceName && failureCount >= policy.FailureThreshold {
            for _, action := range policy.Actions {
                go fm.executeAction(action, serviceName)
            }
        }
    }
}

func (fm *FailoverManager) executeAction(action FailoverAction, serviceName string) {
    switch action.Type {
    case "switch":
        fmt.Printf("Switching traffic for service %s\n", serviceName)
    case "scale":
        fmt.Printf("Scaling service %s\n", serviceName)
    case "notify":
        fmt.Printf("Notifying administrators about service %s failure\n", serviceName)
    }
}

// 使用示例
func ExampleHighAvailabilityArchitecture() {
    // 创建高可用架构
    ha := &HighAvailabilityArchitecture{
        LoadBalancer:    NewLoadBalancer("weighted"),
        ServiceRegistry: NewServiceRegistry(),
        HealthChecker:   NewHealthChecker(30 * time.Second),
        FailoverManager: NewFailoverManager(),
    }

    // 注册服务实例
    ha.ServiceRegistry.RegisterService(ServiceInstance{
        ID:      "mall-api-1",
        Name:    "mall-api",
        Address: "10.0.1.10",
        Port:    8080,
        Tags:    []string{"api", "v1"},
        TTL:     60 * time.Second,
    })

    ha.ServiceRegistry.RegisterService(ServiceInstance{
        ID:      "mall-api-2",
        Name:    "mall-api",
        Address: "10.0.1.11",
        Port:    8080,
        Tags:    []string{"api", "v1"},
        TTL:     60 * time.Second,
    })

    // 添加负载均衡后端
    ha.LoadBalancer.AddBackend(Backend{
        ID:      "mall-api-1",
        Address: "10.0.1.10:8080",
        Weight:  100,
        Healthy: true,
    })

    ha.LoadBalancer.AddBackend(Backend{
        ID:      "mall-api-2",
        Address: "10.0.1.11:8080",
        Weight:  100,
        Healthy: true,
    })

    // 添加健康检查
    ha.HealthChecker.AddCheck(HealthCheck{
        ID:       "mall-api-health",
        Type:     "http",
        Target:   "http://10.0.1.10:8080/health",
        Interval: 30 * time.Second,
        Timeout:  5 * time.Second,
    })

    // 添加故障转移策略
    ha.FailoverManager.AddPolicy(FailoverPolicy{
        ServiceName:      "mall-api",
        FailureThreshold: 3,
        RecoveryTime:     5 * time.Minute,
        Actions: []FailoverAction{
            {Type: "switch", Config: map[string]interface{}{"backup_region": "us-west-2"}},
            {Type: "notify", Config: map[string]interface{}{"channel": "slack"}},
        },
    })

    // 模拟负载均衡
    for i := 0; i < 5; i++ {
        backend, err := ha.LoadBalancer.SelectBackend()
        if err != nil {
            fmt.Printf("Error selecting backend: %v\n", err)
        } else {
            fmt.Printf("Selected backend: %s\n", backend.Address)
        }
    }

    // 模拟服务发现
    instances, err := ha.ServiceRegistry.DiscoverService("mall-api")
    if err != nil {
        fmt.Printf("Error discovering service: %v\n", err)
    } else {
        fmt.Printf("Discovered %d instances\n", len(instances))
    }
}

/*
高可用架构设计要点：
1. 多层冗余：负载均衡、服务实例、数据存储
2. 故障检测：健康检查、监控告警
3. 自动恢复：故障转移、自动扩缩容
4. 数据一致性：主从复制、分布式事务
5. 灾难恢复：跨区域部署、数据备份

扩展思考：
- 如何实现跨区域的故障转移？
- 如何处理网络分区问题？
- 如何保证数据的最终一致性？
- 如何设计合理的降级策略？
*/
```

---

## 📚 章节总结

### 🎯 本章核心要点

通过本章的学习，我们深入掌握了Go语言在生产环境中的实践技能：

#### 1. **容器化部署** 🐳
- **Docker最佳实践**：多阶段构建、非root用户、健康检查
- **Kubernetes编排**：Deployment、Service、Ingress、HPA配置
- **安全配置**：镜像扫描、资源限制、网络策略

#### 2. **监控与可观测性** 📊
- **指标监控**：Prometheus集成、自定义指标、业务监控
- **链路追踪**：OpenTelemetry、分布式追踪、性能分析
- **日志管理**：结构化日志、日志聚合、查询分析
- **告警系统**：规则配置、多渠道通知、故障响应

#### 3. **性能优化** 🔧
- **性能分析**：pprof工具、性能瓶颈识别、优化策略
- **内存优化**：对象池、内存复用、GC调优
- **并发优化**：goroutine管理、channel优化、锁优化
- **缓存策略**：多级缓存、缓存更新、缓存穿透防护

#### 4. **安全实践** 🛡️
- **认证授权**：JWT实现、RBAC权限控制、API密钥管理
- **数据安全**：加密传输、敏感数据保护、安全扫描
- **访问控制**：网络隔离、防火墙配置、安全审计

#### 5. **CI/CD流水线** 🔄
- **自动化构建**：GitHub Actions、多环境部署、质量门禁
- **测试集成**：单元测试、集成测试、安全扫描
- **部署策略**：蓝绿部署、滚动更新、金丝雀发布
- **回滚机制**：版本管理、快速回滚、故障恢复

### 🏢 企业级实践经验

#### **与Java生态对比**
```go
/*
Go vs Java 生产实践对比：

部署方面：
- Go: 单一二进制文件，容器化简单，启动快速
- Java: 需要JVM，WAR/JAR包，启动较慢但生态成熟

监控方面：
- Go: pprof内置，Prometheus原生支持
- Java: JMX监控，APM工具丰富（如Skywalking、Pinpoint）

性能方面：
- Go: 内存占用小，并发性能好，GC延迟低
- Java: 成熟的JVM优化，大量性能调优工具

运维方面：
- Go: 运维简单，依赖少，故障排查相对容易
- Java: 运维复杂，但工具链完善，经验丰富
*/
```

#### **Mall-Go项目生产实践**
```go
// Mall-Go生产环境配置示例
type MallGoProductionConfig struct {
    // 应用配置
    App struct {
        Name         string `yaml:"name"`
        Version      string `yaml:"version"`
        Environment  string `yaml:"environment"`
        LogLevel     string `yaml:"log_level"`
        GracefulTimeout time.Duration `yaml:"graceful_timeout"`
    } `yaml:"app"`

    // 服务配置
    Server struct {
        Host         string        `yaml:"host"`
        Port         int           `yaml:"port"`
        ReadTimeout  time.Duration `yaml:"read_timeout"`
        WriteTimeout time.Duration `yaml:"write_timeout"`
        IdleTimeout  time.Duration `yaml:"idle_timeout"`
    } `yaml:"server"`

    // 数据库配置
    Database struct {
        Master DatabaseConfig `yaml:"master"`
        Slaves []DatabaseConfig `yaml:"slaves"`
    } `yaml:"database"`

    // 缓存配置
    Redis struct {
        Cluster []RedisNode `yaml:"cluster"`
        Sentinel SentinelConfig `yaml:"sentinel"`
    } `yaml:"redis"`

    // 监控配置
    Monitoring struct {
        Prometheus PrometheusConfig `yaml:"prometheus"`
        Jaeger     JaegerConfig     `yaml:"jaeger"`
        Logging    LoggingConfig    `yaml:"logging"`
    } `yaml:"monitoring"`
}

/*
生产环境部署架构：
1. 负载均衡：Nginx + Keepalived
2. 应用层：多实例部署 + 服务发现
3. 数据层：主从复制 + 读写分离
4. 缓存层：Redis集群 + 哨兵模式
5. 监控层：Prometheus + Grafana + Jaeger
*/
```

### 🚀 职业发展建议

#### **技能进阶路径**
1. **初级阶段**：掌握基本的容器化部署和监控配置
2. **中级阶段**：深入理解性能优化和安全实践
3. **高级阶段**：设计高可用架构和故障处理机制
4. **专家阶段**：构建完整的DevOps体系和平台化能力

#### **面试准备重点**
- **生产环境经验**：容器化部署、监控告警、故障处理
- **性能优化案例**：具体的优化实践和效果数据
- **架构设计能力**：高可用、高并发、可扩展的系统设计
- **运维自动化**：CI/CD流水线、基础设施即代码

### 📖 推荐学习资源

#### **官方文档**
- [Go官方文档](https://golang.org/doc/)
- [Docker官方文档](https://docs.docker.com/)
- [Kubernetes官方文档](https://kubernetes.io/docs/)
- [Prometheus官方文档](https://prometheus.io/docs/)

#### **开源项目**
- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [GORM](https://github.com/go-gorm/gorm)
- [Go-Redis](https://github.com/go-redis/redis)
- [OpenTelemetry Go](https://github.com/open-telemetry/opentelemetry-go)

#### **实践项目**
- 构建完整的微服务项目
- 实现监控和告警系统
- 设计CI/CD流水线
- 优化应用性能

---

## 🎉 恭喜完成Go语言学习之旅！

通过完整的学习路径，你已经掌握了：

### 📚 **基础篇**（4章）
- 变量类型与基本语法
- 控制结构与流程控制
- 函数方法与包管理
- 结构体接口与面向对象

### 🚀 **进阶篇**（4章）
- 错误处理与异常管理
- 并发编程与goroutine
- 接口设计与多态实现
- 反射机制与元编程

### 💼 **实战篇**（4章）
- Gin框架Web开发
- GORM数据库操作
- Redis缓存应用
- 消息队列集成

### 🏗️ **架构篇**（1章）
- 微服务架构设计

### 🎯 **高级篇**（1章）
- 生产实践与运维

### 🌟 **你现在具备的能力**
- ✅ **企业级Go开发**：能够独立开发和维护大型Go项目
- ✅ **微服务架构**：具备设计和实现微服务系统的能力
- ✅ **生产环境运维**：掌握容器化部署、监控、性能优化
- ✅ **技术选型能力**：能够根据业务需求选择合适的技术栈
- ✅ **问题解决能力**：具备分析和解决复杂技术问题的能力

### 🎯 **下一步建议**
1. **深入实践**：参与开源项目或构建个人项目
2. **技术分享**：写技术博客或参与技术社区
3. **持续学习**：关注Go语言新特性和生态发展
4. **职业发展**：向高级工程师或架构师方向发展

**记住：技术的学习永无止境，保持好奇心和学习热情，持续精进！** 🚀💪

---

*"代码改变世界，Go语言让你更接近这个目标！继续加油，未来的Go大师！"* 🎊✨
```
```
```
```
```
