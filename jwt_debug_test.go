package main

import (
	"fmt"
	"mall-go/internal/config"
	"mall-go/pkg/auth"
)

func main() {
	fmt.Println("🔧 JWT调试测试工具")
	
	// 初始化配置
	config.Load()
	fmt.Printf("JWT Secret: %s\n", config.GlobalConfig.JWT.Secret)
	fmt.Printf("JWT Expire: %s\n", config.GlobalConfig.JWT.Expire)
	
	// 生成管理员token
	token, err := auth.GenerateToken(1, "admin", "admin")
	if err != nil {
		fmt.Printf("❌ 生成token失败: %v\n", err)
		return
	}
	
	fmt.Printf("✅ 生成的token: %s\n", token)
	
	// 解析token
	claims, err := auth.ParseToken(token)
	if err != nil {
		fmt.Printf("❌ 解析token失败: %v\n", err)
		return
	}
	
	fmt.Printf("✅ 解析成功:\n")
	fmt.Printf("   UserID: %d\n", claims.UserID)
	fmt.Printf("   Username: %s\n", claims.Username)
	fmt.Printf("   Role: %s\n", claims.Role)
	
	// 测试用户信息获取
	userID, username, role, err := auth.GetUserInfoFromToken(token)
	if err != nil {
		fmt.Printf("❌ 获取用户信息失败: %v\n", err)
		return
	}
	
	fmt.Printf("✅ 用户信息获取成功:\n")
	fmt.Printf("   UserID: %d\n", userID)
	fmt.Printf("   Username: %s\n", username)
	fmt.Printf("   Role: %s\n", role)
}
