package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"mall-go/internal/config"
	"mall-go/internal/model"
	"mall-go/pkg/database"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// OrderMigration 订单数据库迁移
type OrderMigration struct {
	db *gorm.DB
}

// NewOrderMigration 创建订单迁移实例
func NewOrderMigration(db *gorm.DB) *OrderMigration {
	return &OrderMigration{db: db}
}

// Run 执行订单相关表的迁移
func (om *OrderMigration) Run() error {
	fmt.Println("🚀 开始执行订单数据库迁移...")

	// 1. 创建订单相关表
	if err := om.createOrderTables(); err != nil {
		return fmt.Errorf("创建订单表失败: %v", err)
	}

	// 2. 创建索引
	if err := om.createIndexes(); err != nil {
		return fmt.Errorf("创建索引失败: %v", err)
	}

	// 3. 插入初始数据
	if err := om.insertInitialData(); err != nil {
		return fmt.Errorf("插入初始数据失败: %v", err)
	}

	// 4. 创建触发器和存储过程
	if err := om.createTriggersAndProcedures(); err != nil {
		return fmt.Errorf("创建触发器和存储过程失败: %v", err)
	}

	fmt.Println("✅ 订单数据库迁移完成！")
	return nil
}

// createOrderTables 创建订单相关表
func (om *OrderMigration) createOrderTables() error {
	fmt.Println("📋 创建订单相关表...")

	// 订单相关表列表
	tables := []interface{}{
		&model.Order{},
		&model.OrderItem{},
		&model.OrderStatusLog{},
		&model.OrderPayment{},
		&model.OrderShipment{},
		&model.OrderAfterSale{},
	}

	// 自动迁移表结构
	for _, table := range tables {
		if err := om.db.AutoMigrate(table); err != nil {
			return fmt.Errorf("迁移表 %T 失败: %v", table, err)
		}
		fmt.Printf("  ✓ 表 %T 创建成功\n", table)
	}

	return nil
}

// createIndexes 创建索引
func (om *OrderMigration) createIndexes() error {
	fmt.Println("🔍 创建数据库索引...")

	// 订单表索引
	orderIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status)",
		"CREATE INDEX IF NOT EXISTS idx_orders_order_time ON orders(order_time)",
		"CREATE INDEX IF NOT EXISTS idx_orders_order_no ON orders(order_no)",
		"CREATE INDEX IF NOT EXISTS idx_orders_pay_time ON orders(pay_time)",
		"CREATE INDEX IF NOT EXISTS idx_orders_ship_time ON orders(ship_time)",
		"CREATE INDEX IF NOT EXISTS idx_orders_receive_time ON orders(receive_time)",
		"CREATE INDEX IF NOT EXISTS idx_orders_finish_time ON orders(finish_time)",
		"CREATE INDEX IF NOT EXISTS idx_orders_cancel_time ON orders(cancel_time)",
		"CREATE INDEX IF NOT EXISTS idx_orders_refund_status ON orders(refund_status)",
		"CREATE INDEX IF NOT EXISTS idx_orders_province_city ON orders(province, city)",
	}

	// 订单商品表索引
	orderItemIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id)",
		"CREATE INDEX IF NOT EXISTS idx_order_items_product_id ON order_items(product_id)",
		"CREATE INDEX IF NOT EXISTS idx_order_items_sku_id ON order_items(sku_id)",
		"CREATE INDEX IF NOT EXISTS idx_order_items_refund_status ON order_items(refund_status)",
	}

	// 订单状态日志表索引
	statusLogIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_order_status_logs_order_id ON order_status_logs(order_id)",
		"CREATE INDEX IF NOT EXISTS idx_order_status_logs_operator ON order_status_logs(operator_id, operator_type)",
		"CREATE INDEX IF NOT EXISTS idx_order_status_logs_status ON order_status_logs(from_status, to_status)",
		"CREATE INDEX IF NOT EXISTS idx_order_status_logs_created_at ON order_status_logs(created_at)",
	}

	// 订单支付表索引
	paymentIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_order_payments_order_id ON order_payments(order_id)",
		"CREATE INDEX IF NOT EXISTS idx_order_payments_payment_no ON order_payments(payment_no)",
		"CREATE INDEX IF NOT EXISTS idx_order_payments_method ON order_payments(payment_method)",
		"CREATE INDEX IF NOT EXISTS idx_order_payments_status ON order_payments(status)",
		"CREATE INDEX IF NOT EXISTS idx_order_payments_third_party_no ON order_payments(third_party_no)",
		"CREATE INDEX IF NOT EXISTS idx_order_payments_pay_time ON order_payments(pay_time)",
	}

	// 订单物流表索引
	shipmentIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_order_shipments_order_id ON order_shipments(order_id)",
		"CREATE INDEX IF NOT EXISTS idx_order_shipments_shipment_no ON order_shipments(shipment_no)",
		"CREATE INDEX IF NOT EXISTS idx_order_shipments_tracking_number ON order_shipments(tracking_number)",
		"CREATE INDEX IF NOT EXISTS idx_order_shipments_company ON order_shipments(shipping_company)",
		"CREATE INDEX IF NOT EXISTS idx_order_shipments_status ON order_shipments(status)",
		"CREATE INDEX IF NOT EXISTS idx_order_shipments_ship_time ON order_shipments(ship_time)",
	}

	// 订单售后表索引
	afterSaleIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_order_after_sales_order_id ON order_after_sales(order_id)",
		"CREATE INDEX IF NOT EXISTS idx_order_after_sales_order_item_id ON order_after_sales(order_item_id)",
		"CREATE INDEX IF NOT EXISTS idx_order_after_sales_after_sale_no ON order_after_sales(after_sale_no)",
		"CREATE INDEX IF NOT EXISTS idx_order_after_sales_type ON order_after_sales(type)",
		"CREATE INDEX IF NOT EXISTS idx_order_after_sales_status ON order_after_sales(status)",
		"CREATE INDEX IF NOT EXISTS idx_order_after_sales_apply_user ON order_after_sales(apply_user_id)",
		"CREATE INDEX IF NOT EXISTS idx_order_after_sales_handle_user ON order_after_sales(handle_user_id)",
		"CREATE INDEX IF NOT EXISTS idx_order_after_sales_created_at ON order_after_sales(created_at)",
	}

	// 执行所有索引创建语句
	allIndexes := append(orderIndexes, orderItemIndexes...)
	allIndexes = append(allIndexes, statusLogIndexes...)
	allIndexes = append(allIndexes, paymentIndexes...)
	allIndexes = append(allIndexes, shipmentIndexes...)
	allIndexes = append(allIndexes, afterSaleIndexes...)

	for _, indexSQL := range allIndexes {
		if err := om.db.Exec(indexSQL).Error; err != nil {
			fmt.Printf("  ⚠️  创建索引失败: %s - %v\n", indexSQL, err)
		} else {
			fmt.Printf("  ✓ 索引创建成功\n")
		}
	}

	return nil
}

// insertInitialData 插入初始数据
func (om *OrderMigration) insertInitialData() error {
	fmt.Println("📊 插入初始数据...")

	// 检查是否已有数据
	var count int64
	om.db.Model(&model.Order{}).Count(&count)
	if count > 0 {
		fmt.Println("  ℹ️  订单表已有数据，跳过初始数据插入")
		return nil
	}

	// 插入示例订单数据（仅用于开发环境）
	if config.GlobalConfig.Server.Mode == "development" {
		if err := om.insertSampleData(); err != nil {
			return fmt.Errorf("插入示例数据失败: %v", err)
		}
	}

	return nil
}

// insertSampleData 插入示例数据
func (om *OrderMigration) insertSampleData() error {
	fmt.Println("  📝 插入示例订单数据...")

	// 示例订单数据
	sampleOrders := []model.Order{
		{
			OrderNo:         "ORD202501200001",
			UserID:          1,
			Status:          model.OrderStatusCompleted,
			OrderType:       model.OrderTypeNormal,
			TotalAmount:     decimal.NewFromFloat(199.99),
			PayableAmount:   decimal.NewFromFloat(199.99),
			PaidAmount:      decimal.NewFromFloat(199.99),
			ReceiverName:    "张三",
			ReceiverPhone:   "13800138000",
			ReceiverAddress: "北京市朝阳区xxx街道xxx号",
			Province:        "北京市",
			City:            "北京市",
			District:        "朝阳区",
			OrderTime:       time.Now().Add(-7 * 24 * time.Hour),
			PayTime:         &[]time.Time{time.Now().Add(-7 * 24 * time.Hour)}[0],
			ShipTime:        &[]time.Time{time.Now().Add(-6 * 24 * time.Hour)}[0],
			ReceiveTime:     &[]time.Time{time.Now().Add(-3 * 24 * time.Hour)}[0],
			FinishTime:      &[]time.Time{time.Now().Add(-1 * 24 * time.Hour)}[0],
			RefundStatus:    model.RefundStatusNone,
		},
		{
			OrderNo:         "ORD202501200002",
			UserID:          2,
			Status:          model.OrderStatusShipped,
			OrderType:       model.OrderTypeNormal,
			TotalAmount:     decimal.NewFromFloat(299.99),
			PayableAmount:   decimal.NewFromFloat(299.99),
			PaidAmount:      decimal.NewFromFloat(299.99),
			ReceiverName:    "李四",
			ReceiverPhone:   "13900139000",
			ReceiverAddress: "上海市浦东新区xxx路xxx号",
			Province:        "上海市",
			City:            "上海市",
			District:        "浦东新区",
			OrderTime:       time.Now().Add(-2 * 24 * time.Hour),
			PayTime:         &[]time.Time{time.Now().Add(-2 * 24 * time.Hour)}[0],
			ShipTime:        &[]time.Time{time.Now().Add(-1 * 24 * time.Hour)}[0],
			RefundStatus:    model.RefundStatusNone,
		},
	}

	// 批量插入订单
	if err := om.db.Create(&sampleOrders).Error; err != nil {
		return fmt.Errorf("插入示例订单失败: %v", err)
	}

	fmt.Printf("  ✓ 插入 %d 条示例订单数据\n", len(sampleOrders))
	return nil
}

// createTriggersAndProcedures 创建触发器和存储过程
func (om *OrderMigration) createTriggersAndProcedures() error {
	fmt.Println("⚙️  创建触发器和存储过程...")

	// 订单状态变更触发器
	orderStatusTrigger := `
		CREATE TRIGGER IF NOT EXISTS tr_order_status_update
		AFTER UPDATE ON orders
		FOR EACH ROW
		BEGIN
			IF OLD.status != NEW.status THEN
				INSERT INTO order_status_logs (
					order_id, from_status, to_status, operator_type, 
					reason, remark, created_at
				) VALUES (
					NEW.id, OLD.status, NEW.status, 'system',
					'状态自动更新', '触发器自动记录', NOW()
				);
			END IF;
		END;
	`

	// 订单金额统计存储过程
	orderStatsProc := `
		CREATE PROCEDURE IF NOT EXISTS sp_get_order_statistics(
			IN start_date DATE,
			IN end_date DATE,
			OUT total_orders INT,
			OUT total_amount DECIMAL(10,2),
			OUT avg_amount DECIMAL(10,2)
		)
		BEGIN
			SELECT 
				COUNT(*),
				COALESCE(SUM(total_amount), 0),
				COALESCE(AVG(total_amount), 0)
			INTO total_orders, total_amount, avg_amount
			FROM orders 
			WHERE DATE(order_time) BETWEEN start_date AND end_date
			AND status != 'cancelled';
		END;
	`

	// 订单自动取消存储过程
	autoCancelProc := `
		CREATE PROCEDURE IF NOT EXISTS sp_auto_cancel_expired_orders()
		BEGIN
			UPDATE orders 
			SET status = 'cancelled', 
				cancel_time = NOW(),
				updated_at = NOW()
			WHERE status = 'pending' 
			AND pay_expire_time < NOW()
			AND pay_expire_time IS NOT NULL;
		END;
	`

	// 执行SQL语句
	procedures := []string{
		orderStatusTrigger,
		orderStatsProc,
		autoCancelProc,
	}

	for _, proc := range procedures {
		if err := om.db.Exec(proc).Error; err != nil {
			fmt.Printf("  ⚠️  创建存储过程/触发器失败: %v\n", err)
		} else {
			fmt.Printf("  ✓ 存储过程/触发器创建成功\n")
		}
	}

	return nil
}

// Rollback 回滚订单相关表
func (om *OrderMigration) Rollback() error {
	fmt.Println("🔄 开始回滚订单数据库迁移...")

	// 删除触发器和存储过程
	dropStatements := []string{
		"DROP TRIGGER IF EXISTS tr_order_status_update",
		"DROP PROCEDURE IF EXISTS sp_get_order_statistics",
		"DROP PROCEDURE IF EXISTS sp_auto_cancel_expired_orders",
	}

	for _, stmt := range dropStatements {
		if err := om.db.Exec(stmt).Error; err != nil {
			fmt.Printf("  ⚠️  删除失败: %s - %v\n", stmt, err)
		}
	}

	// 删除表（注意顺序，先删除有外键依赖的表）
	tables := []string{
		"order_after_sales",
		"order_shipments",
		"order_payments",
		"order_status_logs",
		"order_items",
		"orders",
	}

	for _, table := range tables {
		if err := om.db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table)).Error; err != nil {
			fmt.Printf("  ⚠️  删除表 %s 失败: %v\n", table, err)
		} else {
			fmt.Printf("  ✓ 表 %s 删除成功\n", table)
		}
	}

	fmt.Println("✅ 订单数据库回滚完成！")
	return nil
}

func main() {
	// 初始化配置
	config.Init()

	// 初始化数据库连接
	db := database.Init()

	// 创建迁移实例
	migration := NewOrderMigration(db)

	// 检查命令行参数
	if len(os.Args) > 1 && os.Args[1] == "rollback" {
		// 执行回滚
		if err := migration.Rollback(); err != nil {
			log.Fatalf("回滚失败: %v", err)
		}
	} else {
		// 执行迁移
		if err := migration.Run(); err != nil {
			log.Fatalf("迁移失败: %v", err)
		}
	}
}
