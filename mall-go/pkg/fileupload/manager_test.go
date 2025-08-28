package fileupload

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"testing"

	"mall-go/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupManagerTestDB 设置管理器测试数据库
func setupManagerTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// 自动迁移
	err = db.AutoMigrate(&model.File{})
	require.NoError(t, err)

	return db
}

// createManagerTestFile 创建管理器测试文件
func createManagerTestFile(t *testing.T, filename, content string) *multipart.FileHeader {
	// 创建multipart writer
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 创建文件字段
	part, err := writer.CreateFormFile("file", filename)
	require.NoError(t, err)

	// 写入内容
	_, err = io.WriteString(part, content)
	require.NoError(t, err)

	// 关闭writer
	err = writer.Close()
	require.NoError(t, err)

	// 解析multipart
	reader := multipart.NewReader(body, writer.Boundary())
	form, err := reader.ReadForm(1024 * 1024) // 1MB
	require.NoError(t, err)

	files := form.File["file"]
	require.Len(t, files, 1)

	return files[0]
}

// TestUploadManager_NewUploadManager 测试创建上传管理器
func TestUploadManager_NewUploadManager(t *testing.T) {
	db := setupManagerTestDB(t)

	// 测试使用默认配置
	manager := NewUploadManager(db, nil)
	assert.NotNil(t, manager)
	assert.NotNil(t, manager.config)
	assert.NotNil(t, manager.uploadService)
	assert.NotNil(t, manager.chunkService)
	assert.NotNil(t, manager.validator)

	// 测试使用自定义配置
	config := DefaultUploadConfig()
	config.UploadPath = "/custom/path"
	config.BaseURL = "http://custom.com"

	manager2 := NewUploadManager(db, config)
	assert.NotNil(t, manager2)
	assert.Equal(t, "/custom/path", manager2.config.UploadPath)
	assert.Equal(t, "http://custom.com", manager2.config.BaseURL)

	manager.Stop()
	manager2.Stop()
}

// TestUploadManager_UploadSingleFile 测试单文件上传
func TestUploadManager_UploadSingleFile(t *testing.T) {
	db := setupManagerTestDB(t)
	tempDir := t.TempDir()

	config := DefaultUploadConfig()
	config.UploadPath = tempDir
	config.BaseURL = "http://localhost:8080"

	manager := NewUploadManager(db, config)
	defer manager.Stop()

	// 创建测试文件
	fileHeader := createManagerTestFile(t, "test.txt", "Hello, Manager!")

	// 执行上传
	response, err := manager.UploadSingleFile(
		fileHeader,
		1, // userID
		model.BusinessTypeProduct,
		nil,
		true,
	)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "test.txt", response.OriginalName)
	assert.Equal(t, int64(15), response.FileSize) // "Hello, Manager!" 长度
	assert.Equal(t, model.FileStatusSuccess, response.Status)
}

// TestUploadManager_UploadMultipleFiles 测试多文件上传
func TestUploadManager_UploadMultipleFiles(t *testing.T) {
	db := setupManagerTestDB(t)
	tempDir := t.TempDir()

	config := DefaultUploadConfig()
	config.UploadPath = tempDir
	config.BaseURL = "http://localhost:8080"

	manager := NewUploadManager(db, config)
	defer manager.Stop()

	// 创建多个测试文件
	fileHeaders := []*multipart.FileHeader{
		createManagerTestFile(t, "file1.txt", "Content 1"),
		createManagerTestFile(t, "file2.txt", "Content 2"),
	}

	// 执行批量上传
	responses, err := manager.UploadMultipleFiles(
		fileHeaders,
		1, // userID
		model.BusinessTypeProduct,
		nil,
		true,
	)

	// 验证结果
	assert.NoError(t, err)
	assert.Len(t, responses, 2)

	for i, response := range responses {
		assert.Equal(t, fileHeaders[i].Filename, response.OriginalName)
		assert.Equal(t, model.FileStatusSuccess, response.Status)
	}
}

// TestUploadManager_ValidateFile 测试文件验证
func TestUploadManager_ValidateFile(t *testing.T) {
	db := setupManagerTestDB(t)
	tempDir := t.TempDir()

	config := DefaultUploadConfig()
	config.UploadPath = tempDir

	manager := NewUploadManager(db, config)
	defer manager.Stop()

	tests := []struct {
		name        string
		filename    string
		content     string
		expectError bool
	}{
		{
			name:        "valid text file",
			filename:    "test.txt",
			content:     "Hello, World!",
			expectError: false,
		},
		{
			name:        "valid image file",
			filename:    "test.jpg",
			content:     "fake image content",
			expectError: false,
		},
		{
			name:        "dangerous file",
			filename:    "malware.exe",
			content:     "dangerous content",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileHeader := createManagerTestFile(t, tt.filename, tt.content)
			err := manager.ValidateFile(fileHeader)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestUploadManager_GetFileInfo 测试获取文件信息
func TestUploadManager_GetFileInfo(t *testing.T) {
	db := setupManagerTestDB(t)
	tempDir := t.TempDir()

	config := DefaultUploadConfig()
	config.UploadPath = tempDir
	config.BaseURL = "http://localhost:8080"

	manager := NewUploadManager(db, config)
	defer manager.Stop()

	// 先上传一个文件
	fileHeader := createManagerTestFile(t, "test.txt", "Hello, World!")
	response, err := manager.UploadSingleFile(
		fileHeader,
		1, // userID
		model.BusinessTypeProduct,
		nil,
		true,
	)
	require.NoError(t, err)

	// 获取文件信息
	fileInfo, err := manager.GetFileInfo(response.UUID, 1)
	assert.NoError(t, err)
	assert.Equal(t, response.OriginalName, fileInfo.OriginalName)
	assert.Equal(t, response.FileSize, fileInfo.FileSize)
	assert.Equal(t, response.UUID, fileInfo.UUID)

	// 测试获取不存在的文件
	_, err = manager.GetFileInfo("non-existent-uuid", 1)
	assert.Error(t, err)

	// 测试获取其他用户的文件
	_, err = manager.GetFileInfo(response.UUID, 999)
	assert.Error(t, err)
}

// TestUploadManager_ListUserFiles 测试列出用户文件
func TestUploadManager_ListUserFiles(t *testing.T) {
	db := setupManagerTestDB(t)
	tempDir := t.TempDir()

	config := DefaultUploadConfig()
	config.UploadPath = tempDir
	config.BaseURL = "http://localhost:8080"

	manager := NewUploadManager(db, config)
	defer manager.Stop()

	// 上传多个文件
	for i := 0; i < 5; i++ {
		fileHeader := createManagerTestFile(t, fmt.Sprintf("test%d.txt", i), fmt.Sprintf("Content %d", i))
		_, err := manager.UploadSingleFile(
			fileHeader,
			1, // userID
			model.BusinessTypeProduct,
			nil,
			true,
		)
		require.NoError(t, err)
	}

	// 列出用户文件
	files, total, err := manager.ListUserFiles(1, nil, nil, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, files, 5)

	// 测试分页
	files, total, err = manager.ListUserFiles(1, nil, nil, 1, 3)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, files, 3)

	// 测试按业务类型过滤
	businessType := model.BusinessTypeProduct
	files, total, err = manager.ListUserFiles(1, &businessType, nil, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, files, 5)
}

// TestUploadManager_GetUploadStatistics 测试获取上传统计
func TestUploadManager_GetUploadStatistics(t *testing.T) {
	db := setupManagerTestDB(t)
	tempDir := t.TempDir()

	config := DefaultUploadConfig()
	config.UploadPath = tempDir
	config.BaseURL = "http://localhost:8080"

	manager := NewUploadManager(db, config)
	defer manager.Stop()

	// 上传不同类型的文件
	files := []struct {
		name    string
		content string
	}{
		{"test1.txt", "Content 1"},
		{"test2.txt", "Content 2"},
		{"image.jpg", "Image content"},
	}

	for _, file := range files {
		fileHeader := createManagerTestFile(t, file.name, file.content)
		_, err := manager.UploadSingleFile(
			fileHeader,
			1, // userID
			model.BusinessTypeProduct,
			nil,
			true,
		)
		require.NoError(t, err)
	}

	// 获取统计信息
	stats, err := manager.GetUploadStatistics(1)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), stats.TotalFiles)
	assert.Greater(t, stats.TotalSize, int64(0))
	assert.NotEmpty(t, stats.FileTypeStats)
}

// TestUploadManager_UpdateConfig 测试更新配置
func TestUploadManager_UpdateConfig(t *testing.T) {
	db := setupManagerTestDB(t)

	manager := NewUploadManager(db, nil)
	defer manager.Stop()

	// 获取原始配置
	originalConfig := manager.GetUploadConfig()
	originalPath := originalConfig.UploadPath

	// 更新配置
	newConfig := DefaultUploadConfig()
	newConfig.UploadPath = "/new/path"
	newConfig.BaseURL = "http://new.com"

	err := manager.UpdateConfig(newConfig)
	assert.NoError(t, err)

	// 验证配置已更新
	updatedConfig := manager.GetUploadConfig()
	assert.Equal(t, "/new/path", updatedConfig.UploadPath)
	assert.Equal(t, "http://new.com", updatedConfig.BaseURL)
	assert.NotEqual(t, originalPath, updatedConfig.UploadPath)

	// 测试无效配置
	invalidConfig := &UploadConfig{
		UploadPath: "", // 无效的空路径
	}

	err = manager.UpdateConfig(invalidConfig)
	assert.Error(t, err)
}
