package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	fmt.Println("🛒 开始测试Mall-Go购物车功能...")

	// 等待服务器启动
	time.Sleep(2 * time.Second)

	// 首先登录获取token
	fmt.Println("🔐 用户登录...")
	token := testLogin()
	if token == "" {
		fmt.Println("❌ 登录失败，无法继续测试")
		return
	}

	// 测试添加商品到购物车
	fmt.Println("➕ 测试添加商品到购物车...")
	testAddToCart(token)

	// 测试获取购物车
	fmt.Println("📋 测试获取购物车...")
	testGetCart(token)

	// 测试更新购物车商品数量
	fmt.Println("🔄 测试更新购物车商品数量...")
	testUpdateCartItem(token)

	// 测试删除购物车商品
	fmt.Println("🗑️ 测试删除购物车商品...")
	testRemoveFromCart(token)

	fmt.Println("✅ 购物车功能测试完成!")
}

func testLogin() string {
	loginData := map[string]interface{}{
		"username": "newuser2024",
		"password": "123456789",
	}

	jsonData, _ := json.Marshal(loginData)
	resp, err := http.Post("http://localhost:8080/api/v1/users/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ 登录请求失败: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取登录响应失败: %v\n", err)
		return ""
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("❌ 解析登录响应失败: %v\n", err)
		return ""
	}

	if data, ok := result["data"].(map[string]interface{}); ok {
		if token, ok := data["token"].(string); ok {
			fmt.Printf("✅ 登录成功，获取token: %s...\n", token[:20])
			return token
		}
	}

	fmt.Printf("❌ 登录失败: %s\n", result["message"])
	return ""
}

func testAddToCart(token string) {
	cartData := map[string]interface{}{
		"product_id": 1,
		"quantity":   2,
	}

	jsonData, _ := json.Marshal(cartData)
	req, _ := http.NewRequest("POST", "http://localhost:8080/api/v1/cart/items", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 添加商品到购物车请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("❌ JSON解析失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 添加商品到购物车结果\n")
	fmt.Printf("   响应码: %.0f\n", result["code"])
	fmt.Printf("   消息: %s\n", result["message"])

	if data, ok := result["data"].(map[string]interface{}); ok {
		if productName, ok := data["product_name"]; ok {
			fmt.Printf("   商品名称: %s\n", productName)
		}
		if quantity, ok := data["quantity"]; ok {
			fmt.Printf("   数量: %.0f\n", quantity)
		}
		if price, ok := data["price"]; ok {
			fmt.Printf("   价格: %s\n", price)
		}
	}
}

func testGetCart(token string) {
	req, _ := http.NewRequest("GET", "http://localhost:8080/api/v1/cart", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 获取购物车请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("❌ JSON解析失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 获取购物车结果\n")
	fmt.Printf("   响应码: %.0f\n", result["code"])
	fmt.Printf("   消息: %s\n", result["message"])

	if data, ok := result["data"].(map[string]interface{}); ok {
		if cart, ok := data["cart"].(map[string]interface{}); ok {
			if items, ok := cart["items"].([]interface{}); ok {
				fmt.Printf("   购物车商品数量: %d\n", len(items))
				for i, item := range items {
					if itemMap, ok := item.(map[string]interface{}); ok {
						fmt.Printf("   商品%d: %s (数量: %.0f)\n", i+1, itemMap["product_name"], itemMap["quantity"])
					}
				}
			}
		}
		if summary, ok := data["summary"].(map[string]interface{}); ok {
			if totalAmount, ok := summary["total_amount"]; ok {
				fmt.Printf("   总金额: %s\n", totalAmount)
			}
		}
	}
}

func testUpdateCartItem(token string) {
	updateData := map[string]interface{}{
		"quantity": 3,
	}

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "http://localhost:8080/api/v1/cart/items/1", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 更新购物车商品请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("❌ JSON解析失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 更新购物车商品结果\n")
	fmt.Printf("   响应码: %.0f\n", result["code"])
	fmt.Printf("   消息: %s\n", result["message"])

	if data, ok := result["data"].(map[string]interface{}); ok {
		if quantity, ok := data["quantity"]; ok {
			fmt.Printf("   更新后数量: %.0f\n", quantity)
		}
	}
}

func testRemoveFromCart(token string) {
	req, _ := http.NewRequest("DELETE", "http://localhost:8080/api/v1/cart/items/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 删除购物车商品请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("❌ JSON解析失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 删除购物车商品结果\n")
	fmt.Printf("   响应码: %.0f\n", result["code"])
	fmt.Printf("   消息: %s\n", result["message"])
}
