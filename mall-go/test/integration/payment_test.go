package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mall-go/internal/handler/payment"
	"mall-go/internal/model"
	paymentPkg "mall-go/pkg/payment"
	"mall-go/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// PaymentIntegrationTestSuite 支付集成测试套件
type PaymentIntegrationTestSuite struct {
	suite.Suite
	db             *gorm.DB
	router         *gin.Engine
	paymentService *paymentPkg.Service
	testUser       *model.User
	testOrder      *model.Order
}

// SetupSuite 设置测试套件
func (suite *PaymentIntegrationTestSuite) SetupSuite() {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 设置测试数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	// 自动迁移
	err = db.AutoMigrate(
		&model.User{},
		&model.Order{},
		&model.OrderItem{},
		&model.Payment{},
		&model.PaymentRefund{},
		&model.PaymentLog{},
		&model.PaymentConfig{},
	)
	suite.Require().NoError(err)

	suite.db = db

	// 创建支付服务
	config := paymentPkg.DefaultPaymentConfig()
	config.Alipay.Enabled = true
	config.Wechat.Enabled = true

	paymentService, err := paymentPkg.NewService(db, config)
	suite.Require().NoError(err)
	suite.paymentService = paymentService

	// 设置路由
	router := gin.New()
	api := router.Group("/api/v1")
	payment.RegisterRoutes(api, db, paymentService, nil, nil)
	suite.router = router

	// 创建测试数据
	suite.createTestData()
}

// TearDownSuite 清理测试套件
func (suite *PaymentIntegrationTestSuite) TearDownSuite() {
	// 清理资源
}

// createTestData 创建测试数据
func (suite *PaymentIntegrationTestSuite) createTestData() {
	// 创建测试用户
	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Phone:    "13800138000",
		Status:   "active",
	}
	suite.db.Create(user)
	suite.testUser = user

	// 创建测试订单
	order := &model.Order{
		OrderNo:       "TEST_ORDER_001",
		UserID:        user.ID,
		TotalAmount:   decimal.NewFromFloat(99.99),
		Status:        model.OrderStatusPending,
		PaymentStatus: "pending",
		PaymentMethod: "",
	}
	suite.db.Create(order)
	suite.testOrder = order

	// 创建支付方式配置
	configs := []model.PaymentConfig{
		{
			PaymentMethod: model.PaymentMethodAlipay,
			IsEnabled:     true,
			DisplayName:   "支付宝",
			DisplayOrder:  1,
			MinAmount:     decimal.NewFromFloat(0.01),
			MaxAmount:     decimal.NewFromFloat(50000),
		},
		{
			PaymentMethod: model.PaymentMethodWechat,
			IsEnabled:     true,
			DisplayName:   "微信支付",
			DisplayOrder:  2,
			MinAmount:     decimal.NewFromFloat(0.01),
			MaxAmount:     decimal.NewFromFloat(50000),
		},
	}

	for _, config := range configs {
		suite.db.Create(&config)
	}
}

// TestCreatePayment 测试创建支付
func (suite *PaymentIntegrationTestSuite) TestCreatePayment() {
	// 准备请求数据
	requestData := model.PaymentCreateRequest{
		OrderID:        suite.testOrder.ID,
		PaymentMethod:  model.PaymentMethodAlipay,
		Amount:         decimal.NewFromFloat(99.99),
		Subject:        "测试订单支付",
		Description:    "集成测试订单",
		ExpiredMinutes: 30,
	}

	jsonData, err := json.Marshal(requestData)
	suite.Require().NoError(err)

	// 发送请求
	req, _ := http.NewRequest("POST", "/api/v1/payments", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test_token") // 模拟认证

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// 验证响应
	suite.Equal(http.StatusOK, w.Code)

	var resp response.Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	suite.NoError(err)
	suite.Equal(200, resp.Code)
	suite.Equal("创建支付成功", resp.Message)

	// 验证响应数据
	responseData, ok := resp.Data.(map[string]interface{})
	suite.True(ok)
	suite.NotEmpty(responseData["payment_no"])
	suite.Equal(string(model.PaymentMethodAlipay), responseData["payment_method"])

	// 验证数据库中的支付记录
	var payment model.Payment
	err = suite.db.Where("order_id = ?", suite.testOrder.ID).First(&payment).Error
	suite.NoError(err)
	suite.Equal(model.PaymentMethodAlipay, payment.PaymentMethod)
	suite.Equal(model.PaymentStatusPaying, payment.PaymentStatus)
	suite.Equal(requestData.Amount, payment.Amount)
}

// TestQueryPayment 测试查询支付
func (suite *PaymentIntegrationTestSuite) TestQueryPayment() {
	// 先创建一个支付记录
	payment := &model.Payment{
		PaymentNo:     "TEST_PAY_001",
		OrderID:       suite.testOrder.ID,
		UserID:        suite.testUser.ID,
		PaymentMethod: model.PaymentMethodAlipay,
		PaymentStatus: model.PaymentStatusSuccess,
		Amount:        decimal.NewFromFloat(99.99),
		Subject:       "测试支付",
	}
	suite.db.Create(payment)

	// 测试通过支付ID查询
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/payments/%d", payment.ID), nil)
	req.Header.Set("Authorization", "Bearer test_token")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	suite.NoError(err)
	suite.Equal(200, resp.Code)

	// 验证响应数据
	responseData, ok := resp.Data.(map[string]interface{})
	suite.True(ok)
	suite.Equal(payment.PaymentNo, responseData["payment_no"])
	suite.Equal(string(payment.PaymentMethod), responseData["payment_method"])
}

// TestGetPaymentMethods 测试获取支付方式列表
func (suite *PaymentIntegrationTestSuite) TestGetPaymentMethods() {
	req, _ := http.NewRequest("GET", "/api/v1/payments/methods", nil)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	suite.NoError(err)
	suite.Equal(200, resp.Code)

	// 验证响应数据
	methods, ok := resp.Data.([]interface{})
	suite.True(ok)
	suite.Len(methods, 2) // 应该有2种支付方式

	// 验证第一个支付方式
	method1, ok := methods[0].(map[string]interface{})
	suite.True(ok)
	suite.Equal("支付宝", method1["display_name"])
	suite.Equal(true, method1["is_enabled"])
}

// TestPaymentWorkflow 测试完整的支付流程
func (suite *PaymentIntegrationTestSuite) TestPaymentWorkflow() {
	// 1. 创建支付
	createReq := model.PaymentCreateRequest{
		OrderID:        suite.testOrder.ID,
		PaymentMethod:  model.PaymentMethodWechat,
		Amount:         decimal.NewFromFloat(99.99),
		Subject:        "完整流程测试",
		ExpiredMinutes: 30,
	}

	jsonData, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/v1/payments", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test_token")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	suite.Equal(http.StatusOK, w.Code)

	var createResp response.Response
	json.Unmarshal(w.Body.Bytes(), &createResp)
	createData := createResp.Data.(map[string]interface{})
	paymentID := uint(createData["payment_id"].(float64))

	// 2. 查询支付状态
	req, _ = http.NewRequest("GET", fmt.Sprintf("/api/v1/payments/%d", paymentID), nil)
	req.Header.Set("Authorization", "Bearer test_token")

	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	suite.Equal(http.StatusOK, w.Code)

	var queryResp response.Response
	json.Unmarshal(w.Body.Bytes(), &queryResp)
	queryData := queryResp.Data.(map[string]interface{})
	suite.Equal("wechat", queryData["payment_method"])

	// 3. 模拟支付成功回调（这里简化处理，直接更新数据库）
	suite.db.Model(&model.Payment{}).Where("id = ?", paymentID).Updates(map[string]interface{}{
		"payment_status": model.PaymentStatusSuccess,
		"paid_at":        time.Now(),
	})

	// 4. 再次查询验证状态更新
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	suite.Equal(http.StatusOK, w.Code)

	json.Unmarshal(w.Body.Bytes(), &queryResp)
	queryData = queryResp.Data.(map[string]interface{})
	suite.Equal("success", queryData["payment_status"])

	// 5. 验证订单状态是否同步更新
	var updatedOrder model.Order
	suite.db.First(&updatedOrder, suite.testOrder.ID)
	// 注意：这里需要实际的同步机制才能验证订单状态更新
}

// TestRefundPayment 测试退款流程
func (suite *PaymentIntegrationTestSuite) TestRefundPayment() {
	// 先创建一个成功的支付记录
	payment := &model.Payment{
		PaymentNo:     "TEST_PAY_REFUND",
		OrderID:       suite.testOrder.ID,
		UserID:        suite.testUser.ID,
		PaymentMethod: model.PaymentMethodAlipay,
		PaymentStatus: model.PaymentStatusSuccess,
		Amount:        decimal.NewFromFloat(99.99),
		Subject:       "退款测试支付",
		PaidAt:        &time.Time{},
	}
	now := time.Now()
	payment.PaidAt = &now
	suite.db.Create(payment)

	// 申请退款
	refundReq := model.PaymentRefundRequest{
		PaymentID:    payment.ID,
		RefundAmount: decimal.NewFromFloat(50.00),
		RefundReason: "用户申请退款",
	}

	jsonData, _ := json.Marshal(refundReq)
	req, _ := http.NewRequest("POST", "/api/v1/payments/refund", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test_token")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	suite.NoError(err)
	suite.Equal(200, resp.Code)

	// 验证退款记录
	var refund model.PaymentRefund
	err = suite.db.Where("payment_id = ?", payment.ID).First(&refund).Error
	suite.NoError(err)
	suite.Equal(refundReq.RefundAmount, refund.RefundAmount)
	suite.Equal(refundReq.RefundReason, refund.RefundReason)
}

// TestPaymentValidation 测试支付验证
func (suite *PaymentIntegrationTestSuite) TestPaymentValidation() {
	tests := []struct {
		name           string
		request        model.PaymentCreateRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "无效的支付方式",
			request: model.PaymentCreateRequest{
				OrderID:       suite.testOrder.ID,
				PaymentMethod: "invalid_method",
				Amount:        decimal.NewFromFloat(99.99),
				Subject:       "测试",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "无效的支付方式",
		},
		{
			name: "金额不匹配",
			request: model.PaymentCreateRequest{
				OrderID:       suite.testOrder.ID,
				PaymentMethod: model.PaymentMethodAlipay,
				Amount:        decimal.NewFromFloat(199.99), // 与订单金额不匹配
				Subject:       "测试",
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "无效的金额",
		},
		{
			name: "订单不存在",
			request: model.PaymentCreateRequest{
				OrderID:       99999,
				PaymentMethod: model.PaymentMethodAlipay,
				Amount:        decimal.NewFromFloat(99.99),
				Subject:       "测试",
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			jsonData, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest("POST", "/api/v1/payments", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer test_token")

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			suite.Equal(tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var resp response.Response
				json.Unmarshal(w.Body.Bytes(), &resp)
				suite.Contains(resp.Message, tt.expectedError)
			}
		})
	}
}

// TestConcurrentPayments 测试并发支付
func (suite *PaymentIntegrationTestSuite) TestConcurrentPayments() {
	// 创建多个订单
	orders := make([]*model.Order, 5)
	for i := 0; i < 5; i++ {
		order := &model.Order{
			OrderNo:       fmt.Sprintf("CONCURRENT_ORDER_%d", i),
			UserID:        suite.testUser.ID,
			TotalAmount:   decimal.NewFromFloat(100.00),
			Status:        model.OrderStatusPending,
			PaymentStatus: "pending",
		}
		suite.db.Create(order)
		orders[i] = order
	}

	// 并发创建支付
	done := make(chan bool, 5)
	for i, order := range orders {
		go func(orderID uint, index int) {
			defer func() { done <- true }()

			request := model.PaymentCreateRequest{
				OrderID:       orderID,
				PaymentMethod: model.PaymentMethodAlipay,
				Amount:        decimal.NewFromFloat(100.00),
				Subject:       fmt.Sprintf("并发测试订单 %d", index),
			}

			jsonData, _ := json.Marshal(request)
			req, _ := http.NewRequest("POST", "/api/v1/payments", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer test_token")

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			// 所有请求都应该成功
			suite.Equal(http.StatusOK, w.Code)
		}(order.ID, i)
	}

	// 等待所有协程完成
	for i := 0; i < 5; i++ {
		<-done
	}

	// 验证所有支付记录都已创建
	var count int64
	suite.db.Model(&model.Payment{}).Where("user_id = ?", suite.testUser.ID).Count(&count)
	suite.GreaterOrEqual(count, int64(5))
}

// 运行集成测试套件
func TestPaymentIntegrationSuite(t *testing.T) {
	suite.Run(t, new(PaymentIntegrationTestSuite))
}
