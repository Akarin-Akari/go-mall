package performance

import (
	"net/http"
	"strconv"
	"time"

	"mall-go/internal/middleware"
	"mall-go/internal/service"
	"mall-go/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Handler 性能监控处理器
type Handler struct {
	performanceMonitor *service.PerformanceMonitor
	metrics           *middleware.PerformanceMetrics
}

// NewHandler 创建性能监控处理器
func NewHandler() *Handler {
	return &Handler{
		performanceMonitor: service.GlobalPerformanceMonitor,
		metrics:           middleware.GlobalMetrics,
	}
}

// GetMetrics 获取性能指标
func (h *Handler) GetMetrics(c *gin.Context) {
	if h.performanceMonitor == nil {
		response.Error(c, http.StatusInternalServerError, "性能监控器未初始化")
		return
	}
	
	metricsData, err := h.performanceMonitor.GetMetricsJSON()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取性能指标失败: "+err.Error())
		return
	}
	
	c.Header("Content-Type", "application/json")
	c.Data(http.StatusOK, "application/json", metricsData)
}

// GetSystemMetrics 获取系统指标
func (h *Handler) GetSystemMetrics(c *gin.Context) {
	if h.metrics == nil {
		response.Error(c, http.StatusInternalServerError, "性能指标收集器未初始化")
		return
	}
	
	systemMetrics := h.metrics.GetMetrics()
	response.Success(c, systemMetrics)
}

// GetPerformanceReport 获取性能报告
func (h *Handler) GetPerformanceReport(c *gin.Context) {
	if h.performanceMonitor == nil {
		response.Error(c, http.StatusInternalServerError, "性能监控器未初始化")
		return
	}
	
	// 获取时间范围参数
	timeRange := c.DefaultQuery("time_range", "1h")
	
	report := h.performanceMonitor.GenerateReport(timeRange)
	response.Success(c, report)
}

// GetHealthCheck 健康检查
func (h *Handler) GetHealthCheck(c *gin.Context) {
	healthStatus := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"uptime":    time.Since(startTime),
		"version":   "1.0.0",
		"services": map[string]string{
			"database":           "healthy",
			"performance_monitor": "healthy",
			"metrics_collector":   "healthy",
		},
	}
	
	response.Success(c, healthStatus)
}

// GetPrometheusMetrics 获取Prometheus格式指标
func (h *Handler) GetPrometheusMetrics() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}

// GetDashboard 获取性能监控仪表板数据
func (h *Handler) GetDashboard(c *gin.Context) {
	if h.performanceMonitor == nil {
		response.Error(c, http.StatusInternalServerError, "性能监控器未初始化")
		return
	}
	
	// 获取时间范围参数
	timeRange := c.DefaultQuery("time_range", "1h")
	
	// 生成仪表板数据
	dashboardData := h.generateDashboardData(timeRange)
	response.Success(c, dashboardData)
}

// generateDashboardData 生成仪表板数据
func (h *Handler) generateDashboardData(timeRange string) map[string]interface{} {
	// 获取性能报告
	report := h.performanceMonitor.GenerateReport(timeRange)
	
	// 获取系统指标
	systemMetrics := h.metrics.GetMetrics()
	
	// 构建仪表板数据
	dashboardData := map[string]interface{}{
		"overview": map[string]interface{}{
			"time_range":    timeRange,
			"generated_at":  time.Now(),
			"total_metrics": len(report.Metrics),
			"alert_count":   len(report.Alerts),
		},
		"system": systemMetrics,
		"performance": map[string]interface{}{
			"summary":         report.Summary,
			"recommendations": report.Recommendations,
		},
		"alerts": report.Alerts,
		"charts": h.generateChartData(report),
	}
	
	return dashboardData
}

// generateChartData 生成图表数据
func (h *Handler) generateChartData(report *service.PerformanceReport) map[string]interface{} {
	charts := make(map[string]interface{})
	
	// 内存使用趋势图
	if metric, exists := report.Metrics["memory_usage_mb"]; exists {
		charts["memory_trend"] = map[string]interface{}{
			"title": "内存使用趋势",
			"type":  "line",
			"data":  metric.History,
			"unit":  "MB",
		}
	}
	
	// Goroutine数量趋势图
	if metric, exists := report.Metrics["goroutine_count"]; exists {
		charts["goroutine_trend"] = map[string]interface{}{
			"title": "Goroutine数量趋势",
			"type":  "line",
			"data":  metric.History,
			"unit":  "个",
		}
	}
	
	// 响应时间分布图
	charts["response_time_distribution"] = map[string]interface{}{
		"title": "响应时间分布",
		"type":  "histogram",
		"data":  h.getResponseTimeDistribution(),
		"unit":  "ms",
	}
	
	return charts
}

// getResponseTimeDistribution 获取响应时间分布
func (h *Handler) getResponseTimeDistribution() []map[string]interface{} {
	// 这里应该从实际的指标数据中获取响应时间分布
	// 为了演示，返回模拟数据
	return []map[string]interface{}{
		{"range": "0-100ms", "count": 850},
		{"range": "100-500ms", "count": 120},
		{"range": "500ms-1s", "count": 25},
		{"range": "1s-5s", "count": 4},
		{"range": ">5s", "count": 1},
	}
}

// GetAlerts 获取告警信息
func (h *Handler) GetAlerts(c *gin.Context) {
	if h.performanceMonitor == nil {
		response.Error(c, http.StatusInternalServerError, "性能监控器未初始化")
		return
	}
	
	// 获取查询参数
	severityFilter := c.Query("severity")
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	
	// 生成报告获取告警
	report := h.performanceMonitor.GenerateReport("24h")
	alerts := report.Alerts
	
	// 按严重程度过滤
	if severityFilter != "" {
		filteredAlerts := make([]service.Alert, 0)
		for _, alert := range alerts {
			if alert.Severity == severityFilter {
				filteredAlerts = append(filteredAlerts, alert)
			}
		}
		alerts = filteredAlerts
	}
	
	// 限制数量
	if limit > 0 && len(alerts) > limit {
		alerts = alerts[:limit]
	}
	
	response.Success(c, map[string]interface{}{
		"alerts": alerts,
		"total":  len(alerts),
		"filter": map[string]interface{}{
			"severity": severityFilter,
			"limit":    limit,
		},
	})
}

// GetSlowQueries 获取慢查询信息
func (h *Handler) GetSlowQueries(c *gin.Context) {
	// 获取查询参数
	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(limitStr)
	
	thresholdStr := c.DefaultQuery("threshold", "100")
	threshold, _ := strconv.Atoi(thresholdStr)
	
	// 这里应该从实际的慢查询日志中获取数据
	// 为了演示，返回模拟数据
	slowQueries := []map[string]interface{}{
		{
			"query":     "SELECT * FROM addresses WHERE user_id = ?",
			"duration":  150,
			"timestamp": time.Now().Add(-5 * time.Minute),
			"table":     "addresses",
			"operation": "select",
		},
		{
			"query":     "UPDATE addresses SET is_default = ? WHERE user_id = ?",
			"duration":  120,
			"timestamp": time.Now().Add(-10 * time.Minute),
			"table":     "addresses",
			"operation": "update",
		},
	}
	
	// 按阈值过滤
	filteredQueries := make([]map[string]interface{}, 0)
	for _, query := range slowQueries {
		if duration, ok := query["duration"].(int); ok && duration >= threshold {
			filteredQueries = append(filteredQueries, query)
		}
	}
	
	// 限制数量
	if limit > 0 && len(filteredQueries) > limit {
		filteredQueries = filteredQueries[:limit]
	}
	
	response.Success(c, map[string]interface{}{
		"slow_queries": filteredQueries,
		"total":        len(filteredQueries),
		"filter": map[string]interface{}{
			"threshold": threshold,
			"limit":     limit,
		},
	})
}

// 启动时间，用于计算运行时间
var startTime = time.Now()

// RegisterRoutes 注册性能监控路由
func RegisterRoutes(r *gin.Engine) {
	handler := NewHandler()
	
	// 性能监控API组
	perfGroup := r.Group("/api/v1/performance")
	{
		perfGroup.GET("/metrics", handler.GetMetrics)
		perfGroup.GET("/system", handler.GetSystemMetrics)
		perfGroup.GET("/report", handler.GetPerformanceReport)
		perfGroup.GET("/dashboard", handler.GetDashboard)
		perfGroup.GET("/alerts", handler.GetAlerts)
		perfGroup.GET("/slow-queries", handler.GetSlowQueries)
	}
	
	// 健康检查
	r.GET("/health", handler.GetHealthCheck)
	
	// Prometheus指标端点
	r.GET("/metrics", handler.GetPrometheusMetrics())
}
