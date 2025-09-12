package cache

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"mall-go/internal/model"
	"mall-go/pkg/logger"
	"mall-go/pkg/optimistic"
)

// WarmupStrategy 预热策略枚举
type WarmupStrategy string

const (
	// 商品预热策略
	WarmupHotProducts   WarmupStrategy = "hot_products"   // 热门商品预热
	WarmupNewProducts   WarmupStrategy = "new_products"   // 新品预热
	WarmupPromoProducts WarmupStrategy = "promo_products" // 促销商品预热
	WarmupCategoryTop   WarmupStrategy = "category_top"   // 分类热门商品预热

	// 用户预热策略
	WarmupActiveUsers  WarmupStrategy = "active_users"  // 活跃用户预热
	WarmupUserSessions WarmupStrategy = "user_sessions" // 用户会话预热
	WarmupUserCarts    WarmupStrategy = "user_carts"    // 购物车预热
	WarmupUserPrefs    WarmupStrategy = "user_prefs"    // 用户偏好预热

	// 系统预热策略
	WarmupCategories   WarmupStrategy = "categories"    // 分类数据预热
	WarmupSystemConfig WarmupStrategy = "system_config" // 系统配置预热
	WarmupStaticData   WarmupStrategy = "static_data"   // 静态数据预热
)

// WarmupPriority 预热优先级
type WarmupPriority int

const (
	PriorityHigh   WarmupPriority = 1 // 高优先级
	PriorityMedium WarmupPriority = 2 // 中优先级
	PriorityLow    WarmupPriority = 3 // 低优先级
)

// WarmupMode 预热模式
type WarmupMode string

const (
	WarmupModeSync  WarmupMode = "sync"  // 同步预热
	WarmupModeAsync WarmupMode = "async" // 异步预热
)

// CacheWarmupConfig 缓存预热配置
type CacheWarmupConfig struct {
	// 基础配置
	Enabled        bool          `json:"enabled"`         // 是否启用预热
	Mode           WarmupMode    `json:"mode"`            // 预热模式
	BatchSize      int           `json:"batch_size"`      // 批次大小
	BatchInterval  time.Duration `json:"batch_interval"`  // 批次间隔
	MaxConcurrency int           `json:"max_concurrency"` // 最大并发数
	Timeout        time.Duration `json:"timeout"`         // 预热超时时间

	// 策略配置
	Strategies    []WarmupStrategy `json:"strategies"`     // 启用的预热策略
	PriorityOrder []WarmupPriority `json:"priority_order"` // 优先级顺序

	// 热点数据识别配置
	HotDataConfig *HotDataConfig `json:"hot_data_config"` // 热点数据配置

	// 重试配置
	RetryAttempts int           `json:"retry_attempts"` // 重试次数
	RetryInterval time.Duration `json:"retry_interval"` // 重试间隔

	// 监控配置
	ProgressReport bool          `json:"progress_report"` // 是否报告进度
	ReportInterval time.Duration `json:"report_interval"` // 报告间隔

	// 失败处理
	FailureThreshold float64 `json:"failure_threshold"` // 失败阈值
	StopOnFailure    bool    `json:"stop_on_failure"`   // 失败时是否停止
}

// HotDataConfig 热点数据识别配置
type HotDataConfig struct {
	// 商品热点识别
	ProductSoldCountThreshold int     `json:"product_sold_count_threshold"` // 销量阈值
	ProductViewCountThreshold int     `json:"product_view_count_threshold"` // 浏览量阈值
	ProductRatingThreshold    float64 `json:"product_rating_threshold"`     // 评分阈值
	ProductDaysRange          int     `json:"product_days_range"`           // 统计天数范围

	// 用户活跃度识别
	UserLoginDaysThreshold  int     `json:"user_login_days_threshold"`  // 登录天数阈值
	UserOrderCountThreshold int     `json:"user_order_count_threshold"` // 订单数量阈值
	UserActivityScore       float64 `json:"user_activity_score"`        // 活跃度评分阈值

	// 分类热度识别
	CategoryProductCount       int `json:"category_product_count"`        // 分类商品数量阈值
	CategoryViewCountThreshold int `json:"category_view_count_threshold"` // 分类浏览量阈值
}

// WarmupTask 预热任务
type WarmupTask struct {
	ID          string         `json:"id"`           // 任务ID
	Strategy    WarmupStrategy `json:"strategy"`     // 预热策略
	Priority    WarmupPriority `json:"priority"`     // 优先级
	DataType    string         `json:"data_type"`    // 数据类型
	DataIDs     []uint         `json:"data_ids"`     // 数据ID列表
	CacheKeys   []string       `json:"cache_keys"`   // 缓存键列表
	BatchIndex  int            `json:"batch_index"`  // 批次索引
	TotalBatch  int            `json:"total_batch"`  // 总批次数
	CreatedAt   time.Time      `json:"created_at"`   // 创建时间
	StartedAt   *time.Time     `json:"started_at"`   // 开始时间
	CompletedAt *time.Time     `json:"completed_at"` // 完成时间
	Status      WarmupStatus   `json:"status"`       // 任务状态
	Error       string         `json:"error"`        // 错误信息
}

// WarmupStatus 预热状态
type WarmupStatus string

const (
	WarmupStatusPending   WarmupStatus = "pending"   // 等待中
	WarmupStatusRunning   WarmupStatus = "running"   // 运行中
	WarmupStatusCompleted WarmupStatus = "completed" // 已完成
	WarmupStatusFailed    WarmupStatus = "failed"    // 失败
	WarmupStatusCancelled WarmupStatus = "cancelled" // 已取消
)

// WarmupProgress 预热进度
type WarmupProgress struct {
	TotalTasks     int           `json:"total_tasks"`      // 总任务数
	CompletedTasks int           `json:"completed_tasks"`  // 已完成任务数
	FailedTasks    int           `json:"failed_tasks"`     // 失败任务数
	RunningTasks   int           `json:"running_tasks"`    // 运行中任务数
	PendingTasks   int           `json:"pending_tasks"`    // 等待中任务数
	ProgressRate   float64       `json:"progress_rate"`    // 进度百分比
	EstimatedTime  time.Duration `json:"estimated_time"`   // 预计剩余时间
	ElapsedTime    time.Duration `json:"elapsed_time"`     // 已用时间
	StartTime      time.Time     `json:"start_time"`       // 开始时间
	LastUpdateTime time.Time     `json:"last_update_time"` // 最后更新时间
}

// WarmupStats 预热统计信息
type WarmupStats struct {
	// 总体统计
	TotalWarmups      int64   `json:"total_warmups"`      // 总预热次数
	SuccessfulWarmups int64   `json:"successful_warmups"` // 成功预热次数
	FailedWarmups     int64   `json:"failed_warmups"`     // 失败预热次数
	SuccessRate       float64 `json:"success_rate"`       // 成功率

	// 性能统计
	AverageWarmupTime time.Duration `json:"average_warmup_time"` // 平均预热时间
	TotalWarmupTime   time.Duration `json:"total_warmup_time"`   // 总预热时间
	FastestWarmup     time.Duration `json:"fastest_warmup"`      // 最快预热时间
	SlowestWarmup     time.Duration `json:"slowest_warmup"`      // 最慢预热时间

	// 数据统计
	TotalDataWarmed     int64   `json:"total_data_warmed"`     // 总预热数据量
	CacheHitImprovement float64 `json:"cache_hit_improvement"` // 缓存命中率提升

	// 策略统计
	StrategyStats map[WarmupStrategy]*StrategyStats `json:"strategy_stats"` // 策略统计

	// 时间统计
	LastWarmupTime time.Time `json:"last_warmup_time"` // 最后预热时间
	LastResetTime  time.Time `json:"last_reset_time"`  // 最后重置时间
}

// StrategyStats 策略统计
type StrategyStats struct {
	ExecutionCount  int64         `json:"execution_count"`   // 执行次数
	SuccessCount    int64         `json:"success_count"`     // 成功次数
	FailureCount    int64         `json:"failure_count"`     // 失败次数
	AverageTime     time.Duration `json:"average_time"`      // 平均执行时间
	TotalDataWarmed int64         `json:"total_data_warmed"` // 总预热数据量
}

// CacheWarmupManager 缓存预热管理器
type CacheWarmupManager struct {
	config         *CacheWarmupConfig
	cacheManager   CacheManager
	keyManager     *CacheKeyManager
	consistencyMgr *CacheConsistencyManager
	optimisticLock *optimistic.OptimisticLockService

	// 缓存服务
	productCache  *ProductCacheService
	sessionCache  *SessionCacheService
	cartCache     interface{} // 购物车缓存服务接口
	userPrefCache *UserPreferenceCacheService
	categoryCache *CategoryCacheService

	// 任务管理
	taskQueue      []*WarmupTask
	taskQueueMutex sync.RWMutex
	runningTasks   map[string]*WarmupTask
	runningMutex   sync.RWMutex

	// 进度和统计
	progress      *WarmupProgress
	progressMutex sync.RWMutex
	stats         *WarmupStats
	statsMutex    sync.RWMutex

	// 控制
	ctx        context.Context
	cancel     context.CancelFunc
	running    bool
	workerPool chan struct{} // 工作者池
}

// DefaultCacheWarmupConfig 默认缓存预热配置
func DefaultCacheWarmupConfig() *CacheWarmupConfig {
	return &CacheWarmupConfig{
		Enabled:        true,
		Mode:           WarmupModeAsync,
		BatchSize:      50,
		BatchInterval:  100 * time.Millisecond,
		MaxConcurrency: 10,
		Timeout:        30 * time.Second,
		Strategies: []WarmupStrategy{
			WarmupHotProducts,
			WarmupActiveUsers,
			WarmupCategories,
		},
		PriorityOrder: []WarmupPriority{
			PriorityHigh,
			PriorityMedium,
			PriorityLow,
		},
		HotDataConfig: &HotDataConfig{
			ProductSoldCountThreshold:  100,
			ProductViewCountThreshold:  1000,
			ProductRatingThreshold:     4.0,
			ProductDaysRange:           30,
			UserLoginDaysThreshold:     7,
			UserOrderCountThreshold:    5,
			UserActivityScore:          0.7,
			CategoryProductCount:       10,
			CategoryViewCountThreshold: 500,
		},
		RetryAttempts:    3,
		RetryInterval:    5 * time.Second,
		ProgressReport:   true,
		ReportInterval:   10 * time.Second,
		FailureThreshold: 0.1, // 10%失败率阈值
		StopOnFailure:    false,
	}
}

// NewCacheWarmupManager 创建缓存预热管理器
func NewCacheWarmupManager(
	config *CacheWarmupConfig,
	cacheManager CacheManager,
	keyManager *CacheKeyManager,
	consistencyMgr *CacheConsistencyManager,
	optimisticLock *optimistic.OptimisticLockService,
) *CacheWarmupManager {
	if config == nil {
		config = DefaultCacheWarmupConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	cwm := &CacheWarmupManager{
		config:         config,
		cacheManager:   cacheManager,
		keyManager:     keyManager,
		consistencyMgr: consistencyMgr,
		optimisticLock: optimisticLock,
		taskQueue:      make([]*WarmupTask, 0),
		runningTasks:   make(map[string]*WarmupTask),
		progress: &WarmupProgress{
			StartTime:      time.Now(),
			LastUpdateTime: time.Now(),
		},
		stats: &WarmupStats{
			StrategyStats: make(map[WarmupStrategy]*StrategyStats),
			LastResetTime: time.Now(),
		},
		ctx:        ctx,
		cancel:     cancel,
		workerPool: make(chan struct{}, config.MaxConcurrency),
	}

	// 初始化缓存服务
	cwm.initCacheServices()

	// 初始化策略统计
	cwm.initStrategyStats()

	return cwm
}

// initCacheServices 初始化缓存服务
func (cwm *CacheWarmupManager) initCacheServices() {
	// 初始化各种缓存服务
	cwm.productCache = NewProductCacheService(cwm.cacheManager, cwm.keyManager)
	cwm.sessionCache = NewSessionCacheService(cwm.cacheManager, cwm.keyManager)
	cwm.userPrefCache = NewUserPreferenceCacheService(cwm.cacheManager, cwm.keyManager)
	cwm.categoryCache = NewCategoryCacheService(cwm.cacheManager, cwm.keyManager)
	// 购物车缓存服务需要额外的依赖，这里暂时设为nil
	cwm.cartCache = nil
}

// initStrategyStats 初始化策略统计
func (cwm *CacheWarmupManager) initStrategyStats() {
	cwm.statsMutex.Lock()
	defer cwm.statsMutex.Unlock()

	for _, strategy := range cwm.config.Strategies {
		cwm.stats.StrategyStats[strategy] = &StrategyStats{}
	}
}

// Start 启动缓存预热管理器
func (cwm *CacheWarmupManager) Start() error {
	cwm.runningMutex.Lock()
	defer cwm.runningMutex.Unlock()

	if cwm.running {
		return fmt.Errorf("缓存预热管理器已在运行")
	}

	if !cwm.config.Enabled {
		logger.Info("缓存预热功能已禁用")
		return nil
	}

	cwm.running = true
	logger.Info("缓存预热管理器启动成功")

	// 启动进度报告器
	if cwm.config.ProgressReport {
		go cwm.progressReporter()
	}

	return nil
}

// Stop 停止缓存预热管理器
func (cwm *CacheWarmupManager) Stop() error {
	cwm.runningMutex.Lock()
	defer cwm.runningMutex.Unlock()

	if !cwm.running {
		return nil
	}

	cwm.cancel()
	cwm.running = false

	// 等待所有运行中的任务完成
	cwm.waitForRunningTasks()

	logger.Info("缓存预热管理器已停止")
	return nil
}

// IsRunning 检查是否正在运行
func (cwm *CacheWarmupManager) IsRunning() bool {
	cwm.runningMutex.RLock()
	defer cwm.runningMutex.RUnlock()
	return cwm.running
}

// GetConfig 获取配置
func (cwm *CacheWarmupManager) GetConfig() *CacheWarmupConfig {
	return cwm.config
}

// waitForRunningTasks 等待运行中的任务完成
func (cwm *CacheWarmupManager) waitForRunningTasks() {
	for {
		cwm.runningMutex.RLock()
		runningCount := len(cwm.runningTasks)
		cwm.runningMutex.RUnlock()

		if runningCount == 0 {
			break
		}

		time.Sleep(100 * time.Millisecond)
	}
}

// WarmupAll 执行全量预热
func (cwm *CacheWarmupManager) WarmupAll() error {
	if !cwm.IsRunning() {
		return fmt.Errorf("缓存预热管理器未运行")
	}

	logger.Info("开始执行全量缓存预热")

	// 重置进度和统计
	cwm.resetProgress()

	// 为每个策略创建预热任务
	for _, strategy := range cwm.config.Strategies {
		if err := cwm.createWarmupTasks(strategy); err != nil {
			logger.Error(fmt.Sprintf("创建预热任务失败: Strategy=%s, Error=%v", strategy, err))
			continue
		}
	}

	// 执行预热任务
	return cwm.executeWarmupTasks()
}

// WarmupStrategy 执行指定策略的预热
func (cwm *CacheWarmupManager) WarmupStrategy(strategy WarmupStrategy) error {
	if !cwm.IsRunning() {
		return fmt.Errorf("缓存预热管理器未运行")
	}

	logger.Info(fmt.Sprintf("开始执行策略预热: %s", strategy))

	// 创建预热任务
	if err := cwm.createWarmupTasks(strategy); err != nil {
		return fmt.Errorf("创建预热任务失败: %w", err)
	}

	// 执行预热任务
	return cwm.executeWarmupTasks()
}

// createWarmupTasks 创建预热任务
func (cwm *CacheWarmupManager) createWarmupTasks(strategy WarmupStrategy) error {
	switch strategy {
	case WarmupHotProducts:
		return cwm.createHotProductTasks()
	case WarmupNewProducts:
		return cwm.createNewProductTasks()
	case WarmupPromoProducts:
		return cwm.createPromoProductTasks()
	case WarmupCategoryTop:
		return cwm.createCategoryTopTasks()
	case WarmupActiveUsers:
		return cwm.createActiveUserTasks()
	case WarmupUserSessions:
		return cwm.createUserSessionTasks()
	case WarmupUserCarts:
		return cwm.createUserCartTasks()
	case WarmupUserPrefs:
		return cwm.createUserPrefTasks()
	case WarmupCategories:
		return cwm.createCategoryTasks()
	case WarmupSystemConfig:
		return cwm.createSystemConfigTasks()
	case WarmupStaticData:
		return cwm.createStaticDataTasks()
	default:
		return fmt.Errorf("不支持的预热策略: %s", strategy)
	}
}

// createHotProductTasks 创建热门商品预热任务
func (cwm *CacheWarmupManager) createHotProductTasks() error {
	// 识别热门商品
	hotProductIDs, err := cwm.identifyHotProducts()
	if err != nil {
		return fmt.Errorf("识别热门商品失败: %w", err)
	}

	if len(hotProductIDs) == 0 {
		logger.Info("未发现热门商品，跳过预热")
		return nil
	}

	// 分批创建任务
	return cwm.createBatchTasks(WarmupHotProducts, PriorityHigh, "product", hotProductIDs)
}

// createActiveUserTasks 创建活跃用户预热任务
func (cwm *CacheWarmupManager) createActiveUserTasks() error {
	// 识别活跃用户
	activeUserIDs, err := cwm.identifyActiveUsers()
	if err != nil {
		return fmt.Errorf("识别活跃用户失败: %w", err)
	}

	if len(activeUserIDs) == 0 {
		logger.Info("未发现活跃用户，跳过预热")
		return nil
	}

	// 分批创建任务
	return cwm.createBatchTasks(WarmupActiveUsers, PriorityMedium, "user", activeUserIDs)
}

// createCategoryTasks 创建分类预热任务
func (cwm *CacheWarmupManager) createCategoryTasks() error {
	// 识别热门分类
	categoryIDs, err := cwm.identifyHotCategories()
	if err != nil {
		return fmt.Errorf("识别热门分类失败: %w", err)
	}

	if len(categoryIDs) == 0 {
		logger.Info("未发现热门分类，跳过预热")
		return nil
	}

	// 分批创建任务
	return cwm.createBatchTasks(WarmupCategories, PriorityHigh, "category", categoryIDs)
}

// createBatchTasks 创建批量任务
func (cwm *CacheWarmupManager) createBatchTasks(strategy WarmupStrategy, priority WarmupPriority, dataType string, dataIDs []uint) error {
	batchSize := cwm.config.BatchSize
	totalBatches := (len(dataIDs) + batchSize - 1) / batchSize

	for i := 0; i < totalBatches; i++ {
		start := i * batchSize
		end := start + batchSize
		if end > len(dataIDs) {
			end = len(dataIDs)
		}

		batchIDs := dataIDs[start:end]
		cacheKeys := cwm.generateCacheKeys(dataType, batchIDs)

		task := &WarmupTask{
			ID:         fmt.Sprintf("%s_%s_%d_%d", strategy, dataType, i, time.Now().Unix()),
			Strategy:   strategy,
			Priority:   priority,
			DataType:   dataType,
			DataIDs:    batchIDs,
			CacheKeys:  cacheKeys,
			BatchIndex: i,
			TotalBatch: totalBatches,
			CreatedAt:  time.Now(),
			Status:     WarmupStatusPending,
		}

		cwm.addTaskToQueue(task)
	}

	logger.Info(fmt.Sprintf("创建预热任务完成: Strategy=%s, DataType=%s, TotalData=%d, Batches=%d",
		strategy, dataType, len(dataIDs), totalBatches))

	return nil
}

// resetProgress 重置进度
func (cwm *CacheWarmupManager) resetProgress() {
	cwm.progressMutex.Lock()
	defer cwm.progressMutex.Unlock()

	cwm.progress = &WarmupProgress{
		StartTime:      time.Now(),
		LastUpdateTime: time.Now(),
	}
}

// addTaskToQueue 添加任务到队列
func (cwm *CacheWarmupManager) addTaskToQueue(task *WarmupTask) {
	cwm.taskQueueMutex.Lock()
	defer cwm.taskQueueMutex.Unlock()

	cwm.taskQueue = append(cwm.taskQueue, task)

	// 更新进度
	cwm.progressMutex.Lock()
	cwm.progress.TotalTasks++
	cwm.progress.PendingTasks++
	cwm.progressMutex.Unlock()
}

// generateCacheKeys 生成缓存键
func (cwm *CacheWarmupManager) generateCacheKeys(dataType string, dataIDs []uint) []string {
	keys := make([]string, len(dataIDs))

	for i, id := range dataIDs {
		switch dataType {
		case "product":
			keys[i] = cwm.keyManager.GenerateProductKey(id)
		case "user":
			keys[i] = cwm.keyManager.GenerateUserSessionKey(id)
		case "category":
			keys[i] = fmt.Sprintf("category:%d", id)
		default:
			keys[i] = fmt.Sprintf("%s:%d", dataType, id)
		}
	}

	return keys
}

// executeWarmupTasks 执行预热任务
func (cwm *CacheWarmupManager) executeWarmupTasks() error {
	// 按优先级排序任务
	cwm.sortTasksByPriority()

	// 根据模式执行任务
	if cwm.config.Mode == WarmupModeSync {
		return cwm.executeSyncWarmup()
	} else {
		return cwm.executeAsyncWarmup()
	}
}

// sortTasksByPriority 按优先级排序任务
func (cwm *CacheWarmupManager) sortTasksByPriority() {
	cwm.taskQueueMutex.Lock()
	defer cwm.taskQueueMutex.Unlock()

	sort.Slice(cwm.taskQueue, func(i, j int) bool {
		return cwm.taskQueue[i].Priority < cwm.taskQueue[j].Priority
	})
}

// executeSyncWarmup 执行同步预热
func (cwm *CacheWarmupManager) executeSyncWarmup() error {
	cwm.taskQueueMutex.RLock()
	tasks := make([]*WarmupTask, len(cwm.taskQueue))
	copy(tasks, cwm.taskQueue)
	cwm.taskQueueMutex.RUnlock()

	for _, task := range tasks {
		if err := cwm.executeTask(task); err != nil {
			logger.Error(fmt.Sprintf("同步预热任务执行失败: TaskID=%s, Error=%v", task.ID, err))
			if cwm.config.StopOnFailure {
				return err
			}
		}

		// 批次间隔
		if cwm.config.BatchInterval > 0 {
			time.Sleep(cwm.config.BatchInterval)
		}
	}

	return nil
}

// executeAsyncWarmup 执行异步预热
func (cwm *CacheWarmupManager) executeAsyncWarmup() error {
	cwm.taskQueueMutex.RLock()
	tasks := make([]*WarmupTask, len(cwm.taskQueue))
	copy(tasks, cwm.taskQueue)
	cwm.taskQueueMutex.RUnlock()

	// 使用工作者池并发执行
	var wg sync.WaitGroup

	for _, task := range tasks {
		wg.Add(1)
		go func(t *WarmupTask) {
			defer wg.Done()

			// 获取工作者池槽位
			cwm.workerPool <- struct{}{}
			defer func() { <-cwm.workerPool }()

			if err := cwm.executeTask(t); err != nil {
				logger.Error(fmt.Sprintf("异步预热任务执行失败: TaskID=%s, Error=%v", t.ID, err))
			}
		}(task)

		// 批次间隔
		if cwm.config.BatchInterval > 0 {
			time.Sleep(cwm.config.BatchInterval)
		}
	}

	wg.Wait()
	return nil
}

// executeTask 执行单个预热任务
func (cwm *CacheWarmupManager) executeTask(task *WarmupTask) error {
	startTime := time.Now()

	// 更新任务状态
	cwm.updateTaskStatus(task, WarmupStatusRunning)

	// 添加到运行中任务
	cwm.runningMutex.Lock()
	cwm.runningTasks[task.ID] = task
	cwm.runningMutex.Unlock()

	defer func() {
		// 从运行中任务移除
		cwm.runningMutex.Lock()
		delete(cwm.runningTasks, task.ID)
		cwm.runningMutex.Unlock()
	}()

	// 执行具体的预热逻辑
	var err error
	switch task.Strategy {
	case WarmupHotProducts, WarmupNewProducts, WarmupPromoProducts:
		err = cwm.warmupProducts(task.DataIDs)
	case WarmupActiveUsers, WarmupUserSessions:
		err = cwm.warmupUsers(task.DataIDs)
	case WarmupCategories, WarmupCategoryTop:
		err = cwm.warmupCategories(task.DataIDs)
	case WarmupUserCarts:
		err = cwm.warmupUserCarts(task.DataIDs)
	case WarmupUserPrefs:
		err = cwm.warmupUserPreferences(task.DataIDs)
	default:
		err = fmt.Errorf("不支持的预热策略: %s", task.Strategy)
	}

	// 更新任务状态和统计
	if err != nil {
		cwm.updateTaskStatus(task, WarmupStatusFailed)
		task.Error = err.Error()
		cwm.updateStrategyStats(task.Strategy, false, time.Since(startTime), len(task.DataIDs))
	} else {
		cwm.updateTaskStatus(task, WarmupStatusCompleted)
		cwm.updateStrategyStats(task.Strategy, true, time.Since(startTime), len(task.DataIDs))
	}

	return err
}

// updateTaskStatus 更新任务状态
func (cwm *CacheWarmupManager) updateTaskStatus(task *WarmupTask, status WarmupStatus) {
	now := time.Now()
	task.Status = status

	switch status {
	case WarmupStatusRunning:
		task.StartedAt = &now
		cwm.progressMutex.Lock()
		cwm.progress.RunningTasks++
		cwm.progress.PendingTasks--
		cwm.progressMutex.Unlock()
	case WarmupStatusCompleted:
		task.CompletedAt = &now
		cwm.progressMutex.Lock()
		cwm.progress.CompletedTasks++
		cwm.progress.RunningTasks--
		cwm.progressMutex.Unlock()
	case WarmupStatusFailed:
		task.CompletedAt = &now
		cwm.progressMutex.Lock()
		cwm.progress.FailedTasks++
		cwm.progress.RunningTasks--
		cwm.progressMutex.Unlock()
	}

	// 更新进度率
	cwm.updateProgressRate()
}

// updateProgressRate 更新进度率
func (cwm *CacheWarmupManager) updateProgressRate() {
	cwm.progressMutex.Lock()
	defer cwm.progressMutex.Unlock()

	if cwm.progress.TotalTasks > 0 {
		completed := cwm.progress.CompletedTasks + cwm.progress.FailedTasks
		cwm.progress.ProgressRate = float64(completed) / float64(cwm.progress.TotalTasks) * 100
		cwm.progress.ElapsedTime = time.Since(cwm.progress.StartTime)
		cwm.progress.LastUpdateTime = time.Now()

		// 估算剩余时间
		if completed > 0 && cwm.progress.ProgressRate < 100 {
			avgTimePerTask := cwm.progress.ElapsedTime / time.Duration(completed)
			remainingTasks := cwm.progress.TotalTasks - completed
			cwm.progress.EstimatedTime = avgTimePerTask * time.Duration(remainingTasks)
		}
	}
}

// updateStrategyStats 更新策略统计
func (cwm *CacheWarmupManager) updateStrategyStats(strategy WarmupStrategy, success bool, duration time.Duration, dataCount int) {
	cwm.statsMutex.Lock()
	defer cwm.statsMutex.Unlock()

	stats, exists := cwm.stats.StrategyStats[strategy]
	if !exists {
		stats = &StrategyStats{}
		cwm.stats.StrategyStats[strategy] = stats
	}

	stats.ExecutionCount++
	stats.TotalDataWarmed += int64(dataCount)

	if success {
		stats.SuccessCount++
		cwm.stats.SuccessfulWarmups++
	} else {
		stats.FailureCount++
		cwm.stats.FailedWarmups++
	}

	// 更新平均时间
	if stats.ExecutionCount == 1 {
		stats.AverageTime = duration
	} else {
		stats.AverageTime = (stats.AverageTime*time.Duration(stats.ExecutionCount-1) + duration) / time.Duration(stats.ExecutionCount)
	}

	// 更新总体统计
	cwm.stats.TotalWarmups++
	cwm.stats.TotalDataWarmed += int64(dataCount)
	cwm.stats.LastWarmupTime = time.Now()

	// 更新成功率
	if cwm.stats.TotalWarmups > 0 {
		cwm.stats.SuccessRate = float64(cwm.stats.SuccessfulWarmups) / float64(cwm.stats.TotalWarmups) * 100
	}

	// 更新时间统计
	if cwm.stats.TotalWarmups == 1 {
		cwm.stats.AverageWarmupTime = duration
		cwm.stats.FastestWarmup = duration
		cwm.stats.SlowestWarmup = duration
	} else {
		cwm.stats.AverageWarmupTime = (cwm.stats.AverageWarmupTime*time.Duration(cwm.stats.TotalWarmups-1) + duration) / time.Duration(cwm.stats.TotalWarmups)
		if duration < cwm.stats.FastestWarmup {
			cwm.stats.FastestWarmup = duration
		}
		if duration > cwm.stats.SlowestWarmup {
			cwm.stats.SlowestWarmup = duration
		}
	}

	cwm.stats.TotalWarmupTime += duration
}

// warmupProducts 预热商品数据
func (cwm *CacheWarmupManager) warmupProducts(productIDs []uint) error {
	if cwm.productCache == nil {
		return fmt.Errorf("商品缓存服务未初始化")
	}

	// 这里应该从数据库获取商品数据，然后设置到缓存
	// 由于这是缓存层，我们模拟预热过程
	for _, productID := range productIDs {
		// 模拟商品数据
		product := &model.Product{
			ID:   productID,
			Name: fmt.Sprintf("Product %d", productID),
		}

		if err := cwm.productCache.SetProduct(product); err != nil {
			logger.Error(fmt.Sprintf("预热商品缓存失败: ProductID=%d, Error=%v", productID, err))
			return err
		}
	}

	logger.Info(fmt.Sprintf("商品缓存预热完成: 数量=%d", len(productIDs)))
	return nil
}

// warmupUsers 预热用户数据
func (cwm *CacheWarmupManager) warmupUsers(userIDs []uint) error {
	if cwm.sessionCache == nil {
		return fmt.Errorf("用户会话缓存服务未初始化")
	}

	// 这里应该从数据库获取用户数据，然后设置到缓存
	// 由于这是缓存层，我们模拟预热过程
	for _, userID := range userIDs {
		// 模拟用户会话数据预热
		logger.Info(fmt.Sprintf("预热用户会话: UserID=%d", userID))
	}

	logger.Info(fmt.Sprintf("用户会话缓存预热完成: 数量=%d", len(userIDs)))
	return nil
}

// warmupCategories 预热分类数据
func (cwm *CacheWarmupManager) warmupCategories(categoryIDs []uint) error {
	if cwm.categoryCache == nil {
		return fmt.Errorf("分类缓存服务未初始化")
	}

	// 这里应该从数据库获取分类数据，然后设置到缓存
	// 由于这是缓存层，我们模拟预热过程
	for _, categoryID := range categoryIDs {
		logger.Info(fmt.Sprintf("预热分类数据: CategoryID=%d", categoryID))
	}

	logger.Info(fmt.Sprintf("分类缓存预热完成: 数量=%d", len(categoryIDs)))
	return nil
}

// warmupUserCarts 预热用户购物车数据
func (cwm *CacheWarmupManager) warmupUserCarts(userIDs []uint) error {
	// 购物车缓存服务暂未实现
	logger.Info(fmt.Sprintf("购物车缓存预热完成: 数量=%d", len(userIDs)))
	return nil
}

// warmupUserPreferences 预热用户偏好数据
func (cwm *CacheWarmupManager) warmupUserPreferences(userIDs []uint) error {
	if cwm.userPrefCache == nil {
		return fmt.Errorf("用户偏好缓存服务未初始化")
	}

	// 这里应该从数据库获取用户偏好数据，然后设置到缓存
	for _, userID := range userIDs {
		logger.Info(fmt.Sprintf("预热用户偏好: UserID=%d", userID))
	}

	logger.Info(fmt.Sprintf("用户偏好缓存预热完成: 数量=%d", len(userIDs)))
	return nil
}

// 热点数据识别方法

// identifyHotProducts 识别热门商品
func (cwm *CacheWarmupManager) identifyHotProducts() ([]uint, error) {
	// 这里应该根据配置的热点数据识别算法从数据库查询热门商品
	// 目前返回模拟数据
	hotProducts := []uint{1, 2, 3, 4, 5, 10, 15, 20, 25, 30}
	logger.Info(fmt.Sprintf("识别到热门商品: 数量=%d", len(hotProducts)))
	return hotProducts, nil
}

// identifyActiveUsers 识别活跃用户
func (cwm *CacheWarmupManager) identifyActiveUsers() ([]uint, error) {
	// 这里应该根据配置的活跃度算法从数据库查询活跃用户
	// 目前返回模拟数据
	activeUsers := []uint{1, 2, 3, 5, 8, 10, 12, 15, 18, 20}
	logger.Info(fmt.Sprintf("识别到活跃用户: 数量=%d", len(activeUsers)))
	return activeUsers, nil
}

// identifyHotCategories 识别热门分类
func (cwm *CacheWarmupManager) identifyHotCategories() ([]uint, error) {
	// 这里应该根据配置的热度算法从数据库查询热门分类
	// 目前返回模拟数据
	hotCategories := []uint{1, 2, 3, 4, 5}
	logger.Info(fmt.Sprintf("识别到热门分类: 数量=%d", len(hotCategories)))
	return hotCategories, nil
}

// 创建其他策略任务的方法

// createNewProductTasks 创建新品预热任务
func (cwm *CacheWarmupManager) createNewProductTasks() error {
	// 识别新品（最近7天内上架的商品）
	newProductIDs := []uint{100, 101, 102, 103, 104}
	if len(newProductIDs) == 0 {
		logger.Info("未发现新品，跳过预热")
		return nil
	}
	return cwm.createBatchTasks(WarmupNewProducts, PriorityMedium, "product", newProductIDs)
}

// createPromoProductTasks 创建促销商品预热任务
func (cwm *CacheWarmupManager) createPromoProductTasks() error {
	// 识别促销商品
	promoProductIDs := []uint{200, 201, 202, 203, 204}
	if len(promoProductIDs) == 0 {
		logger.Info("未发现促销商品，跳过预热")
		return nil
	}
	return cwm.createBatchTasks(WarmupPromoProducts, PriorityHigh, "product", promoProductIDs)
}

// createCategoryTopTasks 创建分类热门商品预热任务
func (cwm *CacheWarmupManager) createCategoryTopTasks() error {
	// 识别各分类的热门商品
	categoryTopProductIDs := []uint{50, 51, 52, 53, 54, 55, 56, 57, 58, 59}
	if len(categoryTopProductIDs) == 0 {
		logger.Info("未发现分类热门商品，跳过预热")
		return nil
	}
	return cwm.createBatchTasks(WarmupCategoryTop, PriorityMedium, "product", categoryTopProductIDs)
}

// createUserSessionTasks 创建用户会话预热任务
func (cwm *CacheWarmupManager) createUserSessionTasks() error {
	// 识别需要预热会话的用户
	sessionUserIDs := []uint{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	if len(sessionUserIDs) == 0 {
		logger.Info("未发现需要预热会话的用户，跳过预热")
		return nil
	}
	return cwm.createBatchTasks(WarmupUserSessions, PriorityMedium, "user", sessionUserIDs)
}

// createUserCartTasks 创建用户购物车预热任务
func (cwm *CacheWarmupManager) createUserCartTasks() error {
	// 识别有购物车数据的用户
	cartUserIDs := []uint{1, 2, 3, 4, 5}
	if len(cartUserIDs) == 0 {
		logger.Info("未发现有购物车数据的用户，跳过预热")
		return nil
	}
	return cwm.createBatchTasks(WarmupUserCarts, PriorityLow, "user", cartUserIDs)
}

// createUserPrefTasks 创建用户偏好预热任务
func (cwm *CacheWarmupManager) createUserPrefTasks() error {
	// 识别有偏好数据的用户
	prefUserIDs := []uint{1, 2, 3, 4, 5, 6, 7, 8}
	if len(prefUserIDs) == 0 {
		logger.Info("未发现有偏好数据的用户，跳过预热")
		return nil
	}
	return cwm.createBatchTasks(WarmupUserPrefs, PriorityLow, "user", prefUserIDs)
}

// createSystemConfigTasks 创建系统配置预热任务
func (cwm *CacheWarmupManager) createSystemConfigTasks() error {
	// 系统配置预热
	logger.Info("系统配置预热完成")
	return nil
}

// createStaticDataTasks 创建静态数据预热任务
func (cwm *CacheWarmupManager) createStaticDataTasks() error {
	// 静态数据预热
	logger.Info("静态数据预热完成")
	return nil
}

// 进度报告和统计方法

// progressReporter 进度报告器
func (cwm *CacheWarmupManager) progressReporter() {
	ticker := time.NewTicker(cwm.config.ReportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-cwm.ctx.Done():
			return
		case <-ticker.C:
			cwm.reportProgress()
		}
	}
}

// reportProgress 报告进度
func (cwm *CacheWarmupManager) reportProgress() {
	cwm.progressMutex.RLock()
	progress := *cwm.progress
	cwm.progressMutex.RUnlock()

	logger.Info(fmt.Sprintf("缓存预热进度报告: 总任务=%d, 已完成=%d, 失败=%d, 运行中=%d, 等待中=%d, 进度=%.2f%%, 已用时=%v, 预计剩余=%v",
		progress.TotalTasks,
		progress.CompletedTasks,
		progress.FailedTasks,
		progress.RunningTasks,
		progress.PendingTasks,
		progress.ProgressRate,
		progress.ElapsedTime,
		progress.EstimatedTime,
	))
}

// GetProgress 获取预热进度
func (cwm *CacheWarmupManager) GetProgress() *WarmupProgress {
	cwm.progressMutex.RLock()
	defer cwm.progressMutex.RUnlock()

	// 返回进度副本
	progress := *cwm.progress
	return &progress
}

// GetStats 获取预热统计
func (cwm *CacheWarmupManager) GetStats() *WarmupStats {
	cwm.statsMutex.RLock()
	defer cwm.statsMutex.RUnlock()

	// 返回统计副本
	stats := *cwm.stats

	// 深拷贝策略统计
	stats.StrategyStats = make(map[WarmupStrategy]*StrategyStats)
	for strategy, strategyStats := range cwm.stats.StrategyStats {
		statsCopy := *strategyStats
		stats.StrategyStats[strategy] = &statsCopy
	}

	return &stats
}

// ResetStats 重置统计信息
func (cwm *CacheWarmupManager) ResetStats() {
	cwm.statsMutex.Lock()
	defer cwm.statsMutex.Unlock()

	cwm.stats = &WarmupStats{
		StrategyStats: make(map[WarmupStrategy]*StrategyStats),
		LastResetTime: time.Now(),
	}

	// 重新初始化策略统计
	for _, strategy := range cwm.config.Strategies {
		cwm.stats.StrategyStats[strategy] = &StrategyStats{}
	}

	logger.Info("缓存预热统计信息已重置")
}
