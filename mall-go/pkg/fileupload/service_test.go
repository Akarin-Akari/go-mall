package fileupload

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"testing"

	"mall-go/internal/config"
	"mall-go/internal/model"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// 初始化测试配置
func init() {
	config.GlobalConfig = config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key-for-file-upload-testing",
			Expire: "24h",
		},
	}
}

// setupTestDB 设置测试数据库
func setupTestDB(t *testing.T) *gorm.DB {
	// 使用内存SQLite数据库进行测试
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// 自动迁移表结构
	err = db.AutoMigrate(&model.File{}, &model.User{})
	assert.NoError(t, err)

	return db
}

// createTestFile 创建测试文件
func createTestFile(t *testing.T, filename, content string) *multipart.FileHeader {
	// 创建一个buffer来模拟文件上传
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 创建文件字段
	part, err := writer.CreateFormFile("file", filename)
	assert.NoError(t, err)

	// 写入文件内容
	_, err = part.Write([]byte(content))
	assert.NoError(t, err)

	// 关闭writer
	err = writer.Close()
	assert.NoError(t, err)

	// 解析multipart数据
	reader := multipart.NewReader(body, writer.Boundary())
	form, err := reader.ReadForm(32 << 20) // 32MB
	assert.NoError(t, err)

	// 返回第一个文件
	files := form.File["file"]
	assert.Len(t, files, 1)
	return files[0]
}

// TestNewFileUploadService 测试创建文件上传服务
func TestNewFileUploadService(t *testing.T) {
	db := setupTestDB(t)
	uploadPath := "/tmp/test-uploads"
	baseURL := "http://localhost:8080"

	service := NewFileUploadService(db, uploadPath, baseURL)

	assert.NotNil(t, service)
	assert.Equal(t, db, service.db)
	assert.Equal(t, uploadPath, service.uploadPath)
	assert.Equal(t, baseURL, service.baseURL)
}

// TestUploadFile 测试单文件上传
func TestUploadFile(t *testing.T) {
	db := setupTestDB(t)

	// 创建临时上传目录
	tempDir := t.TempDir()
	service := NewFileUploadService(db, tempDir, "http://localhost:8080")

	// 创建测试文件
	fileHeader := createTestFile(t, "test.txt", "Hello, World!")
	fileHeader.Header.Set("Content-Type", "text/plain")

	// 上传文件
	result, err := service.UploadFile(
		fileHeader,
		1, // userID
		model.BusinessTypeOther,
		nil,  // businessID
		true, // isPublic
	)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test.txt", result.OriginalName)
	assert.Equal(t, int64(13), result.FileSize) // "Hello, World!" 长度
	assert.Equal(t, "text/plain", result.MimeType)
	assert.Equal(t, model.FileTypeDocument, result.FileType) // text/plain 被识别为文档类型
	assert.Equal(t, model.BusinessTypeOther, result.BusinessType)
	assert.True(t, result.IsPublic)
	assert.Equal(t, model.FileStatusSuccess, result.Status)
	assert.NotEmpty(t, result.UUID)
	assert.NotEmpty(t, result.AccessURL)

	// 验证数据库记录
	var fileRecord model.File
	err = db.Where("uuid = ?", result.UUID).First(&fileRecord).Error
	assert.NoError(t, err)
	assert.Equal(t, result.UUID, fileRecord.UUID)
	assert.Equal(t, model.FileStatusSuccess, fileRecord.Status)

	// 验证文件是否实际保存
	assert.FileExists(t, fileRecord.FilePath)
}

// TestUploadMultipleFiles 测试多文件上传
func TestUploadMultipleFiles(t *testing.T) {
	db := setupTestDB(t)

	// 创建临时上传目录
	tempDir := t.TempDir()
	service := NewFileUploadService(db, tempDir, "http://localhost:8080")

	// 创建多个测试文件
	fileHeaders := []*multipart.FileHeader{
		createTestFile(t, "test1.txt", "Content 1"),
		createTestFile(t, "test2.txt", "Content 2"),
	}

	// 设置Content-Type
	for _, fh := range fileHeaders {
		fh.Header.Set("Content-Type", "text/plain")
	}

	// 上传文件
	results, err := service.UploadMultipleFiles(
		fileHeaders,
		1, // userID
		model.BusinessTypeOther,
		nil,   // businessID
		false, // isPublic
	)

	assert.NoError(t, err)
	assert.Len(t, results, 2)

	// 验证每个文件
	for i, result := range results {
		assert.Equal(t, fileHeaders[i].Filename, result.OriginalName)
		assert.Equal(t, model.FileStatusSuccess, result.Status)
		assert.False(t, result.IsPublic)

		// 验证文件存在
		var fileRecord model.File
		err = db.Where("uuid = ?", result.UUID).First(&fileRecord).Error
		assert.NoError(t, err)
		assert.FileExists(t, fileRecord.FilePath)
	}
}

// TestUploadFileExceedsLimit 测试文件大小超限
func TestUploadFileExceedsLimit(t *testing.T) {
	db := setupTestDB(t)

	tempDir := t.TempDir()
	service := NewFileUploadService(db, tempDir, "http://localhost:8080")

	// 创建一个超大的测试文件内容
	largeContent := make([]byte, model.MaxImageSize+1)
	for i := range largeContent {
		largeContent[i] = 'A'
	}

	fileHeader := createTestFile(t, "large.jpg", string(largeContent))
	fileHeader.Header.Set("Content-Type", "image/jpeg")

	// 尝试上传超大文件
	result, err := service.UploadFile(
		fileHeader,
		1,
		model.BusinessTypeProduct,
		nil,
		true,
	)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "文件验证失败")
}

// TestGetFileByUUID 测试根据UUID获取文件
func TestGetFileByUUID(t *testing.T) {
	db := setupTestDB(t)

	tempDir := t.TempDir()
	service := NewFileUploadService(db, tempDir, "http://localhost:8080")

	// 先上传一个文件
	fileHeader := createTestFile(t, "test.txt", "Test content")
	fileHeader.Header.Set("Content-Type", "text/plain")

	result, err := service.UploadFile(
		fileHeader,
		1,
		model.BusinessTypeOther,
		nil,
		true,
	)
	assert.NoError(t, err)

	// 根据UUID获取文件
	file, err := service.GetFileByUUID(result.UUID)
	assert.NoError(t, err)
	assert.NotNil(t, file)
	assert.Equal(t, result.UUID, file.UUID)
	assert.Equal(t, "test.txt", file.OriginalName)

	// 测试不存在的UUID
	_, err = service.GetFileByUUID("non-existent-uuid")
	assert.Error(t, err)
}

// TestGetFilesByUser 测试获取用户文件列表
func TestGetFilesByUser(t *testing.T) {
	db := setupTestDB(t)

	tempDir := t.TempDir()
	service := NewFileUploadService(db, tempDir, "http://localhost:8080")

	// 上传多个文件
	for i := 0; i < 3; i++ {
		fileHeader := createTestFile(t, fmt.Sprintf("test%d.txt", i), fmt.Sprintf("Content %d", i))
		fileHeader.Header.Set("Content-Type", "text/plain")

		_, err := service.UploadFile(
			fileHeader,
			1, // 同一个用户
			model.BusinessTypeOther,
			nil,
			true,
		)
		assert.NoError(t, err)
	}

	// 获取用户文件列表
	files, total, err := service.GetFilesByUser(
		1,   // userID
		"",  // businessType (空表示所有)
		nil, // businessID
		1,   // page
		10,  // pageSize
	)

	assert.NoError(t, err)
	assert.Len(t, files, 3)
	assert.Equal(t, int64(3), total)

	// 测试按业务类型过滤
	files, total, err = service.GetFilesByUser(
		1,
		model.BusinessTypeOther,
		nil,
		1,
		10,
	)

	assert.NoError(t, err)
	assert.Len(t, files, 3)
	assert.Equal(t, int64(3), total)

	// 测试不存在的业务类型
	files, total, err = service.GetFilesByUser(
		1,
		model.BusinessTypeProduct,
		nil,
		1,
		10,
	)

	assert.NoError(t, err)
	assert.Len(t, files, 0)
	assert.Equal(t, int64(0), total)
}

// TestDeleteFile 测试删除文件
func TestDeleteFile(t *testing.T) {
	db := setupTestDB(t)

	tempDir := t.TempDir()
	service := NewFileUploadService(db, tempDir, "http://localhost:8080")

	// 先上传一个文件
	fileHeader := createTestFile(t, "test.txt", "Test content")
	fileHeader.Header.Set("Content-Type", "text/plain")

	result, err := service.UploadFile(
		fileHeader,
		1,
		model.BusinessTypeOther,
		nil,
		true,
	)
	assert.NoError(t, err)

	// 验证文件存在
	var fileRecord model.File
	err = db.Where("uuid = ?", result.UUID).First(&fileRecord).Error
	assert.NoError(t, err)
	assert.FileExists(t, fileRecord.FilePath)

	// 删除文件
	err = service.DeleteFile(result.UUID, 1)
	assert.NoError(t, err)

	// 验证文件状态已更新
	err = db.Where("uuid = ?", result.UUID).First(&fileRecord).Error
	assert.NoError(t, err)
	assert.Equal(t, model.FileStatusDeleted, fileRecord.Status)

	// 测试删除不存在的文件
	err = service.DeleteFile("non-existent-uuid", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "文件不存在或无权限删除")

	// 测试无权限删除其他用户的文件
	err = service.DeleteFile(result.UUID, 2) // 不同的用户ID
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "文件不存在或无权限删除")
}

// TestGenerateStoragePath 测试存储路径生成
func TestGenerateStoragePath(t *testing.T) {
	db := setupTestDB(t)

	tempDir := t.TempDir()
	service := NewFileUploadService(db, tempDir, "http://localhost:8080")

	file := &model.File{
		UUID:         "test-uuid-123",
		OriginalName: "test.jpg",
		FileType:     model.FileTypeImage,
		BusinessType: model.BusinessTypeProduct,
	}

	storedName, filePath, err := service.generateStoragePath(file)
	assert.NoError(t, err)
	assert.Equal(t, "test-uuid-123.jpg", storedName)
	assert.Contains(t, filePath, "images")
	assert.Contains(t, filePath, "product")
	assert.Contains(t, filePath, "test-uuid-123.jpg")

	// 验证目录已创建
	dir := filepath.Dir(filePath)
	assert.DirExists(t, dir)
}

// TestGenerateAccessURL 测试访问URL生成
func TestGenerateAccessURL(t *testing.T) {
	db := setupTestDB(t)
	service := NewFileUploadService(db, "/tmp", "http://localhost:8080")

	// 测试公开文件URL
	publicFile := &model.File{
		UUID:     "public-uuid",
		IsPublic: true,
	}
	url := service.generateAccessURL(publicFile)
	assert.Equal(t, "http://localhost:8080/api/v1/files/public/public-uuid", url)

	// 测试私有文件URL
	privateFile := &model.File{
		UUID:     "private-uuid",
		IsPublic: false,
	}
	url = service.generateAccessURL(privateFile)
	assert.Equal(t, "http://localhost:8080/api/v1/files/private/private-uuid", url)
}
