package performance

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"mall-go/internal/model"

	"github.com/stretchr/testify/assert"
)

// TestUserAPIPerformance 用户API性能测试
func TestUserAPIPerformance(t *testing.T) {
	// 设置性能测试环境
	suite := SetupPerformanceTest(t)
	defer suite.CleanupPerformanceTest()

	// 创建测试数据
	suite.CreateTestData(t)

	t.Run("用户注册性能测试", func(t *testing.T) {
		concurrency := 50
		totalRequests := 500

		requestFunc := func() *RequestResult {
			// 生成随机用户数据
			userID := rand.Intn(10000) + 1000
			registerReq := map[string]interface{}{
				"username": fmt.Sprintf("perfuser%d", userID),
				"email":    fmt.Sprintf("perfuser%d@example.com", userID),
				"password": "password123",
				"phone":    fmt.Sprintf("1380013%04d", userID%10000),
			}

			return suite.MakeHTTPRequest("POST", "/api/auth/register", registerReq)
		}

		result := suite.RunConcurrentTest(t, "用户注册", concurrency, totalRequests, requestFunc)

		// 验证性能指标
		assert.Less(t, result.AverageTime, 200*time.Millisecond, "平均响应时间应小于200ms")
		assert.Less(t, result.P95ResponseTime, 500*time.Millisecond, "P95响应时间应小于500ms")
		assert.Greater(t, result.RequestsPerSec, 100.0, "QPS应大于100")
		assert.Less(t, result.ErrorRate, 5.0, "错误率应小于5%")

		t.Logf("✅ 用户注册性能测试通过 - 平均响应时间: %v, QPS: %.2f",
			result.AverageTime, result.RequestsPerSec)
	})

	t.Run("用户登录性能测试", func(t *testing.T) {
		concurrency := 100
		totalRequests := 1000

		requestFunc := func() *RequestResult {
			// 使用已存在的用户进行登录
			userID := rand.Intn(100) + 1
			loginReq := map[string]interface{}{
				"username": fmt.Sprintf("perfuser%d", userID),
				"password": "hashedpassword",
			}

			return suite.MakeHTTPRequest("POST", "/api/auth/login", loginReq)
		}

		result := suite.RunConcurrentTest(t, "用户登录", concurrency, totalRequests, requestFunc)

		// 验证性能指标
		assert.Less(t, result.AverageTime, 150*time.Millisecond, "平均响应时间应小于150ms")
		assert.Less(t, result.P95ResponseTime, 300*time.Millisecond, "P95响应时间应小于300ms")
		assert.Greater(t, result.RequestsPerSec, 200.0, "QPS应大于200")
		assert.Less(t, result.ErrorRate, 3.0, "错误率应小于3%")

		t.Logf("✅ 用户登录性能测试通过 - 平均响应时间: %v, QPS: %.2f",
			result.AverageTime, result.RequestsPerSec)
	})

	t.Run("用户信息查询性能测试", func(t *testing.T) {
		concurrency := 200
		totalRequests := 2000

		requestFunc := func() *RequestResult {
			// 查询随机用户信息
			userID := rand.Intn(100) + 1
			return suite.MakeHTTPRequest("GET", fmt.Sprintf("/api/user/profile/%d", userID), nil)
		}

		result := suite.RunConcurrentTest(t, "用户信息查询", concurrency, totalRequests, requestFunc)

		// 验证性能指标
		assert.Less(t, result.AverageTime, 100*time.Millisecond, "平均响应时间应小于100ms")
		assert.Less(t, result.P95ResponseTime, 200*time.Millisecond, "P95响应时间应小于200ms")
		assert.Greater(t, result.RequestsPerSec, 500.0, "QPS应大于500")
		assert.Less(t, result.ErrorRate, 2.0, "错误率应小于2%")

		t.Logf("✅ 用户信息查询性能测试通过 - 平均响应时间: %v, QPS: %.2f",
			result.AverageTime, result.RequestsPerSec)
	})

	t.Run("用户并发注册压力测试", func(t *testing.T) {
		concurrency := 100
		totalRequests := 1000

		requestFunc := func() *RequestResult {
			// 生成唯一用户数据
			timestamp := time.Now().UnixNano()
			userID := rand.Intn(100000) + timestamp%100000
			registerReq := map[string]interface{}{
				"username": fmt.Sprintf("stressuser%d", userID),
				"email":    fmt.Sprintf("stressuser%d@example.com", userID),
				"password": "password123",
				"phone":    fmt.Sprintf("1380015%04d", userID%10000),
			}

			return suite.MakeHTTPRequest("POST", "/api/auth/register", registerReq)
		}

		result := suite.RunConcurrentTest(t, "用户并发注册压力测试", concurrency, totalRequests, requestFunc)

		// 压力测试的验收标准相对宽松
		assert.Less(t, result.AverageTime, 500*time.Millisecond, "压力测试平均响应时间应小于500ms")
		assert.Less(t, result.P95ResponseTime, 1*time.Second, "压力测试P95响应时间应小于1s")
		assert.Greater(t, result.RequestsPerSec, 50.0, "压力测试QPS应大于50")
		assert.Less(t, result.ErrorRate, 10.0, "压力测试错误率应小于10%")

		t.Logf("✅ 用户并发注册压力测试通过 - 平均响应时间: %v, QPS: %.2f",
			result.AverageTime, result.RequestsPerSec)
	})

	t.Run("用户登录高并发测试", func(t *testing.T) {
		concurrency := 500
		totalRequests := 5000

		requestFunc := func() *RequestResult {
			// 使用已存在的用户进行高并发登录
			userID := rand.Intn(100) + 1
			loginReq := map[string]interface{}{
				"username": fmt.Sprintf("perfuser%d", userID),
				"password": "hashedpassword",
			}

			return suite.MakeHTTPRequest("POST", "/api/auth/login", loginReq)
		}

		result := suite.RunConcurrentTest(t, "用户登录高并发测试", concurrency, totalRequests, requestFunc)

		// 高并发测试的验收标准
		assert.Less(t, result.AverageTime, 300*time.Millisecond, "高并发测试平均响应时间应小于300ms")
		assert.Less(t, result.P95ResponseTime, 800*time.Millisecond, "高并发测试P95响应时间应小于800ms")
		assert.Greater(t, result.RequestsPerSec, 100.0, "高并发测试QPS应大于100")
		assert.Less(t, result.ErrorRate, 15.0, "高并发测试错误率应小于15%")

		t.Logf("✅ 用户登录高并发测试通过 - 平均响应时间: %v, QPS: %.2f",
			result.AverageTime, result.RequestsPerSec)
	})
}

// BenchmarkUserRegister 用户注册基准测试
func BenchmarkUserRegister(b *testing.B) {
	suite := SetupPerformanceTest(&testing.T{})
	defer suite.CleanupPerformanceTest()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			i++
			registerReq := map[string]interface{}{
				"username": fmt.Sprintf("benchuser%d", i),
				"email":    fmt.Sprintf("benchuser%d@example.com", i),
				"password": "password123",
				"phone":    fmt.Sprintf("1380016%04d", i%10000),
			}

			result := suite.MakeHTTPRequest("POST", "/api/auth/register", registerReq)
			if !result.Success {
				b.Errorf("注册失败: %v", result.Error)
			}
		}
	})
}

// BenchmarkUserLogin 用户登录基准测试
func BenchmarkUserLogin(b *testing.B) {
	suite := SetupPerformanceTest(&testing.T{})
	defer suite.CleanupPerformanceTest()

	// 创建测试用户
	for i := 1; i <= 100; i++ {
		user := &model.User{
			Username: fmt.Sprintf("benchuser%d", i),
			Email:    fmt.Sprintf("benchuser%d@example.com", i),
			Password: "hashedpassword",
			Phone:    fmt.Sprintf("1380017%04d", i),
			Status:   "active",
		}
		suite.db.Create(user)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			i++
			userID := i%100 + 1
			loginReq := map[string]interface{}{
				"username": fmt.Sprintf("benchuser%d", userID),
				"password": "hashedpassword",
			}

			result := suite.MakeHTTPRequest("POST", "/api/auth/login", loginReq)
			if !result.Success {
				b.Errorf("登录失败: %v", result.Error)
			}
		}
	})
}

// BenchmarkUserProfile 用户信息查询基准测试
func BenchmarkUserProfile(b *testing.B) {
	suite := SetupPerformanceTest(&testing.T{})
	defer suite.CleanupPerformanceTest()

	// 创建测试用户
	for i := 1; i <= 100; i++ {
		user := &model.User{
			Username: fmt.Sprintf("benchuser%d", i),
			Email:    fmt.Sprintf("benchuser%d@example.com", i),
			Password: "hashedpassword",
			Phone:    fmt.Sprintf("1380018%04d", i),
			Status:   "active",
		}
		suite.db.Create(user)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			i++
			userID := i%100 + 1
			result := suite.MakeHTTPRequest("GET", fmt.Sprintf("/api/user/profile/%d", userID), nil)
			if !result.Success {
				b.Errorf("查询用户信息失败: %v", result.Error)
			}
		}
	})
}
