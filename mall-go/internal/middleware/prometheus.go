package middleware

import (
	"net/http"
	"strconv"
	"time"

	"mall-go/pkg/metrics"

	"github.com/gin-gonic/gin"
)

// PrometheusMiddleware Prometheus监控中间件
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 获取请求大小
		requestSize := computeApproximateRequestSize(c.Request)

		// 处理请求
		c.Next()

		// 计算响应时间
		duration := time.Since(start).Seconds()

		// 获取响应大小
		responseSize := c.Writer.Size()

		// 获取Prometheus指标
		pm := metrics.GetPrometheusMetrics()

		// 记录HTTP请求指标
		status := strconv.Itoa(c.Writer.Status())
		pm.HTTPRequestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			status,
		).Inc()

		pm.HTTPRequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
		).Observe(duration)

		pm.HTTPRequestSize.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
		).Observe(float64(requestSize))

		if responseSize > 0 {
			pm.HTTPResponseSize.WithLabelValues(
				c.Request.Method,
				c.FullPath(),
			).Observe(float64(responseSize))
		}
	}
}

// computeApproximateRequestSize 计算请求大小（近似值）
func computeApproximateRequestSize(r *http.Request) int {
	s := 0
	if r.URL != nil {
		s = len(r.URL.Path)
	}

	s += len(r.Method)
	s += len(r.Proto)
	for name, values := range r.Header {
		s += len(name)
		for _, value := range values {
			s += len(value)
		}
	}
	s += len(r.Host)

	// 如果有Content-Length，使用它
	if r.ContentLength != -1 {
		s += int(r.ContentLength)
	}

	return s
}
