package main

import (
	"fmt"
	"log"
	"mall-go/internal/config"
	"mall-go/internal/model"
	"mall-go/pkg/database"
)

func main() {
	fmt.Println("🔍 检查数据库中的用户数据")
	
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

	// 查询所有用户
	var users []model.User
	if err := db.Find(&users).Error; err != nil {
		log.Fatalf("查询用户失败: %v", err)
	}

	fmt.Printf("数据库中共有 %d 个用户:\n", len(users))
	fmt.Println("================================================================================")
	
	for _, user := range users {
		fmt.Printf("ID: %d | Username: %s | Email: %s | Role: %s | Status: %s | Created: %s\n",
			user.ID, user.Username, user.Email, user.Role, user.Status, user.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	
	fmt.Println("================================================================================")
	
	// 特别检查admin用户
	var adminUsers []model.User
	if err := db.Where("username = ?", "admin").Find(&adminUsers).Error; err != nil {
		log.Printf("查询admin用户失败: %v", err)
	} else {
		fmt.Printf("\n找到 %d 个admin用户:\n", len(adminUsers))
		for _, admin := range adminUsers {
			fmt.Printf("  ID: %d | Email: %s | Role: %s | Status: %s\n",
				admin.ID, admin.Email, admin.Role, admin.Status)
		}
	}
}
