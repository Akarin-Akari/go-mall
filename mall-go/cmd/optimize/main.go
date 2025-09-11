package main

import (
	"fmt"
	"log"
	"os"

	"mall-go/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseOptimizer 数据库性能优化器
type DatabaseOptimizer struct {
	db *gorm.DB
}

// NewDatabaseOptimizer 创建数据库优化器
func NewDatabaseOptimizer(db *gorm.DB) *DatabaseOptimizer {
	return &DatabaseOptimizer{db: db}
}

// MigrateModels 迁移所有模型
func (do *DatabaseOptimizer) MigrateModels() error {
	fmt.Println("🚀 开始数据库模型迁移...")

	// 迁移所有模型
	err := do.db.AutoMigrate(
		&model.User{},
		&model.Category{},
		&model.Brand{},
		&model.Product{},
		&model.ProductImage{},
		&model.ProductAttr{},
		&model.ProductSKU{},
		&model.ProductReview{},
	)

	if err != nil {
		return fmt.Errorf("模型迁移失败: %v", err)
	}

	fmt.Println("✅ 数据库模型迁移完成")
	return nil
}

// CreateIndexes 创建性能索引
func (do *DatabaseOptimizer) CreateIndexes() error {
	fmt.Println("📊 开始创建性能索引...")

	// 商品表索引
	indexes := []struct {
		table string
		name  string
		sql   string
	}{
		{"products", "idx_products_category_brand", "CREATE INDEX IF NOT EXISTS idx_products_category_brand ON products(category_name, brand_name)"},
		{"products", "idx_products_price_status", "CREATE INDEX IF NOT EXISTS idx_products_price_status ON products(price, status)"},
		{"products", "idx_products_merchant_status", "CREATE INDEX IF NOT EXISTS idx_products_merchant_status ON products(merchant_id, status)"},
		{"products", "idx_products_stock_status", "CREATE INDEX IF NOT EXISTS idx_products_stock_status ON products(stock, status)"},
		{"products", "idx_products_sold_count_desc", "CREATE INDEX IF NOT EXISTS idx_products_sold_count_desc ON products(sold_count DESC)"},
		{"products", "idx_products_category_price", "CREATE INDEX IF NOT EXISTS idx_products_category_price ON products(category_id, price)"},
		{"products", "idx_products_brand_price", "CREATE INDEX IF NOT EXISTS idx_products_brand_price ON products(brand_id, price)"},
		{"products", "idx_products_status_hot", "CREATE INDEX IF NOT EXISTS idx_products_status_hot ON products(status, is_hot)"},
		{"products", "idx_products_status_new", "CREATE INDEX IF NOT EXISTS idx_products_status_new ON products(status, is_new)"},
		{"products", "idx_products_status_recommend", "CREATE INDEX IF NOT EXISTS idx_products_status_recommend ON products(status, is_recommend)"},
		{"products", "idx_products_sort_id", "CREATE INDEX IF NOT EXISTS idx_products_sort_id ON products(sort ASC, id DESC)"},
		{"products", "idx_products_created_desc", "CREATE INDEX IF NOT EXISTS idx_products_created_desc ON products(created_at DESC)"},
		{"products", "idx_products_view_count_desc", "CREATE INDEX IF NOT EXISTS idx_products_view_count_desc ON products(view_count DESC)"},
	}

	for _, idx := range indexes {
		if err := do.db.Exec(idx.sql).Error; err != nil {
			fmt.Printf("⚠️  创建索引 %s 失败: %v\n", idx.name, err)
		} else {
			fmt.Printf("✅ 创建索引 %s 成功\n", idx.name)
		}
	}

	fmt.Println("✅ 性能索引创建完成")
	return nil
}

// CreateTestData 创建测试数据
func (do *DatabaseOptimizer) CreateTestData() error {
	fmt.Println("🧪 开始创建测试数据...")

	// 创建测试分类
	categories := []model.Category{
		{Name: "电子产品", Description: "各类电子设备", Status: "active"},
		{Name: "服装鞋帽", Description: "时尚服饰", Status: "active"},
		{Name: "家居用品", Description: "家庭生活用品", Status: "active"},
		{Name: "运动户外", Description: "运动健身用品", Status: "active"},
		{Name: "美妆护肤", Description: "美容护肤产品", Status: "active"},
	}

	for _, category := range categories {
		var existingCategory model.Category
		if err := do.db.Where("name = ?", category.Name).First(&existingCategory).Error; err != nil {
			if err := do.db.Create(&category).Error; err != nil {
				return fmt.Errorf("创建分类失败: %v", err)
			}
		}
	}

	// 创建测试品牌
	brands := []model.Brand{
		{Name: "Apple", Description: "苹果公司", Status: "active"},
		{Name: "Nike", Description: "耐克运动品牌", Status: "active"},
		{Name: "Samsung", Description: "三星电子", Status: "active"},
		{Name: "Adidas", Description: "阿迪达斯", Status: "active"},
		{Name: "Huawei", Description: "华为技术", Status: "active"},
	}

	for _, brand := range brands {
		var existingBrand model.Brand
		if err := do.db.Where("name = ?", brand.Name).First(&existingBrand).Error; err != nil {
			if err := do.db.Create(&brand).Error; err != nil {
				return fmt.Errorf("创建品牌失败: %v", err)
			}
		}
	}

	// 创建测试商家用户
	merchants := []model.User{
		{Username: "merchant1", Email: "merchant1@example.com", Role: "merchant", Status: "active"},
		{Username: "merchant2", Email: "merchant2@example.com", Role: "merchant", Status: "active"},
		{Username: "merchant3", Email: "merchant3@example.com", Role: "merchant", Status: "active"},
	}

	for _, merchant := range merchants {
		var existingUser model.User
		if err := do.db.Where("username = ?", merchant.Username).First(&existingUser).Error; err != nil {
			if err := do.db.Create(&merchant).Error; err != nil {
				return fmt.Errorf("创建商家用户失败: %v", err)
			}
		}
	}

	fmt.Println("✅ 测试数据创建完成")
	return nil
}

// SyncRedundantData 同步冗余数据
func (do *DatabaseOptimizer) SyncRedundantData() error {
	fmt.Println("🔄 开始同步冗余数据...")

	// 更新商品的分类名称
	result := do.db.Exec(`
		UPDATE products 
		SET category_name = (
			SELECT name FROM categories WHERE id = products.category_id
		) 
		WHERE category_id IS NOT NULL
	`)
	if result.Error != nil {
		return fmt.Errorf("同步分类名称失败: %v", result.Error)
	}
	fmt.Printf("✅ 同步分类名称完成，影响 %d 条记录\n", result.RowsAffected)

	// 更新商品的品牌名称
	result = do.db.Exec(`
		UPDATE products 
		SET brand_name = (
			SELECT name FROM brands WHERE id = products.brand_id
		) 
		WHERE brand_id IS NOT NULL AND brand_id > 0
	`)
	if result.Error != nil {
		return fmt.Errorf("同步品牌名称失败: %v", result.Error)
	}
	fmt.Printf("✅ 同步品牌名称完成，影响 %d 条记录\n", result.RowsAffected)

	// 更新商品的商家名称
	result = do.db.Exec(`
		UPDATE products 
		SET merchant_name = (
			SELECT username FROM users WHERE id = products.merchant_id
		) 
		WHERE merchant_id IS NOT NULL
	`)
	if result.Error != nil {
		return fmt.Errorf("同步商家名称失败: %v", result.Error)
	}
	fmt.Printf("✅ 同步商家名称完成，影响 %d 条记录\n", result.RowsAffected)

	fmt.Println("✅ 冗余数据同步完成")
	return nil
}

// RunOptimization 执行完整优化
func (do *DatabaseOptimizer) RunOptimization() error {
	fmt.Println("🎯 开始Mall-Go数据模型性能优化...")

	// 1. 迁移模型
	if err := do.MigrateModels(); err != nil {
		return err
	}

	// 2. 创建性能索引
	if err := do.CreateIndexes(); err != nil {
		return err
	}

	// 3. 创建测试数据
	if err := do.CreateTestData(); err != nil {
		return err
	}

	// 4. 同步冗余数据
	if err := do.SyncRedundantData(); err != nil {
		return err
	}

	fmt.Println("🎉 Mall-Go数据模型性能优化完成！")
	return nil
}

func main() {
	// 连接数据库
	db, err := gorm.Open(sqlite.Open("mall_optimized.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	// 创建优化器
	optimizer := NewDatabaseOptimizer(db)

	// 执行优化
	if err := optimizer.RunOptimization(); err != nil {
		log.Fatal("优化失败:", err)
		os.Exit(1)
	}

	fmt.Println("✨ 数据库性能优化完成，可以开始性能测试！")
}
