package main

import (
	"fmt"
	"mall-go/internal/config"
	"mall-go/internal/model"
	"mall-go/pkg/database"
	"mall-go/pkg/logger"
)

func main() {
	fmt.Println("🔧 修复admin用户角色权限")
	
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
	
	// 检查是否需要更新角色
	if adminUser.Role == model.RoleAdmin {
		fmt.Println("✅ admin用户角色已经正确，无需修改")
		return
	}
	
	// 更新admin用户角色
	adminUser.Role = model.RoleAdmin
	result = db.Save(&adminUser)
	if result.Error != nil {
		fmt.Printf("❌ 更新admin用户角色失败: %v\n", result.Error)
		return
	}
	
	fmt.Printf("✅ admin用户角色已更新为: %s\n", adminUser.Role)
	
	// 验证更新结果
	var updatedUser model.User
	db.Where("username = ?", "admin").First(&updatedUser)
	fmt.Printf("🔍 验证结果: ID=%d, Username=%s, Role=%s, Status=%s\n", 
		updatedUser.ID, updatedUser.Username, updatedUser.Role, updatedUser.Status)
	
	// 检查是否还有其他需要设置为管理员的用户
	fmt.Println("\n🔍 检查其他可能的管理员用户...")
	var users []model.User
	db.Where("username LIKE ? OR email LIKE ?", "%admin%", "%admin%").Find(&users)
	
	for _, user := range users {
		if user.Role != model.RoleAdmin && (user.Username == "admin" || user.Email == "admin@mall-go.com") {
			user.Role = model.RoleAdmin
			db.Save(&user)
			fmt.Printf("✅ 已将用户 %s 设置为管理员\n", user.Username)
		}
	}
	
	fmt.Println("\n🎉 admin用户角色修复完成！")
	fmt.Println("现在可以使用admin账户测试商品创建API了")
}
