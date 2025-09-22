package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const baseURL = "http://localhost:8081"

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
			Role     string `json:"role"`
		} `json:"user"`
	} `json:"data"`
}

type ProductCreateRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       float64  `json:"price"`
	Stock       int      `json:"stock"`
	CategoryID  int      `json:"category_id"`
	Images      []string `json:"images"`
}

func main() {
	fmt.Println("🔧 详细管理员权限测试")
	fmt.Println("================================================================================")

	// 1. 管理员登录
	fmt.Println("\n=== 步骤1: 管理员登录 ===")
	adminToken, err := loginAsAdmin()
	if err != nil {
		fmt.Printf("❌ 管理员登录失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 管理员登录成功\n")
	fmt.Printf("Token: %s...\n", adminToken[:50])

	// 2. 测试认证中间件
	fmt.Println("\n=== 步骤2: 测试用户资料API（需要认证） ===")
	if err := testUserProfile(adminToken); err != nil {
		fmt.Printf("❌ 用户资料API测试失败: %v\n", err)
	} else {
		fmt.Printf("✅ 用户资料API测试成功\n")
	}

	// 3. 测试商品创建API（需要管理员权限）
	fmt.Println("\n=== 步骤3: 测试商品创建API（需要管理员权限） ===")
	if err := testProductCreateDetailed(adminToken); err != nil {
		fmt.Printf("❌ 商品创建API测试失败: %v\n", err)
	} else {
		fmt.Printf("✅ 商品创建API测试成功\n")
	}

	// 4. 测试商品列表API（公开接口）
	fmt.Println("\n=== 步骤4: 测试商品列表API（公开接口） ===")
	if err := testProductListDetailed(); err != nil {
		fmt.Printf("❌ 商品列表API测试失败: %v\n", err)
	} else {
		fmt.Printf("✅ 商品列表API测试成功\n")
	}

	fmt.Println("\n================================================================================")
	fmt.Println("🎉 详细管理员权限测试完成！")
}

func loginAsAdmin() (string, error) {
	loginReq := LoginRequest{
		Username: "admin",
		Password: "password123",
	}

	jsonData, _ := json.Marshal(loginReq)
	resp, err := http.Post(baseURL+"/api/v1/users/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("登录响应状态码: %d\n", resp.StatusCode)
	fmt.Printf("登录响应内容: %s\n", string(body))
	
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("登录失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return "", fmt.Errorf("解析登录响应失败: %v", err)
	}

	if loginResp.Data.Token == "" {
		return "", fmt.Errorf("登录响应中没有token")
	}

	fmt.Printf("用户信息: ID=%d, Username=%s, Role=%s\n", 
		loginResp.Data.User.ID, loginResp.Data.User.Username, loginResp.Data.User.Role)

	return loginResp.Data.Token, nil
}

func testUserProfile(token string) error {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", baseURL+"/api/v1/users/profile", nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("用户资料API响应状态码: %d\n", resp.StatusCode)
	fmt.Printf("用户资料API响应内容: %s\n", string(body))

	if resp.StatusCode != 200 {
		return fmt.Errorf("状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	return nil
}

func testProductCreateDetailed(token string) error {
	productReq := ProductCreateRequest{
		Name:        "详细测试商品",
		Description: "这是详细测试创建的商品",
		Price:       199.99,
		Stock:       50,
		CategoryID:  1,
		Images:      []string{"https://example.com/detailed-test-product.jpg"},
	}

	jsonData, _ := json.Marshal(productReq)
	fmt.Printf("商品创建请求数据: %s\n", string(jsonData))
	
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", baseURL+"/api/v1/products", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	fmt.Printf("请求头: Authorization: Bearer %s...\n", token[:30])

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("商品创建API响应状态码: %d\n", resp.StatusCode)
	fmt.Printf("商品创建API响应内容: %s\n", string(body))

	if resp.StatusCode != 200 {
		return fmt.Errorf("状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	return nil
}

func testProductListDetailed() error {
	resp, err := http.Get(baseURL + "/api/v1/products?page=1&page_size=10")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("商品列表API响应状态码: %d\n", resp.StatusCode)
	fmt.Printf("商品列表API响应内容: %s\n", string(body))

	if resp.StatusCode != 200 {
		return fmt.Errorf("状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	return nil
}
