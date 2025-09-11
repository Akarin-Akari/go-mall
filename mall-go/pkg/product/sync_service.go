package product

import (
	"fmt"
	"log"

	"mall-go/internal/model"

	"gorm.io/gorm"
)

// SyncService 数据同步服务 - 确保冗余字段一致性
type SyncService struct {
	db *gorm.DB
}

// NewSyncService 创建数据同步服务
func NewSyncService(db *gorm.DB) *SyncService {
	return &SyncService{db: db}
}

// SyncProductRedundantFields 同步商品冗余字段
func (ss *SyncService) SyncProductRedundantFields() error {
	log.Println("🔄 开始同步商品冗余字段...")

	// 同步分类名称
	if err := ss.syncCategoryNames(); err != nil {
		return fmt.Errorf("同步分类名称失败: %v", err)
	}

	// 同步品牌名称
	if err := ss.syncBrandNames(); err != nil {
		return fmt.Errorf("同步品牌名称失败: %v", err)
	}

	// 同步商家名称
	if err := ss.syncMerchantNames(); err != nil {
		return fmt.Errorf("同步商家名称失败: %v", err)
	}

	log.Println("✅ 商品冗余字段同步完成")
	return nil
}

// syncCategoryNames 同步分类名称
func (ss *SyncService) syncCategoryNames() error {
	result := ss.db.Exec(`
		UPDATE products 
		SET category_name = (
			SELECT name FROM categories WHERE id = products.category_id
		) 
		WHERE category_id IS NOT NULL
	`)
	
	if result.Error != nil {
		return result.Error
	}
	
	log.Printf("✅ 同步分类名称完成，影响 %d 条记录", result.RowsAffected)
	return nil
}

// syncBrandNames 同步品牌名称
func (ss *SyncService) syncBrandNames() error {
	result := ss.db.Exec(`
		UPDATE products 
		SET brand_name = (
			SELECT name FROM brands WHERE id = products.brand_id
		) 
		WHERE brand_id IS NOT NULL AND brand_id > 0
	`)
	
	if result.Error != nil {
		return result.Error
	}
	
	log.Printf("✅ 同步品牌名称完成，影响 %d 条记录", result.RowsAffected)
	return nil
}

// syncMerchantNames 同步商家名称
func (ss *SyncService) syncMerchantNames() error {
	result := ss.db.Exec(`
		UPDATE products 
		SET merchant_name = (
			SELECT username FROM users WHERE id = products.merchant_id
		) 
		WHERE merchant_id IS NOT NULL
	`)
	
	if result.Error != nil {
		return result.Error
	}
	
	log.Printf("✅ 同步商家名称完成，影响 %d 条记录", result.RowsAffected)
	return nil
}

// SyncSingleProduct 同步单个商品的冗余字段
func (ss *SyncService) SyncSingleProduct(productID uint) error {
	var product model.Product
	if err := ss.db.First(&product, productID).Error; err != nil {
		return fmt.Errorf("商品不存在: %v", err)
	}

	// 获取分类名称
	var category model.Category
	if product.CategoryID > 0 {
		if err := ss.db.First(&category, product.CategoryID).Error; err == nil {
			product.CategoryName = category.Name
		}
	}

	// 获取品牌名称
	var brand model.Brand
	if product.BrandID > 0 {
		if err := ss.db.First(&brand, product.BrandID).Error; err == nil {
			product.BrandName = brand.Name
		}
	}

	// 获取商家名称
	var merchant model.User
	if product.MerchantID > 0 {
		if err := ss.db.First(&merchant, product.MerchantID).Error; err == nil {
			product.MerchantName = merchant.Username
		}
	}

	// 更新商品
	if err := ss.db.Model(&product).Updates(map[string]interface{}{
		"category_name": product.CategoryName,
		"brand_name":    product.BrandName,
		"merchant_name": product.MerchantName,
	}).Error; err != nil {
		return fmt.Errorf("更新商品冗余字段失败: %v", err)
	}

	return nil
}

// OnCategoryUpdate 分类更新时的回调
func (ss *SyncService) OnCategoryUpdate(categoryID uint, newName string) error {
	result := ss.db.Model(&model.Product{}).
		Where("category_id = ?", categoryID).
		Update("category_name", newName)
	
	if result.Error != nil {
		return fmt.Errorf("更新商品分类名称失败: %v", result.Error)
	}
	
	log.Printf("✅ 分类更新同步完成，影响 %d 个商品", result.RowsAffected)
	return nil
}

// OnBrandUpdate 品牌更新时的回调
func (ss *SyncService) OnBrandUpdate(brandID uint, newName string) error {
	result := ss.db.Model(&model.Product{}).
		Where("brand_id = ?", brandID).
		Update("brand_name", newName)
	
	if result.Error != nil {
		return fmt.Errorf("更新商品品牌名称失败: %v", result.Error)
	}
	
	log.Printf("✅ 品牌更新同步完成，影响 %d 个商品", result.RowsAffected)
	return nil
}

// OnMerchantUpdate 商家更新时的回调
func (ss *SyncService) OnMerchantUpdate(merchantID uint, newUsername string) error {
	result := ss.db.Model(&model.Product{}).
		Where("merchant_id = ?", merchantID).
		Update("merchant_name", newUsername)
	
	if result.Error != nil {
		return fmt.Errorf("更新商品商家名称失败: %v", result.Error)
	}
	
	log.Printf("✅ 商家更新同步完成，影响 %d 个商品", result.RowsAffected)
	return nil
}

// ValidateRedundantData 验证冗余数据一致性
func (ss *SyncService) ValidateRedundantData() (map[string]int, error) {
	result := make(map[string]int)

	// 检查分类名称不一致的商品
	var categoryMismatch int64
	ss.db.Raw(`
		SELECT COUNT(*) FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.category_name != c.name OR (p.category_name IS NULL AND c.name IS NOT NULL)
	`).Scan(&categoryMismatch)
	result["category_mismatch"] = int(categoryMismatch)

	// 检查品牌名称不一致的商品
	var brandMismatch int64
	ss.db.Raw(`
		SELECT COUNT(*) FROM products p
		LEFT JOIN brands b ON p.brand_id = b.id
		WHERE p.brand_name != b.name OR (p.brand_name IS NULL AND b.name IS NOT NULL)
	`).Scan(&brandMismatch)
	result["brand_mismatch"] = int(brandMismatch)

	// 检查商家名称不一致的商品
	var merchantMismatch int64
	ss.db.Raw(`
		SELECT COUNT(*) FROM products p
		LEFT JOIN users u ON p.merchant_id = u.id
		WHERE p.merchant_name != u.username OR (p.merchant_name IS NULL AND u.username IS NOT NULL)
	`).Scan(&merchantMismatch)
	result["merchant_mismatch"] = int(merchantMismatch)

	return result, nil
}

// AutoSyncScheduler 自动同步调度器（可以配合定时任务使用）
func (ss *SyncService) AutoSyncScheduler() error {
	log.Println("🔄 开始自动同步调度...")

	// 验证数据一致性
	validation, err := ss.ValidateRedundantData()
	if err != nil {
		return fmt.Errorf("验证数据一致性失败: %v", err)
	}

	// 如果有不一致的数据，执行同步
	totalMismatch := validation["category_mismatch"] + validation["brand_mismatch"] + validation["merchant_mismatch"]
	if totalMismatch > 0 {
		log.Printf("⚠️  发现 %d 条不一致数据，开始同步...", totalMismatch)
		if err := ss.SyncProductRedundantFields(); err != nil {
			return fmt.Errorf("自动同步失败: %v", err)
		}
	} else {
		log.Println("✅ 数据一致性验证通过，无需同步")
	}

	return nil
}
