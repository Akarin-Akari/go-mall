# 高级篇第三章：监控与日志系统 📊

> *"监控是系统的眼睛，日志是系统的记忆，链路追踪是系统的神经网络。三者结合，构成了现代分布式系统可观测性的完整体系，让我们能够洞察系统的每一个细节！"* 🔍

## 📚 本章学习目标

通过本章学习，你将掌握：

- 📊 **Prometheus+Grafana监控体系**：搭建企业级监控平台，实现全方位系统监控
- 📈 **Go应用指标收集**：掌握runtime metrics、业务指标、自定义指标的收集和分析
- 📋 **ELK/EFK日志聚合**：构建分布式日志收集、存储、分析系统
- 🔗 **分布式链路追踪**：使用Jaeger、Zipkin实现请求链路的完整追踪
- 🚨 **告警系统设计**：配置AlertManager、钉钉、邮件等多渠道告警通知
- 📏 **性能指标定义与SLA**：制定科学的性能基线和服务等级协议
- 🏢 **Mall-Go监控实践**：结合电商项目的完整监控解决方案

---

## 📊 Prometheus+Grafana监控体系

### 监控架构设计

现代分布式系统的监控需要一个完整的可观测性体系，包括指标(Metrics)、日志(Logs)、链路追踪(Traces)三大支柱。

```go
// 监控系统架构定义
package monitoring

import (
    "context"
    "time"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

// 监控系统架构
type MonitoringArchitecture struct {
    // 指标收集层
    MetricsCollection MetricsLayer `json:"metrics_collection"`
    
    // 日志聚合层
    LogAggregation LogLayer `json:"log_aggregation"`
    
    // 链路追踪层
    DistributedTracing TracingLayer `json:"distributed_tracing"`
    
    // 告警通知层
    AlertingSystem AlertingLayer `json:"alerting_system"`
    
    // 可视化展示层
    Visualization VisualizationLayer `json:"visualization"`
    
    // 存储层
    Storage StorageLayer `json:"storage"`
}

// 指标收集层
type MetricsLayer struct {
    // Prometheus配置
    Prometheus struct {
        Server     PrometheusConfig `json:"server"`
        Exporters  []ExporterConfig `json:"exporters"`
        Rules      []RuleConfig     `json:"rules"`
        Targets    []TargetConfig   `json:"targets"`
    } `json:"prometheus"`
    
    // 指标类型
    MetricTypes struct {
        Counter   []CounterMetric   `json:"counter"`
        Gauge     []GaugeMetric     `json:"gauge"`
        Histogram []HistogramMetric `json:"histogram"`
        Summary   []SummaryMetric   `json:"summary"`
    } `json:"metric_types"`
    
    // 业务指标
    BusinessMetrics struct {
        UserMetrics    []string `json:"user_metrics"`
        OrderMetrics   []string `json:"order_metrics"`
        ProductMetrics []string `json:"product_metrics"`
        PaymentMetrics []string `json:"payment_metrics"`
    } `json:"business_metrics"`
}

// Prometheus配置
type PrometheusConfig struct {
    ListenAddress     string        `json:"listen_address"`
    RetentionTime     time.Duration `json:"retention_time"`
    RetentionSize     string        `json:"retention_size"`
    ScrapeInterval    time.Duration `json:"scrape_interval"`
    EvaluationInterval time.Duration `json:"evaluation_interval"`
    ExternalLabels    map[string]string `json:"external_labels"`
    RemoteWrite       []RemoteWriteConfig `json:"remote_write"`
    RemoteRead        []RemoteReadConfig  `json:"remote_read"`
}

// 指标定义
type CounterMetric struct {
    Name        string            `json:"name"`
    Help        string            `json:"help"`
    Labels      []string          `json:"labels"`
    Namespace   string            `json:"namespace"`
    Subsystem   string            `json:"subsystem"`
    ConstLabels map[string]string `json:"const_labels"`
}

type GaugeMetric struct {
    Name        string            `json:"name"`
    Help        string            `json:"help"`
    Labels      []string          `json:"labels"`
    Namespace   string            `json:"namespace"`
    Subsystem   string            `json:"subsystem"`
    ConstLabels map[string]string `json:"const_labels"`
}

type HistogramMetric struct {
    Name        string            `json:"name"`
    Help        string            `json:"help"`
    Labels      []string          `json:"labels"`
    Buckets     []float64         `json:"buckets"`
    Namespace   string            `json:"namespace"`
    Subsystem   string            `json:"subsystem"`
    ConstLabels map[string]string `json:"const_labels"`
}

// Mall-Go监控指标定义
type MallGoMetrics struct {
    // HTTP请求指标
    HTTPRequests struct {
        Total    prometheus.CounterVec   `json:"total"`
        Duration prometheus.HistogramVec `json:"duration"`
        InFlight prometheus.Gauge        `json:"in_flight"`
    } `json:"http_requests"`
    
    // 数据库指标
    Database struct {
        Connections prometheus.GaugeVec     `json:"connections"`
        Queries     prometheus.CounterVec   `json:"queries"`
        Duration    prometheus.HistogramVec `json:"duration"`
        Errors      prometheus.CounterVec   `json:"errors"`
    } `json:"database"`
    
    // 缓存指标
    Cache struct {
        Hits   prometheus.CounterVec `json:"hits"`
        Misses prometheus.CounterVec `json:"misses"`
        Size   prometheus.GaugeVec   `json:"size"`
        TTL    prometheus.GaugeVec   `json:"ttl"`
    } `json:"cache"`
    
    // 业务指标
    Business struct {
        UserRegistrations prometheus.CounterVec `json:"user_registrations"`
        OrdersCreated     prometheus.CounterVec `json:"orders_created"`
        OrdersCompleted   prometheus.CounterVec `json:"orders_completed"`
        PaymentProcessed  prometheus.CounterVec `json:"payment_processed"`
        ProductViews      prometheus.CounterVec `json:"product_views"`
        CartOperations    prometheus.CounterVec `json:"cart_operations"`
    } `json:"business"`
    
    // 系统指标
    System struct {
        CPUUsage    prometheus.GaugeVec `json:"cpu_usage"`
        MemoryUsage prometheus.GaugeVec `json:"memory_usage"`
        DiskUsage   prometheus.GaugeVec `json:"disk_usage"`
        NetworkIO   prometheus.CounterVec `json:"network_io"`
        Goroutines  prometheus.Gauge    `json:"goroutines"`
        GCDuration  prometheus.HistogramVec `json:"gc_duration"`
    } `json:"system"`
}

// 初始化Mall-Go监控指标
func NewMallGoMetrics() *MallGoMetrics {
    return &MallGoMetrics{
        HTTPRequests: struct {
            Total    prometheus.CounterVec   `json:"total"`
            Duration prometheus.HistogramVec `json:"duration"`
            InFlight prometheus.Gauge        `json:"in_flight"`
        }{
            Total: *promauto.NewCounterVec(
                prometheus.CounterOpts{
                    Namespace: "mall_go",
                    Subsystem: "http",
                    Name:      "requests_total",
                    Help:      "Total number of HTTP requests",
                },
                []string{"method", "endpoint", "status_code"},
            ),
            Duration: *promauto.NewHistogramVec(
                prometheus.HistogramOpts{
                    Namespace: "mall_go",
                    Subsystem: "http",
                    Name:      "request_duration_seconds",
                    Help:      "HTTP request duration in seconds",
                    Buckets:   prometheus.DefBuckets,
                },
                []string{"method", "endpoint"},
            ),
            InFlight: promauto.NewGauge(
                prometheus.GaugeOpts{
                    Namespace: "mall_go",
                    Subsystem: "http",
                    Name:      "requests_in_flight",
                    Help:      "Current number of HTTP requests being processed",
                },
            ),
        },
        Database: struct {
            Connections prometheus.GaugeVec     `json:"connections"`
            Queries     prometheus.CounterVec   `json:"queries"`
            Duration    prometheus.HistogramVec `json:"duration"`
            Errors      prometheus.CounterVec   `json:"errors"`
        }{
            Connections: *promauto.NewGaugeVec(
                prometheus.GaugeOpts{
                    Namespace: "mall_go",
                    Subsystem: "database",
                    Name:      "connections",
                    Help:      "Current number of database connections",
                },
                []string{"database", "state"},
            ),
            Queries: *promauto.NewCounterVec(
                prometheus.CounterOpts{
                    Namespace: "mall_go",
                    Subsystem: "database",
                    Name:      "queries_total",
                    Help:      "Total number of database queries",
                },
                []string{"database", "operation", "table"},
            ),
            Duration: *promauto.NewHistogramVec(
                prometheus.HistogramOpts{
                    Namespace: "mall_go",
                    Subsystem: "database",
                    Name:      "query_duration_seconds",
                    Help:      "Database query duration in seconds",
                    Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 5.0},
                },
                []string{"database", "operation"},
            ),
            Errors: *promauto.NewCounterVec(
                prometheus.CounterOpts{
                    Namespace: "mall_go",
                    Subsystem: "database",
                    Name:      "errors_total",
                    Help:      "Total number of database errors",
                },
                []string{"database", "error_type"},
            ),
        },
        Business: struct {
            UserRegistrations prometheus.CounterVec `json:"user_registrations"`
            OrdersCreated     prometheus.CounterVec `json:"orders_created"`
            OrdersCompleted   prometheus.CounterVec `json:"orders_completed"`
            PaymentProcessed  prometheus.CounterVec `json:"payment_processed"`
            ProductViews      prometheus.CounterVec `json:"product_views"`
            CartOperations    prometheus.CounterVec `json:"cart_operations"`
        }{
            UserRegistrations: *promauto.NewCounterVec(
                prometheus.CounterOpts{
                    Namespace: "mall_go",
                    Subsystem: "business",
                    Name:      "user_registrations_total",
                    Help:      "Total number of user registrations",
                },
                []string{"source", "type"},
            ),
            OrdersCreated: *promauto.NewCounterVec(
                prometheus.CounterOpts{
                    Namespace: "mall_go",
                    Subsystem: "business",
                    Name:      "orders_created_total",
                    Help:      "Total number of orders created",
                },
                []string{"category", "payment_method"},
            ),
            OrdersCompleted: *promauto.NewCounterVec(
                prometheus.CounterOpts{
                    Namespace: "mall_go",
                    Subsystem: "business",
                    Name:      "orders_completed_total",
                    Help:      "Total number of orders completed",
                },
                []string{"category", "payment_method"},
            ),
            PaymentProcessed: *promauto.NewCounterVec(
                prometheus.CounterOpts{
                    Namespace: "mall_go",
                    Subsystem: "business",
                    Name:      "payment_processed_total",
                    Help:      "Total amount of payments processed",
                },
                []string{"method", "currency", "status"},
            ),
        },
    }
}
```

### Prometheus配置

```yaml
# prometheus.yml - Prometheus主配置文件
global:
  scrape_interval: 15s
  evaluation_interval: 15s
  external_labels:
    cluster: 'mall-go-prod'
    region: 'us-west-2'

# 告警规则文件
rule_files:
  - "rules/*.yml"

# 告警管理器配置
alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093

# 抓取配置
scrape_configs:
  # Prometheus自身监控
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  # Mall-Go应用监控
  - job_name: 'mall-go'
    metrics_path: '/metrics'
    scrape_interval: 10s
    static_configs:
      - targets: ['mall-go:8080']
    relabel_configs:
      - source_labels: [__address__]
        target_label: instance
      - source_labels: [__meta_kubernetes_pod_name]
        target_label: pod
      - source_labels: [__meta_kubernetes_namespace]
        target_label: namespace

  # Node Exporter监控
  - job_name: 'node-exporter'
    static_configs:
      - targets: 
        - 'node-exporter:9100'
    relabel_configs:
      - source_labels: [__address__]
        regex: '([^:]+):(.*)'
        target_label: __address__
        replacement: '${1}:9100'

  # MySQL监控
  - job_name: 'mysql-exporter'
    static_configs:
      - targets: ['mysql-exporter:9104']

  # Redis监控
  - job_name: 'redis-exporter'
    static_configs:
      - targets: ['redis-exporter:9121']

  # Kubernetes API Server监控
  - job_name: 'kubernetes-apiservers'
    kubernetes_sd_configs:
    - role: endpoints
    scheme: https
    tls_config:
      ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
    bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
    relabel_configs:
    - source_labels: [__meta_kubernetes_namespace, __meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
      action: keep
      regex: default;kubernetes;https

  # Kubernetes节点监控
  - job_name: 'kubernetes-nodes'
    kubernetes_sd_configs:
    - role: node
    scheme: https
    tls_config:
      ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
    bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
    relabel_configs:
    - action: labelmap
      regex: __meta_kubernetes_node_label_(.+)
    - target_label: __address__
      replacement: kubernetes.default.svc:443
    - source_labels: [__meta_kubernetes_node_name]
      regex: (.+)
      target_label: __metrics_path__
      replacement: /api/v1/nodes/${1}/proxy/metrics

  # Kubernetes Pod监控
  - job_name: 'kubernetes-pods'
    kubernetes_sd_configs:
    - role: pod
    relabel_configs:
    - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
      action: keep
      regex: true
    - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
      action: replace
      target_label: __metrics_path__
      regex: (.+)
    - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
      action: replace
      regex: ([^:]+)(?::\d+)?;(\d+)
      replacement: $1:$2
      target_label: __address__
    - action: labelmap
      regex: __meta_kubernetes_pod_label_(.+)
    - source_labels: [__meta_kubernetes_namespace]
      action: replace
      target_label: kubernetes_namespace
    - source_labels: [__meta_kubernetes_pod_name]
      action: replace
      target_label: kubernetes_pod_name

# 远程写入配置（可选）
remote_write:
  - url: "https://prometheus-remote-write.example.com/api/v1/write"
    basic_auth:
      username: "user"
      password: "password"
    write_relabel_configs:
      - source_labels: [__name__]
        regex: 'mall_go_.*'
        action: keep

# 远程读取配置（可选）
remote_read:
  - url: "https://prometheus-remote-read.example.com/api/v1/read"
    basic_auth:
      username: "user"
      password: "password"
```

---

## 📈 Go应用指标收集

### Runtime Metrics收集

Go运行时提供了丰富的内置指标，这些指标对于监控应用性能至关重要。

```go
// Go运行时指标收集器
package metrics

import (
    "context"
    "runtime"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

// Go运行时指标收集器
type GoRuntimeCollector struct {
    // 内存指标
    memStats struct {
        Alloc         prometheus.Gauge
        TotalAlloc    prometheus.Counter
        Sys           prometheus.Gauge
        Lookups       prometheus.Counter
        Mallocs       prometheus.Counter
        Frees         prometheus.Counter
        HeapAlloc     prometheus.Gauge
        HeapSys       prometheus.Gauge
        HeapIdle      prometheus.Gauge
        HeapInuse     prometheus.Gauge
        HeapReleased  prometheus.Gauge
        HeapObjects   prometheus.Gauge
        StackInuse    prometheus.Gauge
        StackSys      prometheus.Gauge
        MSpanInuse    prometheus.Gauge
        MSpanSys      prometheus.Gauge
        MCacheInuse   prometheus.Gauge
        MCacheSys     prometheus.Gauge
        BuckHashSys   prometheus.Gauge
        GCSys         prometheus.Gauge
        OtherSys      prometheus.Gauge
        NextGC        prometheus.Gauge
        LastGC        prometheus.Gauge
        PauseTotalNs  prometheus.Counter
        PauseNs       prometheus.Histogram
        NumGC         prometheus.Counter
        NumForcedGC   prometheus.Counter
        GCCPUFraction prometheus.Gauge
    }

    // Goroutine指标
    goroutines prometheus.Gauge

    // 线程指标
    threads prometheus.Gauge

    // 文件描述符指标
    openFDs prometheus.Gauge
    maxFDs  prometheus.Gauge

    // 收集间隔
    interval time.Duration

    // 停止信号
    stopCh chan struct{}
}

// 创建Go运行时指标收集器
func NewGoRuntimeCollector(interval time.Duration) *GoRuntimeCollector {
    collector := &GoRuntimeCollector{
        interval: interval,
        stopCh:   make(chan struct{}),
    }

    // 初始化内存指标
    collector.memStats.Alloc = promauto.NewGauge(prometheus.GaugeOpts{
        Namespace: "go",
        Subsystem: "memstats",
        Name:      "alloc_bytes",
        Help:      "Number of bytes allocated and still in use.",
    })

    collector.memStats.TotalAlloc = promauto.NewCounter(prometheus.CounterOpts{
        Namespace: "go",
        Subsystem: "memstats",
        Name:      "alloc_bytes_total",
        Help:      "Total number of bytes allocated, even if freed.",
    })

    collector.memStats.Sys = promauto.NewGauge(prometheus.GaugeOpts{
        Namespace: "go",
        Subsystem: "memstats",
        Name:      "sys_bytes",
        Help:      "Number of bytes obtained from system.",
    })

    collector.memStats.HeapAlloc = promauto.NewGauge(prometheus.GaugeOpts{
        Namespace: "go",
        Subsystem: "memstats",
        Name:      "heap_alloc_bytes",
        Help:      "Number of heap bytes allocated and still in use.",
    })

    collector.memStats.HeapSys = promauto.NewGauge(prometheus.GaugeOpts{
        Namespace: "go",
        Subsystem: "memstats",
        Name:      "heap_sys_bytes",
        Help:      "Number of heap bytes obtained from system.",
    })

    collector.memStats.HeapIdle = promauto.NewGauge(prometheus.GaugeOpts{
        Namespace: "go",
        Subsystem: "memstats",
        Name:      "heap_idle_bytes",
        Help:      "Number of heap bytes waiting to be used.",
    })

    collector.memStats.HeapInuse = promauto.NewGauge(prometheus.GaugeOpts{
        Namespace: "go",
        Subsystem: "memstats",
        Name:      "heap_inuse_bytes",
        Help:      "Number of heap bytes that are in use.",
    })

    collector.memStats.HeapObjects = promauto.NewGauge(prometheus.GaugeOpts{
        Namespace: "go",
        Subsystem: "memstats",
        Name:      "heap_objects",
        Help:      "Number of allocated objects.",
    })

    collector.memStats.PauseNs = promauto.NewHistogram(prometheus.HistogramOpts{
        Namespace: "go",
        Subsystem: "gc",
        Name:      "duration_seconds",
        Help:      "Time spent in garbage collection.",
        Buckets:   []float64{1e-9, 1e-8, 1e-7, 1e-6, 1e-5, 1e-4, 1e-3, 1e-2, 1e-1},
    })

    collector.memStats.NumGC = promauto.NewCounter(prometheus.CounterOpts{
        Namespace: "go",
        Subsystem: "gc",
        Name:      "cycles_total",
        Help:      "Number of completed GC cycles.",
    })

    collector.memStats.GCCPUFraction = promauto.NewGauge(prometheus.GaugeOpts{
        Namespace: "go",
        Subsystem: "gc",
        Name:      "cpu_fraction",
        Help:      "Fraction of CPU time used by GC.",
    })

    // 初始化Goroutine指标
    collector.goroutines = promauto.NewGauge(prometheus.GaugeOpts{
        Namespace: "go",
        Name:      "goroutines",
        Help:      "Number of goroutines that currently exist.",
    })

    // 初始化线程指标
    collector.threads = promauto.NewGauge(prometheus.GaugeOpts{
        Namespace: "go",
        Name:      "threads",
        Help:      "Number of OS threads created.",
    })

    return collector
}

// 开始收集指标
func (c *GoRuntimeCollector) Start(ctx context.Context) {
    ticker := time.NewTicker(c.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-c.stopCh:
            return
        case <-ticker.C:
            c.collect()
        }
    }
}

// 停止收集
func (c *GoRuntimeCollector) Stop() {
    close(c.stopCh)
}

// 收集指标
func (c *GoRuntimeCollector) collect() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    // 更新内存指标
    c.memStats.Alloc.Set(float64(m.Alloc))
    c.memStats.TotalAlloc.Add(float64(m.TotalAlloc))
    c.memStats.Sys.Set(float64(m.Sys))
    c.memStats.HeapAlloc.Set(float64(m.HeapAlloc))
    c.memStats.HeapSys.Set(float64(m.HeapSys))
    c.memStats.HeapIdle.Set(float64(m.HeapIdle))
    c.memStats.HeapInuse.Set(float64(m.HeapInuse))
    c.memStats.HeapObjects.Set(float64(m.HeapObjects))

    // 更新GC指标
    c.memStats.NumGC.Add(float64(m.NumGC))
    c.memStats.GCCPUFraction.Set(m.GCCPUFraction)

    // 记录GC暂停时间
    if len(m.PauseNs) > 0 {
        c.memStats.PauseNs.Observe(float64(m.PauseNs[(m.NumGC+255)%256]) / 1e9)
    }

    // 更新Goroutine数量
    c.goroutines.Set(float64(runtime.NumGoroutine()))

    // 更新线程数量
    var numThreads int
    if pprof := runtime.GOMAXPROCS(0); pprof > 0 {
        numThreads = pprof
    }
    c.threads.Set(float64(numThreads))
}
```

### 业务指标收集

```go
// 业务指标收集器
package metrics

import (
    "context"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus"
    "gorm.io/gorm"
)

// 业务指标收集器
type BusinessMetricsCollector struct {
    metrics *MallGoMetrics
    db      *gorm.DB
}

// 创建业务指标收集器
func NewBusinessMetricsCollector(metrics *MallGoMetrics, db *gorm.DB) *BusinessMetricsCollector {
    return &BusinessMetricsCollector{
        metrics: metrics,
        db:      db,
    }
}

// HTTP中间件 - 收集HTTP请求指标
func (c *BusinessMetricsCollector) HTTPMetricsMiddleware() gin.HandlerFunc {
    return func(ctx *gin.Context) {
        start := time.Now()

        // 增加正在处理的请求数
        c.metrics.HTTPRequests.InFlight.Inc()
        defer c.metrics.HTTPRequests.InFlight.Dec()

        // 处理请求
        ctx.Next()

        // 计算请求耗时
        duration := time.Since(start).Seconds()

        // 获取请求信息
        method := ctx.Request.Method
        endpoint := ctx.FullPath()
        statusCode := strconv.Itoa(ctx.Writer.Status())

        // 记录指标
        c.metrics.HTTPRequests.Total.WithLabelValues(method, endpoint, statusCode).Inc()
        c.metrics.HTTPRequests.Duration.WithLabelValues(method, endpoint).Observe(duration)
    }
}

// 数据库中间件 - 收集数据库操作指标
func (c *BusinessMetricsCollector) DatabaseMetricsPlugin() gorm.Plugin {
    return &databaseMetricsPlugin{collector: c}
}

type databaseMetricsPlugin struct {
    collector *BusinessMetricsCollector
}

func (p *databaseMetricsPlugin) Name() string {
    return "metrics"
}

func (p *databaseMetricsPlugin) Initialize(db *gorm.DB) error {
    // 注册回调函数
    db.Callback().Create().Before("gorm:create").Register("metrics:before_create", p.beforeCreate)
    db.Callback().Create().After("gorm:create").Register("metrics:after_create", p.afterCreate)

    db.Callback().Query().Before("gorm:query").Register("metrics:before_query", p.beforeQuery)
    db.Callback().Query().After("gorm:query").Register("metrics:after_query", p.afterQuery)

    db.Callback().Update().Before("gorm:update").Register("metrics:before_update", p.beforeUpdate)
    db.Callback().Update().After("gorm:update").Register("metrics:after_update", p.afterUpdate)

    db.Callback().Delete().Before("gorm:delete").Register("metrics:before_delete", p.beforeDelete)
    db.Callback().Delete().After("gorm:delete").Register("metrics:after_delete", p.afterDelete)

    return nil
}

func (p *databaseMetricsPlugin) beforeCreate(db *gorm.DB) {
    db.Set("metrics:start_time", time.Now())
}

func (p *databaseMetricsPlugin) afterCreate(db *gorm.DB) {
    p.recordMetrics(db, "create")
}

func (p *databaseMetricsPlugin) beforeQuery(db *gorm.DB) {
    db.Set("metrics:start_time", time.Now())
}

func (p *databaseMetricsPlugin) afterQuery(db *gorm.DB) {
    p.recordMetrics(db, "query")
}

func (p *databaseMetricsPlugin) beforeUpdate(db *gorm.DB) {
    db.Set("metrics:start_time", time.Now())
}

func (p *databaseMetricsPlugin) afterUpdate(db *gorm.DB) {
    p.recordMetrics(db, "update")
}

func (p *databaseMetricsPlugin) beforeDelete(db *gorm.DB) {
    db.Set("metrics:start_time", time.Now())
}

func (p *databaseMetricsPlugin) afterDelete(db *gorm.DB) {
    p.recordMetrics(db, "delete")
}

func (p *databaseMetricsPlugin) recordMetrics(db *gorm.DB, operation string) {
    startTime, exists := db.Get("metrics:start_time")
    if !exists {
        return
    }

    start, ok := startTime.(time.Time)
    if !ok {
        return
    }

    duration := time.Since(start).Seconds()
    tableName := db.Statement.Table

    // 记录查询指标
    p.collector.metrics.Database.Queries.WithLabelValues("mysql", operation, tableName).Inc()
    p.collector.metrics.Database.Duration.WithLabelValues("mysql", operation).Observe(duration)

    // 记录错误指标
    if db.Error != nil {
        p.collector.metrics.Database.Errors.WithLabelValues("mysql", "query_error").Inc()
    }
}

// 用户注册指标
func (c *BusinessMetricsCollector) RecordUserRegistration(source, userType string) {
    c.metrics.Business.UserRegistrations.WithLabelValues(source, userType).Inc()
}

// 订单创建指标
func (c *BusinessMetricsCollector) RecordOrderCreated(category, paymentMethod string) {
    c.metrics.Business.OrdersCreated.WithLabelValues(category, paymentMethod).Inc()
}

// 订单完成指标
func (c *BusinessMetricsCollector) RecordOrderCompleted(category, paymentMethod string) {
    c.metrics.Business.OrdersCompleted.WithLabelValues(category, paymentMethod).Inc()
}

// 支付处理指标
func (c *BusinessMetricsCollector) RecordPaymentProcessed(method, currency, status string, amount float64) {
    c.metrics.Business.PaymentProcessed.WithLabelValues(method, currency, status).Add(amount)
}

// 产品浏览指标
func (c *BusinessMetricsCollector) RecordProductView(productID, category string) {
    c.metrics.Business.ProductViews.WithLabelValues(productID, category).Inc()
}

// 购物车操作指标
func (c *BusinessMetricsCollector) RecordCartOperation(operation, userID string) {
    c.metrics.Business.CartOperations.WithLabelValues(operation, userID).Inc()
}
```

### 自定义指标示例

```go
// 自定义指标示例
package metrics

import (
    "time"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

// 自定义指标管理器
type CustomMetricsManager struct {
    // 缓存命中率
    cacheHitRate prometheus.GaugeVec

    // 队列长度
    queueLength prometheus.GaugeVec

    // 连接池状态
    connectionPoolStatus prometheus.GaugeVec

    // 业务处理时间
    businessProcessTime prometheus.HistogramVec

    // 错误率
    errorRate prometheus.GaugeVec

    // 吞吐量
    throughput prometheus.CounterVec
}

// 创建自定义指标管理器
func NewCustomMetricsManager() *CustomMetricsManager {
    return &CustomMetricsManager{
        cacheHitRate: *promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Namespace: "mall_go",
                Subsystem: "cache",
                Name:      "hit_rate",
                Help:      "Cache hit rate percentage",
            },
            []string{"cache_type", "key_pattern"},
        ),

        queueLength: *promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Namespace: "mall_go",
                Subsystem: "queue",
                Name:      "length",
                Help:      "Current queue length",
            },
            []string{"queue_name", "priority"},
        ),

        connectionPoolStatus: *promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Namespace: "mall_go",
                Subsystem: "connection_pool",
                Name:      "connections",
                Help:      "Connection pool status",
            },
            []string{"pool_name", "status"},
        ),

        businessProcessTime: *promauto.NewHistogramVec(
            prometheus.HistogramOpts{
                Namespace: "mall_go",
                Subsystem: "business",
                Name:      "process_duration_seconds",
                Help:      "Business process duration in seconds",
                Buckets:   []float64{0.01, 0.05, 0.1, 0.5, 1.0, 2.0, 5.0, 10.0},
            },
            []string{"process_type", "status"},
        ),

        errorRate: *promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Namespace: "mall_go",
                Subsystem: "error",
                Name:      "rate",
                Help:      "Error rate percentage",
            },
            []string{"service", "error_type"},
        ),

        throughput: *promauto.NewCounterVec(
            prometheus.CounterOpts{
                Namespace: "mall_go",
                Subsystem: "throughput",
                Name:      "operations_total",
                Help:      "Total number of operations processed",
            },
            []string{"operation_type", "result"},
        ),
    }
}

// 记录缓存命中率
func (m *CustomMetricsManager) RecordCacheHitRate(cacheType, keyPattern string, hitRate float64) {
    m.cacheHitRate.WithLabelValues(cacheType, keyPattern).Set(hitRate)
}

// 记录队列长度
func (m *CustomMetricsManager) RecordQueueLength(queueName, priority string, length int) {
    m.queueLength.WithLabelValues(queueName, priority).Set(float64(length))
}

// 记录连接池状态
func (m *CustomMetricsManager) RecordConnectionPoolStatus(poolName, status string, count int) {
    m.connectionPoolStatus.WithLabelValues(poolName, status).Set(float64(count))
}

// 记录业务处理时间
func (m *CustomMetricsManager) RecordBusinessProcessTime(processType, status string, duration time.Duration) {
    m.businessProcessTime.WithLabelValues(processType, status).Observe(duration.Seconds())
}

// 记录错误率
func (m *CustomMetricsManager) RecordErrorRate(service, errorType string, rate float64) {
    m.errorRate.WithLabelValues(service, errorType).Set(rate)
}

// 记录吞吐量
func (m *CustomMetricsManager) RecordThroughput(operationType, result string) {
    m.throughput.WithLabelValues(operationType, result).Inc()
}
```

---

## 📋 ELK/EFK日志聚合架构

### 日志聚合架构设计

ELK(Elasticsearch + Logstash + Kibana)和EFK(Elasticsearch + Fluentd + Kibana)是两种主流的日志聚合解决方案。

```go
// 日志聚合架构定义
package logging

import (
    "context"
    "encoding/json"
    "time"
)

// 日志聚合架构
type LogAggregationArchitecture struct {
    // 日志收集层
    Collection LogCollectionLayer `json:"collection"`

    // 日志处理层
    Processing LogProcessingLayer `json:"processing"`

    // 日志存储层
    Storage LogStorageLayer `json:"storage"`

    // 日志可视化层
    Visualization LogVisualizationLayer `json:"visualization"`

    // 日志告警层
    Alerting LogAlertingLayer `json:"alerting"`
}

// 日志收集层
type LogCollectionLayer struct {
    // 应用日志收集
    ApplicationLogs struct {
        Agents    []LogAgent    `json:"agents"`
        Formats   []LogFormat   `json:"formats"`
        Filters   []LogFilter   `json:"filters"`
        Buffers   []LogBuffer   `json:"buffers"`
    } `json:"application_logs"`

    // 系统日志收集
    SystemLogs struct {
        Syslog    SyslogConfig    `json:"syslog"`
        Journald  JournaldConfig  `json:"journald"`
        Files     []FileConfig    `json:"files"`
    } `json:"system_logs"`

    // 容器日志收集
    ContainerLogs struct {
        Docker     DockerLogConfig     `json:"docker"`
        Kubernetes KubernetesLogConfig `json:"kubernetes"`
        Containerd ContainerdLogConfig `json:"containerd"`
    } `json:"container_logs"`
}

// 日志代理配置
type LogAgent struct {
    Name     string            `json:"name"`     // fluentd, filebeat, logstash
    Version  string            `json:"version"`
    Config   map[string]interface{} `json:"config"`
    Plugins  []string          `json:"plugins"`
    Resources ResourceRequirements `json:"resources"`
}

// 日志格式定义
type LogFormat struct {
    Name        string            `json:"name"`
    Pattern     string            `json:"pattern"`
    Fields      []LogField        `json:"fields"`
    Multiline   MultilineConfig   `json:"multiline"`
    Timestamp   TimestampConfig   `json:"timestamp"`
}

type LogField struct {
    Name     string `json:"name"`
    Type     string `json:"type"`     // string, number, boolean, date
    Required bool   `json:"required"`
    Index    bool   `json:"index"`
}

// Mall-Go结构化日志定义
type MallGoLogEntry struct {
    // 基础字段
    Timestamp   time.Time `json:"@timestamp"`
    Level       string    `json:"level"`
    Message     string    `json:"message"`
    Logger      string    `json:"logger"`

    // 应用字段
    Service     string    `json:"service"`
    Version     string    `json:"version"`
    Environment string    `json:"environment"`

    // 请求字段
    RequestID   string    `json:"request_id"`
    UserID      string    `json:"user_id"`
    SessionID   string    `json:"session_id"`
    IP          string    `json:"ip"`
    UserAgent   string    `json:"user_agent"`

    // HTTP字段
    Method      string    `json:"method"`
    URL         string    `json:"url"`
    StatusCode  int       `json:"status_code"`
    Duration    float64   `json:"duration"`

    // 业务字段
    OrderID     string    `json:"order_id,omitempty"`
    ProductID   string    `json:"product_id,omitempty"`
    CategoryID  string    `json:"category_id,omitempty"`
    PaymentID   string    `json:"payment_id,omitempty"`

    // 错误字段
    Error       string    `json:"error,omitempty"`
    Stack       string    `json:"stack,omitempty"`

    // 自定义字段
    Extra       map[string]interface{} `json:"extra,omitempty"`

    // 基础设施字段
    Host        string    `json:"host"`
    Pod         string    `json:"pod,omitempty"`
    Container   string    `json:"container,omitempty"`
    Namespace   string    `json:"namespace,omitempty"`
}

// 结构化日志记录器
type StructuredLogger struct {
    service     string
    version     string
    environment string
    host        string
    output      LogOutput
}

type LogOutput interface {
    Write(entry *MallGoLogEntry) error
    Close() error
}

// 创建结构化日志记录器
func NewStructuredLogger(service, version, environment, host string, output LogOutput) *StructuredLogger {
    return &StructuredLogger{
        service:     service,
        version:     version,
        environment: environment,
        host:        host,
        output:      output,
    }
}

// 记录信息日志
func (l *StructuredLogger) Info(ctx context.Context, message string, fields map[string]interface{}) {
    entry := l.createLogEntry(ctx, "INFO", message, fields)
    l.output.Write(entry)
}

// 记录警告日志
func (l *StructuredLogger) Warn(ctx context.Context, message string, fields map[string]interface{}) {
    entry := l.createLogEntry(ctx, "WARN", message, fields)
    l.output.Write(entry)
}

// 记录错误日志
func (l *StructuredLogger) Error(ctx context.Context, message string, err error, fields map[string]interface{}) {
    if fields == nil {
        fields = make(map[string]interface{})
    }
    if err != nil {
        fields["error"] = err.Error()
    }
    entry := l.createLogEntry(ctx, "ERROR", message, fields)
    l.output.Write(entry)
}

// 创建日志条目
func (l *StructuredLogger) createLogEntry(ctx context.Context, level, message string, fields map[string]interface{}) *MallGoLogEntry {
    entry := &MallGoLogEntry{
        Timestamp:   time.Now(),
        Level:       level,
        Message:     message,
        Logger:      "mall-go",
        Service:     l.service,
        Version:     l.version,
        Environment: l.environment,
        Host:        l.host,
        Extra:       fields,
    }

    // 从上下文中提取字段
    if requestID := ctx.Value("request_id"); requestID != nil {
        if id, ok := requestID.(string); ok {
            entry.RequestID = id
        }
    }

    if userID := ctx.Value("user_id"); userID != nil {
        if id, ok := userID.(string); ok {
            entry.UserID = id
        }
    }

    if sessionID := ctx.Value("session_id"); sessionID != nil {
        if id, ok := sessionID.(string); ok {
            entry.SessionID = id
        }
    }

    return entry
}
```

### Fluentd配置

```yaml
# fluentd.conf - Fluentd主配置文件
<system>
  workers 4
  root_dir /var/log/fluentd
</system>

# 输入源配置
<source>
  @type tail
  @id mall_go_app_logs
  path /var/log/mall-go/*.log
  pos_file /var/log/fluentd/mall-go.log.pos
  tag mall-go.app
  format json
  time_key @timestamp
  time_format %Y-%m-%dT%H:%M:%S.%L%z
  keep_time_key true
  read_from_head true
  refresh_interval 10

  <parse>
    @type json
    time_key @timestamp
    time_format %Y-%m-%dT%H:%M:%S.%L%z
    keep_time_key true
  </parse>
</source>

# Docker容器日志收集
<source>
  @type tail
  @id docker_container_logs
  path /var/lib/docker/containers/*/*-json.log
  pos_file /var/log/fluentd/docker.log.pos
  tag docker.*
  format json
  time_key time
  time_format %Y-%m-%dT%H:%M:%S.%L%z

  <parse>
    @type json
    time_key time
    time_format %Y-%m-%dT%H:%M:%S.%L%z
  </parse>
</source>

# Kubernetes Pod日志收集
<source>
  @type kubernetes_metadata
  @id kubernetes_metadata
  kubernetes_url "#{ENV['KUBERNETES_SERVICE_HOST']}:#{ENV['KUBERNETES_SERVICE_PORT_HTTPS']}"
  verify_ssl "#{ENV['KUBERNETES_VERIFY_SSL'] || true}"
  ca_file "#{ENV['KUBERNETES_CA_FILE']}"
  bearer_token_file "#{ENV['KUBERNETES_TOKEN_FILE']}"
  cache_size 1000
  cache_ttl 3600
  watch true
  de_dot false
  annotation_match ["fluentd.*"]
  allow_orphans true
</source>

# 系统日志收集
<source>
  @type systemd
  @id systemd_logs
  tag systemd
  path /var/log/journal
  matches [{ "_SYSTEMD_UNIT": "mall-go.service" }]
  read_from_head true
  strip_underscores true

  <storage>
    @type local
    persistent true
    path /var/log/fluentd/systemd.pos
  </storage>

  <entry>
    field_map {"MESSAGE": "message", "_HOSTNAME": "hostname", "_SYSTEMD_UNIT": "unit"}
    fields_strip_underscores true
    fields_lowercase true
  </entry>
</source>

# 过滤器配置
<filter mall-go.**>
  @type record_transformer
  @id mall_go_transformer
  enable_ruby true
  auto_typecast true
  renew_record false
  renew_time_key false
  keep_keys level,message,service,request_id,user_id,method,url,status_code,duration

  <record>
    hostname "#{Socket.gethostname}"
    environment "#{ENV['ENVIRONMENT'] || 'production'}"
    cluster "#{ENV['CLUSTER_NAME'] || 'mall-go-prod'}"
    region "#{ENV['AWS_REGION'] || 'us-west-2'}"
    log_type application
    source fluentd
  </record>
</filter>

# 错误日志特殊处理
<filter mall-go.**>
  @type grep
  @id error_log_filter

  <regexp>
    key level
    pattern ^(ERROR|FATAL)$
  </regexp>
</filter>

# 敏感信息过滤
<filter mall-go.**>
  @type record_modifier
  @id sensitive_data_filter

  <record>
    message ${record["message"].gsub(/password["\s]*[:=]["\s]*[^"\s,}]+/i, 'password=***')}
    message ${record["message"].gsub(/token["\s]*[:=]["\s]*[^"\s,}]+/i, 'token=***')}
    message ${record["message"].gsub(/key["\s]*[:=]["\s]*[^"\s,}]+/i, 'key=***')}
  </record>
</filter>

# 日志采样（减少日志量）
<filter mall-go.**>
  @type sampling
  @id log_sampling
  sampling_rate 10

  <rule>
    key level
    pattern ^(ERROR|FATAL)$
    sample_rate 100
  </rule>

  <rule>
    key level
    pattern ^WARN$
    sample_rate 50
  </rule>

  <rule>
    key level
    pattern ^INFO$
    sample_rate 10
  </rule>

  <rule>
    key level
    pattern ^DEBUG$
    sample_rate 1
  </rule>
</filter>

# 输出配置
<match mall-go.**>
  @type elasticsearch
  @id elasticsearch_output
  host "#{ENV['ELASTICSEARCH_HOST'] || 'elasticsearch'}"
  port "#{ENV['ELASTICSEARCH_PORT'] || 9200}"
  scheme "#{ENV['ELASTICSEARCH_SCHEME'] || 'http'}"
  user "#{ENV['ELASTICSEARCH_USER']}"
  password "#{ENV['ELASTICSEARCH_PASSWORD']}"

  index_name mall-go-logs
  type_name _doc
  time_key @timestamp
  time_key_format %Y-%m-%dT%H:%M:%S.%L%z
  include_timestamp true
  logstash_format true
  logstash_prefix mall-go
  logstash_dateformat %Y.%m.%d

  # 缓冲配置
  <buffer time>
    @type file
    path /var/log/fluentd/buffer/elasticsearch
    timekey 60s
    timekey_wait 10s
    timekey_use_utc true
    chunk_limit_size 32MB
    total_limit_size 1GB
    flush_mode interval
    flush_interval 10s
    flush_thread_count 4
    retry_type exponential_backoff
    retry_wait 1s
    retry_max_interval 60s
    retry_timeout 1h
    overflow_action drop_oldest_chunk
  </buffer>

  # 模板配置
  template_name mall-go-template
  template_file /etc/fluentd/templates/mall-go.json
  template_overwrite true

  # 错误处理
  <secondary>
    @type file
    path /var/log/fluentd/failed_records
    time_slice_format %Y%m%d%H
    time_slice_wait 10m
    time_format %Y-%m-%dT%H:%M:%S%z
    compress gzip
  </secondary>
</match>

# 系统日志输出
<match systemd.**>
  @type elasticsearch
  @id systemd_elasticsearch_output
  host "#{ENV['ELASTICSEARCH_HOST'] || 'elasticsearch'}"
  port "#{ENV['ELASTICSEARCH_PORT'] || 9200}"

  index_name systemd-logs
  logstash_format true
  logstash_prefix systemd
  logstash_dateformat %Y.%m.%d

  <buffer time>
    @type file
    path /var/log/fluentd/buffer/systemd
    timekey 3600s
    timekey_wait 60s
    chunk_limit_size 16MB
    total_limit_size 512MB
    flush_interval 30s
  </buffer>
</match>

# 监控指标输出
<match fluent.**>
  @type prometheus
  @id prometheus_metrics

  <metric>
    name fluentd_input_status_num_records_total
    type counter
    desc The total number of incoming records
    <labels>
      tag ${tag}
      hostname ${hostname}
    </labels>
  </metric>

  <metric>
    name fluentd_output_status_num_records_total
    type counter
    desc The total number of outgoing records
    <labels>
      tag ${tag}
      hostname ${hostname}
    </labels>
  </metric>
</match>
```

### Elasticsearch索引模板

```json
{
  "index_patterns": ["mall-go-*"],
  "settings": {
    "number_of_shards": 3,
    "number_of_replicas": 1,
    "index.refresh_interval": "30s",
    "index.translog.flush_threshold_size": "512mb",
    "index.merge.policy.max_merge_at_once": 5,
    "index.merge.policy.segments_per_tier": 5,
    "index.mapping.total_fields.limit": 2000,
    "index.mapping.depth.limit": 20,
    "index.mapping.nested_fields.limit": 100,
    "index.max_result_window": 50000,
    "index.lifecycle.name": "mall-go-policy",
    "index.lifecycle.rollover_alias": "mall-go-logs"
  },
  "mappings": {
    "properties": {
      "@timestamp": {
        "type": "date",
        "format": "strict_date_optional_time||epoch_millis"
      },
      "level": {
        "type": "keyword",
        "fields": {
          "text": {
            "type": "text"
          }
        }
      },
      "message": {
        "type": "text",
        "analyzer": "standard",
        "fields": {
          "keyword": {
            "type": "keyword",
            "ignore_above": 256
          }
        }
      },
      "service": {
        "type": "keyword"
      },
      "version": {
        "type": "keyword"
      },
      "environment": {
        "type": "keyword"
      },
      "request_id": {
        "type": "keyword"
      },
      "user_id": {
        "type": "keyword"
      },
      "session_id": {
        "type": "keyword"
      },
      "ip": {
        "type": "ip"
      },
      "method": {
        "type": "keyword"
      },
      "url": {
        "type": "keyword",
        "fields": {
          "text": {
            "type": "text"
          }
        }
      },
      "status_code": {
        "type": "integer"
      },
      "duration": {
        "type": "float"
      },
      "order_id": {
        "type": "keyword"
      },
      "product_id": {
        "type": "keyword"
      },
      "category_id": {
        "type": "keyword"
      },
      "payment_id": {
        "type": "keyword"
      },
      "error": {
        "type": "text",
        "analyzer": "standard"
      },
      "stack": {
        "type": "text",
        "analyzer": "keyword"
      },
      "host": {
        "type": "keyword"
      },
      "pod": {
        "type": "keyword"
      },
      "container": {
        "type": "keyword"
      },
      "namespace": {
        "type": "keyword"
      },
      "extra": {
        "type": "object",
        "dynamic": true
      }
    }
  },
  "aliases": {
    "mall-go-logs": {}
  }
}
```

---

## 🔗 分布式链路追踪系统

### Jaeger集成

分布式链路追踪帮助我们理解请求在微服务架构中的完整流程。

```go
// Jaeger链路追踪集成
package tracing

import (
    "context"
    "fmt"
    "io"
    "time"

    "github.com/opentracing/opentracing-go"
    "github.com/opentracing/opentracing-go/ext"
    "github.com/uber/jaeger-client-go"
    "github.com/uber/jaeger-client-go/config"
    "github.com/uber/jaeger-client-go/log"
    "github.com/uber/jaeger-client-go/metrics"
)

// Jaeger配置
type JaegerConfig struct {
    ServiceName     string  `json:"service_name"`
    AgentHost       string  `json:"agent_host"`
    AgentPort       int     `json:"agent_port"`
    CollectorURL    string  `json:"collector_url"`
    SamplingType    string  `json:"sampling_type"`    // const, probabilistic, rateLimiting
    SamplingParam   float64 `json:"sampling_param"`
    LogSpans        bool    `json:"log_spans"`
    MaxTagValueLength int   `json:"max_tag_value_length"`

    // 高级配置
    BufferFlushInterval time.Duration `json:"buffer_flush_interval"`
    QueueSize          int           `json:"queue_size"`
    MaxPacketSize      int           `json:"max_packet_size"`
}

// 链路追踪管理器
type TracingManager struct {
    tracer   opentracing.Tracer
    closer   io.Closer
    config   *JaegerConfig
}

// 创建链路追踪管理器
func NewTracingManager(cfg *JaegerConfig) (*TracingManager, error) {
    // 创建Jaeger配置
    jcfg := config.Configuration{
        ServiceName: cfg.ServiceName,

        // 采样配置
        Sampler: &config.SamplerConfig{
            Type:  cfg.SamplingType,
            Param: cfg.SamplingParam,
        },

        // 报告配置
        Reporter: &config.ReporterConfig{
            LogSpans:            cfg.LogSpans,
            BufferFlushInterval: cfg.BufferFlushInterval,
            LocalAgentHostPort:  fmt.Sprintf("%s:%d", cfg.AgentHost, cfg.AgentPort),
            CollectorEndpoint:   cfg.CollectorURL,
            QueueSize:          cfg.QueueSize,
        },

        // 标签配置
        Tags: []opentracing.Tag{
            {Key: "service.name", Value: cfg.ServiceName},
            {Key: "service.version", Value: "1.0.0"},
            {Key: "deployment.environment", Value: "production"},
        },
    }

    // 创建追踪器
    tracer, closer, err := jcfg.NewTracer(
        config.Logger(jaeger.StdLogger),
        config.Metrics(metrics.NullFactory),
        config.MaxTagValueLength(cfg.MaxTagValueLength),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create tracer: %w", err)
    }

    // 设置全局追踪器
    opentracing.SetGlobalTracer(tracer)

    return &TracingManager{
        tracer: tracer,
        closer: closer,
        config: cfg,
    }, nil
}

// 关闭追踪器
func (tm *TracingManager) Close() error {
    return tm.closer.Close()
}

// 获取追踪器
func (tm *TracingManager) GetTracer() opentracing.Tracer {
    return tm.tracer
}

// HTTP中间件 - 链路追踪
func (tm *TracingManager) HTTPTracingMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 从HTTP头中提取span上下文
        spanCtx, _ := tm.tracer.Extract(
            opentracing.HTTPHeaders,
            opentracing.HTTPHeadersCarrier(c.Request.Header),
        )

        // 创建新的span
        span := tm.tracer.StartSpan(
            fmt.Sprintf("HTTP %s %s", c.Request.Method, c.FullPath()),
            ext.RPCServerOption(spanCtx),
        )
        defer span.Finish()

        // 设置span标签
        ext.HTTPMethod.Set(span, c.Request.Method)
        ext.HTTPUrl.Set(span, c.Request.URL.String())
        ext.Component.Set(span, "gin-http")

        // 将span添加到上下文
        ctx := opentracing.ContextWithSpan(c.Request.Context(), span)
        c.Request = c.Request.WithContext(ctx)

        // 处理请求
        c.Next()

        // 设置响应标签
        ext.HTTPStatusCode.Set(span, uint16(c.Writer.Status()))
        if c.Writer.Status() >= 400 {
            ext.Error.Set(span, true)
            span.SetTag("error.message", c.Errors.String())
        }
    }
}

// 数据库追踪插件
func (tm *TracingManager) DatabaseTracingPlugin() gorm.Plugin {
    return &databaseTracingPlugin{manager: tm}
}

type databaseTracingPlugin struct {
    manager *TracingManager
}

func (p *databaseTracingPlugin) Name() string {
    return "tracing"
}

func (p *databaseTracingPlugin) Initialize(db *gorm.DB) error {
    // 注册回调函数
    db.Callback().Create().Before("gorm:create").Register("tracing:before_create", p.beforeCreate)
    db.Callback().Create().After("gorm:create").Register("tracing:after_create", p.afterCreate)

    db.Callback().Query().Before("gorm:query").Register("tracing:before_query", p.beforeQuery)
    db.Callback().Query().After("gorm:query").Register("tracing:after_query", p.afterQuery)

    db.Callback().Update().Before("gorm:update").Register("tracing:before_update", p.beforeUpdate)
    db.Callback().Update().After("gorm:update").Register("tracing:after_update", p.afterUpdate)

    db.Callback().Delete().Before("gorm:delete").Register("tracing:before_delete", p.beforeDelete)
    db.Callback().Delete().After("gorm:delete").Register("tracing:after_delete", p.afterDelete)

    return nil
}

func (p *databaseTracingPlugin) beforeCreate(db *gorm.DB) {
    p.startSpan(db, "CREATE")
}

func (p *databaseTracingPlugin) afterCreate(db *gorm.DB) {
    p.finishSpan(db)
}

func (p *databaseTracingPlugin) beforeQuery(db *gorm.DB) {
    p.startSpan(db, "SELECT")
}

func (p *databaseTracingPlugin) afterQuery(db *gorm.DB) {
    p.finishSpan(db)
}

func (p *databaseTracingPlugin) beforeUpdate(db *gorm.DB) {
    p.startSpan(db, "UPDATE")
}

func (p *databaseTracingPlugin) afterUpdate(db *gorm.DB) {
    p.finishSpan(db)
}

func (p *databaseTracingPlugin) beforeDelete(db *gorm.DB) {
    p.startSpan(db, "DELETE")
}

func (p *databaseTracingPlugin) afterDelete(db *gorm.DB) {
    p.finishSpan(db)
}

func (p *databaseTracingPlugin) startSpan(db *gorm.DB, operation string) {
    // 从上下文中获取父span
    parentSpan := opentracing.SpanFromContext(db.Statement.Context)
    if parentSpan == nil {
        return
    }

    // 创建数据库操作span
    span := p.manager.tracer.StartSpan(
        fmt.Sprintf("DB %s", operation),
        opentracing.ChildOf(parentSpan.Context()),
    )

    // 设置span标签
    ext.DBType.Set(span, "mysql")
    ext.Component.Set(span, "gorm")
    span.SetTag("db.operation", operation)
    span.SetTag("db.table", db.Statement.Table)

    // 将span保存到上下文
    db.Set("tracing:span", span)
    db.Set("tracing:start_time", time.Now())
}

func (p *databaseTracingPlugin) finishSpan(db *gorm.DB) {
    spanValue, exists := db.Get("tracing:span")
    if !exists {
        return
    }

    span, ok := spanValue.(opentracing.Span)
    if !ok {
        return
    }

    // 设置SQL语句（注意敏感信息过滤）
    if db.Statement.SQL.String() != "" {
        span.SetTag("db.statement", p.sanitizeSQL(db.Statement.SQL.String()))
    }

    // 设置错误信息
    if db.Error != nil {
        ext.Error.Set(span, true)
        span.SetTag("error.message", db.Error.Error())
    }

    // 设置执行时间
    if startTime, exists := db.Get("tracing:start_time"); exists {
        if start, ok := startTime.(time.Time); ok {
            span.SetTag("db.duration", time.Since(start).Milliseconds())
        }
    }

    span.Finish()
}

// 清理SQL语句中的敏感信息
func (p *databaseTracingPlugin) sanitizeSQL(sql string) string {
    // 这里可以实现SQL语句的清理逻辑
    // 例如移除具体的参数值，只保留SQL结构
    return sql
}
```

---

## 🚨 告警系统设计

### AlertManager配置

告警系统是监控体系的重要组成部分，能够及时发现和通知系统异常。

```yaml
# alertmanager.yml - AlertManager主配置文件
global:
  smtp_smarthost: 'smtp.gmail.com:587'
  smtp_from: 'alerts@mall-go.com'
  smtp_auth_username: 'alerts@mall-go.com'
  smtp_auth_password: 'your-app-password'
  smtp_require_tls: true

  # 钉钉机器人配置
  dingtalk_api_url: 'https://oapi.dingtalk.com/robot/send'
  dingtalk_api_secret: 'your-dingtalk-secret'

  # 企业微信配置
  wechat_api_url: 'https://qyapi.weixin.qq.com/cgi-bin'
  wechat_api_corp_id: 'your-corp-id'

# 路由配置
route:
  group_by: ['alertname', 'cluster', 'service']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'default'

  routes:
    # 严重告警 - 立即通知
    - match:
        severity: critical
      receiver: 'critical-alerts'
      group_wait: 0s
      repeat_interval: 5m

    # 警告告警 - 延迟通知
    - match:
        severity: warning
      receiver: 'warning-alerts'
      group_wait: 30s
      repeat_interval: 30m

    # 信息告警 - 批量通知
    - match:
        severity: info
      receiver: 'info-alerts'
      group_wait: 5m
      repeat_interval: 2h

    # 业务告警 - 特殊处理
    - match:
        category: business
      receiver: 'business-alerts'
      group_wait: 1m
      repeat_interval: 15m

# 抑制规则
inhibit_rules:
  # 如果有严重告警，抑制相同服务的警告告警
  - source_match:
      severity: 'critical'
    target_match:
      severity: 'warning'
    equal: ['alertname', 'service', 'instance']

  # 如果服务不可用，抑制该服务的其他告警
  - source_match:
      alertname: 'ServiceDown'
    target_match_re:
      service: '.*'
    equal: ['service']

# 接收器配置
receivers:
  # 默认接收器
  - name: 'default'
    email_configs:
      - to: 'ops@mall-go.com'
        subject: '[Mall-Go] {{ .GroupLabels.alertname }} - {{ .Status | toUpper }}'
        body: |
          {{ range .Alerts }}
          Alert: {{ .Annotations.summary }}
          Description: {{ .Annotations.description }}
          Labels: {{ range .Labels.SortedPairs }}{{ .Name }}={{ .Value }} {{ end }}
          {{ end }}

  # 严重告警接收器
  - name: 'critical-alerts'
    email_configs:
      - to: 'ops@mall-go.com,cto@mall-go.com'
        subject: '🚨 [CRITICAL] {{ .GroupLabels.alertname }}'
        body: |
          严重告警触发！

          {{ range .Alerts }}
          告警名称: {{ .Annotations.summary }}
          告警描述: {{ .Annotations.description }}
          告警时间: {{ .StartsAt.Format "2006-01-02 15:04:05" }}
          告警标签: {{ range .Labels.SortedPairs }}{{ .Name }}={{ .Value }} {{ end }}
          {{ end }}

          请立即处理！

    dingtalk_configs:
      - api_url: 'https://oapi.dingtalk.com/robot/send?access_token=your-token'
        message: |
          🚨 **严重告警**

          {{ range .Alerts }}
          **告警**: {{ .Annotations.summary }}
          **描述**: {{ .Annotations.description }}
          **时间**: {{ .StartsAt.Format "2006-01-02 15:04:05" }}
          **服务**: {{ .Labels.service }}
          **实例**: {{ .Labels.instance }}
          {{ end }}

    webhook_configs:
      - url: 'http://mall-go-alerting:8080/webhook/critical'
        send_resolved: true
        http_config:
          basic_auth:
            username: 'alertmanager'
            password: 'webhook-secret'

  # 警告告警接收器
  - name: 'warning-alerts'
    email_configs:
      - to: 'ops@mall-go.com'
        subject: '⚠️ [WARNING] {{ .GroupLabels.alertname }}'

    dingtalk_configs:
      - api_url: 'https://oapi.dingtalk.com/robot/send?access_token=your-token'
        message: |
          ⚠️ **警告告警**

          {{ range .Alerts }}
          **告警**: {{ .Annotations.summary }}
          **描述**: {{ .Annotations.description }}
          **服务**: {{ .Labels.service }}
          {{ end }}

  # 业务告警接收器
  - name: 'business-alerts'
    email_configs:
      - to: 'business@mall-go.com,ops@mall-go.com'
        subject: '📊 [BUSINESS] {{ .GroupLabels.alertname }}'

    webhook_configs:
      - url: 'http://mall-go-business-monitor:8080/webhook/business'
        send_resolved: true

# 模板配置
templates:
  - '/etc/alertmanager/templates/*.tmpl'
```

### 告警规则定义

```yaml
# alert-rules.yml - Prometheus告警规则
groups:
  # 系统级告警
  - name: system-alerts
    rules:
      # 服务不可用
      - alert: ServiceDown
        expr: up == 0
        for: 1m
        labels:
          severity: critical
          category: system
        annotations:
          summary: "Service {{ $labels.job }} is down"
          description: "Service {{ $labels.job }} on {{ $labels.instance }} has been down for more than 1 minute."

      # CPU使用率过高
      - alert: HighCPUUsage
        expr: 100 - (avg by(instance) (irate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > 80
        for: 5m
        labels:
          severity: warning
          category: system
        annotations:
          summary: "High CPU usage on {{ $labels.instance }}"
          description: "CPU usage is above 80% for more than 5 minutes on {{ $labels.instance }}."

      # 内存使用率过高
      - alert: HighMemoryUsage
        expr: (1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100 > 85
        for: 5m
        labels:
          severity: warning
          category: system
        annotations:
          summary: "High memory usage on {{ $labels.instance }}"
          description: "Memory usage is above 85% for more than 5 minutes on {{ $labels.instance }}."

      # 磁盘空间不足
      - alert: DiskSpaceLow
        expr: (1 - (node_filesystem_avail_bytes / node_filesystem_size_bytes)) * 100 > 90
        for: 2m
        labels:
          severity: critical
          category: system
        annotations:
          summary: "Disk space low on {{ $labels.instance }}"
          description: "Disk usage is above 90% on {{ $labels.instance }} {{ $labels.mountpoint }}."

  # 应用级告警
  - name: application-alerts
    rules:
      # HTTP错误率过高
      - alert: HighHTTPErrorRate
        expr: |
          (
            sum(rate(mall_go_http_requests_total{status_code=~"5.."}[5m])) by (service)
            /
            sum(rate(mall_go_http_requests_total[5m])) by (service)
          ) * 100 > 5
        for: 3m
        labels:
          severity: critical
          category: application
        annotations:
          summary: "High HTTP error rate for {{ $labels.service }}"
          description: "HTTP error rate is above 5% for {{ $labels.service }} for more than 3 minutes."

      # HTTP响应时间过长
      - alert: HighHTTPLatency
        expr: |
          histogram_quantile(0.95,
            sum(rate(mall_go_http_request_duration_seconds_bucket[5m])) by (le, service)
          ) > 2
        for: 5m
        labels:
          severity: warning
          category: application
        annotations:
          summary: "High HTTP latency for {{ $labels.service }}"
          description: "95th percentile latency is above 2 seconds for {{ $labels.service }}."

      # Goroutine数量过多
      - alert: HighGoroutineCount
        expr: go_goroutines > 1000
        for: 5m
        labels:
          severity: warning
          category: application
        annotations:
          summary: "High goroutine count on {{ $labels.instance }}"
          description: "Goroutine count is above 1000 on {{ $labels.instance }}."

      # GC时间过长
      - alert: HighGCDuration
        expr: |
          rate(go_gc_duration_seconds_sum[5m]) / rate(go_gc_duration_seconds_count[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
          category: application
        annotations:
          summary: "High GC duration on {{ $labels.instance }}"
          description: "Average GC duration is above 100ms on {{ $labels.instance }}."

  # 数据库告警
  - name: database-alerts
    rules:
      # 数据库连接数过多
      - alert: HighDatabaseConnections
        expr: mall_go_database_connections{state="active"} > 80
        for: 3m
        labels:
          severity: warning
          category: database
        annotations:
          summary: "High database connections for {{ $labels.database }}"
          description: "Active database connections are above 80 for {{ $labels.database }}."

      # 数据库查询时间过长
      - alert: SlowDatabaseQueries
        expr: |
          histogram_quantile(0.95,
            sum(rate(mall_go_database_query_duration_seconds_bucket[5m])) by (le, database)
          ) > 1
        for: 5m
        labels:
          severity: warning
          category: database
        annotations:
          summary: "Slow database queries for {{ $labels.database }}"
          description: "95th percentile query duration is above 1 second for {{ $labels.database }}."

      # 数据库错误率过高
      - alert: HighDatabaseErrorRate
        expr: |
          (
            sum(rate(mall_go_database_errors_total[5m])) by (database)
            /
            sum(rate(mall_go_database_queries_total[5m])) by (database)
          ) * 100 > 2
        for: 3m
        labels:
          severity: critical
          category: database
        annotations:
          summary: "High database error rate for {{ $labels.database }}"
          description: "Database error rate is above 2% for {{ $labels.database }}."

  # 业务告警
  - name: business-alerts
    rules:
      # 订单创建失败率过高
      - alert: HighOrderFailureRate
        expr: |
          (
            sum(rate(mall_go_business_orders_created_total{status="failed"}[10m]))
            /
            sum(rate(mall_go_business_orders_created_total[10m]))
          ) * 100 > 5
        for: 5m
        labels:
          severity: critical
          category: business
        annotations:
          summary: "High order failure rate"
          description: "Order failure rate is above 5% for more than 5 minutes."

      # 支付失败率过高
      - alert: HighPaymentFailureRate
        expr: |
          (
            sum(rate(mall_go_business_payment_processed_total{status="failed"}[10m]))
            /
            sum(rate(mall_go_business_payment_processed_total[10m]))
          ) * 100 > 3
        for: 3m
        labels:
          severity: critical
          category: business
        annotations:
          summary: "High payment failure rate"
          description: "Payment failure rate is above 3% for more than 3 minutes."

      # 用户注册数量异常
      - alert: LowUserRegistrations
        expr: |
          sum(rate(mall_go_business_user_registrations_total[1h])) < 10
        for: 30m
        labels:
          severity: warning
          category: business
        annotations:
          summary: "Low user registration rate"
          description: "User registration rate is below 10 per hour for more than 30 minutes."

      # 商品浏览量异常
      - alert: LowProductViews
        expr: |
          sum(rate(mall_go_business_product_views_total[1h])) < 100
        for: 30m
        labels:
          severity: info
          category: business
        annotations:
          summary: "Low product view rate"
          description: "Product view rate is below 100 per hour for more than 30 minutes."

  # 基础设施告警
  - name: infrastructure-alerts
    rules:
      # Kubernetes Pod重启过多
      - alert: PodRestartingTooMuch
        expr: increase(kube_pod_container_status_restarts_total[1h]) > 5
        for: 0m
        labels:
          severity: warning
          category: infrastructure
        annotations:
          summary: "Pod {{ $labels.pod }} is restarting too much"
          description: "Pod {{ $labels.pod }} in namespace {{ $labels.namespace }} has restarted more than 5 times in the last hour."

      # Kubernetes节点不可用
      - alert: KubernetesNodeNotReady
        expr: kube_node_status_condition{condition="Ready",status="true"} == 0
        for: 5m
        labels:
          severity: critical
          category: infrastructure
        annotations:
          summary: "Kubernetes node {{ $labels.node }} is not ready"
          description: "Kubernetes node {{ $labels.node }} has been not ready for more than 5 minutes."

      # Elasticsearch集群状态异常
      - alert: ElasticsearchClusterUnhealthy
        expr: elasticsearch_cluster_health_status{color="red"} == 1
        for: 2m
        labels:
          severity: critical
          category: infrastructure
        annotations:
          summary: "Elasticsearch cluster is unhealthy"
          description: "Elasticsearch cluster status is red for more than 2 minutes."
```

---

## 📏 性能指标定义与SLA制定

### 关键性能指标(KPI)定义

```go
// 性能指标定义
package sla

import (
    "time"
)

// 服务等级协议(SLA)定义
type ServiceLevelAgreement struct {
    // 服务基本信息
    ServiceName    string    `json:"service_name"`
    Version        string    `json:"version"`
    Owner          string    `json:"owner"`
    LastUpdated    time.Time `json:"last_updated"`

    // 可用性指标
    Availability   AvailabilityMetrics   `json:"availability"`

    // 性能指标
    Performance    PerformanceMetrics    `json:"performance"`

    // 容量指标
    Capacity       CapacityMetrics       `json:"capacity"`

    // 质量指标
    Quality        QualityMetrics        `json:"quality"`

    // 业务指标
    Business       BusinessMetrics       `json:"business"`
}

// 可用性指标
type AvailabilityMetrics struct {
    // 系统可用性目标
    UptimeTarget        float64 `json:"uptime_target"`        // 99.9%
    UptimeMeasurement   string  `json:"uptime_measurement"`   // monthly

    // 故障恢复时间
    MTTR               time.Duration `json:"mttr"`               // Mean Time To Recovery
    MTTRTarget         time.Duration `json:"mttr_target"`        // 15 minutes

    // 故障间隔时间
    MTBF               time.Duration `json:"mtbf"`               // Mean Time Between Failures
    MTBFTarget         time.Duration `json:"mtbf_target"`        // 30 days

    // 计划内停机时间
    PlannedDowntime    time.Duration `json:"planned_downtime"`   // per month
    PlannedDowntimeMax time.Duration `json:"planned_downtime_max"` // 4 hours
}

// 性能指标
type PerformanceMetrics struct {
    // 响应时间指标
    ResponseTime struct {
        P50Target    time.Duration `json:"p50_target"`    // 100ms
        P95Target    time.Duration `json:"p95_target"`    // 500ms
        P99Target    time.Duration `json:"p99_target"`    // 1000ms
        P999Target   time.Duration `json:"p999_target"`   // 2000ms

        P50Current   time.Duration `json:"p50_current"`
        P95Current   time.Duration `json:"p95_current"`
        P99Current   time.Duration `json:"p99_current"`
        P999Current  time.Duration `json:"p999_current"`
    } `json:"response_time"`

    // 吞吐量指标
    Throughput struct {
        RPSTarget    float64 `json:"rps_target"`    // Requests Per Second
        RPSCurrent   float64 `json:"rps_current"`
        QPSTarget    float64 `json:"qps_target"`    // Queries Per Second
        QPSCurrent   float64 `json:"qps_current"`
        TPSTarget    float64 `json:"tps_target"`    // Transactions Per Second
        TPSCurrent   float64 `json:"tps_current"`
    } `json:"throughput"`

    // 并发指标
    Concurrency struct {
        MaxConcurrentUsers    int `json:"max_concurrent_users"`
        MaxConcurrentRequests int `json:"max_concurrent_requests"`
        CurrentConcurrency    int `json:"current_concurrency"`
    } `json:"concurrency"`
}

// 容量指标
type CapacityMetrics struct {
    // 资源使用率
    ResourceUtilization struct {
        CPUTarget      float64 `json:"cpu_target"`      // 70%
        CPUCurrent     float64 `json:"cpu_current"`
        MemoryTarget   float64 `json:"memory_target"`   // 80%
        MemoryCurrent  float64 `json:"memory_current"`
        DiskTarget     float64 `json:"disk_target"`     // 85%
        DiskCurrent    float64 `json:"disk_current"`
        NetworkTarget  float64 `json:"network_target"`  // 70%
        NetworkCurrent float64 `json:"network_current"`
    } `json:"resource_utilization"`

    // 扩展性指标
    Scalability struct {
        MinInstances     int `json:"min_instances"`
        MaxInstances     int `json:"max_instances"`
        CurrentInstances int `json:"current_instances"`
        AutoScaling      bool `json:"auto_scaling"`
        ScaleUpThreshold float64 `json:"scale_up_threshold"`   // 80%
        ScaleDownThreshold float64 `json:"scale_down_threshold"` // 30%
    } `json:"scalability"`

    // 存储容量
    Storage struct {
        DatabaseSizeLimit   int64 `json:"database_size_limit"`   // bytes
        DatabaseSizeCurrent int64 `json:"database_size_current"`
        LogRetentionDays    int   `json:"log_retention_days"`    // 30 days
        BackupRetentionDays int   `json:"backup_retention_days"` // 90 days
    } `json:"storage"`
}

// 质量指标
type QualityMetrics struct {
    // 错误率指标
    ErrorRate struct {
        HTTPErrorRateTarget    float64 `json:"http_error_rate_target"`    // 0.1%
        HTTPErrorRateCurrent   float64 `json:"http_error_rate_current"`
        DatabaseErrorRateTarget float64 `json:"database_error_rate_target"` // 0.01%
        DatabaseErrorRateCurrent float64 `json:"database_error_rate_current"`
        SystemErrorRateTarget   float64 `json:"system_error_rate_target"`   // 0.05%
        SystemErrorRateCurrent  float64 `json:"system_error_rate_current"`
    } `json:"error_rate"`

    // 数据质量
    DataQuality struct {
        DataAccuracyTarget     float64 `json:"data_accuracy_target"`     // 99.9%
        DataAccuracyCurrent    float64 `json:"data_accuracy_current"`
        DataCompletenessTarget float64 `json:"data_completeness_target"` // 99.5%
        DataCompletenessCurrent float64 `json:"data_completeness_current"`
        DataConsistencyTarget  float64 `json:"data_consistency_target"`  // 99.9%
        DataConsistencyCurrent float64 `json:"data_consistency_current"`
    } `json:"data_quality"`

    // 安全指标
    Security struct {
        SecurityIncidentsTarget int `json:"security_incidents_target"` // 0 per month
        SecurityIncidentsCurrent int `json:"security_incidents_current"`
        VulnerabilitiesTarget   int `json:"vulnerabilities_target"`   // 0 critical
        VulnerabilitiesCurrent  int `json:"vulnerabilities_current"`
        ComplianceScore         float64 `json:"compliance_score"`      // 100%
    } `json:"security"`
}

// 业务指标
type BusinessMetrics struct {
    // 用户体验指标
    UserExperience struct {
        UserSatisfactionTarget float64 `json:"user_satisfaction_target"` // 4.5/5.0
        UserSatisfactionCurrent float64 `json:"user_satisfaction_current"`
        PageLoadTimeTarget     time.Duration `json:"page_load_time_target"` // 2s
        PageLoadTimeCurrent    time.Duration `json:"page_load_time_current"`
        ConversionRateTarget   float64 `json:"conversion_rate_target"`   // 3%
        ConversionRateCurrent  float64 `json:"conversion_rate_current"`
    } `json:"user_experience"`

    // 业务流程指标
    BusinessProcess struct {
        OrderProcessingTimeTarget time.Duration `json:"order_processing_time_target"` // 5 minutes
        OrderProcessingTimeCurrent time.Duration `json:"order_processing_time_current"`
        PaymentSuccessRateTarget  float64 `json:"payment_success_rate_target"`  // 99.5%
        PaymentSuccessRateCurrent float64 `json:"payment_success_rate_current"`
        InventoryAccuracyTarget   float64 `json:"inventory_accuracy_target"`   // 99.9%
        InventoryAccuracyCurrent  float64 `json:"inventory_accuracy_current"`
    } `json:"business_process"`

    // 成本指标
    Cost struct {
        InfrastructureCostTarget float64 `json:"infrastructure_cost_target"` // per month
        InfrastructureCostCurrent float64 `json:"infrastructure_cost_current"`
        OperationalCostTarget    float64 `json:"operational_cost_target"`    // per month
        OperationalCostCurrent   float64 `json:"operational_cost_current"`
        CostPerTransactionTarget float64 `json:"cost_per_transaction_target"`
        CostPerTransactionCurrent float64 `json:"cost_per_transaction_current"`
    } `json:"cost"`
}

// SLA管理器
type SLAManager struct {
    agreements map[string]*ServiceLevelAgreement
    metrics    MetricsCollector
    alerting   AlertingService
}

// 创建SLA管理器
func NewSLAManager(metrics MetricsCollector, alerting AlertingService) *SLAManager {
    return &SLAManager{
        agreements: make(map[string]*ServiceLevelAgreement),
        metrics:    metrics,
        alerting:   alerting,
    }
}

// 注册SLA
func (sm *SLAManager) RegisterSLA(sla *ServiceLevelAgreement) {
    sm.agreements[sla.ServiceName] = sla
}

// 检查SLA合规性
func (sm *SLAManager) CheckCompliance(serviceName string) (*SLAComplianceReport, error) {
    sla, exists := sm.agreements[serviceName]
    if !exists {
        return nil, fmt.Errorf("SLA not found for service: %s", serviceName)
    }

    report := &SLAComplianceReport{
        ServiceName: serviceName,
        CheckTime:   time.Now(),
        Violations:  make([]SLAViolation, 0),
    }

    // 检查可用性
    if err := sm.checkAvailability(sla, report); err != nil {
        return nil, err
    }

    // 检查性能
    if err := sm.checkPerformance(sla, report); err != nil {
        return nil, err
    }

    // 检查质量
    if err := sm.checkQuality(sla, report); err != nil {
        return nil, err
    }

    // 计算总体合规性
    report.OverallCompliance = sm.calculateOverallCompliance(report)

    return report, nil
}

// SLA合规性报告
type SLAComplianceReport struct {
    ServiceName       string         `json:"service_name"`
    CheckTime         time.Time      `json:"check_time"`
    OverallCompliance float64        `json:"overall_compliance"`
    Violations        []SLAViolation `json:"violations"`
}

// SLA违规记录
type SLAViolation struct {
    MetricName    string    `json:"metric_name"`
    TargetValue   float64   `json:"target_value"`
    CurrentValue  float64   `json:"current_value"`
    Severity      string    `json:"severity"`
    Description   string    `json:"description"`
    DetectedAt    time.Time `json:"detected_at"`
}

// Mall-Go SLA配置示例
func CreateMallGoSLA() *ServiceLevelAgreement {
    return &ServiceLevelAgreement{
        ServiceName: "mall-go",
        Version:     "1.0.0",
        Owner:       "Platform Team",
        LastUpdated: time.Now(),

        Availability: AvailabilityMetrics{
            UptimeTarget:       99.9,  // 99.9% 可用性
            UptimeMeasurement:  "monthly",
            MTTRTarget:         15 * time.Minute,  // 15分钟恢复时间
            MTBFTarget:         30 * 24 * time.Hour, // 30天故障间隔
            PlannedDowntimeMax: 4 * time.Hour,     // 每月最多4小时计划停机
        },

        Performance: PerformanceMetrics{
            ResponseTime: struct {
                P50Target    time.Duration `json:"p50_target"`
                P95Target    time.Duration `json:"p95_target"`
                P99Target    time.Duration `json:"p99_target"`
                P999Target   time.Duration `json:"p999_target"`
                P50Current   time.Duration `json:"p50_current"`
                P95Current   time.Duration `json:"p95_current"`
                P99Current   time.Duration `json:"p99_current"`
                P999Current  time.Duration `json:"p999_current"`
            }{
                P50Target:  100 * time.Millisecond,  // 50%请求100ms内响应
                P95Target:  500 * time.Millisecond,  // 95%请求500ms内响应
                P99Target:  1000 * time.Millisecond, // 99%请求1s内响应
                P999Target: 2000 * time.Millisecond, // 99.9%请求2s内响应
            },
            Throughput: struct {
                RPSTarget    float64 `json:"rps_target"`
                RPSCurrent   float64 `json:"rps_current"`
                QPSTarget    float64 `json:"qps_target"`
                QPSCurrent   float64 `json:"qps_current"`
                TPSTarget    float64 `json:"tps_target"`
                TPSCurrent   float64 `json:"tps_current"`
            }{
                RPSTarget: 1000,  // 支持1000 RPS
                QPSTarget: 5000,  // 支持5000 QPS
                TPSTarget: 500,   // 支持500 TPS
            },
            Concurrency: struct {
                MaxConcurrentUsers    int `json:"max_concurrent_users"`
                MaxConcurrentRequests int `json:"max_concurrent_requests"`
                CurrentConcurrency    int `json:"current_concurrency"`
            }{
                MaxConcurrentUsers:    10000, // 支持1万并发用户
                MaxConcurrentRequests: 5000,  // 支持5000并发请求
            },
        },

        Capacity: CapacityMetrics{
            ResourceUtilization: struct {
                CPUTarget      float64 `json:"cpu_target"`
                CPUCurrent     float64 `json:"cpu_current"`
                MemoryTarget   float64 `json:"memory_target"`
                MemoryCurrent  float64 `json:"memory_current"`
                DiskTarget     float64 `json:"disk_target"`
                DiskCurrent    float64 `json:"disk_current"`
                NetworkTarget  float64 `json:"network_target"`
                NetworkCurrent float64 `json:"network_current"`
            }{
                CPUTarget:    70.0, // CPU使用率不超过70%
                MemoryTarget: 80.0, // 内存使用率不超过80%
                DiskTarget:   85.0, // 磁盘使用率不超过85%
                NetworkTarget: 70.0, // 网络使用率不超过70%
            },
            Scalability: struct {
                MinInstances     int `json:"min_instances"`
                MaxInstances     int `json:"max_instances"`
                CurrentInstances int `json:"current_instances"`
                AutoScaling      bool `json:"auto_scaling"`
                ScaleUpThreshold float64 `json:"scale_up_threshold"`
                ScaleDownThreshold float64 `json:"scale_down_threshold"`
            }{
                MinInstances:       3,    // 最少3个实例
                MaxInstances:       20,   // 最多20个实例
                AutoScaling:        true, // 启用自动扩缩容
                ScaleUpThreshold:   80.0, // 80%时扩容
                ScaleDownThreshold: 30.0, // 30%时缩容
            },
        },

        Quality: QualityMetrics{
            ErrorRate: struct {
                HTTPErrorRateTarget    float64 `json:"http_error_rate_target"`
                HTTPErrorRateCurrent   float64 `json:"http_error_rate_current"`
                DatabaseErrorRateTarget float64 `json:"database_error_rate_target"`
                DatabaseErrorRateCurrent float64 `json:"database_error_rate_current"`
                SystemErrorRateTarget   float64 `json:"system_error_rate_target"`
                SystemErrorRateCurrent  float64 `json:"system_error_rate_current"`
            }{
                HTTPErrorRateTarget:    0.1,  // HTTP错误率不超过0.1%
                DatabaseErrorRateTarget: 0.01, // 数据库错误率不超过0.01%
                SystemErrorRateTarget:  0.05, // 系统错误率不超过0.05%
            },
            DataQuality: struct {
                DataAccuracyTarget     float64 `json:"data_accuracy_target"`
                DataAccuracyCurrent    float64 `json:"data_accuracy_current"`
                DataCompletenessTarget float64 `json:"data_completeness_target"`
                DataCompletenessCurrent float64 `json:"data_completeness_current"`
                DataConsistencyTarget  float64 `json:"data_consistency_target"`
                DataConsistencyCurrent float64 `json:"data_consistency_current"`
            }{
                DataAccuracyTarget:     99.9, // 数据准确率99.9%
                DataCompletenessTarget: 99.5, // 数据完整性99.5%
                DataConsistencyTarget:  99.9, // 数据一致性99.9%
            },
        },

        Business: BusinessMetrics{
            UserExperience: struct {
                UserSatisfactionTarget float64 `json:"user_satisfaction_target"`
                UserSatisfactionCurrent float64 `json:"user_satisfaction_current"`
                PageLoadTimeTarget     time.Duration `json:"page_load_time_target"`
                PageLoadTimeCurrent    time.Duration `json:"page_load_time_current"`
                ConversionRateTarget   float64 `json:"conversion_rate_target"`
                ConversionRateCurrent  float64 `json:"conversion_rate_current"`
            }{
                UserSatisfactionTarget: 4.5,                    // 用户满意度4.5/5.0
                PageLoadTimeTarget:     2 * time.Second,        // 页面加载时间2秒
                ConversionRateTarget:   3.0,                    // 转化率3%
            },
            BusinessProcess: struct {
                OrderProcessingTimeTarget time.Duration `json:"order_processing_time_target"`
                OrderProcessingTimeCurrent time.Duration `json:"order_processing_time_current"`
                PaymentSuccessRateTarget  float64 `json:"payment_success_rate_target"`
                PaymentSuccessRateCurrent float64 `json:"payment_success_rate_current"`
                InventoryAccuracyTarget   float64 `json:"inventory_accuracy_target"`
                InventoryAccuracyCurrent  float64 `json:"inventory_accuracy_current"`
            }{
                OrderProcessingTimeTarget: 5 * time.Minute,     // 订单处理时间5分钟
                PaymentSuccessRateTarget:  99.5,                // 支付成功率99.5%
                InventoryAccuracyTarget:   99.9,                // 库存准确率99.9%
            },
        },
    }
}
```

---

## 🏢 Mall-Go监控实践案例

### 完整监控架构实现

```go
// Mall-Go完整监控系统实现
package monitoring

import (
    "context"
    "fmt"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "go.uber.org/zap"
    "gorm.io/gorm"
)

// Mall-Go监控系统
type MallGoMonitoringSystem struct {
    // 指标收集
    metricsCollector *BusinessMetricsCollector
    runtimeCollector *GoRuntimeCollector
    customMetrics    *CustomMetricsManager

    // 日志系统
    logger           *StructuredLogger

    // 链路追踪
    tracingManager   *TracingManager

    // SLA管理
    slaManager       *SLAManager

    // 配置
    config           *MonitoringConfig
}

// 监控配置
type MonitoringConfig struct {
    // Prometheus配置
    Prometheus struct {
        Enabled     bool   `json:"enabled"`
        Port        int    `json:"port"`
        Path        string `json:"path"`
        Namespace   string `json:"namespace"`
    } `json:"prometheus"`

    // Jaeger配置
    Jaeger struct {
        Enabled      bool    `json:"enabled"`
        AgentHost    string  `json:"agent_host"`
        AgentPort    int     `json:"agent_port"`
        SamplingRate float64 `json:"sampling_rate"`
    } `json:"jaeger"`

    // 日志配置
    Logging struct {
        Level       string `json:"level"`
        Format      string `json:"format"`
        Output      string `json:"output"`
        Structured  bool   `json:"structured"`
    } `json:"logging"`

    // 告警配置
    Alerting struct {
        Enabled     bool     `json:"enabled"`
        Webhooks    []string `json:"webhooks"`
        EmailTo     []string `json:"email_to"`
        DingTalkURL string   `json:"dingtalk_url"`
    } `json:"alerting"`
}

// 创建Mall-Go监控系统
func NewMallGoMonitoringSystem(config *MonitoringConfig, db *gorm.DB) (*MallGoMonitoringSystem, error) {
    // 创建指标收集器
    mallGoMetrics := NewMallGoMetrics()
    metricsCollector := NewBusinessMetricsCollector(mallGoMetrics, db)
    runtimeCollector := NewGoRuntimeCollector(10 * time.Second)
    customMetrics := NewCustomMetricsManager()

    // 创建结构化日志记录器
    logger := NewStructuredLogger(
        "mall-go",
        "1.0.0",
        "production",
        "localhost",
        &ConsoleLogOutput{},
    )

    // 创建链路追踪管理器
    var tracingManager *TracingManager
    if config.Jaeger.Enabled {
        jaegerConfig := &JaegerConfig{
            ServiceName:   "mall-go",
            AgentHost:     config.Jaeger.AgentHost,
            AgentPort:     config.Jaeger.AgentPort,
            SamplingType:  "probabilistic",
            SamplingParam: config.Jaeger.SamplingRate,
            LogSpans:      true,
        }

        var err error
        tracingManager, err = NewTracingManager(jaegerConfig)
        if err != nil {
            return nil, fmt.Errorf("failed to create tracing manager: %w", err)
        }
    }

    // 创建SLA管理器
    slaManager := NewSLAManager(metricsCollector, nil)
    slaManager.RegisterSLA(CreateMallGoSLA())

    return &MallGoMonitoringSystem{
        metricsCollector: metricsCollector,
        runtimeCollector: runtimeCollector,
        customMetrics:    customMetrics,
        logger:           logger,
        tracingManager:   tracingManager,
        slaManager:       slaManager,
        config:           config,
    }, nil
}

// 启动监控系统
func (ms *MallGoMonitoringSystem) Start(ctx context.Context) error {
    // 启动运行时指标收集
    go ms.runtimeCollector.Start(ctx)

    // 启动Prometheus指标服务器
    if ms.config.Prometheus.Enabled {
        go ms.startPrometheusServer()
    }

    // 启动SLA检查
    go ms.startSLAMonitoring(ctx)

    ms.logger.Info(ctx, "Mall-Go monitoring system started", map[string]interface{}{
        "prometheus_enabled": ms.config.Prometheus.Enabled,
        "jaeger_enabled":     ms.config.Jaeger.Enabled,
        "logging_level":      ms.config.Logging.Level,
    })

    return nil
}

// 启动Prometheus指标服务器
func (ms *MallGoMonitoringSystem) startPrometheusServer() {
    mux := http.NewServeMux()
    mux.Handle(ms.config.Prometheus.Path, promhttp.Handler())

    // 健康检查端点
    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })

    // 就绪检查端点
    mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
        // 检查各个组件状态
        if ms.isSystemReady() {
            w.WriteHeader(http.StatusOK)
            w.Write([]byte("Ready"))
        } else {
            w.WriteHeader(http.StatusServiceUnavailable)
            w.Write([]byte("Not Ready"))
        }
    })

    server := &http.Server{
        Addr:    fmt.Sprintf(":%d", ms.config.Prometheus.Port),
        Handler: mux,
    }

    if err := server.ListenAndServe(); err != nil {
        ms.logger.Error(context.Background(), "Prometheus server failed", err, nil)
    }
}

// 检查系统就绪状态
func (ms *MallGoMonitoringSystem) isSystemReady() bool {
    // 检查数据库连接
    // 检查缓存连接
    // 检查外部服务连接
    return true
}

// 启动SLA监控
func (ms *MallGoMonitoringSystem) startSLAMonitoring(ctx context.Context) {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            ms.checkSLACompliance(ctx)
        }
    }
}

// 检查SLA合规性
func (ms *MallGoMonitoringSystem) checkSLACompliance(ctx context.Context) {
    report, err := ms.slaManager.CheckCompliance("mall-go")
    if err != nil {
        ms.logger.Error(ctx, "Failed to check SLA compliance", err, nil)
        return
    }

    ms.logger.Info(ctx, "SLA compliance check completed", map[string]interface{}{
        "service":            report.ServiceName,
        "overall_compliance": report.OverallCompliance,
        "violations_count":   len(report.Violations),
    })

    // 如果有违规，发送告警
    if len(report.Violations) > 0 {
        ms.handleSLAViolations(ctx, report)
    }
}

// 处理SLA违规
func (ms *MallGoMonitoringSystem) handleSLAViolations(ctx context.Context, report *SLAComplianceReport) {
    for _, violation := range report.Violations {
        ms.logger.Warn(ctx, "SLA violation detected", map[string]interface{}{
            "metric":        violation.MetricName,
            "target_value":  violation.TargetValue,
            "current_value": violation.CurrentValue,
            "severity":      violation.Severity,
            "description":   violation.Description,
        })

        // 记录违规指标
        ms.customMetrics.RecordErrorRate("sla", violation.Severity, 1.0)
    }
}

// 获取Gin中间件
func (ms *MallGoMonitoringSystem) GetGinMiddlewares() []gin.HandlerFunc {
    middlewares := make([]gin.HandlerFunc, 0)

    // HTTP指标中间件
    middlewares = append(middlewares, ms.metricsCollector.HTTPMetricsMiddleware())

    // 链路追踪中间件
    if ms.tracingManager != nil {
        middlewares = append(middlewares, ms.tracingManager.HTTPTracingMiddleware())
    }

    // 日志中间件
    middlewares = append(middlewares, ms.createLoggingMiddleware())

    return middlewares
}

// 创建日志中间件
func (ms *MallGoMonitoringSystem) createLoggingMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()

        // 生成请求ID
        requestID := generateRequestID()
        c.Set("request_id", requestID)

        // 添加请求ID到上下文
        ctx := context.WithValue(c.Request.Context(), "request_id", requestID)
        c.Request = c.Request.WithContext(ctx)

        // 记录请求开始
        ms.logger.Info(ctx, "Request started", map[string]interface{}{
            "method":     c.Request.Method,
            "url":        c.Request.URL.String(),
            "user_agent": c.Request.UserAgent(),
            "ip":         c.ClientIP(),
        })

        // 处理请求
        c.Next()

        // 记录请求结束
        duration := time.Since(start)
        ms.logger.Info(ctx, "Request completed", map[string]interface{}{
            "method":      c.Request.Method,
            "url":         c.Request.URL.String(),
            "status_code": c.Writer.Status(),
            "duration":    duration.Milliseconds(),
            "size":        c.Writer.Size(),
        })

        // 记录错误
        if len(c.Errors) > 0 {
            ms.logger.Error(ctx, "Request errors", c.Errors.Last(), map[string]interface{}{
                "errors": c.Errors.String(),
            })
        }
    }
}

// 获取GORM插件
func (ms *MallGoMonitoringSystem) GetGORMPlugins() []gorm.Plugin {
    plugins := make([]gorm.Plugin, 0)

    // 数据库指标插件
    plugins = append(plugins, ms.metricsCollector.DatabaseMetricsPlugin())

    // 链路追踪插件
    if ms.tracingManager != nil {
        plugins = append(plugins, ms.tracingManager.DatabaseTracingPlugin())
    }

    return plugins
}

// 业务指标记录方法
func (ms *MallGoMonitoringSystem) RecordUserRegistration(ctx context.Context, userID, source, userType string) {
    ms.metricsCollector.RecordUserRegistration(source, userType)
    ms.logger.Info(ctx, "User registered", map[string]interface{}{
        "user_id":   userID,
        "source":    source,
        "user_type": userType,
    })
}

func (ms *MallGoMonitoringSystem) RecordOrderCreated(ctx context.Context, orderID, userID, category, paymentMethod string, amount float64) {
    ms.metricsCollector.RecordOrderCreated(category, paymentMethod)
    ms.logger.Info(ctx, "Order created", map[string]interface{}{
        "order_id":       orderID,
        "user_id":        userID,
        "category":       category,
        "payment_method": paymentMethod,
        "amount":         amount,
    })
}

func (ms *MallGoMonitoringSystem) RecordPaymentProcessed(ctx context.Context, paymentID, orderID, method, currency, status string, amount float64) {
    ms.metricsCollector.RecordPaymentProcessed(method, currency, status, amount)
    ms.logger.Info(ctx, "Payment processed", map[string]interface{}{
        "payment_id": paymentID,
        "order_id":   orderID,
        "method":     method,
        "currency":   currency,
        "status":     status,
        "amount":     amount,
    })
}

// 控制台日志输出
type ConsoleLogOutput struct{}

func (c *ConsoleLogOutput) Write(entry *MallGoLogEntry) error {
    // 这里可以实现具体的日志输出逻辑
    // 例如输出到控制台、文件或发送到日志聚合系统
    fmt.Printf("[%s] %s - %s\n", entry.Timestamp.Format(time.RFC3339), entry.Level, entry.Message)
    return nil
}

func (c *ConsoleLogOutput) Close() error {
    return nil
}

// 生成请求ID
func generateRequestID() string {
    return fmt.Sprintf("%d", time.Now().UnixNano())
}
```

### Docker Compose监控栈

```yaml
# docker-compose.monitoring.yml - 完整监控栈
version: '3.8'

services:
  # Mall-Go应用
  mall-go:
    build: .
    ports:
      - "8080:8080"
      - "9090:9090"  # Prometheus metrics
    environment:
      - ENVIRONMENT=production
      - PROMETHEUS_ENABLED=true
      - JAEGER_ENABLED=true
      - JAEGER_AGENT_HOST=jaeger
    depends_on:
      - mysql
      - redis
      - jaeger
    networks:
      - monitoring
    labels:
      - "prometheus.io/scrape=true"
      - "prometheus.io/port=9090"
      - "prometheus.io/path=/metrics"

  # MySQL数据库
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword
      MYSQL_DATABASE: mall_go
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
    networks:
      - monitoring

  # Redis缓存
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    networks:
      - monitoring

  # Prometheus监控
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9091:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
      - ./monitoring/rules:/etc/prometheus/rules
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--storage.tsdb.retention.time=30d'
      - '--web.enable-lifecycle'
      - '--web.enable-admin-api'
    networks:
      - monitoring

  # Grafana可视化
  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_USERS_ALLOW_SIGN_UP=false
    volumes:
      - grafana_data:/var/lib/grafana
      - ./monitoring/grafana/dashboards:/etc/grafana/provisioning/dashboards
      - ./monitoring/grafana/datasources:/etc/grafana/provisioning/datasources
    networks:
      - monitoring

  # AlertManager告警
  alertmanager:
    image: prom/alertmanager:latest
    ports:
      - "9093:9093"
    volumes:
      - ./monitoring/alertmanager.yml:/etc/alertmanager/alertmanager.yml
      - alertmanager_data:/alertmanager
    command:
      - '--config.file=/etc/alertmanager/alertmanager.yml'
      - '--storage.path=/alertmanager'
      - '--web.external-url=http://localhost:9093'
    networks:
      - monitoring

  # Node Exporter系统监控
  node-exporter:
    image: prom/node-exporter:latest
    ports:
      - "9100:9100"
    volumes:
      - /proc:/host/proc:ro
      - /sys:/host/sys:ro
      - /:/rootfs:ro
    command:
      - '--path.procfs=/host/proc'
      - '--path.rootfs=/rootfs'
      - '--path.sysfs=/host/sys'
      - '--collector.filesystem.mount-points-exclude=^/(sys|proc|dev|host|etc)($$|/)'
    networks:
      - monitoring

  # MySQL Exporter
  mysql-exporter:
    image: prom/mysqld-exporter:latest
    ports:
      - "9104:9104"
    environment:
      - DATA_SOURCE_NAME=root:rootpassword@(mysql:3306)/
    depends_on:
      - mysql
    networks:
      - monitoring

  # Redis Exporter
  redis-exporter:
    image: oliver006/redis_exporter:latest
    ports:
      - "9121:9121"
    environment:
      - REDIS_ADDR=redis:6379
    depends_on:
      - redis
    networks:
      - monitoring

  # Elasticsearch日志存储
  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.8.0
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=false
      - "ES_JAVA_OPTS=-Xms512m -Xmx512m"
    ports:
      - "9200:9200"
    volumes:
      - elasticsearch_data:/usr/share/elasticsearch/data
    networks:
      - monitoring

  # Kibana日志可视化
  kibana:
    image: docker.elastic.co/kibana/kibana:8.8.0
    ports:
      - "5601:5601"
    environment:
      - ELASTICSEARCH_HOSTS=http://elasticsearch:9200
    depends_on:
      - elasticsearch
    networks:
      - monitoring

  # Fluentd日志收集
  fluentd:
    build: ./monitoring/fluentd
    ports:
      - "24224:24224"
      - "24224:24224/udp"
    volumes:
      - ./monitoring/fluentd/conf:/fluentd/etc
      - /var/log:/var/log:ro
      - /var/lib/docker/containers:/var/lib/docker/containers:ro
    depends_on:
      - elasticsearch
    networks:
      - monitoring

  # Jaeger链路追踪
  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "16686:16686"  # Jaeger UI
      - "14268:14268"  # Jaeger collector
      - "6831:6831/udp"  # Jaeger agent
    environment:
      - COLLECTOR_OTLP_ENABLED=true
    networks:
      - monitoring

volumes:
  mysql_data:
  prometheus_data:
  grafana_data:
  alertmanager_data:
  elasticsearch_data:

networks:
  monitoring:
    driver: bridge
```

---

## 🎯 面试常考知识点

### 1. 监控系统设计原理

**Q: 请解释可观测性的三大支柱及其作用？**

**A: 可观测性的三大支柱包括：**

1. **指标(Metrics)** 📊
   - **作用**: 提供系统运行状态的数值化度量
   - **特点**: 时间序列数据，支持聚合和告警
   - **示例**: CPU使用率、内存使用量、请求QPS、错误率
   - **优势**: 存储成本低，查询速度快，适合长期趋势分析

2. **日志(Logs)** 📋
   - **作用**: 记录系统运行过程中的离散事件
   - **特点**: 结构化或非结构化文本，包含详细上下文信息
   - **示例**: 错误日志、访问日志、业务操作日志
   - **优势**: 信息详细，便于问题定位和根因分析

3. **链路追踪(Traces)** 🔗
   - **作用**: 跟踪请求在分布式系统中的完整调用链路
   - **特点**: 包含时间关系和因果关系的调用图
   - **示例**: HTTP请求跨服务调用、数据库查询、缓存操作
   - **优势**: 直观展示请求流程，快速定位性能瓶颈

**Q: Prometheus的数据模型和存储原理是什么？**

**A: Prometheus数据模型特点：**

```go
// Prometheus指标数据模型
type PrometheusMetric struct {
    // 指标名称
    MetricName string `json:"metric_name"`

    // 标签集合（维度）
    Labels map[string]string `json:"labels"`

    // 时间戳
    Timestamp int64 `json:"timestamp"`

    // 数值
    Value float64 `json:"value"`
}

// 示例：HTTP请求指标
// http_requests_total{method="GET", endpoint="/api/users", status="200"} 1234 @1640995200
```

**存储原理：**
- **时间序列数据库**: 基于时间戳的键值存储
- **标签索引**: 使用倒排索引快速查找时间序列
- **数据压缩**: 使用差值编码和变长编码减少存储空间
- **分块存储**: 按时间范围分块，支持高效查询和压缩

### 2. 分布式链路追踪

**Q: 分布式链路追踪的核心概念有哪些？**

**A: 核心概念包括：**

1. **Trace（链路）**: 一个完整的请求调用链
2. **Span（跨度）**: 链路中的一个操作单元
3. **SpanContext（跨度上下文）**: 跨服务传递的追踪信息
4. **Baggage（行李）**: 跨服务传递的业务数据

```go
// Span结构示例
type Span struct {
    TraceID    string            `json:"trace_id"`
    SpanID     string            `json:"span_id"`
    ParentID   string            `json:"parent_id"`
    Operation  string            `json:"operation"`
    StartTime  time.Time         `json:"start_time"`
    Duration   time.Duration     `json:"duration"`
    Tags       map[string]string `json:"tags"`
    Logs       []LogEntry        `json:"logs"`
}
```

**Q: 如何在Go中实现分布式链路追踪？**

**A: 实现步骤：**

```go
// 1. 创建根Span
span := tracer.StartSpan("http_request")
defer span.Finish()

// 2. 设置标签
span.SetTag("http.method", "GET")
span.SetTag("http.url", "/api/users")

// 3. 传递上下文
ctx := opentracing.ContextWithSpan(context.Background(), span)

// 4. 创建子Span
childSpan, ctx := opentracing.StartSpanFromContext(ctx, "database_query")
defer childSpan.Finish()

// 5. 跨服务传递
carrier := opentracing.HTTPHeadersCarrier(req.Header)
tracer.Inject(span.Context(), opentracing.HTTPHeaders, carrier)
```

### 3. 日志聚合系统

**Q: ELK和EFK的区别是什么？各自的优缺点？**

**A: 对比分析：**

| 特性 | ELK (Logstash) | EFK (Fluentd) |
|------|----------------|---------------|
| **内存使用** | 较高 (JVM) | 较低 (Ruby) |
| **性能** | 高吞吐量 | 中等吞吐量 |
| **插件生态** | 丰富 | 非常丰富 |
| **配置复杂度** | 中等 | 简单 |
| **社区支持** | Elastic官方 | CNCF项目 |
| **适用场景** | 大数据量处理 | 云原生环境 |

**Q: 如何设计高可用的日志聚合架构？**

**A: 高可用设计要点：**

```yaml
# 多层缓冲架构
Application -> Filebeat -> Kafka -> Logstash -> Elasticsearch

# 关键配置
1. 缓冲队列: 防止数据丢失
2. 多副本: 保证数据可靠性
3. 负载均衡: 分散处理压力
4. 监控告警: 及时发现问题
5. 数据备份: 定期备份重要日志
```

### 4. 告警系统设计

**Q: 如何设计一个有效的告警系统？**

**A: 设计原则：**

1. **分级告警**: 根据严重程度分级
   - Critical: 立即处理
   - Warning: 需要关注
   - Info: 信息通知

2. **告警抑制**: 避免告警风暴
   - 相同告警合并
   - 依赖关系抑制
   - 时间窗口限制

3. **多渠道通知**: 确保及时响应
   - 邮件、短信、电话
   - 钉钉、企业微信
   - PagerDuty、OpsGenie

4. **告警收敛**: 减少噪音
   - 智能分组
   - 阈值调优
   - 白名单过滤

### 5. 性能监控指标

**Q: Go应用需要监控哪些关键指标？**

**A: 关键指标分类：**

```go
// 1. 运行时指标
type RuntimeMetrics struct {
    Goroutines    int     `json:"goroutines"`     // Goroutine数量
    HeapAlloc     int64   `json:"heap_alloc"`     // 堆内存分配
    HeapSys       int64   `json:"heap_sys"`       // 堆内存系统
    GCPauseNs     int64   `json:"gc_pause_ns"`    // GC暂停时间
    NumGC         uint32  `json:"num_gc"`         // GC次数
    CPUUsage      float64 `json:"cpu_usage"`      // CPU使用率
}

// 2. 业务指标
type BusinessMetrics struct {
    RequestRate   float64 `json:"request_rate"`   // 请求速率
    ErrorRate     float64 `json:"error_rate"`     // 错误率
    ResponseTime  float64 `json:"response_time"`  // 响应时间
    Throughput    float64 `json:"throughput"`     // 吞吐量
}

// 3. 基础设施指标
type InfraMetrics struct {
    DiskUsage     float64 `json:"disk_usage"`     // 磁盘使用率
    NetworkIO     int64   `json:"network_io"`     // 网络IO
    DatabaseConn  int     `json:"database_conn"`  // 数据库连接数
    CacheHitRate  float64 `json:"cache_hit_rate"` // 缓存命中率
}
```

### 6. SLA设计与监控

**Q: 如何制定合理的SLA指标？**

**A: SLA制定原则：**

1. **SMART原则**:
   - Specific: 具体明确
   - Measurable: 可量化
   - Achievable: 可实现
   - Relevant: 相关性
   - Time-bound: 有时限

2. **关键指标**:
   - 可用性: 99.9% (8.76小时/年)
   - 响应时间: P95 < 500ms
   - 错误率: < 0.1%
   - 吞吐量: > 1000 RPS

3. **监控实现**:
```go
// SLA监控示例
func (s *SLAMonitor) CheckAvailability() float64 {
    totalTime := time.Since(s.startTime)
    downtime := s.calculateDowntime()
    uptime := totalTime - downtime
    return float64(uptime) / float64(totalTime) * 100
}
```

---

## 🏋️ 实战练习题

### 练习1: 构建完整监控系统

**题目**: 为一个电商微服务系统设计并实现完整的监控方案

**要求**:
1. 设计监控架构图，包含指标收集、日志聚合、链路追踪
2. 实现Go应用的指标收集中间件
3. 配置Prometheus + Grafana监控栈
4. 设计告警规则和通知机制
5. 制定SLA指标和监控策略

**技术栈**:
- Prometheus + Grafana
- ELK/EFK日志栈
- Jaeger链路追踪
- AlertManager告警
- Docker + Kubernetes

**评估标准**:
- 架构设计合理性 (25%)
- 代码实现质量 (25%)
- 配置文件完整性 (25%)
- 监控覆盖度 (25%)

### 练习2: 性能问题诊断

**题目**: 某Go Web服务出现性能问题，请设计诊断方案

**问题现象**:
- 响应时间从100ms增加到2s
- CPU使用率持续90%+
- 内存使用量不断增长
- 数据库连接池耗尽

**要求**:
1. 设计问题诊断流程
2. 确定需要收集的监控指标
3. 实现性能分析工具集成
4. 提供问题定位和解决方案
5. 建立预防机制

**技术要点**:
- pprof性能分析
- 链路追踪分析
- 数据库监控
- 缓存监控
- 系统资源监控

### 练习3: 告警系统优化

**题目**: 优化现有告警系统，减少告警噪音，提高响应效率

**现状问题**:
- 告警过多，运维疲劳
- 误报率高，影响响应
- 告警不及时，错过关键问题
- 告警信息不够详细

**要求**:
1. 分析现有告警规则，识别问题
2. 设计告警分级和收敛策略
3. 实现智能告警路由
4. 建立告警效果评估机制
5. 提供告警系统监控方案

**技术要点**:
- 告警规则优化
- 多维度告警收敛
- 动态阈值调整
- 告警疲劳度分析
- 告警响应时间统计

---

## 📚 本章总结

通过本章学习，我们深入掌握了现代分布式系统监控与日志的完整体系：

### 🎯 核心收获

1. **监控体系建设** 📊
   - 掌握了Prometheus+Grafana的完整监控方案
   - 学会了Go应用的多层次指标收集
   - 理解了监控系统的架构设计原则

2. **日志聚合系统** 📋
   - 深入了解ELK/EFK日志聚合架构
   - 掌握了结构化日志的设计和实现
   - 学会了日志的收集、处理、存储、分析

3. **分布式链路追踪** 🔗
   - 理解了分布式追踪的核心概念
   - 掌握了Jaeger的集成和使用
   - 学会了跨服务调用链的监控

4. **告警系统设计** 🚨
   - 掌握了AlertManager的配置和使用
   - 学会了告警规则的设计和优化
   - 理解了告警收敛和通知机制

5. **SLA制定与监控** 📏
   - 学会了性能指标的定义和监控
   - 掌握了SLA的制定和合规性检查
   - 理解了服务质量的量化管理

6. **企业级实践** 🏢
   - 结合Mall-Go项目的完整监控实现
   - 掌握了生产环境的监控最佳实践
   - 学会了监控系统的运维和优化

### 🚀 技术进阶

- **监控即代码**: 将监控配置纳入版本控制
- **智能运维**: 基于机器学习的异常检测
- **成本优化**: 监控数据的存储和查询优化
- **安全监控**: 安全事件的监控和响应

### 💡 最佳实践

1. **渐进式监控**: 从基础指标开始，逐步完善
2. **业务导向**: 监控指标要与业务目标对齐
3. **自动化优先**: 尽可能自动化监控和告警
4. **持续优化**: 定期评估和优化监控效果

监控与日志系统是保障系统稳定运行的重要基础设施，通过本章的学习，你已经具备了构建企业级监控系统的能力！ 🎉

---

*下一章我们将学习《性能优化技巧》，深入探讨Go程序的性能分析和优化策略！* 🚀
