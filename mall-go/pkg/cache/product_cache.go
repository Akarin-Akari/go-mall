package cache

import (
	"encoding/json"
	"fmt"
	"mall-go/internal/model"
	"mall-go/pkg/logger"
	"strconv"
	"time"
)

// ProductCacheService 商品缓存服务
type ProductCacheService struct {
	cacheManager CacheManager
	keyManager   *CacheKeyManager
}

// NewProductCacheService 创建商品缓存服务
func NewProductCacheService(cacheManager CacheManager, keyManager *CacheKeyManager) *ProductCacheService {
	return &ProductCacheService{
		cacheManager: cacheManager,
		keyManager:   keyManager,
	}
}

// ProductCacheData 商品缓存数据结构
type ProductCacheData struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	SubTitle    string `json:"sub_title"`
	Description string `json:"description"`
	Detail      string `json:"detail"`
	CategoryID  uint   `json:"category_id"`
	BrandID     uint   `json:"brand_id"`
	MerchantID  uint   `json:"merchant_id"`

	// 冗余字段
	CategoryName string `json:"category_name"`
	BrandName    string `json:"brand_name"`
	MerchantName string `json:"merchant_name"`

	// 价格信息
	Price       string `json:"price"` // 使用字符串存储decimal
	OriginPrice string `json:"origin_price"`
	CostPrice   string `json:"cost_price"`

	// 库存信息
	Stock     int `json:"stock"`
	MinStock  int `json:"min_stock"`
	MaxStock  int `json:"max_stock"`
	SoldCount int `json:"sold_count"`
	Version   int `json:"version"` // 乐观锁版本号

	// 商品属性
	Weight string `json:"weight"`
	Volume string `json:"volume"`
	Unit   string `json:"unit"`

	// 状态信息
	Status      string `json:"status"`
	IsHot       bool   `json:"is_hot"`
	IsNew       bool   `json:"is_new"`
	IsRecommend bool   `json:"is_recommend"`

	// SEO信息
	SEOTitle       string `json:"seo_title"`
	SEOKeywords    string `json:"seo_keywords"`
	SEODescription string `json:"seo_description"`

	// 排序和显示
	Sort      int `json:"sort"`
	ViewCount int `json:"view_count"`

	// 时间戳
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 缓存元数据
	CachedAt time.Time `json:"cached_at"`
}

// ConvertToProductCacheData 将Product模型转换为缓存数据
func ConvertToProductCacheData(product *model.Product) *ProductCacheData {
	return &ProductCacheData{
		ID:          product.ID,
		Name:        product.Name,
		SubTitle:    product.SubTitle,
		Description: product.Description,
		Detail:      product.Detail,
		CategoryID:  product.CategoryID,
		BrandID:     product.BrandID,
		MerchantID:  product.MerchantID,

		CategoryName: product.CategoryName,
		BrandName:    product.BrandName,
		MerchantName: product.MerchantName,

		Price:       product.Price.String(),
		OriginPrice: product.OriginPrice.String(),
		CostPrice:   product.CostPrice.String(),

		Stock:     product.Stock,
		MinStock:  product.MinStock,
		MaxStock:  product.MaxStock,
		SoldCount: product.SoldCount,
		Version:   product.Version,

		Weight: product.Weight.String(),
		Volume: product.Volume.String(),
		Unit:   product.Unit,

		Status:      product.Status,
		IsHot:       product.IsHot,
		IsNew:       product.IsNew,
		IsRecommend: product.IsRecommend,

		SEOTitle:       product.SEOTitle,
		SEOKeywords:    product.SEOKeywords,
		SEODescription: product.SEODescription,

		Sort:      product.Sort,
		ViewCount: product.ViewCount,

		CreatedAt: product.CreatedAt,
		UpdatedAt: product.UpdatedAt,
		CachedAt:  time.Now(),
	}
}

// GetProduct 获取商品缓存
func (pcs *ProductCacheService) GetProduct(productID uint) (*ProductCacheData, error) {
	key := pcs.keyManager.GenerateProductKey(productID)

	// 从缓存获取
	data, err := pcs.cacheManager.Get(key)
	if err != nil {
		logger.Error(fmt.Sprintf("获取商品缓存失败: %v", err))
		return nil, fmt.Errorf("获取商品缓存失败: %w", err)
	}

	if data == nil {
		return nil, nil // 缓存未命中
	}

	// 反序列化
	var product ProductCacheData
	jsonStr, ok := data.(string)
	if !ok {
		return nil, fmt.Errorf("缓存数据格式错误")
	}

	if err := json.Unmarshal([]byte(jsonStr), &product); err != nil {
		logger.Error(fmt.Sprintf("商品缓存数据反序列化失败: %v", err))
		return nil, fmt.Errorf("商品缓存数据反序列化失败: %w", err)
	}

	return &product, nil
}

// SetProduct 设置商品缓存
func (pcs *ProductCacheService) SetProduct(product *model.Product) error {
	key := pcs.keyManager.GenerateProductKey(product.ID)

	// 转换为缓存数据
	cacheData := ConvertToProductCacheData(product)

	// 序列化
	jsonData, err := json.Marshal(cacheData)
	if err != nil {
		logger.Error(fmt.Sprintf("商品数据序列化失败: %v", err))
		return fmt.Errorf("商品数据序列化失败: %w", err)
	}

	// 获取TTL
	ttl := GetTTL("product")

	// 存储到缓存
	if err := pcs.cacheManager.Set(key, string(jsonData), ttl); err != nil {
		logger.Error(fmt.Sprintf("设置商品缓存失败: %v", err))
		return fmt.Errorf("设置商品缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("商品缓存设置成功: ID=%d, Key=%s, TTL=%v", product.ID, key, ttl))
	return nil
}

// DeleteProduct 删除商品缓存
func (pcs *ProductCacheService) DeleteProduct(productID uint) error {
	key := pcs.keyManager.GenerateProductKey(productID)

	if err := pcs.cacheManager.Delete(key); err != nil {
		logger.Error(fmt.Sprintf("删除商品缓存失败: %v", err))
		return fmt.Errorf("删除商品缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("商品缓存删除成功: ID=%d, Key=%s", productID, key))
	return nil
}

// GetProducts 批量获取商品缓存
func (pcs *ProductCacheService) GetProducts(productIDs []uint) (map[uint]*ProductCacheData, error) {
	if len(productIDs) == 0 {
		return make(map[uint]*ProductCacheData), nil
	}

	// 生成批量键
	keys := pcs.keyManager.GenerateBatchKeys("product", productIDs)

	// 批量获取
	values, err := pcs.cacheManager.MGet(keys)
	if err != nil {
		logger.Error(fmt.Sprintf("批量获取商品缓存失败: %v", err))
		return nil, fmt.Errorf("批量获取商品缓存失败: %w", err)
	}

	// 解析结果
	result := make(map[uint]*ProductCacheData)
	for i, value := range values {
		if value == nil {
			continue // 跳过未命中的缓存
		}

		jsonStr, ok := value.(string)
		if !ok {
			logger.Error(fmt.Sprintf("商品缓存数据格式错误: ID=%d", productIDs[i]))
			continue
		}

		var product ProductCacheData
		if err := json.Unmarshal([]byte(jsonStr), &product); err != nil {
			logger.Error(fmt.Sprintf("商品缓存数据反序列化失败: ID=%d, Error=%v", productIDs[i], err))
			continue
		}

		result[productIDs[i]] = &product
	}

	logger.Info(fmt.Sprintf("批量获取商品缓存完成: 请求=%d, 命中=%d", len(productIDs), len(result)))
	return result, nil
}

// SetProducts 批量设置商品缓存
func (pcs *ProductCacheService) SetProducts(products []*model.Product) error {
	if len(products) == 0 {
		return nil
	}

	// 准备批量数据
	pairs := make(map[string]interface{})
	ttl := GetTTL("product")

	for _, product := range products {
		key := pcs.keyManager.GenerateProductKey(product.ID)
		cacheData := ConvertToProductCacheData(product)

		jsonData, err := json.Marshal(cacheData)
		if err != nil {
			logger.Error(fmt.Sprintf("商品数据序列化失败: ID=%d, Error=%v", product.ID, err))
			continue
		}

		pairs[key] = string(jsonData)
	}

	// 批量设置
	if err := pcs.cacheManager.MSet(pairs, ttl); err != nil {
		logger.Error(fmt.Sprintf("批量设置商品缓存失败: %v", err))
		return fmt.Errorf("批量设置商品缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("批量设置商品缓存成功: 数量=%d, TTL=%v", len(pairs), ttl))
	return nil
}

// DeleteProducts 批量删除商品缓存
func (pcs *ProductCacheService) DeleteProducts(productIDs []uint) error {
	if len(productIDs) == 0 {
		return nil
	}

	// 生成批量键
	keys := pcs.keyManager.GenerateBatchKeys("product", productIDs)

	// 批量删除
	if err := pcs.cacheManager.MDelete(keys); err != nil {
		logger.Error(fmt.Sprintf("批量删除商品缓存失败: %v", err))
		return fmt.Errorf("批量删除商品缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("批量删除商品缓存成功: 数量=%d", len(keys)))
	return nil
}

// ExistsProduct 检查商品缓存是否存在
func (pcs *ProductCacheService) ExistsProduct(productID uint) bool {
	key := pcs.keyManager.GenerateProductKey(productID)
	return pcs.cacheManager.Exists(key)
}

// GetProductTTL 获取商品缓存剩余TTL
func (pcs *ProductCacheService) GetProductTTL(productID uint) (time.Duration, error) {
	key := pcs.keyManager.GenerateProductKey(productID)
	return pcs.cacheManager.TTL(key)
}

// RefreshProductTTL 刷新商品缓存TTL
func (pcs *ProductCacheService) RefreshProductTTL(productID uint) error {
	key := pcs.keyManager.GenerateProductKey(productID)
	ttl := GetTTL("product")

	if err := pcs.cacheManager.Expire(key, ttl); err != nil {
		logger.Error(fmt.Sprintf("刷新商品缓存TTL失败: ID=%d, Error=%v", productID, err))
		return fmt.Errorf("刷新商品缓存TTL失败: %w", err)
	}

	return nil
}

// WarmupProducts 商品缓存预热
func (pcs *ProductCacheService) WarmupProducts(products []*model.Product) error {
	if len(products) == 0 {
		return nil
	}

	logger.Info(fmt.Sprintf("开始商品缓存预热: 数量=%d", len(products)))

	// 批量设置缓存
	if err := pcs.SetProducts(products); err != nil {
		return fmt.Errorf("商品缓存预热失败: %w", err)
	}

	logger.Info(fmt.Sprintf("商品缓存预热完成: 数量=%d", len(products)))
	return nil
}

// GetHotProducts 获取热门商品缓存
func (pcs *ProductCacheService) GetHotProducts() ([]uint, error) {
	key := pcs.keyManager.GenerateHotProductsKey()

	data, err := pcs.cacheManager.Get(key)
	if err != nil {
		return nil, fmt.Errorf("获取热门商品缓存失败: %w", err)
	}

	if data == nil {
		return nil, nil // 缓存未命中
	}

	jsonStr, ok := data.(string)
	if !ok {
		return nil, fmt.Errorf("热门商品缓存数据格式错误")
	}

	var productIDs []uint
	if err := json.Unmarshal([]byte(jsonStr), &productIDs); err != nil {
		return nil, fmt.Errorf("热门商品缓存数据反序列化失败: %w", err)
	}

	return productIDs, nil
}

// SetHotProducts 设置热门商品缓存
func (pcs *ProductCacheService) SetHotProducts(productIDs []uint) error {
	key := pcs.keyManager.GenerateHotProductsKey()

	jsonData, err := json.Marshal(productIDs)
	if err != nil {
		return fmt.Errorf("热门商品数据序列化失败: %w", err)
	}

	ttl := GetTTL("hot")
	if err := pcs.cacheManager.Set(key, string(jsonData), ttl); err != nil {
		return fmt.Errorf("设置热门商品缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("热门商品缓存设置成功: 数量=%d, TTL=%v", len(productIDs), ttl))
	return nil
}

// IncrementViewCount 增加商品浏览量
func (pcs *ProductCacheService) IncrementViewCount(productID uint) error {
	key := pcs.keyManager.GenerateCounterKey("view", productID)

	// 使用Hash结构存储计数器
	field := "count"
	exists := pcs.cacheManager.HExists(key, field)

	if !exists {
		// 初始化计数器
		if err := pcs.cacheManager.HSet(key, field, "1"); err != nil {
			return fmt.Errorf("初始化浏览量计数器失败: %w", err)
		}
	} else {
		// 获取当前值并增加
		currentValue, err := pcs.cacheManager.HGet(key, field)
		if err != nil {
			return fmt.Errorf("获取当前浏览量失败: %w", err)
		}

		count := 0
		if currentValue != nil {
			if countStr, ok := currentValue.(string); ok {
				if c, err := strconv.Atoi(countStr); err == nil {
					count = c
				}
			}
		}

		count++
		if err := pcs.cacheManager.HSet(key, field, strconv.Itoa(count)); err != nil {
			return fmt.Errorf("更新浏览量计数器失败: %w", err)
		}
	}

	// 设置TTL
	ttl := GetTTL("counter")
	if err := pcs.cacheManager.Expire(key, ttl); err != nil {
		logger.Error(fmt.Sprintf("设置浏览量计数器TTL失败: %v", err))
	}

	return nil
}

// GetViewCount 获取商品浏览量
func (pcs *ProductCacheService) GetViewCount(productID uint) (int, error) {
	key := pcs.keyManager.GenerateCounterKey("view", productID)
	field := "count"

	value, err := pcs.cacheManager.HGet(key, field)
	if err != nil {
		return 0, fmt.Errorf("获取浏览量失败: %w", err)
	}

	if value == nil {
		return 0, nil
	}

	countStr, ok := value.(string)
	if !ok {
		return 0, fmt.Errorf("浏览量数据格式错误")
	}

	count, err := strconv.Atoi(countStr)
	if err != nil {
		return 0, fmt.Errorf("浏览量数据转换失败: %w", err)
	}

	return count, nil
}

// GetCacheStats 获取商品缓存统计信息
func (pcs *ProductCacheService) GetCacheStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// 获取缓存管理器统计
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

	return stats
}
