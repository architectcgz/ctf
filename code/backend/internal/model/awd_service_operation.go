package model

import "time"

const (
	AWDServiceOperationTypeStart    = "start"
	AWDServiceOperationTypeRestart  = "restart"
	AWDServiceOperationTypeRecover  = "recover"
	AWDServiceOperationTypeRecreate = "recreate"

	AWDServiceOperationRequestedByUser   = "user"
	AWDServiceOperationRequestedByAdmin  = "admin"
	AWDServiceOperationRequestedBySystem = "system"

	AWDServiceOperationStatusRequested    = "requested"
	AWDServiceOperationStatusProvisioning = "provisioning"
	AWDServiceOperationStatusRecovering   = "recovering"
	AWDServiceOperationStatusRecovered    = "recovered"
	AWDServiceOperationStatusSucceeded    = "succeeded"
	AWDServiceOperationStatusFailed       = "failed"
)

type AWDServiceOperation struct {
	ID            int64      `gorm:"column:id;primaryKey"`
	ContestID     int64      `gorm:"column:contest_id;not null;index:idx_awd_service_operations_scope,priority:1"`
	TeamID        int64      `gorm:"column:team_id;not null;index:idx_awd_service_operations_scope,priority:2"`
	ServiceID     int64      `gorm:"column:service_id;not null;index:idx_awd_service_operations_scope,priority:3"`
	InstanceID    int64      `gorm:"column:instance_id;not null;index"`
	OperationType string     `gorm:"column:operation_type;size:24;not null"`
	RequestedBy   string     `gorm:"column:requested_by;size:16;not null"`
	RequestedByID *int64     `gorm:"column:requested_by_id"`
	Reason        string     `gorm:"column:reason;size:128;not null;default:''"`
	SLABillable   bool       `gorm:"column:sla_billable;not null"`
	Status        string     `gorm:"column:status;size:24;not null"`
	ErrorMessage  string     `gorm:"column:error_message;type:text;not null;default:''"`
	StartedAt     time.Time  `gorm:"column:started_at;not null;index:idx_awd_service_operations_window,priority:1"`
	FinishedAt    *time.Time `gorm:"column:finished_at;index:idx_awd_service_operations_window,priority:2"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at"`
}

func (AWDServiceOperation) TableName() string {
	return "awd_service_operations"
}
