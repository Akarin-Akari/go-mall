package main

import (
	"encoding/json"
	"fmt"
	"mall-go/internal/config"
	"mall-go/pkg/logger"
)

func main() {
	// 初始化日志
	logger.Init()

	// 加载配置
	config.Load()

	// 打印配置信息（隐藏敏感信息）
	cfg := config.GlobalConfig

	fmt.Println("\n========== 配置信息 ==========")
	fmt.Printf("环境模式: %s\n", cfg.Server.Mode)
	fmt.Printf("服务端口: %d\n", cfg.Server.Port)
	fmt.Println()

	fmt.Println("数据库配置:")
	fmt.Printf("  驱动: %s\n", cfg.Database.Driver)
	fmt.Printf("  主机: %s\n", cfg.Database.Host)
	fmt.Printf("  端口: %d\n", cfg.Database.Port)
	fmt.Printf("  数据库名: %s\n", cfg.Database.DBName)
	fmt.Printf("  用户名: %s\n", cfg.Database.Username)
	fmt.Printf("  密码: %s\n", maskPassword(cfg.Database.Password))
	fmt.Printf("  最大连接数: %d\n", cfg.Database.MaxOpenConns)
	fmt.Printf("  最大空闲连接数: %d\n", cfg.Database.MaxIdleConns)
	fmt.Println()

	fmt.Println("Redis配置:")
	fmt.Printf("  主机: %s\n", cfg.Redis.Host)
	fmt.Printf("  端口: %d\n", cfg.Redis.Port)
	fmt.Printf("  密码: %s\n", maskPassword(cfg.Redis.Password))
	fmt.Printf("  数据库: %d\n", cfg.Redis.DB)
	fmt.Printf("  连接池大小: %d\n", cfg.Redis.PoolSize)
	fmt.Println()

	fmt.Println("JWT配置:")
	fmt.Printf("  密钥: %s\n", maskPassword(cfg.JWT.Secret))
	fmt.Printf("  过期时间: %s\n", cfg.JWT.Expire)
	fmt.Println()

	fmt.Println("日志配置:")
	fmt.Printf("  级别: %s\n", cfg.Log.Level)
	fmt.Printf("  格式: %s\n", cfg.Log.Format)
	fmt.Printf("  输出: %s\n", cfg.Log.Output)
	fmt.Println()

	fmt.Println("CORS配置:")
	fmt.Printf("  允许的源: %v\n", cfg.CORS.AllowOrigins)
	fmt.Printf("  允许的方法: %v\n", cfg.CORS.AllowMethods)
	fmt.Printf("  允许凭证: %v\n", cfg.CORS.AllowCredentials)
	fmt.Println()

	fmt.Println("监控配置:")
	fmt.Printf("  启用: %v\n", cfg.Metrics.Enabled)
	fmt.Printf("  路径: %s\n", cfg.Metrics.Path)
	fmt.Println()

	// 输出完整配置（JSON格式，用于调试）
	fmt.Println("========== 完整配置 (JSON) ==========")
	configJSON, _ := json.MarshalIndent(sanitizeConfig(cfg), "", "  ")
	fmt.Println(string(configJSON))
}

// maskPassword 隐藏密码
func maskPassword(password string) string {
	if password == "" {
		return "<未设置>"
	}
	if len(password) <= 4 {
		return "****"
	}
	return password[:2] + "****" + password[len(password)-2:]
}

// sanitizeConfig 清理敏感信息
func sanitizeConfig(cfg config.Config) map[string]interface{} {
	return map[string]interface{}{
		"server": map[string]interface{}{
			"port": cfg.Server.Port,
			"mode": cfg.Server.Mode,
		},
		"database": map[string]interface{}{
			"driver":            cfg.Database.Driver,
			"host":              cfg.Database.Host,
			"port":              cfg.Database.Port,
			"dbname":            cfg.Database.DBName,
			"username":          cfg.Database.Username,
			"password":          maskPassword(cfg.Database.Password),
			"max_open_conns":    cfg.Database.MaxOpenConns,
			"max_idle_conns":    cfg.Database.MaxIdleConns,
			"conn_max_lifetime": cfg.Database.ConnMaxLifetime,
		},
		"redis": map[string]interface{}{
			"host":      cfg.Redis.Host,
			"port":      cfg.Redis.Port,
			"password":  maskPassword(cfg.Redis.Password),
			"db":        cfg.Redis.DB,
			"pool_size": cfg.Redis.PoolSize,
		},
		"jwt": map[string]interface{}{
			"secret": maskPassword(cfg.JWT.Secret),
			"expire": cfg.JWT.Expire,
		},
		"log": map[string]interface{}{
			"level":  cfg.Log.Level,
			"format": cfg.Log.Format,
			"output": cfg.Log.Output,
		},
		"cors": map[string]interface{}{
			"allow_origins":     cfg.CORS.AllowOrigins,
			"allow_methods":     cfg.CORS.AllowMethods,
			"allow_credentials": cfg.CORS.AllowCredentials,
		},
		"metrics": map[string]interface{}{
			"enabled": cfg.Metrics.Enabled,
			"path":    cfg.Metrics.Path,
		},
	}
}

