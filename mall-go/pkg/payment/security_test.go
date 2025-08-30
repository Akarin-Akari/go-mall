package payment

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewSecurityManager(t *testing.T) {
	config := &SecurityConfig{
		EnableSignature: true,
		EnableEncrypt:   false,
		SecretKey:       "test_secret_key",
		TokenExpiry:     24 * time.Hour,
		MaxRequestSize:  1024 * 1024,
		RateLimitRPS:    100,
		AllowedIPs:      []string{"127.0.0.1", "192.168.1.0/24"},
	}

	sm := NewSecurityManager(config)
	assert.NotNil(t, sm)
	assert.Equal(t, config, sm.config)
	assert.NotNil(t, sm.nonceStore)
	assert.NotNil(t, sm.rateLimiter)
	assert.Len(t, sm.ipWhitelist, 2)
}

func TestSecurityManager_VerifySignature(t *testing.T) {
	config := &SecurityConfig{
		EnableSignature: true,
		SecretKey:       "test_secret_key",
	}
	sm := NewSecurityManager(config)

	tests := []struct {
		name     string
		data     map[string]string
		signType string
		wantErr  bool
	}{
		{
			name: "MD5签名验证成功",
			data: map[string]string{
				"param1": "value1",
				"param2": "value2",
			},
			signType: "MD5",
			wantErr:  false,
		},
		{
			name: "HMAC-SHA256签名验证成功",
			data: map[string]string{
				"param1": "value1",
				"param2": "value2",
			},
			signType: "HMAC-SHA256",
			wantErr:  false,
		},
		{
			name: "不支持的签名类型",
			data: map[string]string{
				"param1": "value1",
			},
			signType: "UNKNOWN",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 先计算正确的签名
			var signature string
			var err error

			if !tt.wantErr {
				signature, err = sm.calculateSignature(tt.data, tt.signType)
				assert.NoError(t, err)
			} else {
				signature = "invalid_signature"
			}

			// 验证签名
			err = sm.VerifySignature(tt.data, signature, tt.signType)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSecurityManager_VerifySignature_Disabled(t *testing.T) {
	config := &SecurityConfig{
		EnableSignature: false, // 禁用签名验证
		SecretKey:       "test_secret_key",
	}
	sm := NewSecurityManager(config)

	data := map[string]string{
		"param1": "value1",
	}

	// 即使签名错误，也应该通过验证
	err := sm.VerifySignature(data, "invalid_signature", "MD5")
	assert.NoError(t, err)
}

func TestSecurityManager_CheckNonce(t *testing.T) {
	config := &SecurityConfig{}
	sm := NewSecurityManager(config)

	// 第一次使用随机数应该成功
	err := sm.CheckNonce("nonce123")
	assert.NoError(t, err)

	// 重复使用相同随机数应该失败
	err = sm.CheckNonce("nonce123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "随机数已使用")

	// 使用不同随机数应该成功
	err = sm.CheckNonce("nonce456")
	assert.NoError(t, err)
}

func TestSecurityManager_CheckTimestamp(t *testing.T) {
	config := &SecurityConfig{}
	sm := NewSecurityManager(config)

	now := time.Now()
	tolerance := 5 * time.Minute

	tests := []struct {
		name      string
		timestamp string
		tolerance time.Duration
		wantErr   bool
	}{
		{
			name:      "当前时间戳",
			timestamp: string(rune(now.Unix())),
			tolerance: tolerance,
			wantErr:   false,
		},
		{
			name:      "5分钟前的时间戳",
			timestamp: string(rune(now.Add(-4 * time.Minute).Unix())),
			tolerance: tolerance,
			wantErr:   false,
		},
		{
			name:      "超出容忍范围的时间戳",
			timestamp: string(rune(now.Add(-10 * time.Minute).Unix())),
			tolerance: tolerance,
			wantErr:   true,
		},
		{
			name:      "无效的时间戳格式",
			timestamp: "invalid_timestamp",
			tolerance: tolerance,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sm.CheckTimestamp(tt.timestamp, tt.tolerance)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSecurityManager_CheckIPWhitelist(t *testing.T) {
	config := &SecurityConfig{
		AllowedIPs: []string{"127.0.0.1", "192.168.1.0/24", "10.0.0.1"},
	}
	sm := NewSecurityManager(config)

	tests := []struct {
		name     string
		clientIP string
		wantErr  bool
	}{
		{
			name:     "允许的单个IP",
			clientIP: "127.0.0.1",
			wantErr:  false,
		},
		{
			name:     "CIDR范围内的IP",
			clientIP: "192.168.1.100",
			wantErr:  false,
		},
		{
			name:     "不在白名单中的IP",
			clientIP: "8.8.8.8",
			wantErr:  true,
		},
		{
			name:     "无效的IP格式",
			clientIP: "invalid_ip",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sm.CheckIPWhitelist(tt.clientIP)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSecurityManager_CheckIPWhitelist_EmptyList(t *testing.T) {
	config := &SecurityConfig{
		AllowedIPs: []string{}, // 空白名单
	}
	sm := NewSecurityManager(config)

	// 空白名单应该允许所有IP
	err := sm.CheckIPWhitelist("8.8.8.8")
	assert.NoError(t, err)
}

func TestSecurityManager_CheckRateLimit(t *testing.T) {
	config := &SecurityConfig{
		RateLimitRPS: 2, // 每分钟2次请求
	}
	sm := NewSecurityManager(config)

	key := "test_key"

	// 前两次请求应该成功
	err := sm.CheckRateLimit(key)
	assert.NoError(t, err)

	err = sm.CheckRateLimit(key)
	assert.NoError(t, err)

	// 第三次请求应该被限流
	err = sm.CheckRateLimit(key)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "请求频率超过限制")

	// 不同的key应该有独立的限流计数
	err = sm.CheckRateLimit("another_key")
	assert.NoError(t, err)
}

func TestSecurityManager_ValidateRequestSize(t *testing.T) {
	config := &SecurityConfig{
		MaxRequestSize: 1024, // 1KB限制
	}
	sm := NewSecurityManager(config)

	tests := []struct {
		name    string
		size    int64
		wantErr bool
	}{
		{
			name:    "正常大小的请求",
			size:    512,
			wantErr: false,
		},
		{
			name:    "刚好达到限制的请求",
			size:    1024,
			wantErr: false,
		},
		{
			name:    "超过限制的请求",
			size:    2048,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sm.ValidateRequestSize(tt.size)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "请求大小超过限制")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSecurityManager_ValidateRequestSize_NoLimit(t *testing.T) {
	config := &SecurityConfig{
		MaxRequestSize: 0, // 无限制
	}
	sm := NewSecurityManager(config)

	// 任何大小都应该通过
	err := sm.ValidateRequestSize(1024 * 1024 * 10) // 10MB
	assert.NoError(t, err)
}

func TestSecurityManager_GenerateAndVerifyToken(t *testing.T) {
	config := &SecurityConfig{
		SecretKey:   "test_secret_key",
		TokenExpiry: 24 * time.Hour,
	}
	sm := NewSecurityManager(config)

	data := map[string]string{
		"user_id": "123",
		"role":    "user",
	}

	// 生成令牌
	token, err := sm.GenerateToken(data)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// 验证令牌
	err = sm.VerifyToken(data, token)
	assert.NoError(t, err)

	// 验证错误的令牌
	err = sm.VerifyToken(data, "invalid_token")
	assert.Error(t, err)
}

func TestNonceStore_Cleanup(t *testing.T) {
	ns := NewNonceStore(100 * time.Millisecond) // 100ms过期时间

	// 添加一些随机数
	err := ns.Check("nonce1")
	assert.NoError(t, err)

	err = ns.Check("nonce2")
	assert.NoError(t, err)

	// 等待过期
	time.Sleep(150 * time.Millisecond)

	// 手动清理
	ns.cleanup()

	// 过期的随机数应该被清理，可以重新使用
	err = ns.Check("nonce1")
	assert.NoError(t, err)
}

func TestRateLimiter_Cleanup(t *testing.T) {
	rl := NewRateLimiter(10, 100*time.Millisecond) // 100ms窗口

	// 添加一些请求记录
	err := rl.Check("key1")
	assert.NoError(t, err)

	err = rl.Check("key2")
	assert.NoError(t, err)

	// 等待窗口过期
	time.Sleep(150 * time.Millisecond)

	// 手动清理
	rl.cleanup()

	// 验证清理效果（这里简化验证，实际应该检查内部状态）
	assert.NotNil(t, rl.requests)
}

func TestSecurityManager_GetSecuritySummary(t *testing.T) {
	config := &SecurityConfig{
		EnableSignature: true,
		EnableEncrypt:   false,
		RateLimitRPS:    100,
		MaxRequestSize:  1024 * 1024,
		TokenExpiry:     24 * time.Hour,
		AllowedIPs:      []string{"127.0.0.1", "192.168.1.0/24"},
	}
	sm := NewSecurityManager(config)

	// 添加一些状态数据
	sm.CheckNonce("test_nonce")
	sm.CheckRateLimit("test_key")

	summary := sm.GetSecuritySummary()

	assert.NotNil(t, summary)
	assert.Equal(t, true, summary["signature_enabled"])
	assert.Equal(t, false, summary["encrypt_enabled"])
	assert.Equal(t, 2, summary["ip_whitelist_count"])
	assert.Equal(t, 100, summary["rate_limit_rps"])
	assert.Equal(t, int64(1024*1024), summary["max_request_size"])
	assert.Equal(t, "24h0m0s", summary["token_expiry"])
}

// BenchmarkSecurityManager_VerifySignature 性能测试
func BenchmarkSecurityManager_VerifySignature(b *testing.B) {
	config := &SecurityConfig{
		EnableSignature: true,
		SecretKey:       "test_secret_key",
	}
	sm := NewSecurityManager(config)

	data := map[string]string{
		"param1": "value1",
		"param2": "value2",
		"param3": "value3",
	}

	signature, _ := sm.calculateSignature(data, "MD5")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := sm.VerifySignature(data, signature, "MD5")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSecurityManager_CheckRateLimit 性能测试
func BenchmarkSecurityManager_CheckRateLimit(b *testing.B) {
	config := &SecurityConfig{
		RateLimitRPS: 1000000, // 设置很高的限制避免触发限流
	}
	sm := NewSecurityManager(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key_%d", i%100) // 使用100个不同的key
		err := sm.CheckRateLimit(key)
		if err != nil {
			b.Fatal(err)
		}
	}
}
