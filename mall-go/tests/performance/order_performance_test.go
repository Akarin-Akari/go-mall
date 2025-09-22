package performance

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestOrderAPIPerformance 订单API性能测试
func TestOrderAPIPerformance(t *testing.T) {
	// 设置性能测试环境
	suite := SetupPerformanceTest(t)
	defer suite.CleanupPerformanceTest()

	// 创建测试数据
	suite.CreateTestData(t)

	t.Run("订单创建性能测试", func(t *testing.T) {
		concurrency := 50
		totalRequests := 500

		requestFunc := func() *RequestResult {
			// 随机用户创建订单
			userID := rand.Intn(100) + 1

			// 创建订单项
			orderItems := make([]map[string]interface{}, 0)
			itemCount := rand.Intn(3) + 1 // 1-3个商品

			for i := 0; i < itemCount; i++ {
				productID := rand.Intn(1000) + 1
				quantity := rand.Intn(5) + 1

				orderItems = append(orderItems, map[string]interface{}{
					"product_id": productID,
					"quantity":   quantity,
				})
			}

			createOrderReq := map[string]interface{}{
				"user_id":     userID,
				"order_items": orderItems,
				"address":     "测试地址",
				"phone":       "13800138000",
				"remark":      "性能测试订单",
			}

			return suite.MakeHTTPRequest("POST", "/api/order/create", createOrderReq)
		}

		result := suite.RunConcurrentTest(t, "订单创建", concurrency, totalRequests, requestFunc)

		// 验证性能指标
		assert.Less(t, result.AverageTime, 300*time.Millisecond, "平均响应时间应小于300ms")
		assert.Less(t, result.P95ResponseTime, 800*time.Millisecond, "P95响应时间应小于800ms")
		assert.Greater(t, result.RequestsPerSec, 100.0, "QPS应大于100")
		assert.Less(t, result.ErrorRate, 10.0, "错误率应小于10%")

		t.Logf("✅ 订单创建性能测试通过 - 平均响应时间: %v, QPS: %.2f",
			result.AverageTime, result.RequestsPerSec)
	})

	t.Run("订单查询性能测试", func(t *testing.T) {
		// 先创建一些测试订单
		for i := 1; i <= 100; i++ {
			userID := i
			orderItems := []map[string]interface{}{
				{
					"product_id": rand.Intn(1000) + 1,
					"quantity":   rand.Intn(3) + 1,
				},
			}

			createOrderReq := map[string]interface{}{
				"user_id":     userID,
				"order_items": orderItems,
				"address":     fmt.Sprintf("测试地址%d", i),
				"phone":       "13800138000",
			}

			suite.MakeHTTPRequest("POST", "/api/order/create", createOrderReq)
		}

		concurrency := 150
		totalRequests := 1500

		requestFunc := func() *RequestResult {
			// 随机查询订单
			userID := rand.Intn(100) + 1
			page := rand.Intn(5) + 1
			pageSize := rand.Intn(10) + 10

			path := fmt.Sprintf("/api/order/list?user_id=%d&page=%d&page_size=%d",
				userID, page, pageSize)
			return suite.MakeHTTPRequest("GET", path, nil)
		}

		result := suite.RunConcurrentTest(t, "订单查询", concurrency, totalRequests, requestFunc)

		// 验证性能指标
		assert.Less(t, result.AverageTime, 150*time.Millisecond, "平均响应时间应小于150ms")
		assert.Less(t, result.P95ResponseTime, 400*time.Millisecond, "P95响应时间应小于400ms")
		assert.Greater(t, result.RequestsPerSec, 400.0, "QPS应大于400")
		assert.Less(t, result.ErrorRate, 3.0, "错误率应小于3%")

		t.Logf("✅ 订单查询性能测试通过 - 平均响应时间: %v, QPS: %.2f",
			result.AverageTime, result.RequestsPerSec)
	})

	t.Run("订单详情查询性能测试", func(t *testing.T) {
		concurrency := 200
		totalRequests := 2000

		requestFunc := func() *RequestResult {
			// 查询随机订单详情
			orderID := rand.Intn(100) + 1
			path := fmt.Sprintf("/api/order/%d", orderID)
			return suite.MakeHTTPRequest("GET", path, nil)
		}

		result := suite.RunConcurrentTest(t, "订单详情查询", concurrency, totalRequests, requestFunc)

		// 验证性能指标
		assert.Less(t, result.AverageTime, 100*time.Millisecond, "平均响应时间应小于100ms")
		assert.Less(t, result.P95ResponseTime, 250*time.Millisecond, "P95响应时间应小于250ms")
		assert.Greater(t, result.RequestsPerSec, 600.0, "QPS应大于600")
		assert.Less(t, result.ErrorRate, 2.0, "错误率应小于2%")

		t.Logf("✅ 订单详情查询性能测试通过 - 平均响应时间: %v, QPS: %.2f",
			result.AverageTime, result.RequestsPerSec)
	})

	t.Run("订单状态更新性能测试", func(t *testing.T) {
		concurrency := 80
		totalRequests := 800

		statuses := []string{"paid", "shipped", "delivered", "cancelled"}

		requestFunc := func() *RequestResult {
			// 随机更新订单状态
			orderID := rand.Intn(100) + 1
			status := statuses[rand.Intn(len(statuses))]

			updateReq := map[string]interface{}{
				"status": status,
				"remark": fmt.Sprintf("性能测试状态更新: %s", status),
			}

			path := fmt.Sprintf("/api/order/%d/status", orderID)
			return suite.MakeHTTPRequest("PUT", path, updateReq)
		}

		result := suite.RunConcurrentTest(t, "订单状态更新", concurrency, totalRequests, requestFunc)

		// 验证性能指标
		assert.Less(t, result.AverageTime, 120*time.Millisecond, "平均响应时间应小于120ms")
		assert.Less(t, result.P95ResponseTime, 300*time.Millisecond, "P95响应时间应小于300ms")
		assert.Greater(t, result.RequestsPerSec, 350.0, "QPS应大于350")
		assert.Less(t, result.ErrorRate, 8.0, "错误率应小于8%")

		t.Logf("✅ 订单状态更新性能测试通过 - 平均响应时间: %v, QPS: %.2f",
			result.AverageTime, result.RequestsPerSec)
	})

	t.Run("订单高并发创建压力测试", func(t *testing.T) {
		concurrency := 200
		totalRequests := 2000

		requestFunc := func() *RequestResult {
			// 高并发创建订单
			userID := rand.Intn(100) + 1

			// 使用热门商品测试库存并发
			hotProducts := []int{1, 2, 3, 4, 5}
			productID := hotProducts[rand.Intn(len(hotProducts))]
			quantity := rand.Intn(2) + 1

			orderItems := []map[string]interface{}{
				{
					"product_id": productID,
					"quantity":   quantity,
				},
			}

			createOrderReq := map[string]interface{}{
				"user_id":     userID,
				"order_items": orderItems,
				"address":     "压力测试地址",
				"phone":       "13800138000",
				"remark":      "高并发压力测试订单",
			}

			return suite.MakeHTTPRequest("POST", "/api/order/create", createOrderReq)
		}

		result := suite.RunConcurrentTest(t, "订单高并发创建压力测试", concurrency, totalRequests, requestFunc)

		// 压力测试的验收标准
		assert.Less(t, result.AverageTime, 500*time.Millisecond, "压力测试平均响应时间应小于500ms")
		assert.Less(t, result.P95ResponseTime, 1200*time.Millisecond, "压力测试P95响应时间应小于1.2s")
		assert.Greater(t, result.RequestsPerSec, 50.0, "压力测试QPS应大于50")
		// 高并发创建订单可能因为库存不足导致较高错误率
		assert.Less(t, result.ErrorRate, 30.0, "压力测试错误率应小于30%")

		t.Logf("✅ 订单高并发创建压力测试通过 - 平均响应时间: %v, QPS: %.2f, 错误率: %.2f%%",
			result.AverageTime, result.RequestsPerSec, result.ErrorRate)
	})

	t.Run("订单混合操作性能测试", func(t *testing.T) {
		concurrency := 100
		totalRequests := 1000

		operations := []string{"create", "list", "detail", "update_status"}

		requestFunc := func() *RequestResult {
			// 随机选择操作类型
			operation := operations[rand.Intn(len(operations))]
			userID := rand.Intn(100) + 1

			switch operation {
			case "create":
				orderItems := []map[string]interface{}{
					{
						"product_id": rand.Intn(1000) + 1,
						"quantity":   rand.Intn(3) + 1,
					},
				}
				createOrderReq := map[string]interface{}{
					"user_id":     userID,
					"order_items": orderItems,
					"address":     "混合测试地址",
					"phone":       "13800138000",
				}
				return suite.MakeHTTPRequest("POST", "/api/order/create", createOrderReq)

			case "list":
				path := fmt.Sprintf("/api/order/list?user_id=%d", userID)
				return suite.MakeHTTPRequest("GET", path, nil)

			case "detail":
				orderID := rand.Intn(100) + 1
				path := fmt.Sprintf("/api/order/%d", orderID)
				return suite.MakeHTTPRequest("GET", path, nil)

			case "update_status":
				orderID := rand.Intn(100) + 1
				statuses := []string{"paid", "shipped", "delivered"}
				status := statuses[rand.Intn(len(statuses))]
				updateReq := map[string]interface{}{
					"status": status,
				}
				path := fmt.Sprintf("/api/order/%d/status", orderID)
				return suite.MakeHTTPRequest("PUT", path, updateReq)

			default:
				path := fmt.Sprintf("/api/order/list?user_id=%d", userID)
				return suite.MakeHTTPRequest("GET", path, nil)
			}
		}

		result := suite.RunConcurrentTest(t, "订单混合操作性能测试", concurrency, totalRequests, requestFunc)

		// 验证性能指标
		assert.Less(t, result.AverageTime, 250*time.Millisecond, "平均响应时间应小于250ms")
		assert.Less(t, result.P95ResponseTime, 600*time.Millisecond, "P95响应时间应小于600ms")
		assert.Greater(t, result.RequestsPerSec, 200.0, "QPS应大于200")
		assert.Less(t, result.ErrorRate, 15.0, "错误率应小于15%")

		t.Logf("✅ 订单混合操作性能测试通过 - 平均响应时间: %v, QPS: %.2f",
			result.AverageTime, result.RequestsPerSec)
	})
}

// BenchmarkOrderCreate 订单创建基准测试
func BenchmarkOrderCreate(b *testing.B) {
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

			orderItems := []map[string]interface{}{
				{
					"product_id": productID,
					"quantity":   quantity,
				},
			}

			createOrderReq := map[string]interface{}{
				"user_id":     userID,
				"order_items": orderItems,
				"address":     fmt.Sprintf("基准测试地址%d", i),
				"phone":       "13800138000",
			}

			result := suite.MakeHTTPRequest("POST", "/api/order/create", createOrderReq)
			if !result.Success {
				b.Errorf("订单创建失败: %v", result.Error)
			}
		}
	})
}

// BenchmarkOrderList 订单列表基准测试
func BenchmarkOrderList(b *testing.B) {
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
			page := i%10 + 1
			pageSize := 20

			path := fmt.Sprintf("/api/order/list?user_id=%d&page=%d&page_size=%d",
				userID, page, pageSize)

			result := suite.MakeHTTPRequest("GET", path, nil)
			if !result.Success {
				b.Errorf("订单列表查询失败: %v", result.Error)
			}
		}
	})
}
