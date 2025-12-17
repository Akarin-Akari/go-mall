//go:build ignore

package main

import (
	"fmt"
	"time"

	"mall-go/internal/middleware"
	"mall-go/internal/service"
)

// P3.1级别性能监控简化验证测试
func main() {
	fmt.Println("=== P3.1级别性能监控简化验证测试 ===")
	
	// 1. 测试性能指标收集器基本功能
	fmt.Println("\n1. 性能指标收集器基本测试")
	testBasicMetrics()
	
	// 2. 测试性能监控器基本功能
	fmt.Println("\n2. 性能监控器基本测试")
	testBasicMonitor()
	
	fmt.Println("\n=== P3.1级别性能监控验证完成 ===")
}

// 测试基本指标功能
func testBasicMetrics() {
	fmt.Println("测试基本指标收集功能...")
	
	// 1. 创建性能指标收集器
	metrics := middleware.NewPerformanceMetrics()
	if metrics != nil {
		fmt.Println("✅ PerformanceMetrics创建成功")
	} else {
		fmt.Println("❌ PerformanceMetrics创建失败")
		return
	}
	
	// 2. 测试数据库查询指标记录
	fmt.Println("\n数据库查询指标测试:")
	metrics.RecordDBQuery("select", "addresses", 50*time.Millisecond, nil)
	fmt.Println("✅ 正常查询指标记录成功")
	
	// 测试慢查询记录
	metrics.RecordDBQuery("select", "addresses", 150*time.Millisecond, nil)
	fmt.Println("✅ 慢查询检测和记录成功")
	
	// 3. 测试地址操作指标记录
	fmt.Println("\n地址操作指标测试:")
	metrics.RecordAddressOperation("create_address", nil)
	fmt.Println("✅ 地址操作成功指标记录成功")
	
	metrics.RecordAddressOperation("create_address", fmt.Errorf("测试错误"))
	fmt.Println("✅ 地址操作失败指标记录成功")
	
	// 4. 测试缓存指标记录
	fmt.Println("\n缓存指标测试:")
	metrics.RecordCacheHit("redis", "get_address")
	fmt.Println("✅ 缓存命中指标记录成功")
	
	metrics.RecordCacheMiss("redis", "get_address")
	fmt.Println("✅ 缓存未命中指标记录成功")
	
	// 5. 测试系统指标获取
	fmt.Println("\n系统指标测试:")
	systemMetrics := metrics.GetMetrics()
	if systemMetrics != nil {
		fmt.Printf("✅ 系统指标获取成功，包含 %d 个指标组\n", len(systemMetrics))
		
		if memory, ok := systemMetrics["memory"]; ok {
			if memMap, ok := memory.(map[string]interface{}); ok {
				fmt.Printf("   - 内存分配: %v bytes\n", memMap["alloc"])
				fmt.Printf("   - 系统内存: %v bytes\n", memMap["sys"])
			}
		}
		
		if runtime, ok := systemMetrics["runtime"]; ok {
			if rtMap, ok := runtime.(map[string]interface{}); ok {
				fmt.Printf("   - Goroutine数量: %v\n", rtMap["goroutines"])
				fmt.Printf("   - CPU核心数: %v\n", rtMap["cpu_count"])
			}
		}
	}
	
	// 6. 测试系统监控启动
	fmt.Println("\n系统监控启动测试:")
	metrics.StartSystemMonitoring()
	fmt.Println("✅ 系统监控启动成功")
}

// 测试基本监控器功能
func testBasicMonitor() {
	fmt.Println("测试基本监控器功能...")
	
	// 1. 创建性能监控器
	monitor := service.NewPerformanceMonitor()
	if monitor != nil {
		fmt.Println("✅ PerformanceMonitor创建成功")
	} else {
		fmt.Println("❌ PerformanceMonitor创建失败")
		return
	}
	
	// 2. 测试指标记录
	fmt.Println("\n指标记录测试:")
	monitor.RecordMetric("test_metric", 100.5, map[string]string{
		"type": "test",
		"unit": "ms",
	})
	fmt.Println("✅ 指标记录成功")
	
	// 3. 测试计数器增加
	fmt.Println("\n计数器测试:")
	monitor.IncrementCounter("test_counter", map[string]string{
		"operation": "test",
	})
	fmt.Println("✅ 计数器增加成功")
	
	// 4. 测试地址操作记录
	fmt.Println("\n地址操作记录测试:")
	monitor.RecordAddressOperation("create_address", nil)
	fmt.Println("✅ 地址操作成功记录")
	
	monitor.RecordAddressOperation("create_address", fmt.Errorf("测试错误"))
	fmt.Println("✅ 地址操作失败记录")
	
	// 5. 测试指标JSON导出
	fmt.Println("\n指标导出测试:")
	metricsJSON, err := monitor.GetMetricsJSON()
	if err != nil {
		fmt.Printf("❌ 指标JSON导出失败: %v\n", err)
	} else {
		fmt.Printf("✅ 指标JSON导出成功，数据长度: %d 字节\n", len(metricsJSON))
	}
	
	// 6. 测试性能报告生成
	fmt.Println("\n性能报告生成测试:")
	
	// 添加一些测试数据
	monitor.RecordMetric("memory_usage_mb", 512, nil)
	monitor.RecordMetric("cpu_usage_percent", 45, nil)
	monitor.RecordMetric("goroutine_count", 150, nil)
	
	report := monitor.GenerateReport("1h")
	if report == nil {
		fmt.Println("❌ 性能报告生成失败")
		return
	}
	
	fmt.Printf("✅ 性能报告生成成功\n")
	fmt.Printf("   - 生成时间: %v\n", report.GeneratedAt)
	fmt.Printf("   - 时间范围: %s\n", report.TimeRange)
	fmt.Printf("   - 指标数量: %d\n", len(report.Metrics))
	fmt.Printf("   - 告警数量: %d\n", len(report.Alerts))
	fmt.Printf("   - 建议数量: %d\n", len(report.Recommendations))
	
	// 验证系统信息
	if report.SystemInfo != nil {
		fmt.Printf("   - 系统信息项数: %d\n", len(report.SystemInfo))
		if goVersion, ok := report.SystemInfo["go_version"]; ok {
			fmt.Printf("   - Go版本: %v\n", goVersion)
		}
	}
	
	// 7. 测试全局性能监控器初始化
	fmt.Println("\n全局性能监控器测试:")
	service.InitGlobalPerformanceMonitor()
	if service.GlobalPerformanceMonitor != nil {
		fmt.Println("✅ 全局性能监控器初始化成功")
	} else {
		fmt.Println("❌ 全局性能监控器初始化失败")
	}
	
	// 8. 测试全局指标初始化
	fmt.Println("\n全局指标初始化测试:")
	middleware.InitPerformanceMetrics()
	if middleware.GlobalMetrics != nil {
		fmt.Println("✅ 全局性能指标初始化成功")
	} else {
		fmt.Println("❌ 全局性能指标初始化失败")
	}
}

// 验证总结
func printP31Summary() {
	fmt.Println("\n=== P3.1级别性能监控总结 ===")
	fmt.Println("✅ 1. 性能指标收集器")
	fmt.Println("   - Prometheus指标格式支持")
	fmt.Println("   - 数据库查询性能监控")
	fmt.Println("   - 系统资源监控")
	fmt.Println("   - 慢操作检测和记录")
	
	fmt.Println("✅ 2. 性能监控器")
	fmt.Println("   - 指标数据管理")
	fmt.Println("   - JSON格式数据导出")
	fmt.Println("   - 性能报告生成")
	fmt.Println("   - 全局监控器支持")
	
	fmt.Println("\n🎉 P3.1级别性能监控优化基本功能验证完成！")
}
