package upload

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"mall-go/internal/model"
	"mall-go/pkg/response"
	"mall-go/pkg/upload"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UploadHandler 文件上传处理器
type UploadHandler struct {
	db             *gorm.DB
	storageManager *upload.StorageManager
	validator      *upload.FileValidator
	configManager  *upload.ConfigManager
}

// NewUploadHandler 创建文件上传处理器
func NewUploadHandler(db *gorm.DB, configManager *upload.ConfigManager) (*UploadHandler, error) {
	config := configManager.GetConfig()

	// 创建存储管理器
	storageManager, err := upload.NewStorageManager(config)
	if err != nil {
		return nil, fmt.Errorf("创建存储管理器失败: %v", err)
	}

	// 创建文件验证器
	validator := upload.NewFileValidator(config)

	return &UploadHandler{
		db:             db,
		storageManager: storageManager,
		validator:      validator,
		configManager:  configManager,
	}, nil
}

// UploadSingle 单文件上传
func (h *UploadHandler) UploadSingle(c *gin.Context) {
	// 获取用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "获取上传文件失败: "+err.Error())
		return
	}
	defer file.Close()

	// 获取可选参数
	category := c.DefaultPostForm("category", "general") // 文件分类
	description := c.PostForm("description")             // 文件描述

	// 验证文件
	validationResult, err := h.validator.ValidateFile(header.Filename, file, header.Size)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "文件验证失败: "+err.Error())
		return
	}

	if !validationResult.Valid {
		response.Error(c, http.StatusBadRequest, "文件验证失败: "+strings.Join(validationResult.Errors, "; "))
		return
	}

	// 重置文件指针
	file.Seek(0, io.SeekStart)

	// 生成文件路径
	fileName := h.configManager.GenerateFileName(header.Filename)
	filePath := h.generateFilePath(category, fileName)

	// 上传文件
	if err := h.storageManager.Upload(filePath, file, header.Size); err != nil {
		response.Error(c, http.StatusInternalServerError, "文件上传失败: "+err.Error())
		return
	}

	// 保存文件记录到数据库
	fileRecord := &model.File{
		UserID:       userID.(uint),
		OriginalName: header.Filename,
		FileName:     fileName,
		FilePath:     filePath,
		FileSize:     header.Size,
		ContentType:  validationResult.ContentType,
		Category:     category,
		Description:  description,
		URL:          h.storageManager.GetURL(filePath),
		Status:       "active",
	}

	if err := h.db.Create(fileRecord).Error; err != nil {
		// 如果数据库保存失败，删除已上传的文件
		h.storageManager.Delete(filePath)
		response.Error(c, http.StatusInternalServerError, "保存文件记录失败: "+err.Error())
		return
	}

	// 返回成功响应
	response.Success(c, "文件上传成功", gin.H{
		"file_id":       fileRecord.ID,
		"original_name": fileRecord.OriginalName,
		"file_name":     fileRecord.FileName,
		"file_path":     fileRecord.FilePath,
		"file_size":     fileRecord.FileSize,
		"content_type":  fileRecord.ContentType,
		"category":      fileRecord.Category,
		"url":           fileRecord.URL,
		"validation":    validationResult,
		"uploaded_at":   fileRecord.CreatedAt,
	})
}

// UploadMultiple 多文件上传
func (h *UploadHandler) UploadMultiple(c *gin.Context) {
	// 获取用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	// 解析多文件表单
	form, err := c.MultipartForm()
	if err != nil {
		response.Error(c, http.StatusBadRequest, "解析多文件表单失败: "+err.Error())
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		response.Error(c, http.StatusBadRequest, "未找到上传文件")
		return
	}

	// 验证文件数量
	if err := h.configManager.ValidateFileCount(len(files)); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 获取可选参数
	category := c.DefaultPostForm("category", "general")
	description := c.PostForm("description")

	var uploadedFiles []gin.H
	var failedFiles []gin.H

	// 逐个处理文件
	for i, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			failedFiles = append(failedFiles, gin.H{
				"index":    i,
				"filename": fileHeader.Filename,
				"error":    "打开文件失败: " + err.Error(),
			})
			continue
		}

		// 验证文件
		validationResult, err := h.validator.ValidateFile(fileHeader.Filename, file, fileHeader.Size)
		if err != nil {
			file.Close()
			failedFiles = append(failedFiles, gin.H{
				"index":    i,
				"filename": fileHeader.Filename,
				"error":    "文件验证失败: " + err.Error(),
			})
			continue
		}

		if !validationResult.Valid {
			file.Close()
			failedFiles = append(failedFiles, gin.H{
				"index":    i,
				"filename": fileHeader.Filename,
				"error":    "文件验证失败: " + strings.Join(validationResult.Errors, "; "),
			})
			continue
		}

		// 重置文件指针
		file.Seek(0, io.SeekStart)

		// 生成文件路径
		fileName := h.configManager.GenerateFileName(fileHeader.Filename)
		filePath := h.generateFilePath(category, fileName)

		// 上传文件
		if err := h.storageManager.Upload(filePath, file, fileHeader.Size); err != nil {
			file.Close()
			failedFiles = append(failedFiles, gin.H{
				"index":    i,
				"filename": fileHeader.Filename,
				"error":    "文件上传失败: " + err.Error(),
			})
			continue
		}

		file.Close()

		// 保存文件记录到数据库
		fileRecord := &model.File{
			UserID:       userID.(uint),
			OriginalName: fileHeader.Filename,
			FileName:     fileName,
			FilePath:     filePath,
			FileSize:     fileHeader.Size,
			ContentType:  validationResult.ContentType,
			Category:     category,
			Description:  description,
			URL:          h.storageManager.GetURL(filePath),
			Status:       "active",
		}

		if err := h.db.Create(fileRecord).Error; err != nil {
			// 如果数据库保存失败，删除已上传的文件
			h.storageManager.Delete(filePath)
			failedFiles = append(failedFiles, gin.H{
				"index":    i,
				"filename": fileHeader.Filename,
				"error":    "保存文件记录失败: " + err.Error(),
			})
			continue
		}

		// 添加到成功列表
		uploadedFiles = append(uploadedFiles, gin.H{
			"index":         i,
			"file_id":       fileRecord.ID,
			"original_name": fileRecord.OriginalName,
			"file_name":     fileRecord.FileName,
			"file_path":     fileRecord.FilePath,
			"file_size":     fileRecord.FileSize,
			"content_type":  fileRecord.ContentType,
			"category":      fileRecord.Category,
			"url":           fileRecord.URL,
			"validation":    validationResult,
			"uploaded_at":   fileRecord.CreatedAt,
		})
	}

	// 返回结果
	response.Success(c, "文件上传完成", gin.H{
		"total_files":    len(files),
		"success_count":  len(uploadedFiles),
		"failed_count":   len(failedFiles),
		"uploaded_files": uploadedFiles,
		"failed_files":   failedFiles,
	})
}

// GetFileInfo 获取文件信息
func (h *UploadHandler) GetFileInfo(c *gin.Context) {
	fileIDStr := c.Param("id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的文件ID")
		return
	}

	var file model.File
	if err := h.db.First(&file, uint(fileID)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, "文件不存在")
			return
		}
		response.Error(c, http.StatusInternalServerError, "查询文件失败: "+err.Error())
		return
	}

	response.Success(c, "获取文件信息成功", gin.H{
		"file_id":       file.ID,
		"original_name": file.OriginalName,
		"file_name":     file.FileName,
		"file_path":     file.FilePath,
		"file_size":     file.FileSize,
		"content_type":  file.ContentType,
		"category":      file.Category,
		"description":   file.Description,
		"url":           file.URL,
		"status":        file.Status,
		"created_at":    file.CreatedAt,
		"updated_at":    file.UpdatedAt,
	})
}

// ListFiles 获取文件列表
func (h *UploadHandler) ListFiles(c *gin.Context) {
	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	category := c.Query("category")
	status := c.DefaultQuery("status", "active")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 构建查询
	query := h.db.Model(&model.File{}).Where("status = ?", status)

	if category != "" {
		query = query.Where("category = ?", category)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "查询文件总数失败: "+err.Error())
		return
	}

	// 获取文件列表
	var files []model.File
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&files).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "查询文件列表失败: "+err.Error())
		return
	}

	response.Success(c, "获取文件列表成功", gin.H{
		"files":      files,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// DeleteFile 删除文件
func (h *UploadHandler) DeleteFile(c *gin.Context) {
	fileIDStr := c.Param("id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的文件ID")
		return
	}

	// 获取用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "用户未认证")
		return
	}

	var file model.File
	if err := h.db.First(&file, uint(fileID)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, "文件不存在")
			return
		}
		response.Error(c, http.StatusInternalServerError, "查询文件失败: "+err.Error())
		return
	}

	// 检查权限（只能删除自己的文件，除非是管理员）
	userRole, _ := c.Get("user_role")
	if file.UserID != userID.(uint) && userRole != "admin" {
		response.Error(c, http.StatusForbidden, "无权限删除此文件")
		return
	}

	// 从存储中删除文件
	if err := h.storageManager.Delete(file.FilePath); err != nil {
		response.Error(c, http.StatusInternalServerError, "删除存储文件失败: "+err.Error())
		return
	}

	// 从数据库中删除记录
	if err := h.db.Delete(&file).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "删除文件记录失败: "+err.Error())
		return
	}

	response.Success(c, "文件删除成功", nil)
}

// GetUploadConfig 获取上传配置
func (h *UploadHandler) GetUploadConfig(c *gin.Context) {
	config := h.configManager.GetConfigSummary()
	response.Success(c, "获取上传配置成功", config)
}

// generateFilePath 生成文件路径
func (h *UploadHandler) generateFilePath(category, fileName string) string {
	now := time.Now()
	datePath := fmt.Sprintf("%d/%02d/%02d", now.Year(), now.Month(), now.Day())

	if category != "" && category != "general" {
		return filepath.Join(category, datePath, fileName)
	}

	return filepath.Join(datePath, fileName)
}
