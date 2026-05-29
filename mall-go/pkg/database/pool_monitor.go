package database

import (
	"database/sql"
	"time"

	"mall-go/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PoolStats 连接池统计信息
type PoolStats struct {
	MaxOpenConnections int           `json:"max_open_connections"` // 最大打开连接数
	OpenConnections    int           `json:"open_connections"`     // 当前打开连接数
	InUse              int           `json:"in_use"`               // 正在使用的连接数
	Idle               int           `json:"idle"`                 // 空闲连接数
	WaitCount          int64         `json:"wait_count"`           // 等待连接的总次数
	WaitDuration       time.Duration `json:"wait_duration"`        // 等待连接的总时间
	MaxIdleClosed      int64         `json:"max_idle_closed"`      // 因超过最大空闲连接数而关闭的连接数
	MaxIdleTimeClosed  int64         `json:"max_idle_time_closed"` // 因超过最大空闲时间而关闭的连接数
	MaxLifetimeClosed  int64         `json:"max_lifetime_closed"`  // 因超过最大生命周期而关闭的连接数
}

// GetPoolStats 获取连接池统计信息
func GetPoolStats(db *gorm.DB) (*PoolStats, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	stats := sqlDB.Stats()

	return &PoolStats{
		MaxOpenConnections: stats.MaxOpenConnections,
		OpenConnections:    stats.OpenConnections,
		InUse:              stats.InUse,
		Idle:               stats.Idle,
		WaitCount:          stats.WaitCount,
		WaitDuration:       stats.WaitDuration,
		MaxIdleClosed:      stats.MaxIdleClosed,
		MaxIdleTimeClosed:  stats.MaxIdleTimeClosed,
		MaxLifetimeClosed:  stats.MaxLifetimeClosed,
	}, nil
}

// LogPoolStats 记录连接池统计信息
func LogPoolStats(db *gorm.DB) {
	stats, err := GetPoolStats(db)
	if err != nil {
		logger.Error("获取连接池统计信息失败", zap.Error(err))
		return
	}

	logger.Info("数据库连接池状态",
		zap.Int("max_open", stats.MaxOpenConnections),
		zap.Int("open", stats.OpenConnections),
		zap.Int("in_use", stats.InUse),
		zap.Int("idle", stats.Idle),
		zap.Int64("wait_count", stats.WaitCount),
		zap.Duration("wait_duration", stats.WaitDuration),
		zap.Int64("max_idle_closed", stats.MaxIdleClosed),
		zap.Int64("max_idle_time_closed", stats.MaxIdleTimeClosed),
		zap.Int64("max_lifetime_closed", stats.MaxLifetimeClosed),
	)

	// 检查连接池健康状态
	checkPoolHealth(stats)
}

// checkPoolHealth 检查连接池健康状态
func checkPoolHealth(stats *PoolStats) {
	// 检查连接池是否接近饱和
	if stats.OpenConnections >= stats.MaxOpenConnections*9/10 {
		logger.Warn("数据库连接池接近饱和",
			zap.Int("open", stats.OpenConnections),
			zap.Int("max", stats.MaxOpenConnections),
			zap.Float64("usage_percent", float64(stats.OpenConnections)/float64(stats.MaxOpenConnections)*100),
		)
	}

	// 检查等待时间是否过长
	if stats.WaitCount > 0 {
		avgWaitTime := stats.WaitDuration / time.Duration(stats.WaitCount)
		if avgWaitTime > 100*time.Millisecond {
			logger.Warn("数据库连接等待时间过长",
				zap.Int64("wait_count", stats.WaitCount),
				zap.Duration("avg_wait_time", avgWaitTime),
			)
		}
	}

	// 检查空闲连接是否过少
	if stats.Idle == 0 && stats.InUse > 0 {
		logger.Warn("数据库连接池无空闲连接",
			zap.Int("in_use", stats.InUse),
			zap.Int("idle", stats.Idle),
		)
	}
}

// StartPoolMonitor 启动连接池监控
func StartPoolMonitor(db *gorm.DB, interval time.Duration) {
	if interval == 0 {
		interval = 30 * time.Second // 默认30秒
	}

	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for range ticker.C {
			LogPoolStats(db)
		}
	}()

	logger.Info("数据库连接池监控已启动", zap.Duration("interval", interval))
}

// ConfigureConnectionPool 配置连接池参数
func ConfigureConnectionPool(db *gorm.DB, maxOpen, maxIdle int, maxLifetime time.Duration) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetConnMaxLifetime(maxLifetime)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute) // 空闲连接最大存活时间

	logger.Info("数据库连接池配置完成",
		zap.Int("max_open_conns", maxOpen),
		zap.Int("max_idle_conns", maxIdle),
		zap.Duration("conn_max_lifetime", maxLifetime),
	)

	return nil
}

// PingDB 测试数据库连接
func PingDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	return sqlDB.Ping()
}

// GetDBStats 获取数据库统计信息（用于健康检查）
func GetDBStats(db *gorm.DB) (*sql.DBStats, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	stats := sqlDB.Stats()
	return &stats, nil
}

// IsPoolHealthy 检查连接池是否健康
func IsPoolHealthy(db *gorm.DB) bool {
	stats, err := GetPoolStats(db)
	if err != nil {
		logger.Error("获取连接池状态失败", zap.Error(err))
		return false
	}

	// 检查是否有可用连接
	if stats.OpenConnections == 0 {
		logger.Error("数据库连接池无可用连接")
		return false
	}

	// 检查连接池是否过载
	if stats.OpenConnections >= stats.MaxOpenConnections {
		logger.Warn("数据库连接池已满")
		return false
	}

	return true
}

