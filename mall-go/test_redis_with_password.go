package main

import (
	"fmt"
	"mall-go/internal/config"
	"mall-go/pkg/cache"
	"mall-go/pkg/logger"
)

func main() {
	fmt.Println("🔧 测试Redis配置（密码：123456）")

	// 初始化日志
	logger.Init()

	// 加载配置
	config.Load()

	// 显示当前Redis配置
	fmt.Printf("📋 当前Redis配置:\n")
	fmt.Printf("  Host: %s\n", config.GlobalConfig.Redis.Host)
	fmt.Printf("  Port: %d\n", config.GlobalConfig.Redis.Port)
	fmt.Printf("  Password: %s\n", config.GlobalConfig.Redis.Password)
	fmt.Printf("  DB: %d\n", config.GlobalConfig.Redis.DB)

	// 测试Redis连接
	fmt.Println("\n🔗 测试Redis连接...")
	redisClient, err := cache.NewRedisClient(config.GlobalConfig.Redis)
	if err != nil {
		fmt.Printf("❌ Redis连接失败: %v\n", err)
		fmt.Println("\n💡 请确保:")
		fmt.Println("  1. Redis服务器正在运行")
		fmt.Println("  2. Redis配置了密码 '123456'")
		fmt.Println("  3. 配置文件中的Redis设置正确")
		return
	}

	fmt.Println("✅ Redis连接成功!")

	// 测试基本操作
	fmt.Println("\n🧪 测试基本Redis操作...")

	// 测试SET操作
	client := redisClient.GetClient()
	ctx := redisClient.GetContext()

	err = client.Set(ctx, "test_key", "test_value", 0).Err()
	if err != nil {
		fmt.Printf("❌ SET操作失败: %v\n", err)
		return
	}
	fmt.Println("✅ SET操作成功")

	// 测试GET操作
	val, err := client.Get(ctx, "test_key").Result()
	if err != nil {
		fmt.Printf("❌ GET操作失败: %v\n", err)
		return
	}
	fmt.Printf("✅ GET操作成功，值: %s\n", val)

	// 测试DEL操作
	err = client.Del(ctx, "test_key").Err()
	if err != nil {
		fmt.Printf("❌ DEL操作失败: %v\n", err)
		return
	}
	fmt.Println("✅ DEL操作成功")

	// 关闭连接
	redisClient.Close()

	fmt.Println("\n🎉 Redis配置测试完成！")
	fmt.Println("现在可以重启后端服务以应用新的Redis配置。")
}
