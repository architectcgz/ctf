package entity

import "time"

const (
	ProvisioningEventSeverityInfo    = "info"
	ProvisioningEventSeverityWarning = "warning"
	ProvisioningEventSeverityError   = "error"
)

type ProvisioningEvent struct {
	ID            int64     `gorm:"column:id;primaryKey"`
	InstanceID    int64     `gorm:"column:instance_id;not null;index"`
	Attempt       int       `gorm:"column:attempt;not null;default:0"`
	Stage         string    `gorm:"column:stage;size:64;not null;index"`
	Message       string    `gorm:"column:message;size:255;not null;default:''"`
	Severity      string    `gorm:"column:severity;size:16;not null;default:'info'"`
	RuntimeNodeID *int64    `gorm:"column:runtime_node_id;index"`
	Detail        string    `gorm:"column:detail;type:jsonb;not null;default:'{}'"`
	CreatedAt     time.Time `gorm:"column:created_at"`
}

func (ProvisioningEvent) TableName() string {
	return "instance_provisioning_events"
}
