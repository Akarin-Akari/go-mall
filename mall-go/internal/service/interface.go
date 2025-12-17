package service

import (
	"context"

	"mall-go/internal/model"
)

// IAddressService 地址服务接口
// 定义地址管理的所有业务操作，便于依赖注入和单元测试
type IAddressService interface {
	// CreateAddress 创建新地址
	// 参数:
	//   - ctx: 上下文，用于超时控制和取消操作
	//   - userID: 用户ID，必须大于0
	//   - req: 地址创建请求，包含地址详细信息
	// 返回:
	//   - *model.Address: 创建成功的地址对象
	//   - error: 创建失败时的错误信息
	CreateAddress(ctx context.Context, userID uint, req *model.AddressCreateRequest) (*model.Address, error)

	// GetUserAddresses 获取用户所有地址
	// 参数:
	//   - ctx: 上下文，用于超时控制和取消操作
	//   - userID: 用户ID，必须大于0
	// 返回:
	//   - []*model.Address: 用户的地址列表，按默认地址和创建时间排序
	//   - error: 查询失败时的错误信息
	GetUserAddresses(ctx context.Context, userID uint) ([]*model.Address, error)

	// GetAddressByID 根据ID获取地址
	// 参数:
	//   - ctx: 上下文，用于超时控制和取消操作
	//   - userID: 用户ID，必须大于0
	//   - addressID: 地址ID，必须大于0
	// 返回:
	//   - *model.Address: 查询到的地址对象
	//   - error: 查询失败时的错误信息（如地址不存在或无权限访问）
	GetAddressByID(ctx context.Context, userID, addressID uint) (*model.Address, error)

	// UpdateAddress 更新地址
	// 参数:
	//   - ctx: 上下文，用于超时控制和取消操作
	//   - userID: 用户ID，必须大于0
	//   - addressID: 地址ID，必须大于0
	//   - req: 地址更新请求，包含要更新的字段
	// 返回:
	//   - *model.Address: 更新后的地址对象
	//   - error: 更新失败时的错误信息
	UpdateAddress(ctx context.Context, userID, addressID uint, req *model.AddressUpdateRequest) (*model.Address, error)

	// DeleteAddress 删除地址
	// 参数:
	//   - ctx: 上下文，用于超时控制和取消操作
	//   - userID: 用户ID，必须大于0
	//   - addressID: 地址ID，必须大于0
	// 返回:
	//   - error: 删除失败时的错误信息
	DeleteAddress(ctx context.Context, userID, addressID uint) error

	// SetDefaultAddress 设置默认地址
	// 参数:
	//   - ctx: 上下文，用于超时控制和取消操作
	//   - userID: 用户ID，必须大于0
	//   - addressID: 地址ID，必须大于0
	// 返回:
	//   - *model.Address: 设置为默认的地址对象
	//   - error: 设置失败时的错误信息
	SetDefaultAddress(ctx context.Context, userID, addressID uint) (*model.Address, error)

	// GetDefaultAddress 获取默认地址
	// 参数:
	//   - ctx: 上下文，用于超时控制和取消操作
	//   - userID: 用户ID，必须大于0
	// 返回:
	//   - *model.Address: 用户的默认地址
	//   - error: 查询失败时的错误信息（如没有默认地址）
	GetDefaultAddress(ctx context.Context, userID uint) (*model.Address, error)

	// GetAddressesWithFilter 根据条件获取地址列表
	// 参数:
	//   - ctx: 上下文，用于超时控制和取消操作
	//   - req: 地址列表查询请求，包含过滤条件和分页参数
	// 返回:
	//   - []*model.Address: 符合条件的地址列表
	//   - error: 查询失败时的错误信息
	GetAddressesWithFilter(ctx context.Context, req *model.AddressListRequest) ([]*model.Address, error)

	// ValidateAddressOwnership 验证地址归属
	// 参数:
	//   - ctx: 上下文，用于超时控制和取消操作
	//   - userID: 用户ID，必须大于0
	//   - addressID: 地址ID，必须大于0
	// 返回:
	//   - error: 验证失败时的错误信息（如地址不属于该用户）
	ValidateAddressOwnership(ctx context.Context, userID, addressID uint) error
}

// IAddressServiceFactory 地址服务工厂接口
// 用于创建地址服务实例，支持依赖注入
type IAddressServiceFactory interface {
	// CreateAddressService 创建地址服务实例
	// 参数:
	//   - dependencies: 服务依赖项（如数据库连接、配置等）
	// 返回:
	//   - IAddressService: 地址服务接口实例
	//   - error: 创建失败时的错误信息
	CreateAddressService(dependencies interface{}) (IAddressService, error)
}

// AddressServiceDependencies 地址服务依赖项
// 包含地址服务所需的所有外部依赖
type AddressServiceDependencies struct {
	// DB 数据库连接
	DB interface{}
	
	// Config 配置管理器
	Config interface{}
	
	// Logger 日志记录器
	Logger interface{}
	
	// Cache 缓存服务（可选）
	Cache interface{}
	
	// Metrics 监控指标收集器（可选）
	Metrics interface{}
}

// ServiceRegistry 服务注册表接口
// 用于管理所有服务实例的生命周期
type ServiceRegistry interface {
	// RegisterAddressService 注册地址服务
	RegisterAddressService(service IAddressService) error
	
	// GetAddressService 获取地址服务
	GetAddressService() (IAddressService, error)
	
	// UnregisterAddressService 注销地址服务
	UnregisterAddressService() error
}

// MockAddressService Mock地址服务接口
// 用于单元测试，提供可控的测试行为
type MockAddressService interface {
	IAddressService
	
	// SetMockBehavior 设置Mock行为
	SetMockBehavior(method string, behavior interface{}) error
	
	// ResetMockBehavior 重置Mock行为
	ResetMockBehavior() error
	
	// GetCallHistory 获取调用历史
	GetCallHistory() []interface{}
	
	// ClearCallHistory 清除调用历史
	ClearCallHistory() error
}

// AddressServiceMetrics 地址服务指标接口
// 用于收集和报告服务性能指标
type AddressServiceMetrics interface {
	// RecordOperation 记录操作指标
	RecordOperation(operation string, duration int64, success bool) error
	
	// RecordError 记录错误指标
	RecordError(operation string, errorType string) error
	
	// GetMetrics 获取指标数据
	GetMetrics() (map[string]interface{}, error)
	
	// ResetMetrics 重置指标数据
	ResetMetrics() error
}

// AddressServiceValidator 地址服务验证器接口
// 用于验证业务规则和数据完整性
type AddressServiceValidator interface {
	// ValidateCreateRequest 验证创建请求
	ValidateCreateRequest(req *model.AddressCreateRequest) error
	
	// ValidateUpdateRequest 验证更新请求
	ValidateUpdateRequest(req *model.AddressUpdateRequest) error
	
	// ValidateListRequest 验证列表查询请求
	ValidateListRequest(req *model.AddressListRequest) error
	
	// ValidateBusinessRules 验证业务规则
	ValidateBusinessRules(userID uint, operation string, data interface{}) error
}

// AddressServiceAuditor 地址服务审计接口
// 用于记录和追踪用户操作
type AddressServiceAuditor interface {
	// LogOperation 记录操作日志
	LogOperation(userID uint, operation string, details interface{}) error
	
	// LogError 记录错误日志
	LogError(userID uint, operation string, err error) error
	
	// GetAuditLogs 获取审计日志
	GetAuditLogs(userID uint, startTime, endTime int64) ([]interface{}, error)
}
