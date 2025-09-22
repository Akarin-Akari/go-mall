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

func main() {
	fmt.Println("🧪 测试Mall-Go登录API")
	
	// 测试服务器是否运行
	fmt.Println("检查服务器状态...")
	resp, err := http.Get("http://localhost:8080/health")
	if err != nil {
		fmt.Printf("❌ 服务器未运行: %v\n", err)
		fmt.Println("请先启动服务器: go run cmd/server/main.go")
		return
	}
	resp.Body.Close()
	fmt.Println("✅ 服务器正在运行")
	
	// 测试登录
	loginReq := LoginRequest{
		Username: "admin",
		Password: "admin123",
	}
	
	jsonData, _ := json.Marshal(loginReq)
	
	fmt.Printf("发送登录请求: %s\n", string(jsonData))
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err = client.Post("http://localhost:8080/api/auth/login", "application/json", bytes.NewBuffer(jsonData))
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
	
	fmt.Printf("响应状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应内容: %s\n", string(body))
	
	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		fmt.Printf("❌ 解析响应失败: %v\n", err)
		return
	}
	
	if loginResp.Code == 200 {
		fmt.Println("✅ 登录成功!")
		fmt.Printf("Token: %s\n", loginResp.Data.Token)
		fmt.Printf("用户信息: ID=%d, Username=%s, Email=%s, Role=%s\n",
			loginResp.Data.User.ID,
			loginResp.Data.User.Username,
			loginResp.Data.User.Email,
			loginResp.Data.User.Role)
	} else {
		fmt.Printf("❌ 登录失败: %s\n", loginResp.Msg)
	}
}
