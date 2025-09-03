package product

import (
	"fmt"
	"time"

	"mall-go/internal/model"

	"gorm.io/gorm"
)

// InventoryService 库存管理服务
type InventoryService struct {
	db *gorm.DB
}

// NewInventoryService 创建库存管理服务
func NewInventoryService(db *gorm.DB) *InventoryService {
	return &InventoryService{
		db: db,
	}
}

// InventoryLog 库存变动日志模型
type InventoryLog struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	ProductID  uint      `gorm:"not null;index" json:"product_id"`
	SKUID      uint      `gorm:"index" json:"sku_id"`
	Type       string    `gorm:"size:20;not null;index" json:"type"`        // 变动类型：in/out/adjust
	Quantity   int       `gorm:"not null" json:"quantity"`                  // 变动数量
	BeforeQty  int       `gorm:"not null" json:"before_qty"`                // 变动前数量
	AfterQty   int       `gorm:"not null" json:"after_qty"`                 // 变动后数量
	Reason     string    `gorm:"size:100" json:"reason"`                    // 变动原因
	OrderID    uint      `gorm:"index" json:"order_id"`                     // 关联订单ID
	UserID     uint      `gorm:"not null;index" json:"user_id"`             // 操作用户ID
	Remark     string    `gorm:"size:500" json:"remark"`                    // 备注
	CreatedAt  time.Time `json:"created_at"`
	
	// 关联关系
	Product *model.Product    `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	SKU     *model.ProductSKU `gorm:"foreignKey:SKUID" json:"sku,omitempty"`
	User    *model.User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (InventoryLog) TableName() string {
	return "inventory_logs"
}

// StockAdjustRequest 库存调整请求
type StockAdjustRequest struct {
	ProductID uint   `json:"product_id" binding:"required"`
	SKUID     uint   `json:"sku_id"`
	Quantity  int    `json:"quantity" binding:"required"`
	Reason    string `json:"reason" binding:"required"`
	Remark    string `json:"remark"`
	UserID    uint   `json:"user_id" binding:"required"`
}

// StockInRequest 入库请求
type StockInRequest struct {
	ProductID uint   `json:"product_id" binding:"required"`
	SKUID     uint   `json:"sku_id"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
	Reason    string `json:"reason" binding:"required"`
	Remark    string `json:"remark"`
	UserID    uint   `json:"user_id" binding:"required"`
}

// StockOutRequest 出库请求
type StockOutRequest struct {
	ProductID uint   `json:"product_id" binding:"required"`
	SKUID     uint   `json:"sku_id"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
	Reason    string `json:"reason" binding:"required"`
	OrderID   uint   `json:"order_id"`
	Remark    string `json:"remark"`
	UserID    uint   `json:"user_id" binding:"required"`
}

// InventoryListRequest 库存列表请求
type InventoryListRequest struct {
	ProductID   uint   `form:"product_id"`
	CategoryID  uint   `form:"category_id"`
	MerchantID  uint   `form:"merchant_id"`
	Keyword     string `form:"keyword"`
	LowStock    bool   `form:"low_stock"`    // 是否只显示低库存
	OutOfStock  bool   `form:"out_of_stock"` // 是否只显示缺货
	Page        int    `form:"page" binding:"min=1"`
	PageSize    int    `form:"page_size" binding:"min=1,max=100"`
}

// InventoryLogListRequest 库存日志列表请求
type InventoryLogListRequest struct {
	ProductID uint   `form:"product_id"`
	SKUID     uint   `form:"sku_id"`
	Type      string `form:"type"`
	UserID    uint   `form:"user_id"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	Page      int    `form:"page" binding:"min=1"`
	PageSize  int    `form:"page_size" binding:"min=1,max=100"`
}

// 库存变动类型常量
const (
	InventoryTypeIn     = "in"     // 入库
	InventoryTypeOut    = "out"    // 出库
	InventoryTypeAdjust = "adjust" // 调整
)

// 库存变动原因常量
const (
	InventoryReasonPurchase = "purchase"  // 采购入库
	InventoryReasonReturn   = "return"    // 退货入库
	InventoryReasonSale     = "sale"      // 销售出库
	InventoryReasonDamage   = "damage"    // 损坏出库
	InventoryReasonAdjust   = "adjust"    // 库存调整
	InventoryReasonInit     = "init"      // 初始化
)

// AdjustStock 调整库存
func (is *InventoryService) AdjustStock(req *StockAdjustRequest) error {
	// 开始事务
	tx := is.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var beforeQty, afterQty int
	var err error

	if req.SKUID > 0 {
		// 调整SKU库存
		beforeQty, afterQty, err = is.adjustSKUStock(tx, req.SKUID, req.Quantity)
	} else {
		// 调整商品库存
		beforeQty, afterQty, err = is.adjustProductStock(tx, req.ProductID, req.Quantity)
	}

	if err != nil {
		tx.Rollback()
		return err
	}

	// 记录库存变动日志
	log := &InventoryLog{
		ProductID: req.ProductID,
		SKUID:     req.SKUID,
		Type:      InventoryTypeAdjust,
		Quantity:  req.Quantity,
		BeforeQty: beforeQty,
		AfterQty:  afterQty,
		Reason:    req.Reason,
		UserID:    req.UserID,
		Remark:    req.Remark,
	}

	if err := tx.Create(log).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("记录库存日志失败: %v", err)
	}

	tx.Commit()
	return nil
}

// StockIn 入库
func (is *InventoryService) StockIn(req *StockInRequest) error {
	// 开始事务
	tx := is.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var beforeQty, afterQty int
	var err error

	if req.SKUID > 0 {
		// SKU入库
		beforeQty, afterQty, err = is.adjustSKUStock(tx, req.SKUID, req.Quantity)
	} else {
		// 商品入库
		beforeQty, afterQty, err = is.adjustProductStock(tx, req.ProductID, req.Quantity)
	}

	if err != nil {
		tx.Rollback()
		return err
	}

	// 记录库存变动日志
	log := &InventoryLog{
		ProductID: req.ProductID,
		SKUID:     req.SKUID,
		Type:      InventoryTypeIn,
		Quantity:  req.Quantity,
		BeforeQty: beforeQty,
		AfterQty:  afterQty,
		Reason:    req.Reason,
		UserID:    req.UserID,
		Remark:    req.Remark,
	}

	if err := tx.Create(log).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("记录库存日志失败: %v", err)
	}

	tx.Commit()
	return nil
}

// StockOut 出库
func (is *InventoryService) StockOut(req *StockOutRequest) error {
	// 开始事务
	tx := is.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var beforeQty, afterQty int
	var err error

	if req.SKUID > 0 {
		// SKU出库
		beforeQty, afterQty, err = is.adjustSKUStock(tx, req.SKUID, -req.Quantity)
	} else {
		// 商品出库
		beforeQty, afterQty, err = is.adjustProductStock(tx, req.ProductID, -req.Quantity)
	}

	if err != nil {
		tx.Rollback()
		return err
	}

	// 记录库存变动日志
	log := &InventoryLog{
		ProductID: req.ProductID,
		SKUID:     req.SKUID,
		Type:      InventoryTypeOut,
		Quantity:  req.Quantity,
		BeforeQty: beforeQty,
		AfterQty:  afterQty,
		Reason:    req.Reason,
		OrderID:   req.OrderID,
		UserID:    req.UserID,
		Remark:    req.Remark,
	}

	if err := tx.Create(log).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("记录库存日志失败: %v", err)
	}

	tx.Commit()
	return nil
}

// GetInventoryList 获取库存列表
func (is *InventoryService) GetInventoryList(req *InventoryListRequest) ([]*model.Product, int64, error) {
	query := is.db.Model(&model.Product{})

	// 条件筛选
	if req.ProductID > 0 {
		query = query.Where("id = ?", req.ProductID)
	}

	if req.CategoryID > 0 {
		query = query.Where("category_id = ?", req.CategoryID)
	}

	if req.MerchantID > 0 {
		query = query.Where("merchant_id = ?", req.MerchantID)
	}

	if req.Keyword != "" {
		query = query.Where("name LIKE ?", "%"+req.Keyword+"%")
	}

	if req.LowStock {
		query = query.Where("stock <= min_stock AND min_stock > 0")
	}

	if req.OutOfStock {
		query = query.Where("stock = 0")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询库存总数失败: %v", err)
	}

	// 分页查询
	var products []*model.Product
	offset := (req.Page - 1) * req.PageSize
	if err := query.Preload("Category").
		Preload("SKUs", func(db *gorm.DB) *gorm.DB {
			return db.Where("status = ?", model.SKUStatusActive)
		}).
		Order("stock ASC, id DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&products).Error; err != nil {
		return nil, 0, fmt.Errorf("查询库存列表失败: %v", err)
	}

	return products, total, nil
}

// GetInventoryLogs 获取库存变动日志
func (is *InventoryService) GetInventoryLogs(req *InventoryLogListRequest) ([]*InventoryLog, int64, error) {
	query := is.db.Model(&InventoryLog{})

	// 条件筛选
	if req.ProductID > 0 {
		query = query.Where("product_id = ?", req.ProductID)
	}

	if req.SKUID > 0 {
		query = query.Where("sku_id = ?", req.SKUID)
	}

	if req.Type != "" {
		query = query.Where("type = ?", req.Type)
	}

	if req.UserID > 0 {
		query = query.Where("user_id = ?", req.UserID)
	}

	if req.StartDate != "" {
		query = query.Where("created_at >= ?", req.StartDate)
	}

	if req.EndDate != "" {
		query = query.Where("created_at <= ?", req.EndDate)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询库存日志总数失败: %v", err)
	}

	// 分页查询
	var logs []*InventoryLog
	offset := (req.Page - 1) * req.PageSize
	if err := query.Preload("Product").
		Preload("SKU").
		Preload("User").
		Order("created_at DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("查询库存日志失败: %v", err)
	}

	return logs, total, nil
}

// GetLowStockProducts 获取低库存商品
func (is *InventoryService) GetLowStockProducts(limit int) ([]*model.Product, error) {
	var products []*model.Product
	if err := is.db.Where("stock <= min_stock AND min_stock > 0").
		Preload("Category").
		Order("stock ASC").
		Limit(limit).
		Find(&products).Error; err != nil {
		return nil, fmt.Errorf("查询低库存商品失败: %v", err)
	}

	return products, nil
}

// GetOutOfStockProducts 获取缺货商品
func (is *InventoryService) GetOutOfStockProducts(limit int) ([]*model.Product, error) {
	var products []*model.Product
	if err := is.db.Where("stock = 0").
		Preload("Category").
		Order("id DESC").
		Limit(limit).
		Find(&products).Error; err != nil {
		return nil, fmt.Errorf("查询缺货商品失败: %v", err)
	}

	return products, nil
}

// GetInventoryStatistics 获取库存统计
func (is *InventoryService) GetInventoryStatistics(merchantID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	query := is.db.Model(&model.Product{})
	if merchantID > 0 {
		query = query.Where("merchant_id = ?", merchantID)
	}

	// 总库存
	var totalStock int64
	query.Select("SUM(stock)").Scan(&totalStock)
	stats["total_stock"] = totalStock

	// 低库存商品数
	var lowStockCount int64
	query.Where("stock <= min_stock AND min_stock > 0").Count(&lowStockCount)
	stats["low_stock_count"] = lowStockCount

	// 缺货商品数
	var outStockCount int64
	query.Where("stock = 0").Count(&outStockCount)
	stats["out_stock_count"] = outStockCount

	// 库存总价值（按成本价计算）
	var totalValue float64
	query.Select("SUM(stock * cost_price)").Scan(&totalValue)
	stats["total_value"] = totalValue

	return stats, nil
}

// adjustProductStock 调整商品库存
func (is *InventoryService) adjustProductStock(tx *gorm.DB, productID uint, quantity int) (int, int, error) {
	var product model.Product
	if err := tx.First(&product, productID).Error; err != nil {
		return 0, 0, fmt.Errorf("商品不存在")
	}

	beforeQty := product.Stock
	afterQty := beforeQty + quantity

	if afterQty < 0 {
		return 0, 0, fmt.Errorf("库存不足")
	}

	if err := tx.Model(&product).Update("stock", afterQty).Error; err != nil {
		return 0, 0, fmt.Errorf("更新商品库存失败: %v", err)
	}

	return beforeQty, afterQty, nil
}

// adjustSKUStock 调整SKU库存
func (is *InventoryService) adjustSKUStock(tx *gorm.DB, skuID uint, quantity int) (int, int, error) {
	var sku model.ProductSKU
	if err := tx.First(&sku, skuID).Error; err != nil {
		return 0, 0, fmt.Errorf("SKU不存在")
	}

	beforeQty := sku.Stock
	afterQty := beforeQty + quantity

	if afterQty < 0 {
		return 0, 0, fmt.Errorf("SKU库存不足")
	}

	if err := tx.Model(&sku).Update("stock", afterQty).Error; err != nil {
		return 0, 0, fmt.Errorf("更新SKU库存失败: %v", err)
	}

	return beforeQty, afterQty, nil
}

// 全局库存服务实例
var globalInventoryService *InventoryService

// InitGlobalInventoryService 初始化全局库存服务
func InitGlobalInventoryService(db *gorm.DB) {
	globalInventoryService = NewInventoryService(db)
}

// GetGlobalInventoryService 获取全局库存服务
func GetGlobalInventoryService() *InventoryService {
	return globalInventoryService
}
