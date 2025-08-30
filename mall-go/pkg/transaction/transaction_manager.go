package transaction

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"mall-go/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TransactionManager 事务管理器
type TransactionManager struct {
	db *gorm.DB
}

// NewTransactionManager 创建事务管理器
func NewTransactionManager(db *gorm.DB) *TransactionManager {
	return &TransactionManager{
		db: db,
	}
}

// TransactionFunc 事务执行函数类型
type TransactionFunc func(tx *gorm.DB) error

// CompensationFunc 补偿函数类型
type CompensationFunc func() error

// TransactionOptions 事务选项
type TransactionOptions struct {
	Timeout        time.Duration     // 事务超时时间
	RetryCount     int               // 重试次数
	RetryInterval  time.Duration     // 重试间隔
	Compensations  []CompensationFunc // 补偿函数列表
	EnableLogging  bool              // 是否启用详细日志
	IsolationLevel string            // 事务隔离级别
}

// DefaultTransactionOptions 默认事务选项
func DefaultTransactionOptions() *TransactionOptions {
	return &TransactionOptions{
		Timeout:       30 * time.Second,
		RetryCount:    3,
		RetryInterval: 100 * time.Millisecond,
		Compensations: make([]CompensationFunc, 0),
		EnableLogging: true,
		IsolationLevel: "READ_COMMITTED",
	}
}

// TransactionResult 事务执行结果
type TransactionResult struct {
	Success       bool          `json:"success"`
	Error         error         `json:"error,omitempty"`
	ExecutionTime time.Duration `json:"execution_time"`
	RetryCount    int           `json:"retry_count"`
	CompensationExecuted bool   `json:"compensation_executed"`
}

// ExecuteTransaction 执行事务
func (tm *TransactionManager) ExecuteTransaction(fn TransactionFunc, opts *TransactionOptions) *TransactionResult {
	if opts == nil {
		opts = DefaultTransactionOptions()
	}

	result := &TransactionResult{
		Success: false,
	}

	startTime := time.Now()
	defer func() {
		result.ExecutionTime = time.Since(startTime)
	}()

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()

	// 执行事务（带重试）
	for attempt := 0; attempt <= opts.RetryCount; attempt++ {
		result.RetryCount = attempt

		if opts.EnableLogging {
			logger.Info("开始执行事务",
				zap.Int("attempt", attempt+1),
				zap.Int("max_attempts", opts.RetryCount+1))
		}

		// 执行事务
		err := tm.executeTransactionWithContext(ctx, fn, opts)
		if err == nil {
			result.Success = true
			if opts.EnableLogging {
				logger.Info("事务执行成功",
					zap.Int("attempt", attempt+1),
					zap.Duration("execution_time", result.ExecutionTime))
			}
			return result
		}

		result.Error = err

		// 记录错误
		if opts.EnableLogging {
			logger.Error("事务执行失败",
				zap.Int("attempt", attempt+1),
				zap.Error(err),
				zap.String("stack_trace", tm.getStackTrace()))
		}

		// 如果不是最后一次尝试，等待后重试
		if attempt < opts.RetryCount {
			select {
			case <-ctx.Done():
				result.Error = fmt.Errorf("事务超时: %v", ctx.Err())
				tm.executeCompensations(opts.Compensations)
				result.CompensationExecuted = true
				return result
			case <-time.After(opts.RetryInterval):
				// 继续重试
			}
		}
	}

	// 所有重试都失败，执行补偿
	if len(opts.Compensations) > 0 {
		if opts.EnableLogging {
			logger.Warn("执行补偿操作", zap.Int("compensation_count", len(opts.Compensations)))
		}
		tm.executeCompensations(opts.Compensations)
		result.CompensationExecuted = true
	}

	return result
}

// executeTransactionWithContext 在上下文中执行事务
func (tm *TransactionManager) executeTransactionWithContext(ctx context.Context, fn TransactionFunc, opts *TransactionOptions) error {
	// 开始事务
	tx := tm.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("开始事务失败: %v", tx.Error)
	}

	// 设置事务隔离级别
	if opts.IsolationLevel != "" {
		if err := tx.Exec(fmt.Sprintf("SET TRANSACTION ISOLATION LEVEL %s", opts.IsolationLevel)).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("设置事务隔离级别失败: %v", err)
		}
	}

	// 确保事务会被回滚或提交
	committed := false
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			if opts.EnableLogging {
				logger.Error("事务执行发生panic，已回滚",
					zap.Any("panic", r),
					zap.String("stack_trace", tm.getStackTrace()))
			}
		} else if !committed {
			tx.Rollback()
			if opts.EnableLogging {
				logger.Info("事务已回滚")
			}
		}
	}()

	// 在goroutine中执行事务函数，以便支持超时
	errChan := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errChan <- fmt.Errorf("事务执行panic: %v", r)
			}
		}()
		errChan <- fn(tx)
	}()

	// 等待执行完成或超时
	select {
	case <-ctx.Done():
		return fmt.Errorf("事务执行超时: %v", ctx.Err())
	case err := <-errChan:
		if err != nil {
			return err
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	committed = true
	return nil
}

// executeCompensations 执行补偿操作
func (tm *TransactionManager) executeCompensations(compensations []CompensationFunc) {
	for i, compensation := range compensations {
		if err := compensation(); err != nil {
			logger.Error("补偿操作执行失败",
				zap.Int("compensation_index", i),
				zap.Error(err))
		} else {
			logger.Info("补偿操作执行成功", zap.Int("compensation_index", i))
		}
	}
}

// getStackTrace 获取堆栈跟踪
func (tm *TransactionManager) getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// ExecuteTransactionWithSavepoint 使用保存点执行事务
func (tm *TransactionManager) ExecuteTransactionWithSavepoint(fn TransactionFunc, savepointName string, opts *TransactionOptions) *TransactionResult {
	if opts == nil {
		opts = DefaultTransactionOptions()
	}

	result := &TransactionResult{
		Success: false,
	}

	startTime := time.Now()
	defer func() {
		result.ExecutionTime = time.Since(startTime)
	}()

	// 开始事务
	tx := tm.db.Begin()
	if tx.Error != nil {
		result.Error = fmt.Errorf("开始事务失败: %v", tx.Error)
		return result
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			result.Error = fmt.Errorf("事务执行panic: %v", r)
			if opts.EnableLogging {
				logger.Error("事务执行发生panic，已回滚",
					zap.Any("panic", r),
					zap.String("stack_trace", tm.getStackTrace()))
			}
		}
	}()

	// 创建保存点
	if err := tx.Exec(fmt.Sprintf("SAVEPOINT %s", savepointName)).Error; err != nil {
		tx.Rollback()
		result.Error = fmt.Errorf("创建保存点失败: %v", err)
		return result
	}

	// 执行事务函数
	if err := fn(tx); err != nil {
		// 回滚到保存点
		if rollbackErr := tx.Exec(fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", savepointName)).Error; rollbackErr != nil {
			tx.Rollback()
			result.Error = fmt.Errorf("回滚到保存点失败: %v, 原始错误: %v", rollbackErr, err)
			return result
		}

		result.Error = err
		if opts.EnableLogging {
			logger.Error("事务执行失败，已回滚到保存点",
				zap.String("savepoint", savepointName),
				zap.Error(err))
		}
		return result
	}

	// 释放保存点
	if err := tx.Exec(fmt.Sprintf("RELEASE SAVEPOINT %s", savepointName)).Error; err != nil {
		tx.Rollback()
		result.Error = fmt.Errorf("释放保存点失败: %v", err)
		return result
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		result.Error = fmt.Errorf("提交事务失败: %v", err)
		return result
	}

	result.Success = true
	if opts.EnableLogging {
		logger.Info("带保存点的事务执行成功",
			zap.String("savepoint", savepointName),
			zap.Duration("execution_time", result.ExecutionTime))
	}

	return result
}

// BatchExecuteTransactions 批量执行事务
func (tm *TransactionManager) BatchExecuteTransactions(transactions []TransactionFunc, opts *TransactionOptions) []*TransactionResult {
	results := make([]*TransactionResult, len(transactions))

	for i, fn := range transactions {
		results[i] = tm.ExecuteTransaction(fn, opts)
		
		// 如果某个事务失败且没有启用补偿，停止执行后续事务
		if !results[i].Success && len(opts.Compensations) == 0 {
			// 为剩余的事务创建失败结果
			for j := i + 1; j < len(transactions); j++ {
				results[j] = &TransactionResult{
					Success: false,
					Error:   fmt.Errorf("由于前序事务失败而跳过执行"),
				}
			}
			break
		}
	}

	return results
}

// CreateCompensation 创建补偿函数
func CreateCompensation(description string, fn func() error) CompensationFunc {
	return func() error {
		logger.Info("执行补偿操作", zap.String("description", description))
		return fn()
	}
}

// 全局事务管理器实例
var globalTransactionManager *TransactionManager

// InitGlobalTransactionManager 初始化全局事务管理器
func InitGlobalTransactionManager(db *gorm.DB) {
	globalTransactionManager = NewTransactionManager(db)
}

// GetGlobalTransactionManager 获取全局事务管理器
func GetGlobalTransactionManager() *TransactionManager {
	return globalTransactionManager
}
