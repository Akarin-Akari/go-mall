package upload

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"mall-go/internal/model"

	"gorm.io/gorm"
)

// FileManager 文件管理器
type FileManager struct {
	db             *gorm.DB
	storageManager *StorageManager
	configManager  *ConfigManager
	validator      *FileValidator
}

// NewFileManager 创建文件管理器
func NewFileManager(db *gorm.DB, configManager *ConfigManager) (*FileManager, error) {
	config := configManager.GetConfig()

	// 创建存储管理器
	storageManager, err := NewStorageManager(config)
	if err != nil {
		return nil, fmt.Errorf("创建存储管理器失败: %v", err)
	}

	// 创建文件验证器
	validator := NewFileValidator(config)

	return &FileManager{
		db:             db,
		storageManager: storageManager,
		configManager:  configManager,
		validator:      validator,
	}, nil
}

// UploadFileRequest 上传文件请求
type UploadFileRequest struct {
	UserID      uint   `json:"user_id"`
	Filename    string `json:"filename"`
	Reader      io.Reader
	Size        int64  `json:"size"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

// UploadFileResponse 上传文件响应
type UploadFileResponse struct {
	FileID       uint      `json:"file_id"`
	OriginalName string    `json:"original_name"`
	FileName     string    `json:"file_name"`
	FilePath     string    `json:"file_path"`
	FileSize     int64     `json:"file_size"`
	ContentType  string    `json:"content_type"`
	Category     string    `json:"category"`
	URL          string    `json:"url"`
	UploadedAt   time.Time `json:"uploaded_at"`
}

// UploadFile 上传文件
func (fm *FileManager) UploadFile(req *UploadFileRequest) (*UploadFileResponse, error) {
	// 验证文件
	validationResult, err := fm.validator.ValidateFile(req.Filename, req.Reader, req.Size)
	if err != nil {
		return nil, fmt.Errorf("文件验证失败: %v", err)
	}

	if !validationResult.Valid {
		return nil, fmt.Errorf("文件验证失败: %s", strings.Join(validationResult.Errors, "; "))
	}

	// 生成文件路径
	fileName := fm.configManager.GenerateFileName(req.Filename)
	filePath := fm.generateFilePath(req.Category, fileName)

	// 上传文件到存储
	if err := fm.storageManager.Upload(filePath, req.Reader, req.Size); err != nil {
		return nil, fmt.Errorf("文件上传失败: %v", err)
	}

	// 保存文件记录到数据库
	fileRecord := &model.File{
		UploadUserID: req.UserID,
		OriginalName: req.Filename,
		StoredName:   fileName,
		FilePath:     filePath,
		FileSize:     req.Size,
		MimeType:     validationResult.ContentType,
		BusinessType: model.BusinessType(req.Category),
		AccessURL:    fm.storageManager.GetURL(filePath),
		Status:       model.FileStatusSuccess,
		FileType:     model.GetFileTypeByMime(validationResult.ContentType),
	}

	if err := fm.db.Create(fileRecord).Error; err != nil {
		// 如果数据库保存失败，删除已上传的文件
		fm.storageManager.Delete(filePath)
		return nil, fmt.Errorf("保存文件记录失败: %v", err)
	}

	return &UploadFileResponse{
		FileID:       fileRecord.ID,
		OriginalName: fileRecord.OriginalName,
		FileName:     fileRecord.StoredName,
		FilePath:     fileRecord.FilePath,
		FileSize:     fileRecord.FileSize,
		ContentType:  fileRecord.MimeType,
		Category:     string(fileRecord.BusinessType),
		URL:          fileRecord.AccessURL,
		UploadedAt:   fileRecord.CreatedAt,
	}, nil
}

// GetFileInfo 获取文件信息
func (fm *FileManager) GetFileInfo(fileID uint) (*model.File, error) {
	var file model.File
	if err := fm.db.First(&file, fileID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("文件不存在")
		}
		return nil, fmt.Errorf("查询文件失败: %v", err)
	}

	return &file, nil
}

// GetFileByPath 根据路径获取文件信息
func (fm *FileManager) GetFileByPath(filePath string) (*model.File, error) {
	var file model.File
	if err := fm.db.Where("file_path = ?", filePath).First(&file).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("文件不存在")
		}
		return nil, fmt.Errorf("查询文件失败: %v", err)
	}

	return &file, nil
}

// ListFiles 获取文件列表
func (fm *FileManager) ListFiles(userID uint, category string, page, pageSize int) ([]*model.File, int64, error) {
	query := fm.db.Model(&model.File{}).Where("status = ?", "active")

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	if category != "" {
		query = query.Where("category = ?", category)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询文件总数失败: %v", err)
	}

	// 获取文件列表
	var files []*model.File
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&files).Error; err != nil {
		return nil, 0, fmt.Errorf("查询文件列表失败: %v", err)
	}

	return files, total, nil
}

// DeleteFile 删除文件
func (fm *FileManager) DeleteFile(fileID uint, userID uint, isAdmin bool) error {
	var file model.File
	if err := fm.db.First(&file, fileID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("文件不存在")
		}
		return fmt.Errorf("查询文件失败: %v", err)
	}

	// 检查权限（只能删除自己的文件，除非是管理员）
	if file.UserID != userID && !isAdmin {
		return fmt.Errorf("无权限删除此文件")
	}

	// 从存储中删除文件
	if err := fm.storageManager.Delete(file.FilePath); err != nil {
		return fmt.Errorf("删除存储文件失败: %v", err)
	}

	// 从数据库中删除记录
	if err := fm.db.Delete(&file).Error; err != nil {
		return fmt.Errorf("删除文件记录失败: %v", err)
	}

	return nil
}

// DownloadFile 下载文件
func (fm *FileManager) DownloadFile(fileID uint) (io.ReadCloser, *model.File, error) {
	// 获取文件信息
	file, err := fm.GetFileInfo(fileID)
	if err != nil {
		return nil, nil, err
	}

	// 从存储中下载文件
	reader, err := fm.storageManager.Download(file.FilePath)
	if err != nil {
		return nil, nil, fmt.Errorf("下载文件失败: %v", err)
	}

	return reader, file, nil
}

// GetFileURL 获取文件访问URL
func (fm *FileManager) GetFileURL(fileID uint) (string, error) {
	file, err := fm.GetFileInfo(fileID)
	if err != nil {
		return "", err
	}

	return fm.storageManager.GetURL(file.FilePath), nil
}

// UpdateFileInfo 更新文件信息
func (fm *FileManager) UpdateFileInfo(fileID uint, userID uint, updates map[string]interface{}) error {
	var file model.File
	if err := fm.db.First(&file, fileID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("文件不存在")
		}
		return fmt.Errorf("查询文件失败: %v", err)
	}

	// 检查权限
	if file.UserID != userID {
		return fmt.Errorf("无权限修改此文件")
	}

	// 更新文件信息
	if err := fm.db.Model(&file).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新文件信息失败: %v", err)
	}

	return nil
}

// GetFileStatistics 获取文件统计信息
func (fm *FileManager) GetFileStatistics(userID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总文件数
	var totalFiles int64
	query := fm.db.Model(&model.File{}).Where("status = ?", "active")
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	query.Count(&totalFiles)
	stats["total_files"] = totalFiles

	// 总文件大小
	var totalSize int64
	query = fm.db.Model(&model.File{}).Where("status = ?", "active")
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	query.Select("COALESCE(SUM(file_size), 0)").Scan(&totalSize)
	stats["total_size"] = totalSize

	// 按分类统计
	var categoryStats []struct {
		Category string `json:"category"`
		Count    int64  `json:"count"`
		Size     int64  `json:"size"`
	}
	query = fm.db.Model(&model.File{}).Where("status = ?", "active")
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	query.Select("category, COUNT(*) as count, COALESCE(SUM(file_size), 0) as size").
		Group("category").
		Scan(&categoryStats)
	stats["category_stats"] = categoryStats

	// 今日上传数
	var todayUploads int64
	query = fm.db.Model(&model.File{}).Where("status = ? AND DATE(created_at) = CURDATE()", "active")
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	query.Count(&todayUploads)
	stats["today_uploads"] = todayUploads

	return stats, nil
}

// CleanupInvalidFiles 清理无效文件
func (fm *FileManager) CleanupInvalidFiles() (map[string]interface{}, error) {
	result := map[string]interface{}{
		"checked_files": 0,
		"invalid_files": 0,
		"cleaned_files": 0,
		"cleaned_size":  int64(0),
	}

	// 获取所有活跃文件
	var files []model.File
	if err := fm.db.Where("status = ?", "active").Find(&files).Error; err != nil {
		return nil, fmt.Errorf("查询文件失败: %v", err)
	}

	result["checked_files"] = len(files)

	for _, file := range files {
		// 检查文件是否存在
		exists, err := fm.storageManager.Exists(file.FilePath)
		if err != nil {
			continue
		}

		if !exists {
			result["invalid_files"] = result["invalid_files"].(int) + 1

			// 标记文件为已删除
			if err := fm.db.Model(&file).Update("status", "deleted").Error; err == nil {
				result["cleaned_files"] = result["cleaned_files"].(int) + 1
				result["cleaned_size"] = result["cleaned_size"].(int64) + file.FileSize
			}
		}
	}

	return result, nil
}

// generateFilePath 生成文件路径
func (fm *FileManager) generateFilePath(category, fileName string) string {
	now := time.Now()
	datePath := fmt.Sprintf("%d/%02d/%02d", now.Year(), now.Month(), now.Day())

	if category != "" && category != "general" {
		return filepath.Join(category, datePath, fileName)
	}

	return filepath.Join(datePath, fileName)
}

// 全局文件管理器
var globalFileManager *FileManager

// InitGlobalFileManager 初始化全局文件管理器
func InitGlobalFileManager(db *gorm.DB, configManager *ConfigManager) error {
	var err error
	globalFileManager, err = NewFileManager(db, configManager)
	return err
}

// GetGlobalFileManager 获取全局文件管理器
func GetGlobalFileManager() *FileManager {
	return globalFileManager
}

// 全局便捷函数

// UploadFileGlobal 上传文件（全局函数）
func UploadFileGlobal(req *UploadFileRequest) (*UploadFileResponse, error) {
	if globalFileManager == nil {
		return nil, fmt.Errorf("文件管理器未初始化")
	}
	return globalFileManager.UploadFile(req)
}

// GetFileInfoGlobal 获取文件信息（全局函数）
func GetFileInfoGlobal(fileID uint) (*model.File, error) {
	if globalFileManager == nil {
		return nil, fmt.Errorf("文件管理器未初始化")
	}
	return globalFileManager.GetFileInfo(fileID)
}

// DeleteFileGlobal 删除文件（全局函数）
func DeleteFileGlobal(fileID uint, userID uint, isAdmin bool) error {
	if globalFileManager == nil {
		return fmt.Errorf("文件管理器未初始化")
	}
	return globalFileManager.DeleteFile(fileID, userID, isAdmin)
}
