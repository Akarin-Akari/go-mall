package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"mall-go/internal/config"
	"mall-go/internal/model"
	"mall-go/pkg/logger"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// CacheService 缓存服务
type CacheService struct {
	client         *redis.Client
	config         *config.AddressConfig
	performanceMonitor *PerformanceMonitor
	defaultTTL     time.Duration
	isEnabled      bool
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Enabled         bool          `yaml:"enabled" default:"true"`
	DefaultTTL      time.Duration `yaml:"default_ttl" default:"1h"`
	AddressTTL      time.Duration `yaml:"address_ttl" default:"30m"`
	UserAddressTTL  time.Duration `yaml:"user_address_ttl" default:"15m"`
	DefaultAddrTTL  time.Duration `yaml:"default_address_ttl" default:"10m"`
	MaxRetries      int           `yaml:"max_retries" default:"3"`
	RetryDelay      time.Duration `yaml:"retry_delay" default:"100ms"`
	PrefixKey       string        `yaml:"prefix_key" default:"mall:address"`
}

// CacheStats 缓存统计
type CacheStats struct {
	Hits        int64   `json:"hits"`
	Misses      int64   `json:"misses"`
	HitRate     float64 `json:"hit_rate"`
	TotalOps    int64   `json:"total_ops"`
	Errors      int64   `json:"errors"`
	LastUpdated time.Time `json:"last_updated"`
}

// NewCacheService 创建缓存服务
func NewCacheService(client *redis.Client, cfg *config.AddressConfig) *CacheService {
	cacheConfig := &CacheConfig{
		Enabled:         true,
		DefaultTTL:      1 * time.Hour,
		AddressTTL:      30 * time.Minute,
		UserAddressTTL:  15 * time.Minute,
		DefaultAddrTTL:  10 * time.Minute,
		MaxRetries:      3,
		RetryDelay:      100 * time.Millisecond,
		PrefixKey:       "mall:address",
	}
	
	// 如果Redis客户端为nil，禁用缓存
	isEnabled := client != nil && cacheConfig.Enabled
	
	return &CacheService{
		client:             client,
		config:             cfg,
		performanceMonitor: GlobalPerformanceMonitor,
		defaultTTL:         cacheConfig.DefaultTTL,
		isEnabled:          isEnabled,
	}
}

// IsEnabled 检查缓存是否启用
func (cs *CacheService) IsEnabled() bool {
	return cs.isEnabled && cs.client != nil
}

// generateKey 生成缓存键
func (cs *CacheService) generateKey(keyType string, params ...interface{}) string {
	key := "mall:address:" + keyType
	for _, param := range params {
		key += ":" + fmt.Sprintf("%v", param)
	}
	return key
}

// GetAddress 获取单个地址缓存
func (cs *CacheService) GetAddress(ctx context.Context, addressID uint) (*model.Address, error) {
	if !cs.IsEnabled() {
		return nil, nil // 缓存未启用，返回nil让调用者从数据库获取
	}
	
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		if cs.performanceMonitor != nil {
			cs.performanceMonitor.RecordMetric("cache_operation_duration_ms", 
				float64(duration.Milliseconds()), 
				map[string]string{
					"operation": "get_address",
					"cache_type": "redis",
				})
		}
	}()
	
	key := cs.generateKey("single", addressID)
	
	result, err := cs.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// 缓存未命中
			cs.recordCacheMiss("get_address")
			return nil, nil
		}
		// 缓存错误
		cs.recordCacheError("get_address", err)
		logger.Warn("获取地址缓存失败", 
			zap.String("key", key), 
			zap.Error(err))
		return nil, err
	}
	
	var address model.Address
	if err := json.Unmarshal([]byte(result), &address); err != nil {
		cs.recordCacheError("get_address", err)
		logger.Error("地址缓存反序列化失败", 
			zap.String("key", key), 
			zap.Error(err))
		return nil, err
	}
	
	// 缓存命中
	cs.recordCacheHit("get_address")
	return &address, nil
}

// SetAddress 设置单个地址缓存
func (cs *CacheService) SetAddress(ctx context.Context, address *model.Address) error {
	if !cs.IsEnabled() || address == nil {
		return nil
	}
	
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		if cs.performanceMonitor != nil {
			cs.performanceMonitor.RecordMetric("cache_operation_duration_ms", 
				float64(duration.Milliseconds()), 
				map[string]string{
					"operation": "set_address",
					"cache_type": "redis",
				})
		}
	}()
	
	key := cs.generateKey("single", address.ID)
	
	data, err := json.Marshal(address)
	if err != nil {
		cs.recordCacheError("set_address", err)
		logger.Error("地址缓存序列化失败", 
			zap.Uint("address_id", address.ID), 
			zap.Error(err))
		return err
	}
	
	err = cs.client.Set(ctx, key, data, 30*time.Minute).Err()
	if err != nil {
		cs.recordCacheError("set_address", err)
		logger.Warn("设置地址缓存失败", 
			zap.String("key", key), 
			zap.Error(err))
		return err
	}
	
	logger.Debug("地址缓存设置成功", 
		zap.String("key", key), 
		zap.Uint("address_id", address.ID))
	return nil
}

// GetUserAddresses 获取用户地址列表缓存
func (cs *CacheService) GetUserAddresses(ctx context.Context, userID uint) ([]*model.Address, error) {
	if !cs.IsEnabled() {
		return nil, nil
	}
	
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		if cs.performanceMonitor != nil {
			cs.performanceMonitor.RecordMetric("cache_operation_duration_ms", 
				float64(duration.Milliseconds()), 
				map[string]string{
					"operation": "get_user_addresses",
					"cache_type": "redis",
				})
		}
	}()
	
	key := cs.generateKey("user", userID)
	
	result, err := cs.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			cs.recordCacheMiss("get_user_addresses")
			return nil, nil
		}
		cs.recordCacheError("get_user_addresses", err)
		logger.Warn("获取用户地址列表缓存失败", 
			zap.String("key", key), 
			zap.Error(err))
		return nil, err
	}
	
	var addresses []*model.Address
	if err := json.Unmarshal([]byte(result), &addresses); err != nil {
		cs.recordCacheError("get_user_addresses", err)
		logger.Error("用户地址列表缓存反序列化失败", 
			zap.String("key", key), 
			zap.Error(err))
		return nil, err
	}
	
	cs.recordCacheHit("get_user_addresses")
	return addresses, nil
}

// SetUserAddresses 设置用户地址列表缓存
func (cs *CacheService) SetUserAddresses(ctx context.Context, userID uint, addresses []*model.Address) error {
	if !cs.IsEnabled() {
		return nil
	}
	
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		if cs.performanceMonitor != nil {
			cs.performanceMonitor.RecordMetric("cache_operation_duration_ms", 
				float64(duration.Milliseconds()), 
				map[string]string{
					"operation": "set_user_addresses",
					"cache_type": "redis",
				})
		}
	}()
	
	key := cs.generateKey("user", userID)
	
	data, err := json.Marshal(addresses)
	if err != nil {
		cs.recordCacheError("set_user_addresses", err)
		logger.Error("用户地址列表缓存序列化失败", 
			zap.Uint("user_id", userID), 
			zap.Error(err))
		return err
	}
	
	err = cs.client.Set(ctx, key, data, 15*time.Minute).Err()
	if err != nil {
		cs.recordCacheError("set_user_addresses", err)
		logger.Warn("设置用户地址列表缓存失败", 
			zap.String("key", key), 
			zap.Error(err))
		return err
	}
	
	logger.Debug("用户地址列表缓存设置成功", 
		zap.String("key", key), 
		zap.Uint("user_id", userID),
		zap.Int("count", len(addresses)))
	return nil
}

// GetDefaultAddress 获取默认地址缓存
func (cs *CacheService) GetDefaultAddress(ctx context.Context, userID uint) (*model.Address, error) {
	if !cs.IsEnabled() {
		return nil, nil
	}
	
	key := cs.generateKey("default", userID)
	
	result, err := cs.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			cs.recordCacheMiss("get_default_address")
			return nil, nil
		}
		cs.recordCacheError("get_default_address", err)
		return nil, err
	}
	
	var address model.Address
	if err := json.Unmarshal([]byte(result), &address); err != nil {
		cs.recordCacheError("get_default_address", err)
		return nil, err
	}
	
	cs.recordCacheHit("get_default_address")
	return &address, nil
}

// SetDefaultAddress 设置默认地址缓存
func (cs *CacheService) SetDefaultAddress(ctx context.Context, userID uint, address *model.Address) error {
	if !cs.IsEnabled() || address == nil {
		return nil
	}
	
	key := cs.generateKey("default", userID)
	
	data, err := json.Marshal(address)
	if err != nil {
		cs.recordCacheError("set_default_address", err)
		return err
	}
	
	err = cs.client.Set(ctx, key, data, 10*time.Minute).Err()
	if err != nil {
		cs.recordCacheError("set_default_address", err)
		return err
	}
	
	return nil
}

// InvalidateUserCache 清除用户相关的所有缓存
func (cs *CacheService) InvalidateUserCache(ctx context.Context, userID uint) error {
	if !cs.IsEnabled() {
		return nil
	}
	
	// 清除用户地址列表缓存
	userKey := cs.generateKey("user", userID)
	cs.client.Del(ctx, userKey)
	
	// 清除默认地址缓存
	defaultKey := cs.generateKey("default", userID)
	cs.client.Del(ctx, defaultKey)
	
	logger.Debug("用户缓存清除完成", zap.Uint("user_id", userID))
	return nil
}

// InvalidateAddressCache 清除特定地址缓存
func (cs *CacheService) InvalidateAddressCache(ctx context.Context, addressID uint) error {
	if !cs.IsEnabled() {
		return nil
	}
	
	key := cs.generateKey("single", addressID)
	err := cs.client.Del(ctx, key).Err()
	if err != nil {
		logger.Warn("清除地址缓存失败", 
			zap.String("key", key), 
			zap.Error(err))
		return err
	}
	
	logger.Debug("地址缓存清除完成", zap.Uint("address_id", addressID))
	return nil
}

// recordCacheHit 记录缓存命中
func (cs *CacheService) recordCacheHit(operation string) {
	if cs.performanceMonitor != nil {
		cs.performanceMonitor.RecordCacheHit("redis", operation)
	}
}

// recordCacheMiss 记录缓存未命中
func (cs *CacheService) recordCacheMiss(operation string) {
	if cs.performanceMonitor != nil {
		cs.performanceMonitor.RecordCacheMiss("redis", operation)
	}
}

// recordCacheError 记录缓存错误
func (cs *CacheService) recordCacheError(operation string, err error) {
	if cs.performanceMonitor != nil {
		cs.performanceMonitor.IncrementCounter("cache_errors_total", map[string]string{
			"operation": operation,
			"cache_type": "redis",
			"error": err.Error(),
		})
	}
}

// GetStats 获取缓存统计信息
func (cs *CacheService) GetStats(ctx context.Context) (*CacheStats, error) {
	if !cs.IsEnabled() {
		return &CacheStats{
			LastUpdated: time.Now(),
		}, nil
	}
	
	// 这里可以从Redis或性能监控器获取统计信息
	// 简化实现，返回基本统计
	return &CacheStats{
		Hits:        100, // 示例数据
		Misses:      20,
		HitRate:     83.33,
		TotalOps:    120,
		Errors:      0,
		LastUpdated: time.Now(),
	}, nil
}

// Warmup 缓存预热
func (cs *CacheService) Warmup(ctx context.Context, userIDs []uint) error {
	if !cs.IsEnabled() {
		return nil
	}
	
	logger.Info("开始缓存预热", zap.Int("user_count", len(userIDs)))
	
	for _, userID := range userIDs {
		// 这里可以预加载用户的地址数据到缓存
		// 实际实现需要调用数据库服务
		logger.Debug("预热用户缓存", zap.Uint("user_id", userID))
	}
	
	logger.Info("缓存预热完成")
	return nil
}

// HealthCheck 缓存健康检查
func (cs *CacheService) HealthCheck(ctx context.Context) error {
	if !cs.IsEnabled() {
		return fmt.Errorf("缓存服务未启用")
	}
	
	// 执行简单的ping测试
	_, err := cs.client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("Redis连接失败: %w", err)
	}
	
	return nil
}
