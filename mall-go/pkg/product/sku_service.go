package product

import (
	"encoding/json"
	"fmt"

	"mall-go/internal/model"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// SKUService 商品SKU服务
type SKUService struct {
	db *gorm.DB
}

// NewSKUService 创建SKU服务
func NewSKUService(db *gorm.DB) *SKUService {
	return &SKUService{
		db: db,
	}
}

// CreateSKURequest 创建SKU请求
type CreateSKURequest struct {
	ProductID  uint                   `json:"product_id" binding:"required"`
	SKUCode    string                 `json:"sku_code" binding:"required,min=1,max=100"`
	Name       string                 `json:"name" binding:"required,min=1,max=255"`
	Price      decimal.Decimal        `json:"price" binding:"required"`
	Stock      int                    `json:"stock" binding:"min=0"`
	Image      string                 `json:"image"`
	Weight     decimal.Decimal        `json:"weight"`
	Volume     decimal.Decimal        `json:"volume"`
	Attributes map[string]interface{} `json:"attributes"`
}

// UpdateSKURequest 更新SKU请求
type UpdateSKURequest struct {
	Name       string                 `json:"name" binding:"required,min=1,max=255"`
	Price      decimal.Decimal        `json:"price" binding:"required"`
	Stock      int                    `json:"stock" binding:"min=0"`
	Image      string                 `json:"image"`
	Weight     decimal.Decimal        `json:"weight"`
	Volume     decimal.Decimal        `json:"volume"`
	Status     string                 `json:"status"`
	Attributes map[string]interface{} `json:"attributes"`
}

// SKUListRequest SKU列表请求
type SKUListRequest struct {
	ProductID uint   `form:"product_id"`
	Status    string `form:"status"`
	Keyword   string `form:"keyword"`
	Page      int    `form:"page" binding:"min=1"`
	PageSize  int    `form:"page_size" binding:"min=1,max=100"`
}

// BatchCreateSKURequest 批量创建SKU请求
type BatchCreateSKURequest struct {
	ProductID uint       `json:"product_id" binding:"required"`
	SKUs      []SKUBatch `json:"skus" binding:"required,min=1"`
}

// SKUBatch 批量SKU数据
type SKUBatch struct {
	SKUCode    string                 `json:"sku_code" binding:"required"`
	Name       string                 `json:"name" binding:"required"`
	Price      decimal.Decimal        `json:"price" binding:"required"`
	Stock      int                    `json:"stock" binding:"min=0"`
	Image      string                 `json:"image"`
	Weight     decimal.Decimal        `json:"weight"`
	Volume     decimal.Decimal        `json:"volume"`
	Attributes map[string]interface{} `json:"attributes"`
}

// CreateSKU 创建SKU
func (ss *SKUService) CreateSKU(req *CreateSKURequest) (*model.ProductSKU, error) {
	// 验证商品是否存在
	var product model.Product
	if err := ss.db.First(&product, req.ProductID).Error; err != nil {
		return nil, fmt.Errorf("商品不存在")
	}

	// 检查SKU编码是否已存在
	var existingSKU model.ProductSKU
	if err := ss.db.Where("sku_code = ?", req.SKUCode).First(&existingSKU).Error; err == nil {
		return nil, fmt.Errorf("SKU编码已存在")
	}

	// 序列化属性
	attributesJSON, err := json.Marshal(req.Attributes)
	if err != nil {
		return nil, fmt.Errorf("属性序列化失败: %v", err)
	}

	// 创建SKU
	sku := &model.ProductSKU{
		ProductID:  req.ProductID,
		SKUCode:    req.SKUCode,
		Name:       req.Name,
		Price:      req.Price,
		Stock:      req.Stock,
		Image:      req.Image,
		Weight:     req.Weight,
		Volume:     req.Volume,
		Status:     model.SKUStatusActive,
		Attributes: string(attributesJSON),
	}

	if err := ss.db.Create(sku).Error; err != nil {
		return nil, fmt.Errorf("创建SKU失败: %v", err)
	}

	return sku, nil
}

// UpdateSKU 更新SKU
func (ss *SKUService) UpdateSKU(id uint, req *UpdateSKURequest) (*model.ProductSKU, error) {
	var sku model.ProductSKU
	if err := ss.db.First(&sku, id).Error; err != nil {
		return nil, fmt.Errorf("SKU不存在")
	}

	// 序列化属性
	attributesJSON, err := json.Marshal(req.Attributes)
	if err != nil {
		return nil, fmt.Errorf("属性序列化失败: %v", err)
	}

	// 更新SKU信息
	sku.Name = req.Name
	sku.Price = req.Price
	sku.Stock = req.Stock
	sku.Image = req.Image
	sku.Weight = req.Weight
	sku.Volume = req.Volume
	sku.Attributes = string(attributesJSON)

	if req.Status != "" {
		sku.Status = req.Status
	}

	if err := ss.db.Save(&sku).Error; err != nil {
		return nil, fmt.Errorf("更新SKU失败: %v", err)
	}

	return &sku, nil
}

// DeleteSKU 删除SKU
func (ss *SKUService) DeleteSKU(id uint) error {
	var sku model.ProductSKU
	if err := ss.db.First(&sku, id).Error; err != nil {
		return fmt.Errorf("SKU不存在")
	}

	// 软删除SKU
	if err := ss.db.Delete(&sku).Error; err != nil {
		return fmt.Errorf("删除SKU失败: %v", err)
	}

	return nil
}

// GetSKU 获取SKU详情
func (ss *SKUService) GetSKU(id uint) (*model.ProductSKU, error) {
	var sku model.ProductSKU
	if err := ss.db.Preload("Product").First(&sku, id).Error; err != nil {
		return nil, fmt.Errorf("SKU不存在")
	}

	return &sku, nil
}

// GetSKUByCode 根据SKU编码获取SKU
func (ss *SKUService) GetSKUByCode(skuCode string) (*model.ProductSKU, error) {
	var sku model.ProductSKU
	if err := ss.db.Preload("Product").Where("sku_code = ?", skuCode).First(&sku).Error; err != nil {
		return nil, fmt.Errorf("SKU不存在")
	}

	return &sku, nil
}

// GetSKUList 获取SKU列表
func (ss *SKUService) GetSKUList(req *SKUListRequest) ([]*model.ProductSKU, int64, error) {
	query := ss.db.Model(&model.ProductSKU{})

	// 条件筛选
	if req.ProductID > 0 {
		query = query.Where("product_id = ?", req.ProductID)
	}

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	if req.Keyword != "" {
		query = query.Where("name LIKE ? OR sku_code LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询SKU总数失败: %v", err)
	}

	// 分页查询
	var skus []*model.ProductSKU
	offset := (req.Page - 1) * req.PageSize
	if err := query.Preload("Product").
		Order("id DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&skus).Error; err != nil {
		return nil, 0, fmt.Errorf("查询SKU列表失败: %v", err)
	}

	return skus, total, nil
}

// GetProductSKUs 获取商品的所有SKU
func (ss *SKUService) GetProductSKUs(productID uint) ([]*model.ProductSKU, error) {
	var skus []*model.ProductSKU
	if err := ss.db.Where("product_id = ? AND status = ?", productID, model.SKUStatusActive).
		Order("id ASC").
		Find(&skus).Error; err != nil {
		return nil, fmt.Errorf("查询商品SKU失败: %v", err)
	}

	return skus, nil
}

// BatchCreateSKUs 批量创建SKU
func (ss *SKUService) BatchCreateSKUs(req *BatchCreateSKURequest) ([]*model.ProductSKU, error) {
	// 验证商品是否存在
	var product model.Product
	if err := ss.db.First(&product, req.ProductID).Error; err != nil {
		return nil, fmt.Errorf("商品不存在")
	}

	// 检查SKU编码是否重复
	skuCodes := make([]string, len(req.SKUs))
	for i, skuReq := range req.SKUs {
		skuCodes[i] = skuReq.SKUCode
	}

	var existingCount int64
	if err := ss.db.Model(&model.ProductSKU{}).Where("sku_code IN ?", skuCodes).Count(&existingCount).Error; err != nil {
		return nil, fmt.Errorf("检查SKU编码失败: %v", err)
	}
	if existingCount > 0 {
		return nil, fmt.Errorf("存在重复的SKU编码")
	}

	// 开始事务
	tx := ss.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var createdSKUs []*model.ProductSKU

	// 批量创建SKU
	for _, skuReq := range req.SKUs {
		// 序列化属性
		attributesJSON, err := json.Marshal(skuReq.Attributes)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("属性序列化失败: %v", err)
		}

		sku := &model.ProductSKU{
			ProductID:  req.ProductID,
			SKUCode:    skuReq.SKUCode,
			Name:       skuReq.Name,
			Price:      skuReq.Price,
			Stock:      skuReq.Stock,
			Image:      skuReq.Image,
			Weight:     skuReq.Weight,
			Volume:     skuReq.Volume,
			Status:     model.SKUStatusActive,
			Attributes: string(attributesJSON),
		}

		if err := tx.Create(sku).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("创建SKU失败: %v", err)
		}

		createdSKUs = append(createdSKUs, sku)
	}

	tx.Commit()
	return createdSKUs, nil
}

// UpdateSKUStatus 更新SKU状态
func (ss *SKUService) UpdateSKUStatus(id uint, status string) error {
	var sku model.ProductSKU
	if err := ss.db.First(&sku, id).Error; err != nil {
		return fmt.Errorf("SKU不存在")
	}

	// 验证状态值
	validStatuses := []string{model.SKUStatusActive, model.SKUStatusInactive}
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

	if err := ss.db.Model(&sku).Update("status", status).Error; err != nil {
		return fmt.Errorf("更新SKU状态失败: %v", err)
	}

	return nil
}

// UpdateSKUStock 更新SKU库存
func (ss *SKUService) UpdateSKUStock(id uint, stock int) error {
	var sku model.ProductSKU
	if err := ss.db.First(&sku, id).Error; err != nil {
		return fmt.Errorf("SKU不存在")
	}

	if stock < 0 {
		return fmt.Errorf("库存不能为负数")
	}

	if err := ss.db.Model(&sku).Update("stock", stock).Error; err != nil {
		return fmt.Errorf("更新SKU库存失败: %v", err)
	}

	return nil
}

// DeductSKUStock 扣减SKU库存
func (ss *SKUService) DeductSKUStock(id uint, quantity int) error {
	if quantity <= 0 {
		return fmt.Errorf("扣减数量必须大于0")
	}

	// 使用乐观锁更新库存
	result := ss.db.Model(&model.ProductSKU{}).
		Where("id = ? AND stock >= ?", id, quantity).
		Update("stock", gorm.Expr("stock - ?", quantity))

	if result.Error != nil {
		return fmt.Errorf("扣减SKU库存失败: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("SKU库存不足或不存在")
	}

	return nil
}

// RestoreSKUStock 恢复SKU库存
func (ss *SKUService) RestoreSKUStock(id uint, quantity int) error {
	if quantity <= 0 {
		return fmt.Errorf("恢复数量必须大于0")
	}

	result := ss.db.Model(&model.ProductSKU{}).
		Where("id = ?", id).
		Update("stock", gorm.Expr("stock + ?", quantity))

	if result.Error != nil {
		return fmt.Errorf("恢复SKU库存失败: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("SKU不存在")
	}

	return nil
}

// BatchUpdateSKUStatus 批量更新SKU状态
func (ss *SKUService) BatchUpdateSKUStatus(ids []uint, status string) error {
	if len(ids) == 0 {
		return fmt.Errorf("SKU ID列表不能为空")
	}

	// 验证状态值
	validStatuses := []string{model.SKUStatusActive, model.SKUStatusInactive}
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

	if err := ss.db.Model(&model.ProductSKU{}).Where("id IN ?", ids).Update("status", status).Error; err != nil {
		return fmt.Errorf("批量更新SKU状态失败: %v", err)
	}

	return nil
}

// GetSKUStatistics 获取SKU统计信息
func (ss *SKUService) GetSKUStatistics(productID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	query := ss.db.Model(&model.ProductSKU{})
	if productID > 0 {
		query = query.Where("product_id = ?", productID)
	}

	// 总SKU数
	var totalCount int64
	query.Count(&totalCount)
	stats["total_skus"] = totalCount

	// 各状态SKU数
	var statusStats []struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	query.Select("status, COUNT(*) as count").Group("status").Scan(&statusStats)
	stats["status_stats"] = statusStats

	// 库存统计
	var stockStats struct {
		TotalStock    int64 `json:"total_stock"`
		OutStockCount int64 `json:"out_stock_count"`
	}
	query.Select("SUM(stock) as total_stock").Scan(&stockStats.TotalStock)
	query.Where("stock = 0").Count(&stockStats.OutStockCount)
	stats["stock_stats"] = stockStats

	return stats, nil
}

// 全局SKU服务实例
var globalSKUService *SKUService

// InitGlobalSKUService 初始化全局SKU服务
func InitGlobalSKUService(db *gorm.DB) {
	globalSKUService = NewSKUService(db)
}

// GetGlobalSKUService 获取全局SKU服务
func GetGlobalSKUService() *SKUService {
	return globalSKUService
}
