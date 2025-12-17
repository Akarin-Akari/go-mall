package performance

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestCartAPIPerformance 购物车API性能测试
func TestCartAPIPerformance(t *testing.T) {
	// 设置性能测试环境
	suite := SetupPerformanceTest(t)
	defer suite.CleanupPerformanceTest()

	// 创建测试数据
	suite.CreateTestData(t)

	t.Run("添加商品到购物车性能测试", func(t *testing.T) {
		concurrency := 100
		totalRequests := 1000

		requestFunc := func() *RequestResult {
			// 随机用户和商品
			userID := rand.Intn(100) + 1
			productID := rand.Intn(1000) + 1
			quantity := rand.Intn(5) + 1

			addToCartReq := map[string]interface{}{
				"product_id": productID,
				"quantity":   quantity,
			}

			path := fmt.Sprintf("/api/cart/add?user_id=%d", userID)
			return suite.MakeHTTPRequest("POST", path, addToCartReq)
		}

		result := suite.RunConcurrentTest(t, "添加商品到购物车", concurrency, totalRequests, requestFunc)

		// 验证性能指标
		assert.Less(t, result.AverageTime, 150*time.Millisecond, "平均响应时间应小于150ms")
		assert.Less(t, result.P95ResponseTime, 400*time.Millisecond, "P95响应时间应小于400ms")
		assert.Greater(t, result.RequestsPerSec, 300.0, "QPS应大于300")
		assert.Less(t, result.ErrorRate, 5.0, "错误率应小于5%")

		t.Logf("✅ 添加商品到购物车性能测试通过 - 平均响应时间: %v, QPS: %.2f",
			result.AverageTime, result.RequestsPerSec)
	})

	t.Run("获取购物车性能测试", func(t *testing.T) {
		// 先为用户添加一些购物车商品
		for userID := 1; userID <= 100; userID++ {
			for i := 0; i < 5; i++ {
				productID := rand.Intn(1000) + 1
				quantity := rand.Intn(3) + 1
				addToCartReq := map[string]interface{}{
					"product_id": productID,
					"quantity":   quantity,
				}
				path := fmt.Sprintf("/api/cart/add?user_id=%d", userID)
				suite.MakeHTTPRequest("POST", path, addToCartReq)
			}
		}

		concurrency := 200
		totalRequests := 2000

		requestFunc := func() *RequestResult {
			// 随机用户获取购物车
			userID := rand.Intn(100) + 1
			path := fmt.Sprintf("/api/cart?user_id=%d", userID)
			return suite.MakeHTTPRequest("GET", path, nil)
		}

		result := suite.RunConcurrentTest(t, "获取购物车", concurrency, totalRequests, requestFunc)

		// 验证性能指标
		assert.Less(t, result.AverageTime, 100*time.Millisecond, "平均响应时间应小于100ms")
		assert.Less(t, result.P95ResponseTime, 250*time.Millisecond, "P95响应时间应小于250ms")
		assert.Greater(t, result.RequestsPerSec, 500.0, "QPS应大于500")
		assert.Less(t, result.ErrorRate, 2.0, "错误率应小于2%")

		t.Logf("✅ 获取购物车性能测试通过 - 平均响应时间: %v, QPS: %.2f",
			result.AverageTime, result.RequestsPerSec)
	})

	t.Run("更新购物车商品数量性能测试", func(t *testing.T) {
		concurrency := 80
		totalRequests := 800

		requestFunc := func() *RequestResult {
			// 随机用户和商品
			userID := rand.Intn(100) + 1
			productID := rand.Intn(1000) + 1
			quantity := rand.Intn(10) + 1

			updateReq := map[string]interface{}{
				"product_id": productID,
				"quantity":   quantity,
			}

			path := fmt.Sprintf("/api/cart/update?user_id=%d", userID)
			return suite.MakeHTTPRequest("PUT", path, updateReq)
		}

		result := suite.RunConcurrentTest(t, "更新购物车商品数量", concurrency, totalRequests, requestFunc)

		// 验证性能指标
		assert.Less(t, result.AverageTime, 120*time.Millisecond, "平均响应时间应小于120ms")
		assert.Less(t, result.P95ResponseTime, 300*time.Millisecond, "P95响应时间应小于300ms")
		assert.Greater(t, result.RequestsPerSec, 350.0, "QPS应大于350")
		assert.Less(t, result.ErrorRate, 8.0, "错误率应小于8%")

		t.Logf("✅ 更新购物车商品数量性能测试通过 - 平均响应时间: %v, QPS: %.2f",
			result.AverageTime, result.RequestsPerSec)
	})

	t.Run("删除购物车商品性能测试", func(t *testing.T) {
		concurrency := 60
		totalRequests := 600

		requestFunc := func() *RequestResult {
			// 随机用户和商品
			userID := rand.Intn(100) + 1
			productID := rand.Intn(1000) + 1

			path := fmt.Sprintf("/api/cart/remove?user_id=%d&product_id=%d", userID, productID)
			return suite.MakeHTTPRequest("DELETE", path, nil)
		}

		result := suite.RunConcurrentTest(t, "删除购物车商品", concurrency, totalRequests, requestFunc)

		// 验证性能指标
		assert.Less(t, result.AverageTime, 100*time.Millisecond, "平均响应时间应小于100ms")
		assert.Less(t, result.P95ResponseTime, 250*time.Millisecond, "P95响应时间应小于250ms")
		assert.Greater(t, result.RequestsPerSec, 400.0, "QPS应大于400")
		assert.Less(t, result.ErrorRate, 10.0, "错误率应小于10%")

		t.Logf("✅ 删除购物车商品性能测试通过 - 平均响应时间: %v, QPS: %.2f",
			result.AverageTime, result.RequestsPerSec)
	})

	t.Run("购物车高并发操作压力测试", func(t *testing.T) {
		concurrency := 300
		totalRequests := 3000

		operations := []string{"add", "get", "update", "remove"}

		requestFunc := func() *RequestResult {
			// 随机选择操作类型
			operation := operations[rand.Intn(len(operations))]
			userID := rand.Intn(100) + 1
			productID := rand.Intn(1000) + 1

			switch operation {
			case "add":
				quantity := rand.Intn(5) + 1
				addToCartReq := map[string]interface{}{
					"product_id": productID,
					"quantity":   quantity,
				}
				path := fmt.Sprintf("/api/cart/add?user_id=%d", userID)
				return suite.MakeHTTPRequest("POST", path, addToCartReq)

			case "get":
				path := fmt.Sprintf("/api/cart?user_id=%d", userID)
				return suite.MakeHTTPRequest("GET", path, nil)

			case "update":
				quantity := rand.Intn(10) + 1
				updateReq := map[string]interface{}{
					"product_id": productID,
					"quantity":   quantity,
				}
				path := fmt.Sprintf("/api/cart/update?user_id=%d", userID)
				return suite.MakeHTTPRequest("PUT", path, updateReq)

			case "remove":
				path := fmt.Sprintf("/api/cart/remove?user_id=%d&product_id=%d", userID, productID)
				return suite.MakeHTTPRequest("DELETE", path, nil)

			default:
				path := fmt.Sprintf("/api/cart?user_id=%d", userID)
				return suite.MakeHTTPRequest("GET", path, nil)
			}
		}

		result := suite.RunConcurrentTest(t, "购物车高并发操作压力测试", concurrency, totalRequests, requestFunc)

		// 压力测试的验收标准
		assert.Less(t, result.AverageTime, 400*time.Millisecond, "压力测试平均响应时间应小于400ms")
		assert.Less(t, result.P95ResponseTime, 1*time.Second, "压力测试P95响应时间应小于1s")
		assert.Greater(t, result.RequestsPerSec, 150.0, "压力测试QPS应大于150")
		assert.Less(t, result.ErrorRate, 20.0, "压力测试错误率应小于20%")

		t.Logf("✅ 购物车高并发操作压力测试通过 - 平均响应时间: %v, QPS: %.2f",
			result.AverageTime, result.RequestsPerSec)
	})

	t.Run("购物车并发库存检查测试", func(t *testing.T) {
		// 创建一个库存有限的商品
		limitedProductID := 1001
		concurrency := 50
		totalRequests := 500

		requestFunc := func() *RequestResult {
			// 多个用户同时添加同一个商品到购物车
			userID := rand.Intn(100) + 1
			quantity := rand.Intn(3) + 1

			addToCartReq := map[string]interface{}{
				"product_id": limitedProductID,
				"quantity":   quantity,
			}

			path := fmt.Sprintf("/api/cart/add?user_id=%d", userID)
			return suite.MakeHTTPRequest("POST", path, addToCartReq)
		}

		result := suite.RunConcurrentTest(t, "购物车并发库存检查测试", concurrency, totalRequests, requestFunc)

		// 并发库存检查的验收标准
		assert.Less(t, result.AverageTime, 200*time.Millisecond, "平均响应时间应小于200ms")
		assert.Less(t, result.P95ResponseTime, 500*time.Millisecond, "P95响应时间应小于500ms")
		assert.Greater(t, result.RequestsPerSec, 100.0, "QPS应大于100")
		// 库存检查可能导致较高的错误率，这是正常的
		assert.Less(t, result.ErrorRate, 50.0, "错误率应小于50%")

		t.Logf("✅ 购物车并发库存检查测试通过 - 平均响应时间: %v, QPS: %.2f, 错误率: %.2f%%",
			result.AverageTime, result.RequestsPerSec, result.ErrorRate)
	})
}

// BenchmarkCartAdd 添加购物车基准测试
func BenchmarkCartAdd(b *testing.B) {
	suite := SetupPerformanceTest(&testing.T{})
	defer suite.CleanupPerformanceTest()

	// 创建测试数据
	suite.CreateTestData(&testing.T{})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			i++
			userID := i%100 + 1
			productID := i%1000 + 1
			quantity := i%5 + 1

			addToCartReq := map[string]interface{}{
				"product_id": productID,
				"quantity":   quantity,
			}

			path := fmt.Sprintf("/api/cart/add?user_id=%d", userID)
			result := suite.MakeHTTPRequest("POST", path, addToCartReq)
			if !result.Success {
				b.Errorf("添加购物车失败: %v", result.Error)
			}
		}
	})
}

// BenchmarkCartGet 获取购物车基准测试
func BenchmarkCartGet(b *testing.B) {
	suite := SetupPerformanceTest(&testing.T{})
	defer suite.CleanupPerformanceTest()

	// 创建测试数据
	suite.CreateTestData(&testing.T{})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			i++
			userID := i%100 + 1
			path := fmt.Sprintf("/api/cart?user_id=%d", userID)

			result := suite.MakeHTTPRequest("GET", path, nil)
			if !result.Success {
				b.Errorf("获取购物车失败: %v", result.Error)
			}
		}
	})
}
