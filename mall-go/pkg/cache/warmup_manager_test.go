package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"mall-go/pkg/optimistic"
)

// 创建测试用的缓存预热管理器
func createTestWarmupManager() (*CacheWarmupManager, *SharedMockCacheManager, *SharedMockOptimisticLockService) {
	mockCache := &SharedMockCacheManager{}
	mockOptimistic := &SharedMockOptimisticLockService{}

	// 初始化键管理器
	InitKeyManager("test")
	keyManager := GetKeyManager()

	// 创建一致性管理器
	consistencyConfig := DefaultCacheConsistencyConfig()
	consistencyMgr := NewCacheConsistencyManager(consistencyConfig, mockCache, keyManager, mockOptimistic)

	// 创建预热配置
	config := DefaultCacheWarmupConfig()
	config.BatchSize = 5                           // 减少批次大小用于测试
	config.MaxConcurrency = 2                      // 减少并发数用于测试
	config.ReportInterval = 100 * time.Millisecond // 缩短报告间隔

	cwm := NewCacheWarmupManager(config, mockCache, keyManager, consistencyMgr, mockOptimistic)

	return cwm, mockCache, mockOptimistic
}

// TestCacheWarmupManagerCreation 测试缓存预热管理器创建
func TestCacheWarmupManagerCreation(t *testing.T) {
	cwm, _, _ := createTestWarmupManager()

	assert.NotNil(t, cwm)
	assert.NotNil(t, cwm.config)
	assert.NotNil(t, cwm.cacheManager)
	assert.NotNil(t, cwm.keyManager)
	assert.NotNil(t, cwm.consistencyMgr)
	assert.NotNil(t, cwm.optimisticLock)
	assert.NotNil(t, cwm.progress)
	assert.NotNil(t, cwm.stats)
	assert.False(t, cwm.IsRunning())
}

// TestCacheWarmupManagerStartStop 测试启动和停止
func TestCacheWarmupManagerStartStop(t *testing.T) {
	cwm, _, _ := createTestWarmupManager()

	// 测试启动
	err := cwm.Start()
	assert.NoError(t, err)
	assert.True(t, cwm.IsRunning())

	// 测试重复启动
	err = cwm.Start()
	assert.Error(t, err)

	// 等待一小段时间让进度报告器启动
	time.Sleep(50 * time.Millisecond)

	// 测试停止
	err = cwm.Stop()
	assert.NoError(t, err)
	assert.False(t, cwm.IsRunning())

	// 测试重复停止
	err = cwm.Stop()
	assert.NoError(t, err)
}

// TestCacheWarmupManagerConfig 测试配置管理
func TestCacheWarmupManagerConfig(t *testing.T) {
	cwm, _, _ := createTestWarmupManager()

	config := cwm.GetConfig()
	assert.NotNil(t, config)
	assert.True(t, config.Enabled)
	assert.Equal(t, WarmupModeAsync, config.Mode)
	assert.Equal(t, 5, config.BatchSize)
	assert.Equal(t, 2, config.MaxConcurrency)
}

// TestCacheWarmupManagerHotDataIdentification 测试热点数据识别
func TestCacheWarmupManagerHotDataIdentification(t *testing.T) {
	cwm, _, _ := createTestWarmupManager()

	t.Run("识别热门商品", func(t *testing.T) {
		hotProducts, err := cwm.identifyHotProducts()
		assert.NoError(t, err)
		assert.NotEmpty(t, hotProducts)
		assert.Contains(t, hotProducts, uint(1))
	})

	t.Run("识别活跃用户", func(t *testing.T) {
		activeUsers, err := cwm.identifyActiveUsers()
		assert.NoError(t, err)
		assert.NotEmpty(t, activeUsers)
		assert.Contains(t, activeUsers, uint(1))
	})

	t.Run("识别热门分类", func(t *testing.T) {
		hotCategories, err := cwm.identifyHotCategories()
		assert.NoError(t, err)
		assert.NotEmpty(t, hotCategories)
		assert.Contains(t, hotCategories, uint(1))
	})
}

// TestCacheWarmupManagerTaskCreation 测试任务创建
func TestCacheWarmupManagerTaskCreation(t *testing.T) {
	cwm, _, _ := createTestWarmupManager()

	t.Run("创建热门商品任务", func(t *testing.T) {
		err := cwm.createHotProductTasks()
		assert.NoError(t, err)

		// 检查任务队列
		cwm.taskQueueMutex.RLock()
		taskCount := len(cwm.taskQueue)
		cwm.taskQueueMutex.RUnlock()

		assert.Greater(t, taskCount, 0)
	})

	t.Run("创建活跃用户任务", func(t *testing.T) {
		err := cwm.createActiveUserTasks()
		assert.NoError(t, err)

		// 检查任务队列
		cwm.taskQueueMutex.RLock()
		taskCount := len(cwm.taskQueue)
		cwm.taskQueueMutex.RUnlock()

		assert.Greater(t, taskCount, 0)
	})

	t.Run("创建分类任务", func(t *testing.T) {
		err := cwm.createCategoryTasks()
		assert.NoError(t, err)

		// 检查任务队列
		cwm.taskQueueMutex.RLock()
		taskCount := len(cwm.taskQueue)
		cwm.taskQueueMutex.RUnlock()

		assert.Greater(t, taskCount, 0)
	})
}

// TestCacheWarmupManagerBatchTasks 测试批量任务创建
func TestCacheWarmupManagerBatchTasks(t *testing.T) {
	cwm, _, _ := createTestWarmupManager()

	dataIDs := []uint{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}

	err := cwm.createBatchTasks(WarmupHotProducts, PriorityHigh, "product", dataIDs)
	assert.NoError(t, err)

	// 检查任务队列
	cwm.taskQueueMutex.RLock()
	tasks := cwm.taskQueue
	cwm.taskQueueMutex.RUnlock()

	// 应该创建3个批次（每批5个，总共12个数据）
	expectedBatches := 3
	assert.Equal(t, expectedBatches, len(tasks))

	// 检查第一个任务
	firstTask := tasks[0]
	assert.Equal(t, WarmupHotProducts, firstTask.Strategy)
	assert.Equal(t, PriorityHigh, firstTask.Priority)
	assert.Equal(t, "product", firstTask.DataType)
	assert.Equal(t, 5, len(firstTask.DataIDs))
	assert.Equal(t, 0, firstTask.BatchIndex)
	assert.Equal(t, expectedBatches, firstTask.TotalBatch)
	assert.Equal(t, WarmupStatusPending, firstTask.Status)
}

// TestCacheWarmupManagerCacheKeyGeneration 测试缓存键生成
func TestCacheWarmupManagerCacheKeyGeneration(t *testing.T) {
	cwm, _, _ := createTestWarmupManager()

	t.Run("商品缓存键", func(t *testing.T) {
		dataIDs := []uint{1, 2, 3}
		keys := cwm.generateCacheKeys("product", dataIDs)

		assert.Equal(t, len(dataIDs), len(keys))
		assert.Equal(t, "test:product:1", keys[0])
		assert.Equal(t, "test:product:2", keys[1])
		assert.Equal(t, "test:product:3", keys[2])
	})

	t.Run("用户缓存键", func(t *testing.T) {
		dataIDs := []uint{1, 2}
		keys := cwm.generateCacheKeys("user", dataIDs)

		assert.Equal(t, len(dataIDs), len(keys))
		assert.Equal(t, "test:user:session:1", keys[0])
		assert.Equal(t, "test:user:session:2", keys[1])
	})

	t.Run("分类缓存键", func(t *testing.T) {
		dataIDs := []uint{1, 2}
		keys := cwm.generateCacheKeys("category", dataIDs)

		assert.Equal(t, len(dataIDs), len(keys))
		assert.Equal(t, "category:1", keys[0])
		assert.Equal(t, "category:2", keys[1])
	})
}

// TestCacheWarmupManagerTaskSorting 测试任务排序
func TestCacheWarmupManagerTaskSorting(t *testing.T) {
	cwm, _, _ := createTestWarmupManager()

	// 创建不同优先级的任务
	task1 := &WarmupTask{ID: "task1", Priority: PriorityLow}
	task2 := &WarmupTask{ID: "task2", Priority: PriorityHigh}
	task3 := &WarmupTask{ID: "task3", Priority: PriorityMedium}

	cwm.taskQueue = []*WarmupTask{task1, task2, task3}

	// 排序
	cwm.sortTasksByPriority()

	// 检查排序结果（高优先级应该在前面）
	assert.Equal(t, PriorityHigh, cwm.taskQueue[0].Priority)
	assert.Equal(t, PriorityMedium, cwm.taskQueue[1].Priority)
	assert.Equal(t, PriorityLow, cwm.taskQueue[2].Priority)
}

// TestCacheWarmupManagerProgress 测试进度管理
func TestCacheWarmupManagerProgress(t *testing.T) {
	cwm, _, _ := createTestWarmupManager()

	// 重置进度
	cwm.resetProgress()

	progress := cwm.GetProgress()
	assert.NotNil(t, progress)
	assert.Equal(t, 0, progress.TotalTasks)
	assert.Equal(t, 0, progress.CompletedTasks)
	assert.Equal(t, 0, progress.FailedTasks)
	assert.Equal(t, 0, progress.RunningTasks)
	assert.Equal(t, 0, progress.PendingTasks)
	assert.Equal(t, float64(0), progress.ProgressRate)
}

// TestCacheWarmupManagerStats 测试统计功能
func TestCacheWarmupManagerStats(t *testing.T) {
	cwm, _, _ := createTestWarmupManager()

	// 获取初始统计
	stats := cwm.GetStats()
	assert.NotNil(t, stats)
	assert.Equal(t, int64(0), stats.TotalWarmups)
	assert.Equal(t, int64(0), stats.SuccessfulWarmups)
	assert.Equal(t, int64(0), stats.FailedWarmups)
	assert.Equal(t, float64(0), stats.SuccessRate)

	// 更新统计
	cwm.updateStrategyStats(WarmupHotProducts, true, 100*time.Millisecond, 5)

	// 检查统计更新
	stats = cwm.GetStats()
	assert.Equal(t, int64(1), stats.TotalWarmups)
	assert.Equal(t, int64(1), stats.SuccessfulWarmups)
	assert.Equal(t, int64(5), stats.TotalDataWarmed)
	assert.Equal(t, float64(100), stats.SuccessRate)

	// 检查策略统计
	strategyStats := stats.StrategyStats[WarmupHotProducts]
	assert.NotNil(t, strategyStats)
	assert.Equal(t, int64(1), strategyStats.ExecutionCount)
	assert.Equal(t, int64(1), strategyStats.SuccessCount)
	assert.Equal(t, int64(5), strategyStats.TotalDataWarmed)

	// 重置统计
	cwm.ResetStats()
	stats = cwm.GetStats()
	assert.Equal(t, int64(0), stats.TotalWarmups)
	assert.Equal(t, int64(0), stats.SuccessfulWarmups)
}
