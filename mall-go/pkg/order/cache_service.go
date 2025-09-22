package order

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"mall-go/internal/model"

	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// CacheService 订单缓存服务
type CacheService struct {
	rdb          *redis.Client
	orderService *OrderService
	ctx          context.Context
}

// NewCacheService 创建订单缓存服务
func NewCacheService(rdb *redis.Client, orderService *OrderService) *CacheService {
	return &CacheService{
		rdb:          rdb,
		orderService: orderService,
		ctx:          context.Background(),
	}
}

// 缓存键前缀
const (
	OrderCachePrefix      = "order:"
	OrderListCachePrefix  = "order_list:"
	OrderStatsCachePrefix = "order_stats:"
	OrderLockPrefix       = "order_lock:"
	UserOrderCachePrefix  = "user_orders:"
)

// 缓存过期时间
const (
	OrderCacheExpire      = 1 * time.Hour    // 订单缓存1小时
	OrderListCacheExpire  = 10 * time.Minute // 订单列表缓存10分钟
	OrderStatsCacheExpire = 5 * time.Minute  // 订单统计缓存5分钟
	OrderLockExpire       = 30 * time.Second // 订单锁30秒
)

// OrderCacheData 订单缓存数据
type OrderCacheData struct {
	Order     *model.Order `json:"order"`
	Version   int64        `json:"version"`
	UpdatedAt time.Time    `json:"updated_at"`
}

// GetOrderWithCache 获取订单（带缓存）
func (cs *CacheService) GetOrderWithCache(orderID uint) (*model.Order, error) {
	cacheKey := cs.getOrderCacheKey(orderID)

	// 尝试从缓存获取
	cachedData, err := cs.getOrderFromCache(cacheKey)
	if err == nil && cachedData != nil {
		return cachedData.Order, nil
	}

	// 缓存未命中，从数据库获取
	order, err := cs.getOrderFromDB(orderID)
	if err != nil {
		return nil, err
	}

	// 异步更新缓存
	go cs.updateOrderCache(cacheKey, order)

	return order, nil
}

// GetOrderByNoWithCache 根据订单号获取订单（带缓存）
func (cs *CacheService) GetOrderByNoWithCache(orderNo string) (*model.Order, error) {
	cacheKey := fmt.Sprintf("%sno_%s", OrderCachePrefix, orderNo)

	// 尝试从缓存获取
	cachedData, err := cs.getOrderFromCache(cacheKey)
	if err == nil && cachedData != nil {
		return cachedData.Order, nil
	}

	// 缓存未命中，从数据库获取
	order, err := cs.getOrderByNoFromDB(orderNo)
	if err != nil {
		return nil, err
	}

	// 异步更新缓存
	go cs.updateOrderCache(cacheKey, order)

	return order, nil
}

// GetUserOrdersWithCache 获取用户订单列表（带缓存）
func (cs *CacheService) GetUserOrdersWithCache(userID uint, status string, page, pageSize int) ([]*model.Order, int64, error) {
	// 如果Redis不可用，直接从数据库获取
	if cs.rdb == nil {
		return cs.getUserOrdersFromDB(userID, status, page, pageSize)
	}

	cacheKey := cs.getUserOrdersCacheKey(userID, status, page, pageSize)

	// 尝试从缓存获取
	cachedData, err := cs.getUserOrdersFromCache(cacheKey)
	if err == nil && cachedData != nil {
		return cachedData.Orders, cachedData.Total, nil
	}

	// 缓存未命中，从数据库获取
	orders, total, err := cs.getUserOrdersFromDB(userID, status, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// 异步更新缓存（仅在Redis可用时）
	if cs.rdb != nil {
		go cs.updateUserOrdersCache(cacheKey, orders, total)
	}

	return orders, total, nil
}

// GetOrderStatisticsWithCache 获取订单统计（带缓存）
func (cs *CacheService) GetOrderStatisticsWithCache() (*model.OrderStatisticsResponse, error) {
	cacheKey := fmt.Sprintf("%sall", OrderStatsCachePrefix)

	// 尝试从缓存获取
	cachedStats, err := cs.getOrderStatsFromCache(cacheKey)
	if err == nil && cachedStats != nil {
		return cachedStats, nil
	}

	// 缓存未命中，从数据库获取
	stats, err := cs.getOrderStatsFromDB()
	if err != nil {
		return nil, err
	}

	// 异步更新缓存
	go cs.updateOrderStatsCache(cacheKey, stats)

	return stats, nil
}

// InvalidateOrderCache 使订单缓存失效
func (cs *CacheService) InvalidateOrderCache(orderID uint) {
	cacheKey := cs.getOrderCacheKey(orderID)
	cs.rdb.Del(cs.ctx, cacheKey)

	// 同时清除相关的列表缓存
	cs.clearRelatedListCache(orderID)
}

// InvalidateOrderCacheByNo 根据订单号使缓存失效
func (cs *CacheService) InvalidateOrderCacheByNo(orderNo string) {
	cacheKey := fmt.Sprintf("%sno_%s", OrderCachePrefix, orderNo)
	cs.rdb.Del(cs.ctx, cacheKey)
}

// InvalidateUserOrdersCache 使用户订单列表缓存失效
func (cs *CacheService) InvalidateUserOrdersCache(userID uint) {
	pattern := fmt.Sprintf("%suser_%d_*", UserOrderCachePrefix, userID)
	keys, err := cs.rdb.Keys(cs.ctx, pattern).Result()
	if err == nil && len(keys) > 0 {
		cs.rdb.Del(cs.ctx, keys...)
	}
}

// InvalidateOrderStatsCache 使订单统计缓存失效
func (cs *CacheService) InvalidateOrderStatsCache() {
	pattern := fmt.Sprintf("%s*", OrderStatsCachePrefix)
	keys, err := cs.rdb.Keys(cs.ctx, pattern).Result()
	if err == nil && len(keys) > 0 {
		cs.rdb.Del(cs.ctx, keys...)
	}
}

// AcquireOrderLock 获取订单锁
func (cs *CacheService) AcquireOrderLock(orderID uint) (string, error) {
	lockKey := cs.getOrderLockKey(orderID)
	lockValue := fmt.Sprintf("%d", time.Now().UnixNano())

	success, err := cs.rdb.SetNX(cs.ctx, lockKey, lockValue, OrderLockExpire).Result()
	if err != nil {
		return "", err
	}

	if !success {
		return "", fmt.Errorf("获取订单锁失败，订单正在处理中")
	}

	return lockValue, nil
}

// ReleaseOrderLock 释放订单锁
func (cs *CacheService) ReleaseOrderLock(orderID uint, lockValue string) {
	lockKey := cs.getOrderLockKey(orderID)
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	cs.rdb.Eval(cs.ctx, script, []string{lockKey}, lockValue)
}

// WarmupOrderCache 预热订单缓存
func (cs *CacheService) WarmupOrderCache(orderIDs []uint) error {
	for _, orderID := range orderIDs {
		order, err := cs.getOrderFromDB(orderID)
		if err != nil {
			continue
		}

		cacheKey := cs.getOrderCacheKey(orderID)
		cs.updateOrderCache(cacheKey, order)
	}

	return nil
}

// getOrderCacheKey 获取订单缓存键
func (cs *CacheService) getOrderCacheKey(orderID uint) string {
	return fmt.Sprintf("%sid_%d", OrderCachePrefix, orderID)
}

// getUserOrdersCacheKey 获取用户订单列表缓存键
func (cs *CacheService) getUserOrdersCacheKey(userID uint, status string, page, pageSize int) string {
	return fmt.Sprintf("%suser_%d_%s_%d_%d", UserOrderCachePrefix, userID, status, page, pageSize)
}

// getOrderLockKey 获取订单锁键
func (cs *CacheService) getOrderLockKey(orderID uint) string {
	return fmt.Sprintf("%s%d", OrderLockPrefix, orderID)
}

// getOrderFromCache 从缓存获取订单
func (cs *CacheService) getOrderFromCache(cacheKey string) (*OrderCacheData, error) {
	// 如果Redis不可用，返回缓存未命中
	if cs.rdb == nil {
		return nil, fmt.Errorf("cache not available")
	}

	data, err := cs.rdb.Get(cs.ctx, cacheKey).Result()
	if err != nil {
		return nil, err
	}

	var cacheData OrderCacheData
	if err := json.Unmarshal([]byte(data), &cacheData); err != nil {
		return nil, err
	}

	return &cacheData, nil
}

// updateOrderCache 更新订单缓存
func (cs *CacheService) updateOrderCache(cacheKey string, order *model.Order) {
	cacheData := &OrderCacheData{
		Order:     order,
		Version:   time.Now().Unix(),
		UpdatedAt: time.Now(),
	}

	data, err := json.Marshal(cacheData)
	if err != nil {
		return
	}

	cs.rdb.Set(cs.ctx, cacheKey, data, OrderCacheExpire)
}

// getUserOrdersFromCache 从缓存获取用户订单列表
func (cs *CacheService) getUserOrdersFromCache(cacheKey string) (*UserOrdersCacheData, error) {
	// 如果Redis不可用，返回缓存未命中
	if cs.rdb == nil {
		return nil, fmt.Errorf("cache not available")
	}

	data, err := cs.rdb.Get(cs.ctx, cacheKey).Result()
	if err != nil {
		return nil, err
	}

	var cacheData UserOrdersCacheData
	if err := json.Unmarshal([]byte(data), &cacheData); err != nil {
		return nil, err
	}

	return &cacheData, nil
}

// updateUserOrdersCache 更新用户订单列表缓存
func (cs *CacheService) updateUserOrdersCache(cacheKey string, orders []*model.Order, total int64) {
	cacheData := &UserOrdersCacheData{
		Orders:    orders,
		Total:     total,
		UpdatedAt: time.Now(),
	}

	data, err := json.Marshal(cacheData)
	if err != nil {
		return
	}

	cs.rdb.Set(cs.ctx, cacheKey, data, OrderListCacheExpire)
}

// getOrderStatsFromCache 从缓存获取订单统计
func (cs *CacheService) getOrderStatsFromCache(cacheKey string) (*model.OrderStatisticsResponse, error) {
	data, err := cs.rdb.Get(cs.ctx, cacheKey).Result()
	if err != nil {
		return nil, err
	}

	var stats model.OrderStatisticsResponse
	if err := json.Unmarshal([]byte(data), &stats); err != nil {
		return nil, err
	}

	return &stats, nil
}

// updateOrderStatsCache 更新订单统计缓存
func (cs *CacheService) updateOrderStatsCache(cacheKey string, stats *model.OrderStatisticsResponse) {
	data, err := json.Marshal(stats)
	if err != nil {
		return
	}

	cs.rdb.Set(cs.ctx, cacheKey, data, OrderStatsCacheExpire)
}

// getOrderFromDB 从数据库获取订单
func (cs *CacheService) getOrderFromDB(orderID uint) (*model.Order, error) {
	var order model.Order
	err := cs.orderService.db.Preload("OrderItems.Product").
		Preload("OrderItems.SKU").
		First(&order, orderID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("订单不存在")
		}
		return nil, fmt.Errorf("查询订单失败: %v", err)
	}
	return &order, nil
}

// getOrderByNoFromDB 从数据库根据订单号获取订单
func (cs *CacheService) getOrderByNoFromDB(orderNo string) (*model.Order, error) {
	var order model.Order
	err := cs.orderService.db.Preload("OrderItems.Product").
		Preload("OrderItems.SKU").
		Where("order_no = ?", orderNo).
		First(&order).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("订单不存在")
		}
		return nil, fmt.Errorf("查询订单失败: %v", err)
	}
	return &order, nil
}

// getUserOrdersFromDB 从数据库获取用户订单列表
func (cs *CacheService) getUserOrdersFromDB(userID uint, status string, page, pageSize int) ([]*model.Order, int64, error) {
	var orders []*model.Order
	var total int64

	// 构建查询条件
	query := cs.orderService.db.Model(&model.Order{}).Where("user_id = ?", userID)

	// 如果指定了状态，添加状态过滤
	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询订单总数失败: %v", err)
	}

	// 分页查询订单列表
	offset := (page - 1) * pageSize
	err := query.Preload("OrderItems.Product").
		Preload("OrderItems.SKU").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&orders).Error

	if err != nil {
		return nil, 0, fmt.Errorf("查询订单列表失败: %v", err)
	}

	return orders, total, nil
}

// getOrderStatsFromDB 从数据库获取订单统计
func (cs *CacheService) getOrderStatsFromDB() (*model.OrderStatisticsResponse, error) {
	stats := &model.OrderStatisticsResponse{}

	// 查询总订单数
	if err := cs.orderService.db.Model(&model.Order{}).Count(&stats.TotalOrders).Error; err != nil {
		return nil, fmt.Errorf("查询总订单数失败: %v", err)
	}

	// 查询各状态订单数
	cs.orderService.db.Model(&model.Order{}).Where("status = ?", model.OrderStatusPending).Count(&stats.PendingOrders)
	cs.orderService.db.Model(&model.Order{}).Where("status = ?", model.OrderStatusPaid).Count(&stats.PaidOrders)
	cs.orderService.db.Model(&model.Order{}).Where("status = ?", model.OrderStatusShipped).Count(&stats.ShippedOrders)
	cs.orderService.db.Model(&model.Order{}).Where("status = ?", model.OrderStatusCompleted).Count(&stats.CompletedOrders)

	// 查询总金额
	var totalAmount decimal.Decimal
	cs.orderService.db.Model(&model.Order{}).Select("COALESCE(SUM(total_amount), 0)").Scan(&totalAmount)
	stats.TotalAmount = totalAmount

	// 查询今日订单数和金额
	today := time.Now().Format("2006-01-02")
	cs.orderService.db.Model(&model.Order{}).Where("DATE(created_at) = ?", today).Count(&stats.TodayOrders)

	var todayAmount decimal.Decimal
	cs.orderService.db.Model(&model.Order{}).Where("DATE(created_at) = ?", today).Select("COALESCE(SUM(total_amount), 0)").Scan(&todayAmount)
	stats.TodayAmount = todayAmount

	return stats, nil
}

// clearRelatedListCache 清除相关的列表缓存
func (cs *CacheService) clearRelatedListCache(orderID uint) {
	// 清除用户订单列表缓存
	// 这里需要知道订单属于哪个用户，实际实现中需要查询数据库

	// 清除订单统计缓存
	cs.InvalidateOrderStatsCache()
}

// UserOrdersCacheData 用户订单列表缓存数据
type UserOrdersCacheData struct {
	Orders    []*model.Order `json:"orders"`
	Total     int64          `json:"total"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// GetCacheStats 获取缓存统计信息
func (cs *CacheService) GetCacheStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 获取订单缓存数量
	orderKeys, err := cs.rdb.Keys(cs.ctx, OrderCachePrefix+"*").Result()
	if err == nil {
		stats["order_cache_count"] = len(orderKeys)
	}

	// 获取订单列表缓存数量
	listKeys, err := cs.rdb.Keys(cs.ctx, OrderListCachePrefix+"*").Result()
	if err == nil {
		stats["order_list_cache_count"] = len(listKeys)
	}

	// 获取用户订单缓存数量
	userOrderKeys, err := cs.rdb.Keys(cs.ctx, UserOrderCachePrefix+"*").Result()
	if err == nil {
		stats["user_order_cache_count"] = len(userOrderKeys)
	}

	// 获取活跃锁数量
	lockKeys, err := cs.rdb.Keys(cs.ctx, OrderLockPrefix+"*").Result()
	if err == nil {
		stats["active_locks"] = len(lockKeys)
	}

	return stats, nil
}

// CleanExpiredCache 清理过期缓存
func (cs *CacheService) CleanExpiredCache() error {
	// Redis会自动清理过期键，这里可以添加额外的清理逻辑

	// 清理长时间未访问的缓存
	patterns := []string{
		OrderCachePrefix + "*",
		OrderListCachePrefix + "*",
		UserOrderCachePrefix + "*",
	}

	for _, pattern := range patterns {
		keys, err := cs.rdb.Keys(cs.ctx, pattern).Result()
		if err != nil {
			continue
		}

		for _, key := range keys {
			ttl, err := cs.rdb.TTL(cs.ctx, key).Result()
			if err != nil {
				continue
			}

			// 如果TTL小于1分钟，提前清理
			if ttl < time.Minute {
				cs.rdb.Del(cs.ctx, key)
			}
		}
	}

	return nil
}

// RefreshOrderCache 刷新订单缓存
func (cs *CacheService) RefreshOrderCache(orderID uint) error {
	// 清除现有缓存
	cs.InvalidateOrderCache(orderID)

	// 重新加载缓存
	_, err := cs.GetOrderWithCache(orderID)
	return err
}

// BatchInvalidateOrderCache 批量使订单缓存失效
func (cs *CacheService) BatchInvalidateOrderCache(orderIDs []uint) error {
	if len(orderIDs) == 0 {
		return nil
	}

	var keys []string
	for _, orderID := range orderIDs {
		keys = append(keys, cs.getOrderCacheKey(orderID))
	}

	return cs.rdb.Del(cs.ctx, keys...).Err()
}

// 全局订单缓存服务实例
var globalCacheService *CacheService

// InitGlobalCacheService 初始化全局订单缓存服务
func InitGlobalCacheService(rdb *redis.Client, orderService *OrderService) {
	globalCacheService = NewCacheService(rdb, orderService)
}

// GetGlobalCacheService 获取全局订单缓存服务
func GetGlobalCacheService() *CacheService {
	return globalCacheService
}
