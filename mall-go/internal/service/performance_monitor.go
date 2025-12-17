package service

import (
	"context"
	"encoding/json"
	"runtime"
	"sync"
	"time"

	"mall-go/pkg/logger"

	"go.uber.org/zap"
)

// PerformanceMonitor 性能监控器
type PerformanceMonitor struct {
	metrics     map[string]*MetricData
	alerts      []AlertRule
	mu          sync.RWMutex
	isRunning   bool
	stopChan    chan struct{}
	alertChan   chan Alert
}

// MetricData 指标数据
type MetricData struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"` // counter, gauge, histogram
	Value       float64                `json:"value"`
	Labels      map[string]string      `json:"labels"`
	Timestamp   time.Time              `json:"timestamp"`
	History     []float64              `json:"history"`
	MaxHistory  int                    `json:"max_history"`
}

// AlertRule 告警规则
type AlertRule struct {
	Name        string        `json:"name"`
	MetricName  string        `json:"metric_name"`
	Condition   string        `json:"condition"` // >, <, >=, <=, ==
	Threshold   float64       `json:"threshold"`
	Duration    time.Duration `json:"duration"`
	Enabled     bool          `json:"enabled"`
	LastFired   time.Time     `json:"last_fired"`
	CoolDown    time.Duration `json:"cool_down"`
}

// Alert 告警信息
type Alert struct {
	RuleName    string                 `json:"rule_name"`
	MetricName  string                 `json:"metric_name"`
	Value       float64                `json:"value"`
	Threshold   float64                `json:"threshold"`
	Condition   string                 `json:"condition"`
	Timestamp   time.Time              `json:"timestamp"`
	Severity    string                 `json:"severity"`
	Message     string                 `json:"message"`
	Labels      map[string]string      `json:"labels"`
}

// PerformanceReport 性能报告
type PerformanceReport struct {
	GeneratedAt     time.Time                  `json:"generated_at"`
	TimeRange       string                     `json:"time_range"`
	Summary         map[string]interface{}     `json:"summary"`
	Metrics         map[string]*MetricData     `json:"metrics"`
	Alerts          []Alert                    `json:"alerts"`
	Recommendations []string                   `json:"recommendations"`
	SystemInfo      map[string]interface{}     `json:"system_info"`
}

// NewPerformanceMonitor 创建性能监控器
func NewPerformanceMonitor() *PerformanceMonitor {
	pm := &PerformanceMonitor{
		metrics:   make(map[string]*MetricData),
		alerts:    make([]AlertRule, 0),
		stopChan:  make(chan struct{}),
		alertChan: make(chan Alert, 100),
	}
	
	// 添加默认告警规则
	pm.addDefaultAlertRules()
	
	return pm
}

// addDefaultAlertRules 添加默认告警规则
func (pm *PerformanceMonitor) addDefaultAlertRules() {
	defaultRules := []AlertRule{
		{
			Name:       "高内存使用率",
			MetricName: "memory_usage_mb",
			Condition:  ">",
			Threshold:  1024, // 1GB
			Duration:   5 * time.Minute,
			Enabled:    true,
			CoolDown:   10 * time.Minute,
		},
		{
			Name:       "高CPU使用率",
			MetricName: "cpu_usage_percent",
			Condition:  ">",
			Threshold:  80,
			Duration:   3 * time.Minute,
			Enabled:    true,
			CoolDown:   5 * time.Minute,
		},
		{
			Name:       "慢响应时间",
			MetricName: "avg_response_time_ms",
			Condition:  ">",
			Threshold:  1000, // 1秒
			Duration:   2 * time.Minute,
			Enabled:    true,
			CoolDown:   5 * time.Minute,
		},
		{
			Name:       "高错误率",
			MetricName: "error_rate_percent",
			Condition:  ">",
			Threshold:  5, // 5%
			Duration:   1 * time.Minute,
			Enabled:    true,
			CoolDown:   3 * time.Minute,
		},
		{
			Name:       "Goroutine泄漏",
			MetricName: "goroutine_count",
			Condition:  ">",
			Threshold:  1000,
			Duration:   10 * time.Minute,
			Enabled:    true,
			CoolDown:   15 * time.Minute,
		},
	}
	
	pm.alerts = append(pm.alerts, defaultRules...)
}

// RecordMetric 记录指标
func (pm *PerformanceMonitor) RecordMetric(name string, value float64, labels map[string]string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	metric, exists := pm.metrics[name]
	if !exists {
		metric = &MetricData{
			Name:       name,
			Type:       "gauge",
			Labels:     labels,
			History:    make([]float64, 0),
			MaxHistory: 100,
		}
		pm.metrics[name] = metric
	}
	
	metric.Value = value
	metric.Timestamp = time.Now()
	
	// 添加到历史记录
	metric.History = append(metric.History, value)
	if len(metric.History) > metric.MaxHistory {
		metric.History = metric.History[1:]
	}
	
	// 检查告警
	pm.checkAlerts(name, value)
}

// IncrementCounter 增加计数器
func (pm *PerformanceMonitor) IncrementCounter(name string, labels map[string]string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	metric, exists := pm.metrics[name]
	if !exists {
		metric = &MetricData{
			Name:       name,
			Type:       "counter",
			Labels:     labels,
			History:    make([]float64, 0),
			MaxHistory: 100,
		}
		pm.metrics[name] = metric
	}
	
	metric.Value++
	metric.Timestamp = time.Now()
}

// checkAlerts 检查告警规则
func (pm *PerformanceMonitor) checkAlerts(metricName string, value float64) {
	for i := range pm.alerts {
		rule := &pm.alerts[i]
		if !rule.Enabled || rule.MetricName != metricName {
			continue
		}
		
		// 检查冷却时间
		if time.Since(rule.LastFired) < rule.CoolDown {
			continue
		}
		
		// 检查条件
		triggered := false
		switch rule.Condition {
		case ">":
			triggered = value > rule.Threshold
		case "<":
			triggered = value < rule.Threshold
		case ">=":
			triggered = value >= rule.Threshold
		case "<=":
			triggered = value <= rule.Threshold
		case "==":
			triggered = value == rule.Threshold
		}
		
		if triggered {
			alert := Alert{
				RuleName:   rule.Name,
				MetricName: metricName,
				Value:      value,
				Threshold:  rule.Threshold,
				Condition:  rule.Condition,
				Timestamp:  time.Now(),
				Severity:   pm.getSeverity(rule.Name),
				Message:    pm.getAlertMessage(rule, value),
			}
			
			// 发送告警
			select {
			case pm.alertChan <- alert:
				rule.LastFired = time.Now()
				logger.Warn("性能告警触发",
					zap.String("rule", rule.Name),
					zap.String("metric", metricName),
					zap.Float64("value", value),
					zap.Float64("threshold", rule.Threshold),
				)
			default:
				logger.Error("告警通道已满，丢弃告警", zap.String("rule", rule.Name))
			}
		}
	}
}

// getSeverity 获取告警严重程度
func (pm *PerformanceMonitor) getSeverity(ruleName string) string {
	switch ruleName {
	case "高内存使用率", "高CPU使用率":
		return "high"
	case "慢响应时间", "高错误率":
		return "medium"
	case "Goroutine泄漏":
		return "critical"
	default:
		return "low"
	}
}

// getAlertMessage 获取告警消息
func (pm *PerformanceMonitor) getAlertMessage(rule *AlertRule, value float64) string {
	switch rule.Name {
	case "高内存使用率":
		return "系统内存使用率过高，当前使用 %.2f MB，超过阈值 %.2f MB"
	case "高CPU使用率":
		return "系统CPU使用率过高，当前使用率 %.2f%%，超过阈值 %.2f%%"
	case "慢响应时间":
		return "API响应时间过慢，当前平均响应时间 %.2f ms，超过阈值 %.2f ms"
	case "高错误率":
		return "系统错误率过高，当前错误率 %.2f%%，超过阈值 %.2f%%"
	case "Goroutine泄漏":
		return "检测到可能的Goroutine泄漏，当前Goroutine数量 %.0f，超过阈值 %.0f"
	default:
		return "指标 %s 当前值 %.2f %s 阈值 %.2f"
	}
}

// Start 启动性能监控
func (pm *PerformanceMonitor) Start(ctx context.Context) {
	if pm.isRunning {
		return
	}
	
	pm.isRunning = true
	logger.Info("性能监控器启动")
	
	// 启动指标收集
	go pm.collectSystemMetrics(ctx)
	
	// 启动告警处理
	go pm.handleAlerts(ctx)
}

// Stop 停止性能监控
func (pm *PerformanceMonitor) Stop() {
	if !pm.isRunning {
		return
	}
	
	pm.isRunning = false
	close(pm.stopChan)
	logger.Info("性能监控器停止")
}

// collectSystemMetrics 收集系统指标
func (pm *PerformanceMonitor) collectSystemMetrics(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-pm.stopChan:
			return
		case <-ticker.C:
			pm.collectMetrics()
		}
	}
}

// collectMetrics 收集指标
func (pm *PerformanceMonitor) collectMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	// 内存指标
	pm.RecordMetric("memory_usage_mb", float64(m.Alloc)/1024/1024, nil)
	pm.RecordMetric("memory_sys_mb", float64(m.Sys)/1024/1024, nil)
	pm.RecordMetric("memory_heap_mb", float64(m.HeapAlloc)/1024/1024, nil)
	
	// 运行时指标
	pm.RecordMetric("goroutine_count", float64(runtime.NumGoroutine()), nil)
	pm.RecordMetric("cpu_count", float64(runtime.NumCPU()), nil)
	
	// GC指标
	pm.RecordMetric("gc_count", float64(m.NumGC), nil)
	pm.RecordMetric("gc_pause_ms", float64(m.PauseTotalNs)/1000000, nil)
}

// handleAlerts 处理告警
func (pm *PerformanceMonitor) handleAlerts(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-pm.stopChan:
			return
		case alert := <-pm.alertChan:
			pm.processAlert(alert)
		}
	}
}

// processAlert 处理告警
func (pm *PerformanceMonitor) processAlert(alert Alert) {
	// 记录告警日志
	logger.Error("性能告警",
		zap.String("rule", alert.RuleName),
		zap.String("metric", alert.MetricName),
		zap.Float64("value", alert.Value),
		zap.Float64("threshold", alert.Threshold),
		zap.String("severity", alert.Severity),
		zap.String("message", alert.Message),
	)
	
	// 这里可以添加更多告警处理逻辑，如：
	// - 发送邮件通知
	// - 发送钉钉/企业微信消息
	// - 调用外部告警系统API
	// - 自动执行恢复脚本
}

// GenerateReport 生成性能报告
func (pm *PerformanceMonitor) GenerateReport(timeRange string) *PerformanceReport {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	report := &PerformanceReport{
		GeneratedAt: time.Now(),
		TimeRange:   timeRange,
		Metrics:     make(map[string]*MetricData),
		Alerts:      make([]Alert, 0),
		SystemInfo: map[string]interface{}{
			"go_version":    runtime.Version(),
			"os":           runtime.GOOS,
			"arch":         runtime.GOARCH,
			"cpu_count":    runtime.NumCPU(),
			"goroutines":   runtime.NumGoroutine(),
			"memory_alloc": m.Alloc,
			"memory_sys":   m.Sys,
		},
	}
	
	// 复制指标数据
	for name, metric := range pm.metrics {
		report.Metrics[name] = &MetricData{
			Name:      metric.Name,
			Type:      metric.Type,
			Value:     metric.Value,
			Labels:    metric.Labels,
			Timestamp: metric.Timestamp,
			History:   append([]float64(nil), metric.History...),
		}
	}
	
	// 生成摘要
	report.Summary = pm.generateSummary()
	
	// 生成建议
	report.Recommendations = pm.generateRecommendations()
	
	return report
}

// generateSummary 生成摘要
func (pm *PerformanceMonitor) generateSummary() map[string]interface{} {
	summary := make(map[string]interface{})
	
	if metric, exists := pm.metrics["memory_usage_mb"]; exists {
		summary["avg_memory_usage_mb"] = pm.calculateAverage(metric.History)
	}
	
	if metric, exists := pm.metrics["goroutine_count"]; exists {
		summary["avg_goroutine_count"] = pm.calculateAverage(metric.History)
	}
	
	return summary
}

// generateRecommendations 生成建议
func (pm *PerformanceMonitor) generateRecommendations() []string {
	recommendations := make([]string, 0)
	
	if metric, exists := pm.metrics["memory_usage_mb"]; exists {
		if metric.Value > 512 {
			recommendations = append(recommendations, "内存使用率较高，建议检查是否存在内存泄漏")
		}
	}
	
	if metric, exists := pm.metrics["goroutine_count"]; exists {
		if metric.Value > 500 {
			recommendations = append(recommendations, "Goroutine数量较多，建议检查是否存在Goroutine泄漏")
		}
	}
	
	return recommendations
}

// calculateAverage 计算平均值
func (pm *PerformanceMonitor) calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// GetMetricsJSON 获取指标JSON
func (pm *PerformanceMonitor) GetMetricsJSON() ([]byte, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	return json.Marshal(pm.metrics)
}

// 全局性能监控器实例
var GlobalPerformanceMonitor *PerformanceMonitor

// InitGlobalPerformanceMonitor 初始化全局性能监控器
func InitGlobalPerformanceMonitor() {
	GlobalPerformanceMonitor = NewPerformanceMonitor()
	logger.Info("全局性能监控器已初始化")
}
