package cache

import (
	"context"
	"fmt"
	"mall-go/internal/config"
	"mall-go/pkg/logger"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient Redis客户端封装
type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisClient 创建Redis客户端
func NewRedisClient(cfg config.RedisConfig) (*RedisClient, error) {
	// 创建Redis客户端配置
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		// 连接池配置
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,
		// 超时配置
		DialTimeout:     time.Duration(cfg.DialTimeout) * time.Second,
		ReadTimeout:     time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout:    time.Duration(cfg.WriteTimeout) * time.Second,
		ConnMaxIdleTime: time.Duration(cfg.IdleTimeout) * time.Second,
		ConnMaxLifetime: time.Duration(cfg.MaxConnAge) * time.Second,
		// 性能优化配置
		PoolTimeout: time.Duration(cfg.PoolTimeout) * time.Second,
		// 重试配置优化
		MaxRetryBackoff: 512 * time.Millisecond,
		MinRetryBackoff: 8 * time.Millisecond,
	})

	ctx := context.Background()

	// 测试连接
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Error("Redis连接失败: " + err.Error())
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	logger.Info("Redis连接成功")

	return &RedisClient{
		client: rdb,
		ctx:    ctx,
	}, nil
}

// GetClient 获取Redis客户端
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}

// GetContext 获取上下文
func (r *RedisClient) GetContext() context.Context {
	return r.ctx
}

// Ping 测试连接
func (r *RedisClient) Ping() error {
	return r.client.Ping(r.ctx).Err()
}

// Close 关闭连接
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// GetConnectionStats 获取连接池统计信息
func (r *RedisClient) GetConnectionStats() *redis.PoolStats {
	return r.client.PoolStats()
}

// LogConnectionStats 记录连接池统计信息
func (r *RedisClient) LogConnectionStats() {
	stats := r.GetConnectionStats()
	logger.Info(fmt.Sprintf("Redis连接池统计 - 总连接数: %d, 空闲连接数: %d, 过期连接数: %d, 命中数: %d, 未命中数: %d, 超时数: %d",
		stats.TotalConns,
		stats.IdleConns,
		stats.StaleConns,
		stats.Hits,
		stats.Misses,
		stats.Timeouts,
	))
}

// HealthCheck 健康检查
func (r *RedisClient) HealthCheck() error {
	// 执行简单的ping命令
	if err := r.Ping(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	// 检查连接池状态
	stats := r.GetConnectionStats()
	if stats.TotalConns == 0 {
		return fmt.Errorf("no redis connections available")
	}

	// 检查是否有过多的超时
	if stats.Timeouts > 100 {
		logger.Warn(fmt.Sprintf("Redis连接池超时次数过多: %d", stats.Timeouts))
	}

	return nil
}

// 全局Redis客户端实例
var GlobalRedisClient *RedisClient

// InitRedis 初始化Redis连接
func InitRedis(cfg config.RedisConfig) error {
	client, err := NewRedisClient(cfg)
	if err != nil {
		return err
	}

	GlobalRedisClient = client

	// 启动连接池监控
	go monitorConnectionPool(client)

	return nil
}

// monitorConnectionPool 监控连接池状态
func monitorConnectionPool(client *RedisClient) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := client.HealthCheck(); err != nil {
			logger.Error("Redis健康检查失败: " + err.Error())
		} else {
			client.LogConnectionStats()
		}
	}
}

// GetRedisClient 获取全局Redis客户端
func GetRedisClient() *RedisClient {
	return GlobalRedisClient
}
