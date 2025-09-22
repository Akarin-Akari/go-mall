package errors

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"mall-go/pkg/logger"

	"go.uber.org/zap"
)

// RetryConfig 重试配置
type RetryConfig struct {
	MaxAttempts   int                    `json:"max_attempts"`   // 最大重试次数
	BaseDelay     time.Duration          `json:"base_delay"`     // 基础延迟时间
	MaxDelay      time.Duration          `json:"max_delay"`      // 最大延迟时间
	BackoffFactor float64                `json:"backoff_factor"` // 退避因子
	Jitter        bool                   `json:"jitter"`         // 是否启用抖动
	RetryableErrors []ErrorCode          `json:"retryable_errors"` // 可重试的错误码
	OnRetry       func(attempt int, err error) // 重试回调
}

// CircuitBreakerConfig 熔断器配置
type CircuitBreakerConfig struct {
	FailureThreshold   int           `json:"failure_threshold"`   // 失败阈值
	RecoveryTimeout    time.Duration `json:"recovery_timeout"`    // 恢复超时时间
	HalfOpenMaxCalls   int           `json:"half_open_max_calls"` // 半开状态最大调用数
	MinRequestsToTrip  int           `json:"min_requests_to_trip"` // 最小触发请求数
}

// CircuitBreakerState 熔断器状态
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	config          CircuitBreakerConfig
	state           CircuitBreakerState
	failureCount    int
	requestCount    int
	lastFailureTime time.Time
	halfOpenCalls   int
}

// RecoveryManager 错误恢复管理器
type RecoveryManager struct {
	retryConfigs      map[string]RetryConfig
	circuitBreakers   map[string]*CircuitBreaker
	defaultRetryConfig RetryConfig
}

// NewRecoveryManager 创建错误恢复管理器
func NewRecoveryManager() *RecoveryManager {
	return &RecoveryManager{
		retryConfigs:    make(map[string]RetryConfig),
		circuitBreakers: make(map[string]*CircuitBreaker),
		defaultRetryConfig: RetryConfig{
			MaxAttempts:   3,
			BaseDelay:     100 * time.Millisecond,
			MaxDelay:      5 * time.Second,
			BackoffFactor: 2.0,
			Jitter:        true,
			RetryableErrors: []ErrorCode{
				ErrCodeSystemTimeout,
				ErrCodeNetworkTimeout,
				ErrCodeThirdPartyUnavailable,
				ErrCodeDatabaseConnection,
				ErrCodeCacheConnectionFailed,
			},
		},
	}
}

// SetRetryConfig 设置重试配置
func (rm *RecoveryManager) SetRetryConfig(operation string, config RetryConfig) {
	rm.retryConfigs[operation] = config
}

// SetCircuitBreaker 设置熔断器
func (rm *RecoveryManager) SetCircuitBreaker(operation string, config CircuitBreakerConfig) {
	rm.circuitBreakers[operation] = &CircuitBreaker{
		config: config,
		state:  StateClosed,
	}
}

// ExecuteWithRetry 执行带重试的操作
func (rm *RecoveryManager) ExecuteWithRetry(ctx context.Context, operation string, fn func() error) error {
	config := rm.getRetryConfig(operation)
	
	var lastErr error
	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		// 检查上下文是否已取消
		if ctx.Err() != nil {
			return ctx.Err()
		}
		
		// 执行操作
		err := fn()
		if err == nil {
			// 成功执行
			if attempt > 1 {
				logger.Info("重试成功",
					zap.String("operation", operation),
					zap.Int("attempt", attempt))
			}
			return nil
		}
		
		lastErr = err
		
		// 检查是否为可重试错误
		if !rm.isRetryableError(err, config) {
			logger.Info("错误不可重试",
				zap.String("operation", operation),
				zap.Error(err))
			return err
		}
		
		// 最后一次尝试，不再延迟
		if attempt == config.MaxAttempts {
			break
		}
		
		// 计算延迟时间
		delay := rm.calculateDelay(attempt, config)
		
		// 执行重试回调
		if config.OnRetry != nil {
			config.OnRetry(attempt, err)
		}
		
		logger.Warn("操作失败，准备重试",
			zap.String("operation", operation),
			zap.Int("attempt", attempt),
			zap.Duration("delay", delay),
			zap.Error(err))
		
		// 等待延迟时间
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// 继续下一次重试
		}
	}
	
	logger.Error("重试失败，已达到最大重试次数",
		zap.String("operation", operation),
		zap.Int("max_attempts", config.MaxAttempts),
		zap.Error(lastErr))
	
	return lastErr
}

// ExecuteWithCircuitBreaker 执行带熔断器的操作
func (rm *RecoveryManager) ExecuteWithCircuitBreaker(operation string, fn func() error) error {
	breaker := rm.getCircuitBreaker(operation)
	if breaker == nil {
		// 没有熔断器配置，直接执行
		return fn()
	}
	
	// 检查熔断器状态
	if !breaker.allowRequest() {
		return NewSystemError(ErrCodeSystemOverload, "服务熔断中，请稍后再试").
			WithSuggestion("系统正在恢复中，请稍后重试")
	}
	
	// 执行操作
	err := fn()
	
	// 记录执行结果
	breaker.recordResult(err == nil)
	
	return err
}

// ExecuteWithRecovery 执行带完整恢复机制的操作
func (rm *RecoveryManager) ExecuteWithRecovery(ctx context.Context, operation string, fn func() error) error {
	// 先检查熔断器
	return rm.ExecuteWithCircuitBreaker(operation, func() error {
		// 再执行重试机制
		return rm.ExecuteWithRetry(ctx, operation, fn)
	})
}

// getRetryConfig 获取重试配置
func (rm *RecoveryManager) getRetryConfig(operation string) RetryConfig {
	if config, exists := rm.retryConfigs[operation]; exists {
		return config
	}
	return rm.defaultRetryConfig
}

// getCircuitBreaker 获取熔断器
func (rm *RecoveryManager) getCircuitBreaker(operation string) *CircuitBreaker {
	return rm.circuitBreakers[operation]
}

// isRetryableError 检查错误是否可重试
func (rm *RecoveryManager) isRetryableError(err error, config RetryConfig) bool {
	if be, ok := err.(*BusinessError); ok {
		// 检查错误是否标记为可重试
		if be.Retryable {
			return true
		}
		
		// 检查错误码是否在可重试列表中
		for _, code := range config.RetryableErrors {
			if be.Code == code {
				return true
			}
		}
		
		// 根据错误分类判断
		switch be.Category {
		case CategoryNetwork, CategoryThirdParty:
			return true
		case CategorySystem:
			// 系统错误中的特定错误可重试
			switch be.Code {
			case ErrCodeSystemTimeout, ErrCodeSystemOverload:
				return true
			}
		}
	}
	
	return false
}

// calculateDelay 计算延迟时间
func (rm *RecoveryManager) calculateDelay(attempt int, config RetryConfig) time.Duration {
	// 指数退避算法
	delay := time.Duration(float64(config.BaseDelay) * math.Pow(config.BackoffFactor, float64(attempt-1)))
	
	// 限制最大延迟
	if delay > config.MaxDelay {
		delay = config.MaxDelay
	}
	
	// 添加抖动
	if config.Jitter {
		jitter := time.Duration(rand.Float64() * float64(delay) * 0.1)
		delay += jitter
	}
	
	return delay
}

// CircuitBreaker 方法

// allowRequest 检查是否允许请求
func (cb *CircuitBreaker) allowRequest() bool {
	now := time.Now()
	
	switch cb.state {
	case StateClosed:
		return true
		
	case StateOpen:
		// 检查是否到达恢复时间
		if now.Sub(cb.lastFailureTime) > cb.config.RecoveryTimeout {
			cb.state = StateHalfOpen
			cb.halfOpenCalls = 0
			return true
		}
		return false
		
	case StateHalfOpen:
		// 半开状态下限制调用数量
		return cb.halfOpenCalls < cb.config.HalfOpenMaxCalls
		
	default:
		return false
	}
}

// recordResult 记录执行结果
func (cb *CircuitBreaker) recordResult(success bool) {
	cb.requestCount++
	
	if success {
		cb.onSuccess()
	} else {
		cb.onFailure()
	}
}

// onSuccess 处理成功结果
func (cb *CircuitBreaker) onSuccess() {
	if cb.state == StateHalfOpen {
		cb.halfOpenCalls++
		// 如果半开状态下连续成功，关闭熔断器
		if cb.halfOpenCalls >= cb.config.HalfOpenMaxCalls {
			cb.state = StateClosed
			cb.failureCount = 0
			logger.Info("熔断器已关闭")
		}
	} else if cb.state == StateClosed {
		// 闭合状态下成功，重置失败计数
		cb.failureCount = 0
	}
}

// onFailure 处理失败结果
func (cb *CircuitBreaker) onFailure() {
	cb.failureCount++
	cb.lastFailureTime = time.Now()
	
	if cb.state == StateHalfOpen {
		// 半开状态下失败，立即打开熔断器
		cb.state = StateOpen
		logger.Warn("熔断器重新打开")
	} else if cb.state == StateClosed {
		// 闭合状态下检查是否需要打开熔断器
		if cb.requestCount >= cb.config.MinRequestsToTrip &&
			cb.failureCount >= cb.config.FailureThreshold {
			cb.state = StateOpen
			logger.Warn("熔断器已打开",
				zap.Int("failure_count", cb.failureCount),
				zap.Int("request_count", cb.requestCount))
		}
	}
}

// GetState 获取熔断器状态
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	return cb.state
}

// GetStatistics 获取熔断器统计信息
func (cb *CircuitBreaker) GetStatistics() map[string]interface{} {
	return map[string]interface{}{
		"state":          cb.state,
		"failure_count":  cb.failureCount,
		"request_count":  cb.requestCount,
		"half_open_calls": cb.halfOpenCalls,
		"last_failure_time": cb.lastFailureTime,
	}
}

// 预定义的操作配置
var (
	// 数据库操作重试配置
	DatabaseRetryConfig = RetryConfig{
		MaxAttempts:   3,
		BaseDelay:     50 * time.Millisecond,
		MaxDelay:      1 * time.Second,
		BackoffFactor: 2.0,
		Jitter:        true,
		RetryableErrors: []ErrorCode{
			ErrCodeDatabaseConnection,
			ErrCodeDatabaseQueryTimeout,
		},
	}
	
	// 第三方服务重试配置
	ThirdPartyRetryConfig = RetryConfig{
		MaxAttempts:   3,
		BaseDelay:     200 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 2.0,
		Jitter:        true,
		RetryableErrors: []ErrorCode{
			ErrCodeThirdPartyUnavailable,
			ErrCodeThirdPartyTimeout,
			ErrCodeNetworkTimeout,
		},
	}
	
	// 支付服务重试配置
	PaymentRetryConfig = RetryConfig{
		MaxAttempts:   2,
		BaseDelay:     500 * time.Millisecond,
		MaxDelay:      3 * time.Second,
		BackoffFactor: 1.5,
		Jitter:        false,
		RetryableErrors: []ErrorCode{
			ErrCodePaymentChannelUnavailable,
			ErrCodeThirdPartyTimeout,
		},
	}
	
	// 数据库熔断器配置
	DatabaseCircuitBreakerConfig = CircuitBreakerConfig{
		FailureThreshold:  10,
		RecoveryTimeout:   30 * time.Second,
		HalfOpenMaxCalls:  3,
		MinRequestsToTrip: 20,
	}
	
	// 第三方服务熔断器配置
	ThirdPartyCircuitBreakerConfig = CircuitBreakerConfig{
		FailureThreshold:  5,
		RecoveryTimeout:   60 * time.Second,
		HalfOpenMaxCalls:  2,
		MinRequestsToTrip: 10,
	}
)

// 全局恢复管理器实例
var GlobalRecoveryManager = NewRecoveryManager()

// init 初始化默认配置
func init() {
	// 设置预定义的重试配置
	GlobalRecoveryManager.SetRetryConfig("database", DatabaseRetryConfig)
	GlobalRecoveryManager.SetRetryConfig("third_party", ThirdPartyRetryConfig)
	GlobalRecoveryManager.SetRetryConfig("payment", PaymentRetryConfig)
	
	// 设置预定义的熔断器配置
	GlobalRecoveryManager.SetCircuitBreaker("database", DatabaseCircuitBreakerConfig)
	GlobalRecoveryManager.SetCircuitBreaker("third_party", ThirdPartyCircuitBreakerConfig)
}

// ExecuteWithRetry 全局重试执行函数
func ExecuteWithRetry(ctx context.Context, operation string, fn func() error) error {
	return GlobalRecoveryManager.ExecuteWithRetry(ctx, operation, fn)
}

// ExecuteWithCircuitBreaker 全局熔断器执行函数
func ExecuteWithCircuitBreaker(operation string, fn func() error) error {
	return GlobalRecoveryManager.ExecuteWithCircuitBreaker(operation, fn)
}

// ExecuteWithRecovery 全局恢复机制执行函数
func ExecuteWithRecovery(ctx context.Context, operation string, fn func() error) error {
	return GlobalRecoveryManager.ExecuteWithRecovery(ctx, operation, fn)
}