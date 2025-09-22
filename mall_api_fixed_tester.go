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

// 模块测试报告
type ModuleReport struct {
	ModuleName   string          `json:"module_name"`
	TotalTests   int             `json:"total_tests"`
	SuccessCount int             `json:"success_count"`
	FailureCount int             `json:"failure_count"`
	SuccessRate  float64         `json:"success_rate"`
	Results      []APITestResult `json:"results"`
}

// API测试客户端
type MallAPITester struct {
	BaseURL    string
	Client     *http.Client
	UserToken  string
	AdminToken string
	Results    []APITestResult
	Reports    map[string]*ModuleReport
}

// 创建新的API测试器
func NewMallAPITester(baseURL string) *MallAPITester {
	return &MallAPITester{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
		Results: make([]APITestResult, 0),
		Reports: make(map[string]*ModuleReport),
	}
}

// 执行HTTP请求测试
func (mat *MallAPITester) TestRequest(module, method, endpoint string, body interface{}, token string) APITestResult {
	start := time.Now()

	var reqBody io.Reader
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	}

	url := mat.BaseURL + endpoint
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

	resp, err := mat.Client.Do(req)
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

	mat.Results = append(mat.Results, result)
	mat.addToModuleReport(result)
	return result
}

// 添加结果到模块报告
func (mat *MallAPITester) addToModuleReport(result APITestResult) {
	if mat.Reports[result.Module] == nil {
		mat.Reports[result.Module] = &ModuleReport{
			ModuleName: result.Module,
			Results:    make([]APITestResult, 0),
		}
	}

	report := mat.Reports[result.Module]
	report.Results = append(report.Results, result)
	report.TotalTests++
	if result.Success {
		report.SuccessCount++
	} else {
		report.FailureCount++
	}
	report.SuccessRate = float64(report.SuccessCount) / float64(report.TotalTests) * 100
}

// 打印测试结果
func (mat *MallAPITester) PrintResult(result APITestResult) {
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
func (mat *MallAPITester) GetUserToken() error {
	// 先注册一个测试用户
	timestamp := time.Now().Unix()
	username := "testuser" + strconv.FormatInt(timestamp, 10)

	registerBody := map[string]interface{}{
		"username": username,
		"email":    username + "@example.com",
		"password": "12345678",
		"nickname": "Test User " + strconv.FormatInt(timestamp, 10),
	}

	result := mat.TestRequest("Auth", "POST", "/api/v1/users/register", registerBody, "")
	if !result.Success {
		return fmt.Errorf("用户注册失败: %s", result.Error)
	}

	// 登录获取Token
	loginBody := map[string]interface{}{
		"username": username,
		"password": "12345678",
	}

	result = mat.TestRequest("Auth", "POST", "/api/v1/users/login", loginBody, "")
	if !result.Success {
		return fmt.Errorf("用户登录失败: %s", result.Error)
	}

	if respMap, ok := result.Response.(map[string]interface{}); ok {
		if data, exists := respMap["data"]; exists {
			if dataMap, ok := data.(map[string]interface{}); ok {
				if token, exists := dataMap["token"]; exists {
					mat.UserToken = token.(string)
					fmt.Printf("✅ 获取用户Token成功: %s...\n", mat.UserToken[:30])
					return nil
				}
			}
		}
	}

	return fmt.Errorf("无法从登录响应中提取Token")
}

// 获取管理员Token（如果需要）
func (mat *MallAPITester) GetAdminToken() error {
	// 这里可以尝试创建管理员用户或使用预设的管理员账户
	// 暂时使用普通用户Token
	mat.AdminToken = mat.UserToken
	fmt.Printf("✅ 使用用户Token作为管理员Token\n")
	return nil
}

// 1. 商品管理模块测试（修复版）
func (mat *MallAPITester) TestProductModuleFixed() {
	fmt.Println("=== 1. 商品管理模块测试（修复版） ===")

	// 测试商品列表API（添加必需的分页参数）
	result := mat.TestRequest("Product", "GET", "/api/v1/products?page=1&page_size=10", nil, "")
	mat.PrintResult(result)

	// 测试商品详情API（使用ID 1）
	result = mat.TestRequest("Product", "GET", "/api/v1/products/1", nil, "")
	mat.PrintResult(result)

	// 测试商品创建API（需要管理员权限）
	productData := map[string]interface{}{
		"name":        "测试商品",
		"description": "这是一个测试商品",
		"price":       99.99,
		"stock":       100,
		"category_id": 1,
	}
	result = mat.TestRequest("Product", "POST", "/api/v1/products", productData, mat.AdminToken)
	mat.PrintResult(result)
}

// 2. 购物车模块测试（修复版）
func (mat *MallAPITester) TestCartModuleFixed() {
	fmt.Println("=== 2. 购物车模块测试（修复版） ===")

	// 测试获取购物车
	result := mat.TestRequest("Cart", "GET", "/api/v1/cart", nil, mat.UserToken)
	mat.PrintResult(result)

	// 测试购物车商品数量
	result = mat.TestRequest("Cart", "GET", "/api/v1/cart/count", nil, mat.UserToken)
	mat.PrintResult(result)

	// 测试清空购物车
	result = mat.TestRequest("Cart", "DELETE", "/api/v1/cart/clear", nil, mat.UserToken)
	mat.PrintResult(result)
}

// 3. 订单管理模块测试（修复版）
func (mat *MallAPITester) TestOrderModuleFixed() {
	fmt.Println("=== 3. 订单管理模块测试（修复版） ===")

	// 测试订单列表（添加必需的分页参数）
	result := mat.TestRequest("Order", "GET", "/api/v1/orders?page=1&page_size=10", nil, mat.UserToken)
	mat.PrintResult(result)

	// 测试创建订单（使用正确的参数格式）
	orderData := map[string]interface{}{
		"cart_item_ids":    []int{1, 2},
		"receiver_name":    "张三",
		"receiver_phone":   "13800138000",
		"receiver_address": "某某街道123号",
		"province":         "北京市",
		"city":             "北京市",
		"district":         "朝阳区",
	}
	result = mat.TestRequest("Order", "POST", "/api/v1/orders", orderData, mat.UserToken)
	mat.PrintResult(result)
}

// 4. 支付模块测试（修复版）
func (mat *MallAPITester) TestPaymentModuleFixed() {
	fmt.Println("=== 4. 支付模块测试（修复版） ===")

	// 测试支付列表（添加分页参数）
	result := mat.TestRequest("Payment", "GET", "/api/v1/payments?page=1&page_size=10", nil, mat.UserToken)
	mat.PrintResult(result)

	// 测试查询支付状态
	result = mat.TestRequest("Payment", "GET", "/api/v1/payments/query?order_id=1", nil, mat.UserToken)
	mat.PrintResult(result)
}

// 5. 地址管理模块测试（修复版）
func (mat *MallAPITester) TestAddressModuleFixed() {
	fmt.Println("=== 5. 地址管理模块测试（修复版） ===")

	// 测试地址列表
	result := mat.TestRequest("Address", "GET", "/api/v1/addresses", nil, mat.UserToken)
	mat.PrintResult(result)

	// 测试获取地区数据（添加认证Token）
	result = mat.TestRequest("Address", "GET", "/api/v1/addresses/regions", nil, mat.UserToken)
	mat.PrintResult(result)
}

// 生成最终报告
func (mat *MallAPITester) GenerateFinalReport() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("🎯 Mall-Go电商系统API联调测试修复版报告")
	fmt.Println(strings.Repeat("=", 80))

	totalTests := len(mat.Results)
	totalSuccess := 0
	totalFailure := 0
	totalTime := time.Duration(0)

	for _, result := range mat.Results {
		if result.Success {
			totalSuccess++
		} else {
			totalFailure++
		}
		totalTime += result.ResponseTime
	}

	fmt.Printf("📊 总体统计:\n")
	fmt.Printf("   总测试数: %d\n", totalTests)
	fmt.Printf("   成功: %d\n", totalSuccess)
	fmt.Printf("   失败: %d\n", totalFailure)
	fmt.Printf("   成功率: %.2f%%\n", float64(totalSuccess)/float64(totalTests)*100)
	fmt.Printf("   平均响应时间: %v\n", totalTime/time.Duration(totalTests))
	fmt.Println()

	// 按模块生成报告
	fmt.Println("📋 各模块详细报告:")
	for moduleName, report := range mat.Reports {
		fmt.Printf("\n🔸 %s模块:\n", moduleName)
		fmt.Printf("   测试数: %d | 成功: %d | 失败: %d | 成功率: %.2f%%\n",
			report.TotalTests, report.SuccessCount, report.FailureCount, report.SuccessRate)

		// 显示失败的测试
		if report.FailureCount > 0 {
			fmt.Printf("   ❌ 失败的测试:\n")
			for _, result := range report.Results {
				if !result.Success {
					fmt.Printf("      - %s %s (HTTP %d)\n", result.Method, result.Endpoint, result.StatusCode)
				}
			}
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("✅ 修复版测试报告生成完成")
	fmt.Println(strings.Repeat("=", 80))
}

func main() {
	fmt.Println("🚀 开始Mall-Go电商系统修复版API联调测试...")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	tester := NewMallAPITester("http://localhost:8081")

	// 首先进行基础健康检查
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

	if err := tester.GetAdminToken(); err != nil {
		fmt.Printf("❌ 获取管理员Token失败: %v\n", err)
		return
	}

	fmt.Println()
	fmt.Println("🎯 开始各模块API修复版测试...")
	fmt.Println()

	// 1. 商品管理模块测试（修复版）
	tester.TestProductModuleFixed()

	// 2. 购物车模块测试（修复版）
	tester.TestCartModuleFixed()

	// 3. 订单管理模块测试（修复版）
	tester.TestOrderModuleFixed()

	// 4. 支付模块测试（修复版）
	tester.TestPaymentModuleFixed()

	// 5. 地址管理模块测试（修复版）
	tester.TestAddressModuleFixed()

	// 生成最终报告
	tester.GenerateFinalReport()

	fmt.Println("🎉 Mall-Go电商系统修复版API联调测试完成！")
}
