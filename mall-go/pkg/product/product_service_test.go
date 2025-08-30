package product

import (
	"testing"

	"mall-go/internal/model"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ProductServiceTestSuite 商品服务测试套件
type ProductServiceTestSuite struct {
	suite.Suite
	db             *gorm.DB
	productService *ProductService
}

// SetupSuite 设置测试套件
func (suite *ProductServiceTestSuite) SetupSuite() {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)
	suite.db = db

	// 自动迁移
	err = db.AutoMigrate(
		&model.Product{},
		&model.Category{},
		&model.Brand{},
		&model.ProductImage{},
		&model.ProductAttr{},
		&model.ProductSKU{},
		&model.ProductReview{},
		&model.User{},
	)
	suite.Require().NoError(err)

	// 创建测试数据
	suite.createTestData()

	// 创建商品服务
	suite.productService = NewProductService(db)
}

// createTestData 创建测试数据
func (suite *ProductServiceTestSuite) createTestData() {
	// 创建测试分类
	category := &model.Category{
		Name:     "测试分类",
		ParentID: 0,
		Level:    1,
		Status:   model.CategoryStatusActive,
	}
	suite.db.Create(category)

	// 创建测试品牌
	brand := &model.Brand{
		Name:   "测试品牌",
		Status: model.BrandStatusActive,
	}
	suite.db.Create(brand)

	// 创建测试用户（商家）
	user := &model.User{
		Username: "testmerchant",
		Email:    "merchant@test.com",
		Role:     model.RoleMerchant,
		Status:   model.StatusActive,
	}
	suite.db.Create(user)
}

// TestProductService_CreateProduct 测试创建商品
func (suite *ProductServiceTestSuite) TestProductService_CreateProduct() {
	req := &CreateProductRequest{
		Name:        "测试商品",
		Description: "这是一个测试商品",
		CategoryID:  1,
		BrandID:     1,
		MerchantID:  1,
		Price:       decimal.NewFromFloat(99.99),
		Stock:       100,
		Images:      []string{"http://example.com/image1.jpg", "http://example.com/image2.jpg"},
		Attributes: []ProductAttributeRequest{
			{AttrName: "颜色", AttrValue: "红色", Sort: 1},
			{AttrName: "尺寸", AttrValue: "L", Sort: 2},
		},
	}

	product, err := suite.productService.CreateProduct(req)
	suite.NoError(err)
	suite.NotNil(product)

	// 验证商品基本信息
	suite.Equal("测试商品", product.Name)
	suite.Equal("这是一个测试商品", product.Description)
	suite.Equal(uint(1), product.CategoryID)
	suite.Equal(uint(1), product.BrandID)
	suite.Equal(uint(1), product.MerchantID)
	suite.True(product.Price.Equal(decimal.NewFromFloat(99.99)))
	suite.Equal(100, product.Stock)
	suite.Equal(model.ProductStatusDraft, product.Status)

	// 验证数据库记录
	var dbProduct model.Product
	err = suite.db.Preload("Images").Preload("Attributes").First(&dbProduct, product.ID).Error
	suite.NoError(err)
	suite.Equal("测试商品", dbProduct.Name)
	suite.Len(dbProduct.Images, 2)
	suite.Len(dbProduct.Attributes, 2)
}

// TestProductService_UpdateProduct 测试更新商品
func (suite *ProductServiceTestSuite) TestProductService_UpdateProduct() {
	// 先创建一个商品
	createReq := &CreateProductRequest{
		Name:       "原始商品",
		CategoryID: 1,
		MerchantID: 1,
		Price:      decimal.NewFromFloat(50.00),
		Stock:      50,
	}

	product, err := suite.productService.CreateProduct(createReq)
	suite.NoError(err)

	// 更新商品
	updateReq := &UpdateProductRequest{
		Name:        "更新后的商品",
		CategoryID:  1,
		Price:       decimal.NewFromFloat(75.00),
		Stock:       75,
		Status:      model.ProductStatusActive,
		IsHot:       true,
		IsRecommend: true,
	}

	updatedProduct, err := suite.productService.UpdateProduct(product.ID, updateReq)
	suite.NoError(err)
	suite.NotNil(updatedProduct)

	// 验证更新结果
	suite.Equal("更新后的商品", updatedProduct.Name)
	suite.True(updatedProduct.Price.Equal(decimal.NewFromFloat(75.00)))
	suite.Equal(75, updatedProduct.Stock)
	suite.Equal(model.ProductStatusActive, updatedProduct.Status)
	suite.True(updatedProduct.IsHot)
	suite.True(updatedProduct.IsRecommend)
}

// TestProductService_DeleteProduct 测试删除商品
func (suite *ProductServiceTestSuite) TestProductService_DeleteProduct() {
	// 先创建一个商品
	createReq := &CreateProductRequest{
		Name:       "待删除商品",
		CategoryID: 1,
		MerchantID: 1,
		Price:      decimal.NewFromFloat(30.00),
		Stock:      30,
	}

	product, err := suite.productService.CreateProduct(createReq)
	suite.NoError(err)

	// 删除商品
	err = suite.productService.DeleteProduct(product.ID)
	suite.NoError(err)

	// 验证商品已被软删除
	var deletedProduct model.Product
	err = suite.db.Unscoped().First(&deletedProduct, product.ID).Error
	suite.NoError(err)
	suite.NotNil(deletedProduct.DeletedAt)
}

// TestProductService_GetProduct 测试获取商品详情
func (suite *ProductServiceTestSuite) TestProductService_GetProduct() {
	// 先创建一个商品
	createReq := &CreateProductRequest{
		Name:       "详情测试商品",
		CategoryID: 1,
		MerchantID: 1,
		Price:      decimal.NewFromFloat(40.00),
		Stock:      40,
	}

	product, err := suite.productService.CreateProduct(createReq)
	suite.NoError(err)

	// 获取商品详情
	productDetail, err := suite.productService.GetProduct(product.ID)
	suite.NoError(err)
	suite.NotNil(productDetail)

	// 验证商品信息
	suite.Equal("详情测试商品", productDetail.Name)
	suite.True(productDetail.Price.Equal(decimal.NewFromFloat(40.00)))
	suite.Equal(40, productDetail.Stock)

	// 验证关联数据已加载
	suite.NotNil(productDetail.Category)
	suite.NotNil(productDetail.Brand)
}

// TestProductService_GetProductList 测试获取商品列表
func (suite *ProductServiceTestSuite) TestProductService_GetProductList() {
	// 创建多个测试商品
	for i := 1; i <= 5; i++ {
		createReq := &CreateProductRequest{
			Name:       fmt.Sprintf("列表测试商品%d", i),
			CategoryID: 1,
			MerchantID: 1,
			Price:      decimal.NewFromFloat(float64(i * 10)),
			Stock:      i * 10,
		}
		suite.productService.CreateProduct(createReq)
	}

	// 测试基本列表查询
	req := &ProductListRequest{
		Page:     1,
		PageSize: 10,
	}

	products, total, err := suite.productService.GetProductList(req)
	suite.NoError(err)
	suite.NotNil(products)
	suite.True(total >= 5) // 至少有5个商品

	// 测试分类筛选
	categoryID := uint(1)
	req.CategoryID = &categoryID
	products, total, err = suite.productService.GetProductList(req)
	suite.NoError(err)
	suite.True(total >= 5)

	// 测试关键词搜索
	req.Keyword = "列表测试"
	products, total, err = suite.productService.GetProductList(req)
	suite.NoError(err)
	suite.True(total >= 5)
}

// TestProductService_UpdateProductStatus 测试更新商品状态
func (suite *ProductServiceTestSuite) TestProductService_UpdateProductStatus() {
	// 先创建一个商品
	createReq := &CreateProductRequest{
		Name:       "状态测试商品",
		CategoryID: 1,
		MerchantID: 1,
		Price:      decimal.NewFromFloat(60.00),
		Stock:      60,
	}

	product, err := suite.productService.CreateProduct(createReq)
	suite.NoError(err)
	suite.Equal(model.ProductStatusDraft, product.Status)

	// 更新状态为上架
	err = suite.productService.UpdateProductStatus(product.ID, model.ProductStatusActive)
	suite.NoError(err)

	// 验证状态已更新
	updatedProduct, err := suite.productService.GetProduct(product.ID)
	suite.NoError(err)
	suite.Equal(model.ProductStatusActive, updatedProduct.Status)

	// 测试无效状态
	err = suite.productService.UpdateProductStatus(product.ID, "invalid_status")
	suite.Error(err)
	suite.Contains(err.Error(), "无效的状态值")
}

// TestProductService_StockOperations 测试库存操作
func (suite *ProductServiceTestSuite) TestProductService_StockOperations() {
	// 先创建一个商品
	createReq := &CreateProductRequest{
		Name:       "库存测试商品",
		CategoryID: 1,
		MerchantID: 1,
		Price:      decimal.NewFromFloat(80.00),
		Stock:      100,
	}

	product, err := suite.productService.CreateProduct(createReq)
	suite.NoError(err)

	// 测试扣减库存
	err = suite.productService.DeductStock(product.ID, 30)
	suite.NoError(err)

	// 验证库存已扣减
	updatedProduct, err := suite.productService.GetProduct(product.ID)
	suite.NoError(err)
	suite.Equal(70, updatedProduct.Stock)
	suite.Equal(30, updatedProduct.SoldCount)

	// 测试恢复库存
	err = suite.productService.RestoreStock(product.ID, 10)
	suite.NoError(err)

	// 验证库存已恢复
	updatedProduct, err = suite.productService.GetProduct(product.ID)
	suite.NoError(err)
	suite.Equal(80, updatedProduct.Stock)
	suite.Equal(20, updatedProduct.SoldCount)

	// 测试库存不足的情况
	err = suite.productService.DeductStock(product.ID, 100)
	suite.Error(err)
	suite.Contains(err.Error(), "库存不足")
}

// TestProductService_GetHotProducts 测试获取热销商品
func (suite *ProductServiceTestSuite) TestProductService_GetHotProducts() {
	// 创建热销商品
	createReq := &CreateProductRequest{
		Name:       "热销商品",
		CategoryID: 1,
		MerchantID: 1,
		Price:      decimal.NewFromFloat(120.00),
		Stock:      50,
		IsHot:      true,
	}

	product, err := suite.productService.CreateProduct(createReq)
	suite.NoError(err)

	// 设置为上架状态
	suite.productService.UpdateProductStatus(product.ID, model.ProductStatusActive)

	// 获取热销商品
	hotProducts, err := suite.productService.GetHotProducts(10)
	suite.NoError(err)
	suite.True(len(hotProducts) >= 1)

	// 验证返回的是热销商品
	found := false
	for _, p := range hotProducts {
		if p.ID == product.ID {
			found = true
			suite.True(p.IsHot)
			break
		}
	}
	suite.True(found, "应该找到创建的热销商品")
}

// TestProductService_BatchUpdateProductStatus 测试批量更新商品状态
func (suite *ProductServiceTestSuite) TestProductService_BatchUpdateProductStatus() {
	// 创建多个商品
	var productIDs []uint
	for i := 1; i <= 3; i++ {
		createReq := &CreateProductRequest{
			Name:       fmt.Sprintf("批量测试商品%d", i),
			CategoryID: 1,
			MerchantID: 1,
			Price:      decimal.NewFromFloat(float64(i * 20)),
			Stock:      i * 20,
		}
		product, err := suite.productService.CreateProduct(createReq)
		suite.NoError(err)
		productIDs = append(productIDs, product.ID)
	}

	// 批量更新状态
	err := suite.productService.BatchUpdateProductStatus(productIDs, model.ProductStatusActive)
	suite.NoError(err)

	// 验证所有商品状态都已更新
	for _, id := range productIDs {
		product, err := suite.productService.GetProduct(id)
		suite.NoError(err)
		suite.Equal(model.ProductStatusActive, product.Status)
	}
}

// 运行商品服务测试套件
func TestProductServiceSuite(t *testing.T) {
	suite.Run(t, new(ProductServiceTestSuite))
}

// TestNewProductService 测试创建商品服务
func TestNewProductService(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	service := NewProductService(db)
	assert.NotNil(t, service)
	assert.Equal(t, db, service.db)
}
