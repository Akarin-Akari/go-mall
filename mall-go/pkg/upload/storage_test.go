package upload

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// StorageTestSuite 存储测试套件
type StorageTestSuite struct {
	suite.Suite
	tempDir string
	config  *UploadConfig
	storage *LocalStorage
}

// SetupSuite 设置测试套件
func (suite *StorageTestSuite) SetupSuite() {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "upload_test_")
	suite.Require().NoError(err)
	suite.tempDir = tempDir

	// 创建测试配置
	suite.config = &UploadConfig{
		StorageType: StorageTypeLocal,
		Local: LocalConfig{
			UploadPath: tempDir,
			URLPrefix:  "/uploads",
		},
	}

	// 创建本地存储
	storage, err := NewLocalStorage(suite.config)
	suite.Require().NoError(err)
	suite.storage = storage
}

// TearDownSuite 清理测试套件
func (suite *StorageTestSuite) TearDownSuite() {
	// 清理临时目录
	os.RemoveAll(suite.tempDir)
}

// TestLocalStorage_Upload 测试文件上传
func (suite *StorageTestSuite) TestLocalStorage_Upload() {
	testData := "Hello, World!"
	reader := strings.NewReader(testData)
	filePath := "test/upload.txt"

	// 上传文件
	err := suite.storage.Upload(filePath, reader, int64(len(testData)))
	suite.NoError(err)

	// 验证文件是否存在
	fullPath := filepath.Join(suite.tempDir, filePath)
	_, err = os.Stat(fullPath)
	suite.NoError(err)

	// 验证文件内容
	content, err := os.ReadFile(fullPath)
	suite.NoError(err)
	suite.Equal(testData, string(content))
}

// TestLocalStorage_Download 测试文件下载
func (suite *StorageTestSuite) TestLocalStorage_Download() {
	testData := "Download test content"
	filePath := "test/download.txt"

	// 先上传文件
	reader := strings.NewReader(testData)
	err := suite.storage.Upload(filePath, reader, int64(len(testData)))
	suite.NoError(err)

	// 下载文件
	downloadReader, err := suite.storage.Download(filePath)
	suite.NoError(err)
	defer downloadReader.Close()

	// 验证下载内容
	content, err := io.ReadAll(downloadReader)
	suite.NoError(err)
	suite.Equal(testData, string(content))
}

// TestLocalStorage_Delete 测试文件删除
func (suite *StorageTestSuite) TestLocalStorage_Delete() {
	testData := "Delete test content"
	filePath := "test/delete.txt"

	// 先上传文件
	reader := strings.NewReader(testData)
	err := suite.storage.Upload(filePath, reader, int64(len(testData)))
	suite.NoError(err)

	// 验证文件存在
	exists, err := suite.storage.Exists(filePath)
	suite.NoError(err)
	suite.True(exists)

	// 删除文件
	err = suite.storage.Delete(filePath)
	suite.NoError(err)

	// 验证文件不存在
	exists, err = suite.storage.Exists(filePath)
	suite.NoError(err)
	suite.False(exists)
}

// TestLocalStorage_Exists 测试文件存在检查
func (suite *StorageTestSuite) TestLocalStorage_Exists() {
	filePath := "test/exists.txt"

	// 文件不存在
	exists, err := suite.storage.Exists(filePath)
	suite.NoError(err)
	suite.False(exists)

	// 上传文件
	testData := "Exists test content"
	reader := strings.NewReader(testData)
	err = suite.storage.Upload(filePath, reader, int64(len(testData)))
	suite.NoError(err)

	// 文件存在
	exists, err = suite.storage.Exists(filePath)
	suite.NoError(err)
	suite.True(exists)
}

// TestLocalStorage_GetFileInfo 测试获取文件信息
func (suite *StorageTestSuite) TestLocalStorage_GetFileInfo() {
	testData := "File info test content"
	filePath := "test/fileinfo.txt"

	// 上传文件
	reader := strings.NewReader(testData)
	err := suite.storage.Upload(filePath, reader, int64(len(testData)))
	suite.NoError(err)

	// 获取文件信息
	fileInfo, err := suite.storage.GetFileInfo(filePath)
	suite.NoError(err)
	suite.NotNil(fileInfo)

	suite.Equal(filePath, fileInfo.Path)
	suite.Equal("fileinfo.txt", fileInfo.Name)
	suite.Equal(int64(len(testData)), fileInfo.Size)
	suite.Equal("text/plain", fileInfo.ContentType)
	suite.NotEmpty(fileInfo.ETag)
	suite.NotEmpty(fileInfo.URL)
}

// TestLocalStorage_ListFiles 测试文件列表
func (suite *StorageTestSuite) TestLocalStorage_ListFiles() {
	// 上传多个测试文件
	testFiles := []struct {
		path    string
		content string
	}{
		{"list/file1.txt", "Content 1"},
		{"list/file2.txt", "Content 2"},
		{"list/subdir/file3.txt", "Content 3"},
		{"other/file4.txt", "Content 4"},
	}

	for _, tf := range testFiles {
		reader := strings.NewReader(tf.content)
		err := suite.storage.Upload(tf.path, reader, int64(len(tf.content)))
		suite.NoError(err)
	}

	// 列出所有文件
	files, err := suite.storage.ListFiles("", 0)
	suite.NoError(err)
	suite.GreaterOrEqual(len(files), 4)

	// 列出特定前缀的文件
	files, err = suite.storage.ListFiles("list", 0)
	suite.NoError(err)
	suite.GreaterOrEqual(len(files), 3)

	// 限制文件数量
	files, err = suite.storage.ListFiles("list", 2)
	suite.NoError(err)
	suite.LessOrEqual(len(files), 2)
}

// TestLocalStorage_GetURL 测试获取文件URL
func (suite *StorageTestSuite) TestLocalStorage_GetURL() {
	filePath := "test/url.txt"
	expectedURL := "/uploads/" + filePath

	url := suite.storage.GetURL(filePath)
	suite.Equal(expectedURL, url)
}

// TestLocalStorage_ErrorCases 测试错误情况
func (suite *StorageTestSuite) TestLocalStorage_ErrorCases() {
	// 下载不存在的文件
	_, err := suite.storage.Download("nonexistent/file.txt")
	suite.Error(err)
	suite.Contains(err.Error(), "文件不存在")

	// 获取不存在文件的信息
	_, err = suite.storage.GetFileInfo("nonexistent/file.txt")
	suite.Error(err)
	suite.Contains(err.Error(), "文件不存在")

	// 删除不存在的文件（应该成功，不报错）
	err = suite.storage.Delete("nonexistent/file.txt")
	suite.NoError(err)
}

// TestStorageManager 测试存储管理器
func (suite *StorageTestSuite) TestStorageManager() {
	// 创建存储管理器
	manager, err := NewStorageManager(suite.config)
	suite.NoError(err)
	suite.NotNil(manager)

	// 测试上传
	testData := "Storage manager test"
	reader := strings.NewReader(testData)
	filePath := "manager/test.txt"

	err = manager.Upload(filePath, reader, int64(len(testData)))
	suite.NoError(err)

	// 测试下载
	downloadReader, err := manager.Download(filePath)
	suite.NoError(err)
	defer downloadReader.Close()

	content, err := io.ReadAll(downloadReader)
	suite.NoError(err)
	suite.Equal(testData, string(content))

	// 测试获取URL
	url := manager.GetURL(filePath)
	suite.Equal("/uploads/"+filePath, url)

	// 测试删除
	err = manager.Delete(filePath)
	suite.NoError(err)

	exists, err := manager.Exists(filePath)
	suite.NoError(err)
	suite.False(exists)
}

// 运行存储测试套件
func TestStorageSuite(t *testing.T) {
	suite.Run(t, new(StorageTestSuite))
}

// TestNewStorageManager 测试创建存储管理器
func TestNewStorageManager(t *testing.T) {
	tests := []struct {
		name        string
		config      *UploadConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "本地存储配置",
			config: &UploadConfig{
				StorageType: StorageTypeLocal,
				Local: LocalConfig{
					UploadPath: "/tmp/test",
				},
			},
			expectError: false,
		},
		{
			name: "不支持的存储类型",
			config: &UploadConfig{
				StorageType: "unsupported",
			},
			expectError: true,
			errorMsg:    "不支持的存储类型",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewStorageManager(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, manager)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, manager)
			}
		})
	}
}

// TestGetContentType 测试内容类型检测
func TestGetContentType(t *testing.T) {
	tests := []struct {
		filename     string
		expectedType string
	}{
		{"test.jpg", "image/jpeg"},
		{"test.jpeg", "image/jpeg"},
		{"test.png", "image/png"},
		{"test.gif", "image/gif"},
		{"test.webp", "image/webp"},
		{"test.pdf", "application/pdf"},
		{"test.txt", "text/plain"},
		{"test.json", "application/json"},
		{"test.xml", "application/xml"},
		{"test.zip", "application/zip"},
		{"test.mp4", "video/mp4"},
		{"test.mp3", "audio/mpeg"},
		{"test.unknown", "application/octet-stream"},
		{"", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			contentType := getContentType(tt.filename)
			assert.Equal(t, tt.expectedType, contentType)
		})
	}
}

// BenchmarkLocalStorage_Upload 性能测试
func BenchmarkLocalStorage_Upload(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "upload_bench_")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	config := &UploadConfig{
		StorageType: StorageTypeLocal,
		Local: LocalConfig{
			UploadPath: tempDir,
		},
	}

	storage, err := NewLocalStorage(config)
	if err != nil {
		b.Fatal(err)
	}

	testData := bytes.Repeat([]byte("test data "), 1000) // 10KB数据

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(testData)
		filePath := fmt.Sprintf("bench/file_%d.txt", i)
		err := storage.Upload(filePath, reader, int64(len(testData)))
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkLocalStorage_Download 性能测试
func BenchmarkLocalStorage_Download(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "download_bench_")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	config := &UploadConfig{
		StorageType: StorageTypeLocal,
		Local: LocalConfig{
			UploadPath: tempDir,
		},
	}

	storage, err := NewLocalStorage(config)
	if err != nil {
		b.Fatal(err)
	}

	// 准备测试文件
	testData := bytes.Repeat([]byte("test data "), 1000) // 10KB数据
	filePath := "bench/download_test.txt"
	reader := bytes.NewReader(testData)
	err = storage.Upload(filePath, reader, int64(len(testData)))
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		downloadReader, err := storage.Download(filePath)
		if err != nil {
			b.Fatal(err)
		}

		// 读取所有数据
		_, err = io.ReadAll(downloadReader)
		if err != nil {
			b.Fatal(err)
		}

		downloadReader.Close()
	}
}
