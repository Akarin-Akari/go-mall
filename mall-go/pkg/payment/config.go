package payment

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"mall-go/internal/model"

	"github.com/shopspring/decimal"
)

// PaymentConfig 支付系统配置
type PaymentConfig struct {
	// 基础配置
	Environment     string                    `json:"environment" yaml:"environment"`         // 环境: dev, test, prod
	Debug           bool                      `json:"debug" yaml:"debug"`                     // 调试模式
	LogLevel        string                    `json:"log_level" yaml:"log_level"`             // 日志级别
	
	// 通用配置
	DefaultCurrency string                    `json:"default_currency" yaml:"default_currency"` // 默认货币
	DefaultTimeout  time.Duration             `json:"default_timeout" yaml:"default_timeout"`   // 默认超时时间
	MaxRetries      int                       `json:"max_retries" yaml:"max_retries"`           // 最大重试次数
	
	// 支付方式配置
	Alipay          AlipayConfig              `json:"alipay" yaml:"alipay"`                   // 支付宝配置
	Wechat          WechatConfig              `json:"wechat" yaml:"wechat"`                   // 微信支付配置
	UnionPay        UnionPayConfig            `json:"unionpay" yaml:"unionpay"`               // 银联配置
	
	// 回调配置
	CallbackConfig  CallbackConfig            `json:"callback" yaml:"callback"`               // 回调配置
	
	// 安全配置
	Security        SecurityConfig            `json:"security" yaml:"security"`               // 安全配置
	
	// 限额配置
	Limits          LimitsConfig              `json:"limits" yaml:"limits"`                   // 限额配置
}

// AlipayConfig 支付宝配置
type AlipayConfig struct {
	Enabled         bool                      `json:"enabled" yaml:"enabled"`                 // 是否启用
	AppID           string                    `json:"app_id" yaml:"app_id"`                   // 应用ID
	PrivateKey      string                    `json:"private_key" yaml:"private_key"`         // 应用私钥
	PublicKey       string                    `json:"public_key" yaml:"public_key"`           // 支付宝公钥
	SignType        string                    `json:"sign_type" yaml:"sign_type"`             // 签名类型
	Format          string                    `json:"format" yaml:"format"`                   // 数据格式
	Charset         string                    `json:"charset" yaml:"charset"`                 // 编码格式
	GatewayURL      string                    `json:"gateway_url" yaml:"gateway_url"`         // 网关地址
	NotifyURL       string                    `json:"notify_url" yaml:"notify_url"`           // 异步通知地址
	ReturnURL       string                    `json:"return_url" yaml:"return_url"`           // 同步跳转地址
	Timeout         time.Duration             `json:"timeout" yaml:"timeout"`                 // 超时时间
}

// WechatConfig 微信支付配置
type WechatConfig struct {
	Enabled         bool                      `json:"enabled" yaml:"enabled"`                 // 是否启用
	AppID           string                    `json:"app_id" yaml:"app_id"`                   // 应用ID
	MchID           string                    `json:"mch_id" yaml:"mch_id"`                   // 商户号
	APIKey          string                    `json:"api_key" yaml:"api_key"`                 // API密钥
	CertPath        string                    `json:"cert_path" yaml:"cert_path"`             // 证书路径
	KeyPath         string                    `json:"key_path" yaml:"key_path"`               // 私钥路径
	SignType        string                    `json:"sign_type" yaml:"sign_type"`             // 签名类型
	GatewayURL      string                    `json:"gateway_url" yaml:"gateway_url"`         // 网关地址
	NotifyURL       string                    `json:"notify_url" yaml:"notify_url"`           // 异步通知地址
	Timeout         time.Duration             `json:"timeout" yaml:"timeout"`                 // 超时时间
}

// UnionPayConfig 银联配置
type UnionPayConfig struct {
	Enabled         bool                      `json:"enabled" yaml:"enabled"`                 // 是否启用
	MerchantID      string                    `json:"merchant_id" yaml:"merchant_id"`         // 商户号
	CertPath        string                    `json:"cert_path" yaml:"cert_path"`             // 证书路径
	KeyPath         string                    `json:"key_path" yaml:"key_path"`               // 私钥路径
	GatewayURL      string                    `json:"gateway_url" yaml:"gateway_url"`         // 网关地址
	NotifyURL       string                    `json:"notify_url" yaml:"notify_url"`           // 异步通知地址
	Timeout         time.Duration             `json:"timeout" yaml:"timeout"`                 // 超时时间
}

// CallbackConfig 回调配置
type CallbackConfig struct {
	MaxRetries      int                       `json:"max_retries" yaml:"max_retries"`         // 最大重试次数
	RetryInterval   time.Duration             `json:"retry_interval" yaml:"retry_interval"`   // 重试间隔
	Timeout         time.Duration             `json:"timeout" yaml:"timeout"`                 // 超时时间
	VerifySignature bool                      `json:"verify_signature" yaml:"verify_signature"` // 验证签名
	AllowedIPs      []string                  `json:"allowed_ips" yaml:"allowed_ips"`         // 允许的IP
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	EnableSignature bool                      `json:"enable_signature" yaml:"enable_signature"` // 启用签名验证
	EnableEncrypt   bool                      `json:"enable_encrypt" yaml:"enable_encrypt"`     // 启用加密
	SecretKey       string                    `json:"secret_key" yaml:"secret_key"`             // 密钥
	TokenExpiry     time.Duration             `json:"token_expiry" yaml:"token_expiry"`         // Token过期时间
	MaxRequestSize  int64                     `json:"max_request_size" yaml:"max_request_size"` // 最大请求大小
	RateLimitRPS    int                       `json:"rate_limit_rps" yaml:"rate_limit_rps"`     // 限流RPS
}

// LimitsConfig 限额配置
type LimitsConfig struct {
	// 单笔限额
	MinAmount       decimal.Decimal           `json:"min_amount" yaml:"min_amount"`             // 最小金额
	MaxAmount       decimal.Decimal           `json:"max_amount" yaml:"max_amount"`             // 最大金额
	
	// 日限额
	DailyMaxAmount  decimal.Decimal           `json:"daily_max_amount" yaml:"daily_max_amount"` // 日最大金额
	DailyMaxCount   int                       `json:"daily_max_count" yaml:"daily_max_count"`   // 日最大笔数
	
	// 月限额
	MonthlyMaxAmount decimal.Decimal          `json:"monthly_max_amount" yaml:"monthly_max_amount"` // 月最大金额
	MonthlyMaxCount  int                      `json:"monthly_max_count" yaml:"monthly_max_count"`   // 月最大笔数
	
	// 支付方式限额
	MethodLimits    map[model.PaymentMethod]MethodLimit `json:"method_limits" yaml:"method_limits"` // 支付方式限额
}

// MethodLimit 支付方式限额
type MethodLimit struct {
	MinAmount       decimal.Decimal           `json:"min_amount" yaml:"min_amount"`             // 最小金额
	MaxAmount       decimal.Decimal           `json:"max_amount" yaml:"max_amount"`             // 最大金额
	DailyMaxAmount  decimal.Decimal           `json:"daily_max_amount" yaml:"daily_max_amount"` // 日最大金额
	DailyMaxCount   int                       `json:"daily_max_count" yaml:"daily_max_count"`   // 日最大笔数
}

// DefaultPaymentConfig 默认支付配置
func DefaultPaymentConfig() *PaymentConfig {
	return &PaymentConfig{
		Environment:     "dev",
		Debug:           true,
		LogLevel:        "info",
		DefaultCurrency: "CNY",
		DefaultTimeout:  30 * time.Second,
		MaxRetries:      3,
		
		Alipay: AlipayConfig{
			Enabled:    false,
			SignType:   "RSA2",
			Format:     "JSON",
			Charset:    "utf-8",
			GatewayURL: "https://openapi.alipaydev.com/gateway.do", // 沙箱环境
			Timeout:    30 * time.Second,
		},
		
		Wechat: WechatConfig{
			Enabled:    false,
			SignType:   "MD5",
			GatewayURL: "https://api.mch.weixin.qq.com",
			Timeout:    30 * time.Second,
		},
		
		UnionPay: UnionPayConfig{
			Enabled:    false,
			GatewayURL: "https://gateway.test.95516.com", // 测试环境
			Timeout:    30 * time.Second,
		},
		
		CallbackConfig: CallbackConfig{
			MaxRetries:      3,
			RetryInterval:   5 * time.Second,
			Timeout:         10 * time.Second,
			VerifySignature: true,
			AllowedIPs:      []string{}, // 空表示允许所有IP
		},
		
		Security: SecurityConfig{
			EnableSignature: true,
			EnableEncrypt:   false,
			TokenExpiry:     24 * time.Hour,
			MaxRequestSize:  1024 * 1024, // 1MB
			RateLimitRPS:    100,
		},
		
		Limits: LimitsConfig{
			MinAmount:        decimal.NewFromFloat(0.01),
			MaxAmount:        decimal.NewFromFloat(50000),
			DailyMaxAmount:   decimal.NewFromFloat(100000),
			DailyMaxCount:    1000,
			MonthlyMaxAmount: decimal.NewFromFloat(1000000),
			MonthlyMaxCount:  10000,
			MethodLimits: map[model.PaymentMethod]MethodLimit{
				model.PaymentMethodAlipay: {
					MinAmount:      decimal.NewFromFloat(0.01),
					MaxAmount:      decimal.NewFromFloat(50000),
					DailyMaxAmount: decimal.NewFromFloat(100000),
					DailyMaxCount:  1000,
				},
				model.PaymentMethodWechat: {
					MinAmount:      decimal.NewFromFloat(0.01),
					MaxAmount:      decimal.NewFromFloat(50000),
					DailyMaxAmount: decimal.NewFromFloat(100000),
					DailyMaxCount:  1000,
				},
			},
		},
	}
}

// LoadConfigFromFile 从文件加载配置
func LoadConfigFromFile(configPath string) (*PaymentConfig, error) {
	// 如果文件不存在，返回默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultPaymentConfig(), nil
	}
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}
	
	config := &PaymentConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}
	
	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %v", err)
	}
	
	return config, nil
}

// LoadConfigFromEnv 从环境变量加载配置
func LoadConfigFromEnv() *PaymentConfig {
	config := DefaultPaymentConfig()
	
	// 从环境变量覆盖配置
	if env := os.Getenv("PAYMENT_ENVIRONMENT"); env != "" {
		config.Environment = env
	}
	
	if debug := os.Getenv("PAYMENT_DEBUG"); debug == "true" {
		config.Debug = true
	}
	
	// 支付宝配置
	if appID := os.Getenv("ALIPAY_APP_ID"); appID != "" {
		config.Alipay.AppID = appID
		config.Alipay.Enabled = true
	}
	
	if privateKey := os.Getenv("ALIPAY_PRIVATE_KEY"); privateKey != "" {
		config.Alipay.PrivateKey = privateKey
	}
	
	if publicKey := os.Getenv("ALIPAY_PUBLIC_KEY"); publicKey != "" {
		config.Alipay.PublicKey = publicKey
	}
	
	// 微信支付配置
	if appID := os.Getenv("WECHAT_APP_ID"); appID != "" {
		config.Wechat.AppID = appID
		config.Wechat.Enabled = true
	}
	
	if mchID := os.Getenv("WECHAT_MCH_ID"); mchID != "" {
		config.Wechat.MchID = mchID
	}
	
	if apiKey := os.Getenv("WECHAT_API_KEY"); apiKey != "" {
		config.Wechat.APIKey = apiKey
	}
	
	return config
}

// SaveToFile 保存配置到文件
func (c *PaymentConfig) SaveToFile(configPath string) error {
	// 创建目录
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}
	
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}
	
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}
	
	return nil
}

// Validate 验证配置
func (c *PaymentConfig) Validate() error {
	if c.Environment == "" {
		return fmt.Errorf("环境配置不能为空")
	}
	
	if c.DefaultCurrency == "" {
		c.DefaultCurrency = "CNY"
	}
	
	if c.DefaultTimeout <= 0 {
		c.DefaultTimeout = 30 * time.Second
	}
	
	if c.MaxRetries <= 0 {
		c.MaxRetries = 3
	}
	
	// 验证支付宝配置
	if c.Alipay.Enabled {
		if c.Alipay.AppID == "" {
			return fmt.Errorf("支付宝AppID不能为空")
		}
		if c.Alipay.PrivateKey == "" {
			return fmt.Errorf("支付宝私钥不能为空")
		}
		if c.Alipay.PublicKey == "" {
			return fmt.Errorf("支付宝公钥不能为空")
		}
	}
	
	// 验证微信支付配置
	if c.Wechat.Enabled {
		if c.Wechat.AppID == "" {
			return fmt.Errorf("微信AppID不能为空")
		}
		if c.Wechat.MchID == "" {
			return fmt.Errorf("微信商户号不能为空")
		}
		if c.Wechat.APIKey == "" {
			return fmt.Errorf("微信API密钥不能为空")
		}
	}
	
	return nil
}

// IsProduction 是否生产环境
func (c *PaymentConfig) IsProduction() bool {
	return c.Environment == "prod" || c.Environment == "production"
}

// IsTest 是否测试环境
func (c *PaymentConfig) IsTest() bool {
	return c.Environment == "test"
}

// IsDevelopment 是否开发环境
func (c *PaymentConfig) IsDevelopment() bool {
	return c.Environment == "dev" || c.Environment == "development"
}

// GetMethodLimit 获取支付方式限额
func (c *PaymentConfig) GetMethodLimit(method model.PaymentMethod) MethodLimit {
	if limit, exists := c.Limits.MethodLimits[method]; exists {
		return limit
	}
	
	// 返回默认限额
	return MethodLimit{
		MinAmount:      c.Limits.MinAmount,
		MaxAmount:      c.Limits.MaxAmount,
		DailyMaxAmount: c.Limits.DailyMaxAmount,
		DailyMaxCount:  c.Limits.DailyMaxCount,
	}
}
