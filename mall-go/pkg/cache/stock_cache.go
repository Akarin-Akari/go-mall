package cache

import (
	"encoding/json"
	"fmt"
	"mall-go/internal/model"
	"mall-go/pkg/logger"
	"strconv"
	"time"
)

// StockCacheService 库存缓存服务
type StockCacheService struct {
	cacheManager CacheManager
	keyManager   *CacheKeyManager
}

// NewStockCacheService 创建库存缓存服务
func NewStockCacheService(cacheManager CacheManager, keyManager *CacheKeyManager) *StockCacheService {
	return &StockCacheService{
		cacheManager: cacheManager,
		keyManager:   keyManager,
	}
}

// StockCacheData 库存缓存数据结构
type StockCacheData struct {
	ProductID uint `json:"product_id"`
	Stock     int  `json:"stock"`
	MinStock  int  `json:"min_stock"`
	MaxStock  int  `json:"max_stock"`
	SoldCount int  `json:"sold_count"`
	Version   int  `json:"version"` // 乐观锁版本号

	// 库存状态
	Status       string    `json:"status"`         // active, inactive, out_of_stock
	IsLowStock   bool      `json:"is_low_stock"`   // 是否低库存
	LastSoldTime time.Time `json:"last_sold_time"` // 最后销售时间

	// 缓存元数据
	CachedAt  time.Time `json:"cached_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// StockDeductionRequest 库存扣减请求
type StockDeductionRequest struct {
	ProductID uint   `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Reason    string `json:"reason"` // 扣减原因：order, return, adjust等
}

// StockDeductionResult 库存扣减结果
type StockDeductionResult struct {
	ProductID  uint   `json:"product_id"`
	Success    bool   `json:"success"`
	OldStock   int    `json:"old_stock"`
	NewStock   int    `json:"new_stock"`
	OldVersion int    `json:"old_version"`
	NewVersion int    `json:"new_version"`
	Retries    int    `json:"retries"`
	Error      string `json:"error,omitempty"`
}

// LowStockAlert 低库存预警
type LowStockAlert struct {
	ProductID    uint      `json:"product_id"`
	ProductName  string    `json:"product_name"`
	CurrentStock int       `json:"current_stock"`
	MinStock     int       `json:"min_stock"`
	AlertTime    time.Time `json:"alert_time"`
}

// ConvertToStockCacheData 将Product模型转换为库存缓存数据
func ConvertToStockCacheData(product *model.Product) *StockCacheData {
	isLowStock := product.Stock <= product.MinStock && product.MinStock > 0
	status := "active"
	if product.Stock <= 0 {
		status = "out_of_stock"
	} else if product.Status != "active" {
		status = "inactive"
	}

	return &StockCacheData{
		ProductID: product.ID,
		Stock:     product.Stock,
		MinStock:  product.MinStock,
		MaxStock:  product.MaxStock,
		SoldCount: product.SoldCount,
		Version:   product.Version,

		Status:       status,
		IsLowStock:   isLowStock,
		LastSoldTime: time.Now(), // 这里应该从实际业务数据获取

		CachedAt:  time.Now(),
		UpdatedAt: product.UpdatedAt,
	}
}

// GetStock 获取库存缓存
func (scs *StockCacheService) GetStock(productID uint) (*StockCacheData, error) {
	key := scs.keyManager.GenerateProductStockKey(productID)

	// 从缓存获取
	data, err := scs.cacheManager.Get(key)
	if err != nil {
		logger.Error(fmt.Sprintf("获取库存缓存失败: %v", err))
		return nil, fmt.Errorf("获取库存缓存失败: %w", err)
	}

	if data == nil {
		return nil, nil // 缓存未命中
	}

	// 反序列化
	var stock StockCacheData
	jsonStr, ok := data.(string)
	if !ok {
		return nil, fmt.Errorf("库存缓存数据格式错误")
	}

	if err := json.Unmarshal([]byte(jsonStr), &stock); err != nil {
		logger.Error(fmt.Sprintf("库存缓存数据反序列化失败: %v", err))
		return nil, fmt.Errorf("库存缓存数据反序列化失败: %w", err)
	}

	return &stock, nil
}

// SetStock 设置库存缓存
func (scs *StockCacheService) SetStock(product *model.Product) error {
	key := scs.keyManager.GenerateProductStockKey(product.ID)

	// 转换为缓存数据
	cacheData := ConvertToStockCacheData(product)

	// 序列化
	jsonData, err := json.Marshal(cacheData)
	if err != nil {
		logger.Error(fmt.Sprintf("库存数据序列化失败: %v", err))
		return fmt.Errorf("库存数据序列化失败: %w", err)
	}

	// 获取TTL
	ttl := GetTTL("stock")

	// 存储到缓存
	if err := scs.cacheManager.Set(key, string(jsonData), ttl); err != nil {
		logger.Error(fmt.Sprintf("设置库存缓存失败: %v", err))
		return fmt.Errorf("设置库存缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("库存缓存设置成功: ProductID=%d, Stock=%d, Version=%d, TTL=%v",
		product.ID, product.Stock, product.Version, ttl))
	return nil
}

// UpdateStock 更新库存缓存（原子操作）
func (scs *StockCacheService) UpdateStock(productID uint, newStock int, newVersion int) error {
	key := scs.keyManager.GenerateProductStockKey(productID)

	// 获取当前缓存数据
	currentData, err := scs.GetStock(productID)
	if err != nil {
		return fmt.Errorf("获取当前库存缓存失败: %w", err)
	}

	if currentData == nil {
		return fmt.Errorf("库存缓存不存在: ProductID=%d", productID)
	}

	// 更新数据
	currentData.Stock = newStock
	currentData.Version = newVersion
	currentData.UpdatedAt = time.Now()
	currentData.CachedAt = time.Now()

	// 更新状态
	if newStock <= 0 {
		currentData.Status = "out_of_stock"
	} else if currentData.Status == "out_of_stock" {
		currentData.Status = "active"
	}

	// 更新低库存状态
	currentData.IsLowStock = newStock <= currentData.MinStock && currentData.MinStock > 0

	// 序列化并保存
	jsonData, err := json.Marshal(currentData)
	if err != nil {
		return fmt.Errorf("库存数据序列化失败: %w", err)
	}

	ttl := GetTTL("stock")
	if err := scs.cacheManager.Set(key, string(jsonData), ttl); err != nil {
		return fmt.Errorf("更新库存缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("库存缓存更新成功: ProductID=%d, Stock=%d, Version=%d",
		productID, newStock, newVersion))
	return nil
}

// DeleteStock 删除库存缓存
func (scs *StockCacheService) DeleteStock(productID uint) error {
	key := scs.keyManager.GenerateProductStockKey(productID)

	if err := scs.cacheManager.Delete(key); err != nil {
		logger.Error(fmt.Sprintf("删除库存缓存失败: %v", err))
		return fmt.Errorf("删除库存缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("库存缓存删除成功: ProductID=%d", productID))
	return nil
}

// GetStocks 批量获取库存缓存
func (scs *StockCacheService) GetStocks(productIDs []uint) (map[uint]*StockCacheData, error) {
	if len(productIDs) == 0 {
		return make(map[uint]*StockCacheData), nil
	}

	// 生成批量键
	keys := scs.keyManager.GenerateBatchKeys("stock", productIDs)

	// 批量获取
	values, err := scs.cacheManager.MGet(keys)
	if err != nil {
		logger.Error(fmt.Sprintf("批量获取库存缓存失败: %v", err))
		return nil, fmt.Errorf("批量获取库存缓存失败: %w", err)
	}

	// 解析结果
	result := make(map[uint]*StockCacheData)
	for i, value := range values {
		if value == nil {
			continue // 跳过未命中的缓存
		}

		jsonStr, ok := value.(string)
		if !ok {
			logger.Error(fmt.Sprintf("库存缓存数据格式错误: ProductID=%d", productIDs[i]))
			continue
		}

		var stock StockCacheData
		if err := json.Unmarshal([]byte(jsonStr), &stock); err != nil {
			logger.Error(fmt.Sprintf("库存缓存数据反序列化失败: ProductID=%d, Error=%v", productIDs[i], err))
			continue
		}

		result[productIDs[i]] = &stock
	}

	logger.Info(fmt.Sprintf("批量获取库存缓存完成: 请求=%d, 命中=%d", len(productIDs), len(result)))
	return result, nil
}

// SetStocks 批量设置库存缓存
func (scs *StockCacheService) SetStocks(products []*model.Product) error {
	if len(products) == 0 {
		return nil
	}

	// 准备批量数据
	pairs := make(map[string]interface{})
	ttl := GetTTL("stock")

	for _, product := range products {
		key := scs.keyManager.GenerateProductStockKey(product.ID)
		cacheData := ConvertToStockCacheData(product)

		jsonData, err := json.Marshal(cacheData)
		if err != nil {
			logger.Error(fmt.Sprintf("库存数据序列化失败: ProductID=%d, Error=%v", product.ID, err))
			continue
		}

		pairs[key] = string(jsonData)
	}

	// 批量设置
	if err := scs.cacheManager.MSet(pairs, ttl); err != nil {
		logger.Error(fmt.Sprintf("批量设置库存缓存失败: %v", err))
		return fmt.Errorf("批量设置库存缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("批量设置库存缓存成功: 数量=%d, TTL=%v", len(pairs), ttl))
	return nil
}

// DeductStockWithOptimisticLock 使用乐观锁扣减库存（缓存层）
func (scs *StockCacheService) DeductStockWithOptimisticLock(request *StockDeductionRequest) (*StockDeductionResult, error) {
	if request.Quantity <= 0 {
		return nil, fmt.Errorf("扣减数量必须大于0")
	}

	result := &StockDeductionResult{
		ProductID: request.ProductID,
		Success:   false,
	}

	maxRetries := 5 // 缓存层重试次数
	for retries := 0; retries < maxRetries; retries++ {
		result.Retries = retries

		// 获取当前库存缓存
		currentStock, err := scs.GetStock(request.ProductID)
		if err != nil {
			result.Error = fmt.Sprintf("获取库存缓存失败: %v", err)
			return result, err
		}

		if currentStock == nil {
			result.Error = "库存缓存不存在"
			return result, fmt.Errorf("库存缓存不存在: ProductID=%d", request.ProductID)
		}

		// 检查库存是否足够
		if currentStock.Stock < request.Quantity {
			result.Error = fmt.Sprintf("库存不足，当前库存：%d，需要：%d", currentStock.Stock, request.Quantity)
			return result, fmt.Errorf(result.Error)
		}

		// 记录原始值
		result.OldStock = currentStock.Stock
		result.OldVersion = currentStock.Version

		// 计算新值
		newStock := currentStock.Stock - request.Quantity
		newVersion := currentStock.Version + 1

		// 使用乐观锁更新（通过版本号验证）
		key := scs.keyManager.GenerateProductStockKey(request.ProductID)

		// 构建更新后的数据
		updatedData := *currentStock
		updatedData.Stock = newStock
		updatedData.Version = newVersion
		updatedData.SoldCount += request.Quantity
		updatedData.UpdatedAt = time.Now()
		updatedData.CachedAt = time.Now()
		updatedData.LastSoldTime = time.Now()

		// 更新状态
		if newStock <= 0 {
			updatedData.Status = "out_of_stock"
		}
		updatedData.IsLowStock = newStock <= updatedData.MinStock && updatedData.MinStock > 0

		// 使用Hash操作进行原子性更新（模拟乐观锁）

		// 检查版本号是否匹配
		cachedVersion, err := scs.cacheManager.HGet(key, "version")
		if err != nil {
			// 如果Hash不存在，说明是第一次设置，直接设置
			if err := scs.setStockWithVersion(key, &updatedData); err != nil {
				result.Error = fmt.Sprintf("设置库存缓存失败: %v", err)
				return result, err
			}
		} else {
			// 验证版本号
			if cachedVersionStr, ok := cachedVersion.(string); ok {
				if cachedVersionInt, parseErr := strconv.Atoi(cachedVersionStr); parseErr == nil {
					if cachedVersionInt != currentStock.Version {
						// 版本号不匹配，需要重试
						if retries == maxRetries-1 {
							result.Error = fmt.Sprintf("库存更新失败，并发冲突过多，已重试%d次", maxRetries)
							return result, fmt.Errorf(result.Error)
						}

						// 退避等待
						backoffTime := time.Millisecond * time.Duration(10*(retries+1))
						time.Sleep(backoffTime)
						continue
					}
				}
			}

			// 版本号匹配，执行更新
			if err := scs.setStockWithVersion(key, &updatedData); err != nil {
				result.Error = fmt.Sprintf("更新库存缓存失败: %v", err)
				return result, err
			}
		}

		// 更新成功
		result.Success = true
		result.NewStock = newStock
		result.NewVersion = newVersion

		logger.Info(fmt.Sprintf("库存扣减成功: ProductID=%d, 数量=%d, 库存=%d→%d, 版本=%d→%d, 重试=%d",
			request.ProductID, request.Quantity, result.OldStock, result.NewStock,
			result.OldVersion, result.NewVersion, retries))

		// 检查是否需要低库存预警
		if updatedData.IsLowStock {
			scs.addLowStockAlert(request.ProductID, newStock, updatedData.MinStock)
		}

		return result, nil
	}

	result.Error = fmt.Sprintf("库存更新失败，超过最大重试次数: %d", maxRetries)
	return result, fmt.Errorf(result.Error)
}

// setStockWithVersion 使用Hash结构原子性设置库存和版本
func (scs *StockCacheService) setStockWithVersion(key string, stockData *StockCacheData) error {
	// 使用Hash结构存储，确保原子性
	fields := map[string]interface{}{
		"product_id":     strconv.Itoa(int(stockData.ProductID)),
		"stock":          strconv.Itoa(stockData.Stock),
		"min_stock":      strconv.Itoa(stockData.MinStock),
		"max_stock":      strconv.Itoa(stockData.MaxStock),
		"sold_count":     strconv.Itoa(stockData.SoldCount),
		"version":        strconv.Itoa(stockData.Version),
		"status":         stockData.Status,
		"is_low_stock":   strconv.FormatBool(stockData.IsLowStock),
		"last_sold_time": stockData.LastSoldTime.Format(time.RFC3339),
		"cached_at":      stockData.CachedAt.Format(time.RFC3339),
		"updated_at":     stockData.UpdatedAt.Format(time.RFC3339),
	}

	if err := scs.cacheManager.HMSet(key, fields); err != nil {
		return fmt.Errorf("设置Hash库存数据失败: %w", err)
	}

	// 设置TTL
	ttl := GetTTL("stock")
	if err := scs.cacheManager.Expire(key, ttl); err != nil {
		logger.Error(fmt.Sprintf("设置库存缓存TTL失败: %v", err))
	}

	return nil
}

// BatchDeductStock 批量扣减库存
func (scs *StockCacheService) BatchDeductStock(requests []*StockDeductionRequest) ([]*StockDeductionResult, error) {
	if len(requests) == 0 {
		return []*StockDeductionResult{}, nil
	}

	results := make([]*StockDeductionResult, len(requests))

	// 逐个处理扣减请求（保证原子性）
	for i, request := range requests {
		result, err := scs.DeductStockWithOptimisticLock(request)
		if err != nil {
			// 如果有任何一个失败，需要回滚之前的操作
			scs.rollbackBatchDeduction(requests[:i], results[:i])
			return results, fmt.Errorf("批量库存扣减失败: %w", err)
		}
		results[i] = result
	}

	logger.Info(fmt.Sprintf("批量库存扣减成功: 数量=%d", len(requests)))
	return results, nil
}

// rollbackBatchDeduction 回滚批量扣减操作
func (scs *StockCacheService) rollbackBatchDeduction(requests []*StockDeductionRequest, results []*StockDeductionResult) {
	for i := len(requests) - 1; i >= 0; i-- {
		if results[i] != nil && results[i].Success {
			// 回滚：增加库存
			rollbackRequest := &StockDeductionRequest{
				ProductID: requests[i].ProductID,
				Quantity:  -requests[i].Quantity, // 负数表示增加
				Reason:    "rollback",
			}

			if _, err := scs.DeductStockWithOptimisticLock(rollbackRequest); err != nil {
				logger.Error(fmt.Sprintf("回滚库存扣减失败: ProductID=%d, Error=%v",
					requests[i].ProductID, err))
			}
		}
	}
}

// addLowStockAlert 添加低库存预警
func (scs *StockCacheService) addLowStockAlert(productID uint, currentStock, minStock int) {
	alert := &LowStockAlert{
		ProductID:    productID,
		CurrentStock: currentStock,
		MinStock:     minStock,
		AlertTime:    time.Now(),
	}

	// 将预警信息存储到Redis List中
	alertKey := fmt.Sprintf("%s:low_stock_alerts", "mall") // 使用固定前缀
	alertData, err := json.Marshal(alert)
	if err != nil {
		logger.Error(fmt.Sprintf("低库存预警数据序列化失败: %v", err))
		return
	}

	if err := scs.cacheManager.LPush(alertKey, string(alertData)); err != nil {
		logger.Error(fmt.Sprintf("添加低库存预警失败: %v", err))
		return
	}

	logger.Info(fmt.Sprintf("低库存预警: ProductID=%d, 当前库存=%d, 最小库存=%d",
		productID, currentStock, minStock))
}

// GetLowStockAlerts 获取低库存预警列表
func (scs *StockCacheService) GetLowStockAlerts(limit int64) ([]*LowStockAlert, error) {
	alertKey := fmt.Sprintf("%s:low_stock_alerts", "mall")

	// 获取预警列表
	alertsData, err := scs.cacheManager.LRange(alertKey, 0, limit-1)
	if err != nil {
		return nil, fmt.Errorf("获取低库存预警失败: %w", err)
	}

	var alerts []*LowStockAlert
	for _, data := range alertsData {
		if dataStr, ok := data.(string); ok {
			var alert LowStockAlert
			if err := json.Unmarshal([]byte(dataStr), &alert); err != nil {
				logger.Error(fmt.Sprintf("低库存预警数据反序列化失败: %v", err))
				continue
			}
			alerts = append(alerts, &alert)
		}
	}

	return alerts, nil
}

// ClearLowStockAlerts 清空低库存预警
func (scs *StockCacheService) ClearLowStockAlerts() error {
	alertKey := fmt.Sprintf("%s:low_stock_alerts", "mall")
	return scs.cacheManager.Delete(alertKey)
}

// ExistsStock 检查库存缓存是否存在
func (scs *StockCacheService) ExistsStock(productID uint) bool {
	key := scs.keyManager.GenerateProductStockKey(productID)
	return scs.cacheManager.Exists(key)
}

// GetStockTTL 获取库存缓存剩余TTL
func (scs *StockCacheService) GetStockTTL(productID uint) (time.Duration, error) {
	key := scs.keyManager.GenerateProductStockKey(productID)
	return scs.cacheManager.TTL(key)
}

// RefreshStockTTL 刷新库存缓存TTL
func (scs *StockCacheService) RefreshStockTTL(productID uint) error {
	key := scs.keyManager.GenerateProductStockKey(productID)
	ttl := GetTTL("stock")
	return scs.cacheManager.Expire(key, ttl)
}

// WarmupStocks 库存缓存预热
func (scs *StockCacheService) WarmupStocks(products []*model.Product) error {
	return scs.SetStocks(products)
}

// GetOutOfStockProducts 获取缺货商品列表
func (scs *StockCacheService) GetOutOfStockProducts() ([]uint, error) {
	// 使用Set存储缺货商品ID
	key := fmt.Sprintf("%s:out_of_stock_products", "mall")

	members, err := scs.cacheManager.SMembers(key)
	if err != nil {
		return nil, fmt.Errorf("获取缺货商品列表失败: %w", err)
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

// AddOutOfStockProduct 添加缺货商品
func (scs *StockCacheService) AddOutOfStockProduct(productID uint) error {
	key := fmt.Sprintf("%s:out_of_stock_products", "mall")
	return scs.cacheManager.SAdd(key, strconv.Itoa(int(productID)))
}

// RemoveOutOfStockProduct 移除缺货商品
func (scs *StockCacheService) RemoveOutOfStockProduct(productID uint) error {
	key := fmt.Sprintf("%s:out_of_stock_products", "mall")
	return scs.cacheManager.SRem(key, strconv.Itoa(int(productID)))
}

// GetStockStats 获取库存统计信息
func (scs *StockCacheService) GetStockStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// 获取缓存指标
	if metrics := scs.cacheManager.GetMetrics(); metrics != nil {
		stats["total_ops"] = metrics.TotalOps
		stats["hit_count"] = metrics.HitCount
		stats["miss_count"] = metrics.MissCount
		stats["hit_rate"] = metrics.HitRate
		stats["error_count"] = metrics.ErrorCount
		stats["last_updated"] = metrics.LastUpdated
	}

	// 获取连接池统计
	if connStats := scs.cacheManager.GetConnectionStats(); connStats != nil {
		stats["total_conns"] = connStats.TotalConns
		stats["idle_conns"] = connStats.IdleConns
		stats["hits"] = connStats.Hits
		stats["misses"] = connStats.Misses
	}

	// 获取低库存预警数量
	if alerts, err := scs.GetLowStockAlerts(100); err == nil {
		stats["low_stock_alerts"] = len(alerts)
	}

	// 获取缺货商品数量
	if outOfStockProducts, err := scs.GetOutOfStockProducts(); err == nil {
		stats["out_of_stock_count"] = len(outOfStockProducts)
	}

	return stats
}

// DeleteStocks 批量删除库存缓存
func (scs *StockCacheService) DeleteStocks(productIDs []uint) error {
	if len(productIDs) == 0 {
		return nil
	}

	keys := scs.keyManager.GenerateBatchKeys("stock", productIDs)

	if err := scs.cacheManager.MDelete(keys); err != nil {
		logger.Error(fmt.Sprintf("批量删除库存缓存失败: %v", err))
		return fmt.Errorf("批量删除库存缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("批量删除库存缓存成功: 数量=%d", len(productIDs)))
	return nil
}

// SyncStockFromDB 从数据库同步库存到缓存（用于缓存失效后的恢复）
func (scs *StockCacheService) SyncStockFromDB(productID uint, product *model.Product) error {
	// 设置库存缓存
	if err := scs.SetStock(product); err != nil {
		return fmt.Errorf("同步库存缓存失败: %w", err)
	}

	// 更新缺货状态
	if product.Stock <= 0 {
		scs.AddOutOfStockProduct(productID)
	} else {
		scs.RemoveOutOfStockProduct(productID)
	}

	// 检查低库存预警
	if product.Stock <= product.MinStock && product.MinStock > 0 {
		scs.addLowStockAlert(productID, product.Stock, product.MinStock)
	}

	logger.Info(fmt.Sprintf("库存同步完成: ProductID=%d, Stock=%d, Version=%d",
		productID, product.Stock, product.Version))
	return nil
}
