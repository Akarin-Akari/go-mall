package payment

import (
	"context"
	"fmt"
	"sync"
	"time"

	"mall-go/internal/model"
	"mall-go/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SyncManager 同步管理器
type SyncManager struct {
	db         *gorm.DB
	eventQueue chan *SyncEvent
	workers    int
	retryQueue chan *SyncEvent
	maxRetries int
	retryDelay time.Duration
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

// SyncEvent 同步事件
type SyncEvent struct {
	ID         string                 `json:"id"`
	Type       SyncEventType          `json:"type"`
	PaymentID  uint                   `json:"payment_id"`
	OrderID    uint                   `json:"order_id"`
	UserID     uint                   `json:"user_id"`
	Data       map[string]interface{} `json:"data"`
	Timestamp  time.Time              `json:"timestamp"`
	RetryCount int                    `json:"retry_count"`
	MaxRetries int                    `json:"max_retries"`
	NextRetry  time.Time              `json:"next_retry"`
}

// SyncEventType 同步事件类型
type SyncEventType string

const (
	SyncEventPaymentSuccess  SyncEventType = "payment_success"  // 支付成功
	SyncEventPaymentFailed   SyncEventType = "payment_failed"   // 支付失败
	SyncEventPaymentCanceled SyncEventType = "payment_canceled" // 支付取消
	SyncEventRefundSuccess   SyncEventType = "refund_success"   // 退款成功
	SyncEventRefundFailed    SyncEventType = "refund_failed"    // 退款失败
)

// SyncResult 同步结果
type SyncResult struct {
	Success   bool                   `json:"success"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewSyncManager 创建同步管理器
func NewSyncManager(db *gorm.DB, workers int) *SyncManager {
	ctx, cancel := context.WithCancel(context.Background())

	sm := &SyncManager{
		db:         db,
		eventQueue: make(chan *SyncEvent, 1000),
		workers:    workers,
		retryQueue: make(chan *SyncEvent, 500),
		maxRetries: 3,
		retryDelay: time.Second * 30,
		ctx:        ctx,
		cancel:     cancel,
	}

	// 启动工作协程
	sm.startWorkers()

	// 启动重试协程
	sm.startRetryWorker()

	return sm
}

// startWorkers 启动工作协程
func (sm *SyncManager) startWorkers() {
	for i := 0; i < sm.workers; i++ {
		sm.wg.Add(1)
		go sm.worker(i)
	}
}

// startRetryWorker 启动重试协程
func (sm *SyncManager) startRetryWorker() {
	sm.wg.Add(1)
	go sm.retryWorker()
}

// worker 工作协程
func (sm *SyncManager) worker(id int) {
	defer sm.wg.Done()

	logger.Info("同步工作协程启动", zap.Int("worker_id", id))

	for {
		select {
		case <-sm.ctx.Done():
			logger.Info("同步工作协程停止", zap.Int("worker_id", id))
			return
		case event := <-sm.eventQueue:
			sm.processEvent(event, id)
		}
	}
}

// retryWorker 重试工作协程
func (sm *SyncManager) retryWorker() {
	defer sm.wg.Done()

	logger.Info("重试工作协程启动")

	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for {
		select {
		case <-sm.ctx.Done():
			logger.Info("重试工作协程停止")
			return
		case <-ticker.C:
			sm.processRetryEvents()
		case event := <-sm.retryQueue:
			if time.Now().After(event.NextRetry) {
				sm.eventQueue <- event
			} else {
				// 重新放入重试队列
				go func() {
					time.Sleep(time.Until(event.NextRetry))
					sm.retryQueue <- event
				}()
			}
		}
	}
}

// processEvent 处理同步事件
func (sm *SyncManager) processEvent(event *SyncEvent, workerID int) {
	logger.Info("处理同步事件",
		zap.String("event_id", event.ID),
		zap.String("event_type", string(event.Type)),
		zap.Uint("payment_id", event.PaymentID),
		zap.Int("worker_id", workerID))

	var result *SyncResult
	var err error

	// 根据事件类型处理
	switch event.Type {
	case SyncEventPaymentSuccess:
		result, err = sm.handlePaymentSuccess(event)
	case SyncEventPaymentFailed:
		result, err = sm.handlePaymentFailed(event)
	case SyncEventPaymentCanceled:
		result, err = sm.handlePaymentCanceled(event)
	case SyncEventRefundSuccess:
		result, err = sm.handleRefundSuccess(event)
	case SyncEventRefundFailed:
		result, err = sm.handleRefundFailed(event)
	default:
		err = fmt.Errorf("未知的事件类型: %s", event.Type)
	}

	// 处理结果
	if err != nil {
		logger.Error("同步事件处理失败",
			zap.String("event_id", event.ID),
			zap.Error(err))

		// 重试逻辑
		if event.RetryCount < sm.maxRetries {
			event.RetryCount++
			event.NextRetry = time.Now().Add(sm.retryDelay * time.Duration(event.RetryCount))
			sm.retryQueue <- event
			logger.Info("事件加入重试队列",
				zap.String("event_id", event.ID),
				zap.Int("retry_count", event.RetryCount))
		} else {
			logger.Error("事件重试次数超限，放弃处理",
				zap.String("event_id", event.ID),
				zap.Int("retry_count", event.RetryCount))
			sm.recordFailedEvent(event, err)
		}
	} else {
		logger.Info("同步事件处理成功",
			zap.String("event_id", event.ID),
			zap.Any("result", result))
		sm.recordSuccessEvent(event, result)
	}
}

// handlePaymentSuccess 处理支付成功事件
func (sm *SyncManager) handlePaymentSuccess(event *SyncEvent) (*SyncResult, error) {
	// 开启事务
	tx := sm.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新订单状态
	err := tx.Model(&model.Order{}).Where("id = ?", event.OrderID).Updates(map[string]interface{}{
		"status":         model.OrderStatusPaid,
		"payment_status": model.PaymentStatusPaid,
		"payment_method": event.Data["payment_method"],
		"paid_at":        time.Now(),
	}).Error

	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("更新订单状态失败: %v", err)
	}

	// 更新商品库存（如果需要）
	if shouldUpdateStock(event) {
		if err := sm.updateProductStock(tx, event); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("更新商品库存失败: %v", err)
		}
	}

	// TODO: 创建订单日志功能需要实现OrderLog模型
	// orderLog := &model.OrderLog{
	// 	OrderID:     event.OrderID,
	// 	UserID:      event.UserID,
	// 	Action:      "PAYMENT_SUCCESS",
	// 	Description: "支付成功",
	// 	CreatedAt:   time.Now(),
	// }

	// if err := tx.Create(orderLog).Error; err != nil {
	// 	logger.Error("创建订单日志失败", zap.Error(err))
	// 	// 不影响主流程
	// }

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("提交事务失败: %v", err)
	}

	return &SyncResult{
		Success:   true,
		Message:   "订单状态同步成功",
		Data:      map[string]interface{}{"order_id": event.OrderID},
		Timestamp: time.Now(),
	}, nil
}

// handlePaymentFailed 处理支付失败事件
func (sm *SyncManager) handlePaymentFailed(event *SyncEvent) (*SyncResult, error) {
	// 更新订单状态
	err := sm.db.Model(&model.Order{}).Where("id = ?", event.OrderID).Updates(map[string]interface{}{
		"status":         model.OrderStatusCancelled,
		"payment_status": model.PaymentStatusFailed,
		"cancelled_at":   time.Now(),
	}).Error

	if err != nil {
		return nil, fmt.Errorf("更新订单状态失败: %v", err)
	}

	return &SyncResult{
		Success:   true,
		Message:   "支付失败状态同步成功",
		Data:      map[string]interface{}{"order_id": event.OrderID},
		Timestamp: time.Now(),
	}, nil
}

// handlePaymentCanceled 处理支付取消事件
func (sm *SyncManager) handlePaymentCanceled(event *SyncEvent) (*SyncResult, error) {
	// 更新订单状态
	err := sm.db.Model(&model.Order{}).Where("id = ?", event.OrderID).Updates(map[string]interface{}{
		"status":         model.OrderStatusCancelled,
		"payment_status": model.PaymentStatusCancelled,
		"cancelled_at":   time.Now(),
	}).Error

	if err != nil {
		return nil, fmt.Errorf("更新订单状态失败: %v", err)
	}

	return &SyncResult{
		Success:   true,
		Message:   "支付取消状态同步成功",
		Data:      map[string]interface{}{"order_id": event.OrderID},
		Timestamp: time.Now(),
	}, nil
}

// handleRefundSuccess 处理退款成功事件
func (sm *SyncManager) handleRefundSuccess(event *SyncEvent) (*SyncResult, error) {
	// 开启事务
	tx := sm.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新订单状态
	err := tx.Model(&model.Order{}).Where("id = ?", event.OrderID).Updates(map[string]interface{}{
		"status":         model.OrderStatusRefunded,
		"payment_status": model.PaymentStatusRefunded,
		"refunded_at":    time.Now(),
	}).Error

	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("更新订单状态失败: %v", err)
	}

	// 恢复商品库存（如果需要）
	if shouldRestoreStock(event) {
		if err := sm.restoreProductStock(tx, event); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("恢复商品库存失败: %v", err)
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("提交事务失败: %v", err)
	}

	return &SyncResult{
		Success:   true,
		Message:   "退款成功状态同步成功",
		Data:      map[string]interface{}{"order_id": event.OrderID},
		Timestamp: time.Now(),
	}, nil
}

// handleRefundFailed 处理退款失败事件
func (sm *SyncManager) handleRefundFailed(event *SyncEvent) (*SyncResult, error) {
	// 记录退款失败日志
	logger.Error("退款失败",
		zap.Uint("order_id", event.OrderID),
		zap.Uint("payment_id", event.PaymentID))

	return &SyncResult{
		Success:   true,
		Message:   "退款失败状态记录成功",
		Data:      map[string]interface{}{"order_id": event.OrderID},
		Timestamp: time.Now(),
	}, nil
}

// PublishEvent 发布同步事件
func (sm *SyncManager) PublishEvent(eventType SyncEventType, paymentID, orderID, userID uint, data map[string]interface{}) error {
	event := &SyncEvent{
		ID:         generateEventID(),
		Type:       eventType,
		PaymentID:  paymentID,
		OrderID:    orderID,
		UserID:     userID,
		Data:       data,
		Timestamp:  time.Now(),
		RetryCount: 0,
		MaxRetries: sm.maxRetries,
	}

	select {
	case sm.eventQueue <- event:
		logger.Info("同步事件已发布",
			zap.String("event_id", event.ID),
			zap.String("event_type", string(eventType)))
		return nil
	default:
		return fmt.Errorf("事件队列已满")
	}
}

// updateProductStock 更新商品库存
func (sm *SyncManager) updateProductStock(tx *gorm.DB, event *SyncEvent) error {
	// 这里应该根据订单项更新商品库存
	// 简化处理，实际应该查询订单项
	return nil
}

// restoreProductStock 恢复商品库存
func (sm *SyncManager) restoreProductStock(tx *gorm.DB, event *SyncEvent) error {
	// 这里应该根据订单项恢复商品库存
	// 简化处理，实际应该查询订单项
	return nil
}

// shouldUpdateStock 是否应该更新库存
func shouldUpdateStock(event *SyncEvent) bool {
	// 根据业务规则判断是否需要更新库存
	return true
}

// shouldRestoreStock 是否应该恢复库存
func shouldRestoreStock(event *SyncEvent) bool {
	// 根据业务规则判断是否需要恢复库存
	return true
}

// processRetryEvents 处理重试事件
func (sm *SyncManager) processRetryEvents() {
	// 从数据库中查询需要重试的事件
	// 这里简化处理
}

// recordSuccessEvent 记录成功事件
func (sm *SyncManager) recordSuccessEvent(event *SyncEvent, result *SyncResult) {
	// 记录成功的同步事件
	logger.Info("同步事件成功",
		zap.String("event_id", event.ID),
		zap.Any("result", result))
}

// recordFailedEvent 记录失败事件
func (sm *SyncManager) recordFailedEvent(event *SyncEvent, err error) {
	// 记录失败的同步事件
	logger.Error("同步事件最终失败",
		zap.String("event_id", event.ID),
		zap.Error(err))
}

// generateEventID 生成事件ID
func generateEventID() string {
	return fmt.Sprintf("sync_%d", time.Now().UnixNano())
}

// Stop 停止同步管理器
func (sm *SyncManager) Stop() {
	logger.Info("停止同步管理器")
	sm.cancel()
	sm.wg.Wait()
	close(sm.eventQueue)
	close(sm.retryQueue)
	logger.Info("同步管理器已停止")
}
