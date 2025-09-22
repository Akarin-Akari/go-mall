package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// 简化的模型结构用于数据初始化
type User struct {
	ID        uint      `gorm:"primaryKey"`
	Username  string    `gorm:"uniqueIndex"`
	Email     string    `gorm:"uniqueIndex"`
	Password  string
	Nickname  string
	Role      string    `gorm:"default:user"`
	Status    string    `gorm:"default:active"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Category struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string
	Description string
	Status      string `gorm:"default:active"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Product struct {
	ID          uint    `gorm:"primaryKey"`
	Name        string
	Description string
	Price       float64
	Stock       int
	CategoryID  uint
	Status      string `gorm:"default:active"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Address struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint
	Name      string
	Phone     string
	Province  string
	City      string
	District  string
	Detail    string
	IsDefault bool `gorm:"default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func main() {
	fmt.Println("🚀 开始初始化Mall-Go测试数据...")

	// 连接数据库
	db, err := gorm.Open(sqlite.Open("mall-go/mall_go.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	fmt.Println("✅ 数据库连接成功")

	// 自动迁移表结构
	err = db.AutoMigrate(&User{}, &Category{}, &Product{}, &Address{})
	if err != nil {
		log.Fatal("数据库迁移失败:", err)
	}

	fmt.Println("✅ 数据库表结构迁移完成")

	// 初始化分类数据
	categories := []Category{
		{ID: 1, Name: "电子产品", Description: "手机、电脑、数码产品等"},
		{ID: 2, Name: "服装鞋帽", Description: "男装、女装、鞋子、配饰等"},
		{ID: 3, Name: "家居用品", Description: "家具、装饰、生活用品等"},
		{ID: 4, Name: "图书音像", Description: "图书、音乐、影视等"},
		{ID: 5, Name: "运动户外", Description: "运动器材、户外用品等"},
	}

	for _, category := range categories {
		var existingCategory Category
		if err := db.First(&existingCategory, category.ID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				db.Create(&category)
				fmt.Printf("✅ 创建分类: %s\n", category.Name)
			}
		} else {
			fmt.Printf("⚠️ 分类已存在: %s\n", category.Name)
		}
	}

	// 初始化商品数据
	products := []Product{
		{ID: 1, Name: "iPhone 15 Pro", Description: "苹果最新旗舰手机", Price: 7999.00, Stock: 50, CategoryID: 1},
		{ID: 2, Name: "MacBook Pro", Description: "苹果笔记本电脑", Price: 12999.00, Stock: 30, CategoryID: 1},
		{ID: 3, Name: "Nike运动鞋", Description: "舒适透气运动鞋", Price: 599.00, Stock: 100, CategoryID: 2},
		{ID: 4, Name: "无线蓝牙耳机", Description: "高音质无线耳机", Price: 299.00, Stock: 200, CategoryID: 1},
		{ID: 5, Name: "智能手表", Description: "多功能智能手表", Price: 1299.00, Stock: 80, CategoryID: 1},
		{ID: 6, Name: "休闲T恤", Description: "纯棉舒适T恤", Price: 89.00, Stock: 150, CategoryID: 2},
		{ID: 7, Name: "办公椅", Description: "人体工学办公椅", Price: 899.00, Stock: 40, CategoryID: 3},
		{ID: 8, Name: "编程入门书籍", Description: "适合初学者的编程书", Price: 59.00, Stock: 120, CategoryID: 4},
		{ID: 9, Name: "瑜伽垫", Description: "防滑瑜伽垫", Price: 129.00, Stock: 90, CategoryID: 5},
		{ID: 10, Name: "咖啡杯", Description: "陶瓷咖啡杯", Price: 39.00, Stock: 200, CategoryID: 3},
	}

	for _, product := range products {
		var existingProduct Product
		if err := db.First(&existingProduct, product.ID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				db.Create(&product)
				fmt.Printf("✅ 创建商品: %s (¥%.2f)\n", product.Name, product.Price)
			}
		} else {
			fmt.Printf("⚠️ 商品已存在: %s\n", product.Name)
		}
	}

	// 创建管理员用户
	adminUser := User{
		ID:       1,
		Username: "admin",
		Email:    "admin@mall-go.com",
		Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
		Nickname: "系统管理员",
		Role:     "admin",
		Status:   "active",
	}

	var existingAdmin User
	if err := db.First(&existingAdmin, adminUser.ID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			db.Create(&adminUser)
			fmt.Printf("✅ 创建管理员用户: %s\n", adminUser.Username)
		}
	} else {
		fmt.Printf("⚠️ 管理员用户已存在: %s\n", adminUser.Username)
	}

	// 创建测试用户
	testUser := User{
		ID:       2,
		Username: "testuser",
		Email:    "test@example.com",
		Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
		Nickname: "测试用户",
		Role:     "user",
		Status:   "active",
	}

	var existingTestUser User
	if err := db.First(&existingTestUser, testUser.ID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			db.Create(&testUser)
			fmt.Printf("✅ 创建测试用户: %s\n", testUser.Username)
		}
	} else {
		fmt.Printf("⚠️ 测试用户已存在: %s\n", testUser.Username)
	}

	// 创建测试地址
	addresses := []Address{
		{ID: 1, UserID: 2, Name: "张三", Phone: "13800138000", Province: "北京市", City: "北京市", District: "朝阳区", Detail: "某某街道123号", IsDefault: true},
		{ID: 2, UserID: 2, Name: "李四", Phone: "13900139000", Province: "上海市", City: "上海市", District: "浦东新区", Detail: "某某路456号", IsDefault: false},
	}

	for _, address := range addresses {
		var existingAddress Address
		if err := db.First(&existingAddress, address.ID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				db.Create(&address)
				fmt.Printf("✅ 创建地址: %s - %s\n", address.Name, address.Detail)
			}
		} else {
			fmt.Printf("⚠️ 地址已存在: %s\n", address.Detail)
		}
	}

	fmt.Println("\n🎉 测试数据初始化完成！")
	fmt.Println("📊 数据统计:")
	
	var categoryCount, productCount, userCount, addressCount int64
	db.Model(&Category{}).Count(&categoryCount)
	db.Model(&Product{}).Count(&productCount)
	db.Model(&User{}).Count(&userCount)
	db.Model(&Address{}).Count(&addressCount)
	
	fmt.Printf("   分类数量: %d\n", categoryCount)
	fmt.Printf("   商品数量: %d\n", productCount)
	fmt.Printf("   用户数量: %d\n", userCount)
	fmt.Printf("   地址数量: %d\n", addressCount)
	
	fmt.Println("\n💡 测试账户信息:")
	fmt.Println("   管理员: admin / password")
	fmt.Println("   测试用户: testuser / password")
	
	fmt.Println("\n✅ 现在可以重新运行API测试了！")
}
