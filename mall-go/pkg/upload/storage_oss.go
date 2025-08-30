package upload

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// OSSStorage 阿里云OSS存储实现
type OSSStorage struct {
	client *oss.Client
	bucket *oss.Bucket
	config *UploadConfig
}

// NewOSSStorage 创建OSS存储
func NewOSSStorage(config *UploadConfig) (*OSSStorage, error) {
	if config.OSS.Endpoint == "" || config.OSS.AccessKeyID == "" || 
	   config.OSS.AccessKeySecret == "" || config.OSS.BucketName == "" {
		return nil, fmt.Errorf("OSS配置不完整")
	}

	// 创建OSS客户端
	client, err := oss.New(config.OSS.Endpoint, config.OSS.AccessKeyID, config.OSS.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("创建OSS客户端失败: %v", err)
	}

	// 获取存储桶
	bucket, err := client.Bucket(config.OSS.BucketName)
	if err != nil {
		return nil, fmt.Errorf("获取OSS存储桶失败: %v", err)
	}

	return &OSSStorage{
		client: client,
		bucket: bucket,
		config: config,
	}, nil
}

// Upload 上传文件到OSS
func (os *OSSStorage) Upload(filePath string, reader io.Reader, size int64) error {
	// 设置上传选项
	options := []oss.Option{
		oss.ContentLength(size),
		oss.ContentType(getContentType(filePath)),
	}

	// 上传文件
	err := os.bucket.PutObject(filePath, reader, options...)
	if err != nil {
		return fmt.Errorf("上传文件到OSS失败: %v", err)
	}

	return nil
}

// Download 从OSS下载文件
func (os *OSSStorage) Download(filePath string) (io.ReadCloser, error) {
	reader, err := os.bucket.GetObject(filePath)
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchKey") {
			return nil, fmt.Errorf("文件不存在: %s", filePath)
		}
		return nil, fmt.Errorf("从OSS下载文件失败: %v", err)
	}

	return reader, nil
}

// Delete 删除OSS文件
func (os *OSSStorage) Delete(filePath string) error {
	err := os.bucket.DeleteObject(filePath)
	if err != nil {
		return fmt.Errorf("删除OSS文件失败: %v", err)
	}

	return nil
}

// Exists 检查OSS文件是否存在
func (os *OSSStorage) Exists(filePath string) (bool, error) {
	exists, err := os.bucket.IsObjectExist(filePath)
	if err != nil {
		return false, fmt.Errorf("检查OSS文件失败: %v", err)
	}

	return exists, nil
}

// GetFileInfo 获取OSS文件信息
func (os *OSSStorage) GetFileInfo(filePath string) (*FileInfo, error) {
	header, err := os.bucket.GetObjectMeta(filePath)
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchKey") {
			return nil, fmt.Errorf("文件不存在: %s", filePath)
		}
		return nil, fmt.Errorf("获取OSS文件信息失败: %v", err)
	}

	// 解析文件大小
	var size int64
	if contentLength := header.Get("Content-Length"); contentLength != "" {
		fmt.Sscanf(contentLength, "%d", &size)
	}

	// 解析最后修改时间
	var lastModified time.Time
	if lastModifiedStr := header.Get("Last-Modified"); lastModifiedStr != "" {
		lastModified, _ = time.Parse(time.RFC1123, lastModifiedStr)
	}

	return &FileInfo{
		Path:         filePath,
		Name:         getFileName(filePath),
		Size:         size,
		ContentType:  header.Get("Content-Type"),
		LastModified: lastModified,
		ETag:         strings.Trim(header.Get("ETag"), "\""),
		URL:          os.GetURL(filePath),
	}, nil
}

// GetURL 获取OSS文件访问URL
func (os *OSSStorage) GetURL(filePath string) string {
	if os.config.OSS.URLPrefix != "" {
		return os.config.OSS.URLPrefix + "/" + filePath
	}
	
	// 生成默认URL
	return fmt.Sprintf("https://%s.%s/%s", os.config.OSS.BucketName, os.config.OSS.Endpoint, filePath)
}

// ListFiles 列出OSS文件
func (os *OSSStorage) ListFiles(prefix string, limit int) ([]*FileInfo, error) {
	options := []oss.Option{
		oss.Prefix(prefix),
	}
	
	if limit > 0 {
		options = append(options, oss.MaxKeys(limit))
	}

	result, err := os.bucket.ListObjects(options...)
	if err != nil {
		return nil, fmt.Errorf("列出OSS文件失败: %v", err)
	}

	var files []*FileInfo
	for _, object := range result.Objects {
		files = append(files, &FileInfo{
			Path:         object.Key,
			Name:         getFileName(object.Key),
			Size:         object.Size,
			ContentType:  object.Type,
			LastModified: object.LastModified,
			ETag:         strings.Trim(object.ETag, "\""),
			URL:          os.GetURL(object.Key),
		})
	}

	return files, nil
}

// S3Storage AWS S3存储实现（简化版，需要AWS SDK）
type S3Storage struct {
	config *UploadConfig
}

// NewS3Storage 创建S3存储
func NewS3Storage(config *UploadConfig) (*S3Storage, error) {
	// TODO: 实现AWS S3存储
	// 需要引入 github.com/aws/aws-sdk-go/service/s3
	return nil, fmt.Errorf("S3存储暂未实现，请使用本地存储或OSS")
}

// Upload 上传文件到S3
func (s3 *S3Storage) Upload(filePath string, reader io.Reader, size int64) error {
	return fmt.Errorf("S3存储暂未实现")
}

// Download 从S3下载文件
func (s3 *S3Storage) Download(filePath string) (io.ReadCloser, error) {
	return nil, fmt.Errorf("S3存储暂未实现")
}

// Delete 删除S3文件
func (s3 *S3Storage) Delete(filePath string) error {
	return fmt.Errorf("S3存储暂未实现")
}

// Exists 检查S3文件是否存在
func (s3 *S3Storage) Exists(filePath string) (bool, error) {
	return false, fmt.Errorf("S3存储暂未实现")
}

// GetFileInfo 获取S3文件信息
func (s3 *S3Storage) GetFileInfo(filePath string) (*FileInfo, error) {
	return nil, fmt.Errorf("S3存储暂未实现")
}

// GetURL 获取S3文件访问URL
func (s3 *S3Storage) GetURL(filePath string) string {
	return filePath
}

// ListFiles 列出S3文件
func (s3 *S3Storage) ListFiles(prefix string, limit int) ([]*FileInfo, error) {
	return nil, fmt.Errorf("S3存储暂未实现")
}

// COSStorage 腾讯云COS存储实现（简化版，需要COS SDK）
type COSStorage struct {
	config *UploadConfig
}

// NewCOSStorage 创建COS存储
func NewCOSStorage(config *UploadConfig) (*COSStorage, error) {
	// TODO: 实现腾讯云COS存储
	// 需要引入 github.com/tencentyun/cos-go-sdk-v5
	return nil, fmt.Errorf("COS存储暂未实现，请使用本地存储或OSS")
}

// Upload 上传文件到COS
func (cos *COSStorage) Upload(filePath string, reader io.Reader, size int64) error {
	return fmt.Errorf("COS存储暂未实现")
}

// Download 从COS下载文件
func (cos *COSStorage) Download(filePath string) (io.ReadCloser, error) {
	return nil, fmt.Errorf("COS存储暂未实现")
}

// Delete 删除COS文件
func (cos *COSStorage) Delete(filePath string) error {
	return fmt.Errorf("COS存储暂未实现")
}

// Exists 检查COS文件是否存在
func (cos *COSStorage) Exists(filePath string) (bool, error) {
	return false, fmt.Errorf("COS存储暂未实现")
}

// GetFileInfo 获取COS文件信息
func (cos *COSStorage) GetFileInfo(filePath string) (*FileInfo, error) {
	return nil, fmt.Errorf("COS存储暂未实现")
}

// GetURL 获取COS文件访问URL
func (cos *COSStorage) GetURL(filePath string) string {
	return filePath
}

// ListFiles 列出COS文件
func (cos *COSStorage) ListFiles(prefix string, limit int) ([]*FileInfo, error) {
	return nil, fmt.Errorf("COS存储暂未实现")
}

// 辅助函数

// getFileName 从文件路径获取文件名
func getFileName(filePath string) string {
	parts := strings.Split(filePath, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return filePath
}
