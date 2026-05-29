package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"

	"mall-go/internal/config"
	"mall-go/pkg/logger"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	// 定义命令行参数
	command := flag.String("cmd", "up", "迁移命令: up, down, version, force")
	steps := flag.Int("steps", 0, "迁移步数（用于 down 命令）")
	forceVersion := flag.Int("force", -1, "强制设置版本号")
	flag.Parse()

	// 初始化日志
	logger.Init()

	// 加载配置
	config.Load()
	cfg := config.GlobalConfig

	// 构建数据库连接字符串
	var dsn string
	if cfg.Database.Driver == "mysql" {
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local&multiStatements=true",
			cfg.Database.Username,
			cfg.Database.Password,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.DBName,
			cfg.Database.Charset,
		)
	} else {
		log.Fatal("数据库迁移工具仅支持 MySQL")
	}

	// 打开数据库连接
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
	}

	// 创建迁移驱动
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		log.Fatalf("创建迁移驱动失败: %v", err)
	}

	// 创建迁移实例
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"mysql",
		driver,
	)
	if err != nil {
		log.Fatalf("创建迁移实例失败: %v", err)
	}

	// 执行迁移命令
	switch *command {
	case "up":
		if err := m.Up(); err != nil {
			if err == migrate.ErrNoChange {
				log.Println("✅ 数据库已是最新版本，无需迁移")
			} else {
				log.Fatalf("❌ 执行迁移失败: %v", err)
			}
		} else {
			version, dirty, _ := m.Version()
			log.Printf("✅ 迁移成功完成，当前版本: %d (dirty: %v)", version, dirty)
		}

	case "down":
		if *steps > 0 {
			if err := m.Steps(-*steps); err != nil {
				log.Fatalf("❌ 回滚失败: %v", err)
			}
			version, dirty, _ := m.Version()
			log.Printf("✅ 回滚成功，当前版本: %d (dirty: %v)", version, dirty)
		} else {
			if err := m.Down(); err != nil {
				if err == migrate.ErrNoChange {
					log.Println("✅ 已经是最早版本，无法继续回滚")
				} else {
					log.Fatalf("❌ 回滚失败: %v", err)
				}
			} else {
				log.Println("✅ 已回滚所有迁移")
			}
		}

	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("❌ 获取版本失败: %v", err)
		}
		log.Printf("📌 当前迁移版本: %d (dirty: %v)", version, dirty)

	case "force":
		if *forceVersion < 0 {
			log.Fatal("❌ 请使用 -force 参数指定版本号")
		}
		if err := m.Force(*forceVersion); err != nil {
			log.Fatalf("❌ 强制设置版本失败: %v", err)
		}
		log.Printf("✅ 已强制设置版本为: %d", *forceVersion)

	default:
		log.Fatalf("❌ 未知命令: %s\n可用命令: up, down, version, force", *command)
	}
}
