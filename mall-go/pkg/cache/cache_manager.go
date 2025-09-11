package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheManager 缓存管理器接口
type CacheManager interface {
	// 基础CRUD操作
	Get(key string) (interface{}, error)
	Set(key string, value interface{}, ttl time.Duration) error
	Delete(key string) error
	Exists(key string) bool
	Expire(key string, ttl time.Duration) error
	TTL(key string) (time.Duration, error)

	// 批量操作
	MGet(keys []string) ([]interface{}, error)
	MSet(pairs map[string]interface{}, ttl time.Duration) error
	MDelete(keys []string) error

	// Hash操作
	HGet(key, field string) (interface{}, error)
	HSet(key, field string, value interface{}) error
	HMGet(key string, fields []string) ([]interface{}, error)
	HMSet(key string, fields map[string]interface{}) error
	HDelete(key string, fields []string) error
	HExists(key, field string) bool

	// List操作
	LPush(key string, values ...interface{}) error
	RPush(key string, values ...interface{}) error
	LPop(key string) (interface{}, error)
	RPop(key string) (interface{}, error)
	LRange(key string, start, stop int64) ([]interface{}, error)
	LLen(key string) (int64, error)

	// Set操作
	SAdd(key string, members ...interface{}) error
	SMembers(key string) ([]interface{}, error)
	SIsMember(key string, member interface{}) bool
	SRem(key string, members ...interface{}) error

	// ZSet操作
	ZAdd(key string, members ...redis.Z) error
	ZRange(key string, start, stop int64) ([]interface{}, error)
	ZRangeByScore(key string, min, max string) ([]interface{}, error)
	ZRem(key string, members ...interface{}) error
	ZScore(key string, member string) (float64, error)

	// 统计功能
	GetMetrics() *CacheMetrics
	GetConnectionStats() *redis.PoolStats
	HealthCheck() error

	// 管理功能
	FlushDB() error
	FlushAll() error
	Close() error
}

// CacheMetrics 缓存指标
type CacheMetrics struct {
	HitCount    int64     `json:"hit_count"`    // 命中次数
	MissCount   int64     `json:"miss_count"`   // 未命中次数
	HitRate     float64   `json:"hit_rate"`     // 命中率
	TotalOps    int64     `json:"total_ops"`    // 总操作数
	ErrorCount  int64     `json:"error_count"`  // 错误次数
	LastUpdated time.Time `json:"last_updated"` // 最后更新时间
}

// RedisCacheManager Redis缓存管理器实现
type RedisCacheManager struct {
	client  *RedisClient
	metrics *CacheMetrics
	ctx     context.Context
}

// NewRedisCacheManager 创建Redis缓存管理器
func NewRedisCacheManager(client *RedisClient) CacheManager {
	return &RedisCacheManager{
		client: client,
		metrics: &CacheMetrics{
			LastUpdated: time.Now(),
		},
		ctx: client.GetContext(),
	}
}

// 基础CRUD操作实现

// Get 获取缓存值
func (r *RedisCacheManager) Get(key string) (interface{}, error) {
	r.metrics.TotalOps++

	result, err := r.client.GetClient().Get(r.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			r.metrics.MissCount++
			return nil, nil // 缓存未命中
		}
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("get cache failed: %w", err)
	}

	r.metrics.HitCount++
	r.updateHitRate()

	// 尝试反序列化JSON
	var value interface{}
	if err := json.Unmarshal([]byte(result), &value); err != nil {
		// 如果不是JSON，直接返回字符串
		return result, nil
	}

	return value, nil
}

// Set 设置缓存值
func (r *RedisCacheManager) Set(key string, value interface{}, ttl time.Duration) error {
	r.metrics.TotalOps++

	// 序列化值
	var data string
	switch v := value.(type) {
	case string:
		data = v
	case []byte:
		data = string(v)
	default:
		jsonData, err := json.Marshal(value)
		if err != nil {
			r.metrics.ErrorCount++
			return fmt.Errorf("marshal value failed: %w", err)
		}
		data = string(jsonData)
	}

	err := r.client.GetClient().Set(r.ctx, key, data, ttl).Err()
	if err != nil {
		r.metrics.ErrorCount++
		return fmt.Errorf("set cache failed: %w", err)
	}

	return nil
}

// Delete 删除缓存
func (r *RedisCacheManager) Delete(key string) error {
	r.metrics.TotalOps++

	err := r.client.GetClient().Del(r.ctx, key).Err()
	if err != nil {
		r.metrics.ErrorCount++
		return fmt.Errorf("delete cache failed: %w", err)
	}

	return nil
}

// Exists 检查键是否存在
func (r *RedisCacheManager) Exists(key string) bool {
	r.metrics.TotalOps++

	result, err := r.client.GetClient().Exists(r.ctx, key).Result()
	if err != nil {
		r.metrics.ErrorCount++
		return false
	}

	return result > 0
}

// Expire 设置过期时间
func (r *RedisCacheManager) Expire(key string, ttl time.Duration) error {
	r.metrics.TotalOps++

	err := r.client.GetClient().Expire(r.ctx, key, ttl).Err()
	if err != nil {
		r.metrics.ErrorCount++
		return fmt.Errorf("set expire failed: %w", err)
	}

	return nil
}

// TTL 获取剩余过期时间
func (r *RedisCacheManager) TTL(key string) (time.Duration, error) {
	r.metrics.TotalOps++

	result, err := r.client.GetClient().TTL(r.ctx, key).Result()
	if err != nil {
		r.metrics.ErrorCount++
		return 0, fmt.Errorf("get ttl failed: %w", err)
	}

	return result, nil
}

// 批量操作实现

// MGet 批量获取
func (r *RedisCacheManager) MGet(keys []string) ([]interface{}, error) {
	r.metrics.TotalOps++

	if len(keys) == 0 {
		return []interface{}{}, nil
	}

	results, err := r.client.GetClient().MGet(r.ctx, keys...).Result()
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("mget failed: %w", err)
	}

	// 统计命中和未命中
	for _, result := range results {
		if result == nil {
			r.metrics.MissCount++
		} else {
			r.metrics.HitCount++
		}
	}

	r.updateHitRate()
	return results, nil
}

// MSet 批量设置
func (r *RedisCacheManager) MSet(pairs map[string]interface{}, ttl time.Duration) error {
	r.metrics.TotalOps++

	if len(pairs) == 0 {
		return nil
	}

	// 准备数据
	args := make([]interface{}, 0, len(pairs)*2)
	for key, value := range pairs {
		// 序列化值
		var data string
		switch v := value.(type) {
		case string:
			data = v
		case []byte:
			data = string(v)
		default:
			jsonData, err := json.Marshal(value)
			if err != nil {
				r.metrics.ErrorCount++
				return fmt.Errorf("marshal value failed: %w", err)
			}
			data = string(jsonData)
		}

		args = append(args, key, data)
	}

	// 批量设置
	err := r.client.GetClient().MSet(r.ctx, args...).Err()
	if err != nil {
		r.metrics.ErrorCount++
		return fmt.Errorf("mset failed: %w", err)
	}

	// 如果需要设置TTL，逐个设置
	if ttl > 0 {
		for key := range pairs {
			if err := r.Expire(key, ttl); err != nil {
				return err
			}
		}
	}

	return nil
}

// MDelete 批量删除
func (r *RedisCacheManager) MDelete(keys []string) error {
	r.metrics.TotalOps++

	if len(keys) == 0 {
		return nil
	}

	err := r.client.GetClient().Del(r.ctx, keys...).Err()
	if err != nil {
		r.metrics.ErrorCount++
		return fmt.Errorf("mdelete failed: %w", err)
	}

	return nil
}

// updateHitRate 更新命中率
func (r *RedisCacheManager) updateHitRate() {
	total := r.metrics.HitCount + r.metrics.MissCount
	if total > 0 {
		r.metrics.HitRate = float64(r.metrics.HitCount) / float64(total)
	}
	r.metrics.LastUpdated = time.Now()
}

// GetMetrics 获取缓存指标
func (r *RedisCacheManager) GetMetrics() *CacheMetrics {
	r.updateHitRate()
	return r.metrics
}

// GetConnectionStats 获取连接池统计
func (r *RedisCacheManager) GetConnectionStats() *redis.PoolStats {
	return r.client.GetConnectionStats()
}

// HealthCheck 健康检查
func (r *RedisCacheManager) HealthCheck() error {
	return r.client.HealthCheck()
}

// FlushDB 清空当前数据库
func (r *RedisCacheManager) FlushDB() error {
	return r.client.GetClient().FlushDB(r.ctx).Err()
}

// FlushAll 清空所有数据库
func (r *RedisCacheManager) FlushAll() error {
	return r.client.GetClient().FlushAll(r.ctx).Err()
}

// Close 关闭连接
func (r *RedisCacheManager) Close() error {
	return r.client.Close()
}

// Hash操作实现

// HGet 获取Hash字段值
func (r *RedisCacheManager) HGet(key, field string) (interface{}, error) {
	r.metrics.TotalOps++

	result, err := r.client.GetClient().HGet(r.ctx, key, field).Result()
	if err != nil {
		if err == redis.Nil {
			r.metrics.MissCount++
			return nil, nil
		}
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("hget failed: %w", err)
	}

	r.metrics.HitCount++
	r.updateHitRate()

	// 尝试反序列化JSON
	var value interface{}
	if err := json.Unmarshal([]byte(result), &value); err != nil {
		return result, nil
	}

	return value, nil
}

// HSet 设置Hash字段值
func (r *RedisCacheManager) HSet(key, field string, value interface{}) error {
	r.metrics.TotalOps++

	// 序列化值
	var data string
	switch v := value.(type) {
	case string:
		data = v
	case []byte:
		data = string(v)
	default:
		jsonData, err := json.Marshal(value)
		if err != nil {
			r.metrics.ErrorCount++
			return fmt.Errorf("marshal value failed: %w", err)
		}
		data = string(jsonData)
	}

	err := r.client.GetClient().HSet(r.ctx, key, field, data).Err()
	if err != nil {
		r.metrics.ErrorCount++
		return fmt.Errorf("hset failed: %w", err)
	}

	return nil
}

// HMGet 批量获取Hash字段
func (r *RedisCacheManager) HMGet(key string, fields []string) ([]interface{}, error) {
	r.metrics.TotalOps++

	if len(fields) == 0 {
		return []interface{}{}, nil
	}

	results, err := r.client.GetClient().HMGet(r.ctx, key, fields...).Result()
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("hmget failed: %w", err)
	}

	// 统计命中和未命中
	for _, result := range results {
		if result == nil {
			r.metrics.MissCount++
		} else {
			r.metrics.HitCount++
		}
	}

	r.updateHitRate()
	return results, nil
}

// HMSet 批量设置Hash字段
func (r *RedisCacheManager) HMSet(key string, fields map[string]interface{}) error {
	r.metrics.TotalOps++

	if len(fields) == 0 {
		return nil
	}

	// 序列化所有值
	serializedFields := make(map[string]interface{})
	for field, value := range fields {
		switch v := value.(type) {
		case string:
			serializedFields[field] = v
		case []byte:
			serializedFields[field] = string(v)
		default:
			jsonData, err := json.Marshal(value)
			if err != nil {
				r.metrics.ErrorCount++
				return fmt.Errorf("marshal value failed: %w", err)
			}
			serializedFields[field] = string(jsonData)
		}
	}

	err := r.client.GetClient().HMSet(r.ctx, key, serializedFields).Err()
	if err != nil {
		r.metrics.ErrorCount++
		return fmt.Errorf("hmset failed: %w", err)
	}

	return nil
}

// HDelete 删除Hash字段
func (r *RedisCacheManager) HDelete(key string, fields []string) error {
	r.metrics.TotalOps++

	if len(fields) == 0 {
		return nil
	}

	err := r.client.GetClient().HDel(r.ctx, key, fields...).Err()
	if err != nil {
		r.metrics.ErrorCount++
		return fmt.Errorf("hdel failed: %w", err)
	}

	return nil
}

// HExists 检查Hash字段是否存在
func (r *RedisCacheManager) HExists(key, field string) bool {
	r.metrics.TotalOps++

	result, err := r.client.GetClient().HExists(r.ctx, key, field).Result()
	if err != nil {
		r.metrics.ErrorCount++
		return false
	}

	return result
}

// List操作实现

// LPush 从左侧推入列表
func (r *RedisCacheManager) LPush(key string, values ...interface{}) error {
	r.metrics.TotalOps++

	if len(values) == 0 {
		return nil
	}

	// 序列化所有值
	serializedValues := make([]interface{}, len(values))
	for i, value := range values {
		switch v := value.(type) {
		case string:
			serializedValues[i] = v
		case []byte:
			serializedValues[i] = string(v)
		default:
			jsonData, err := json.Marshal(value)
			if err != nil {
				r.metrics.ErrorCount++
				return fmt.Errorf("marshal value failed: %w", err)
			}
			serializedValues[i] = string(jsonData)
		}
	}

	err := r.client.GetClient().LPush(r.ctx, key, serializedValues...).Err()
	if err != nil {
		r.metrics.ErrorCount++
		return fmt.Errorf("lpush failed: %w", err)
	}

	return nil
}

// RPush 从右侧推入列表
func (r *RedisCacheManager) RPush(key string, values ...interface{}) error {
	r.metrics.TotalOps++

	if len(values) == 0 {
		return nil
	}

	// 序列化所有值
	serializedValues := make([]interface{}, len(values))
	for i, value := range values {
		switch v := value.(type) {
		case string:
			serializedValues[i] = v
		case []byte:
			serializedValues[i] = string(v)
		default:
			jsonData, err := json.Marshal(value)
			if err != nil {
				r.metrics.ErrorCount++
				return fmt.Errorf("marshal value failed: %w", err)
			}
			serializedValues[i] = string(jsonData)
		}
	}

	err := r.client.GetClient().RPush(r.ctx, key, serializedValues...).Err()
	if err != nil {
		r.metrics.ErrorCount++
		return fmt.Errorf("rpush failed: %w", err)
	}

	return nil
}

// LPop 从左侧弹出列表元素
func (r *RedisCacheManager) LPop(key string) (interface{}, error) {
	r.metrics.TotalOps++

	result, err := r.client.GetClient().LPop(r.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			r.metrics.MissCount++
			return nil, nil
		}
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("lpop failed: %w", err)
	}

	r.metrics.HitCount++
	r.updateHitRate()

	// 尝试反序列化JSON
	var value interface{}
	if err := json.Unmarshal([]byte(result), &value); err != nil {
		return result, nil
	}

	return value, nil
}

// RPop 从右侧弹出列表元素
func (r *RedisCacheManager) RPop(key string) (interface{}, error) {
	r.metrics.TotalOps++

	result, err := r.client.GetClient().RPop(r.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			r.metrics.MissCount++
			return nil, nil
		}
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("rpop failed: %w", err)
	}

	r.metrics.HitCount++
	r.updateHitRate()

	// 尝试反序列化JSON
	var value interface{}
	if err := json.Unmarshal([]byte(result), &value); err != nil {
		return result, nil
	}

	return value, nil
}

// LRange 获取列表范围内的元素
func (r *RedisCacheManager) LRange(key string, start, stop int64) ([]interface{}, error) {
	r.metrics.TotalOps++

	results, err := r.client.GetClient().LRange(r.ctx, key, start, stop).Result()
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("lrange failed: %w", err)
	}

	// 转换结果
	values := make([]interface{}, len(results))
	for i, result := range results {
		var value interface{}
		if err := json.Unmarshal([]byte(result), &value); err != nil {
			values[i] = result
		} else {
			values[i] = value
		}
	}

	if len(values) > 0 {
		r.metrics.HitCount++
	} else {
		r.metrics.MissCount++
	}
	r.updateHitRate()

	return values, nil
}

// LLen 获取列表长度
func (r *RedisCacheManager) LLen(key string) (int64, error) {
	r.metrics.TotalOps++

	result, err := r.client.GetClient().LLen(r.ctx, key).Result()
	if err != nil {
		r.metrics.ErrorCount++
		return 0, fmt.Errorf("llen failed: %w", err)
	}

	return result, nil
}

// Set操作实现

// SAdd 添加集合成员
func (r *RedisCacheManager) SAdd(key string, members ...interface{}) error {
	r.metrics.TotalOps++

	if len(members) == 0 {
		return nil
	}

	// 序列化所有成员
	serializedMembers := make([]interface{}, len(members))
	for i, member := range members {
		switch v := member.(type) {
		case string:
			serializedMembers[i] = v
		case []byte:
			serializedMembers[i] = string(v)
		default:
			jsonData, err := json.Marshal(member)
			if err != nil {
				r.metrics.ErrorCount++
				return fmt.Errorf("marshal member failed: %w", err)
			}
			serializedMembers[i] = string(jsonData)
		}
	}

	err := r.client.GetClient().SAdd(r.ctx, key, serializedMembers...).Err()
	if err != nil {
		r.metrics.ErrorCount++
		return fmt.Errorf("sadd failed: %w", err)
	}

	return nil
}

// SMembers 获取集合所有成员
func (r *RedisCacheManager) SMembers(key string) ([]interface{}, error) {
	r.metrics.TotalOps++

	results, err := r.client.GetClient().SMembers(r.ctx, key).Result()
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("smembers failed: %w", err)
	}

	// 转换结果
	values := make([]interface{}, len(results))
	for i, result := range results {
		var value interface{}
		if err := json.Unmarshal([]byte(result), &value); err != nil {
			values[i] = result
		} else {
			values[i] = value
		}
	}

	if len(values) > 0 {
		r.metrics.HitCount++
	} else {
		r.metrics.MissCount++
	}
	r.updateHitRate()

	return values, nil
}

// SIsMember 检查是否为集合成员
func (r *RedisCacheManager) SIsMember(key string, member interface{}) bool {
	r.metrics.TotalOps++

	// 序列化成员
	var data string
	switch v := member.(type) {
	case string:
		data = v
	case []byte:
		data = string(v)
	default:
		jsonData, err := json.Marshal(member)
		if err != nil {
			r.metrics.ErrorCount++
			return false
		}
		data = string(jsonData)
	}

	result, err := r.client.GetClient().SIsMember(r.ctx, key, data).Result()
	if err != nil {
		r.metrics.ErrorCount++
		return false
	}

	return result
}

// SRem 删除集合成员
func (r *RedisCacheManager) SRem(key string, members ...interface{}) error {
	r.metrics.TotalOps++

	if len(members) == 0 {
		return nil
	}

	// 序列化所有成员
	serializedMembers := make([]interface{}, len(members))
	for i, member := range members {
		switch v := member.(type) {
		case string:
			serializedMembers[i] = v
		case []byte:
			serializedMembers[i] = string(v)
		default:
			jsonData, err := json.Marshal(member)
			if err != nil {
				r.metrics.ErrorCount++
				return fmt.Errorf("marshal member failed: %w", err)
			}
			serializedMembers[i] = string(jsonData)
		}
	}

	err := r.client.GetClient().SRem(r.ctx, key, serializedMembers...).Err()
	if err != nil {
		r.metrics.ErrorCount++
		return fmt.Errorf("srem failed: %w", err)
	}

	return nil
}

// ZSet操作实现

// ZAdd 添加有序集合成员
func (r *RedisCacheManager) ZAdd(key string, members ...redis.Z) error {
	r.metrics.TotalOps++

	if len(members) == 0 {
		return nil
	}

	// 序列化所有成员
	serializedMembers := make([]redis.Z, len(members))
	for i, member := range members {
		switch v := member.Member.(type) {
		case string:
			serializedMembers[i] = redis.Z{Score: member.Score, Member: v}
		case []byte:
			serializedMembers[i] = redis.Z{Score: member.Score, Member: string(v)}
		default:
			jsonData, err := json.Marshal(member.Member)
			if err != nil {
				r.metrics.ErrorCount++
				return fmt.Errorf("marshal member failed: %w", err)
			}
			serializedMembers[i] = redis.Z{Score: member.Score, Member: string(jsonData)}
		}
	}

	err := r.client.GetClient().ZAdd(r.ctx, key, serializedMembers...).Err()
	if err != nil {
		r.metrics.ErrorCount++
		return fmt.Errorf("zadd failed: %w", err)
	}

	return nil
}

// ZRange 获取有序集合范围内的成员
func (r *RedisCacheManager) ZRange(key string, start, stop int64) ([]interface{}, error) {
	r.metrics.TotalOps++

	results, err := r.client.GetClient().ZRange(r.ctx, key, start, stop).Result()
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("zrange failed: %w", err)
	}

	// 转换结果
	values := make([]interface{}, len(results))
	for i, result := range results {
		var value interface{}
		if err := json.Unmarshal([]byte(result), &value); err != nil {
			values[i] = result
		} else {
			values[i] = value
		}
	}

	if len(values) > 0 {
		r.metrics.HitCount++
	} else {
		r.metrics.MissCount++
	}
	r.updateHitRate()

	return values, nil
}

// ZRangeByScore 按分数范围获取有序集合成员
func (r *RedisCacheManager) ZRangeByScore(key string, min, max string) ([]interface{}, error) {
	r.metrics.TotalOps++

	results, err := r.client.GetClient().ZRangeByScore(r.ctx, key, &redis.ZRangeBy{
		Min: min,
		Max: max,
	}).Result()
	if err != nil {
		r.metrics.ErrorCount++
		return nil, fmt.Errorf("zrangebyscore failed: %w", err)
	}

	// 转换结果
	values := make([]interface{}, len(results))
	for i, result := range results {
		var value interface{}
		if err := json.Unmarshal([]byte(result), &value); err != nil {
			values[i] = result
		} else {
			values[i] = value
		}
	}

	if len(values) > 0 {
		r.metrics.HitCount++
	} else {
		r.metrics.MissCount++
	}
	r.updateHitRate()

	return values, nil
}

// ZRem 删除有序集合成员
func (r *RedisCacheManager) ZRem(key string, members ...interface{}) error {
	r.metrics.TotalOps++

	if len(members) == 0 {
		return nil
	}

	// 序列化所有成员
	serializedMembers := make([]interface{}, len(members))
	for i, member := range members {
		switch v := member.(type) {
		case string:
			serializedMembers[i] = v
		case []byte:
			serializedMembers[i] = string(v)
		default:
			jsonData, err := json.Marshal(member)
			if err != nil {
				r.metrics.ErrorCount++
				return fmt.Errorf("marshal member failed: %w", err)
			}
			serializedMembers[i] = string(jsonData)
		}
	}

	err := r.client.GetClient().ZRem(r.ctx, key, serializedMembers...).Err()
	if err != nil {
		r.metrics.ErrorCount++
		return fmt.Errorf("zrem failed: %w", err)
	}

	return nil
}

// ZScore 获取有序集合成员分数
func (r *RedisCacheManager) ZScore(key string, member string) (float64, error) {
	r.metrics.TotalOps++

	result, err := r.client.GetClient().ZScore(r.ctx, key, member).Result()
	if err != nil {
		if err == redis.Nil {
			r.metrics.MissCount++
			return 0, nil
		}
		r.metrics.ErrorCount++
		return 0, fmt.Errorf("zscore failed: %w", err)
	}

	r.metrics.HitCount++
	r.updateHitRate()

	return result, nil
}
