package config

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// AddressConfig 地址管理配置
type AddressConfig struct {
	// 业务配置
	MaxAddressPerUser int `yaml:"max_address_per_user" default:"20"`
	MaxPageSize       int `yaml:"max_page_size" default:"100"`
	
	// 验证规则配置
	PhoneRegexPattern     string `yaml:"phone_regex_pattern" default:"^1[3-9]\\d{9}$"`
	PostalCodeRegexPattern string `yaml:"postal_code_regex_pattern" default:"^\\d{6}$"`
	
	// 操作超时配置
	DatabaseTimeout time.Duration `yaml:"database_timeout" default:"5s"`
	
	// 日志配置
	EnableDetailedLog bool `yaml:"enable_detailed_log" default:"true"`
	LogLevel          string `yaml:"log_level" default:"info"`
}

// DefaultAddressConfig 返回默认配置
func DefaultAddressConfig() *AddressConfig {
	return &AddressConfig{
		MaxAddressPerUser:      20,
		MaxPageSize:            100,
		PhoneRegexPattern:      "^1[3-9]\\d{9}$",
		PostalCodeRegexPattern: "^\\d{6}$",
		DatabaseTimeout:        5 * time.Second,
		EnableDetailedLog:      true,
		LogLevel:               "info",
	}
}

// LoadAddressConfig 从配置文件加载地址配置
func LoadAddressConfig(configPath string) (*AddressConfig, error) {
	// 如果配置文件不存在，返回默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultAddressConfig(), nil
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// 解析YAML配置
	config := DefaultAddressConfig() // 先设置默认值
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}

	// 验证配置的合理性
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// LoadAddressConfigFromDir 从指定目录加载地址配置
func LoadAddressConfigFromDir(configDir string) (*AddressConfig, error) {
	configPath := filepath.Join(configDir, "address.yaml")
	return LoadAddressConfig(configPath)
}

// Validate 验证配置的合理性
func (c *AddressConfig) Validate() error {
	if c.MaxAddressPerUser <= 0 || c.MaxAddressPerUser > 1000 {
		c.MaxAddressPerUser = 20 // 重置为默认值
	}
	
	if c.MaxPageSize <= 0 || c.MaxPageSize > 1000 {
		c.MaxPageSize = 100 // 重置为默认值
	}
	
	if c.PhoneRegexPattern == "" {
		c.PhoneRegexPattern = "^1[3-9]\\d{9}$"
	}
	
	if c.PostalCodeRegexPattern == "" {
		c.PostalCodeRegexPattern = "^\\d{6}$"
	}
	
	if c.DatabaseTimeout <= 0 {
		c.DatabaseTimeout = 5 * time.Second
	}
	
	if c.LogLevel == "" {
		c.LogLevel = "info"
	}
	
	return nil
}

// ToYAML 将配置导出为YAML格式
func (c *AddressConfig) ToYAML() ([]byte, error) {
	return yaml.Marshal(c)
}

// SaveToFile 将配置保存到文件
func (c *AddressConfig) SaveToFile(configPath string) error {
	// 确保目录存在
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 转换为YAML
	data, err := c.ToYAML()
	if err != nil {
		return err
	}

	// 写入文件
	return os.WriteFile(configPath, data, 0644)
}

// Clone 克隆配置
func (c *AddressConfig) Clone() *AddressConfig {
	return &AddressConfig{
		MaxAddressPerUser:      c.MaxAddressPerUser,
		MaxPageSize:            c.MaxPageSize,
		PhoneRegexPattern:      c.PhoneRegexPattern,
		PostalCodeRegexPattern: c.PostalCodeRegexPattern,
		DatabaseTimeout:        c.DatabaseTimeout,
		EnableDetailedLog:      c.EnableDetailedLog,
		LogLevel:               c.LogLevel,
	}
}

// IsValidLogLevel 检查日志级别是否有效
func (c *AddressConfig) IsValidLogLevel() bool {
	validLevels := []string{"debug", "info", "warn", "error", "fatal"}
	for _, level := range validLevels {
		if c.LogLevel == level {
			return true
		}
	}
	return false
}

// GetEffectiveMaxPageSize 获取有效的最大页面大小
func (c *AddressConfig) GetEffectiveMaxPageSize(requestedSize int) int {
	if requestedSize <= 0 {
		return 10 // 默认页面大小
	}
	if requestedSize > c.MaxPageSize {
		return c.MaxPageSize
	}
	return requestedSize
}
