package metrics

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"mall-go/pkg/logger"

	"go.uber.org/zap"
)

// MetricType 指标类型
type MetricType string

const (
	CounterMetric   MetricType = "counter"   // 计数器
	GaugeMetric     MetricType = "gauge"     // 计量器
	HistogramMetric MetricType = "histogram" // 直方图
	SummaryMetric   MetricType = "summary"   // 摘要
)

// Metric 指标接口
type Metric interface {
	Name() string
	Type() MetricType
	Value() interface{}
	Labels() map[string]string
	Timestamp() time.Time
}

// Counter 计数器指标
type Counter struct {
	name      string
	value     int64
	labels    map[string]string
	timestamp time.Time
	mutex     sync.RWMutex
}

// NewCounter 创建计数器
func NewCounter(name string, labels map[string]string) *Counter {
	return &Counter{
		name:      name,
		value:     0,
		labels:    labels,
		timestamp: time.Now(),
	}
}

func (c *Counter) Name() string                { return c.name }
func (c *Counter) Type() MetricType            { return CounterMetric }
func (c *Counter) Labels() map[string]string  { return c.labels }
func (c *Counter) Timestamp() time.Time       { return c.timestamp }

func (c *Counter) Value() interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.value
}

// Inc 增加计数
func (c *Counter) Inc() {
	c.Add(1)
}

// Add 增加指定值
func (c *Counter) Add(value int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.value += value
	c.timestamp = time.Now()
}

// Gauge 计量器指标
type Gauge struct {
	name      string
	value     float64
	labels    map[string]string
	timestamp time.Time
	mutex     sync.RWMutex
}

// NewGauge 创建计量器
func NewGauge(name string, labels map[string]string) *Gauge {
	return &Gauge{
		name:      name,
		value:     0,
		labels:    labels,
		timestamp: time.Now(),
	}
}

func (g *Gauge) Name() string                { return g.name }
func (g *Gauge) Type() MetricType            { return GaugeMetric }
func (g *Gauge) Labels() map[string]string  { return g.labels }
func (g *Gauge) Timestamp() time.Time       { return g.timestamp }

func (g *Gauge) Value() interface{} {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.value
}

// Set 设置值
func (g *Gauge) Set(value float64) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.value = value
	g.timestamp = time.Now()
}

// Inc 增加1
func (g *Gauge) Inc() {
	g.Add(1)
}

// Dec 减少1
func (g *Gauge) Dec() {
	g.Add(-1)
}

// Add 增加指定值
func (g *Gauge) Add(value float64) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.value += value
	g.timestamp = time.Now()
}

// Histogram 直方图指标
type Histogram struct {
	name      string
	buckets   []float64
	counts    []int64
	sum       float64
	count     int64
	labels    map[string]string
	timestamp time.Time
	mutex     sync.RWMutex
}

// NewHistogram 创建直方图
func NewHistogram(name string, buckets []float64, labels map[string]string) *Histogram {
	if buckets == nil {
		// 默认桶
		buckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}
	}
	
	return &Histogram{
		name:      name,
		buckets:   buckets,
		counts:    make([]int64, len(buckets)+1), // +1 for +Inf bucket
		sum:       0,
		count:     0,
		labels:    labels,
		timestamp: time.Now(),
	}
}

func (h *Histogram) Name() string                { return h.name }
func (h *Histogram) Type() MetricType            { return HistogramMetric }
func (h *Histogram) Labels() map[string]string  { return h.labels }
func (h *Histogram) Timestamp() time.Time       { return h.timestamp }

func (h *Histogram) Value() interface{} {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return map[string]interface{}{
		"count":   h.count,
		"sum":     h.sum,
		"buckets": h.buckets,
		"counts":  h.counts,
	}
}

// Observe 观察值
func (h *Histogram) Observe(value float64) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	h.sum += value
	h.count++
	h.timestamp = time.Now()
	
	// 找到对应的桶
	for i, bucket := range h.buckets {
		if value <= bucket {
			h.counts[i]++
		}
	}
	// +Inf bucket
	h.counts[len(h.buckets)]++
}

// GetPercentile 获取百分位数
func (h *Histogram) GetPercentile(percentile float64) float64 {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	if h.count == 0 {
		return 0
	}
	
	target := float64(h.count) * percentile / 100
	cumulative := int64(0)
	
	for i, count := range h.counts {
		cumulative += count
		if float64(cumulative) >= target {
			if i < len(h.buckets) {
				return h.buckets[i]
			}
			return h.buckets[len(h.buckets)-1]
		}
	}
	
	return h.buckets[len(h.buckets)-1]
}

// MetricsRegistry 指标注册表
type MetricsRegistry struct {
	metrics map[string]Metric
	mutex   sync.RWMutex
}

// NewMetricsRegistry 创建指标注册表
func NewMetricsRegistry() *MetricsRegistry {
	return &MetricsRegistry{
		metrics: make(map[string]Metric),
	}
}

// Register 注册指标
func (r *MetricsRegistry) Register(metric Metric) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	key := r.getMetricKey(metric.Name(), metric.Labels())
	if _, exists := r.metrics[key]; exists {
		return fmt.Errorf("metric already exists: %s", key)
	}
	
	r.metrics[key] = metric
	return nil
}

// GetMetric 获取指标
func (r *MetricsRegistry) GetMetric(name string, labels map[string]string) Metric {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	key := r.getMetricKey(name, labels)
	return r.metrics[key]
}

// GetAllMetrics 获取所有指标
func (r *MetricsRegistry) GetAllMetrics() map[string]Metric {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	result := make(map[string]Metric)
	for k, v := range r.metrics {
		result[k] = v
	}
	return result
}

// getMetricKey 生成指标键
func (r *MetricsRegistry) getMetricKey(name string, labels map[string]string) string {
	key := name
	for k, v := range labels {
		key += fmt.Sprintf(",%s=%s", k, v)
	}
	return key
}

// SystemMetrics 系统指标
type SystemMetrics struct {
	cpuUsage    *Gauge
	memUsage    *Gauge
	goroutines  *Gauge
	registry    *MetricsRegistry
	ticker      *time.Ticker
	stopChan    chan struct{}
}

// NewSystemMetrics 创建系统指标
func NewSystemMetrics(registry *MetricsRegistry) *SystemMetrics {
	sm := &SystemMetrics{
		cpuUsage:   NewGauge("system_cpu_usage_percent", nil),
		memUsage:   NewGauge("system_memory_usage_bytes", nil),
		goroutines: NewGauge("system_goroutines_count", nil),
		registry:   registry,
		stopChan:   make(chan struct{}),
	}
	
	// 注册指标
	registry.Register(sm.cpuUsage)
	registry.Register(sm.memUsage)
	registry.Register(sm.goroutines)
	
	return sm
}

// Start 开始收集系统指标
func (sm *SystemMetrics) Start(interval time.Duration) {
	sm.ticker = time.NewTicker(interval)
	
	go func() {
		for {
			select {
			case <-sm.ticker.C:
				sm.collectMetrics()
			case <-sm.stopChan:
				return
			}
		}
	}()
}

// Stop 停止收集系统指标
func (sm *SystemMetrics) Stop() {
	if sm.ticker != nil {
		sm.ticker.Stop()
	}
	close(sm.stopChan)
}

// collectMetrics 收集系统指标
func (sm *SystemMetrics) collectMetrics() {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	
	// 内存使用量
	sm.memUsage.Set(float64(mem.Alloc))
	
	// Goroutine数量
	sm.goroutines.Set(float64(runtime.NumGoroutine()))
	
	// CPU使用率需要通过其他方式获取，这里简化处理
	// 在实际生产环境中，可以使用第三方库如 gopsutil
}

// ApplicationMetrics 应用指标
type ApplicationMetrics struct {
	httpRequests     *Counter
	httpDuration     *Histogram
	dbConnections    *Gauge
	cacheHitRate     *Gauge
	activeUsers      *Gauge
	errorRate        *Counter
	registry         *MetricsRegistry
}

// NewApplicationMetrics 创建应用指标
func NewApplicationMetrics(registry *MetricsRegistry) *ApplicationMetrics {
	am := &ApplicationMetrics{
		httpRequests:  NewCounter("http_requests_total", map[string]string{"method": "", "status": ""}),
		httpDuration:  NewHistogram("http_request_duration_seconds", nil, map[string]string{"method": "", "endpoint": ""}),
		dbConnections: NewGauge("database_connections_active", nil),
		cacheHitRate:  NewGauge("cache_hit_rate_percent", nil),
		activeUsers:   NewGauge("active_users_count", nil),
		errorRate:     NewCounter("errors_total", map[string]string{"type": ""}),
		registry:      registry,
	}
	
	// 注册指标
	registry.Register(am.httpRequests)
	registry.Register(am.httpDuration)
	registry.Register(am.dbConnections)
	registry.Register(am.cacheHitRate)
	registry.Register(am.activeUsers)
	registry.Register(am.errorRate)
	
	return am
}

// RecordHTTPRequest 记录HTTP请求
func (am *ApplicationMetrics) RecordHTTPRequest(method, status string, duration time.Duration) {
	// 创建带标签的计数器
	labels := map[string]string{"method": method, "status": status}
	counter := NewCounter("http_requests_total", labels)
	counter.Inc()
	
	// 记录请求时长
	histLabels := map[string]string{"method": method, "endpoint": ""}
	hist := NewHistogram("http_request_duration_seconds", nil, histLabels)
	hist.Observe(duration.Seconds())
}

// BusinessMetrics 业务指标
type BusinessMetrics struct {
	orderTotal      *Counter
	orderValue      *Histogram
	userRegistration *Counter
	productViews    *Counter
	cartAdditions   *Counter
	paymentSuccess  *Counter
	paymentFailure  *Counter
	registry        *MetricsRegistry
}

// NewBusinessMetrics 创建业务指标
func NewBusinessMetrics(registry *MetricsRegistry) *BusinessMetrics {
	bm := &BusinessMetrics{
		orderTotal:       NewCounter("orders_total", map[string]string{"status": ""}),
		orderValue:       NewHistogram("order_value_yuan", []float64{10, 50, 100, 500, 1000, 5000, 10000}, nil),
		userRegistration: NewCounter("user_registrations_total", nil),
		productViews:     NewCounter("product_views_total", map[string]string{"product_id": ""}),
		cartAdditions:    NewCounter("cart_additions_total", map[string]string{"product_id": ""}),
		paymentSuccess:   NewCounter("payments_success_total", map[string]string{"method": ""}),
		paymentFailure:   NewCounter("payments_failure_total", map[string]string{"method": "", "reason": ""}),
		registry:         registry,
	}
	
	// 注册指标
	registry.Register(bm.orderTotal)
	registry.Register(bm.orderValue)
	registry.Register(bm.userRegistration)
	registry.Register(bm.productViews)
	registry.Register(bm.cartAdditions)
	registry.Register(bm.paymentSuccess)
	registry.Register(bm.paymentFailure)
	
	return bm
}

// RecordOrder 记录订单
func (bm *BusinessMetrics) RecordOrder(status string, value float64) {
	labels := map[string]string{"status": status}
	counter := NewCounter("orders_total", labels)
	counter.Inc()
	
	bm.orderValue.Observe(value)
}

// MetricsCollector 指标收集器
type MetricsCollector struct {
	registry     *MetricsRegistry
	systemMetrics *SystemMetrics
	appMetrics   *ApplicationMetrics
	bizMetrics   *BusinessMetrics
	ticker       *time.Ticker
	stopChan     chan struct{}
}

// NewMetricsCollector 创建指标收集器
func NewMetricsCollector() *MetricsCollector {
	registry := NewMetricsRegistry()
	
	return &MetricsCollector{
		registry:      registry,
		systemMetrics: NewSystemMetrics(registry),
		appMetrics:    NewApplicationMetrics(registry),
		bizMetrics:    NewBusinessMetrics(registry),
		stopChan:      make(chan struct{}),
	}
}

// Start 开始收集指标
func (mc *MetricsCollector) Start(reportInterval time.Duration) {
	// 启动系统指标收集
	mc.systemMetrics.Start(30 * time.Second)
	
	// 启动定期报告
	mc.ticker = time.NewTicker(reportInterval)
	
	go func() {
		for {
			select {
			case <-mc.ticker.C:
				mc.reportMetrics()
			case <-mc.stopChan:
				return
			}
		}
	}()
	
	logger.Info("Metrics collector started",
		zap.Duration("report_interval", reportInterval))
}

// Stop 停止收集指标
func (mc *MetricsCollector) Stop() {
	mc.systemMetrics.Stop()
	
	if mc.ticker != nil {
		mc.ticker.Stop()
	}
	close(mc.stopChan)
	
	logger.Info("Metrics collector stopped")
}

// GetSystemMetrics 获取系统指标
func (mc *MetricsCollector) GetSystemMetrics() *SystemMetrics {
	return mc.systemMetrics
}

// GetApplicationMetrics 获取应用指标
func (mc *MetricsCollector) GetApplicationMetrics() *ApplicationMetrics {
	return mc.appMetrics
}

// GetBusinessMetrics 获取业务指标
func (mc *MetricsCollector) GetBusinessMetrics() *BusinessMetrics {
	return mc.bizMetrics
}

// GetRegistry 获取注册表
func (mc *MetricsCollector) GetRegistry() *MetricsRegistry {
	return mc.registry
}

// reportMetrics 报告指标
func (mc *MetricsCollector) reportMetrics() {
	ctx := context.Background()
	metrics := mc.registry.GetAllMetrics()
	
	for name, metric := range metrics {
		logger.LogPerformance(ctx, "metrics_report", 0, map[string]interface{}{
			"metric_name":  name,
			"metric_type":  string(metric.Type()),
			"metric_value": metric.Value(),
			"labels":       metric.Labels(),
			"timestamp":    metric.Timestamp(),
		})
	}
}

// GetMetricsSummary 获取指标摘要
func (mc *MetricsCollector) GetMetricsSummary() map[string]interface{} {
	summary := make(map[string]interface{})
	
	// 系统指标摘要
	summary["system"] = map[string]interface{}{
		"memory_usage_mb": float64(mc.systemMetrics.memUsage.Value().(float64)) / 1024 / 1024,
		"goroutines":      mc.systemMetrics.goroutines.Value(),
	}
	
	// 应用指标摘要
	summary["application"] = map[string]interface{}{
		"db_connections": mc.appMetrics.dbConnections.Value(),
		"cache_hit_rate": mc.appMetrics.cacheHitRate.Value(),
		"active_users":   mc.appMetrics.activeUsers.Value(),
	}
	
	return summary
}

// Alert 告警
type Alert struct {
	Name        string            `json:"name"`
	Level       string            `json:"level"`       // info, warning, critical
	Message     string            `json:"message"`
	Labels      map[string]string `json:"labels"`
	Timestamp   time.Time         `json:"timestamp"`
	Resolved    bool              `json:"resolved"`
	ResolvedAt  *time.Time        `json:"resolved_at,omitempty"`
}

// AlertManager 告警管理器
type AlertManager struct {
	alerts   []Alert
	mutex    sync.RWMutex
	handlers []AlertHandler
}

// AlertHandler 告警处理器接口
type AlertHandler interface {
	Handle(alert Alert) error
}

// LogAlertHandler 日志告警处理器
type LogAlertHandler struct{}

// Handle 处理告警
func (h *LogAlertHandler) Handle(alert Alert) error {
	ctx := context.Background()
	
	severity := "info"
	switch alert.Level {
	case "critical":
		severity = "critical"
	case "warning":
		severity = "medium"
	}
	
	details := map[string]interface{}{
		"alert_name": alert.Name,
		"labels":     alert.Labels,
		"resolved":   alert.Resolved,
	}
	
	logger.LogSecurityEvent(ctx, alert.Message, severity, details)
	return nil
}

// NewAlertManager 创建告警管理器
func NewAlertManager() *AlertManager {
	am := &AlertManager{
		alerts:   make([]Alert, 0),
		handlers: make([]AlertHandler, 0),
	}
	
	// 添加默认的日志处理器
	am.AddHandler(&LogAlertHandler{})
	
	return am
}

// AddHandler 添加告警处理器
func (am *AlertManager) AddHandler(handler AlertHandler) {
	am.handlers = append(am.handlers, handler)
}

// FireAlert 触发告警
func (am *AlertManager) FireAlert(name, level, message string, labels map[string]string) {
	alert := Alert{
		Name:      name,
		Level:     level,
		Message:   message,
		Labels:    labels,
		Timestamp: time.Now(),
		Resolved:  false,
	}
	
	am.mutex.Lock()
	am.alerts = append(am.alerts, alert)
	am.mutex.Unlock()
	
	// 处理告警
	for _, handler := range am.handlers {
		if err := handler.Handle(alert); err != nil {
			logger.Error("Failed to handle alert",
				zap.String("alert_name", name),
				zap.Error(err))
		}
	}
}

// ResolveAlert 解决告警
func (am *AlertManager) ResolveAlert(name string, labels map[string]string) {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	
	for i := range am.alerts {
		if am.alerts[i].Name == name && !am.alerts[i].Resolved {
			// 检查标签匹配
			match := true
			for k, v := range labels {
				if am.alerts[i].Labels[k] != v {
					match = false
					break
				}
			}
			
			if match {
				now := time.Now()
				am.alerts[i].Resolved = true
				am.alerts[i].ResolvedAt = &now
				break
			}
		}
	}
}

// GetActiveAlerts 获取活跃告警
func (am *AlertManager) GetActiveAlerts() []Alert {
	am.mutex.RLock()
	defer am.mutex.RUnlock()
	
	var active []Alert
	for _, alert := range am.alerts {
		if !alert.Resolved {
			active = append(active, alert)
		}
	}
	
	return active
}

// GetAllAlerts 获取所有告警
func (am *AlertManager) GetAllAlerts() []Alert {
	am.mutex.RLock()
	defer am.mutex.RUnlock()
	
	result := make([]Alert, len(am.alerts))
	copy(result, am.alerts)
	
	return result
}

// MonitoringService 监控服务
type MonitoringService struct {
	collector    *MetricsCollector
	alertManager *AlertManager
	thresholds   map[string]float64
}

// NewMonitoringService 创建监控服务
func NewMonitoringService() *MonitoringService {
	return &MonitoringService{
		collector:    NewMetricsCollector(),
		alertManager: NewAlertManager(),
		thresholds: map[string]float64{
			"memory_usage_mb":    1000,  // 1GB
			"response_time_ms":   2000,  // 2秒
			"error_rate_percent": 5,     // 5%
			"cpu_usage_percent":  80,    // 80%
		},
	}
}

// Start 启动监控服务
func (ms *MonitoringService) Start() {
	ms.collector.Start(1 * time.Minute)
	
	// 启动阈值检查
	go ms.checkThresholds()
	
	logger.Info("Monitoring service started")
}

// Stop 停止监控服务
func (ms *MonitoringService) Stop() {
	ms.collector.Stop()
	logger.Info("Monitoring service stopped")
}

// checkThresholds 检查阈值
func (ms *MonitoringService) checkThresholds() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		summary := ms.collector.GetMetricsSummary()
		
		// 检查内存使用
		if memUsage, ok := summary["system"].(map[string]interface{})["memory_usage_mb"].(float64); ok {
			if memUsage > ms.thresholds["memory_usage_mb"] {
				ms.alertManager.FireAlert(
					"high_memory_usage",
					"warning",
					fmt.Sprintf("Memory usage is high: %.2f MB", memUsage),
					map[string]string{"threshold": fmt.Sprintf("%.0f", ms.thresholds["memory_usage_mb"])},
				)
			}
		}
	}
}

// GetCollector 获取指标收集器
func (ms *MonitoringService) GetCollector() *MetricsCollector {
	return ms.collector
}

// GetAlertManager 获取告警管理器
func (ms *MonitoringService) GetAlertManager() *AlertManager {
	return ms.alertManager
}

// GetHealthStatus 获取健康状态
func (ms *MonitoringService) GetHealthStatus() map[string]interface{} {
	summary := ms.collector.GetMetricsSummary()
	alerts := ms.alertManager.GetActiveAlerts()
	
	status := "healthy"
	if len(alerts) > 0 {
		for _, alert := range alerts {
			if alert.Level == "critical" {
				status = "critical"
				break
			} else if alert.Level == "warning" && status == "healthy" {
				status = "warning"
			}
		}
	}
	
	return map[string]interface{}{
		"status":        status,
		"metrics":       summary,
		"active_alerts": len(alerts),
		"timestamp":     time.Now(),
	}
}