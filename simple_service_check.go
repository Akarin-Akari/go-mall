package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	fmt.Println("🧪 Mall-Go 服务检查")
	fmt.Println("====================")
	
	// 检查后端服务
	fmt.Println("📡 检查后端服务 (http://localhost:8080)...")
	client := &http.Client{Timeout: 3 * time.Second}
	
	resp, err := client.Get("http://localhost:8080/health")
	if err != nil {
		fmt.Printf("❌ 后端服务离线: %v\n", err)
		fmt.Println("💡 请启动后端服务: cd mall-go && go run cmd/server/main.go")
	} else {
		fmt.Printf("✅ 后端服务在线 (状态码: %d)\n", resp.StatusCode)
		resp.Body.Close()
	}
	
	// 检查前端服务
	fmt.Println("\n🌐 检查前端服务...")
	ports := []string{"3000", "8081"}
	
	for _, port := range ports {
		url := "http://localhost:" + port
		resp, err := client.Get(url)
		if err != nil {
			fmt.Printf("❌ 端口 %s 服务离线: %v\n", port, err)
		} else {
			fmt.Printf("✅ 端口 %s 服务在线 (状态码: %d)\n", port, resp.StatusCode)
			resp.Body.Close()
		}
	}
	
	fmt.Println("\n====================")
	fmt.Println("检查完成")
}
