package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	fmt.Println("🚀 开始导入数据库表结构...")

	// 连接到MySQL服务器（使用root用户）
	rootDSN := "root:123456@tcp(localhost:3306)/?charset=utf8mb4&parseTime=True&loc=Local&timeout=30s"

	db, err := sql.Open("mysql", rootDSN)
	if err != nil {
		log.Fatalf("连接MySQL服务器失败: %v", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatalf("MySQL服务器连接测试失败: %v", err)
	}

	fmt.Println("✅ MySQL服务器连接成功")

	// 读取并执行SQL文件
	if err := executeSQLFile(db, "schema.sql"); err != nil {
		log.Fatalf("执行SQL文件失败: %v", err)
	}

	fmt.Println("🎉 数据库表结构导入完成!")
}

func executeSQLFile(db *sql.DB, filename string) error {
	// 读取SQL文件
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("打开SQL文件失败: %v", err)
	}
	defer file.Close()

	fmt.Printf("📋 正在执行SQL文件: %s\n", filename)

	scanner := bufio.NewScanner(file)
	var sqlBuilder strings.Builder
	lineCount := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineCount++

		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}

		// 累积SQL语句
		sqlBuilder.WriteString(line)
		sqlBuilder.WriteString(" ")

		// 如果行以分号结尾，执行SQL语句
		if strings.HasSuffix(line, ";") {
			sqlStatement := strings.TrimSpace(sqlBuilder.String())
			if sqlStatement != "" {
				// 执行SQL语句
				if err := executeSQLStatement(db, sqlStatement); err != nil {
					log.Printf("执行SQL语句失败 (行 %d): %v", lineCount, err)
					log.Printf("SQL语句: %s", sqlStatement)
					// 继续执行其他语句，不中断整个过程
				}
			}
			sqlBuilder.Reset()
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取SQL文件失败: %v", err)
	}

	return nil
}

func executeSQLStatement(db *sql.DB, sqlStatement string) error {
	// 跳过SET语句和其他配置语句
	upperSQL := strings.ToUpper(strings.TrimSpace(sqlStatement))
	if strings.HasPrefix(upperSQL, "SET ") {
		fmt.Printf("⚙️  执行配置: %s\n", sqlStatement[:50])
		_, err := db.Exec(sqlStatement)
		return err
	}

	// 处理DROP语句
	if strings.HasPrefix(upperSQL, "DROP ") {
		fmt.Printf("🗑️  删除表: %s\n", sqlStatement[:50])
		_, err := db.Exec(sqlStatement)
		if err != nil {
			// DROP语句失败不是致命错误，可能表不存在
			fmt.Printf("⚠️  删除表警告: %v\n", err)
			return nil
		}
		return nil
	}

	// 处理CREATE语句
	if strings.HasPrefix(upperSQL, "CREATE ") {
		if strings.Contains(upperSQL, "CREATE DATABASE") {
			fmt.Printf("🏗️  创建数据库: %s\n", sqlStatement[:50])
		} else {
			fmt.Printf("📋 创建表: %s\n", sqlStatement[:50])
		}
		_, err := db.Exec(sqlStatement)
		if err != nil {
			return fmt.Errorf("创建失败: %v", err)
		}
		return nil
	}

	// 处理INSERT语句
	if strings.HasPrefix(upperSQL, "INSERT ") {
		fmt.Printf("📝 插入数据: %s\n", sqlStatement[:50])
		_, err := db.Exec(sqlStatement)
		if err != nil {
			return fmt.Errorf("插入数据失败: %v", err)
		}
		return nil
	}

	// 其他SQL语句
	fmt.Printf("⚡ 执行SQL: %s\n", sqlStatement[:50])
	_, err := db.Exec(sqlStatement)
	return err
}
