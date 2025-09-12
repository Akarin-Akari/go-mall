package cache

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"mall-go/pkg/logger"
	"mall-go/pkg/optimistic"
)

// ProtectionStrategy 防护策略枚举
type ProtectionStrategy string

const (
	// 缓存穿透防护
	ProtectionBloomFilter ProtectionStrategy = "bloom_filter" // 布隆过滤器
	ProtectionNullCache   ProtectionStrategy = "null_cache"   // 空值缓存
	ProtectionWhitelist   ProtectionStrategy = "whitelist"    // 白名单机制

	// 缓存击穿防护
	ProtectionDistributedLock ProtectionStrategy = "distributed_lock" // 分布式锁
	ProtectionMutexLock       ProtectionStrategy = "mutex_lock"       // 互斥锁
	ProtectionSingleFlight    ProtectionStrategy = "single_flight"    // 单飞模式

	// 缓存雪崩防护
	ProtectionRandomTTL      ProtectionStrategy = "random_ttl"      // 随机TTL
	ProtectionMultiLevel     ProtectionStrategy = "multi_level"     // 多级缓存
	ProtectionCircuitBreaker ProtectionStrategy = "circuit_breaker" // 熔断器
)

// ProtectionLevel 防护级别
type ProtectionLevel int

const (
	ProtectionLevelBasic    ProtectionLevel = 1 // 基础防护
	ProtectionLevelStandard ProtectionLevel = 2 // 标准防护
	ProtectionLevelAdvanced ProtectionLevel = 3 // 高级防护
)

// CacheProtectionConfig 缓存防护配置
type CacheProtectionConfig struct {
	// 基础配置
	Enabled    bool                 `json:"enabled"`    // 是否启用防护
	Level      ProtectionLevel      `json:"level"`      // 防护级别
	Strategies []ProtectionStrategy `json:"strategies"` // 启用的防护策略

	// 布隆过滤器配置
	BloomFilterConfig *BloomFilterConfig `json:"bloom_filter_config"`

	// 分布式锁配置
	DistributedLockConfig *DistributedLockConfig `json:"distributed_lock_config"`

	// 空值缓存配置
	NullCacheConfig *NullCacheConfig `json:"null_cache_config"`

	// 随机TTL配置
	RandomTTLConfig *RandomTTLConfig `json:"random_ttl_config"`

	// 熔断器配置
	CircuitBreakerConfig *CircuitBreakerConfig `json:"circuit_breaker_config"`

	// 监控配置
	MonitoringEnabled bool          `json:"monitoring_enabled"` // 是否启用监控
	MetricsInterval   time.Duration `json:"metrics_interval"`   // 指标收集间隔
	AlertThreshold    float64       `json:"alert_threshold"`    // 告警阈值
}

// BloomFilterConfig 布隆过滤器配置
type BloomFilterConfig struct {
	ExpectedElements  uint64        `json:"expected_elements"`   // 预期元素数量
	FalsePositiveRate float64       `json:"false_positive_rate"` // 误判率
	HashFunctions     uint64        `json:"hash_functions"`      // 哈希函数数量
	BitArraySize      uint64        `json:"bit_array_size"`      // 位数组大小
	RedisKey          string        `json:"redis_key"`           // Redis键名
	RefreshInterval   time.Duration `json:"refresh_interval"`    // 刷新间隔
}

// DistributedLockConfig 分布式锁配置
type DistributedLockConfig struct {
	LockTimeout    time.Duration `json:"lock_timeout"`    // 锁超时时间
	AcquireTimeout time.Duration `json:"acquire_timeout"` // 获取锁超时时间
	RetryInterval  time.Duration `json:"retry_interval"`  // 重试间隔
	MaxRetries     int           `json:"max_retries"`     // 最大重试次数
	AutoRenew      bool          `json:"auto_renew"`      // 是否自动续期
	RenewInterval  time.Duration `json:"renew_interval"`  // 续期间隔
}

// NullCacheConfig 空值缓存配置
type NullCacheConfig struct {
	TTL             time.Duration `json:"ttl"`              // 空值缓存TTL
	MaxNullKeys     int           `json:"max_null_keys"`    // 最大空值键数量
	CleanupInterval time.Duration `json:"cleanup_interval"` // 清理间隔
}

// RandomTTLConfig 随机TTL配置
type RandomTTLConfig struct {
	BaseTTL     time.Duration `json:"base_ttl"`     // 基础TTL
	RandomRange time.Duration `json:"random_range"` // 随机范围
	MinTTL      time.Duration `json:"min_ttl"`      // 最小TTL
	MaxTTL      time.Duration `json:"max_ttl"`      // 最大TTL
}

// CircuitBreakerConfig 熔断器配置
type CircuitBreakerConfig struct {
	FailureThreshold int           `json:"failure_threshold"`  // 失败阈值
	RecoveryTimeout  time.Duration `json:"recovery_timeout"`   // 恢复超时时间
	HalfOpenRequests int           `json:"half_open_requests"` // 半开状态请求数
	MonitoringPeriod time.Duration `json:"monitoring_period"`  // 监控周期
}

// ProtectionMetrics 防护指标
type ProtectionMetrics struct {
	// 总体指标
	TotalRequests     int64   `json:"total_requests"`     // 总请求数
	ProtectedRequests int64   `json:"protected_requests"` // 被防护的请求数
	ProtectionRate    float64 `json:"protection_rate"`    // 防护率

	// 穿透防护指标
	PenetrationAttempts int64 `json:"penetration_attempts"` // 穿透尝试次数
	PenetrationBlocked  int64 `json:"penetration_blocked"`  // 穿透被阻止次数
	BloomFilterHits     int64 `json:"bloom_filter_hits"`    // 布隆过滤器命中次数
	BloomFilterMisses   int64 `json:"bloom_filter_misses"`  // 布隆过滤器未命中次数

	// 击穿防护指标
	BreakdownAttempts int64 `json:"breakdown_attempts"` // 击穿尝试次数
	BreakdownBlocked  int64 `json:"breakdown_blocked"`  // 击穿被阻止次数
	LockAcquisitions  int64 `json:"lock_acquisitions"`  // 锁获取次数
	LockTimeouts      int64 `json:"lock_timeouts"`      // 锁超时次数

	// 雪崩防护指标
	AvalancheDetected   int64 `json:"avalanche_detected"`    // 雪崩检测次数
	AvalanchePrevented  int64 `json:"avalanche_prevented"`   // 雪崩预防次数
	CircuitBreakerTrips int64 `json:"circuit_breaker_trips"` // 熔断器触发次数

	// 性能指标
	AverageResponseTime time.Duration `json:"average_response_time"` // 平均响应时间
	MaxResponseTime     time.Duration `json:"max_response_time"`     // 最大响应时间
	MinResponseTime     time.Duration `json:"min_response_time"`     // 最小响应时间

	// 时间统计
	LastUpdated   time.Time `json:"last_updated"`    // 最后更新时间
	LastResetTime time.Time `json:"last_reset_time"` // 最后重置时间
}

// DistributedLock 分布式锁
type DistributedLock struct {
	rdb        *redis.Client
	key        string
	value      string
	expiration time.Duration
	ctx        context.Context
}

// BloomFilter 布隆过滤器
type BloomFilter struct {
	rdb           *redis.Client
	key           string
	hashFunctions uint64
	bitArraySize  uint64
	ctx           context.Context
}

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	failureCount    int64
	lastFailureTime time.Time
	state           CircuitBreakerState
	config          *CircuitBreakerConfig
	mutex           sync.RWMutex
}

// CircuitBreakerState 熔断器状态
type CircuitBreakerState int

const (
	CircuitBreakerClosed   CircuitBreakerState = 0 // 关闭状态
	CircuitBreakerOpen     CircuitBreakerState = 1 // 开启状态
	CircuitBreakerHalfOpen CircuitBreakerState = 2 // 半开状态
)

// CacheProtectionManager 缓存防护管理器
type CacheProtectionManager struct {
	config         *CacheProtectionConfig
	cacheManager   CacheManager
	keyManager     *CacheKeyManager
	consistencyMgr *CacheConsistencyManager
	warmupMgr      *CacheWarmupManager
	optimisticLock *optimistic.OptimisticLockService

	// 防护组件
	bloomFilter      *BloomFilter
	circuitBreaker   *CircuitBreaker
	distributedLocks map[string]*DistributedLock
	lockMutex        sync.RWMutex

	// 空值缓存
	nullCache      map[string]time.Time
	nullCacheMutex sync.RWMutex

	// 指标统计
	metrics      *ProtectionMetrics
	metricsMutex sync.RWMutex

	// 控制
	ctx          context.Context
	cancel       context.CancelFunc
	running      bool
	runningMutex sync.RWMutex
}

// DefaultCacheProtectionConfig 默认缓存防护配置
func DefaultCacheProtectionConfig() *CacheProtectionConfig {
	return &CacheProtectionConfig{
		Enabled: true,
		Level:   ProtectionLevelStandard,
		Strategies: []ProtectionStrategy{
			ProtectionBloomFilter,
			ProtectionDistributedLock,
			ProtectionRandomTTL,
		},
		BloomFilterConfig: &BloomFilterConfig{
			ExpectedElements:  1000000,
			FalsePositiveRate: 0.01,
			HashFunctions:     7,
			BitArraySize:      9585059,
			RedisKey:          "bloom:cache:protection",
			RefreshInterval:   1 * time.Hour,
		},
		DistributedLockConfig: &DistributedLockConfig{
			LockTimeout:    30 * time.Second,
			AcquireTimeout: 5 * time.Second,
			RetryInterval:  50 * time.Millisecond,
			MaxRetries:     10,
			AutoRenew:      true,
			RenewInterval:  10 * time.Second,
		},
		NullCacheConfig: &NullCacheConfig{
			TTL:             5 * time.Minute,
			MaxNullKeys:     10000,
			CleanupInterval: 10 * time.Minute,
		},
		RandomTTLConfig: &RandomTTLConfig{
			BaseTTL:     1 * time.Hour,
			RandomRange: 30 * time.Minute,
			MinTTL:      30 * time.Minute,
			MaxTTL:      2 * time.Hour,
		},
		CircuitBreakerConfig: &CircuitBreakerConfig{
			FailureThreshold: 50,
			RecoveryTimeout:  30 * time.Second,
			HalfOpenRequests: 5,
			MonitoringPeriod: 1 * time.Minute,
		},
		MonitoringEnabled: true,
		MetricsInterval:   30 * time.Second,
		AlertThreshold:    0.1, // 10%失败率告警
	}
}

// NewCacheProtectionManager 创建缓存防护管理器
func NewCacheProtectionManager(
	config *CacheProtectionConfig,
	cacheManager CacheManager,
	keyManager *CacheKeyManager,
	consistencyMgr *CacheConsistencyManager,
	warmupMgr *CacheWarmupManager,
	optimisticLock *optimistic.OptimisticLockService,
) *CacheProtectionManager {
	ctx, cancel := context.WithCancel(context.Background())

	cpm := &CacheProtectionManager{
		config:           config,
		cacheManager:     cacheManager,
		keyManager:       keyManager,
		consistencyMgr:   consistencyMgr,
		warmupMgr:        warmupMgr,
		optimisticLock:   optimisticLock,
		distributedLocks: make(map[string]*DistributedLock),
		nullCache:        make(map[string]time.Time),
		metrics:          &ProtectionMetrics{LastResetTime: time.Now()},
		ctx:              ctx,
		cancel:           cancel,
		running:          false,
	}

	// 初始化防护组件
	cpm.initializeComponents()

	return cpm
}

// initializeComponents 初始化防护组件
func (cpm *CacheProtectionManager) initializeComponents() {
	// 初始化布隆过滤器
	if cpm.isStrategyEnabled(ProtectionBloomFilter) {
		cpm.initBloomFilter()
	}

	// 初始化熔断器
	if cpm.isStrategyEnabled(ProtectionCircuitBreaker) {
		cpm.initCircuitBreaker()
	}
}

// initBloomFilter 初始化布隆过滤器
func (cpm *CacheProtectionManager) initBloomFilter() {
	if cpm.config.BloomFilterConfig == nil {
		return
	}

	// 获取Redis客户端
	var redisClient *redis.Client
	if redisCacheManager, ok := cpm.cacheManager.(*RedisCacheManager); ok {
		redisClient = redisCacheManager.client.GetClient()
	} else {
		logger.Error("缓存防护管理器需要Redis缓存管理器来支持布隆过滤器")
		return
	}

	cpm.bloomFilter = &BloomFilter{
		rdb:           redisClient,
		key:           cpm.config.BloomFilterConfig.RedisKey,
		hashFunctions: cpm.config.BloomFilterConfig.HashFunctions,
		bitArraySize:  cpm.config.BloomFilterConfig.BitArraySize,
		ctx:           cpm.ctx,
	}

	logger.Info("布隆过滤器初始化完成")
}

// initCircuitBreaker 初始化熔断器
func (cpm *CacheProtectionManager) initCircuitBreaker() {
	if cpm.config.CircuitBreakerConfig == nil {
		return
	}

	cpm.circuitBreaker = &CircuitBreaker{
		state:  CircuitBreakerClosed,
		config: cpm.config.CircuitBreakerConfig,
	}

	logger.Info("熔断器初始化完成")
}

// isStrategyEnabled 检查策略是否启用
func (cpm *CacheProtectionManager) isStrategyEnabled(strategy ProtectionStrategy) bool {
	for _, s := range cpm.config.Strategies {
		if s == strategy {
			return true
		}
	}
	return false
}

// Start 启动缓存防护管理器
func (cpm *CacheProtectionManager) Start() error {
	cpm.runningMutex.Lock()
	defer cpm.runningMutex.Unlock()

	if cpm.running {
		return fmt.Errorf("缓存防护管理器已在运行中")
	}

	if !cpm.config.Enabled {
		return fmt.Errorf("缓存防护管理器未启用")
	}

	// 启动监控
	if cpm.config.MonitoringEnabled {
		go cpm.startMetricsCollector()
	}

	// 启动空值缓存清理
	if cpm.isStrategyEnabled(ProtectionNullCache) {
		go cpm.startNullCacheCleanup()
	}

	// 启动布隆过滤器刷新
	if cpm.isStrategyEnabled(ProtectionBloomFilter) && cpm.bloomFilter != nil {
		go cpm.startBloomFilterRefresh()
	}

	cpm.running = true
	logger.Info("缓存防护管理器启动成功")

	return nil
}

// Stop 停止缓存防护管理器
func (cpm *CacheProtectionManager) Stop() error {
	cpm.runningMutex.Lock()
	defer cpm.runningMutex.Unlock()

	if !cpm.running {
		return nil
	}

	// 取消上下文
	cpm.cancel()

	// 释放所有分布式锁
	cpm.lockMutex.Lock()
	for _, lock := range cpm.distributedLocks {
		lock.Release()
	}
	cpm.distributedLocks = make(map[string]*DistributedLock)
	cpm.lockMutex.Unlock()

	cpm.running = false
	logger.Info("缓存防护管理器停止成功")

	return nil
}

// IsRunning 检查是否运行中
func (cpm *CacheProtectionManager) IsRunning() bool {
	cpm.runningMutex.RLock()
	defer cpm.runningMutex.RUnlock()
	return cpm.running
}

// GetConfig 获取配置
func (cpm *CacheProtectionManager) GetConfig() *CacheProtectionConfig {
	return cpm.config
}

// ProtectedGet 受保护的缓存获取
func (cpm *CacheProtectionManager) ProtectedGet(key string, loader func() (interface{}, error)) (interface{}, error) {
	startTime := time.Now()
	defer func() {
		cpm.updateMetrics(time.Since(startTime))
	}()

	// 1. 检查熔断器状态
	if cpm.isStrategyEnabled(ProtectionCircuitBreaker) && cpm.circuitBreaker != nil {
		if !cpm.circuitBreaker.AllowRequest() {
			cpm.incrementMetric("circuit_breaker_trips")
			return nil, fmt.Errorf("熔断器开启，拒绝请求")
		}
	}

	// 2. 尝试从缓存获取
	value, err := cpm.cacheManager.Get(key)
	if err == nil && value != nil {
		// 缓存命中
		cpm.incrementMetric("total_requests")
		if cpm.circuitBreaker != nil {
			cpm.circuitBreaker.RecordSuccess()
		}
		return value, nil
	}

	// 3. 缓存未命中，进行防护检查
	cpm.incrementMetric("total_requests")

	// 检查布隆过滤器（防穿透）
	if cpm.isStrategyEnabled(ProtectionBloomFilter) && cpm.bloomFilter != nil {
		if !cpm.bloomFilter.MightContain(key) {
			cpm.incrementMetric("penetration_blocked")
			cpm.incrementMetric("bloom_filter_misses")

			// 缓存空值
			if cpm.isStrategyEnabled(ProtectionNullCache) {
				cpm.cacheNullValue(key)
			}

			return nil, fmt.Errorf("数据不存在（布隆过滤器检查）")
		}
		cpm.incrementMetric("bloom_filter_hits")
	}

	// 检查空值缓存
	if cpm.isStrategyEnabled(ProtectionNullCache) && cpm.isNullCached(key) {
		cpm.incrementMetric("penetration_blocked")
		return nil, fmt.Errorf("数据不存在（空值缓存）")
	}

	// 4. 使用分布式锁防止击穿
	if cpm.isStrategyEnabled(ProtectionDistributedLock) {
		return cpm.getWithDistributedLock(key, loader)
	}

	// 5. 直接加载数据
	return cpm.loadAndCache(key, loader)
}

// getWithDistributedLock 使用分布式锁获取数据
func (cpm *CacheProtectionManager) getWithDistributedLock(key string, loader func() (interface{}, error)) (interface{}, error) {
	lockKey := cpm.keyManager.GenerateLockKey(key)

	// 获取或创建分布式锁
	lock := cpm.getOrCreateLock(lockKey)

	// 尝试获取锁
	if err := lock.TryAcquire(cpm.config.DistributedLockConfig.AcquireTimeout); err != nil {
		cpm.incrementMetric("lock_timeouts")

		// 获取锁失败，等待后重试获取缓存
		time.Sleep(cpm.config.DistributedLockConfig.RetryInterval)
		value, err := cpm.cacheManager.Get(key)
		if err == nil && value != nil {
			return value, nil
		}

		return nil, fmt.Errorf("获取分布式锁失败: %w", err)
	}

	cpm.incrementMetric("lock_acquisitions")
	defer lock.Release()

	// 双重检查缓存
	value, err := cpm.cacheManager.Get(key)
	if err == nil && value != nil {
		return value, nil
	}

	// 加载数据并缓存
	return cpm.loadAndCache(key, loader)
}

// loadAndCache 加载数据并缓存
func (cpm *CacheProtectionManager) loadAndCache(key string, loader func() (interface{}, error)) (interface{}, error) {
	// 加载数据
	data, err := loader()
	if err != nil {
		if cpm.circuitBreaker != nil {
			cpm.circuitBreaker.RecordFailure()
		}

		// 缓存空值（防穿透）
		if cpm.isStrategyEnabled(ProtectionNullCache) {
			cpm.cacheNullValue(key)
		}

		return nil, err
	}

	// 成功加载数据
	if cpm.circuitBreaker != nil {
		cpm.circuitBreaker.RecordSuccess()
	}

	// 添加到布隆过滤器
	if cpm.isStrategyEnabled(ProtectionBloomFilter) && cpm.bloomFilter != nil {
		cpm.bloomFilter.Add(key)
	}

	// 计算TTL（防雪崩）
	ttl := cpm.calculateTTL(key)

	// 缓存数据
	if err := cpm.cacheManager.Set(key, data, ttl); err != nil {
		logger.Error(fmt.Sprintf("缓存数据失败: %v", err))
	}

	return data, nil
}

// calculateTTL 计算TTL（防雪崩）
func (cpm *CacheProtectionManager) calculateTTL(key string) time.Duration {
	if !cpm.isStrategyEnabled(ProtectionRandomTTL) || cpm.config.RandomTTLConfig == nil {
		return GetTTL("default")
	}

	config := cpm.config.RandomTTLConfig

	// 生成随机TTL
	randomDuration := time.Duration(cpm.hashString(key)%int64(config.RandomRange.Seconds())) * time.Second
	ttl := config.BaseTTL + randomDuration

	// 确保在最小和最大TTL范围内
	if ttl < config.MinTTL {
		ttl = config.MinTTL
	}
	if ttl > config.MaxTTL {
		ttl = config.MaxTTL
	}

	return ttl
}

// hashString 字符串哈希
func (cpm *CacheProtectionManager) hashString(s string) int64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return int64(h.Sum64())
}

// getOrCreateLock 获取或创建分布式锁
func (cpm *CacheProtectionManager) getOrCreateLock(lockKey string) *DistributedLock {
	cpm.lockMutex.Lock()
	defer cpm.lockMutex.Unlock()

	if lock, exists := cpm.distributedLocks[lockKey]; exists {
		return lock
	}

	// 获取Redis客户端
	var redisClient *redis.Client
	if redisCacheManager, ok := cpm.cacheManager.(*RedisCacheManager); ok {
		redisClient = redisCacheManager.client.GetClient()
	} else {
		logger.Error("缓存防护管理器需要Redis缓存管理器来支持分布式锁")
		return nil
	}

	lock := &DistributedLock{
		rdb:        redisClient,
		key:        lockKey,
		value:      cpm.generateLockValue(),
		expiration: cpm.config.DistributedLockConfig.LockTimeout,
		ctx:        cpm.ctx,
	}

	cpm.distributedLocks[lockKey] = lock
	return lock
}

// generateLockValue 生成锁的唯一值
func (cpm *CacheProtectionManager) generateLockValue() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// cacheNullValue 缓存空值
func (cpm *CacheProtectionManager) cacheNullValue(key string) {
	if cpm.config.NullCacheConfig == nil {
		return
	}

	cpm.nullCacheMutex.Lock()
	defer cpm.nullCacheMutex.Unlock()

	// 检查空值缓存大小限制
	if len(cpm.nullCache) >= cpm.config.NullCacheConfig.MaxNullKeys {
		// 清理最旧的条目
		oldestTime := time.Now()
		oldestKey := ""
		for k, t := range cpm.nullCache {
			if t.Before(oldestTime) {
				oldestTime = t
				oldestKey = k
			}
		}
		if oldestKey != "" {
			delete(cpm.nullCache, oldestKey)
		}
	}

	cpm.nullCache[key] = time.Now()
}

// isNullCached 检查是否为空值缓存
func (cpm *CacheProtectionManager) isNullCached(key string) bool {
	if cpm.config.NullCacheConfig == nil {
		return false
	}

	cpm.nullCacheMutex.RLock()
	defer cpm.nullCacheMutex.RUnlock()

	cachedTime, exists := cpm.nullCache[key]
	if !exists {
		return false
	}

	// 检查是否过期
	if time.Since(cachedTime) > cpm.config.NullCacheConfig.TTL {
		// 异步清理过期条目
		go func() {
			cpm.nullCacheMutex.Lock()
			delete(cpm.nullCache, key)
			cpm.nullCacheMutex.Unlock()
		}()
		return false
	}

	return true
}

// incrementMetric 增加指标计数
func (cpm *CacheProtectionManager) incrementMetric(metricName string) {
	cpm.metricsMutex.Lock()
	defer cpm.metricsMutex.Unlock()

	switch metricName {
	case "total_requests":
		cpm.metrics.TotalRequests++
	case "protected_requests":
		cpm.metrics.ProtectedRequests++
	case "penetration_attempts":
		cpm.metrics.PenetrationAttempts++
	case "penetration_blocked":
		cpm.metrics.PenetrationBlocked++
	case "bloom_filter_hits":
		cpm.metrics.BloomFilterHits++
	case "bloom_filter_misses":
		cpm.metrics.BloomFilterMisses++
	case "breakdown_attempts":
		cpm.metrics.BreakdownAttempts++
	case "breakdown_blocked":
		cpm.metrics.BreakdownBlocked++
	case "lock_acquisitions":
		cpm.metrics.LockAcquisitions++
	case "lock_timeouts":
		cpm.metrics.LockTimeouts++
	case "avalanche_detected":
		cpm.metrics.AvalancheDetected++
	case "avalanche_prevented":
		cpm.metrics.AvalanchePrevented++
	case "circuit_breaker_trips":
		cpm.metrics.CircuitBreakerTrips++
	}

	// 更新防护率
	if cpm.metrics.TotalRequests > 0 {
		cpm.metrics.ProtectionRate = float64(cpm.metrics.ProtectedRequests) / float64(cpm.metrics.TotalRequests) * 100
	}

	cpm.metrics.LastUpdated = time.Now()
}

// updateMetrics 更新性能指标
func (cpm *CacheProtectionManager) updateMetrics(responseTime time.Duration) {
	cpm.metricsMutex.Lock()
	defer cpm.metricsMutex.Unlock()

	// 更新响应时间统计
	if cpm.metrics.MinResponseTime == 0 || responseTime < cpm.metrics.MinResponseTime {
		cpm.metrics.MinResponseTime = responseTime
	}
	if responseTime > cpm.metrics.MaxResponseTime {
		cpm.metrics.MaxResponseTime = responseTime
	}

	// 计算平均响应时间（简化版本）
	if cpm.metrics.AverageResponseTime == 0 {
		cpm.metrics.AverageResponseTime = responseTime
	} else {
		cpm.metrics.AverageResponseTime = (cpm.metrics.AverageResponseTime + responseTime) / 2
	}
}

// DistributedLock 方法实现

// TryAcquire 尝试获取锁
func (dl *DistributedLock) TryAcquire(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(dl.ctx, timeout)
	defer cancel()

	// 使用SET命令的NX和EX选项实现原子操作
	success, err := dl.rdb.SetNX(ctx, dl.key, dl.value, dl.expiration).Result()
	if err != nil {
		return fmt.Errorf("获取锁失败: %w", err)
	}

	if !success {
		return fmt.Errorf("锁已被占用")
	}

	return nil
}

// Release 释放锁
func (dl *DistributedLock) Release() error {
	// 使用Lua脚本确保只释放自己的锁
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`

	result, err := dl.rdb.Eval(dl.ctx, script, []string{dl.key}, dl.value).Result()
	if err != nil {
		return fmt.Errorf("释放锁失败: %w", err)
	}

	if result.(int64) == 0 {
		return fmt.Errorf("锁不存在或不属于当前持有者")
	}

	return nil
}

// BloomFilter 方法实现

// Add 添加元素到布隆过滤器
func (bf *BloomFilter) Add(item string) error {
	hashes := bf.getHashes(item)

	pipe := bf.rdb.Pipeline()
	for _, hash := range hashes {
		pipe.SetBit(bf.ctx, bf.key, int64(hash), 1)
	}

	_, err := pipe.Exec(bf.ctx)
	return err
}

// MightContain 检查元素是否可能存在
func (bf *BloomFilter) MightContain(item string) bool {
	hashes := bf.getHashes(item)

	pipe := bf.rdb.Pipeline()
	cmds := make([]*redis.IntCmd, len(hashes))
	for i, hash := range hashes {
		cmds[i] = pipe.GetBit(bf.ctx, bf.key, int64(hash))
	}

	_, err := pipe.Exec(bf.ctx)
	if err != nil {
		return false
	}

	for _, cmd := range cmds {
		if cmd.Val() == 0 {
			return false
		}
	}

	return true
}

// getHashes 获取哈希值
func (bf *BloomFilter) getHashes(item string) []uint64 {
	hashes := make([]uint64, bf.hashFunctions)

	h1 := fnv.New64()
	h1.Write([]byte(item))
	hash1 := h1.Sum64()

	h2 := fnv.New64a()
	h2.Write([]byte(item))
	hash2 := h2.Sum64()

	for i := uint64(0); i < bf.hashFunctions; i++ {
		hashes[i] = (hash1 + i*hash2) % bf.bitArraySize
	}

	return hashes
}

// CircuitBreaker 方法实现

// AllowRequest 检查是否允许请求
func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	switch cb.state {
	case CircuitBreakerClosed:
		return true
	case CircuitBreakerOpen:
		// 检查是否可以进入半开状态
		if time.Since(cb.lastFailureTime) > cb.config.RecoveryTimeout {
			cb.mutex.RUnlock()
			cb.mutex.Lock()
			cb.state = CircuitBreakerHalfOpen
			cb.mutex.Unlock()
			cb.mutex.RLock()
			return true
		}
		return false
	case CircuitBreakerHalfOpen:
		return true
	default:
		return false
	}
}

// RecordSuccess 记录成功
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if cb.state == CircuitBreakerHalfOpen {
		cb.state = CircuitBreakerClosed
		cb.failureCount = 0
	}
}

// RecordFailure 记录失败
func (cb *CircuitBreaker) RecordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.failureCount >= int64(cb.config.FailureThreshold) {
		cb.state = CircuitBreakerOpen
	}
}

// 后台任务方法

// startMetricsCollector 启动指标收集器
func (cpm *CacheProtectionManager) startMetricsCollector() {
	ticker := time.NewTicker(cpm.config.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cpm.collectMetrics()
		case <-cpm.ctx.Done():
			return
		}
	}
}

// collectMetrics 收集指标
func (cpm *CacheProtectionManager) collectMetrics() {
	cpm.metricsMutex.RLock()
	metrics := *cpm.metrics
	cpm.metricsMutex.RUnlock()

	// 检查告警阈值
	if metrics.TotalRequests > 0 {
		failureRate := float64(metrics.TotalRequests-metrics.ProtectedRequests) / float64(metrics.TotalRequests)
		if failureRate > cpm.config.AlertThreshold {
			logger.Warn(fmt.Sprintf("缓存防护失败率过高: %.2f%%, 阈值: %.2f%%",
				failureRate*100, cpm.config.AlertThreshold*100))
		}
	}

	logger.Info(fmt.Sprintf("缓存防护指标: 总请求=%d, 防护率=%.2f%%, 穿透阻止=%d, 击穿阻止=%d",
		metrics.TotalRequests, metrics.ProtectionRate, metrics.PenetrationBlocked, metrics.BreakdownBlocked))
}

// startNullCacheCleanup 启动空值缓存清理
func (cpm *CacheProtectionManager) startNullCacheCleanup() {
	if cpm.config.NullCacheConfig == nil {
		return
	}

	ticker := time.NewTicker(cpm.config.NullCacheConfig.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cpm.cleanupNullCache()
		case <-cpm.ctx.Done():
			return
		}
	}
}

// cleanupNullCache 清理过期的空值缓存
func (cpm *CacheProtectionManager) cleanupNullCache() {
	if cpm.config.NullCacheConfig == nil {
		return
	}

	cpm.nullCacheMutex.Lock()
	defer cpm.nullCacheMutex.Unlock()

	now := time.Now()
	expiredKeys := make([]string, 0)

	for key, cachedTime := range cpm.nullCache {
		if now.Sub(cachedTime) > cpm.config.NullCacheConfig.TTL {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		delete(cpm.nullCache, key)
	}

	if len(expiredKeys) > 0 {
		logger.Info(fmt.Sprintf("清理过期空值缓存: %d个", len(expiredKeys)))
	}
}

// startBloomFilterRefresh 启动布隆过滤器刷新
func (cpm *CacheProtectionManager) startBloomFilterRefresh() {
	if cpm.config.BloomFilterConfig == nil {
		return
	}

	ticker := time.NewTicker(cpm.config.BloomFilterConfig.RefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cpm.refreshBloomFilter()
		case <-cpm.ctx.Done():
			return
		}
	}
}

// refreshBloomFilter 刷新布隆过滤器
func (cpm *CacheProtectionManager) refreshBloomFilter() {
	if cpm.bloomFilter == nil {
		return
	}

	// 这里可以实现布隆过滤器的刷新逻辑
	// 例如：重新加载热点数据到布隆过滤器
	logger.Info("布隆过滤器刷新完成")
}

// GetMetrics 获取防护指标
func (cpm *CacheProtectionManager) GetMetrics() *ProtectionMetrics {
	cpm.metricsMutex.RLock()
	defer cpm.metricsMutex.RUnlock()

	// 返回指标的副本
	metrics := *cpm.metrics
	return &metrics
}

// ResetMetrics 重置指标
func (cpm *CacheProtectionManager) ResetMetrics() {
	cpm.metricsMutex.Lock()
	defer cpm.metricsMutex.Unlock()

	cpm.metrics = &ProtectionMetrics{
		LastResetTime: time.Now(),
	}

	logger.Info("缓存防护指标已重置")
}

// AddToBloomFilter 添加键到布隆过滤器
func (cpm *CacheProtectionManager) AddToBloomFilter(key string) error {
	if !cpm.isStrategyEnabled(ProtectionBloomFilter) || cpm.bloomFilter == nil {
		return fmt.Errorf("布隆过滤器未启用")
	}

	return cpm.bloomFilter.Add(key)
}

// CheckBloomFilter 检查布隆过滤器
func (cpm *CacheProtectionManager) CheckBloomFilter(key string) bool {
	if !cpm.isStrategyEnabled(ProtectionBloomFilter) || cpm.bloomFilter == nil {
		return true // 如果未启用，默认允许
	}

	return cpm.bloomFilter.MightContain(key)
}
