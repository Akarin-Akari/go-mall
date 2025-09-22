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

func main() {
	fmt.Println("🚀 Mall-Go订单API修复验证测试")
	fmt.Println("========================================")

	// 1. 健康检查
	fmt.Println("\n🔍 步骤1: 健康检查")
	if !testHealth() {
		fmt.Println("❌ 服务器未运行，测试终止")
		return
	}

	// 2. 用户注册
	fmt.Println("\n📝 步骤2: 用户注册")
	testRegister()

	// 3. 用户登录获取Token
	fmt.Println("\n🔐 步骤3: 用户登录")
	token := testLogin()
	if token == "" {
		fmt.Println("❌ 登录失败，测试终止")
		return
	}

	// 4. 测试订单列表API
	fmt.Println("\n📋 步骤4: 测试订单列表API")
	orderListSuccess := testOrderList(token)

	// 5. 测试订单创建API
	fmt.Println("\n🛒 步骤5: 测试订单创建API")
	orderCreateSuccess := testOrderCreate(token)

	// 6. 总结测试结果
	fmt.Println("\n📊 测试结果总结:")
	fmt.Println("========================================")
	if orderListSuccess {
		fmt.Println("✅ 订单列表API: 正常")
	} else {
		fmt.Println("❌ 订单列表API: 异常")
	}

	if orderCreateSuccess {
		fmt.Println("✅ 订单创建API: 正常")
	} else {
		fmt.Println("⚠️ 订单创建API: 有问题（可能是业务逻辑错误）")
	}

	// 计算成功率
	successCount := 0
	if orderListSuccess {
		successCount++
	}
	if orderCreateSuccess {
		successCount++
	}
	successRate := float64(successCount) / 2.0 * 100
	fmt.Printf("📈 订单模块成功率: %.1f%%\n", successRate)

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

func testRegister() {
	registerData := map[string]string{
		"username": "newuser2024",
		"email":    "newuser2024@example.com",
		"password": "123456789",
	}

	jsonData, _ := json.Marshal(registerData)
	resp, err := http.Post(baseURL+"/api/v1/users/register", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ 注册请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("注册响应状态: %d\n", resp.StatusCode)

	if resp.StatusCode == 200 || resp.StatusCode == 201 {
		fmt.Println("✅ 注册成功")
	} else {
		fmt.Println("⚠️ 注册失败或用户已存在（这是正常的）")
	}
}

func testLogin() string {
	loginData := map[string]string{
		"username": "newuser2024",
		"password": "123456789",
	}

	jsonData, _ := json.Marshal(loginData)
	resp, err := http.Post(baseURL+"/api/v1/users/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ 登录请求失败: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("登录响应状态: %d\n", resp.StatusCode)

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

	fmt.Printf("❌ 登录失败，响应: %s\n", string(body))
	return ""
}

func testOrderList(token string) bool {
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", baseURL+"/api/v1/orders?page=1&page_size=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 订单列表请求失败: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("订单列表响应状态: %d\n", resp.StatusCode)

	if resp.StatusCode == 200 {
		fmt.Println("✅ 订单列表API正常")
		return true
	} else {
		fmt.Printf("❌ 订单列表API异常，响应: %s\n", string(body))
		return false
	}
}

func testOrderCreate(token string) bool {
	// 创建订单请求数据
	orderData := map[string]interface{}{
		"cart_item_ids":  []int{1, 2}, // 假设的购物车商品项ID
		"address_id":     1,           // 假设的地址ID
		"payment_method": "alipay",
		"remark":         "测试订单",
	}

	jsonData, _ := json.Marshal(orderData)
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("POST", baseURL+"/api/v1/orders", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 订单创建请求失败: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("订单创建响应状态: %d\n", resp.StatusCode)
	fmt.Printf("订单创建响应内容: %s\n", string(body))

	if resp.StatusCode == 200 {
		fmt.Println("✅ 订单创建API正常")
		return true
	} else if resp.StatusCode == 400 {
		fmt.Println("⚠️ 订单创建API返回400 - 业务逻辑错误（如购物车为空）")
		return false
	} else if resp.StatusCode == 500 {
		fmt.Println("❌ 订单创建API返回500 - 服务器内部错误")
		return false
	} else {
		fmt.Printf("❌ 订单创建API异常: %d\n", resp.StatusCode)
		return false
	}
}
