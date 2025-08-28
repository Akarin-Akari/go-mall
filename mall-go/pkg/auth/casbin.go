package auth

import (
	"fmt"
	"strconv"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

// CasbinManager Casbin权限管理器
type CasbinManager struct {
	enforcer *casbin.Enforcer
	adapter  *gormadapter.Adapter
}

var globalCasbinManager *CasbinManager

// InitCasbin 初始化Casbin
func InitCasbin(db *gorm.DB) (*CasbinManager, error) {
	// 创建GORM适配器
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, fmt.Errorf("创建Casbin适配器失败: %v", err)
	}

	// 创建执行器
	enforcer, err := casbin.NewEnforcer("configs/rbac_model.conf", adapter)
	if err != nil {
		return nil, fmt.Errorf("创建Casbin执行器失败: %v", err)
	}

	// 加载策略
	err = enforcer.LoadPolicy()
	if err != nil {
		return nil, fmt.Errorf("加载Casbin策略失败: %v", err)
	}

	manager := &CasbinManager{
		enforcer: enforcer,
		adapter:  adapter,
	}

	globalCasbinManager = manager
	return manager, nil
}

// GetEnforcer 获取Casbin执行器
func (cm *CasbinManager) GetEnforcer() *casbin.Enforcer {
	return cm.enforcer
}

// CheckPermission 检查权限
func (cm *CasbinManager) CheckPermission(subject, object, action string) (bool, error) {
	return cm.enforcer.Enforce(subject, object, action)
}

// AddPolicy 添加权限策略
func (cm *CasbinManager) AddPolicy(subject, object, action string) (bool, error) {
	return cm.enforcer.AddPolicy(subject, object, action)
}

// RemovePolicy 移除权限策略
func (cm *CasbinManager) RemovePolicy(subject, object, action string) (bool, error) {
	return cm.enforcer.RemovePolicy(subject, object, action)
}

// AddRoleForUser 为用户添加角色
func (cm *CasbinManager) AddRoleForUser(user, role string) (bool, error) {
	return cm.enforcer.AddRoleForUser(user, role)
}

// DeleteRoleForUser 删除用户角色
func (cm *CasbinManager) DeleteRoleForUser(user, role string) (bool, error) {
	return cm.enforcer.DeleteRoleForUser(user, role)
}

// GetRolesForUser 获取用户的所有角色
func (cm *CasbinManager) GetRolesForUser(user string) ([]string, error) {
	roles, err := cm.enforcer.GetRolesForUser(user)
	return roles, err
}

// GetUsersForRole 获取角色的所有用户
func (cm *CasbinManager) GetUsersForRole(role string) ([]string, error) {
	users, err := cm.enforcer.GetUsersForRole(role)
	return users, err
}

// HasRoleForUser 检查用户是否有指定角色
func (cm *CasbinManager) HasRoleForUser(user, role string) (bool, error) {
	hasRole, err := cm.enforcer.HasRoleForUser(user, role)
	return hasRole, err
}

// SavePolicy 保存策略到数据库
func (cm *CasbinManager) SavePolicy() error {
	return cm.enforcer.SavePolicy()
}

// LoadPolicy 从数据库加载策略
func (cm *CasbinManager) LoadPolicy() error {
	return cm.enforcer.LoadPolicy()
}

// 全局函数，用于方便调用

// CheckUserPermission 检查用户权限
func CheckUserPermission(userID uint, resource, action string) (bool, error) {
	if globalCasbinManager == nil {
		return false, fmt.Errorf("Casbin未初始化")
	}

	userSubject := fmt.Sprintf("user:%d", userID)
	return globalCasbinManager.CheckPermission(userSubject, resource, action)
}

// AddUserRole 为用户添加角色
func AddUserRole(userID uint, role string) (bool, error) {
	if globalCasbinManager == nil {
		return false, fmt.Errorf("Casbin未初始化")
	}

	userSubject := fmt.Sprintf("user:%d", userID)
	return globalCasbinManager.AddRoleForUser(userSubject, role)
}

// RemoveUserRole 移除用户角色
func RemoveUserRole(userID uint, role string) (bool, error) {
	if globalCasbinManager == nil {
		return false, fmt.Errorf("Casbin未初始化")
	}

	userSubject := fmt.Sprintf("user:%d", userID)
	return globalCasbinManager.DeleteRoleForUser(userSubject, role)
}

// GetUserRoles 获取用户角色
func GetUserRoles(userID uint) ([]string, error) {
	if globalCasbinManager == nil {
		return nil, fmt.Errorf("Casbin未初始化")
	}

	userSubject := fmt.Sprintf("user:%d", userID)
	return globalCasbinManager.GetRolesForUser(userSubject)
}

// HasUserRole 检查用户是否有指定角色
func HasUserRole(userID uint, role string) (bool, error) {
	if globalCasbinManager == nil {
		return false, fmt.Errorf("Casbin未初始化")
	}

	userSubject := fmt.Sprintf("user:%d", userID)
	return globalCasbinManager.HasRoleForUser(userSubject, role)
}

// AddRolePermission 为角色添加权限
func AddRolePermission(role, resource, action string) (bool, error) {
	if globalCasbinManager == nil {
		return false, fmt.Errorf("Casbin未初始化")
	}

	return globalCasbinManager.AddPolicy(role, resource, action)
}

// RemoveRolePermission 移除角色权限
func RemoveRolePermission(role, resource, action string) (bool, error) {
	if globalCasbinManager == nil {
		return false, fmt.Errorf("Casbin未初始化")
	}

	return globalCasbinManager.RemovePolicy(role, resource, action)
}

// InitDefaultPolicies 初始化默认权限策略
func InitDefaultPolicies() error {
	if globalCasbinManager == nil {
		return fmt.Errorf("Casbin未初始化")
	}

	// 管理员权限
	adminPolicies := [][]string{
		{"admin", "user", "create"},
		{"admin", "user", "read"},
		{"admin", "user", "update"},
		{"admin", "user", "delete"},
		{"admin", "product", "create"},
		{"admin", "product", "read"},
		{"admin", "product", "update"},
		{"admin", "product", "delete"},
		{"admin", "order", "create"},
		{"admin", "order", "read"},
		{"admin", "order", "update"},
		{"admin", "order", "delete"},
		{"admin", "file", "create"},
		{"admin", "file", "read"},
		{"admin", "file", "update"},
		{"admin", "file", "delete"},
	}

	// 普通用户权限
	userPolicies := [][]string{
		{"user", "product", "read"},
		{"user", "order", "create"},
		{"user", "order", "read"},
		{"user", "file", "create"},
		{"user", "file", "read"},
	}

	// 添加管理员权限
	for _, policy := range adminPolicies {
		globalCasbinManager.AddPolicy(policy[0], policy[1], policy[2])
	}

	// 添加用户权限
	for _, policy := range userPolicies {
		globalCasbinManager.AddPolicy(policy[0], policy[1], policy[2])
	}

	// 保存策略
	return globalCasbinManager.SavePolicy()
}

// GetCasbinManager 获取全局Casbin管理器
func GetCasbinManager() *CasbinManager {
	return globalCasbinManager
}

// ParseUserID 从用户主体中解析用户ID
func ParseUserID(subject string) (uint, error) {
	if len(subject) < 5 || subject[:5] != "user:" {
		return 0, fmt.Errorf("无效的用户主体格式: %s", subject)
	}

	userIDStr := subject[5:]
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("解析用户ID失败: %v", err)
	}

	return uint(userID), nil
}

// FormatUserSubject 格式化用户主体
func FormatUserSubject(userID uint) string {
	return fmt.Sprintf("user:%d", userID)
}
