package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"mall-go/internal/handler/upload"
	"mall-go/internal/model"
	"mall-go/pkg/response"
	uploadPkg "mall-go/pkg/upload"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// UploadIntegrationTestSuite 文件上传集成测试套件
type UploadIntegrationTestSuite struct {
	suite.Suite
	db            *gorm.DB
	router        *gin.Engine
	tempDir       string
	configManager *uploadPkg.ConfigManager
	testUser      *model.User
}

// SetupSuite 设置测试套件
func (suite *UploadIntegrationTestSuite) SetupSuite() {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)
	suite.db = db

	// 自动迁移
	err = db.AutoMigrate(&model.User{}, &model.File{})
	suite.Require().NoError(err)

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "upload_integration_test_")
	suite.Require().NoError(err)
	suite.tempDir = tempDir

	// 创建配置
	config := uploadPkg.DefaultUploadConfig()
	config.Local.UploadPath = tempDir
	config.MaxFileSize = 1024 * 1024 // 1MB
	config.MaxFiles = 5

	// 创建配置管理器
	configManager := &uploadPkg.ConfigManager{}
	configManager = &uploadPkg.ConfigManager{}
	// 简化配置管理器初始化
	suite.configManager = configManager

	// 设置路由
	router := gin.New()
	
	// 添加认证中间件模拟
	router.Use(func(c *gin.Context) {
		// 模拟用户认证
		c.Set("user_id", uint(1))
		c.Set("user_role", "user")
		c.Next()
	})

	api := router.Group("/api/v1")
	err = upload.RegisterRoutes(api, db, configManager)
	suite.Require().NoError(err)

	suite.router = router

	// 创建测试用户
	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
	}
	db.Create(user)
	suite.testUser = user
}

// TearDownSuite 清理测试套件
func (suite *UploadIntegrationTestSuite) TearDownSuite() {
	// 清理临时目录
	os.RemoveAll(suite.tempDir)
}

// TestUploadSingle 测试单文件上传API
func (suite *UploadIntegrationTestSuite) TestUploadSingle() {
	// 创建multipart表单
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加文件
	fileWriter, err := writer.CreateFormFile("file", "test.txt")
	suite.NoError(err)
	fileWriter.Write([]byte("Hello, World!"))

	// 添加其他字段
	writer.WriteField("category", "document")
	writer.WriteField("description", "测试文件上传")
	writer.Close()

	// 发送请求
	req, _ := http.NewRequest("POST", "/api/v1/upload/single", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// 验证响应
	suite.Equal(http.StatusOK, w.Code)

	var resp response.Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	suite.NoError(err)
	suite.Equal(200, resp.Code)
	suite.Equal("文件上传成功", resp.Message)

	// 验证响应数据
	data, ok := resp.Data.(map[string]interface{})
	suite.True(ok)
	suite.NotEmpty(data["file_id"])
	suite.Equal("test.txt", data["original_name"])
	suite.Equal("document", data["category"])
	suite.NotEmpty(data["url"])
}

// TestUploadMultiple 测试多文件上传API
func (suite *UploadIntegrationTestSuite) TestUploadMultiple() {
	// 创建multipart表单
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加多个文件
	files := []struct {
		name    string
		content string
	}{
		{"file1.txt", "Content 1"},
		{"file2.txt", "Content 2"},
		{"file3.txt", "Content 3"},
	}

	for _, file := range files {
		fileWriter, err := writer.CreateFormFile("files", file.name)
		suite.NoError(err)
		fileWriter.Write([]byte(file.content))
	}

	// 添加其他字段
	writer.WriteField("category", "document")
	writer.WriteField("description", "批量上传测试")
	writer.Close()

	// 发送请求
	req, _ := http.NewRequest("POST", "/api/v1/upload/multiple", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// 验证响应
	suite.Equal(http.StatusOK, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	suite.NoError(err)
	suite.Equal(200, resp.Code)

	// 验证响应数据
	data, ok := resp.Data.(map[string]interface{})
	suite.True(ok)
	suite.Equal(float64(3), data["total_files"])
	suite.Equal(float64(3), data["success_count"])
	suite.Equal(float64(0), data["failed_count"])
}

// TestGetFileInfo 测试获取文件信息API
func (suite *UploadIntegrationTestSuite) TestGetFileInfo() {
	// 先上传一个文件
	fileID := suite.uploadTestFile("info_test.txt", "File info test")

	// 获取文件信息
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/upload/file/%d", fileID), nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// 验证响应
	suite.Equal(http.StatusOK, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	suite.NoError(err)
	suite.Equal(200, resp.Code)

	// 验证响应数据
	data, ok := resp.Data.(map[string]interface{})
	suite.True(ok)
	suite.Equal(float64(fileID), data["file_id"])
	suite.Equal("info_test.txt", data["original_name"])
}

// TestListFiles 测试获取文件列表API
func (suite *UploadIntegrationTestSuite) TestListFiles() {
	// 上传多个测试文件
	for i := 0; i < 3; i++ {
		suite.uploadTestFile(fmt.Sprintf("list_test_%d.txt", i), "List test content")
	}

	// 获取文件列表
	req, _ := http.NewRequest("GET", "/api/v1/upload/files?page=1&page_size=10", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// 验证响应
	suite.Equal(http.StatusOK, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	suite.NoError(err)
	suite.Equal(200, resp.Code)

	// 验证响应数据
	data, ok := resp.Data.(map[string]interface{})
	suite.True(ok)
	suite.GreaterOrEqual(data["total"].(float64), float64(3))
	
	files, ok := data["files"].([]interface{})
	suite.True(ok)
	suite.GreaterOrEqual(len(files), 3)
}

// TestDeleteFile 测试删除文件API
func (suite *UploadIntegrationTestSuite) TestDeleteFile() {
	// 先上传一个文件
	fileID := suite.uploadTestFile("delete_test.txt", "Delete me")

	// 删除文件
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/upload/file/%d", fileID), nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// 验证响应
	suite.Equal(http.StatusOK, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	suite.NoError(err)
	suite.Equal(200, resp.Code)
	suite.Equal("文件删除成功", resp.Message)

	// 验证文件已删除
	req, _ = http.NewRequest("GET", fmt.Sprintf("/api/v1/upload/file/%d", fileID), nil)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	suite.Equal(http.StatusNotFound, w.Code)
}

// TestGetUploadConfig 测试获取上传配置API
func (suite *UploadIntegrationTestSuite) TestGetUploadConfig() {
	req, _ := http.NewRequest("GET", "/api/v1/upload/config", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// 验证响应
	suite.Equal(http.StatusOK, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	suite.NoError(err)
	suite.Equal(200, resp.Code)

	// 验证响应数据
	data, ok := resp.Data.(map[string]interface{})
	suite.True(ok)
	suite.NotEmpty(data["storage_type"])
	suite.NotEmpty(data["max_file_size"])
	suite.NotEmpty(data["max_files"])
}

// TestUploadValidation 测试上传验证
func (suite *UploadIntegrationTestSuite) TestUploadValidation() {
	tests := []struct {
		name           string
		filename       string
		content        string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "文件过大",
			filename:       "large.txt",
			content:        strings.Repeat("a", 2*1024*1024), // 2MB
			expectedStatus: http.StatusBadRequest,
			expectedError:  "文件大小超过限制",
		},
		{
			name:           "不支持的文件类型",
			filename:       "test.xyz",
			content:        "unsupported file",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "不支持的文件扩展名",
		},
		{
			name:           "空文件",
			filename:       "empty.txt",
			content:        "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "文件为空",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// 创建multipart表单
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)

			fileWriter, err := writer.CreateFormFile("file", tt.filename)
			suite.NoError(err)
			fileWriter.Write([]byte(tt.content))
			writer.Close()

			// 发送请求
			req, _ := http.NewRequest("POST", "/api/v1/upload/single", &buf)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			// 验证响应
			suite.Equal(tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var resp response.Response
				json.Unmarshal(w.Body.Bytes(), &resp)
				suite.Contains(resp.Message, tt.expectedError)
			}
		})
	}
}

// TestUploadWithoutAuth 测试未认证的上传请求
func (suite *UploadIntegrationTestSuite) TestUploadWithoutAuth() {
	// 创建新的路由器，不包含认证中间件
	router := gin.New()
	api := router.Group("/api/v1")
	upload.RegisterRoutes(api, suite.db, suite.configManager)

	// 创建上传请求
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	fileWriter, _ := writer.CreateFormFile("file", "test.txt")
	fileWriter.Write([]byte("test content"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/v1/upload/single", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 应该返回未认证错误
	suite.Equal(http.StatusUnauthorized, w.Code)
}

// TestConcurrentUploads 测试并发上传
func (suite *UploadIntegrationTestSuite) TestConcurrentUploads() {
	const numUploads = 5
	done := make(chan bool, numUploads)
	results := make([]int, numUploads)

	// 并发上传文件
	for i := 0; i < numUploads; i++ {
		go func(index int) {
			defer func() { done <- true }()

			// 创建上传请求
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)
			fileWriter, _ := writer.CreateFormFile("file", fmt.Sprintf("concurrent_%d.txt", index))
			fileWriter.Write([]byte(fmt.Sprintf("Concurrent upload %d", index)))
			writer.Close()

			req, _ := http.NewRequest("POST", "/api/v1/upload/single", &buf)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			results[index] = w.Code
		}(i)
	}

	// 等待所有上传完成
	for i := 0; i < numUploads; i++ {
		<-done
	}

	// 验证所有上传都成功
	for i, status := range results {
		suite.Equal(http.StatusOK, status, "Upload %d failed", i)
	}
}

// uploadTestFile 辅助方法：上传测试文件
func (suite *UploadIntegrationTestSuite) uploadTestFile(filename, content string) uint {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	fileWriter, err := writer.CreateFormFile("file", filename)
	suite.NoError(err)
	fileWriter.Write([]byte(content))
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/v1/upload/single", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var resp response.Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp.Data.(map[string]interface{})
	
	return uint(data["file_id"].(float64))
}

// 运行上传集成测试套件
func TestUploadIntegrationSuite(t *testing.T) {
	suite.Run(t, new(UploadIntegrationTestSuite))
}

// TestUploadMiddleware 测试上传中间件
func TestUploadMiddleware(t *testing.T) {
	config := uploadPkg.DefaultUploadConfig()
	middleware := upload.NewUploadMiddleware(&uploadPkg.ConfigManager{})

	// 测试速率限制中间件
	rateLimitMiddleware := middleware.RateLimitMiddleware(2, time.Minute)
	assert.NotNil(t, rateLimitMiddleware)

	// 测试验证上传请求中间件
	validateMiddleware := middleware.ValidateUploadRequest()
	assert.NotNil(t, validateMiddleware)
}
