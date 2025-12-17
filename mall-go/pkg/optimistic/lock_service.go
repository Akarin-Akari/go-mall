package optimistic

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// OptimisticLockService 乐观锁服务
type OptimisticLockService struct {
	db *gorm.DB
}

// NewOptimisticLockService 创建乐观锁服务
func NewOptimisticLockService(db *gorm.DB) *OptimisticLockService {
	return &OptimisticLockService{
		db: db,
	}
}

// UpdateConfig 更新配置
type UpdateConfig struct {
	MaxRetries    int           // 最大重试次数，默认3次
	RetryInterval time.Duration // 重试间隔，默认10ms
	BackoffFactor float64       // 退避因子，默认1.5
}

// DefaultUpdateConfig 默认更新配置
func DefaultUpdateConfig() *UpdateConfig {
	return &UpdateConfig{
		MaxRetries:    3,
		RetryInterval: 10 * time.Millisecond,
		BackoffFactor: 1.5,
	}
}

// UpdateResult 更新结果
type UpdateResult struct {
	Success      bool  // 是否成功
	RowsAffected int64 // 影响行数
	Retries      int   // 重试次数
	Error        error // 错误信息
}

// UpdateWithOptimisticLock 使用乐观锁更新记录
// tableName: 表名
// id: 记录ID
// updates: 更新字段映射
// config: 更新配置，可为nil使用默认配置
func (ols *OptimisticLockService) UpdateWithOptimisticLock(
	tableName string,
	id uint,
	updates map[string]interface{},
	config *UpdateConfig,
) *UpdateResult {
	if config == nil {
		config = DefaultUpdateConfig()
	}

	result := &UpdateResult{}

	for retries := 0; retries < config.MaxRetries; retries++ {
		result.Retries = retries

		// 获取当前记录的版本号
		var currentVersion int
		if err := ols.db.Table(tableName).
			Select("version").
			Where("id = ?", id).
			Scan(&currentVersion).Error; err != nil {
			result.Error = fmt.Errorf("获取版本号失败: %v", err)
			return result
		}

		// 准备更新数据，包含版本号递增
		updateData := make(map[string]interface{})
		for k, v := range updates {
			updateData[k] = v
		}
		updateData["version"] = currentVersion + 1
		updateData["updated_at"] = time.Now()

		// 执行乐观锁更新
		dbResult := ols.db.Table(tableName).
			Where("id = ? AND version = ?", id, currentVersion).
			Updates(updateData)

		if dbResult.Error != nil {
			result.Error = fmt.Errorf("更新失败: %v", dbResult.Error)
			return result
		}

		// 检查是否更新成功
		if dbResult.RowsAffected > 0 {
			result.Success = true
			result.RowsAffected = dbResult.RowsAffected
			return result
		}

		// 更新失败，说明版本号已变化，需要重试
		if retries == config.MaxRetries-1 {
			result.Error = fmt.Errorf("更新失败，并发冲突过多，已重试%d次", config.MaxRetries)
			return result
		}

		// 计算退避时间并等待
		backoffTime := time.Duration(float64(config.RetryInterval) *
			(1 + float64(retries)*config.BackoffFactor))
		time.Sleep(backoffTime)
	}

	result.Error = fmt.Errorf("更新失败，超过最大重试次数")
	return result
}

// UpdateWithOptimisticLockTx 在事务中使用乐观锁更新记录
func (ols *OptimisticLockService) UpdateWithOptimisticLockTx(
	tx *gorm.DB,
	tableName string,
	id uint,
	updates map[string]interface{},
	config *UpdateConfig,
) *UpdateResult {
	if config == nil {
		config = DefaultUpdateConfig()
	}

	result := &UpdateResult{}

	for retries := 0; retries < config.MaxRetries; retries++ {
		result.Retries = retries

		// 获取当前记录的版本号
		var currentVersion int
		if err := tx.Table(tableName).
			Select("version").
			Where("id = ?", id).
			Scan(&currentVersion).Error; err != nil {
			result.Error = fmt.Errorf("获取版本号失败: %v", err)
			return result
		}

		// 准备更新数据，包含版本号递增
		updateData := make(map[string]interface{})
		for k, v := range updates {
			updateData[k] = v
		}
		updateData["version"] = currentVersion + 1
		updateData["updated_at"] = time.Now()

		// 执行乐观锁更新
		dbResult := tx.Table(tableName).
			Where("id = ? AND version = ?", id, currentVersion).
			Updates(updateData)

		if dbResult.Error != nil {
			result.Error = fmt.Errorf("更新失败: %v", dbResult.Error)
			return result
		}

		// 检查是否更新成功
		if dbResult.RowsAffected > 0 {
			result.Success = true
			result.RowsAffected = dbResult.RowsAffected
			return result
		}

		// 更新失败，说明版本号已变化，需要重试
		if retries == config.MaxRetries-1 {
			result.Error = fmt.Errorf("更新失败，并发冲突过多，已重试%d次", config.MaxRetries)
			return result
		}

		// 在事务中不能sleep太久，使用较短的等待时间
		time.Sleep(time.Millisecond * time.Duration(retries+1))
	}

	result.Error = fmt.Errorf("更新失败，超过最大重试次数")
	return result
}

// BatchUpdateWithOptimisticLock 批量乐观锁更新
type BatchUpdateItem struct {
	TableName string
	ID        uint
	Updates   map[string]interface{}
}

// BatchUpdateResult 批量更新结果
type BatchUpdateResult struct {
	Success      bool
	TotalItems   int
	SuccessItems int
	FailedItems  []BatchUpdateItem
	Results      []*UpdateResult
	Error        error
}

// BatchUpdateWithOptimisticLock 批量使用乐观锁更新记录
func (ols *OptimisticLockService) BatchUpdateWithOptimisticLock(
	items []BatchUpdateItem,
	config *UpdateConfig,
) *BatchUpdateResult {
	if config == nil {
		config = DefaultUpdateConfig()
	}

	batchResult := &BatchUpdateResult{
		TotalItems: len(items),
		Results:    make([]*UpdateResult, len(items)),
	}

	// 在事务中执行批量更新
	err := ols.db.Transaction(func(tx *gorm.DB) error {
		for i, item := range items {
			result := ols.UpdateWithOptimisticLockTx(tx, item.TableName, item.ID, item.Updates, config)
			batchResult.Results[i] = result

			if result.Success {
				batchResult.SuccessItems++
			} else {
				batchResult.FailedItems = append(batchResult.FailedItems, item)
				// 如果有失败的项目，回滚整个事务
				return result.Error
			}
		}
		return nil
	})

	if err != nil {
		batchResult.Error = err
		batchResult.Success = false
	} else {
		batchResult.Success = batchResult.SuccessItems == batchResult.TotalItems
	}

	return batchResult
}

// ConflictStatistics 冲突统计
type ConflictStatistics struct {
	TotalUpdates  int64   // 总更新次数
	ConflictCount int64   // 冲突次数
	ConflictRate  float64 // 冲突率
	AvgRetries    float64 // 平均重试次数
	MaxRetries    int     // 最大重试次数
}

// GetConflictStatistics 获取冲突统计信息（需要应用层记录）
func (ols *OptimisticLockService) GetConflictStatistics() *ConflictStatistics {
	// 这里可以实现统计逻辑，比如从Redis或数据库中获取统计数据
	// 暂时返回空统计
	return &ConflictStatistics{}
}
