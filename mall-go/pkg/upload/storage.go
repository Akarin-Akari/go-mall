package upload

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// StorageInterface 存储接口
type StorageInterface interface {
	// Upload 上传文件
	Upload(filePath string, reader io.Reader, size int64) error

	// Download 下载文件
	Download(filePath string) (io.ReadCloser, error)

	// Delete 删除文件
	Delete(filePath string) error

	// Exists 检查文件是否存在
	Exists(filePath string) (bool, error)

	// GetFileInfo 获取文件信息
	GetFileInfo(filePath string) (*FileInfo, error)

	// GetURL 获取文件访问URL
	GetURL(filePath string) string

	// ListFiles 列出文件
	ListFiles(prefix string, limit int) ([]*FileInfo, error)
}

// FileInfo 文件信息
type FileInfo struct {
	Path         string    `json:"path"`          // 文件路径
	Name         string    `json:"name"`          // 文件名
	Size         int64     `json:"size"`          // 文件大小
	ContentType  string    `json:"content_type"`  // 文件类型
	LastModified time.Time `json:"last_modified"` // 最后修改时间
	ETag         string    `json:"etag"`          // 文件标识
	URL          string    `json:"url"`           // 访问URL
}

// StorageManager 存储管理器
type StorageManager struct {
	storage StorageInterface
	config  *UploadConfig
}

// NewStorageManager 创建存储管理器
func NewStorageManager(config *UploadConfig) (*StorageManager, error) {
	var storage StorageInterface
	var err error

	switch config.StorageType {
	case StorageTypeLocal:
		storage, err = NewLocalStorage(config)
	case StorageTypeOSS:
		storage, err = NewOSSStorage(config)
	case StorageTypeS3:
		storage, err = NewS3Storage(config)
	case StorageTypeCOS:
		storage, err = NewCOSStorage(config)
	default:
		return nil, fmt.Errorf("不支持的存储类型: %s", config.StorageType)
	}

	if err != nil {
		return nil, fmt.Errorf("创建存储服务失败: %v", err)
	}

	return &StorageManager{
		storage: storage,
		config:  config,
	}, nil
}

// Upload 上传文件
func (sm *StorageManager) Upload(filePath string, reader io.Reader, size int64) error {
	return sm.storage.Upload(filePath, reader, size)
}

// Download 下载文件
func (sm *StorageManager) Download(filePath string) (io.ReadCloser, error) {
	return sm.storage.Download(filePath)
}

// Delete 删除文件
func (sm *StorageManager) Delete(filePath string) error {
	return sm.storage.Delete(filePath)
}

// Exists 检查文件是否存在
func (sm *StorageManager) Exists(filePath string) (bool, error) {
	return sm.storage.Exists(filePath)
}

// GetFileInfo 获取文件信息
func (sm *StorageManager) GetFileInfo(filePath string) (*FileInfo, error) {
	return sm.storage.GetFileInfo(filePath)
}

// GetURL 获取文件访问URL
func (sm *StorageManager) GetURL(filePath string) string {
	return sm.storage.GetURL(filePath)
}

// ListFiles 列出文件
func (sm *StorageManager) ListFiles(prefix string, limit int) ([]*FileInfo, error) {
	return sm.storage.ListFiles(prefix, limit)
}

// GetStorage 获取底层存储接口
func (sm *StorageManager) GetStorage() StorageInterface {
	return sm.storage
}

// GetStorageType 获取存储类型
func (sm *StorageManager) GetStorageType() StorageType {
	return sm.config.StorageType
}

// LocalStorage 本地存储实现
type LocalStorage struct {
	config *UploadConfig
}

// NewLocalStorage 创建本地存储
func NewLocalStorage(config *UploadConfig) (*LocalStorage, error) {
	// 确保上传目录存在
	if err := os.MkdirAll(config.Local.UploadPath, 0755); err != nil {
		return nil, fmt.Errorf("创建上传目录失败: %v", err)
	}

	return &LocalStorage{
		config: config,
	}, nil
}

// Upload 上传文件到本地
func (ls *LocalStorage) Upload(filePath string, reader io.Reader, size int64) error {
	fullPath := filepath.Join(ls.config.Local.UploadPath, filePath)

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 创建文件
	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	// 复制数据
	_, err = io.Copy(file, reader)
	if err != nil {
		// 删除部分写入的文件
		os.Remove(fullPath)
		return fmt.Errorf("写入文件失败: %v", err)
	}

	return nil
}

// Download 从本地下载文件
func (ls *LocalStorage) Download(filePath string) (io.ReadCloser, error) {
	fullPath := filepath.Join(ls.config.Local.UploadPath, filePath)

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("文件不存在: %s", filePath)
		}
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}

	return file, nil
}

// Delete 删除本地文件
func (ls *LocalStorage) Delete(filePath string) error {
	fullPath := filepath.Join(ls.config.Local.UploadPath, filePath)

	err := os.Remove(fullPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除文件失败: %v", err)
	}

	return nil
}

// Exists 检查本地文件是否存在
func (ls *LocalStorage) Exists(filePath string) (bool, error) {
	fullPath := filepath.Join(ls.config.Local.UploadPath, filePath)

	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("检查文件失败: %v", err)
	}

	return true, nil
}

// GetFileInfo 获取本地文件信息
func (ls *LocalStorage) GetFileInfo(filePath string) (*FileInfo, error) {
	fullPath := filepath.Join(ls.config.Local.UploadPath, filePath)

	stat, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("文件不存在: %s", filePath)
		}
		return nil, fmt.Errorf("获取文件信息失败: %v", err)
	}

	return &FileInfo{
		Path:         filePath,
		Name:         stat.Name(),
		Size:         stat.Size(),
		ContentType:  getContentType(filePath),
		LastModified: stat.ModTime(),
		ETag:         fmt.Sprintf("%d-%d", stat.Size(), stat.ModTime().Unix()),
		URL:          ls.GetURL(filePath),
	}, nil
}

// GetURL 获取本地文件访问URL
func (ls *LocalStorage) GetURL(filePath string) string {
	return ls.config.Local.URLPrefix + "/" + filePath
}

// ListFiles 列出本地文件
func (ls *LocalStorage) ListFiles(prefix string, limit int) ([]*FileInfo, error) {
	searchPath := filepath.Join(ls.config.Local.UploadPath, prefix)

	var files []*FileInfo
	count := 0

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if limit > 0 && count >= limit {
			return filepath.SkipDir
		}

		// 计算相对路径
		relPath, err := filepath.Rel(ls.config.Local.UploadPath, path)
		if err != nil {
			return err
		}

		files = append(files, &FileInfo{
			Path:         filepath.ToSlash(relPath),
			Name:         info.Name(),
			Size:         info.Size(),
			ContentType:  getContentType(path),
			LastModified: info.ModTime(),
			ETag:         fmt.Sprintf("%d-%d", info.Size(), info.ModTime().Unix()),
			URL:          ls.GetURL(filepath.ToSlash(relPath)),
		})

		count++
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("列出文件失败: %v", err)
	}

	return files, nil
}

// getContentType 根据文件扩展名获取内容类型
func getContentType(filename string) string {
	ext := filepath.Ext(filename)
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".pdf":
		return "application/pdf"
	case ".txt":
		return "text/plain"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".zip":
		return "application/zip"
	case ".mp4":
		return "video/mp4"
	case ".mp3":
		return "audio/mpeg"
	default:
		return "application/octet-stream"
	}
}

// 全局存储管理器
var globalStorageManager *StorageManager

// InitGlobalStorageManager 初始化全局存储管理器
func InitGlobalStorageManager(config *UploadConfig) error {
	var err error
	globalStorageManager, err = NewStorageManager(config)
	return err
}

// GetGlobalStorageManager 获取全局存储管理器
func GetGlobalStorageManager() *StorageManager {
	return globalStorageManager
}

// 全局便捷函数

// UploadFile 上传文件（全局函数）
func UploadFile(filePath string, reader io.Reader, size int64) error {
	if globalStorageManager == nil {
		return fmt.Errorf("存储管理器未初始化")
	}
	return globalStorageManager.Upload(filePath, reader, size)
}

// DownloadFile 下载文件（全局函数）
func DownloadFile(filePath string) (io.ReadCloser, error) {
	if globalStorageManager == nil {
		return nil, fmt.Errorf("存储管理器未初始化")
	}
	return globalStorageManager.Download(filePath)
}

// DeleteFile 删除文件（全局函数）
func DeleteFile(filePath string) error {
	if globalStorageManager == nil {
		return fmt.Errorf("存储管理器未初始化")
	}
	return globalStorageManager.Delete(filePath)
}

// FileExists 检查文件是否存在（全局函数）
func FileExists(filePath string) (bool, error) {
	if globalStorageManager == nil {
		return false, fmt.Errorf("存储管理器未初始化")
	}
	return globalStorageManager.Exists(filePath)
}
