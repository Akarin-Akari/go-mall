package service

import (
	"context"
	"time"

	"mall-go/internal/config"
	"mall-go/internal/model"

	"gorm.io/gorm"
)

// MockDB Mock数据库接口
type MockDB struct {
	addresses       map[uint]*model.Address
	userAddresses   map[uint][]*model.Address
	defaultAddresses map[uint]*model.Address
	nextID          uint
	shouldError     bool
	errorMessage    string
	queryDelay      time.Duration
}

// NewMockDB 创建Mock数据库
func NewMockDB() *MockDB {
	return &MockDB{
		addresses:       make(map[uint]*model.Address),
		userAddresses:   make(map[uint][]*model.Address),
		defaultAddresses: make(map[uint]*model.Address),
		nextID:          1,
		shouldError:     false,
		queryDelay:      0,
	}
}

// SetError 设置Mock数据库返回错误
func (m *MockDB) SetError(shouldError bool, message string) {
	m.shouldError = shouldError
	m.errorMessage = message
}

// SetQueryDelay 设置查询延迟（用于测试超时）
func (m *MockDB) SetQueryDelay(delay time.Duration) {
	m.queryDelay = delay
}

// AddTestData 添加测试数据
func (m *MockDB) AddTestData(userID uint, addresses []*model.Address) {
	for _, addr := range addresses {
		if addr.ID == 0 {
			addr.ID = m.nextID
			m.nextID++
		}
		addr.UserID = userID
		
		m.addresses[addr.ID] = addr
		
		if m.userAddresses[userID] == nil {
			m.userAddresses[userID] = make([]*model.Address, 0)
		}
		m.userAddresses[userID] = append(m.userAddresses[userID], addr)
		
		if addr.IsDefault {
			m.defaultAddresses[userID] = addr
		}
	}
}

// GetAddress 获取地址
func (m *MockDB) GetAddress(ctx context.Context, addressID uint) (*model.Address, error) {
	if m.queryDelay > 0 {
		time.Sleep(m.queryDelay)
	}
	
	if m.shouldError {
		return nil, gorm.ErrRecordNotFound
	}
	
	if addr, exists := m.addresses[addressID]; exists {
		return addr, nil
	}
	
	return nil, gorm.ErrRecordNotFound
}

// GetUserAddresses 获取用户地址列表
func (m *MockDB) GetUserAddresses(ctx context.Context, userID uint) ([]*model.Address, error) {
	if m.queryDelay > 0 {
		time.Sleep(m.queryDelay)
	}
	
	if m.shouldError {
		return nil, gorm.ErrInvalidDB
	}
	
	if addresses, exists := m.userAddresses[userID]; exists {
		// 返回副本，避免外部修改
		result := make([]*model.Address, len(addresses))
		copy(result, addresses)
		return result, nil
	}
	
	return []*model.Address{}, nil
}

// CreateAddress 创建地址
func (m *MockDB) CreateAddress(ctx context.Context, address *model.Address) error {
	if m.queryDelay > 0 {
		time.Sleep(m.queryDelay)
	}
	
	if m.shouldError {
		return gorm.ErrInvalidTransaction
	}
	
	if address.ID == 0 {
		address.ID = m.nextID
		m.nextID++
	}
	
	m.addresses[address.ID] = address
	
	if m.userAddresses[address.UserID] == nil {
		m.userAddresses[address.UserID] = make([]*model.Address, 0)
	}
	m.userAddresses[address.UserID] = append(m.userAddresses[address.UserID], address)
	
	if address.IsDefault {
		// 清除其他默认地址
		for _, addr := range m.userAddresses[address.UserID] {
			if addr.ID != address.ID {
				addr.IsDefault = false
			}
		}
		m.defaultAddresses[address.UserID] = address
	}
	
	return nil
}

// UpdateAddress 更新地址
func (m *MockDB) UpdateAddress(ctx context.Context, address *model.Address) error {
	if m.queryDelay > 0 {
		time.Sleep(m.queryDelay)
	}
	
	if m.shouldError {
		return gorm.ErrInvalidTransaction
	}
	
	if _, exists := m.addresses[address.ID]; !exists {
		return gorm.ErrRecordNotFound
	}
	
	m.addresses[address.ID] = address
	
	// 更新用户地址列表中的地址
	for i, addr := range m.userAddresses[address.UserID] {
		if addr.ID == address.ID {
			m.userAddresses[address.UserID][i] = address
			break
		}
	}
	
	if address.IsDefault {
		m.defaultAddresses[address.UserID] = address
	}
	
	return nil
}

// DeleteAddress 删除地址
func (m *MockDB) DeleteAddress(ctx context.Context, addressID uint) error {
	if m.queryDelay > 0 {
		time.Sleep(m.queryDelay)
	}
	
	if m.shouldError {
		return gorm.ErrInvalidTransaction
	}
	
	address, exists := m.addresses[addressID]
	if !exists {
		return gorm.ErrRecordNotFound
	}
	
	delete(m.addresses, addressID)
	
	// 从用户地址列表中删除
	userAddresses := m.userAddresses[address.UserID]
	for i, addr := range userAddresses {
		if addr.ID == addressID {
			m.userAddresses[address.UserID] = append(userAddresses[:i], userAddresses[i+1:]...)
			break
		}
	}
	
	// 如果删除的是默认地址，清除默认地址记录
	if defaultAddr, exists := m.defaultAddresses[address.UserID]; exists && defaultAddr.ID == addressID {
		delete(m.defaultAddresses, address.UserID)
	}
	
	return nil
}

// SetDefaultAddress 设置默认地址
func (m *MockDB) SetDefaultAddress(ctx context.Context, userID, addressID uint) error {
	if m.queryDelay > 0 {
		time.Sleep(m.queryDelay)
	}
	
	if m.shouldError {
		return gorm.ErrInvalidTransaction
	}
	
	address, exists := m.addresses[addressID]
	if !exists || address.UserID != userID {
		return gorm.ErrRecordNotFound
	}
	
	// 清除所有默认地址
	for _, addr := range m.userAddresses[userID] {
		addr.IsDefault = false
	}
	
	// 设置新的默认地址
	address.IsDefault = true
	m.defaultAddresses[userID] = address
	
	return nil
}

// GetDefaultAddress 获取默认地址
func (m *MockDB) GetDefaultAddress(ctx context.Context, userID uint) (*model.Address, error) {
	if m.queryDelay > 0 {
		time.Sleep(m.queryDelay)
	}
	
	if m.shouldError {
		return nil, gorm.ErrInvalidDB
	}
	
	if addr, exists := m.defaultAddresses[userID]; exists {
		return addr, nil
	}
	
	return nil, gorm.ErrRecordNotFound
}

// MockCacheService Mock缓存服务
type MockCacheService struct {
	cache       map[string]interface{}
	shouldError bool
	hitRate     float64
	enabled     bool
}

// NewMockCacheService 创建Mock缓存服务
func NewMockCacheService() *MockCacheService {
	return &MockCacheService{
		cache:   make(map[string]interface{}),
		enabled: true,
		hitRate: 0.8, // 80%命中率
	}
}

// SetEnabled 设置缓存启用状态
func (m *MockCacheService) SetEnabled(enabled bool) {
	m.enabled = enabled
}

// SetError 设置是否返回错误
func (m *MockCacheService) SetError(shouldError bool) {
	m.shouldError = shouldError
}

// IsEnabled 检查缓存是否启用
func (m *MockCacheService) IsEnabled() bool {
	return m.enabled
}

// Get 获取缓存
func (m *MockCacheService) Get(key string) (interface{}, bool) {
	if !m.enabled {
		return nil, false
	}
	
	if m.shouldError {
		return nil, false
	}
	
	value, exists := m.cache[key]
	return value, exists
}

// Set 设置缓存
func (m *MockCacheService) Set(key string, value interface{}) error {
	if !m.enabled {
		return nil
	}
	
	if m.shouldError {
		return gorm.ErrInvalidDB
	}
	
	m.cache[key] = value
	return nil
}

// Delete 删除缓存
func (m *MockCacheService) Delete(key string) error {
	if !m.enabled {
		return nil
	}
	
	delete(m.cache, key)
	return nil
}

// MockPerformanceMonitor Mock性能监控器
type MockPerformanceMonitor struct {
	metrics map[string]float64
	counters map[string]int64
}

// NewMockPerformanceMonitor 创建Mock性能监控器
func NewMockPerformanceMonitor() *MockPerformanceMonitor {
	return &MockPerformanceMonitor{
		metrics:  make(map[string]float64),
		counters: make(map[string]int64),
	}
}

// RecordMetric 记录指标
func (m *MockPerformanceMonitor) RecordMetric(name string, value float64, labels map[string]string) {
	m.metrics[name] = value
}

// IncrementCounter 增加计数器
func (m *MockPerformanceMonitor) IncrementCounter(name string, labels map[string]string) {
	m.counters[name]++
}

// RecordAddressOperation 记录地址操作
func (m *MockPerformanceMonitor) RecordAddressOperation(operation string, err error) {
	if err != nil {
		m.counters[operation+"_error"]++
	} else {
		m.counters[operation+"_success"]++
	}
}

// RecordCacheHit 记录缓存命中
func (m *MockPerformanceMonitor) RecordCacheHit(cacheType, operation string) {
	m.counters["cache_hit"]++
}

// RecordCacheMiss 记录缓存未命中
func (m *MockPerformanceMonitor) RecordCacheMiss(cacheType, operation string) {
	m.counters["cache_miss"]++
}

// GetMetric 获取指标值
func (m *MockPerformanceMonitor) GetMetric(name string) float64 {
	return m.metrics[name]
}

// GetCounter 获取计数器值
func (m *MockPerformanceMonitor) GetCounter(name string) int64 {
	return m.counters[name]
}

// TestDataFactory 测试数据工厂
type TestDataFactory struct{}

// NewTestDataFactory 创建测试数据工厂
func NewTestDataFactory() *TestDataFactory {
	return &TestDataFactory{}
}

// CreateTestAddress 创建测试地址
func (f *TestDataFactory) CreateTestAddress(userID uint, isDefault bool) *model.Address {
	return &model.Address{
		UserID:        userID,
		ReceiverName:  "测试用户",
		ReceiverPhone: "13800138000",
		Province:      "北京市",
		City:          "北京市",
		District:      "朝阳区",
		DetailAddress: "测试详细地址",
		IsDefault:     isDefault,
	}
}

// CreateTestAddresses 创建多个测试地址
func (f *TestDataFactory) CreateTestAddresses(userID uint, count int) []*model.Address {
	addresses := make([]*model.Address, count)
	for i := 0; i < count; i++ {
		addresses[i] = f.CreateTestAddress(userID, i == 0) // 第一个设为默认地址
		addresses[i].ReceiverName = fmt.Sprintf("测试用户%d", i+1)
		addresses[i].DetailAddress = fmt.Sprintf("测试详细地址%d", i+1)
	}
	return addresses
}

// CreateTestAddressCreateRequest 创建测试地址创建请求
func (f *TestDataFactory) CreateTestAddressCreateRequest() *model.AddressCreateRequest {
	return &model.AddressCreateRequest{
		ReceiverName:  "新测试用户",
		ReceiverPhone: "13900139000",
		Province:      "上海市",
		City:          "上海市",
		District:      "浦东新区",
		DetailAddress: "新测试详细地址",
		IsDefault:     false,
	}
}
