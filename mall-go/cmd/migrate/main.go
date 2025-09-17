package main

import (
	"fmt"
	"log"

	"mall-go/internal/model"

	"github.com/shopspring/decimal"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("🚀 开始数据库迁移...")

	// 数据库配置
	dbConfig := struct {
		Driver   string
		Host     string
		Port     int
		Username string
		Password string
		DBName   string
		Charset  string
	}{
		Driver:   "mysql",
		Host:     "localhost",
		Port:     3306,
		Username: "gomall",
		Password: "123456",
		DBName:   "gomall",
		Charset:  "utf8mb4",
	}

	// 直接使用root用户连接到MySQL服务器（不指定数据库）
	rootDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=%s&parseTime=True&loc=Local&timeout=30s",
		"root", "123456", dbConfig.Host, dbConfig.Port, dbConfig.Charset)

	rootDB, err := gorm.Open(mysql.Open(rootDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接MySQL服务器失败: %v", err)
	}

	// 删除并重新创建数据库以避免格式问题
	dropDBSQL := fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", dbConfig.DBName)
	if err := rootDB.Exec(dropDBSQL).Error; err != nil {
		log.Printf("删除数据库警告: %v", err)
	}

	createDBSQL := fmt.Sprintf("CREATE DATABASE `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", dbConfig.DBName)
	if err := rootDB.Exec(createDBSQL).Error; err != nil {
		log.Fatalf("创建数据库失败: %v", err)
	}
	fmt.Printf("✅ 数据库 '%s' 重新创建成功\n", dbConfig.DBName)

	// 直接使用root用户连接到目标数据库进行迁移
	targetDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local&timeout=30s&readTimeout=30s&writeTimeout=30s",
		"root", "123456", dbConfig.Host, dbConfig.Port, dbConfig.DBName, dbConfig.Charset)

	db, err := gorm.Open(mysql.Open(targetDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接目标数据库失败: %v", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("获取数据库连接失败: %v", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(300) // 5分钟

	fmt.Println("✅ 数据库连接成功")

	// 执行数据库迁移
	if err := migrate(db); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	fmt.Println("🎉 数据库迁移完成!")
}

func migrate(db *gorm.DB) error {
	fmt.Println("📋 开始创建数据表...")

	// 定义要迁移的模型
	models := []interface{}{
		&model.User{},
		&model.Category{},
		&model.Product{},
		&model.ProductImage{},
		&model.Cart{},
		&model.CartItem{},
		&model.Order{},
		&model.OrderItem{},
		&model.Payment{},
		&model.File{},
	}

	// 执行自动迁移
	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("迁移模型 %T 失败: %v", model, err)
		}
		fmt.Printf("✅ 表 %T 创建成功\n", model)
	}

	// 插入初始数据
	if err := seedData(db); err != nil {
		return fmt.Errorf("插入初始数据失败: %v", err)
	}

	return nil
}

func seedData(db *gorm.DB) error {
	fmt.Println("📋 开始插入初始数据...")

	// 检查是否已有数据
	var userCount int64
	db.Model(&model.User{}).Count(&userCount)
	if userCount > 0 {
		fmt.Println("⚠️  数据已存在，跳过初始数据插入")
		return nil
	}

	// 创建管理员用户
	adminUser := &model.User{
		Username: "admin",
		Email:    "admin@example.com",
		Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
		Role:     "admin",
		Status:   "active",
	}

	if err := db.Create(adminUser).Error; err != nil {
		return fmt.Errorf("创建管理员用户失败: %v", err)
	}
	fmt.Println("✅ 管理员用户创建成功")

	// 创建测试用户
	testUser := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
		Role:     "user",
		Status:   "active",
	}

	if err := db.Create(testUser).Error; err != nil {
		return fmt.Errorf("创建测试用户失败: %v", err)
	}
	fmt.Println("✅ 测试用户创建成功")

	// 创建示例商品
	price1, _ := decimal.NewFromString("8999.00")
	price2, _ := decimal.NewFromString("13999.00")
	price3, _ := decimal.NewFromString("699.00")
	price4, _ := decimal.NewFromString("1999.00")

	products := []*model.Product{
		{
			Name:        "iPhone 15 Pro Max 256GB 深空黑色",
			Description: "搭载A17 Pro芯片，支持5G网络，拥有强大的摄影系统和超长续航能力。",
			Price:       price1,
			Stock:       50,
			CategoryID:  1,
			Status:      "active",
		},
		{
			Name:        "MacBook Pro 14英寸 M3芯片",
			Description: "全新M3芯片，14英寸Liquid Retina XDR显示屏，专业级性能，适合创意工作者。",
			Price:       price2,
			Stock:       30,
			CategoryID:  1,
			Status:      "active",
		},
		{
			Name:        "Nike Air Max 270 运动鞋",
			Description: "经典Air Max气垫设计，舒适透气，适合日常运动和休闲穿着。",
			Price:       price3,
			Stock:       120,
			CategoryID:  2,
			Status:      "active",
		},
		{
			Name:        "Sony WH-1000XM5 无线降噪耳机",
			Description: "业界领先的降噪技术，30小时续航，高解析度音质，舒适佩戴。",
			Price:       price4,
			Stock:       45,
			CategoryID:  1,
			Status:      "active",
		},
	}

	for _, product := range products {
		if err := db.Create(product).Error; err != nil {
			return fmt.Errorf("创建商品失败: %v", err)
		}
	}
	fmt.Printf("✅ %d 个示例商品创建成功\n", len(products))

	// 创建示例购物车
	totalAmount, _ := decimal.NewFromString("0.00")
	cart := &model.Cart{
		UserID:      testUser.ID,
		TotalAmount: totalAmount,
		Status:      "active",
	}

	if err := db.Create(cart).Error; err != nil {
		return fmt.Errorf("创建购物车失败: %v", err)
	}

	// 添加购物车项目
	cartItems := []*model.CartItem{
		{
			CartID:    cart.ID,
			ProductID: products[0].ID,
			Quantity:  1,
			Price:     products[0].Price,
			Selected:  true,
		},
		{
			CartID:    cart.ID,
			ProductID: products[3].ID,
			Quantity:  1,
			Price:     products[3].Price,
			Selected:  true,
		},
	}

	for _, item := range cartItems {
		if err := db.Create(item).Error; err != nil {
			return fmt.Errorf("创建购物车项目失败: %v", err)
		}
	}
	fmt.Printf("✅ %d 个购物车项目创建成功\n", len(cartItems))

	return nil
}
