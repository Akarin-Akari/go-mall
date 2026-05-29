package metrics

import (
	"runtime"
	"time"

	"mall-go/pkg/cache"
	"mall-go/pkg/database"
	"mall-go/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SystemMetricsCollector 系统指标收集器
type SystemMetricsCollector struct {
	db       *gorm.DB
	interval time.Duration
	stopChan chan struct{}
}

// NewSystemMetricsCollector 创建系统指标收集器
func NewSystemMetricsCollector(db *gorm.DB, interval time.Duration) *SystemMetricsCollector {
	if interval == 0 {
		interval = 15 * time.Second // 默认15秒
	}

	return &SystemMetricsCollector{
		db:       db,
		interval: interval,
		stopChan: make(chan struct{}),
	}
}

// Start 启动指标收集
func (smc *SystemMetricsCollector) Start() {
	ticker := time.NewTicker(smc.interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				smc.collectMetrics()
			case <-smc.stopChan:
				return
			}
		}
	}()

	logger.Info("系统指标收集器已启动", zap.Duration("interval", smc.interval))
}

// Stop 停止指标收集
func (smc *SystemMetricsCollector) Stop() {
	close(smc.stopChan)
	logger.Info("系统指标收集器已停止")
}

// collectMetrics 收集所有指标
func (smc *SystemMetricsCollector) collectMetrics() {
	pm := GetPrometheusMetrics()

	// 收集Go运行时指标
	smc.collectRuntimeMetrics(pm)

	// 收集数据库指标
	if smc.db != nil {
		smc.collectDatabaseMetrics(pm)
	}

	// 收集Redis指标
	smc.collectRedisMetrics(pm)
}

// collectRuntimeMetrics 收集Go运行时指标
func (smc *SystemMetricsCollector) collectRuntimeMetrics(pm *PrometheusMetrics) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Goroutines数量
	pm.GoRoutines.Set(float64(runtime.NumGoroutine()))

	// 内存指标
	pm.MemoryAlloc.Set(float64(m.Alloc))
	pm.MemorySys.Set(float64(m.Sys))

	// GC指标
	pm.GCPauseTotal.Add(float64(m.NumGC))
	if m.NumGC > 0 {
		// 最近一次GC暂停时间
		lastPause := m.PauseNs[(m.NumGC+255)%256]
		pm.GCPauseDuration.Observe(float64(lastPause) / 1e9) // 转换为秒
	}
}

// collectDatabaseMetrics 收集数据库指标
func (smc *SystemMetricsCollector) collectDatabaseMetrics(pm *PrometheusMetrics) {
	stats, err := database.GetPoolStats(smc.db)
	if err != nil {
		logger.Error("获取数据库连接池统计失败", zap.Error(err))
		return
	}

	pm.DBConnectionsOpen.Set(float64(stats.OpenConnections))
	pm.DBConnectionsInUse.Set(float64(stats.InUse))
	pm.DBConnectionsIdle.Set(float64(stats.Idle))
}

// collectRedisMetrics 收集Redis指标
func (smc *SystemMetricsCollector) collectRedisMetrics(pm *PrometheusMetrics) {
	redisClient := cache.GetRedisClient()
	if redisClient == nil {
		return
	}

	stats := redisClient.GetConnectionStats()
	pm.RedisConnectionsTotal.Set(float64(stats.TotalConns))
	pm.RedisConnectionsIdle.Set(float64(stats.IdleConns))
}

// 全局系统指标收集器
var GlobalSystemMetricsCollector *SystemMetricsCollector

// InitSystemMetricsCollector 初始化系统指标收集器
func InitSystemMetricsCollector(db *gorm.DB, interval time.Duration) {
	GlobalSystemMetricsCollector = NewSystemMetricsCollector(db, interval)
	GlobalSystemMetricsCollector.Start()
}

// StopSystemMetricsCollector 停止系统指标收集器
func StopSystemMetricsCollector() {
	if GlobalSystemMetricsCollector != nil {
		GlobalSystemMetricsCollector.Stop()
	}
}

