package payment

import (
	"fmt"
	"testing"
	"time"

	"mall-go/internal/model"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockAlipayClient 模拟支付宝客户端
type MockAlipayClient struct {
	mock.Mock
}

func (m *MockAlipayClient) CreatePayment(req interface{}) (interface{}, error) {
	args := m.Called(req)
	return args.Get(0), args.Error(1)
}

func (m *MockAlipayClient) QueryPayment(outTradeNo string) (interface{}, error) {
	args := m.Called(outTradeNo)
	return args.Get(0), args.Error(1)
}

func (m *MockAlipayClient) VerifyCallback(params map[string]string) error {
	args := m.Called(params)
	return args.Error(0)
}

// MockWechatClient 模拟微信支付客户端
type MockWechatClient struct {
	mock.Mock
}

func (m *MockWechatClient) CreatePayment(req interface{}) (interface{}, error) {
	args := m.Called(req)
	return args.Get(0), args.Error(1)
}

func (m *MockWechatClient) QueryPayment(outTradeNo string) (interface{}, error) {
	args := m.Called(outTradeNo)
	return args.Get(0), args.Error(1)
}

func (m *MockWechatClient) VerifyCallback(params map[string]string) error {
	args := m.Called(params)
	return args.Error(0)
}

func (m *MockWechatClient) ParseCallback(data []byte) (interface{}, error) {
	args := m.Called(data)
	return args.Get(0), args.Error(1)
}

// setupTestDB 设置测试数据库
func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 自动迁移表结构
	err = db.AutoMigrate(
		&model.User{},
		&model.Order{},
		&model.Payment{},
		&model.PaymentRefund{},
		&model.PaymentLog{},
		&model.PaymentConfig{},
	)
	if err != nil {
		panic("failed to migrate database")
	}

	return db
}

// createTestData 创建测试数据
func createTestData(db *gorm.DB) (*model.User, *model.Order) {
	// 创建测试用户
	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Status:   "active",
	}
	db.Create(user)

	// 创建测试订单
	order := &model.Order{
		OrderNo:       "ORDER123456",
		UserID:        user.ID,
		TotalAmount:   decimal.NewFromFloat(100.00),
		Status:        model.OrderStatusPending,
		PaymentStatus: model.PaymentStatusPending,
	}
	db.Create(order)

	return user, order
}

func TestService_CreatePayment(t *testing.T) {
	db := setupTestDB()
	user, order := createTestData(db)

	// 创建配置
	config := DefaultPaymentConfig()
	config.Alipay.Enabled = true

	// 创建服务
	service, err := NewService(db, config)
	assert.NoError(t, err)

	tests := []struct {
		name    string
		request *model.PaymentCreateRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "成功创建支付宝支付",
			request: &model.PaymentCreateRequest{
				OrderID:        order.ID,
				PaymentMethod:  model.PaymentMethodAlipay,
				Amount:         decimal.NewFromFloat(100.00),
				Subject:        "测试订单",
				Description:    "测试订单描述",
				ExpiredMinutes: 30,
			},
			wantErr: false,
		},
		{
			name: "无效的支付方式",
			request: &model.PaymentCreateRequest{
				OrderID:        order.ID,
				PaymentMethod:  "invalid_method",
				Amount:         decimal.NewFromFloat(100.00),
				Subject:        "测试订单",
				ExpiredMinutes: 30,
			},
			wantErr: true,
			errMsg:  "无效的支付方式",
		},
		{
			name: "金额不匹配",
			request: &model.PaymentCreateRequest{
				OrderID:        order.ID,
				PaymentMethod:  model.PaymentMethodAlipay,
				Amount:         decimal.NewFromFloat(200.00), // 与订单金额不匹配
				Subject:        "测试订单",
				ExpiredMinutes: 30,
			},
			wantErr: true,
			errMsg:  "无效的金额",
		},
		{
			name: "订单不存在",
			request: &model.PaymentCreateRequest{
				OrderID:        99999, // 不存在的订单ID
				PaymentMethod:  model.PaymentMethodAlipay,
				Amount:         decimal.NewFromFloat(100.00),
				Subject:        "测试订单",
				ExpiredMinutes: 30,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := service.CreatePayment(tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.request.OrderID, resp.PaymentID)
				assert.Equal(t, tt.request.PaymentMethod, resp.PaymentMethod)
				assert.Equal(t, tt.request.Amount, resp.Amount)
			}
		})
	}
}

func TestService_QueryPayment(t *testing.T) {
	db := setupTestDB()
	user, order := createTestData(db)

	// 创建测试支付记录
	payment := &model.Payment{
		PaymentNo:     "PAY123456",
		OrderID:       order.ID,
		UserID:        user.ID,
		PaymentMethod: model.PaymentMethodAlipay,
		PaymentStatus: model.PaymentStatusSuccess,
		Amount:        decimal.NewFromFloat(100.00),
		Subject:       "测试支付",
	}
	db.Create(payment)

	config := DefaultPaymentConfig()
	service, err := NewService(db, config)
	assert.NoError(t, err)

	tests := []struct {
		name    string
		request *model.PaymentQueryRequest
		wantErr bool
	}{
		{
			name: "通过支付ID查询",
			request: &model.PaymentQueryRequest{
				PaymentID: payment.ID,
			},
			wantErr: false,
		},
		{
			name: "通过支付单号查询",
			request: &model.PaymentQueryRequest{
				PaymentNo: payment.PaymentNo,
			},
			wantErr: false,
		},
		{
			name: "通过订单ID查询",
			request: &model.PaymentQueryRequest{
				OrderID: payment.OrderID,
			},
			wantErr: false,
		},
		{
			name: "查询不存在的支付",
			request: &model.PaymentQueryRequest{
				PaymentID: 99999,
			},
			wantErr: true,
		},
		{
			name:    "空查询参数",
			request: &model.PaymentQueryRequest{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := service.QueryPayment(tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, payment.PaymentNo, resp.PaymentNo)
				assert.Equal(t, payment.PaymentMethod, resp.PaymentMethod)
				assert.Equal(t, payment.Amount, resp.Amount)
			}
		})
	}
}

func TestService_RefundPayment(t *testing.T) {
	db := setupTestDB()
	user, order := createTestData(db)

	// 创建成功的支付记录
	payment := &model.Payment{
		PaymentNo:     "PAY123456",
		OrderID:       order.ID,
		UserID:        user.ID,
		PaymentMethod: model.PaymentMethodAlipay,
		PaymentStatus: model.PaymentStatusSuccess,
		Amount:        decimal.NewFromFloat(100.00),
		Subject:       "测试支付",
	}
	db.Create(payment)

	config := DefaultPaymentConfig()
	service, err := NewService(db, config)
	assert.NoError(t, err)

	tests := []struct {
		name    string
		request *model.PaymentRefundRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "成功申请退款",
			request: &model.PaymentRefundRequest{
				PaymentID:    payment.ID,
				RefundAmount: decimal.NewFromFloat(50.00),
				RefundReason: "用户申请退款",
			},
			wantErr: false,
		},
		{
			name: "退款金额超过支付金额",
			request: &model.PaymentRefundRequest{
				PaymentID:    payment.ID,
				RefundAmount: decimal.NewFromFloat(200.00),
				RefundReason: "用户申请退款",
			},
			wantErr: true,
			errMsg:  "退款金额不能大于支付金额",
		},
		{
			name: "支付记录不存在",
			request: &model.PaymentRefundRequest{
				PaymentID:    99999,
				RefundAmount: decimal.NewFromFloat(50.00),
				RefundReason: "用户申请退款",
			},
			wantErr: true,
		},
		{
			name: "无效的退款金额",
			request: &model.PaymentRefundRequest{
				PaymentID:    payment.ID,
				RefundAmount: decimal.NewFromFloat(-10.00),
				RefundReason: "用户申请退款",
			},
			wantErr: true,
			errMsg:  "无效的金额",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := service.RefundPayment(tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.request.PaymentID, resp.PaymentID)
				assert.Equal(t, tt.request.RefundAmount, resp.RefundAmount)
				assert.Equal(t, tt.request.RefundReason, resp.RefundReason)
			}
		})
	}
}

func TestService_generatePaymentNo(t *testing.T) {
	db := setupTestDB()
	config := DefaultPaymentConfig()
	service, err := NewService(db, config)
	assert.NoError(t, err)

	// 生成多个支付单号，确保唯一性
	paymentNos := make(map[string]bool)
	for i := 0; i < 100; i++ {
		paymentNo := service.generatePaymentNo()
		assert.NotEmpty(t, paymentNo)
		assert.True(t, len(paymentNo) > 10) // 确保长度足够
		assert.False(t, paymentNos[paymentNo], "支付单号应该是唯一的")
		paymentNos[paymentNo] = true
	}
}

func TestService_calculateExpiredTime(t *testing.T) {
	db := setupTestDB()
	config := DefaultPaymentConfig()
	service, err := NewService(db, config)
	assert.NoError(t, err)

	tests := []struct {
		name    string
		minutes int
		want    time.Duration
	}{
		{
			name:    "正常过期时间",
			minutes: 30,
			want:    30 * time.Minute,
		},
		{
			name:    "零值使用默认时间",
			minutes: 0,
			want:    30 * time.Minute, // 默认30分钟
		},
		{
			name:    "负值使用默认时间",
			minutes: -10,
			want:    30 * time.Minute, // 默认30分钟
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now()
			expiredAt := service.calculateExpiredTime(tt.minutes)

			assert.NotNil(t, expiredAt)
			duration := expiredAt.Sub(now)

			// 允许1秒的误差
			assert.True(t, duration >= tt.want-time.Second && duration <= tt.want+time.Second)
		})
	}
}

// BenchmarkService_CreatePayment 性能测试
func BenchmarkService_CreatePayment(b *testing.B) {
	db := setupTestDB()
	user, order := createTestData(db)

	config := DefaultPaymentConfig()
	service, err := NewService(db, config)
	if err != nil {
		b.Fatal(err)
	}

	request := &model.PaymentCreateRequest{
		OrderID:        order.ID,
		PaymentMethod:  model.PaymentMethodAlipay,
		Amount:         decimal.NewFromFloat(100.00),
		Subject:        "性能测试订单",
		ExpiredMinutes: 30,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 为每次测试创建新的订单
		newOrder := &model.Order{
			OrderNo:       fmt.Sprintf("BENCH%d", i),
			UserID:        user.ID,
			TotalAmount:   decimal.NewFromFloat(100.00),
			Status:        model.OrderStatusPending,
			PaymentStatus: model.PaymentStatusPending,
		}
		db.Create(newOrder)

		request.OrderID = newOrder.ID
		_, err := service.CreatePayment(request)
		if err != nil {
			b.Fatal(err)
		}
	}
}
