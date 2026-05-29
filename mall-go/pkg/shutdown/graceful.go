package shutdown

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mall-go/pkg/logger"
	"mall-go/pkg/metrics"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// GracefulShutdown 优雅关闭管理器
type GracefulShutdown struct {
	server           *http.Server
	db               *gorm.DB
	rdb              *redis.Client
	shutdownTimeout  time.Duration
	signalChan       chan os.Signal
	shutdownComplete chan struct{}
}

// NewGracefulShutdown 创建优雅关闭管理器
func NewGracefulShutdown(
	server *http.Server,
	db *gorm.DB,
	rdb *redis.Client,
	shutdownTimeout time.Duration,
) *GracefulShutdown {
	if shutdownTimeout == 0 {
		shutdownTimeout = 30 * time.Second // 默认30秒
	}

	return &GracefulShutdown{
		server:           server,
		db:               db,
		rdb:              rdb,
		shutdownTimeout:  shutdownTimeout,
		signalChan:       make(chan os.Signal, 1),
		shutdownComplete: make(chan struct{}),
	}
}

// Start 启动优雅关闭监听
func (gs *GracefulShutdown) Start() {
	// 注册信号处理
	signal.Notify(gs.signalChan, syscall.SIGTERM, syscall.SIGINT)

	// 在goroutine中监听信号
	go func() {
		sig := <-gs.signalChan
		logger.Info("收到关闭信号", zap.String("signal", sig.String()))
		gs.shutdown()
		close(gs.shutdownComplete)
	}()

	logger.Info("优雅关闭监听已启动", zap.Duration("timeout", gs.shutdownTimeout))
}

// Wait 等待关闭完成
func (gs *GracefulShutdown) Wait() {
	<-gs.shutdownComplete
}

// shutdown 执行关闭流程
func (gs *GracefulShutdown) shutdown() {
	logger.Info("开始执行优雅关闭流程", zap.Duration("timeout", gs.shutdownTimeout))

	// 创建带超时的context
	ctx, cancel := context.WithTimeout(context.Background(), gs.shutdownTimeout)
	defer cancel()

	// 1. 关闭HTTP服务器（停止接收新请求，等待现有请求完成）
	logger.Info("关闭HTTP服务器...")
	if err := gs.shutdownHTTPServer(ctx); err != nil {
		logger.Error("HTTP服务器关闭失败", zap.Error(err))
	} else {
		logger.Info("HTTP服务器已关闭")
	}

	// 2. 停止系统指标收集器
	logger.Info("停止系统指标收集器...")
	gs.stopMetricsCollector()
	logger.Info("系统指标收集器已停止")

	// 3. 关闭Redis连接
	logger.Info("关闭Redis连接...")
	if err := gs.closeRedis(ctx); err != nil {
		logger.Error("Redis连接关闭失败", zap.Error(err))
	} else {
		logger.Info("Redis连接已关闭")
	}

	// 4. 关闭数据库连接
	logger.Info("关闭数据库连接...")
	if err := gs.closeDatabase(); err != nil {
		logger.Error("数据库连接关闭失败", zap.Error(err))
	} else {
		logger.Info("数据库连接已关闭")
	}

	logger.Info("优雅关闭流程完成")
}

// shutdownHTTPServer 关闭HTTP服务器
func (gs *GracefulShutdown) shutdownHTTPServer(ctx context.Context) error {
	if gs.server == nil {
		return nil
	}

	// 设置一个稍短的超时，以便有时间清理其他资源
	shutdownCtx, cancel := context.WithTimeout(ctx, gs.shutdownTimeout-5*time.Second)
	defer cancel()

	return gs.server.Shutdown(shutdownCtx)
}

// closeRedis 关闭Redis连接
func (gs *GracefulShutdown) closeRedis(ctx context.Context) error {
	if gs.rdb == nil {
		return nil
	}

	// 关闭底层redis.Client
	return gs.rdb.Close()
}

// closeDatabase 关闭数据库连接
func (gs *GracefulShutdown) closeDatabase() error {
	if gs.db == nil {
		return nil
	}

	sqlDB, err := gs.db.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

// stopMetricsCollector 停止系统指标收集器
func (gs *GracefulShutdown) stopMetricsCollector() {
	metrics.StopSystemMetricsCollector()
}

// ForceShutdown 强制关闭（用于测试或紧急情况）
func (gs *GracefulShutdown) ForceShutdown() {
	logger.Warn("执行强制关闭")
	gs.shutdown()
	close(gs.shutdownComplete)
}

// IsShutdownComplete 检查关闭是否完成
func (gs *GracefulShutdown) IsShutdownComplete() bool {
	select {
	case <-gs.shutdownComplete:
		return true
	default:
		return false
	}
}
