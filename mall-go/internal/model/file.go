package model

import (
	"time"

	"gorm.io/gorm"
)

// FileType 文件类型枚举
type FileType string

const (
	FileTypeImage    FileType = "image"    // 图片文件
	FileTypeDocument FileType = "document" // 文档文件
	FileTypeVideo    FileType = "video"    // 视频文件
	FileTypeOther    FileType = "other"    // 其他文件
)

// StorageType 存储类型枚举
type StorageType string

const (
	StorageTypeLocal StorageType = "local" // 本地存储
	StorageTypeOSS   StorageType = "oss"   // 阿里云OSS
	StorageTypeCOS   StorageType = "cos"   // 腾讯云COS
	StorageTypeS3    StorageType = "s3"    // AWS S3
)

// FileStatus 文件状态枚举
type FileStatus string

const (
	FileStatusUploading FileStatus = "uploading" // 上传中
	FileStatusSuccess   FileStatus = "success"   // 上传成功
	FileStatusFailed    FileStatus = "failed"    // 上传失败
	FileStatusDeleted   FileStatus = "deleted"   // 已删除
)

// BusinessType 业务类型枚举
type BusinessType string

const (
	BusinessTypeAvatar  BusinessType = "avatar"  // 用户头像
	BusinessTypeProduct BusinessType = "product" // 商品图片
	BusinessTypeStore   BusinessType = "store"   // 店铺图片
	BusinessTypeOther   BusinessType = "other"   // 其他业务
)

// File 文件信息模型
type File struct {
	ID           uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	UUID         string         `json:"uuid" gorm:"type:varchar(36);uniqueIndex;not null;comment:文件唯一标识"`
	OriginalName string         `json:"original_name" gorm:"type:varchar(255);not null;comment:原始文件名"`
	StoredName   string         `json:"stored_name" gorm:"type:varchar(255);not null;comment:存储文件名"`
	FilePath     string         `json:"file_path" gorm:"type:varchar(500);not null;comment:文件路径"`
	FileSize     int64          `json:"file_size" gorm:"not null;comment:文件大小(字节)"`
	MimeType     string         `json:"mime_type" gorm:"type:varchar(100);not null;comment:MIME类型"`
	FileType     FileType       `json:"file_type" gorm:"type:varchar(20);not null;comment:文件类型"`
	StorageType  StorageType    `json:"storage_type" gorm:"type:varchar(20);default:local;comment:存储类型"`
	UploadUserID uint           `json:"upload_user_id" gorm:"not null;index;comment:上传用户ID"`
	BusinessType BusinessType   `json:"business_type" gorm:"type:varchar(50);index:idx_business;comment:业务类型"`
	BusinessID   *uint          `json:"business_id" gorm:"index:idx_business;comment:关联业务ID"`
	AccessURL    string         `json:"access_url" gorm:"type:varchar(500);comment:访问URL"`
	IsPublic     bool           `json:"is_public" gorm:"default:false;comment:是否公开访问"`
	Status       FileStatus     `json:"status" gorm:"type:varchar(20);default:uploading;index"`
	CreatedAt    time.Time      `json:"created_at" gorm:"index"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// 关联关系
	UploadUser User `json:"upload_user,omitempty" gorm:"foreignKey:UploadUserID;references:ID"`
}

// TableName 指定表名
func (File) TableName() string {
	return "files"
}

// IsImage 检查文件是否为图片
func (f *File) IsImage() bool {
	imageTypes := []string{"image/jpeg", "image/jpg", "image/png", "image/gif", "image/webp", "image/bmp"}
	for _, imageType := range imageTypes {
		if f.MimeType == imageType {
			return true
		}
	}
	return false
}

// GetURL 获取文件访问URL (兼容性方法)
func (f *File) GetURL() string {
	return f.AccessURL
}

// URL 获取文件访问URL (属性访问器)
func (f *File) URL() string {
	return f.AccessURL
}

// FileUploadRequest 文件上传请求结构
type FileUploadRequest struct {
	BusinessType BusinessType `json:"business_type" binding:"required" example:"product"`
	BusinessID   *uint        `json:"business_id" example:"123"`
	IsPublic     bool         `json:"is_public" example:"true"`
	Description  string       `json:"description" example:"商品主图"`
}

// FileUploadResponse 文件上传响应结构
type FileUploadResponse struct {
	ID           uint         `json:"id"`
	UUID         string       `json:"uuid"`
	OriginalName string       `json:"original_name"`
	FileSize     int64        `json:"file_size"`
	MimeType     string       `json:"mime_type"`
	FileType     FileType     `json:"file_type"`
	BusinessType BusinessType `json:"business_type"`
	BusinessID   *uint        `json:"business_id"`
	AccessURL    string       `json:"access_url"`
	IsPublic     bool         `json:"is_public"`
	Status       FileStatus   `json:"status"`
	CreatedAt    time.Time    `json:"created_at"`
}

// FileListRequest 文件列表请求结构
type FileListRequest struct {
	BusinessType BusinessType `form:"business_type" example:"product"`
	BusinessID   *uint        `form:"business_id" example:"123"`
	FileType     FileType     `form:"file_type" example:"image"`
	IsPublic     *bool        `form:"is_public" example:"true"`
	Page         int          `form:"page" example:"1"`
	PageSize     int          `form:"page_size" example:"10"`
}

// FileInfo 文件信息结构（用于列表展示）
type FileInfo struct {
	ID           uint         `json:"id"`
	UUID         string       `json:"uuid"`
	OriginalName string       `json:"original_name"`
	FileSize     int64        `json:"file_size"`
	MimeType     string       `json:"mime_type"`
	FileType     FileType     `json:"file_type"`
	BusinessType BusinessType `json:"business_type"`
	BusinessID   *uint        `json:"business_id"`
	AccessURL    string       `json:"access_url"`
	IsPublic     bool         `json:"is_public"`
	Status       FileStatus   `json:"status"`
	CreatedAt    time.Time    `json:"created_at"`
}

// 文件类型验证相关常量
var (
	// SupportedImageTypes 支持的图片类型
	SupportedImageTypes = []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}

	// SupportedDocumentTypes 支持的文档类型
	SupportedDocumentTypes = []string{".pdf", ".doc", ".docx", ".txt"}

	// SupportedVideoTypes 支持的视频类型
	SupportedVideoTypes = []string{".mp4", ".avi", ".mov", ".wmv"}

	// ImageMimeTypes 图片MIME类型
	ImageMimeTypes = []string{
		"image/jpeg", "image/jpg", "image/png",
		"image/gif", "image/webp",
	}

	// DocumentMimeTypes 文档MIME类型
	DocumentMimeTypes = []string{
		"application/pdf", "application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"text/plain",
	}

	// VideoMimeTypes 视频MIME类型
	VideoMimeTypes = []string{
		"video/mp4", "video/avi", "video/quicktime", "video/x-ms-wmv",
	}
)

// 文件大小限制常量（字节）
const (
	MaxImageSize    = 5 * 1024 * 1024  // 5MB
	MaxDocumentSize = 10 * 1024 * 1024 // 10MB
	MaxVideoSize    = 50 * 1024 * 1024 // 50MB
	MaxFileCount    = 10               // 单次最多上传10个文件
)

// GetFileTypeByMime 根据MIME类型获取文件类型
func GetFileTypeByMime(mimeType string) FileType {
	for _, mime := range ImageMimeTypes {
		if mime == mimeType {
			return FileTypeImage
		}
	}

	for _, mime := range DocumentMimeTypes {
		if mime == mimeType {
			return FileTypeDocument
		}
	}

	for _, mime := range VideoMimeTypes {
		if mime == mimeType {
			return FileTypeVideo
		}
	}

	return FileTypeOther
}

// GetMaxSizeByType 根据文件类型获取最大大小限制
func GetMaxSizeByType(fileType FileType) int64 {
	switch fileType {
	case FileTypeImage:
		return MaxImageSize
	case FileTypeDocument:
		return MaxDocumentSize
	case FileTypeVideo:
		return MaxVideoSize
	default:
		return MaxImageSize // 默认使用图片限制
	}
}

// IsValidBusinessType 验证业务类型是否有效
func IsValidBusinessType(businessType BusinessType) bool {
	validTypes := []BusinessType{
		BusinessTypeAvatar, BusinessTypeProduct,
		BusinessTypeStore, BusinessTypeOther,
	}

	for _, validType := range validTypes {
		if businessType == validType {
			return true
		}
	}
	return false
}

// IsValidFileType 验证文件类型是否有效
func IsValidFileType(fileType FileType) bool {
	validTypes := []FileType{
		FileTypeImage, FileTypeDocument,
		FileTypeVideo, FileTypeOther,
	}

	for _, validType := range validTypes {
		if fileType == validType {
			return true
		}
	}
	return false
}

// IsValidStorageType 验证存储类型是否有效
func IsValidStorageType(storageType StorageType) bool {
	validTypes := []StorageType{
		StorageTypeLocal, StorageTypeOSS,
		StorageTypeCOS, StorageTypeS3,
	}

	for _, validType := range validTypes {
		if storageType == validType {
			return true
		}
	}
	return false
}
