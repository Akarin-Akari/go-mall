//go:build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// API测试结果结构
type TestResult struct {
	Endpoint     string        `json:"endpoint"`
	Method       string        `json:"method"`
	StatusCode   int           `json:"status_code"`
	Success      bool          `json:"success"`
	ResponseTime time.Duration `json:"response_time"`
	Error        string        `json:"error,omitempty"`
	Response     interface{}   `json:"response,omitempty"`
}

// API测试客户端
type APITester struct {
	BaseURL string
	Client  *http.Client
	Results []TestResult
}

// 创建新的API测试器
func NewAPITester(baseURL string) *APITester {
	return &APITester{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
		Results: make([]TestResult, 0),
	}
}

// 执行HTTP请求测试
func (at *APITester) TestRequest(method, endpoint string, body interface{}) TestResult {
	start := time.Now()

	var reqBody io.Reader
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	}

	url := at.BaseURL + endpoint
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return TestResult{
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

	resp, err := at.Client.Do(req)
	responseTime := time.Since(start)

	if err != nil {
		return TestResult{
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

	result := TestResult{
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

	at.Results = append(at.Results, result)
	return result
}

// 打印测试结果
func (at *APITester) PrintResult(result TestResult) {
	status := "✅"
	if !result.Success {
		status = "❌"
	}

	fmt.Printf("%s %s %s - %d (%v)\n",
		status, result.Method, result.Endpoint,
		result.StatusCode, result.ResponseTime)

	if result.Error != "" {
		fmt.Printf("   错误: %s\n", result.Error)
	}

	if result.Response != nil && result.Success {
		if respMap, ok := result.Response.(map[string]interface{}); ok {
			if message, exists := respMap["message"]; exists {
				fmt.Printf("   响应: %v\n", message)
			}
		}
	}
	fmt.Println()
}

// 生成测试报告
func (at *APITester) GenerateReport() {
	fmt.Println("=== API测试报告 ===")
	fmt.Printf("总测试数: %d\n", len(at.Results))

	successCount := 0
	totalTime := time.Duration(0)

	for _, result := range at.Results {
		if result.Success {
			successCount++
		}
		totalTime += result.ResponseTime
	}

	fmt.Printf("成功: %d\n", successCount)
	fmt.Printf("失败: %d\n", len(at.Results)-successCount)
	fmt.Printf("成功率: %.2f%%\n", float64(successCount)/float64(len(at.Results))*100)
	fmt.Printf("平均响应时间: %v\n", totalTime/time.Duration(len(at.Results)))
	fmt.Println()
}

func main() {
	fmt.Println("🚀 开始Mall-Go API全面测试...")
	fmt.Println()

	tester := NewAPITester("http://localhost:8080")

	// 1. 基础健康检查测试
	fmt.Println("=== 1. 基础健康检查测试 ===")
	result := tester.TestRequest("GET", "/health", nil)
	tester.PrintResult(result)

	result = tester.TestRequest("GET", "/", nil)
	tester.PrintResult(result)

	// 2. 产品API测试
	fmt.Println("=== 2. 产品API测试 ===")
	result = tester.TestRequest("GET", "/api/v1/products", nil)
	tester.PrintResult(result)

	// 3. 用户API测试（如果存在）
	fmt.Println("=== 3. 用户API测试 ===")
	result = tester.TestRequest("GET", "/api/v1/users", nil)
	tester.PrintResult(result)

	// 4. 订单API测试（如果存在）
	fmt.Println("=== 4. 订单API测试 ===")
	result = tester.TestRequest("GET", "/api/v1/orders", nil)
	tester.PrintResult(result)

	// 5. 支付API测试（如果存在）
	fmt.Println("=== 5. 支付API测试 ===")
	result = tester.TestRequest("GET", "/api/v1/payments", nil)
	tester.PrintResult(result)

	// 6. 文件上传API测试（如果存在）
	fmt.Println("=== 6. 文件上传API测试 ===")
	result = tester.TestRequest("GET", "/api/v1/files", nil)
	tester.PrintResult(result)

	// 7. 错误处理测试
	fmt.Println("=== 7. 错误处理测试 ===")
	result = tester.TestRequest("GET", "/api/v1/nonexistent", nil)
	tester.PrintResult(result)

	result = tester.TestRequest("POST", "/api/v1/products", map[string]interface{}{
		"invalid": "data",
	})
	tester.PrintResult(result)

	// 生成最终报告
	tester.GenerateReport()

	fmt.Println("✅ API测试完成！")
}
