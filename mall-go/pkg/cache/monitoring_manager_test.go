package cache

import (
	"testing"
	"time"

	"mall-go/pkg/optimistic"

	"github.com/stretchr/testify/assert"
)

// 使用已有的SharedMockCacheManager

// setupTestMonitoringManager 设置测试监控管理器
func setupTestMonitoringManager(t *testing.T) *CacheMonitoringManager {
	config := DefaultCacheMonitoringConfig()
	config.CollectInterval = 100 * time.Millisecond // 快速收集用于测试
	config.RetentionPeriod = 1 * time.Hour

	mockCacheManager := &SharedMockCacheManager{}
	mockCacheManager.On("GetMetrics").Return(&CacheMetrics{
		HitCount:    100,
		MissCount:   20,
		HitRate:     0.83,
		TotalOps:    120,
		ErrorCount:  2,
		LastUpdated: time.Now(),
	})

	InitKeyManager("test")
	keyManager := GetKeyManager()
	var optimisticLock *optimistic.OptimisticLockService = nil

	cmm := NewCacheMonitoringManager(
		config,
		mockCacheManager,
		keyManager,
		nil, // consistencyMgr
		nil, // warmupMgr
		nil, // protectionMgr
		optimisticLock,
	)

	return cmm
}

func TestNewCacheMonitoringManager(t *testing.T) {
	cmm := setupTestMonitoringManager(t)

	assert.NotNil(t, cmm)
	assert.NotNil(t, cmm.config)
	assert.NotNil(t, cmm.stats)
	assert.NotNil(t, cmm.timeSeriesData)
	assert.NotNil(t, cmm.hotKeys)
	assert.NotNil(t, cmm.activeAlerts)
	assert.False(t, cmm.IsRunning())
}

func TestCacheMonitoringManagerStartStop(t *testing.T) {
	cmm := setupTestMonitoringManager(t)

	// 测试启动
	err := cmm.Start()
	assert.NoError(t, err)
	assert.True(t, cmm.IsRunning())

	// 等待一段时间让数据收集器运行
	time.Sleep(200 * time.Millisecond)

	// 测试停止
	err = cmm.Stop()
	assert.NoError(t, err)
	assert.False(t, cmm.IsRunning())

	// 测试重复启动
	err = cmm.Start()
	assert.NoError(t, err)

	err = cmm.Start()
	assert.Error(t, err)

	cmm.Stop()
}

func TestCacheMonitoringManagerMetricsCollection(t *testing.T) {
	cmm := setupTestMonitoringManager(t)

	err := cmm.Start()
	assert.NoError(t, err)
	defer cmm.Stop()

	// 等待数据收集
	time.Sleep(300 * time.Millisecond)

	// 检查统计信息
	stats := cmm.GetStats()
	assert.NotNil(t, stats)
	assert.Equal(t, 0.83, stats.HitRate)
	assert.Equal(t, 0.17, stats.MissRate)
	assert.True(t, stats.CollectionCount > 0)

	// 检查时间序列数据
	hitRateData := cmm.GetTimeSeriesData(MetricHitRate)
	assert.NotNil(t, hitRateData)
	assert.True(t, len(hitRateData.DataPoints) > 0)
}

func TestCacheMonitoringManagerResponseTimeRecording(t *testing.T) {
	cmm := setupTestMonitoringManager(t)

	// 记录响应时间
	cmm.RecordResponseTime(50 * time.Millisecond)
	cmm.RecordResponseTime(100 * time.Millisecond)
	cmm.RecordResponseTime(150 * time.Millisecond)

	// 更新响应时间统计
	cmm.updateResponseTimeStats()

	stats := cmm.GetStats()
	assert.True(t, stats.AvgResponseTime > 0)
	assert.True(t, stats.MinResponseTime > 0)
	assert.True(t, stats.MaxResponseTime > 0)
}

func TestCacheMonitoringManagerHotKeyTracking(t *testing.T) {
	cmm := setupTestMonitoringManager(t)

	// 记录热点键访问
	cmm.RecordHotKey("product:123", true)
	cmm.RecordHotKey("product:123", true)
	cmm.RecordHotKey("product:456", false)
	cmm.RecordHotKey("product:123", false)

	// 获取热点键
	hotKeys := cmm.GetHotKeys(10)
	assert.Len(t, hotKeys, 2)

	// 检查排序（按访问次数降序）
	assert.Equal(t, "product:123", hotKeys[0].Key)
	assert.Equal(t, int64(3), hotKeys[0].AccessCount)
	assert.Equal(t, "product:456", hotKeys[1].Key)
	assert.Equal(t, int64(1), hotKeys[1].AccessCount)
}

func TestCacheMonitoringManagerAlertSystem(t *testing.T) {
	cmm := setupTestMonitoringManager(t)

	// 创建一个低阈值的告警规则用于测试
	rule := &AlertRule{
		ID:          "test_alert",
		Name:        "测试告警",
		MetricType:  MetricHitRate,
		Operator:    "<",
		Threshold:   0.9, // 90%
		Duration:    1 * time.Second,
		Level:       AlertLevelWarning,
		Enabled:     true,
		Description: "命中率过低测试",
	}

	// 触发告警
	cmm.triggerAlert(rule, 0.8)

	// 检查活跃告警
	alerts := cmm.GetActiveAlerts()
	assert.Len(t, alerts, 1)
	assert.Equal(t, rule.ID, alerts[0].RuleID)
	assert.Equal(t, "active", alerts[0].Status)
}

func TestCacheMonitoringManagerPerformanceReport(t *testing.T) {
	cmm := setupTestMonitoringManager(t)

	// 添加一些测试数据
	cmm.RecordResponseTime(50 * time.Millisecond)
	cmm.RecordHotKey("product:123", true)

	// 生成性能报告
	report := cmm.GeneratePerformanceReport("daily")

	assert.NotNil(t, report)
	assert.NotEmpty(t, report.ReportID)
	assert.Equal(t, "daily", report.Period)
	assert.NotNil(t, report.Summary)
	assert.NotNil(t, report.TrendAnalysis)
	assert.NotNil(t, report.Recommendations)
}

func TestCacheMonitoringManagerDataExport(t *testing.T) {
	cmm := setupTestMonitoringManager(t)

	// 获取监控数据
	data := cmm.GetMonitoringData()

	assert.NotNil(t, data)
	assert.Contains(t, data, "stats")
	assert.Contains(t, data, "time_series")
	assert.Contains(t, data, "active_alerts")
	assert.Contains(t, data, "hot_keys")
	assert.Contains(t, data, "config")
}

func TestCacheMonitoringManagerResetStats(t *testing.T) {
	cmm := setupTestMonitoringManager(t)

	// 添加一些数据
	cmm.RecordResponseTime(100 * time.Millisecond)
	cmm.RecordHotKey("test:key", true)

	// 重置统计
	cmm.ResetStats()

	// 检查数据已清空
	stats := cmm.GetStats()
	assert.Equal(t, int64(0), stats.TotalRequests)
	assert.Equal(t, int64(0), stats.CollectionCount)

	hotKeys := cmm.GetHotKeys(10)
	assert.Len(t, hotKeys, 0)
}
