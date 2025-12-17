package service

import (
	"fmt"

	"mall-go/internal/config"
	"mall-go/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AddressServiceFactory 地址服务工厂实现
type AddressServiceFactory struct{}

// NewAddressServiceFactory 创建地址服务工厂
func NewAddressServiceFactory() IAddressServiceFactory {
	return &AddressServiceFactory{}
}

// CreateAddressService 创建地址服务实例
func (f *AddressServiceFactory) CreateAddressService(dependencies interface{}) (IAddressService, error) {
	deps, ok := dependencies.(*AddressServiceDependencies)
	if !ok {
		return nil, fmt.Errorf("invalid dependencies type, expected *AddressServiceDependencies")
	}

	// 验证必需的依赖项
	if deps.DB == nil {
		return nil, fmt.Errorf("database connection is required")
	}

	db, ok := deps.DB.(*gorm.DB)
	if !ok {
		return nil, fmt.Errorf("invalid database type, expected *gorm.DB")
	}

	// 处理配置依赖
	var configManager *config.AddressConfigManager
	if deps.Config != nil {
		if cfg, ok := deps.Config.(*config.AddressConfig); ok {
			var err error
			configManager, err = config.NewAddressConfigManager(cfg)
			if err != nil {
				logger.Warn("创建配置管理器失败，使用默认配置", zap.Error(err))
				configManager, _ = config.NewAddressConfigManager(nil)
			}
		} else {
			logger.Warn("配置类型不正确，使用默认配置")
			configManager, _ = config.NewAddressConfigManager(nil)
		}
	} else {
		// 使用默认配置
		configManager, _ = config.NewAddressConfigManager(nil)
	}

	// 创建服务实例
	service := &AddressService{
		db:            db,
		configManager: configManager,
	}

	logger.Info("地址服务创建成功")
	return service, nil
}

// DefaultAddressServiceRegistry 默认服务注册表实现
type DefaultAddressServiceRegistry struct {
	addressService IAddressService
}

// NewDefaultAddressServiceRegistry 创建默认服务注册表
func NewDefaultAddressServiceRegistry() ServiceRegistry {
	return &DefaultAddressServiceRegistry{}
}

// RegisterAddressService 注册地址服务
func (r *DefaultAddressServiceRegistry) RegisterAddressService(service IAddressService) error {
	if service == nil {
		return fmt.Errorf("service cannot be nil")
	}
	r.addressService = service
	logger.Info("地址服务注册成功")
	return nil
}

// GetAddressService 获取地址服务
func (r *DefaultAddressServiceRegistry) GetAddressService() (IAddressService, error) {
	if r.addressService == nil {
		return nil, fmt.Errorf("address service not registered")
	}
	return r.addressService, nil
}

// UnregisterAddressService 注销地址服务
func (r *DefaultAddressServiceRegistry) UnregisterAddressService() error {
	r.addressService = nil
	logger.Info("地址服务注销成功")
	return nil
}

// ServiceContainer 服务容器
// 提供依赖注入和服务管理功能
type ServiceContainer struct {
	registry ServiceRegistry
	factory  IAddressServiceFactory
	deps     *AddressServiceDependencies
}

// NewServiceContainer 创建服务容器
func NewServiceContainer() *ServiceContainer {
	return &ServiceContainer{
		registry: NewDefaultAddressServiceRegistry(),
		factory:  NewAddressServiceFactory(),
		deps:     &AddressServiceDependencies{},
	}
}

// SetDatabase 设置数据库连接
func (c *ServiceContainer) SetDatabase(db *gorm.DB) *ServiceContainer {
	c.deps.DB = db
	return c
}

// SetConfig 设置配置
func (c *ServiceContainer) SetConfig(cfg *config.AddressConfig) *ServiceContainer {
	c.deps.Config = cfg
	return c
}

// SetLogger 设置日志记录器
func (c *ServiceContainer) SetLogger(log interface{}) *ServiceContainer {
	c.deps.Logger = log
	return c
}

// SetCache 设置缓存服务
func (c *ServiceContainer) SetCache(cache interface{}) *ServiceContainer {
	c.deps.Cache = cache
	return c
}

// SetMetrics 设置监控指标收集器
func (c *ServiceContainer) SetMetrics(metrics interface{}) *ServiceContainer {
	c.deps.Metrics = metrics
	return c
}

// Build 构建并注册服务
func (c *ServiceContainer) Build() (IAddressService, error) {
	// 创建服务实例
	service, err := c.factory.CreateAddressService(c.deps)
	if err != nil {
		return nil, fmt.Errorf("failed to create address service: %w", err)
	}

	// 注册服务
	if err := c.registry.RegisterAddressService(service); err != nil {
		return nil, fmt.Errorf("failed to register address service: %w", err)
	}

	return service, nil
}

// GetService 获取已注册的服务
func (c *ServiceContainer) GetService() (IAddressService, error) {
	return c.registry.GetAddressService()
}

// Cleanup 清理资源
func (c *ServiceContainer) Cleanup() error {
	return c.registry.UnregisterAddressService()
}

// CreateAddressServiceWithDefaults 使用默认配置创建地址服务
// 这是一个便捷函数，用于快速创建服务实例
func CreateAddressServiceWithDefaults(db *gorm.DB) (IAddressService, error) {
	container := NewServiceContainer()
	return container.SetDatabase(db).Build()
}

// CreateAddressServiceWithConfig 使用指定配置创建地址服务
// 这是一个便捷函数，用于创建带配置的服务实例
func CreateAddressServiceWithConfig(db *gorm.DB, cfg *config.AddressConfig) (IAddressService, error) {
	container := NewServiceContainer()
	return container.SetDatabase(db).SetConfig(cfg).Build()
}

// CreateAddressServiceWithDependencies 使用完整依赖创建地址服务
// 这是一个便捷函数，用于创建带完整依赖的服务实例
func CreateAddressServiceWithDependencies(deps *AddressServiceDependencies) (IAddressService, error) {
	factory := NewAddressServiceFactory()
	return factory.CreateAddressService(deps)
}

// ValidateServiceDependencies 验证服务依赖项
func ValidateServiceDependencies(deps *AddressServiceDependencies) error {
	if deps == nil {
		return fmt.Errorf("dependencies cannot be nil")
	}

	if deps.DB == nil {
		return fmt.Errorf("database connection is required")
	}

	if _, ok := deps.DB.(*gorm.DB); !ok {
		return fmt.Errorf("invalid database type, expected *gorm.DB")
	}

	// 配置是可选的，但如果提供了需要验证类型
	if deps.Config != nil {
		if _, ok := deps.Config.(*config.AddressConfig); !ok {
			return fmt.Errorf("invalid config type, expected *config.AddressConfig")
		}
	}

	logger.Info("服务依赖项验证通过")
	return nil
}

// GetServiceInfo 获取服务信息
func GetServiceInfo() map[string]interface{} {
	return map[string]interface{}{
		"service_name":    "AddressService",
		"interface_name":  "IAddressService",
		"factory_name":    "AddressServiceFactory",
		"registry_name":   "DefaultAddressServiceRegistry",
		"container_name":  "ServiceContainer",
		"version":         "1.0.0",
		"description":     "地址管理服务，支持依赖注入和接口抽象",
		"features": []string{
			"依赖注入",
			"接口抽象",
			"服务注册表",
			"工厂模式",
			"配置管理",
			"错误处理",
		},
	}
}
