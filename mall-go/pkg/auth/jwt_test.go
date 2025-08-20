package auth

import (
	"testing"
	"time"
	"mall-go/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

// 初始化测试配置
func init() {
	config.GlobalConfig = config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key-for-jwt-testing",
			Expire: "24h",
		},
	}
}

func TestGenerateToken(t *testing.T) {
	userID := uint(1)
	username := "testuser"
	role := "user"

	token, err := GenerateToken(userID, username, role)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	if token == "" {
		t.Error("GenerateToken() returned empty token")
	}

	// 验证生成的token格式
	if len(token) < 100 {
		t.Error("GenerateToken() returned token too short")
	}
}

func TestParseToken(t *testing.T) {
	userID := uint(1)
	username := "testuser"
	role := "user"

	// 生成token
	token, err := GenerateToken(userID, username, role)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	// 解析token
	claims, err := ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}

	// 验证claims内容
	if claims.UserID != userID {
		t.Errorf("ParseToken() UserID = %v, want %v", claims.UserID, userID)
	}
	if claims.Username != username {
		t.Errorf("ParseToken() Username = %v, want %v", claims.Username, username)
	}
	if claims.Role != role {
		t.Errorf("ParseToken() Role = %v, want %v", claims.Role, role)
	}

	// 验证标准claims
	if claims.Issuer != "mall-go" {
		t.Errorf("ParseToken() Issuer = %v, want %v", claims.Issuer, "mall-go")
	}
	if claims.Subject != username {
		t.Errorf("ParseToken() Subject = %v, want %v", claims.Subject, username)
	}
}

func TestValidateToken(t *testing.T) {
	userID := uint(1)
	username := "testuser"
	role := "user"

	// 生成有效token
	validToken, err := GenerateToken(userID, username, role)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	// 测试有效token
	if !ValidateToken(validToken) {
		t.Error("ValidateToken() with valid token should return true")
	}

	// 测试无效token
	invalidToken := "invalid.token.string"
	if ValidateToken(invalidToken) {
		t.Error("ValidateToken() with invalid token should return false")
	}

	// 测试空token
	if ValidateToken("") {
		t.Error("ValidateToken() with empty token should return false")
	}
}

func TestParseTokenWithInvalidSignature(t *testing.T) {
	// 使用不同的密钥生成token
	originalSecret := config.GlobalConfig.JWT.Secret
	config.GlobalConfig.JWT.Secret = "different-secret"

	token, err := GenerateToken(1, "testuser", "user")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	// 恢复原始密钥
	config.GlobalConfig.JWT.Secret = originalSecret

	// 尝试解析用不同密钥生成的token
	_, err = ParseToken(token)
	if err == nil {
		t.Error("ParseToken() with invalid signature should return error")
	}
}

func TestTokenExpiration(t *testing.T) {
	// 设置短过期时间
	originalExpire := config.GlobalConfig.JWT.Expire
	config.GlobalConfig.JWT.Expire = "1ms" // 1毫秒后过期

	token, err := GenerateToken(1, "testuser", "user")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	// 等待token过期
	time.Sleep(10 * time.Millisecond)

	// 验证过期token
	if ValidateToken(token) {
		t.Error("ValidateToken() with expired token should return false")
	}

	// 恢复原始过期时间
	config.GlobalConfig.JWT.Expire = originalExpire
}

func TestRefreshToken(t *testing.T) {
	// 设置较短的过期时间以便测试刷新
	originalExpire := config.GlobalConfig.JWT.Expire
	config.GlobalConfig.JWT.Expire = "2h" // 2小时过期

	userID := uint(1)
	username := "testuser"
	role := "user"

	// 生成原始token
	originalToken, err := GenerateToken(userID, username, role)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	// 创建一个接近过期的token（手动设置过期时间）
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)), // 30分钟后过期
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "mall-go",
			Subject:   username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	nearExpiryToken, err := token.SignedString([]byte(config.GlobalConfig.JWT.Secret))
	if err != nil {
		t.Fatalf("Failed to create near expiry token: %v", err)
	}

	// 尝试刷新token
	newToken, err := RefreshToken(nearExpiryToken)
	if err != nil {
		t.Fatalf("RefreshToken() error = %v", err)
	}

	if newToken == "" {
		t.Error("RefreshToken() returned empty token")
	}

	// 验证新token有效
	if !ValidateToken(newToken) {
		t.Error("RefreshToken() returned invalid token")
	}

	// 验证新token包含相同的用户信息
	newClaims, err := ParseToken(newToken)
	if err != nil {
		t.Fatalf("ParseToken() on refreshed token error = %v", err)
	}

	if newClaims.UserID != userID {
		t.Errorf("Refreshed token UserID = %v, want %v", newClaims.UserID, userID)
	}
	if newClaims.Username != username {
		t.Errorf("Refreshed token Username = %v, want %v", newClaims.Username, username)
	}
	if newClaims.Role != role {
		t.Errorf("Refreshed token Role = %v, want %v", newClaims.Role, role)
	}

	// 恢复原始过期时间
	config.GlobalConfig.JWT.Expire = originalExpire
}

func TestGetClaimsFromToken(t *testing.T) {
	userID := uint(1)
	username := "testuser"
	role := "user"

	token, err := GenerateToken(userID, username, role)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	claims, err := GetClaimsFromToken(token)
	if err != nil {
		t.Fatalf("GetClaimsFromToken() error = %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("GetClaimsFromToken() UserID = %v, want %v", claims.UserID, userID)
	}
	if claims.Username != username {
		t.Errorf("GetClaimsFromToken() Username = %v, want %v", claims.Username, username)
	}
	if claims.Role != role {
		t.Errorf("GetClaimsFromToken() Role = %v, want %v", claims.Role, role)
	}
}

func BenchmarkGenerateToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateToken(1, "testuser", "user")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseToken(b *testing.B) {
	token, err := GenerateToken(1, "testuser", "user")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseToken(token)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidateToken(b *testing.B) {
	token, err := GenerateToken(1, "testuser", "user")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateToken(token)
	}
}
