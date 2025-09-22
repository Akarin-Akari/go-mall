package main

import (
	"fmt"
	"mall-go/internal/config"
	"mall-go/internal/model"
	"mall-go/pkg/database"
	"mall-go/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	fmt.Println("🔐 修复testuser用户密码")
	
	// 初始化配置
	config.Load()
	
	// 初始化日志
	logger.Init()
	
	// 初始化数据库
	db := database.Init()
	if db == nil {
		fmt.Println("❌ 数据库初始化失败")
		return
	}
	
	fmt.Println("✅ 数据库连接成功")
	
	// 查找testuser用户
	var testUser model.User
	result := db.Where("username = ?", "testuser").First(&testUser)
	if result.Error != nil {
		fmt.Printf("❌ 找不到testuser用户: %v\n", result.Error)
		return
	}
	
	fmt.Printf("📋 找到testuser用户: ID=%d, Username=%s, Email=%s, Role=%s\n", 
		testUser.ID, testUser.Username, testUser.Email, testUser.Role)
	
	// 重新设置密码为password123
	fmt.Println("\n🔧 重新设置testuser密码为 'password123'")
	newPassword := "password123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("❌ 密码哈希生成失败: %v\n", err)
		return
	}
	
	testUser.Password = string(hashedPassword)
	result = db.Save(&testUser)
	if result.Error != nil {
		fmt.Printf("❌ 更新密码失败: %v\n", result.Error)
		return
	}
	
	fmt.Println("✅ testuser密码已重新设置为 'password123'")
	
	// 验证新密码
	err = bcrypt.CompareHashAndPassword([]byte(testUser.Password), []byte(newPassword))
	if err == nil {
		fmt.Println("✅ 新密码验证成功！")
	} else {
		fmt.Printf("❌ 新密码验证失败: %v\n", err)
	}
	
	fmt.Println("\n🎉 testuser密码重置完成！现在可以使用 testuser/password123 登录")
}
