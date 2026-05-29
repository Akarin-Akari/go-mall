package audit

import (
	"mall-go/internal/model"
	"mall-go/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AuditService 审计日志服务
type AuditService struct {
	db *gorm.DB
}

// NewAuditService 创建审计日志服务
func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{db: db}
}

// LogEntry 审计日志写入参数
type LogEntry struct {
	UserID     uint
	Action     string
	TargetType string
	TargetID   string
	IP         string
	UserAgent  string
	OldValue   string
	NewValue   string
	Remark     string
}

// Log 写入一条审计日志。
//
// 设计原则：审计写入失败绝不能阻断主业务流程（如修改密码本身已成功），
// 因此这里吞掉错误仅记录到应用日志，由运维通过日志告警发现审计落库异常。
func (as *AuditService) Log(entry LogEntry) {
	record := &model.AuditLog{
		UserID:     entry.UserID,
		Action:     entry.Action,
		TargetType: entry.TargetType,
		TargetID:   entry.TargetID,
		IP:         entry.IP,
		UserAgent:  entry.UserAgent,
		OldValue:   entry.OldValue,
		NewValue:   entry.NewValue,
		Remark:     entry.Remark,
	}
	if err := as.db.Create(record).Error; err != nil {
		logger.Error("写入审计日志失败",
			zap.String("action", entry.Action),
			zap.Uint("user_id", entry.UserID),
			zap.Error(err))
	}
}

// 全局审计服务实例
var globalAuditService *AuditService

// InitGlobalAuditService 初始化全局审计服务
func InitGlobalAuditService(db *gorm.DB) {
	globalAuditService = NewAuditService(db)
}

// GetGlobalAuditService 获取全局审计服务（可能为 nil，调用方需判空）
func GetGlobalAuditService() *AuditService {
	return globalAuditService
}
