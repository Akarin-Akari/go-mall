package middleware

import (
	"runtime"
	"strconv"
	"sync"
	"time"

	"mall-go/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

// PerformanceMetrics 性能指标收集器
type PerformanceMetrics struct {
	// HTTP请求指标
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	httpRequestSize      *prometheus.HistogramVec
	httpResponseSize     *prometheus.HistogramVec
	
	// 数据库指标
	dbQueryDuration      *prometheus.HistogramVec
	dbQueryTotal         *prometheus.CounterVec
	dbConnectionsActive  prometheus.Gauge
	
	// 系统指标
	memoryUsage          prometheus.Gauge
	cpuUsage             prometheus.Gauge
	goroutineCount       prometheus.Gauge
	
	// 业务指标
	addressOperations    *prometheus.CounterVec
	slowOperations       *prometheus.CounterVec
	
	// 缓存指标
	cacheHits            *prometheus.CounterVec
	cacheMisses          *prometheus.CounterVec
	
	mu                   sync.RWMutex
	lastCPUTime          time.Time
	lastCPUUsage         float64
}

// NewPerformanceMetrics 创建性能指标收集器
func NewPerformanceMetrics() *PerformanceMetrics {
	return &PerformanceMetrics{
		httpRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		),
		httpRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		httpRequestSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_size_bytes",
				Help:    "HTTP request size in bytes",
				Buckets: prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"method", "endpoint"},
		),
		httpResponseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_response_size_bytes",
				Help:    "HTTP response size in bytes",
				Buckets: prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"method", "endpoint"},
		),
		dbQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "db_query_duration_seconds",
				Help:    "Database query duration in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 5.0},
			},
			[]string{"operation", "table"},
		),
		dbQueryTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "db_queries_total",
				Help: "Total number of database queries",
			},
			[]string{"operation", "table", "status"},
		),
		dbConnectionsActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "db_connections_active",
				Help: "Number of active database connections",
			},
		),
		memoryUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "memory_usage_bytes",
				Help: "Current memory usage in bytes",
			},
		),
		cpuUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "cpu_usage_percent",
				Help: "Current CPU usage percentage",
			},
		),
		goroutineCount: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "goroutines_count",
				Help: "Current number of goroutines",
			},
		),
		addressOperations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "address_operations_total",
				Help: "Total number of address operations",
			},
			[]string{"operation", "status"},
		),
		slowOperations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slow_operations_total",
				Help: "Total number of slow operations",
			},
			[]string{"operation", "threshold"},
		),
		cacheHits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_hits_total",
				Help: "Total number of cache hits",
			},
			[]string{"cache_type", "operation"},
		),
		cacheMisses: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_misses_total",
				Help: "Total number of cache misses",
			},
			[]string{"cache_type", "operation"},
		),
		lastCPUTime: time.Now(),
	}
}

// PerformanceMiddleware 性能监控中间件
func (pm *PerformanceMetrics) PerformanceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// 记录请求大小
		requestSize := c.Request.ContentLength
		if requestSize > 0 {
			pm.httpRequestSize.WithLabelValues(
				c.Request.Method,
				c.FullPath(),
			).Observe(float64(requestSize))
		}
		
		// 处理请求
		c.Next()
		
		// 计算响应时间
		duration := time.Since(start)
		
		// 记录HTTP指标
		pm.httpRequestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			strconv.Itoa(c.Writer.Status()),
		).Inc()
		
		pm.httpRequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
		).Observe(duration.Seconds())
		
		// 记录响应大小
		responseSize := c.Writer.Size()
		if responseSize > 0 {
			pm.httpResponseSize.WithLabelValues(
				c.Request.Method,
				c.FullPath(),
			).Observe(float64(responseSize))
		}
		
		// 记录慢请求
		if duration > 1*time.Second {
			logger.Warn("慢请求检测",
				zap.String("method", c.Request.Method),
				zap.String("path", c.FullPath()),
				zap.Duration("duration", duration),
				zap.Int("status", c.Writer.Status()),
			)
		}
		
		// 更新系统指标
		pm.updateSystemMetrics()
	}
}

// RecordDBQuery 记录数据库查询指标
func (pm *PerformanceMetrics) RecordDBQuery(operation, table string, duration time.Duration, err error) {
	status := "success"
	if err != nil {
		status = "error"
	}
	
	pm.dbQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
	pm.dbQueryTotal.WithLabelValues(operation, table, status).Inc()
	
	// 记录慢查询
	if duration > 100*time.Millisecond {
		pm.slowOperations.WithLabelValues(operation, "100ms").Inc()
		logger.Warn("慢查询检测",
			zap.String("operation", operation),
			zap.String("table", table),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
	}
}

// RecordAddressOperation 记录地址操作指标
func (pm *PerformanceMetrics) RecordAddressOperation(operation string, err error) {
	status := "success"
	if err != nil {
		status = "error"
	}
	pm.addressOperations.WithLabelValues(operation, status).Inc()
}

// RecordCacheHit 记录缓存命中
func (pm *PerformanceMetrics) RecordCacheHit(cacheType, operation string) {
	pm.cacheHits.WithLabelValues(cacheType, operation).Inc()
}

// RecordCacheMiss 记录缓存未命中
func (pm *PerformanceMetrics) RecordCacheMiss(cacheType, operation string) {
	pm.cacheMisses.WithLabelValues(cacheType, operation).Inc()
}

// updateSystemMetrics 更新系统指标
func (pm *PerformanceMetrics) updateSystemMetrics() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	// 更新内存使用
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	pm.memoryUsage.Set(float64(m.Alloc))
	
	// 更新Goroutine数量
	pm.goroutineCount.Set(float64(runtime.NumGoroutine()))
	
	// 更新CPU使用率（简化版本）
	now := time.Now()
	if now.Sub(pm.lastCPUTime) > 5*time.Second {
		// 这里可以实现更精确的CPU使用率计算
		// 简化版本：基于Goroutine数量估算
		goroutines := float64(runtime.NumGoroutine())
		cpuPercent := (goroutines / 1000) * 100
		if cpuPercent > 100 {
			cpuPercent = 100
		}
		pm.cpuUsage.Set(cpuPercent)
		pm.lastCPUTime = now
	}
}

// GetMetrics 获取当前指标快照
func (pm *PerformanceMetrics) GetMetrics() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return map[string]interface{}{
		"memory": map[string]interface{}{
			"alloc":       m.Alloc,
			"total_alloc": m.TotalAlloc,
			"sys":         m.Sys,
			"heap_alloc":  m.HeapAlloc,
			"heap_sys":    m.HeapSys,
		},
		"runtime": map[string]interface{}{
			"goroutines": runtime.NumGoroutine(),
			"cpu_count":  runtime.NumCPU(),
		},
		"gc": map[string]interface{}{
			"num_gc":     m.NumGC,
			"pause_ns":   m.PauseTotalNs,
			"last_gc":    time.Unix(0, int64(m.LastGC)),
		},
	}
}

// StartSystemMonitoring 启动系统监控
func (pm *PerformanceMetrics) StartSystemMonitoring() {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			pm.updateSystemMetrics()
		}
	}()
}

// 全局性能指标实例
var GlobalMetrics *PerformanceMetrics

// InitPerformanceMetrics 初始化全局性能指标
func InitPerformanceMetrics() {
	GlobalMetrics = NewPerformanceMetrics()
	GlobalMetrics.StartSystemMonitoring()
	logger.Info("性能监控系统已启动")
}
