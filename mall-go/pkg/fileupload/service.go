package fileupload

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"mall-go/internal/model"
	"mall-go/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// FileUploadService 文件上传服务
type FileUploadService struct {
	db         *gorm.DB
	uploadPath string
	baseURL    string
	config     *UploadConfig  // 上传配置
	validator  *FileValidator // 文件验证器
	// 并发安全机制
	uploadMutex sync.RWMutex // 保护文件上传操作
	pathMutex   sync.Mutex   // 保护路径生成操作
}

// NewFileUploadService 创建文件上传服务实例（使用默认配置）
func NewFileUploadService(db *gorm.DB, uploadPath, baseURL string) *FileUploadService {
	config := DefaultUploadConfig()
	config.UploadPath = uploadPath
	config.BaseURL = baseURL

	return NewFileUploadServiceWithConfig(db, config)
}

// NewFileUploadServiceWithConfig 创建文件上传服务实例（使用自定义配置）
func NewFileUploadServiceWithConfig(db *gorm.DB, config *UploadConfig) *FileUploadService {
	// 验证配置
	if err := config.Validate(); err != nil {
		logger.Error("上传配置验证失败", zap.Error(err))
		config = DefaultUploadConfig() // 使用默认配置
	}

	// 创建验证器
	validator := NewFileValidatorWithConfig(config.ToValidationConfig())

	return &FileUploadService{
		db:         db,
		uploadPath: config.UploadPath,
		baseURL:    config.BaseURL,
		config:     config,
		validator:  validator,
	}
}

// UploadFile 上传单个文件（并发安全，支持事务）
func (s *FileUploadService) UploadFile(
	fileHeader *multipart.FileHeader,
	userID uint,
	businessType model.BusinessType,
	businessID *uint,
	isPublic bool,
) (*model.FileUploadResponse, error) {
	// 并发安全：获取上传锁
	s.uploadMutex.Lock()
	defer s.uploadMutex.Unlock()

	// 记录上传开始
	logger.Info("开始文件上传",
		zap.String("filename", fileHeader.Filename),
		zap.Int64("size", fileHeader.Size),
		zap.Uint("user_id", userID),
		zap.String("business_type", string(businessType)))

	// 1. 基础验证
	if err := s.validateFile(fileHeader); err != nil {
		logger.Error("文件验证失败",
			zap.String("filename", fileHeader.Filename),
			zap.Error(err))
		return nil, fmt.Errorf("文件验证失败: %v", err)
	}

	// 2. 打开上传的文件
	file, err := fileHeader.Open()
	if err != nil {
		logger.Error("打开文件失败",
			zap.String("filename", fileHeader.Filename),
			zap.Error(err))
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 3. 创建文件记录
	fileRecord := &model.File{
		UUID:         uuid.New().String(),
		OriginalName: fileHeader.Filename,
		FileSize:     fileHeader.Size,
		MimeType:     fileHeader.Header.Get("Content-Type"),
		FileType:     model.GetFileTypeByMime(fileHeader.Header.Get("Content-Type")),
		StorageType:  model.StorageTypeLocal,
		UploadUserID: userID,
		BusinessType: businessType,
		BusinessID:   businessID,
		IsPublic:     isPublic,
		Status:       model.FileStatusUploading,
	}

	// 4. 生成存储路径和文件名（线程安全）
	storedName, filePath, err := s.generateStoragePathSafe(fileRecord)
	if err != nil {
		logger.Error("生成存储路径失败",
			zap.String("uuid", fileRecord.UUID),
			zap.Error(err))
		return nil, fmt.Errorf("生成存储路径失败: %v", err)
	}

	fileRecord.StoredName = storedName
	fileRecord.FilePath = filePath
	fileRecord.AccessURL = s.generateAccessURL(fileRecord)

	// 5. 使用数据库事务确保原子性
	tx := s.db.Begin()
	if tx.Error != nil {
		logger.Error("开始事务失败", zap.Error(tx.Error))
		return nil, fmt.Errorf("开始事务失败: %v", tx.Error)
	}

	// 确保事务回滚
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			logger.Error("文件上传过程中发生panic", zap.Any("panic", r))
		}
	}()

	// 6. 保存文件记录到数据库
	if err := tx.Create(fileRecord).Error; err != nil {
		tx.Rollback()
		logger.Error("保存文件记录失败",
			zap.String("uuid", fileRecord.UUID),
			zap.Error(err))
		return nil, fmt.Errorf("保存文件记录失败: %v", err)
	}

	// 7. 保存文件到磁盘
	if err := s.saveFileToDisk(file, filePath); err != nil {
		tx.Rollback()
		logger.Error("保存文件到磁盘失败",
			zap.String("uuid", fileRecord.UUID),
			zap.String("path", filePath),
			zap.Error(err))
		return nil, fmt.Errorf("保存文件失败: %v", err)
	}

	// 8. 更新文件状态为成功
	fileRecord.Status = model.FileStatusSuccess
	if err := tx.Model(fileRecord).Update("status", model.FileStatusSuccess).Error; err != nil {
		tx.Rollback()
		// 清理已保存的文件
		os.Remove(filePath)
		logger.Error("更新文件状态失败",
			zap.String("uuid", fileRecord.UUID),
			zap.Error(err))
		return nil, fmt.Errorf("更新文件状态失败: %v", err)
	}

	// 9. 提交事务
	if err := tx.Commit().Error; err != nil {
		// 清理已保存的文件
		os.Remove(filePath)
		logger.Error("提交事务失败",
			zap.String("uuid", fileRecord.UUID),
			zap.Error(err))
		return nil, fmt.Errorf("提交事务失败: %v", err)
	}

	// 10. 返回响应
	response := &model.FileUploadResponse{
		ID:           fileRecord.ID,
		UUID:         fileRecord.UUID,
		OriginalName: fileRecord.OriginalName,
		FileSize:     fileRecord.FileSize,
		MimeType:     fileRecord.MimeType,
		FileType:     fileRecord.FileType,
		BusinessType: fileRecord.BusinessType,
		BusinessID:   fileRecord.BusinessID,
		AccessURL:    fileRecord.AccessURL,
		IsPublic:     fileRecord.IsPublic,
		Status:       fileRecord.Status,
		CreatedAt:    fileRecord.CreatedAt,
	}

	// 记录上传成功
	logger.Info("文件上传成功",
		zap.String("uuid", fileRecord.UUID),
		zap.String("filename", fileRecord.OriginalName),
		zap.Uint("user_id", userID),
		zap.Int64("size", fileRecord.FileSize))

	return response, nil
}

// UploadMultipleFiles 上传多个文件（并发安全，支持批量事务）
func (s *FileUploadService) UploadMultipleFiles(
	fileHeaders []*multipart.FileHeader,
	userID uint,
	businessType model.BusinessType,
	businessID *uint,
	isPublic bool,
) ([]*model.FileUploadResponse, error) {
	// 验证文件数量
	if len(fileHeaders) > model.MaxFileCount {
		logger.Error("文件数量超过限制",
			zap.Int("count", len(fileHeaders)),
			zap.Int("max", model.MaxFileCount))
		return nil, fmt.Errorf("文件数量超过限制，最多允许%d个文件", model.MaxFileCount)
	}

	logger.Info("开始批量文件上传",
		zap.Int("file_count", len(fileHeaders)),
		zap.Uint("user_id", userID),
		zap.String("business_type", string(businessType)))

	var responses []*model.FileUploadResponse
	var errors []string
	var successCount int

	// 顺序上传文件，避免过度并发导致资源竞争
	for i, fileHeader := range fileHeaders {
		response, err := s.UploadFile(fileHeader, userID, businessType, businessID, isPublic)
		if err != nil {
			errors = append(errors, fmt.Sprintf("文件%d(%s)上传失败: %v",
				i+1, fileHeader.Filename, err))
			logger.Error("单个文件上传失败",
				zap.Int("index", i+1),
				zap.String("filename", fileHeader.Filename),
				zap.Error(err))
			continue
		}
		responses = append(responses, response)
		successCount++
	}

	logger.Info("批量文件上传完成",
		zap.Int("total", len(fileHeaders)),
		zap.Int("success", successCount),
		zap.Int("failed", len(errors)))

	// 如果有错误，返回错误信息
	if len(errors) > 0 {
		return responses, fmt.Errorf("部分文件上传失败: %s", strings.Join(errors, "; "))
	}

	return responses, nil
}

// GetFileByUUID 根据UUID获取文件信息
func (s *FileUploadService) GetFileByUUID(uuid string) (*model.File, error) {
	var file model.File
	err := s.db.Where("uuid = ? AND status = ?", uuid, model.FileStatusSuccess).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// GetFilesByUser 获取用户的文件列表
func (s *FileUploadService) GetFilesByUser(
	userID uint,
	businessType model.BusinessType,
	businessID *uint,
	page, pageSize int,
) ([]*model.File, int64, error) {
	var files []*model.File
	var total int64

	query := s.db.Model(&model.File{}).Where("upload_user_id = ? AND status = ?", userID, model.FileStatusSuccess)

	// 添加业务类型过滤
	if businessType != "" {
		query = query.Where("business_type = ?", businessType)
	}

	// 添加业务ID过滤
	if businessID != nil {
		query = query.Where("business_id = ?", *businessID)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&files).Error; err != nil {
		return nil, 0, err
	}

	return files, total, nil
}

// DeleteFile 删除文件（支持事务和结构化日志）
func (s *FileUploadService) DeleteFile(uuid string, userID uint) error {
	logger.Info("开始删除文件",
		zap.String("uuid", uuid),
		zap.Uint("user_id", userID))

	// 查找文件
	var file model.File
	err := s.db.Where("uuid = ? AND upload_user_id = ?", uuid, userID).First(&file).Error
	if err != nil {
		logger.Error("文件不存在或无权限删除",
			zap.String("uuid", uuid),
			zap.Uint("user_id", userID),
			zap.Error(err))
		return fmt.Errorf("文件不存在或无权限删除")
	}

	// 使用事务确保数据一致性
	tx := s.db.Begin()
	if tx.Error != nil {
		logger.Error("开始删除事务失败", zap.Error(tx.Error))
		return fmt.Errorf("开始删除事务失败: %v", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			logger.Error("删除文件过程中发生panic", zap.Any("panic", r))
		}
	}()

	// 更新文件状态为已删除
	if err := tx.Model(&file).Update("status", model.FileStatusDeleted).Error; err != nil {
		tx.Rollback()
		logger.Error("更新文件状态失败",
			zap.String("uuid", uuid),
			zap.Error(err))
		return fmt.Errorf("更新文件状态失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		logger.Error("提交删除事务失败",
			zap.String("uuid", uuid),
			zap.Error(err))
		return fmt.Errorf("提交删除事务失败: %v", err)
	}

	// 删除物理文件（在事务提交后）
	if err := os.Remove(file.FilePath); err != nil {
		logger.Warn("删除物理文件失败",
			zap.String("uuid", uuid),
			zap.String("path", file.FilePath),
			zap.Error(err))
		// 物理文件删除失败不影响逻辑删除的成功
	}

	logger.Info("文件删除成功",
		zap.String("uuid", uuid),
		zap.Uint("user_id", userID),
		zap.String("filename", file.OriginalName))

	return nil
}

// validateFile 验证文件（使用配置化验证器）
func (s *FileUploadService) validateFile(fileHeader *multipart.FileHeader) error {
	// 使用配置化的验证器
	if s.validator != nil {
		return s.validator.ValidateFile(fileHeader)
	}

	// 兜底验证逻辑（如果验证器未初始化）
	mimeType := fileHeader.Header.Get("Content-Type")
	fileType := model.GetFileTypeByMime(mimeType)

	// 使用配置中的大小限制
	var maxSize int64
	if s.config != nil {
		maxSize = s.config.GetMaxSizeByFileType(fileType)
	} else {
		maxSize = model.GetMaxSizeByType(fileType)
	}

	if fileHeader.Size > maxSize {
		return fmt.Errorf("文件大小超过限制，最大允许%d字节", maxSize)
	}

	// 验证文件类型
	if !s.isValidMimeType(mimeType) {
		return fmt.Errorf("不支持的文件类型: %s", mimeType)
	}

	// 验证文件扩展名
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !s.isValidExtension(ext, fileType) {
		return fmt.Errorf("不支持的文件扩展名: %s", ext)
	}

	return nil
}

// isValidMimeType 验证MIME类型
func (s *FileUploadService) isValidMimeType(mimeType string) bool {
	validMimes := append(append(model.ImageMimeTypes, model.DocumentMimeTypes...), model.VideoMimeTypes...)
	for _, validMime := range validMimes {
		if mimeType == validMime {
			return true
		}
	}
	return false
}

// isValidExtension 验证文件扩展名
func (s *FileUploadService) isValidExtension(ext string, fileType model.FileType) bool {
	var validExts []string
	switch fileType {
	case model.FileTypeImage:
		validExts = model.SupportedImageTypes
	case model.FileTypeDocument:
		validExts = model.SupportedDocumentTypes
	case model.FileTypeVideo:
		validExts = model.SupportedVideoTypes
	default:
		validExts = append(append(model.SupportedImageTypes, model.SupportedDocumentTypes...), model.SupportedVideoTypes...)
	}

	for _, validExt := range validExts {
		if ext == validExt {
			return true
		}
	}
	return false
}

// generateStoragePath 生成存储路径（原始方法，保持兼容性）
func (s *FileUploadService) generateStoragePath(file *model.File) (string, string, error) {
	return s.generateStoragePathSafe(file)
}

// generateStoragePathSafe 生成存储路径（线程安全版本）
func (s *FileUploadService) generateStoragePathSafe(file *model.File) (string, string, error) {
	// 使用互斥锁保护路径生成过程
	s.pathMutex.Lock()
	defer s.pathMutex.Unlock()

	// 根据文件类型和业务类型生成目录结构
	now := time.Now()
	typeDir := strings.ToLower(string(file.FileType)) + "s" // images, documents, videos
	businessDir := strings.ToLower(string(file.BusinessType))
	dateDir := now.Format("2006/01")

	// 构建完整目录路径
	fullDir := filepath.Join(s.uploadPath, typeDir, businessDir, dateDir)

	// 创建目录（线程安全）
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		logger.Error("创建目录失败",
			zap.String("dir", fullDir),
			zap.Error(err))
		return "", "", fmt.Errorf("创建目录失败: %v", err)
	}

	// 生成存储文件名（使用UUID + 原始扩展名）
	ext := filepath.Ext(file.OriginalName)
	storedName := file.UUID + ext
	fullPath := filepath.Join(fullDir, storedName)

	// 检查文件是否已存在（防止UUID冲突）
	if _, err := os.Stat(fullPath); err == nil {
		// 文件已存在，重新生成UUID
		file.UUID = uuid.New().String()
		storedName = file.UUID + ext
		fullPath = filepath.Join(fullDir, storedName)

		logger.Warn("检测到文件名冲突，重新生成UUID",
			zap.String("original_uuid", file.UUID),
			zap.String("new_uuid", file.UUID))
	}

	return storedName, fullPath, nil
}

// saveFileToDisk 保存文件到磁盘
func (s *FileUploadService) saveFileToDisk(src io.Reader, destPath string) error {
	// 创建目标文件
	dst, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer dst.Close()

	// 复制文件内容
	_, err = io.Copy(dst, src)
	if err != nil {
		return fmt.Errorf("复制文件内容失败: %v", err)
	}

	return nil
}

// generateAccessURL 生成访问URL
func (s *FileUploadService) generateAccessURL(file *model.File) string {
	if file.IsPublic {
		return fmt.Sprintf("%s/api/v1/files/public/%s", s.baseURL, file.UUID)
	}
	return fmt.Sprintf("%s/api/v1/files/private/%s", s.baseURL, file.UUID)
}
