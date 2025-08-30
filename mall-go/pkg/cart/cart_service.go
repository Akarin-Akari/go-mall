package cart

import (
	"fmt"
	"time"

	"mall-go/internal/model"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// CartService 购物车服务
type CartService struct {
	db *gorm.DB
}

// NewCartService 创建购物车服务
func NewCartService(db *gorm.DB) *CartService {
	return &CartService{
		db: db,
	}
}

// AddToCart 添加商品到购物车
func (cs *CartService) AddToCart(userID uint, sessionID string, req *model.AddToCartRequest) (*model.CartItem, error) {
	// 验证商品是否存在且可购买
	var product model.Product
	if err := cs.db.Where("id = ? AND status = ?", req.ProductID, model.ProductStatusActive).First(&product).Error; err != nil {
		return nil, fmt.Errorf("商品不存在或已下架")
	}

	// 如果指定了SKU，验证SKU
	var sku *model.ProductSKU
	if req.SKUID > 0 {
		var skuModel model.ProductSKU
		if err := cs.db.Where("id = ? AND product_id = ? AND status = ?", req.SKUID, req.ProductID, model.SKUStatusActive).First(&skuModel).Error; err != nil {
			return nil, fmt.Errorf("商品规格不存在或已下架")
		}
		sku = &skuModel
	}

	// 检查库存
	availableStock := product.Stock
	currentPrice := product.Price
	if sku != nil {
		availableStock = sku.Stock
		currentPrice = sku.Price
	}

	if availableStock < req.Quantity {
		return nil, fmt.Errorf("库存不足，当前库存：%d", availableStock)
	}

	// 获取或创建购物车
	cart, err := cs.getOrCreateCart(userID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("获取购物车失败: %v", err)
	}

	// 检查是否已存在相同商品
	var existingItem model.CartItem
	query := cs.db.Where("cart_id = ? AND product_id = ?", cart.ID, req.ProductID)
	if req.SKUID > 0 {
		query = query.Where("sku_id = ?", req.SKUID)
	} else {
		query = query.Where("sku_id = 0")
	}

	if err := query.First(&existingItem).Error; err == nil {
		// 商品已存在，更新数量
		newQuantity := existingItem.Quantity + req.Quantity
		if availableStock < newQuantity {
			return nil, fmt.Errorf("库存不足，当前库存：%d，购物车已有：%d", availableStock, existingItem.Quantity)
		}

		existingItem.Quantity = newQuantity
		existingItem.Price = currentPrice // 更新为当前价格
		existingItem.UpdatedAt = time.Now()

		if err := cs.db.Save(&existingItem).Error; err != nil {
			return nil, fmt.Errorf("更新购物车商品失败: %v", err)
		}

		// 更新购物车统计
		cs.updateCartSummary(cart.ID)
		return &existingItem, nil
	}

	// 创建新的购物车商品项
	cartItem := &model.CartItem{
		CartID:    cart.ID,
		ProductID: req.ProductID,
		SKUID:     req.SKUID,
		Quantity:  req.Quantity,
		Price:     currentPrice,
		Selected:  true,
		Status:    model.CartItemStatusNormal,

		// 商品快照信息
		ProductName:  product.Name,
		ProductImage: product.GetMainImage(),
	}

	// 如果有SKU，填充SKU信息
	if sku != nil {
		cartItem.SKUName = sku.Name
		cartItem.SKUImage = sku.Image
		cartItem.SKUAttrs = sku.Attributes
	}

	if err := cs.db.Create(cartItem).Error; err != nil {
		return nil, fmt.Errorf("添加商品到购物车失败: %v", err)
	}

	// 更新购物车统计
	cs.updateCartSummary(cart.ID)

	return cartItem, nil
}

// RemoveFromCart 从购物车移除商品
func (cs *CartService) RemoveFromCart(userID uint, sessionID string, itemID uint) error {
	cart, err := cs.getCart(userID, sessionID)
	if err != nil {
		return fmt.Errorf("购物车不存在")
	}

	// 删除购物车商品项
	result := cs.db.Where("id = ? AND cart_id = ?", itemID, cart.ID).Delete(&model.CartItem{})
	if result.Error != nil {
		return fmt.Errorf("删除商品失败: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("商品不存在")
	}

	// 更新购物车统计
	cs.updateCartSummary(cart.ID)

	return nil
}

// UpdateCartItem 更新购物车商品
func (cs *CartService) UpdateCartItem(userID uint, sessionID string, itemID uint, req *model.UpdateCartItemRequest) (*model.CartItem, error) {
	cart, err := cs.getCart(userID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("购物车不存在")
	}

	// 查找购物车商品项
	var cartItem model.CartItem
	if err := cs.db.Where("id = ? AND cart_id = ?", itemID, cart.ID).First(&cartItem).Error; err != nil {
		return nil, fmt.Errorf("商品不存在")
	}

	// 检查库存
	if req.Quantity > 0 {
		availableStock, err := cs.getAvailableStock(cartItem.ProductID, cartItem.SKUID)
		if err != nil {
			return nil, fmt.Errorf("获取库存信息失败: %v", err)
		}

		if availableStock < req.Quantity {
			return nil, fmt.Errorf("库存不足，当前库存：%d", availableStock)
		}

		cartItem.Quantity = req.Quantity
	}

	cartItem.Selected = req.Selected
	cartItem.UpdatedAt = time.Now()

	if err := cs.db.Save(&cartItem).Error; err != nil {
		return nil, fmt.Errorf("更新购物车商品失败: %v", err)
	}

	// 更新购物车统计
	cs.updateCartSummary(cart.ID)

	return &cartItem, nil
}

// ClearCart 清空购物车
func (cs *CartService) ClearCart(userID uint, sessionID string) error {
	cart, err := cs.getCart(userID, sessionID)
	if err != nil {
		return fmt.Errorf("购物车不存在")
	}

	// 删除所有购物车商品项
	if err := cs.db.Where("cart_id = ?", cart.ID).Delete(&model.CartItem{}).Error; err != nil {
		return fmt.Errorf("清空购物车失败: %v", err)
	}

	// 更新购物车统计
	cs.updateCartSummary(cart.ID)

	return nil
}

// GetCart 获取购物车
func (cs *CartService) GetCart(userID uint, sessionID string, includeInvalid bool) (*model.CartResponse, error) {
	cart, err := cs.getCart(userID, sessionID)
	if err != nil {
		// 如果购物车不存在，返回空购物车
		return &model.CartResponse{
			Cart: &model.Cart{
				UserID:      userID,
				SessionID:   sessionID,
				Status:      model.CartStatusActive,
				ItemCount:   0,
				TotalQty:    0,
				TotalAmount: decimal.Zero,
				Items:       []model.CartItem{},
			},
			Summary: &model.CartSummary{
				ItemCount:      0,
				TotalQty:       0,
				SelectedCount:  0,
				SelectedQty:    0,
				TotalAmount:    decimal.Zero,
				SelectedAmount: decimal.Zero,
				DiscountAmount: decimal.Zero,
				ShippingFee:    decimal.Zero,
				FinalAmount:    decimal.Zero,
				InvalidItems:   []model.CartItem{},
			},
		}, nil
	}

	// 加载购物车商品项
	query := cs.db.Where("cart_id = ?", cart.ID)
	if !includeInvalid {
		query = query.Where("status = ?", model.CartItemStatusNormal)
	}

	var items []model.CartItem
	if err := query.Preload("Product").Preload("SKU").Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("获取购物车商品失败: %v", err)
	}

	cart.Items = items

	// 计算购物车汇总信息
	summary := cs.calculateCartSummary(cart)

	return &model.CartResponse{
		Cart:    cart,
		Summary: summary,
	}, nil
}

// BatchUpdateCart 批量更新购物车
func (cs *CartService) BatchUpdateCart(userID uint, sessionID string, req *model.BatchUpdateCartRequest) error {
	cart, err := cs.getCart(userID, sessionID)
	if err != nil {
		return fmt.Errorf("购物车不存在")
	}

	// 开始事务
	tx := cs.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, item := range req.Items {
		var cartItem model.CartItem
		if err := tx.Where("id = ? AND cart_id = ?", item.ID, cart.ID).First(&cartItem).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("商品ID %d 不存在", item.ID)
		}

		// 检查库存
		if item.Quantity > 0 {
			availableStock, err := cs.getAvailableStock(cartItem.ProductID, cartItem.SKUID)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("获取商品ID %d 库存信息失败: %v", item.ID, err)
			}

			if availableStock < item.Quantity {
				tx.Rollback()
				return fmt.Errorf("商品ID %d 库存不足，当前库存：%d", item.ID, availableStock)
			}

			cartItem.Quantity = item.Quantity
		}

		cartItem.Selected = item.Selected
		cartItem.UpdatedAt = time.Now()

		if err := tx.Save(&cartItem).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("更新商品ID %d 失败: %v", item.ID, err)
		}
	}

	tx.Commit()

	// 更新购物车统计
	cs.updateCartSummary(cart.ID)

	return nil
}

// SelectAllItems 全选/取消全选购物车商品
func (cs *CartService) SelectAllItems(userID uint, sessionID string, selected bool) error {
	cart, err := cs.getCart(userID, sessionID)
	if err != nil {
		return fmt.Errorf("购物车不存在")
	}

	// 更新所有有效商品的选中状态
	if err := cs.db.Model(&model.CartItem{}).
		Where("cart_id = ? AND status = ?", cart.ID, model.CartItemStatusNormal).
		Update("selected", selected).Error; err != nil {
		return fmt.Errorf("更新选中状态失败: %v", err)
	}

	// 更新购物车统计
	cs.updateCartSummary(cart.ID)

	return nil
}

// RemoveInvalidItems 移除失效商品
func (cs *CartService) RemoveInvalidItems(userID uint, sessionID string) error {
	cart, err := cs.getCart(userID, sessionID)
	if err != nil {
		return fmt.Errorf("购物车不存在")
	}

	// 删除失效商品
	if err := cs.db.Where("cart_id = ? AND status != ?", cart.ID, model.CartItemStatusNormal).Delete(&model.CartItem{}).Error; err != nil {
		return fmt.Errorf("删除失效商品失败: %v", err)
	}

	// 更新购物车统计
	cs.updateCartSummary(cart.ID)

	return nil
}

// GetCartItemCount 获取购物车商品数量
func (cs *CartService) GetCartItemCount(userID uint, sessionID string) (int, error) {
	cart, err := cs.getCart(userID, sessionID)
	if err != nil {
		return 0, nil // 购物车不存在返回0
	}

	var count int64
	if err := cs.db.Model(&model.CartItem{}).
		Where("cart_id = ? AND status = ?", cart.ID, model.CartItemStatusNormal).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("获取购物车商品数量失败: %v", err)
	}

	return int(count), nil
}

// getOrCreateCart 获取或创建购物车
func (cs *CartService) getOrCreateCart(userID uint, sessionID string) (*model.Cart, error) {
	// 先尝试获取现有购物车
	cart, err := cs.getCart(userID, sessionID)
	if err == nil {
		return cart, nil
	}

	// 创建新购物车
	cart = &model.Cart{
		UserID:      userID,
		SessionID:   sessionID,
		Status:      model.CartStatusActive,
		ItemCount:   0,
		TotalQty:    0,
		TotalAmount: decimal.Zero,
	}

	if err := cs.db.Create(cart).Error; err != nil {
		return nil, fmt.Errorf("创建购物车失败: %v", err)
	}

	return cart, nil
}

// getCart 获取购物车
func (cs *CartService) getCart(userID uint, sessionID string) (*model.Cart, error) {
	var cart model.Cart
	query := cs.db.Where("status = ?", model.CartStatusActive)

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	} else if sessionID != "" {
		query = query.Where("user_id = 0 AND session_id = ?", sessionID)
	} else {
		return nil, fmt.Errorf("用户ID和会话ID不能同时为空")
	}

	if err := query.First(&cart).Error; err != nil {
		return nil, err
	}

	return &cart, nil
}

// getAvailableStock 获取可用库存
func (cs *CartService) getAvailableStock(productID, skuID uint) (int, error) {
	if skuID > 0 {
		var sku model.ProductSKU
		if err := cs.db.Where("id = ? AND status = ?", skuID, model.SKUStatusActive).First(&sku).Error; err != nil {
			return 0, fmt.Errorf("SKU不存在")
		}
		return sku.Stock, nil
	}

	var product model.Product
	if err := cs.db.Where("id = ? AND status = ?", productID, model.ProductStatusActive).First(&product).Error; err != nil {
		return 0, fmt.Errorf("商品不存在")
	}
	return product.Stock, nil
}

// updateCartSummary 更新购物车统计信息
func (cs *CartService) updateCartSummary(cartID uint) error {
	var cart model.Cart
	if err := cs.db.Preload("Items").First(&cart, cartID).Error; err != nil {
		return err
	}

	// 计算统计信息
	itemCount := len(cart.Items)
	totalQty := 0
	totalAmount := decimal.Zero

	for _, item := range cart.Items {
		if item.Status == model.CartItemStatusNormal {
			totalQty += item.Quantity
			if item.Selected {
				itemTotal := item.Price.Mul(decimal.NewFromInt(int64(item.Quantity)))
				totalAmount = totalAmount.Add(itemTotal)
			}
		}
	}

	// 更新购物车统计
	return cs.db.Model(&cart).Updates(map[string]interface{}{
		"item_count":   itemCount,
		"total_qty":    totalQty,
		"total_amount": totalAmount,
		"updated_at":   time.Now(),
	}).Error
}

// calculateCartSummary 计算购物车汇总信息
func (cs *CartService) calculateCartSummary(cart *model.Cart) *model.CartSummary {
	summary := &model.CartSummary{
		ItemCount:       0,
		TotalQty:        0,
		SelectedCount:   0,
		SelectedQty:     0,
		TotalAmount:     decimal.Zero,
		SelectedAmount:  decimal.Zero,
		DiscountAmount:  decimal.Zero,
		ShippingFee:     decimal.Zero,
		FinalAmount:     decimal.Zero,
		InvalidItems:    []model.CartItem{},
	}

	for _, item := range cart.Items {
		if item.Status == model.CartItemStatusNormal {
			summary.ItemCount++
			summary.TotalQty += item.Quantity
			itemTotal := item.Price.Mul(decimal.NewFromInt(int64(item.Quantity)))
			summary.TotalAmount = summary.TotalAmount.Add(itemTotal)

			if item.Selected {
				summary.SelectedCount++
				summary.SelectedQty += item.Quantity
				summary.SelectedAmount = summary.SelectedAmount.Add(itemTotal)
			}
		} else {
			summary.InvalidItems = append(summary.InvalidItems, item)
		}
	}

	// 计算最终金额（暂时不考虑优惠和运费）
	summary.FinalAmount = summary.SelectedAmount.Sub(summary.DiscountAmount).Add(summary.ShippingFee)

	return summary
}

// 全局购物车服务实例
var globalCartService *CartService

// InitGlobalCartService 初始化全局购物车服务
func InitGlobalCartService(db *gorm.DB) {
	globalCartService = NewCartService(db)
}

// GetGlobalCartService 获取全局购物车服务
func GetGlobalCartService() *CartService {
	return globalCartService
}
