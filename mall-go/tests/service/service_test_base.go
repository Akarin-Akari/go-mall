package service

import (
	"testing"

	"mall-go/internal/config"
	"mall-go/internal/model"
	testConfig "mall-go/tests/config"
	"mall-go/tests/helpers"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// ServiceTestSuite Service层测试基础套件
type ServiceTestSuite struct {
	suite.Suite
	db     *gorm.DB
	helper *helpers.TestHelper
}

// SetupSuite 测试套件初始化
func (suite *ServiceTestSuite) SetupSuite() {
	// 初始化全局配置
	config.GlobalConfig = config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key-for-service-layer-testing",
			Expire: "24h",
		},
		Database: config.DatabaseConfig{
			Driver: "sqlite",
			DBName: ":memory:",
		},
	}

	// 初始化测试数据库
	suite.db = testConfig.SetupTestDB()

	// 自动迁移所有测试表
	err := suite.db.AutoMigrate(
		// 用户相关
		&model.User{},

		// 商品相关
		&model.Product{},
		&model.ProductImage{},
		&model.Category{},

		// 购物车相关
		&model.Cart{},
		&model.CartItem{},

		// 订单相关
		&model.Order{},
		&model.OrderItem{},

		// 支付相关
		&model.Payment{},

		// 其他
		&model.File{},
	)
	suite.Require().NoError(err, "数据库迁移失败")

	// 初始化测试辅助工具
	suite.helper = helpers.NewTestHelper(suite.db)
}

// TearDownSuite 测试套件清理
func (suite *ServiceTestSuite) TearDownSuite() {
	if suite.db != nil {
		testConfig.CleanupTestDB(suite.db)
	}
}

// SetupTest 每个测试前的准备
func (suite *ServiceTestSuite) SetupTest() {
	// 清理测试数据
	suite.helper.CleanupTestData()
}

// TearDownTest 每个测试后的清理
func (suite *ServiceTestSuite) TearDownTest() {
	// 清理测试数据
	suite.helper.CleanupTestData()
}

// GetDB 获取数据库连接
func (suite *ServiceTestSuite) GetDB() *gorm.DB {
	return suite.db
}

// GetHelper 获取测试辅助工具
func (suite *ServiceTestSuite) GetHelper() *helpers.TestHelper {
	return suite.helper
}

// CreateTestUser 创建测试用户
func (suite *ServiceTestSuite) CreateTestUser(username, email, password string) *model.User {
	return suite.helper.CreateTestUser(username, email, password)
}

// CreateTestCategory 创建测试分类
func (suite *ServiceTestSuite) CreateTestCategory(name, description string) *model.Category {
	return suite.helper.CreateTestCategory(name, description)
}

// CreateTestProduct 创建测试商品
func (suite *ServiceTestSuite) CreateTestProduct(name, price string, stock int) *model.Product {
	return suite.helper.CreateTestProduct(name, price, stock)
}

// AssertBusinessRule 断言业务规则
func (suite *ServiceTestSuite) AssertBusinessRule(condition bool, message string) {
	suite.True(condition, "业务规则验证失败: "+message)
}

// AssertErrorContains 断言错误包含指定信息
func (suite *ServiceTestSuite) AssertErrorContains(err error, expectedMsg string) {
	suite.Error(err, "期望有错误发生")
	suite.Contains(err.Error(), expectedMsg, "错误信息不匹配")
}

// AssertNoError 断言无错误
func (suite *ServiceTestSuite) AssertNoError(err error, message string) {
	suite.NoError(err, message)
}

// RunServiceTest 运行Service层测试的通用方法
func RunServiceTest(t *testing.T, testSuite suite.TestingSuite) {
	suite.Run(t, testSuite)
}

// MockServiceConfig Service层测试的Mock配置
type MockServiceConfig struct {
	EnableRedis   bool
	EnablePayment bool
	EnableSMS     bool
	EnableEmail   bool
}

// DefaultMockServiceConfig 默认Mock配置
func DefaultMockServiceConfig() *MockServiceConfig {
	return &MockServiceConfig{
		EnableRedis:   false, // 默认不启用Redis
		EnablePayment: false, // 默认不启用支付
		EnableSMS:     false, // 默认不启用短信
		EnableEmail:   false, // 默认不启用邮件
	}
}

// ServiceTestCase Service层测试用例结构
type ServiceTestCase struct {
	Name           string
	Description    string
	SetupData      func(*ServiceTestSuite)
	ExecuteAction  func(*ServiceTestSuite) (interface{}, error)
	ValidateResult func(*ServiceTestSuite, interface{}, error)
	CleanupData    func(*ServiceTestSuite)
}

// RunServiceTestCases 批量运行Service层测试用例
func (suite *ServiceTestSuite) RunServiceTestCases(testCases []ServiceTestCase) {
	for _, tc := range testCases {
		suite.Run(tc.Name, func() {
			// 设置测试数据
			if tc.SetupData != nil {
				tc.SetupData(suite)
			}

			// 执行测试动作
			result, err := tc.ExecuteAction(suite)

			// 验证结果
			tc.ValidateResult(suite, result, err)

			// 清理测试数据
			if tc.CleanupData != nil {
				tc.CleanupData(suite)
			}
		})
	}
}

// BusinessLogicTestHelper 业务逻辑测试辅助工具
type BusinessLogicTestHelper struct {
	suite *ServiceTestSuite
}

// NewBusinessLogicTestHelper 创建业务逻辑测试辅助工具
func NewBusinessLogicTestHelper(suite *ServiceTestSuite) *BusinessLogicTestHelper {
	return &BusinessLogicTestHelper{suite: suite}
}

// ValidateBusinessRule 验证业务规则
func (h *BusinessLogicTestHelper) ValidateBusinessRule(ruleName string, condition bool, errorMsg string) {
	h.suite.True(condition, "业务规则[%s]验证失败: %s", ruleName, errorMsg)
}

// ValidateDataConsistency 验证数据一致性
func (h *BusinessLogicTestHelper) ValidateDataConsistency(description string, validator func() bool) {
	h.suite.True(validator(), "数据一致性验证失败: "+description)
}

// ValidateConcurrencyScenario 验证并发场景
func (h *BusinessLogicTestHelper) ValidateConcurrencyScenario(description string, concurrentAction func() error) {
	err := concurrentAction()
	h.suite.NoError(err, "并发场景验证失败: "+description)
}
