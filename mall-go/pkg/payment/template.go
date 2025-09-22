package payment

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/shopspring/decimal"
	"mall-go/internal/model"
)

// ConfigTemplate 配置模板管理器
type ConfigTemplate struct {
	Environment string `json:"environment"`
	Templates   map[string]*PaymentConfig `json:"templates"`
}

// GetDevelopmentTemplate 获取开发环境配置模板
func GetDevelopmentTemplate() *PaymentConfig {
	return &PaymentConfig{
		Environment:     "dev",
		Debug:           true,
		LogLevel:        "debug",
		DefaultCurrency: "CNY",
		DefaultTimeout:  30 * time.Second,
		MaxRetries:      3,

		Alipay: AlipayConfig{
			Enabled:    true,
			AppID:      "2021000000000000", // 沙箱AppID示例
			SignType:   "RSA2",
			Format:     "JSON",
			Charset:    "utf-8",
			GatewayURL: "https://openapi.alipaydev.com/gateway.do", // 沙箱环境
			NotifyURL:  "http://localhost:8080/api/v1/payment/alipay/notify",
			ReturnURL:  "http://localhost:3000/payment/success",
			Timeout:    30 * time.Second,
		},

		Wechat: WechatConfig{
			Enabled:    true,
			AppID:      "wx1234567890abcdef", // 测试AppID示例
			MchID:      "1234567890",
			SignType:   "MD5",
			GatewayURL: "https://api.mch.weixin.qq.com",
			NotifyURL:  "http://localhost:8080/api/v1/payment/wechat/notify",
			Timeout:    30 * time.Second,
		},

		UnionPay: UnionPayConfig{
			Enabled:    false, // 开发环境默认关闭银联
			GatewayURL: "https://gateway.test.95516.com", // 测试环境
			Timeout:    30 * time.Second,
		},

		CallbackConfig: CallbackConfig{
			MaxRetries:      3,
			RetryInterval:   5 * time.Second,
			Timeout:         10 * time.Second,
			VerifySignature: true,
			AllowedIPs:      []string{}, // 开发环境允许所有IP
		},

		Security: SecurityConfig{
			EnableSignature: true,
			EnableEncrypt:   false,
			TokenExpiry:     24 * time.Hour,
			MaxRequestSize:  1024 * 1024, // 1MB
			RateLimitRPS:    1000,        // 开发环境更宽松的限流
		},

		Limits: LimitsConfig{
			MinAmount:        decimal.NewFromFloat(0.01),
			MaxAmount:        decimal.NewFromFloat(10000), // 开发环境较小限额
			DailyMaxAmount:   decimal.NewFromFloat(50000),
			DailyMaxCount:    500,
			MonthlyMaxAmount: decimal.NewFromFloat(200000),
			MonthlyMaxCount:  5000,
			MethodLimits: map[model.PaymentMethod]MethodLimit{
				model.PaymentMethodAlipay: {
					MinAmount:      decimal.NewFromFloat(0.01),
					MaxAmount:      decimal.NewFromFloat(10000),
					DailyMaxAmount: decimal.NewFromFloat(50000),
					DailyMaxCount:  500,
				},
				model.PaymentMethodWechat: {
					MinAmount:      decimal.NewFromFloat(0.01),
					MaxAmount:      decimal.NewFromFloat(10000),
					DailyMaxAmount: decimal.NewFromFloat(50000),
					DailyMaxCount:  500,
				},
			},
		},
	}
}

// GetTestTemplate 获取测试环境配置模板
func GetTestTemplate() *PaymentConfig {
	config := GetDevelopmentTemplate()
	config.Environment = "test"
	config.Debug = true
	config.LogLevel = "info"

	// 测试环境使用更真实的配置但仍然是沙箱
	config.Alipay.NotifyURL = "https://test-api.yourdomain.com/api/v1/payment/alipay/notify"
	config.Alipay.ReturnURL = "https://test.yourdomain.com/payment/success"
	config.Wechat.NotifyURL = "https://test-api.yourdomain.com/api/v1/payment/wechat/notify"

	// 测试环境更严格的安全配置
	config.Security.RateLimitRPS = 500
	config.CallbackConfig.AllowedIPs = []string{
		"110.75.143.101",  // 支付宝回调IP段示例
		"110.75.143.102",
		"182.254.11.170",  // 微信回调IP段示例
		"182.254.11.171",
	}

	return config
}

// GetProductionTemplate 获取生产环境配置模板
func GetProductionTemplate() *PaymentConfig {
	return &PaymentConfig{
		Environment:     "prod",
		Debug:           false,
		LogLevel:        "warn",
		DefaultCurrency: "CNY",
		DefaultTimeout:  30 * time.Second,
		MaxRetries:      3,

		Alipay: AlipayConfig{
			Enabled:    true,
			SignType:   "RSA2",
			Format:     "JSON",
			Charset:    "utf-8",
			GatewayURL: "https://openapi.alipay.com/gateway.do", // 生产环境
			NotifyURL:  "https://api.yourdomain.com/api/v1/payment/alipay/notify",
			ReturnURL:  "https://yourdomain.com/payment/success",
			Timeout:    30 * time.Second,
		},

		Wechat: WechatConfig{
			Enabled:    true,
			SignType:   "MD5",
			GatewayURL: "https://api.mch.weixin.qq.com",
			NotifyURL:  "https://api.yourdomain.com/api/v1/payment/wechat/notify",
			Timeout:    30 * time.Second,
		},

		UnionPay: UnionPayConfig{
			Enabled:    true,
			GatewayURL: "https://gateway.95516.com", // 生产环境
			Timeout:    30 * time.Second,
		},

		CallbackConfig: CallbackConfig{
			MaxRetries:      5, // 生产环境更多重试
			RetryInterval:   10 * time.Second,
			Timeout:         15 * time.Second,
			VerifySignature: true,
			AllowedIPs: []string{
				"110.75.143.101",
				"110.75.143.102", 
				"110.75.143.103", // 支付宝生产环境IP
				"182.254.11.170", 
				"182.254.11.171", // 微信生产环境IP
			},
		},

		Security: SecurityConfig{
			EnableSignature: true,
			EnableEncrypt:   true, // 生产环境启用加密
			TokenExpiry:     1 * time.Hour, // 生产环境较短过期时间
			MaxRequestSize:  512 * 1024, // 更严格的请求大小限制
			RateLimitRPS:    200,         // 生产环境严格限流
		},

		Limits: LimitsConfig{
			MinAmount:        decimal.NewFromFloat(0.01),
			MaxAmount:        decimal.NewFromFloat(50000), // 生产环境正常限额
			DailyMaxAmount:   decimal.NewFromFloat(500000),
			DailyMaxCount:    5000,
			MonthlyMaxAmount: decimal.NewFromFloat(5000000),
			MonthlyMaxCount:  50000,
			MethodLimits: map[model.PaymentMethod]MethodLimit{
				model.PaymentMethodAlipay: {
					MinAmount:      decimal.NewFromFloat(0.01),
					MaxAmount:      decimal.NewFromFloat(50000),
					DailyMaxAmount: decimal.NewFromFloat(500000),
					DailyMaxCount:  5000,
				},
				model.PaymentMethodWechat: {
					MinAmount:      decimal.NewFromFloat(0.01),
					MaxAmount:      decimal.NewFromFloat(20000), // 微信单笔限额较小
					DailyMaxAmount: decimal.NewFromFloat(200000),
					DailyMaxCount:  3000,
				},
				model.PaymentMethodUnionPay: {
					MinAmount:      decimal.NewFromFloat(0.01),
					MaxAmount:      decimal.NewFromFloat(100000), // 银联限额较大
					DailyMaxAmount: decimal.NewFromFloat(1000000),
					DailyMaxCount:  10000,
				},
			},
		},
	}
}

// LoadTemplateByEnvironment 根据环境加载配置模板
func LoadTemplateByEnvironment(env string) *PaymentConfig {
	switch env {
	case "dev", "development":
		return GetDevelopmentTemplate()
	case "test", "testing":
		return GetTestTemplate()
	case "prod", "production":
		return GetProductionTemplate()
	default:
		return GetDevelopmentTemplate()
	}
}

// GenerateConfigFile 生成配置文件
func GenerateConfigFile(env string, filePath string) error {
	template := LoadTemplateByEnvironment(env)
	
	data, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置模板失败: %v", err)
	}
	
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}
	
	return nil
}

// ValidateEnvironmentConfig 验证环境配置完整性
func ValidateEnvironmentConfig(config *PaymentConfig) []ValidationError {
	var errors []ValidationError
	
	// 基础配置验证
	if config.Environment == "" {
		errors = append(errors, ValidationError{
			Field:   "environment",
			Message: "环境配置不能为空",
			Code:    "ENV_REQUIRED",
		})
	}
	
	// 生产环境特殊验证
	if config.IsProduction() {
		if config.Debug {
			errors = append(errors, ValidationError{
				Field:   "debug",
				Message: "生产环境不应启用调试模式",
				Code:    "PROD_DEBUG_ENABLED",
			})
		}
		
		if config.Security.EnableEncrypt == false {
			errors = append(errors, ValidationError{
				Field:   "security.enable_encrypt",
				Message: "生产环境必须启用数据加密",
				Code:    "PROD_ENCRYPT_REQUIRED",
			})
		}
		
		if len(config.CallbackConfig.AllowedIPs) == 0 {
			errors = append(errors, ValidationError{
				Field:   "callback.allowed_ips",
				Message: "生产环境必须配置回调IP白名单",
				Code:    "PROD_IP_WHITELIST_REQUIRED",
			})
		}
	}
	
	// 支付宝配置验证
	if config.Alipay.Enabled {
		if config.Alipay.AppID == "" {
			errors = append(errors, ValidationError{
				Field:   "alipay.app_id",
				Message: "支付宝AppID不能为空",
				Code:    "ALIPAY_APPID_REQUIRED",
			})
		}
		
		if config.Alipay.PrivateKey == "" {
			errors = append(errors, ValidationError{
				Field:   "alipay.private_key",
				Message: "支付宝应用私钥不能为空",
				Code:    "ALIPAY_PRIVATE_KEY_REQUIRED",
			})
		}
		
		if config.Alipay.PublicKey == "" {
			errors = append(errors, ValidationError{
				Field:   "alipay.public_key",
				Message: "支付宝公钥不能为空",
				Code:    "ALIPAY_PUBLIC_KEY_REQUIRED",
			})
		}
	}
	
	// 微信支付配置验证
	if config.Wechat.Enabled {
		if config.Wechat.AppID == "" {
			errors = append(errors, ValidationError{
				Field:   "wechat.app_id",
				Message: "微信AppID不能为空",
				Code:    "WECHAT_APPID_REQUIRED",
			})
		}
		
		if config.Wechat.MchID == "" {
			errors = append(errors, ValidationError{
				Field:   "wechat.mch_id",
				Message: "微信商户号不能为空",
				Code:    "WECHAT_MCHID_REQUIRED",
			})
		}
		
		if config.Wechat.APIKey == "" {
			errors = append(errors, ValidationError{
				Field:   "wechat.api_key",
				Message: "微信API密钥不能为空",
				Code:    "WECHAT_APIKEY_REQUIRED",
			})
		}
	}
	
	return errors
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// Error 实现error接口
func (e ValidationError) Error() string {
	return fmt.Sprintf("字段 %s: %s (错误码: %s)", e.Field, e.Message, e.Code)
}