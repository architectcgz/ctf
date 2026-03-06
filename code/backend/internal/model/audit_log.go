package model

import "time"

const (
	AuditActionLogin   = "login"
	AuditActionLogout  = "logout"
	AuditActionCreate  = "create"
	AuditActionUpdate  = "update"
	AuditActionDelete  = "delete"
	AuditActionSubmit  = "submit"
	AuditActionAdminOp = "admin_op"
)

type AuditLog struct {
	ID           int64     `gorm:"column:id;primaryKey"`
	UserID       *int64    `gorm:"column:user_id"`
	Action       string    `gorm:"column:action"`
	ResourceType string    `gorm:"column:resource_type"`
	ResourceID   *int64    `gorm:"column:resource_id"`
	Detail       string    `gorm:"column:detail;type:jsonb;default:'{}'"`
	IPAddress    string    `gorm:"column:ip_address"`
	UserAgent    *string   `gorm:"column:user_agent"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
