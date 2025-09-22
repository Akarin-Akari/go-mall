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

type Result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
	} `json:"user"`
}

func main() {
	fmt.Println("🚀 Mall-Go订单API快速测试")
	fmt.Println("=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=" + "=")

	// 1. 健康检查
	fmt.Println("\n🔍 步骤1: 健康检查")
	if !testHealth() {
		fmt.Println("❌ 服务器未运行，测试终止")
		return
	}

	// 2. 用户登录获取Token
	fmt.Println("\n🔐 步骤2: 用户登录")
	token := testLogin()
	if token == "" {
		fmt.Println("❌ 登录失败，测试终止")
		return
	}

	// 3. 测试订单列表API
	fmt.Println("\n📋 步骤3: 测试订单列表API")
	testOrderList(token)

	// 4. 测试订单创建API
	fmt.Println("\n🛒 步骤4: 测试订单创建API")
	testOrderCreate(token)

	fmt.Println("\n✅ 测试完成！")
}

func testHealth() bool {
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		fmt.Printf("❌ 健康检查失败: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Println("✅ 服务器运行正常")
		return true
	}
	fmt.Printf("❌ 服务器状态异常: %d\n", resp.StatusCode)
	return false
}

func testLogin() string {
	loginData := map[string]string{
		"username": "newuser2024",
		"password": "123456789",
	}

	jsonData, _ := json.Marshal(loginData)
	resp, err := http.Post(baseURL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ 登录请求失败: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("登录响应状态: %d\n", resp.StatusCode)
	fmt.Printf("登录响应内容: %s\n", string(body))

	if resp.StatusCode == 200 {
		var result Result
		if err := json.Unmarshal(body, &result); err == nil {
			if loginResp, ok := result.Data.(map[string]interface{}); ok {
				if token, exists := loginResp["token"]; exists {
					fmt.Println("✅ 登录成功")
					return token.(string)
				}
			}
		}
	}

	fmt.Println("❌ 登录失败")
	return ""
}

func testOrderList(token string) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", baseURL+"/api/v1/orders?page=1&page_size=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 订单列表请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("订单列表响应状态: %d\n", resp.StatusCode)
	fmt.Printf("订单列表响应内容: %s\n", string(body))

	if resp.StatusCode == 200 {
		fmt.Println("✅ 订单列表API正常")
	} else {
		fmt.Printf("❌ 订单列表API异常: %d\n", resp.StatusCode)
	}
}

func testOrderCreate(token string) {
	// 创建订单请求数据
	orderData := map[string]interface{}{
		"cart_item_ids": []int{1, 2}, // 假设的购物车商品项ID
		"address_id":    1,           // 假设的地址ID
		"payment_method": "alipay",
		"remark":        "测试订单",
	}

	jsonData, _ := json.Marshal(orderData)
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("POST", baseURL+"/api/v1/orders", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 订单创建请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("订单创建响应状态: %d\n", resp.StatusCode)
	fmt.Printf("订单创建响应内容: %s\n", string(body))

	if resp.StatusCode == 200 {
		fmt.Println("✅ 订单创建API正常")
	} else if resp.StatusCode == 400 {
		fmt.Println("⚠️ 订单创建API返回400 - 可能是业务逻辑错误（如购物车为空）")
	} else if resp.StatusCode == 500 {
		fmt.Println("❌ 订单创建API返回500 - 服务器内部错误")
	} else {
		fmt.Printf("❌ 订单创建API异常: %d\n", resp.StatusCode)
	}
}
