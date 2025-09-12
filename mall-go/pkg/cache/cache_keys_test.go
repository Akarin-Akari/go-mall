package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestKeyBuilder(t *testing.T) {
	// 测试基础键构建
	key := NewKeyBuilder().Add("product").AddUint(123).Build()
	assert.Equal(t, "product:123", key)

	// 测试带前缀的键构建
	keyWithPrefix := NewKeyBuilder().Add("product").AddUint(123).BuildWithPrefix("mall")
	assert.Equal(t, "mall:product:123", keyWithPrefix)

	// 测试复杂键构建
	complexKey := NewKeyBuilder().
		Add("user").
		Add("session").
		Add("abc123").
		AddFormat("exp_%d", 1640995200).
		Build()
	assert.Equal(t, "user:session:abc123:exp_1640995200", complexKey)

	// 测试空部分过滤
	keyWithEmpty := NewKeyBuilder().
		Add("product").
		Add("").
		AddUint(456).
		Build()
	assert.Equal(t, "product:456", keyWithEmpty)
}

func TestCacheKeyManager(t *testing.T) {
	manager := NewCacheKeyManager("test")

	// 测试商品相关键生成
	productKey := manager.GenerateProductKey(123)
	assert.Equal(t, "test:product:123", productKey)

	stockKey := manager.GenerateProductStockKey(123)
	assert.Equal(t, "test:stock:123", stockKey)

	priceKey := manager.GenerateProductPriceKey(123)
	assert.Equal(t, "test:price:123", priceKey)

	// 测试用户相关键生成
	sessionKey := manager.GenerateUserSessionKey(123)
	assert.Equal(t, "test:user:session:123", sessionKey)

	cartKey := manager.GenerateUserCartKey(456)
	assert.Equal(t, "test:cart:456", cartKey)

	profileKey := manager.GenerateUserProfileKey(456)
	assert.Equal(t, "test:user:profile:456", profileKey)

	// 测试业务相关键生成
	categoryKey := manager.GenerateCategoryProductsKey(789)
	assert.Equal(t, "test:category:789:products", categoryKey)

	hotKey := manager.GenerateHotProductsKey()
	assert.Equal(t, "test:hot:products", hotKey)

	orderKey := manager.GenerateOrderKey(101112)
	assert.Equal(t, "test:order:101112", orderKey)

	// 测试统计和锁键生成
	statsKey := manager.GenerateStatsKey("daily", "2025-01-10")
	assert.Equal(t, "test:stats:daily:2025-01-10", statsKey)

	counterKey := manager.GenerateCounterKey("view", 123)
	assert.Equal(t, "test:counter:view:123", counterKey)

	lockKey := manager.GenerateLockKey("order:123")
	assert.Equal(t, "test:lock:order:123", lockKey)
}

func TestTTLManagement(t *testing.T) {
	// 测试预定义TTL
	productTTL := GetTTL("product")
	assert.Equal(t, 30*time.Minute, productTTL)

	stockTTL := GetTTL("stock")
	assert.Equal(t, 10*time.Minute, stockTTL)

	sessionTTL := GetTTL("session")
	assert.Equal(t, 2*time.Hour, sessionTTL)

	lockTTL := GetTTL("lock")
	assert.Equal(t, 30*time.Second, lockTTL)

	// 测试未知类型的默认TTL
	unknownTTL := GetTTL("unknown_type")
	assert.Equal(t, 30*time.Minute, unknownTTL)

	// 测试根据键名获取TTL
	keyTTL := GetTTLByKey("mall:product:123")
	assert.Equal(t, 30*time.Minute, keyTTL)

	keyTTL2 := GetTTLByKey("stock:456")
	assert.Equal(t, 10*time.Minute, keyTTL2)

	// 测试自定义TTL
	SetCustomTTL("custom", 5*time.Minute)
	customTTL := GetTTL("custom")
	assert.Equal(t, 5*time.Minute, customTTL)
}

func TestKeyValidation(t *testing.T) {
	// 测试有效键
	err := ValidateKey("valid:key:123")
	assert.NoError(t, err)

	// 测试空键
	err = ValidateKey("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")

	// 测试过长键
	longKey := make([]byte, 251)
	for i := range longKey {
		longKey[i] = 'a'
	}
	err = ValidateKey(string(longKey))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too long")

	// 测试包含非法字符的键
	err = ValidateKey("invalid key with space")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid character")

	err = ValidateKey("invalid\tkey\twith\ttab")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid character")

	err = ValidateKey("invalid\nkey\nwith\nnewline")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid character")
}

func TestKeyParsing(t *testing.T) {
	// 测试完整键解析
	key := "mall:product:123:info"
	parsed := ParseKey(key)

	assert.Equal(t, "mall", parsed["prefix"])
	assert.Equal(t, "product", parsed["type"])
	assert.Equal(t, "123", parsed["id"])
	assert.Equal(t, "info", parsed["subtype"])
	assert.Equal(t, key, parsed["full"])
	assert.Equal(t, "4", parsed["parts_count"])

	// 测试简单键解析
	simpleKey := "product:456"
	simpleParsed := ParseKey(simpleKey)

	assert.Equal(t, "product", simpleParsed["prefix"])
	assert.Equal(t, "456", simpleParsed["type"])
	assert.Equal(t, simpleKey, simpleParsed["full"])
	assert.Equal(t, "2", simpleParsed["parts_count"])

	// 测试单部分键解析
	singleKey := "simple"
	singleParsed := ParseKey(singleKey)

	assert.Equal(t, "simple", singleParsed["prefix"])
	assert.Equal(t, singleKey, singleParsed["full"])
	assert.Equal(t, "1", singleParsed["parts_count"])
}

func TestBatchKeyGeneration(t *testing.T) {
	manager := NewCacheKeyManager("test")

	// 测试批量生成商品键
	productIDs := []uint{1, 2, 3, 4, 5}
	productKeys := manager.GenerateBatchKeys("product", productIDs)

	expectedKeys := []string{
		"test:product:1",
		"test:product:2",
		"test:product:3",
		"test:product:4",
		"test:product:5",
	}
	assert.Equal(t, expectedKeys, productKeys)

	// 测试批量生成库存键
	stockKeys := manager.GenerateBatchKeys("stock", productIDs)
	expectedStockKeys := []string{
		"test:stock:1",
		"test:stock:2",
		"test:stock:3",
		"test:stock:4",
		"test:stock:5",
	}
	assert.Equal(t, expectedStockKeys, stockKeys)

	// 测试批量生成用户键
	userIDs := []uint{10, 20, 30}
	userKeys := manager.GenerateBatchKeys("user", userIDs)
	expectedUserKeys := []string{
		"test:user:profile:10",
		"test:user:profile:20",
		"test:user:profile:30",
	}
	assert.Equal(t, expectedUserKeys, userKeys)

	// 测试批量生成自定义键
	customKeys := manager.GenerateBatchKeys("custom", []uint{100, 200})
	expectedCustomKeys := []string{
		"test:custom:100",
		"test:custom:200",
	}
	assert.Equal(t, expectedCustomKeys, customKeys)
}

func TestGlobalKeyManager(t *testing.T) {
	// 测试初始化全局键管理器
	InitKeyManager("global_test")

	manager := GetKeyManager()
	assert.NotNil(t, manager)
	assert.Equal(t, "global_test", manager.prefix)

	// 测试使用全局键管理器
	key := manager.GenerateProductKey(999)
	assert.Equal(t, "global_test:product:999", key)

	// 测试默认全局键管理器
	GlobalKeyManager = nil
	defaultManager := GetKeyManager()
	assert.NotNil(t, defaultManager)
	assert.Equal(t, "mall", defaultManager.prefix)
}

func TestKeyConstants(t *testing.T) {
	// 验证键常量格式正确
	assert.Equal(t, "product:{id}", ProductInfoKey)
	assert.Equal(t, "stock:{id}", ProductStockKey)
	assert.Equal(t, "price:{id}", ProductPriceKey)
	assert.Equal(t, "user:session:{token}", UserSessionKey)
	assert.Equal(t, "cart:{user_id}", UserCartKey)
	assert.Equal(t, "category:{id}:products", CategoryProductsKey)
	assert.Equal(t, "hot:products", HotProductsKey)
	assert.Equal(t, "order:{id}", OrderKey)
	assert.Equal(t, "stats:daily:{date}", DailyStatsKey)
	assert.Equal(t, "counter:{type}:{id}", CounterKey)
	assert.Equal(t, "lock:{key}", CacheLockKey)
}

func BenchmarkKeyGeneration(b *testing.B) {
	manager := NewCacheKeyManager("bench")

	b.Run("ProductKey", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			manager.GenerateProductKey(uint(i))
		}
	})

	b.Run("UserSessionKey", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			manager.GenerateUserSessionKey(uint(i))
		}
	})

	b.Run("KeyBuilder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			NewKeyBuilder().Add("test").AddInt(i).Add("data").Build()
		}
	})
}
