package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"mall-go/internal/handler/user"
	"mall-go/internal/model"
	"mall-go/tests/config"
	"mall-go/tests/helpers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestUserRegisterDebug 调试用户注册测试
func TestUserRegisterDebug(t *testing.T) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 初始化测试数据库
	db := config.SetupTestDB()
	defer config.CleanupTestDB(db)

	// 自动迁移测试表
	err := db.AutoMigrate(
		&model.User{},
		&model.Product{},
		&model.Category{},
		&model.Order{},
		&model.OrderItem{},
		&model.CartItem{},
		&model.Cart{},
	)
	assert.NoError(t, err)

	// 初始化测试辅助工具
	helper := helpers.NewTestHelper(db)

	// 初始化Handler
	handler := user.NewHandler(db)

	// 设置路由
	router := gin.New()
	v1 := router.Group("/api/v1")
	{
		users := v1.Group("/users")
		{
			users.POST("/register", handler.Register)
		}
	}

	// 准备测试数据
	requestBody := model.UserRegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		Nickname: "测试用户",
		Role:     "user",
	}

	// 发送请求
	reqBodyBytes, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", "/api/v1/users/register", bytes.NewBuffer(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 打印响应内容用于调试
	t.Logf("Response Status: %d", w.Code)
	t.Logf("Response Body: %s", w.Body.String())

	// 解析响应
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// 打印响应结构
	t.Logf("Response Structure: %+v", response)

	// 检查响应
	if w.Code == http.StatusOK {
		assert.Equal(t, "用户注册成功", response["msg"])
		assert.NotNil(t, response["data"])
	} else {
		t.Logf("Request failed with status %d", w.Code)
		if msg, ok := response["msg"]; ok {
			t.Logf("Error message: %v", msg)
		}
		if err, ok := response["error"]; ok {
			t.Logf("Error: %v", err)
		}
	}

	// 清理
	helper.CleanupTestData()
}
