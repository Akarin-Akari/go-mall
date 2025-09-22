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
			ID       uint   `json:"id"`
			Username string `json:"username"`
			Email    string `json:"email"`
		} `json:"user"`
	} `json:"data"`
}

func main() {
	fmt.Println("🧪 测试Mall-Go API与Redis集成")
	
	// 等待服务器启动
	fmt.Println("⏳ 等待服务器启动...")
	time.Sleep(3 * time.Second)
	
	// 测试健康检查
	fmt.Println("\n🔍 测试健康检查...")
	resp, err := http.Get("http://localhost:8081/health")
	if err != nil {
		fmt.Printf("❌ 健康检查失败: %v\n", err)
		fmt.Println("💡 请确保后端服务正在运行在端口8081")
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 200 {
		fmt.Println("✅ 健康检查成功")
	} else {
		fmt.Printf("❌ 健康检查失败，状态码: %d\n", resp.StatusCode)
		return
	}
	
	// 测试登录API
	fmt.Println("\n🔐 测试登录API...")
	loginReq := LoginRequest{
		Username: "admin",
		Password: "admin123",
	}
	
	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		fmt.Printf("❌ JSON序列化失败: %v\n", err)
		return
	}
	
	resp, err = http.Post("http://localhost:8081/api/v1/users/login", "application/json", bytes.NewBuffer(jsonData))
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
	
	fmt.Printf("📊 登录响应状态码: %d\n", resp.StatusCode)
	fmt.Printf("📊 登录响应内容: %s\n", string(body))
	
	if resp.StatusCode == 200 {
		var loginResp LoginResponse
		err = json.Unmarshal(body, &loginResp)
		if err != nil {
			fmt.Printf("❌ 解析登录响应失败: %v\n", err)
			return
		}
		
		if loginResp.Code == 200 {
			fmt.Println("✅ 登录成功")
			fmt.Printf("🎫 JWT Token: %s...\n", loginResp.Data.Token[:50])
			fmt.Printf("👤 用户信息: %s (%s)\n", loginResp.Data.User.Username, loginResp.Data.User.Email)
			
			// 测试需要认证的API
			fmt.Println("\n🔒 测试需要认证的API...")
			testAuthenticatedAPI(loginResp.Data.Token)
		} else {
			fmt.Printf("❌ 登录失败: %s\n", loginResp.Msg)
		}
	} else {
		fmt.Printf("❌ 登录失败，状态码: %d\n", resp.StatusCode)
	}
	
	fmt.Println("\n🎉 API测试完成！")
}

func testAuthenticatedAPI(token string) {
	// 测试获取用户信息API
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://localhost:8081/api/v1/users/profile", nil)
	if err != nil {
		fmt.Printf("❌ 创建请求失败: %v\n", err)
		return
	}
	
	req.Header.Set("Authorization", "Bearer "+token)
	
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 获取用户信息失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		return
	}
	
	fmt.Printf("📊 用户信息响应状态码: %d\n", resp.StatusCode)
	fmt.Printf("📊 用户信息响应内容: %s\n", string(body))
	
	if resp.StatusCode == 200 {
		fmt.Println("✅ 认证API测试成功")
	} else {
		fmt.Printf("❌ 认证API测试失败，状态码: %d\n", resp.StatusCode)
	}
}
