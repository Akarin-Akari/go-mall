package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	fmt.Println("🔍 Mall-Go 数据库连接诊断工具")
	fmt.Println("================================")

	// 测试配置
	configs := []struct {
		name   string
		dbname string
		dsn    string
	}{
		{
			name:   "配置文件中的数据库名 (gomall)",
			dbname: "gomall",
			dsn:    "root:123456@tcp(localhost:3306)/gomall?charset=utf8mb4&parseTime=True&loc=Local",
		},
		{
			name:   "初始化脚本中的数据库名 (MALL_GO)",
			dbname: "MALL_GO",
			dsn:    "root:123456@tcp(localhost:3306)/MALL_GO?charset=utf8mb4&parseTime=True&loc=Local",
		},
		{
			name:   "不指定数据库名 (测试MySQL连接)",
			dbname: "",
			dsn:    "root:123456@tcp(localhost:3306)/?charset=utf8mb4&parseTime=True&loc=Local",
		},
	}

	// 1. 测试MySQL服务连接
	fmt.Println("\n📡 测试MySQL服务连接...")
	testMySQLConnection()

	// 2. 测试各种数据库配置
	fmt.Println("\n🗄️ 测试数据库配置...")
	for _, config := range configs {
		fmt.Printf("\n测试: %s\n", config.name)
		testDatabaseConnection(config.dsn, config.dbname)
	}

	// 3. 列出所有数据库
	fmt.Println("\n📋 列出所有数据库...")
	listDatabases()

	// 4. 检查表结构
	fmt.Println("\n🏗️ 检查表结构...")
	checkTables("gomall")
	checkTables("MALL_GO")

	fmt.Println("\n✅ 数据库诊断完成!")
}

func testMySQLConnection() {
	dsn := "root:123456@tcp(localhost:3306)/"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("❌ MySQL连接失败: %v\n", err)
		return
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		fmt.Printf("❌ MySQL服务不可用: %v\n", err)
		return
	}

	fmt.Println("✅ MySQL服务连接成功")
}

func testDatabaseConnection(dsn, dbname string) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("❌ 连接失败: %v\n", err)
		return
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		fmt.Printf("❌ 数据库不可用: %v\n", err)
		return
	}

	if dbname != "" {
		fmt.Printf("✅ 数据库 '%s' 连接成功\n", dbname)
	} else {
		fmt.Println("✅ MySQL服务连接成功")
	}
}

func listDatabases() {
	dsn := "root:123456@tcp(localhost:3306)/"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("❌ 连接失败: %v\n", err)
		return
	}
	defer db.Close()

	rows, err := db.Query("SHOW DATABASES")
	if err != nil {
		fmt.Printf("❌ 查询数据库列表失败: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Println("📋 现有数据库:")
	for rows.Next() {
		var dbname string
		if err := rows.Scan(&dbname); err != nil {
			continue
		}
		fmt.Printf("  - %s\n", dbname)
	}
}

func checkTables(dbname string) {
	dsn := fmt.Sprintf("root:123456@tcp(localhost:3306)/%s", dbname)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("❌ 连接数据库 '%s' 失败: %v\n", dbname, err)
		return
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		fmt.Printf("❌ 数据库 '%s' 不存在或不可用\n", dbname)
		return
	}

	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		fmt.Printf("❌ 查询表列表失败: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Printf("🏗️ 数据库 '%s' 中的表:\n", dbname)
	tableCount := 0
	for rows.Next() {
		var tablename string
		if err := rows.Scan(&tablename); err != nil {
			continue
		}
		fmt.Printf("  - %s\n", tablename)
		tableCount++
	}

	if tableCount == 0 {
		fmt.Printf("  (无表)\n")
	}
}
