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

type AddToCartRequest struct {
	ProductID uint `json:"product_id"`
	Quantity  int  `json:"quantity"`
}

type UpdateCartRequest struct {
	Quantity int `json:"quantity"`
}

func main() {
	fmt.Println("🛒 测试购物车模块API")

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

	// 登录获取普通用户token
	fmt.Println("\n🔐 使用普通用户账户登录...")
	token, err := loginAsUser(baseURL)
	if err != nil {
		fmt.Printf("❌ 登录失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 登录成功，Token: %s...\n", token[:50])

	// 测试添加商品到购物车
	fmt.Println("\n➕ 测试添加商品到购物车...")
	testAddToCart(baseURL, token)

	// 测试获取购物车列表
	fmt.Println("\n📋 测试获取购物车列表...")
	testGetCart(baseURL, token)

	// 测试更新购物车商品数量
	fmt.Println("\n🔄 测试更新购物车商品数量...")
	testUpdateCart(baseURL, token)

	// 测试清空购物车
	fmt.Println("\n🗑️ 测试清空购物车...")
	testClearCart(baseURL, token)

	fmt.Println("\n🎉 购物车模块API测试完成！")
}

func loginAsUser(baseURL string) (string, error) {
	// 使用第一个普通用户账户
	loginReq := LoginRequest{
		Username: "testuser",
		Password: "password123",
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

func testAddToCart(baseURL, token string) {
	addReq := AddToCartRequest{
		ProductID: 1, // 使用存在的商品ID
		Quantity:  2,
	}

	jsonData, err := json.Marshal(addReq)
	if err != nil {
		fmt.Printf("  ❌ JSON序列化失败: %v\n", err)
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", baseURL+"/api/v1/cart", bytes.NewBuffer(jsonData))
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
		fmt.Println("  ✅ 添加商品到购物车API正常")
	} else {
		fmt.Printf("  ❌ 添加商品到购物车失败，状态码: %d\n", resp.StatusCode)
	}
}

func testGetCart(baseURL, token string) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", baseURL+"/api/v1/cart", nil)
	if err != nil {
		fmt.Printf("  ❌ 创建请求失败: %v\n", err)
		return
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("  ❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("  📊 状态码: %d\n", resp.StatusCode)

	if resp.StatusCode == 200 {
		fmt.Println("  ✅ 获取购物车列表API正常")
	} else {
		fmt.Printf("  ❌ 响应: %s\n", string(body))
	}
}

func testUpdateCart(baseURL, token string) {
	// 假设购物车中有ID为1的商品
	updateReq := UpdateCartRequest{
		Quantity: 3,
	}

	jsonData, err := json.Marshal(updateReq)
	if err != nil {
		fmt.Printf("  ❌ JSON序列化失败: %v\n", err)
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("PUT", baseURL+"/api/v1/cart/1", bytes.NewBuffer(jsonData))
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

	if resp.StatusCode == 200 {
		fmt.Println("  ✅ 更新购物车商品数量API正常")
	} else {
		fmt.Printf("  ❌ 更新购物车商品数量失败，状态码: %d\n", resp.StatusCode)
	}
}

func testClearCart(baseURL, token string) {
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", baseURL+"/api/v1/cart/clear", nil)
	if err != nil {
		fmt.Printf("  ❌ 创建请求失败: %v\n", err)
		return
	}

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

	if resp.StatusCode == 200 {
		fmt.Println("  ✅ 清空购物车API正常")
	} else {
		fmt.Printf("  ❌ 清空购物车失败，状态码: %d\n", resp.StatusCode)
	}
}
