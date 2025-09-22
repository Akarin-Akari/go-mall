package auth

import (
	"testing"

	"mall-go/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// CasbinTestSuite Casbin权限系统测试套件
type CasbinTestSuite struct {
	suite.Suite
	db            *gorm.DB
	casbinManager *CasbinManager
}

// SetupSuite 设置测试套件
func (suite *CasbinTestSuite) SetupSuite() {
	// 设置测试数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)
	suite.db = db

	// 初始化Casbin
	casbinManager, err := InitCasbin(db)
	suite.Require().NoError(err)
	suite.casbinManager = casbinManager

	// 初始化测试权限数据
	suite.initTestPermissions()
}

// TearDownSuite 清理测试套件
func (suite *CasbinTestSuite) TearDownSuite() {
	// 清理资源
}

// initTestPermissions 初始化测试权限数据
func (suite *CasbinTestSuite) initTestPermissions() {
	// 添加角色权限策略
	policies := [][]string{
		// 管理员权限
		{model.RoleAdmin, model.ResourceUser, model.ActionManage},
		{model.RoleAdmin, model.ResourceProduct, model.ActionManage},
		{model.RoleAdmin, model.ResourceOrder, model.ActionManage},

		// 商家权限
		{model.RoleMerchant, model.ResourceProduct, model.ActionCreate},
		{model.RoleMerchant, model.ResourceProduct, model.ActionRead},
		{model.RoleMerchant, model.ResourceProduct, model.ActionWrite},
		{model.RoleMerchant, model.ResourceOrder, model.ActionRead},

		// 用户权限
		{model.RoleUser, model.ResourceProduct, model.ActionRead},
		{model.RoleUser, model.ResourceOrder, model.ActionCreate},
		{model.RoleUser, model.ResourceOrder, model.ActionRead},
	}

	for _, policy := range policies {
		suite.casbinManager.AddPolicy(policy[0], policy[1], policy[2])
	}

	// 添加用户角色关系
	userRoles := [][]string{
		{"user:1", model.RoleAdmin},
		{"user:2", model.RoleMerchant},
		{"user:3", model.RoleUser},
	}

	for _, userRole := range userRoles {
		suite.casbinManager.AddRoleForUser(userRole[0], userRole[1])
	}

	// 保存策略
	suite.casbinManager.SavePolicy()
}

// TestCasbinManager_CheckPermission 测试权限检查
func (suite *CasbinTestSuite) TestCasbinManager_CheckPermission() {
	tests := []struct {
		name     string
		subject  string
		object   string
		action   string
		expected bool
	}{
		{
			name:     "管理员管理用户权限",
			subject:  "user:1",
			object:   model.ResourceUser,
			action:   model.ActionManage,
			expected: true,
		},
		{
			name:     "商家创建商品权限",
			subject:  "user:2",
			object:   model.ResourceProduct,
			action:   model.ActionCreate,
			expected: true,
		},
		{
			name:     "用户查看商品权限",
			subject:  "user:3",
			object:   model.ResourceProduct,
			action:   model.ActionRead,
			expected: true,
		},
		{
			name:     "用户无管理权限",
			subject:  "user:3",
			object:   model.ResourceUser,
			action:   model.ActionManage,
			expected: false,
		},
		{
			name:     "商家无用户管理权限",
			subject:  "user:2",
			object:   model.ResourceUser,
			action:   model.ActionManage,
			expected: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			hasPermission, err := suite.casbinManager.CheckPermission(tt.subject, tt.object, tt.action)
			suite.NoError(err)
			suite.Equal(tt.expected, hasPermission)
		})
	}
}

// TestCasbinManager_RoleManagement 测试角色管理
func (suite *CasbinTestSuite) TestCasbinManager_RoleManagement() {
	// 测试添加用户角色
	added, err := suite.casbinManager.AddRoleForUser("user:4", model.RoleUser)
	suite.NoError(err)
	suite.True(added)

	// 测试检查用户角色
	hasRole, err := suite.casbinManager.HasRoleForUser("user:4", model.RoleUser)
	suite.NoError(err)
	suite.True(hasRole)

	// 测试获取用户角色
	roles, err := suite.casbinManager.GetRolesForUser("user:4")
	suite.NoError(err)
	suite.Contains(roles, model.RoleUser)

	// 测试删除用户角色
	deleted, err := suite.casbinManager.DeleteRoleForUser("user:4", model.RoleUser)
	suite.NoError(err)
	suite.True(deleted)

	// 验证角色已删除
	hasRole, err = suite.casbinManager.HasRoleForUser("user:4", model.RoleUser)
	suite.NoError(err)
	suite.False(hasRole)
}

// TestCasbinManager_PolicyManagement 测试策略管理
func (suite *CasbinTestSuite) TestCasbinManager_PolicyManagement() {
	// 测试添加策略
	added, err := suite.casbinManager.AddPolicy("test_role", "test_resource", "test_action")
	suite.NoError(err)
	suite.True(added)

	// 测试检查策略
	hasPermission, err := suite.casbinManager.CheckPermission("test_role", "test_resource", "test_action")
	suite.NoError(err)
	suite.True(hasPermission)

	// 测试删除策略
	removed, err := suite.casbinManager.RemovePolicy("test_role", "test_resource", "test_action")
	suite.NoError(err)
	suite.True(removed)

	// 验证策略已删除
	hasPermission, err = suite.casbinManager.CheckPermission("test_role", "test_resource", "test_action")
	suite.NoError(err)
	suite.False(hasPermission)
}

// TestGlobalFunctions 测试全局函数
func (suite *CasbinTestSuite) TestGlobalFunctions() {
	// 测试检查用户权限
	hasPermission, err := CheckUserPermission(1, model.ResourceUser, model.ActionManage)
	suite.NoError(err)
	suite.True(hasPermission)

	// 测试添加用户角色
	added, err := AddUserRole(5, model.RoleUser)
	suite.NoError(err)
	suite.True(added)

	// 测试获取用户角色
	roles, err := GetUserRoles(5)
	suite.NoError(err)
	suite.Contains(roles, model.RoleUser)

	// 测试检查用户角色
	hasRole, err := HasUserRole(5, model.RoleUser)
	suite.NoError(err)
	suite.True(hasRole)

	// 测试删除用户角色
	removed, err := RemoveUserRole(5, model.RoleUser)
	suite.NoError(err)
	suite.True(removed)
}

// TestPermissionInheritance 测试权限继承
func (suite *CasbinTestSuite) TestPermissionInheritance() {
	// 创建角色继承关系：super_admin 继承 admin
	suite.casbinManager.AddRoleForUser("super_admin", model.RoleAdmin)

	// 为用户分配 super_admin 角色
	suite.casbinManager.AddRoleForUser("user:6", "super_admin")

	// 验证用户通过角色继承获得权限
	hasPermission, err := suite.casbinManager.CheckPermission("user:6", model.ResourceUser, model.ActionManage)
	suite.NoError(err)
	suite.True(hasPermission)
}

// TestComplexPermissionScenarios 测试复杂权限场景
func (suite *CasbinTestSuite) TestComplexPermissionScenarios() {
	// 场景1：用户拥有多个角色
	suite.casbinManager.AddRoleForUser("user:7", model.RoleUser)
	suite.casbinManager.AddRoleForUser("user:7", model.RoleMerchant)

	// 验证用户拥有两个角色的权限
	// 用户权限：查看商品
	hasUserPermission, err := suite.casbinManager.CheckPermission("user:7", model.ResourceProduct, model.ActionRead)
	suite.NoError(err)
	suite.True(hasUserPermission)

	// 商家权限：创建商品
	hasMerchantPermission, err := suite.casbinManager.CheckPermission("user:7", model.ResourceProduct, model.ActionCreate)
	suite.NoError(err)
	suite.True(hasMerchantPermission)

	// 场景2：动态权限变更
	// 临时给用户添加管理员权限
	suite.casbinManager.AddRoleForUser("user:7", model.RoleAdmin)

	// 验证用户现在拥有管理员权限
	hasAdminPermission, err := suite.casbinManager.CheckPermission("user:7", model.ResourceUser, model.ActionManage)
	suite.NoError(err)
	suite.True(hasAdminPermission)

	// 撤销管理员权限
	suite.casbinManager.DeleteRoleForUser("user:7", model.RoleAdmin)

	// 验证管理员权限已撤销
	hasAdminPermission, err = suite.casbinManager.CheckPermission("user:7", model.ResourceUser, model.ActionManage)
	suite.NoError(err)
	suite.False(hasAdminPermission)
}

// TestPolicyPersistence 测试策略持久化
func (suite *CasbinTestSuite) TestPolicyPersistence() {
	// 添加新策略
	suite.casbinManager.AddPolicy("test_persistence", "test_resource", "test_action")

	// 保存到数据库
	err := suite.casbinManager.SavePolicy()
	suite.NoError(err)

	// 重新加载策略
	err = suite.casbinManager.LoadPolicy()
	suite.NoError(err)

	// 验证策略仍然存在
	hasPermission, err := suite.casbinManager.CheckPermission("test_persistence", "test_resource", "test_action")
	suite.NoError(err)
	suite.True(hasPermission)
}

// TestErrorHandling 测试错误处理
func (suite *CasbinTestSuite) TestErrorHandling() {
	// 测试未初始化的全局管理器
	originalManager := globalCasbinManager
	globalCasbinManager = nil

	// 应该返回错误
	_, err := CheckUserPermission(1, model.ResourceUser, model.ActionRead)
	suite.Error(err)
	suite.Contains(err.Error(), "Casbin未初始化")

	// 恢复全局管理器
	globalCasbinManager = originalManager
}

// TestPermissionValidation 测试权限验证
func (suite *CasbinTestSuite) TestPermissionValidation() {
	// 测试有效资源
	suite.True(model.IsValidResource(model.ResourceUser))
	suite.True(model.IsValidResource(model.ResourceProduct))
	suite.False(model.IsValidResource("invalid_resource"))

	// 测试有效操作
	suite.True(model.IsValidAction(model.ActionRead))
	suite.True(model.IsValidAction(model.ActionManage))
	suite.False(model.IsValidAction("invalid_action"))
}

// 运行Casbin测试套件
func TestCasbinSuite(t *testing.T) {
	suite.Run(t, new(CasbinTestSuite))
}

// TestInitCasbin 测试Casbin初始化
func TestInitCasbin(t *testing.T) {
	// 设置测试数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// 测试初始化
	manager, err := InitCasbin(db)
	assert.NoError(t, err)
	assert.NotNil(t, manager)
	assert.NotNil(t, manager.enforcer)
	assert.NotNil(t, manager.adapter)

	// 测试全局管理器设置
	assert.Equal(t, manager, GetCasbinManager())
}

// TestInitDefaultPolicies 测试默认策略初始化
func TestInitDefaultPolicies(t *testing.T) {
	// 设置测试数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// 初始化Casbin
	_, err = InitCasbin(db)
	assert.NoError(t, err)

	// 初始化默认策略
	err = InitDefaultPolicies()
	assert.NoError(t, err)

	// 验证一些默认策略
	hasPermission, err := globalCasbinManager.CheckPermission(model.RoleAdmin, model.ResourceUser, model.ActionRead)
	assert.NoError(t, err)
	assert.True(t, hasPermission)

	hasPermission, err = globalCasbinManager.CheckPermission(model.RoleUser, model.ResourceProduct, model.ActionRead)
	assert.NoError(t, err)
	assert.True(t, hasPermission)
}
