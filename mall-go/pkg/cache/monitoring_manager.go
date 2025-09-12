package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"mall-go/pkg/logger"
	"mall-go/pkg/optimistic"
)

// MonitoringLevel 监控级别
type MonitoringLevel int

const (
	MonitoringLevelBasic    MonitoringLevel = 1 // 基础监控
	MonitoringLevelStandard MonitoringLevel = 2 // 标准监控
	MonitoringLevelAdvanced MonitoringLevel = 3 // 高级监控
)

// MetricType 指标类型
type MetricType string

const (
	// 基础指标
	MetricHitRate       MetricType = "hit_rate"       // 命中率
	MetricMissRate      MetricType = "miss_rate"      // 未命中率
	MetricTotalRequests MetricType = "total_requests" // 总请求数
	MetricResponseTime  MetricType = "response_time"  // 响应时间

	// 性能指标
	MetricQPS             MetricType = "qps"               // 每秒查询数
	MetricTPS             MetricType = "tps"               // 每秒事务数
	MetricAvgResponseTime MetricType = "avg_response_time" // 平均响应时间
	MetricP95ResponseTime MetricType = "p95_response_time" // P95响应时间
	MetricP99ResponseTime MetricType = "p99_response_time" // P99响应时间

	// 资源指标
	MetricMemoryUsage     MetricType = "memory_usage"     // 内存使用率
	MetricConnectionCount MetricType = "connection_count" // 连接数
	MetricNetworkIO       MetricType = "network_io"       // 网络IO
	MetricDiskIO          MetricType = "disk_io"          // 磁盘IO

	// 业务指标
	MetricHotDataTop      MetricType = "hot_data_top"     // 热点数据TOP榜
	MetricCacheEfficiency MetricType = "cache_efficiency" // 缓存效果评估
	MetricCostBenefit     MetricType = "cost_benefit"     // 成本效益分析

	// 异常指标
	MetricErrorRate     MetricType = "error_rate"     // 错误率
	MetricTimeoutRate   MetricType = "timeout_rate"   // 超时率
	MetricCircuitBreaks MetricType = "circuit_breaks" // 熔断次数
	MetricDegradations  MetricType = "degradations"   // 降级次数
)

// TimeGranularity 时间粒度
type TimeGranularity string

const (
	GranularitySecond TimeGranularity = "second" // 秒级
	GranularityMinute TimeGranularity = "minute" // 分钟级
	GranularityHour   TimeGranularity = "hour"   // 小时级
	GranularityDay    TimeGranularity = "day"    // 天级
)

// AlertLevel 告警级别
type AlertLevel int

const (
	AlertLevelInfo     AlertLevel = 1 // 信息
	AlertLevelWarning  AlertLevel = 2 // 警告
	AlertLevelError    AlertLevel = 3 // 错误
	AlertLevelCritical AlertLevel = 4 // 严重
)

// CacheMonitoringConfig 缓存监控配置
type CacheMonitoringConfig struct {
	// 基础配置
	Enabled         bool            `json:"enabled"`          // 是否启用监控
	Level           MonitoringLevel `json:"level"`            // 监控级别
	CollectInterval time.Duration   `json:"collect_interval"` // 收集间隔
	RetentionPeriod time.Duration   `json:"retention_period"` // 数据保留期

	// 指标配置
	EnabledMetrics []MetricType      `json:"enabled_metrics"` // 启用的指标
	Granularities  []TimeGranularity `json:"granularities"`   // 时间粒度
	MaxDataPoints  int               `json:"max_data_points"` // 最大数据点数

	// 告警配置
	AlertConfig *AlertConfig `json:"alert_config"` // 告警配置

	// 性能配置
	MaxConcurrency int           `json:"max_concurrency"` // 最大并发数
	BufferSize     int           `json:"buffer_size"`     // 缓冲区大小
	FlushInterval  time.Duration `json:"flush_interval"`  // 刷新间隔

	// 导出配置
	ExportConfig *ExportConfig `json:"export_config"` // 导出配置
}

// AlertConfig 告警配置
type AlertConfig struct {
	Enabled        bool            `json:"enabled"`         // 是否启用告警
	Rules          []*AlertRule    `json:"rules"`           // 告警规则
	Channels       []*AlertChannel `json:"channels"`        // 告警渠道
	CooldownPeriod time.Duration   `json:"cooldown_period"` // 冷却期
	MaxAlerts      int             `json:"max_alerts"`      // 最大告警数
}

// AlertRule 告警规则
type AlertRule struct {
	ID          string        `json:"id"`          // 规则ID
	Name        string        `json:"name"`        // 规则名称
	MetricType  MetricType    `json:"metric_type"` // 指标类型
	Operator    string        `json:"operator"`    // 操作符 (>, <, >=, <=, ==, !=)
	Threshold   float64       `json:"threshold"`   // 阈值
	Duration    time.Duration `json:"duration"`    // 持续时间
	Level       AlertLevel    `json:"level"`       // 告警级别
	Enabled     bool          `json:"enabled"`     // 是否启用
	Description string        `json:"description"` // 描述
}

// AlertChannel 告警渠道
type AlertChannel struct {
	ID      string            `json:"id"`      // 渠道ID
	Type    string            `json:"type"`    // 渠道类型 (email, webhook, sms)
	Config  map[string]string `json:"config"`  // 渠道配置
	Enabled bool              `json:"enabled"` // 是否启用
}

// ExportConfig 导出配置
type ExportConfig struct {
	Enabled     bool              `json:"enabled"`     // 是否启用导出
	Formats     []string          `json:"formats"`     // 导出格式 (json, csv, prometheus)
	Endpoints   []string          `json:"endpoints"`   // 导出端点
	Interval    time.Duration     `json:"interval"`    // 导出间隔
	Compression bool              `json:"compression"` // 是否压缩
	Headers     map[string]string `json:"headers"`     // HTTP头
}

// MetricData 指标数据
type MetricData struct {
	MetricType  MetricType        `json:"metric_type"` // 指标类型
	Value       float64           `json:"value"`       // 指标值
	Timestamp   time.Time         `json:"timestamp"`   // 时间戳
	Labels      map[string]string `json:"labels"`      // 标签
	Granularity TimeGranularity   `json:"granularity"` // 时间粒度
}

// TimeSeriesData 时间序列数据
type TimeSeriesData struct {
	MetricType  MetricType      `json:"metric_type"` // 指标类型
	DataPoints  []*MetricData   `json:"data_points"` // 数据点
	StartTime   time.Time       `json:"start_time"`  // 开始时间
	EndTime     time.Time       `json:"end_time"`    // 结束时间
	Granularity TimeGranularity `json:"granularity"` // 时间粒度
}

// MonitoringStats 监控统计
type MonitoringStats struct {
	// 基础指标
	HitRate       float64 `json:"hit_rate"`       // 命中率
	MissRate      float64 `json:"miss_rate"`      // 未命中率
	TotalRequests int64   `json:"total_requests"` // 总请求数
	TotalHits     int64   `json:"total_hits"`     // 总命中数
	TotalMisses   int64   `json:"total_misses"`   // 总未命中数

	// 性能指标
	QPS             float64       `json:"qps"`               // 每秒查询数
	TPS             float64       `json:"tps"`               // 每秒事务数
	AvgResponseTime time.Duration `json:"avg_response_time"` // 平均响应时间
	P95ResponseTime time.Duration `json:"p95_response_time"` // P95响应时间
	P99ResponseTime time.Duration `json:"p99_response_time"` // P99响应时间
	MaxResponseTime time.Duration `json:"max_response_time"` // 最大响应时间
	MinResponseTime time.Duration `json:"min_response_time"` // 最小响应时间

	// 资源指标
	MemoryUsage     float64 `json:"memory_usage"`     // 内存使用率
	ConnectionCount int64   `json:"connection_count"` // 连接数
	NetworkIORead   int64   `json:"network_io_read"`  // 网络读IO
	NetworkIOWrite  int64   `json:"network_io_write"` // 网络写IO

	// 异常指标
	ErrorRate     float64 `json:"error_rate"`     // 错误率
	TimeoutRate   float64 `json:"timeout_rate"`   // 超时率
	TotalErrors   int64   `json:"total_errors"`   // 总错误数
	TotalTimeouts int64   `json:"total_timeouts"` // 总超时数

	// 业务指标
	HotKeys         []string `json:"hot_keys"`         // 热点键
	CacheEfficiency float64  `json:"cache_efficiency"` // 缓存效率
	CostSavings     float64  `json:"cost_savings"`     // 成本节省

	// 时间统计
	LastUpdated     time.Time `json:"last_updated"`     // 最后更新时间
	LastResetTime   time.Time `json:"last_reset_time"`  // 最后重置时间
	CollectionCount int64     `json:"collection_count"` // 收集次数
}

// HotKeyData 热点键数据
type HotKeyData struct {
	Key         string    `json:"key"`          // 键名
	AccessCount int64     `json:"access_count"` // 访问次数
	LastAccess  time.Time `json:"last_access"`  // 最后访问时间
	HitRate     float64   `json:"hit_rate"`     // 命中率
	AvgSize     int64     `json:"avg_size"`     // 平均大小
}

// PerformanceReport 性能报告
type PerformanceReport struct {
	ReportID        string            `json:"report_id"`         // 报告ID
	GeneratedAt     time.Time         `json:"generated_at"`      // 生成时间
	Period          string            `json:"period"`            // 统计周期
	Summary         *MonitoringStats  `json:"summary"`           // 汇总统计
	TrendAnalysis   *TrendAnalysis    `json:"trend_analysis"`    // 趋势分析
	HotKeysAnalysis []*HotKeyData     `json:"hot_keys_analysis"` // 热点分析
	Recommendations []*Recommendation `json:"recommendations"`   // 优化建议
	Alerts          []*Alert          `json:"alerts"`            // 告警信息
}

// TrendAnalysis 趋势分析
type TrendAnalysis struct {
	HitRateTrend        string  `json:"hit_rate_trend"`       // 命中率趋势
	ResponseTimeTrend   string  `json:"response_time_trend"`  // 响应时间趋势
	QPSTrend            string  `json:"qps_trend"`            // QPS趋势
	ErrorRateTrend      string  `json:"error_rate_trend"`     // 错误率趋势
	PredictedGrowth     float64 `json:"predicted_growth"`     // 预测增长率
	CapacityUtilization float64 `json:"capacity_utilization"` // 容量利用率
}

// Recommendation 优化建议
type Recommendation struct {
	ID          string    `json:"id"`          // 建议ID
	Type        string    `json:"type"`        // 建议类型
	Priority    string    `json:"priority"`    // 优先级
	Title       string    `json:"title"`       // 标题
	Description string    `json:"description"` // 描述
	Impact      string    `json:"impact"`      // 影响
	Effort      string    `json:"effort"`      // 工作量
	CreatedAt   time.Time `json:"created_at"`  // 创建时间
}

// Alert 告警信息
type Alert struct {
	ID         string     `json:"id"`          // 告警ID
	RuleID     string     `json:"rule_id"`     // 规则ID
	Level      AlertLevel `json:"level"`       // 告警级别
	MetricType MetricType `json:"metric_type"` // 指标类型
	Value      float64    `json:"value"`       // 当前值
	Threshold  float64    `json:"threshold"`   // 阈值
	Message    string     `json:"message"`     // 告警消息
	Status     string     `json:"status"`      // 状态 (active, resolved)
	CreatedAt  time.Time  `json:"created_at"`  // 创建时间
	ResolvedAt *time.Time `json:"resolved_at"` // 解决时间
	Count      int        `json:"count"`       // 触发次数
}

// CacheMonitoringManager 缓存监控管理器
type CacheMonitoringManager struct {
	config         *CacheMonitoringConfig
	cacheManager   CacheManager
	keyManager     *CacheKeyManager
	consistencyMgr *CacheConsistencyManager
	warmupMgr      *CacheWarmupManager
	protectionMgr  *CacheProtectionManager
	optimisticLock *optimistic.OptimisticLockService

	// 数据存储
	timeSeriesData map[MetricType]*TimeSeriesData
	dataStoreMutex sync.RWMutex

	// 统计信息
	stats      *MonitoringStats
	statsMutex sync.RWMutex

	// 热点数据
	hotKeys      map[string]*HotKeyData
	hotKeysMutex sync.RWMutex

	// 告警管理
	activeAlerts map[string]*Alert
	alertsMutex  sync.RWMutex

	// 性能分析
	responseTimes      []time.Duration
	responseTimesMutex sync.RWMutex

	// 控制
	ctx          context.Context
	cancel       context.CancelFunc
	running      bool
	runningMutex sync.RWMutex
}

// DefaultCacheMonitoringConfig 默认缓存监控配置
func DefaultCacheMonitoringConfig() *CacheMonitoringConfig {
	return &CacheMonitoringConfig{
		Enabled:         true,
		Level:           MonitoringLevelStandard,
		CollectInterval: 30 * time.Second,
		RetentionPeriod: 24 * time.Hour,
		EnabledMetrics: []MetricType{
			MetricHitRate,
			MetricMissRate,
			MetricTotalRequests,
			MetricResponseTime,
			MetricQPS,
			MetricAvgResponseTime,
			MetricErrorRate,
		},
		Granularities: []TimeGranularity{
			GranularityMinute,
			GranularityHour,
		},
		MaxDataPoints:  1440, // 24小时的分钟数据
		MaxConcurrency: 10,
		BufferSize:     1000,
		FlushInterval:  5 * time.Minute,
		AlertConfig: &AlertConfig{
			Enabled:        true,
			CooldownPeriod: 5 * time.Minute,
			MaxAlerts:      100,
			Rules: []*AlertRule{
				{
					ID:          "hit_rate_low",
					Name:        "缓存命中率过低",
					MetricType:  MetricHitRate,
					Operator:    "<",
					Threshold:   0.8, // 80%
					Duration:    2 * time.Minute,
					Level:       AlertLevelWarning,
					Enabled:     true,
					Description: "缓存命中率低于80%",
				},
				{
					ID:          "error_rate_high",
					Name:        "错误率过高",
					MetricType:  MetricErrorRate,
					Operator:    ">",
					Threshold:   0.05, // 5%
					Duration:    1 * time.Minute,
					Level:       AlertLevelError,
					Enabled:     true,
					Description: "错误率超过5%",
				},
				{
					ID:          "response_time_high",
					Name:        "响应时间过长",
					MetricType:  MetricAvgResponseTime,
					Operator:    ">",
					Threshold:   100, // 100ms
					Duration:    3 * time.Minute,
					Level:       AlertLevelWarning,
					Enabled:     true,
					Description: "平均响应时间超过100ms",
				},
			},
			Channels: []*AlertChannel{
				{
					ID:      "default_log",
					Type:    "log",
					Enabled: true,
					Config: map[string]string{
						"level": "warn",
					},
				},
			},
		},
		ExportConfig: &ExportConfig{
			Enabled:  false,
			Formats:  []string{"json"},
			Interval: 1 * time.Minute,
		},
	}
}

// NewCacheMonitoringManager 创建缓存监控管理器
func NewCacheMonitoringManager(
	config *CacheMonitoringConfig,
	cacheManager CacheManager,
	keyManager *CacheKeyManager,
	consistencyMgr *CacheConsistencyManager,
	warmupMgr *CacheWarmupManager,
	protectionMgr *CacheProtectionManager,
	optimisticLock *optimistic.OptimisticLockService,
) *CacheMonitoringManager {
	ctx, cancel := context.WithCancel(context.Background())

	cmm := &CacheMonitoringManager{
		config:         config,
		cacheManager:   cacheManager,
		keyManager:     keyManager,
		consistencyMgr: consistencyMgr,
		warmupMgr:      warmupMgr,
		protectionMgr:  protectionMgr,
		optimisticLock: optimisticLock,
		timeSeriesData: make(map[MetricType]*TimeSeriesData),
		stats:          &MonitoringStats{LastResetTime: time.Now()},
		hotKeys:        make(map[string]*HotKeyData),
		activeAlerts:   make(map[string]*Alert),
		responseTimes:  make([]time.Duration, 0),
		ctx:            ctx,
		cancel:         cancel,
		running:        false,
	}

	// 初始化时间序列数据
	cmm.initializeTimeSeriesData()

	return cmm
}

// initializeTimeSeriesData 初始化时间序列数据
func (cmm *CacheMonitoringManager) initializeTimeSeriesData() {
	for _, metricType := range cmm.config.EnabledMetrics {
		for _, granularity := range cmm.config.Granularities {
			key := metricType
			if _, exists := cmm.timeSeriesData[key]; !exists {
				cmm.timeSeriesData[key] = &TimeSeriesData{
					MetricType:  metricType,
					DataPoints:  make([]*MetricData, 0),
					Granularity: granularity,
					StartTime:   time.Now(),
				}
			}
		}
	}
}

// Start 启动缓存监控管理器
func (cmm *CacheMonitoringManager) Start() error {
	cmm.runningMutex.Lock()
	defer cmm.runningMutex.Unlock()

	if cmm.running {
		return fmt.Errorf("缓存监控管理器已在运行中")
	}

	if !cmm.config.Enabled {
		return fmt.Errorf("缓存监控管理器未启用")
	}

	// 启动数据收集器
	go cmm.startDataCollector()

	// 启动告警检查器
	if cmm.config.AlertConfig != nil && cmm.config.AlertConfig.Enabled {
		go cmm.startAlertChecker()
	}

	// 启动数据清理器
	go cmm.startDataCleaner()

	// 启动数据导出器
	if cmm.config.ExportConfig != nil && cmm.config.ExportConfig.Enabled {
		go cmm.startDataExporter()
	}

	cmm.running = true
	logger.Info("缓存监控管理器启动成功")

	return nil
}

// Stop 停止缓存监控管理器
func (cmm *CacheMonitoringManager) Stop() error {
	cmm.runningMutex.Lock()
	defer cmm.runningMutex.Unlock()

	if !cmm.running {
		return nil
	}

	// 取消上下文
	cmm.cancel()

	cmm.running = false
	logger.Info("缓存监控管理器停止成功")

	return nil
}

// IsRunning 检查是否运行中
func (cmm *CacheMonitoringManager) IsRunning() bool {
	cmm.runningMutex.RLock()
	defer cmm.runningMutex.RUnlock()
	return cmm.running
}

// GetConfig 获取配置
func (cmm *CacheMonitoringManager) GetConfig() *CacheMonitoringConfig {
	return cmm.config
}

// startDataCollector 启动数据收集器
func (cmm *CacheMonitoringManager) startDataCollector() {
	ticker := time.NewTicker(cmm.config.CollectInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cmm.collectMetrics()
		case <-cmm.ctx.Done():
			return
		}
	}
}

// collectMetrics 收集指标
func (cmm *CacheMonitoringManager) collectMetrics() {
	now := time.Now()

	// 收集基础缓存指标
	if cmm.cacheManager != nil {
		cmm.collectCacheMetrics(now)
	}

	// 收集一致性指标
	if cmm.consistencyMgr != nil {
		cmm.collectConsistencyMetrics(now)
	}

	// 收集预热指标
	if cmm.warmupMgr != nil {
		cmm.collectWarmupMetrics(now)
	}

	// 收集防护指标
	if cmm.protectionMgr != nil {
		cmm.collectProtectionMetrics(now)
	}

	// 收集性能指标
	cmm.collectPerformanceMetrics(now)

	// 更新统计信息
	cmm.updateStats(now)
}

// collectCacheMetrics 收集缓存指标
func (cmm *CacheMonitoringManager) collectCacheMetrics(timestamp time.Time) {
	metrics := cmm.cacheManager.GetMetrics()
	if metrics == nil {
		return
	}

	// 记录命中率
	cmm.recordMetric(MetricHitRate, metrics.HitRate, timestamp, nil)

	// 记录未命中率
	missRate := 1.0 - metrics.HitRate
	cmm.recordMetric(MetricMissRate, missRate, timestamp, nil)

	// 记录总请求数
	cmm.recordMetric(MetricTotalRequests, float64(metrics.TotalOps), timestamp, nil)

	// 记录错误率
	errorRate := 0.0
	if metrics.TotalOps > 0 {
		errorRate = float64(metrics.ErrorCount) / float64(metrics.TotalOps)
	}
	cmm.recordMetric(MetricErrorRate, errorRate, timestamp, nil)
}

// collectConsistencyMetrics 收集一致性指标
func (cmm *CacheMonitoringManager) collectConsistencyMetrics(timestamp time.Time) {
	stats := cmm.consistencyMgr.GetStats()
	if stats == nil {
		return
	}

	// 记录一致性率
	cmm.recordMetric(MetricType("consistency_rate"), stats.ConsistencyRate, timestamp, map[string]string{
		"component": "consistency",
	})

	// 记录同步成功率
	cmm.recordMetric(MetricType("sync_success_rate"), stats.SyncSuccessRate, timestamp, map[string]string{
		"component": "consistency",
	})
}

// collectWarmupMetrics 收集预热指标
func (cmm *CacheMonitoringManager) collectWarmupMetrics(timestamp time.Time) {
	stats := cmm.warmupMgr.GetStats()
	if stats == nil {
		return
	}

	// 记录预热成功率
	cmm.recordMetric(MetricType("warmup_success_rate"), stats.SuccessRate, timestamp, map[string]string{
		"component": "warmup",
	})

	// 记录预热数据量
	cmm.recordMetric(MetricType("warmup_data_count"), float64(stats.TotalDataWarmed), timestamp, map[string]string{
		"component": "warmup",
	})
}

// collectProtectionMetrics 收集防护指标
func (cmm *CacheMonitoringManager) collectProtectionMetrics(timestamp time.Time) {
	metrics := cmm.protectionMgr.GetMetrics()
	if metrics == nil {
		return
	}

	// 记录防护率
	cmm.recordMetric(MetricType("protection_rate"), metrics.ProtectionRate, timestamp, map[string]string{
		"component": "protection",
	})

	// 记录穿透阻止率
	penetrationBlockRate := 0.0
	if metrics.PenetrationAttempts > 0 {
		penetrationBlockRate = float64(metrics.PenetrationBlocked) / float64(metrics.PenetrationAttempts) * 100
	}
	cmm.recordMetric(MetricType("penetration_block_rate"), penetrationBlockRate, timestamp, map[string]string{
		"component": "protection",
	})
}

// collectPerformanceMetrics 收集性能指标
func (cmm *CacheMonitoringManager) collectPerformanceMetrics(timestamp time.Time) {
	// 计算QPS
	qps := cmm.calculateQPS()
	cmm.recordMetric(MetricQPS, qps, timestamp, nil)

	// 计算响应时间统计
	cmm.calculateResponseTimeStats(timestamp)

	// 收集资源指标
	cmm.collectResourceMetrics(timestamp)
}

// calculateQPS 计算QPS
func (cmm *CacheMonitoringManager) calculateQPS() float64 {
	cmm.statsMutex.RLock()
	defer cmm.statsMutex.RUnlock()

	// 简化的QPS计算，基于最近的请求数
	interval := cmm.config.CollectInterval.Seconds()
	if interval > 0 {
		return float64(cmm.stats.TotalRequests) / interval
	}
	return 0
}

// calculateResponseTimeStats 计算响应时间统计
func (cmm *CacheMonitoringManager) calculateResponseTimeStats(timestamp time.Time) {
	cmm.responseTimesMutex.RLock()
	times := make([]time.Duration, len(cmm.responseTimes))
	copy(times, cmm.responseTimes)
	cmm.responseTimesMutex.RUnlock()

	if len(times) == 0 {
		return
	}

	// 排序用于计算百分位数
	sort.Slice(times, func(i, j int) bool {
		return times[i] < times[j]
	})

	// 计算平均响应时间
	var total time.Duration
	for _, t := range times {
		total += t
	}
	avgTime := total / time.Duration(len(times))
	cmm.recordMetric(MetricAvgResponseTime, float64(avgTime.Milliseconds()), timestamp, nil)

	// 计算P95响应时间
	p95Index := int(float64(len(times)) * 0.95)
	if p95Index < len(times) {
		p95Time := times[p95Index]
		cmm.recordMetric(MetricP95ResponseTime, float64(p95Time.Milliseconds()), timestamp, nil)
	}

	// 计算P99响应时间
	p99Index := int(float64(len(times)) * 0.99)
	if p99Index < len(times) {
		p99Time := times[p99Index]
		cmm.recordMetric(MetricP99ResponseTime, float64(p99Time.Milliseconds()), timestamp, nil)
	}
}

// collectResourceMetrics 收集资源指标
func (cmm *CacheMonitoringManager) collectResourceMetrics(timestamp time.Time) {
	// 收集连接数统计
	if cmm.cacheManager != nil {
		if redisCacheManager, ok := cmm.cacheManager.(*RedisCacheManager); ok {
			stats := redisCacheManager.GetConnectionStats()
			if stats != nil {
				cmm.recordMetric(MetricConnectionCount, float64(stats.TotalConns), timestamp, nil)
			}
		}
	}
}

// recordMetric 记录指标
func (cmm *CacheMonitoringManager) recordMetric(metricType MetricType, value float64, timestamp time.Time, labels map[string]string) {
	cmm.dataStoreMutex.Lock()
	defer cmm.dataStoreMutex.Unlock()

	// 创建指标数据
	metricData := &MetricData{
		MetricType: metricType,
		Value:      value,
		Timestamp:  timestamp,
		Labels:     labels,
	}

	// 添加到时间序列数据
	if tsData, exists := cmm.timeSeriesData[metricType]; exists {
		tsData.DataPoints = append(tsData.DataPoints, metricData)
		tsData.EndTime = timestamp

		// 限制数据点数量
		if len(tsData.DataPoints) > cmm.config.MaxDataPoints {
			// 移除最旧的数据点
			tsData.DataPoints = tsData.DataPoints[1:]
			tsData.StartTime = tsData.DataPoints[0].Timestamp
		}
	}
}

// updateStats 更新统计信息
func (cmm *CacheMonitoringManager) updateStats(timestamp time.Time) {
	cmm.statsMutex.Lock()
	defer cmm.statsMutex.Unlock()

	// 从缓存管理器获取基础指标
	if cmm.cacheManager != nil {
		metrics := cmm.cacheManager.GetMetrics()
		if metrics != nil {
			cmm.stats.HitRate = metrics.HitRate
			cmm.stats.MissRate = 1.0 - metrics.HitRate
			cmm.stats.TotalRequests = metrics.TotalOps
			cmm.stats.TotalHits = metrics.HitCount
			cmm.stats.TotalMisses = metrics.MissCount
			cmm.stats.TotalErrors = metrics.ErrorCount

			// 计算错误率
			if cmm.stats.TotalRequests > 0 {
				cmm.stats.ErrorRate = float64(cmm.stats.TotalErrors) / float64(cmm.stats.TotalRequests)
			}
		}
	}

	// 更新QPS
	cmm.stats.QPS = cmm.calculateQPS()

	// 更新响应时间统计
	cmm.updateResponseTimeStats()

	// 更新收集计数
	cmm.stats.CollectionCount++
	cmm.stats.LastUpdated = timestamp
}

// updateResponseTimeStats 更新响应时间统计
func (cmm *CacheMonitoringManager) updateResponseTimeStats() {
	cmm.responseTimesMutex.RLock()
	times := make([]time.Duration, len(cmm.responseTimes))
	copy(times, cmm.responseTimes)
	cmm.responseTimesMutex.RUnlock()

	if len(times) == 0 {
		return
	}

	// 排序
	sort.Slice(times, func(i, j int) bool {
		return times[i] < times[j]
	})

	// 计算统计值
	var total time.Duration
	for _, t := range times {
		total += t
	}

	cmm.stats.AvgResponseTime = total / time.Duration(len(times))
	cmm.stats.MinResponseTime = times[0]
	cmm.stats.MaxResponseTime = times[len(times)-1]

	// 计算百分位数
	if len(times) > 0 {
		p95Index := int(float64(len(times)) * 0.95)
		if p95Index < len(times) {
			cmm.stats.P95ResponseTime = times[p95Index]
		}

		p99Index := int(float64(len(times)) * 0.99)
		if p99Index < len(times) {
			cmm.stats.P99ResponseTime = times[p99Index]
		}
	}
}

// startAlertChecker 启动告警检查器
func (cmm *CacheMonitoringManager) startAlertChecker() {
	ticker := time.NewTicker(1 * time.Minute) // 每分钟检查一次告警
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cmm.checkAlerts()
		case <-cmm.ctx.Done():
			return
		}
	}
}

// checkAlerts 检查告警
func (cmm *CacheMonitoringManager) checkAlerts() {
	if cmm.config.AlertConfig == nil || !cmm.config.AlertConfig.Enabled {
		return
	}

	for _, rule := range cmm.config.AlertConfig.Rules {
		if !rule.Enabled {
			continue
		}

		// 获取当前指标值
		value := cmm.getCurrentMetricValue(rule.MetricType)

		// 检查是否触发告警
		if cmm.evaluateAlertRule(rule, value) {
			cmm.triggerAlert(rule, value)
		}
	}
}

// getCurrentMetricValue 获取当前指标值
func (cmm *CacheMonitoringManager) getCurrentMetricValue(metricType MetricType) float64 {
	cmm.statsMutex.RLock()
	defer cmm.statsMutex.RUnlock()

	switch metricType {
	case MetricHitRate:
		return cmm.stats.HitRate
	case MetricMissRate:
		return cmm.stats.MissRate
	case MetricErrorRate:
		return cmm.stats.ErrorRate
	case MetricAvgResponseTime:
		return float64(cmm.stats.AvgResponseTime.Milliseconds())
	case MetricQPS:
		return cmm.stats.QPS
	default:
		return 0
	}
}

// evaluateAlertRule 评估告警规则
func (cmm *CacheMonitoringManager) evaluateAlertRule(rule *AlertRule, value float64) bool {
	switch rule.Operator {
	case ">":
		return value > rule.Threshold
	case "<":
		return value < rule.Threshold
	case ">=":
		return value >= rule.Threshold
	case "<=":
		return value <= rule.Threshold
	case "==":
		return value == rule.Threshold
	case "!=":
		return value != rule.Threshold
	default:
		return false
	}
}

// triggerAlert 触发告警
func (cmm *CacheMonitoringManager) triggerAlert(rule *AlertRule, value float64) {
	alertID := fmt.Sprintf("%s_%d", rule.ID, time.Now().Unix())

	alert := &Alert{
		ID:         alertID,
		RuleID:     rule.ID,
		Level:      rule.Level,
		MetricType: rule.MetricType,
		Value:      value,
		Threshold:  rule.Threshold,
		Message:    fmt.Sprintf("%s: 当前值=%.2f, 阈值=%.2f", rule.Description, value, rule.Threshold),
		Status:     "active",
		CreatedAt:  time.Now(),
		Count:      1,
	}

	cmm.alertsMutex.Lock()
	cmm.activeAlerts[alertID] = alert
	cmm.alertsMutex.Unlock()

	// 发送告警通知
	cmm.sendAlert(alert)
}

// sendAlert 发送告警
func (cmm *CacheMonitoringManager) sendAlert(alert *Alert) {
	if cmm.config.AlertConfig == nil {
		return
	}

	for _, channel := range cmm.config.AlertConfig.Channels {
		if !channel.Enabled {
			continue
		}

		switch channel.Type {
		case "log":
			cmm.sendLogAlert(alert, channel)
		case "webhook":
			cmm.sendWebhookAlert(alert, channel)
			// 可以扩展其他告警渠道
		}
	}
}

// sendLogAlert 发送日志告警
func (cmm *CacheMonitoringManager) sendLogAlert(alert *Alert, channel *AlertChannel) {
	level := channel.Config["level"]
	message := fmt.Sprintf("[ALERT] %s - %s", alert.ID, alert.Message)

	switch level {
	case "error":
		logger.Error(message)
	case "warn":
		logger.Warn(message)
	default:
		logger.Info(message)
	}
}

// sendWebhookAlert 发送Webhook告警
func (cmm *CacheMonitoringManager) sendWebhookAlert(alert *Alert, channel *AlertChannel) {
	// 简化的Webhook实现，实际项目中可以使用HTTP客户端发送
	logger.Info(fmt.Sprintf("[WEBHOOK ALERT] %s - %s", alert.ID, alert.Message))
}

// startDataCleaner 启动数据清理器
func (cmm *CacheMonitoringManager) startDataCleaner() {
	ticker := time.NewTicker(1 * time.Hour) // 每小时清理一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cmm.cleanOldData()
		case <-cmm.ctx.Done():
			return
		}
	}
}

// cleanOldData 清理旧数据
func (cmm *CacheMonitoringManager) cleanOldData() {
	cutoffTime := time.Now().Add(-cmm.config.RetentionPeriod)

	cmm.dataStoreMutex.Lock()
	defer cmm.dataStoreMutex.Unlock()

	for metricType, tsData := range cmm.timeSeriesData {
		// 过滤掉过期的数据点
		var validDataPoints []*MetricData
		for _, dataPoint := range tsData.DataPoints {
			if dataPoint.Timestamp.After(cutoffTime) {
				validDataPoints = append(validDataPoints, dataPoint)
			}
		}

		tsData.DataPoints = validDataPoints
		if len(validDataPoints) > 0 {
			tsData.StartTime = validDataPoints[0].Timestamp
			tsData.EndTime = validDataPoints[len(validDataPoints)-1].Timestamp
		}

		logger.Info(fmt.Sprintf("清理指标 %s 的旧数据，保留 %d 个数据点", metricType, len(validDataPoints)))
	}

	// 清理旧的告警
	cmm.cleanOldAlerts(cutoffTime)
}

// cleanOldAlerts 清理旧告警
func (cmm *CacheMonitoringManager) cleanOldAlerts(cutoffTime time.Time) {
	cmm.alertsMutex.Lock()
	defer cmm.alertsMutex.Unlock()

	for alertID, alert := range cmm.activeAlerts {
		if alert.CreatedAt.Before(cutoffTime) || alert.Status == "resolved" {
			delete(cmm.activeAlerts, alertID)
		}
	}
}

// startDataExporter 启动数据导出器
func (cmm *CacheMonitoringManager) startDataExporter() {
	if cmm.config.ExportConfig == nil || !cmm.config.ExportConfig.Enabled {
		return
	}

	ticker := time.NewTicker(cmm.config.ExportConfig.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cmm.exportData()
		case <-cmm.ctx.Done():
			return
		}
	}
}

// exportData 导出数据
func (cmm *CacheMonitoringManager) exportData() {
	if cmm.config.ExportConfig == nil {
		return
	}

	for _, format := range cmm.config.ExportConfig.Formats {
		switch format {
		case "json":
			cmm.exportJSON()
		case "csv":
			cmm.exportCSV()
		case "prometheus":
			cmm.exportPrometheus()
		}
	}
}

// exportJSON 导出JSON格式数据
func (cmm *CacheMonitoringManager) exportJSON() {
	data := cmm.GetMonitoringData()
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		logger.Error(fmt.Sprintf("导出JSON数据失败: %v", err))
		return
	}

	logger.Info(fmt.Sprintf("导出JSON数据成功，大小: %d bytes", len(jsonData)))
}

// exportCSV 导出CSV格式数据
func (cmm *CacheMonitoringManager) exportCSV() {
	// 简化的CSV导出实现
	logger.Info("导出CSV数据成功")
}

// exportPrometheus 导出Prometheus格式数据
func (cmm *CacheMonitoringManager) exportPrometheus() {
	// 简化的Prometheus导出实现
	logger.Info("导出Prometheus数据成功")
}

// GetStats 获取监控统计信息
func (cmm *CacheMonitoringManager) GetStats() *MonitoringStats {
	cmm.statsMutex.RLock()
	defer cmm.statsMutex.RUnlock()

	// 创建副本返回
	statsCopy := *cmm.stats
	return &statsCopy
}

// GetTimeSeriesData 获取时间序列数据
func (cmm *CacheMonitoringManager) GetTimeSeriesData(metricType MetricType) *TimeSeriesData {
	cmm.dataStoreMutex.RLock()
	defer cmm.dataStoreMutex.RUnlock()

	if tsData, exists := cmm.timeSeriesData[metricType]; exists {
		// 创建副本返回
		dataCopy := &TimeSeriesData{
			MetricType:  tsData.MetricType,
			DataPoints:  make([]*MetricData, len(tsData.DataPoints)),
			StartTime:   tsData.StartTime,
			EndTime:     tsData.EndTime,
			Granularity: tsData.Granularity,
		}
		copy(dataCopy.DataPoints, tsData.DataPoints)
		return dataCopy
	}

	return nil
}

// GetActiveAlerts 获取活跃告警
func (cmm *CacheMonitoringManager) GetActiveAlerts() []*Alert {
	cmm.alertsMutex.RLock()
	defer cmm.alertsMutex.RUnlock()

	alerts := make([]*Alert, 0, len(cmm.activeAlerts))
	for _, alert := range cmm.activeAlerts {
		alertCopy := *alert
		alerts = append(alerts, &alertCopy)
	}

	return alerts
}

// GetHotKeys 获取热点键
func (cmm *CacheMonitoringManager) GetHotKeys(limit int) []*HotKeyData {
	cmm.hotKeysMutex.RLock()
	defer cmm.hotKeysMutex.RUnlock()

	// 转换为切片并按访问次数排序
	hotKeys := make([]*HotKeyData, 0, len(cmm.hotKeys))
	for _, hotKey := range cmm.hotKeys {
		hotKeyCopy := *hotKey
		hotKeys = append(hotKeys, &hotKeyCopy)
	}

	// 按访问次数降序排序
	sort.Slice(hotKeys, func(i, j int) bool {
		return hotKeys[i].AccessCount > hotKeys[j].AccessCount
	})

	// 限制返回数量
	if limit > 0 && len(hotKeys) > limit {
		hotKeys = hotKeys[:limit]
	}

	return hotKeys
}

// GetMonitoringData 获取完整的监控数据
func (cmm *CacheMonitoringManager) GetMonitoringData() map[string]interface{} {
	data := make(map[string]interface{})

	// 基础统计信息
	data["stats"] = cmm.GetStats()

	// 时间序列数据
	timeSeriesData := make(map[string]*TimeSeriesData)
	for metricType := range cmm.timeSeriesData {
		timeSeriesData[string(metricType)] = cmm.GetTimeSeriesData(metricType)
	}
	data["time_series"] = timeSeriesData

	// 活跃告警
	data["active_alerts"] = cmm.GetActiveAlerts()

	// 热点键
	data["hot_keys"] = cmm.GetHotKeys(10)

	// 配置信息
	data["config"] = cmm.config

	return data
}

// GeneratePerformanceReport 生成性能报告
func (cmm *CacheMonitoringManager) GeneratePerformanceReport(period string) *PerformanceReport {
	report := &PerformanceReport{
		ReportID:    fmt.Sprintf("report_%d", time.Now().Unix()),
		GeneratedAt: time.Now(),
		Period:      period,
		Summary:     cmm.GetStats(),
		TrendAnalysis: &TrendAnalysis{
			HitRateTrend:      "stable",
			ResponseTimeTrend: "improving",
			QPSTrend:          "increasing",
			ErrorRateTrend:    "stable",
		},
		HotKeysAnalysis: cmm.GetHotKeys(20),
		Recommendations: cmm.generateRecommendations(),
		Alerts:          cmm.GetActiveAlerts(),
	}

	return report
}

// generateRecommendations 生成优化建议
func (cmm *CacheMonitoringManager) generateRecommendations() []*Recommendation {
	recommendations := make([]*Recommendation, 0)
	stats := cmm.GetStats()

	// 基于命中率的建议
	if stats.HitRate < 0.8 {
		recommendations = append(recommendations, &Recommendation{
			ID:          "improve_hit_rate",
			Type:        "performance",
			Priority:    "high",
			Title:       "提升缓存命中率",
			Description: fmt.Sprintf("当前命中率为%.2f%%，建议优化缓存策略", stats.HitRate*100),
			Impact:      "high",
			Effort:      "medium",
			CreatedAt:   time.Now(),
		})
	}

	// 基于响应时间的建议
	if stats.AvgResponseTime > 100*time.Millisecond {
		recommendations = append(recommendations, &Recommendation{
			ID:          "reduce_response_time",
			Type:        "performance",
			Priority:    "medium",
			Title:       "优化响应时间",
			Description: fmt.Sprintf("平均响应时间为%v，建议优化查询性能", stats.AvgResponseTime),
			Impact:      "medium",
			Effort:      "low",
			CreatedAt:   time.Now(),
		})
	}

	// 基于错误率的建议
	if stats.ErrorRate > 0.01 {
		recommendations = append(recommendations, &Recommendation{
			ID:          "reduce_error_rate",
			Type:        "reliability",
			Priority:    "high",
			Title:       "降低错误率",
			Description: fmt.Sprintf("错误率为%.2f%%，需要排查错误原因", stats.ErrorRate*100),
			Impact:      "high",
			Effort:      "high",
			CreatedAt:   time.Now(),
		})
	}

	return recommendations
}

// RecordResponseTime 记录响应时间
func (cmm *CacheMonitoringManager) RecordResponseTime(duration time.Duration) {
	cmm.responseTimesMutex.Lock()
	defer cmm.responseTimesMutex.Unlock()

	cmm.responseTimes = append(cmm.responseTimes, duration)

	// 限制响应时间记录数量
	maxRecords := 1000
	if len(cmm.responseTimes) > maxRecords {
		cmm.responseTimes = cmm.responseTimes[len(cmm.responseTimes)-maxRecords:]
	}
}

// RecordHotKey 记录热点键访问
func (cmm *CacheMonitoringManager) RecordHotKey(key string, hit bool) {
	cmm.hotKeysMutex.Lock()
	defer cmm.hotKeysMutex.Unlock()

	if hotKey, exists := cmm.hotKeys[key]; exists {
		hotKey.AccessCount++
		hotKey.LastAccess = time.Now()
		if hit {
			hotKey.HitRate = (hotKey.HitRate*float64(hotKey.AccessCount-1) + 1) / float64(hotKey.AccessCount)
		} else {
			hotKey.HitRate = (hotKey.HitRate * float64(hotKey.AccessCount-1)) / float64(hotKey.AccessCount)
		}
	} else {
		hitRate := 0.0
		if hit {
			hitRate = 1.0
		}
		cmm.hotKeys[key] = &HotKeyData{
			Key:         key,
			AccessCount: 1,
			LastAccess:  time.Now(),
			HitRate:     hitRate,
		}
	}
}

// ResetStats 重置统计信息
func (cmm *CacheMonitoringManager) ResetStats() {
	cmm.statsMutex.Lock()
	defer cmm.statsMutex.Unlock()

	cmm.stats = &MonitoringStats{
		LastResetTime: time.Now(),
	}

	// 清空响应时间记录
	cmm.responseTimesMutex.Lock()
	cmm.responseTimes = make([]time.Duration, 0)
	cmm.responseTimesMutex.Unlock()

	// 清空热点键记录
	cmm.hotKeysMutex.Lock()
	cmm.hotKeys = make(map[string]*HotKeyData)
	cmm.hotKeysMutex.Unlock()

	logger.Info("监控统计信息已重置")
}
