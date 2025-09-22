package cache

import (
	"mall-go/internal/config"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRedisClient(t *testing.T) {
	// 测试配置
	cfg := config.RedisConfig{
		Host:         "localhost",
		Port:         6379,
		Password:     "",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 2,
		MaxRetries:   3,
		DialTimeout:  5,
		ReadTimeout:  3,
		WriteTimeout: 3,
		IdleTimeout:  300,
		MaxConnAge:   3600,
	}

	// 创建Redis客户端
	client, err := NewRedisClient(cfg)

	// 如果Redis服务未启动，跳过测试
	if err != nil {
		t.Skipf("Redis服务未启动，跳过测试: %v", err)
		return
	}

	require.NoError(t, err)
	require.NotNil(t, client)

	// 测试连接
	err = client.Ping()
	assert.NoError(t, err)

	// 测试连接池统计
	stats := client.GetConnectionStats()
	assert.NotNil(t, stats)
	t.Logf("连接池统计: 总连接数=%d, 空闲连接数=%d", stats.TotalConns, stats.IdleConns)

	// 测试健康检查
	err = client.HealthCheck()
	assert.NoError(t, err)

	// 关闭连接
	err = client.Close()
	assert.NoError(t, err)
}

func TestRedisClientConcurrency(t *testing.T) {
	cfg := config.RedisConfig{
		Host:         "localhost",
		Port:         6379,
		Password:     "",
		DB:           0,
		PoolSize:     50,
		MinIdleConns: 10,
		MaxRetries:   3,
		DialTimeout:  5,
		ReadTimeout:  3,
		WriteTimeout: 3,
		IdleTimeout:  300,
		MaxConnAge:   3600,
	}

	client, err := NewRedisClient(cfg)
	if err != nil {
		t.Skipf("Redis服务未启动，跳过测试: %v", err)
		return
	}
	defer client.Close()

	// 并发测试
	concurrency := 100
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// 执行ping操作
			err := client.Ping()
			assert.NoError(t, err, "Goroutine %d ping failed", id)
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < concurrency; i++ {
		select {
		case <-done:
			// 成功完成
		case <-time.After(10 * time.Second):
			t.Fatalf("并发测试超时")
		}
	}

	// 检查连接池状态
	stats := client.GetConnectionStats()
	t.Logf("并发测试后连接池统计: 总连接数=%d, 空闲连接数=%d, 命中数=%d, 未命中数=%d",
		stats.TotalConns, stats.IdleConns, stats.Hits, stats.Misses)

	// 验证连接池没有泄漏
	assert.True(t, stats.TotalConns <= 50, "连接数不应超过池大小")
}

func TestInitRedis(t *testing.T) {
	cfg := config.RedisConfig{
		Host:         "localhost",
		Port:         6379,
		Password:     "",
		DB:           0,
		PoolSize:     20,
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  5,
		ReadTimeout:  3,
		WriteTimeout: 3,
		IdleTimeout:  300,
		MaxConnAge:   3600,
	}

	// 初始化Redis
	err := InitRedis(cfg)
	if err != nil {
		t.Skipf("Redis服务未启动，跳过测试: %v", err)
		return
	}

	// 验证全局客户端
	assert.NotNil(t, GlobalRedisClient)

	// 测试全局客户端
	client := GetRedisClient()
	assert.NotNil(t, client)
	assert.Equal(t, GlobalRedisClient, client)

	// 测试连接
	err = client.Ping()
	assert.NoError(t, err)

	// 清理
	if GlobalRedisClient != nil {
		GlobalRedisClient.Close()
		GlobalRedisClient = nil
	}
}

// BenchmarkRedisConnection 基准测试Redis连接性能
func BenchmarkRedisConnection(b *testing.B) {
	cfg := config.RedisConfig{
		Host:         "localhost",
		Port:         6379,
		Password:     "",
		DB:           0,
		PoolSize:     100,
		MinIdleConns: 10,
		MaxRetries:   3,
		DialTimeout:  5,
		ReadTimeout:  3,
		WriteTimeout: 3,
		IdleTimeout:  300,
		MaxConnAge:   3600,
	}

	client, err := NewRedisClient(cfg)
	if err != nil {
		b.Skipf("Redis服务未启动，跳过基准测试: %v", err)
		return
	}
	defer client.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := client.Ping()
			if err != nil {
				b.Errorf("Ping failed: %v", err)
			}
		}
	})
}
