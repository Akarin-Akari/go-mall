package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	fmt.Println("🔧 创建MALL_GO数据库...")

	// 连接MySQL服务器（不指定数据库）
	dsn := "root:123456@tcp(localhost:3306)/?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("❌ 连接MySQL失败: %v", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatalf("❌ MySQL服务不可用: %v", err)
	}
	fmt.Println("✅ MySQL服务连接成功")

	// 创建数据库
	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS MALL_GO CHARACTER SET UTF8MB4 COLLATE UTF8MB4_UNICODE_CI")
	if err != nil {
		log.Fatalf("❌ 创建数据库失败: %v", err)
	}
	fmt.Println("✅ 数据库MALL_GO创建成功")

	// 验证数据库是否存在
	var dbName string
	err = db.QueryRow("SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = 'MALL_GO'").Scan(&dbName)
	if err != nil {
		log.Fatalf("❌ 验证数据库失败: %v", err)
	}
	fmt.Printf("✅ 数据库验证成功: %s\n", dbName)

	// 连接到新创建的数据库
	dsn = "root:123456@tcp(localhost:3306)/MALL_GO?charset=utf8mb4&parseTime=True&loc=Local"
	mallDB, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("❌ 连接MALL_GO数据库失败: %v", err)
	}
	defer mallDB.Close()

	if err := mallDB.Ping(); err != nil {
		log.Fatalf("❌ MALL_GO数据库不可用: %v", err)
	}
	fmt.Println("✅ MALL_GO数据库连接成功")

	// 创建基础表结构
	tables := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(50) NOT NULL UNIQUE,
			email VARCHAR(100) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL,
			nickname VARCHAR(50),
			avatar VARCHAR(255),
			phone VARCHAR(20),
			role VARCHAR(20) DEFAULT 'user',
			status VARCHAR(20) DEFAULT 'active',
			email_verified BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP NULL,
			INDEX idx_username (username),
			INDEX idx_email (email),
			INDEX idx_status (status)
		)`,
		`CREATE TABLE IF NOT EXISTS categories (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			description VARCHAR(500),
			parent_id BIGINT UNSIGNED NULL,
			level INT DEFAULT 1,
			sort INT DEFAULT 0,
			status VARCHAR(20) DEFAULT 'active',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP NULL,
			INDEX idx_parent_id (parent_id),
			INDEX idx_status (status),
			INDEX idx_sort (sort)
		)`,
		`CREATE TABLE IF NOT EXISTS products (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(200) NOT NULL,
			description TEXT,
			price DECIMAL(10, 2) NOT NULL,
			stock INT NOT NULL DEFAULT 0,
			category_id BIGINT UNSIGNED NOT NULL,
			status VARCHAR(20) DEFAULT 'active',
			images JSON,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP NULL,
			INDEX idx_category_id (category_id),
			INDEX idx_status (status),
			INDEX idx_price (price)
		)`,
		`CREATE TABLE IF NOT EXISTS orders (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			order_no VARCHAR(50) NOT NULL UNIQUE,
			user_id BIGINT UNSIGNED NOT NULL,
			total_amount DECIMAL(10, 2) NOT NULL,
			status VARCHAR(20) DEFAULT 'pending',
			payment_status VARCHAR(20) DEFAULT 'unpaid',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP NULL,
			INDEX idx_order_no (order_no),
			INDEX idx_user_id (user_id),
			INDEX idx_status (status)
		)`,
		`CREATE TABLE IF NOT EXISTS order_items (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			order_id BIGINT UNSIGNED NOT NULL,
			product_id BIGINT UNSIGNED NOT NULL,
			quantity INT NOT NULL,
			price DECIMAL(10, 2) NOT NULL,
			refund_quantity INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_order_id (order_id),
			INDEX idx_product_id (product_id)
		)`,
		`CREATE TABLE IF NOT EXISTS files (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			filename VARCHAR(255) NOT NULL,
			original_name VARCHAR(255) NOT NULL,
			file_path VARCHAR(500) NOT NULL,
			file_size BIGINT NOT NULL,
			mime_type VARCHAR(100) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_filename (filename)
		)`,
	}

	// 创建表
	for i, tableSQL := range tables {
		_, err = mallDB.Exec(tableSQL)
		if err != nil {
			log.Printf("❌ 创建表%d失败: %v", i+1, err)
		} else {
			fmt.Printf("✅ 表%d创建成功\n", i+1)
		}
	}

	// 插入测试数据
	testData := []string{
		`INSERT IGNORE INTO users (username, email, password, nickname, role, status) VALUES
		('admin', 'admin@example.com', '$2a$10$N.zmdr9k7uOCQb376NoUnuTJ8iKXISwKQv2BsLEKTRAZKoOqGDQwi', '管理员', 'admin', 'active'),
		('testuser', 'test@example.com', '$2a$10$N.zmdr9k7uOCQb376NoUnuTJ8iKXISwKQv2BsLEKTRAZKoOqGDQwi', '测试用户', 'user', 'active')`,
		`INSERT IGNORE INTO categories (id, name, description, status) VALUES
		(1, '电子产品', '各种电子设备和配件', 'active'),
		(2, '服装鞋帽', '时尚服装和鞋帽', 'active'),
		(3, '家居用品', '家庭生活用品', 'active')`,
		`INSERT IGNORE INTO products (id, name, description, price, stock, category_id, status) VALUES
		(1, '智能手机', '高性能智能手机，拍照清晰，运行流畅', 2999.00, 100, 1, 'active'),
		(2, '笔记本电脑', '轻薄便携笔记本电脑，适合办公和学习', 5999.00, 50, 1, 'active'),
		(3, '运动鞋', '舒适透气运动鞋，适合各种运动', 299.00, 200, 2, 'active'),
		(4, '咖啡杯', '精美陶瓷咖啡杯，保温效果好', 59.00, 500, 3, 'active')`,
	}

	// 插入测试数据
	for i, dataSQL := range testData {
		_, err = mallDB.Exec(dataSQL)
		if err != nil {
			log.Printf("❌ 插入测试数据%d失败: %v", i+1, err)
		} else {
			fmt.Printf("✅ 测试数据%d插入成功\n", i+1)
		}
	}

	fmt.Println("\n🎉 数据库初始化完成!")
	fmt.Println("数据库名称: MALL_GO")
	fmt.Println("表数量: 6个")
	fmt.Println("测试数据: 已插入")
}
