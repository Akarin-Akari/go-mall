package auth

import (
	"errors"
	"time"
	"mall-go/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT声明
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT令牌
func GenerateToken(userID uint, username, role string) (string, error) {
	// 获取JWT配置
	jwtConfig := config.GlobalConfig.JWT
	
	// 解析过期时间
	expireDuration, err := time.ParseDuration(jwtConfig.Expire)
	if err != nil {
		return "", err
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
	tokenString, err := token.SignedString([]byte(jwtConfig.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseToken 解析JWT令牌
func ParseToken(tokenString string) (*Claims, error) {
	// 获取JWT配置
	jwtConfig := config.GlobalConfig.JWT

	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("无效的签名方法")
		}
		return []byte(jwtConfig.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	// 验证令牌
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的令牌")
}

// ValidateToken 验证令牌
func ValidateToken(tokenString string) bool {
	_, err := ParseToken(tokenString)
	return err == nil
}

// RefreshToken 刷新JWT令牌
// 如果token在过期前的刷新窗口期内，则生成新token
func RefreshToken(tokenString string) (string, error) {
	// 解析现有token（即使过期也要解析）
	claims, err := parseTokenIgnoreExpiry(tokenString)
	if err != nil {
		return "", err
	}

	// 检查token是否在可刷新的时间窗口内
	// 这里设置为过期前1小时内可以刷新
	refreshWindow := time.Hour
	if time.Until(claims.ExpiresAt.Time) > refreshWindow {
		return "", errors.New("token还未到刷新时间")
	}

	// 生成新的token
	return GenerateToken(claims.UserID, claims.Username, claims.Role)
}

// parseTokenIgnoreExpiry 解析token但忽略过期时间
func parseTokenIgnoreExpiry(tokenString string) (*Claims, error) {
	jwtConfig := config.GlobalConfig.JWT

	// 解析令牌，忽略过期时间
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("无效的签名方法")
		}
		return []byte(jwtConfig.Secret), nil
	}, jwt.WithoutClaimsValidation())

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok {
		return claims, nil
	}

	return nil, errors.New("无效的令牌格式")
}

// GetClaimsFromToken 从token中获取Claims（不验证过期时间）
func GetClaimsFromToken(tokenString string) (*Claims, error) {
	return parseTokenIgnoreExpiry(tokenString)
}
