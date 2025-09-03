package payment

import (
	"fmt"
	"sync"
	"time"

	"mall-go/internal/model"
	"mall-go/pkg/logger"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ConfigManager 配置管理器
type ConfigManager struct {
	db           *gorm.DB
	config       *PaymentConfig
	dbConfigs    map[model.PaymentMethod]*model.PaymentConfig
	mutex        sync.RWMutex
	lastUpdate   time.Time
	updateTicker *time.Ticker
}

// NewConfigManager 创建配置管理器
func NewConfigManager(db *gorm.DB, config *PaymentConfig) *ConfigManager {
	manager := &ConfigManager{
		db:        db,
		config:    config,
		dbConfigs: make(map[model.PaymentMethod]*model.PaymentConfig),
	}

	// 加载数据库配置
	if err := manager.LoadDBConfigs(); err != nil {
		logger.Error("加载数据库配置失败", zap.Error(err))
	}

	// 启动定时更新
	manager.startAutoUpdate()

	return manager
}

// LoadDBConfigs 加载数据库配置
func (cm *ConfigManager) LoadDBConfigs() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	var configs []model.PaymentConfig
	if err := cm.db.Find(&configs).Error; err != nil {
		return fmt.Errorf("查询支付配置失败: %v", err)
	}

	// 更新内存配置
	for _, config := range configs {
		cm.dbConfigs[config.PaymentMethod] = &config
	}

	cm.lastUpdate = time.Now()
	logger.Info("数据库配置加载完成", zap.Int("count", len(configs)))

	return nil
}

// GetConfig 获取支付配置
func (cm *ConfigManager) GetConfig() *PaymentConfig {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	return cm.config
}

// GetMethodConfig 获取支付方式配置
func (cm *ConfigManager) GetMethodConfig(method model.PaymentMethod) (*model.PaymentConfig, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	if config, exists := cm.dbConfigs[method]; exists {
		return config, nil
	}

	return nil, fmt.Errorf("支付方式 %s 配置不存在", method)
}

// IsMethodEnabled 检查支付方式是否启用
func (cm *ConfigManager) IsMethodEnabled(method model.PaymentMethod) bool {
	config, err := cm.GetMethodConfig(method)
	if err != nil {
		return false
	}

	return config.IsEnabled
}

// GetMethodLimit 获取支付方式限额
func (cm *ConfigManager) GetMethodLimit(method model.PaymentMethod) (model.PaymentConfig, error) {
	config, err := cm.GetMethodConfig(method)
	if err != nil {
		return model.PaymentConfig{}, err
	}

	return *config, nil
}

// ValidateAmount 验证支付金额
func (cm *ConfigManager) ValidateAmount(method model.PaymentMethod, amount decimal.Decimal) error {
	config, err := cm.GetMethodConfig(method)
	if err != nil {
		return err
	}

	if amount.LessThan(config.MinAmount) {
		return fmt.Errorf("支付金额不能小于 %s", config.MinAmount.String())
	}

	if amount.GreaterThan(config.MaxAmount) {
		return fmt.Errorf("支付金额不能大于 %s", config.MaxAmount.String())
	}

	return nil
}

// UpdateMethodConfig 更新支付方式配置
func (cm *ConfigManager) UpdateMethodConfig(method model.PaymentMethod, config *model.PaymentConfig) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 更新数据库
	if err := cm.db.Save(config).Error; err != nil {
		return fmt.Errorf("更新支付配置失败: %v", err)
	}

	// 更新内存配置
	cm.dbConfigs[method] = config

	logger.Info("支付方式配置已更新",
		zap.String("method", string(method)),
		zap.Bool("enabled", config.IsEnabled))

	return nil
}

// CreateMethodConfig 创建支付方式配置
func (cm *ConfigManager) CreateMethodConfig(config *model.PaymentConfig) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 检查是否已存在
	if _, exists := cm.dbConfigs[config.PaymentMethod]; exists {
		return fmt.Errorf("支付方式 %s 配置已存在", config.PaymentMethod)
	}

	// 创建数据库记录
	if err := cm.db.Create(config).Error; err != nil {
		return fmt.Errorf("创建支付配置失败: %v", err)
	}

	// 更新内存配置
	cm.dbConfigs[config.PaymentMethod] = config

	logger.Info("支付方式配置已创建",
		zap.String("method", string(config.PaymentMethod)),
		zap.String("display_name", config.DisplayName))

	return nil
}

// DeleteMethodConfig 删除支付方式配置
func (cm *ConfigManager) DeleteMethodConfig(method model.PaymentMethod) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 删除数据库记录
	if err := cm.db.Where("payment_method = ?", method).Delete(&model.PaymentConfig{}).Error; err != nil {
		return fmt.Errorf("删除支付配置失败: %v", err)
	}

	// 删除内存配置
	delete(cm.dbConfigs, method)

	logger.Info("支付方式配置已删除", zap.String("method", string(method)))

	return nil
}

// GetEnabledMethods 获取启用的支付方式
func (cm *ConfigManager) GetEnabledMethods() []model.PaymentMethod {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	var methods []model.PaymentMethod
	for method, config := range cm.dbConfigs {
		if config.IsEnabled {
			methods = append(methods, method)
		}
	}

	return methods
}

// GetAllMethods 获取所有支付方式配置
func (cm *ConfigManager) GetAllMethods() map[model.PaymentMethod]*model.PaymentConfig {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	// 复制配置避免并发问题
	configs := make(map[model.PaymentMethod]*model.PaymentConfig)
	for method, config := range cm.dbConfigs {
		configCopy := *config
		configs[method] = &configCopy
	}

	return configs
}

// InitDefaultConfigs 初始化默认配置
func (cm *ConfigManager) InitDefaultConfigs() error {
	defaultConfigs := []model.PaymentConfig{
		{
			PaymentMethod: model.PaymentMethodAlipay,
			IsEnabled:     false,
			DisplayName:   "支付宝",
			DisplayOrder:  1,
			Icon:          "/static/icons/alipay.png",
			MinAmount:     decimal.NewFromFloat(0.01),
			MaxAmount:     decimal.NewFromFloat(50000),
		},
		{
			PaymentMethod: model.PaymentMethodWechat,
			IsEnabled:     false,
			DisplayName:   "微信支付",
			DisplayOrder:  2,
			Icon:          "/static/icons/wechat.png",
			MinAmount:     decimal.NewFromFloat(0.01),
			MaxAmount:     decimal.NewFromFloat(50000),
		},
		{
			PaymentMethod: model.PaymentMethodUnionPay,
			IsEnabled:     false,
			DisplayName:   "银联支付",
			DisplayOrder:  3,
			Icon:          "/static/icons/unionpay.png",
			MinAmount:     decimal.NewFromFloat(0.01),
			MaxAmount:     decimal.NewFromFloat(50000),
		},
		{
			PaymentMethod: model.PaymentMethodBalance,
			IsEnabled:     true,
			DisplayName:   "余额支付",
			DisplayOrder:  4,
			Icon:          "/static/icons/balance.png",
			MinAmount:     decimal.NewFromFloat(0.01),
			MaxAmount:     decimal.NewFromFloat(10000),
		},
	}

	for _, config := range defaultConfigs {
		// 检查是否已存在
		var existingConfig model.PaymentConfig
		err := cm.db.Where("payment_method = ?", config.PaymentMethod).First(&existingConfig).Error
		if err == nil {
			continue // 已存在，跳过
		}

		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("查询支付配置失败: %v", err)
		}

		// 创建默认配置
		if err := cm.CreateMethodConfig(&config); err != nil {
			return fmt.Errorf("创建默认配置失败: %v", err)
		}
	}

	logger.Info("默认支付配置初始化完成")
	return nil
}

// startAutoUpdate 启动自动更新
func (cm *ConfigManager) startAutoUpdate() {
	cm.updateTicker = time.NewTicker(5 * time.Minute) // 每5分钟更新一次

	go func() {
		for range cm.updateTicker.C {
			if err := cm.LoadDBConfigs(); err != nil {
				logger.Error("自动更新配置失败", zap.Error(err))
			}
		}
	}()
}

// Stop 停止配置管理器
func (cm *ConfigManager) Stop() {
	if cm.updateTicker != nil {
		cm.updateTicker.Stop()
	}
	logger.Info("配置管理器已停止")
}

// GetConfigSummary 获取配置摘要
func (cm *ConfigManager) GetConfigSummary() map[string]interface{} {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	summary := map[string]interface{}{
		"environment":      cm.config.Environment,
		"debug":            cm.config.Debug,
		"default_currency": cm.config.DefaultCurrency,
		"last_update":      cm.lastUpdate,
		"enabled_methods":  len(cm.GetEnabledMethods()),
		"total_methods":    len(cm.dbConfigs),
		"methods":          make(map[string]interface{}),
	}

	// 添加各支付方式状态
	for method, config := range cm.dbConfigs {
		summary["methods"].(map[string]interface{})[string(method)] = map[string]interface{}{
			"enabled":      config.IsEnabled,
			"display_name": config.DisplayName,
			"min_amount":   config.MinAmount,
			"max_amount":   config.MaxAmount,
		}
	}

	return summary
}

// ReloadConfig 重新加载配置
func (cm *ConfigManager) ReloadConfig() error {
	logger.Info("开始重新加载配置")

	if err := cm.LoadDBConfigs(); err != nil {
		return fmt.Errorf("重新加载配置失败: %v", err)
	}

	logger.Info("配置重新加载完成")
	return nil
}
