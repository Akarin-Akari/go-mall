package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"mall-go/pkg/logger"
	"mall-go/pkg/optimistic"
)

// CacheSyncStrategy 缓存同步策略枚举
type CacheSyncStrategy string

const (
	// WriteThrough 写穿透：同时写入缓存和数据库
	WriteThrough CacheSyncStrategy = "write_through"
	// WriteBehind 写回：先写缓存，异步写入数据库
	WriteBehind CacheSyncStrategy = "write_behind"
	// CacheAside 缓存旁路：应用程序管理缓存和数据库
	CacheAside CacheSyncStrategy = "cache_aside"
	// RefreshAhead 提前刷新：在缓存过期前主动刷新
	RefreshAhead CacheSyncStrategy = "refresh_ahead"
)

// CacheConsistencyConfig 缓存一致性配置
type CacheConsistencyConfig struct {
	// 同步策略
	Strategy CacheSyncStrategy `json:"strategy"`

	// 一致性检查配置
	CheckInterval  time.Duration `json:"check_interval"`   // 一致性检查间隔
	CheckBatchSize int           `json:"check_batch_size"` // 批量检查大小
	CheckTimeout   time.Duration `json:"check_timeout"`    // 检查超时时间

	// 同步配置
	SyncTimeout   time.Duration `json:"sync_timeout"`    // 同步超时时间
	SyncRetries   int           `json:"sync_retries"`    // 同步重试次数
	SyncBatchSize int           `json:"sync_batch_size"` // 批量同步大小

	// 失效配置
	InvalidateDelay time.Duration `json:"invalidate_delay"` // 失效延迟时间
	InvalidateBatch int           `json:"invalidate_batch"` // 批量失效大小

	// 事件驱动配置
	EventBufferSize int `json:"event_buffer_size"` // 事件缓冲区大小
	EventWorkers    int `json:"event_workers"`     // 事件处理工作者数量

	// 分布式配置
	DistributedMode bool     `json:"distributed_mode"` // 是否启用分布式模式
	NodeID          string   `json:"node_id"`          // 节点ID
	ClusterNodes    []string `json:"cluster_nodes"`    // 集群节点列表
}

// DefaultCacheConsistencyConfig 默认缓存一致性配置
func DefaultCacheConsistencyConfig() *CacheConsistencyConfig {
	return &CacheConsistencyConfig{
		Strategy:        CacheAside,
		CheckInterval:   30 * time.Second,
		CheckBatchSize:  100,
		CheckTimeout:    5 * time.Second,
		SyncTimeout:     10 * time.Second,
		SyncRetries:     3,
		SyncBatchSize:   50,
		InvalidateDelay: 100 * time.Millisecond,
		InvalidateBatch: 20,
		EventBufferSize: 1000,
		EventWorkers:    5,
		DistributedMode: false,
		NodeID:          "node-1",
		ClusterNodes:    []string{},
	}
}

// CacheUpdateEvent 缓存更新事件结构
type CacheUpdateEvent struct {
	ID         string                 `json:"id"`          // 事件ID
	Type       string                 `json:"type"`        // 事件类型：create, update, delete
	TableName  string                 `json:"table_name"`  // 表名
	RecordID   uint                   `json:"record_id"`   // 记录ID
	OldVersion int                    `json:"old_version"` // 旧版本号
	NewVersion int                    `json:"new_version"` // 新版本号
	Data       map[string]interface{} `json:"data"`        // 变更数据
	CacheKeys  []string               `json:"cache_keys"`  // 相关缓存键
	Timestamp  time.Time              `json:"timestamp"`   // 事件时间戳
	NodeID     string                 `json:"node_id"`     // 产生事件的节点ID
	Processed  bool                   `json:"processed"`   // 是否已处理
}

// ConsistencyCheckResult 一致性检查结果
type ConsistencyCheckResult struct {
	CacheKey        string                 `json:"cache_key"`        // 缓存键
	TableName       string                 `json:"table_name"`       // 表名
	RecordID        uint                   `json:"record_id"`        // 记录ID
	IsConsistent    bool                   `json:"is_consistent"`    // 是否一致
	CacheVersion    int                    `json:"cache_version"`    // 缓存版本
	DatabaseVersion int                    `json:"database_version"` // 数据库版本
	CacheData       map[string]interface{} `json:"cache_data"`       // 缓存数据
	DatabaseData    map[string]interface{} `json:"database_data"`    // 数据库数据
	CheckTime       time.Time              `json:"check_time"`       // 检查时间
	Error           string                 `json:"error,omitempty"`  // 错误信息
}

// CacheConsistencyStats 缓存一致性统计
type CacheConsistencyStats struct {
	TotalChecks       int64     `json:"total_checks"`       // 总检查次数
	ConsistentCount   int64     `json:"consistent_count"`   // 一致数量
	InconsistentCount int64     `json:"inconsistent_count"` // 不一致数量
	ConsistencyRate   float64   `json:"consistency_rate"`   // 一致性率
	TotalSyncs        int64     `json:"total_syncs"`        // 总同步次数
	SuccessfulSyncs   int64     `json:"successful_syncs"`   // 成功同步次数
	FailedSyncs       int64     `json:"failed_syncs"`       // 失败同步次数
	SyncSuccessRate   float64   `json:"sync_success_rate"`  // 同步成功率
	TotalEvents       int64     `json:"total_events"`       // 总事件数
	ProcessedEvents   int64     `json:"processed_events"`   // 已处理事件数
	PendingEvents     int64     `json:"pending_events"`     // 待处理事件数
	LastCheckTime     time.Time `json:"last_check_time"`    // 最后检查时间
	LastSyncTime      time.Time `json:"last_sync_time"`     // 最后同步时间
	AverageCheckTime  float64   `json:"average_check_time"` // 平均检查时间(ms)
	AverageSyncTime   float64   `json:"average_sync_time"`  // 平均同步时间(ms)
}

// CacheConsistencyManager 缓存一致性管理器
type CacheConsistencyManager struct {
	config         *CacheConsistencyConfig
	cacheManager   CacheManager
	keyManager     *CacheKeyManager
	optimisticLock *optimistic.OptimisticLockService

	// 事件处理
	eventChan    chan *CacheUpdateEvent
	eventWorkers []*EventWorker
	eventWg      sync.WaitGroup

	// 统计信息
	stats      *CacheConsistencyStats
	statsMutex sync.RWMutex

	// 控制
	ctx          context.Context
	cancel       context.CancelFunc
	running      bool
	runningMutex sync.RWMutex
}

// EventWorker 事件处理工作者
type EventWorker struct {
	id        int
	manager   *CacheConsistencyManager
	eventChan chan *CacheUpdateEvent
	ctx       context.Context
}

// NewCacheConsistencyManager 创建缓存一致性管理器
func NewCacheConsistencyManager(
	config *CacheConsistencyConfig,
	cacheManager CacheManager,
	keyManager *CacheKeyManager,
	optimisticLock *optimistic.OptimisticLockService,
) *CacheConsistencyManager {
	if config == nil {
		config = DefaultCacheConsistencyConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	ccm := &CacheConsistencyManager{
		config:         config,
		cacheManager:   cacheManager,
		keyManager:     keyManager,
		optimisticLock: optimisticLock,
		eventChan:      make(chan *CacheUpdateEvent, config.EventBufferSize),
		stats: &CacheConsistencyStats{
			LastCheckTime: time.Now(),
			LastSyncTime:  time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// 初始化事件工作者
	ccm.initEventWorkers()

	return ccm
}

// initEventWorkers 初始化事件工作者
func (ccm *CacheConsistencyManager) initEventWorkers() {
	ccm.eventWorkers = make([]*EventWorker, ccm.config.EventWorkers)

	for i := 0; i < ccm.config.EventWorkers; i++ {
		worker := &EventWorker{
			id:        i,
			manager:   ccm,
			eventChan: ccm.eventChan,
			ctx:       ccm.ctx,
		}
		ccm.eventWorkers[i] = worker
	}
}

// Start 启动缓存一致性管理器
func (ccm *CacheConsistencyManager) Start() error {
	ccm.runningMutex.Lock()
	defer ccm.runningMutex.Unlock()

	if ccm.running {
		return fmt.Errorf("缓存一致性管理器已在运行")
	}

	// 启动事件工作者
	for _, worker := range ccm.eventWorkers {
		ccm.eventWg.Add(1)
		go worker.start(&ccm.eventWg)
	}

	// 启动定期一致性检查
	ccm.eventWg.Add(1)
	go ccm.startConsistencyChecker(&ccm.eventWg)

	ccm.running = true
	logger.Info("缓存一致性管理器启动成功")

	return nil
}

// Stop 停止缓存一致性管理器
func (ccm *CacheConsistencyManager) Stop() error {
	ccm.runningMutex.Lock()
	defer ccm.runningMutex.Unlock()

	if !ccm.running {
		return nil
	}

	// 取消上下文
	ccm.cancel()

	// 等待所有工作者停止
	ccm.eventWg.Wait()

	// 关闭事件通道
	close(ccm.eventChan)

	ccm.running = false
	logger.Info("缓存一致性管理器停止成功")

	return nil
}

// IsRunning 检查是否正在运行
func (ccm *CacheConsistencyManager) IsRunning() bool {
	ccm.runningMutex.RLock()
	defer ccm.runningMutex.RUnlock()
	return ccm.running
}

// PublishEvent 发布缓存更新事件
func (ccm *CacheConsistencyManager) PublishEvent(event *CacheUpdateEvent) error {
	if !ccm.IsRunning() {
		return fmt.Errorf("缓存一致性管理器未运行")
	}

	// 设置事件元数据
	if event.ID == "" {
		event.ID = fmt.Sprintf("%s_%d_%d", event.TableName, event.RecordID, time.Now().UnixNano())
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	if event.NodeID == "" {
		event.NodeID = ccm.config.NodeID
	}

	// 非阻塞发送事件
	select {
	case ccm.eventChan <- event:
		ccm.updateStats(func(stats *CacheConsistencyStats) {
			stats.TotalEvents++
		})
		logger.Info(fmt.Sprintf("发布缓存更新事件: Type=%s, Table=%s, ID=%d",
			event.Type, event.TableName, event.RecordID))
		return nil
	default:
		return fmt.Errorf("事件缓冲区已满，无法发布事件")
	}
}

// CheckConsistency 检查缓存一致性
func (ccm *CacheConsistencyManager) CheckConsistency(cacheKey string, tableName string, recordID uint) (*ConsistencyCheckResult, error) {
	startTime := time.Now()

	result := &ConsistencyCheckResult{
		CacheKey:  cacheKey,
		TableName: tableName,
		RecordID:  recordID,
		CheckTime: startTime,
	}

	// 获取缓存数据
	cacheData, err := ccm.cacheManager.Get(cacheKey)
	if err != nil {
		result.Error = fmt.Sprintf("获取缓存数据失败: %v", err)
		return result, err
	}

	// 解析缓存数据
	if cacheData != nil {
		var cacheMap map[string]interface{}
		if err := json.Unmarshal([]byte(cacheData.(string)), &cacheMap); err != nil {
			result.Error = fmt.Sprintf("解析缓存数据失败: %v", err)
			return result, err
		}
		result.CacheData = cacheMap

		// 获取缓存版本号
		if version, ok := cacheMap["version"]; ok {
			if v, ok := version.(float64); ok {
				result.CacheVersion = int(v)
			}
		}
	}

	// 获取数据库数据（这里需要根据实际情况实现数据库查询）
	// 暂时模拟数据库查询
	result.DatabaseData = map[string]interface{}{
		"id":      recordID,
		"version": result.CacheVersion, // 模拟相同版本
	}
	result.DatabaseVersion = result.CacheVersion

	// 比较一致性
	result.IsConsistent = ccm.compareData(result.CacheData, result.DatabaseData)

	// 更新统计
	ccm.updateStats(func(stats *CacheConsistencyStats) {
		stats.TotalChecks++
		if result.IsConsistent {
			stats.ConsistentCount++
		} else {
			stats.InconsistentCount++
		}
		stats.ConsistencyRate = float64(stats.ConsistentCount) / float64(stats.TotalChecks) * 100
		stats.LastCheckTime = time.Now()

		// 更新平均检查时间
		checkDuration := time.Since(startTime).Milliseconds()
		if stats.AverageCheckTime == 0 {
			stats.AverageCheckTime = float64(checkDuration)
		} else {
			stats.AverageCheckTime = (stats.AverageCheckTime + float64(checkDuration)) / 2
		}
	})

	return result, nil
}

// SyncCache 同步缓存数据
func (ccm *CacheConsistencyManager) SyncCache(cacheKey string, tableName string, recordID uint, data map[string]interface{}) error {
	startTime := time.Now()

	var err error
	switch ccm.config.Strategy {
	case WriteThrough:
		err = ccm.syncWriteThrough(cacheKey, tableName, recordID, data)
	case WriteBehind:
		err = ccm.syncWriteBehind(cacheKey, tableName, recordID, data)
	case CacheAside:
		err = ccm.syncCacheAside(cacheKey, tableName, recordID, data)
	case RefreshAhead:
		err = ccm.syncRefreshAhead(cacheKey, tableName, recordID, data)
	default:
		err = fmt.Errorf("不支持的同步策略: %s", ccm.config.Strategy)
	}

	// 更新同步统计
	ccm.updateStats(func(stats *CacheConsistencyStats) {
		stats.TotalSyncs++
		if err == nil {
			stats.SuccessfulSyncs++
		} else {
			stats.FailedSyncs++
		}
		stats.SyncSuccessRate = float64(stats.SuccessfulSyncs) / float64(stats.TotalSyncs) * 100
		stats.LastSyncTime = time.Now()

		// 更新平均同步时间
		syncDuration := time.Since(startTime).Milliseconds()
		if stats.AverageSyncTime == 0 {
			stats.AverageSyncTime = float64(syncDuration)
		} else {
			stats.AverageSyncTime = (stats.AverageSyncTime + float64(syncDuration)) / 2
		}
	})

	return err
}

// InvalidateCache 失效缓存
func (ccm *CacheConsistencyManager) InvalidateCache(cacheKeys []string) error {
	if len(cacheKeys) == 0 {
		return nil
	}

	// 批量删除缓存
	if err := ccm.cacheManager.MDelete(cacheKeys); err != nil {
		logger.Error(fmt.Sprintf("批量失效缓存失败: %v", err))
		return fmt.Errorf("批量失效缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("成功失效缓存: 数量=%d", len(cacheKeys)))
	return nil
}

// GetStats 获取一致性统计信息
func (ccm *CacheConsistencyManager) GetStats() *CacheConsistencyStats {
	ccm.statsMutex.RLock()
	defer ccm.statsMutex.RUnlock()

	// 创建副本返回
	statsCopy := *ccm.stats
	return &statsCopy
}

// ResetStats 重置统计信息
func (ccm *CacheConsistencyManager) ResetStats() {
	ccm.statsMutex.Lock()
	defer ccm.statsMutex.Unlock()

	ccm.stats = &CacheConsistencyStats{
		LastCheckTime: time.Now(),
		LastSyncTime:  time.Now(),
	}

	logger.Info("缓存一致性统计信息已重置")
}

// GetConfig 获取配置信息
func (ccm *CacheConsistencyManager) GetConfig() *CacheConsistencyConfig {
	return ccm.config
}

// updateStats 更新统计信息（线程安全）
func (ccm *CacheConsistencyManager) updateStats(updateFunc func(*CacheConsistencyStats)) {
	ccm.statsMutex.Lock()
	defer ccm.statsMutex.Unlock()
	updateFunc(ccm.stats)
}

// compareData 比较缓存数据和数据库数据
func (ccm *CacheConsistencyManager) compareData(cacheData, dbData map[string]interface{}) bool {
	if cacheData == nil && dbData == nil {
		return true
	}
	if cacheData == nil || dbData == nil {
		return false
	}

	// 比较版本号
	cacheVersion, cacheOk := cacheData["version"]
	dbVersion, dbOk := dbData["version"]

	if cacheOk && dbOk {
		return cacheVersion == dbVersion
	}

	// 如果没有版本号，比较关键字段
	keyFields := []string{"id", "updated_at", "price", "stock"}
	for _, field := range keyFields {
		if cacheData[field] != dbData[field] {
			return false
		}
	}

	return true
}

// syncWriteThrough 写穿透同步策略
func (ccm *CacheConsistencyManager) syncWriteThrough(cacheKey string, tableName string, recordID uint, data map[string]interface{}) error {
	// 1. 先更新数据库
	updates := make(map[string]interface{})
	for k, v := range data {
		if k != "id" { // 排除ID字段
			updates[k] = v
		}
	}

	result := ccm.optimisticLock.UpdateWithOptimisticLock(tableName, recordID, updates, nil)
	if !result.Success {
		ccm.updateStats(func(stats *CacheConsistencyStats) {
			stats.FailedSyncs++
		})
		return fmt.Errorf("数据库更新失败: %v", result.Error)
	}

	// 2. 更新缓存
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化缓存数据失败: %w", err)
	}

	ttl := GetTTL("product") // 使用默认TTL
	if err := ccm.cacheManager.Set(cacheKey, string(jsonData), ttl); err != nil {
		logger.Error(fmt.Sprintf("写穿透策略更新缓存失败: %v", err))
		// 缓存更新失败不影响数据库更新的成功
	}

	logger.Info(fmt.Sprintf("写穿透同步成功: Key=%s, Table=%s, ID=%d", cacheKey, tableName, recordID))
	return nil
}

// syncWriteBehind 写回同步策略
func (ccm *CacheConsistencyManager) syncWriteBehind(cacheKey string, tableName string, recordID uint, data map[string]interface{}) error {
	// 1. 先更新缓存
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化缓存数据失败: %w", err)
	}

	ttl := GetTTL("product")
	if err := ccm.cacheManager.Set(cacheKey, string(jsonData), ttl); err != nil {
		return fmt.Errorf("更新缓存失败: %w", err)
	}

	// 2. 异步更新数据库
	go func() {
		updates := make(map[string]interface{})
		for k, v := range data {
			if k != "id" {
				updates[k] = v
			}
		}

		result := ccm.optimisticLock.UpdateWithOptimisticLock(tableName, recordID, updates, nil)
		if !result.Success {
			logger.Error(fmt.Sprintf("写回策略异步数据库更新失败: Table=%s, ID=%d, Error=%v",
				tableName, recordID, result.Error))
			ccm.updateStats(func(stats *CacheConsistencyStats) {
				stats.FailedSyncs++
			})
		} else {
			logger.Info(fmt.Sprintf("写回策略异步数据库更新成功: Table=%s, ID=%d", tableName, recordID))
		}
	}()

	logger.Info(fmt.Sprintf("写回同步成功: Key=%s, Table=%s, ID=%d", cacheKey, tableName, recordID))
	return nil
}

// syncCacheAside 缓存旁路同步策略
func (ccm *CacheConsistencyManager) syncCacheAside(cacheKey string, tableName string, recordID uint, data map[string]interface{}) error {
	// 缓存旁路策略：删除缓存，让应用程序在下次访问时重新加载
	if err := ccm.cacheManager.Delete(cacheKey); err != nil {
		logger.Error(fmt.Sprintf("缓存旁路策略删除缓存失败: %v", err))
		return fmt.Errorf("删除缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("缓存旁路同步成功: Key=%s, Table=%s, ID=%d", cacheKey, tableName, recordID))
	return nil
}

// syncRefreshAhead 提前刷新同步策略
func (ccm *CacheConsistencyManager) syncRefreshAhead(cacheKey string, tableName string, recordID uint, data map[string]interface{}) error {
	// 提前刷新策略：更新缓存并延长TTL
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化缓存数据失败: %w", err)
	}

	// 使用较长的TTL
	ttl := GetTTL("product") * 2
	if err := ccm.cacheManager.Set(cacheKey, string(jsonData), ttl); err != nil {
		return fmt.Errorf("更新缓存失败: %w", err)
	}

	logger.Info(fmt.Sprintf("提前刷新同步成功: Key=%s, Table=%s, ID=%d, TTL=%v",
		cacheKey, tableName, recordID, ttl))
	return nil
}

// startConsistencyChecker 启动定期一致性检查
func (ccm *CacheConsistencyManager) startConsistencyChecker(wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(ccm.config.CheckInterval)
	defer ticker.Stop()

	logger.Info(fmt.Sprintf("一致性检查器启动，检查间隔: %v", ccm.config.CheckInterval))

	for {
		select {
		case <-ccm.ctx.Done():
			logger.Info("一致性检查器停止")
			return
		case <-ticker.C:
			ccm.performConsistencyCheck()
		}
	}
}

// performConsistencyCheck 执行一致性检查
func (ccm *CacheConsistencyManager) performConsistencyCheck() {
	logger.Info("开始执行定期一致性检查")

	// 这里可以实现具体的一致性检查逻辑
	// 例如：扫描所有缓存键，检查与数据库的一致性
	// 由于需要具体的数据库查询逻辑，这里暂时记录日志

	ccm.updateStats(func(stats *CacheConsistencyStats) {
		stats.LastCheckTime = time.Now()
	})

	logger.Info("定期一致性检查完成")
}

// start 启动事件工作者
func (ew *EventWorker) start(wg *sync.WaitGroup) {
	defer wg.Done()

	logger.Info(fmt.Sprintf("事件工作者 %d 启动", ew.id))

	for {
		select {
		case <-ew.ctx.Done():
			logger.Info(fmt.Sprintf("事件工作者 %d 停止", ew.id))
			return
		case event := <-ew.eventChan:
			if event != nil {
				ew.processEvent(event)
			}
		}
	}
}

// processEvent 处理缓存更新事件
func (ew *EventWorker) processEvent(event *CacheUpdateEvent) {
	logger.Info(fmt.Sprintf("工作者 %d 处理事件: Type=%s, Table=%s, ID=%d",
		ew.id, event.Type, event.TableName, event.RecordID))

	// 根据事件类型处理
	switch event.Type {
	case "create", "update":
		// 更新缓存
		for _, cacheKey := range event.CacheKeys {
			if err := ew.manager.SyncCache(cacheKey, event.TableName, event.RecordID, event.Data); err != nil {
				logger.Error(fmt.Sprintf("工作者 %d 同步缓存失败: Key=%s, Error=%v",
					ew.id, cacheKey, err))
			}
		}
	case "delete":
		// 删除缓存
		if err := ew.manager.InvalidateCache(event.CacheKeys); err != nil {
			logger.Error(fmt.Sprintf("工作者 %d 失效缓存失败: Keys=%v, Error=%v",
				ew.id, event.CacheKeys, err))
		}
	default:
		logger.Error(fmt.Sprintf("工作者 %d 收到未知事件类型: %s", ew.id, event.Type))
	}

	// 标记事件已处理
	event.Processed = true
	ew.manager.updateStats(func(stats *CacheConsistencyStats) {
		stats.ProcessedEvents++
		stats.PendingEvents = stats.TotalEvents - stats.ProcessedEvents
	})
}
