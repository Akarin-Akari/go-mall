package cache

import (
	"encoding/json"
	"fmt"
	"mall-go/internal/model"
	"mall-go/pkg/logger"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
)

// PriceCacheService 价格缓存服务
type PriceCacheService struct {
	cacheManager CacheManager
	keyManager   *CacheKeyManager
}

// NewPriceCacheService 创建价格缓存服务
func NewPriceCacheService(cacheManager CacheManager, keyManager *CacheKeyManager) *PriceCacheService {
	return &PriceCacheService{
		cacheManager: cacheManager,
		keyManager:   keyManager,
	}
}

// PriceCacheData 价格缓存数据结构
type PriceCacheData struct {
	ProductID uint `json:"product_id"`

	// 基础价格信息
	Price       string `json:"price"`        // 当前价格（使用字符串存储decimal）
	OriginPrice string `json:"origin_price"` // 原价
	CostPrice   string `json:"cost_price"`   // 成本价

	// 促销价格信息
	PromotionPrice     string    `json:"promotion_price,omitempty"`      // 促销价格
	PromotionStartTime time.Time `json:"promotion_start_time,omitempty"` // 促销开始时间
	PromotionEndTime   time.Time `json:"promotion_end_time,omitempty"`   // 促销结束时间
	PromotionType      string    `json:"promotion_type,omitempty"`       // 促销类型：discount, fixed, percentage
	PromotionValue     string    `json:"promotion_value,omitempty"`      // 促销值

	// 会员价格信息
	VipPrice    string `json:"vip_price,omitempty"`    // VIP价格
	VipDiscount string `json:"vip_discount,omitempty"` // VIP折扣率

	// 价格状态
	PriceStatus string `json:"price_status"` // normal, promotion, vip_only
	IsPromotion bool   `json:"is_promotion"` // 是否促销中
	IsVipPrice  bool   `json:"is_vip_price"` // 是否有VIP价格

	// 价格变更信息
	LastPriceChange  time.Time `json:"last_price_change"`  // 最后价格变更时间
	PriceChangeCount int       `json:"price_change_count"` // 价格变更次数

	// 缓存元数据
	CachedAt  time.Time `json:"cached_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Version   int       `json:"version"` // 乐观锁版本号
}

// PriceUpdateRequest 价格更新请求
type PriceUpdateRequest struct {
	ProductID   uint            `json:"product_id"`
	Price       decimal.Decimal `json:"price"`
	OriginPrice decimal.Decimal `json:"origin_price,omitempty"`
	CostPrice   decimal.Decimal `json:"cost_price,omitempty"`
	Reason      string          `json:"reason"` // 更新原因：manual, promotion, batch等
}

// PromotionPriceRequest 促销价格请求
type PromotionPriceRequest struct {
	ProductID      uint            `json:"product_id"`
	PromotionPrice decimal.Decimal `json:"promotion_price"`
	StartTime      time.Time       `json:"start_time"`
	EndTime        time.Time       `json:"end_time"`
	PromotionType  string          `json:"promotion_type"` // discount, fixed, percentage
	PromotionValue decimal.Decimal `json:"promotion_value,omitempty"`
}

// PriceHistoryRecord 价格历史记录
type PriceHistoryRecord struct {
	ProductID  uint            `json:"product_id"`
	OldPrice   decimal.Decimal `json:"old_price"`
	NewPrice   decimal.Decimal `json:"new_price"`
	ChangeType string          `json:"change_type"` // increase, decrease, promotion_start, promotion_end
	Reason     string          `json:"reason"`
	ChangeTime time.Time       `json:"change_time"`
	OperatorID uint            `json:"operator_id,omitempty"`
}

// ConvertToPriceCacheData 将Product模型转换为价格缓存数据
func ConvertToPriceCacheData(product *model.Product) *PriceCacheData {
	priceStatus := "normal"
	isPromotion := false

	// 这里可以根据实际业务逻辑判断促销状态
	// 暂时简化处理

	return &PriceCacheData{
		ProductID:   product.ID,
		Price:       product.Price.String(),
		OriginPrice: product.OriginPrice.String(),
		CostPrice:   product.CostPrice.String(),

		PriceStatus:      priceStatus,
		IsPromotion:      isPromotion,
		IsVipPrice:       false,
		LastPriceChange:  product.UpdatedAt,
		PriceChangeCount: 0, // 这个需要从实际业务数据获取

		CachedAt:  time.Now(),
		UpdatedAt: product.UpdatedAt,
		Version:   product.Version,
	}
}

// GetPrice 获取价格缓存
func (pcs *PriceCacheService) GetPrice(productID uint) (*PriceCacheData, error) {
	key := pcs.keyManager.GenerateProductPriceKey(productID)

	// 从缓存获取
	data, err := pcs.cacheManager.Get(key)
	if err != nil {
		logger.Error(fmt.Sprintf("获取价格缓存失败: %v", err))
		return nil, fmt.Errorf("获取价格缓存失败: %w", err)
	}

	if data == nil {
		return nil, nil // 缓存未命中
	}

	// 反序列化
	var price PriceCacheData
	jsonStr, ok := data.(string)
	if !ok {
		return nil, fmt.Errorf("价格缓存数据格式错误")
	}

	if err := json.Unmarshal([]byte(jsonStr), &price); err != nil {
		logger.Error(fmt.Sprintf("价格缓存数据反序列化失败: %v", err))
		return nil, fmt.Errorf("价格缓存数据反序列化失败: %w", err)
	}

	// 检查促销价格是否过期
	if price.IsPromotion && !price.PromotionEndTime.IsZero() {
		if time.Now().After(price.PromotionEndTime) {
			// 促销已过期，更新状态
			price.IsPromotion = false
			price.PriceStatus = "normal"
			price.PromotionPrice = ""

			// 更新缓存
			if err := pcs.updatePriceCache(productID, &price); err != nil {
				logger.Error(fmt.Sprintf("更新过期促销价格失败: %v", err))
			}
		}
	}

	return &price, nil
}

// SetPrice 设置价格缓存
func (pcs *PriceCacheService) SetPrice(product *model.Product) error {
	key := pcs.keyManager.GenerateProductPriceKey(product.ID)

	// 转换为缓存数据
	cacheData := ConvertToPriceCacheData(product)

	// 序列化
	jsonData, err := json.Marshal(cacheData)
	if err != nil {
		logger.Error(fmt.Sprintf("价格数据序列化失败: %v", err))
		return fmt.Errorf("价格数据序列化失败: %w", err)
	}

	// 获取TTL
	ttl := GetTTL("price")

	// 存储到缓存
	if err := pcs.cacheManager.Set(key, string(jsonData), ttl); err != nil {
		logger.Error(fmt.Sprintf("设置价格缓存失败: %v", err))
		return fmt.Errorf("设置价格缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("价格缓存设置成功: ProductID=%d, Price=%s, Version=%d, TTL=%v",
		product.ID, product.Price.String(), product.Version, ttl))
	return nil
}

// UpdatePrice 更新价格缓存
func (pcs *PriceCacheService) UpdatePrice(request *PriceUpdateRequest) error {
	// 获取当前价格缓存
	currentPrice, err := pcs.GetPrice(request.ProductID)
	if err != nil {
		return fmt.Errorf("获取当前价格缓存失败: %w", err)
	}

	if currentPrice == nil {
		return fmt.Errorf("价格缓存不存在: ProductID=%d", request.ProductID)
	}

	// 记录价格变更历史
	oldPrice, _ := decimal.NewFromString(currentPrice.Price)
	if !oldPrice.Equal(request.Price) {
		changeType := "increase"
		if request.Price.LessThan(oldPrice) {
			changeType = "decrease"
		}

		historyRecord := &PriceHistoryRecord{
			ProductID:  request.ProductID,
			OldPrice:   oldPrice,
			NewPrice:   request.Price,
			ChangeType: changeType,
			Reason:     request.Reason,
			ChangeTime: time.Now(),
		}

		// 记录价格变更历史
		pcs.recordPriceHistory(historyRecord)
	}

	// 更新价格数据
	currentPrice.Price = request.Price.String()
	if !request.OriginPrice.IsZero() {
		currentPrice.OriginPrice = request.OriginPrice.String()
	}
	if !request.CostPrice.IsZero() {
		currentPrice.CostPrice = request.CostPrice.String()
	}
	currentPrice.LastPriceChange = time.Now()
	currentPrice.PriceChangeCount++
	currentPrice.UpdatedAt = time.Now()
	currentPrice.CachedAt = time.Now()
	currentPrice.Version++

	// 更新缓存
	return pcs.updatePriceCache(request.ProductID, currentPrice)
}

// updatePriceCache 内部方法：更新价格缓存
func (pcs *PriceCacheService) updatePriceCache(productID uint, priceData *PriceCacheData) error {
	key := pcs.keyManager.GenerateProductPriceKey(productID)

	// 序列化并保存
	jsonData, err := json.Marshal(priceData)
	if err != nil {
		return fmt.Errorf("价格数据序列化失败: %w", err)
	}

	ttl := GetTTL("price")
	if err := pcs.cacheManager.Set(key, string(jsonData), ttl); err != nil {
		return fmt.Errorf("更新价格缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("价格缓存更新成功: ProductID=%d, Price=%s, Version=%d",
		productID, priceData.Price, priceData.Version))
	return nil
}

// recordPriceHistory 记录价格变更历史
func (pcs *PriceCacheService) recordPriceHistory(record *PriceHistoryRecord) {
	// 将价格历史记录存储到Redis List中
	historyKey := fmt.Sprintf("%s:price_history:%d", "mall", record.ProductID)
	historyData, err := json.Marshal(record)
	if err != nil {
		logger.Error(fmt.Sprintf("价格历史数据序列化失败: %v", err))
		return
	}

	if err := pcs.cacheManager.LPush(historyKey, string(historyData)); err != nil {
		logger.Error(fmt.Sprintf("记录价格历史失败: %v", err))
		return
	}

	// 保持历史记录数量限制（最多保留100条）
	if length, err := pcs.cacheManager.LLen(historyKey); err == nil && length > 100 {
		// 删除最旧的记录
		pcs.cacheManager.RPop(historyKey)
	}

	logger.Info(fmt.Sprintf("价格变更历史记录: ProductID=%d, %s→%s, 原因=%s",
		record.ProductID, record.OldPrice.String(), record.NewPrice.String(), record.Reason))
}

// SetPromotionPrice 设置促销价格
func (pcs *PriceCacheService) SetPromotionPrice(request *PromotionPriceRequest) error {
	// 获取当前价格缓存
	currentPrice, err := pcs.GetPrice(request.ProductID)
	if err != nil {
		return fmt.Errorf("获取当前价格缓存失败: %w", err)
	}

	if currentPrice == nil {
		return fmt.Errorf("价格缓存不存在: ProductID=%d", request.ProductID)
	}

	// 验证促销时间
	if request.StartTime.After(request.EndTime) {
		return fmt.Errorf("促销开始时间不能晚于结束时间")
	}

	// 更新促销信息
	currentPrice.PromotionPrice = request.PromotionPrice.String()
	currentPrice.PromotionStartTime = request.StartTime
	currentPrice.PromotionEndTime = request.EndTime
	currentPrice.PromotionType = request.PromotionType
	currentPrice.PromotionValue = request.PromotionValue.String()
	currentPrice.IsPromotion = true
	currentPrice.PriceStatus = "promotion"
	currentPrice.UpdatedAt = time.Now()
	currentPrice.CachedAt = time.Now()
	currentPrice.Version++

	// 记录促销开始历史
	historyRecord := &PriceHistoryRecord{
		ProductID:  request.ProductID,
		OldPrice:   decimal.RequireFromString(currentPrice.Price),
		NewPrice:   request.PromotionPrice,
		ChangeType: "promotion_start",
		Reason:     fmt.Sprintf("促销活动开始: %s", request.PromotionType),
		ChangeTime: time.Now(),
	}
	pcs.recordPriceHistory(historyRecord)

	// 更新缓存
	return pcs.updatePriceCache(request.ProductID, currentPrice)
}

// ClearPromotionPrice 清除促销价格
func (pcs *PriceCacheService) ClearPromotionPrice(productID uint) error {
	// 获取当前价格缓存
	currentPrice, err := pcs.GetPrice(productID)
	if err != nil {
		return fmt.Errorf("获取当前价格缓存失败: %w", err)
	}

	if currentPrice == nil {
		return fmt.Errorf("价格缓存不存在: ProductID=%d", productID)
	}

	if !currentPrice.IsPromotion {
		return nil // 没有促销价格，无需清除
	}

	// 记录促销结束历史
	historyRecord := &PriceHistoryRecord{
		ProductID:  productID,
		OldPrice:   decimal.RequireFromString(currentPrice.PromotionPrice),
		NewPrice:   decimal.RequireFromString(currentPrice.Price),
		ChangeType: "promotion_end",
		Reason:     "促销活动结束",
		ChangeTime: time.Now(),
	}
	pcs.recordPriceHistory(historyRecord)

	// 清除促销信息
	currentPrice.PromotionPrice = ""
	currentPrice.PromotionStartTime = time.Time{}
	currentPrice.PromotionEndTime = time.Time{}
	currentPrice.PromotionType = ""
	currentPrice.PromotionValue = ""
	currentPrice.IsPromotion = false
	currentPrice.PriceStatus = "normal"
	currentPrice.UpdatedAt = time.Now()
	currentPrice.CachedAt = time.Now()
	currentPrice.Version++

	// 更新缓存
	return pcs.updatePriceCache(productID, currentPrice)
}

// GetEffectivePrice 获取有效价格（考虑促销、会员等）
func (pcs *PriceCacheService) GetEffectivePrice(productID uint, userType string) (*decimal.Decimal, error) {
	priceData, err := pcs.GetPrice(productID)
	if err != nil {
		return nil, err
	}

	if priceData == nil {
		return nil, fmt.Errorf("价格缓存不存在: ProductID=%d", productID)
	}

	// 检查促销价格
	if priceData.IsPromotion && !priceData.PromotionEndTime.IsZero() {
		now := time.Now()
		if now.After(priceData.PromotionStartTime) && now.Before(priceData.PromotionEndTime) {
			promotionPrice, err := decimal.NewFromString(priceData.PromotionPrice)
			if err == nil {
				return &promotionPrice, nil
			}
		}
	}

	// 检查VIP价格
	if userType == "vip" && priceData.IsVipPrice && priceData.VipPrice != "" {
		vipPrice, err := decimal.NewFromString(priceData.VipPrice)
		if err == nil {
			return &vipPrice, nil
		}
	}

	// 返回普通价格
	normalPrice, err := decimal.NewFromString(priceData.Price)
	if err != nil {
		return nil, fmt.Errorf("价格数据格式错误: %w", err)
	}

	return &normalPrice, nil
}

// DeletePrice 删除价格缓存
func (pcs *PriceCacheService) DeletePrice(productID uint) error {
	key := pcs.keyManager.GenerateProductPriceKey(productID)

	if err := pcs.cacheManager.Delete(key); err != nil {
		logger.Error(fmt.Sprintf("删除价格缓存失败: %v", err))
		return fmt.Errorf("删除价格缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("价格缓存删除成功: ProductID=%d", productID))
	return nil
}

// GetPrices 批量获取价格缓存
func (pcs *PriceCacheService) GetPrices(productIDs []uint) (map[uint]*PriceCacheData, error) {
	if len(productIDs) == 0 {
		return make(map[uint]*PriceCacheData), nil
	}

	// 生成批量键
	keys := pcs.keyManager.GenerateBatchKeys("price", productIDs)

	// 批量获取
	values, err := pcs.cacheManager.MGet(keys)
	if err != nil {
		logger.Error(fmt.Sprintf("批量获取价格缓存失败: %v", err))
		return nil, fmt.Errorf("批量获取价格缓存失败: %w", err)
	}

	// 解析结果
	result := make(map[uint]*PriceCacheData)
	for i, value := range values {
		if value == nil {
			continue // 跳过未命中的缓存
		}

		jsonStr, ok := value.(string)
		if !ok {
			logger.Error(fmt.Sprintf("价格缓存数据格式错误: ProductID=%d", productIDs[i]))
			continue
		}

		var price PriceCacheData
		if err := json.Unmarshal([]byte(jsonStr), &price); err != nil {
			logger.Error(fmt.Sprintf("价格缓存数据反序列化失败: ProductID=%d, Error=%v", productIDs[i], err))
			continue
		}

		result[productIDs[i]] = &price
	}

	logger.Info(fmt.Sprintf("批量获取价格缓存完成: 请求=%d, 命中=%d", len(productIDs), len(result)))
	return result, nil
}

// SetPrices 批量设置价格缓存
func (pcs *PriceCacheService) SetPrices(products []*model.Product) error {
	if len(products) == 0 {
		return nil
	}

	// 准备批量数据
	pairs := make(map[string]interface{})
	ttl := GetTTL("price")

	for _, product := range products {
		key := pcs.keyManager.GenerateProductPriceKey(product.ID)
		cacheData := ConvertToPriceCacheData(product)

		jsonData, err := json.Marshal(cacheData)
		if err != nil {
			logger.Error(fmt.Sprintf("价格数据序列化失败: ProductID=%d, Error=%v", product.ID, err))
			continue
		}

		pairs[key] = string(jsonData)
	}

	// 批量设置
	if err := pcs.cacheManager.MSet(pairs, ttl); err != nil {
		logger.Error(fmt.Sprintf("批量设置价格缓存失败: %v", err))
		return fmt.Errorf("批量设置价格缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("批量设置价格缓存成功: 数量=%d, TTL=%v", len(pairs), ttl))
	return nil
}

// GetPriceHistory 获取价格变更历史
func (pcs *PriceCacheService) GetPriceHistory(productID uint, limit int64) ([]*PriceHistoryRecord, error) {
	historyKey := fmt.Sprintf("%s:price_history:%d", "mall", productID)

	// 获取历史记录
	historyData, err := pcs.cacheManager.LRange(historyKey, 0, limit-1)
	if err != nil {
		return nil, fmt.Errorf("获取价格历史失败: %w", err)
	}

	var history []*PriceHistoryRecord
	for _, data := range historyData {
		if dataStr, ok := data.(string); ok {
			var record PriceHistoryRecord
			if err := json.Unmarshal([]byte(dataStr), &record); err != nil {
				logger.Error(fmt.Sprintf("价格历史数据反序列化失败: %v", err))
				continue
			}
			history = append(history, &record)
		}
	}

	return history, nil
}

// ClearPriceHistory 清空价格变更历史
func (pcs *PriceCacheService) ClearPriceHistory(productID uint) error {
	historyKey := fmt.Sprintf("%s:price_history:%d", "mall", productID)
	return pcs.cacheManager.Delete(historyKey)
}

// ExistsPrice 检查价格缓存是否存在
func (pcs *PriceCacheService) ExistsPrice(productID uint) bool {
	key := pcs.keyManager.GenerateProductPriceKey(productID)
	return pcs.cacheManager.Exists(key)
}

// GetPriceTTL 获取价格缓存剩余TTL
func (pcs *PriceCacheService) GetPriceTTL(productID uint) (time.Duration, error) {
	key := pcs.keyManager.GenerateProductPriceKey(productID)
	return pcs.cacheManager.TTL(key)
}

// RefreshPriceTTL 刷新价格缓存TTL
func (pcs *PriceCacheService) RefreshPriceTTL(productID uint) error {
	key := pcs.keyManager.GenerateProductPriceKey(productID)
	ttl := GetTTL("price")
	return pcs.cacheManager.Expire(key, ttl)
}

// WarmupPrices 价格缓存预热
func (pcs *PriceCacheService) WarmupPrices(products []*model.Product) error {
	return pcs.SetPrices(products)
}

// GetPromotionProducts 获取促销商品列表
func (pcs *PriceCacheService) GetPromotionProducts() ([]uint, error) {
	// 使用Set存储促销商品ID
	key := fmt.Sprintf("%s:promotion_products", "mall")

	members, err := pcs.cacheManager.SMembers(key)
	if err != nil {
		return nil, fmt.Errorf("获取促销商品列表失败: %w", err)
	}

	var productIDs []uint
	for _, member := range members {
		if memberStr, ok := member.(string); ok {
			if id, err := strconv.ParseUint(memberStr, 10, 32); err == nil {
				productIDs = append(productIDs, uint(id))
			}
		}
	}

	return productIDs, nil
}

// AddPromotionProduct 添加促销商品
func (pcs *PriceCacheService) AddPromotionProduct(productID uint) error {
	key := fmt.Sprintf("%s:promotion_products", "mall")
	return pcs.cacheManager.SAdd(key, strconv.Itoa(int(productID)))
}

// RemovePromotionProduct 移除促销商品
func (pcs *PriceCacheService) RemovePromotionProduct(productID uint) error {
	key := fmt.Sprintf("%s:promotion_products", "mall")
	return pcs.cacheManager.SRem(key, strconv.Itoa(int(productID)))
}

// BatchUpdatePromotionPrices 批量更新促销价格
func (pcs *PriceCacheService) BatchUpdatePromotionPrices(requests []*PromotionPriceRequest) error {
	if len(requests) == 0 {
		return nil
	}

	// 逐个处理促销价格设置
	for _, request := range requests {
		if err := pcs.SetPromotionPrice(request); err != nil {
			logger.Error(fmt.Sprintf("批量设置促销价格失败: ProductID=%d, Error=%v",
				request.ProductID, err))
			return fmt.Errorf("批量设置促销价格失败: %w", err)
		}

		// 添加到促销商品列表
		pcs.AddPromotionProduct(request.ProductID)
	}

	logger.Info(fmt.Sprintf("批量设置促销价格成功: 数量=%d", len(requests)))
	return nil
}

// GetPriceStats 获取价格统计信息
func (pcs *PriceCacheService) GetPriceStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// 获取缓存指标
	if metrics := pcs.cacheManager.GetMetrics(); metrics != nil {
		stats["total_ops"] = metrics.TotalOps
		stats["hit_count"] = metrics.HitCount
		stats["miss_count"] = metrics.MissCount
		stats["hit_rate"] = metrics.HitRate
		stats["error_count"] = metrics.ErrorCount
		stats["last_updated"] = metrics.LastUpdated
	}

	// 获取连接池统计
	if connStats := pcs.cacheManager.GetConnectionStats(); connStats != nil {
		stats["total_conns"] = connStats.TotalConns
		stats["idle_conns"] = connStats.IdleConns
		stats["hits"] = connStats.Hits
		stats["misses"] = connStats.Misses
	}

	// 获取促销商品数量
	if promotionProducts, err := pcs.GetPromotionProducts(); err == nil {
		stats["promotion_count"] = len(promotionProducts)
	}

	return stats
}

// DeletePrices 批量删除价格缓存
func (pcs *PriceCacheService) DeletePrices(productIDs []uint) error {
	if len(productIDs) == 0 {
		return nil
	}

	keys := pcs.keyManager.GenerateBatchKeys("price", productIDs)

	if err := pcs.cacheManager.MDelete(keys); err != nil {
		logger.Error(fmt.Sprintf("批量删除价格缓存失败: %v", err))
		return fmt.Errorf("批量删除价格缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("批量删除价格缓存成功: 数量=%d", len(productIDs)))
	return nil
}

// SyncPriceFromDB 从数据库同步价格到缓存（用于缓存失效后的恢复）
func (pcs *PriceCacheService) SyncPriceFromDB(productID uint, product *model.Product) error {
	// 设置价格缓存
	if err := pcs.SetPrice(product); err != nil {
		return fmt.Errorf("同步价格缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("价格同步完成: ProductID=%d, Price=%s, Version=%d",
		productID, product.Price.String(), product.Version))
	return nil
}

// CheckPromotionExpiry 检查并清理过期促销
func (pcs *PriceCacheService) CheckPromotionExpiry() error {
	// 获取所有促销商品
	promotionProducts, err := pcs.GetPromotionProducts()
	if err != nil {
		return fmt.Errorf("获取促销商品列表失败: %w", err)
	}

	now := time.Now()
	expiredCount := 0

	for _, productID := range promotionProducts {
		priceData, err := pcs.GetPrice(productID)
		if err != nil {
			logger.Error(fmt.Sprintf("获取促销商品价格失败: ProductID=%d, Error=%v", productID, err))
			continue
		}

		if priceData != nil && priceData.IsPromotion && !priceData.PromotionEndTime.IsZero() {
			if now.After(priceData.PromotionEndTime) {
				// 促销已过期，清除促销价格
				if err := pcs.ClearPromotionPrice(productID); err != nil {
					logger.Error(fmt.Sprintf("清除过期促销价格失败: ProductID=%d, Error=%v", productID, err))
				} else {
					// 从促销商品列表中移除
					pcs.RemovePromotionProduct(productID)
					expiredCount++
				}
			}
		}
	}

	if expiredCount > 0 {
		logger.Info(fmt.Sprintf("清理过期促销完成: 数量=%d", expiredCount))
	}

	return nil
}
