package service

import (
	"context"
	"testing"
	"time"

	"mall-go/internal/config"
	"mall-go/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAddressService_CreateAddress 测试创建地址
func TestAddressService_CreateAddress(t *testing.T) {
	// 准备测试环境
	mockDB := NewMockDB()
	factory := NewTestDataFactory()
	service := createTestAddressService(mockDB)
	
	tests := []struct {
		name        string
		userID      uint
		request     *model.AddressCreateRequest
		setupMock   func()
		expectError bool
		errorType   error
	}{
		{
			name:   "成功创建地址",
			userID: 123,
			request: factory.CreateTestAddressCreateRequest(),
			setupMock: func() {
				mockDB.SetError(false, "")
			},
			expectError: false,
		},
		{
			name:   "用户ID无效",
			userID: 0,
			request: factory.CreateTestAddressCreateRequest(),
			setupMock: func() {
				mockDB.SetError(false, "")
			},
			expectError: true,
			errorType:   ErrInvalidUserID,
		},
		{
			name:   "请求对象为nil",
			userID: 123,
			request: nil,
			setupMock: func() {
				mockDB.SetError(false, "")
			},
			expectError: true,
			errorType:   ErrInvalidRequest,
		},
		{
			name:   "数据库错误",
			userID: 123,
			request: factory.CreateTestAddressCreateRequest(),
			setupMock: func() {
				mockDB.SetError(true, "数据库连接失败")
			},
			expectError: true,
		},
		{
			name:   "操作超时",
			userID: 123,
			request: factory.CreateTestAddressCreateRequest(),
			setupMock: func() {
				mockDB.SetError(false, "")
				mockDB.SetQueryDelay(2 * time.Second) // 设置超时
			},
			expectError: true,
			errorType:   ErrOperationTimeout,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置Mock
			tt.setupMock()
			
			// 执行测试
			ctx := context.Background()
			result, err := service.CreateAddress(ctx, tt.userID, tt.request)
			
			// 验证结果
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errorType != nil {
					assert.Equal(t, tt.errorType, err)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.userID, result.UserID)
				if tt.request != nil {
					assert.Equal(t, tt.request.ReceiverName, result.ReceiverName)
					assert.Equal(t, tt.request.ReceiverPhone, result.ReceiverPhone)
				}
			}
			
			// 重置Mock状态
			mockDB.SetError(false, "")
			mockDB.SetQueryDelay(0)
		})
	}
}

// TestAddressService_GetUserAddresses 测试获取用户地址列表
func TestAddressService_GetUserAddresses(t *testing.T) {
	// 准备测试环境
	mockDB := NewMockDB()
	factory := NewTestDataFactory()
	service := createTestAddressService(mockDB)
	
	// 添加测试数据
	userID := uint(123)
	testAddresses := factory.CreateTestAddresses(userID, 3)
	mockDB.AddTestData(userID, testAddresses)
	
	tests := []struct {
		name        string
		userID      uint
		setupMock   func()
		expectError bool
		expectCount int
	}{
		{
			name:   "成功获取地址列表",
			userID: userID,
			setupMock: func() {
				mockDB.SetError(false, "")
			},
			expectError: false,
			expectCount: 3,
		},
		{
			name:   "用户ID无效",
			userID: 0,
			setupMock: func() {
				mockDB.SetError(false, "")
			},
			expectError: true,
		},
		{
			name:   "用户无地址",
			userID: 999,
			setupMock: func() {
				mockDB.SetError(false, "")
			},
			expectError: false,
			expectCount: 0,
		},
		{
			name:   "数据库错误",
			userID: userID,
			setupMock: func() {
				mockDB.SetError(true, "数据库查询失败")
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置Mock
			tt.setupMock()
			
			// 执行测试
			ctx := context.Background()
			result, err := service.GetUserAddresses(ctx, tt.userID)
			
			// 验证结果
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, tt.expectCount)
				
				// 验证地址排序（默认地址在前）
				if len(result) > 0 && tt.expectCount > 0 {
					assert.True(t, result[0].IsDefault)
				}
			}
			
			// 重置Mock状态
			mockDB.SetError(false, "")
		})
	}
}

// TestAddressService_GetAddressByID 测试根据ID获取地址
func TestAddressService_GetAddressByID(t *testing.T) {
	// 准备测试环境
	mockDB := NewMockDB()
	factory := NewTestDataFactory()
	service := createTestAddressService(mockDB)
	
	// 添加测试数据
	userID := uint(123)
	testAddress := factory.CreateTestAddress(userID, true)
	testAddress.ID = 1
	mockDB.AddTestData(userID, []*model.Address{testAddress})
	
	tests := []struct {
		name        string
		userID      uint
		addressID   uint
		setupMock   func()
		expectError bool
		errorType   error
	}{
		{
			name:      "成功获取地址",
			userID:    userID,
			addressID: 1,
			setupMock: func() {
				mockDB.SetError(false, "")
			},
			expectError: false,
		},
		{
			name:      "用户ID无效",
			userID:    0,
			addressID: 1,
			setupMock: func() {
				mockDB.SetError(false, "")
			},
			expectError: true,
			errorType:   ErrInvalidUserID,
		},
		{
			name:      "地址ID无效",
			userID:    userID,
			addressID: 0,
			setupMock: func() {
				mockDB.SetError(false, "")
			},
			expectError: true,
			errorType:   ErrInvalidAddressID,
		},
		{
			name:      "地址不存在",
			userID:    userID,
			addressID: 999,
			setupMock: func() {
				mockDB.SetError(false, "")
			},
			expectError: true,
			errorType:   ErrAddressNotFound,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置Mock
			tt.setupMock()
			
			// 执行测试
			ctx := context.Background()
			result, err := service.GetAddressByID(ctx, tt.userID, tt.addressID)
			
			// 验证结果
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errorType != nil {
					assert.Equal(t, tt.errorType, err)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.addressID, result.ID)
				assert.Equal(t, tt.userID, result.UserID)
			}
			
			// 重置Mock状态
			mockDB.SetError(false, "")
		})
	}
}

// TestAddressService_SetDefaultAddress 测试设置默认地址
func TestAddressService_SetDefaultAddress(t *testing.T) {
	// 准备测试环境
	mockDB := NewMockDB()
	factory := NewTestDataFactory()
	service := createTestAddressService(mockDB)
	
	// 添加测试数据
	userID := uint(123)
	testAddresses := factory.CreateTestAddresses(userID, 3)
	for i, addr := range testAddresses {
		addr.ID = uint(i + 1)
		addr.IsDefault = (i == 0) // 第一个为默认地址
	}
	mockDB.AddTestData(userID, testAddresses)
	
	tests := []struct {
		name        string
		userID      uint
		addressID   uint
		setupMock   func()
		expectError bool
		errorType   error
	}{
		{
			name:      "成功设置默认地址",
			userID:    userID,
			addressID: 2, // 设置第二个地址为默认
			setupMock: func() {
				mockDB.SetError(false, "")
			},
			expectError: false,
		},
		{
			name:      "用户ID无效",
			userID:    0,
			addressID: 1,
			setupMock: func() {
				mockDB.SetError(false, "")
			},
			expectError: true,
			errorType:   ErrInvalidUserID,
		},
		{
			name:      "地址ID无效",
			userID:    userID,
			addressID: 0,
			setupMock: func() {
				mockDB.SetError(false, "")
			},
			expectError: true,
			errorType:   ErrInvalidAddressID,
		},
		{
			name:      "地址不存在",
			userID:    userID,
			addressID: 999,
			setupMock: func() {
				mockDB.SetError(false, "")
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置Mock
			tt.setupMock()
			
			// 执行测试
			ctx := context.Background()
			result, err := service.SetDefaultAddress(ctx, tt.userID, tt.addressID)
			
			// 验证结果
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errorType != nil {
					assert.Equal(t, tt.errorType, err)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.addressID, result.ID)
				assert.True(t, result.IsDefault)
			}
			
			// 重置Mock状态
			mockDB.SetError(false, "")
		})
	}
}

// TestAddressService_DeleteAddress 测试删除地址
func TestAddressService_DeleteAddress(t *testing.T) {
	// 准备测试环境
	mockDB := NewMockDB()
	factory := NewTestDataFactory()
	service := createTestAddressService(mockDB)
	
	// 添加测试数据
	userID := uint(123)
	testAddress := factory.CreateTestAddress(userID, false)
	testAddress.ID = 1
	mockDB.AddTestData(userID, []*model.Address{testAddress})
	
	tests := []struct {
		name        string
		userID      uint
		addressID   uint
		setupMock   func()
		expectError bool
		errorType   error
	}{
		{
			name:      "成功删除地址",
			userID:    userID,
			addressID: 1,
			setupMock: func() {
				mockDB.SetError(false, "")
			},
			expectError: false,
		},
		{
			name:      "用户ID无效",
			userID:    0,
			addressID: 1,
			setupMock: func() {
				mockDB.SetError(false, "")
			},
			expectError: true,
			errorType:   ErrInvalidUserID,
		},
		{
			name:      "地址ID无效",
			userID:    userID,
			addressID: 0,
			setupMock: func() {
				mockDB.SetError(false, "")
			},
			expectError: true,
			errorType:   ErrInvalidAddressID,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置Mock
			tt.setupMock()
			
			// 执行测试
			ctx := context.Background()
			err := service.DeleteAddress(ctx, tt.userID, tt.addressID)
			
			// 验证结果
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != nil {
					assert.Equal(t, tt.errorType, err)
				}
			} else {
				assert.NoError(t, err)
			}
			
			// 重置Mock状态
			mockDB.SetError(false, "")
		})
	}
}

// createTestAddressService 创建测试用的AddressService
func createTestAddressService(mockDB *MockDB) *AddressService {
	// 创建配置
	cfg := config.DefaultAddressConfig()
	configManager, _ := config.NewAddressConfigManager(cfg)
	
	// 创建审计日志记录器
	auditLogger := NewAuditLogger()
	
	// 创建超时管理器
	timeoutManager := NewTimeoutManager(cfg)
	timeoutWrapper := NewTimeoutWrapper(timeoutManager)
	
	// 创建性能监控器
	performanceMonitor := NewMockPerformanceMonitor()
	
	// 创建缓存服务（禁用状态）
	cacheService := NewCacheService(nil, cfg)
	
	return &AddressService{
		db:                 nil, // 使用Mock，不需要真实DB
		configManager:      configManager,
		auditLogger:        auditLogger,
		timeoutManager:     timeoutManager,
		timeoutWrapper:     timeoutWrapper,
		performanceMonitor: performanceMonitor,
		cacheService:       cacheService,
	}
}
