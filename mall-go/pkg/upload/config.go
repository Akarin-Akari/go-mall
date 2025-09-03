package upload

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// StorageType 存储类型
type StorageType string

const (
	StorageTypeLocal StorageType = "local" // 本地存储
	StorageTypeOSS   StorageType = "oss"   // 阿里云OSS
	StorageTypeS3    StorageType = "s3"    // AWS S3
	StorageTypeCOS   StorageType = "cos"   // 腾讯云COS
)

// UploadConfig 文件上传配置
type UploadConfig struct {
	// 基础配置
	StorageType     StorageType `json:"storage_type" yaml:"storage_type"`         // 存储类型
	MaxFileSize     int64       `json:"max_file_size" yaml:"max_file_size"`       // 最大文件大小(字节)
	MaxFiles        int         `json:"max_files" yaml:"max_files"`               // 单次最大上传文件数
	AllowedTypes    []string    `json:"allowed_types" yaml:"allowed_types"`       // 允许的文件类型
	AllowedExts     []string    `json:"allowed_exts" yaml:"allowed_exts"`         // 允许的文件扩展名
	EnableSecurity  bool        `json:"enable_security" yaml:"enable_security"`   // 启用安全检查
	EnableThumbnail bool        `json:"enable_thumbnail" yaml:"enable_thumbnail"` // 启用缩略图生成

	// 本地存储配置
	Local LocalConfig `json:"local" yaml:"local"`

	// 阿里云OSS配置
	OSS OSSConfig `json:"oss" yaml:"oss"`

	// AWS S3配置
	S3 S3Config `json:"s3" yaml:"s3"`

	// 腾讯云COS配置
	COS COSConfig `json:"cos" yaml:"cos"`

	// 安全配置
	Security SecurityConfig `json:"security" yaml:"security"`

	// 缩略图配置
	Thumbnail ThumbnailConfig `json:"thumbnail" yaml:"thumbnail"`
}

// LocalConfig 本地存储配置
type LocalConfig struct {
	UploadPath string `json:"upload_path" yaml:"upload_path"` // 上传路径
	URLPrefix  string `json:"url_prefix" yaml:"url_prefix"`   // URL前缀
}

// OSSConfig 阿里云OSS配置
type OSSConfig struct {
	Endpoint        string `json:"endpoint" yaml:"endpoint"`                 // OSS端点
	AccessKeyID     string `json:"access_key_id" yaml:"access_key_id"`       // AccessKey ID
	AccessKeySecret string `json:"access_key_secret" yaml:"access_key_secret"` // AccessKey Secret
	BucketName      string `json:"bucket_name" yaml:"bucket_name"`           // 存储桶名称
	URLPrefix       string `json:"url_prefix" yaml:"url_prefix"`             // URL前缀
}

// S3Config AWS S3配置
type S3Config struct {
	Region          string `json:"region" yaml:"region"`                     // AWS区域
	AccessKeyID     string `json:"access_key_id" yaml:"access_key_id"`       // AccessKey ID
	SecretAccessKey string `json:"secret_access_key" yaml:"secret_access_key"` // Secret Access Key
	BucketName      string `json:"bucket_name" yaml:"bucket_name"`           // 存储桶名称
	URLPrefix       string `json:"url_prefix" yaml:"url_prefix"`             // URL前缀
}

// COSConfig 腾讯云COS配置
type COSConfig struct {
	Region    string `json:"region" yaml:"region"`         // COS区域
	SecretID  string `json:"secret_id" yaml:"secret_id"`   // Secret ID
	SecretKey string `json:"secret_key" yaml:"secret_key"` // Secret Key
	Bucket    string `json:"bucket" yaml:"bucket"`         // 存储桶名称
	URLPrefix string `json:"url_prefix" yaml:"url_prefix"` // URL前缀
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	EnableVirusScan   bool     `json:"enable_virus_scan" yaml:"enable_virus_scan"`     // 启用病毒扫描
	EnableContentScan bool     `json:"enable_content_scan" yaml:"enable_content_scan"` // 启用内容扫描
	ForbiddenExts     []string `json:"forbidden_exts" yaml:"forbidden_exts"`           // 禁止的扩展名
	MaxImageWidth     int      `json:"max_image_width" yaml:"max_image_width"`         // 最大图片宽度
	MaxImageHeight    int      `json:"max_image_height" yaml:"max_image_height"`       // 最大图片高度
}

// ThumbnailConfig 缩略图配置
type ThumbnailConfig struct {
	Sizes   []ThumbnailSize `json:"sizes" yaml:"sizes"`     // 缩略图尺寸
	Quality int             `json:"quality" yaml:"quality"` // 图片质量(1-100)
	Format  string          `json:"format" yaml:"format"`   // 输出格式(jpeg, png, webp)
}

// ThumbnailSize 缩略图尺寸
type ThumbnailSize struct {
	Name   string `json:"name" yaml:"name"`     // 尺寸名称
	Width  int    `json:"width" yaml:"width"`   // 宽度
	Height int    `json:"height" yaml:"height"` // 高度
}

// DefaultUploadConfig 返回默认配置
func DefaultUploadConfig() *UploadConfig {
	return &UploadConfig{
		StorageType:     StorageTypeLocal,
		MaxFileSize:     10 * 1024 * 1024, // 10MB
		MaxFiles:        5,
		AllowedTypes:    []string{"image/jpeg", "image/png", "image/gif", "image/webp", "application/pdf"},
		AllowedExts:     []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".pdf"},
		EnableSecurity:  true,
		EnableThumbnail: true,

		Local: LocalConfig{
			UploadPath: "./uploads",
			URLPrefix:  "/static/uploads",
		},

		OSS: OSSConfig{
			Endpoint:        "",
			AccessKeyID:     "",
			AccessKeySecret: "",
			BucketName:      "",
			URLPrefix:       "",
		},

		S3: S3Config{
			Region:          "us-east-1",
			AccessKeyID:     "",
			SecretAccessKey: "",
			BucketName:      "",
			URLPrefix:       "",
		},

		COS: COSConfig{
			Region:    "",
			SecretID:  "",
			SecretKey: "",
			Bucket:    "",
			URLPrefix: "",
		},

		Security: SecurityConfig{
			EnableVirusScan:   false,
			EnableContentScan: true,
			ForbiddenExts:     []string{".exe", ".bat", ".cmd", ".com", ".pif", ".scr", ".vbs", ".js"},
			MaxImageWidth:     4096,
			MaxImageHeight:    4096,
		},

		Thumbnail: ThumbnailConfig{
			Sizes: []ThumbnailSize{
				{Name: "small", Width: 150, Height: 150},
				{Name: "medium", Width: 300, Height: 300},
				{Name: "large", Width: 800, Height: 600},
			},
			Quality: 85,
			Format:  "jpeg",
		},
	}
}

// LoadConfigFromEnv 从环境变量加载配置
func LoadConfigFromEnv() *UploadConfig {
	config := DefaultUploadConfig()

	// 基础配置
	if storageType := os.Getenv("UPLOAD_STORAGE_TYPE"); storageType != "" {
		config.StorageType = StorageType(storageType)
	}

	if maxFileSize := os.Getenv("UPLOAD_MAX_FILE_SIZE"); maxFileSize != "" {
		if size, err := strconv.ParseInt(maxFileSize, 10, 64); err == nil {
			config.MaxFileSize = size
		}
	}

	if maxFiles := os.Getenv("UPLOAD_MAX_FILES"); maxFiles != "" {
		if count, err := strconv.Atoi(maxFiles); err == nil {
			config.MaxFiles = count
		}
	}

	if allowedTypes := os.Getenv("UPLOAD_ALLOWED_TYPES"); allowedTypes != "" {
		config.AllowedTypes = strings.Split(allowedTypes, ",")
	}

	if allowedExts := os.Getenv("UPLOAD_ALLOWED_EXTS"); allowedExts != "" {
		config.AllowedExts = strings.Split(allowedExts, ",")
	}

	// 本地存储配置
	if uploadPath := os.Getenv("UPLOAD_LOCAL_PATH"); uploadPath != "" {
		config.Local.UploadPath = uploadPath
	}

	if urlPrefix := os.Getenv("UPLOAD_LOCAL_URL_PREFIX"); urlPrefix != "" {
		config.Local.URLPrefix = urlPrefix
	}

	// OSS配置
	if endpoint := os.Getenv("OSS_ENDPOINT"); endpoint != "" {
		config.OSS.Endpoint = endpoint
	}

	if accessKeyID := os.Getenv("OSS_ACCESS_KEY_ID"); accessKeyID != "" {
		config.OSS.AccessKeyID = accessKeyID
	}

	if accessKeySecret := os.Getenv("OSS_ACCESS_KEY_SECRET"); accessKeySecret != "" {
		config.OSS.AccessKeySecret = accessKeySecret
	}

	if bucketName := os.Getenv("OSS_BUCKET_NAME"); bucketName != "" {
		config.OSS.BucketName = bucketName
	}

	if urlPrefix := os.Getenv("OSS_URL_PREFIX"); urlPrefix != "" {
		config.OSS.URLPrefix = urlPrefix
	}

	// S3配置
	if region := os.Getenv("S3_REGION"); region != "" {
		config.S3.Region = region
	}

	if accessKeyID := os.Getenv("S3_ACCESS_KEY_ID"); accessKeyID != "" {
		config.S3.AccessKeyID = accessKeyID
	}

	if secretAccessKey := os.Getenv("S3_SECRET_ACCESS_KEY"); secretAccessKey != "" {
		config.S3.SecretAccessKey = secretAccessKey
	}

	if bucketName := os.Getenv("S3_BUCKET_NAME"); bucketName != "" {
		config.S3.BucketName = bucketName
	}

	// 安全配置
	if enableSecurity := os.Getenv("UPLOAD_ENABLE_SECURITY"); enableSecurity != "" {
		config.EnableSecurity = enableSecurity == "true"
	}

	return config
}

// LoadConfigFromFile 从文件加载配置
func LoadConfigFromFile(filename string) (*UploadConfig, error) {
	// 如果文件不存在，返回默认配置
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return DefaultUploadConfig(), nil
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	config := DefaultUploadConfig()
	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return config, nil
}

// SaveToFile 保存配置到文件
func (c *UploadConfig) SaveToFile(filename string) error {
	// 确保目录存在
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}

// Validate 验证配置
func (c *UploadConfig) Validate() error {
	if c.MaxFileSize <= 0 {
		return fmt.Errorf("最大文件大小必须大于0")
	}

	if c.MaxFiles <= 0 {
		return fmt.Errorf("最大文件数必须大于0")
	}

	if len(c.AllowedTypes) == 0 && len(c.AllowedExts) == 0 {
		return fmt.Errorf("必须指定允许的文件类型或扩展名")
	}

	// 验证存储类型特定配置
	switch c.StorageType {
	case StorageTypeLocal:
		if c.Local.UploadPath == "" {
			return fmt.Errorf("本地存储路径不能为空")
		}
	case StorageTypeOSS:
		if c.OSS.Endpoint == "" || c.OSS.AccessKeyID == "" || c.OSS.AccessKeySecret == "" || c.OSS.BucketName == "" {
			return fmt.Errorf("OSS配置不完整")
		}
	case StorageTypeS3:
		if c.S3.AccessKeyID == "" || c.S3.SecretAccessKey == "" || c.S3.BucketName == "" {
			return fmt.Errorf("S3配置不完整")
		}
	case StorageTypeCOS:
		if c.COS.SecretID == "" || c.COS.SecretKey == "" || c.COS.Bucket == "" {
			return fmt.Errorf("COS配置不完整")
		}
	default:
		return fmt.Errorf("不支持的存储类型: %s", c.StorageType)
	}

	return nil
}

// IsAllowedType 检查文件类型是否允许
func (c *UploadConfig) IsAllowedType(contentType string) bool {
	if len(c.AllowedTypes) == 0 {
		return true
	}

	for _, allowedType := range c.AllowedTypes {
		if contentType == allowedType {
			return true
		}
	}

	return false
}

// IsAllowedExt 检查文件扩展名是否允许
func (c *UploadConfig) IsAllowedExt(filename string) bool {
	if len(c.AllowedExts) == 0 {
		return true
	}

	ext := strings.ToLower(filepath.Ext(filename))
	for _, allowedExt := range c.AllowedExts {
		if ext == strings.ToLower(allowedExt) {
			return true
		}
	}

	return false
}

// IsForbiddenExt 检查文件扩展名是否被禁止
func (c *UploadConfig) IsForbiddenExt(filename string) bool {
	if len(c.Security.ForbiddenExts) == 0 {
		return false
	}

	ext := strings.ToLower(filepath.Ext(filename))
	for _, forbiddenExt := range c.Security.ForbiddenExts {
		if ext == strings.ToLower(forbiddenExt) {
			return true
		}
	}

	return false
}

// GetUploadPath 获取上传路径
func (c *UploadConfig) GetUploadPath(subPath string) string {
	switch c.StorageType {
	case StorageTypeLocal:
		return filepath.Join(c.Local.UploadPath, subPath)
	default:
		return subPath
	}
}

// GetFileURL 获取文件访问URL
func (c *UploadConfig) GetFileURL(filePath string) string {
	switch c.StorageType {
	case StorageTypeLocal:
		return c.Local.URLPrefix + "/" + strings.TrimPrefix(filePath, c.Local.UploadPath)
	case StorageTypeOSS:
		return c.OSS.URLPrefix + "/" + filePath
	case StorageTypeS3:
		return c.S3.URLPrefix + "/" + filePath
	case StorageTypeCOS:
		return c.COS.URLPrefix + "/" + filePath
	default:
		return filePath
	}
}

// GenerateFileName 生成文件名
func (c *UploadConfig) GenerateFileName(originalName string) string {
	ext := filepath.Ext(originalName)
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%d%s", timestamp, ext)
}

// GetStorageConfig 获取存储配置摘要
func (c *UploadConfig) GetStorageConfig() map[string]interface{} {
	return map[string]interface{}{
		"storage_type":     c.StorageType,
		"max_file_size":    c.MaxFileSize,
		"max_files":        c.MaxFiles,
		"allowed_types":    c.AllowedTypes,
		"allowed_exts":     c.AllowedExts,
		"enable_security":  c.EnableSecurity,
		"enable_thumbnail": c.EnableThumbnail,
	}
}
