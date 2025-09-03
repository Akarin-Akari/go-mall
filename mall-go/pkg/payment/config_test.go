package payment

import (
	"os"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestDefaultPaymentConfig(t *testing.T) {
	config := DefaultPaymentConfig()

	assert.NotNil(t, config)
	assert.Equal(t, "dev", config.Environment)
	assert.True(t, config.Debug)
	assert.Equal(t, "info", config.LogLevel)
	assert.Equal(t, "CNY", config.DefaultCurrency)
	assert.Equal(t, 30*time.Second, config.DefaultTimeout)
	assert.Equal(t, 3, config.MaxRetries)

	// 检查支付宝配置
	assert.False(t, config.Alipay.Enabled)
	assert.Equal(t, "RSA2", config.Alipay.SignType)
	assert.Equal(t, "JSON", config.Alipay.Format)
	assert.Equal(t, "utf-8", config.Alipay.Charset)

	// 检查微信支付配置
	assert.False(t, config.Wechat.Enabled)
	assert.Equal(t, "MD5", config.Wechat.SignType)

	// 检查限额配置
	assert.Equal(t, decimal.NewFromFloat(0.01), config.Limits.MinAmount)
	assert.Equal(t, decimal.NewFromFloat(50000), config.Limits.MaxAmount)
}

func TestPaymentConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *PaymentConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "默认配置验证通过",
			config:  DefaultPaymentConfig(),
			wantErr: false,
		},
		{
			name: "环境配置为空",
			config: &PaymentConfig{
				Environment: "",
			},
			wantErr: true,
			errMsg:  "环境配置不能为空",
		},
		{
			name: "支付宝配置不完整",
			config: &PaymentConfig{
				Environment: "test",
				Alipay: AlipayConfig{
					Enabled: true,
					AppID:   "", // 缺少AppID
				},
			},
			wantErr: true,
			errMsg:  "支付宝AppID不能为空",
		},
		{
			name: "微信支付配置不完整",
			config: &PaymentConfig{
				Environment: "test",
				Wechat: WechatConfig{
					Enabled: true,
					AppID:   "test_app_id",
					MchID:   "", // 缺少商户号
				},
			},
			wantErr: true,
			errMsg:  "微信商户号不能为空",
		},
		{
			name: "完整的支付宝配置",
			config: &PaymentConfig{
				Environment: "test",
				Alipay: AlipayConfig{
					Enabled:    true,
					AppID:      "test_app_id",
					PrivateKey: "test_private_key",
					PublicKey:  "test_public_key",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPaymentConfig_IsProduction(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		want        bool
	}{
		{"生产环境 prod", "prod", true},
		{"生产环境 production", "production", true},
		{"测试环境", "test", false},
		{"开发环境", "dev", false},
		{"其他环境", "staging", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &PaymentConfig{Environment: tt.environment}
			assert.Equal(t, tt.want, config.IsProduction())
		})
	}
}

func TestPaymentConfig_IsTest(t *testing.T) {
	config := &PaymentConfig{Environment: "test"}
	assert.True(t, config.IsTest())

	config.Environment = "prod"
	assert.False(t, config.IsTest())
}

func TestPaymentConfig_IsDevelopment(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		want        bool
	}{
		{"开发环境 dev", "dev", true},
		{"开发环境 development", "development", true},
		{"生产环境", "prod", false},
		{"测试环境", "test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &PaymentConfig{Environment: tt.environment}
			assert.Equal(t, tt.want, config.IsDevelopment())
		})
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	// 保存原始环境变量
	originalEnv := map[string]string{
		"PAYMENT_ENVIRONMENT": os.Getenv("PAYMENT_ENVIRONMENT"),
		"PAYMENT_DEBUG":       os.Getenv("PAYMENT_DEBUG"),
		"ALIPAY_APP_ID":       os.Getenv("ALIPAY_APP_ID"),
		"ALIPAY_PRIVATE_KEY":  os.Getenv("ALIPAY_PRIVATE_KEY"),
		"WECHAT_APP_ID":       os.Getenv("WECHAT_APP_ID"),
		"WECHAT_MCH_ID":       os.Getenv("WECHAT_MCH_ID"),
	}

	// 清理环境变量
	defer func() {
		for key, value := range originalEnv {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	// 设置测试环境变量
	os.Setenv("PAYMENT_ENVIRONMENT", "test")
	os.Setenv("PAYMENT_DEBUG", "true")
	os.Setenv("ALIPAY_APP_ID", "test_alipay_app_id")
	os.Setenv("ALIPAY_PRIVATE_KEY", "test_private_key")
	os.Setenv("WECHAT_APP_ID", "test_wechat_app_id")
	os.Setenv("WECHAT_MCH_ID", "test_mch_id")

	config := LoadConfigFromEnv()

	assert.Equal(t, "test", config.Environment)
	assert.True(t, config.Debug)
	assert.Equal(t, "test_alipay_app_id", config.Alipay.AppID)
	assert.Equal(t, "test_private_key", config.Alipay.PrivateKey)
	assert.True(t, config.Alipay.Enabled)
	assert.Equal(t, "test_wechat_app_id", config.Wechat.AppID)
	assert.Equal(t, "test_mch_id", config.Wechat.MchID)
	assert.True(t, config.Wechat.Enabled)
}

func TestLoadConfigFromFile(t *testing.T) {
	// 测试不存在的文件
	config, err := LoadConfigFromFile("nonexistent.json")
	assert.NoError(t, err)
	assert.NotNil(t, config)
	// 应该返回默认配置
	assert.Equal(t, "dev", config.Environment)

	// 创建临时配置文件
	tempFile := "test_config.json"
	configJSON := `{
		"environment": "test",
		"debug": false,
		"default_currency": "USD",
		"alipay": {
			"enabled": true,
			"app_id": "test_app_id"
		}
	}`

	err = os.WriteFile(tempFile, []byte(configJSON), 0644)
	assert.NoError(t, err)
	defer os.Remove(tempFile)

	// 测试加载配置文件
	config, err = LoadConfigFromFile(tempFile)
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "test", config.Environment)
	assert.False(t, config.Debug)
	assert.Equal(t, "USD", config.DefaultCurrency)
	assert.True(t, config.Alipay.Enabled)
	assert.Equal(t, "test_app_id", config.Alipay.AppID)
}

func TestPaymentConfig_SaveToFile(t *testing.T) {
	config := DefaultPaymentConfig()
	config.Environment = "test"
	config.Debug = false

	tempFile := "test_save_config.json"
	defer os.Remove(tempFile)

	// 保存配置到文件
	err := config.SaveToFile(tempFile)
	assert.NoError(t, err)

	// 验证文件是否存在
	_, err = os.Stat(tempFile)
	assert.NoError(t, err)

	// 重新加载配置验证
	loadedConfig, err := LoadConfigFromFile(tempFile)
	assert.NoError(t, err)
	assert.Equal(t, config.Environment, loadedConfig.Environment)
	assert.Equal(t, config.Debug, loadedConfig.Debug)
}

func TestPaymentConfig_GetMethodLimit(t *testing.T) {
	config := DefaultPaymentConfig()

	// 测试存在的支付方式限额
	limit := config.GetMethodLimit("alipay")
	assert.Equal(t, decimal.NewFromFloat(0.01), limit.MinAmount)
	assert.Equal(t, decimal.NewFromFloat(50000), limit.MaxAmount)

	// 测试不存在的支付方式，应该返回默认限额
	limit = config.GetMethodLimit("unknown")
	assert.Equal(t, config.Limits.MinAmount, limit.MinAmount)
	assert.Equal(t, config.Limits.MaxAmount, limit.MaxAmount)
}

// BenchmarkLoadConfigFromEnv 性能测试
func BenchmarkLoadConfigFromEnv(b *testing.B) {
	// 设置一些环境变量
	os.Setenv("PAYMENT_ENVIRONMENT", "test")
	os.Setenv("ALIPAY_APP_ID", "test_app_id")
	defer func() {
		os.Unsetenv("PAYMENT_ENVIRONMENT")
		os.Unsetenv("ALIPAY_APP_ID")
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config := LoadConfigFromEnv()
		_ = config
	}
}

// BenchmarkPaymentConfig_Validate 性能测试
func BenchmarkPaymentConfig_Validate(b *testing.B) {
	config := DefaultPaymentConfig()
	config.Alipay.Enabled = true
	config.Alipay.AppID = "test_app_id"
	config.Alipay.PrivateKey = "test_private_key"
	config.Alipay.PublicKey = "test_public_key"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := config.Validate()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestPaymentConfig_AutoCorrection(t *testing.T) {
	config := &PaymentConfig{}

	// 验证自动修正功能
	err := config.Validate()
	assert.NoError(t, err)

	// 检查默认值是否被设置
	assert.Equal(t, "CNY", config.DefaultCurrency)
	assert.Equal(t, 30*time.Second, config.DefaultTimeout)
	assert.Equal(t, 3, config.MaxRetries)
}

func TestPaymentConfig_EdgeCases(t *testing.T) {
	// 测试边界情况
	config := &PaymentConfig{
		Environment:    "test",
		DefaultTimeout: -1 * time.Second, // 负值
		MaxRetries:     -1,               // 负值
	}

	err := config.Validate()
	assert.NoError(t, err)

	// 验证负值被修正为默认值
	assert.Equal(t, 30*time.Second, config.DefaultTimeout)
	assert.Equal(t, 3, config.MaxRetries)
}
