package cache

import (
	"encoding/json"
	"fmt"
	"mall-go/internal/model"
	"mall-go/pkg/logger"
	"time"

	"github.com/shopspring/decimal"
)

// CartCacheService 购物车缓存服务
type CartCacheService struct {
	cacheManager CacheManager
	keyManager   *CacheKeyManager
}

// NewCartCacheService 创建购物车缓存服务
func NewCartCacheService(cacheManager CacheManager, keyManager *CacheKeyManager) *CartCacheService {
	return &CartCacheService{
		cacheManager: cacheManager,
		keyManager:   keyManager,
	}
}

// CartCacheData 购物车缓存数据结构
type CartCacheData struct {
	CartID      uint                  `json:"cart_id"`
	UserID      uint                  `json:"user_id"`
	SessionID   string                `json:"session_id"`
	Status      string                `json:"status"`
	ItemCount   int                   `json:"item_count"`
	TotalQty    int                   `json:"total_qty"`
	TotalAmount string                `json:"total_amount"` // 使用字符串存储decimal
	Items       []CartItemCacheData   `json:"items"`
	Summary     *CartSummaryCacheData `json:"summary,omitempty"`
	CachedAt    time.Time             `json:"cached_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	Version     int                   `json:"version"`
}

// CartItemCacheData 购物车商品项缓存数据结构
type CartItemCacheData struct {
	ID           uint      `json:"id"`
	CartID       uint      `json:"cart_id"`
	ProductID    uint      `json:"product_id"`
	SKUID        uint      `json:"sku_id"`
	Quantity     int       `json:"quantity"`
	Price        string    `json:"price"` // 使用字符串存储decimal
	ProductName  string    `json:"product_name"`
	ProductImage string    `json:"product_image"`
	SKUName      string    `json:"sku_name"`
	SKUImage     string    `json:"sku_image"`
	SKUAttrs     string    `json:"sku_attrs"`
	Selected     bool      `json:"selected"`
	Status       string    `json:"status"`
	CachedAt     time.Time `json:"cached_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Version      int       `json:"version"`
}

// CartSummaryCacheData 购物车汇总缓存数据结构
type CartSummaryCacheData struct {
	ItemCount      int                 `json:"item_count"`
	TotalQty       int                 `json:"total_qty"`
	SelectedCount  int                 `json:"selected_count"`
	SelectedQty    int                 `json:"selected_qty"`
	TotalAmount    string              `json:"total_amount"`    // 使用字符串存储decimal
	SelectedAmount string              `json:"selected_amount"` // 使用字符串存储decimal
	DiscountAmount string              `json:"discount_amount"` // 使用字符串存储decimal
	ShippingFee    string              `json:"shipping_fee"`    // 使用字符串存储decimal
	FinalAmount    string              `json:"final_amount"`    // 使用字符串存储decimal
	InvalidItems   []CartItemCacheData `json:"invalid_items"`
	CachedAt       time.Time           `json:"cached_at"`
}

// ConvertToCartCacheData 转换为购物车缓存数据
func ConvertToCartCacheData(cart *model.Cart) *CartCacheData {
	now := time.Now()

	// 转换购物车商品项
	items := make([]CartItemCacheData, len(cart.Items))
	for i, item := range cart.Items {
		items[i] = CartItemCacheData{
			ID:           item.ID,
			CartID:       item.CartID,
			ProductID:    item.ProductID,
			SKUID:        item.SKUID,
			Quantity:     item.Quantity,
			Price:        item.Price.String(),
			ProductName:  item.ProductName,
			ProductImage: item.ProductImage,
			SKUName:      item.SKUName,
			SKUImage:     item.SKUImage,
			SKUAttrs:     item.SKUAttrs,
			Selected:     item.Selected,
			Status:       item.Status,
			CachedAt:     now,
			UpdatedAt:    item.UpdatedAt,
			Version:      item.Version,
		}
	}

	return &CartCacheData{
		CartID:      cart.ID,
		UserID:      cart.UserID,
		SessionID:   cart.SessionID,
		Status:      cart.Status,
		ItemCount:   cart.ItemCount,
		TotalQty:    cart.TotalQty,
		TotalAmount: cart.TotalAmount.String(),
		Items:       items,
		CachedAt:    now,
		UpdatedAt:   cart.UpdatedAt,
		Version:     cart.Version,
	}
}

// ConvertToCartSummaryCacheData 转换为购物车汇总缓存数据
func ConvertToCartSummaryCacheData(summary *model.CartSummary) *CartSummaryCacheData {
	now := time.Now()

	// 转换失效商品列表
	invalidItems := make([]CartItemCacheData, len(summary.InvalidItems))
	for i, item := range summary.InvalidItems {
		invalidItems[i] = CartItemCacheData{
			ID:           item.ID,
			CartID:       item.CartID,
			ProductID:    item.ProductID,
			SKUID:        item.SKUID,
			Quantity:     item.Quantity,
			Price:        item.Price.String(),
			ProductName:  item.ProductName,
			ProductImage: item.ProductImage,
			SKUName:      item.SKUName,
			SKUImage:     item.SKUImage,
			SKUAttrs:     item.SKUAttrs,
			Selected:     item.Selected,
			Status:       item.Status,
			CachedAt:     now,
			UpdatedAt:    item.UpdatedAt,
			Version:      item.Version,
		}
	}

	return &CartSummaryCacheData{
		ItemCount:      summary.ItemCount,
		TotalQty:       summary.TotalQty,
		SelectedCount:  summary.SelectedCount,
		SelectedQty:    summary.SelectedQty,
		TotalAmount:    summary.TotalAmount.String(),
		SelectedAmount: summary.SelectedAmount.String(),
		DiscountAmount: summary.DiscountAmount.String(),
		ShippingFee:    summary.ShippingFee.String(),
		FinalAmount:    summary.FinalAmount.String(),
		InvalidItems:   invalidItems,
		CachedAt:       now,
	}
}

// GetUserCart 获取用户购物车缓存
func (ccs *CartCacheService) GetUserCart(userID uint) (*CartCacheData, error) {
	key := ccs.keyManager.GenerateUserCartKey(userID)

	result, err := ccs.cacheManager.Get(key)
	if err != nil {
		logger.Error(fmt.Sprintf("获取用户购物车缓存失败: UserID=%d, Error=%v", userID, err))
		return nil, fmt.Errorf("获取用户购物车缓存失败: %w", err)
	}

	if result == nil {
		return nil, nil // 缓存未命中
	}

	// 反序列化
	var cartData CartCacheData
	if err := json.Unmarshal([]byte(result.(string)), &cartData); err != nil {
		logger.Error(fmt.Sprintf("用户购物车数据反序列化失败: UserID=%d, Error=%v", userID, err))
		return nil, fmt.Errorf("用户购物车数据反序列化失败: %w", err)
	}

	logger.Info(fmt.Sprintf("用户购物车缓存命中: UserID=%d, CartID=%d, ItemCount=%d",
		userID, cartData.CartID, cartData.ItemCount))
	return &cartData, nil
}

// SetUserCart 设置用户购物车缓存
func (ccs *CartCacheService) SetUserCart(cart *model.Cart) error {
	key := ccs.keyManager.GenerateUserCartKey(cart.UserID)

	// 转换为缓存数据
	cacheData := ConvertToCartCacheData(cart)

	// 序列化
	data, err := json.Marshal(cacheData)
	if err != nil {
		logger.Error(fmt.Sprintf("用户购物车数据序列化失败: UserID=%d, Error=%v", cart.UserID, err))
		return fmt.Errorf("用户购物车数据序列化失败: %w", err)
	}

	// 设置缓存，使用cart类型的TTL（24小时）
	ttl := CacheTTL["cart"]
	if err := ccs.cacheManager.Set(key, string(data), ttl); err != nil {
		logger.Error(fmt.Sprintf("设置用户购物车缓存失败: UserID=%d, Error=%v", cart.UserID, err))
		return fmt.Errorf("设置用户购物车缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("用户购物车缓存设置成功: UserID=%d, CartID=%d, ItemCount=%d",
		cart.UserID, cart.ID, cart.ItemCount))
	return nil
}

// GetGuestCart 获取游客购物车缓存
func (ccs *CartCacheService) GetGuestCart(sessionID string) (*CartCacheData, error) {
	key := ccs.keyManager.GenerateGuestCartKey(sessionID)

	result, err := ccs.cacheManager.Get(key)
	if err != nil {
		logger.Error(fmt.Sprintf("获取游客购物车缓存失败: SessionID=%s, Error=%v", sessionID, err))
		return nil, fmt.Errorf("获取游客购物车缓存失败: %w", err)
	}

	if result == nil {
		return nil, nil // 缓存未命中
	}

	// 反序列化
	var cartData CartCacheData
	if err := json.Unmarshal([]byte(result.(string)), &cartData); err != nil {
		logger.Error(fmt.Sprintf("游客购物车数据反序列化失败: SessionID=%s, Error=%v", sessionID, err))
		return nil, fmt.Errorf("游客购物车数据反序列化失败: %w", err)
	}

	logger.Info(fmt.Sprintf("游客购物车缓存命中: SessionID=%s, CartID=%d, ItemCount=%d",
		sessionID, cartData.CartID, cartData.ItemCount))
	return &cartData, nil
}

// SetGuestCart 设置游客购物车缓存
func (ccs *CartCacheService) SetGuestCart(cart *model.Cart) error {
	if cart.SessionID == "" {
		return fmt.Errorf("游客购物车SessionID不能为空")
	}

	key := ccs.keyManager.GenerateGuestCartKey(cart.SessionID)

	// 转换为缓存数据
	cacheData := ConvertToCartCacheData(cart)

	// 序列化
	data, err := json.Marshal(cacheData)
	if err != nil {
		logger.Error(fmt.Sprintf("游客购物车数据序列化失败: SessionID=%s, Error=%v", cart.SessionID, err))
		return fmt.Errorf("游客购物车数据序列化失败: %w", err)
	}

	// 设置缓存，使用cart类型的TTL（24小时）
	ttl := CacheTTL["cart"]
	if err := ccs.cacheManager.Set(key, string(data), ttl); err != nil {
		logger.Error(fmt.Sprintf("设置游客购物车缓存失败: SessionID=%s, Error=%v", cart.SessionID, err))
		return fmt.Errorf("设置游客购物车缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("游客购物车缓存设置成功: SessionID=%s, CartID=%d, ItemCount=%d",
		cart.SessionID, cart.ID, cart.ItemCount))
	return nil
}

// GetCartSummary 获取购物车汇总缓存
func (ccs *CartCacheService) GetCartSummary(cartID uint) (*CartSummaryCacheData, error) {
	key := ccs.keyManager.GenerateCartSummaryKey(cartID)

	result, err := ccs.cacheManager.Get(key)
	if err != nil {
		logger.Error(fmt.Sprintf("获取购物车汇总缓存失败: CartID=%d, Error=%v", cartID, err))
		return nil, fmt.Errorf("获取购物车汇总缓存失败: %w", err)
	}

	if result == nil {
		return nil, nil // 缓存未命中
	}

	// 反序列化
	var summaryData CartSummaryCacheData
	if err := json.Unmarshal([]byte(result.(string)), &summaryData); err != nil {
		logger.Error(fmt.Sprintf("购物车汇总数据反序列化失败: CartID=%d, Error=%v", cartID, err))
		return nil, fmt.Errorf("购物车汇总数据反序列化失败: %w", err)
	}

	logger.Info(fmt.Sprintf("购物车汇总缓存命中: CartID=%d, ItemCount=%d, SelectedCount=%d",
		cartID, summaryData.ItemCount, summaryData.SelectedCount))
	return &summaryData, nil
}

// SetCartSummary 设置购物车汇总缓存
func (ccs *CartCacheService) SetCartSummary(cartID uint, summary *model.CartSummary) error {
	key := ccs.keyManager.GenerateCartSummaryKey(cartID)

	// 转换为缓存数据
	cacheData := ConvertToCartSummaryCacheData(summary)

	// 序列化
	data, err := json.Marshal(cacheData)
	if err != nil {
		logger.Error(fmt.Sprintf("购物车汇总数据序列化失败: CartID=%d, Error=%v", cartID, err))
		return fmt.Errorf("购物车汇总数据序列化失败: %w", err)
	}

	// 设置缓存，使用cart类型的TTL（24小时）
	ttl := CacheTTL["cart"]
	if err := ccs.cacheManager.Set(key, string(data), ttl); err != nil {
		logger.Error(fmt.Sprintf("设置购物车汇总缓存失败: CartID=%d, Error=%v", cartID, err))
		return fmt.Errorf("设置购物车汇总缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("购物车汇总缓存设置成功: CartID=%d, ItemCount=%d, SelectedCount=%d",
		cartID, summary.ItemCount, summary.SelectedCount))
	return nil
}

// GetCartItem 获取购物车商品项缓存
func (ccs *CartCacheService) GetCartItem(cartID uint, itemID uint) (*CartItemCacheData, error) {
	key := ccs.keyManager.GenerateCartItemKey(cartID, itemID)

	result, err := ccs.cacheManager.Get(key)
	if err != nil {
		logger.Error(fmt.Sprintf("获取购物车商品项缓存失败: CartID=%d, ItemID=%d, Error=%v", cartID, itemID, err))
		return nil, fmt.Errorf("获取购物车商品项缓存失败: %w", err)
	}

	if result == nil {
		return nil, nil // 缓存未命中
	}

	// 反序列化
	var itemData CartItemCacheData
	if err := json.Unmarshal([]byte(result.(string)), &itemData); err != nil {
		logger.Error(fmt.Sprintf("购物车商品项数据反序列化失败: CartID=%d, ItemID=%d, Error=%v", cartID, itemID, err))
		return nil, fmt.Errorf("购物车商品项数据反序列化失败: %w", err)
	}

	logger.Info(fmt.Sprintf("购物车商品项缓存命中: CartID=%d, ItemID=%d, ProductID=%d",
		cartID, itemID, itemData.ProductID))
	return &itemData, nil
}

// SetCartItem 设置购物车商品项缓存
func (ccs *CartCacheService) SetCartItem(item *model.CartItem) error {
	key := ccs.keyManager.GenerateCartItemKey(item.CartID, item.ID)

	// 转换为缓存数据
	now := time.Now()
	cacheData := CartItemCacheData{
		ID:           item.ID,
		CartID:       item.CartID,
		ProductID:    item.ProductID,
		SKUID:        item.SKUID,
		Quantity:     item.Quantity,
		Price:        item.Price.String(),
		ProductName:  item.ProductName,
		ProductImage: item.ProductImage,
		SKUName:      item.SKUName,
		SKUImage:     item.SKUImage,
		SKUAttrs:     item.SKUAttrs,
		Selected:     item.Selected,
		Status:       item.Status,
		CachedAt:     now,
		UpdatedAt:    item.UpdatedAt,
		Version:      item.Version,
	}

	// 序列化
	data, err := json.Marshal(cacheData)
	if err != nil {
		logger.Error(fmt.Sprintf("购物车商品项数据序列化失败: CartID=%d, ItemID=%d, Error=%v", item.CartID, item.ID, err))
		return fmt.Errorf("购物车商品项数据序列化失败: %w", err)
	}

	// 设置缓存，使用cart类型的TTL（24小时）
	ttl := CacheTTL["cart"]
	if err := ccs.cacheManager.Set(key, string(data), ttl); err != nil {
		logger.Error(fmt.Sprintf("设置购物车商品项缓存失败: CartID=%d, ItemID=%d, Error=%v", item.CartID, item.ID, err))
		return fmt.Errorf("设置购物车商品项缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("购物车商品项缓存设置成功: CartID=%d, ItemID=%d, ProductID=%d",
		item.CartID, item.ID, item.ProductID))
	return nil
}

// DeleteCartItem 删除购物车商品项缓存
func (ccs *CartCacheService) DeleteCartItem(cartID uint, itemID uint) error {
	key := ccs.keyManager.GenerateCartItemKey(cartID, itemID)

	if err := ccs.cacheManager.Delete(key); err != nil {
		logger.Error(fmt.Sprintf("删除购物车商品项缓存失败: CartID=%d, ItemID=%d, Error=%v", cartID, itemID, err))
		return fmt.Errorf("删除购物车商品项缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("购物车商品项缓存删除成功: CartID=%d, ItemID=%d", cartID, itemID))
	return nil
}

// UpdateCartItemQuantity 更新购物车商品项数量
func (ccs *CartCacheService) UpdateCartItemQuantity(cartID uint, itemID uint, quantity int) error {
	// 获取现有商品项
	itemData, err := ccs.GetCartItem(cartID, itemID)
	if err != nil {
		return fmt.Errorf("获取购物车商品项失败: %w", err)
	}

	if itemData == nil {
		return fmt.Errorf("购物车商品项不存在: CartID=%d, ItemID=%d", cartID, itemID)
	}

	// 更新数量
	itemData.Quantity = quantity
	itemData.CachedAt = time.Now()
	itemData.Version++

	// 序列化并保存
	key := ccs.keyManager.GenerateCartItemKey(cartID, itemID)
	data, err := json.Marshal(itemData)
	if err != nil {
		return fmt.Errorf("购物车商品项数据序列化失败: %w", err)
	}

	ttl := CacheTTL["cart"]
	if err := ccs.cacheManager.Set(key, string(data), ttl); err != nil {
		return fmt.Errorf("更新购物车商品项缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("购物车商品项数量更新成功: CartID=%d, ItemID=%d, Quantity=%d",
		cartID, itemID, quantity))
	return nil
}

// UpdateCartItemSelection 更新购物车商品项选中状态
func (ccs *CartCacheService) UpdateCartItemSelection(cartID uint, itemID uint, selected bool) error {
	// 获取现有商品项
	itemData, err := ccs.GetCartItem(cartID, itemID)
	if err != nil {
		return fmt.Errorf("获取购物车商品项失败: %w", err)
	}

	if itemData == nil {
		return fmt.Errorf("购物车商品项不存在: CartID=%d, ItemID=%d", cartID, itemID)
	}

	// 更新选中状态
	itemData.Selected = selected
	itemData.CachedAt = time.Now()
	itemData.Version++

	// 序列化并保存
	key := ccs.keyManager.GenerateCartItemKey(cartID, itemID)
	data, err := json.Marshal(itemData)
	if err != nil {
		return fmt.Errorf("购物车商品项数据序列化失败: %w", err)
	}

	ttl := CacheTTL["cart"]
	if err := ccs.cacheManager.Set(key, string(data), ttl); err != nil {
		return fmt.Errorf("更新购物车商品项缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("购物车商品项选中状态更新成功: CartID=%d, ItemID=%d, Selected=%v",
		cartID, itemID, selected))
	return nil
}

// BatchUpdateCartItems 批量更新购物车商品项
func (ccs *CartCacheService) BatchUpdateCartItems(updates []CartItemUpdate) error {
	for _, update := range updates {
		if update.Quantity > 0 {
			if err := ccs.UpdateCartItemQuantity(update.CartID, update.ItemID, update.Quantity); err != nil {
				logger.Error(fmt.Sprintf("批量更新购物车商品项数量失败: CartID=%d, ItemID=%d, Error=%v",
					update.CartID, update.ItemID, err))
				return fmt.Errorf("批量更新购物车商品项数量失败: %w", err)
			}
		}

		if err := ccs.UpdateCartItemSelection(update.CartID, update.ItemID, update.Selected); err != nil {
			logger.Error(fmt.Sprintf("批量更新购物车商品项选中状态失败: CartID=%d, ItemID=%d, Error=%v",
				update.CartID, update.ItemID, err))
			return fmt.Errorf("批量更新购物车商品项选中状态失败: %w", err)
		}
	}

	logger.Info(fmt.Sprintf("批量更新购物车商品项成功: 更新数量=%d", len(updates)))
	return nil
}

// BatchDeleteCartItems 批量删除购物车商品项
func (ccs *CartCacheService) BatchDeleteCartItems(cartID uint, itemIDs []uint) error {
	keys := make([]string, len(itemIDs))
	for i, itemID := range itemIDs {
		keys[i] = ccs.keyManager.GenerateCartItemKey(cartID, itemID)
	}

	if err := ccs.cacheManager.MDelete(keys); err != nil {
		logger.Error(fmt.Sprintf("批量删除购物车商品项缓存失败: CartID=%d, Error=%v", cartID, err))
		return fmt.Errorf("批量删除购物车商品项缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("批量删除购物车商品项缓存成功: CartID=%d, 删除数量=%d", cartID, len(itemIDs)))
	return nil
}

// DeleteUserCart 删除用户购物车缓存
func (ccs *CartCacheService) DeleteUserCart(userID uint) error {
	key := ccs.keyManager.GenerateUserCartKey(userID)

	if err := ccs.cacheManager.Delete(key); err != nil {
		logger.Error(fmt.Sprintf("删除用户购物车缓存失败: UserID=%d, Error=%v", userID, err))
		return fmt.Errorf("删除用户购物车缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("用户购物车缓存删除成功: UserID=%d", userID))
	return nil
}

// DeleteGuestCart 删除游客购物车缓存
func (ccs *CartCacheService) DeleteGuestCart(sessionID string) error {
	key := ccs.keyManager.GenerateGuestCartKey(sessionID)

	if err := ccs.cacheManager.Delete(key); err != nil {
		logger.Error(fmt.Sprintf("删除游客购物车缓存失败: SessionID=%s, Error=%v", sessionID, err))
		return fmt.Errorf("删除游客购物车缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("游客购物车缓存删除成功: SessionID=%s", sessionID))
	return nil
}

// DeleteCartSummary 删除购物车汇总缓存
func (ccs *CartCacheService) DeleteCartSummary(cartID uint) error {
	key := ccs.keyManager.GenerateCartSummaryKey(cartID)

	if err := ccs.cacheManager.Delete(key); err != nil {
		logger.Error(fmt.Sprintf("删除购物车汇总缓存失败: CartID=%d, Error=%v", cartID, err))
		return fmt.Errorf("删除购物车汇总缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("购物车汇总缓存删除成功: CartID=%d", cartID))
	return nil
}

// ExistsUserCart 检查用户购物车缓存是否存在
func (ccs *CartCacheService) ExistsUserCart(userID uint) bool {
	key := ccs.keyManager.GenerateUserCartKey(userID)
	return ccs.cacheManager.Exists(key)
}

// ExistsGuestCart 检查游客购物车缓存是否存在
func (ccs *CartCacheService) ExistsGuestCart(sessionID string) bool {
	key := ccs.keyManager.GenerateGuestCartKey(sessionID)
	return ccs.cacheManager.Exists(key)
}

// ExistsCartItem 检查购物车商品项缓存是否存在
func (ccs *CartCacheService) ExistsCartItem(cartID uint, itemID uint) bool {
	key := ccs.keyManager.GenerateCartItemKey(cartID, itemID)
	return ccs.cacheManager.Exists(key)
}

// GetUserCartTTL 获取用户购物车缓存TTL
func (ccs *CartCacheService) GetUserCartTTL(userID uint) (time.Duration, error) {
	key := ccs.keyManager.GenerateUserCartKey(userID)
	return ccs.cacheManager.TTL(key)
}

// RefreshUserCartTTL 刷新用户购物车缓存TTL
func (ccs *CartCacheService) RefreshUserCartTTL(userID uint) error {
	key := ccs.keyManager.GenerateUserCartKey(userID)
	ttl := CacheTTL["cart"]
	return ccs.cacheManager.Expire(key, ttl)
}

// CartItemUpdate 购物车商品项更新结构
type CartItemUpdate struct {
	CartID   uint `json:"cart_id"`
	ItemID   uint `json:"item_id"`
	Quantity int  `json:"quantity"`
	Selected bool `json:"selected"`
}

// CartConsistencyResult 购物车一致性检查结果
type CartConsistencyResult struct {
	IsConsistent bool                  `json:"is_consistent"`
	InvalidItems []CartItemCacheData   `json:"invalid_items"`
	PriceChanges []CartItemPriceChange `json:"price_changes"`
	StockIssues  []CartItemStockIssue  `json:"stock_issues"`
	CheckedAt    time.Time             `json:"checked_at"`
}

// CartItemPriceChange 购物车商品项价格变动
type CartItemPriceChange struct {
	ItemID       uint   `json:"item_id"`
	ProductID    uint   `json:"product_id"`
	ProductName  string `json:"product_name"`
	OldPrice     string `json:"old_price"`
	NewPrice     string `json:"new_price"`
	PriceChanged bool   `json:"price_changed"`
}

// CartItemStockIssue 购物车商品项库存问题
type CartItemStockIssue struct {
	ItemID          uint   `json:"item_id"`
	ProductID       uint   `json:"product_id"`
	ProductName     string `json:"product_name"`
	RequestQuantity int    `json:"request_quantity"`
	AvailableStock  int    `json:"available_stock"`
	StockSufficient bool   `json:"stock_sufficient"`
}

// ValidateCartConsistency 验证购物车数据一致性
func (ccs *CartCacheService) ValidateCartConsistency(cartData *CartCacheData, productCache *ProductCacheService, stockCache *StockCacheService) (*CartConsistencyResult, error) {
	result := &CartConsistencyResult{
		IsConsistent: true,
		InvalidItems: []CartItemCacheData{},
		PriceChanges: []CartItemPriceChange{},
		StockIssues:  []CartItemStockIssue{},
		CheckedAt:    time.Now(),
	}

	for _, item := range cartData.Items {
		// 检查商品是否存在
		productData, err := productCache.GetProduct(item.ProductID)
		if err != nil {
			logger.Error(fmt.Sprintf("获取商品信息失败: ProductID=%d, Error=%v", item.ProductID, err))
			continue
		}

		if productData == nil {
			// 商品不存在，标记为失效
			result.IsConsistent = false
			result.InvalidItems = append(result.InvalidItems, item)
			continue
		}

		// 检查价格变动
		currentPrice, err := decimal.NewFromString(productData.Price)
		if err != nil {
			logger.Error(fmt.Sprintf("解析商品价格失败: ProductID=%d, Price=%s, Error=%v",
				item.ProductID, productData.Price, err))
			continue
		}

		itemPrice, err := decimal.NewFromString(item.Price)
		if err != nil {
			logger.Error(fmt.Sprintf("解析购物车商品价格失败: ItemID=%d, Price=%s, Error=%v",
				item.ID, item.Price, err))
			continue
		}

		if !currentPrice.Equal(itemPrice) {
			result.IsConsistent = false
			result.PriceChanges = append(result.PriceChanges, CartItemPriceChange{
				ItemID:       item.ID,
				ProductID:    item.ProductID,
				ProductName:  item.ProductName,
				OldPrice:     item.Price,
				NewPrice:     productData.Price,
				PriceChanged: true,
			})
		}

		// 检查库存
		stockData, err := stockCache.GetStock(item.ProductID)
		if err != nil {
			logger.Error(fmt.Sprintf("获取商品库存失败: ProductID=%d, Error=%v", item.ProductID, err))
			continue
		}

		if stockData != nil && stockData.Stock < item.Quantity {
			result.IsConsistent = false
			result.StockIssues = append(result.StockIssues, CartItemStockIssue{
				ItemID:          item.ID,
				ProductID:       item.ProductID,
				ProductName:     item.ProductName,
				RequestQuantity: item.Quantity,
				AvailableStock:  stockData.Stock,
				StockSufficient: false,
			})
		}
	}

	logger.Info(fmt.Sprintf("购物车一致性检查完成: CartID=%d, IsConsistent=%v, InvalidItems=%d, PriceChanges=%d, StockIssues=%d",
		cartData.CartID, result.IsConsistent, len(result.InvalidItems), len(result.PriceChanges), len(result.StockIssues)))

	return result, nil
}

// UpdateCartItemStatus 更新购物车商品项状态
func (ccs *CartCacheService) UpdateCartItemStatus(cartID uint, itemID uint, status string) error {
	// 获取现有商品项
	itemData, err := ccs.GetCartItem(cartID, itemID)
	if err != nil {
		return fmt.Errorf("获取购物车商品项失败: %w", err)
	}

	if itemData == nil {
		return fmt.Errorf("购物车商品项不存在: CartID=%d, ItemID=%d", cartID, itemID)
	}

	// 更新状态
	itemData.Status = status
	itemData.CachedAt = time.Now()
	itemData.Version++

	// 序列化并保存
	key := ccs.keyManager.GenerateCartItemKey(cartID, itemID)
	data, err := json.Marshal(itemData)
	if err != nil {
		return fmt.Errorf("购物车商品项数据序列化失败: %w", err)
	}

	ttl := CacheTTL["cart"]
	if err := ccs.cacheManager.Set(key, string(data), ttl); err != nil {
		return fmt.Errorf("更新购物车商品项缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("购物车商品项状态更新成功: CartID=%d, ItemID=%d, Status=%s",
		cartID, itemID, status))
	return nil
}

// RefreshCartWithConsistencyCheck 刷新购物车并进行一致性检查
func (ccs *CartCacheService) RefreshCartWithConsistencyCheck(cartData *CartCacheData, productCache *ProductCacheService, stockCache *StockCacheService) (*CartCacheData, error) {
	// 进行一致性检查
	consistencyResult, err := ccs.ValidateCartConsistency(cartData, productCache, stockCache)
	if err != nil {
		return nil, fmt.Errorf("购物车一致性检查失败: %w", err)
	}

	// 更新失效商品状态
	for _, invalidItem := range consistencyResult.InvalidItems {
		if err := ccs.UpdateCartItemStatus(cartData.CartID, invalidItem.ID, model.CartItemStatusInvalid); err != nil {
			logger.Error(fmt.Sprintf("更新失效商品状态失败: CartID=%d, ItemID=%d, Error=%v",
				cartData.CartID, invalidItem.ID, err))
		}
	}

	// 更新价格变动商品状态
	for _, priceChange := range consistencyResult.PriceChanges {
		if err := ccs.UpdateCartItemStatus(cartData.CartID, priceChange.ItemID, model.CartItemStatusPriceChange); err != nil {
			logger.Error(fmt.Sprintf("更新价格变动商品状态失败: CartID=%d, ItemID=%d, Error=%v",
				cartData.CartID, priceChange.ItemID, err))
		}
	}

	// 更新库存不足商品状态
	for _, stockIssue := range consistencyResult.StockIssues {
		if err := ccs.UpdateCartItemStatus(cartData.CartID, stockIssue.ItemID, model.CartItemStatusOutOfStock); err != nil {
			logger.Error(fmt.Sprintf("更新库存不足商品状态失败: CartID=%d, ItemID=%d, Error=%v",
				cartData.CartID, stockIssue.ItemID, err))
		}
	}

	// 重新获取更新后的购物车数据
	var updatedCartData *CartCacheData
	if cartData.UserID > 0 {
		updatedCartData, err = ccs.GetUserCart(cartData.UserID)
	} else {
		updatedCartData, err = ccs.GetGuestCart(cartData.SessionID)
	}

	if err != nil {
		return nil, fmt.Errorf("获取更新后的购物车数据失败: %w", err)
	}

	logger.Info(fmt.Sprintf("购物车一致性检查和刷新完成: CartID=%d, 一致性=%v",
		cartData.CartID, consistencyResult.IsConsistent))

	return updatedCartData, nil
}
