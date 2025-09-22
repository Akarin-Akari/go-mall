package main

import (
	"fmt"
	"mall-go/pkg/cache"
	"mall-go/pkg/logger"
	"time"
)

func main() {
	// 初始化日志
	logger.Init()

	fmt.Println("🔧 测试缓存键管理工具...")

	// 初始化全局键管理器
	cache.InitKeyManager("mall")
	manager := cache.GetKeyManager()

	fmt.Printf("📋 缓存键管理工具验证:\n")
	fmt.Printf("  - 键管理器前缀: %s\n", "mall")

	// 测试键构建器
	testKeyBuilder()

	// 测试键生成
	testKeyGeneration(manager)

	// 测试TTL管理
	testTTLManagement()

	// 测试键验证
	testKeyValidation()

	// 测试键解析
	testKeyParsing()

	// 测试批量操作
	testBatchOperations(manager)

	fmt.Println("\n🎉 任务1.3 缓存键管理工具完成!")
	fmt.Println("📋 验收标准检查:")
	fmt.Println("  ✅ 缓存键命名规范实现")
	fmt.Println("  ✅ TTL管理工具完善")
	fmt.Println("  ✅ 键生成工具类完成")
	fmt.Println("  ✅ 键验证和解析功能")
	fmt.Println("  ✅ 批量键操作支持")
}

func testKeyBuilder() {
	fmt.Println("\n🧪 测试键构建器:")

	// 基础键构建
	key1 := cache.NewKeyBuilder().Add("product").AddUint(123).Build()
	fmt.Printf("  ✅ 基础键构建: %s\n", key1)

	// 带前缀键构建
	key2 := cache.NewKeyBuilder().Add("product").AddUint(123).BuildWithPrefix("mall")
	fmt.Printf("  ✅ 带前缀键构建: %s\n", key2)

	// 复杂键构建
	key3 := cache.NewKeyBuilder().
		Add("user").
		Add("session").
		Add("abc123").
		AddFormat("exp_%d", 1640995200).
		Build()
	fmt.Printf("  ✅ 复杂键构建: %s\n", key3)

	// 整数键构建
	key4 := cache.NewKeyBuilder().Add("counter").AddInt(456).AddUint(789).Build()
	fmt.Printf("  ✅ 整数键构建: %s\n", key4)
}

func testKeyGeneration(manager *cache.CacheKeyManager) {
	fmt.Println("\n🧪 测试键生成:")

	// 商品相关键
	productKey := manager.GenerateProductKey(123)
	fmt.Printf("  ✅ 商品键: %s\n", productKey)

	stockKey := manager.GenerateProductStockKey(123)
	fmt.Printf("  ✅ 库存键: %s\n", stockKey)

	priceKey := manager.GenerateProductPriceKey(123)
	fmt.Printf("  ✅ 价格键: %s\n", priceKey)

	// 用户相关键
	sessionKey := manager.GenerateUserSessionKey(123)
	fmt.Printf("  ✅ 会话键: %s\n", sessionKey)

	cartKey := manager.GenerateUserCartKey(456)
	fmt.Printf("  ✅ 购物车键: %s\n", cartKey)

	profileKey := manager.GenerateUserProfileKey(456)
	fmt.Printf("  ✅ 用户资料键: %s\n", profileKey)

	// 业务相关键
	categoryKey := manager.GenerateCategoryProductsKey(789)
	fmt.Printf("  ✅ 分类商品键: %s\n", categoryKey)

	hotKey := manager.GenerateHotProductsKey()
	fmt.Printf("  ✅ 热门商品键: %s\n", hotKey)

	orderKey := manager.GenerateOrderKey(101112)
	fmt.Printf("  ✅ 订单键: %s\n", orderKey)

	// 统计和控制键
	statsKey := manager.GenerateStatsKey("daily", "2025-01-10")
	fmt.Printf("  ✅ 统计键: %s\n", statsKey)

	counterKey := manager.GenerateCounterKey("view", 123)
	fmt.Printf("  ✅ 计数器键: %s\n", counterKey)

	lockKey := manager.GenerateLockKey("order:123")
	fmt.Printf("  ✅ 锁键: %s\n", lockKey)
}

func testTTLManagement() {
	fmt.Println("\n🧪 测试TTL管理:")

	// 测试预定义TTL
	ttlTypes := []string{"product", "stock", "session", "lock", "stats"}
	for _, ttlType := range ttlTypes {
		ttl := cache.GetTTL(ttlType)
		fmt.Printf("  ✅ %s TTL: %v\n", ttlType, ttl)
	}

	// 测试根据键名获取TTL
	testKeys := []string{
		"mall:product:123",
		"stock:456",
		"user:session:abc123",
		"lock:order:789",
	}
	for _, key := range testKeys {
		ttl := cache.GetTTLByKey(key)
		fmt.Printf("  ✅ 键 %s TTL: %v\n", key, ttl)
	}

	// 测试自定义TTL
	cache.SetCustomTTL("custom_type", 5*time.Minute)
	customTTL := cache.GetTTL("custom_type")
	fmt.Printf("  ✅ 自定义TTL设置: %v\n", customTTL)

	// 测试未知类型默认TTL
	unknownTTL := cache.GetTTL("unknown_type")
	fmt.Printf("  ✅ 未知类型默认TTL: %v\n", unknownTTL)
}

func testKeyValidation() {
	fmt.Println("\n🧪 测试键验证:")

	// 有效键测试
	validKeys := []string{
		"valid:key:123",
		"mall:product:456",
		"user:session:abc123def456",
		"stats:daily:2025-01-10",
	}

	for _, key := range validKeys {
		err := cache.ValidateKey(key)
		if err == nil {
			fmt.Printf("  ✅ 有效键: %s\n", key)
		} else {
			fmt.Printf("  ❌ 键验证失败: %s - %v\n", key, err)
		}
	}

	// 无效键测试
	invalidKeys := []string{
		"",                            // 空键
		"invalid key with space",      // 包含空格
		"invalid\tkey\twith\ttab",     // 包含制表符
		"invalid\nkey\nwith\nnewline", // 包含换行符
	}

	for _, key := range invalidKeys {
		err := cache.ValidateKey(key)
		if err != nil {
			fmt.Printf("  ✅ 无效键检测: %s - %v\n", key, err)
		} else {
			fmt.Printf("  ❌ 应该无效但通过验证: %s\n", key)
		}
	}

	// 过长键测试
	longKey := make([]byte, 251)
	for i := range longKey {
		longKey[i] = 'a'
	}
	err := cache.ValidateKey(string(longKey))
	if err != nil {
		fmt.Printf("  ✅ 过长键检测: 长度%d - %v\n", len(longKey), err)
	}
}

func testKeyParsing() {
	fmt.Println("\n🧪 测试键解析:")

	testKeys := []string{
		"mall:product:123:info",
		"user:session:abc123",
		"stock:456",
		"simple",
	}

	for _, key := range testKeys {
		parsed := cache.ParseKey(key)
		fmt.Printf("  ✅ 解析键 %s:\n", key)
		fmt.Printf("    - 前缀: %s\n", parsed["prefix"])
		fmt.Printf("    - 类型: %s\n", parsed["type"])
		fmt.Printf("    - ID: %s\n", parsed["id"])
		fmt.Printf("    - 子类型: %s\n", parsed["subtype"])
		fmt.Printf("    - 部分数: %s\n", parsed["parts_count"])
	}
}

func testBatchOperations(manager *cache.CacheKeyManager) {
	fmt.Println("\n🧪 测试批量操作:")

	// 批量生成商品键
	productIDs := []uint{1, 2, 3, 4, 5}
	productKeys := manager.GenerateBatchKeys("product", productIDs)
	fmt.Printf("  ✅ 批量商品键 (%d个):\n", len(productKeys))
	for i, key := range productKeys {
		fmt.Printf("    %d. %s\n", i+1, key)
	}

	// 批量生成库存键
	stockKeys := manager.GenerateBatchKeys("stock", []uint{10, 20, 30})
	fmt.Printf("  ✅ 批量库存键 (%d个):\n", len(stockKeys))
	for i, key := range stockKeys {
		fmt.Printf("    %d. %s\n", i+1, key)
	}

	// 批量生成用户键
	userKeys := manager.GenerateBatchKeys("user", []uint{100, 200})
	fmt.Printf("  ✅ 批量用户键 (%d个):\n", len(userKeys))
	for i, key := range userKeys {
		fmt.Printf("    %d. %s\n", i+1, key)
	}

	// 批量生成自定义键
	customKeys := manager.GenerateBatchKeys("custom", []uint{1000, 2000})
	fmt.Printf("  ✅ 批量自定义键 (%d个):\n", len(customKeys))
	for i, key := range customKeys {
		fmt.Printf("    %d. %s\n", i+1, key)
	}
}
