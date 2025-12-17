package performance

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestProductAPIPerformance 商品API性能测试
func TestProductAPIPerformance(t *testing.T) {
	// 设置性能测试环境
	suite := SetupPerformanceTest(t)
	defer suite.CleanupPerformanceTest()

	// 创建测试数据
	suite.CreateTestData(t)

	t.Run("商品列表查询性能测试", func(t *testing.T) {
		concurrency := 100
		totalRequests := 1000

		requestFunc := func() *RequestResult {
			// 随机分页参数
			page := rand.Intn(10) + 1
			pageSize := rand.Intn(20) + 10
			path := fmt.Sprintf("/api/product/list?page=%d&page_size=%d", page, pageSize)

			return suite.MakeHTTPRequest("GET", path, nil)
		}

		result := suite.RunConcurrentTest(t, "商品列表查询", concurrency, totalRequests, requestFunc)

		// 验证性能指标
		assert.Less(t, result.AverageTime, 150*time.Millisecond, "平均响应时间应小于150ms")
		assert.Less(t, result.P95ResponseTime, 400*time.Millisecond, "P95响应时间应小于400ms")
		assert.Greater(t, result.RequestsPerSec, 300.0, "QPS应大于300")
		assert.Less(t, result.ErrorRate, 3.0, "错误率应小于3%")

		t.Logf("✅ 商品列表查询性能测试通过 - 平均响应时间: %v, QPS: %.2f",
			result.AverageTime, result.RequestsPerSec)
	})

	t.Run("商品详情查询性能测试", func(t *testing.T) {
		concurrency := 200
		totalRequests := 2000

		requestFunc := func() *RequestResult {
			// 查询随机商品详情
			productID := rand.Intn(1000) + 1
			path := fmt.Sprintf("/api/product/%d", productID)

			return suite.MakeHTTPRequest("GET", path, nil)
		}

		result := suite.RunConcurrentTest(t, "商品详情查询", concurrency, totalRequests, requestFunc)

		// 验证性能指标
		assert.Less(t, result.AverageTime, 100*time.Millisecond, "平均响应时间应小于100ms")
		assert.Less(t, result.P95ResponseTime, 250*time.Millisecond, "P95响应时间应小于250ms")
		assert.Greater(t, result.RequestsPerSec, 500.0, "QPS应大于500")
		assert.Less(t, result.ErrorRate, 2.0, "错误率应小于2%")

		t.Logf("✅ 商品详情查询性能测试通过 - 平均响应时间: %v, QPS: %.2f",
			result.AverageTime, result.RequestsPerSec)
	})

	t.Run("商品搜索性能测试", func(t *testing.T) {
		concurrency := 80
		totalRequests := 800

		searchKeywords := []string{
			"性能测试", "商品", "测试", "电商", "购物",
			"手机", "电脑", "服装", "食品", "家居",
		}

		requestFunc := func() *RequestResult {
			// 随机搜索关键词
			keyword := searchKeywords[rand.Intn(len(searchKeywords))]
			page := rand.Intn(5) + 1
			pageSize := rand.Intn(10) + 10
			path := fmt.Sprintf("/api/product/search?keyword=%s&page=%d&page_size=%d",
				keyword, page, pageSize)

			return suite.MakeHTTPRequest("GET", path, nil)
		}

		result := suite.RunConcurrentTest(t, "商品搜索", concurrency, totalRequests, requestFunc)

		// 验证性能指标
		assert.Less(t, result.AverageTime, 200*time.Millisecond, "平均响应时间应小于200ms")
		assert.Less(t, result.P95ResponseTime, 500*time.Millisecond, "P95响应时间应小于500ms")
		assert.Greater(t, result.RequestsPerSec, 200.0, "QPS应大于200")
		assert.Less(t, result.ErrorRate, 5.0, "错误率应小于5%")

		t.Logf("✅ 商品搜索性能测试通过 - 平均响应时间: %v, QPS: %.2f",
			result.AverageTime, result.RequestsPerSec)
	})

	t.Run("分类商品查询性能测试", func(t *testing.T) {
		concurrency := 150
		totalRequests := 1500

		requestFunc := func() *RequestResult {
			// 查询随机分类的商品
			categoryID := rand.Intn(20) + 1
			page := rand.Intn(10) + 1
			pageSize := rand.Intn(15) + 10
			path := fmt.Sprintf("/api/product/category/%d?page=%d&page_size=%d",
				categoryID, page, pageSize)

			return suite.MakeHTTPRequest("GET", path, nil)
		}

		result := suite.RunConcurrentTest(t, "分类商品查询", concurrency, totalRequests, requestFunc)

		// 验证性能指标
		assert.Less(t, result.AverageTime, 120*time.Millisecond, "平均响应时间应小于120ms")
		assert.Less(t, result.P95ResponseTime, 300*time.Millisecond, "P95响应时间应小于300ms")
		assert.Greater(t, result.RequestsPerSec, 400.0, "QPS应大于400")
		assert.Less(t, result.ErrorRate, 3.0, "错误率应小于3%")

		t.Logf("✅ 分类商品查询性能测试通过 - 平均响应时间: %v, QPS: %.2f",
			result.AverageTime, result.RequestsPerSec)
	})

	t.Run("商品高并发查询压力测试", func(t *testing.T) {
		concurrency := 500
		totalRequests := 5000

		requestFunc := func() *RequestResult {
			// 高并发查询热门商品
			hotProductIDs := []int{1, 2, 3, 4, 5, 10, 20, 50, 100, 200}
			productID := hotProductIDs[rand.Intn(len(hotProductIDs))]
			path := fmt.Sprintf("/api/product/%d", productID)

			return suite.MakeHTTPRequest("GET", path, nil)
		}

		result := suite.RunConcurrentTest(t, "商品高并发查询压力测试", concurrency, totalRequests, requestFunc)

		// 压力测试的验收标准
		assert.Less(t, result.AverageTime, 300*time.Millisecond, "压力测试平均响应时间应小于300ms")
		assert.Less(t, result.P95ResponseTime, 800*time.Millisecond, "压力测试P95响应时间应小于800ms")
		assert.Greater(t, result.RequestsPerSec, 200.0, "压力测试QPS应大于200")
		assert.Less(t, result.ErrorRate, 10.0, "压力测试错误率应小于10%")

		t.Logf("✅ 商品高并发查询压力测试通过 - 平均响应时间: %v, QPS: %.2f",
			result.AverageTime, result.RequestsPerSec)
	})

	t.Run("商品混合操作性能测试", func(t *testing.T) {
		concurrency := 100
		totalRequests := 1000

		operations := []string{"list", "detail", "search", "category"}

		requestFunc := func() *RequestResult {
			// 随机选择操作类型
			operation := operations[rand.Intn(len(operations))]

			switch operation {
			case "list":
				page := rand.Intn(5) + 1
				pageSize := rand.Intn(10) + 10
				path := fmt.Sprintf("/api/product/list?page=%d&page_size=%d", page, pageSize)
				return suite.MakeHTTPRequest("GET", path, nil)

			case "detail":
				productID := rand.Intn(1000) + 1
				path := fmt.Sprintf("/api/product/%d", productID)
				return suite.MakeHTTPRequest("GET", path, nil)

			case "search":
				keywords := []string{"测试", "商品", "电商"}
				keyword := keywords[rand.Intn(len(keywords))]
				path := fmt.Sprintf("/api/product/search?keyword=%s", keyword)
				return suite.MakeHTTPRequest("GET", path, nil)

			case "category":
				categoryID := rand.Intn(20) + 1
				path := fmt.Sprintf("/api/product/category/%d", categoryID)
				return suite.MakeHTTPRequest("GET", path, nil)

			default:
				return suite.MakeHTTPRequest("GET", "/api/product/list", nil)
			}
		}

		result := suite.RunConcurrentTest(t, "商品混合操作性能测试", concurrency, totalRequests, requestFunc)

		// 验证性能指标
		assert.Less(t, result.AverageTime, 180*time.Millisecond, "平均响应时间应小于180ms")
		assert.Less(t, result.P95ResponseTime, 450*time.Millisecond, "P95响应时间应小于450ms")
		assert.Greater(t, result.RequestsPerSec, 250.0, "QPS应大于250")
		assert.Less(t, result.ErrorRate, 5.0, "错误率应小于5%")

		t.Logf("✅ 商品混合操作性能测试通过 - 平均响应时间: %v, QPS: %.2f",
			result.AverageTime, result.RequestsPerSec)
	})
}

// BenchmarkProductList 商品列表基准测试
func BenchmarkProductList(b *testing.B) {
	suite := SetupPerformanceTest(&testing.T{})
	defer suite.CleanupPerformanceTest()

	// 创建测试数据
	suite.CreateTestData(&testing.T{})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			i++
			page := i%10 + 1
			pageSize := 20
			path := fmt.Sprintf("/api/product/list?page=%d&page_size=%d", page, pageSize)

			result := suite.MakeHTTPRequest("GET", path, nil)
			if !result.Success {
				b.Errorf("商品列表查询失败: %v", result.Error)
			}
		}
	})
}

// BenchmarkProductDetail 商品详情基准测试
func BenchmarkProductDetail(b *testing.B) {
	suite := SetupPerformanceTest(&testing.T{})
	defer suite.CleanupPerformanceTest()

	// 创建测试数据
	suite.CreateTestData(&testing.T{})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			i++
			productID := i%1000 + 1
			path := fmt.Sprintf("/api/product/%d", productID)

			result := suite.MakeHTTPRequest("GET", path, nil)
			if !result.Success {
				b.Errorf("商品详情查询失败: %v", result.Error)
			}
		}
	})
}

// BenchmarkProductSearch 商品搜索基准测试
func BenchmarkProductSearch(b *testing.B) {
	suite := SetupPerformanceTest(&testing.T{})
	defer suite.CleanupPerformanceTest()

	// 创建测试数据
	suite.CreateTestData(&testing.T{})

	keywords := []string{"测试", "商品", "电商", "购物", "手机"}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			i++
			keyword := keywords[i%len(keywords)]
			path := fmt.Sprintf("/api/product/search?keyword=%s", keyword)

			result := suite.MakeHTTPRequest("GET", path, nil)
			if !result.Success {
				b.Errorf("商品搜索失败: %v", result.Error)
			}
		}
	})
}
