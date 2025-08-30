package upload

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultUploadConfig(t *testing.T) {
	config := DefaultUploadConfig()

	assert.NotNil(t, config)
	assert.Equal(t, StorageTypeLocal, config.StorageType)
	assert.Equal(t, int64(10*1024*1024), config.MaxFileSize) // 10MB
	assert.Equal(t, 5, config.MaxFiles)
	assert.True(t, config.EnableSecurity)
	assert.True(t, config.EnableThumbnail)
	assert.NotEmpty(t, config.AllowedTypes)
	assert.NotEmpty(t, config.AllowedExts)
}

func TestUploadConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *UploadConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "默认配置验证通过",
			config:  DefaultUploadConfig(),
			wantErr: false,
		},
		{
			name: "最大文件大小为0",
			config: &UploadConfig{
				MaxFileSize: 0,
				MaxFiles:    1,
				AllowedTypes: []string{"image/jpeg"},
				StorageType: StorageTypeLocal,
				Local: LocalConfig{
					UploadPath: "/tmp",
				},
			},
			wantErr: true,
			errMsg:  "最大文件大小必须大于0",
		},
		{
			name: "最大文件数为0",
			config: &UploadConfig{
				MaxFileSize: 1024,
				MaxFiles:    0,
				AllowedTypes: []string{"image/jpeg"},
				StorageType: StorageTypeLocal,
				Local: LocalConfig{
					UploadPath: "/tmp",
				},
			},
			wantErr: true,
			errMsg:  "最大文件数必须大于0",
		},
		{
			name: "本地存储路径为空",
			config: &UploadConfig{
				MaxFileSize: 1024,
				MaxFiles:    1,
				AllowedTypes: []string{"image/jpeg"},
				StorageType: StorageTypeLocal,
				Local: LocalConfig{
					UploadPath: "",
				},
			},
			wantErr: true,
			errMsg:  "本地存储路径不能为空",
		},
		{
			name: "OSS配置不完整",
			config: &UploadConfig{
				MaxFileSize: 1024,
				MaxFiles:    1,
				AllowedTypes: []string{"image/jpeg"},
				StorageType: StorageTypeOSS,
				OSS: OSSConfig{
					Endpoint: "oss-cn-hangzhou.aliyuncs.com",
					// 缺少其他必需字段
				},
			},
			wantErr: true,
			errMsg:  "OSS配置不完整",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUploadConfig_IsAllowedType(t *testing.T) {
	config := &UploadConfig{
		AllowedTypes: []string{"image/jpeg", "image/png", "application/pdf"},
	}

	tests := []struct {
		contentType string
		expected    bool
	}{
		{"image/jpeg", true},
		{"image/png", true},
		{"application/pdf", true},
		{"image/gif", false},
		{"text/plain", false},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			result := config.IsAllowedType(tt.contentType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUploadConfig_IsAllowedExt(t *testing.T) {
	config := &UploadConfig{
		AllowedExts: []string{".jpg", ".jpeg", ".png", ".pdf"},
	}

	tests := []struct {
		filename string
		expected bool
	}{
		{"test.jpg", true},
		{"test.JPEG", true}, // 大小写不敏感
		{"test.png", true},
		{"test.pdf", true},
		{"test.gif", false},
		{"test.txt", false},
		{"test", false}, // 无扩展名
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := config.IsAllowedExt(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUploadConfig_IsForbiddenExt(t *testing.T) {
	config := &UploadConfig{
		Security: SecurityConfig{
			ForbiddenExts: []string{".exe", ".bat", ".cmd"},
		},
	}

	tests := []struct {
		filename string
		expected bool
	}{
		{"test.exe", true},
		{"test.EXE", true}, // 大小写不敏感
		{"test.bat", true},
		{"test.cmd", true},
		{"test.jpg", false},
		{"test.pdf", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := config.IsForbiddenExt(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUploadConfig_GenerateFileName(t *testing.T) {
	config := DefaultUploadConfig()

	tests := []struct {
		originalName string
		expectExt    string
	}{
		{"test.jpg", ".jpg"},
		{"document.pdf", ".pdf"},
		{"image.PNG", ".PNG"},
		{"noext", ""},
	}

	for _, tt := range tests {
		t.Run(tt.originalName, func(t *testing.T) {
			fileName := config.GenerateFileName(tt.originalName)
			
			assert.NotEmpty(t, fileName)
			assert.NotEqual(t, tt.originalName, fileName) // 应该生成新的文件名
			
			if tt.expectExt != "" {
				assert.True(t, len(fileName) > len(tt.expectExt))
				assert.Equal(t, tt.expectExt, fileName[len(fileName)-len(tt.expectExt):])
			}
		})
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	// 保存原始环境变量
	originalEnv := map[string]string{
		"UPLOAD_STORAGE_TYPE":   os.Getenv("UPLOAD_STORAGE_TYPE"),
		"UPLOAD_MAX_FILE_SIZE":  os.Getenv("UPLOAD_MAX_FILE_SIZE"),
		"UPLOAD_MAX_FILES":      os.Getenv("UPLOAD_MAX_FILES"),
		"UPLOAD_ALLOWED_TYPES":  os.Getenv("UPLOAD_ALLOWED_TYPES"),
		"UPLOAD_LOCAL_PATH":     os.Getenv("UPLOAD_LOCAL_PATH"),
		"OSS_ENDPOINT":          os.Getenv("OSS_ENDPOINT"),
		"OSS_ACCESS_KEY_ID":     os.Getenv("OSS_ACCESS_KEY_ID"),
	}

	// 清理环境变量
	defer func() {
		for key, value := range originalEnv {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	// 设置测试环境变量
	os.Setenv("UPLOAD_STORAGE_TYPE", "oss")
	os.Setenv("UPLOAD_MAX_FILE_SIZE", "20971520") // 20MB
	os.Setenv("UPLOAD_MAX_FILES", "10")
	os.Setenv("UPLOAD_ALLOWED_TYPES", "image/jpeg,image/png")
	os.Setenv("UPLOAD_LOCAL_PATH", "/custom/upload")
	os.Setenv("OSS_ENDPOINT", "oss-cn-beijing.aliyuncs.com")
	os.Setenv("OSS_ACCESS_KEY_ID", "test_access_key")

	config := LoadConfigFromEnv()

	assert.Equal(t, StorageTypeOSS, config.StorageType)
	assert.Equal(t, int64(20971520), config.MaxFileSize)
	assert.Equal(t, 10, config.MaxFiles)
	assert.Equal(t, []string{"image/jpeg", "image/png"}, config.AllowedTypes)
	assert.Equal(t, "/custom/upload", config.Local.UploadPath)
	assert.Equal(t, "oss-cn-beijing.aliyuncs.com", config.OSS.Endpoint)
	assert.Equal(t, "test_access_key", config.OSS.AccessKeyID)
}

func TestUploadConfig_SaveAndLoadFromFile(t *testing.T) {
	// 创建测试配置
	config := DefaultUploadConfig()
	config.StorageType = StorageTypeOSS
	config.MaxFileSize = 20 * 1024 * 1024
	config.OSS.Endpoint = "oss-cn-hangzhou.aliyuncs.com"

	// 保存到临时文件
	tempFile := "/tmp/test_upload_config.json"
	defer os.Remove(tempFile)

	err := config.SaveToFile(tempFile)
	assert.NoError(t, err)

	// 从文件加载
	loadedConfig, err := LoadConfigFromFile(tempFile)
	assert.NoError(t, err)
	assert.NotNil(t, loadedConfig)

	// 验证配置
	assert.Equal(t, config.StorageType, loadedConfig.StorageType)
	assert.Equal(t, config.MaxFileSize, loadedConfig.MaxFileSize)
	assert.Equal(t, config.OSS.Endpoint, loadedConfig.OSS.Endpoint)
}

func TestUploadConfig_GetFileURL(t *testing.T) {
	tests := []struct {
		name        string
		config      *UploadConfig
		filePath    string
		expectedURL string
	}{
		{
			name: "本地存储URL",
			config: &UploadConfig{
				StorageType: StorageTypeLocal,
				Local: LocalConfig{
					UploadPath: "/uploads",
					URLPrefix:  "/static/uploads",
				},
			},
			filePath:    "2023/12/01/test.jpg",
			expectedURL: "/static/uploads/2023/12/01/test.jpg",
		},
		{
			name: "OSS存储URL",
			config: &UploadConfig{
				StorageType: StorageTypeOSS,
				OSS: OSSConfig{
					URLPrefix: "https://bucket.oss-cn-hangzhou.aliyuncs.com",
				},
			},
			filePath:    "2023/12/01/test.jpg",
			expectedURL: "https://bucket.oss-cn-hangzhou.aliyuncs.com/2023/12/01/test.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := tt.config.GetFileURL(tt.filePath)
			assert.Equal(t, tt.expectedURL, url)
		})
	}
}

func TestUploadConfig_GetUploadPath(t *testing.T) {
	config := &UploadConfig{
		StorageType: StorageTypeLocal,
		Local: LocalConfig{
			UploadPath: "/var/uploads",
		},
	}

	tests := []struct {
		subPath      string
		expectedPath string
	}{
		{"2023/12/01/test.jpg", "/var/uploads/2023/12/01/test.jpg"},
		{"avatar/user123.png", "/var/uploads/avatar/user123.png"},
		{"", "/var/uploads"},
	}

	for _, tt := range tests {
		t.Run(tt.subPath, func(t *testing.T) {
			path := config.GetUploadPath(tt.subPath)
			assert.Equal(t, tt.expectedPath, path)
		})
	}
}

// BenchmarkGenerateFileName 性能测试
func BenchmarkGenerateFileName(b *testing.B) {
	config := DefaultUploadConfig()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fileName := config.GenerateFileName("test.jpg")
		_ = fileName
	}
}

// BenchmarkIsAllowedType 性能测试
func BenchmarkIsAllowedType(b *testing.B) {
	config := DefaultUploadConfig()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		allowed := config.IsAllowedType("image/jpeg")
		_ = allowed
	}
}

func TestUploadConfig_EdgeCases(t *testing.T) {
	// 测试边界情况
	config := &UploadConfig{}

	// 空的允许类型列表
	assert.True(t, config.IsAllowedType("any/type"))

	// 空的允许扩展名列表
	assert.True(t, config.IsAllowedExt("any.ext"))

	// 空的禁止扩展名列表
	assert.False(t, config.IsForbiddenExt("any.ext"))

	// 生成文件名时的边界情况
	fileName := config.GenerateFileName("")
	assert.NotEmpty(t, fileName)

	fileName = config.GenerateFileName("noext")
	assert.NotEmpty(t, fileName)
	assert.Equal(t, "noext", fileName[len(fileName)-5:]) // 应该保留原始名称
}
