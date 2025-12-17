package service

import (
	"context"
	"encoding/json"
	"time"

	"mall-go/pkg/logger"

	"go.uber.org/zap"
)

// AuditLogger 审计日志记录器
type AuditLogger struct {
	logger *zap.Logger
}

// NewAuditLogger 创建审计日志记录器
func NewAuditLogger() *AuditLogger {
	return &AuditLogger{
		logger: logger.GetLogger(),
	}
}

// AuditLogEntry 审计日志条目
type AuditLogEntry struct {
	Timestamp   time.Time   `json:"timestamp"`
	UserID      uint        `json:"user_id"`
	Operation   string      `json:"operation"`
	Resource    string      `json:"resource"`
	ResourceID  uint        `json:"resource_id,omitempty"`
	Action      string      `json:"action"`
	Status      string      `json:"status"`
	Duration    int64       `json:"duration_ms"`
	IPAddress   string      `json:"ip_address,omitempty"`
	UserAgent   string      `json:"user_agent,omitempty"`
	RequestID   string      `json:"request_id,omitempty"`
	Details     interface{} `json:"details,omitempty"`
	Error       string      `json:"error,omitempty"`
	Changes     interface{} `json:"changes,omitempty"`
}

// OperationContext 操作上下文
type OperationContext struct {
	UserID      uint
	Operation   string
	Resource    string
	ResourceID  uint
	Action      string
	IPAddress   string
	UserAgent   string
	RequestID   string
	StartTime   time.Time
	Details     interface{}
	Changes     interface{}
}

// LogOperation 记录操作日志
func (a *AuditLogger) LogOperation(ctx context.Context, opCtx *OperationContext, status string, err error) {
	duration := time.Since(opCtx.StartTime).Milliseconds()
	
	entry := AuditLogEntry{
		Timestamp:  time.Now(),
		UserID:     opCtx.UserID,
		Operation:  opCtx.Operation,
		Resource:   opCtx.Resource,
		ResourceID: opCtx.ResourceID,
		Action:     opCtx.Action,
		Status:     status,
		Duration:   duration,
		IPAddress:  opCtx.IPAddress,
		UserAgent:  opCtx.UserAgent,
		RequestID:  opCtx.RequestID,
		Details:    opCtx.Details,
		Changes:    opCtx.Changes,
	}
	
	if err != nil {
		entry.Error = err.Error()
	}
	
	// 序列化为JSON
	entryJSON, _ := json.Marshal(entry)
	
	// 根据状态选择日志级别
	switch status {
	case "success":
		a.logger.Info("操作审计日志",
			zap.String("audit_log", string(entryJSON)),
			zap.Uint("user_id", opCtx.UserID),
			zap.String("operation", opCtx.Operation),
			zap.String("action", opCtx.Action),
			zap.Int64("duration_ms", duration),
		)
	case "error":
		a.logger.Error("操作审计日志",
			zap.String("audit_log", string(entryJSON)),
			zap.Uint("user_id", opCtx.UserID),
			zap.String("operation", opCtx.Operation),
			zap.String("action", opCtx.Action),
			zap.Int64("duration_ms", duration),
			zap.Error(err),
		)
	case "warning":
		a.logger.Warn("操作审计日志",
			zap.String("audit_log", string(entryJSON)),
			zap.Uint("user_id", opCtx.UserID),
			zap.String("operation", opCtx.Operation),
			zap.String("action", opCtx.Action),
			zap.Int64("duration_ms", duration),
		)
	default:
		a.logger.Info("操作审计日志",
			zap.String("audit_log", string(entryJSON)),
			zap.Uint("user_id", opCtx.UserID),
			zap.String("operation", opCtx.Operation),
			zap.String("action", opCtx.Action),
			zap.Int64("duration_ms", duration),
		)
	}
}

// LogSlowOperation 记录慢操作日志
func (a *AuditLogger) LogSlowOperation(ctx context.Context, opCtx *OperationContext, threshold time.Duration) {
	duration := time.Since(opCtx.StartTime)
	if duration > threshold {
		a.logger.Warn("慢操作检测",
			zap.Uint("user_id", opCtx.UserID),
			zap.String("operation", opCtx.Operation),
			zap.String("action", opCtx.Action),
			zap.Duration("duration", duration),
			zap.Duration("threshold", threshold),
			zap.String("resource", opCtx.Resource),
			zap.Uint("resource_id", opCtx.ResourceID),
			zap.Any("details", opCtx.Details),
		)
	}
}

// LogUserBehavior 记录用户行为日志
func (a *AuditLogger) LogUserBehavior(ctx context.Context, userID uint, behavior string, details interface{}) {
	a.logger.Info("用户行为追踪",
		zap.Uint("user_id", userID),
		zap.String("behavior", behavior),
		zap.Time("timestamp", time.Now()),
		zap.Any("details", details),
	)
}

// LogSecurityEvent 记录安全事件日志
func (a *AuditLogger) LogSecurityEvent(ctx context.Context, userID uint, event string, severity string, details interface{}) {
	switch severity {
	case "high":
		a.logger.Error("安全事件",
			zap.Uint("user_id", userID),
			zap.String("event", event),
			zap.String("severity", severity),
			zap.Time("timestamp", time.Now()),
			zap.Any("details", details),
		)
	case "medium":
		a.logger.Warn("安全事件",
			zap.Uint("user_id", userID),
			zap.String("event", event),
			zap.String("severity", severity),
			zap.Time("timestamp", time.Now()),
			zap.Any("details", details),
		)
	default:
		a.logger.Info("安全事件",
			zap.Uint("user_id", userID),
			zap.String("event", event),
			zap.String("severity", severity),
			zap.Time("timestamp", time.Now()),
			zap.Any("details", details),
		)
	}
}

// LogDataChange 记录数据变更日志
func (a *AuditLogger) LogDataChange(ctx context.Context, userID uint, table string, recordID uint, action string, oldData, newData interface{}) {
	a.logger.Info("数据变更日志",
		zap.Uint("user_id", userID),
		zap.String("table", table),
		zap.Uint("record_id", recordID),
		zap.String("action", action),
		zap.Time("timestamp", time.Now()),
		zap.Any("old_data", oldData),
		zap.Any("new_data", newData),
	)
}

// LogPerformanceMetrics 记录性能指标日志
func (a *AuditLogger) LogPerformanceMetrics(ctx context.Context, operation string, metrics map[string]interface{}) {
	a.logger.Info("性能指标",
		zap.String("operation", operation),
		zap.Time("timestamp", time.Now()),
		zap.Any("metrics", metrics),
	)
}

// LogError 记录错误日志
func (a *AuditLogger) LogError(ctx context.Context, userID uint, operation string, err error, details interface{}) {
	a.logger.Error("操作错误",
		zap.Uint("user_id", userID),
		zap.String("operation", operation),
		zap.Error(err),
		zap.Time("timestamp", time.Now()),
		zap.Any("details", details),
	)
}

// CreateOperationContext 创建操作上下文
func CreateOperationContext(userID uint, operation, resource, action string) *OperationContext {
	return &OperationContext{
		UserID:    userID,
		Operation: operation,
		Resource:  resource,
		Action:    action,
		StartTime: time.Now(),
	}
}

// WithResourceID 设置资源ID
func (oc *OperationContext) WithResourceID(resourceID uint) *OperationContext {
	oc.ResourceID = resourceID
	return oc
}

// WithIPAddress 设置IP地址
func (oc *OperationContext) WithIPAddress(ipAddress string) *OperationContext {
	oc.IPAddress = ipAddress
	return oc
}

// WithUserAgent 设置用户代理
func (oc *OperationContext) WithUserAgent(userAgent string) *OperationContext {
	oc.UserAgent = userAgent
	return oc
}

// WithRequestID 设置请求ID
func (oc *OperationContext) WithRequestID(requestID string) *OperationContext {
	oc.RequestID = requestID
	return oc
}

// WithDetails 设置详细信息
func (oc *OperationContext) WithDetails(details interface{}) *OperationContext {
	oc.Details = details
	return oc
}

// WithChanges 设置变更信息
func (oc *OperationContext) WithChanges(changes interface{}) *OperationContext {
	oc.Changes = changes
	return oc
}

// GetDuration 获取操作持续时间
func (oc *OperationContext) GetDuration() time.Duration {
	return time.Since(oc.StartTime)
}
