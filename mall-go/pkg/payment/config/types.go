package config

import (
	"time"

	"github.com/shopspring/decimal"
)

// PaymentConfig 支付系统配置
type PaymentConfig struct {
	// 基础配置
	Environment string `json:"environment" yaml:"environment"` // 环境: dev, test, prod
	Debug       bool   `json:"debug" yaml:"debug"`             // 调试模式
	LogLevel    string `json:"log_level" yaml:"log_level"`     // 日志级别

	// 通用配置
	DefaultCurrency string        `json:"default_currency" yaml:"default_currency"` // 默认货币
	DefaultTimeout  time.Duration `json:"default_timeout" yaml:"default_timeout"`   // 默认超时时间
	MaxRetries      int           `json:"max_retries" yaml:"max_retries"`           // 最大重试次数

	// 支付方式配置
	Alipay   AlipayConfig   `json:"alipay" yaml:"alipay"`     // 支付宝配置
	Wechat   WechatConfig   `json:"wechat" yaml:"wechat"`     // 微信支付配置
	UnionPay UnionPayConfig `json:"unionpay" yaml:"unionpay"` // 银联配置

	// 回调配置
	CallbackConfig CallbackConfig `json:"callback" yaml:"callback"` // 回调配置

	// 安全配置
	Security SecurityConfig `json:"security" yaml:"security"` // 安全配置

	// 限额配置
	Limits LimitsConfig `json:"limits" yaml:"limits"` // 限额配置
}

// AlipayConfig 支付宝配置
type AlipayConfig struct {
	Enabled    bool          `json:"enabled" yaml:"enabled"`         // 是否启用
	AppID      string        `json:"app_id" yaml:"app_id"`           // 应用ID
	PrivateKey string        `json:"private_key" yaml:"private_key"` // 应用私钥
	PublicKey  string        `json:"public_key" yaml:"public_key"`   // 支付宝公钥
	SignType   string        `json:"sign_type" yaml:"sign_type"`     // 签名类型
	Format     string        `json:"format" yaml:"format"`           // 数据格式
	Charset    string        `json:"charset" yaml:"charset"`         // 编码格式
	GatewayURL string        `json:"gateway_url" yaml:"gateway_url"` // 网关地址
	NotifyURL  string        `json:"notify_url" yaml:"notify_url"`   // 异步通知地址
	ReturnURL  string        `json:"return_url" yaml:"return_url"`   // 同步跳转地址
	Timeout    time.Duration `json:"timeout" yaml:"timeout"`         // 超时时间
}

// WechatConfig 微信支付配置
type WechatConfig struct {
	Enabled    bool          `json:"enabled" yaml:"enabled"`         // 是否启用
	AppID      string        `json:"app_id" yaml:"app_id"`           // 应用ID
	MchID      string        `json:"mch_id" yaml:"mch_id"`           // 商户号
	APIKey     string        `json:"api_key" yaml:"api_key"`         // API密钥
	CertPath   string        `json:"cert_path" yaml:"cert_path"`     // 证书路径
	KeyPath    string        `json:"key_path" yaml:"key_path"`       // 私钥路径
	SignType   string        `json:"sign_type" yaml:"sign_type"`     // 签名类型
	GatewayURL string        `json:"gateway_url" yaml:"gateway_url"` // 网关地址
	NotifyURL  string        `json:"notify_url" yaml:"notify_url"`   // 异步通知地址
	Timeout    time.Duration `json:"timeout" yaml:"timeout"`         // 超时时间
}

// UnionPayConfig 银联配置
type UnionPayConfig struct {
	Enabled    bool          `json:"enabled" yaml:"enabled"`         // 是否启用
	MerchantID string        `json:"merchant_id" yaml:"merchant_id"` // 商户号
	CertPath   string        `json:"cert_path" yaml:"cert_path"`     // 证书路径
	KeyPath    string        `json:"key_path" yaml:"key_path"`       // 私钥路径
	GatewayURL string        `json:"gateway_url" yaml:"gateway_url"` // 网关地址
	NotifyURL  string        `json:"notify_url" yaml:"notify_url"`   // 异步通知地址
	Timeout    time.Duration `json:"timeout" yaml:"timeout"`         // 超时时间
}

// CallbackConfig 回调配置
type CallbackConfig struct {
	MaxRetries      int           `json:"max_retries" yaml:"max_retries"`           // 最大重试次数
	RetryInterval   time.Duration `json:"retry_interval" yaml:"retry_interval"`     // 重试间隔
	Timeout         time.Duration `json:"timeout" yaml:"timeout"`                   // 超时时间
	VerifySignature bool          `json:"verify_signature" yaml:"verify_signature"` // 验证签名
	AllowedIPs      []string      `json:"allowed_ips" yaml:"allowed_ips"`           // 允许的IP
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	EnableSignature bool          `json:"enable_signature" yaml:"enable_signature"` // 启用签名验证
	EnableEncrypt   bool          `json:"enable_encrypt" yaml:"enable_encrypt"`     // 启用加密
	SecretKey       string        `json:"secret_key" yaml:"secret_key"`             // 密钥
	TokenExpiry     time.Duration `json:"token_expiry" yaml:"token_expiry"`         // Token过期时间
	MaxRequestSize  int64         `json:"max_request_size" yaml:"max_request_size"` // 最大请求大小
	RateLimitRPS    int           `json:"rate_limit_rps" yaml:"rate_limit_rps"`     // 限流RPS
	AllowedIPs      []string      `json:"allowed_ips" yaml:"allowed_ips"`           // 允许的IP列表
}

// LimitsConfig 限额配置
type LimitsConfig struct {
	// 单笔限额
	SingleMinAmount decimal.Decimal `json:"single_min_amount" yaml:"single_min_amount"` // 单笔最小金额
	SingleMaxAmount decimal.Decimal `json:"single_max_amount" yaml:"single_max_amount"` // 单笔最大金额

	// 日限额
	DailyMaxAmount decimal.Decimal `json:"daily_max_amount" yaml:"daily_max_amount"` // 日最大金额
	DailyMaxCount  int             `json:"daily_max_count" yaml:"daily_max_count"`   // 日最大笔数

	// 月限额
	MonthlyMaxAmount decimal.Decimal `json:"monthly_max_amount" yaml:"monthly_max_amount"` // 月最大金额
	MonthlyMaxCount  int             `json:"monthly_max_count" yaml:"monthly_max_count"`   // 月最大笔数

	// 支付方式限额
	MethodLimits map[string]MethodLimit `json:"method_limits" yaml:"method_limits"` // 支付方式限额
}

// MethodLimit 支付方式限额
type MethodLimit struct {
	MinAmount      decimal.Decimal `json:"min_amount" yaml:"min_amount"`             // 最小金额
	MaxAmount      decimal.Decimal `json:"max_amount" yaml:"max_amount"`             // 最大金额
	DailyMaxAmount decimal.Decimal `json:"daily_max_amount" yaml:"daily_max_amount"` // 日最大金额
	DailyMaxCount  int             `json:"daily_max_count" yaml:"daily_max_count"`   // 日最大笔数
}
