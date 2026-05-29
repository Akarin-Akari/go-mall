package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// PrometheusMetrics Prometheus指标集合
type PrometheusMetrics struct {
	// HTTP请求指标
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPRequestSize     *prometheus.HistogramVec
	HTTPResponseSize    *prometheus.HistogramVec

	// 业务指标
	UserRegistrations *prometheus.CounterVec
	UserLogins        *prometheus.CounterVec
	OrdersCreated     *prometheus.CounterVec
	OrdersCompleted   *prometheus.CounterVec
	PaymentsProcessed *prometheus.CounterVec
	ProductViews      *prometheus.CounterVec
	CartOperations    *prometheus.CounterVec

	// 数据库指标
	DBConnectionsOpen  prometheus.Gauge
	DBConnectionsInUse prometheus.Gauge
	DBConnectionsIdle  prometheus.Gauge
	DBQueryDuration    *prometheus.HistogramVec
	DBQueryTotal       *prometheus.CounterVec

	// Redis指标
	RedisConnectionsTotal prometheus.Gauge
	RedisConnectionsIdle  prometheus.Gauge
	RedisCommandDuration  *prometheus.HistogramVec
	RedisCommandTotal     *prometheus.CounterVec
	RedisErrors           *prometheus.CounterVec

	// 系统指标
	GoRoutines    prometheus.Gauge
	MemoryAlloc   prometheus.Gauge
	MemorySys     prometheus.Gauge
	GCPauseTotal  prometheus.Counter
	GCPauseDuration prometheus.Histogram
}

// NewPrometheusMetrics 创建Prometheus指标集合
func NewPrometheusMetrics() *PrometheusMetrics {
	return &PrometheusMetrics{
		// HTTP请求指标
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "mall_go",
				Subsystem: "http",
				Name:      "requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "mall_go",
				Subsystem: "http",
				Name:      "request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"method", "endpoint"},
		),
		HTTPRequestSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "mall_go",
				Subsystem: "http",
				Name:      "request_size_bytes",
				Help:      "HTTP request size in bytes",
				Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"method", "endpoint"},
		),
		HTTPResponseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "mall_go",
				Subsystem: "http",
				Name:      "response_size_bytes",
				Help:      "HTTP response size in bytes",
				Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"method", "endpoint"},
		),

		// 业务指标
		UserRegistrations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "mall_go",
				Subsystem: "business",
				Name:      "user_registrations_total",
				Help:      "Total number of user registrations",
			},
			[]string{"source"},
		),
		UserLogins: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "mall_go",
				Subsystem: "business",
				Name:      "user_logins_total",
				Help:      "Total number of user logins",
			},
			[]string{"status"},
		),
		OrdersCreated: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "mall_go",
				Subsystem: "business",
				Name:      "orders_created_total",
				Help:      "Total number of orders created",
			},
			[]string{"type"},
		),
		OrdersCompleted: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "mall_go",
				Subsystem: "business",
				Name:      "orders_completed_total",
				Help:      "Total number of orders completed",
			},
			[]string{"type"},
		),
		PaymentsProcessed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "mall_go",
				Subsystem: "business",
				Name:      "payments_processed_total",
				Help:      "Total number of payments processed",
			},
			[]string{"method", "status"},
		),
		ProductViews: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "mall_go",
				Subsystem: "business",
				Name:      "product_views_total",
				Help:      "Total number of product views",
			},
			[]string{"category"},
		),
		CartOperations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "mall_go",
				Subsystem: "business",
				Name:      "cart_operations_total",
				Help:      "Total number of cart operations",
			},
			[]string{"operation"},
		),

		// 数据库指标
		DBConnectionsOpen: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "mall_go",
				Subsystem: "database",
				Name:      "connections_open",
				Help:      "Number of open database connections",
			},
		),
		DBConnectionsInUse: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "mall_go",
				Subsystem: "database",
				Name:      "connections_in_use",
				Help:      "Number of database connections in use",
			},
		),
		DBConnectionsIdle: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "mall_go",
				Subsystem: "database",
				Name:      "connections_idle",
				Help:      "Number of idle database connections",
			},
		),
		DBQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "mall_go",
				Subsystem: "database",
				Name:      "query_duration_seconds",
				Help:      "Database query duration in seconds",
				Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
			},
			[]string{"operation"},
		),
		DBQueryTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "mall_go",
				Subsystem: "database",
				Name:      "query_total",
				Help:      "Total number of database queries",
			},
			[]string{"operation", "status"},
		),

		// Redis指标
		RedisConnectionsTotal: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "mall_go",
				Subsystem: "redis",
				Name:      "connections_total",
				Help:      "Total number of Redis connections",
			},
		),
		RedisConnectionsIdle: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "mall_go",
				Subsystem: "redis",
				Name:      "connections_idle",
				Help:      "Number of idle Redis connections",
			},
		),
		RedisCommandDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "mall_go",
				Subsystem: "redis",
				Name:      "command_duration_seconds",
				Help:      "Redis command duration in seconds",
				Buckets:   []float64{.0001, .0005, .001, .005, .01, .025, .05, .1},
			},
			[]string{"command"},
		),
		RedisCommandTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "mall_go",
				Subsystem: "redis",
				Name:      "command_total",
				Help:      "Total number of Redis commands",
			},
			[]string{"command", "status"},
		),
		RedisErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "mall_go",
				Subsystem: "redis",
				Name:      "errors_total",
				Help:      "Total number of Redis errors",
			},
			[]string{"type"},
		),

		// 系统指标
		GoRoutines: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "mall_go",
				Subsystem: "runtime",
				Name:      "goroutines",
				Help:      "Number of goroutines",
			},
		),
		MemoryAlloc: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "mall_go",
				Subsystem: "runtime",
				Name:      "memory_alloc_bytes",
				Help:      "Bytes of allocated heap objects",
			},
		),
		MemorySys: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "mall_go",
				Subsystem: "runtime",
				Name:      "memory_sys_bytes",
				Help:      "Total bytes of memory obtained from the OS",
			},
		),
		GCPauseTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: "mall_go",
				Subsystem: "runtime",
				Name:      "gc_pause_total",
				Help:      "Total number of GC pauses",
			},
		),
		GCPauseDuration: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: "mall_go",
				Subsystem: "runtime",
				Name:      "gc_pause_duration_seconds",
				Help:      "GC pause duration in seconds",
				Buckets:   []float64{.00001, .00005, .0001, .0005, .001, .005, .01},
			},
		),
	}
}

// 全局Prometheus指标实例
var GlobalPrometheusMetrics *PrometheusMetrics

// InitPrometheusMetrics 初始化Prometheus指标
func InitPrometheusMetrics() {
	GlobalPrometheusMetrics = NewPrometheusMetrics()
}

// GetPrometheusMetrics 获取全局Prometheus指标
func GetPrometheusMetrics() *PrometheusMetrics {
	if GlobalPrometheusMetrics == nil {
		InitPrometheusMetrics()
	}
	return GlobalPrometheusMetrics
}

