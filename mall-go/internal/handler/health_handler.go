package handler

import (
	"net/http"
	"time"

	"mall-go/pkg/cache"
	"mall-go/pkg/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HealthHandler 健康检查处理器
type HealthHandler struct {
	db *gorm.DB
}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Services  map[string]interface{} `json:"services"`
}

// Health 健康检查端点
func (h *HealthHandler) Health(c *gin.Context) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Services:  make(map[string]interface{}),
	}

	// 检查数据库
	dbStatus := h.checkDatabase()
	response.Services["database"] = dbStatus

	// 检查Redis
	redisStatus := h.checkRedis()
	response.Services["redis"] = redisStatus

	// 如果任何服务不健康，返回503
	if dbStatus["status"] != "healthy" {
		response.Status = "unhealthy"
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	// Redis降级不影响整体健康状态
	c.JSON(http.StatusOK, response)
}

// checkDatabase 检查数据库健康状态
func (h *HealthHandler) checkDatabase() map[string]interface{} {
	result := map[string]interface{}{
		"status": "healthy",
	}

	// 测试连接
	if err := database.PingDB(h.db); err != nil {
		result["status"] = "unhealthy"
		result["error"] = err.Error()
		return result
	}

	// 获取连接池统计
	stats, err := database.GetPoolStats(h.db)
	if err != nil {
		result["status"] = "degraded"
		result["error"] = err.Error()
		return result
	}

	result["pool"] = map[string]interface{}{
		"max_open":    stats.MaxOpenConnections,
		"open":        stats.OpenConnections,
		"in_use":      stats.InUse,
		"idle":        stats.Idle,
		"wait_count":  stats.WaitCount,
		"usage_percent": float64(stats.OpenConnections) / float64(stats.MaxOpenConnections) * 100,
	}

	// 检查连接池健康状态
	if !database.IsPoolHealthy(h.db) {
		result["status"] = "degraded"
		result["warning"] = "connection pool is under pressure"
	}

	return result
}

// checkRedis 检查Redis健康状态
func (h *HealthHandler) checkRedis() map[string]interface{} {
	result := map[string]interface{}{
		"status": "healthy",
	}

	// 检查容错客户端
	fallbackClient := cache.GetFallbackRedisClient()
	if fallbackClient == nil {
		result["status"] = "disabled"
		result["message"] = "redis client not initialized"
		return result
	}

	// 检查是否启用
	if !fallbackClient.IsEnabled() {
		result["status"] = "degraded"
		result["message"] = "redis is temporarily disabled due to failures"
		result["stats"] = fallbackClient.GetStats()
		return result
	}

	// 测试连接
	redisClient := cache.GetRedisClient()
	if redisClient != nil {
		if err := redisClient.Ping(); err != nil {
			result["status"] = "unhealthy"
			result["error"] = err.Error()
			return result
		}

		// 获取连接池统计
		stats := redisClient.GetConnectionStats()
		result["pool"] = map[string]interface{}{
			"total_conns": stats.TotalConns,
			"idle_conns":  stats.IdleConns,
			"stale_conns": stats.StaleConns,
			"hits":        stats.Hits,
			"misses":      stats.Misses,
			"timeouts":    stats.Timeouts,
		}
	}

	return result
}

// Readiness 就绪检查端点（用于Kubernetes）
func (h *HealthHandler) Readiness(c *gin.Context) {
	// 检查数据库连接
	if err := database.PingDB(h.db); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

// Liveness 存活检查端点（用于Kubernetes）
func (h *HealthHandler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}

