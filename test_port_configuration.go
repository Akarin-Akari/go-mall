package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// 测试端口配置的脚本
func main() {
	fmt.Println("🔧 Mall-Go端口配置验证工具")
	fmt.Println("=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=")
	
	// 测试后端服务（应该在8081端口）
	fmt.Println("\n🚀 测试后端服务配置...")
	testBackendService()
	
	// 测试前端服务（应该在3000端口）
	fmt.Println("\n🌐 测试前端服务配置...")
	testFrontendService()
	
	fmt.Println("\n✅ 端口配置验证完成!")
}

func testBackendService() {
	// 测试8081端口（新配置）
	fmt.Println("  📡 测试后端服务 - 端口8081...")
	
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	
	resp, err := client.Get("http://localhost:8081/health")
	if err != nil {
		fmt.Printf("  ❌ 后端服务8081端口连接失败: %v\n", err)
		
		// 尝试8080端口（旧配置）
		fmt.Println("  🔄 尝试旧端口8080...")
		resp2, err2 := client.Get("http://localhost:8080/health")
		if err2 != nil {
			fmt.Printf("  ❌ 后端服务8080端口也连接失败: %v\n", err2)
			fmt.Println("  💡 建议: 请启动后端服务")
		} else {
			fmt.Println("  ⚠️  后端服务仍在8080端口运行，需要重启到8081端口")
			resp2.Body.Close()
		}
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 200 {
		body, _ := io.ReadAll(resp.Body)
		var healthResp map[string]interface{}
		if err := json.Unmarshal(body, &healthResp); err == nil {
			fmt.Printf("  ✅ 后端服务8081端口正常运行\n")
			fmt.Printf("  📊 响应状态: %s\n", healthResp["status"])
			fmt.Printf("  📝 响应消息: %s\n", healthResp["message"])
		} else {
			fmt.Printf("  ✅ 后端服务8081端口响应正常 (状态码: %d)\n", resp.StatusCode)
		}
	} else {
		fmt.Printf("  ⚠️  后端服务8081端口响应异常 (状态码: %d)\n", resp.StatusCode)
	}
}

func testFrontendService() {
	// 测试3000端口（新配置）
	fmt.Println("  🌐 测试前端服务 - 端口3000...")
	
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	
	resp, err := client.Get("http://localhost:3000")
	if err != nil {
		fmt.Printf("  ❌ 前端服务3000端口连接失败: %v\n", err)
		
		// 尝试3001端口（可能的旧配置）
		fmt.Println("  🔄 尝试端口3001...")
		resp2, err2 := client.Get("http://localhost:3001")
		if err2 != nil {
			fmt.Printf("  ❌ 前端服务3001端口也连接失败: %v\n", err2)
			fmt.Println("  💡 建议: 请启动前端服务")
		} else {
			fmt.Println("  ⚠️  前端服务仍在3001端口运行，需要重启到3000端口")
			resp2.Body.Close()
		}
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 200 {
		fmt.Printf("  ✅ 前端服务3000端口正常运行 (状态码: %d)\n", resp.StatusCode)
	} else {
		fmt.Printf("  ⚠️  前端服务3000端口响应异常 (状态码: %d)\n", resp.StatusCode)
	}
}
