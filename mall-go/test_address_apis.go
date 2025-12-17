//go:build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"time"

	"mall-go/internal/handler"
	"mall-go/internal/model"
	"mall-go/pkg/payment"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// 地址管理API测试程序
func main() {
	// 创建内存数据库用于测试
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 自动迁移数据库表
	if err := db.AutoMigrate(&model.User{}, &model.Address{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// 创建测试用户
	testUser := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		Role:     model.RoleUser,
	}
	if err := db.Create(testUser).Error; err != nil {
		log.Fatal("Failed to create test user:", err)
	}

	// 创建Redis客户端（可以为nil）
	var rdb *redis.Client = nil

	// 创建支付服务
	paymentService := payment.NewService(db)

	// 创建Gin引擎
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 注册中间件
	handler.RegisterMiddleware(r)

	// 注册路由
	handler.RegisterRoutes(r, db, rdb, paymentService)

	fmt.Println("=== 地址管理API测试开始 ===")

	// 模拟JWT token（实际应用中需要真实的JWT）
	token := "Bearer test-jwt-token"

	// 测试1: 创建地址
	fmt.Println("\n1. 测试创建地址")
	createAddressReq := model.AddressCreateRequest{
		ReceiverName:  "张三",
		ReceiverPhone: "13800138000",
		Province:      "广东省",
		City:          "深圳市",
		District:      "南山区",
		DetailAddress: "科技园南区深南大道10000号",
		PostalCode:    "518000",
		IsDefault:     true,
		AddressType:   model.AddressTypeHome,
	}
	
	addressID := testCreateAddress(r, createAddressReq, token)

	// 测试2: 获取地址列表
	fmt.Println("\n2. 测试获取地址列表")
	testGetAddresses(r, token)

	// 测试3: 更新地址
	fmt.Println("\n3. 测试更新地址")
	updateAddressReq := model.AddressUpdateRequest{
		ReceiverName:  "李四",
		ReceiverPhone: "13900139000",
		Province:      "北京市",
		City:          "北京市",
		District:      "朝阳区",
		DetailAddress: "三里屯SOHO",
		PostalCode:    "100000",
	}
	testUpdateAddress(r, addressID, updateAddressReq, token)

	// 测试4: 设置默认地址
	fmt.Println("\n4. 测试设置默认地址")
	testSetDefaultAddress(r, addressID, token)

	// 测试5: 创建第二个地址
	fmt.Println("\n5. 测试创建第二个地址")
	createAddressReq2 := model.AddressCreateRequest{
		ReceiverName:  "王五",
		ReceiverPhone: "13700137000",
		Province:      "上海市",
		City:          "上海市",
		District:      "浦东新区",
		DetailAddress: "陆家嘴金融中心",
		PostalCode:    "200000",
		IsDefault:     false,
		AddressType:   model.AddressTypeCompany,
	}
	addressID2 := testCreateAddress(r, createAddressReq2, token)

	// 测试6: 再次获取地址列表
	fmt.Println("\n6. 测试获取更新后的地址列表")
	testGetAddresses(r, token)

	// 测试7: 获取地区数据
	fmt.Println("\n7. 测试获取地区数据")
	testGetRegions(r)

	// 测试8: 删除地址
	fmt.Println("\n8. 测试删除地址")
	testDeleteAddress(r, addressID2, token)

	// 测试9: 最终地址列表
	fmt.Println("\n9. 测试删除后的地址列表")
	testGetAddresses(r, token)

	fmt.Println("\n=== 地址管理API测试完成 ===")
}

// 测试创建地址
func testCreateAddress(r *gin.Engine, req model.AddressCreateRequest, token string) uint {
	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/api/v1/addresses", bytes.NewBuffer(reqBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", token)
	
	r.ServeHTTP(w, request)
	
	fmt.Printf("状态码: %d\n", w.Code)
	body, _ := io.ReadAll(w.Body)
	fmt.Printf("响应: %s\n", string(body))
	
	// 解析响应获取地址ID
	var response struct {
		Data struct {
			ID uint `json:"id"`
		} `json:"data"`
	}
	json.Unmarshal(body, &response)
	return response.Data.ID
}

// 测试获取地址列表
func testGetAddresses(r *gin.Engine, token string) {
	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/api/v1/addresses", nil)
	request.Header.Set("Authorization", token)
	
	r.ServeHTTP(w, request)
	
	fmt.Printf("状态码: %d\n", w.Code)
	body, _ := io.ReadAll(w.Body)
	fmt.Printf("响应: %s\n", string(body))
}

// 测试更新地址
func testUpdateAddress(r *gin.Engine, addressID uint, req model.AddressUpdateRequest, token string) {
	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	request, _ := http.NewRequest("PUT", fmt.Sprintf("/api/v1/addresses/%d", addressID), bytes.NewBuffer(reqBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", token)
	
	r.ServeHTTP(w, request)
	
	fmt.Printf("状态码: %d\n", w.Code)
	body, _ := io.ReadAll(w.Body)
	fmt.Printf("响应: %s\n", string(body))
}

// 测试设置默认地址
func testSetDefaultAddress(r *gin.Engine, addressID uint, token string) {
	w := httptest.NewRecorder()
	request, _ := http.NewRequest("PUT", fmt.Sprintf("/api/v1/addresses/%d/default", addressID), nil)
	request.Header.Set("Authorization", token)
	
	r.ServeHTTP(w, request)
	
	fmt.Printf("状态码: %d\n", w.Code)
	body, _ := io.ReadAll(w.Body)
	fmt.Printf("响应: %s\n", string(body))
}

// 测试删除地址
func testDeleteAddress(r *gin.Engine, addressID uint, token string) {
	w := httptest.NewRecorder()
	request, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/addresses/%d", addressID), nil)
	request.Header.Set("Authorization", token)
	
	r.ServeHTTP(w, request)
	
	fmt.Printf("状态码: %d\n", w.Code)
	body, _ := io.ReadAll(w.Body)
	fmt.Printf("响应: %s\n", string(body))
}

// 测试获取地区数据
func testGetRegions(r *gin.Engine) {
	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/api/v1/addresses/regions", nil)
	
	r.ServeHTTP(w, request)
	
	fmt.Printf("状态码: %d\n", w.Code)
	body, _ := io.ReadAll(w.Body)
	fmt.Printf("响应: %s\n", string(body))
}
