package security

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"mall-go/pkg/logger"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RateLimitConfig 速率限制配置
type RateLimitConfig struct {
	RequestsPerMinute int           // 每分钟请求数
	BurstSize         int           // 突发大小
	WindowDuration    time.Duration // 时间窗口
}

// RateLimiter 速率限制器
type RateLimiter struct {
	rdb    *redis.Client
	config RateLimitConfig
}

// NewRateLimiter 创建速率限制器
func NewRateLimiter(rdb *redis.Client, config RateLimitConfig) *RateLimiter {
	if config.WindowDuration == 0 {
		config.WindowDuration = time.Minute
	}
	if config.RequestsPerMinute == 0 {
		config.RequestsPerMinute = 100
	}
	if config.BurstSize == 0 {
		config.BurstSize = 10
	}

	return &RateLimiter{
		rdb:    rdb,
		config: config,
	}
}

// CheckLimit 检查是否超过限制
func (rl *RateLimiter) CheckLimit(ctx context.Context, key string) (bool, int, error) {
	if rl.rdb == nil {
		return true, rl.config.RequestsPerMinute, nil
	}

	redisKey := fmt.Sprintf("ratelimit:%s", key)

	// 使用Redis的INCR和EXPIRE实现速率限制
	pipe := rl.rdb.Pipeline()

	// 增加计数
	pipe.Incr(ctx, redisKey)
	// 设置过期时间
	pipe.Expire(ctx, redisKey, rl.config.WindowDuration)

	_, err := pipe.Exec(ctx)
	if err != nil {
		logger.Error("速率限制检查失败", zap.Error(err), zap.String("key", key))
		return true, rl.config.RequestsPerMinute, err
	}

	// 获取当前计数
	countStr, err := rl.rdb.Get(ctx, redisKey).Result()
	if err != nil && err != redis.Nil {
		logger.Error("获取速率限制计数失败", zap.Error(err))
		return true, rl.config.RequestsPerMinute, err
	}

	count := 0
	if countStr != "" {
		count, _ = strconv.Atoi(countStr)
	}

	// 检查是否超过限制
	allowed := count <= rl.config.RequestsPerMinute
	remaining := rl.config.RequestsPerMinute - count
	if remaining < 0 {
		remaining = 0
	}

	return allowed, remaining, nil
}

// GetRemaining 获取剩余请求数
func (rl *RateLimiter) GetRemaining(ctx context.Context, key string) (int, error) {
	if rl.rdb == nil {
		return rl.config.RequestsPerMinute, nil
	}

	redisKey := fmt.Sprintf("ratelimit:%s", key)
	countStr, err := rl.rdb.Get(ctx, redisKey).Result()
	if err != nil && err != redis.Nil {
		logger.Error("获取剩余请求数失败", zap.Error(err))
		return 0, err
	}

	count := 0
	if countStr != "" {
		count, _ = strconv.Atoi(countStr)
	}

	remaining := rl.config.RequestsPerMinute - count
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}

// Reset 重置限制计数
func (rl *RateLimiter) Reset(ctx context.Context, key string) error {
	if rl.rdb == nil {
		return nil
	}

	redisKey := fmt.Sprintf("ratelimit:%s", key)
	err := rl.rdb.Del(ctx, redisKey).Err()
	if err != nil {
		logger.Error("重置速率限制失败", zap.Error(err))
		return err
	}

	return nil
}

// 全局速率限制器实例
var GlobalRateLimiter *RateLimiter

// InitRateLimiter 初始化全局速率限制器
func InitRateLimiter(rdb *redis.Client, config RateLimitConfig) {
	GlobalRateLimiter = NewRateLimiter(rdb, config)
}

// GetRateLimiter 获取全局速率限制器
func GetRateLimiter() *RateLimiter {
	if GlobalRateLimiter == nil {
		GlobalRateLimiter = NewRateLimiter(nil, RateLimitConfig{
			RequestsPerMinute: 100,
			BurstSize:         10,
			WindowDuration:    time.Minute,
		})
	}
	return GlobalRateLimiter
}
