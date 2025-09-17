package cart

import (
	"fmt"
	"time"

	"mall-go/internal/model"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// SyncService 购物车商品同步服务
type SyncService struct {
	db           *gorm.DB
	cartService  *CartService
	cacheService *CacheService
}

// NewSyncService 创建购物车商品同步服务
func NewSyncService(db *gorm.DB, cartService *CartService, cacheService *CacheService) *SyncService {
	return &SyncService{
		db:           db,
		cartService:  cartService,
		cacheService: cacheService,
	}
}

// SyncResult 同步结果
type SyncResult struct {
	TotalItems      int                  `json:"total_items"`
	UpdatedItems    int                  `json:"updated_items"`
	InvalidItems    int                  `json:"invalid_items"`
	PriceChanges    []PriceChangeItem    `json:"price_changes"`
	StockIssues     []StockIssueItem     `json:"stock_issues"`
	InvalidProducts []InvalidProductItem `json:"invalid_products"`
	ProcessedCarts  int                  `json:"processed_carts"`
	Errors          []string             `json:"errors"`
}

// PriceChangeItem 价格变动商品
type PriceChangeItem struct {
	CartItemID  uint            `json:"cart_item_id"`
	ProductID   uint            `json:"product_id"`
	SKUID       uint            `json:"sku_id"`
	ProductName string          `json:"product_name"`
	OldPrice    decimal.Decimal `json:"old_price"`
	NewPrice    decimal.Decimal `json:"new_price"`
	Quantity    int             `json:"quantity"`
}

// StockIssueItem 库存问题商品
type StockIssueItem struct {
	CartItemID     uint   `json:"cart_item_id"`
	ProductID      uint   `json:"product_id"`
	SKUID          uint   `json:"sku_id"`
	ProductName    string `json:"product_name"`
	RequestedQty   int    `json:"requested_qty"`
	AvailableStock int    `json:"available_stock"`
}

// InvalidProductItem 失效商品
type InvalidProductItem struct {
	CartItemID  uint   `json:"cart_item_id"`
	ProductID   uint   `json:"product_id"`
	SKUID       uint   `json:"sku_id"`
	ProductName string `json:"product_name"`
	Reason      string `json:"reason"`
}

// SyncCartItems 同步购物车商品信息
func (ss *SyncService) SyncCartItems(userID uint, sessionID string) (*SyncResult, error) {
	result := &SyncResult{
		PriceChanges:    []PriceChangeItem{},
		StockIssues:     []StockIssueItem{},
		InvalidProducts: []InvalidProductItem{},
		Errors:          []string{},
	}

	// 获取购物车
	cart, err := ss.cartService.getCart(userID, sessionID)
	if err != nil {
		return result, fmt.Errorf("购物车不存在")
	}

	// 获取购物车商品项
	var cartItems []model.CartItem
	if err := ss.db.Where("cart_id = ?", cart.ID).Find(&cartItems).Error; err != nil {
		return result, fmt.Errorf("获取购物车商品失败: %v", err)
	}

	result.TotalItems = len(cartItems)

	// 开始事务
	tx := ss.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, item := range cartItems {
		if err := ss.syncCartItem(tx, &item, result); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("同步商品ID %d 失败: %v", item.ID, err))
		}
	}

	tx.Commit()
	result.ProcessedCarts = 1

	// 清除缓存
	if ss.cacheService != nil {
		ss.cacheService.clearCartCache(userID, sessionID)
	}

	return result, nil
}

// SyncAllCartItems 同步所有购物车商品信息
func (ss *SyncService) SyncAllCartItems() (*SyncResult, error) {
	result := &SyncResult{
		PriceChanges:    []PriceChangeItem{},
		StockIssues:     []StockIssueItem{},
		InvalidProducts: []InvalidProductItem{},
		Errors:          []string{},
	}

	// 获取所有活跃购物车
	var carts []model.Cart
	if err := ss.db.Where("status = ?", model.CartStatusActive).Find(&carts).Error; err != nil {
		return result, fmt.Errorf("获取购物车列表失败: %v", err)
	}

	for _, cart := range carts {
		cartResult, err := ss.SyncCartItems(cart.UserID, cart.SessionID)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("同步购物车ID %d 失败: %v", cart.ID, err))
			continue
		}

		// 合并结果
		result.TotalItems += cartResult.TotalItems
		result.UpdatedItems += cartResult.UpdatedItems
		result.InvalidItems += cartResult.InvalidItems
		result.PriceChanges = append(result.PriceChanges, cartResult.PriceChanges...)
		result.StockIssues = append(result.StockIssues, cartResult.StockIssues...)
		result.InvalidProducts = append(result.InvalidProducts, cartResult.InvalidProducts...)
		result.Errors = append(result.Errors, cartResult.Errors...)
		result.ProcessedCarts++
	}

	return result, nil
}

// syncCartItem 同步单个购物车商品项
func (ss *SyncService) syncCartItem(tx *gorm.DB, item *model.CartItem, result *SyncResult) error {
	updated := false
	_ = item.Status // 避免未使用变量警告

	// 检查商品状态
	var product model.Product
	if err := tx.Where("id = ?", item.ProductID).First(&product).Error; err != nil {
		// 商品不存在
		item.Status = model.CartItemStatusInvalid
		result.InvalidProducts = append(result.InvalidProducts, InvalidProductItem{
			CartItemID:  item.ID,
			ProductID:   item.ProductID,
			SKUID:       item.SKUID,
			ProductName: item.ProductName,
			Reason:      "商品已删除",
		})
		updated = true
	} else if product.Status != model.ProductStatusActive {
		// 商品已下架
		item.Status = model.CartItemStatusInvalid
		result.InvalidProducts = append(result.InvalidProducts, InvalidProductItem{
			CartItemID:  item.ID,
			ProductID:   item.ProductID,
			SKUID:       item.SKUID,
			ProductName: item.ProductName,
			Reason:      "商品已下架",
		})
		updated = true
	} else {
		// 商品正常，检查价格和库存
		currentPrice := product.Price
		availableStock := product.Stock

		// 如果有SKU，检查SKU状态
		if item.SKUID > 0 {
			var sku model.ProductSKU
			if err := tx.Where("id = ? AND product_id = ?", item.SKUID, item.ProductID).First(&sku).Error; err != nil {
				// SKU不存在
				item.Status = model.CartItemStatusInvalid
				result.InvalidProducts = append(result.InvalidProducts, InvalidProductItem{
					CartItemID:  item.ID,
					ProductID:   item.ProductID,
					SKUID:       item.SKUID,
					ProductName: item.ProductName,
					Reason:      "商品规格已删除",
				})
				updated = true
			} else if sku.Status != model.SKUStatusActive {
				// SKU已下架
				item.Status = model.CartItemStatusInvalid
				result.InvalidProducts = append(result.InvalidProducts, InvalidProductItem{
					CartItemID:  item.ID,
					ProductID:   item.ProductID,
					SKUID:       item.SKUID,
					ProductName: item.ProductName,
					Reason:      "商品规格已下架",
				})
				updated = true
			} else {
				currentPrice = sku.Price
				availableStock = sku.Stock
			}
		}

		// 如果商品/SKU正常，检查价格和库存
		if item.Status != model.CartItemStatusInvalid {
			// 检查价格变化
			if !item.Price.Equal(currentPrice) {
				result.PriceChanges = append(result.PriceChanges, PriceChangeItem{
					CartItemID:  item.ID,
					ProductID:   item.ProductID,
					SKUID:       item.SKUID,
					ProductName: item.ProductName,
					OldPrice:    item.Price,
					NewPrice:    currentPrice,
					Quantity:    item.Quantity,
				})

				// 更新价格并标记价格变动
				item.Price = currentPrice
				if item.Status == model.CartItemStatusNormal {
					item.Status = model.CartItemStatusPriceChange
				}
				updated = true
			}

			// 检查库存
			if availableStock < item.Quantity {
				result.StockIssues = append(result.StockIssues, StockIssueItem{
					CartItemID:     item.ID,
					ProductID:      item.ProductID,
					SKUID:          item.SKUID,
					ProductName:    item.ProductName,
					RequestedQty:   item.Quantity,
					AvailableStock: availableStock,
				})

				// 标记库存不足
				item.Status = model.CartItemStatusOutOfStock
				updated = true
			} else if item.Status == model.CartItemStatusOutOfStock {
				// 库存恢复正常
				item.Status = model.CartItemStatusNormal
				updated = true
			}

			// 如果价格没有变化且库存充足，恢复正常状态
			if item.Status == model.CartItemStatusPriceChange && availableStock >= item.Quantity {
				item.Status = model.CartItemStatusNormal
				updated = true
			}
		}
	}

	// 更新商品项
	if updated {
		item.UpdatedAt = time.Now()
		if err := tx.Save(item).Error; err != nil {
			return fmt.Errorf("更新购物车商品失败: %v", err)
		}

		result.UpdatedItems++
		if item.Status != model.CartItemStatusNormal {
			result.InvalidItems++
		}
	}

	return nil
}

// ValidateCartItems 验证购物车商品
func (ss *SyncService) ValidateCartItems(userID uint, sessionID string) (*model.CartResponse, error) {
	// 先同步商品信息
	_, err := ss.SyncCartItems(userID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("同步购物车商品失败: %v", err)
	}

	// 获取最新的购物车信息
	return ss.cartService.GetCart(userID, sessionID, true)
}

// CleanInvalidItems 清理失效商品
func (ss *SyncService) CleanInvalidItems(userID uint, sessionID string) error {
	cart, err := ss.cartService.getCart(userID, sessionID)
	if err != nil {
		return fmt.Errorf("购物车不存在")
	}

	// 删除失效商品
	if err := ss.db.Where("cart_id = ? AND status != ?", cart.ID, model.CartItemStatusNormal).Delete(&model.CartItem{}).Error; err != nil {
		return fmt.Errorf("清理失效商品失败: %v", err)
	}

	// 更新购物车统计
	ss.cartService.updateCartSummary(cart.ID)

	// 清除缓存
	if ss.cacheService != nil {
		ss.cacheService.clearCartCache(userID, sessionID)
	}

	return nil
}

// AutoFixStockIssues 自动修复库存问题
func (ss *SyncService) AutoFixStockIssues(userID uint, sessionID string) (*SyncResult, error) {
	result := &SyncResult{
		PriceChanges:    []PriceChangeItem{},
		StockIssues:     []StockIssueItem{},
		InvalidProducts: []InvalidProductItem{},
		Errors:          []string{},
	}

	cart, err := ss.cartService.getCart(userID, sessionID)
	if err != nil {
		return result, fmt.Errorf("购物车不存在")
	}

	// 获取库存不足的商品
	var cartItems []model.CartItem
	if err := ss.db.Where("cart_id = ? AND status = ?", cart.ID, model.CartItemStatusOutOfStock).Find(&cartItems).Error; err != nil {
		return result, fmt.Errorf("获取库存不足商品失败: %v", err)
	}

	result.TotalItems = len(cartItems)

	// 开始事务
	tx := ss.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, item := range cartItems {
		// 获取当前可用库存
		availableStock := 0
		if item.SKUID > 0 {
			var sku model.ProductSKU
			if err := tx.Where("id = ?", item.SKUID).First(&sku).Error; err == nil {
				availableStock = sku.Stock
			}
		} else {
			var product model.Product
			if err := tx.Where("id = ?", item.ProductID).First(&product).Error; err == nil {
				availableStock = product.Stock
			}
		}

		if availableStock > 0 {
			// 调整数量为可用库存
			oldQuantity := item.Quantity
			item.Quantity = availableStock
			item.Status = model.CartItemStatusNormal
			item.UpdatedAt = time.Now()

			if err := tx.Save(&item).Error; err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("修复商品ID %d 失败: %v", item.ID, err))
				continue
			}

			result.StockIssues = append(result.StockIssues, StockIssueItem{
				CartItemID:     item.ID,
				ProductID:      item.ProductID,
				SKUID:          item.SKUID,
				ProductName:    item.ProductName,
				RequestedQty:   oldQuantity,
				AvailableStock: availableStock,
			})

			result.UpdatedItems++
		} else {
			// 库存为0，标记为失效
			item.Status = model.CartItemStatusInvalid
			item.UpdatedAt = time.Now()

			if err := tx.Save(&item).Error; err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("标记商品ID %d 失效失败: %v", item.ID, err))
				continue
			}

			result.InvalidProducts = append(result.InvalidProducts, InvalidProductItem{
				CartItemID:  item.ID,
				ProductID:   item.ProductID,
				SKUID:       item.SKUID,
				ProductName: item.ProductName,
				Reason:      "库存为0",
			})

			result.InvalidItems++
		}
	}

	tx.Commit()

	// 更新购物车统计
	ss.cartService.updateCartSummary(cart.ID)

	// 清除缓存
	if ss.cacheService != nil {
		ss.cacheService.clearCartCache(userID, sessionID)
	}

	result.ProcessedCarts = 1
	return result, nil
}

// ScheduledSync 定时同步任务
func (ss *SyncService) ScheduledSync() error {
	// 同步所有购物车商品信息
	result, err := ss.SyncAllCartItems()
	if err != nil {
		return fmt.Errorf("定时同步失败: %v", err)
	}

	// 记录同步结果日志
	fmt.Printf("定时同步完成: 处理购物车 %d 个, 更新商品 %d 个, 失效商品 %d 个, 错误 %d 个\n",
		result.ProcessedCarts, result.UpdatedItems, result.InvalidItems, len(result.Errors))

	return nil
}

// GetSyncStats 获取同步统计信息
func (ss *SyncService) GetSyncStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 统计各状态的购物车商品数量
	var statusStats []struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}

	if err := ss.db.Model(&model.CartItem{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusStats).Error; err != nil {
		return nil, fmt.Errorf("获取状态统计失败: %v", err)
	}

	stats["status_stats"] = statusStats

	// 统计需要同步的商品数量
	var needSyncCount int64
	if err := ss.db.Model(&model.CartItem{}).
		Where("status != ?", model.CartItemStatusNormal).
		Count(&needSyncCount).Error; err != nil {
		return nil, fmt.Errorf("获取需要同步商品数量失败: %v", err)
	}

	stats["need_sync_count"] = needSyncCount

	return stats, nil
}

// 全局购物车同步服务实例
var globalSyncService *SyncService

// InitGlobalSyncService 初始化全局购物车同步服务
func InitGlobalSyncService(db *gorm.DB, cartService *CartService, cacheService *CacheService) {
	globalSyncService = NewSyncService(db, cartService, cacheService)
}

// GetGlobalSyncService 获取全局购物车同步服务
func GetGlobalSyncService() *SyncService {
	return globalSyncService
}
