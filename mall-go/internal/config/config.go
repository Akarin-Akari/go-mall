package config

import (
	"flag"
	"fmt"
	"os"

	"mall-go/pkg/logger"

	"github.com/spf13/viper"
)

// Config 配置结构体
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Log       LogConfig       `mapstructure:"log"`
	CORS      CORSConfig      `mapstructure:"cors"`
	Upload    UploadConfig    `mapstructure:"upload"`
	Metrics   MetricsConfig   `mapstructure:"metrics"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
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

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
	Compress   bool   `mapstructure:"compress"`
}

// CORSConfig CORS配置
type CORSConfig struct {
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowMethods     []string `mapstructure:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"`
}

// UploadConfig 文件上传配置
type UploadConfig struct {
	MaxSize      string   `mapstructure:"max_size"`
	AllowedTypes []string `mapstructure:"allowed_types"`
	UploadPath   string   `mapstructure:"upload_path"`
}

// MetricsConfig 监控配置
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled           bool `mapstructure:"enabled"`
	RequestsPerSecond int  `mapstructure:"requests_per_second"`
	Burst             int  `mapstructure:"burst"`
}

var GlobalConfig Config

// Load 加载配置
func Load() {
	// 解析命令行参数
	configFile := flag.String("config", "", "配置文件路径")
	flag.Parse()

	// 设置默认值
	setDefaults()

	// 确定配置文件名
	configName := "config"
	env := os.Getenv("ENV")
	if env == "prod" || env == "production" {
		configName = "config-prod"
		logger.Info("使用生产环境配置")
	} else if env == "dev" || env == "development" {
		configName = "config"
		logger.Info("使用开发环境配置")
	} else {
		logger.Info("未指定环境变量 ENV，使用默认配置")
	}

	// 如果指定了配置文件路径，直接使用
	if *configFile != "" {
		viper.SetConfigFile(*configFile)
		logger.Info("使用指定的配置文件: " + *configFile)
	} else {
		viper.SetConfigName(configName)
		viper.SetConfigType("yaml")
		viper.AddConfigPath("./configs")
		viper.AddConfigPath(".")
	}

	// 支持从环境变量读取配置（优先级最高）
	viper.AutomaticEnv()
	viper.SetEnvPrefix("MALL") // 环境变量前缀，如 MALL_SERVER_PORT

	// 绑定特定的环境变量（用于敏感信息）
	bindEnvVars()

	// 尝试读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		logger.Warn("配置文件未找到或读取失败，将使用默认值和环境变量: " + err.Error())
	} else {
		logger.Info("配置文件加载成功: " + viper.ConfigFileUsed())
	}

	// 解析配置到结构体
	if err := viper.Unmarshal(&GlobalConfig); err != nil {
		logger.Error("解析配置失败: " + err.Error())
		panic(err)
	}

	// 从环境变量覆盖敏感信息
	overrideFromEnv()

	// 验证配置
	if err := validateConfig(); err != nil {
		logger.Error("配置验证失败: " + err.Error())
		panic(err)
	}

	logger.Info("配置加载完成")
}

// bindEnvVars 绑定环境变量
func bindEnvVars() {
	// 数据库配置
	viper.BindEnv("database.driver", "DB_DRIVER") // 数据库驱动（sqlite/mysql）
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("database.username", "DB_USERNAME")
	viper.BindEnv("database.host", "DB_HOST")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.dbname", "DB_NAME")

	// Redis 敏感信息
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
	viper.BindEnv("redis.host", "REDIS_HOST")
	viper.BindEnv("redis.port", "REDIS_PORT")

	// JWT 密钥
	viper.BindEnv("jwt.secret", "JWT_SECRET")
}

// overrideFromEnv 从环境变量覆盖敏感信息
func overrideFromEnv() {
	// 数据库密码
	if dbPassword := os.Getenv("DB_PASSWORD"); dbPassword != "" {
		GlobalConfig.Database.Password = dbPassword
	}

	// Redis 密码
	if redisPassword := os.Getenv("REDIS_PASSWORD"); redisPassword != "" {
		GlobalConfig.Redis.Password = redisPassword
	}

	// JWT 密钥
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		GlobalConfig.JWT.Secret = jwtSecret
	}
}

// validateConfig 验证配置
func validateConfig() error {
	// 生产环境必须设置 JWT 密钥
	if GlobalConfig.Server.Mode == "release" {
		if GlobalConfig.JWT.Secret == "" || GlobalConfig.JWT.Secret == "your-secret-key-change-in-production" {
			return fmt.Errorf("生产环境必须通过环境变量 JWT_SECRET 设置 JWT 密钥")
		}

		// 生产环境禁止使用通配符 CORS
		for _, origin := range GlobalConfig.CORS.AllowOrigins {
			if origin == "*" {
				return fmt.Errorf("生产环境禁止使用通配符 '*' 作为 CORS AllowOrigins")
			}
		}
	}

	return nil
}

// setDefaults 设置默认配置
func setDefaults() {
	// 服务器配置
	viper.SetDefault("server.port", 8081)
	viper.SetDefault("server.mode", "debug")

	// 数据库配置
	viper.SetDefault("database.driver", "mysql")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("database.charset", "utf8mb4")
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.max_open_conns", 50)
	viper.SetDefault("database.conn_max_lifetime", 1800)

	// Redis基础配置
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)
	// Redis连接池配置
	viper.SetDefault("redis.pool_size", 100)
	viper.SetDefault("redis.min_idle_conns", 10)
	viper.SetDefault("redis.max_retries", 3)
	// Redis超时配置
	viper.SetDefault("redis.dial_timeout", 5)
	viper.SetDefault("redis.read_timeout", 3)
	viper.SetDefault("redis.write_timeout", 3)
	viper.SetDefault("redis.idle_timeout", 300)
	viper.SetDefault("redis.max_conn_age", 3600)
	// Redis性能优化配置
	viper.SetDefault("redis.pool_timeout", 10)
	viper.SetDefault("redis.idle_check_frequency", 60)
	viper.SetDefault("redis.max_redirect", 8)
	viper.SetDefault("redis.read_only", false)
	viper.SetDefault("redis.route_by_latency", true)
	viper.SetDefault("redis.route_randomly", false)

	// JWT配置
	viper.SetDefault("jwt.secret", "your-secret-key")
	viper.SetDefault("jwt.expire", "24h")

	// 日志配置
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.output", "stdout")
	viper.SetDefault("log.max_size", 100)
	viper.SetDefault("log.max_age", 30)
	viper.SetDefault("log.max_backups", 10)
	viper.SetDefault("log.compress", true)

	// CORS配置
	viper.SetDefault("cors.allow_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	viper.SetDefault("cors.allow_headers", []string{"Authorization", "Content-Type", "X-Request-ID"})
	viper.SetDefault("cors.allow_credentials", true)
	viper.SetDefault("cors.max_age", 86400)

	// 文件上传配置
	viper.SetDefault("upload.max_size", "10MB")
	viper.SetDefault("upload.allowed_types", []string{"image/jpeg", "image/png", "image/gif"})
	viper.SetDefault("upload.upload_path", "./uploads")

	// 监控配置
	viper.SetDefault("metrics.enabled", true)
	viper.SetDefault("metrics.path", "/metrics")

	// 限流配置
	viper.SetDefault("rate_limit.enabled", false)
	viper.SetDefault("rate_limit.requests_per_second", 100)
	viper.SetDefault("rate_limit.burst", 200)
}
