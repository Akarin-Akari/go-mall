package config

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/glebarez/sqlite"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestConfig 测试配置结构
type TestConfig struct {
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Redis    RedisConfig    `mapstructure:"redis"`
}

type DatabaseConfig struct {
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	Charset  string `mapstructure:"charset"`
}

type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	ExpireTime int    `mapstructure:"expire_time"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

var testConfig *TestConfig

// LoadTestConfig 加载测试配置
func LoadTestConfig() *TestConfig {
	if testConfig != nil {
		return testConfig
	}

	// 设置测试环境变量
	os.Setenv("GIN_MODE", "test")
	os.Setenv("LOG_LEVEL", "error")

	// 获取项目根目录
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(filename)))

	// 配置viper
	viper.SetConfigName("test")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(filepath.Join(projectRoot, "configs"))
	viper.AddConfigPath(".")

	// 设置默认值
	viper.SetDefault("database.driver", "sqlite")
	viper.SetDefault("database.database", ":memory:")
	viper.SetDefault("jwt.secret", "test-secret-key-for-mall-go")
	viper.SetDefault("jwt.expire_time", 3600)
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 1)

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		// 如果配置文件不存在，使用默认值
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic(err)
		}
	}

	testConfig = &TestConfig{}
	if err := viper.Unmarshal(testConfig); err != nil {
		panic(err)
	}

	return testConfig
}

// SetupTestDB 设置测试数据库
func SetupTestDB() *gorm.DB {
	config := LoadTestConfig()

	var db *gorm.DB
	var err error

	switch config.Database.Driver {
	case "sqlite":
		// 使用内存数据库进行测试
		db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
	default:
		// 默认使用SQLite内存数据库
		db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
	}

	if err != nil {
		panic("failed to connect test database: " + err.Error())
	}

	return db
}

// CleanupTestDB 清理测试数据库
func CleanupTestDB(db *gorm.DB) {
	if db != nil {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}

// GetTestConfig 获取测试配置
func GetTestConfig() *TestConfig {
	if testConfig == nil {
		return LoadTestConfig()
	}
	return testConfig
}
