package fileupload

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"mall-go/internal/model"
)

// UploadConfig 文件上传配置
type UploadConfig struct {
	// 基础配置
	UploadPath string `json:"upload_path" yaml:"upload_path"`   // 上传路径
	BaseURL    string `json:"base_url" yaml:"base_url"`         // 基础URL
	
	// 存储配置
	StorageType    string `json:"storage_type" yaml:"storage_type"`       // 存储类型: local, oss, s3
	MaxConcurrency int    `json:"max_concurrency" yaml:"max_concurrency"` // 最大并发数
	
	// 文件大小限制（字节）
	MaxImageSize    int64 `json:"max_image_size" yaml:"max_image_size"`       // 图片最大大小
	MaxDocumentSize int64 `json:"max_document_size" yaml:"max_document_size"` // 文档最大大小
	MaxVideoSize    int64 `json:"max_video_size" yaml:"max_video_size"`       // 视频最大大小
	MaxFileCount    int   `json:"max_file_count" yaml:"max_file_count"`       // 单次最多上传文件数
	MaxTotalSize    int64 `json:"max_total_size" yaml:"max_total_size"`       // 批量上传总大小限制
	
	// 文件类型配置
	AllowedImageTypes    []string `json:"allowed_image_types" yaml:"allowed_image_types"`       // 允许的图片类型
	AllowedDocumentTypes []string `json:"allowed_document_types" yaml:"allowed_document_types"` // 允许的文档类型
	AllowedVideoTypes    []string `json:"allowed_video_types" yaml:"allowed_video_types"`       // 允许的视频类型
	
	// MIME类型配置
	ImageMimeTypes    []string `json:"image_mime_types" yaml:"image_mime_types"`       // 图片MIME类型
	DocumentMimeTypes []string `json:"document_mime_types" yaml:"document_mime_types"` // 文档MIME类型
	VideoMimeTypes    []string `json:"video_mime_types" yaml:"video_mime_types"`       // 视频MIME类型
	
	// 安全配置
	DangerousExts    []string `json:"dangerous_exts" yaml:"dangerous_exts"`       // 危险的扩展名
	EnableMagicCheck bool     `json:"enable_magic_check" yaml:"enable_magic_check"` // 是否启用魔数检查
	EnableVirusScan  bool     `json:"enable_virus_scan" yaml:"enable_virus_scan"`   // 是否启用病毒扫描
	
	// 性能配置
	EnableCompression bool    `json:"enable_compression" yaml:"enable_compression"` // 是否启用压缩
	CompressionQuality int   `json:"compression_quality" yaml:"compression_quality"` // 压缩质量(1-100)
	EnableThumbnail   bool    `json:"enable_thumbnail" yaml:"enable_thumbnail"`     // 是否生成缩略图
	ThumbnailSizes    []int   `json:"thumbnail_sizes" yaml:"thumbnail_sizes"`       // 缩略图尺寸
	
	// 清理配置
	EnableAutoCleanup bool `json:"enable_auto_cleanup" yaml:"enable_auto_cleanup"` // 是否启用自动清理
	CleanupDays       int  `json:"cleanup_days" yaml:"cleanup_days"`               // 清理天数
}

// DefaultUploadConfig 默认上传配置
func DefaultUploadConfig() *UploadConfig {
	return &UploadConfig{
		// 基础配置
		UploadPath: "./uploads",
		BaseURL:    "http://localhost:8080",
		
		// 存储配置
		StorageType:    "local",
		MaxConcurrency: 3,
		
		// 文件大小限制
		MaxImageSize:    model.MaxImageSize,
		MaxDocumentSize: model.MaxDocumentSize,
		MaxVideoSize:    model.MaxVideoSize,
		MaxFileCount:    model.MaxFileCount,
		MaxTotalSize:    100 * 1024 * 1024, // 100MB
		
		// 文件类型配置
		AllowedImageTypes:    model.SupportedImageTypes,
		AllowedDocumentTypes: model.SupportedDocumentTypes,
		AllowedVideoTypes:    model.SupportedVideoTypes,
		
		// MIME类型配置
		ImageMimeTypes:    model.ImageMimeTypes,
		DocumentMimeTypes: model.DocumentMimeTypes,
		VideoMimeTypes:    model.VideoMimeTypes,
		
		// 安全配置
		DangerousExts: []string{
			".exe", ".bat", ".cmd", ".com", ".pif", ".scr", ".vbs", ".js", ".jar",
			".php", ".asp", ".aspx", ".jsp", ".py", ".rb", ".pl", ".sh",
		},
		EnableMagicCheck: true,
		EnableVirusScan:  false,
		
		// 性能配置
		EnableCompression:  false,
		CompressionQuality: 80,
		EnableThumbnail:    false,
		ThumbnailSizes:     []int{150, 300, 600},
		
		// 清理配置
		EnableAutoCleanup: false,
		CleanupDays:       30,
	}
}

// LoadConfigFromFile 从文件加载配置
func LoadConfigFromFile(configPath string) (*UploadConfig, error) {
	// 如果文件不存在，返回默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultUploadConfig(), nil
	}
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}
	
	config := &UploadConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}
	
	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %v", err)
	}
	
	return config, nil
}

// LoadConfigFromEnv 从环境变量加载配置
func LoadConfigFromEnv() *UploadConfig {
	config := DefaultUploadConfig()
	
	// 从环境变量覆盖配置
	if uploadPath := os.Getenv("UPLOAD_PATH"); uploadPath != "" {
		config.UploadPath = uploadPath
	}
	
	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		config.BaseURL = baseURL
	}
	
	if storageType := os.Getenv("STORAGE_TYPE"); storageType != "" {
		config.StorageType = storageType
	}
	
	return config
}

// SaveToFile 保存配置到文件
func (c *UploadConfig) SaveToFile(configPath string) error {
	// 创建目录
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}
	
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}
	
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}
	
	return nil
}

// Validate 验证配置
func (c *UploadConfig) Validate() error {
	if c.UploadPath == "" {
		return fmt.Errorf("上传路径不能为空")
	}
	
	if c.BaseURL == "" {
		return fmt.Errorf("基础URL不能为空")
	}
	
	if c.MaxImageSize <= 0 {
		return fmt.Errorf("图片最大大小必须大于0")
	}
	
	if c.MaxDocumentSize <= 0 {
		return fmt.Errorf("文档最大大小必须大于0")
	}
	
	if c.MaxVideoSize <= 0 {
		return fmt.Errorf("视频最大大小必须大于0")
	}
	
	if c.MaxFileCount <= 0 {
		return fmt.Errorf("最大文件数量必须大于0")
	}
	
	if c.MaxConcurrency <= 0 {
		c.MaxConcurrency = 3 // 设置默认值
	}
	
	return nil
}

// GetMaxSizeByFileType 根据文件类型获取最大大小
func (c *UploadConfig) GetMaxSizeByFileType(fileType model.FileType) int64 {
	switch fileType {
	case model.FileTypeImage:
		return c.MaxImageSize
	case model.FileTypeDocument:
		return c.MaxDocumentSize
	case model.FileTypeVideo:
		return c.MaxVideoSize
	default:
		return c.MaxImageSize // 默认使用图片大小限制
	}
}

// ToValidationConfig 转换为验证配置
func (c *UploadConfig) ToValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		MaxImageSize:    c.MaxImageSize,
		MaxDocumentSize: c.MaxDocumentSize,
		MaxVideoSize:    c.MaxVideoSize,
		MaxFileCount:    c.MaxFileCount,
		AllowedTypes: append(append(c.AllowedImageTypes, c.AllowedDocumentTypes...), 
			c.AllowedVideoTypes...),
		AllowedMimes: append(append(c.ImageMimeTypes, c.DocumentMimeTypes...), 
			c.VideoMimeTypes...),
		DangerousExts:    c.DangerousExts,
		EnableMagicCheck: c.EnableMagicCheck,
	}
}
