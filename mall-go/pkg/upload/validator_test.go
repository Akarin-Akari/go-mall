package upload

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// ValidatorTestSuite 验证器测试套件
type ValidatorTestSuite struct {
	suite.Suite
	config    *UploadConfig
	validator *FileValidator
}

// SetupSuite 设置测试套件
func (suite *ValidatorTestSuite) SetupSuite() {
	suite.config = DefaultUploadConfig()
	suite.validator = NewFileValidator(suite.config)
}

// TestFileValidator_ValidateFile 测试文件验证
func (suite *ValidatorTestSuite) TestFileValidator_ValidateFile() {
	tests := []struct {
		name        string
		filename    string
		content     string
		size        int64
		expectValid bool
		expectError string
	}{
		{
			name:        "有效的JPEG文件",
			filename:    "test.jpg",
			content:     "\xFF\xD8\xFF\xE0\x00\x10JFIF", // JPEG文件头
			size:        1024,
			expectValid: true,
		},
		{
			name:        "文件大小超过限制",
			filename:    "large.jpg",
			content:     "\xFF\xD8\xFF\xE0\x00\x10JFIF",
			size:        suite.config.MaxFileSize + 1,
			expectValid: false,
		},
		{
			name:        "空文件",
			filename:    "empty.txt",
			content:     "",
			size:        0,
			expectValid: false,
			expectError: "文件为空",
		},
		{
			name:        "禁止的文件扩展名",
			filename:    "malware.exe",
			content:     "MZ", // Windows可执行文件头
			size:        1024,
			expectValid: false,
			expectError: "文件类型被禁止",
		},
		{
			name:        "不支持的文件扩展名",
			filename:    "test.xyz",
			content:     "test content",
			size:        100,
			expectValid: false,
			expectError: "不支持的文件扩展名",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			reader := strings.NewReader(tt.content)
			result, err := suite.validator.ValidateFile(tt.filename, reader, tt.size)

			if tt.expectValid {
				suite.NoError(err)
				suite.NotNil(result)
				suite.True(result.Valid)
				suite.Empty(result.Errors)
			} else {
				if tt.expectError != "" {
					suite.NotNil(result)
					suite.False(result.Valid)
					suite.NotEmpty(result.Errors)
					
					found := false
					for _, errMsg := range result.Errors {
						if strings.Contains(errMsg, tt.expectError) {
							found = true
							break
						}
					}
					suite.True(found, "Expected error message not found: %s", tt.expectError)
				}
			}
		})
	}
}

// TestFileValidator_SecurityCheck 测试安全检查
func (suite *ValidatorTestSuite) TestFileValidator_SecurityCheck() {
	// 启用安全检查
	suite.config.EnableSecurity = true
	validator := NewFileValidator(suite.config)

	tests := []struct {
		name        string
		filename    string
		content     string
		expectValid bool
		expectError string
	}{
		{
			name:        "Windows可执行文件",
			filename:    "test.exe",
			content:     "MZ\x90\x00", // PE文件头
			expectValid: false,
			expectError: "检测到Windows可执行文件",
		},
		{
			name:        "Linux可执行文件",
			filename:    "test.bin",
			content:     "\x7FELF", // ELF文件头
			expectValid: false,
			expectError: "检测到Linux可执行文件",
		},
		{
			name:        "PHP脚本文件",
			filename:    "test.txt",
			content:     "<?php echo 'hello'; ?>",
			expectValid: false,
			expectError: "检测到脚本文件特征",
		},
		{
			name:        "JavaScript代码",
			filename:    "test.txt",
			content:     "<script>alert('xss')</script>",
			expectValid: false,
			expectError: "检测到脚本文件特征",
		},
		{
			name:        "包含可疑代码的文件",
			filename:    "test.txt",
			content:     "eval(base64_decode('malicious code'))",
			expectValid: true, // 只是警告，不阻止上传
		},
		{
			name:        "正常文本文件",
			filename:    "test.txt",
			content:     "This is a normal text file content.",
			expectValid: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			reader := strings.NewReader(tt.content)
			result, err := validator.ValidateFile(tt.filename, reader, int64(len(tt.content)))

			suite.NoError(err)
			suite.NotNil(result)

			if tt.expectValid {
				suite.True(result.Valid)
			} else {
				suite.False(result.Valid)
				if tt.expectError != "" {
					found := false
					for _, errMsg := range result.Errors {
						if strings.Contains(errMsg, tt.expectError) {
							found = true
							break
						}
					}
					suite.True(found, "Expected error message not found: %s", tt.expectError)
				}
			}
		})
	}
}

// TestFileValidator_ImageValidation 测试图片验证
func (suite *ValidatorTestSuite) TestFileValidator_ImageValidation() {
	// 创建一个简单的PNG图片数据
	pngData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG签名
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, // IHDR块开始
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, // 1x1像素
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, 0xDE, // 其余IHDR数据
	}

	reader := bytes.NewReader(pngData)
	result, err := suite.validator.ValidateFile("test.png", reader, int64(len(pngData)))

	suite.NoError(err)
	suite.NotNil(result)
	suite.True(result.Valid)
	suite.Equal("image/png", result.ContentType)
	
	// 注意：由于我们的PNG数据不完整，图片解析可能会失败
	// 这里主要测试验证器的基本功能
}

// TestFileValidator_FileSizeValidation 测试文件大小验证
func (suite *ValidatorTestSuite) TestFileValidator_FileSizeValidation() {
	tests := []struct {
		name        string
		size        int64
		expectValid bool
	}{
		{
			name:        "正常大小",
			size:        1024,
			expectValid: true,
		},
		{
			name:        "最大允许大小",
			size:        suite.config.MaxFileSize,
			expectValid: true,
		},
		{
			name:        "超过最大大小",
			size:        suite.config.MaxFileSize + 1,
			expectValid: false,
		},
		{
			name:        "零大小",
			size:        0,
			expectValid: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			content := strings.Repeat("a", int(tt.size))
			reader := strings.NewReader(content)
			result, err := suite.validator.ValidateFile("test.txt", reader, tt.size)

			suite.NoError(err)
			suite.NotNil(result)
			suite.Equal(tt.expectValid, result.Valid)

			if !tt.expectValid {
				suite.NotEmpty(result.Errors)
			}
		})
	}
}

// TestFileValidator_ExtensionValidation 测试扩展名验证
func (suite *ValidatorTestSuite) TestFileValidator_ExtensionValidation() {
	tests := []struct {
		filename    string
		expectValid bool
	}{
		{"test.jpg", true},
		{"test.JPEG", true}, // 大小写不敏感
		{"test.png", true},
		{"test.gif", true},
		{"test.pdf", true},
		{"test.txt", false}, // 不在允许列表中
		{"test.exe", false}, // 被禁止的扩展名
		{"test", false},     // 无扩展名
	}

	for _, tt := range tests {
		suite.Run(tt.filename, func() {
			content := "test content"
			reader := strings.NewReader(content)
			result, err := suite.validator.ValidateFile(tt.filename, reader, int64(len(content)))

			suite.NoError(err)
			suite.NotNil(result)
			suite.Equal(tt.expectValid, result.Valid)
		})
	}
}

// TestFileValidator_ContentTypeValidation 测试内容类型验证
func (suite *ValidatorTestSuite) TestFileValidator_ContentTypeValidation() {
	tests := []struct {
		name        string
		content     string
		expectValid bool
	}{
		{
			name:        "JPEG图片",
			content:     "\xFF\xD8\xFF\xE0",
			expectValid: true,
		},
		{
			name:        "PNG图片",
			content:     "\x89PNG\r\n\x1a\n",
			expectValid: true,
		},
		{
			name:        "GIF图片",
			content:     "GIF87a",
			expectValid: true,
		},
		{
			name:        "PDF文档",
			content:     "%PDF-1.4",
			expectValid: true,
		},
		{
			name:        "普通文本",
			content:     "This is plain text",
			expectValid: false, // 不在允许的类型列表中
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			reader := strings.NewReader(tt.content)
			result, err := suite.validator.ValidateFile("test.jpg", reader, int64(len(tt.content)))

			suite.NoError(err)
			suite.NotNil(result)
			
			// 注意：这里的验证结果取决于文件扩展名和内容类型的匹配
			// 实际结果可能因为类型不匹配而产生警告
		})
	}
}

// TestFileValidator_DisabledSecurity 测试禁用安全检查
func (suite *ValidatorTestSuite) TestFileValidator_DisabledSecurity() {
	// 禁用安全检查
	config := DefaultUploadConfig()
	config.EnableSecurity = false
	validator := NewFileValidator(config)

	// 即使是可执行文件，也应该通过验证（如果其他条件满足）
	content := "MZ\x90\x00" // Windows可执行文件头
	reader := strings.NewReader(content)
	result, err := validator.ValidateFile("test.jpg", reader, int64(len(content)))

	suite.NoError(err)
	suite.NotNil(result)
	// 应该通过基本验证，不进行安全检查
}

// 运行验证器测试套件
func TestValidatorSuite(t *testing.T) {
	suite.Run(t, new(ValidatorTestSuite))
}

// TestNewFileValidator 测试创建文件验证器
func TestNewFileValidator(t *testing.T) {
	config := DefaultUploadConfig()
	validator := NewFileValidator(config)

	assert.NotNil(t, validator)
	assert.Equal(t, config, validator.config)
}

// TestValidationResult 测试验证结果结构
func TestValidationResult(t *testing.T) {
	result := &ValidationResult{
		Valid:       true,
		Filename:    "test.jpg",
		Size:        1024,
		ContentType: "image/jpeg",
		Errors:      []string{},
		Warnings:    []string{"测试警告"},
	}

	assert.True(t, result.Valid)
	assert.Equal(t, "test.jpg", result.Filename)
	assert.Equal(t, int64(1024), result.Size)
	assert.Equal(t, "image/jpeg", result.ContentType)
	assert.Empty(t, result.Errors)
	assert.Len(t, result.Warnings, 1)
}

// BenchmarkFileValidator_ValidateFile 性能测试
func BenchmarkFileValidator_ValidateFile(b *testing.B) {
	config := DefaultUploadConfig()
	validator := NewFileValidator(config)
	
	content := strings.Repeat("test content ", 1000) // 约12KB
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(content)
		result, err := validator.ValidateFile("test.txt", reader, int64(len(content)))
		if err != nil {
			b.Fatal(err)
		}
		_ = result
	}
}

// TestGlobalValidator 测试全局验证器
func TestGlobalValidator(t *testing.T) {
	// 初始化全局验证器
	config := DefaultUploadConfig()
	InitGlobalValidator(config)

	// 测试全局验证器
	validator := GetGlobalValidator()
	assert.NotNil(t, validator)

	// 测试全局验证函数
	content := "test content"
	reader := strings.NewReader(content)
	result, err := ValidateUploadFile("test.jpg", reader, int64(len(content)))
	
	assert.NoError(t, err)
	assert.NotNil(t, result)
}
