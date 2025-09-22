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
			Role     string `json:"role"`
		} `json:"user"`
	} `json:"data"`
}

type CreateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	CategoryID  uint    `json:"category_id"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	Status      string  `json:"status"`
}

func main() {
	fmt.Println("🧪 测试商品管理模块API")
	
	baseURL := "http://localhost:8081"
	
	// 等待服务器启动
	fmt.Println("⏳ 等待服务器启动...")
	time.Sleep(2 * time.Second)
	
	// 测试健康检查
	fmt.Println("\n🔍 测试健康检查...")
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		fmt.Printf("❌ 健康检查失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 200 {
		fmt.Println("✅ 健康检查成功")
	} else {
		fmt.Printf("❌ 健康检查失败，状态码: %d\n", resp.StatusCode)
		return
	}
	
	// 登录获取admin token
	fmt.Println("\n🔐 使用admin账户登录...")
	token, err := loginAsAdmin(baseURL)
	if err != nil {
		fmt.Printf("❌ 登录失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 登录成功，Token: %s...\n", token[:50])
	
	// 测试商品列表API
	fmt.Println("\n📋 测试商品列表API...")
	testProductList(baseURL, token)
	
	// 测试商品详情API
	fmt.Println("\n🔍 测试商品详情API...")
	testProductDetail(baseURL, token)
	
	// 测试商品创建API
	fmt.Println("\n➕ 测试商品创建API...")
	testProductCreate(baseURL, token)
	
	fmt.Println("\n🎉 商品管理模块API测试完成！")
}

func loginAsAdmin(baseURL string) (string, error) {
	loginReq := LoginRequest{
		Username: "admin",
		Password: "admin123",
	}
	
	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		return "", err
	}
	
	resp, err := http.Post(baseURL+"/api/v1/users/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("登录失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}
	
	var loginResp LoginResponse
	err = json.Unmarshal(body, &loginResp)
	if err != nil {
		return "", err
	}
	
	if loginResp.Code != 200 {
		return "", fmt.Errorf("登录失败: %s", loginResp.Msg)
	}
	
	fmt.Printf("👤 用户信息: %s (%s) - 角色: %s\n", 
		loginResp.Data.User.Username, loginResp.Data.User.Email, loginResp.Data.User.Role)
	
	return loginResp.Data.Token, nil
}

func testProductList(baseURL, token string) {
	// 测试不带分页参数
	fmt.Println("  测试1: 不带分页参数")
	resp, err := http.Get(baseURL + "/api/v1/products")
	if err != nil {
		fmt.Printf("  ❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("  📊 状态码: %d\n", resp.StatusCode)
	if resp.StatusCode == 200 {
		fmt.Println("  ✅ 商品列表API（无分页）正常")
	} else {
		fmt.Printf("  ❌ 响应: %s\n", string(body))
	}
	
	// 测试带分页参数
	fmt.Println("  测试2: 带分页参数")
	resp2, err := http.Get(baseURL + "/api/v1/products?page=1&page_size=5")
	if err != nil {
		fmt.Printf("  ❌ 请求失败: %v\n", err)
		return
	}
	defer resp2.Body.Close()
	
	body2, _ := io.ReadAll(resp2.Body)
	fmt.Printf("  📊 状态码: %d\n", resp2.StatusCode)
	if resp2.StatusCode == 200 {
		fmt.Println("  ✅ 商品列表API（带分页）正常")
	} else {
		fmt.Printf("  ❌ 响应: %s\n", string(body2))
	}
}

func testProductDetail(baseURL, token string) {
	// 测试存在的商品ID
	fmt.Println("  测试1: 获取商品ID=1的详情")
	resp, err := http.Get(baseURL + "/api/v1/products/1")
	if err != nil {
		fmt.Printf("  ❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("  📊 状态码: %d\n", resp.StatusCode)
	if resp.StatusCode == 200 {
		fmt.Println("  ✅ 商品详情API正常")
	} else {
		fmt.Printf("  ❌ 响应: %s\n", string(body))
	}
	
	// 测试不存在的商品ID
	fmt.Println("  测试2: 获取不存在的商品ID=999")
	resp2, err := http.Get(baseURL + "/api/v1/products/999")
	if err != nil {
		fmt.Printf("  ❌ 请求失败: %v\n", err)
		return
	}
	defer resp2.Body.Close()
	
	fmt.Printf("  📊 状态码: %d\n", resp2.StatusCode)
	if resp2.StatusCode == 404 {
		fmt.Println("  ✅ 不存在商品返回404正常")
	} else {
		body2, _ := io.ReadAll(resp2.Body)
		fmt.Printf("  ⚠️  预期404，实际: %d, 响应: %s\n", resp2.StatusCode, string(body2))
	}
}

func testProductCreate(baseURL, token string) {
	createReq := CreateProductRequest{
		Name:        "API测试商品",
		Description: "通过API创建的测试商品",
		CategoryID:  1,
		Price:       99.99,
		Stock:       50,
		Status:      "active",
	}
	
	jsonData, err := json.Marshal(createReq)
	if err != nil {
		fmt.Printf("  ❌ JSON序列化失败: %v\n", err)
		return
	}
	
	client := &http.Client{}
	req, err := http.NewRequest("POST", baseURL+"/api/v1/products", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("  ❌ 创建请求失败: %v\n", err)
		return
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("  ❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("  📊 状态码: %d\n", resp.StatusCode)
	fmt.Printf("  📄 响应: %s\n", string(body))
	
	if resp.StatusCode == 200 || resp.StatusCode == 201 {
		fmt.Println("  ✅ 商品创建API正常")
	} else {
		fmt.Printf("  ❌ 商品创建失败，状态码: %d\n", resp.StatusCode)
	}
}
