package cache

import (
	"fmt"
	"strings"
	"time"
)

// CacheKeyManager 缓存键管理器
type CacheKeyManager struct {
	prefix    string
	separator string
}

// NewCacheKeyManager 创建缓存键管理器
func NewCacheKeyManager(prefix string) *CacheKeyManager {
	return &CacheKeyManager{
		prefix:    prefix,
		separator: ":",
	}
}

// 缓存键命名规范常量
const (
	// 商品相关键
	ProductInfoKey    = "product:{id}"           // 商品基础信息
	ProductStockKey   = "stock:{id}"             // 商品库存
	ProductPriceKey   = "price:{id}"             // 商品价格
	ProductViewKey    = "stats:view:{id}"        // 商品浏览量
	
	// 用户相关键
	UserSessionKey    = "user:session:{token}"   // 用户会话
	UserCartKey       = "cart:{user_id}"         // 购物车
	UserProfileKey    = "user:profile:{id}"      // 用户资料
	UserPreferenceKey = "user:pref:{id}"         // 用户偏好
	
	// 业务相关键
	CategoryProductsKey = "category:{id}:products" // 分类商品
	HotProductsKey     = "hot:products"            // 热门商品
	OrderKey          = "order:{id}"               // 订单信息
	OrderLockKey      = "lock:order:{id}"          // 订单锁
	
	// 统计相关键
	DailyStatsKey     = "stats:daily:{date}"     // 日统计
	HourlyStatsKey    = "stats:hourly:{hour}"    // 小时统计
	CounterKey        = "counter:{type}:{id}"    // 计数器
	
	// 缓存控制键
	CacheVersionKey   = "version:{type}"         // 缓存版本
	CacheLockKey      = "lock:{key}"             // 缓存锁
	CacheWarmupKey    = "warmup:{type}"          // 预热标记
)

// TTL配置映射
var CacheTTL = map[string]time.Duration{
	// 基础数据 - 较长TTL
	"product":     30 * time.Minute,  // 商品信息
	"category":    60 * time.Minute,  // 分类信息
	"user":        120 * time.Minute, // 用户资料
	
	// 动态数据 - 较短TTL
	"stock":       10 * time.Minute,  // 库存数据
	"price":       15 * time.Minute,  // 价格信息
	"cart":        24 * time.Hour,    // 购物车
	
	// 会话数据 - 中等TTL
	"session":     2 * time.Hour,     // 用户会话
	"token":       30 * time.Minute,  // 访问令牌
	
	// 统计数据 - 短TTL
	"stats":       5 * time.Minute,   // 实时统计
	"hot":         60 * time.Minute,  // 热点数据
	"counter":     1 * time.Hour,     // 计数器
	
	// 锁和控制 - 很短TTL
	"lock":        30 * time.Second,  // 分布式锁
	"version":     24 * time.Hour,    // 版本控制
	"warmup":      10 * time.Minute,  // 预热标记
}

// KeyBuilder 键构建器
type KeyBuilder struct {
	parts []string
}

// NewKeyBuilder 创建键构建器
func NewKeyBuilder() *KeyBuilder {
	return &KeyBuilder{
		parts: make([]string, 0),
	}
}

// Add 添加键部分
func (kb *KeyBuilder) Add(part string) *KeyBuilder {
	if part != "" {
		kb.parts = append(kb.parts, part)
	}
	return kb
}

// AddInt 添加整数键部分
func (kb *KeyBuilder) AddInt(value int) *KeyBuilder {
	return kb.Add(fmt.Sprintf("%d", value))
}

// AddUint 添加无符号整数键部分
func (kb *KeyBuilder) AddUint(value uint) *KeyBuilder {
	return kb.Add(fmt.Sprintf("%d", value))
}

// AddFormat 添加格式化键部分
func (kb *KeyBuilder) AddFormat(format string, args ...interface{}) *KeyBuilder {
	return kb.Add(fmt.Sprintf(format, args...))
}

// Build 构建最终键
func (kb *KeyBuilder) Build() string {
	return strings.Join(kb.parts, ":")
}

// BuildWithPrefix 构建带前缀的键
func (kb *KeyBuilder) BuildWithPrefix(prefix string) string {
	if prefix == "" {
		return kb.Build()
	}
	return prefix + ":" + kb.Build()
}

// 键生成函数

// GenerateProductKey 生成商品相关键
func (ckm *CacheKeyManager) GenerateProductKey(productID uint) string {
	return NewKeyBuilder().Add("product").AddUint(productID).BuildWithPrefix(ckm.prefix)
}

// GenerateProductStockKey 生成商品库存键
func (ckm *CacheKeyManager) GenerateProductStockKey(productID uint) string {
	return NewKeyBuilder().Add("stock").AddUint(productID).BuildWithPrefix(ckm.prefix)
}

// GenerateProductPriceKey 生成商品价格键
func (ckm *CacheKeyManager) GenerateProductPriceKey(productID uint) string {
	return NewKeyBuilder().Add("price").AddUint(productID).BuildWithPrefix(ckm.prefix)
}

// GenerateUserSessionKey 生成用户会话键
func (ckm *CacheKeyManager) GenerateUserSessionKey(token string) string {
	return NewKeyBuilder().Add("user").Add("session").Add(token).BuildWithPrefix(ckm.prefix)
}

// GenerateUserCartKey 生成用户购物车键
func (ckm *CacheKeyManager) GenerateUserCartKey(userID uint) string {
	return NewKeyBuilder().Add("cart").AddUint(userID).BuildWithPrefix(ckm.prefix)
}

// GenerateUserProfileKey 生成用户资料键
func (ckm *CacheKeyManager) GenerateUserProfileKey(userID uint) string {
	return NewKeyBuilder().Add("user").Add("profile").AddUint(userID).BuildWithPrefix(ckm.prefix)
}

// GenerateCategoryProductsKey 生成分类商品键
func (ckm *CacheKeyManager) GenerateCategoryProductsKey(categoryID uint) string {
	return NewKeyBuilder().Add("category").AddUint(categoryID).Add("products").BuildWithPrefix(ckm.prefix)
}

// GenerateHotProductsKey 生成热门商品键
func (ckm *CacheKeyManager) GenerateHotProductsKey() string {
	return NewKeyBuilder().Add("hot").Add("products").BuildWithPrefix(ckm.prefix)
}

// GenerateOrderKey 生成订单键
func (ckm *CacheKeyManager) GenerateOrderKey(orderID uint) string {
	return NewKeyBuilder().Add("order").AddUint(orderID).BuildWithPrefix(ckm.prefix)
}

// GenerateStatsKey 生成统计键
func (ckm *CacheKeyManager) GenerateStatsKey(statsType string, date string) string {
	return NewKeyBuilder().Add("stats").Add(statsType).Add(date).BuildWithPrefix(ckm.prefix)
}

// GenerateCounterKey 生成计数器键
func (ckm *CacheKeyManager) GenerateCounterKey(counterType string, id uint) string {
	return NewKeyBuilder().Add("counter").Add(counterType).AddUint(id).BuildWithPrefix(ckm.prefix)
}

// GenerateLockKey 生成锁键
func (ckm *CacheKeyManager) GenerateLockKey(resource string) string {
	return NewKeyBuilder().Add("lock").Add(resource).BuildWithPrefix(ckm.prefix)
}

// TTL管理函数

// GetTTL 获取键类型对应的TTL
func GetTTL(keyType string) time.Duration {
	if ttl, exists := CacheTTL[keyType]; exists {
		return ttl
	}
	return 30 * time.Minute // 默认TTL
}

// GetTTLByKey 根据键名获取TTL
func GetTTLByKey(key string) time.Duration {
	parts := strings.Split(key, ":")
	if len(parts) > 0 {
		// 去掉前缀，获取键类型
		keyType := parts[0]
		if len(parts) > 1 && parts[0] != "" {
			keyType = parts[1] // 如果有前缀，取第二部分
		}
		return GetTTL(keyType)
	}
	return GetTTL("default")
}

// SetCustomTTL 设置自定义TTL
func SetCustomTTL(keyType string, ttl time.Duration) {
	CacheTTL[keyType] = ttl
}

// 键验证函数

// ValidateKey 验证键名是否符合规范
func ValidateKey(key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}
	
	if len(key) > 250 {
		return fmt.Errorf("key too long: %d characters (max 250)", len(key))
	}
	
	// 检查是否包含非法字符
	invalidChars := []string{" ", "\t", "\n", "\r"}
	for _, char := range invalidChars {
		if strings.Contains(key, char) {
			return fmt.Errorf("key contains invalid character: %q", char)
		}
	}
	
	return nil
}

// ParseKey 解析键名，提取各部分
func ParseKey(key string) map[string]string {
	parts := strings.Split(key, ":")
	result := make(map[string]string)
	
	if len(parts) >= 1 {
		result["prefix"] = parts[0]
	}
	if len(parts) >= 2 {
		result["type"] = parts[1]
	}
	if len(parts) >= 3 {
		result["id"] = parts[2]
	}
	if len(parts) >= 4 {
		result["subtype"] = parts[3]
	}
	
	result["full"] = key
	result["parts_count"] = fmt.Sprintf("%d", len(parts))
	
	return result
}

// 批量键操作

// GenerateBatchKeys 批量生成键
func (ckm *CacheKeyManager) GenerateBatchKeys(keyType string, ids []uint) []string {
	keys := make([]string, len(ids))
	for i, id := range ids {
		switch keyType {
		case "product":
			keys[i] = ckm.GenerateProductKey(id)
		case "stock":
			keys[i] = ckm.GenerateProductStockKey(id)
		case "price":
			keys[i] = ckm.GenerateProductPriceKey(id)
		case "order":
			keys[i] = ckm.GenerateOrderKey(id)
		case "user":
			keys[i] = ckm.GenerateUserProfileKey(id)
		case "cart":
			keys[i] = ckm.GenerateUserCartKey(id)
		default:
			keys[i] = NewKeyBuilder().Add(keyType).AddUint(id).BuildWithPrefix(ckm.prefix)
		}
	}
	return keys
}

// 全局键管理器实例
var GlobalKeyManager *CacheKeyManager

// InitKeyManager 初始化全局键管理器
func InitKeyManager(prefix string) {
	GlobalKeyManager = NewCacheKeyManager(prefix)
}

// GetKeyManager 获取全局键管理器
func GetKeyManager() *CacheKeyManager {
	if GlobalKeyManager == nil {
		GlobalKeyManager = NewCacheKeyManager("mall")
	}
	return GlobalKeyManager
}
