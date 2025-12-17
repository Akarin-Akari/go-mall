package inventory

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"mall-go/internal/model"
	"mall-go/pkg/database"

	"github.com/go-redis/redis/v8"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// InventoryTestSuite 库存测试套件
type InventoryTestSuite struct {
	suite.Suite
	db               *gorm.DB
	rdb              *redis.Client
	inventoryService *InventoryService
	testProduct      *model.Product
	testSKU          *model.ProductSKU
}

// SetupSuite 设置测试套件
func (suite *InventoryTestSuite) SetupSuite() {
	// 初始化测试数据库
	suite.db = database.InitTestDB()

	// 初始化Redis
	suite.rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1, // 使用测试数据库
	})

	// 创建库存服务
	suite.inventoryService = NewInventoryService(suite.db, suite.rdb)

	// 迁移表结构
	suite.db.AutoMigrate(&model.Product{}, &model.ProductSKU{})
}

// TearDownSuite 清理测试套件
func (suite *InventoryTestSuite) TearDownSuite() {
	// 清理测试数据
	suite.db.Exec("DROP TABLE IF EXISTS products")
	suite.db.Exec("DROP TABLE IF EXISTS product_skus")

	// 关闭连接
	sqlDB, _ := suite.db.DB()
	sqlDB.Close()
	suite.rdb.Close()
}

// SetupTest 设置每个测试
func (suite *InventoryTestSuite) SetupTest() {
	// 创建测试商品
	suite.testProduct = &model.Product{
		Name:        "测试商品",
		Price:       decimal.NewFromFloat(99.99),
		Stock:       1000,
		Version:     0,
		Status:      model.ProductStatusActive,
		Description: "测试商品描述",
		CategoryID:  1,
	}
	suite.db.Create(suite.testProduct)

	// 创建测试SKU
	suite.testSKU = &model.ProductSKU{
		ProductID: suite.testProduct.ID,
		SKUCode:   "TEST-SKU-001",
		Name:      "测试SKU",
		Price:     decimal.NewFromFloat(89.99),
		Stock:     500,
		Version:   0,
		Status:    model.SKUStatusActive,
	}
	suite.db.Create(suite.testSKU)
}

// TearDownTest 清理每个测试
func (suite *InventoryTestSuite) TearDownTest() {
	// 清理测试数据
	suite.db.Delete(&model.ProductSKU{}, "product_id = ?", suite.testProduct.ID)
	suite.db.Delete(&model.Product{}, "id = ?", suite.testProduct.ID)

	// 清理Redis锁
	suite.rdb.FlushDB(context.Background())
}

// TestConcurrentStockDeduction 测试并发库存扣减
func (suite *InventoryTestSuite) TestConcurrentStockDeduction() {
	t := suite.T()

	// 测试参数
	goroutineCount := 100
	deductQuantity := 5
	expectedFinalStock := suite.testProduct.Stock - (goroutineCount * deductQuantity)

	var wg sync.WaitGroup
	var successCount int32
	var failureCount int32
	var mutex sync.Mutex

	// 启动多个goroutine并发扣减库存
	for i := 0; i < goroutineCount; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			requests := []StockDeductionRequest{
				{
					ProductID: suite.testProduct.ID,
					Quantity:  deductQuantity,
				},
			}

			results, err := suite.inventoryService.DeductStockWithOptimisticLock(requests)

			mutex.Lock()
			if err != nil || !results[0].Success {
				failureCount++
				t.Logf("Goroutine %d failed: %v", index, err)
			} else {
				successCount++
				t.Logf("Goroutine %d succeeded", index)
			}
			mutex.Unlock()
		}(i)
	}

	// 等待所有goroutine完成
	wg.Wait()

	// 验证结果
	var finalProduct model.Product
	suite.db.First(&finalProduct, suite.testProduct.ID)

	t.Logf("Initial stock: %d", suite.testProduct.Stock)
	t.Logf("Final stock: %d", finalProduct.Stock)
	t.Logf("Success count: %d", successCount)
	t.Logf("Failure count: %d", failureCount)
	t.Logf("Expected final stock: %d", expectedFinalStock)

	// 断言：最终库存应该正确
	assert.Equal(t, expectedFinalStock, finalProduct.Stock, "最终库存应该正确")

	// 断言：成功次数应该等于goroutine数量
	assert.Equal(t, int32(goroutineCount), successCount, "所有扣减操作都应该成功")

	// 断言：失败次数应该为0（因为库存充足）
	assert.Equal(t, int32(0), failureCount, "不应该有失败的扣减操作")
}

// TestConcurrentStockDeductionWithInsufficientStock 测试库存不足时的并发扣减
func (suite *InventoryTestSuite) TestConcurrentStockDeductionWithInsufficientStock() {
	t := suite.T()

	// 设置较小的初始库存
	suite.db.Model(suite.testProduct).Update("stock", 50)

	// 测试参数
	goroutineCount := 100
	deductQuantity := 1

	var wg sync.WaitGroup
	var successCount int32
	var failureCount int32
	var mutex sync.Mutex

	// 启动多个goroutine并发扣减库存
	for i := 0; i < goroutineCount; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			requests := []StockDeductionRequest{
				{
					ProductID: suite.testProduct.ID,
					Quantity:  deductQuantity,
				},
			}

			results, err := suite.inventoryService.DeductStockWithOptimisticLock(requests)

			mutex.Lock()
			if err != nil || !results[0].Success {
				failureCount++
			} else {
				successCount++
			}
			mutex.Unlock()
		}(i)
	}

	// 等待所有goroutine完成
	wg.Wait()

	// 验证结果
	var finalProduct model.Product
	suite.db.First(&finalProduct, suite.testProduct.ID)

	t.Logf("Initial stock: 50")
	t.Logf("Final stock: %d", finalProduct.Stock)
	t.Logf("Success count: %d", successCount)
	t.Logf("Failure count: %d", failureCount)

	// 断言：最终库存不应该为负数
	assert.GreaterOrEqual(t, finalProduct.Stock, 0, "最终库存不应该为负数")

	// 断言：成功次数应该等于初始库存
	assert.Equal(t, int32(50), successCount, "成功次数应该等于初始库存")

	// 断言：失败次数应该等于剩余的goroutine数量
	assert.Equal(t, int32(50), failureCount, "失败次数应该正确")

	// 断言：最终库存应该为0
	assert.Equal(t, 0, finalProduct.Stock, "最终库存应该为0")
}

// TestSKUConcurrentStockDeduction 测试SKU并发库存扣减
func (suite *InventoryTestSuite) TestSKUConcurrentStockDeduction() {
	t := suite.T()

	// 测试参数
	goroutineCount := 50
	deductQuantity := 2
	expectedFinalStock := suite.testSKU.Stock - (goroutineCount * deductQuantity)

	var wg sync.WaitGroup
	var successCount int32
	var failureCount int32
	var mutex sync.Mutex

	// 启动多个goroutine并发扣减SKU库存
	for i := 0; i < goroutineCount; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			requests := []StockDeductionRequest{
				{
					ProductID: suite.testProduct.ID,
					SKUID:     suite.testSKU.ID,
					Quantity:  deductQuantity,
				},
			}

			results, err := suite.inventoryService.DeductStockWithOptimisticLock(requests)

			mutex.Lock()
			if err != nil || !results[0].Success {
				failureCount++
			} else {
				successCount++
			}
			mutex.Unlock()
		}(i)
	}

	// 等待所有goroutine完成
	wg.Wait()

	// 验证结果
	var finalSKU model.ProductSKU
	suite.db.First(&finalSKU, suite.testSKU.ID)

	t.Logf("Initial SKU stock: %d", suite.testSKU.Stock)
	t.Logf("Final SKU stock: %d", finalSKU.Stock)
	t.Logf("Success count: %d", successCount)
	t.Logf("Failure count: %d", failureCount)

	// 断言：最终SKU库存应该正确
	assert.Equal(t, expectedFinalStock, finalSKU.Stock, "最终SKU库存应该正确")

	// 断言：成功次数应该等于goroutine数量
	assert.Equal(t, int32(goroutineCount), successCount, "所有扣减操作都应该成功")
}

// TestStockRestore 测试库存恢复
func (suite *InventoryTestSuite) TestStockRestore() {
	t := suite.T()

	// 先扣减一些库存
	requests := []StockDeductionRequest{
		{
			ProductID: suite.testProduct.ID,
			Quantity:  100,
		},
	}

	_, err := suite.inventoryService.DeductStockWithOptimisticLock(requests)
	assert.NoError(t, err, "库存扣减应该成功")

	// 验证库存已扣减
	var product model.Product
	suite.db.First(&product, suite.testProduct.ID)
	assert.Equal(t, suite.testProduct.Stock-100, product.Stock, "库存应该已扣减")

	// 恢复库存
	err = suite.inventoryService.RestoreStock(requests)
	assert.NoError(t, err, "库存恢复应该成功")

	// 验证库存已恢复
	suite.db.First(&product, suite.testProduct.ID)
	assert.Equal(t, suite.testProduct.Stock, product.Stock, "库存应该已恢复")
}

// TestCheckStock 测试库存检查
func (suite *InventoryTestSuite) TestCheckStock() {
	t := suite.T()

	// 测试库存充足的情况
	requests := []StockDeductionRequest{
		{
			ProductID: suite.testProduct.ID,
			Quantity:  100,
		},
	}

	sufficient, err := suite.inventoryService.CheckStock(requests)
	assert.NoError(t, err, "库存检查不应该出错")
	assert.True(t, sufficient, "库存应该充足")

	// 测试库存不足的情况
	requests[0].Quantity = 2000
	sufficient, err = suite.inventoryService.CheckStock(requests)
	assert.Error(t, err, "库存检查应该返回错误")
	assert.False(t, sufficient, "库存应该不足")
}

// TestGetStockInfo 测试获取库存信息
func (suite *InventoryTestSuite) TestGetStockInfo() {
	t := suite.T()

	// 测试获取商品库存
	stock, err := suite.inventoryService.GetStockInfo(suite.testProduct.ID, 0)
	assert.NoError(t, err, "获取商品库存不应该出错")
	assert.Equal(t, suite.testProduct.Stock, stock, "商品库存应该正确")

	// 测试获取SKU库存
	stock, err = suite.inventoryService.GetStockInfo(suite.testProduct.ID, suite.testSKU.ID)
	assert.NoError(t, err, "获取SKU库存不应该出错")
	assert.Equal(t, suite.testSKU.Stock, stock, "SKU库存应该正确")
}

// TestInventoryTestSuite 运行库存测试套件
func TestInventoryTestSuite(t *testing.T) {
	suite.Run(t, new(InventoryTestSuite))
}
