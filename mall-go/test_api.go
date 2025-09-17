package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	fmt.Println("🚀 开始测试Mall-Go API...")

	// 等待服务器启动
	time.Sleep(2 * time.Second)

	// 测试健康检查
	fmt.Println("📋 测试健康检查API...")
	testHealthCheck()

	// 测试商品列表
	fmt.Println("📦 测试商品列表API...")
	testProductList()

	// 测试商品详情
	fmt.Println("🔍 测试商品详情API...")
	testProductDetail()

	fmt.Println("✅ API测试完成!")
}

func testHealthCheck() {
	resp, err := http.Get("http://localhost:8080/health")
	if err != nil {
		fmt.Printf("❌ 健康检查失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 健康检查成功: %s\n", string(body))
}

func testProductList() {
	resp, err := http.Get("http://localhost:8080/api/v1/products?page=1&page_size=5")
	if err != nil {
		fmt.Printf("❌ 商品列表请求失败: %v\n", err)
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

	fmt.Printf("✅ 商品列表API成功\n")
	fmt.Printf("   响应码: %.0f\n", result["code"])
	fmt.Printf("   消息: %s\n", result["message"])

	if data, ok := result["data"].(map[string]interface{}); ok {
		if list, ok := data["list"].([]interface{}); ok {
			fmt.Printf("   商品数量: %d\n", len(list))
			if len(list) > 0 {
				if product, ok := list[0].(map[string]interface{}); ok {
					fmt.Printf("   第一个商品: %s\n", product["name"])
				}
			}
		}
		if total, ok := data["total"]; ok {
			fmt.Printf("   总商品数: %.0f\n", total)
		}
	}
}

func testProductDetail() {
	resp, err := http.Get("http://localhost:8080/api/v1/products/1")
	if err != nil {
		fmt.Printf("❌ 商品详情请求失败: %v\n", err)
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

	fmt.Printf("✅ 商品详情API成功\n")
	fmt.Printf("   响应码: %.0f\n", result["code"])
	fmt.Printf("   消息: %s\n", result["message"])

	if data, ok := result["data"].(map[string]interface{}); ok {
		if name, ok := data["name"]; ok {
			fmt.Printf("   商品名称: %s\n", name)
		}
		if price, ok := data["price"]; ok {
			fmt.Printf("   商品价格: %s\n", price)
		}
		if stock, ok := data["stock"]; ok {
			fmt.Printf("   库存数量: %.0f\n", stock)
		}
	}
}
