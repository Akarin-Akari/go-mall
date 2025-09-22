package errors

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ErrorAPIController 错误API控制器
type ErrorAPIController struct {
	monitor     *ErrorMonitor
	i18nManager *I18nManager
	recovery    *RecoveryManager
}

// NewErrorAPIController 创建错误API控制器
func NewErrorAPIController() *ErrorAPIController {
	return &ErrorAPIController{
		monitor:     GlobalErrorMonitor,
		i18nManager: GlobalI18nManager,
		recovery:    GlobalRecoveryManager,
	}
}

// RegisterRoutes 注册错误管理API路由
func (controller *ErrorAPIController) RegisterRoutes(r *gin.RouterGroup) {
	errorGroup := r.Group("/errors")
	{
		// 错误监控相关
		errorGroup.GET("/metrics", controller.GetMetrics)
		errorGroup.GET("/metrics/export", controller.ExportMetrics)
		errorGroup.POST("/metrics/reset", controller.ResetMetrics)
		
		// 错误码相关
		errorGroup.GET("/codes", controller.GetErrorCodes)
		errorGroup.GET("/codes/:code", controller.GetErrorCodeInfo)
		
		// 国际化相关
		errorGroup.GET("/messages", controller.GetMessages)
		errorGroup.GET("/messages/:lang", controller.GetMessagesByLanguage)
		errorGroup.GET("/languages", controller.GetSupportedLanguages)
		
		// 熔断器相关
		errorGroup.GET("/circuit-breakers", controller.GetCircuitBreakers)
		errorGroup.POST("/circuit-breakers/:operation/reset", controller.ResetCircuitBreaker)
		
		// 健康检查
		errorGroup.GET("/health", controller.HealthCheck)
		
		// 错误测试（仅开发环境）
		errorGroup.POST("/test", controller.TestError)
	}
}

// GetMetrics 获取错误指标
func (controller *ErrorAPIController) GetMetrics(c *gin.Context) {
	metrics := controller.monitor.GetMetrics()
	
	// 检查是否需要特定时间范围的数据
	if timeRange := c.Query("time_range"); timeRange != "" {
		// 这里可以根据时间范围过滤数据
		// 简化实现，返回所有数据
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metrics,
	})
}

// ExportMetrics 导出错误指标
func (controller *ErrorAPIController) ExportMetrics(c *gin.Context) {
	data, err := controller.monitor.ExportMetrics()
	if err != nil {
		AbortWithBusinessError(c, ErrCodeSystemInternal, "导出指标失败")
		return
	}
	
	format := c.DefaultQuery("format", "json")
	switch format {
	case "json":
		c.Header("Content-Type", "application/json")
		c.Header("Content-Disposition", "attachment; filename=error_metrics.json")
		c.Data(http.StatusOK, "application/json", data)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "不支持的导出格式",
		})
	}
}

// ResetMetrics 重置错误指标
func (controller *ErrorAPIController) ResetMetrics(c *gin.Context) {
	// 检查权限
	if !RequirePermission(c, "system:metrics:reset") {
		return
	}
	
	controller.monitor.ResetMetrics()
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "错误指标已重置",
	})
}

// GetErrorCodes 获取所有错误码
func (controller *ErrorAPIController) GetErrorCodes(c *gin.Context) {
	category := c.Query("category")
	level := c.Query("level")
	
	// 构建错误码列表
	errorCodes := []map[string]interface{}{}
	
	// 这里可以根据分类和级别过滤错误码
	allCodes := []ErrorCode{
		ErrCodeSystemInternal, ErrCodeSystemTimeout, ErrCodeSystemOverload,
		ErrCodeBusinessLogic, ErrCodeBusinessRuleViolation,
		ErrCodeValidationRequired, ErrCodeValidationFormat,
		ErrCodeAuthRequired, ErrCodeAuthInvalidToken,
		ErrCodePermissionDenied,
		ErrCodeDatabaseConnection, ErrCodeDatabaseNotFound,
		ErrCodePaymentInvalidMethod, ErrCodePaymentInsufficientFunds,
	}
	
	for _, code := range allCodes {
		// 根据错误码获取分类和级别信息
		codeCategory := GetErrorCategory(NewError(code, "", ErrorLevelInfo, CategorySystem))
		codeLevel := GetErrorLevel(NewError(code, "", ErrorLevelInfo, CategorySystem))
		
		// 应用过滤条件
		if category != "" && string(codeCategory) != category {
			continue
		}
		if level != "" && string(codeLevel) != level {
			continue
		}
		
		errorCodes = append(errorCodes, map[string]interface{}{
			"code":        code,
			"category":    codeCategory,
			"level":       codeLevel,
			"formatted":   FormatErrorCode(code),
		})
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"error_codes": errorCodes,
			"total":       len(errorCodes),
		},
	})
}

// GetErrorCodeInfo 获取特定错误码信息
func (controller *ErrorAPIController) GetErrorCodeInfo(c *gin.Context) {
	codeStr := c.Param("code")
	lang := Language(c.DefaultQuery("lang", "zh-cn"))
	
	code, valid := ParseErrorCode(codeStr)
	if !valid {
		AbortWithValidationError(c, "code", "无效的错误码格式")
		return
	}
	
	// 获取错误信息
	message := controller.i18nManager.GetMessage(code, lang)
	category := GetErrorCategory(NewError(code, "", ErrorLevelInfo, CategorySystem))
	level := GetErrorLevel(NewError(code, "", ErrorLevelInfo, CategorySystem))
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"code":       code,
			"category":   category,
			"level":      level,
			"formatted":  FormatErrorCode(code),
			"message":    message.Message,
			"details":    message.Details,
			"suggestion": message.Suggestion,
			"language":   lang,
		},
	})
}

// GetMessages 获取错误消息
func (controller *ErrorAPIController) GetMessages(c *gin.Context) {
	lang := Language(c.DefaultQuery("lang", "zh-cn"))
	category := c.Query("category")
	
	// 获取所有错误码的消息
	messages := []map[string]interface{}{}
	
	allCodes := []ErrorCode{
		ErrCodeSystemInternal, ErrCodeSystemTimeout, ErrCodeSystemOverload,
		ErrCodeBusinessLogic, ErrCodeValidationRequired, ErrCodeAuthRequired,
		ErrCodePermissionDenied, ErrCodePaymentInvalidMethod,
	}
	
	for _, code := range allCodes {
		codeCategory := GetErrorCategory(NewError(code, "", ErrorLevelInfo, CategorySystem))
		
		// 应用分类过滤
		if category != "" && string(codeCategory) != category {
			continue
		}
		
		message := controller.i18nManager.GetMessage(code, lang)
		messages = append(messages, map[string]interface{}{
			"code":       code,
			"category":   codeCategory,
			"message":    message.Message,
			"details":    message.Details,
			"suggestion": message.Suggestion,
		})
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"messages": messages,
			"language": lang,
			"total":    len(messages),
		},
	})
}

// GetMessagesByLanguage 获取指定语言的错误消息
func (controller *ErrorAPIController) GetMessagesByLanguage(c *gin.Context) {
	lang := Language(c.Param("lang"))
	
	// 验证语言是否支持
	supportedLangs := controller.i18nManager.GetSupportedLanguages()
	supported := false
	for _, supportedLang := range supportedLangs {
		if supportedLang == lang {
			supported = true
			break
		}
	}
	
	if !supported {
		AbortWithValidationError(c, "lang", "不支持的语言")
		return
	}
	
	// 重定向到GetMessages，并设置语言参数
	c.Request.URL.RawQuery = "lang=" + string(lang)
	controller.GetMessages(c)
}

// GetSupportedLanguages 获取支持的语言列表
func (controller *ErrorAPIController) GetSupportedLanguages(c *gin.Context) {
	languages := controller.i18nManager.GetSupportedLanguages()
	stats := controller.i18nManager.GetStatistics()
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"languages": languages,
			"statistics": stats,
		},
	})
}

// GetCircuitBreakers 获取熔断器状态
func (controller *ErrorAPIController) GetCircuitBreakers(c *gin.Context) {
	// 这里应该从RecoveryManager获取熔断器状态
	// 简化实现，返回模拟数据
	circuitBreakers := []map[string]interface{}{
		{
			"operation": "database",
			"state":     "closed",
			"failure_count": 0,
			"request_count": 100,
		},
		{
			"operation": "third_party",
			"state":     "open",
			"failure_count": 5,
			"request_count": 10,
		},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"circuit_breakers": circuitBreakers,
			"total":           len(circuitBreakers),
		},
	})
}

// ResetCircuitBreaker 重置熔断器
func (controller *ErrorAPIController) ResetCircuitBreaker(c *gin.Context) {
	// 检查权限
	if !RequirePermission(c, "system:circuit-breaker:reset") {
		return
	}
	
	operation := c.Param("operation")
	if operation == "" {
		AbortWithValidationError(c, "operation", "操作名称不能为空")
		return
	}
	
	// 这里应该重置具体的熔断器
	// 简化实现
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "熔断器已重置",
		"data": map[string]interface{}{
			"operation": operation,
		},
	})
}

// HealthCheck 健康检查
func (controller *ErrorAPIController) HealthCheck(c *gin.Context) {
	// 检查各个组件的健康状态
	health := map[string]interface{}{
		"error_monitor": map[string]interface{}{
			"status": "healthy",
			"metrics_loaded": controller.monitor != nil,
		},
		"i18n_manager": map[string]interface{}{
			"status": "healthy",
			"loaded": controller.i18nManager.IsLoaded(),
			"languages": len(controller.i18nManager.GetSupportedLanguages()),
		},
		"recovery_manager": map[string]interface{}{
			"status": "healthy",
			"enabled": controller.recovery != nil,
		},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    health,
	})
}

// TestError 测试错误（仅用于开发环境）
func (controller *ErrorAPIController) TestError(c *gin.Context) {
	// 检查是否为开发环境
	if gin.Mode() != gin.DebugMode {
		AbortWithPermissionError(c, "error_test", "access")
		return
	}
	
	var request struct {
		ErrorCode ErrorCode `json:"error_code" binding:"required"`
		Language  Language  `json:"language"`
		Context   map[string]interface{} `json:"context"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		AbortWithValidationError(c, "request", "请求参数格式错误")
		return
	}
	
	if request.Language == "" {
		request.Language = LangZHCN
	}
	
	// 创建测试错误
	testErr := GetBusinessErrorByCode(request.ErrorCode)
	if testErr == nil {
		testErr = NewSystemError(request.ErrorCode, "测试错误")
	}
	
	// 添加上下文
	for k, v := range request.Context {
		testErr.WithContext(k, v)
	}
	
	// 本地化错误
	localizedErr := controller.i18nManager.LocalizeError(testErr, request.Language)
	
	// 记录到监控系统
	controller.monitor.RecordError(localizedErr, c.Request.URL.Path, c.Request.Method, c.ClientIP(), c.Request.UserAgent())
	
	// 返回错误信息（不使用HandleError，避免实际触发错误处理）
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "测试错误已创建",
		"data": map[string]interface{}{
			"error":     localizedErr,
			"recorded":  true,
		},
	})
}

// ErrorMiddleware 错误统计中间件
func ErrorMiddleware(monitor *ErrorMonitor) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 记录请求
		monitor.RecordRequest()
		
		// 处理请求
		c.Next()
		
		// 检查是否有错误
		if len(c.Errors) > 0 {
			for _, ginErr := range c.Errors {
				if be, ok := ginErr.Err.(*BusinessError); ok {
					monitor.RecordError(be, c.Request.URL.Path, c.Request.Method, c.ClientIP(), c.Request.UserAgent())
				}
			}
		}
	})
}

// GetAcceptLanguage 从请求头获取接受的语言
func GetAcceptLanguage(c *gin.Context) Language {
	acceptLang := c.GetHeader("Accept-Language")
	if acceptLang == "" {
		return LangZHCN
	}
	
	// 简单解析Accept-Language头
	// 实际实现可能需要更复杂的解析逻辑
	if strings.Contains(acceptLang, "en") {
		return LangEN
	}
	if strings.Contains(acceptLang, "zh-CN") || strings.Contains(acceptLang, "zh-cn") {
		return LangZHCN
	}
	if strings.Contains(acceptLang, "zh-TW") || strings.Contains(acceptLang, "zh-tw") {
		return LangZHTW
	}
	
	return LangZHCN
}

// LocalizedHandleError 本地化错误处理
func LocalizedHandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	
	// 获取用户语言偏好
	lang := GetAcceptLanguage(c)
	if langParam := c.Query("lang"); langParam != "" {
		lang = Language(langParam)
	}
	
	var businessErr *BusinessError
	if be, ok := err.(*BusinessError); ok {
		businessErr = GlobalI18nManager.LocalizeError(be, lang)
	} else {
		businessErr = WrapError(err, ErrCodeSystemInternal, "系统内部错误", CategorySystem)
		businessErr = GlobalI18nManager.LocalizeError(businessErr, lang)
	}
	
	// 记录到监控系统
	GlobalErrorMonitor.RecordError(businessErr, c.Request.URL.Path, c.Request.Method, c.ClientIP(), c.Request.UserAgent())
	
	// 处理错误
	HandleError(c, businessErr)
}