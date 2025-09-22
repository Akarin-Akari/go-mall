package errors

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"mall-go/pkg/logger"

	"go.uber.org/zap"
)

// ErrorMetrics 错误指标统计
type ErrorMetrics struct {
	mu sync.RWMutex

	// 总体统计
	TotalErrors     int64                    `json:"total_errors"`      // 总错误数
	TotalRequests   int64                    `json:"total_requests"`    // 总请求数
	ErrorRate       float64                  `json:"error_rate"`        // 错误率
	LastResetTime   time.Time                `json:"last_reset_time"`   // 上次重置时间

	// 按错误码统计
	ErrorCodeStats  map[ErrorCode]int64      `json:"error_code_stats"`  // 错误码统计
	
	// 按错误级别统计
	ErrorLevelStats map[ErrorLevel]int64     `json:"error_level_stats"` // 错误级别统计
	
	// 按错误分类统计
	CategoryStats   map[ErrorCategory]int64  `json:"category_stats"`    // 错误分类统计
	
	// 按时间统计
	HourlyStats     map[int]int64           `json:"hourly_stats"`      // 小时统计
	DailyStats      map[string]int64        `json:"daily_stats"`       // 日统计
	
	// 用户错误统计
	UserErrorStats  map[uint]int64          `json:"user_error_stats"`  // 用户错误统计
	
	// 路径错误统计
	PathErrorStats  map[string]int64        `json:"path_error_stats"`  // 路径错误统计
	
	// 最近错误
	RecentErrors    []*ErrorRecord          `json:"recent_errors"`     // 最近错误记录
}

// ErrorRecord 错误记录
type ErrorRecord struct {
	Timestamp   time.Time     `json:"timestamp"`   // 时间戳
	Code        ErrorCode     `json:"code"`        // 错误码
	Level       ErrorLevel    `json:"level"`       // 错误级别
	Category    ErrorCategory `json:"category"`    // 错误分类
	Message     string        `json:"message"`     // 错误消息
	UserID      uint          `json:"user_id"`     // 用户ID
	RequestID   string        `json:"request_id"`  // 请求ID
	Path        string        `json:"path"`        // 请求路径
	Method      string        `json:"method"`      // 请求方法
	ClientIP    string        `json:"client_ip"`   // 客户端IP
	UserAgent   string        `json:"user_agent"`  // 用户代理
}

// ErrorMonitor 错误监控器
type ErrorMonitor struct {
	metrics    *ErrorMetrics
	collectors []ErrorCollector
	alertRules []AlertRule
	ticker     *time.Ticker
	stopCh     chan struct{}
}

// ErrorCollector 错误收集器接口
type ErrorCollector interface {
	CollectError(record *ErrorRecord) error
	Name() string
}

// AlertRule 告警规则
type AlertRule struct {
	Name        string                              `json:"name"`         // 规则名称
	Description string                              `json:"description"`  // 规则描述
	Condition   func(metrics *ErrorMetrics) bool    `json:"-"`           // 告警条件
	Action      func(metrics *ErrorMetrics) error   `json:"-"`           // 告警动作
	Enabled     bool                                `json:"enabled"`     // 是否启用
	LastTrigger *time.Time                          `json:"last_trigger"` // 上次触发时间
	CoolDown    time.Duration                       `json:"cool_down"`   // 冷却时间
}

// NewErrorMonitor 创建错误监控器
func NewErrorMonitor() *ErrorMonitor {
	return &ErrorMonitor{
		metrics: &ErrorMetrics{
			ErrorCodeStats:  make(map[ErrorCode]int64),
			ErrorLevelStats: make(map[ErrorLevel]int64),
			CategoryStats:   make(map[ErrorCategory]int64),
			HourlyStats:     make(map[int]int64),
			DailyStats:      make(map[string]int64),
			UserErrorStats:  make(map[uint]int64),
			PathErrorStats:  make(map[string]int64),
			RecentErrors:    make([]*ErrorRecord, 0),
			LastResetTime:   time.Now(),
		},
		collectors: make([]ErrorCollector, 0),
		alertRules: make([]AlertRule, 0),
		stopCh:     make(chan struct{}),
	}
}

// RecordError 记录错误
func (em *ErrorMonitor) RecordError(err *BusinessError, path, method, clientIP, userAgent string) {
	em.metrics.mu.Lock()
	defer em.metrics.mu.Unlock()

	// 更新总体统计
	em.metrics.TotalErrors++
	
	// 更新错误码统计
	em.metrics.ErrorCodeStats[err.Code]++
	
	// 更新错误级别统计
	em.metrics.ErrorLevelStats[err.Level]++
	
	// 更新错误分类统计
	em.metrics.CategoryStats[err.Category]++
	
	// 更新小时统计
	hour := err.Timestamp.Hour()
	em.metrics.HourlyStats[hour]++
	
	// 更新日统计
	day := err.Timestamp.Format("2006-01-02")
	em.metrics.DailyStats[day]++
	
	// 更新用户错误统计
	if err.UserID > 0 {
		em.metrics.UserErrorStats[err.UserID]++
	}
	
	// 更新路径错误统计
	if path != "" {
		em.metrics.PathErrorStats[path]++
	}
	
	// 创建错误记录
	record := &ErrorRecord{
		Timestamp: err.Timestamp,
		Code:      err.Code,
		Level:     err.Level,
		Category:  err.Category,
		Message:   err.Message,
		UserID:    err.UserID,
		RequestID: err.RequestID,
		Path:      path,
		Method:    method,
		ClientIP:  clientIP,
		UserAgent: userAgent,
	}
	
	// 添加到最近错误列表
	em.metrics.RecentErrors = append(em.metrics.RecentErrors, record)
	
	// 保持最近错误列表长度不超过1000
	if len(em.metrics.RecentErrors) > 1000 {
		em.metrics.RecentErrors = em.metrics.RecentErrors[100:]
	}
	
	// 发送到收集器
	for _, collector := range em.collectors {
		go func(c ErrorCollector) {
			if err := c.CollectError(record); err != nil {
				logger.Error("错误收集器失败",
					zap.String("collector", c.Name()),
					zap.Error(err))
			}
		}(collector)
	}
	
	// 检查告警规则
	em.checkAlertRules()
}

// RecordRequest 记录请求
func (em *ErrorMonitor) RecordRequest() {
	em.metrics.mu.Lock()
	defer em.metrics.mu.Unlock()
	
	em.metrics.TotalRequests++
	
	// 更新错误率
	if em.metrics.TotalRequests > 0 {
		em.metrics.ErrorRate = float64(em.metrics.TotalErrors) / float64(em.metrics.TotalRequests) * 100
	}
}

// GetMetrics 获取错误指标
func (em *ErrorMonitor) GetMetrics() *ErrorMetrics {
	em.metrics.mu.RLock()
	defer em.metrics.mu.RUnlock()
	
	// 返回副本
	metrics := &ErrorMetrics{
		TotalErrors:     em.metrics.TotalErrors,
		TotalRequests:   em.metrics.TotalRequests,
		ErrorRate:       em.metrics.ErrorRate,
		LastResetTime:   em.metrics.LastResetTime,
		ErrorCodeStats:  make(map[ErrorCode]int64),
		ErrorLevelStats: make(map[ErrorLevel]int64),
		CategoryStats:   make(map[ErrorCategory]int64),
		HourlyStats:     make(map[int]int64),
		DailyStats:      make(map[string]int64),
		UserErrorStats:  make(map[uint]int64),
		PathErrorStats:  make(map[string]int64),
		RecentErrors:    make([]*ErrorRecord, len(em.metrics.RecentErrors)),
	}
	
	// 复制map数据
	for k, v := range em.metrics.ErrorCodeStats {
		metrics.ErrorCodeStats[k] = v
	}
	for k, v := range em.metrics.ErrorLevelStats {
		metrics.ErrorLevelStats[k] = v
	}
	for k, v := range em.metrics.CategoryStats {
		metrics.CategoryStats[k] = v
	}
	for k, v := range em.metrics.HourlyStats {
		metrics.HourlyStats[k] = v
	}
	for k, v := range em.metrics.DailyStats {
		metrics.DailyStats[k] = v
	}
	for k, v := range em.metrics.UserErrorStats {
		metrics.UserErrorStats[k] = v
	}
	for k, v := range em.metrics.PathErrorStats {
		metrics.PathErrorStats[k] = v
	}
	
	// 复制最近错误
	copy(metrics.RecentErrors, em.metrics.RecentErrors)
	
	return metrics
}

// ResetMetrics 重置指标
func (em *ErrorMonitor) ResetMetrics() {
	em.metrics.mu.Lock()
	defer em.metrics.mu.Unlock()
	
	em.metrics.TotalErrors = 0
	em.metrics.TotalRequests = 0
	em.metrics.ErrorRate = 0
	em.metrics.LastResetTime = time.Now()
	em.metrics.ErrorCodeStats = make(map[ErrorCode]int64)
	em.metrics.ErrorLevelStats = make(map[ErrorLevel]int64)
	em.metrics.CategoryStats = make(map[ErrorCategory]int64)
	em.metrics.HourlyStats = make(map[int]int64)
	em.metrics.DailyStats = make(map[string]int64)
	em.metrics.UserErrorStats = make(map[uint]int64)
	em.metrics.PathErrorStats = make(map[string]int64)
	em.metrics.RecentErrors = make([]*ErrorRecord, 0)
	
	logger.Info("错误监控指标已重置")
}

// AddCollector 添加错误收集器
func (em *ErrorMonitor) AddCollector(collector ErrorCollector) {
	em.collectors = append(em.collectors, collector)
	logger.Info("添加错误收集器", zap.String("name", collector.Name()))
}

// AddAlertRule 添加告警规则
func (em *ErrorMonitor) AddAlertRule(rule AlertRule) {
	em.alertRules = append(em.alertRules, rule)
	logger.Info("添加告警规则", zap.String("name", rule.Name))
}

// checkAlertRules 检查告警规则
func (em *ErrorMonitor) checkAlertRules() {
	now := time.Now()
	
	for i, rule := range em.alertRules {
		if !rule.Enabled {
			continue
		}
		
		// 检查冷却时间
		if rule.LastTrigger != nil && now.Sub(*rule.LastTrigger) < rule.CoolDown {
			continue
		}
		
		// 检查告警条件
		if rule.Condition(em.metrics) {
			// 触发告警
			if err := rule.Action(em.metrics); err != nil {
				logger.Error("告警动作执行失败",
					zap.String("rule", rule.Name),
					zap.Error(err))
			} else {
				logger.Warn("触发告警",
					zap.String("rule", rule.Name),
					zap.String("description", rule.Description))
				
				// 更新最后触发时间
				em.alertRules[i].LastTrigger = &now
			}
		}
	}
}

// Start 启动错误监控
func (em *ErrorMonitor) Start(ctx context.Context) {
	// 启动定时任务
	em.ticker = time.NewTicker(1 * time.Minute)
	
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-em.stopCh:
				return
			case <-em.ticker.C:
				// 定期检查告警规则
				em.metrics.mu.RLock()
				em.checkAlertRules()
				em.metrics.mu.RUnlock()
				
				// 定期输出统计信息
				em.logStatistics()
			}
		}
	}()
	
	logger.Info("错误监控系统已启动")
}

// Stop 停止错误监控
func (em *ErrorMonitor) Stop() {
	if em.ticker != nil {
		em.ticker.Stop()
	}
	close(em.stopCh)
	logger.Info("错误监控系统已停止")
}

// logStatistics 输出统计信息
func (em *ErrorMonitor) logStatistics() {
	metrics := em.GetMetrics()
	
	logger.Info("错误监控统计",
		zap.Int64("total_errors", metrics.TotalErrors),
		zap.Int64("total_requests", metrics.TotalRequests),
		zap.Float64("error_rate", metrics.ErrorRate),
		zap.Int("error_types", len(metrics.ErrorCodeStats)),
		zap.Int("recent_errors", len(metrics.RecentErrors)))
}

// ExportMetrics 导出指标数据
func (em *ErrorMonitor) ExportMetrics() ([]byte, error) {
	metrics := em.GetMetrics()
	return json.MarshalIndent(metrics, "", "  ")
}

// 默认告警规则
var DefaultAlertRules = []AlertRule{
	{
		Name:        "高错误率告警",
		Description: "错误率超过5%时触发告警",
		Condition: func(metrics *ErrorMetrics) bool {
			return metrics.ErrorRate > 5.0 && metrics.TotalRequests > 100
		},
		Action: func(metrics *ErrorMetrics) error {
			logger.Error("高错误率告警",
				zap.Float64("error_rate", metrics.ErrorRate),
				zap.Int64("total_errors", metrics.TotalErrors),
				zap.Int64("total_requests", metrics.TotalRequests))
			return nil
		},
		Enabled:  true,
		CoolDown: 5 * time.Minute,
	},
	{
		Name:        "系统错误激增告警",
		Description: "系统错误数量快速增长时触发告警",
		Condition: func(metrics *ErrorMetrics) bool {
			systemErrors := metrics.CategoryStats[CategorySystem]
			return systemErrors > 50
		},
		Action: func(metrics *ErrorMetrics) error {
			systemErrors := metrics.CategoryStats[CategorySystem]
			logger.Error("系统错误激增告警",
				zap.Int64("system_errors", systemErrors))
			return nil
		},
		Enabled:  true,
		CoolDown: 10 * time.Minute,
	},
	{
		Name:        "支付错误告警",
		Description: "支付相关错误过多时触发告警",
		Condition: func(metrics *ErrorMetrics) bool {
			paymentErrors := metrics.CategoryStats[CategoryPayment]
			return paymentErrors > 10
		},
		Action: func(metrics *ErrorMetrics) error {
			paymentErrors := metrics.CategoryStats[CategoryPayment]
			logger.Error("支付错误告警",
				zap.Int64("payment_errors", paymentErrors))
			return nil
		},
		Enabled:  true,
		CoolDown: 3 * time.Minute,
	},
}

// 全局错误监控实例
var GlobalErrorMonitor = NewErrorMonitor()

// init 初始化默认告警规则
func init() {
	for _, rule := range DefaultAlertRules {
		GlobalErrorMonitor.AddAlertRule(rule)
	}
}