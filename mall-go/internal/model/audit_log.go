package model

import "time"

// AuditLog 敏感操作审计日志
//
// 用于记录密码修改、邮箱变更、角色分配、账户锁定/解锁等安全敏感操作，
// 与仅记录登录行为的 UserLoginLog 互补。敏感字段（如密码）只记录"已变更"事实，
// 绝不记录明文值。
type AuditLog struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	UserID     uint      `gorm:"index;not null" json:"user_id"`        // 操作者用户ID
	Action     string    `gorm:"size:50;not null;index" json:"action"` // 操作类型，见 AuditAction* 常量
	TargetType string    `gorm:"size:50" json:"target_type"`           // 目标类型：user/order/role 等
	TargetID   string    `gorm:"size:64;index" json:"target_id"`       // 目标ID（字符串以兼容各类主键）
	IP         string    `gorm:"size:45" json:"ip"`                    // 操作来源IP（兼容IPv6）
	UserAgent  string    `gorm:"size:255" json:"user_agent"`           // 客户端UA
	OldValue   string    `gorm:"type:text" json:"old_value"`           // 变更前值（脱敏后，密码类不记录）
	NewValue   string    `gorm:"type:text" json:"new_value"`           // 变更后值（脱敏后，密码类不记录）
	Remark     string    `gorm:"size:255" json:"remark"`               // 备注
	CreatedAt  time.Time `json:"created_at"`
}

// TableName 指定表名
func (AuditLog) TableName() string {
	return "audit_logs"
}

// 审计操作类型常量
const (
	AuditActionChangePassword = "change_password" // 修改密码
	AuditActionChangeEmail    = "change_email"    // 修改邮箱
	AuditActionChangePhone    = "change_phone"    // 修改手机号
	AuditActionResetPassword  = "reset_password"  // 重置密码（忘记密码流程）
	AuditActionAssignRole     = "assign_role"     // 分配角色
	AuditActionLockAccount    = "lock_account"    // 锁定账户
	AuditActionUnlockAccount  = "unlock_account"  // 解锁账户
)
