package fileupload

import (
	"database/sql"
	"fmt"
	"mime/multipart"

	"mall-go/internal/model"
	"mall-go/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UploadManager 文件上传管理器（统一入口）
type UploadManager struct {
	db            *gorm.DB
	config        *UploadConfig
	uploadService *FileUploadService
	chunkService  *ChunkUploadService
	validator     *FileValidator
}

// NewUploadManager 创建文件上传管理器
func NewUploadManager(db *gorm.DB, config *UploadConfig) *UploadManager {
	if config == nil {
		config = DefaultUploadConfig()
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		logger.Error("上传配置验证失败", zap.Error(err))
		config = DefaultUploadConfig()
	}

	return &UploadManager{
		db:            db,
		config:        config,
		uploadService: NewFileUploadServiceWithConfig(db, config),
		chunkService:  NewChunkUploadService(db, config),
		validator:     NewFileValidatorWithConfig(config.ToValidationConfig()),
	}
}

// UploadSingleFile 上传单个文件
func (m *UploadManager) UploadSingleFile(
	fileHeader *multipart.FileHeader,
	userID uint,
	businessType model.BusinessType,
	businessID *uint,
	isPublic bool,
) (*model.FileUploadResponse, error) {
	logger.Info("开始单文件上传",
		zap.String("filename", fileHeader.Filename),
		zap.Int64("size", fileHeader.Size),
		zap.Uint("user_id", userID))

	return m.uploadService.UploadFile(fileHeader, userID, businessType, businessID, isPublic)
}

// UploadMultipleFiles 上传多个文件
func (m *UploadManager) UploadMultipleFiles(
	fileHeaders []*multipart.FileHeader,
	userID uint,
	businessType model.BusinessType,
	businessID *uint,
	isPublic bool,
) ([]*model.FileUploadResponse, error) {
	logger.Info("开始多文件上传",
		zap.Int("file_count", len(fileHeaders)),
		zap.Uint("user_id", userID))

	return m.uploadService.UploadMultipleFiles(fileHeaders, userID, businessType, businessID, isPublic)
}

// InitiateChunkUpload 初始化分片上传
func (m *UploadManager) InitiateChunkUpload(
	fileName string,
	fileSize int64,
	chunkSize int64,
	userID uint,
	businessType model.BusinessType,
	businessID *uint,
	isPublic bool,
) (*ChunkSession, error) {
	logger.Info("初始化分片上传",
		zap.String("filename", fileName),
		zap.Int64("file_size", fileSize),
		zap.Int64("chunk_size", chunkSize),
		zap.Uint("user_id", userID))

	return m.chunkService.InitiateChunkUpload(fileName, fileSize, chunkSize, userID, businessType, businessID, isPublic)
}

// UploadChunk 上传分片
func (m *UploadManager) UploadChunk(
	sessionID string,
	chunkIndex int,
	chunkData *multipart.FileHeader,
	chunkHash string,
) (*ChunkUploadResponse, error) {
	return m.chunkService.UploadChunk(sessionID, chunkIndex, chunkData, chunkHash)
}

// DeleteFile 删除文件
func (m *UploadManager) DeleteFile(uuid string, userID uint) error {
	logger.Info("删除文件",
		zap.String("uuid", uuid),
		zap.Uint("user_id", userID))

	return m.uploadService.DeleteFile(uuid, userID)
}

// ValidateFile 验证文件
func (m *UploadManager) ValidateFile(fileHeader *multipart.FileHeader) error {
	return m.validator.ValidateFile(fileHeader)
}

// GetUploadConfig 获取上传配置
func (m *UploadManager) GetUploadConfig() *UploadConfig {
	return m.config
}

// UpdateConfig 更新配置
func (m *UploadManager) UpdateConfig(config *UploadConfig) error {
	if err := config.Validate(); err != nil {
		return fmt.Errorf("配置验证失败: %v", err)
	}

	m.config = config

	// 重新创建服务实例
	m.uploadService = NewFileUploadServiceWithConfig(m.db, config)
	m.chunkService = NewChunkUploadService(m.db, config)
	m.validator = NewFileValidatorWithConfig(config.ToValidationConfig())

	logger.Info("上传配置已更新")
	return nil
}

// GetFileInfo 获取文件信息
func (m *UploadManager) GetFileInfo(uuid string, userID uint) (*model.File, error) {
	var file model.File
	err := m.db.Where("uuid = ? AND upload_user_id = ?", uuid, userID).First(&file).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("文件不存在或无权限访问")
		}
		return nil, fmt.Errorf("查询文件信息失败: %v", err)
	}

	return &file, nil
}

// ListUserFiles 列出用户文件
func (m *UploadManager) ListUserFiles(
	userID uint,
	businessType *model.BusinessType,
	businessID *uint,
	page, pageSize int,
) ([]*model.File, int64, error) {
	query := m.db.Where("upload_user_id = ? AND status != ?", userID, model.FileStatusDeleted)

	if businessType != nil {
		query = query.Where("business_type = ?", *businessType)
	}

	if businessID != nil {
		query = query.Where("business_id = ?", *businessID)
	}

	// 计算总数
	var total int64
	if err := query.Model(&model.File{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计文件数量失败: %v", err)
	}

	// 分页查询
	var files []*model.File
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&files).Error
	if err != nil {
		return nil, 0, fmt.Errorf("查询文件列表失败: %v", err)
	}

	return files, total, nil
}

// GetUploadStatistics 获取上传统计信息
func (m *UploadManager) GetUploadStatistics(userID uint) (*UploadStatistics, error) {
	stats := &UploadStatistics{
		UserID: userID,
	}

	// 统计总文件数
	err := m.db.Model(&model.File{}).
		Where("upload_user_id = ? AND status = ?", userID, model.FileStatusSuccess).
		Count(&stats.TotalFiles).Error
	if err != nil {
		return nil, fmt.Errorf("统计文件数量失败: %v", err)
	}

	// 统计总大小
	var totalSize sql.NullInt64
	err = m.db.Model(&model.File{}).
		Where("upload_user_id = ? AND status = ?", userID, model.FileStatusSuccess).
		Select("SUM(file_size)").Scan(&totalSize).Error
	if err != nil {
		return nil, fmt.Errorf("统计文件大小失败: %v", err)
	}

	if totalSize.Valid {
		stats.TotalSize = totalSize.Int64
	}

	// 按文件类型统计
	var typeStats []struct {
		FileType string `json:"file_type"`
		Count    int64  `json:"count"`
		Size     int64  `json:"size"`
	}

	err = m.db.Model(&model.File{}).
		Where("upload_user_id = ? AND status = ?", userID, model.FileStatusSuccess).
		Select("file_type, COUNT(*) as count, SUM(file_size) as size").
		Group("file_type").
		Scan(&typeStats).Error
	if err != nil {
		return nil, fmt.Errorf("统计文件类型失败: %v", err)
	}

	stats.FileTypeStats = make(map[string]FileTypeStat)
	for _, stat := range typeStats {
		stats.FileTypeStats[stat.FileType] = FileTypeStat{
			Count: stat.Count,
			Size:  stat.Size,
		}
	}

	return stats, nil
}

// Stop 停止管理器
func (m *UploadManager) Stop() {
	if m.chunkService != nil {
		m.chunkService.Stop()
	}
	logger.Info("文件上传管理器已停止")
}

// UploadStatistics 上传统计信息
type UploadStatistics struct {
	UserID        uint                    `json:"user_id"`
	TotalFiles    int64                   `json:"total_files"`
	TotalSize     int64                   `json:"total_size"`
	FileTypeStats map[string]FileTypeStat `json:"file_type_stats"`
}

// FileTypeStat 文件类型统计
type FileTypeStat struct {
	Count int64 `json:"count"`
	Size  int64 `json:"size"`
}
