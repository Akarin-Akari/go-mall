package order

import (
	"encoding/json"
	"fmt"
	"time"

	"mall-go/internal/model"

	"gorm.io/gorm"
)

// ShippingService 订单物流跟踪服务
type ShippingService struct {
	db            *gorm.DB
	statusService *StatusService
}

// NewShippingService 创建订单物流跟踪服务
func NewShippingService(db *gorm.DB, statusService *StatusService) *ShippingService {
	return &ShippingService{
		db:            db,
		statusService: statusService,
	}
}

// ShippingCompany 物流公司信息
type ShippingCompany struct {
	Code    string `json:"code"`
	Name    string `json:"name"`
	Website string `json:"website"`
	Phone   string `json:"phone"`
}

// TrackingInfo 物流轨迹信息
type TrackingInfo struct {
	Time        time.Time `json:"time"`
	Status      string    `json:"status"`
	Location    string    `json:"location"`
	Description string    `json:"description"`
	Operator    string    `json:"operator"`
}

// ShippingRequest 发货请求
type ShippingRequest struct {
	OrderID         uint   `json:"order_id" binding:"required"`
	ShippingCompany string `json:"shipping_company" binding:"required"`
	TrackingNumber  string `json:"tracking_number" binding:"required"`
	ShippingMethod  string `json:"shipping_method"`
	SenderName      string `json:"sender_name"`
	SenderPhone     string `json:"sender_phone"`
	SenderAddress   string `json:"sender_address"`
	EstimatedDays   int    `json:"estimated_days"`
}

// ShippingResponse 发货响应
type ShippingResponse struct {
	ShipmentNo      string           `json:"shipment_no"`
	ShippingCompany string           `json:"shipping_company"`
	TrackingNumber  string           `json:"tracking_number"`
	Status          string           `json:"status"`
	ShipTime        *time.Time       `json:"ship_time"`
	EstimatedArrival *time.Time      `json:"estimated_arrival"`
	TrackingInfo    []TrackingInfo   `json:"tracking_info"`
}

// CreateShipment 创建发货记录
func (ss *ShippingService) CreateShipment(req *ShippingRequest) (*ShippingResponse, error) {
	// 开始事务
	tx := ss.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取订单
	var order model.Order
	if err := tx.First(&order, req.OrderID).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("订单不存在")
	}

	// 检查订单状态
	if !order.CanShip() {
		tx.Rollback()
		return nil, fmt.Errorf("订单状态不允许发货")
	}

	// 检查是否已经发货
	var existingShipment model.OrderShipment
	if err := tx.Where("order_id = ?", req.OrderID).First(&existingShipment).Error; err == nil {
		tx.Rollback()
		return nil, fmt.Errorf("订单已发货")
	}

	// 创建发货记录
	now := time.Now()
	shipment := &model.OrderShipment{
		OrderID:         req.OrderID,
		ShipmentNo:      ss.generateShipmentNo(),
		ShippingCompany: req.ShippingCompany,
		TrackingNumber:  req.TrackingNumber,
		Status:          model.ShippingStatusShipped,
		SenderName:      req.SenderName,
		SenderPhone:     req.SenderPhone,
		SenderAddress:   req.SenderAddress,
		ReceiverName:    order.ReceiverName,
		ReceiverPhone:   order.ReceiverPhone,
		ReceiverAddress: order.ReceiverAddress,
		ShipTime:        &now,
	}

	// 计算预计到达时间
	if req.EstimatedDays > 0 {
		estimatedArrival := now.Add(time.Duration(req.EstimatedDays) * 24 * time.Hour)
		shipment.DeliveryTime = &estimatedArrival
	}

	// 初始化物流轨迹
	initialTracking := []TrackingInfo{
		{
			Time:        now,
			Status:      "shipped",
			Location:    req.SenderAddress,
			Description: "商品已发货",
			Operator:    req.SenderName,
		},
	}
	trackingDataJSON, _ := json.Marshal(initialTracking)
	shipment.TrackingData = string(trackingDataJSON)

	if err := tx.Create(shipment).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建发货记录失败: %v", err)
	}

	// 更新订单信息
	orderUpdates := map[string]interface{}{
		"shipping_company":  req.ShippingCompany,
		"tracking_number":   req.TrackingNumber,
		"shipping_status":   model.ShippingStatusShipped,
		"shipping_method":   req.ShippingMethod,
		"ship_time":         &now,
	}

	if shipment.DeliveryTime != nil {
		orderUpdates["estimated_arrival"] = shipment.DeliveryTime
	}

	if err := tx.Model(&order).Updates(orderUpdates).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("更新订单发货信息失败: %v", err)
	}

	// 更新订单状态为已发货
	if err := ss.statusService.UpdateOrderStatus(req.OrderID, model.OrderStatusShipped, 
		0, model.OperatorTypeAdmin, "商品发货", "管理员发货操作"); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("更新订单状态失败: %v", err)
	}

	tx.Commit()

	return &ShippingResponse{
		ShipmentNo:       shipment.ShipmentNo,
		ShippingCompany:  shipment.ShippingCompany,
		TrackingNumber:   shipment.TrackingNumber,
		Status:           shipment.Status,
		ShipTime:         shipment.ShipTime,
		EstimatedArrival: shipment.DeliveryTime,
		TrackingInfo:     initialTracking,
	}, nil
}

// UpdateShippingStatus 更新物流状态
func (ss *ShippingService) UpdateShippingStatus(shipmentNo, status, location, description string) error {
	// 开始事务
	tx := ss.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取发货记录
	var shipment model.OrderShipment
	if err := tx.Preload("Order").Where("shipment_no = ?", shipmentNo).First(&shipment).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("发货记录不存在")
	}

	// 解析现有物流轨迹
	var trackingInfo []TrackingInfo
	if shipment.TrackingData != "" {
		json.Unmarshal([]byte(shipment.TrackingData), &trackingInfo)
	}

	// 添加新的物流轨迹
	newTracking := TrackingInfo{
		Time:        time.Now(),
		Status:      status,
		Location:    location,
		Description: description,
		Operator:    "物流公司",
	}
	trackingInfo = append(trackingInfo, newTracking)

	// 更新物流轨迹数据
	trackingDataJSON, _ := json.Marshal(trackingInfo)
	updates := map[string]interface{}{
		"status":        status,
		"tracking_data": string(trackingDataJSON),
	}

	// 根据状态更新时间字段
	now := time.Now()
	switch status {
	case model.ShippingStatusDelivered:
		updates["delivery_time"] = &now
	case model.ShippingStatusReceived:
		updates["receive_time"] = &now
	}

	if err := tx.Model(&shipment).Updates(updates).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新物流状态失败: %v", err)
	}

	// 更新订单物流状态
	if err := tx.Model(&shipment.Order).Update("shipping_status", status).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新订单物流状态失败: %v", err)
	}

	// 根据物流状态更新订单状态
	var orderStatus string
	var reason string
	switch status {
	case model.ShippingStatusDelivered:
		orderStatus = model.OrderStatusDelivered
		reason = "商品已配送"
		tx.Model(&shipment.Order).Update("delivery_time", &now)
	case model.ShippingStatusReceived:
		orderStatus = model.OrderStatusReceived
		reason = "用户已签收"
		tx.Model(&shipment.Order).Update("receive_time", &now)
	}

	if orderStatus != "" {
		if err := ss.statusService.UpdateOrderStatus(shipment.OrderID, orderStatus, 
			0, model.OperatorTypeSystem, reason, "物流状态自动更新"); err != nil {
			tx.Rollback()
			return fmt.Errorf("更新订单状态失败: %v", err)
		}
	}

	tx.Commit()
	return nil
}

// GetShippingInfo 获取物流信息
func (ss *ShippingService) GetShippingInfo(orderID uint) (*ShippingResponse, error) {
	var shipment model.OrderShipment
	if err := ss.db.Where("order_id = ?", orderID).First(&shipment).Error; err != nil {
		return nil, fmt.Errorf("物流信息不存在")
	}

	// 解析物流轨迹
	var trackingInfo []TrackingInfo
	if shipment.TrackingData != "" {
		json.Unmarshal([]byte(shipment.TrackingData), &trackingInfo)
	}

	return &ShippingResponse{
		ShipmentNo:       shipment.ShipmentNo,
		ShippingCompany:  shipment.ShippingCompany,
		TrackingNumber:   shipment.TrackingNumber,
		Status:           shipment.Status,
		ShipTime:         shipment.ShipTime,
		EstimatedArrival: shipment.DeliveryTime,
		TrackingInfo:     trackingInfo,
	}, nil
}

// TrackShipment 跟踪物流
func (ss *ShippingService) TrackShipment(trackingNumber, shippingCompany string) (*ShippingResponse, error) {
	// 从数据库获取物流信息
	var shipment model.OrderShipment
	if err := ss.db.Where("tracking_number = ? AND shipping_company = ?", 
		trackingNumber, shippingCompany).First(&shipment).Error; err != nil {
		return nil, fmt.Errorf("物流信息不存在")
	}

	// 调用第三方物流接口获取最新信息
	latestTracking, err := ss.queryThirdPartyTracking(trackingNumber, shippingCompany)
	if err != nil {
		// 如果第三方接口调用失败，返回数据库中的信息
		return ss.GetShippingInfo(shipment.OrderID)
	}

	// 更新数据库中的物流信息
	if len(latestTracking) > 0 {
		trackingDataJSON, _ := json.Marshal(latestTracking)
		ss.db.Model(&shipment).Update("tracking_data", string(trackingDataJSON))

		// 检查是否有状态变化
		latestStatus := latestTracking[len(latestTracking)-1].Status
		if latestStatus != shipment.Status {
			ss.UpdateShippingStatus(shipment.ShipmentNo, latestStatus, 
				latestTracking[len(latestTracking)-1].Location,
				latestTracking[len(latestTracking)-1].Description)
		}
	}

	return &ShippingResponse{
		ShipmentNo:       shipment.ShipmentNo,
		ShippingCompany:  shipment.ShippingCompany,
		TrackingNumber:   shipment.TrackingNumber,
		Status:           shipment.Status,
		ShipTime:         shipment.ShipTime,
		EstimatedArrival: shipment.DeliveryTime,
		TrackingInfo:     latestTracking,
	}, nil
}

// queryThirdPartyTracking 查询第三方物流接口
func (ss *ShippingService) queryThirdPartyTracking(trackingNumber, shippingCompany string) ([]TrackingInfo, error) {
	// 这里应该调用真实的第三方物流查询接口
	// 为了演示，返回模拟数据
	
	mockTracking := []TrackingInfo{
		{
			Time:        time.Now().Add(-2 * 24 * time.Hour),
			Status:      "shipped",
			Location:    "深圳市",
			Description: "商品已发货",
			Operator:    "发货仓库",
		},
		{
			Time:        time.Now().Add(-1 * 24 * time.Hour),
			Status:      "transit",
			Location:    "广州市",
			Description: "商品运输中",
			Operator:    "广州转运中心",
		},
		{
			Time:        time.Now().Add(-12 * time.Hour),
			Status:      "transit",
			Location:    "上海市",
			Description: "商品到达上海转运中心",
			Operator:    "上海转运中心",
		},
		{
			Time:        time.Now().Add(-2 * time.Hour),
			Status:      "delivered",
			Location:    "上海市浦东新区",
			Description: "商品正在派送中",
			Operator:    "派送员张三",
		},
	}

	return mockTracking, nil
}

// GetShippingCompanies 获取支持的物流公司列表
func (ss *ShippingService) GetShippingCompanies() []ShippingCompany {
	return []ShippingCompany{
		{
			Code:    "SF",
			Name:    "顺丰速运",
			Website: "https://www.sf-express.com",
			Phone:   "95338",
		},
		{
			Code:    "YTO",
			Name:    "圆通速递",
			Website: "https://www.yto.net.cn",
			Phone:   "95554",
		},
		{
			Code:    "ZTO",
			Name:    "中通快递",
			Website: "https://www.zto.com",
			Phone:   "95311",
		},
		{
			Code:    "STO",
			Name:    "申通快递",
			Website: "https://www.sto.cn",
			Phone:   "95543",
		},
		{
			Code:    "YD",
			Name:    "韵达速递",
			Website: "https://www.yunda.com",
			Phone:   "95546",
		},
		{
			Code:    "HTKY",
			Name:    "百世快递",
			Website: "https://www.800best.com",
			Phone:   "400-885-6561",
		},
		{
			Code:    "EMS",
			Name:    "中国邮政",
			Website: "https://www.ems.com.cn",
			Phone:   "11183",
		},
		{
			Code:    "JD",
			Name:    "京东物流",
			Website: "https://www.jdl.com",
			Phone:   "950616",
		},
	}
}

// ConfirmReceipt 确认收货
func (ss *ShippingService) ConfirmReceipt(orderID uint, userID uint) error {
	// 获取订单
	var order model.Order
	if err := ss.db.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
		return fmt.Errorf("订单不存在")
	}

	// 检查订单状态
	if !order.CanReceive() {
		return fmt.Errorf("订单状态不允许确认收货")
	}

	// 更新物流状态
	var shipment model.OrderShipment
	if err := ss.db.Where("order_id = ?", orderID).First(&shipment).Error; err == nil {
		ss.UpdateShippingStatus(shipment.ShipmentNo, model.ShippingStatusReceived, 
			order.ReceiverAddress, "用户确认收货")
	}

	return nil
}

// GetShippingStatistics 获取物流统计信息
func (ss *ShippingService) GetShippingStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 统计各物流状态的订单数量
	var statusStats []struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}

	if err := ss.db.Model(&model.OrderShipment{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusStats).Error; err != nil {
		return nil, fmt.Errorf("获取物流状态统计失败: %v", err)
	}

	stats["status_stats"] = statusStats

	// 统计各物流公司的使用情况
	var companyStats []struct {
		ShippingCompany string `json:"shipping_company"`
		Count           int64  `json:"count"`
	}

	if err := ss.db.Model(&model.OrderShipment{}).
		Select("shipping_company, COUNT(*) as count").
		Group("shipping_company").
		Order("count DESC").
		Limit(10).
		Scan(&companyStats).Error; err != nil {
		return nil, fmt.Errorf("获取物流公司统计失败: %v", err)
	}

	stats["company_stats"] = companyStats

	// 统计今日发货数量
	var todayShipments int64
	today := time.Now().Format("2006-01-02")
	if err := ss.db.Model(&model.OrderShipment{}).
		Where("DATE(ship_time) = ?", today).
		Count(&todayShipments).Error; err != nil {
		return nil, fmt.Errorf("获取今日发货统计失败: %v", err)
	}

	stats["today_shipments"] = todayShipments

	return stats, nil
}

// generateShipmentNo 生成发货单号
func (ss *ShippingService) generateShipmentNo() string {
	return fmt.Sprintf("SHIP%d", time.Now().UnixNano())
}

// AutoUpdateShippingStatus 自动更新物流状态
func (ss *ShippingService) AutoUpdateShippingStatus() error {
	// 获取所有运输中的物流记录
	var shipments []model.OrderShipment
	if err := ss.db.Where("status IN ?", []string{
		model.ShippingStatusShipped,
		model.ShippingStatusTransit,
	}).Find(&shipments).Error; err != nil {
		return fmt.Errorf("获取物流记录失败: %v", err)
	}

	for _, shipment := range shipments {
		// 查询最新物流信息
		latestTracking, err := ss.queryThirdPartyTracking(shipment.TrackingNumber, shipment.ShippingCompany)
		if err != nil {
			continue // 跳过查询失败的记录
		}

		if len(latestTracking) > 0 {
			latestStatus := latestTracking[len(latestTracking)-1].Status
			if latestStatus != shipment.Status {
				// 更新物流状态
				ss.UpdateShippingStatus(shipment.ShipmentNo, latestStatus,
					latestTracking[len(latestTracking)-1].Location,
					latestTracking[len(latestTracking)-1].Description)
			}
		}
	}

	return nil
}

// 全局订单物流服务实例
var globalShippingService *ShippingService

// InitGlobalShippingService 初始化全局订单物流服务
func InitGlobalShippingService(db *gorm.DB, statusService *StatusService) {
	globalShippingService = NewShippingService(db, statusService)
}

// GetGlobalShippingService 获取全局订单物流服务
func GetGlobalShippingService() *ShippingService {
	return globalShippingService
}
