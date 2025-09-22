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

type ServiceStatus struct {
	Name    string
	URL     string
	Status  string
	Details string
}

func checkService(name, url string) ServiceStatus {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return ServiceStatus{
			Name:    name,
			URL:     url,
			Status:  "❌ 离线",
			Details: err.Error(),
		}
	}
	defer resp.Body.Close()
	
	return ServiceStatus{
		Name:    name,
		URL:     url,
		Status:  "✅ 在线",
		Details: fmt.Sprintf("状态码: %d", resp.StatusCode),
	}
}

func testLogin(baseURL string) {
	fmt.Println("\n🔐 测试登录功能:")
	fmt.Println("=====================================")
	
	loginReq := LoginRequest{
		Username: "admin",
		Password: "admin123",
	}
	
	jsonData, _ := json.Marshal(loginReq)
	fmt.Printf("📤 发送登录请求: %s\n", string(jsonData))
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(baseURL+"/api/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ 登录请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		return
	}
	
	fmt.Printf("📥 响应状态码: %d\n", resp.StatusCode)
	fmt.Printf("📥 响应内容: %s\n", string(body))
	
	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		fmt.Printf("❌ 解析响应失败: %v\n", err)
		return
	}
	
	if loginResp.Code == 200 {
		fmt.Println("✅ 登录成功!")
		fmt.Printf("🎫 Token: %s\n", loginResp.Data.Token)
		fmt.Printf("👤 用户信息: ID=%d, Username=%s, Email=%s, Role=%s\n",
			loginResp.Data.User.ID,
			loginResp.Data.User.Username,
			loginResp.Data.User.Email,
			loginResp.Data.User.Role)
		
		// 测试带Token的API请求
		testAuthenticatedAPI(baseURL, loginResp.Data.Token)
	} else {
		fmt.Printf("❌ 登录失败: %s\n", loginResp.Msg)
	}
}

func testAuthenticatedAPI(baseURL, token string) {
	fmt.Println("\n🔒 测试认证API:")
	fmt.Println("=====================================")
	
	client := &http.Client{Timeout: 5 * time.Second}
	req, _ := http.NewRequest("GET", baseURL+"/api/user/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 认证API请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("📥 认证API响应 (状态码: %d): %s\n", resp.StatusCode, string(body))
}

func testCORS(baseURL string) {
	fmt.Println("\n🌐 测试CORS配置:")
	fmt.Println("=====================================")
	
	client := &http.Client{Timeout: 5 * time.Second}
	req, _ := http.NewRequest("OPTIONS", baseURL+"/api/auth/login", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")
	
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ CORS预检请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	fmt.Printf("📥 CORS预检响应状态码: %d\n", resp.StatusCode)
	fmt.Printf("🔧 CORS Headers:\n")
	for key, values := range resp.Header {
		if key == "Access-Control-Allow-Origin" || 
		   key == "Access-Control-Allow-Methods" || 
		   key == "Access-Control-Allow-Headers" {
			fmt.Printf("  %s: %v\n", key, values)
		}
	}
}

func main() {
	fmt.Println("🧪 Mall-Go 前后端联调测试")
	fmt.Println("=============================================")
	
	// 检查服务状态
	services := []ServiceStatus{
		checkService("后端服务", "http://localhost:8080/health"),
		checkService("前端服务(3000)", "http://localhost:3000"),
		checkService("前端服务(8081)", "http://localhost:8081"),
	}
	
	fmt.Println("📊 服务状态检查:")
	for _, service := range services {
		fmt.Printf("  %s %s - %s (%s)\n", service.Status, service.Name, service.URL, service.Details)
	}
	
	// 如果后端服务在线，进行API测试
	backendURL := "http://localhost:8080"
	if services[0].Status == "✅ 在线" {
		testLogin(backendURL)
		testCORS(backendURL)
	} else {
		fmt.Println("\n❌ 后端服务离线，无法进行API测试")
		fmt.Println("💡 请先启动后端服务: cd mall-go && go run cmd/server/main.go")
	}
	
	fmt.Println("\n=============================================")
	fmt.Println("🎯 联调测试完成")
}
