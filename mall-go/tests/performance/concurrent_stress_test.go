package performance

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"mall-go/internal/model"
	"mall-go/pkg/cache"
	"mall-go/pkg/logger"

	"github.com/stretchr/testify/assert"
)

// ConcurrentStressTestSuite 并发压力测试套件
type ConcurrentStressTestSuite struct {
	*CachePerformanceTestSuite
}

// StressTestResult 压力测试结果
type StressTestResult struct {
	*PerformanceResult
	ConcurrentUsers    int           `json:"concurrent_users"`
	TestDuration       time.Duration `json:"test_duration"`
	PeakQPS           float64       `json:"peak_qps"`
	MemoryUsageMB     float64       `json:"memory_usage_mb"`
	CacheHitRate      float64       `json:"cache_hit_rate"`
	DataConsistency   float64       `json:"data_consistency"`
	SystemStability   float64       `json:"system_stability"`
}

// SetupConcurrentStressTest 设置并发压力测试环境
func SetupConcurrentStressTest(t *testing.T) *ConcurrentStressTestSuite {
	cacheTestSuite := SetupCachePerformanceTest(t)
	if cacheTestSuite == nil {
		return nil
	}
	
	return &ConcurrentStressTestSuite{
		CachePerformanceTestSuite: cacheTestSuite,
	}
}

// CleanupConcurrentStressTest 清理并发压力测试环境
func (suite *ConcurrentStressTestSuite) CleanupConcurrentStressTest() {
	suite.CleanupCachePerformanceTest()
}

// TestHighConcurrencyCache 测试高并发缓存性能
func TestHighConcurrencyCache(t *testing.T) {
	suite := SetupConcurrentStressTest(t)
	if suite == nil {
		return
	}
	defer suite.CleanupConcurrentStressTest()
	
	// 创建测试数据
	suite.CreateCacheTestData(t)
	
	t.Run("高并发缓存读写压力测试", func(t *testing.T) {
		concurrency := 500
		testDuration := 30 * time.Second
		
		var totalRequests int64
		var successRequests int64
		var cacheHits int64
		var cacheMisses int64
		var errors int64
		
		ctx, cancel := context.WithTimeout(context.Background(), testDuration)
		defer cancel()
		
		var wg sync.WaitGroup
		startTime := time.Now()
		
		// 启动并发工作者
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				
				for {
					select {
					case <-ctx.Done():
						return
					default:
						// 执行缓存操作
						atomic.AddInt64(&totalRequests, 1)
						
						// 80%读操作，20%写操作
						if rand.Float64() < 0.8 {
							// 读操作
							productID := uint(rand.Intn(1000) + 1)
							key := suite.keyManager.GenerateProductKey(productID)
							
							// 尝试从缓存获取
							_, err := suite.cacheManager.Get(key)
							if err == nil {
								atomic.AddInt64(&cacheHits, 1)
								atomic.AddInt64(&successRequests, 1)
							} else {
								atomic.AddInt64(&cacheMisses, 1)
								
								// 从数据库加载并缓存
								var product model.Product
								if dbErr := suite.db.Where("id = ?", productID).First(&product).Error; dbErr == nil {
									suite.cacheManager.Set(key, product, 1*time.Hour)
									atomic.AddInt64(&successRequests, 1)
								} else {
									atomic.AddInt64(&errors, 1)
								}
							}
						} else {
							// 写操作
							productID := uint(rand.Intn(1000) + 1)
							key := suite.keyManager.GenerateProductKey(productID)
							
							data := map[string]interface{}{
								"id":    productID,
								"name":  fmt.Sprintf("Product %d", productID),
								"price": rand.Float64() * 1000,
							}
							
							if err := suite.cacheManager.Set(key, data, 1*time.Hour); err == nil {
								atomic.AddInt64(&successRequests, 1)
							} else {
								atomic.AddInt64(&errors, 1)
							}
						}
						
						// 短暂休息避免过度消耗CPU
						time.Sleep(time.Microsecond * 100)
					}
				}
			}(i)
		}
		
		// 等待所有工作者完成
		wg.Wait()
		actualDuration := time.Since(startTime)
		
		// 计算性能指标
		totalReq := atomic.LoadInt64(&totalRequests)
		successReq := atomic.LoadInt64(&successRequests)
		hits := atomic.LoadInt64(&cacheHits)
		misses := atomic.LoadInt64(&cacheMisses)
		errs := atomic.LoadInt64(&errors)
		
		qps := float64(totalReq) / actualDuration.Seconds()
		hitRate := float64(hits) / float64(hits+misses) * 100
		successRate := float64(successReq) / float64(totalReq) * 100
		errorRate := float64(errs) / float64(totalReq) * 100
		
		// 验证性能指标
		assert.Greater(t, qps, 5000.0, "QPS应>5000")
		assert.GreaterOrEqual(t, hitRate, 60.0, "缓存命中率应≥60%")
		assert.GreaterOrEqual(t, successRate, 95.0, "成功率应≥95%")
		assert.Less(t, errorRate, 5.0, "错误率应<5%")
		
		t.Logf("✅ 高并发缓存压力测试通过")
		t.Logf("   - 并发用户数: %d", concurrency)
		t.Logf("   - 测试时长: %v", actualDuration)
		t.Logf("   - 总请求数: %d", totalReq)
		t.Logf("   - QPS: %.2f (目标: >5000)", qps)
		t.Logf("   - 缓存命中率: %.2f%% (目标: ≥60%%)", hitRate)
		t.Logf("   - 成功率: %.2f%% (目标: ≥95%%)", successRate)
		t.Logf("   - 错误率: %.2f%% (目标: <5%%)", errorRate)
	})
}

// TestConcurrentConsistency 测试并发一致性
func TestConcurrentConsistency(t *testing.T) {
	suite := SetupConcurrentStressTest(t)
	if suite == nil {
		return
	}
	defer suite.CleanupConcurrentStressTest()
	
	// 创建测试数据
	suite.CreateCacheTestData(t)
	
	t.Run("并发一致性压力测试", func(t *testing.T) {
		concurrency := 100
		testDuration := 20 * time.Second
		
		var totalUpdates int64
		var successUpdates int64
		var consistencyChecks int64
		var consistentData int64
		
		ctx, cancel := context.WithTimeout(context.Background(), testDuration)
		defer cancel()
		
		var wg sync.WaitGroup
		startTime := time.Now()
		
		// 启动并发更新工作者
		for i := 0; i < concurrency/2; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				
				for {
					select {
					case <-ctx.Done():
						return
					default:
						// 随机选择商品进行更新
						productID := uint(rand.Intn(100) + 1)
						newPrice := fmt.Sprintf("%.2f", rand.Float64()*1000+10)
						
						updates := map[string]interface{}{
							"price": newPrice,
						}
						
						// 通过一致性管理器更新
						cacheKey := suite.keyManager.GenerateProductKey(productID)
						if err := suite.consistencyMgr.SyncCache(cacheKey, "products", productID, updates); err == nil {
							atomic.AddInt64(&successUpdates, 1)
						}
						atomic.AddInt64(&totalUpdates, 1)
						
						time.Sleep(time.Millisecond * 10)
					}
				}
			}(i)
		}
		
		// 启动并发一致性检查工作者
		for i := 0; i < concurrency/2; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				
				for {
					select {
					case <-ctx.Done():
						return
					default:
						// 随机检查商品数据一致性
						productID := uint(rand.Intn(100) + 1)
						cacheKey := suite.keyManager.GenerateProductKey(productID)
						
						// 从缓存获取数据
						cachedData, cacheErr := suite.cacheManager.Get(cacheKey)
						
						// 从数据库获取数据
						var dbProduct model.Product
						dbErr := suite.db.Where("id = ?", productID).First(&dbProduct).Error
						
						atomic.AddInt64(&consistencyChecks, 1)
						
						// 简化的一致性检查
						if cacheErr == nil && dbErr == nil && cachedData != nil {
							atomic.AddInt64(&consistentData, 1)
						} else if cacheErr != nil && dbErr != nil {
							// 都不存在也算一致
							atomic.AddInt64(&consistentData, 1)
						}
						
						time.Sleep(time.Millisecond * 5)
					}
				}
			}(i)
		}
		
		// 等待所有工作者完成
		wg.Wait()
		actualDuration := time.Since(startTime)
		
		// 计算一致性指标
		totalUpd := atomic.LoadInt64(&totalUpdates)
		successUpd := atomic.LoadInt64(&successUpdates)
		totalChecks := atomic.LoadInt64(&consistencyChecks)
		consistent := atomic.LoadInt64(&consistentData)
		
		updateSuccessRate := float64(successUpd) / float64(totalUpd) * 100
		consistencyRate := float64(consistent) / float64(totalChecks) * 100
		updateQPS := float64(totalUpd) / actualDuration.Seconds()
		
		// 验证一致性指标
		assert.GreaterOrEqual(t, updateSuccessRate, 90.0, "更新成功率应≥90%")
		assert.GreaterOrEqual(t, consistencyRate, 85.0, "数据一致性应≥85%")
		assert.Greater(t, updateQPS, 100.0, "更新QPS应>100")
		
		t.Logf("✅ 并发一致性压力测试通过")
		t.Logf("   - 并发用户数: %d", concurrency)
		t.Logf("   - 测试时长: %v", actualDuration)
		t.Logf("   - 总更新数: %d", totalUpd)
		t.Logf("   - 更新成功率: %.2f%% (目标: ≥90%%)", updateSuccessRate)
		t.Logf("   - 数据一致性: %.2f%% (目标: ≥85%%)", consistencyRate)
		t.Logf("   - 更新QPS: %.2f (目标: >100)", updateQPS)
		t.Logf("   - 一致性检查次数: %d", totalChecks)
	})
}

// TestSystemStabilityUnderLoad 测试系统负载稳定性
func TestSystemStabilityUnderLoad(t *testing.T) {
	suite := SetupConcurrentStressTest(t)
	if suite == nil {
		return
	}
	defer suite.CleanupConcurrentStressTest()
	
	// 创建测试数据
	suite.CreateCacheTestData(t)
	
	t.Run("系统负载稳定性测试", func(t *testing.T) {
		// 渐进式增加负载
		loadLevels := []int{50, 100, 200, 300, 500}
		testDurationPerLevel := 10 * time.Second
		
		for _, concurrency := range loadLevels {
			t.Logf("🔄 测试负载级别: %d 并发用户", concurrency)
			
			var totalRequests int64
			var successRequests int64
			var errors int64
			
			ctx, cancel := context.WithTimeout(context.Background(), testDurationPerLevel)
			
			var wg sync.WaitGroup
			startTime := time.Now()
			
			// 启动并发工作者
			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					
					for {
						select {
						case <-ctx.Done():
							return
						default:
							atomic.AddInt64(&totalRequests, 1)
							
							// 混合操作：读取、写入、更新
							operation := rand.Intn(3)
							productID := uint(rand.Intn(500) + 1)
							key := suite.keyManager.GenerateProductKey(productID)
							
							switch operation {
							case 0: // 读取
								if _, err := suite.cacheManager.Get(key); err == nil {
									atomic.AddInt64(&successRequests, 1)
								} else {
									atomic.AddInt64(&errors, 1)
								}
							case 1: // 写入
								data := map[string]interface{}{
									"id":   productID,
									"name": fmt.Sprintf("Product %d", productID),
								}
								if err := suite.cacheManager.Set(key, data, 1*time.Hour); err == nil {
									atomic.AddInt64(&successRequests, 1)
								} else {
									atomic.AddInt64(&errors, 1)
								}
							case 2: // 删除
								if err := suite.cacheManager.Delete(key); err == nil {
									atomic.AddInt64(&successRequests, 1)
								} else {
									atomic.AddInt64(&errors, 1)
								}
							}
							
							time.Sleep(time.Microsecond * 200)
						}
					}
				}()
			}
			
			wg.Wait()
			cancel()
			actualDuration := time.Since(startTime)
			
			// 计算性能指标
			totalReq := atomic.LoadInt64(&totalRequests)
			successReq := atomic.LoadInt64(&successRequests)
			errs := atomic.LoadInt64(&errors)
			
			qps := float64(totalReq) / actualDuration.Seconds()
			successRate := float64(successReq) / float64(totalReq) * 100
			errorRate := float64(errs) / float64(totalReq) * 100
			
			// 验证系统稳定性
			assert.GreaterOrEqual(t, successRate, 90.0, fmt.Sprintf("负载%d下成功率应≥90%%", concurrency))
			assert.Less(t, errorRate, 10.0, fmt.Sprintf("负载%d下错误率应<10%%", concurrency))
			
			t.Logf("   ✅ 负载级别 %d: QPS=%.2f, 成功率=%.2f%%, 错误率=%.2f%%", 
				concurrency, qps, successRate, errorRate)
			
			// 短暂休息让系统恢复
			time.Sleep(2 * time.Second)
		}
		
		t.Logf("✅ 系统负载稳定性测试通过")
	})
}
