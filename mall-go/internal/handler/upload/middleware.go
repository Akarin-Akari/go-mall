package upload

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"mall-go/pkg/response"
	"mall-go/pkg/upload"

	"github.com/gin-gonic/gin"
)

// UploadMiddleware 文件上传中间件
type UploadMiddleware struct {
	configManager *upload.ConfigManager
	rateLimiter   map[string]*RateLimiter
}

// RateLimiter 简单的速率限制器
type RateLimiter struct {
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

// NewUploadMiddleware 创建文件上传中间件
func NewUploadMiddleware(configManager *upload.ConfigManager) *UploadMiddleware {
	return &UploadMiddleware{
		configManager: configManager,
		rateLimiter:   make(map[string]*RateLimiter),
	}
}

// NewRateLimiter 创建速率限制器
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(key string) bool {
	now := time.Now()
	
	// 清理过期的请求记录
	if requests, exists := rl.requests[key]; exists {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if now.Sub(reqTime) < rl.window {
				validRequests = append(validRequests, reqTime)
			}
		}
		rl.requests[key] = validRequests
	}

	// 检查是否超过限制
	if len(rl.requests[key]) >= rl.limit {
		return false
	}

	// 记录当前请求
	rl.requests[key] = append(rl.requests[key], now)
	return true
}

// ValidateUploadRequest 验证上传请求中间件
func (um *UploadMiddleware) ValidateUploadRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查Content-Type
		contentType := c.GetHeader("Content-Type")
		if !strings.Contains(contentType, "multipart/form-data") {
			response.Error(c, http.StatusBadRequest, "请求必须是multipart/form-data格式")
			c.Abort()
			return
		}

		// 检查Content-Length
		contentLengthStr := c.GetHeader("Content-Length")
		if contentLengthStr == "" {
			response.Error(c, http.StatusBadRequest, "缺少Content-Length头")
			c.Abort()
			return
		}

		contentLength, err := strconv.ParseInt(contentLengthStr, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "无效的Content-Length")
			c.Abort()
			return
		}

		// 检查请求大小
		maxSize := um.configManager.GetMaxFileSize()
		if contentLength > maxSize {
			response.Error(c, http.StatusRequestEntityTooLarge, 
				fmt.Sprintf("请求大小超过限制，最大允许 %d 字节", maxSize))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitMiddleware 速率限制中间件
func (um *UploadMiddleware) RateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(limit, window)
	
	return func(c *gin.Context) {
		// 使用IP地址作为限制键
		clientIP := c.ClientIP()
		
		if !limiter.Allow(clientIP) {
			response.Error(c, http.StatusTooManyRequests, 
				fmt.Sprintf("请求频率超过限制，每%v最多%d次请求", window, limit))
			c.Abort()
			return
		}

		c.Next()
	}
}

// CheckUploadPermission 检查上传权限中间件
func (um *UploadMiddleware) CheckUploadPermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户信息
		userID, exists := c.Get("user_id")
		if !exists {
			response.Error(c, http.StatusUnauthorized, "用户未认证")
			c.Abort()
			return
		}

		userRole, _ := c.Get("user_role")
		
		// 检查用户是否有上传权限
		if !um.hasUploadPermission(userID.(uint), userRole) {
			response.Error(c, http.StatusForbidden, "无文件上传权限")
			c.Abort()
			return
		}

		c.Next()
	}
}

// CheckFileCategory 检查文件分类权限中间件
func (um *UploadMiddleware) CheckFileCategory() gin.HandlerFunc {
	return func(c *gin.Context) {
		category := c.PostForm("category")
		if category == "" {
			category = "general"
		}

		userRole, _ := c.Get("user_role")
		
		// 检查分类权限
		if !um.hasCategoryPermission(category, userRole) {
			response.Error(c, http.StatusForbidden, 
				fmt.Sprintf("无权限上传到分类: %s", category))
			c.Abort()
			return
		}

		c.Next()
	}
}

// LogUploadActivity 记录上传活动中间件
func (um *UploadMiddleware) LogUploadActivity() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// 记录请求信息
		userID, _ := c.Get("user_id")
		clientIP := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")
		
		c.Next()
		
		// 记录响应信息
		duration := time.Since(start)
		status := c.Writer.Status()
		
		// 这里可以记录到数据库或日志文件
		logData := map[string]interface{}{
			"user_id":      userID,
			"client_ip":    clientIP,
			"user_agent":   userAgent,
			"method":       c.Request.Method,
			"path":         c.Request.URL.Path,
			"status":       status,
			"duration_ms":  duration.Milliseconds(),
			"timestamp":    start,
		}
		
		// 输出到控制台（生产环境应该记录到日志系统）
		fmt.Printf("Upload Activity: %+v\n", logData)
	}
}

// ValidateFileType 验证文件类型中间件
func (um *UploadMiddleware) ValidateFileType() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 解析multipart表单
		err := c.Request.ParseMultipartForm(32 << 20) // 32MB
		if err != nil {
			response.Error(c, http.StatusBadRequest, "解析表单失败: "+err.Error())
			c.Abort()
			return
		}

		// 检查单文件上传
		if file, _, err := c.Request.FormFile("file"); err == nil {
			defer file.Close()
			
			// 读取文件头部检测类型
			buffer := make([]byte, 512)
			_, err := file.Read(buffer)
			if err != nil {
				response.Error(c, http.StatusBadRequest, "读取文件失败")
				c.Abort()
				return
			}
			
			// 重置文件指针
			file.Seek(0, 0)
			
			// 检测文件类型
			contentType := http.DetectContentType(buffer)
			if !um.configManager.IsAllowedType(contentType) {
				response.Error(c, http.StatusBadRequest, 
					fmt.Sprintf("不支持的文件类型: %s", contentType))
				c.Abort()
				return
			}
		}

		// 检查多文件上传
		if form := c.Request.MultipartForm; form != nil {
			files := form.File["files"]
			for _, fileHeader := range files {
				file, err := fileHeader.Open()
				if err != nil {
					continue
				}
				
				buffer := make([]byte, 512)
				file.Read(buffer)
				file.Close()
				
				contentType := http.DetectContentType(buffer)
				if !um.configManager.IsAllowedType(contentType) {
					response.Error(c, http.StatusBadRequest, 
						fmt.Sprintf("文件 %s 类型不支持: %s", fileHeader.Filename, contentType))
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}

// CheckStorageQuota 检查存储配额中间件
func (um *UploadMiddleware) CheckStorageQuota() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		// 检查用户存储配额
		if !um.checkUserQuota(userID.(uint)) {
			response.Error(c, http.StatusForbidden, "存储配额已满")
			c.Abort()
			return
		}

		c.Next()
	}
}

// SecurityScanMiddleware 安全扫描中间件
func (um *UploadMiddleware) SecurityScanMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !um.configManager.IsSecurityEnabled() {
			c.Next()
			return
		}

		// 这里可以集成病毒扫描、恶意文件检测等
		// 简化实现，只检查文件扩展名
		
		// 解析表单获取文件
		err := c.Request.ParseMultipartForm(32 << 20)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "解析表单失败")
			c.Abort()
			return
		}

		// 检查单文件
		if _, fileHeader, err := c.Request.FormFile("file"); err == nil {
			if um.configManager.IsForbiddenExt(fileHeader.Filename) {
				response.Error(c, http.StatusBadRequest, 
					fmt.Sprintf("文件类型被禁止: %s", fileHeader.Filename))
				c.Abort()
				return
			}
		}

		// 检查多文件
		if form := c.Request.MultipartForm; form != nil {
			files := form.File["files"]
			for _, fileHeader := range files {
				if um.configManager.IsForbiddenExt(fileHeader.Filename) {
					response.Error(c, http.StatusBadRequest, 
						fmt.Sprintf("文件 %s 类型被禁止", fileHeader.Filename))
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}

// hasUploadPermission 检查用户是否有上传权限
func (um *UploadMiddleware) hasUploadPermission(userID uint, userRole interface{}) bool {
	// 简化实现，实际应该查询数据库或权限系统
	if userRole == "admin" || userRole == "user" {
		return true
	}
	return false
}

// hasCategoryPermission 检查分类权限
func (um *UploadMiddleware) hasCategoryPermission(category string, userRole interface{}) bool {
	// 管理员可以上传到任何分类
	if userRole == "admin" {
		return true
	}

	// 普通用户的分类限制
	allowedCategories := []string{"general", "avatar", "document"}
	for _, allowed := range allowedCategories {
		if category == allowed {
			return true
		}
	}

	return false
}

// checkUserQuota 检查用户存储配额
func (um *UploadMiddleware) checkUserQuota(userID uint) bool {
	// 简化实现，实际应该查询数据库计算用户已使用的存储空间
	// 这里假设每个用户有100MB的配额
	const userQuota = 100 * 1024 * 1024 // 100MB
	
	// TODO: 实现实际的配额检查逻辑
	// 1. 查询用户已上传文件的总大小
	// 2. 与配额进行比较
	
	return true // 简化实现，总是返回true
}

// 全局中间件实例
var globalUploadMiddleware *UploadMiddleware

// InitGlobalUploadMiddleware 初始化全局上传中间件
func InitGlobalUploadMiddleware(configManager *upload.ConfigManager) {
	globalUploadMiddleware = NewUploadMiddleware(configManager)
}

// GetGlobalUploadMiddleware 获取全局上传中间件
func GetGlobalUploadMiddleware() *UploadMiddleware {
	return globalUploadMiddleware
}
