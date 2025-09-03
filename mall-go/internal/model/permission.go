package model

// 权限资源和操作定义
// 用于Casbin权限管理系统

// ========== 角色定义 ==========
// 角色常量已在user.go中定义，此处不重复定义

// ========== 资源定义 ==========
const (
	// 用户管理资源
	ResourceUser = "user"

	// 商品管理资源
	ResourceProduct  = "product"
	ResourceCategory = "category"

	// 订单管理资源
	ResourceOrder = "order"

	// 店铺管理资源
	ResourceStore = "store"

	// 系统管理资源
	ResourceSystem = "system"
	ResourceConfig = "config"

	// 文件管理资源
	ResourceFile = "file"

	// 统计报表资源
	ResourceReport = "report"
)

// ========== 操作定义 ==========
const (
	ActionRead   = "read"   // 查看/读取
	ActionWrite  = "write"  // 编辑/修改
	ActionCreate = "create" // 创建/新增
	ActionDelete = "delete" // 删除
	ActionManage = "manage" // 管理（包含所有操作）
)

// ========== 权限组合定义 ==========

// UserPermissions 用户权限定义
var UserPermissions = [][]string{
	// 用户可以查看和编辑自己的信息
	{RoleUser, ResourceUser, ActionRead},
	{RoleUser, ResourceUser, ActionWrite},

	// 用户可以查看商品
	{RoleUser, ResourceProduct, ActionRead},
	{RoleUser, ResourceCategory, ActionRead},

	// 用户可以管理自己的订单
	{RoleUser, ResourceOrder, ActionRead},
	{RoleUser, ResourceOrder, ActionCreate},

	// 用户可以上传文件（头像等）
	{RoleUser, ResourceFile, ActionCreate},
}

// MerchantPermissions 商家权限定义
var MerchantPermissions = [][]string{
	// 继承用户权限，另外添加商家特有权限

	// 商家可以管理商品
	{RoleMerchant, ResourceProduct, ActionRead},
	{RoleMerchant, ResourceProduct, ActionWrite},
	{RoleMerchant, ResourceProduct, ActionCreate},
	{RoleMerchant, ResourceProduct, ActionDelete},

	// 商家可以管理分类
	{RoleMerchant, ResourceCategory, ActionRead},
	{RoleMerchant, ResourceCategory, ActionWrite},
	{RoleMerchant, ResourceCategory, ActionCreate},

	// 商家可以查看和处理订单
	{RoleMerchant, ResourceOrder, ActionRead},
	{RoleMerchant, ResourceOrder, ActionWrite},

	// 商家可以管理店铺
	{RoleMerchant, ResourceStore, ActionRead},
	{RoleMerchant, ResourceStore, ActionWrite},

	// 商家可以上传商品图片
	{RoleMerchant, ResourceFile, ActionCreate},
	{RoleMerchant, ResourceFile, ActionRead},

	// 商家可以查看报表
	{RoleMerchant, ResourceReport, ActionRead},
}

// AdminPermissions 管理员权限定义
var AdminPermissions = [][]string{
	// 管理员拥有所有权限

	// 用户管理
	{RoleAdmin, ResourceUser, ActionManage},

	// 商品管理
	{RoleAdmin, ResourceProduct, ActionManage},
	{RoleAdmin, ResourceCategory, ActionManage},

	// 订单管理
	{RoleAdmin, ResourceOrder, ActionManage},

	// 店铺管理
	{RoleAdmin, ResourceStore, ActionManage},

	// 系统管理
	{RoleAdmin, ResourceSystem, ActionManage},
	{RoleAdmin, ResourceConfig, ActionManage},

	// 文件管理
	{RoleAdmin, ResourceFile, ActionManage},

	// 报表管理
	{RoleAdmin, ResourceReport, ActionManage},
}

// ========== 权限检查辅助函数 ==========

// GetAllPermissions 获取所有权限定义
func GetAllPermissions() [][]string {
	var allPermissions [][]string
	allPermissions = append(allPermissions, UserPermissions...)
	allPermissions = append(allPermissions, MerchantPermissions...)
	allPermissions = append(allPermissions, AdminPermissions...)
	return allPermissions
}

// GetRolePermissions 根据角色获取权限
func GetRolePermissions(role string) [][]string {
	switch role {
	case RoleUser:
		return UserPermissions
	case RoleMerchant:
		return MerchantPermissions
	case RoleAdmin:
		return AdminPermissions
	default:
		return [][]string{}
	}
}

// IsValidRole 检查角色是否有效
func IsValidRole(role string) bool {
	return role == RoleUser || role == RoleMerchant || role == RoleAdmin
}

// IsValidResource 检查资源是否有效
func IsValidResource(resource string) bool {
	validResources := []string{
		ResourceUser, ResourceProduct, ResourceCategory,
		ResourceOrder, ResourceStore, ResourceSystem,
		ResourceConfig, ResourceFile, ResourceReport,
	}

	for _, validResource := range validResources {
		if resource == validResource {
			return true
		}
	}
	return false
}

// IsValidAction 检查操作是否有效
func IsValidAction(action string) bool {
	validActions := []string{
		ActionRead, ActionWrite, ActionCreate,
		ActionDelete, ActionManage,
	}

	for _, validAction := range validActions {
		if action == validAction {
			return true
		}
	}
	return false
}

// ========== 权限描述映射 ==========

// ResourceDescriptions 资源描述映射
var ResourceDescriptions = map[string]string{
	ResourceUser:     "用户管理",
	ResourceProduct:  "商品管理",
	ResourceCategory: "分类管理",
	ResourceOrder:    "订单管理",
	ResourceStore:    "店铺管理",
	ResourceSystem:   "系统管理",
	ResourceConfig:   "配置管理",
	ResourceFile:     "文件管理",
	ResourceReport:   "报表管理",
}

// ActionDescriptions 操作描述映射
var ActionDescriptions = map[string]string{
	ActionRead:   "查看",
	ActionWrite:  "编辑",
	ActionCreate: "创建",
	ActionDelete: "删除",
	ActionManage: "管理",
}

// RoleDescriptions 角色描述映射
var RoleDescriptions = map[string]string{
	RoleUser:     "普通用户",
	RoleMerchant: "商家",
	RoleAdmin:    "管理员",
}

// GetPermissionDescription 获取权限描述
func GetPermissionDescription(role, resource, action string) string {
	roleDesc := RoleDescriptions[role]
	resourceDesc := ResourceDescriptions[resource]
	actionDesc := ActionDescriptions[action]

	if roleDesc == "" || resourceDesc == "" || actionDesc == "" {
		return "未知权限"
	}

	return roleDesc + "可以" + actionDesc + resourceDesc
}
