package order

import (
	"fmt"
	"time"

	"mall-go/internal/model"

	"gorm.io/gorm"
)

// StatusService 订单状态管理服务
type StatusService struct {
	db *gorm.DB
}

// NewStatusService 创建订单状态管理服务
func NewStatusService(db *gorm.DB) *StatusService {
	return &StatusService{
		db: db,
	}
}

// StatusTransition 状态流转规则
type StatusTransition struct {
	From      string
	To        string
	Condition func(*model.Order) bool
	Action    func(*gorm.DB, *model.Order) error
}

// getStatusTransitions 获取状态流转规则
func (ss *StatusService) getStatusTransitions() []StatusTransition {
	return []StatusTransition{
		// 待支付 -> 已支付
		{
			From: model.OrderStatusPending,
			To:   model.OrderStatusPaid,
			Condition: func(order *model.Order) bool {
				return order.PaidAmount.GreaterThanOrEqual(order.PayableAmount)
			},
			Action: func(tx *gorm.DB, order *model.Order) error {
				now := time.Now()
				order.PayTime = &now
				return nil
			},
		},
		// 已支付 -> 已发货
		{
			From: model.OrderStatusPaid,
			To:   model.OrderStatusShipped,
			Condition: func(order *model.Order) bool {
				return true // 管理员手动发货
			},
			Action: func(tx *gorm.DB, order *model.Order) error {
				now := time.Now()
				order.ShipTime = &now
				order.ShippingStatus = model.ShippingStatusShipped
				// 设置收货超时时间（7天后自动确认收货）
				receiveExpireTime := now.Add(7 * 24 * time.Hour)
				order.ReceiveExpireTime = &receiveExpireTime
				return nil
			},
		},
		// 已发货 -> 已配送
		{
			From: model.OrderStatusShipped,
			To:   model.OrderStatusDelivered,
			Condition: func(order *model.Order) bool {
				return order.ShippingStatus == model.ShippingStatusDelivered
			},
			Action: func(tx *gorm.DB, order *model.Order) error {
				now := time.Now()
				order.DeliveryTime = &now
				return nil
			},
		},
		// 已配送 -> 已收货
		{
			From: model.OrderStatusDelivered,
			To:   model.OrderStatusReceived,
			Condition: func(order *model.Order) bool {
				return true // 用户确认收货或自动确认
			},
			Action: func(tx *gorm.DB, order *model.Order) error {
				now := time.Now()
				order.ReceiveTime = &now
				order.ShippingStatus = model.ShippingStatusReceived
				// 设置评价超时时间（15天后不能评价）
				reviewExpireTime := now.Add(15 * 24 * time.Hour)
				order.ReviewExpireTime = &reviewExpireTime
				return nil
			},
		},
		// 已收货 -> 已完成
		{
			From: model.OrderStatusReceived,
			To:   model.OrderStatusCompleted,
			Condition: func(order *model.Order) bool {
				return true // 用户评价后或超过售后期
			},
			Action: func(tx *gorm.DB, order *model.Order) error {
				now := time.Now()
				order.FinishTime = &now
				return nil
			},
		},
		// 待支付 -> 已取消
		{
			From: model.OrderStatusPending,
			To:   model.OrderStatusCancelled,
			Condition: func(order *model.Order) bool {
				return true // 用户取消或支付超时
			},
			Action: func(tx *gorm.DB, order *model.Order) error {
				now := time.Now()
				order.CancelTime = &now
				// 恢复库存
				return ss.restoreStock(tx, order)
			},
		},
		// 已支付 -> 已取消（需要退款）
		{
			From: model.OrderStatusPaid,
			To:   model.OrderStatusCancelled,
			Condition: func(order *model.Order) bool {
				return true // 管理员取消订单
			},
			Action: func(tx *gorm.DB, order *model.Order) error {
				now := time.Now()
				order.CancelTime = &now
				order.RefundStatus = model.RefundStatusPending
				// 恢复库存
				return ss.restoreStock(tx, order)
			},
		},
		// 任何状态 -> 退款中
		{
			From: "*", // 通配符表示任何已支付状态
			To:   model.OrderStatusRefunding,
			Condition: func(order *model.Order) bool {
				return order.IsPaid() && order.RefundStatus == model.RefundStatusPending
			},
			Action: func(tx *gorm.DB, order *model.Order) error {
				// 处理退款逻辑
				return nil
			},
		},
		// 退款中 -> 已退款
		{
			From: model.OrderStatusRefunding,
			To:   model.OrderStatusRefunded,
			Condition: func(order *model.Order) bool {
				return order.RefundStatus == model.RefundStatusCompleted
			},
			Action: func(tx *gorm.DB, order *model.Order) error {
				now := time.Now()
				order.RefundTime = &now
				return nil
			},
		},
	}
}

// UpdateOrderStatus 更新订单状态
func (ss *StatusService) UpdateOrderStatus(orderID uint, toStatus string, operatorID uint, operatorType string, reason, remark string) error {
	// 开始事务
	tx := ss.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取订单
	var order model.Order
	if err := tx.Preload("OrderItems").First(&order, orderID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("订单不存在")
	}

	fromStatus := order.Status

	// 检查状态流转是否合法
	if !ss.isValidTransition(fromStatus, toStatus) {
		tx.Rollback()
		return fmt.Errorf("不能从状态 %s 转换到 %s", fromStatus, toStatus)
	}

	// 查找对应的状态流转规则
	transitions := ss.getStatusTransitions()
	var transition *StatusTransition
	for _, t := range transitions {
		if (t.From == fromStatus || t.From == "*") && t.To == toStatus {
			if t.Condition(&order) {
				transition = &t
				break
			}
		}
	}

	if transition == nil {
		tx.Rollback()
		return fmt.Errorf("状态流转条件不满足")
	}

	// 执行状态流转动作
	if transition.Action != nil {
		if err := transition.Action(tx, &order); err != nil {
			tx.Rollback()
			return fmt.Errorf("执行状态流转动作失败: %v", err)
		}
	}

	// 更新订单状态
	order.Status = toStatus
	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新订单状态失败: %v", err)
	}

	// 记录状态流转日志
	statusLog := &model.OrderStatusLog{
		OrderID:      orderID,
		FromStatus:   fromStatus,
		ToStatus:     toStatus,
		OperatorID:   operatorID,
		OperatorType: operatorType,
		Reason:       reason,
		Remark:       remark,
	}

	if err := tx.Create(statusLog).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("记录状态日志失败: %v", err)
	}

	tx.Commit()
	return nil
}

// isValidTransition 检查状态流转是否合法
func (ss *StatusService) isValidTransition(from, to string) bool {
	// 定义合法的状态流转
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

// restoreStock 恢复库存
func (ss *StatusService) restoreStock(tx *gorm.DB, order *model.Order) error {
	for _, item := range order.OrderItems {
		if item.SKUID > 0 {
			// 恢复SKU库存
			if err := tx.Model(&model.ProductSKU{}).
				Where("id = ?", item.SKUID).
				UpdateColumn("stock", gorm.Expr("stock + ?", item.Quantity)).Error; err != nil {
				return fmt.Errorf("恢复SKU库存失败: %v", err)
			}
		} else {
			// 恢复商品库存
			if err := tx.Model(&model.Product{}).
				Where("id = ?", item.ProductID).
				UpdateColumns(map[string]interface{}{
					"stock":      gorm.Expr("stock + ?", item.Quantity),
					"sold_count": gorm.Expr("sold_count - ?", item.Quantity),
				}).Error; err != nil {
				return fmt.Errorf("恢复商品库存失败: %v", err)
			}
		}
	}

	return nil
}

// AutoUpdateExpiredOrders 自动更新过期订单
func (ss *StatusService) AutoUpdateExpiredOrders() error {
	now := time.Now()

	// 处理支付超时的订单
	var expiredOrders []model.Order
	if err := ss.db.Where("status = ? AND pay_expire_time < ?",
		model.OrderStatusPending, now).Find(&expiredOrders).Error; err != nil {
		return fmt.Errorf("查询过期订单失败: %v", err)
	}

	for _, order := range expiredOrders {
		if err := ss.UpdateOrderStatus(order.ID, model.OrderStatusCancelled, 0,
			model.OperatorTypeSystem, "支付超时", "系统自动取消"); err != nil {
			// 记录错误但继续处理其他订单
			fmt.Printf("自动取消订单 %s 失败: %v\n", order.OrderNo, err)
		}
	}

	// 处理自动确认收货的订单
	var autoReceiveOrders []model.Order
	if err := ss.db.Where("status = ? AND receive_expire_time < ?",
		model.OrderStatusDelivered, now).Find(&autoReceiveOrders).Error; err != nil {
		return fmt.Errorf("查询自动确认收货订单失败: %v", err)
	}

	for _, order := range autoReceiveOrders {
		if err := ss.UpdateOrderStatus(order.ID, model.OrderStatusReceived, 0,
			model.OperatorTypeSystem, "自动确认收货", "系统自动确认收货"); err != nil {
			fmt.Printf("自动确认收货订单 %s 失败: %v\n", order.OrderNo, err)
		}
	}

	return nil
}

// GetOrderStatusFlow 获取订单状态流转记录
func (ss *StatusService) GetOrderStatusFlow(orderID uint) ([]model.OrderStatusLog, error) {
	var statusLogs []model.OrderStatusLog
	if err := ss.db.Where("order_id = ?", orderID).
		Preload("Operator").
		Order("created_at ASC").
		Find(&statusLogs).Error; err != nil {
		return nil, fmt.Errorf("获取状态流转记录失败: %v", err)
	}

	return statusLogs, nil
}

// CanTransitionTo 检查订单是否可以转换到指定状态
func (ss *StatusService) CanTransitionTo(orderID uint, toStatus string) (bool, string) {
	var order model.Order
	if err := ss.db.First(&order, orderID).Error; err != nil {
		return false, "订单不存在"
	}

	if !ss.isValidTransition(order.Status, toStatus) {
		return false, fmt.Sprintf("不能从状态 %s 转换到 %s", order.Status, toStatus)
	}

	// 检查特定条件
	switch toStatus {
	case model.OrderStatusPaid:
		if order.PaidAmount.LessThan(order.PayableAmount) {
			return false, "订单未完成支付"
		}
	case model.OrderStatusCancelled:
		if order.Status != model.OrderStatusPending && order.Status != model.OrderStatusPaid {
			return false, "订单状态不允许取消"
		}
	case model.OrderStatusReceived:
		if order.Status != model.OrderStatusDelivered && order.Status != model.OrderStatusShipped {
			return false, "订单未配送完成"
		}
	}

	return true, ""
}

// GetStatusStatistics 获取订单状态统计
func (ss *StatusService) GetStatusStatistics() (map[string]int64, error) {
	var results []struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}

	if err := ss.db.Model(&model.Order{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("获取状态统计失败: %v", err)
	}

	statistics := make(map[string]int64)
	for _, result := range results {
		statistics[result.Status] = result.Count
	}

	return statistics, nil
}

// BatchUpdateOrderStatus 批量更新订单状态
func (ss *StatusService) BatchUpdateOrderStatus(orderIDs []uint, toStatus string, operatorID uint, operatorType string, reason, remark string) error {
	if len(orderIDs) == 0 {
		return fmt.Errorf("订单ID列表不能为空")
	}

	// 开始事务
	tx := ss.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	successCount := 0
	var errors []string

	for _, orderID := range orderIDs {
		if err := ss.UpdateOrderStatus(orderID, toStatus, operatorID, operatorType, reason, remark); err != nil {
			errors = append(errors, fmt.Sprintf("订单ID %d: %v", orderID, err))
		} else {
			successCount++
		}
	}

	if len(errors) > 0 {
		tx.Rollback()
		return fmt.Errorf("批量更新失败，成功：%d，失败：%d，错误：%v", successCount, len(errors), errors)
	}

	tx.Commit()
	return nil
}

// 全局订单状态服务实例
var globalStatusService *StatusService

// InitGlobalStatusService 初始化全局订单状态服务
func InitGlobalStatusService(db *gorm.DB) {
	globalStatusService = NewStatusService(db)
}

// GetGlobalStatusService 获取全局订单状态服务
func GetGlobalStatusService() *StatusService {
	return globalStatusService
}
