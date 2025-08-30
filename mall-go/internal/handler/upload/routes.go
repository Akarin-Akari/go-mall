package upload

import (
	"net/http"
	"strconv"

	"mall-go/internal/handler/middleware"
	"mall-go/internal/model"
	"mall-go/pkg/response"
	"mall-go/pkg/upload"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes 注册文件上传路由
func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB, configManager *upload.ConfigManager) error {
	// 创建上传处理器
	handler, err := NewUploadHandler(db, configManager)
	if err != nil {
		return err
	}

	// 文件上传路由组
	uploadGroup := router.Group("/upload")
	{
		// 需要认证的路由
		uploadGroup.Use(middleware.AuthMiddleware())
		{
			// 单文件上传
			uploadGroup.POST("/single", handler.UploadSingle)

			// 多文件上传
			uploadGroup.POST("/multiple", handler.UploadMultiple)

			// 获取文件信息
			uploadGroup.GET("/file/:id", handler.GetFileInfo)

			// 获取文件列表
			uploadGroup.GET("/files", handler.ListFiles)

			// 删除文件
			uploadGroup.DELETE("/file/:id", handler.DeleteFile)

			// 获取上传配置
			uploadGroup.GET("/config", handler.GetUploadConfig)
		}

		// 管理员路由
		adminGroup := uploadGroup.Group("/admin")
		adminGroup.Use(middleware.AdminMiddleware())
		{
			// 管理员获取所有文件
			adminGroup.GET("/files", handler.ListAllFiles)

			// 管理员删除任意文件
			adminGroup.DELETE("/file/:id", handler.AdminDeleteFile)

			// 获取上传统计
			adminGroup.GET("/statistics", handler.GetUploadStatistics)

			// 清理无效文件
			adminGroup.POST("/cleanup", handler.CleanupInvalidFiles)
		}
	}

	return nil
}

// ListAllFiles 管理员获取所有文件
func (h *UploadHandler) ListAllFiles(c *gin.Context) {
	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	category := c.Query("category")
	status := c.Query("status")
	userID := c.Query("user_id")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 构建查询
	query := h.db.Model(&model.File{}).Preload("User")

	if category != "" {
		query = query.Where("category = ?", category)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if userID != "" {
		query = query.Where("user_id = ?", userID)
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

// AdminDeleteFile 管理员删除文件
func (h *UploadHandler) AdminDeleteFile(c *gin.Context) {
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

// GetUploadStatistics 获取上传统计
func (h *UploadHandler) GetUploadStatistics(c *gin.Context) {
	var stats struct {
		TotalFiles   int64 `json:"total_files"`
		TotalSize    int64 `json:"total_size"`
		ActiveFiles  int64 `json:"active_files"`
		DeletedFiles int64 `json:"deleted_files"`
		TodayUploads int64 `json:"today_uploads"`
		MonthUploads int64 `json:"month_uploads"`
	}

	// 总文件数
	h.db.Model(&model.File{}).Count(&stats.TotalFiles)

	// 总文件大小
	h.db.Model(&model.File{}).Select("COALESCE(SUM(file_size), 0)").Scan(&stats.TotalSize)

	// 活跃文件数
	h.db.Model(&model.File{}).Where("status = ?", "active").Count(&stats.ActiveFiles)

	// 已删除文件数
	h.db.Model(&model.File{}).Where("status = ?", "deleted").Count(&stats.DeletedFiles)

	// 今日上传数
	h.db.Model(&model.File{}).Where("DATE(created_at) = CURDATE()").Count(&stats.TodayUploads)

	// 本月上传数
	h.db.Model(&model.File{}).Where("YEAR(created_at) = YEAR(NOW()) AND MONTH(created_at) = MONTH(NOW())").Count(&stats.MonthUploads)

	// 按分类统计
	var categoryStats []struct {
		Category string `json:"category"`
		Count    int64  `json:"count"`
		Size     int64  `json:"size"`
	}
	h.db.Model(&model.File{}).
		Select("category, COUNT(*) as count, COALESCE(SUM(file_size), 0) as size").
		Where("status = ?", "active").
		Group("category").
		Scan(&categoryStats)

	// 按用户统计（前10名）
	var userStats []struct {
		UserID   uint   `json:"user_id"`
		Username string `json:"username"`
		Count    int64  `json:"count"`
		Size     int64  `json:"size"`
	}
	h.db.Table("files").
		Select("files.user_id, users.username, COUNT(*) as count, COALESCE(SUM(files.file_size), 0) as size").
		Joins("LEFT JOIN users ON files.user_id = users.id").
		Where("files.status = ?", "active").
		Group("files.user_id, users.username").
		Order("count DESC").
		Limit(10).
		Scan(&userStats)

	response.Success(c, "获取上传统计成功", gin.H{
		"overview":       stats,
		"category_stats": categoryStats,
		"user_stats":     userStats,
	})
}

// CleanupInvalidFiles 清理无效文件
func (h *UploadHandler) CleanupInvalidFiles(c *gin.Context) {
	var cleanupResult struct {
		CheckedFiles int   `json:"checked_files"`
		InvalidFiles int   `json:"invalid_files"`
		CleanedFiles int   `json:"cleaned_files"`
		CleanedSize  int64 `json:"cleaned_size"`
	}

	// 获取所有活跃文件
	var files []model.File
	if err := h.db.Where("status = ?", "active").Find(&files).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "查询文件失败: "+err.Error())
		return
	}

	cleanupResult.CheckedFiles = len(files)

	for _, file := range files {
		// 检查文件是否存在
		exists, err := h.storageManager.Exists(file.FilePath)
		if err != nil {
			continue
		}

		if !exists {
			cleanupResult.InvalidFiles++

			// 标记文件为已删除
			if err := h.db.Model(&file).Update("status", "deleted").Error; err == nil {
				cleanupResult.CleanedFiles++
				cleanupResult.CleanedSize += file.FileSize
			}
		}
	}

	response.Success(c, "清理完成", cleanupResult)
}

// RegisterStaticRoutes 注册静态文件路由
func RegisterStaticRoutes(router *gin.Engine, configManager *upload.ConfigManager) {
	config := configManager.GetConfig()

	if config.StorageType == upload.StorageTypeLocal {
		// 提供静态文件访问
		router.Static(config.Local.URLPrefix, config.Local.UploadPath)
	}
}
