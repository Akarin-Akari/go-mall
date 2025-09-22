package inventory

import (
	"context"
	"fmt"
	"time"

	"mall-go/internal/model"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// InventoryService 库存管理服务
type InventoryService struct {
	db  *gorm.DB
	rdb *redis.Client
	ctx context.Context
}

// NewInventoryService 创建库存管理服务
func NewInventoryService(db *gorm.DB, rdb *redis.Client) *InventoryService {
	return &InventoryService{
		db:  db,
		rdb: rdb,
		ctx: context.Background(),
	}
}

// StockDeductionRequest 库存扣减请求
type StockDeductionRequest struct {
	ProductID uint `json:"product_id"`
	SKUID     uint `json:"sku_id,omitempty"`
	Quantity  int  `json:"quantity"`
}

// StockDeductionResult 库存扣减结果
type StockDeductionResult struct {
	Success   bool   `json:"success"`
	ProductID uint   `json:"product_id"`
	SKUID     uint   `json:"sku_id,omitempty"`
	Quantity  int    `json:"quantity"`
	Error     string `json:"error,omitempty"`
}

// DeductStockWithOptimisticLock 使用乐观锁扣减库存（优化版 - 移除Redis锁，避免双重事务）
func (is *InventoryService) DeductStockWithOptimisticLock(requests []StockDeductionRequest) ([]StockDeductionResult, error) {
	results := make([]StockDeductionResult, len(requests))

	// 直接使用数据库乐观锁，不使用Redis分布式锁，避免双重锁定
	for i, req := range requests {
		result := StockDeductionResult{
			ProductID: req.ProductID,
			SKUID:     req.SKUID,
			Quantity:  req.Quantity,
		}

		var err error
		if req.SKUID > 0 {
			// 扣减SKU库存 - 使用独立事务
			err = is.deductSKUStockWithRetry(is.db, req.SKUID, req.Quantity, 3)
		} else {
			// 扣减商品库存 - 使用独立事务
			err = is.deductProductStockWithRetry(is.db, req.ProductID, req.Quantity, 3)
		}

		if err != nil {
			result.Success = false
			result.Error = err.Error()
			// 如果某个商品扣减失败，需要回滚之前成功的扣减
			if i > 0 {
				is.rollbackPreviousDeductions(requests[:i])
			}
			return results, fmt.Errorf("库存扣减失败: %v", err)
		}

		result.Success = true
		results[i] = result
	}

	return results, nil
}

// rollbackPreviousDeductions 回滚之前成功的库存扣减
func (is *InventoryService) rollbackPreviousDeductions(requests []StockDeductionRequest) {
	for _, req := range requests {
		if req.SKUID > 0 {
			is.restoreSKUStock(is.db, req.SKUID, req.Quantity)
		} else {
			is.restoreProductStock(is.db, req.ProductID, req.Quantity)
		}
	}
}

// deductProductStockWithRetry 使用乐观锁扣减商品库存（带重试）- 优化版，使用独立事务
func (is *InventoryService) deductProductStockWithRetry(db *gorm.DB, productID uint, quantity int, maxRetries int) error {
	for retries := 0; retries < maxRetries; retries++ {
		// 使用独立的短事务，减少锁持有时间
		err := db.Transaction(func(tx *gorm.DB) error {
			// 获取当前商品信息
			var product model.Product
			if err := tx.Where("id = ?", productID).First(&product).Error; err != nil {
				return fmt.Errorf("商品不存在: %v", err)
			}

			// 检查库存是否足够
			if product.Stock < quantity {
				return fmt.Errorf("商品库存不足，当前库存：%d，需要：%d", product.Stock, quantity)
			}

			// 使用乐观锁更新库存
			result := tx.Model(&product).
				Where("id = ? AND version = ?", product.ID, product.Version).
				Updates(map[string]interface{}{
					"stock":      product.Stock - quantity,
					"sold_count": product.SoldCount + quantity,
					"version":    product.Version + 1,
					"updated_at": time.Now(),
				})

			if result.Error != nil {
				return fmt.Errorf("更新商品库存失败: %v", result.Error)
			}

			// 检查是否更新成功
			if result.RowsAffected == 0 {
				return fmt.Errorf("版本冲突，需要重试")
			}

			return nil
		})

		// 如果成功，直接返回
		if err == nil {
			return nil
		}

		// 如果是版本冲突，进行重试
		if retries < maxRetries-1 && (err.Error() == "版本冲突，需要重试") {
			// 短暂等待后重试，使用指数退避
			time.Sleep(time.Millisecond * time.Duration(5*(1<<retries)))
			continue
		}

		// 其他错误或达到最大重试次数，直接返回错误
		return err
	}

	return fmt.Errorf("库存更新失败，超过最大重试次数")
}

// deductSKUStockWithRetry 使用乐观锁扣减SKU库存（带重试）
func (is *InventoryService) deductSKUStockWithRetry(tx *gorm.DB, skuID uint, quantity int, maxRetries int) error {
	for retries := 0; retries < maxRetries; retries++ {
		// 获取当前SKU信息
		var sku model.ProductSKU
		if err := tx.Where("id = ?", skuID).First(&sku).Error; err != nil {
			return fmt.Errorf("SKU不存在: %v", err)
		}

		// 检查库存是否足够
		if sku.Stock < quantity {
			return fmt.Errorf("SKU库存不足，当前库存：%d，需要：%d", sku.Stock, quantity)
		}

		// 使用乐观锁更新SKU库存
		result := tx.Model(&sku).
			Where("id = ? AND version = ?", sku.ID, sku.Version).
			Updates(map[string]interface{}{
				"stock":      sku.Stock - quantity,
				"version":    sku.Version + 1,
				"updated_at": time.Now(),
			})

		if result.Error != nil {
			return fmt.Errorf("更新SKU库存失败: %v", result.Error)
		}

		// 更新成功
		if result.RowsAffected > 0 {
			// 同时更新商品的销售数量
			tx.Model(&model.Product{}).
				Where("id = ?", sku.ProductID).
				UpdateColumn("sold_count", gorm.Expr("sold_count + ?", quantity))
			return nil
		}

		// 更新失败，说明版本号已变化，需要重试
		if retries == maxRetries-1 {
			return fmt.Errorf("SKU库存更新失败，并发冲突过多，请重试")
		}

		// 短暂等待后重试
		time.Sleep(time.Millisecond * time.Duration(10*(retries+1)))
	}

	return fmt.Errorf("SKU库存更新失败，超过最大重试次数")
}

// RestoreStock 恢复库存（用于订单取消等场景）
func (is *InventoryService) RestoreStock(requests []StockDeductionRequest) error {
	// 开始事务
	tx := is.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, req := range requests {
		if req.SKUID > 0 {
			// 恢复SKU库存
			if err := is.restoreSKUStock(tx, req.SKUID, req.Quantity); err != nil {
				tx.Rollback()
				return fmt.Errorf("恢复SKU库存失败: %v", err)
			}
		} else {
			// 恢复商品库存
			if err := is.restoreProductStock(tx, req.ProductID, req.Quantity); err != nil {
				tx.Rollback()
				return fmt.Errorf("恢复商品库存失败: %v", err)
			}
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// restoreProductStock 恢复商品库存
func (is *InventoryService) restoreProductStock(tx *gorm.DB, productID uint, quantity int) error {
	result := tx.Model(&model.Product{}).
		Where("id = ?", productID).
		Updates(map[string]interface{}{
			"stock":      gorm.Expr("stock + ?", quantity),
			"sold_count": gorm.Expr("GREATEST(sold_count - ?, 0)", quantity),
			"version":    gorm.Expr("version + 1"),
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("恢复商品库存失败: %v", result.Error)
	}

	return nil
}

// restoreSKUStock 恢复SKU库存
func (is *InventoryService) restoreSKUStock(tx *gorm.DB, skuID uint, quantity int) error {
	// 获取SKU信息
	var sku model.ProductSKU
	if err := tx.Where("id = ?", skuID).First(&sku).Error; err != nil {
		return fmt.Errorf("SKU不存在: %v", err)
	}

	// 恢复SKU库存
	result := tx.Model(&sku).
		Where("id = ?", skuID).
		Updates(map[string]interface{}{
			"stock":      gorm.Expr("stock + ?", quantity),
			"version":    gorm.Expr("version + 1"),
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("恢复SKU库存失败: %v", result.Error)
	}

	// 同时更新商品的销售数量
	tx.Model(&model.Product{}).
		Where("id = ?", sku.ProductID).
		UpdateColumn("sold_count", gorm.Expr("GREATEST(sold_count - ?, 0)", quantity))

	return nil
}

// CheckStock 检查库存是否充足
func (is *InventoryService) CheckStock(requests []StockDeductionRequest) (bool, error) {
	for _, req := range requests {
		if req.SKUID > 0 {
			// 检查SKU库存
			var sku model.ProductSKU
			if err := is.db.Where("id = ?", req.SKUID).First(&sku).Error; err != nil {
				return false, fmt.Errorf("SKU不存在: %v", err)
			}
			if sku.Stock < req.Quantity {
				return false, fmt.Errorf("SKU库存不足，当前库存：%d，需要：%d", sku.Stock, req.Quantity)
			}
		} else {
			// 检查商品库存
			var product model.Product
			if err := is.db.Where("id = ?", req.ProductID).First(&product).Error; err != nil {
				return false, fmt.Errorf("商品不存在: %v", err)
			}
			if product.Stock < req.Quantity {
				return false, fmt.Errorf("商品库存不足，当前库存：%d，需要：%d", product.Stock, req.Quantity)
			}
		}
	}
	return true, nil
}

// GetStockInfo 获取库存信息
func (is *InventoryService) GetStockInfo(productID uint, skuID uint) (int, error) {
	if skuID > 0 {
		var sku model.ProductSKU
		if err := is.db.Where("id = ?", skuID).First(&sku).Error; err != nil {
			return 0, fmt.Errorf("SKU不存在: %v", err)
		}
		return sku.Stock, nil
	}

	var product model.Product
	if err := is.db.Where("id = ?", productID).First(&product).Error; err != nil {
		return 0, fmt.Errorf("商品不存在: %v", err)
	}
	return product.Stock, nil
}

// releaseLocks 释放分布式锁
func (is *InventoryService) releaseLocks(lockKeys []string, lockValues []string) {
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`

	for i, lockKey := range lockKeys {
		if i < len(lockValues) {
			is.rdb.Eval(is.ctx, script, []string{lockKey}, lockValues[i])
		}
	}
}

// 全局库存服务实例
var globalInventoryService *InventoryService

// InitGlobalInventoryService 初始化全局库存服务
func InitGlobalInventoryService(db *gorm.DB, rdb *redis.Client) {
	globalInventoryService = NewInventoryService(db, rdb)
}

// GetGlobalInventoryService 获取全局库存服务
func GetGlobalInventoryService() *InventoryService {
	return globalInventoryService
}
