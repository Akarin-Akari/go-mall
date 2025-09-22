package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"mall-go/internal/config"
	"mall-go/internal/model"
	"mall-go/pkg/database"
	"mall-go/pkg/logger"
)

func main() {
	fmt.Println("🔐 验证admin用户密码")

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

	// 查找admin用户
	var adminUser model.User
	result := db.Where("username = ?", "admin").First(&adminUser)
	if result.Error != nil {
		fmt.Printf("❌ 找不到admin用户: %v\n", result.Error)
		return
	}

	fmt.Printf("📋 找到admin用户: ID=%d, Username=%s, Email=%s, Role=%s\n",
		adminUser.ID, adminUser.Username, adminUser.Email, adminUser.Role)

	// 测试密码
	testPasswords := []string{"admin123", "admin", "123456", "password"}

	fmt.Println("\n🧪 测试常见密码:")
	for _, password := range testPasswords {
		err := bcrypt.CompareHashAndPassword([]byte(adminUser.Password), []byte(password))
		if err == nil {
			fmt.Printf("✅ 密码 '%s' 匹配成功！\n", password)
			return
		} else {
			fmt.Printf("❌ 密码 '%s' 不匹配\n", password)
		}
	}

	fmt.Println("\n⚠️  所有测试密码都不匹配")
	fmt.Printf("存储的密码哈希: %s\n", adminUser.Password)

	// 重新设置密码为admin123
	fmt.Println("\n🔧 重新设置admin密码为 'admin123'")
	newPassword := "admin123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("❌ 密码哈希生成失败: %v\n", err)
		return
	}

	adminUser.Password = string(hashedPassword)
	result = db.Save(&adminUser)
	if result.Error != nil {
		fmt.Printf("❌ 更新密码失败: %v\n", result.Error)
		return
	}

	fmt.Println("✅ admin密码已重新设置为 'admin123'")

	// 验证新密码
	err = bcrypt.CompareHashAndPassword([]byte(adminUser.Password), []byte(newPassword))
	if err == nil {
		fmt.Println("✅ 新密码验证成功！")
	} else {
		fmt.Printf("❌ 新密码验证失败: %v\n", err)
	}

	fmt.Println("\n🎉 密码重置完成！现在可以使用 admin/admin123 登录")
}
