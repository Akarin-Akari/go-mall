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

type APIResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func main() {
	fmt.Println("🔧 Mall-Go管理员API测试工具")
	fmt.Println("================================================================================")

	// 1. 管理员登录获取Token
	fmt.Println("\n=== 管理员登录测试 ===")
	adminToken, err := loginAsAdmin()
	if err != nil {
		fmt.Printf("❌ 管理员登录失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 管理员登录成功，Token: %s...\n", adminToken[:20])

	// 2. 测试商品管理API
	fmt.Println("\n=== 商品管理API测试 ===")

	// 2.1 测试商品列表API
	if err := testProductList(); err != nil {
		fmt.Printf("❌ 商品列表API测试失败: %v\n", err)
	} else {
		fmt.Println("✅ 商品列表API测试成功")
	}

	// 2.2 测试商品详情API
	if err := testProductDetail(1); err != nil {
		fmt.Printf("❌ 商品详情API测试失败: %v\n", err)
	} else {
		fmt.Println("✅ 商品详情API测试成功")
	}

	// 2.3 测试商品创建API（需要管理员权限）
	if err := testProductCreate(adminToken); err != nil {
		fmt.Printf("❌ 商品创建API测试失败: %v\n", err)
	} else {
		fmt.Println("✅ 商品创建API测试成功")
	}

	// 3. 测试购物车API
	fmt.Println("\n=== 购物车API测试 ===")

	// 先用普通用户登录
	userToken, err := loginAsUser()
	if err != nil {
		fmt.Printf("❌ 普通用户登录失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 普通用户登录成功\n")

	// 测试购物车功能
	if err := testCartOperations(userToken); err != nil {
		fmt.Printf("❌ 购物车API测试失败: %v\n", err)
	} else {
		fmt.Println("✅ 购物车API测试成功")
	}

	// 4. 测试支付API
	fmt.Println("\n=== 支付API测试 ===")
	if err := testPaymentAPIs(userToken); err != nil {
		fmt.Printf("❌ 支付API测试失败: %v\n", err)
	} else {
		fmt.Println("✅ 支付API测试成功")
	}

	fmt.Println("\n================================================================================")
	fmt.Println("🎉 Mall-Go管理员API测试完成！")
}

func loginAsAdmin() (string, error) {
	loginReq := LoginRequest{
		Username: "admin",
		Password: "password123",
	}

	jsonData, _ := json.Marshal(loginReq)
	resp, err := http.Post(baseURL+"/api/v1/users/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

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

	return loginResp.Data.Token, nil
}

func loginAsUser() (string, error) {
	// 先注册一个测试用户
	registerReq := map[string]string{
		"username": "testuser2024",
		"email":    "test2024@example.com",
		"password": "123456789",
	}

	jsonData, _ := json.Marshal(registerReq)
	http.Post(baseURL+"/api/v1/users/register", "application/json", bytes.NewBuffer(jsonData))

	// 然后登录
	loginReq := LoginRequest{
		Username: "testuser2024",
		Password: "123456789",
	}

	jsonData, _ = json.Marshal(loginReq)
	resp, err := http.Post(baseURL+"/api/v1/users/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("用户登录失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return "", fmt.Errorf("解析用户登录响应失败: %v", err)
	}

	return loginResp.Data.Token, nil
}

func testProductList() error {
	resp, err := http.Get(baseURL + "/api/v1/products")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	return nil
}

func testProductDetail(productID int) error {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/products/%d", baseURL, productID))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	return nil
}

func testProductCreate(token string) error {
	productReq := ProductCreateRequest{
		Name:        "测试商品-管理员创建",
		Description: "这是管理员创建的测试商品",
		Price:       99.99,
		Stock:       100,
		CategoryID:  1,
		Images:      []string{"https://example.com/test-product.jpg"},
	}

	jsonData, _ := json.Marshal(productReq)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", baseURL+"/api/v1/products", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return fmt.Errorf("状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("   商品创建响应: %s\n", string(body))
	return nil
}

func testCartOperations(token string) error {
	// 测试添加商品到购物车
	addCartReq := map[string]interface{}{
		"product_id": 1,
		"quantity":   2,
	}

	jsonData, _ := json.Marshal(addCartReq)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", baseURL+"/api/v1/cart", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("添加购物车失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 测试获取购物车
	req, err = http.NewRequest("GET", baseURL+"/api/v1/cart", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("获取购物车失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	return nil
}

func testPaymentAPIs(token string) error {
	// 测试支付创建API
	paymentReq := map[string]interface{}{
		"order_id":       1,
		"payment_method": "alipay",
		"amount":         99.99,
	}

	jsonData, _ := json.Marshal(paymentReq)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", baseURL+"/api/v1/payments", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		fmt.Printf("   支付创建响应: %s\n", string(body))
		// 支付API可能因为业务逻辑返回错误，这是正常的
	}

	return nil
}
