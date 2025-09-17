package order

import (
	"fmt"
	"time"

	"mall-go/internal/model"
	"mall-go/pkg/cart"
	"mall-go/pkg/inventory"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// OrderService 订单服务
type OrderService struct {
	db                 *gorm.DB
	cartService        *cart.CartService
	calculationService *cart.CalculationService
	inventoryService   *inventory.InventoryService
}

// NewOrderService 创建订单服务
func NewOrderService(db *gorm.DB, cartService *cart.CartService, calculationService *cart.CalculationService, inventoryService *inventory.InventoryService) *OrderService {
	return &OrderService{
		db:                 db,
		cartService:        cartService,
		calculationService: calculationService,
		inventoryService:   inventoryService,
	}
}

// CreateOrder 创建订单 - 并发安全版本
func (os *OrderService) CreateOrder(userID uint, req *model.OrderCreateRequest) (*model.Order, error) {
	// 使用事务确保数据一致性
	var order *model.Order
	err := os.db.Transaction(func(tx *gorm.DB) error {
		// 获取购物车商品项（使用悲观锁防止并发修改）
		var cartItems []model.CartItem
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("id IN ? AND cart_id IN (SELECT id FROM carts WHERE user_id = ?)",
				req.CartItemIDs, userID).
			Preload("Product").
			Preload("SKU").
			Find(&cartItems).Error; err != nil {
			return fmt.Errorf("获取购物车商品失败: %v", err)
		}

		if len(cartItems) == 0 {
			return fmt.Errorf("购物车商品为空")
		}

		// 预先验证所有商品和库存（在扣减前检查）
		if err := os.validateCartItemsForOrder(tx, cartItems); err != nil {
			return err
		}

		// 创建订单对象
		var createErr error
		order, createErr = os.createOrderWithItems(tx, userID, req, cartItems)
		if createErr != nil {
			return createErr
		}

		// 使用库存服务扣减库存（带乐观锁重试）
		if err := os.deductStockWithInventoryService(cartItems); err != nil {
			return fmt.Errorf("扣减库存失败: %v", err)
		}

		// 清理购物车商品项
		if err := tx.Where("id IN ?", req.CartItemIDs).Delete(&model.CartItem{}).Error; err != nil {
			return fmt.Errorf("清理购物车失败: %v", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return order, nil
}

// validateCartItemsForOrder 验证购物车商品项是否可以下单
func (os *OrderService) validateCartItemsForOrder(tx *gorm.DB, cartItems []model.CartItem) error {
	for _, item := range cartItems {
		if item.Product == nil {
			return fmt.Errorf("商品ID %d 不存在", item.ProductID)
		}

		if item.Product.Status != model.ProductStatusActive {
			return fmt.Errorf("商品 %s 已下架", item.Product.Name)
		}

		// 检查库存
		availableStock := item.Product.Stock
		if item.SKUID > 0 {
			if item.SKU == nil {
				return fmt.Errorf("商品规格ID %d 不存在", item.SKUID)
			}
			if item.SKU.Status != model.SKUStatusActive {
				return fmt.Errorf("商品规格 %s 已下架", item.SKU.Name)
			}
			availableStock = item.SKU.Stock
		}

		if availableStock < item.Quantity {
			return fmt.Errorf("商品 %s 库存不足，当前库存：%d", item.Product.Name, availableStock)
		}
	}
	return nil
}

// createOrderWithItems 创建订单和订单商品项
func (os *OrderService) createOrderWithItems(tx *gorm.DB, userID uint, req *model.OrderCreateRequest, cartItems []model.CartItem) (*model.Order, error) {
	// 更新商品价格为当前价格
	for i := range cartItems {
		currentPrice := cartItems[i].Product.Price
		if cartItems[i].SKUID > 0 && cartItems[i].SKU != nil {
			currentPrice = cartItems[i].SKU.Price
		}
		cartItems[i].Price = currentPrice
	}

	// 计算订单金额
	calculation, err := os.calculateOrderAmount(cartItems, req)
	if err != nil {
		return nil, fmt.Errorf("计算订单金额失败: %v", err)
	}

	// 创建订单
	order := &model.Order{
		OrderNo:         os.generateOrderNo(userID),
		UserID:          userID,
		Status:          model.OrderStatusPending,
		OrderType:       model.OrderTypeNormal,
		TotalAmount:     calculation.TotalAmount,
		PayableAmount:   calculation.PayableAmount,
		DiscountAmount:  calculation.DiscountAmount,
		ShippingFee:     calculation.ShippingFee,
		TaxAmount:       calculation.TaxAmount,
		CouponID:        req.CouponID,
		CouponAmount:    calculation.CouponAmount,
		PointsUsed:      req.PointsUsed,
		PointsAmount:    calculation.PointsAmount,
		ReceiverName:    req.ReceiverName,
		ReceiverPhone:   req.ReceiverPhone,
		ReceiverAddress: req.ReceiverAddress,
		ReceiverZipCode: req.ReceiverZipCode,
		Province:        req.Province,
		City:            req.City,
		District:        req.District,
		ShippingMethod:  req.ShippingMethod,
		BuyerMessage:    req.BuyerMessage,
		OrderTime:       time.Now(),
		PayExpireTime:   os.getPayExpireTime(),
		RefundStatus:    model.RefundStatusNone,
	}

	if err := tx.Create(order).Error; err != nil {
		return nil, fmt.Errorf("创建订单失败: %v", err)
	}

	// 创建订单商品项
	var orderItems []model.OrderItem
	for _, cartItem := range cartItems {
		orderItem := model.OrderItem{
			OrderID:      order.ID,
			ProductID:    cartItem.ProductID,
			SKUID:        cartItem.SKUID,
			Quantity:     cartItem.Quantity,
			ProductName:  cartItem.Product.Name,
			ProductImage: cartItem.Product.GetMainImage(),
			Price:        cartItem.Price,
			TotalPrice:   cartItem.Price.Mul(decimal.NewFromInt(int64(cartItem.Quantity))),
			RefundStatus: model.RefundStatusNone,
		}

		// 如果有SKU，填充SKU信息
		if cartItem.SKU != nil {
			orderItem.SKUName = cartItem.SKU.Name
			orderItem.SKUImage = cartItem.SKU.Image
			orderItem.SKUAttrs = cartItem.SKU.Attributes
		}

		orderItems = append(orderItems, orderItem)
	}

	if err := tx.Create(&orderItems).Error; err != nil {
		return nil, fmt.Errorf("创建订单商品失败: %v", err)
	}

	// 记录订单状态日志
	statusLog := &model.OrderStatusLog{
		OrderID:      order.ID,
		ToStatus:     model.OrderStatusPending,
		OperatorID:   userID,
		OperatorType: model.OperatorTypeUser,
		Reason:       "用户下单",
		Remark:       "订单创建成功",
	}

	if err := tx.Create(statusLog).Error; err != nil {
		return nil, fmt.Errorf("记录订单日志失败: %v", err)
	}

	return order, nil
}

// calculateOrderAmount 计算订单金额
func (os *OrderService) calculateOrderAmount(cartItems []model.CartItem, req *model.OrderCreateRequest) (*OrderCalculation, error) {
	calculation := &OrderCalculation{
		TotalAmount:    decimal.Zero,
		DiscountAmount: decimal.Zero,
		ShippingFee:    decimal.Zero,
		TaxAmount:      decimal.Zero,
		CouponAmount:   decimal.Zero,
		PointsAmount:   decimal.Zero,
	}

	// 计算商品总金额
	for _, item := range cartItems {
		itemTotal := item.Price.Mul(decimal.NewFromInt(int64(item.Quantity)))
		calculation.TotalAmount = calculation.TotalAmount.Add(itemTotal)
	}

	// 计算优惠券折扣
	if req.CouponID > 0 {
		couponAmount, err := os.applyCoupon(req.CouponID, calculation.TotalAmount)
		if err != nil {
			return nil, fmt.Errorf("应用优惠券失败: %v", err)
		}
		calculation.CouponAmount = couponAmount
		calculation.DiscountAmount = calculation.DiscountAmount.Add(couponAmount)
	}

	// 计算积分抵扣
	if req.PointsUsed > 0 {
		pointsAmount := os.calculatePointsAmount(req.PointsUsed)
		calculation.PointsAmount = pointsAmount
		calculation.DiscountAmount = calculation.DiscountAmount.Add(pointsAmount)
	}

	// 计算运费
	calculation.ShippingFee = os.calculateShippingFee(calculation.TotalAmount, req.Province)

	// 计算应付金额
	calculation.PayableAmount = calculation.TotalAmount.
		Sub(calculation.DiscountAmount).
		Add(calculation.ShippingFee).
		Add(calculation.TaxAmount)

	// 确保应付金额不为负数
	if calculation.PayableAmount.LessThan(decimal.Zero) {
		calculation.PayableAmount = decimal.Zero
	}

	return calculation, nil
}

// deductStock 扣减库存
func (os *OrderService) deductStock(tx *gorm.DB, cartItems []model.CartItem) error {
	for _, item := range cartItems {
		if item.SKUID > 0 {
			// 扣减SKU库存
			result := tx.Model(&model.ProductSKU{}).
				Where("id = ? AND stock >= ?", item.SKUID, item.Quantity).
				UpdateColumn("stock", gorm.Expr("stock - ?", item.Quantity))

			if result.Error != nil {
				return fmt.Errorf("扣减SKU库存失败: %v", result.Error)
			}

			if result.RowsAffected == 0 {
				return fmt.Errorf("SKU库存不足")
			}
		} else {
			// 扣减商品库存
			result := tx.Model(&model.Product{}).
				Where("id = ? AND stock >= ?", item.ProductID, item.Quantity).
				UpdateColumns(map[string]interface{}{
					"stock":      gorm.Expr("stock - ?", item.Quantity),
					"sold_count": gorm.Expr("sold_count + ?", item.Quantity),
				})

			if result.Error != nil {
				return fmt.Errorf("扣减商品库存失败: %v", result.Error)
			}

			if result.RowsAffected == 0 {
				return fmt.Errorf("商品库存不足")
			}
		}
	}

	return nil
}

// deductStockWithInventoryService 使用库存服务扣减库存
func (os *OrderService) deductStockWithInventoryService(cartItems []model.CartItem) error {
	// 构建库存扣减请求
	var requests []inventory.StockDeductionRequest
	for _, item := range cartItems {
		req := inventory.StockDeductionRequest{
			ProductID: item.ProductID,
			SKUID:     item.SKUID,
			Quantity:  item.Quantity,
		}
		requests = append(requests, req)
	}

	// 使用库存服务扣减库存
	results, err := os.inventoryService.DeductStockWithOptimisticLock(requests)
	if err != nil {
		return err
	}

	// 检查扣减结果
	for _, result := range results {
		if !result.Success {
			return fmt.Errorf("库存扣减失败: %s", result.Error)
		}
	}

	return nil
}

// applyCoupon 应用优惠券
func (os *OrderService) applyCoupon(couponID uint, totalAmount decimal.Decimal) (decimal.Decimal, error) {
	// 这里简化处理，实际应该查询优惠券表
	// 模拟优惠券数据
	coupons := map[uint]struct {
		Name        string
		Type        string // fixed, percentage
		Value       decimal.Decimal
		MinAmount   decimal.Decimal
		MaxDiscount decimal.Decimal
		Used        bool
	}{
		1: {"新用户专享券", "fixed", decimal.NewFromFloat(10.0), decimal.NewFromFloat(50.0), decimal.Zero, false},
		2: {"满100减20", "fixed", decimal.NewFromFloat(20.0), decimal.NewFromFloat(100.0), decimal.Zero, false},
		3: {"9折优惠券", "percentage", decimal.NewFromFloat(0.1), decimal.NewFromFloat(30.0), decimal.NewFromFloat(50.0), false},
	}

	coupon, exists := coupons[couponID]
	if !exists {
		return decimal.Zero, fmt.Errorf("优惠券不存在")
	}

	if coupon.Used {
		return decimal.Zero, fmt.Errorf("优惠券已使用")
	}

	if totalAmount.LessThan(coupon.MinAmount) {
		return decimal.Zero, fmt.Errorf("未达到优惠券使用门槛：%.2f元", coupon.MinAmount.InexactFloat64())
	}

	var discount decimal.Decimal
	if coupon.Type == "fixed" {
		discount = coupon.Value
	} else if coupon.Type == "percentage" {
		discount = totalAmount.Mul(coupon.Value)
		if coupon.MaxDiscount.GreaterThan(decimal.Zero) && discount.GreaterThan(coupon.MaxDiscount) {
			discount = coupon.MaxDiscount
		}
	}

	return discount, nil
}

// calculatePointsAmount 计算积分抵扣金额
func (os *OrderService) calculatePointsAmount(points int) decimal.Decimal {
	// 简化处理：100积分=1元
	return decimal.NewFromInt(int64(points)).Div(decimal.NewFromInt(100))
}

// calculateShippingFee 计算运费
func (os *OrderService) calculateShippingFee(totalAmount decimal.Decimal, province string) decimal.Decimal {
	// 满99元包邮
	freeShippingThreshold := decimal.NewFromFloat(99.0)
	if totalAmount.GreaterThanOrEqual(freeShippingThreshold) {
		return decimal.Zero
	}

	// 基础运费
	baseFee := decimal.NewFromFloat(8.0)

	// 根据地区调整运费
	switch province {
	case "新疆", "西藏", "内蒙古":
		return baseFee.Mul(decimal.NewFromFloat(1.5))
	case "海南":
		return baseFee.Mul(decimal.NewFromFloat(1.2))
	default:
		return baseFee
	}
}

// generateOrderNo 生成订单号
func (os *OrderService) generateOrderNo(userID uint) string {
	return fmt.Sprintf("ORD%d%d", time.Now().Unix(), userID)
}

// getPayExpireTime 获取支付超时时间
func (os *OrderService) getPayExpireTime() *time.Time {
	expireTime := time.Now().Add(30 * time.Minute) // 30分钟后过期
	return &expireTime
}

// OrderCalculation 订单计算结果
type OrderCalculation struct {
	TotalAmount    decimal.Decimal `json:"total_amount"`
	DiscountAmount decimal.Decimal `json:"discount_amount"`
	ShippingFee    decimal.Decimal `json:"shipping_fee"`
	TaxAmount      decimal.Decimal `json:"tax_amount"`
	CouponAmount   decimal.Decimal `json:"coupon_amount"`
	PointsAmount   decimal.Decimal `json:"points_amount"`
	PayableAmount  decimal.Decimal `json:"payable_amount"`
}

// 全局订单服务实例
var globalOrderService *OrderService

// InitGlobalOrderService 初始化全局订单服务
func InitGlobalOrderService(db *gorm.DB, cartService *cart.CartService, calculationService *cart.CalculationService, inventoryService *inventory.InventoryService) {
	globalOrderService = NewOrderService(db, cartService, calculationService, inventoryService)
}

// GetGlobalOrderService 获取全局订单服务
func GetGlobalOrderService() *OrderService {
	return globalOrderService
}
