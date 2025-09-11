package config

import (
	"mall-go/pkg/logger"

	"github.com/spf13/viper"
)

// Config 配置结构体
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver          string `mapstructure:"driver"`
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	Charset         string `mapstructure:"charset"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	// 连接池配置
	PoolSize     int `mapstructure:"pool_size"`      // 连接池大小
	MinIdleConns int `mapstructure:"min_idle_conns"` // 最小空闲连接数
	MaxRetries   int `mapstructure:"max_retries"`    // 最大重试次数
	// 超时配置
	DialTimeout  int `mapstructure:"dial_timeout"`  // 连接超时(秒)
	ReadTimeout  int `mapstructure:"read_timeout"`  // 读取超时(秒)
	WriteTimeout int `mapstructure:"write_timeout"` // 写入超时(秒)
	IdleTimeout  int `mapstructure:"idle_timeout"`  // 空闲连接超时(秒)
	MaxConnAge   int `mapstructure:"max_conn_age"`  // 连接最大存活时间(秒)
	// 性能优化配置
	PoolTimeout        int  `mapstructure:"pool_timeout"`         // 获取连接超时时间(秒)
	IdleCheckFrequency int  `mapstructure:"idle_check_frequency"` // 空闲连接检查频率(秒)
	MaxRedirect        int  `mapstructure:"max_redirect"`         // 集群模式最大重定向次数
	ReadOnly           bool `mapstructure:"read_only"`            // 是否只读模式
	RouteByLatency     bool `mapstructure:"route_by_latency"`     // 按延迟路由(集群模式)
	RouteRandomly      bool `mapstructure:"route_randomly"`       // 随机路由(集群模式)
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret string `mapstructure:"secret"`
	Expire string `mapstructure:"expire"`
}

var GlobalConfig Config

// Load 加载配置
func Load() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// 设置默认值
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		logger.Error("读取配置文件失败: " + err.Error())
		panic(err)
	}

	if err := viper.Unmarshal(&GlobalConfig); err != nil {
		logger.Error("解析配置文件失败: " + err.Error())
		panic(err)
	}

	logger.Info("配置加载成功")
}

// setDefaults 设置默认配置
func setDefaults() {
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("database.driver", "mysql")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("database.charset", "utf8mb4")
	// Redis基础配置
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)
	// Redis连接池配置
	viper.SetDefault("redis.pool_size", 200)
	viper.SetDefault("redis.min_idle_conns", 20)
	viper.SetDefault("redis.max_retries", 5)
	// Redis超时配置
	viper.SetDefault("redis.dial_timeout", 10)
	viper.SetDefault("redis.read_timeout", 5)
	viper.SetDefault("redis.write_timeout", 5)
	viper.SetDefault("redis.idle_timeout", 600)
	viper.SetDefault("redis.max_conn_age", 7200)
	// Redis性能优化配置
	viper.SetDefault("redis.pool_timeout", 30)
	viper.SetDefault("redis.idle_check_frequency", 60)
	viper.SetDefault("redis.max_redirect", 8)
	viper.SetDefault("redis.read_only", false)
	viper.SetDefault("redis.route_by_latency", true)
	viper.SetDefault("redis.route_randomly", false)
	viper.SetDefault("jwt.secret", "your-secret-key")
	viper.SetDefault("jwt.expire", "24h")
}
