package auth

import (
	"errors"
	"time"

	"mall-go/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT声明结构
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT令牌
func GenerateToken(userID uint, username, role string) (string, error) {
	// 获取配置
	cfg := config.GlobalConfig
	if cfg.JWT.Secret == "" {
		return "", errors.New("JWT密钥未配置")
	}

	// 解析过期时间
	expireDuration, err := time.ParseDuration(cfg.JWT.Expire)
	if err != nil {
		expireDuration = 24 * time.Hour // 默认24小时
	}

	// 创建声明
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "mall-go",
			Subject:   username,
		},
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名令牌
	tokenString, err := token.SignedString([]byte(cfg.JWT.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseToken 解析JWT令牌
func ParseToken(tokenString string) (*Claims, error) {
	// 获取配置
	cfg := config.GlobalConfig
	if cfg.JWT.Secret == "" {
		return nil, errors.New("JWT密钥未配置")
	}

	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("无效的签名方法")
		}
		return []byte(cfg.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	// 验证令牌有效性
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的令牌")
}

// ValidateToken 验证令牌是否有效
func ValidateToken(tokenString string) bool {
	_, err := ParseToken(tokenString)
	return err == nil
}

// RefreshToken 刷新令牌
func RefreshToken(tokenString string) (string, error) {
	// 解析现有令牌
	claims, err := ParseToken(tokenString)
	if err != nil {
		return "", err
	}

	// 检查令牌是否即将过期（在30分钟内过期才允许刷新）
	if time.Until(claims.ExpiresAt.Time) > 30*time.Minute {
		return "", errors.New("令牌尚未到刷新时间")
	}

	// 生成新令牌
	return GenerateToken(claims.UserID, claims.Username, claims.Role)
}

// GetUserInfoFromToken 从令牌中获取用户信息
func GetUserInfoFromToken(tokenString string) (uint, string, string, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return 0, "", "", err
	}

	return claims.UserID, claims.Username, claims.Role, nil
}

// GenerateTokenWithExpiry 生成带有自定义过期时间的JWT令牌
func GenerateTokenWithExpiry(userID uint, username, role string, expiry time.Duration) (string, time.Time, error) {
	// 获取配置
	cfg := config.GlobalConfig
	if cfg.JWT.Secret == "" {
		return "", time.Time{}, errors.New("JWT密钥未配置")
	}

	// 计算过期时间
	expiresAt := time.Now().Add(expiry)

	// 创建声明
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "mall-go",
			Subject:   username,
		},
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名令牌
	tokenString, err := token.SignedString([]byte(cfg.JWT.Secret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// GenerateRefreshToken 生成刷新令牌
func GenerateRefreshToken(userID uint, username, role string) (string, time.Time, error) {
	// 刷新令牌有效期为7天
	return GenerateTokenWithExpiry(userID, username, role, 7*24*time.Hour)
}

// ValidateRefreshToken 验证刷新令牌
func ValidateRefreshToken(tokenString string) (*Claims, error) {
	// 刷新令牌的验证逻辑与普通令牌相同
	return ParseToken(tokenString)
}
