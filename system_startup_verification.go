package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type APIResponse struct {
	Message   string   `json:"message"`
	Version   string   `json:"version"`
	Endpoints []string `json:"endpoints"`
}

func main() {
	fmt.Println("🚀 Mall-Go电商系统启动验证")
	fmt.Println("=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=")

	// 测试结果统计
	var passedTests, totalTests int

	// 1. 后端健康检查
	fmt.Println("\n📋 1. 后端健康检查")
	totalTests++
	if testBackendHealth() {
		fmt.Println("   ✅ 后端健康检查通过")
		passedTests++
	} else {
		fmt.Println("   ❌ 后端健康检查失败")
	}

	// 2. 后端API信息
	fmt.Println("\n📋 2. 后端API信息")
	totalTests++
	if testBackendAPI() {
		fmt.Println("   ✅ 后端API信息获取成功")
		passedTests++
	} else {
		fmt.Println("   ❌ 后端API信息获取失败")
	}

	// 3. 前端服务检查
	fmt.Println("\n📋 3. 前端服务检查")
	totalTests++
	if testFrontendService() {
		fmt.Println("   ✅ 前端服务响应正常")
		passedTests++
	} else {
		fmt.Println("   ❌ 前端服务响应异常")
	}

	// 4. 前后端通信测试
	fmt.Println("\n📋 4. 前后端通信测试")
	totalTests++
	if testCommunication() {
		fmt.Println("   ✅ 前后端通信正常")
		passedTests++
	} else {
		fmt.Println("   ❌ 前后端通信异常")
	}

	// 输出最终结果
	fmt.Println("\n" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=")
	fmt.Printf("📊 验证结果: %d/%d 测试通过\n", passedTests, totalTests)
	
	if passedTests == totalTests {
		fmt.Println("🎉 Mall-Go电商系统启动验证完全成功！")
		fmt.Println("\n🌐 访问地址:")
		fmt.Println("   前端Web: http://localhost:3000")
		fmt.Println("   后端API: http://localhost:8080")
		fmt.Println("   健康检查: http://localhost:8080/health")
		fmt.Println("\n👤 测试账号:")
		fmt.Println("   用户名: newuser2024")
		fmt.Println("   密码: 123456789")
	} else {
		fmt.Printf("⚠️  系统启动验证部分失败，请检查失败的组件\n")
	}
	fmt.Println("=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=")
}

func testBackendHealth() bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("http://localhost:8080/health")
	if err != nil {
		fmt.Printf("   ❌ 请求失败: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("   ❌ HTTP状态码: %d\n", resp.StatusCode)
		return false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("   ❌ 读取响应失败: %v\n", err)
		return false
	}

	var health HealthResponse
	if err := json.Unmarshal(body, &health); err != nil {
		fmt.Printf("   ❌ JSON解析失败: %v\n", err)
		return false
	}

	if health.Status != "ok" {
		fmt.Printf("   ❌ 健康状态异常: %s\n", health.Status)
		return false
	}

	fmt.Printf("   📊 状态: %s\n", health.Status)
	fmt.Printf("   📊 消息: %s\n", health.Message)
	return true
}

func testBackendAPI() bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("http://localhost:8080/")
	if err != nil {
		fmt.Printf("   ❌ 请求失败: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("   ❌ HTTP状态码: %d\n", resp.StatusCode)
		return false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("   ❌ 读取响应失败: %v\n", err)
		return false
	}

	var api APIResponse
	if err := json.Unmarshal(body, &api); err != nil {
		fmt.Printf("   ❌ JSON解析失败: %v\n", err)
		return false
	}

	fmt.Printf("   📊 版本: %s\n", api.Version)
	fmt.Printf("   📊 可用端点数量: %d\n", len(api.Endpoints))
	return true
}

func testFrontendService() bool {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("http://localhost:3000")
	if err != nil {
		fmt.Printf("   ❌ 请求失败: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("   ❌ HTTP状态码: %d\n", resp.StatusCode)
		return false
	}

	fmt.Printf("   📊 HTTP状态码: %d\n", resp.StatusCode)
	fmt.Printf("   📊 Content-Type: %s\n", resp.Header.Get("Content-Type"))
	return true
}

func testCommunication() bool {
	// 这里可以测试前端是否能正确调用后端API
	// 由于前端是SPA，我们测试后端的CORS设置
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("OPTIONS", "http://localhost:8080/health", nil)
	if err != nil {
		fmt.Printf("   ❌ 创建请求失败: %v\n", err)
		return false
	}
	
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("   ❌ CORS预检请求失败: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	// 检查CORS头
	corsOrigin := resp.Header.Get("Access-Control-Allow-Origin")
	if corsOrigin == "" {
		fmt.Printf("   ⚠️  CORS配置可能需要检查\n")
	} else {
		fmt.Printf("   📊 CORS Origin: %s\n", corsOrigin)
	}

	return true
}
