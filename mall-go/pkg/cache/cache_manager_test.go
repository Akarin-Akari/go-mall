package cache

import (
	"mall-go/internal/config"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func setupTestCacheManager(t *testing.T) CacheManager {
	cfg := config.RedisConfig{
		Host:         "localhost",
		Port:         6379,
		Password:     "",
		DB:           1, // 使用测试数据库
		PoolSize:     10,
		MinIdleConns: 2,
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
		return nil
	}

	manager := NewRedisCacheManager(client)

	// 清空测试数据库
	manager.FlushDB()

	return manager
}

func TestCacheManagerBasicOperations(t *testing.T) {
	manager := setupTestCacheManager(t)
	if manager == nil {
		return
	}
	defer manager.Close()

	// 测试Set和Get
	err := manager.Set("test:key1", "test_value", 10*time.Second)
	assert.NoError(t, err)

	value, err := manager.Get("test:key1")
	assert.NoError(t, err)
	assert.Equal(t, "test_value", value)

	// 测试Exists
	exists := manager.Exists("test:key1")
	assert.True(t, exists)

	exists = manager.Exists("test:nonexistent")
	assert.False(t, exists)

	// 测试Delete
	err = manager.Delete("test:key1")
	assert.NoError(t, err)

	value, err = manager.Get("test:key1")
	assert.NoError(t, err)
	assert.Nil(t, value)
}

func TestCacheManagerJSONSerialization(t *testing.T) {
	manager := setupTestCacheManager(t)
	if manager == nil {
		return
	}
	defer manager.Close()

	// 测试复杂对象序列化
	testData := map[string]interface{}{
		"name":  "test_product",
		"price": 99.99,
		"tags":  []string{"electronics", "mobile"},
	}

	err := manager.Set("test:json", testData, 10*time.Second)
	assert.NoError(t, err)

	value, err := manager.Get("test:json")
	assert.NoError(t, err)
	assert.NotNil(t, value)

	// 验证反序列化结果
	result, ok := value.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "test_product", result["name"])
	assert.Equal(t, 99.99, result["price"])
}

func TestCacheManagerBatchOperations(t *testing.T) {
	manager := setupTestCacheManager(t)
	if manager == nil {
		return
	}
	defer manager.Close()

	// 测试MSet
	pairs := map[string]interface{}{
		"test:batch1": "value1",
		"test:batch2": "value2",
		"test:batch3": "value3",
	}

	err := manager.MSet(pairs, 10*time.Second)
	assert.NoError(t, err)

	// 测试MGet
	keys := []string{"test:batch1", "test:batch2", "test:batch3", "test:nonexistent"}
	values, err := manager.MGet(keys)
	assert.NoError(t, err)
	assert.Len(t, values, 4)
	assert.Equal(t, "value1", values[0])
	assert.Equal(t, "value2", values[1])
	assert.Equal(t, "value3", values[2])
	assert.Nil(t, values[3]) // 不存在的键

	// 测试MDelete
	deleteKeys := []string{"test:batch1", "test:batch2"}
	err = manager.MDelete(deleteKeys)
	assert.NoError(t, err)

	// 验证删除结果
	value, err := manager.Get("test:batch1")
	assert.NoError(t, err)
	assert.Nil(t, value)
}

func TestCacheManagerHashOperations(t *testing.T) {
	manager := setupTestCacheManager(t)
	if manager == nil {
		return
	}
	defer manager.Close()

	key := "test:hash"

	// 测试HSet和HGet
	err := manager.HSet(key, "field1", "value1")
	assert.NoError(t, err)

	value, err := manager.HGet(key, "field1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", value)

	// 测试HMSet
	fields := map[string]interface{}{
		"field2": "value2",
		"field3": map[string]string{"nested": "data"},
	}
	err = manager.HMSet(key, fields)
	assert.NoError(t, err)

	// 测试HMGet
	fieldNames := []string{"field1", "field2", "field3"}
	values, err := manager.HMGet(key, fieldNames)
	assert.NoError(t, err)
	assert.Len(t, values, 3)
	assert.Equal(t, "value1", values[0])
	assert.Equal(t, "value2", values[1])

	// 测试HExists
	exists := manager.HExists(key, "field1")
	assert.True(t, exists)

	exists = manager.HExists(key, "nonexistent")
	assert.False(t, exists)

	// 测试HDelete
	err = manager.HDelete(key, []string{"field1"})
	assert.NoError(t, err)

	value, err = manager.HGet(key, "field1")
	assert.NoError(t, err)
	assert.Nil(t, value)
}

func TestCacheManagerListOperations(t *testing.T) {
	manager := setupTestCacheManager(t)
	if manager == nil {
		return
	}
	defer manager.Close()

	key := "test:list"

	// 测试RPush
	err := manager.RPush(key, "item1", "item2", "item3")
	assert.NoError(t, err)

	// 测试LLen
	length, err := manager.LLen(key)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), length)

	// 测试LRange
	values, err := manager.LRange(key, 0, -1)
	assert.NoError(t, err)
	assert.Len(t, values, 3)
	assert.Equal(t, "item1", values[0])
	assert.Equal(t, "item2", values[1])
	assert.Equal(t, "item3", values[2])

	// 测试LPush
	err = manager.LPush(key, "item0")
	assert.NoError(t, err)

	// 测试LPop
	value, err := manager.LPop(key)
	assert.NoError(t, err)
	assert.Equal(t, "item0", value)

	// 测试RPop
	value, err = manager.RPop(key)
	assert.NoError(t, err)
	assert.Equal(t, "item3", value)
}

func TestCacheManagerSetOperations(t *testing.T) {
	manager := setupTestCacheManager(t)
	if manager == nil {
		return
	}
	defer manager.Close()

	key := "test:set"

	// 测试SAdd
	err := manager.SAdd(key, "member1", "member2", "member3")
	assert.NoError(t, err)

	// 测试SMembers
	members, err := manager.SMembers(key)
	assert.NoError(t, err)
	assert.Len(t, members, 3)

	// 测试SIsMember
	isMember := manager.SIsMember(key, "member1")
	assert.True(t, isMember)

	isMember = manager.SIsMember(key, "nonexistent")
	assert.False(t, isMember)

	// 测试SRem
	err = manager.SRem(key, "member1")
	assert.NoError(t, err)

	isMember = manager.SIsMember(key, "member1")
	assert.False(t, isMember)
}

func TestCacheManagerZSetOperations(t *testing.T) {
	manager := setupTestCacheManager(t)
	if manager == nil {
		return
	}
	defer manager.Close()

	key := "test:zset"

	// 测试ZAdd
	members := []redis.Z{
		{Score: 1.0, Member: "member1"},
		{Score: 2.0, Member: "member2"},
		{Score: 3.0, Member: "member3"},
	}
	err := manager.ZAdd(key, members...)
	assert.NoError(t, err)

	// 测试ZRange
	values, err := manager.ZRange(key, 0, -1)
	assert.NoError(t, err)
	assert.Len(t, values, 3)

	// 测试ZScore
	score, err := manager.ZScore(key, "member2")
	assert.NoError(t, err)
	assert.Equal(t, 2.0, score)

	// 测试ZRangeByScore
	values, err = manager.ZRangeByScore(key, "1", "2")
	assert.NoError(t, err)
	assert.Len(t, values, 2)

	// 测试ZRem
	err = manager.ZRem(key, "member1")
	assert.NoError(t, err)

	values, err = manager.ZRange(key, 0, -1)
	assert.NoError(t, err)
	assert.Len(t, values, 2)
}

func TestCacheManagerMetrics(t *testing.T) {
	manager := setupTestCacheManager(t)
	if manager == nil {
		return
	}
	defer manager.Close()

	// 执行一些操作来生成指标
	manager.Set("test:metrics1", "value1", 10*time.Second)
	manager.Set("test:metrics2", "value2", 10*time.Second)
	manager.Get("test:metrics1")    // 命中
	manager.Get("test:metrics2")    // 命中
	manager.Get("test:nonexistent") // 未命中

	// 获取指标
	metrics := manager.GetMetrics()
	assert.NotNil(t, metrics)
	assert.True(t, metrics.TotalOps > 0)
	assert.True(t, metrics.HitCount > 0)
	assert.True(t, metrics.MissCount > 0)
	assert.True(t, metrics.HitRate > 0 && metrics.HitRate < 1)

	t.Logf("缓存指标: 总操作数=%d, 命中数=%d, 未命中数=%d, 命中率=%.2f",
		metrics.TotalOps, metrics.HitCount, metrics.MissCount, metrics.HitRate)
}

func TestCacheManagerHealthCheck(t *testing.T) {
	manager := setupTestCacheManager(t)
	if manager == nil {
		return
	}
	defer manager.Close()

	// 测试健康检查
	err := manager.HealthCheck()
	assert.NoError(t, err)

	// 测试连接池统计
	stats := manager.GetConnectionStats()
	assert.NotNil(t, stats)
	t.Logf("连接池统计: 总连接数=%d, 空闲连接数=%d", stats.TotalConns, stats.IdleConns)
}
