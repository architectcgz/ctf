package model

import "time"

const (
	AWDScopeControlScopeTeam        = "team"
	AWDScopeControlScopeTeamService = "team_service"

	AWDScopeControlTypeRetired                    = "retired"
	AWDScopeControlTypeServiceDisabled            = "service_disabled"
	AWDScopeControlTypeDesiredReconcileSuppressed = "desired_reconcile_suppressed"
)

type AWDScopeControl struct {
	ID          int64      `gorm:"column:id;primaryKey"`
	ContestID   int64      `gorm:"column:contest_id;not null;index:idx_awd_scope_controls_scope,priority:1;uniqueIndex:uk_awd_scope_controls"`
	TeamID      int64      `gorm:"column:team_id;not null;index:idx_awd_scope_controls_scope,priority:2;uniqueIndex:uk_awd_scope_controls"`
	ScopeType   string     `gorm:"column:scope_type;size:24;not null;index:idx_awd_scope_controls_scope,priority:3;uniqueIndex:uk_awd_scope_controls"`
	ServiceID   int64      `gorm:"column:service_id;not null;default:0;index:idx_awd_scope_controls_scope,priority:4;uniqueIndex:uk_awd_scope_controls"`
	ControlType string     `gorm:"column:control_type;size:48;not null;uniqueIndex:uk_awd_scope_controls"`
	Reason      string     `gorm:"column:reason;type:text;not null;default:''"`
	UpdatedBy   *int64     `gorm:"column:updated_by"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at"`
}

func (AWDScopeControl) TableName() string {
	return "awd_scope_controls"
}
