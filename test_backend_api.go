package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Token string `json:"token"`
		User  struct {
			ID       int    `json:"id"`
			Username string `json:"username"`
			Email    string `json:"email"`
			Role     string `json:"role"`
		} `json:"user"`
	} `json:"data"`
}

func waitForBackend() bool {
	fmt.Println("⏳ 等待后端服务启动...")
	client := &http.Client{Timeout: 2 * time.Second}
	
	for i := 0; i < 30; i++ { // 等待最多60秒
		resp, err := client.Get("http://localhost:8080/health")
		if err == nil {
			resp.Body.Close()
			fmt.Println("✅ 后端服务已启动!")
			return true
		}
		fmt.Printf(".")
		time.Sleep(2 * time.Second)
	}
	fmt.Println("\n❌ 后端服务启动超时")
	return false
}

func testLogin() bool {
	fmt.Println("\n🔐 测试登录API...")
	
	loginReq := LoginRequest{
		Username: "admin",
		Password: "admin123",
	}
	
	jsonData, _ := json.Marshal(loginReq)
	fmt.Printf("📤 发送登录请求: %s\n", string(jsonData))
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post("http://localhost:8080/api/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ 登录请求失败: %v\n", err)
		return false
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		return false
	}
	
	fmt.Printf("📥 响应状态码: %d\n", resp.StatusCode)
	fmt.Printf("📥 响应内容: %s\n", string(body))
	
	if resp.StatusCode != 200 {
		fmt.Printf("❌ HTTP状态码错误: %d\n", resp.StatusCode)
		return false
	}
	
	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		fmt.Printf("❌ 解析JSON失败: %v\n", err)
		return false
	}
	
	if loginResp.Code == 200 {
		fmt.Println("✅ 登录成功!")
		fmt.Printf("🎫 Token: %s\n", loginResp.Data.Token)
		fmt.Printf("👤 用户信息: ID=%d, Username=%s, Email=%s, Role=%s\n",
			loginResp.Data.User.ID,
			loginResp.Data.User.Username,
			loginResp.Data.User.Email,
			loginResp.Data.User.Role)
		return true
	} else {
		fmt.Printf("❌ 登录失败: %s (Code: %d)\n", loginResp.Msg, loginResp.Code)
		return false
	}
}

func testCORS() {
	fmt.Println("\n🌐 测试CORS配置...")
	
	client := &http.Client{Timeout: 5 * time.Second}
	req, _ := http.NewRequest("OPTIONS", "http://localhost:8080/api/auth/login", nil)
	req.Header.Set("Origin", "http://localhost:8081")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")
	
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ CORS预检请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	fmt.Printf("📥 CORS预检响应状态码: %d\n", resp.StatusCode)
	
	allowOrigin := resp.Header.Get("Access-Control-Allow-Origin")
	allowMethods := resp.Header.Get("Access-Control-Allow-Methods")
	allowHeaders := resp.Header.Get("Access-Control-Allow-Headers")
	
	fmt.Printf("🔧 CORS配置:\n")
	fmt.Printf("  Allow-Origin: %s\n", allowOrigin)
	fmt.Printf("  Allow-Methods: %s\n", allowMethods)
	fmt.Printf("  Allow-Headers: %s\n", allowHeaders)
	
	if allowOrigin == "*" || allowOrigin == "http://localhost:8081" {
		fmt.Println("✅ CORS配置正确")
	} else {
		fmt.Println("⚠️  CORS配置可能有问题")
	}
}

func main() {
	fmt.Println("🧪 Mall-Go 后端API测试")
	fmt.Println("========================")
	
	// 检查后端服务是否运行
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://localhost:8080/health")
	if err != nil {
		fmt.Printf("❌ 后端服务未运行: %v\n", err)
		fmt.Println("💡 请在另一个终端启动: cd mall-go && go run cmd/server/main.go")
		
		if !waitForBackend() {
			return
		}
	} else {
		fmt.Println("✅ 后端服务正在运行")
		resp.Body.Close()
	}
	
	// 执行API测试
	if testLogin() {
		testCORS()
		fmt.Println("\n🎉 后端API测试完成 - 可以进行前端联调!")
	} else {
		fmt.Println("\n❌ 登录测试失败 - 需要检查后端配置")
	}
	
	fmt.Println("\n========================")
}
