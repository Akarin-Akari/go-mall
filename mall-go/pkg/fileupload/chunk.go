package fileupload

import (
	"crypto/md5"
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

// ChunkUploadService 分片上传服务
type ChunkUploadService struct {
	db         *gorm.DB
	uploadPath string
	baseURL    string
	config     *UploadConfig
	// 分片管理
	chunkSessions sync.Map // 存储分片上传会话
	cleanupTicker *time.Ticker
}

// ChunkSession 分片上传会话
type ChunkSession struct {
	SessionID      string             `json:"session_id"`
	FileUUID       string             `json:"file_uuid"`
	FileName       string             `json:"file_name"`
	FileSize       int64              `json:"file_size"`
	ChunkSize      int64              `json:"chunk_size"`
	TotalChunks    int                `json:"total_chunks"`
	UploadedChunks map[int]bool       `json:"uploaded_chunks"`
	ChunkHashes    map[int]string     `json:"chunk_hashes"`
	UserID         uint               `json:"user_id"`
	BusinessType   model.BusinessType `json:"business_type"`
	BusinessID     *uint              `json:"business_id"`
	IsPublic       bool               `json:"is_public"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
	mutex          sync.RWMutex       `json:"-"`
}

// ChunkUploadRequest 分片上传请求
type ChunkUploadRequest struct {
	SessionID   string `json:"session_id" binding:"required"`
	ChunkIndex  int    `json:"chunk_index" binding:"required"`
	ChunkHash   string `json:"chunk_hash" binding:"required"`
	TotalChunks int    `json:"total_chunks" binding:"required"`
}

// ChunkUploadResponse 分片上传响应
type ChunkUploadResponse struct {
	SessionID      string  `json:"session_id"`
	ChunkIndex     int     `json:"chunk_index"`
	Uploaded       bool    `json:"uploaded"`
	Progress       float64 `json:"progress"`
	UploadedChunks int     `json:"uploaded_chunks"`
	TotalChunks    int     `json:"total_chunks"`
	IsComplete     bool    `json:"is_complete"`
}

// NewChunkUploadService 创建分片上传服务
func NewChunkUploadService(db *gorm.DB, config *UploadConfig) *ChunkUploadService {
	service := &ChunkUploadService{
		db:         db,
		uploadPath: config.UploadPath,
		baseURL:    config.BaseURL,
		config:     config,
	}

	// 启动清理定时器
	service.startCleanupTimer()

	return service
}

// InitiateChunkUpload 初始化分片上传
func (s *ChunkUploadService) InitiateChunkUpload(
	fileName string,
	fileSize int64,
	chunkSize int64,
	userID uint,
	businessType model.BusinessType,
	businessID *uint,
	isPublic bool,
) (*ChunkSession, error) {
	// 验证参数
	if fileName == "" {
		return nil, fmt.Errorf("文件名不能为空")
	}

	if fileSize <= 0 {
		return nil, fmt.Errorf("文件大小必须大于0")
	}

	if chunkSize <= 0 {
		chunkSize = 1024 * 1024 // 默认1MB
	}

	// 计算分片数量
	totalChunks := int((fileSize + chunkSize - 1) / chunkSize)

	// 创建会话
	session := &ChunkSession{
		SessionID:      uuid.New().String(),
		FileUUID:       uuid.New().String(),
		FileName:       fileName,
		FileSize:       fileSize,
		ChunkSize:      chunkSize,
		TotalChunks:    totalChunks,
		UploadedChunks: make(map[int]bool),
		ChunkHashes:    make(map[int]string),
		UserID:         userID,
		BusinessType:   businessType,
		BusinessID:     businessID,
		IsPublic:       isPublic,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 存储会话
	s.chunkSessions.Store(session.SessionID, session)

	logger.Info("初始化分片上传会话",
		zap.String("session_id", session.SessionID),
		zap.String("file_name", fileName),
		zap.Int64("file_size", fileSize),
		zap.Int("total_chunks", totalChunks),
		zap.Uint("user_id", userID))

	return session, nil
}

// UploadChunk 上传分片
func (s *ChunkUploadService) UploadChunk(
	sessionID string,
	chunkIndex int,
	chunkData *multipart.FileHeader,
	chunkHash string,
) (*ChunkUploadResponse, error) {
	// 获取会话
	sessionInterface, exists := s.chunkSessions.Load(sessionID)
	if !exists {
		return nil, fmt.Errorf("分片上传会话不存在")
	}

	session := sessionInterface.(*ChunkSession)
	session.mutex.Lock()
	defer session.mutex.Unlock()

	// 验证分片索引
	if chunkIndex < 0 || chunkIndex >= session.TotalChunks {
		return nil, fmt.Errorf("无效的分片索引: %d", chunkIndex)
	}

	// 检查分片是否已上传
	if session.UploadedChunks[chunkIndex] {
		logger.Info("分片已存在，跳过上传",
			zap.String("session_id", sessionID),
			zap.Int("chunk_index", chunkIndex))

		return s.buildChunkResponse(session, chunkIndex, true), nil
	}

	// 验证分片哈希
	if err := s.validateChunkHash(chunkData, chunkHash); err != nil {
		return nil, fmt.Errorf("分片哈希验证失败: %v", err)
	}

	// 保存分片文件
	chunkPath := s.getChunkPath(session.SessionID, chunkIndex)
	if err := s.saveChunkToDisk(chunkData, chunkPath); err != nil {
		return nil, fmt.Errorf("保存分片失败: %v", err)
	}

	// 更新会话状态
	session.UploadedChunks[chunkIndex] = true
	session.ChunkHashes[chunkIndex] = chunkHash
	session.UpdatedAt = time.Now()

	logger.Info("分片上传成功",
		zap.String("session_id", sessionID),
		zap.Int("chunk_index", chunkIndex),
		zap.String("chunk_path", chunkPath))

	response := s.buildChunkResponse(session, chunkIndex, true)

	// 检查是否所有分片都已上传
	if response.IsComplete {
		// 异步合并文件
		go s.mergeChunksAsync(session)
	}

	return response, nil
}

// validateChunkHash 验证分片哈希
func (s *ChunkUploadService) validateChunkHash(chunkData *multipart.FileHeader, expectedHash string) error {
	file, err := chunkData.Open()
	if err != nil {
		return fmt.Errorf("打开分片文件失败: %v", err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("计算分片哈希失败: %v", err)
	}

	actualHash := fmt.Sprintf("%x", hash.Sum(nil))
	if actualHash != expectedHash {
		return fmt.Errorf("分片哈希不匹配，期望: %s, 实际: %s", expectedHash, actualHash)
	}

	return nil
}

// getChunkPath 获取分片文件路径
func (s *ChunkUploadService) getChunkPath(sessionID string, chunkIndex int) string {
	chunkDir := filepath.Join(s.uploadPath, "chunks", sessionID)
	os.MkdirAll(chunkDir, 0755)
	return filepath.Join(chunkDir, fmt.Sprintf("chunk_%d", chunkIndex))
}

// saveChunkToDisk 保存分片到磁盘
func (s *ChunkUploadService) saveChunkToDisk(chunkData *multipart.FileHeader, chunkPath string) error {
	src, err := chunkData.Open()
	if err != nil {
		return fmt.Errorf("打开分片文件失败: %v", err)
	}
	defer src.Close()

	dst, err := os.Create(chunkPath)
	if err != nil {
		return fmt.Errorf("创建分片文件失败: %v", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("复制分片数据失败: %v", err)
	}

	return nil
}

// buildChunkResponse 构建分片响应
func (s *ChunkUploadService) buildChunkResponse(session *ChunkSession, chunkIndex int, uploaded bool) *ChunkUploadResponse {
	uploadedCount := len(session.UploadedChunks)
	progress := float64(uploadedCount) / float64(session.TotalChunks) * 100
	isComplete := uploadedCount == session.TotalChunks

	return &ChunkUploadResponse{
		SessionID:      session.SessionID,
		ChunkIndex:     chunkIndex,
		Uploaded:       uploaded,
		Progress:       progress,
		UploadedChunks: uploadedCount,
		TotalChunks:    session.TotalChunks,
		IsComplete:     isComplete,
	}
}

// startCleanupTimer 启动清理定时器
func (s *ChunkUploadService) startCleanupTimer() {
	s.cleanupTicker = time.NewTicker(1 * time.Hour) // 每小时清理一次

	go func() {
		for range s.cleanupTicker.C {
			s.cleanupExpiredSessions()
		}
	}()
}

// cleanupExpiredSessions 清理过期会话
func (s *ChunkUploadService) cleanupExpiredSessions() {
	expireTime := time.Now().Add(-24 * time.Hour) // 24小时过期

	s.chunkSessions.Range(func(key, value interface{}) bool {
		session := value.(*ChunkSession)
		if session.UpdatedAt.Before(expireTime) {
			// 删除分片文件
			chunkDir := filepath.Join(s.uploadPath, "chunks", session.SessionID)
			os.RemoveAll(chunkDir)

			// 删除会话
			s.chunkSessions.Delete(key)

			logger.Info("清理过期分片上传会话",
				zap.String("session_id", session.SessionID),
				zap.String("file_name", session.FileName))
		}
		return true
	})
}

// mergeChunksAsync 异步合并分片
func (s *ChunkUploadService) mergeChunksAsync(session *ChunkSession) {
	logger.Info("开始合并分片",
		zap.String("session_id", session.SessionID),
		zap.String("file_name", session.FileName))

	// 创建最终文件记录
	fileRecord := &model.File{
		UUID:         session.FileUUID,
		OriginalName: session.FileName,
		FileSize:     session.FileSize,
		MimeType:     "application/octet-stream", // 临时设置，后续可以根据文件内容检测
		FileType:     model.GetFileTypeByMime("application/octet-stream"),
		StorageType:  model.StorageTypeLocal,
		UploadUserID: session.UserID,
		BusinessType: session.BusinessType,
		BusinessID:   session.BusinessID,
		IsPublic:     session.IsPublic,
		Status:       model.FileStatusUploading,
	}

	// 生成最终文件路径
	storedName, finalPath, err := s.generateFinalPath(fileRecord)
	if err != nil {
		logger.Error("生成最终文件路径失败",
			zap.String("session_id", session.SessionID),
			zap.Error(err))
		return
	}

	fileRecord.StoredName = storedName
	fileRecord.FilePath = finalPath
	fileRecord.AccessURL = s.generateAccessURL(fileRecord)

	// 使用事务
	tx := s.db.Begin()
	if tx.Error != nil {
		logger.Error("开始合并事务失败",
			zap.String("session_id", session.SessionID),
			zap.Error(tx.Error))
		return
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			logger.Error("合并分片过程中发生panic",
				zap.String("session_id", session.SessionID),
				zap.Any("panic", r))
		}
	}()

	// 保存文件记录
	if err := tx.Create(fileRecord).Error; err != nil {
		tx.Rollback()
		logger.Error("保存文件记录失败",
			zap.String("session_id", session.SessionID),
			zap.Error(err))
		return
	}

	// 合并分片文件
	if err := s.mergeChunkFiles(session, finalPath); err != nil {
		tx.Rollback()
		logger.Error("合并分片文件失败",
			zap.String("session_id", session.SessionID),
			zap.Error(err))
		return
	}

	// 更新文件状态
	fileRecord.Status = model.FileStatusSuccess
	if err := tx.Model(fileRecord).Update("status", model.FileStatusSuccess).Error; err != nil {
		tx.Rollback()
		os.Remove(finalPath) // 清理文件
		logger.Error("更新文件状态失败",
			zap.String("session_id", session.SessionID),
			zap.Error(err))
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		os.Remove(finalPath) // 清理文件
		logger.Error("提交合并事务失败",
			zap.String("session_id", session.SessionID),
			zap.Error(err))
		return
	}

	// 清理分片文件
	s.cleanupChunkFiles(session.SessionID)

	// 删除会话
	s.chunkSessions.Delete(session.SessionID)

	logger.Info("分片合并完成",
		zap.String("session_id", session.SessionID),
		zap.String("file_uuid", fileRecord.UUID),
		zap.String("final_path", finalPath))
}

// mergeChunkFiles 合并分片文件
func (s *ChunkUploadService) mergeChunkFiles(session *ChunkSession, finalPath string) error {
	// 创建最终文件
	finalFile, err := os.Create(finalPath)
	if err != nil {
		return fmt.Errorf("创建最终文件失败: %v", err)
	}
	defer finalFile.Close()

	// 按顺序合并分片
	for i := 0; i < session.TotalChunks; i++ {
		chunkPath := s.getChunkPath(session.SessionID, i)

		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			return fmt.Errorf("打开分片文件失败 (chunk %d): %v", i, err)
		}

		if _, err := io.Copy(finalFile, chunkFile); err != nil {
			chunkFile.Close()
			return fmt.Errorf("复制分片数据失败 (chunk %d): %v", i, err)
		}

		chunkFile.Close()
	}

	return nil
}

// generateFinalPath 生成最终文件路径
func (s *ChunkUploadService) generateFinalPath(file *model.File) (string, string, error) {
	// 根据文件类型和业务类型生成目录结构
	now := time.Now()
	typeDir := strings.ToLower(string(file.FileType)) + "s"
	businessDir := strings.ToLower(string(file.BusinessType))
	dateDir := now.Format("2006/01")

	// 构建完整目录路径
	fullDir := filepath.Join(s.uploadPath, typeDir, businessDir, dateDir)

	// 创建目录
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		return "", "", fmt.Errorf("创建目录失败: %v", err)
	}

	// 生成存储文件名
	ext := filepath.Ext(file.OriginalName)
	storedName := file.UUID + ext
	fullPath := filepath.Join(fullDir, storedName)

	return storedName, fullPath, nil
}

// generateAccessURL 生成访问URL
func (s *ChunkUploadService) generateAccessURL(file *model.File) string {
	return fmt.Sprintf("%s/api/files/%s", s.baseURL, file.UUID)
}

// cleanupChunkFiles 清理分片文件
func (s *ChunkUploadService) cleanupChunkFiles(sessionID string) {
	chunkDir := filepath.Join(s.uploadPath, "chunks", sessionID)
	if err := os.RemoveAll(chunkDir); err != nil {
		logger.Warn("清理分片文件失败",
			zap.String("session_id", sessionID),
			zap.String("chunk_dir", chunkDir),
			zap.Error(err))
	}
}

// Stop 停止服务
func (s *ChunkUploadService) Stop() {
	if s.cleanupTicker != nil {
		s.cleanupTicker.Stop()
	}
}
