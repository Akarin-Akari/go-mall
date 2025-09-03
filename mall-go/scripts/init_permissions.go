package main

import (
	"fmt"
	"log"

	"mall-go/internal/config"
	"mall-go/internal/model"
	"mall-go/pkg/auth"
	"mall-go/pkg/database"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("Mall-Go权限系统初始化程序")
	fmt.Println("========================================")

	// 加载配置
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 连接数据库
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// 初始化Casbin
	casbinManager, err := auth.InitCasbin(db)
	if err != nil {
		log.Fatalf("Casbin初始化失败: %v", err)
	}

	fmt.Println("✅ 数据库和Casbin初始化成功")

	// 清空现有权限数据（可选）
	fmt.Println("\n[1] 清空现有权限数据...")
	if err := clearExistingPolicies(casbinManager); err != nil {
		log.Printf("⚠️ 清空权限数据失败: %v", err)
	} else {
		fmt.Println("✅ 现有权限数据已清空")
	}

	// 初始化角色权限
	fmt.Println("\n[2] 初始化角色权限...")
	if err := initRolePermissions(casbinManager); err != nil {
		log.Fatalf("❌ 初始化角色权限失败: %v", err)
	}
	fmt.Println("✅ 角色权限初始化完成")

	// 初始化示例用户角色
	fmt.Println("\n[3] 初始化示例用户角色...")
	if err := initUserRoles(casbinManager); err != nil {
		log.Printf("⚠️ 初始化用户角色失败: %v", err)
	} else {
		fmt.Println("✅ 示例用户角色初始化完成")
	}

	// 保存策略到数据库
	fmt.Println("\n[4] 保存策略到数据库...")
	if err := casbinManager.SavePolicy(); err != nil {
		log.Fatalf("❌ 保存策略失败: %v", err)
	}
	fmt.Println("✅ 策略保存成功")

	// 验证权限设置
	fmt.Println("\n[5] 验证权限设置...")
	if err := verifyPermissions(casbinManager); err != nil {
		log.Printf("⚠️ 权限验证失败: %v", err)
	} else {
		fmt.Println("✅ 权限验证通过")
	}

	fmt.Println("\n========================================")
	fmt.Println("🎉 权限系统初始化完成！")
	fmt.Println("========================================")
}

// clearExistingPolicies 清空现有权限策略
func clearExistingPolicies(cm *auth.CasbinManager) error {
	enforcer := cm.GetEnforcer()
	
	// 清空所有策略
	_, err := enforcer.RemoveFilteredPolicy(0)
	if err != nil {
		return fmt.Errorf("清空策略失败: %v", err)
	}

	// 清空所有角色分配
	_, err = enforcer.RemoveFilteredGroupingPolicy(0)
	if err != nil {
		return fmt.Errorf("清空角色分配失败: %v", err)
	}

	return nil
}

// initRolePermissions 初始化角色权限
func initRolePermissions(cm *auth.CasbinManager) error {
	// 获取所有权限定义
	allPermissions := model.GetAllPermissions()

	// 添加权限策略
	for _, perm := range allPermissions {
		if len(perm) >= 3 {
			role, resource, action := perm[0], perm[1], perm[2]
			_, err := cm.AddPolicy(role, resource, action)
			if err != nil {
				return fmt.Errorf("添加权限策略失败 [%s, %s, %s]: %v", role, resource, action, err)
			}
			fmt.Printf("  ✓ 添加权限: %s 可以 %s %s\n", role, action, resource)
		}
	}

	return nil
}

// initUserRoles 初始化示例用户角色
func initUserRoles(cm *auth.CasbinManager) error {
	// 示例用户角色分配
	userRoles := []struct {
		UserID uint
		Role   string
		Desc   string
	}{
		{1, model.RoleAdmin, "管理员用户"},
		{2, model.RoleMerchant, "商家用户"},
		{3, model.RoleUser, "普通用户"},
	}

	for _, ur := range userRoles {
		_, err := cm.AddRoleForUser(ur.UserID, ur.Role)
		if err != nil {
			return fmt.Errorf("为用户%d分配角色%s失败: %v", ur.UserID, ur.Role, err)
		}
		fmt.Printf("  ✓ 用户%d -> %s (%s)\n", ur.UserID, ur.Role, ur.Desc)
	}

	return nil
}

// verifyPermissions 验证权限设置
func verifyPermissions(cm *auth.CasbinManager) error {
	// 测试用例
	testCases := []struct {
		UserID   uint
		Resource string
		Action   string
		Expected bool
		Desc     string
	}{
		{1, model.ResourceUser, model.ActionManage, true, "管理员管理用户"},
		{1, model.ResourceProduct, model.ActionManage, true, "管理员管理商品"},
		{2, model.ResourceProduct, model.ActionCreate, true, "商家创建商品"},
		{2, model.ResourceUser, model.ActionManage, false, "商家不能管理用户"},
		{3, model.ResourceProduct, model.ActionRead, true, "用户查看商品"},
		{3, model.ResourceProduct, model.ActionCreate, false, "用户不能创建商品"},
	}

	for _, tc := range testCases {
		hasPermission, err := cm.CheckPermission(tc.UserID, tc.Resource, tc.Action)
		if err != nil {
			return fmt.Errorf("权限检查失败: %v", err)
		}

		if hasPermission != tc.Expected {
			return fmt.Errorf("权限验证失败: %s - 期望%v，实际%v", tc.Desc, tc.Expected, hasPermission)
		}

		status := "✓"
		if !tc.Expected {
			status = "✗"
		}
		fmt.Printf("  %s %s\n", status, tc.Desc)
	}

	return nil
}

// 辅助函数：显示所有权限
func showAllPermissions(cm *auth.CasbinManager) {
	fmt.Println("\n========== 所有权限策略 ==========")
	
	enforcer := cm.GetEnforcer()
	policies := enforcer.GetPolicy()
	
	for _, policy := range policies {
		if len(policy) >= 3 {
			fmt.Printf("权限: %s 可以 %s %s\n", policy[0], policy[2], policy[1])
		}
	}

	fmt.Println("\n========== 所有用户角色 ==========")
	
	groupingPolicies := enforcer.GetGroupingPolicy()
	for _, gp := range groupingPolicies {
		if len(gp) >= 2 {
			fmt.Printf("用户角色: %s -> %s\n", gp[0], gp[1])
		}
	}
}
