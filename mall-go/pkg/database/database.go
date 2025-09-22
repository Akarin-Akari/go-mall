package database

import (
	"fmt"
	"log"
	"time"

	"mall-go/internal/config"
	"mall-go/internal/model"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Init 初始化数据库连接
func Init() *gorm.DB {
	var err error

	// 获取配置
	cfg := config.GlobalConfig

	// 配置GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// 根据驱动类型连接数据库
	if cfg.Database.Driver == "sqlite" {
		// SQLite连接
		DB, err = gorm.Open(sqlite.Open(cfg.Database.DBName), gormConfig)
		if err == nil {
			// 配置SQLite以支持并发访问
			configureSQLiteForConcurrency(DB)
		}
	} else if cfg.Database.Driver == "memory" {
		// 内存SQLite连接（用于测试）
		DB, err = gorm.Open(sqlite.Open(":memory:"), gormConfig)
		if err == nil {
			// 配置SQLite以支持并发访问
			configureSQLiteForConcurrency(DB)
		}
		log.Println("使用内存数据库模式（仅用于测试）")
	} else {
		// MySQL连接 - 先尝试连接指定数据库
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=30s&readTimeout=30s&writeTimeout=30s",
			cfg.Database.Username,
			cfg.Database.Password,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.DBName,
		)
		DB, err = gorm.Open(mysql.Open(dsn), gormConfig)

		// 如果数据库不存在，尝试创建数据库
		if err != nil && (err.Error() == "Error 1049: Unknown database '"+cfg.Database.DBName+"'" ||
			fmt.Sprintf("Error 1049 (42000): Unknown database '%s'", cfg.Database.DBName) == err.Error()) {
			log.Printf("数据库 %s 不存在，尝试创建...", cfg.Database.DBName)

			// 连接MySQL服务器（不指定数据库）
			rootDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8mb4&parseTime=True&loc=Local&timeout=30s&readTimeout=30s&writeTimeout=30s",
				cfg.Database.Username,
				cfg.Database.Password,
				cfg.Database.Host,
				cfg.Database.Port,
			)
			rootDB, rootErr := gorm.Open(mysql.Open(rootDSN), gormConfig)
			if rootErr != nil {
				log.Printf("连接MySQL服务器失败: %v", rootErr)
			} else {
				// 创建数据库
				createSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET UTF8MB4 COLLATE UTF8MB4_UNICODE_CI", cfg.Database.DBName)
				if createErr := rootDB.Exec(createSQL).Error; createErr != nil {
					log.Printf("创建数据库失败: %v", createErr)
				} else {
					log.Printf("数据库 %s 创建成功", cfg.Database.DBName)
					// 重新尝试连接
					DB, err = gorm.Open(mysql.Open(dsn), gormConfig)
				}
			}
		}
	}
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// 获取底层sql.DB对象进行连接池配置
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("获取数据库实例失败: %v", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Second)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
	}

	log.Println("数据库连接成功")

	// 自动迁移
	if err := autoMigrate(DB); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	return DB
}

// configureSQLiteForConcurrency 配置SQLite以支持并发访问
func configureSQLiteForConcurrency(db *gorm.DB) {
	// 获取底层sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("获取sql.DB失败: %v", err)
		return
	}

	// 配置SQLite连接池 - SQLite只支持单个写连接
	sqlDB.SetMaxOpenConns(1)    // SQLite只支持单个写连接
	sqlDB.SetMaxIdleConns(1)    // 保持连接池简单
	sqlDB.SetConnMaxLifetime(0) // 连接不过期

	// 启用WAL模式以提高并发性能
	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA synchronous=NORMAL")
	db.Exec("PRAGMA cache_size=1000")
	db.Exec("PRAGMA temp_store=memory")
	db.Exec("PRAGMA busy_timeout=30000") // 30秒超时，避免database locked错误

	log.Println("SQLite并发配置已启用: WAL模式, 30秒超时")
}

// InitDB 初始化数据库连接（别名，保持兼容性）
func InitDB() (*gorm.DB, error) {
	return Init(), nil
}

// autoMigrate 自动迁移数据库表结构
func autoMigrate(db *gorm.DB) error {
	log.Println("开始数据库迁移...")

	// 检查表是否存在，如果存在则跳过迁移
	if db.Migrator().HasTable(&model.User{}) {
		log.Println("数据库表已存在，跳过迁移")
		return nil
	}

	// 迁移所有模型
	err := db.AutoMigrate(
		&model.User{},
		&model.Product{},
		&model.ProductImage{}, // 添加商品图片表
		&model.Category{},
		&model.Order{},
		&model.OrderItem{},
		&model.OrderStatusLog{}, // 添加订单状态日志表
		&model.File{},
		// 支付相关模型
		&model.Payment{},
		&model.PaymentRefund{},
		&model.PaymentLog{},
		&model.PaymentConfig{},
		// 其他相关模型
		&model.Address{},
		&model.Cart{},
		&model.CartItem{},
	)

	if err != nil {
		return fmt.Errorf("数据库迁移失败: %v", err)
	}

	log.Println("数据库迁移完成")
	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	if DB == nil {
		log.Fatal("数据库未初始化，请先调用Init()")
	}
	return DB
}

// Close 关闭数据库连接
func Close() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

// Ping 测试数据库连接
func Ping() error {
	if DB == nil {
		return fmt.Errorf("数据库未初始化")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Ping()
}

// GetStats 获取数据库连接统计信息
func GetStats() map[string]interface{} {
	if DB == nil {
		return map[string]interface{}{
			"error": "数据库未初始化",
		}
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}

// Transaction 执行事务
func Transaction(fn func(*gorm.DB) error) error {
	if DB == nil {
		return fmt.Errorf("数据库未初始化")
	}

	return DB.Transaction(fn)
}

// IsHealthy 检查数据库健康状态
func IsHealthy() bool {
	if DB == nil {
		return false
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return false
	}

	err = sqlDB.Ping()
	return err == nil
}

// CreateTables 创建表（如果不存在）
func CreateTables() error {
	if DB == nil {
		return fmt.Errorf("数据库未初始化")
	}

	// 创建用户表
	if !DB.Migrator().HasTable(&model.User{}) {
		if err := DB.Migrator().CreateTable(&model.User{}); err != nil {
			return fmt.Errorf("创建用户表失败: %v", err)
		}
	}

	// 创建商品表
	if !DB.Migrator().HasTable(&model.Product{}) {
		if err := DB.Migrator().CreateTable(&model.Product{}); err != nil {
			return fmt.Errorf("创建商品表失败: %v", err)
		}
	}

	// 创建分类表
	if !DB.Migrator().HasTable(&model.Category{}) {
		if err := DB.Migrator().CreateTable(&model.Category{}); err != nil {
			return fmt.Errorf("创建分类表失败: %v", err)
		}
	}

	// 创建订单表
	if !DB.Migrator().HasTable(&model.Order{}) {
		if err := DB.Migrator().CreateTable(&model.Order{}); err != nil {
			return fmt.Errorf("创建订单表失败: %v", err)
		}
	}

	// 创建订单项表
	if !DB.Migrator().HasTable(&model.OrderItem{}) {
		if err := DB.Migrator().CreateTable(&model.OrderItem{}); err != nil {
			return fmt.Errorf("创建订单项表失败: %v", err)
		}
	}

	// 创建文件表
	if !DB.Migrator().HasTable(&model.File{}) {
		if err := DB.Migrator().CreateTable(&model.File{}); err != nil {
			return fmt.Errorf("创建文件表失败: %v", err)
		}
	}

	return nil
}

// DropTables 删除所有表（危险操作，仅用于开发环境）
func DropTables() error {
	if DB == nil {
		return fmt.Errorf("数据库未初始化")
	}

	tables := []interface{}{
		&model.OrderItem{},
		&model.Order{},
		&model.File{},
		&model.Product{},
		&model.Category{},
		&model.User{},
	}

	for _, table := range tables {
		if err := DB.Migrator().DropTable(table); err != nil {
			return fmt.Errorf("删除表失败: %v", err)
		}
	}

	return nil
}
