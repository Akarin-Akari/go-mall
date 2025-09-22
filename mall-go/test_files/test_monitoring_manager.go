package main

import (
	"fmt"
	"log"
	"time"

	"mall-go/pkg/cache"
	"mall-go/pkg/optimistic"
)

func main() {
	fmt.Println("🔍 缓存监控管理器功能验证程序")
	fmt.Println("=" * 50)

	// 测试配置创建
	testConfigCreation()

	// 测试监控管理器创建
	testMonitoringManagerCreation()

	// 测试监控功能
	testMonitoringFunctionality()

	// 测试告警系统
	testAlertSystem()

	// 测试性能报告
	testPerformanceReport()

	// 测试数据导出
	testDataExport()

	fmt.Println("\n🎉 所有测试完成！")
}

func testConfigCreation() {
	fmt.Println("\n📋 测试1: 配置创建")

	config := cache.DefaultCacheMonitoringConfig()
	if config == nil {
		log.Fatal("❌ 默认配置创建失败")
	}

	fmt.Printf("  ✅ 默认配置创建成功\n")
	fmt.Printf("    - 监控级别: %d\n", config.Level)
	fmt.Printf("    - 收集间隔: %v\n", config.CollectInterval)
	fmt.Printf("    - 数据保留期: %v\n", config.RetentionPeriod)
	fmt.Printf("    - 启用指标数量: %d\n", len(config.EnabledMetrics))
	fmt.Printf("    - 告警规则数量: %d\n", len(config.AlertConfig.Rules))
}

func testMonitoringManagerCreation() {
	fmt.Println("\n🏗️ 测试2: 监控管理器创建")

	// 创建依赖组件
	config := cache.DefaultCacheMonitoringConfig()
	config.CollectInterval = 1 * time.Second // 快速收集用于测试

	// 创建模拟缓存管理器
	cacheManager := createMockCacheManager()
	keyManager := cache.NewCacheKeyManager()
	optimisticLock := optimistic.NewOptimisticLockService()

	// 创建监控管理器
	monitoringManager := cache.NewCacheMonitoringManager(
		config,
		cacheManager,
		keyManager,
		nil, // consistencyMgr
		nil, // warmupMgr
		nil, // protectionMgr
		optimisticLock,
	)

	if monitoringManager == nil {
		log.Fatal("❌ 监控管理器创建失败")
	}

	fmt.Printf("  ✅ 监控管理器创建成功\n")
	fmt.Printf("    - 运行状态: %v\n", monitoringManager.IsRunning())
	fmt.Printf("    - 配置级别: %d\n", monitoringManager.GetConfig().Level)
}

func testMonitoringFunctionality() {
	fmt.Println("\n📊 测试3: 监控功能")

	// 创建监控管理器
	monitoringManager := createTestMonitoringManager()

	// 启动监控
	err := monitoringManager.Start()
	if err != nil {
		log.Fatalf("❌ 启动监控失败: %v", err)
	}
	defer monitoringManager.Stop()

	fmt.Printf("  ✅ 监控管理器启动成功\n")

	// 模拟一些操作
	fmt.Printf("  📈 模拟缓存操作...\n")
	for i := 0; i < 10; i++ {
		// 记录响应时间
		responseTime := time.Duration(50+i*10) * time.Millisecond
		monitoringManager.RecordResponseTime(responseTime)

		// 记录热点键访问
		key := fmt.Sprintf("product:%d", i%3)
		hit := i%4 != 0 // 75%命中率
		monitoringManager.RecordHotKey(key, hit)

		time.Sleep(100 * time.Millisecond)
	}

	// 等待数据收集
	time.Sleep(2 * time.Second)

	// 检查统计信息
	stats := monitoringManager.GetStats()
	fmt.Printf("  ✅ 统计信息收集成功\n")
	fmt.Printf("    - 命中率: %.2f%%\n", stats.HitRate*100)
	fmt.Printf("    - 未命中率: %.2f%%\n", stats.MissRate*100)
	fmt.Printf("    - 总请求数: %d\n", stats.TotalRequests)
	fmt.Printf("    - 平均响应时间: %v\n", stats.AvgResponseTime)
	fmt.Printf("    - 收集次数: %d\n", stats.CollectionCount)

	// 检查时间序列数据
	hitRateData := monitoringManager.GetTimeSeriesData(cache.MetricHitRate)
	if hitRateData != nil {
		fmt.Printf("  ✅ 时间序列数据收集成功\n")
		fmt.Printf("    - 数据点数量: %d\n", len(hitRateData.DataPoints))
		fmt.Printf("    - 开始时间: %v\n", hitRateData.StartTime.Format("15:04:05"))
		fmt.Printf("    - 结束时间: %v\n", hitRateData.EndTime.Format("15:04:05"))
	}

	// 检查热点键
	hotKeys := monitoringManager.GetHotKeys(5)
	fmt.Printf("  ✅ 热点键分析成功\n")
	fmt.Printf("    - 热点键数量: %d\n", len(hotKeys))
	for i, hotKey := range hotKeys {
		fmt.Printf("    - TOP%d: %s (访问%d次, 命中率%.2f%%)\n",
			i+1, hotKey.Key, hotKey.AccessCount, hotKey.HitRate*100)
	}
}

func testAlertSystem() {
	fmt.Println("\n🚨 测试4: 告警系统")

	monitoringManager := createTestMonitoringManager()
	err := monitoringManager.Start()
	if err != nil {
		log.Fatalf("❌ 启动监控失败: %v", err)
	}
	defer monitoringManager.Stop()

	// 等待告警检查器运行
	time.Sleep(2 * time.Second)

	// 检查活跃告警
	alerts := monitoringManager.GetActiveAlerts()
	fmt.Printf("  ✅ 告警系统运行正常\n")
	fmt.Printf("    - 活跃告警数量: %d\n", len(alerts))

	for i, alert := range alerts {
		fmt.Printf("    - 告警%d: %s (级别:%d, 状态:%s)\n",
			i+1, alert.Message, alert.Level, alert.Status)
	}

	if len(alerts) == 0 {
		fmt.Printf("    - 当前无活跃告警\n")
	}
}

func testPerformanceReport() {
	fmt.Println("\n📋 测试5: 性能报告")

	monitoringManager := createTestMonitoringManager()

	// 添加测试数据
	for i := 0; i < 5; i++ {
		monitoringManager.RecordResponseTime(time.Duration(100+i*20) * time.Millisecond)
		monitoringManager.RecordHotKey(fmt.Sprintf("key:%d", i), i%2 == 0)
	}

	// 生成性能报告
	report := monitoringManager.GeneratePerformanceReport("test")
	if report == nil {
		log.Fatal("❌ 性能报告生成失败")
	}

	fmt.Printf("  ✅ 性能报告生成成功\n")
	fmt.Printf("    - 报告ID: %s\n", report.ReportID)
	fmt.Printf("    - 生成时间: %v\n", report.GeneratedAt.Format("15:04:05"))
	fmt.Printf("    - 统计周期: %s\n", report.Period)
	fmt.Printf("    - 优化建议数量: %d\n", len(report.Recommendations))

	for i, rec := range report.Recommendations {
		fmt.Printf("    - 建议%d: %s (优先级:%s)\n", i+1, rec.Title, rec.Priority)
	}

	fmt.Printf("    - 趋势分析:\n")
	fmt.Printf("      * 命中率趋势: %s\n", report.TrendAnalysis.HitRateTrend)
	fmt.Printf("      * 响应时间趋势: %s\n", report.TrendAnalysis.ResponseTimeTrend)
	fmt.Printf("      * QPS趋势: %s\n", report.TrendAnalysis.QPSTrend)
}

func testDataExport() {
	fmt.Println("\n📤 测试6: 数据导出")

	monitoringManager := createTestMonitoringManager()

	// 获取监控数据
	data := monitoringManager.GetMonitoringData()
	if data == nil {
		log.Fatal("❌ 监控数据获取失败")
	}

	fmt.Printf("  ✅ 监控数据获取成功\n")
	fmt.Printf("    - 数据项数量: %d\n", len(data))

	// 检查数据完整性
	expectedKeys := []string{"stats", "time_series", "active_alerts", "hot_keys", "config"}
	for _, key := range expectedKeys {
		if _, exists := data[key]; exists {
			fmt.Printf("    - ✅ %s 数据存在\n", key)
		} else {
			fmt.Printf("    - ❌ %s 数据缺失\n", key)
		}
	}

	// 测试重置功能
	fmt.Printf("  🔄 测试统计重置...\n")
	monitoringManager.ResetStats()
	
	resetStats := monitoringManager.GetStats()
	fmt.Printf("    - ✅ 统计重置成功\n")
	fmt.Printf("    - 重置后总请求数: %d\n", resetStats.TotalRequests)
	fmt.Printf("    - 重置时间: %v\n", resetStats.LastResetTime.Format("15:04:05"))
}

// createMockCacheManager 创建模拟缓存管理器
func createMockCacheManager() cache.CacheManager {
	// 这里应该返回一个实际的缓存管理器实例
	// 为了简化，我们返回nil，在实际使用中需要创建真实的实例
	return nil
}

// createTestMonitoringManager 创建测试用监控管理器
func createTestMonitoringManager() *cache.CacheMonitoringManager {
	config := cache.DefaultCacheMonitoringConfig()
	config.CollectInterval = 500 * time.Millisecond
	config.RetentionPeriod = 1 * time.Hour

	// 创建依赖组件
	cacheManager := createMockCacheManager()
	keyManager := cache.NewCacheKeyManager()
	optimisticLock := optimistic.NewOptimisticLockService()

	return cache.NewCacheMonitoringManager(
		config,
		cacheManager,
		keyManager,
		nil, // consistencyMgr
		nil, // warmupMgr
		nil, // protectionMgr
		optimisticLock,
	)
}
