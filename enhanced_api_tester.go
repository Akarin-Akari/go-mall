package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// API测试结果结构
type APITestResult struct {
	Module       string        `json:"module"`
	Endpoint     string        `json:"endpoint"`
	Method       string        `json:"method"`
	StatusCode   int           `json:"status_code"`
	Success      bool          `json:"success"`
	ResponseTime time.Duration `json:"response_time"`
	Error        string        `json:"error,omitempty"`
	Response     interface{}   `json:"response,omitempty"`
}

// API测试客户端
type EnhancedAPITester struct {
	BaseURL   string
	Client    *http.Client
	UserToken string
	Results   []APITestResult
}

// 创建新的API测试器
func NewEnhancedAPITester(baseURL string) *EnhancedAPITester {
	return &EnhancedAPITester{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
		Results: make([]APITestResult, 0),
	}
}

// 执行HTTP请求测试
func (eat *EnhancedAPITester) TestRequest(module, method, endpoint string, body interface{}, token string) APITestResult {
	start := time.Now()

	var reqBody io.Reader
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	}

	url := eat.BaseURL + endpoint
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return APITestResult{
			Module:       module,
			Endpoint:     endpoint,
			Method:       method,
			Success:      false,
			Error:        fmt.Sprintf("创建请求失败: %v", err),
			ResponseTime: time.Since(start),
		}
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := eat.Client.Do(req)
	responseTime := time.Since(start)

	if err != nil {
		return APITestResult{
			Module:       module,
			Endpoint:     endpoint,
			Method:       method,
			Success:      false,
			Error:        fmt.Sprintf("请求失败: %v", err),
			ResponseTime: responseTime,
		}
	}
	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(resp.Body)
	var responseData interface{}
	json.Unmarshal(responseBody, &responseData)

	result := APITestResult{
		Module:       module,
		Endpoint:     endpoint,
		Method:       method,
		StatusCode:   resp.StatusCode,
		Success:      resp.StatusCode >= 200 && resp.StatusCode < 300,
		ResponseTime: responseTime,
		Response:     responseData,
	}

	if !result.Success {
		result.Error = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(responseBody))
	}

	eat.Results = append(eat.Results, result)
	return result
}

// 打印测试结果
func (eat *EnhancedAPITester) PrintResult(result APITestResult) {
	status := "✅"
	if !result.Success {
		status = "❌"
	}

	fmt.Printf("%s [%s] %s %s - %d (%v)\n",
		status, result.Module, result.Method, result.Endpoint,
		result.StatusCode, result.ResponseTime)

	if result.Error != "" {
		fmt.Printf("   错误: %s\n", result.Error)
	}

	if result.Response != nil && result.Success {
		if respMap, ok := result.Response.(map[string]interface{}); ok {
			if message, exists := respMap["message"]; exists {
				fmt.Printf("   响应: %v\n", message)
			}
			if data, exists := respMap["data"]; exists {
				if dataMap, ok := data.(map[string]interface{}); ok {
					if id, exists := dataMap["id"]; exists {
						fmt.Printf("   ID: %v\n", id)
					}
				}
			}
		}
	}
	fmt.Println()
}

// 获取用户Token
func (eat *EnhancedAPITester) GetUserToken() error {
	// 先注册一个测试用户
	timestamp := time.Now().Unix()
	username := "testuser" + strconv.FormatInt(timestamp, 10)

	registerBody := map[string]interface{}{
		"username": username,
		"email":    username + "@example.com",
		"password": "12345678",
		"nickname": "Test User " + strconv.FormatInt(timestamp, 10),
	}

	result := eat.TestRequest("Auth", "POST", "/api/v1/users/register", registerBody, "")
	if !result.Success {
		return fmt.Errorf("用户注册失败: %s", result.Error)
	}

	// 登录获取Token
	loginBody := map[string]interface{}{
		"username": username,
		"password": "12345678",
	}

	result = eat.TestRequest("Auth", "POST", "/api/v1/users/login", loginBody, "")
	if !result.Success {
		return fmt.Errorf("用户登录失败: %s", result.Error)
	}

	if respMap, ok := result.Response.(map[string]interface{}); ok {
		if data, exists := respMap["data"]; exists {
			if dataMap, ok := data.(map[string]interface{}); ok {
				if token, exists := dataMap["token"]; exists {
					eat.UserToken = token.(string)
					fmt.Printf("✅ 获取用户Token成功: %s...\n", eat.UserToken[:30])
					return nil
				}
			}
		}
	}

	return fmt.Errorf("无法从登录响应中提取Token")
}

// 初始化测试数据
func (eat *EnhancedAPITester) InitializeTestData() error {
	fmt.Println("🔧 初始化测试数据...")

	// 1. 添加商品到购物车
	fmt.Println("   添加商品到购物车...")
	cartItem1 := map[string]interface{}{
		"product_id": 1,
		"quantity":   2,
	}

	result := eat.TestRequest("Cart", "POST", "/api/v1/cart/add", cartItem1, eat.UserToken)
	if result.Success {
		fmt.Println("   ✅ 商品1添加到购物车成功")
	} else {
		fmt.Printf("   ⚠️ 商品1添加失败: %s\n", result.Error)
	}

	cartItem2 := map[string]interface{}{
		"product_id": 2,
		"quantity":   1,
	}

	result = eat.TestRequest("Cart", "POST", "/api/v1/cart/add", cartItem2, eat.UserToken)
	if result.Success {
		fmt.Println("   ✅ 商品2添加到购物车成功")
	} else {
		fmt.Printf("   ⚠️ 商品2添加失败: %s\n", result.Error)
	}

	// 2. 获取购物车内容以获取商品项ID
	result = eat.TestRequest("Cart", "GET", "/api/v1/cart", nil, eat.UserToken)
	if result.Success {
		fmt.Println("   ✅ 购物车数据获取成功")
	} else {
		fmt.Printf("   ⚠️ 购物车数据获取失败: %s\n", result.Error)
	}

	fmt.Println("✅ 测试数据初始化完成")
	return nil
}

// 测试订单模块（增强版）
func (eat *EnhancedAPITester) TestOrderModuleEnhanced() {
	fmt.Println("=== 订单管理模块测试（增强版） ===")

	// 测试订单列表
	result := eat.TestRequest("Order", "GET", "/api/v1/orders?page=1&page_size=10", nil, eat.UserToken)
	eat.PrintResult(result)

	// 先获取购物车内容以获取实际的商品项ID
	cartResult := eat.TestRequest("Cart", "GET", "/api/v1/cart", nil, eat.UserToken)

	var cartItemIDs []int
	if cartResult.Success && cartResult.Response != nil {
		if respMap, ok := cartResult.Response.(map[string]interface{}); ok {
			if data, exists := respMap["data"]; exists {
				if dataMap, ok := data.(map[string]interface{}); ok {
					if cart, exists := dataMap["cart"]; exists {
						if cartMap, ok := cart.(map[string]interface{}); ok {
							if items, exists := cartMap["items"]; exists {
								if itemsArray, ok := items.([]interface{}); ok {
									for _, item := range itemsArray {
										if itemMap, ok := item.(map[string]interface{}); ok {
											if id, exists := itemMap["id"]; exists {
												if idFloat, ok := id.(float64); ok {
													cartItemIDs = append(cartItemIDs, int(idFloat))
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	fmt.Printf("   获取到的购物车商品项ID: %v\n", cartItemIDs)

	// 测试创建订单（使用实际的购物车商品项ID）
	orderData := map[string]interface{}{
		"cart_item_ids":    cartItemIDs,
		"receiver_name":    "张三",
		"receiver_phone":   "13800138000",
		"receiver_address": "某某街道123号",
		"province":         "北京市",
		"city":             "北京市",
		"district":         "朝阳区",
	}
	result = eat.TestRequest("Order", "POST", "/api/v1/orders", orderData, eat.UserToken)
	eat.PrintResult(result)
}

// 生成测试报告
func (eat *EnhancedAPITester) GenerateReport() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("🎯 Mall-Go增强版API测试报告")
	fmt.Println(strings.Repeat("=", 80))

	totalTests := len(eat.Results)
	totalSuccess := 0
	totalFailure := 0

	for _, result := range eat.Results {
		if result.Success {
			totalSuccess++
		} else {
			totalFailure++
		}
	}

	fmt.Printf("📊 总体统计:\n")
	fmt.Printf("   总测试数: %d\n", totalTests)
	fmt.Printf("   成功: %d\n", totalSuccess)
	fmt.Printf("   失败: %d\n", totalFailure)
	fmt.Printf("   成功率: %.2f%%\n", float64(totalSuccess)/float64(totalTests)*100)

	// 显示失败的测试
	if totalFailure > 0 {
		fmt.Printf("\n❌ 失败的测试:\n")
		for _, result := range eat.Results {
			if !result.Success {
				fmt.Printf("   - %s %s (HTTP %d)\n", result.Method, result.Endpoint, result.StatusCode)
			}
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("✅ 增强版测试报告生成完成")
	fmt.Println(strings.Repeat("=", 80))
}

func main() {
	fmt.Println("🚀 开始Mall-Go增强版API测试...")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	tester := NewEnhancedAPITester("http://localhost:8081")

	// 健康检查
	fmt.Println("=== 基础健康检查 ===")
	result := tester.TestRequest("System", "GET", "/health", nil, "")
	tester.PrintResult(result)

	if !result.Success {
		fmt.Println("❌ 服务器未运行，请先启动后端服务器")
		return
	}

	// 获取认证Token
	fmt.Println("=== 获取认证Token ===")
	if err := tester.GetUserToken(); err != nil {
		fmt.Printf("❌ 获取用户Token失败: %v\n", err)
		return
	}

	// 初始化测试数据
	if err := tester.InitializeTestData(); err != nil {
		fmt.Printf("❌ 初始化测试数据失败: %v\n", err)
		return
	}

	fmt.Println()
	fmt.Println("🎯 开始增强版API测试...")
	fmt.Println()

	// 测试订单模块
	tester.TestOrderModuleEnhanced()

	// 生成报告
	tester.GenerateReport()

	fmt.Println("🎉 Mall-Go增强版API测试完成！")
}
