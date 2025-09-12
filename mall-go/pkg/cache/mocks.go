package cache

import (
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"

	"mall-go/pkg/optimistic"
)

// SharedMockCacheManager 共享的模拟缓存管理器
type SharedMockCacheManager struct {
	mock.Mock
}

func (m *SharedMockCacheManager) Get(key string) (interface{}, error) {
	args := m.Called(key)
	return args.Get(0), args.Error(1)
}

func (m *SharedMockCacheManager) Set(key string, value interface{}, ttl time.Duration) error {
	args := m.Called(key, value, ttl)
	return args.Error(0)
}

func (m *SharedMockCacheManager) Delete(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *SharedMockCacheManager) Exists(key string) bool {
	args := m.Called(key)
	return args.Bool(0)
}

func (m *SharedMockCacheManager) Expire(key string, ttl time.Duration) error {
	args := m.Called(key, ttl)
	return args.Error(0)
}

func (m *SharedMockCacheManager) TTL(key string) (time.Duration, error) {
	args := m.Called(key)
	return args.Get(0).(time.Duration), args.Error(1)
}

func (m *SharedMockCacheManager) MGet(keys []string) ([]interface{}, error) {
	args := m.Called(keys)
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *SharedMockCacheManager) MSet(pairs map[string]interface{}, ttl time.Duration) error {
	args := m.Called(pairs, ttl)
	return args.Error(0)
}

func (m *SharedMockCacheManager) MDelete(keys []string) error {
	args := m.Called(keys)
	return args.Error(0)
}

func (m *SharedMockCacheManager) HGet(key, field string) (interface{}, error) {
	args := m.Called(key, field)
	return args.Get(0), args.Error(1)
}

func (m *SharedMockCacheManager) HSet(key, field string, value interface{}) error {
	args := m.Called(key, field, value)
	return args.Error(0)
}

func (m *SharedMockCacheManager) HMGet(key string, fields []string) ([]interface{}, error) {
	args := m.Called(key, fields)
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *SharedMockCacheManager) HMSet(key string, fields map[string]interface{}) error {
	args := m.Called(key, fields)
	return args.Error(0)
}

func (m *SharedMockCacheManager) HDelete(key string, fields []string) error {
	args := m.Called(key, fields)
	return args.Error(0)
}

func (m *SharedMockCacheManager) HExists(key, field string) bool {
	args := m.Called(key, field)
	return args.Bool(0)
}

func (m *SharedMockCacheManager) LPush(key string, values ...interface{}) error {
	args := m.Called(key, values)
	return args.Error(0)
}

func (m *SharedMockCacheManager) RPush(key string, values ...interface{}) error {
	args := m.Called(key, values)
	return args.Error(0)
}

func (m *SharedMockCacheManager) LPop(key string) (interface{}, error) {
	args := m.Called(key)
	return args.Get(0), args.Error(1)
}

func (m *SharedMockCacheManager) RPop(key string) (interface{}, error) {
	args := m.Called(key)
	return args.Get(0), args.Error(1)
}

func (m *SharedMockCacheManager) LRange(key string, start, stop int64) ([]interface{}, error) {
	args := m.Called(key, start, stop)
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *SharedMockCacheManager) LLen(key string) (int64, error) {
	args := m.Called(key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *SharedMockCacheManager) SAdd(key string, members ...interface{}) error {
	args := m.Called(key, members)
	return args.Error(0)
}

func (m *SharedMockCacheManager) SMembers(key string) ([]interface{}, error) {
	args := m.Called(key)
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *SharedMockCacheManager) SIsMember(key string, member interface{}) bool {
	args := m.Called(key, member)
	return args.Bool(0)
}

func (m *SharedMockCacheManager) SRem(key string, members ...interface{}) error {
	args := m.Called(key, members)
	return args.Error(0)
}

func (m *SharedMockCacheManager) ZAdd(key string, members ...redis.Z) error {
	args := m.Called(key, members)
	return args.Error(0)
}

func (m *SharedMockCacheManager) ZRange(key string, start, stop int64) ([]interface{}, error) {
	args := m.Called(key, start, stop)
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *SharedMockCacheManager) ZRangeByScore(key string, min, max string) ([]interface{}, error) {
	args := m.Called(key, min, max)
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *SharedMockCacheManager) ZRem(key string, members ...interface{}) error {
	args := m.Called(key, members)
	return args.Error(0)
}

func (m *SharedMockCacheManager) ZScore(key string, member string) (float64, error) {
	args := m.Called(key, member)
	return args.Get(0).(float64), args.Error(1)
}

func (m *SharedMockCacheManager) GetMetrics() *CacheMetrics {
	args := m.Called()
	return args.Get(0).(*CacheMetrics)
}

func (m *SharedMockCacheManager) GetConnectionStats() *redis.PoolStats {
	args := m.Called()
	return args.Get(0).(*redis.PoolStats)
}

func (m *SharedMockCacheManager) HealthCheck() error {
	args := m.Called()
	return args.Error(0)
}

func (m *SharedMockCacheManager) FlushDB() error {
	args := m.Called()
	return args.Error(0)
}

func (m *SharedMockCacheManager) FlushAll() error {
	args := m.Called()
	return args.Error(0)
}

func (m *SharedMockCacheManager) Close() error {
	args := m.Called()
	return args.Error(0)
}

// SharedMockOptimisticLockService 共享的模拟乐观锁服务
type SharedMockOptimisticLockService struct {
	mock.Mock
}

func (m *SharedMockOptimisticLockService) UpdateWithOptimisticLock(
	tableName string,
	id uint,
	updates map[string]interface{},
	config *optimistic.UpdateConfig,
) *optimistic.UpdateResult {
	args := m.Called(tableName, id, updates, config)
	return args.Get(0).(*optimistic.UpdateResult)
}
