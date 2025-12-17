package cart

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"mall-go/internal/model"

	"github.com/redis/go-redis/v9"
)

// CacheService 购物车缓存服务
type CacheService struct {
	rdb         *redis.Client
	cartService *CartService
	ctx         context.Context
}

// NewCacheService 创建购物车缓存服务
func NewCacheService(rdb *redis.Client, cartService *CartService) *CacheService {
	return &CacheService{
		rdb:         rdb,
		cartService: cartService,
		ctx:         context.Background(),
	}
}

// 缓存键前缀
const (
	CartCachePrefix     = "cart:"
	CartItemCachePrefix = "cart_items:"
	CartCountPrefix     = "cart_count:"
	CartLockPrefix      = "cart_lock:"
)

// 缓存过期时间
const (
	CartCacheExpire = 24 * time.Hour     // 购物车缓存24小时
	CartCountExpire = 1 * time.Hour      // 购物车数量缓存1小时
	CartLockExpire  = 30 * time.Second   // 购物车锁30秒
	GuestCartExpire = 7 * 24 * time.Hour // 游客购物车7天
)

// CartCacheData 购物车缓存数据
type CartCacheData struct {
	Cart      *model.Cart        `json:"cart"`
	Items     []model.CartItem   `json:"items"`
	Summary   *model.CartSummary `json:"summary"`
	Version   int64              `json:"version"`
	UpdatedAt time.Time          `json:"updated_at"`
}

// AddToCartWithCache 添加商品到购物车（带缓存）
func (cs *CacheService) AddToCartWithCache(userID uint, sessionID string, req *model.AddToCartRequest) (*model.CartItem, error) {
	// 获取分布式锁
	lockKey := cs.getCartLockKey(userID, sessionID)
	lock, err := cs.acquireLock(lockKey, CartLockExpire)
	if err != nil {
		return nil, fmt.Errorf("获取购物车锁失败: %v", err)
	}
	defer cs.releaseLock(lockKey, lock)

	// 调用核心服务添加商品
	cartItem, err := cs.cartService.AddToCart(userID, sessionID, req)
	if err != nil {
		return nil, err
	}

	// 清除相关缓存
	cs.clearCartCache(userID, sessionID)

	return cartItem, nil
}

// GetCartWithCache 获取购物车（带缓存）
func (cs *CacheService) GetCartWithCache(userID uint, sessionID string, includeInvalid bool) (*model.CartResponse, error) {
	cacheKey := cs.getCartCacheKey(userID, sessionID)

	// 尝试从缓存获取
	cachedData, err := cs.getCartFromCache(cacheKey)
	if err == nil && cachedData != nil {
		// 缓存命中，检查是否需要包含失效商品
		if includeInvalid || len(cachedData.Summary.InvalidItems) == 0 {
			return &model.CartResponse{
				Cart:    cachedData.Cart,
				Summary: cachedData.Summary,
			}, nil
		}
	}

	// 缓存未命中或需要包含失效商品，从数据库获取
	cartResponse, err := cs.cartService.GetCart(userID, sessionID, includeInvalid)
	if err != nil {
		return nil, err
	}

	// 异步更新缓存
	go cs.updateCartCache(cacheKey, cartResponse, userID == 0)

	return cartResponse, nil
}

// UpdateCartItemWithCache 更新购物车商品（带缓存）
func (cs *CacheService) UpdateCartItemWithCache(userID uint, sessionID string, itemID uint, req *model.UpdateCartItemRequest) (*model.CartItem, error) {
	// 获取分布式锁
	lockKey := cs.getCartLockKey(userID, sessionID)
	lock, err := cs.acquireLock(lockKey, CartLockExpire)
	if err != nil {
		return nil, fmt.Errorf("获取购物车锁失败: %v", err)
	}
	defer cs.releaseLock(lockKey, lock)

	// 调用核心服务更新商品
	cartItem, err := cs.cartService.UpdateCartItem(userID, sessionID, itemID, req)
	if err != nil {
		return nil, err
	}

	// 清除相关缓存
	cs.clearCartCache(userID, sessionID)

	return cartItem, nil
}

// RemoveFromCartWithCache 从购物车移除商品（带缓存）
func (cs *CacheService) RemoveFromCartWithCache(userID uint, sessionID string, itemID uint) error {
	// 获取分布式锁
	lockKey := cs.getCartLockKey(userID, sessionID)
	lock, err := cs.acquireLock(lockKey, CartLockExpire)
	if err != nil {
		return fmt.Errorf("获取购物车锁失败: %v", err)
	}
	defer cs.releaseLock(lockKey, lock)

	// 调用核心服务移除商品
	err = cs.cartService.RemoveFromCart(userID, sessionID, itemID)
	if err != nil {
		return err
	}

	// 清除相关缓存
	cs.clearCartCache(userID, sessionID)

	return nil
}

// ClearCartWithCache 清空购物车（带缓存）
func (cs *CacheService) ClearCartWithCache(userID uint, sessionID string) error {
	// 获取分布式锁
	lockKey := cs.getCartLockKey(userID, sessionID)
	lock, err := cs.acquireLock(lockKey, CartLockExpire)
	if err != nil {
		return fmt.Errorf("获取购物车锁失败: %v", err)
	}
	defer cs.releaseLock(lockKey, lock)

	// 调用核心服务清空购物车
	err = cs.cartService.ClearCart(userID, sessionID)
	if err != nil {
		return err
	}

	// 清除相关缓存
	cs.clearCartCache(userID, sessionID)

	return nil
}

// GetCartItemCountWithCache 获取购物车商品数量（带缓存）
func (cs *CacheService) GetCartItemCountWithCache(userID uint, sessionID string) (int, error) {
	countKey := cs.getCartCountKey(userID, sessionID)

	// 尝试从缓存获取
	countStr, err := cs.rdb.Get(cs.ctx, countKey).Result()
	if err == nil {
		count, err := strconv.Atoi(countStr)
		if err == nil {
			return count, nil
		}
	}

	// 缓存未命中，从数据库获取
	count, err := cs.cartService.GetCartItemCount(userID, sessionID)
	if err != nil {
		return 0, err
	}

	// 异步更新缓存
	go cs.rdb.Set(cs.ctx, countKey, count, CartCountExpire)

	return count, nil
}

// BatchUpdateCartWithCache 批量更新购物车（带缓存）
func (cs *CacheService) BatchUpdateCartWithCache(userID uint, sessionID string, req *model.BatchUpdateCartRequest) error {
	// 获取分布式锁
	lockKey := cs.getCartLockKey(userID, sessionID)
	lock, err := cs.acquireLock(lockKey, CartLockExpire)
	if err != nil {
		return fmt.Errorf("获取购物车锁失败: %v", err)
	}
	defer cs.releaseLock(lockKey, lock)

	// 调用核心服务批量更新
	err = cs.cartService.BatchUpdateCart(userID, sessionID, req)
	if err != nil {
		return err
	}

	// 清除相关缓存
	cs.clearCartCache(userID, sessionID)

	return nil
}

// getCartCacheKey 获取购物车缓存键
func (cs *CacheService) getCartCacheKey(userID uint, sessionID string) string {
	if userID > 0 {
		return fmt.Sprintf("%suser_%d", CartCachePrefix, userID)
	}
	return fmt.Sprintf("%sguest_%s", CartCachePrefix, sessionID)
}

// getCartCountKey 获取购物车数量缓存键
func (cs *CacheService) getCartCountKey(userID uint, sessionID string) string {
	if userID > 0 {
		return fmt.Sprintf("%suser_%d", CartCountPrefix, userID)
	}
	return fmt.Sprintf("%sguest_%s", CartCountPrefix, sessionID)
}

// getCartLockKey 获取购物车锁键
func (cs *CacheService) getCartLockKey(userID uint, sessionID string) string {
	if userID > 0 {
		return fmt.Sprintf("%suser_%d", CartLockPrefix, userID)
	}
	return fmt.Sprintf("%sguest_%s", CartLockPrefix, sessionID)
}

// getCartFromCache 从缓存获取购物车
func (cs *CacheService) getCartFromCache(cacheKey string) (*CartCacheData, error) {
	data, err := cs.rdb.Get(cs.ctx, cacheKey).Result()
	if err != nil {
		return nil, err
	}

	var cacheData CartCacheData
	if err := json.Unmarshal([]byte(data), &cacheData); err != nil {
		return nil, err
	}

	return &cacheData, nil
}

// updateCartCache 更新购物车缓存
func (cs *CacheService) updateCartCache(cacheKey string, cartResponse *model.CartResponse, isGuest bool) {
	cacheData := &CartCacheData{
		Cart:      cartResponse.Cart,
		Items:     cartResponse.Cart.Items,
		Summary:   cartResponse.Summary,
		Version:   time.Now().Unix(),
		UpdatedAt: time.Now(),
	}

	data, err := json.Marshal(cacheData)
	if err != nil {
		return
	}

	expire := CartCacheExpire
	if isGuest {
		expire = GuestCartExpire
	}

	cs.rdb.Set(cs.ctx, cacheKey, data, expire)
}

// clearCartCache 清除购物车相关缓存
func (cs *CacheService) clearCartCache(userID uint, sessionID string) {
	cartKey := cs.getCartCacheKey(userID, sessionID)
	countKey := cs.getCartCountKey(userID, sessionID)

	cs.rdb.Del(cs.ctx, cartKey, countKey)
}

// acquireLock 获取分布式锁
func (cs *CacheService) acquireLock(lockKey string, expire time.Duration) (string, error) {
	lockValue := fmt.Sprintf("%d", time.Now().UnixNano())

	success, err := cs.rdb.SetNX(cs.ctx, lockKey, lockValue, expire).Result()
	if err != nil {
		return "", err
	}

	if !success {
		return "", fmt.Errorf("获取锁失败，锁已被占用")
	}

	return lockValue, nil
}

// releaseLock 释放分布式锁
func (cs *CacheService) releaseLock(lockKey, lockValue string) {
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	cs.rdb.Eval(cs.ctx, script, []string{lockKey}, lockValue)
}

// WarmupCartCache 预热购物车缓存
func (cs *CacheService) WarmupCartCache(userID uint, sessionID string) error {
	cartResponse, err := cs.cartService.GetCart(userID, sessionID, false)
	if err != nil {
		return err
	}

	cacheKey := cs.getCartCacheKey(userID, sessionID)
	cs.updateCartCache(cacheKey, cartResponse, userID == 0)

	return nil
}

// RefreshCartCache 刷新购物车缓存
func (cs *CacheService) RefreshCartCache(userID uint, sessionID string) error {
	// 清除现有缓存
	cs.clearCartCache(userID, sessionID)

	// 重新加载缓存
	return cs.WarmupCartCache(userID, sessionID)
}

// GetCacheStats 获取缓存统计信息
func (cs *CacheService) GetCacheStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 获取购物车缓存数量
	cartKeys, err := cs.rdb.Keys(cs.ctx, CartCachePrefix+"*").Result()
	if err == nil {
		stats["cart_cache_count"] = len(cartKeys)
	}

	// 获取购物车数量缓存数量
	countKeys, err := cs.rdb.Keys(cs.ctx, CartCountPrefix+"*").Result()
	if err == nil {
		stats["count_cache_count"] = len(countKeys)
	}

	// 获取活跃锁数量
	lockKeys, err := cs.rdb.Keys(cs.ctx, CartLockPrefix+"*").Result()
	if err == nil {
		stats["active_locks"] = len(lockKeys)
	}

	return stats, nil
}

// CleanExpiredCache 清理过期缓存
func (cs *CacheService) CleanExpiredCache() error {
	// Redis会自动清理过期键，这里可以添加额外的清理逻辑
	// 比如清理长时间未使用的游客购物车缓存

	// 获取所有游客购物车缓存键
	guestKeys, err := cs.rdb.Keys(cs.ctx, CartCachePrefix+"guest_*").Result()
	if err != nil {
		return err
	}

	// 检查每个键的最后访问时间，清理超过7天未访问的缓存
	for _, key := range guestKeys {
		ttl, err := cs.rdb.TTL(cs.ctx, key).Result()
		if err != nil {
			continue
		}

		// 如果TTL小于1天，说明缓存即将过期，可以提前清理
		if ttl < 24*time.Hour {
			cs.rdb.Del(cs.ctx, key)
		}
	}

	return nil
}

// 全局购物车缓存服务实例
var globalCacheService *CacheService

// InitGlobalCacheService 初始化全局购物车缓存服务
func InitGlobalCacheService(rdb *redis.Client, cartService *CartService) {
	globalCacheService = NewCacheService(rdb, cartService)
}

// GetGlobalCacheService 获取全局购物车缓存服务
func GetGlobalCacheService() *CacheService {
	return globalCacheService
}
