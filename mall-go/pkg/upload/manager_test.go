package upload

import (
	"os"
	"strings"
	"testing"
	"time"

	"mall-go/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// FileManagerTestSuite 文件管理器测试套件
type FileManagerTestSuite struct {
	suite.Suite
	db            *gorm.DB
	tempDir       string
	configManager *ConfigManager
	fileManager   *FileManager
}

// SetupSuite 设置测试套件
func (suite *FileManagerTestSuite) SetupSuite() {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)
	suite.db = db

	// 自动迁移
	err = db.AutoMigrate(&model.User{}, &model.File{})
	suite.Require().NoError(err)

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "file_manager_test_")
	suite.Require().NoError(err)
	suite.tempDir = tempDir

	// 创建配置
	config := DefaultUploadConfig()
	config.Local.UploadPath = tempDir
	config.MaxFileSize = 1024 * 1024 // 1MB

	// 创建配置管理器
	configManager := &ConfigManager{
		config: config,
	}
	suite.configManager = configManager

	// 创建文件管理器
	fileManager, err := NewFileManager(db, configManager)
	suite.Require().NoError(err)
	suite.fileManager = fileManager

	// 创建测试用户
	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
	}
	suite.db.Create(user)
}

// TearDownSuite 清理测试套件
func (suite *FileManagerTestSuite) TearDownSuite() {
	// 清理临时目录
	os.RemoveAll(suite.tempDir)
}

// TestFileManager_UploadFile 测试文件上传
func (suite *FileManagerTestSuite) TestFileManager_UploadFile() {
	req := &UploadFileRequest{
		UserID:      1,
		Filename:    "test.txt",
		Reader:      strings.NewReader("Hello, World!"),
		Size:        13,
		Category:    "document",
		Description: "测试文件",
	}

	resp, err := suite.fileManager.UploadFile(req)
	suite.NoError(err)
	suite.NotNil(resp)

	// 验证响应
	suite.Equal(uint(1), resp.FileID)
	suite.Equal("test.txt", resp.OriginalName)
	suite.NotEmpty(resp.FileName)
	suite.NotEmpty(resp.FilePath)
	suite.Equal(int64(13), resp.FileSize)
	suite.Equal("document", resp.Category)
	suite.NotEmpty(resp.URL)

	// 验证数据库记录
	var file model.File
	err = suite.db.First(&file, resp.FileID).Error
	suite.NoError(err)
	suite.Equal(req.UserID, file.UserID)
	suite.Equal(req.Filename, file.OriginalName)
	suite.Equal(req.Category, file.Category)
	suite.Equal(req.Description, file.Description)
	suite.Equal("active", file.Status)

	// 验证文件是否实际存储
	exists, err := suite.fileManager.storageManager.Exists(file.FilePath)
	suite.NoError(err)
	suite.True(exists)
}

// TestFileManager_GetFileInfo 测试获取文件信息
func (suite *FileManagerTestSuite) TestFileManager_GetFileInfo() {
	// 先上传一个文件
	req := &UploadFileRequest{
		UserID:   1,
		Filename: "info_test.txt",
		Reader:   strings.NewReader("File info test"),
		Size:     14,
		Category: "general",
	}

	resp, err := suite.fileManager.UploadFile(req)
	suite.NoError(err)

	// 获取文件信息
	fileInfo, err := suite.fileManager.GetFileInfo(resp.FileID)
	suite.NoError(err)
	suite.NotNil(fileInfo)

	suite.Equal(resp.FileID, fileInfo.ID)
	suite.Equal(req.UserID, fileInfo.UserID)
	suite.Equal(req.Filename, fileInfo.OriginalName)
	suite.Equal(req.Category, fileInfo.Category)
	suite.Equal("active", fileInfo.Status)
}

// TestFileManager_GetFileInfo_NotFound 测试获取不存在的文件信息
func (suite *FileManagerTestSuite) TestFileManager_GetFileInfo_NotFound() {
	fileInfo, err := suite.fileManager.GetFileInfo(99999)
	suite.Error(err)
	suite.Nil(fileInfo)
	suite.Contains(err.Error(), "文件不存在")
}

// TestFileManager_ListFiles 测试获取文件列表
func (suite *FileManagerTestSuite) TestFileManager_ListFiles() {
	// 上传多个测试文件
	testFiles := []struct {
		filename string
		category string
		userID   uint
	}{
		{"file1.txt", "document", 1},
		{"file2.jpg", "image", 1},
		{"file3.pdf", "document", 1},
		{"file4.txt", "document", 2}, // 不同用户
	}

	for _, tf := range testFiles {
		req := &UploadFileRequest{
			UserID:   tf.userID,
			Filename: tf.filename,
			Reader:   strings.NewReader("test content"),
			Size:     12,
			Category: tf.category,
		}
		_, err := suite.fileManager.UploadFile(req)
		suite.NoError(err)
	}

	// 测试获取用户1的所有文件
	files, total, err := suite.fileManager.ListFiles(1, "", 1, 10)
	suite.NoError(err)
	suite.Equal(int64(3), total) // 用户1有3个文件
	suite.Len(files, 3)

	// 测试按分类过滤
	files, total, err = suite.fileManager.ListFiles(1, "document", 1, 10)
	suite.NoError(err)
	suite.Equal(int64(2), total) // 用户1有2个document文件
	suite.Len(files, 2)

	// 测试分页
	files, total, err = suite.fileManager.ListFiles(1, "", 1, 2)
	suite.NoError(err)
	suite.Equal(int64(3), total)
	suite.Len(files, 2) // 每页2个文件

	// 测试获取所有用户的文件
	files, total, err = suite.fileManager.ListFiles(0, "", 1, 10)
	suite.NoError(err)
	suite.Equal(int64(4), total) // 总共4个文件
	suite.Len(files, 4)
}

// TestFileManager_DeleteFile 测试删除文件
func (suite *FileManagerTestSuite) TestFileManager_DeleteFile() {
	// 上传测试文件
	req := &UploadFileRequest{
		UserID:   1,
		Filename: "delete_test.txt",
		Reader:   strings.NewReader("Delete me"),
		Size:     9,
		Category: "general",
	}

	resp, err := suite.fileManager.UploadFile(req)
	suite.NoError(err)

	// 验证文件存在
	exists, err := suite.fileManager.storageManager.Exists(resp.FilePath)
	suite.NoError(err)
	suite.True(exists)

	// 删除文件
	err = suite.fileManager.DeleteFile(resp.FileID, 1, false)
	suite.NoError(err)

	// 验证文件已从存储中删除
	exists, err = suite.fileManager.storageManager.Exists(resp.FilePath)
	suite.NoError(err)
	suite.False(exists)

	// 验证数据库记录已删除
	var file model.File
	err = suite.db.First(&file, resp.FileID).Error
	suite.Error(err)
	suite.Equal(gorm.ErrRecordNotFound, err)
}

// TestFileManager_DeleteFile_Permission 测试删除文件权限检查
func (suite *FileManagerTestSuite) TestFileManager_DeleteFile_Permission() {
	// 用户1上传文件
	req := &UploadFileRequest{
		UserID:   1,
		Filename: "permission_test.txt",
		Reader:   strings.NewReader("Permission test"),
		Size:     15,
		Category: "general",
	}

	resp, err := suite.fileManager.UploadFile(req)
	suite.NoError(err)

	// 用户2尝试删除用户1的文件（非管理员）
	err = suite.fileManager.DeleteFile(resp.FileID, 2, false)
	suite.Error(err)
	suite.Contains(err.Error(), "无权限删除此文件")

	// 管理员删除文件
	err = suite.fileManager.DeleteFile(resp.FileID, 2, true)
	suite.NoError(err)
}

// TestFileManager_DownloadFile 测试下载文件
func (suite *FileManagerTestSuite) TestFileManager_DownloadFile() {
	testContent := "Download test content"
	
	// 上传测试文件
	req := &UploadFileRequest{
		UserID:   1,
		Filename: "download_test.txt",
		Reader:   strings.NewReader(testContent),
		Size:     int64(len(testContent)),
		Category: "general",
	}

	resp, err := suite.fileManager.UploadFile(req)
	suite.NoError(err)

	// 下载文件
	reader, file, err := suite.fileManager.DownloadFile(resp.FileID)
	suite.NoError(err)
	suite.NotNil(reader)
	suite.NotNil(file)
	defer reader.Close()

	// 验证文件信息
	suite.Equal(resp.FileID, file.ID)
	suite.Equal(req.Filename, file.OriginalName)

	// 验证文件内容
	content := make([]byte, len(testContent))
	n, err := reader.Read(content)
	suite.NoError(err)
	suite.Equal(len(testContent), n)
	suite.Equal(testContent, string(content))
}

// TestFileManager_GetFileURL 测试获取文件URL
func (suite *FileManagerTestSuite) TestFileManager_GetFileURL() {
	// 上传测试文件
	req := &UploadFileRequest{
		UserID:   1,
		Filename: "url_test.txt",
		Reader:   strings.NewReader("URL test"),
		Size:     8,
		Category: "general",
	}

	resp, err := suite.fileManager.UploadFile(req)
	suite.NoError(err)

	// 获取文件URL
	url, err := suite.fileManager.GetFileURL(resp.FileID)
	suite.NoError(err)
	suite.NotEmpty(url)
	suite.Contains(url, "/uploads/")
}

// TestFileManager_UpdateFileInfo 测试更新文件信息
func (suite *FileManagerTestSuite) TestFileManager_UpdateFileInfo() {
	// 上传测试文件
	req := &UploadFileRequest{
		UserID:   1,
		Filename: "update_test.txt",
		Reader:   strings.NewReader("Update test"),
		Size:     11,
		Category: "general",
	}

	resp, err := suite.fileManager.UploadFile(req)
	suite.NoError(err)

	// 更新文件信息
	updates := map[string]interface{}{
		"description": "更新后的描述",
		"category":    "document",
	}

	err = suite.fileManager.UpdateFileInfo(resp.FileID, 1, updates)
	suite.NoError(err)

	// 验证更新
	file, err := suite.fileManager.GetFileInfo(resp.FileID)
	suite.NoError(err)
	suite.Equal("更新后的描述", file.Description)
	suite.Equal("document", file.Category)
}

// TestFileManager_GetFileStatistics 测试获取文件统计
func (suite *FileManagerTestSuite) TestFileManager_GetFileStatistics() {
	// 上传多个测试文件
	categories := []string{"document", "image", "document", "video"}
	for i, category := range categories {
		req := &UploadFileRequest{
			UserID:   1,
			Filename: fmt.Sprintf("stats_test_%d.txt", i),
			Reader:   strings.NewReader("Stats test"),
			Size:     10,
			Category: category,
		}
		_, err := suite.fileManager.UploadFile(req)
		suite.NoError(err)
	}

	// 获取统计信息
	stats, err := suite.fileManager.GetFileStatistics(1)
	suite.NoError(err)
	suite.NotNil(stats)

	// 验证统计数据
	suite.GreaterOrEqual(stats["total_files"].(int64), int64(4))
	suite.GreaterOrEqual(stats["total_size"].(int64), int64(40))
	
	categoryStats, ok := stats["category_stats"].([]struct {
		Category string `json:"category"`
		Count    int64  `json:"count"`
		Size     int64  `json:"size"`
	})
	suite.True(ok)
	suite.NotEmpty(categoryStats)
}

// TestFileManager_CleanupInvalidFiles 测试清理无效文件
func (suite *FileManagerTestSuite) TestFileManager_CleanupInvalidFiles() {
	// 上传测试文件
	req := &UploadFileRequest{
		UserID:   1,
		Filename: "cleanup_test.txt",
		Reader:   strings.NewReader("Cleanup test"),
		Size:     12,
		Category: "general",
	}

	resp, err := suite.fileManager.UploadFile(req)
	suite.NoError(err)

	// 手动删除存储文件（模拟文件丢失）
	err = suite.fileManager.storageManager.Delete(resp.FilePath)
	suite.NoError(err)

	// 执行清理
	result, err := suite.fileManager.CleanupInvalidFiles()
	suite.NoError(err)
	suite.NotNil(result)

	// 验证清理结果
	suite.GreaterOrEqual(result["checked_files"].(int), 1)
	suite.GreaterOrEqual(result["invalid_files"].(int), 1)
	suite.GreaterOrEqual(result["cleaned_files"].(int), 1)

	// 验证文件状态已更新
	file, err := suite.fileManager.GetFileInfo(resp.FileID)
	suite.NoError(err)
	suite.Equal("deleted", file.Status)
}

// 运行文件管理器测试套件
func TestFileManagerSuite(t *testing.T) {
	suite.Run(t, new(FileManagerTestSuite))
}

// TestNewFileManager 测试创建文件管理器
func TestNewFileManager(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// 创建配置管理器
	config := DefaultUploadConfig()
	configManager := &ConfigManager{config: config}

	// 创建文件管理器
	fileManager, err := NewFileManager(db, configManager)
	assert.NoError(t, err)
	assert.NotNil(t, fileManager)
	assert.Equal(t, db, fileManager.db)
	assert.NotNil(t, fileManager.storageManager)
	assert.NotNil(t, fileManager.validator)
}

// TestGlobalFileManager 测试全局文件管理器
func TestGlobalFileManager(t *testing.T) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// 创建配置管理器
	config := DefaultUploadConfig()
	configManager := &ConfigManager{config: config}

	// 初始化全局文件管理器
	err = InitGlobalFileManager(db, configManager)
	assert.NoError(t, err)

	// 获取全局文件管理器
	manager := GetGlobalFileManager()
	assert.NotNil(t, manager)
}
