package service

import (
	"context"
	"mall-go/internal/model"
	"gorm.io/gorm"
)

// InventoryService 库存服务
type InventoryService struct {
	db *gorm.DB
}

// NewInventoryService 创建新的库存服务实例
func NewInventoryService(db *gorm.DB) *InventoryService {
	return &InventoryService{
		db: db,
	}
}

// CheckStock 检查库存
func (s *InventoryService) CheckStock(ctx context.Context, productID uint, quantity int) (bool, error) {
	var product model.Product
	if err := s.db.Where("id = ?", productID).First(&product).Error; err != nil {
		return false, err
	}
	
	return product.Stock >= quantity, nil
}

// DeductStock 扣减库存
func (s *InventoryService) DeductStock(ctx context.Context, productID uint, quantity int) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var product model.Product
		if err := tx.Where("id = ?", productID).First(&product).Error; err != nil {
			return err
		}
		
		if product.Stock < quantity {
			return model.ErrInvalidOperation
		}
		
		return tx.Model(&product).Update("stock", product.Stock-quantity).Error
	})
}

// RestoreStock 恢复库存
func (s *InventoryService) RestoreStock(ctx context.Context, productID uint, quantity int) error {
	return s.db.Model(&model.Product{}).
		Where("id = ?", productID).
		Update("stock", gorm.Expr("stock + ?", quantity)).Error
}