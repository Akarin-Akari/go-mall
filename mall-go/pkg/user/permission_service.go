package user

import (
	"fmt"
	"time"

	"mall-go/internal/model"

	"gorm.io/gorm"
)

// PermissionService 用户权限服务
type PermissionService struct {
	db *gorm.DB
}

// NewPermissionService 创建用户权限服务
func NewPermissionService(db *gorm.DB) *PermissionService {
	return &PermissionService{
		db: db,
	}
}

// CasbinPolicy 权限策略结构
type CasbinPolicy struct {
	Subject string `json:"subject"`
	Object  string `json:"object"`
	Action  string `json:"action"`
}

// AssignRoleRequest 分配角色请求
type AssignRoleRequest struct {
	UserID uint   `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required"`
	Reason string `json:"reason"` // 分配原因
}

// RevokeRoleRequest 撤销角色请求
type RevokeRoleRequest struct {
	UserID uint   `json:"user_id" binding:"required"`
	Reason string `json:"reason"` // 撤销原因
}

// CheckPermissionRequest 检查权限请求
type CheckPermissionRequest struct {
	UserID   uint   `json:"user_id" binding:"required"`
	Resource string `json:"resource" binding:"required"`
	Action   string `json:"action" binding:"required"`
}

// UserPermissionInfo 用户权限信息
type UserPermissionInfo struct {
	UserID      uint                `json:"user_id"`
	Username    string              `json:"username"`
	Role        string              `json:"role"`
	Permissions []string            `json:"permissions"`
	Resources   map[string][]string `json:"resources"` // 资源和对应的操作
	Policies    []CasbinPolicy      `json:"policies"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

// RoleChangeLog 角色变更日志
type RoleChangeLog struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	AdminID   uint      `gorm:"not null" json:"admin_id"` // 操作管理员ID
	OldRole   string    `gorm:"size:50" json:"old_role"`  // 原角色
	NewRole   string    `gorm:"size:50" json:"new_role"`  // 新角色
	Reason    string    `gorm:"size:500" json:"reason"`   // 变更原因
	CreatedAt time.Time `json:"created_at"`

	// 关联关系
	User  *model.User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Admin *model.User `gorm:"foreignKey:AdminID" json:"admin,omitempty"`
}

// TableName 指定表名
func (RoleChangeLog) TableName() string {
	return "role_change_logs"
}

// AssignRole 分配角色
func (ps *PermissionService) AssignRole(adminID uint, req *AssignRoleRequest) error {
	// 验证角色是否有效
	if !ps.isValidRole(req.Role) {
		return fmt.Errorf("无效的角色: %s", req.Role)
	}

	// 查询用户
	var user model.User
	if err := ps.db.First(&user, req.UserID).Error; err != nil {
		return fmt.Errorf("用户不存在")
	}

	// 记录原角色
	oldRole := user.Role

	// 检查是否需要更新
	if oldRole == req.Role {
		return fmt.Errorf("用户已经是 %s 角色", req.Role)
	}

	// 开始事务
	tx := ps.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新用户角色
	user.Role = req.Role
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新用户角色失败: %v", err)
	}

	// 权限更新已通过数据库角色字段完成

	// 记录角色变更日志
	changeLog := &RoleChangeLog{
		UserID:  req.UserID,
		AdminID: adminID,
		OldRole: oldRole,
		NewRole: req.Role,
		Reason:  req.Reason,
	}
	if err := tx.Create(changeLog).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("记录角色变更日志失败: %v", err)
	}

	tx.Commit()
	return nil
}

// RevokeRole 撤销角色（设为普通用户）
func (ps *PermissionService) RevokeRole(adminID uint, req *RevokeRoleRequest) error {
	// 查询用户
	var user model.User
	if err := ps.db.First(&user, req.UserID).Error; err != nil {
		return fmt.Errorf("用户不存在")
	}

	// 检查是否已经是普通用户
	if user.Role == model.RoleUser {
		return fmt.Errorf("用户已经是普通用户")
	}

	// 记录原角色
	oldRole := user.Role

	// 开始事务
	tx := ps.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新用户角色为普通用户
	user.Role = model.RoleUser
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("撤销用户角色失败: %v", err)
	}

	// 权限更新已通过数据库角色字段完成

	// 记录角色变更日志
	changeLog := &RoleChangeLog{
		UserID:  req.UserID,
		AdminID: adminID,
		OldRole: oldRole,
		NewRole: model.RoleUser,
		Reason:  req.Reason,
	}
	if err := tx.Create(changeLog).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("记录角色变更日志失败: %v", err)
	}

	tx.Commit()
	return nil
}

// CheckPermission 检查用户权限
func (ps *PermissionService) CheckPermission(req *CheckPermissionRequest) (bool, error) {
	// 使用简单的角色检查
	return ps.checkPermissionByRole(req.UserID, req.Resource, req.Action)
}

// GetUserPermissions 获取用户权限信息
func (ps *PermissionService) GetUserPermissions(userID uint) (*UserPermissionInfo, error) {
	// 查询用户
	var user model.User
	if err := ps.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	permInfo := &UserPermissionInfo{
		UserID:    user.ID,
		Username:  user.Username,
		Role:      user.Role,
		UpdatedAt: user.UpdatedAt,
	}

	// 简单的权限获取
	permInfo.Permissions = ps.getPermissionsByRole(user.Role)
	permInfo.Resources = ps.getResourcesByRole(user.Role)
	permInfo.Policies = []CasbinPolicy{} // 空的策略列表

	return permInfo, nil
}

// GetRoleChangeLogs 获取角色变更日志
func (ps *PermissionService) GetRoleChangeLogs(userID uint, page, pageSize int) ([]RoleChangeLog, int64, error) {
	var logs []RoleChangeLog
	var total int64

	query := ps.db.Model(&RoleChangeLog{})
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询角色变更日志总数失败: %v", err)
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	if err := query.Preload("User").Preload("Admin").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("查询角色变更日志失败: %v", err)
	}

	return logs, total, nil
}

// GetUsersByRole 根据角色获取用户列表
func (ps *PermissionService) GetUsersByRole(role string, page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	// 验证角色
	if !ps.isValidRole(role) {
		return nil, 0, fmt.Errorf("无效的角色: %s", role)
	}

	// 获取总数
	if err := ps.db.Model(&model.User{}).Where("role = ?", role).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询用户总数失败: %v", err)
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	if err := ps.db.Where("role = ?", role).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("查询用户列表失败: %v", err)
	}

	return users, total, nil
}

// checkPermissionByRole 通过角色检查权限（简单实现）
func (ps *PermissionService) checkPermissionByRole(userID uint, resource, action string) (bool, error) {
	var user model.User
	if err := ps.db.First(&user, userID).Error; err != nil {
		return false, fmt.Errorf("用户不存在")
	}

	// 管理员拥有所有权限
	if user.Role == model.RoleAdmin {
		return true, nil
	}

	// 商家权限检查
	if user.Role == model.RoleMerchant {
		merchantResources := []string{"product", "order", "store"}
		for _, res := range merchantResources {
			if resource == res {
				return true, nil
			}
		}
	}

	// 普通用户权限检查
	if user.Role == model.RoleUser {
		userResources := []string{"profile", "order"}
		userActions := []string{"read", "update"}

		for _, res := range userResources {
			if resource == res {
				for _, act := range userActions {
					if action == act {
						return true, nil
					}
				}
			}
		}
	}

	return false, nil
}

// isValidRole 验证角色是否有效
func (ps *PermissionService) isValidRole(role string) bool {
	validRoles := []string{model.RoleUser, model.RoleMerchant, model.RoleAdmin}
	for _, validRole := range validRoles {
		if role == validRole {
			return true
		}
	}
	return false
}

// getPermissionsByRole 根据角色获取权限列表
func (ps *PermissionService) getPermissionsByRole(role string) []string {
	switch role {
	case model.RoleAdmin:
		return []string{
			"user:read", "user:write", "user:delete",
			"product:read", "product:write", "product:delete",
			"order:read", "order:write", "order:delete",
			"system:read", "system:write",
		}
	case model.RoleMerchant:
		return []string{
			"product:read", "product:write",
			"order:read", "order:write",
			"store:read", "store:write",
		}
	case model.RoleUser:
		return []string{
			"profile:read", "profile:write",
			"order:read",
		}
	default:
		return []string{}
	}
}

// getResourcesByRole 根据角色获取资源权限映射
func (ps *PermissionService) getResourcesByRole(role string) map[string][]string {
	switch role {
	case model.RoleAdmin:
		return map[string][]string{
			"user":    {"read", "write", "delete"},
			"product": {"read", "write", "delete"},
			"order":   {"read", "write", "delete"},
			"system":  {"read", "write"},
		}
	case model.RoleMerchant:
		return map[string][]string{
			"product": {"read", "write"},
			"order":   {"read", "write"},
			"store":   {"read", "write"},
		}
	case model.RoleUser:
		return map[string][]string{
			"profile": {"read", "write"},
			"order":   {"read"},
		}
	default:
		return map[string][]string{}
	}
}

// 全局权限服务实例
var globalPermissionService *PermissionService

// InitGlobalPermissionService 初始化全局权限服务
func InitGlobalPermissionService(db *gorm.DB) {
	globalPermissionService = NewPermissionService(db)
}

// GetGlobalPermissionService 获取全局权限服务
func GetGlobalPermissionService() *PermissionService {
	return globalPermissionService
}
