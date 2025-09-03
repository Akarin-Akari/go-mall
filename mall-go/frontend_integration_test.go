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

// 测试结果结构
type TestResult struct {
	TestName string      `json:"test_name"`
	Success  bool        `json:"success"`
	Message  string      `json:"message"`
	Data     interface{} `json:"data,omitempty"`
}

// 全局变量
var (
	baseURL     = "http://localhost:8080"
	token       = ""
	userID      uint
	testResults []TestResult
)

func main() {
	fmt.Println("🚀 Mall-Go前后端联调测试开始...")
	fmt.Println(strings.Repeat("=", 60))

	// 等待服务器启动
	time.Sleep(3 * time.Second)

	// 执行完整的业务流程测试
	runFullIntegrationTest()

	// 输出测试结果
	printTestResults()
}

// 执行完整的集成测试
func runFullIntegrationTest() {
	// 1. 健康检查
	testHealthCheck()

	// 2. 用户注册
	testUserRegister()

	// 3. 用户登录
	testUserLogin()

	// 4. 商品数据验证
	testProductData()

	// 5. 购物车操作
	testCartOperations()

	// 6. 订单创建
	testOrderCreation()

	// 7. 订单管理
	testOrderManagement()
}

// 测试健康检查
func testHealthCheck() {
	fmt.Println("📋 测试1: 健康检查API")

	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		addTestResult("健康检查", false, "请求失败: "+err.Error(), nil)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if resp.StatusCode == 200 {
		addTestResult("健康检查", true, "API服务器正常运行", result)
		fmt.Printf("   ✅ 健康检查成功: %s\n", result["message"])
	} else {
		addTestResult("健康检查", false, fmt.Sprintf("状态码: %d", resp.StatusCode), result)
		fmt.Printf("   ❌ 健康检查失败: %d\n", resp.StatusCode)
	}
}

// 测试用户注册
func testUserRegister() {
	fmt.Println("📋 测试2: 用户注册API")

	registerData := map[string]interface{}{
		"username": "testuser_" + fmt.Sprintf("%d", time.Now().Unix()),
		"email":    "test_" + fmt.Sprintf("%d", time.Now().Unix()) + "@example.com",
		"password": "123456789",
		"nickname": "测试用户",
	}

	jsonData, _ := json.Marshal(registerData)
	resp, err := http.Post(baseURL+"/api/v1/users/register", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		addTestResult("用户注册", false, "请求失败: "+err.Error(), nil)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if resp.StatusCode == 200 {
		addTestResult("用户注册", true, "用户注册成功", result)
		fmt.Printf("   ✅ 用户注册成功: %s\n", registerData["username"])

		// 保存用户信息用于登录
		if data, ok := result["data"].(map[string]interface{}); ok {
			if id, ok := data["id"].(float64); ok {
				userID = uint(id)
			}
		}
	} else {
		addTestResult("用户注册", false, fmt.Sprintf("状态码: %d", resp.StatusCode), result)
		fmt.Printf("   ❌ 用户注册失败: %s\n", result["message"])
	}
}

// 测试用户登录
func testUserLogin() {
	fmt.Println("📋 测试3: 用户登录API")

	// 使用已知的测试用户
	loginData := map[string]interface{}{
		"username": "newuser2024",
		"password": "123456789",
	}

	jsonData, _ := json.Marshal(loginData)
	resp, err := http.Post(baseURL+"/api/v1/users/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		addTestResult("用户登录", false, "请求失败: "+err.Error(), nil)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if resp.StatusCode == 200 {
		if data, ok := result["data"].(map[string]interface{}); ok {
			if tokenStr, ok := data["token"].(string); ok {
				token = tokenStr
				addTestResult("用户登录", true, "登录成功，获取Token", map[string]string{"token": token[:20] + "..."})
				fmt.Printf("   ✅ 用户登录成功，Token: %s...\n", token[:20])
				return
			}
		}
	}

	addTestResult("用户登录", false, fmt.Sprintf("状态码: %d", resp.StatusCode), result)
	fmt.Printf("   ❌ 用户登录失败: %s\n", result["message"])
}

// 测试商品数据
func testProductData() {
	fmt.Println("📋 测试4: 商品数据验证")

	resp, err := http.Get(baseURL + "/api/v1/products?page=1&page_size=10")
	if err != nil {
		addTestResult("商品数据验证", false, "请求失败: "+err.Error(), nil)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if resp.StatusCode == 200 {
		if data, ok := result["data"].(map[string]interface{}); ok {
			if list, ok := data["list"].([]interface{}); ok {
				productCount := len(list)
				if productCount >= 15 {
					addTestResult("商品数据验证", true, fmt.Sprintf("商品数据完整，共%d个商品", productCount), map[string]int{"count": productCount})
					fmt.Printf("   ✅ 商品数据验证成功: %d个商品\n", productCount)

					// 显示前3个商品
					for i, item := range list[:3] {
						if product, ok := item.(map[string]interface{}); ok {
							fmt.Printf("      商品%d: %s (价格: %v)\n", i+1, product["name"], product["price"])
						}
					}
				} else {
					addTestResult("商品数据验证", false, fmt.Sprintf("商品数量不足，仅有%d个", productCount), map[string]int{"count": productCount})
					fmt.Printf("   ⚠️ 商品数量不足: %d个\n", productCount)
				}
			}
		}
	} else {
		addTestResult("商品数据验证", false, fmt.Sprintf("状态码: %d", resp.StatusCode), result)
		fmt.Printf("   ❌ 商品数据验证失败: %d\n", resp.StatusCode)
	}
}

// 测试购物车操作
func testCartOperations() {
	fmt.Println("📋 测试5: 购物车操作")

	if token == "" {
		addTestResult("购物车操作", false, "缺少认证Token", nil)
		fmt.Println("   ❌ 购物车测试跳过: 缺少认证Token")
		return
	}

	// 5.1 添加商品到购物车
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
		addTestResult("购物车-添加商品", false, "请求失败: "+err.Error(), nil)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if resp.StatusCode == 200 {
		addTestResult("购物车-添加商品", true, "商品添加到购物车成功", result)
		fmt.Printf("   ✅ 添加商品到购物车成功\n")
	} else {
		addTestResult("购物车-添加商品", false, fmt.Sprintf("状态码: %d", resp.StatusCode), result)
		fmt.Printf("   ❌ 添加商品到购物车失败: %s\n", result["message"])
		return
	}

	// 5.2 获取购物车
	req, _ = http.NewRequest("GET", baseURL+"/api/v1/cart", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = client.Do(req)
	if err != nil {
		addTestResult("购物车-获取购物车", false, "请求失败: "+err.Error(), nil)
		return
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	json.Unmarshal(body, &result)

	if resp.StatusCode == 200 {
		addTestResult("购物车-获取购物车", true, "获取购物车成功", result)
		fmt.Printf("   ✅ 获取购物车成功\n")
	} else {
		addTestResult("购物车-获取购物车", false, fmt.Sprintf("状态码: %d", resp.StatusCode), result)
		fmt.Printf("   ❌ 获取购物车失败: %s\n", result["message"])
	}
}

// 测试订单创建
func testOrderCreation() {
	fmt.Println("📋 测试6: 订单创建")

	if token == "" {
		addTestResult("订单创建", false, "缺少认证Token", nil)
		fmt.Println("   ❌ 订单创建测试跳过: 缺少认证Token")
		return
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
		addTestResult("订单创建", false, "请求失败: "+err.Error(), nil)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if resp.StatusCode == 200 {
		addTestResult("订单创建", true, "订单创建成功", result)
		fmt.Printf("   ✅ 订单创建成功\n")

		if data, ok := result["data"].(map[string]interface{}); ok {
			if orderNo, ok := data["order_no"].(string); ok {
				fmt.Printf("      订单号: %s\n", orderNo)
			}
			if totalAmount, ok := data["total_amount"]; ok {
				fmt.Printf("      订单金额: %v\n", totalAmount)
			}
		}
	} else {
		addTestResult("订单创建", false, fmt.Sprintf("状态码: %d", resp.StatusCode), result)
		fmt.Printf("   ❌ 订单创建失败: %s\n", result["message"])
	}
}

// 测试订单管理
func testOrderManagement() {
	fmt.Println("📋 测试7: 订单管理")

	if token == "" {
		addTestResult("订单管理", false, "缺少认证Token", nil)
		fmt.Println("   ❌ 订单管理测试跳过: 缺少认证Token")
		return
	}

	req, _ := http.NewRequest("GET", baseURL+"/api/v1/orders?page=1&page_size=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		addTestResult("订单管理", false, "请求失败: "+err.Error(), nil)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if resp.StatusCode == 200 {
		addTestResult("订单管理", true, "获取订单列表成功", result)
		fmt.Printf("   ✅ 获取订单列表成功\n")

		if data, ok := result["data"].(map[string]interface{}); ok {
			if orders, ok := data["orders"].([]interface{}); ok {
				fmt.Printf("      订单数量: %d\n", len(orders))
			}
		}
	} else {
		addTestResult("订单管理", false, fmt.Sprintf("状态码: %d", resp.StatusCode), result)
		fmt.Printf("   ❌ 获取订单列表失败: %s\n", result["message"])
	}
}

// 添加测试结果
func addTestResult(testName string, success bool, message string, data interface{}) {
	testResults = append(testResults, TestResult{
		TestName: testName,
		Success:  success,
		Message:  message,
		Data:     data,
	})
}

// 打印测试结果
func printTestResults() {
	fmt.Println("\n" + "="*60)
	fmt.Println("📊 前后端联调测试结果汇总")
	fmt.Println("=" * 60)

	successCount := 0
	totalCount := len(testResults)

	for i, result := range testResults {
		status := "❌"
		if result.Success {
			status = "✅"
			successCount++
		}
		fmt.Printf("%d. %s %s - %s\n", i+1, status, result.TestName, result.Message)
	}

	fmt.Println("=" * 60)
	fmt.Printf("📈 测试通过率: %d/%d (%.1f%%)\n", successCount, totalCount, float64(successCount)/float64(totalCount)*100)

	if successCount == totalCount {
		fmt.Println("🎉 所有测试通过！前后端联调准备就绪！")
	} else {
		fmt.Printf("⚠️ %d个测试失败，需要修复后再进行前后端联调\n", totalCount-successCount)
	}
}
