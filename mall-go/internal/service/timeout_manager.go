package service

import (
	"context"
	"time"

	"mall-go/internal/config"
	"mall-go/pkg/logger"

	"go.uber.org/zap"
)

// TimeoutManager 超时管理器
type TimeoutManager struct {
	config *config.AddressConfig
}

// NewTimeoutManager 创建超时管理器
func NewTimeoutManager(cfg *config.AddressConfig) *TimeoutManager {
	if cfg == nil {
		cfg = config.DefaultAddressConfig()
	}
	return &TimeoutManager{
		config: cfg,
	}
}

// WithOperationTimeout 为操作添加超时控制
func (tm *TimeoutManager) WithOperationTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, tm.config.OperationTimeout)
}

// WithQueryTimeout 为查询添加超时控制
func (tm *TimeoutManager) WithQueryTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, tm.config.QueryTimeout)
}

// WithTransactionTimeout 为事务添加超时控制
func (tm *TimeoutManager) WithTransactionTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, tm.config.TransactionTimeout)
}

// WithDatabaseTimeout 为数据库操作添加超时控制
func (tm *TimeoutManager) WithDatabaseTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, tm.config.DatabaseTimeout)
}

// WithCustomTimeout 使用自定义超时时间
func (tm *TimeoutManager) WithCustomTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}

// CheckTimeout 检查上下文是否超时
func (tm *TimeoutManager) CheckTimeout(ctx context.Context, operation string) error {
	select {
	case <-ctx.Done():
		err := ctx.Err()
		if err == context.DeadlineExceeded {
			logger.Warn("操作超时",
				zap.String("operation", operation),
				zap.Error(err))
			return ErrOperationTimeout
		}
		return err
	default:
		return nil
	}
}

// MonitorTimeout 监控操作超时
func (tm *TimeoutManager) MonitorTimeout(ctx context.Context, operation string, startTime time.Time) {
	select {
	case <-ctx.Done():
		duration := time.Since(startTime)
		if ctx.Err() == context.DeadlineExceeded {
			logger.Error("操作超时监控",
				zap.String("operation", operation),
				zap.Duration("duration", duration),
				zap.Duration("timeout", tm.getTimeoutForOperation(operation)),
				zap.Error(ctx.Err()))
		}
	default:
		// 操作正常完成
	}
}

// getTimeoutForOperation 根据操作类型获取超时时间
func (tm *TimeoutManager) getTimeoutForOperation(operation string) time.Duration {
	switch operation {
	case "query", "select", "find":
		return tm.config.QueryTimeout
	case "transaction", "tx":
		return tm.config.TransactionTimeout
	case "database", "db":
		return tm.config.DatabaseTimeout
	default:
		return tm.config.OperationTimeout
	}
}

// TimeoutConfig 超时配置
type TimeoutConfig struct {
	Operation   time.Duration
	Query       time.Duration
	Transaction time.Duration
	Database    time.Duration
}

// GetTimeoutConfig 获取超时配置
func (tm *TimeoutManager) GetTimeoutConfig() TimeoutConfig {
	return TimeoutConfig{
		Operation:   tm.config.OperationTimeout,
		Query:       tm.config.QueryTimeout,
		Transaction: tm.config.TransactionTimeout,
		Database:    tm.config.DatabaseTimeout,
	}
}

// SetTimeoutConfig 设置超时配置
func (tm *TimeoutManager) SetTimeoutConfig(cfg TimeoutConfig) {
	tm.config.OperationTimeout = cfg.Operation
	tm.config.QueryTimeout = cfg.Query
	tm.config.TransactionTimeout = cfg.Transaction
	tm.config.DatabaseTimeout = cfg.Database
}

// TimeoutWrapper 超时包装器
type TimeoutWrapper struct {
	manager *TimeoutManager
}

// NewTimeoutWrapper 创建超时包装器
func NewTimeoutWrapper(manager *TimeoutManager) *TimeoutWrapper {
	return &TimeoutWrapper{
		manager: manager,
	}
}

// WrapOperation 包装操作，添加超时控制
func (tw *TimeoutWrapper) WrapOperation(ctx context.Context, operation string, fn func(context.Context) error) error {
	timeoutCtx, cancel := tw.manager.WithOperationTimeout(ctx)
	defer cancel()

	startTime := time.Now()
	
	// 启动超时监控
	done := make(chan struct{})
	go func() {
		defer close(done)
		tw.manager.MonitorTimeout(timeoutCtx, operation, startTime)
	}()

	// 执行操作
	err := fn(timeoutCtx)
	
	// 检查是否超时
	if timeoutErr := tw.manager.CheckTimeout(timeoutCtx, operation); timeoutErr != nil {
		return timeoutErr
	}

	return err
}

// WrapQuery 包装查询操作
func (tw *TimeoutWrapper) WrapQuery(ctx context.Context, operation string, fn func(context.Context) error) error {
	timeoutCtx, cancel := tw.manager.WithQueryTimeout(ctx)
	defer cancel()

	startTime := time.Now()
	
	// 启动超时监控
	go tw.manager.MonitorTimeout(timeoutCtx, operation, startTime)

	// 执行查询
	err := fn(timeoutCtx)
	
	// 检查是否超时
	if timeoutErr := tw.manager.CheckTimeout(timeoutCtx, operation); timeoutErr != nil {
		return timeoutErr
	}

	return err
}

// WrapTransaction 包装事务操作
func (tw *TimeoutWrapper) WrapTransaction(ctx context.Context, operation string, fn func(context.Context) error) error {
	timeoutCtx, cancel := tw.manager.WithTransactionTimeout(ctx)
	defer cancel()

	startTime := time.Now()
	
	// 启动超时监控
	go tw.manager.MonitorTimeout(timeoutCtx, operation, startTime)

	// 执行事务
	err := fn(timeoutCtx)
	
	// 检查是否超时
	if timeoutErr := tw.manager.CheckTimeout(timeoutCtx, operation); timeoutErr != nil {
		return timeoutErr
	}

	return err
}

// TimeoutMetrics 超时指标
type TimeoutMetrics struct {
	TotalOperations   int64
	TimeoutOperations int64
	AverageLatency    time.Duration
	MaxLatency        time.Duration
}

// GetTimeoutMetrics 获取超时指标
func (tm *TimeoutManager) GetTimeoutMetrics() TimeoutMetrics {
	// 这里可以实现实际的指标收集逻辑
	return TimeoutMetrics{
		TotalOperations:   0,
		TimeoutOperations: 0,
		AverageLatency:    0,
		MaxLatency:        0,
	}
}

// ResetTimeoutMetrics 重置超时指标
func (tm *TimeoutManager) ResetTimeoutMetrics() {
	// 这里可以实现实际的指标重置逻辑
	logger.Info("超时指标已重置")
}

// IsTimeoutError 判断是否为超时错误
func IsTimeoutError(err error) bool {
	return err == context.DeadlineExceeded || err == ErrOperationTimeout
}

// HandleTimeoutError 处理超时错误
func HandleTimeoutError(ctx context.Context, operation string, err error) error {
	if IsTimeoutError(err) {
		logger.Error("操作超时",
			zap.String("operation", operation),
			zap.Error(err))
		return ErrOperationTimeout
	}
	return err
}

// TimeoutAwareContext 超时感知的上下文
type TimeoutAwareContext struct {
	context.Context
	operation string
	startTime time.Time
	manager   *TimeoutManager
}

// NewTimeoutAwareContext 创建超时感知的上下文
func NewTimeoutAwareContext(ctx context.Context, operation string, manager *TimeoutManager) *TimeoutAwareContext {
	return &TimeoutAwareContext{
		Context:   ctx,
		operation: operation,
		startTime: time.Now(),
		manager:   manager,
	}
}

// CheckDeadline 检查截止时间
func (tac *TimeoutAwareContext) CheckDeadline() error {
	return tac.manager.CheckTimeout(tac.Context, tac.operation)
}

// GetElapsed 获取已经过的时间
func (tac *TimeoutAwareContext) GetElapsed() time.Duration {
	return time.Since(tac.startTime)
}

// GetRemaining 获取剩余时间
func (tac *TimeoutAwareContext) GetRemaining() time.Duration {
	if deadline, ok := tac.Deadline(); ok {
		return time.Until(deadline)
	}
	return 0
}

// IsNearDeadline 判断是否接近截止时间
func (tac *TimeoutAwareContext) IsNearDeadline(threshold time.Duration) bool {
	remaining := tac.GetRemaining()
	return remaining > 0 && remaining <= threshold
}
