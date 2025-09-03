package recovery

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"mall-go/internal/model"
	"mall-go/pkg/logger"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ErrorRecoveryService 错误恢复服务
type ErrorRecoveryService struct {
	db  *gorm.DB
	rdb *redis.Client
	ctx context.Context
}

// NewErrorRecoveryService 创建错误恢复服务
func NewErrorRecoveryService(db *gorm.DB, rdb *redis.Client) *ErrorRecoveryService {
	return &ErrorRecoveryService{
		db:  db,
		rdb: rdb,
		ctx: context.Background(),
	}
}

// RecoveryTask 恢复任务
type RecoveryTask struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Status      string                 `json:"status"`
	Data        map[string]interface{} `json:"data"`
	Error       string                 `json:"error,omitempty"`
	RetryCount  int                    `json:"retry_count"`
	MaxRetries  int                    `json:"max_retries"`
	NextRetry   time.Time              `json:"next_retry"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// RecoveryTaskStatus 恢复任务状态
const (
	TaskStatusPending   = "pending"
	TaskStatusRunning   = "running"
	TaskStatusCompleted = "completed"
	TaskStatusFailed    = "failed"
	TaskStatusCancelled = "cancelled"
)

// RecoveryTaskType 恢复任务类型
const (
	TaskTypeOrderRecovery    = "order_recovery"
	TaskTypePaymentRecovery  = "payment_recovery"
	TaskTypeInventoryRecovery = "inventory_recovery"
	TaskTypeRefundRecovery   = "refund_recovery"
)

// CreateRecoveryTask 创建恢复任务
func (ers *ErrorRecoveryService) CreateRecoveryTask(taskType string, data map[string]interface{}, maxRetries int) (*RecoveryTask, error) {
	task := &RecoveryTask{
		ID:         fmt.Sprintf("%s_%d", taskType, time.Now().UnixNano()),
		Type:       taskType,
		Status:     TaskStatusPending,
		Data:       data,
		RetryCount: 0,
		MaxRetries: maxRetries,
		NextRetry:  time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// 保存到Redis
	taskJSON, err := json.Marshal(task)
	if err != nil {
		return nil, fmt.Errorf("序列化恢复任务失败: %v", err)
	}

	key := fmt.Sprintf("recovery_task:%s", task.ID)
	if err := ers.rdb.Set(ers.ctx, key, taskJSON, 24*time.Hour).Err(); err != nil {
		return nil, fmt.Errorf("保存恢复任务失败: %v", err)
	}

	// 添加到待处理队列
	if err := ers.rdb.LPush(ers.ctx, "recovery_queue", task.ID).Err(); err != nil {
		return nil, fmt.Errorf("添加到恢复队列失败: %v", err)
	}

	logger.Info("创建恢复任务",
		zap.String("task_id", task.ID),
		zap.String("task_type", taskType),
		zap.Int("max_retries", maxRetries))

	return task, nil
}

// ProcessRecoveryTasks 处理恢复任务
func (ers *ErrorRecoveryService) ProcessRecoveryTasks() {
	for {
		// 从队列中获取任务
		taskID, err := ers.rdb.BRPop(ers.ctx, 5*time.Second, "recovery_queue").Result()
		if err != nil {
			if err != redis.Nil {
				logger.Error("获取恢复任务失败", zap.Error(err))
			}
			continue
		}

		if len(taskID) < 2 {
			continue
		}

		// 处理任务
		if err := ers.processTask(taskID[1]); err != nil {
			logger.Error("处理恢复任务失败",
				zap.String("task_id", taskID[1]),
				zap.Error(err))
		}
	}
}

// processTask 处理单个任务
func (ers *ErrorRecoveryService) processTask(taskID string) error {
	// 获取任务详情
	task, err := ers.getTask(taskID)
	if err != nil {
		return fmt.Errorf("获取任务失败: %v", err)
	}

	// 检查是否到了重试时间
	if time.Now().Before(task.NextRetry) {
		// 重新放回队列
		ers.rdb.LPush(ers.ctx, "recovery_queue", taskID)
		return nil
	}

	// 更新任务状态为运行中
	task.Status = TaskStatusRunning
	task.UpdatedAt = time.Now()
	if err := ers.updateTask(task); err != nil {
		return fmt.Errorf("更新任务状态失败: %v", err)
	}

	logger.Info("开始处理恢复任务",
		zap.String("task_id", taskID),
		zap.String("task_type", task.Type),
		zap.Int("retry_count", task.RetryCount))

	// 根据任务类型执行恢复逻辑
	var recoveryErr error
	switch task.Type {
	case TaskTypeOrderRecovery:
		recoveryErr = ers.recoverOrder(task)
	case TaskTypePaymentRecovery:
		recoveryErr = ers.recoverPayment(task)
	case TaskTypeInventoryRecovery:
		recoveryErr = ers.recoverInventory(task)
	case TaskTypeRefundRecovery:
		recoveryErr = ers.recoverRefund(task)
	default:
		recoveryErr = fmt.Errorf("未知的恢复任务类型: %s", task.Type)
	}

	// 处理恢复结果
	if recoveryErr == nil {
		// 恢复成功
		task.Status = TaskStatusCompleted
		task.UpdatedAt = time.Now()
		ers.updateTask(task)

		logger.Info("恢复任务执行成功",
			zap.String("task_id", taskID),
			zap.String("task_type", task.Type))
	} else {
		// 恢复失败
		task.RetryCount++
		task.Error = recoveryErr.Error()
		task.UpdatedAt = time.Now()

		if task.RetryCount >= task.MaxRetries {
			// 超过最大重试次数，标记为失败
			task.Status = TaskStatusFailed
			ers.updateTask(task)

			logger.Error("恢复任务最终失败",
				zap.String("task_id", taskID),
				zap.String("task_type", task.Type),
				zap.Int("retry_count", task.RetryCount),
				zap.Error(recoveryErr))

			// 发送告警通知
			ers.sendAlertNotification(task, recoveryErr)
		} else {
			// 计算下次重试时间（指数退避）
			backoffDuration := time.Duration(task.RetryCount*task.RetryCount) * time.Minute
			task.NextRetry = time.Now().Add(backoffDuration)
			task.Status = TaskStatusPending
			ers.updateTask(task)

			// 重新放回队列
			ers.rdb.LPush(ers.ctx, "recovery_queue", taskID)

			logger.Warn("恢复任务失败，将重试",
				zap.String("task_id", taskID),
				zap.String("task_type", task.Type),
				zap.Int("retry_count", task.RetryCount),
				zap.Time("next_retry", task.NextRetry),
				zap.Error(recoveryErr))
		}
	}

	return nil
}

// recoverOrder 恢复订单
func (ers *ErrorRecoveryService) recoverOrder(task *RecoveryTask) error {
	orderID, ok := task.Data["order_id"].(float64)
	if !ok {
		return fmt.Errorf("无效的订单ID")
	}

	var order model.Order
	if err := ers.db.First(&order, uint(orderID)).Error; err != nil {
		return fmt.Errorf("查询订单失败: %v", err)
	}

	// 根据订单状态执行相应的恢复逻辑
	switch order.Status {
	case model.OrderStatusPending:
		// 检查支付超时
		if order.PayExpireTime != nil && time.Now().After(*order.PayExpireTime) {
			return ers.cancelExpiredOrder(&order)
		}
	case model.OrderStatusPaid:
		// 检查是否需要自动发货
		return ers.checkAutoShipping(&order)
	case model.OrderStatusShipped:
		// 检查物流状态
		return ers.updateShippingStatus(&order)
	}

	return nil
}

// recoverPayment 恢复支付
func (ers *ErrorRecoveryService) recoverPayment(task *RecoveryTask) error {
	paymentID, ok := task.Data["payment_id"].(float64)
	if !ok {
		return fmt.Errorf("无效的支付ID")
	}

	var payment model.Payment
	if err := ers.db.First(&payment, uint(paymentID)).Error; err != nil {
		return fmt.Errorf("查询支付记录失败: %v", err)
	}

	// 查询第三方支付状态
	return ers.syncPaymentStatus(&payment)
}

// recoverInventory 恢复库存
func (ers *ErrorRecoveryService) recoverInventory(task *RecoveryTask) error {
	productID, ok := task.Data["product_id"].(float64)
	if !ok {
		return fmt.Errorf("无效的商品ID")
	}

	quantity, ok := task.Data["quantity"].(float64)
	if !ok {
		return fmt.Errorf("无效的数量")
	}

	// 恢复库存
	result := ers.db.Model(&model.Product{}).
		Where("id = ?", uint(productID)).
		Updates(map[string]interface{}{
			"stock":      gorm.Expr("stock + ?", int(quantity)),
			"sold_count": gorm.Expr("GREATEST(sold_count - ?, 0)", int(quantity)),
			"version":    gorm.Expr("version + 1"),
		})

	if result.Error != nil {
		return fmt.Errorf("恢复库存失败: %v", result.Error)
	}

	logger.Info("库存恢复成功",
		zap.Uint("product_id", uint(productID)),
		zap.Int("quantity", int(quantity)))

	return nil
}

// recoverRefund 恢复退款
func (ers *ErrorRecoveryService) recoverRefund(task *RecoveryTask) error {
	refundID, ok := task.Data["refund_id"].(float64)
	if !ok {
		return fmt.Errorf("无效的退款ID")
	}

	// 查询退款记录并处理
	var refund model.PaymentRefund
	if err := ers.db.First(&refund, uint(refundID)).Error; err != nil {
		return fmt.Errorf("查询退款记录失败: %v", err)
	}

	// 调用第三方退款接口
	return ers.processThirdPartyRefund(&refund)
}

// 辅助方法实现
func (ers *ErrorRecoveryService) getTask(taskID string) (*RecoveryTask, error) {
	key := fmt.Sprintf("recovery_task:%s", taskID)
	taskJSON, err := ers.rdb.Get(ers.ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var task RecoveryTask
	if err := json.Unmarshal([]byte(taskJSON), &task); err != nil {
		return nil, err
	}

	return &task, nil
}

func (ers *ErrorRecoveryService) updateTask(task *RecoveryTask) error {
	taskJSON, err := json.Marshal(task)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("recovery_task:%s", task.ID)
	return ers.rdb.Set(ers.ctx, key, taskJSON, 24*time.Hour).Err()
}

func (ers *ErrorRecoveryService) cancelExpiredOrder(order *model.Order) error {
	// 实现订单取消逻辑
	return nil
}

func (ers *ErrorRecoveryService) checkAutoShipping(order *model.Order) error {
	// 实现自动发货检查逻辑
	return nil
}

func (ers *ErrorRecoveryService) updateShippingStatus(order *model.Order) error {
	// 实现物流状态更新逻辑
	return nil
}

func (ers *ErrorRecoveryService) syncPaymentStatus(payment *model.Payment) error {
	// 实现支付状态同步逻辑
	return nil
}

func (ers *ErrorRecoveryService) processThirdPartyRefund(refund *model.PaymentRefund) error {
	// 实现第三方退款处理逻辑
	return nil
}

func (ers *ErrorRecoveryService) sendAlertNotification(task *RecoveryTask, err error) {
	// 实现告警通知逻辑
	logger.Error("发送恢复任务失败告警",
		zap.String("task_id", task.ID),
		zap.String("task_type", task.Type),
		zap.Error(err))
}

// GetTaskStatus 获取任务状态
func (ers *ErrorRecoveryService) GetTaskStatus(taskID string) (*RecoveryTask, error) {
	return ers.getTask(taskID)
}

// CancelTask 取消任务
func (ers *ErrorRecoveryService) CancelTask(taskID string) error {
	task, err := ers.getTask(taskID)
	if err != nil {
		return err
	}

	task.Status = TaskStatusCancelled
	task.UpdatedAt = time.Now()

	return ers.updateTask(task)
}

// 全局错误恢复服务实例
var globalErrorRecoveryService *ErrorRecoveryService

// InitGlobalErrorRecoveryService 初始化全局错误恢复服务
func InitGlobalErrorRecoveryService(db *gorm.DB, rdb *redis.Client) {
	globalErrorRecoveryService = NewErrorRecoveryService(db, rdb)
	
	// 启动后台任务处理器
	go globalErrorRecoveryService.ProcessRecoveryTasks()
}

// GetGlobalErrorRecoveryService 获取全局错误恢复服务
func GetGlobalErrorRecoveryService() *ErrorRecoveryService {
	return globalErrorRecoveryService
}
