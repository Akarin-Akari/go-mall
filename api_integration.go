package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	baseURL = "http://localhost:8080"
	token   = ""
)

func main() {
	fmt.Println("🚀 Mall-Go前后端联调测试开始...")
	fmt.Println(strings.Repeat("=", 60))

	// 等待服务器启动
	time.Sleep(3 * time.Second)

	// 执行测试
	testResults := []string{}

	// 1. 健康检查
	if testHealthCheck() {
		testResults = append(testResults, "✅ 健康检查")
	} else {
		testResults = append(testResults, "❌ 健康检查")
	}

	// 2. 用户登录
	if testUserLogin() {
		testResults = append(testResults, "✅ 用户登录")
	} else {
		testResults = append(testResults, "❌ 用户登录")
	}

	// 3. 商品数据
	if testProductData() {
		testResults = append(testResults, "✅ 商品数据")
	} else {
		testResults = append(testResults, "❌ 商品数据")
	}

	// 4. 购物车操作
	if testCartOperations() {
		testResults = append(testResults, "✅ 购物车操作")
	} else {
		testResults = append(testResults, "❌ 购物车操作")
	}

	// 5. 订单创建
	if testOrderCreation() {
		testResults = append(testResults, "✅ 订单创建")
	} else {
		testResults = append(testResults, "❌ 订单创建")
	}

	// 输出结果
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("📊 测试结果汇总:")
	for _, result := range testResults {
		fmt.Println("  " + result)
	}
	fmt.Println(strings.Repeat("=", 60))
}

func testHealthCheck() bool {
	fmt.Println("📋 测试1: 健康检查API")
	
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		fmt.Printf("   ❌ 请求失败: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Println("   ✅ 健康检查成功")
		return true
	}
	
	fmt.Printf("   ❌ 健康检查失败: %d\n", resp.StatusCode)
	return false
}

func testUserLogin() bool {
	fmt.Println("📋 测试2: 用户登录API")
	
	loginData := map[string]interface{}{
		"username": "newuser2024",
		"password": "123456789",
	}

	jsonData, _ := json.Marshal(loginData)
	resp, err := http.Post(baseURL+"/api/v1/users/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("   ❌ 请求失败: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if resp.StatusCode == 200 {
		if data, ok := result["data"].(map[string]interface{}); ok {
			if tokenStr, ok := data["token"].(string); ok {
				token = tokenStr
				fmt.Printf("   ✅ 登录成功，Token: %s...\n", token[:20])
				return true
			}
		}
	}
	
	fmt.Printf("   ❌ 登录失败: %v\n", result["message"])
	return false
}

func testProductData() bool {
	fmt.Println("📋 测试3: 商品数据验证")
	
	resp, err := http.Get(baseURL + "/api/v1/products?page=1&page_size=10")
	if err != nil {
		fmt.Printf("   ❌ 请求失败: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if resp.StatusCode == 200 {
		if data, ok := result["data"].(map[string]interface{}); ok {
			if list, ok := data["list"].([]interface{}); ok {
				productCount := len(list)
				fmt.Printf("   ✅ 商品数据验证成功: %d个商品\n", productCount)
				
				// 显示前3个商品
				for i, item := range list {
					if i >= 3 {
						break
					}
					if product, ok := item.(map[string]interface{}); ok {
						fmt.Printf("      商品%d: %s (价格: %v)\n", i+1, product["name"], product["price"])
					}
				}
				return productCount >= 10
			}
		}
	}
	
	fmt.Printf("   ❌ 商品数据验证失败: %d\n", resp.StatusCode)
	return false
}

func testCartOperations() bool {
	fmt.Println("📋 测试4: 购物车操作")
	
	if token == "" {
		fmt.Println("   ❌ 缺少认证Token")
		return false
	}

	// 添加商品到购物车
	cartData := map[string]interface{}{
		"product_id": 1,
		"quantity":   2,
	}

	jsonData, _ := json.Marshal(cartData)
	req, _ := http.NewRequest("POST", baseURL+"/api/v1/cart/items", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("   ❌ 请求失败: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Println("   ✅ 购物车操作成功")
		return true
	}
	
	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	fmt.Printf("   ❌ 购物车操作失败: %v\n", result["message"])
	return false
}

func testOrderCreation() bool {
	fmt.Println("📋 测试5: 订单创建")
	
	if token == "" {
		fmt.Println("   ❌ 缺少认证Token")
		return false
	}

	orderData := map[string]interface{}{
		"cart_item_ids":    []uint{1},
		"receiver_name":    "张三",
		"receiver_phone":   "13800138000",
		"receiver_address": "北京市朝阳区测试地址123号",
		"province":         "北京市",
		"city":             "北京市",
		"district":         "朝阳区",
		"shipping_method":  "standard",
		"buyer_message":    "测试订单",
	}

	jsonData, _ := json.Marshal(orderData)
	req, _ := http.NewRequest("POST", baseURL+"/api/v1/orders", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("   ❌ 请求失败: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if resp.StatusCode == 200 {
		fmt.Println("   ✅ 订单创建成功")
		if data, ok := result["data"].(map[string]interface{}); ok {
			if orderNo, ok := data["order_no"].(string); ok {
				fmt.Printf("      订单号: %s\n", orderNo)
			}
		}
		return true
	}
	
	fmt.Printf("   ❌ 订单创建失败: %v\n", result["message"])
	return false
}
