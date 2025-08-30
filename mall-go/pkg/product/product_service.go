package product

import (
	"fmt"

	"mall-go/internal/model"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ProductService 商品服务
type ProductService struct {
	db *gorm.DB
}

// NewProductService 创建商品服务
func NewProductService(db *gorm.DB) *ProductService {
	return &ProductService{
		db: db,
	}
}

// CreateProductRequest 创建商品请求
type CreateProductRequest struct {
	Name           string                    `json:"name" binding:"required,min=1,max=255"`
	SubTitle       string                    `json:"sub_title"`
	Description    string                    `json:"description"`
	Detail         string                    `json:"detail"`
	CategoryID     uint                      `json:"category_id" binding:"required"`
	BrandID        uint                      `json:"brand_id"`
	MerchantID     uint                      `json:"merchant_id" binding:"required"`
	Price          decimal.Decimal           `json:"price" binding:"required"`
	OriginPrice    decimal.Decimal           `json:"origin_price"`
	CostPrice      decimal.Decimal           `json:"cost_price"`
	Stock          int                       `json:"stock" binding:"min=0"`
	MinStock       int                       `json:"min_stock"`
	MaxStock       int                       `json:"max_stock"`
	Weight         decimal.Decimal           `json:"weight"`
	Volume         decimal.Decimal           `json:"volume"`
	Unit           string                    `json:"unit"`
	SEOTitle       string                    `json:"seo_title"`
	SEOKeywords    string                    `json:"seo_keywords"`
	SEODescription string                    `json:"seo_description"`
	Sort           int                       `json:"sort"`
	Images         []string                  `json:"images"`
	Attributes     []ProductAttributeRequest `json:"attributes"`
}

// UpdateProductRequest 更新商品请求
type UpdateProductRequest struct {
	Name           string                    `json:"name" binding:"required,min=1,max=255"`
	SubTitle       string                    `json:"sub_title"`
	Description    string                    `json:"description"`
	Detail         string                    `json:"detail"`
	CategoryID     uint                      `json:"category_id" binding:"required"`
	BrandID        uint                      `json:"brand_id"`
	Price          decimal.Decimal           `json:"price" binding:"required"`
	OriginPrice    decimal.Decimal           `json:"origin_price"`
	CostPrice      decimal.Decimal           `json:"cost_price"`
	Stock          int                       `json:"stock" binding:"min=0"`
	MinStock       int                       `json:"min_stock"`
	MaxStock       int                       `json:"max_stock"`
	Weight         decimal.Decimal           `json:"weight"`
	Volume         decimal.Decimal           `json:"volume"`
	Unit           string                    `json:"unit"`
	Status         string                    `json:"status"`
	IsHot          bool                      `json:"is_hot"`
	IsNew          bool                      `json:"is_new"`
	IsRecommend    bool                      `json:"is_recommend"`
	SEOTitle       string                    `json:"seo_title"`
	SEOKeywords    string                    `json:"seo_keywords"`
	SEODescription string                    `json:"seo_description"`
	Sort           int                       `json:"sort"`
	Images         []string                  `json:"images"`
	Attributes     []ProductAttributeRequest `json:"attributes"`
}

// ProductAttributeRequest 商品属性请求
type ProductAttributeRequest struct {
	AttrName  string `json:"attr_name" binding:"required"`
	AttrValue string `json:"attr_value" binding:"required"`
	Sort      int    `json:"sort"`
}

// ProductListRequest 商品列表请求
type ProductListRequest struct {
	Page        int     `form:"page" binding:"min=1"`
	PageSize    int     `form:"page_size" binding:"min=1,max=100"`
	CategoryID  *uint   `form:"category_id"`
	BrandID     *uint   `form:"brand_id"`
	MerchantID  *uint   `form:"merchant_id"`
	Keyword     string  `form:"keyword"`
	Status      string  `form:"status"`
	IsHot       *bool   `form:"is_hot"`
	IsNew       *bool   `form:"is_new"`
	IsRecommend *bool   `form:"is_recommend"`
	MinPrice    *string `form:"min_price"`
	MaxPrice    *string `form:"max_price"`
	SortBy      string  `form:"sort_by"` // price_asc, price_desc, sales_desc, created_desc
}

// ProductSearchRequest 商品搜索请求
type ProductSearchRequest struct {
	Keyword    string   `form:"keyword" binding:"required"`
	CategoryID *uint    `form:"category_id"`
	BrandID    *uint    `form:"brand_id"`
	MinPrice   *string  `form:"min_price"`
	MaxPrice   *string  `form:"max_price"`
	Tags       []string `form:"tags"`
	Page       int      `form:"page" binding:"min=1"`
	PageSize   int      `form:"page_size" binding:"min=1,max=100"`
	SortBy     string   `form:"sort_by"`
}

// CreateProduct 创建商品
func (ps *ProductService) CreateProduct(req *CreateProductRequest) (*model.Product, error) {
	// 验证分类是否存在
	var category model.Category
	if err := ps.db.First(&category, req.CategoryID).Error; err != nil {
		return nil, fmt.Errorf("分类不存在")
	}

	// 验证品牌是否存在（如果提供了品牌ID）
	if req.BrandID > 0 {
		var brand model.Brand
		if err := ps.db.First(&brand, req.BrandID).Error; err != nil {
			return nil, fmt.Errorf("品牌不存在")
		}
	}

	// 验证商家是否存在
	var merchant model.User
	if err := ps.db.First(&merchant, req.MerchantID).Error; err != nil {
		return nil, fmt.Errorf("商家不存在")
	}

	// 开始事务
	tx := ps.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建商品
	product := &model.Product{
		Name:           req.Name,
		SubTitle:       req.SubTitle,
		Description:    req.Description,
		Detail:         req.Detail,
		CategoryID:     req.CategoryID,
		BrandID:        req.BrandID,
		MerchantID:     req.MerchantID,
		Price:          req.Price,
		OriginPrice:    req.OriginPrice,
		CostPrice:      req.CostPrice,
		Stock:          req.Stock,
		MinStock:       req.MinStock,
		MaxStock:       req.MaxStock,
		Weight:         req.Weight,
		Volume:         req.Volume,
		Unit:           req.Unit,
		Status:         model.ProductStatusDraft,
		SEOTitle:       req.SEOTitle,
		SEOKeywords:    req.SEOKeywords,
		SEODescription: req.SEODescription,
		Sort:           req.Sort,
	}

	if err := tx.Create(product).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建商品失败: %v", err)
	}

	// 创建商品图片
	if len(req.Images) > 0 {
		for i, imageURL := range req.Images {
			image := &model.ProductImage{
				ProductID: product.ID,
				URL:       imageURL,
				Sort:      i,
				IsMain:    i == 0, // 第一张图片设为主图
			}
			if err := tx.Create(image).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("创建商品图片失败: %v", err)
			}
		}
	}

	// 创建商品属性
	if len(req.Attributes) > 0 {
		for _, attrReq := range req.Attributes {
			attr := &model.ProductAttr{
				ProductID: product.ID,
				AttrName:  attrReq.AttrName,
				AttrValue: attrReq.AttrValue,
				Sort:      attrReq.Sort,
			}
			if err := tx.Create(attr).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("创建商品属性失败: %v", err)
			}
		}
	}

	tx.Commit()

	// 重新查询商品（包含关联数据）
	return ps.GetProduct(product.ID)
}

// UpdateProduct 更新商品
func (ps *ProductService) UpdateProduct(id uint, req *UpdateProductRequest) (*model.Product, error) {
	var product model.Product
	if err := ps.db.First(&product, id).Error; err != nil {
		return nil, fmt.Errorf("商品不存在")
	}

	// 验证分类是否存在
	var category model.Category
	if err := ps.db.First(&category, req.CategoryID).Error; err != nil {
		return nil, fmt.Errorf("分类不存在")
	}

	// 验证品牌是否存在（如果提供了品牌ID）
	if req.BrandID > 0 {
		var brand model.Brand
		if err := ps.db.First(&brand, req.BrandID).Error; err != nil {
			return nil, fmt.Errorf("品牌不存在")
		}
	}

	// 开始事务
	tx := ps.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新商品基本信息
	product.Name = req.Name
	product.SubTitle = req.SubTitle
	product.Description = req.Description
	product.Detail = req.Detail
	product.CategoryID = req.CategoryID
	product.BrandID = req.BrandID
	product.Price = req.Price
	product.OriginPrice = req.OriginPrice
	product.CostPrice = req.CostPrice
	product.Stock = req.Stock
	product.MinStock = req.MinStock
	product.MaxStock = req.MaxStock
	product.Weight = req.Weight
	product.Volume = req.Volume
	product.Unit = req.Unit
	product.IsHot = req.IsHot
	product.IsNew = req.IsNew
	product.IsRecommend = req.IsRecommend
	product.SEOTitle = req.SEOTitle
	product.SEOKeywords = req.SEOKeywords
	product.SEODescription = req.SEODescription
	product.Sort = req.Sort

	if req.Status != "" {
		product.Status = req.Status
	}

	if err := tx.Save(&product).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("更新商品失败: %v", err)
	}

	// 更新商品图片
	if len(req.Images) > 0 {
		// 删除原有图片
		if err := tx.Where("product_id = ?", product.ID).Delete(&model.ProductImage{}).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("删除原有图片失败: %v", err)
		}

		// 创建新图片
		for i, imageURL := range req.Images {
			image := &model.ProductImage{
				ProductID: product.ID,
				URL:       imageURL,
				Sort:      i,
				IsMain:    i == 0,
			}
			if err := tx.Create(image).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("创建商品图片失败: %v", err)
			}
		}
	}

	// 更新商品属性
	if len(req.Attributes) > 0 {
		// 删除原有属性
		if err := tx.Where("product_id = ?", product.ID).Delete(&model.ProductAttr{}).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("删除原有属性失败: %v", err)
		}

		// 创建新属性
		for _, attrReq := range req.Attributes {
			attr := &model.ProductAttr{
				ProductID: product.ID,
				AttrName:  attrReq.AttrName,
				AttrValue: attrReq.AttrValue,
				Sort:      attrReq.Sort,
			}
			if err := tx.Create(attr).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("创建商品属性失败: %v", err)
			}
		}
	}

	tx.Commit()

	// 重新查询商品（包含关联数据）
	return ps.GetProduct(product.ID)
}

// DeleteProduct 删除商品
func (ps *ProductService) DeleteProduct(id uint) error {
	var product model.Product
	if err := ps.db.First(&product, id).Error; err != nil {
		return fmt.Errorf("商品不存在")
	}

	// 软删除商品
	if err := ps.db.Delete(&product).Error; err != nil {
		return fmt.Errorf("删除商品失败: %v", err)
	}

	return nil
}

// GetProduct 获取商品详情
func (ps *ProductService) GetProduct(id uint) (*model.Product, error) {
	var product model.Product
	if err := ps.db.Preload("Category").
		Preload("Brand").
		Preload("Merchant").
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort ASC")
		}).
		Preload("Attributes", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort ASC")
		}).
		Preload("SKUs", func(db *gorm.DB) *gorm.DB {
			return db.Where("status = ?", model.SKUStatusActive)
		}).
		First(&product, id).Error; err != nil {
		return nil, fmt.Errorf("商品不存在")
	}

	// 增加浏览次数
	ps.db.Model(&product).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1))

	return &product, nil
}

// GetProductList 获取商品列表
func (ps *ProductService) GetProductList(req *ProductListRequest) ([]*model.Product, int64, error) {
	query := ps.db.Model(&model.Product{})

	// 条件筛选
	if req.CategoryID != nil {
		query = query.Where("category_id = ?", *req.CategoryID)
	}

	if req.BrandID != nil {
		query = query.Where("brand_id = ?", *req.BrandID)
	}

	if req.MerchantID != nil {
		query = query.Where("merchant_id = ?", *req.MerchantID)
	}

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	if req.IsHot != nil {
		query = query.Where("is_hot = ?", *req.IsHot)
	}

	if req.IsNew != nil {
		query = query.Where("is_new = ?", *req.IsNew)
	}

	if req.IsRecommend != nil {
		query = query.Where("is_recommend = ?", *req.IsRecommend)
	}

	if req.Keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	// 价格范围筛选
	if req.MinPrice != nil && *req.MinPrice != "" {
		if minPrice, err := decimal.NewFromString(*req.MinPrice); err == nil {
			query = query.Where("price >= ?", minPrice)
		}
	}

	if req.MaxPrice != nil && *req.MaxPrice != "" {
		if maxPrice, err := decimal.NewFromString(*req.MaxPrice); err == nil {
			query = query.Where("price <= ?", maxPrice)
		}
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询商品总数失败: %v", err)
	}

	// 排序
	orderBy := "sort ASC, id DESC"
	switch req.SortBy {
	case "price_asc":
		orderBy = "price ASC"
	case "price_desc":
		orderBy = "price DESC"
	case "sales_desc":
		orderBy = "sold_count DESC"
	case "created_desc":
		orderBy = "created_at DESC"
	}

	// 分页查询
	var products []*model.Product
	offset := (req.Page - 1) * req.PageSize
	if err := query.Preload("Category").
		Preload("Brand").
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_main = ?", true).Order("sort ASC").Limit(1)
		}).
		Order(orderBy).
		Offset(offset).
		Limit(req.PageSize).
		Find(&products).Error; err != nil {
		return nil, 0, fmt.Errorf("查询商品列表失败: %v", err)
	}

	return products, total, nil
}

// UpdateProductStatus 更新商品状态
func (ps *ProductService) UpdateProductStatus(id uint, status string) error {
	var product model.Product
	if err := ps.db.First(&product, id).Error; err != nil {
		return fmt.Errorf("商品不存在")
	}

	// 验证状态值
	validStatuses := []string{model.ProductStatusDraft, model.ProductStatusActive, model.ProductStatusInactive}
	isValid := false
	for _, validStatus := range validStatuses {
		if status == validStatus {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("无效的状态值")
	}

	if err := ps.db.Model(&product).Update("status", status).Error; err != nil {
		return fmt.Errorf("更新商品状态失败: %v", err)
	}

	return nil
}

// UpdateProductStock 更新商品库存
func (ps *ProductService) UpdateProductStock(id uint, stock int) error {
	var product model.Product
	if err := ps.db.First(&product, id).Error; err != nil {
		return fmt.Errorf("商品不存在")
	}

	if stock < 0 {
		return fmt.Errorf("库存不能为负数")
	}

	if err := ps.db.Model(&product).Update("stock", stock).Error; err != nil {
		return fmt.Errorf("更新商品库存失败: %v", err)
	}

	return nil
}

// DeductStock 扣减库存
func (ps *ProductService) DeductStock(id uint, quantity int) error {
	if quantity <= 0 {
		return fmt.Errorf("扣减数量必须大于0")
	}

	// 使用乐观锁更新库存
	result := ps.db.Model(&model.Product{}).
		Where("id = ? AND stock >= ?", id, quantity).
		Updates(map[string]interface{}{
			"stock":      gorm.Expr("stock - ?", quantity),
			"sold_count": gorm.Expr("sold_count + ?", quantity),
		})

	if result.Error != nil {
		return fmt.Errorf("扣减库存失败: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("库存不足或商品不存在")
	}

	return nil
}

// RestoreStock 恢复库存
func (ps *ProductService) RestoreStock(id uint, quantity int) error {
	if quantity <= 0 {
		return fmt.Errorf("恢复数量必须大于0")
	}

	result := ps.db.Model(&model.Product{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"stock":      gorm.Expr("stock + ?", quantity),
			"sold_count": gorm.Expr("CASE WHEN sold_count >= ? THEN sold_count - ? ELSE 0 END", quantity, quantity),
		})

	if result.Error != nil {
		return fmt.Errorf("恢复库存失败: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("商品不存在")
	}

	return nil
}

// GetProductsByCategory 根据分类获取商品
func (ps *ProductService) GetProductsByCategory(categoryID uint, page, pageSize int) ([]*model.Product, int64, error) {
	req := &ProductListRequest{
		CategoryID: &categoryID,
		Status:     model.ProductStatusActive,
		Page:       page,
		PageSize:   pageSize,
	}

	return ps.GetProductList(req)
}

// GetProductsByBrand 根据品牌获取商品
func (ps *ProductService) GetProductsByBrand(brandID uint, page, pageSize int) ([]*model.Product, int64, error) {
	req := &ProductListRequest{
		BrandID:  &brandID,
		Status:   model.ProductStatusActive,
		Page:     page,
		PageSize: pageSize,
	}

	return ps.GetProductList(req)
}

// GetHotProducts 获取热销商品
func (ps *ProductService) GetHotProducts(limit int) ([]*model.Product, error) {
	var products []*model.Product
	isHot := true

	if err := ps.db.Where("status = ? AND is_hot = ?", model.ProductStatusActive, isHot).
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_main = ?", true).Order("sort ASC").Limit(1)
		}).
		Order("sold_count DESC, sort ASC").
		Limit(limit).
		Find(&products).Error; err != nil {
		return nil, fmt.Errorf("查询热销商品失败: %v", err)
	}

	return products, nil
}

// GetNewProducts 获取新品商品
func (ps *ProductService) GetNewProducts(limit int) ([]*model.Product, error) {
	var products []*model.Product
	isNew := true

	if err := ps.db.Where("status = ? AND is_new = ?", model.ProductStatusActive, isNew).
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_main = ?", true).Order("sort ASC").Limit(1)
		}).
		Order("created_at DESC, sort ASC").
		Limit(limit).
		Find(&products).Error; err != nil {
		return nil, fmt.Errorf("查询新品商品失败: %v", err)
	}

	return products, nil
}

// GetRecommendProducts 获取推荐商品
func (ps *ProductService) GetRecommendProducts(limit int) ([]*model.Product, error) {
	var products []*model.Product
	isRecommend := true

	if err := ps.db.Where("status = ? AND is_recommend = ?", model.ProductStatusActive, isRecommend).
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_main = ?", true).Order("sort ASC").Limit(1)
		}).
		Order("sort ASC, created_at DESC").
		Limit(limit).
		Find(&products).Error; err != nil {
		return nil, fmt.Errorf("查询推荐商品失败: %v", err)
	}

	return products, nil
}

// BatchUpdateProductStatus 批量更新商品状态
func (ps *ProductService) BatchUpdateProductStatus(ids []uint, status string) error {
	if len(ids) == 0 {
		return fmt.Errorf("商品ID列表不能为空")
	}

	// 验证状态值
	validStatuses := []string{model.ProductStatusDraft, model.ProductStatusActive, model.ProductStatusInactive}
	isValid := false
	for _, validStatus := range validStatuses {
		if status == validStatus {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("无效的状态值")
	}

	if err := ps.db.Model(&model.Product{}).Where("id IN ?", ids).Update("status", status).Error; err != nil {
		return fmt.Errorf("批量更新商品状态失败: %v", err)
	}

	return nil
}

// GetProductStatistics 获取商品统计信息
func (ps *ProductService) GetProductStatistics(merchantID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	query := ps.db.Model(&model.Product{})
	if merchantID > 0 {
		query = query.Where("merchant_id = ?", merchantID)
	}

	// 总商品数
	var totalCount int64
	query.Count(&totalCount)
	stats["total_products"] = totalCount

	// 各状态商品数
	var statusStats []struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	query.Select("status, COUNT(*) as count").Group("status").Scan(&statusStats)
	stats["status_stats"] = statusStats

	// 库存统计
	var stockStats struct {
		TotalStock    int64 `json:"total_stock"`
		LowStockCount int64 `json:"low_stock_count"`
		OutStockCount int64 `json:"out_stock_count"`
	}
	query.Select("SUM(stock) as total_stock").Scan(&stockStats.TotalStock)
	query.Where("stock <= min_stock AND min_stock > 0").Count(&stockStats.LowStockCount)
	query.Where("stock = 0").Count(&stockStats.OutStockCount)
	stats["stock_stats"] = stockStats

	return stats, nil
}

// 全局商品服务实例
var globalProductService *ProductService

// InitGlobalProductService 初始化全局商品服务
func InitGlobalProductService(db *gorm.DB) {
	globalProductService = NewProductService(db)
}

// GetGlobalProductService 获取全局商品服务
func GetGlobalProductService() *ProductService {
	return globalProductService
}
