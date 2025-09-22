package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// 打开数据库连接
	db, err := sql.Open("sqlite3", "./mall-go/mall_go.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 获取所有表名
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table'")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			log.Fatal(err)
		}
		tables = append(tables, tableName)
	}

	// 检查每个表的字段
	for _, table := range tables {
		fmt.Printf("\n=== Table: %s ===\n", table)
		
		pragma := fmt.Sprintf("PRAGMA table_info(%s)", table)
		infoRows, err := db.Query(pragma)
		if err != nil {
			fmt.Printf("Error getting info for table %s: %v\n", table, err)
			continue
		}
		
		for infoRows.Next() {
			var cid int
			var name, dataType string
			var notNull, pk int
			var defaultValue sql.NullString
			
			if err := infoRows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
				fmt.Printf("Error scanning row: %v\n", err)
				continue
			}
			
			// 只显示包含sku的字段
			if name == "sku_id" || name == "sk_uid" || name == "SKUID" {
				fmt.Printf("  Field: %s, Type: %s, NotNull: %d, PK: %d\n", name, dataType, notNull, pk)
			}
		}
		infoRows.Close()
	}
}