package main

import (
	"fmt"
	"log"
	"mall-go/internal/config"
	"mall-go/internal/model"
	"mall-go/pkg/database"

	"gorm.io/gorm"
)

func main() {
	fmt.Println("🚀 开始初始化Mall-Go管理员用户...")

	// 初始化配置
	cfg := config.Config{
		Database: config.DatabaseConfig{
			Driver: "sqlite",
			DBName: "mall.db",
		},
	}
	config.GlobalConfig = cfg

	// 初始化数据库
	db := database.Init()
	if db == nil {
		log.Fatal("数据库初始化失败")
	}

	// 自动迁移用户表
	err := db.AutoMigrate(&model.User{})
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 创建管理员用户
	if err := createAdminUser(db); err != nil {
		log.Fatalf("创建管理员用户失败: %v", err)
	}

	fmt.Println("✅ 管理员用户初始化完成！")
}

func createAdminUser(db *gorm.DB) error {
	// 检查是否已存在管理员用户
	var existingAdmin model.User
	if err := db.Where("username = ? OR email = ?", "admin", "admin@mall-go.com").First(&existingAdmin).Error; err == nil {
		fmt.Printf("⚠️ 管理员用户已存在: %s (ID: %d, Role: %s)\n", existingAdmin.Username, existingAdmin.ID, existingAdmin.Role)

		// 如果存在但角色不是admin，更新角色
		if existingAdmin.Role != model.RoleAdmin {
			existingAdmin.Role = model.RoleAdmin
			if err := db.Save(&existingAdmin).Error; err != nil {
				return fmt.Errorf("更新管理员角色失败: %v", err)
			}
			fmt.Printf("✅ 已更新用户 %s 的角色为管理员\n", existingAdmin.Username)
		}
		return nil
	}

	// 创建新的管理员用户
	adminUser := model.User{
		Username: "admin",
		Email:    "admin@mall-go.com",
		Nickname: "系统管理员",
		Role:     model.RoleAdmin,
		Status:   model.StatusActive,
	}

	// 设置密码 (password123)
	if err := adminUser.SetPassword("password123"); err != nil {
		return fmt.Errorf("设置管理员密码失败: %v", err)
	}

	// 创建用户
	if err := db.Create(&adminUser).Error; err != nil {
		return fmt.Errorf("创建管理员用户失败: %v", err)
	}

	fmt.Printf("✅ 创建管理员用户成功: %s (ID: %d)\n", adminUser.Username, adminUser.ID)
	fmt.Println("📝 管理员登录信息:")
	fmt.Println("   用户名: admin")
	fmt.Println("   邮箱: admin@mall-go.com")
	fmt.Println("   密码: password123")
	fmt.Println("   角色: admin")

	return nil
}
