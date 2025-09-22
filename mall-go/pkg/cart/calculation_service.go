package cart

import (
	"fmt"
	"math"

	"mall-go/internal/model"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// CalculationService 购物车计算服务
type CalculationService struct {
	db *gorm.DB
}

// NewCalculationService 创建购物车计算服务
func NewCalculationService(db *gorm.DB) *CalculationService {
	return &CalculationService{
		db: db,
	}
}

// CartCalculation 购物车计算结果
type CartCalculation struct {
	// 基础金额
	SubtotalAmount decimal.Decimal `json:"subtotal_amount"` // 小计金额
	SelectedAmount decimal.Decimal `json:"selected_amount"` // 选中商品金额

	// 优惠信息
	CouponDiscount    decimal.Decimal `json:"coupon_discount"`    // 优惠券折扣
	PromotionDiscount decimal.Decimal `json:"promotion_discount"` // 促销折扣
	MemberDiscount    decimal.Decimal `json:"member_discount"`    // 会员折扣
	TotalDiscount     decimal.Decimal `json:"total_discount"`     // 总折扣

	// 运费信息
	ShippingFee           decimal.Decimal `json:"shipping_fee"`            // 运费
	FreeShippingThreshold decimal.Decimal `json:"free_shipping_threshold"` // 包邮门槛

	// 税费信息
	TaxAmount decimal.Decimal `json:"tax_amount"` // 税费
	TaxRate   decimal.Decimal `json:"tax_rate"`   // 税率

	// 最终金额
	PayableAmount decimal.Decimal `json:"payable_amount"` // 应付金额
	SavedAmount   decimal.Decimal `json:"saved_amount"`   // 节省金额

	// 商品统计
	ItemCount        int `json:"item_count"`        // 商品种类数
	TotalQuantity    int `json:"total_quantity"`    // 商品总数量
	SelectedCount    int `json:"selected_count"`    // 选中商品种类数
	SelectedQuantity int `json:"selected_quantity"` // 选中商品数量

	// 重量信息
	TotalWeight decimal.Decimal `json:"total_weight"` // 总重量

	// 积分信息
	EarnPoints     int             `json:"earn_points"`     // 可获得积分
	UsedPoints     int             `json:"used_points"`     // 使用积分
	PointsDiscount decimal.Decimal `json:"points_discount"` // 积分抵扣金额
}

// PromotionRule 促销规则
type PromotionRule struct {
	ID          uint            `json:"id"`
	Name        string          `json:"name"`
	Type        string          `json:"type"`         // full_reduction, discount, gift
	Threshold   decimal.Decimal `json:"threshold"`    // 门槛金额
	Discount    decimal.Decimal `json:"discount"`     // 折扣金额或比例
	MaxDiscount decimal.Decimal `json:"max_discount"` // 最大折扣金额
	StartTime   string          `json:"start_time"`
	EndTime     string          `json:"end_time"`
	Status      string          `json:"status"`
}

// ShippingRule 运费规则
type ShippingRule struct {
	ID               uint            `json:"id"`
	Name             string          `json:"name"`
	BaseWeight       decimal.Decimal `json:"base_weight"`       // 首重
	BaseFee          decimal.Decimal `json:"base_fee"`          // 首重费用
	AdditionalWeight decimal.Decimal `json:"additional_weight"` // 续重
	AdditionalFee    decimal.Decimal `json:"additional_fee"`    // 续重费用
	FreeThreshold    decimal.Decimal `json:"free_threshold"`    // 包邮门槛
	MaxFee           decimal.Decimal `json:"max_fee"`           // 最高运费
}

// CalculateCart 计算购物车
func (cs *CalculationService) CalculateCart(cart *model.Cart, userID uint, region string) (*CartCalculation, error) {
	calculation := &CartCalculation{
		SubtotalAmount:        decimal.Zero,
		SelectedAmount:        decimal.Zero,
		CouponDiscount:        decimal.Zero,
		PromotionDiscount:     decimal.Zero,
		MemberDiscount:        decimal.Zero,
		TotalDiscount:         decimal.Zero,
		ShippingFee:           decimal.Zero,
		FreeShippingThreshold: decimal.NewFromFloat(99.0), // 默认99元包邮
		TaxAmount:             decimal.Zero,
		TaxRate:               decimal.Zero,
		PayableAmount:         decimal.Zero,
		SavedAmount:           decimal.Zero,
		TotalWeight:           decimal.Zero,
		EarnPoints:            0,
		UsedPoints:            0,
		PointsDiscount:        decimal.Zero,
	}

	// 计算基础金额和统计信息
	cs.calculateBasicAmount(cart, calculation)

	// 计算会员折扣
	if userID > 0 {
		cs.calculateMemberDiscount(userID, calculation)
	}

	// 计算促销折扣
	cs.calculatePromotionDiscount(cart, calculation)

	// 计算运费
	cs.calculateShippingFee(cart, region, calculation)

	// 计算税费
	cs.calculateTax(calculation)

	// 计算积分
	cs.calculatePoints(userID, calculation)

	// 计算最终金额
	cs.calculateFinalAmount(calculation)

	return calculation, nil
}

// calculateBasicAmount 计算基础金额
func (cs *CalculationService) calculateBasicAmount(cart *model.Cart, calc *CartCalculation) {
	for _, item := range cart.Items {
		if item.Status == model.CartItemStatusNormal {
			calc.ItemCount++
			calc.TotalQuantity += item.Quantity

			itemTotal := item.Price.Mul(decimal.NewFromInt(int64(item.Quantity)))
			calc.SubtotalAmount = calc.SubtotalAmount.Add(itemTotal)

			if item.Selected {
				calc.SelectedCount++
				calc.SelectedQuantity += item.Quantity
				calc.SelectedAmount = calc.SelectedAmount.Add(itemTotal)
			}

			// 计算重量
			if item.Product != nil {
				weight := item.Product.Weight.Mul(decimal.NewFromInt(int64(item.Quantity)))
				calc.TotalWeight = calc.TotalWeight.Add(weight)
			}
		}
	}
}

// calculateMemberDiscount 计算会员折扣
func (cs *CalculationService) calculateMemberDiscount(userID uint, calc *CartCalculation) {
	// 查询用户会员等级
	var user model.User
	if err := cs.db.First(&user, userID).Error; err != nil {
		return
	}

	// 根据会员等级计算折扣
	discountRate := decimal.Zero
	switch user.Role {
	case "vip1":
		discountRate = decimal.NewFromFloat(0.05) // 5%折扣
	case "vip2":
		discountRate = decimal.NewFromFloat(0.08) // 8%折扣
	case "vip3":
		discountRate = decimal.NewFromFloat(0.10) // 10%折扣
	}

	if discountRate.GreaterThan(decimal.Zero) {
		calc.MemberDiscount = calc.SelectedAmount.Mul(discountRate)
	}
}

// calculatePromotionDiscount 计算促销折扣
func (cs *CalculationService) calculatePromotionDiscount(cart *model.Cart, calc *CartCalculation) {
	// 这里简化处理，实际应该查询数据库中的促销规则
	promotions := []PromotionRule{
		{
			ID:        1,
			Name:      "满100减10",
			Type:      "full_reduction",
			Threshold: decimal.NewFromFloat(100.0),
			Discount:  decimal.NewFromFloat(10.0),
			Status:    "active",
		},
		{
			ID:        2,
			Name:      "满200减25",
			Type:      "full_reduction",
			Threshold: decimal.NewFromFloat(200.0),
			Discount:  decimal.NewFromFloat(25.0),
			Status:    "active",
		},
		{
			ID:        3,
			Name:      "满500减60",
			Type:      "full_reduction",
			Threshold: decimal.NewFromFloat(500.0),
			Discount:  decimal.NewFromFloat(60.0),
			Status:    "active",
		},
	}

	// 找到最优促销规则
	bestDiscount := decimal.Zero
	for _, promotion := range promotions {
		if promotion.Status == "active" && calc.SelectedAmount.GreaterThanOrEqual(promotion.Threshold) {
			if promotion.Type == "full_reduction" {
				if promotion.Discount.GreaterThan(bestDiscount) {
					bestDiscount = promotion.Discount
				}
			} else if promotion.Type == "discount" {
				discount := calc.SelectedAmount.Mul(promotion.Discount)
				if promotion.MaxDiscount.GreaterThan(decimal.Zero) && discount.GreaterThan(promotion.MaxDiscount) {
					discount = promotion.MaxDiscount
				}
				if discount.GreaterThan(bestDiscount) {
					bestDiscount = discount
				}
			}
		}
	}

	calc.PromotionDiscount = bestDiscount
}

// calculateShippingFee 计算运费
func (cs *CalculationService) calculateShippingFee(cart *model.Cart, region string, calc *CartCalculation) {
	// 如果选中商品金额达到包邮门槛，免运费
	if calc.SelectedAmount.GreaterThanOrEqual(calc.FreeShippingThreshold) {
		calc.ShippingFee = decimal.Zero
		return
	}

	// 简化的运费计算规则
	shippingRule := ShippingRule{
		BaseWeight:       decimal.NewFromFloat(1.0),  // 首重1kg
		BaseFee:          decimal.NewFromFloat(8.0),  // 首重8元
		AdditionalWeight: decimal.NewFromFloat(1.0),  // 续重1kg
		AdditionalFee:    decimal.NewFromFloat(3.0),  // 续重3元
		MaxFee:           decimal.NewFromFloat(50.0), // 最高50元
	}

	// 根据地区调整运费（简化处理）
	regionMultiplier := decimal.NewFromFloat(1.0)
	switch region {
	case "remote": // 偏远地区
		regionMultiplier = decimal.NewFromFloat(1.5)
	case "international": // 国际
		regionMultiplier = decimal.NewFromFloat(3.0)
	}

	// 计算运费
	if calc.TotalWeight.LessThanOrEqual(shippingRule.BaseWeight) {
		calc.ShippingFee = shippingRule.BaseFee.Mul(regionMultiplier)
	} else {
		additionalWeight := calc.TotalWeight.Sub(shippingRule.BaseWeight)
		additionalUnits := decimal.NewFromFloat(math.Ceil(additionalWeight.InexactFloat64() / shippingRule.AdditionalWeight.InexactFloat64()))
		additionalFee := additionalUnits.Mul(shippingRule.AdditionalFee)
		calc.ShippingFee = shippingRule.BaseFee.Add(additionalFee).Mul(regionMultiplier)
	}

	// 限制最高运费
	if calc.ShippingFee.GreaterThan(shippingRule.MaxFee.Mul(regionMultiplier)) {
		calc.ShippingFee = shippingRule.MaxFee.Mul(regionMultiplier)
	}
}

// calculateTax 计算税费
func (cs *CalculationService) calculateTax(calc *CartCalculation) {
	// 简化处理，实际应该根据商品类别和地区计算税费
	// 这里假设统一税率为0%（国内一般商品不单独收税）
	calc.TaxRate = decimal.Zero
	calc.TaxAmount = decimal.Zero
}

// calculatePoints 计算积分
func (cs *CalculationService) calculatePoints(userID uint, calc *CartCalculation) {
	if userID == 0 {
		return
	}

	// 简化处理：每消费1元获得1积分
	calc.EarnPoints = int(calc.SelectedAmount.InexactFloat64())

	// 这里可以添加积分使用逻辑
	// calc.UsedPoints = ...
	// calc.PointsDiscount = decimal.NewFromInt(int64(calc.UsedPoints)).Div(decimal.NewFromInt(100)) // 100积分=1元
}

// calculateFinalAmount 计算最终金额
func (cs *CalculationService) calculateFinalAmount(calc *CartCalculation) {
	// 计算总折扣
	calc.TotalDiscount = calc.CouponDiscount.Add(calc.PromotionDiscount).Add(calc.MemberDiscount).Add(calc.PointsDiscount)

	// 计算应付金额
	calc.PayableAmount = calc.SelectedAmount.Sub(calc.TotalDiscount).Add(calc.ShippingFee).Add(calc.TaxAmount)

	// 确保应付金额不为负数
	if calc.PayableAmount.LessThan(decimal.Zero) {
		calc.PayableAmount = decimal.Zero
	}

	// 计算节省金额
	calc.SavedAmount = calc.TotalDiscount
}

// CalculateCartWithCoupon 使用优惠券计算购物车
func (cs *CalculationService) CalculateCartWithCoupon(cart *model.Cart, userID uint, region string, couponID uint) (*CartCalculation, error) {
	// 先进行基础计算
	calculation, err := cs.CalculateCart(cart, userID, region)
	if err != nil {
		return nil, err
	}

	// 应用优惠券
	if couponID > 0 {
		couponDiscount, err := cs.applyCoupon(couponID, userID, calculation.SelectedAmount)
		if err != nil {
			return calculation, err // 返回基础计算结果，但包含错误信息
		}
		calculation.CouponDiscount = couponDiscount
	}

	// 重新计算最终金额
	cs.calculateFinalAmount(calculation)

	return calculation, nil
}

// applyCoupon 应用优惠券
func (cs *CalculationService) applyCoupon(couponID, userID uint, selectedAmount decimal.Decimal) (decimal.Decimal, error) {
	// 这里简化处理，实际应该查询数据库中的优惠券信息
	// 验证优惠券是否有效、是否已使用、是否满足使用条件等

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

	if selectedAmount.LessThan(coupon.MinAmount) {
		return decimal.Zero, fmt.Errorf("未达到优惠券使用门槛：%.2f元", coupon.MinAmount.InexactFloat64())
	}

	var discount decimal.Decimal
	if coupon.Type == "fixed" {
		discount = coupon.Value
	} else if coupon.Type == "percentage" {
		discount = selectedAmount.Mul(coupon.Value)
		if coupon.MaxDiscount.GreaterThan(decimal.Zero) && discount.GreaterThan(coupon.MaxDiscount) {
			discount = coupon.MaxDiscount
		}
	}

	return discount, nil
}

// EstimateShipping 估算运费
func (cs *CalculationService) EstimateShipping(cart *model.Cart, region string) (decimal.Decimal, error) {
	calc := &CartCalculation{
		SelectedAmount:        decimal.Zero,
		TotalWeight:           decimal.Zero,
		FreeShippingThreshold: decimal.NewFromFloat(99.0),
	}

	// 计算选中商品金额和重量
	for _, item := range cart.Items {
		if item.Selected && item.Status == model.CartItemStatusNormal {
			itemTotal := item.Price.Mul(decimal.NewFromInt(int64(item.Quantity)))
			calc.SelectedAmount = calc.SelectedAmount.Add(itemTotal)

			if item.Product != nil {
				weight := item.Product.Weight.Mul(decimal.NewFromInt(int64(item.Quantity)))
				calc.TotalWeight = calc.TotalWeight.Add(weight)
			}
		}
	}

	// 计算运费
	cs.calculateShippingFee(cart, region, calc)

	return calc.ShippingFee, nil
}

// GetPromotionSuggestions 获取促销建议
func (cs *CalculationService) GetPromotionSuggestions(selectedAmount decimal.Decimal) []map[string]interface{} {
	suggestions := []map[string]interface{}{}

	// 满减促销建议
	promotions := []struct {
		Threshold decimal.Decimal
		Discount  decimal.Decimal
		Name      string
	}{
		{decimal.NewFromFloat(100.0), decimal.NewFromFloat(10.0), "满100减10"},
		{decimal.NewFromFloat(200.0), decimal.NewFromFloat(25.0), "满200减25"},
		{decimal.NewFromFloat(500.0), decimal.NewFromFloat(60.0), "满500减60"},
	}

	for _, promo := range promotions {
		if selectedAmount.LessThan(promo.Threshold) {
			needAmount := promo.Threshold.Sub(selectedAmount)
			suggestions = append(suggestions, map[string]interface{}{
				"type":        "promotion",
				"name":        promo.Name,
				"need_amount": needAmount,
				"discount":    promo.Discount,
				"message":     fmt.Sprintf("再买%.2f元即可享受%s优惠", needAmount.InexactFloat64(), promo.Name),
			})
			break // 只推荐最近的一个促销
		}
	}

	// 包邮建议
	freeShippingThreshold := decimal.NewFromFloat(99.0)
	if selectedAmount.LessThan(freeShippingThreshold) {
		needAmount := freeShippingThreshold.Sub(selectedAmount)
		suggestions = append(suggestions, map[string]interface{}{
			"type":        "free_shipping",
			"name":        "包邮",
			"need_amount": needAmount,
			"message":     fmt.Sprintf("再买%.2f元即可包邮", needAmount.InexactFloat64()),
		})
	}

	return suggestions
}

// 全局购物车计算服务实例
var globalCalculationService *CalculationService

// InitGlobalCalculationService 初始化全局购物车计算服务
func InitGlobalCalculationService(db *gorm.DB) {
	globalCalculationService = NewCalculationService(db)
}

// GetGlobalCalculationService 获取全局购物车计算服务
func GetGlobalCalculationService() *CalculationService {
	return globalCalculationService
}
