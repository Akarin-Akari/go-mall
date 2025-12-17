package middleware

import (
	"strconv"
	"time"

	"mall-go/pkg/metrics"

	"github.com/gin-gonic/gin"
)

// MonitoringMiddleware 监控中间件
func MonitoringMiddleware(collector *metrics.MetricsCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 处理请求
		c.Next()

		// 记录指标
		duration := time.Since(start)
		status := strconv.Itoa(c.Writer.Status())

		// 记录HTTP请求指标
		appMetrics := collector.GetApplicationMetrics()
		appMetrics.RecordHTTPRequest(c.Request.Method, status, duration)

		// 记录慢请求告警
		if duration > 2*time.Second {
			alertManager := metrics.NewAlertManager()
			alertManager.FireAlert(
				"slow_request",
				"warning",
				"Slow HTTP request detected",
				map[string]string{
					"method":   c.Request.Method,
					"path":     c.Request.URL.Path,
					"duration": duration.String(),
				},
			)
		}

		// 记录错误率
		if c.Writer.Status() >= 500 {
			errorCounter := metrics.NewCounter("http_errors_total", map[string]string{
				"method": c.Request.Method,
				"status": status,
				"path":   c.Request.URL.Path,
			})
			errorCounter.Inc()
		}
	}
}
