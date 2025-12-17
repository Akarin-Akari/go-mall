package payment

import (
	"sync"
	"time"

	"mall-go/internal/model"
	"mall-go/pkg/logger"

	"go.uber.org/zap"
)

// PaymentMetrics 支付系统指标
type PaymentMetrics struct {
	mu sync.RWMutex

	// 支付创建指标
	CreatePaymentTotal    map[string]int64           `json:"create_payment_total"`
	CreatePaymentSuccess  map[string]int64           `json:"create_payment_success"`
	CreatePaymentFailed   map[string]int64           `json:"create_payment_failed"`
	CreatePaymentDuration map[string][]time.Duration `json:"-"` // 不序列化，内存中统计用

	// 支付查询指标
	QueryPaymentTotal    map[string]int64           `json:"query_payment_total"`
	QueryPaymentSuccess  map[string]int64           `json:"query_payment_success"`
	QueryPaymentFailed   map[string]int64           `json:"query_payment_failed"`
	QueryPaymentDuration map[string][]time.Duration `json:"-"`

	// 支付回调指标
	CallbackTotal   map[string]int64 `json:"callback_total"`
	CallbackSuccess map[string]int64 `json:"callback_success"`
	CallbackFailed  map[string]int64 `json:"callback_failed"`

	// 系统指标
	LastResetTime      time.Time `json:"last_reset_time"`
	CurrentConnections int64     `json:"current_connections"`
	TotalRequests      int64     `json:"total_requests"`
}

// PaymentMetricsSummary 支付指标摘要
type PaymentMetricsSummary struct {
	Method            string        `json:"method"`
	CreateTotal       int64         `json:"create_total"`
	CreateSuccess     int64         `json:"create_success"`
	CreateFailed      int64         `json:"create_failed"`
	CreateSuccessRate float64       `json:"create_success_rate"`
	CreateAvgDuration time.Duration `json:"create_avg_duration"`
	CreateP95Duration time.Duration `json:"create_p95_duration"`

	QueryTotal       int64         `json:"query_total"`
	QuerySuccess     int64         `json:"query_success"`
	QueryFailed      int64         `json:"query_failed"`
	QuerySuccessRate float64       `json:"query_success_rate"`
	QueryAvgDuration time.Duration `json:"query_avg_duration"`

	CallbackTotal       int64   `json:"callback_total"`
	CallbackSuccess     int64   `json:"callback_success"`
	CallbackFailed      int64   `json:"callback_failed"`
	CallbackSuccessRate float64 `json:"callback_success_rate"`
}

// NewPaymentMetrics 创建支付指标实例
func NewPaymentMetrics() *PaymentMetrics {
	return &PaymentMetrics{
		CreatePaymentTotal:    make(map[string]int64),
		CreatePaymentSuccess:  make(map[string]int64),
		CreatePaymentFailed:   make(map[string]int64),
		CreatePaymentDuration: make(map[string][]time.Duration),

		QueryPaymentTotal:    make(map[string]int64),
		QueryPaymentSuccess:  make(map[string]int64),
		QueryPaymentFailed:   make(map[string]int64),
		QueryPaymentDuration: make(map[string][]time.Duration),

		CallbackTotal:   make(map[string]int64),
		CallbackSuccess: make(map[string]int64),
		CallbackFailed:  make(map[string]int64),

		LastResetTime: time.Now(),
	}
}

// RecordCreatePayment 记录创建支付指标
func (pm *PaymentMetrics) RecordCreatePayment(method model.PaymentMethod, status string, duration time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	methodStr := string(method)
	pm.CreatePaymentTotal[methodStr]++
	pm.TotalRequests++

	switch status {
	case "success":
		pm.CreatePaymentSuccess[methodStr]++
	case "error", "failed":
		pm.CreatePaymentFailed[methodStr]++
	}

	// 记录响应时间
	if pm.CreatePaymentDuration[methodStr] == nil {
		pm.CreatePaymentDuration[methodStr] = make([]time.Duration, 0)
	}

	// 保留最近1000条记录用于统计
	if len(pm.CreatePaymentDuration[methodStr]) >= 1000 {
		pm.CreatePaymentDuration[methodStr] = pm.CreatePaymentDuration[methodStr][100:]
	}
	pm.CreatePaymentDuration[methodStr] = append(pm.CreatePaymentDuration[methodStr], duration)

	// 记录慢请求日志
	if duration > 5*time.Second {
		logger.Warn("创建支付请求响应缓慢",
			zap.String("method", methodStr),
			zap.String("status", status),
			zap.Duration("duration", duration))
	}
}

// RecordQueryPayment 记录查询支付指标
func (pm *PaymentMetrics) RecordQueryPayment(method model.PaymentMethod, status string, duration time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	methodStr := string(method)
	pm.QueryPaymentTotal[methodStr]++
	pm.TotalRequests++

	switch status {
	case "success":
		pm.QueryPaymentSuccess[methodStr]++
	case "error", "failed":
		pm.QueryPaymentFailed[methodStr]++
	}

	// 记录响应时间
	if pm.QueryPaymentDuration[methodStr] == nil {
		pm.QueryPaymentDuration[methodStr] = make([]time.Duration, 0)
	}

	if len(pm.QueryPaymentDuration[methodStr]) >= 1000 {
		pm.QueryPaymentDuration[methodStr] = pm.QueryPaymentDuration[methodStr][100:]
	}
	pm.QueryPaymentDuration[methodStr] = append(pm.QueryPaymentDuration[methodStr], duration)

	// 记录慢请求日志
	if duration > 3*time.Second {
		logger.Warn("查询支付请求响应缓慢",
			zap.String("method", methodStr),
			zap.String("status", status),
			zap.Duration("duration", duration))
	}
}

// RecordCallback 记录回调指标
func (pm *PaymentMetrics) RecordCallback(method model.PaymentMethod, status string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	methodStr := string(method)
	pm.CallbackTotal[methodStr]++

	switch status {
	case "success":
		pm.CallbackSuccess[methodStr]++
	case "error", "failed":
		pm.CallbackFailed[methodStr]++
	}
}

// IncrementConnections 增加连接数
func (pm *PaymentMetrics) IncrementConnections() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.CurrentConnections++
}

// DecrementConnections 减少连接数
func (pm *PaymentMetrics) DecrementConnections() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if pm.CurrentConnections > 0 {
		pm.CurrentConnections--
	}
}

// GetSummary 获取指标摘要
func (pm *PaymentMetrics) GetSummary() []PaymentMetricsSummary {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	methods := []string{"alipay", "wechat", "unionpay"}
	summaries := make([]PaymentMetricsSummary, 0, len(methods))

	for _, method := range methods {
		summary := PaymentMetricsSummary{
			Method:          method,
			CreateTotal:     pm.CreatePaymentTotal[method],
			CreateSuccess:   pm.CreatePaymentSuccess[method],
			CreateFailed:    pm.CreatePaymentFailed[method],
			QueryTotal:      pm.QueryPaymentTotal[method],
			QuerySuccess:    pm.QueryPaymentSuccess[method],
			QueryFailed:     pm.QueryPaymentFailed[method],
			CallbackTotal:   pm.CallbackTotal[method],
			CallbackSuccess: pm.CallbackSuccess[method],
			CallbackFailed:  pm.CallbackFailed[method],
		}

		// 计算成功率
		if summary.CreateTotal > 0 {
			summary.CreateSuccessRate = float64(summary.CreateSuccess) / float64(summary.CreateTotal) * 100
		}
		if summary.QueryTotal > 0 {
			summary.QuerySuccessRate = float64(summary.QuerySuccess) / float64(summary.QueryTotal) * 100
		}
		if summary.CallbackTotal > 0 {
			summary.CallbackSuccessRate = float64(summary.CallbackSuccess) / float64(summary.CallbackTotal) * 100
		}

		// 计算平均响应时间
		if durations := pm.CreatePaymentDuration[method]; len(durations) > 0 {
			var total time.Duration
			for _, d := range durations {
				total += d
			}
			summary.CreateAvgDuration = total / time.Duration(len(durations))
			summary.CreateP95Duration = calculateP95(durations)
		}

		if durations := pm.QueryPaymentDuration[method]; len(durations) > 0 {
			var total time.Duration
			for _, d := range durations {
				total += d
			}
			summary.QueryAvgDuration = total / time.Duration(len(durations))
		}

		// 只有有数据的方法才添加到摘要中
		if summary.CreateTotal > 0 || summary.QueryTotal > 0 || summary.CallbackTotal > 0 {
			summaries = append(summaries, summary)
		}
	}

	return summaries
}

// calculateP95 计算P95响应时间
func calculateP95(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}

	// 简单排序（生产环境建议使用更高效的排序算法）
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)

	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j] > sorted[j+1] {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	p95Index := int(float64(len(sorted)) * 0.95)
	if p95Index >= len(sorted) {
		p95Index = len(sorted) - 1
	}

	return sorted[p95Index]
}

// Reset 重置指标
func (pm *PaymentMetrics) Reset() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.CreatePaymentTotal = make(map[string]int64)
	pm.CreatePaymentSuccess = make(map[string]int64)
	pm.CreatePaymentFailed = make(map[string]int64)
	pm.CreatePaymentDuration = make(map[string][]time.Duration)

	pm.QueryPaymentTotal = make(map[string]int64)
	pm.QueryPaymentSuccess = make(map[string]int64)
	pm.QueryPaymentFailed = make(map[string]int64)
	pm.QueryPaymentDuration = make(map[string][]time.Duration)

	pm.CallbackTotal = make(map[string]int64)
	pm.CallbackSuccess = make(map[string]int64)
	pm.CallbackFailed = make(map[string]int64)

	pm.LastResetTime = time.Now()
	pm.TotalRequests = 0

	logger.Info("支付系统指标已重置")
}

// LogSummary 记录指标摘要到日志
func (pm *PaymentMetrics) LogSummary() {
	summaries := pm.GetSummary()

	logger.Info("📊 支付系统指标摘要",
		zap.Int("method_count", len(summaries)),
		zap.Int64("current_connections", pm.CurrentConnections),
		zap.Int64("total_requests", pm.TotalRequests),
		zap.Time("last_reset", pm.LastResetTime))

	for _, summary := range summaries {
		logger.Info("支付方式指标",
			zap.String("method", summary.Method),
			zap.Int64("create_total", summary.CreateTotal),
			zap.Float64("create_success_rate", summary.CreateSuccessRate),
			zap.Duration("create_avg_duration", summary.CreateAvgDuration),
			zap.Duration("create_p95_duration", summary.CreateP95Duration),
			zap.Int64("query_total", summary.QueryTotal),
			zap.Float64("query_success_rate", summary.QuerySuccessRate),
			zap.Duration("query_avg_duration", summary.QueryAvgDuration),
			zap.Int64("callback_total", summary.CallbackTotal),
			zap.Float64("callback_success_rate", summary.CallbackSuccessRate))
	}
}

// StartPeriodicLogging 启动周期性指标日志记录
func (pm *PaymentMetrics) StartPeriodicLogging(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				pm.LogSummary()
			}
		}
	}()

	logger.Info("支付系统指标周期性日志记录已启动",
		zap.Duration("interval", interval))
}
