package performance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"mall-go/internal/config"
	"mall-go/internal/handler"
	"mall-go/internal/model"

	"github.com/glebarez/sqlite"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PerformanceTestSuite 性能测试套件
type PerformanceTestSuite struct {
	db     *gorm.DB
	router *gin.Engine
	server *httptest.Server
}

// PerformanceResult 性能测试结果
type PerformanceResult struct {
	TotalRequests    int           `json:"total_requests"`
	SuccessRequests  int           `json:"success_requests"`
	FailedRequests   int           `json:"failed_requests"`
	TotalTime        time.Duration `json:"total_time"`
	AverageTime      time.Duration `json:"average_time"`
	MinTime          time.Duration `json:"min_time"`
	MaxTime          time.Duration `json:"max_time"`
	RequestsPerSec   float64       `json:"requests_per_sec"`
	P95ResponseTime  time.Duration `json:"p95_response_time"`
	P99ResponseTime  time.Duration `json:"p99_response_time"`
	ErrorRate        float64       `json:"error_rate"`
}

// RequestResult 单个请求结果
type RequestResult struct {
	Duration   time.Duration
	StatusCode int
	Success    bool
	Error      error
}

// SetupPerformanceTest 设置性能测试环境
func SetupPerformanceTest(t *testing.T) *PerformanceTestSuite {
	// 初始化配置
	config.GlobalConfig = config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key-for-performance-testing",
			Expire: "24h",
		},
		Database: config.DatabaseConfig{
			Driver:   "sqlite",
			Host:     "",
			Port:     0,
			Username: "",
			Password: "",
			DBName:   ":memory:",
		},
	}

	// 初始化测试数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	assert.NoError(t, err, "数据库连接失败")

	// 自动迁移
	err = db.AutoMigrate(
		&model.User{},
		&model.Product{},
		&model.ProductImage{},
		&model.ProductSKU{},
		&model.Category{},
		&model.Brand{},
		&model.Cart{},
		&model.CartItem{},
		&model.Order{},
		&model.OrderItem{},
		&model.OrderStatusLog{},
		&model.Payment{},
	)
	assert.NoError(t, err, "数据库迁移失败")

	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 创建路由
	r := gin.Default()
	
	// 注册简化的路由用于测试
	handler.RegisterMiddleware(r)
	handler.RegisterRoutes(r, db, nil, nil)

	// 创建测试服务器
	server := httptest.NewServer(r)

	return &PerformanceTestSuite{
		db:     db,
		router: r,
		server: server,
	}
}

// CleanupPerformanceTest 清理性能测试环境
func (pts *PerformanceTestSuite) CleanupPerformanceTest() {
	if pts.server != nil {
		pts.server.Close()
	}
	if pts.db != nil {
		sqlDB, err := pts.db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}

// CreateTestData 创建测试数据
func (pts *PerformanceTestSuite) CreateTestData(t *testing.T) {
	// 创建测试用户
	for i := 1; i <= 100; i++ {
		user := &model.User{
			Username: fmt.Sprintf("perfuser%d", i),
			Email:    fmt.Sprintf("perf%d@example.com", i),
			Password: "hashedpassword",
			Phone:    fmt.Sprintf("1380013%04d", i),
			Status:   "active",
		}
		err := pts.db.Create(user).Error
		assert.NoError(t, err, "创建测试用户失败")
	}

	// 创建商家用户
	for i := 1; i <= 10; i++ {
		merchant := &model.User{
			Username: fmt.Sprintf("perfmerchant%d", i),
			Email:    fmt.Sprintf("perfmerchant%d@example.com", i),
			Password: "hashedpassword",
			Phone:    fmt.Sprintf("1380014%04d", i),
			Role:     "merchant",
			Status:   "active",
		}
		err := pts.db.Create(merchant).Error
		assert.NoError(t, err, "创建商家用户失败")
	}

	// 创建分类
	for i := 1; i <= 20; i++ {
		category := &model.Category{
			Name:        fmt.Sprintf("性能测试分类%d", i),
			Description: fmt.Sprintf("performance-test-category-%d", i),
			Status:      "active",
		}
		err := pts.db.Create(category).Error
		assert.NoError(t, err, "创建分类失败")
	}

	// 创建商品
	for i := 1; i <= 1000; i++ {
		price, _ := decimal.NewFromString(fmt.Sprintf("%.2f", float64(i)*9.99))
		product := &model.Product{
			Name:        fmt.Sprintf("性能测试商品%d", i),
			Description: fmt.Sprintf("用于性能测试的商品%d", i),
			CategoryID:  uint((i-1)%20 + 1), // 分配到不同分类
			MerchantID:  uint((i-1)%10 + 101), // 分配到不同商家
			Price:       price,
			Stock:       1000,
			Status:      "active",
		}
		err := pts.db.Create(product).Error
		assert.NoError(t, err, "创建测试商品失败")
	}

	t.Logf("✅ 测试数据创建完成 - 用户: 100, 商家: 10, 分类: 20, 商品: 1000")
}

// RunConcurrentTest 运行并发测试
func (pts *PerformanceTestSuite) RunConcurrentTest(
	t *testing.T,
	testName string,
	concurrency int,
	totalRequests int,
	requestFunc func() *RequestResult,
) *PerformanceResult {
	
	t.Logf("🚀 开始性能测试: %s - 并发数: %d, 总请求数: %d", testName, concurrency, totalRequests)
	
	results := make(chan *RequestResult, totalRequests)
	var wg sync.WaitGroup
	
	startTime := time.Now()
	
	// 控制并发数
	semaphore := make(chan struct{}, concurrency)
	
	// 发送请求
	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			semaphore <- struct{}{} // 获取信号量
			defer func() { <-semaphore }() // 释放信号量
			
			result := requestFunc()
			results <- result
		}()
	}
	
	// 等待所有请求完成
	wg.Wait()
	close(results)
	
	totalTime := time.Since(startTime)
	
	// 统计结果
	var responseTimes []time.Duration
	successCount := 0
	failedCount := 0
	minTime := time.Duration(0)
	maxTime := time.Duration(0)
	
	for result := range results {
		responseTimes = append(responseTimes, result.Duration)
		
		if result.Success {
			successCount++
		} else {
			failedCount++
		}
		
		if minTime == 0 || result.Duration < minTime {
			minTime = result.Duration
		}
		if result.Duration > maxTime {
			maxTime = result.Duration
		}
	}
	
	// 计算平均时间
	var totalResponseTime time.Duration
	for _, duration := range responseTimes {
		totalResponseTime += duration
	}
	averageTime := totalResponseTime / time.Duration(len(responseTimes))
	
	// 计算P95和P99
	p95Time, p99Time := calculatePercentiles(responseTimes)
	
	// 计算QPS
	requestsPerSec := float64(totalRequests) / totalTime.Seconds()
	
	// 计算错误率
	errorRate := float64(failedCount) / float64(totalRequests) * 100
	
	result := &PerformanceResult{
		TotalRequests:   totalRequests,
		SuccessRequests: successCount,
		FailedRequests:  failedCount,
		TotalTime:       totalTime,
		AverageTime:     averageTime,
		MinTime:         minTime,
		MaxTime:         maxTime,
		RequestsPerSec:  requestsPerSec,
		P95ResponseTime: p95Time,
		P99ResponseTime: p99Time,
		ErrorRate:       errorRate,
	}
	
	// 输出测试结果
	t.Logf("📊 %s 性能测试结果:", testName)
	t.Logf("   总请求数: %d", result.TotalRequests)
	t.Logf("   成功请求: %d", result.SuccessRequests)
	t.Logf("   失败请求: %d", result.FailedRequests)
	t.Logf("   总耗时: %v", result.TotalTime)
	t.Logf("   平均响应时间: %v", result.AverageTime)
	t.Logf("   最小响应时间: %v", result.MinTime)
	t.Logf("   最大响应时间: %v", result.MaxTime)
	t.Logf("   QPS: %.2f", result.RequestsPerSec)
	t.Logf("   P95响应时间: %v", result.P95ResponseTime)
	t.Logf("   P99响应时间: %v", result.P99ResponseTime)
	t.Logf("   错误率: %.2f%%", result.ErrorRate)
	
	return result
}

// calculatePercentiles 计算百分位数
func calculatePercentiles(times []time.Duration) (p95, p99 time.Duration) {
	if len(times) == 0 {
		return 0, 0
	}
	
	// 简单排序
	for i := 0; i < len(times)-1; i++ {
		for j := 0; j < len(times)-i-1; j++ {
			if times[j] > times[j+1] {
				times[j], times[j+1] = times[j+1], times[j]
			}
		}
	}
	
	p95Index := int(float64(len(times)) * 0.95)
	p99Index := int(float64(len(times)) * 0.99)
	
	if p95Index >= len(times) {
		p95Index = len(times) - 1
	}
	if p99Index >= len(times) {
		p99Index = len(times) - 1
	}
	
	return times[p95Index], times[p99Index]
}

// MakeHTTPRequest 发送HTTP请求
func (pts *PerformanceTestSuite) MakeHTTPRequest(method, path string, body interface{}) *RequestResult {
	startTime := time.Now()
	
	var reqBody []byte
	var err error
	
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return &RequestResult{
				Duration:   time.Since(startTime),
				StatusCode: 0,
				Success:    false,
				Error:      err,
			}
		}
	}
	
	req, err := http.NewRequest(method, pts.server.URL+path, bytes.NewBuffer(reqBody))
	if err != nil {
		return &RequestResult{
			Duration:   time.Since(startTime),
			StatusCode: 0,
			Success:    false,
			Error:      err,
		}
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return &RequestResult{
			Duration:   time.Since(startTime),
			StatusCode: 0,
			Success:    false,
			Error:      err,
		}
	}
	defer resp.Body.Close()
	
	duration := time.Since(startTime)
	success := resp.StatusCode >= 200 && resp.StatusCode < 300
	
	return &RequestResult{
		Duration:   duration,
		StatusCode: resp.StatusCode,
		Success:    success,
		Error:      nil,
	}
}
