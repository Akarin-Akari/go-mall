package order

import (
	"context"
	"fmt"
	"time"

	"mall-go/internal/model"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// OrderStatusManager 订单状态管理器
type OrderStatusManager struct {
	db  *gorm.DB
	rdb *redis.Client
	ctx context.Context
}

// NewOrderStatusManager 创建订单状态管理器
func NewOrderStatusManager(db *gorm.DB, rdb *redis.Client) *OrderStatusManager {
	return &OrderStatusManager{
		db:  db,
		rdb: rdb,
		ctx: context.Background(),
	}
}

// StatusUpdateRequest 状态更新请求
type StatusUpdateRequest struct {
	OrderID      uint   `json:"order_id"`
	ToStatus     string `json:"to_status"`
	OperatorID   uint   `json:"operator_id"`
	OperatorType string `json:"operator_type"`
	Reason       string `json:"reason"`
	Remark       string `json:"remark"`
}

// StatusUpdateResult 状态更新结果
type StatusUpdateResult struct {
	Success      bool   `json:"success"`
	OrderID      uint   `json:"order_id"`
	FromStatus   string `json:"from_status"`
	ToStatus     string `json:"to_status"`
	Error        string `json:"error,omitempty"`
	RetryCount   int    `json:"retry_count"`
}

// UpdateOrderStatusWithLock 使用分布式锁和乐观锁更新订单状态
func (osm *OrderStatusManager) UpdateOrderStatusWithLock(req *StatusUpdateRequest) (*StatusUpdateResult, error) {
	result := &StatusUpdateResult{
		OrderID:  req.OrderID,
		ToStatus: req.ToStatus,
	}

	// 获取分布式锁
	lockKey := fmt.Sprintf("order_status_lock:%d", req.OrderID)
	lockValue := fmt.Sprintf("%d", time.Now().UnixNano())
	
	success, err := osm.rdb.SetNX(osm.ctx, lockKey, lockValue, 30*time.Second).Result()
	if err != nil || !success {
		result.Error = "获取订单状态锁失败"
		return result, fmt.Errorf("获取订单状态锁失败")
	}

	// 确保释放锁
	defer osm.releaseLock(lockKey, lockValue)

	// 使用乐观锁更新状态（带重试）
	maxRetries := 3
	for retries := 0; retries < maxRetries; retries++ {
		result.RetryCount = retries + 1
		
		updateResult, err := osm.updateOrderStatusWithOptimisticLock(req)
		if err == nil {
			result.Success = true
			result.FromStatus = updateResult.FromStatus
			return result, nil
		}

		// 如果是版本冲突，重试
		if retries < maxRetries-1 {
			time.Sleep(time.Millisecond * time.Duration(10*(retries+1)))
			continue
		}

		result.Error = err.Error()
		return result, err
	}

	result.Error = "状态更新失败，超过最大重试次数"
	return result, fmt.Errorf("状态更新失败，超过最大重试次数")
}

// updateOrderStatusWithOptimisticLock 使用乐观锁更新订单状态
func (osm *OrderStatusManager) updateOrderStatusWithOptimisticLock(req *StatusUpdateRequest) (*StatusUpdateResult, error) {
	result := &StatusUpdateResult{
		OrderID:  req.OrderID,
		ToStatus: req.ToStatus,
	}

	// 开始事务
	tx := osm.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取当前订单信息
	var order model.Order
	if err := tx.Preload("OrderItems").First(&order, req.OrderID).Error; err != nil {
		tx.Rollback()
		return result, fmt.Errorf("订单不存在")
	}

	result.FromStatus = order.Status

	// 检查状态流转是否合法
	if !osm.isValidTransition(order.Status, req.ToStatus) {
		tx.Rollback()
		return result, fmt.Errorf("不能从状态 %s 转换到 %s", order.Status, req.ToStatus)
	}

	// 检查状态流转条件
	if !osm.checkTransitionCondition(&order, req.ToStatus) {
		tx.Rollback()
		return result, fmt.Errorf("状态流转条件不满足")
	}

	// 执行状态流转前置动作
	if err := osm.executePreAction(tx, &order, req.ToStatus); err != nil {
		tx.Rollback()
		return result, fmt.Errorf("执行状态流转前置动作失败: %v", err)
	}

	// 使用乐观锁更新订单状态
	updateResult := tx.Model(&order).
		Where("id = ? AND version = ?", order.ID, order.Version).
		Updates(map[string]interface{}{
			"status":     req.ToStatus,
			"version":    order.Version + 1,
			"updated_at": time.Now(),
		})

	if updateResult.Error != nil {
		tx.Rollback()
		return result, fmt.Errorf("更新订单状态失败: %v", updateResult.Error)
	}

	if updateResult.RowsAffected == 0 {
		tx.Rollback()
		return result, fmt.Errorf("订单状态已被其他操作修改，请重试")
	}

	// 执行状态流转后置动作
	if err := osm.executePostAction(tx, &order, req.ToStatus); err != nil {
		tx.Rollback()
		return result, fmt.Errorf("执行状态流转后置动作失败: %v", err)
	}

	// 记录状态流转日志
	statusLog := &model.OrderStatusLog{
		OrderID:      req.OrderID,
		FromStatus:   order.Status,
		ToStatus:     req.ToStatus,
		OperatorID:   req.OperatorID,
		OperatorType: req.OperatorType,
		Reason:       req.Reason,
		Remark:       req.Remark,
	}

	if err := tx.Create(statusLog).Error; err != nil {
		tx.Rollback()
		return result, fmt.Errorf("记录状态日志失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return result, fmt.Errorf("提交事务失败: %v", err)
	}

	result.Success = true
	return result, nil
}

// isValidTransition 检查状态流转是否合法
func (osm *OrderStatusManager) isValidTransition(from, to string) bool {
	validTransitions := map[string][]string{
		model.OrderStatusPending: {
			model.OrderStatusPaid,
			model.OrderStatusCancelled,
			model.OrderStatusClosed,
		},
		model.OrderStatusPaid: {
			model.OrderStatusShipped,
			model.OrderStatusCancelled,
			model.OrderStatusRefunding,
		},
		model.OrderStatusShipped: {
			model.OrderStatusDelivered,
			model.OrderStatusReceived,
			model.OrderStatusRefunding,
		},
		model.OrderStatusDelivered: {
			model.OrderStatusReceived,
			model.OrderStatusRefunding,
		},
		model.OrderStatusReceived: {
			model.OrderStatusCompleted,
			model.OrderStatusRefunding,
		},
		model.OrderStatusRefunding: {
			model.OrderStatusRefunded,
			model.OrderStatusCancelled,
		},
	}

	allowedStatuses, exists := validTransitions[from]
	if !exists {
		return false
	}

	for _, status := range allowedStatuses {
		if status == to {
			return true
		}
	}

	return false
}

// checkTransitionCondition 检查状态流转条件
func (osm *OrderStatusManager) checkTransitionCondition(order *model.Order, toStatus string) bool {
	switch toStatus {
	case model.OrderStatusPaid:
		return order.PaidAmount.GreaterThanOrEqual(order.PayableAmount)
	case model.OrderStatusCancelled:
		return order.Status == model.OrderStatusPending || order.Status == model.OrderStatusPaid
	case model.OrderStatusReceived:
		return order.Status == model.OrderStatusDelivered || order.Status == model.OrderStatusShipped
	default:
		return true
	}
}

// executePreAction 执行状态流转前置动作
func (osm *OrderStatusManager) executePreAction(tx *gorm.DB, order *model.Order, toStatus string) error {
	switch toStatus {
	case model.OrderStatusPaid:
		now := time.Now()
		order.PayTime = &now
	case model.OrderStatusShipped:
		now := time.Now()
		order.ShipTime = &now
		order.ShippingStatus = model.ShippingStatusShipped
		// 设置收货超时时间（7天后自动确认收货）
		receiveExpireTime := now.Add(7 * 24 * time.Hour)
		order.ReceiveExpireTime = &receiveExpireTime
	case model.OrderStatusDelivered:
		now := time.Now()
		order.DeliveryTime = &now
	case model.OrderStatusReceived:
		now := time.Now()
		order.ReceiveTime = &now
		order.ShippingStatus = model.ShippingStatusReceived
		// 设置评价超时时间（15天后不能评价）
		reviewExpireTime := now.Add(15 * 24 * time.Hour)
		order.ReviewExpireTime = &reviewExpireTime
	case model.OrderStatusCompleted:
		now := time.Now()
		order.FinishTime = &now
	case model.OrderStatusCancelled:
		now := time.Now()
		order.CancelTime = &now
		// 如果已支付，需要退款
		if order.Status == model.OrderStatusPaid {
			order.RefundStatus = model.RefundStatusPending
		}
	}
	return nil
}

// executePostAction 执行状态流转后置动作
func (osm *OrderStatusManager) executePostAction(tx *gorm.DB, order *model.Order, toStatus string) error {
	switch toStatus {
	case model.OrderStatusCancelled:
		// 恢复库存
		return osm.restoreOrderStock(tx, order)
	default:
		return nil
	}
}

// restoreOrderStock 恢复订单库存
func (osm *OrderStatusManager) restoreOrderStock(tx *gorm.DB, order *model.Order) error {
	for _, item := range order.OrderItems {
		if item.SKUID > 0 {
			// 恢复SKU库存
			result := tx.Model(&model.ProductSKU{}).
				Where("id = ?", item.SKUID).
				Updates(map[string]interface{}{
					"stock":   gorm.Expr("stock + ?", item.Quantity),
					"version": gorm.Expr("version + 1"),
				})
			if result.Error != nil {
				return fmt.Errorf("恢复SKU库存失败: %v", result.Error)
			}
		} else {
			// 恢复商品库存
			result := tx.Model(&model.Product{}).
				Where("id = ?", item.ProductID).
				Updates(map[string]interface{}{
					"stock":      gorm.Expr("stock + ?", item.Quantity),
					"sold_count": gorm.Expr("GREATEST(sold_count - ?, 0)", item.Quantity),
					"version":    gorm.Expr("version + 1"),
				})
			if result.Error != nil {
				return fmt.Errorf("恢复商品库存失败: %v", result.Error)
			}
		}
	}
	return nil
}

// releaseLock 释放分布式锁
func (osm *OrderStatusManager) releaseLock(lockKey, lockValue string) {
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	osm.rdb.Eval(osm.ctx, script, []string{lockKey}, lockValue)
}

// BatchUpdateOrderStatus 批量更新订单状态
func (osm *OrderStatusManager) BatchUpdateOrderStatus(requests []*StatusUpdateRequest) ([]*StatusUpdateResult, error) {
	results := make([]*StatusUpdateResult, len(requests))
	
	for i, req := range requests {
		result, err := osm.UpdateOrderStatusWithLock(req)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
		}
		results[i] = result
	}
	
	return results, nil
}

// GetOrderStatusFlow 获取订单状态流转记录
func (osm *OrderStatusManager) GetOrderStatusFlow(orderID uint) ([]model.OrderStatusLog, error) {
	var statusLogs []model.OrderStatusLog
	if err := osm.db.Where("order_id = ?", orderID).
		Preload("Operator").
		Order("created_at ASC").
		Find(&statusLogs).Error; err != nil {
		return nil, fmt.Errorf("获取状态流转记录失败: %v", err)
	}

	return statusLogs, nil
}

// CanTransitionTo 检查订单是否可以转换到指定状态
func (osm *OrderStatusManager) CanTransitionTo(orderID uint, toStatus string) (bool, string) {
	var order model.Order
	if err := osm.db.First(&order, orderID).Error; err != nil {
		return false, "订单不存在"
	}

	if !osm.isValidTransition(order.Status, toStatus) {
		return false, fmt.Sprintf("不能从状态 %s 转换到 %s", order.Status, toStatus)
	}

	if !osm.checkTransitionCondition(&order, toStatus) {
		return false, "状态流转条件不满足"
	}

	return true, ""
}
