package security

import (
	"context"
	"fmt"
	"time"

	"mall-go/pkg/logger"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// JWTBlacklist JWT黑名单管理器
type JWTBlacklist struct {
	rdb *redis.Client
}

// NewJWTBlacklist 创建JWT黑名单管理器
func NewJWTBlacklist(rdb *redis.Client) *JWTBlacklist {
	return &JWTBlacklist{
		rdb: rdb,
	}
}

// AddToBlacklist 将token添加到黑名单
func (jb *JWTBlacklist) AddToBlacklist(ctx context.Context, token string, expiry time.Duration) error {
	if jb.rdb == nil {
		logger.Warn("Redis客户端未初始化，跳过黑名单操作")
		return nil
	}

	key := fmt.Sprintf("jwt:blacklist:%s", token)
	err := jb.rdb.Set(ctx, key, "1", expiry).Err()
	if err != nil {
		logger.Error("添加JWT黑名单失败", zap.Error(err), zap.String("token", token[:20]+"..."))
		return err
	}

	logger.Debug("JWT已添加到黑名单", zap.String("token", token[:20]+"..."))
	return nil
}

// IsBlacklisted 检查token是否在黑名单中
func (jb *JWTBlacklist) IsBlacklisted(ctx context.Context, token string) bool {
	if jb.rdb == nil {
		return false
	}

	key := fmt.Sprintf("jwt:blacklist:%s", token)
	val, err := jb.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return false
		}
		logger.Error("检查JWT黑名单失败", zap.Error(err))
		return false
	}

	return val == "1"
}

// RemoveFromBlacklist 从黑名单中移除token
func (jb *JWTBlacklist) RemoveFromBlacklist(ctx context.Context, token string) error {
	if jb.rdb == nil {
		return nil
	}

	key := fmt.Sprintf("jwt:blacklist:%s", token)
	err := jb.rdb.Del(ctx, key).Err()
	if err != nil {
		logger.Error("从黑名单移除JWT失败", zap.Error(err))
		return err
	}

	return nil
}

// 全局JWT黑名单实例
var GlobalJWTBlacklist *JWTBlacklist

// InitJWTBlacklist 初始化全局JWT黑名单
func InitJWTBlacklist(rdb *redis.Client) {
	GlobalJWTBlacklist = NewJWTBlacklist(rdb)
}

// GetJWTBlacklist 获取全局JWT黑名单
func GetJWTBlacklist() *JWTBlacklist {
	if GlobalJWTBlacklist == nil {
		GlobalJWTBlacklist = NewJWTBlacklist(nil)
	}
	return GlobalJWTBlacklist
}

