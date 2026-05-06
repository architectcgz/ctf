package model

import "time"

const (
	AWDDefenseWorkspaceStatusPending      = "pending"
	AWDDefenseWorkspaceStatusProvisioning = "provisioning"
	AWDDefenseWorkspaceStatusRunning      = "running"
	AWDDefenseWorkspaceStatusFailed       = "failed"
)

type AWDDefenseWorkspace struct {
	ID                int64     `gorm:"column:id;primaryKey"`
	ContestID         int64     `gorm:"column:contest_id;not null;uniqueIndex:uk_awd_defense_workspaces_scope,priority:1"`
	TeamID            int64     `gorm:"column:team_id;not null;uniqueIndex:uk_awd_defense_workspaces_scope,priority:2"`
	ServiceID         int64     `gorm:"column:service_id;not null;uniqueIndex:uk_awd_defense_workspaces_scope,priority:3"`
	InstanceID        int64     `gorm:"column:instance_id;not null;index:idx_awd_defense_workspaces_instance_id"`
	WorkspaceRevision int64     `gorm:"column:workspace_revision;not null;default:1"`
	Status            string    `gorm:"column:status;size:24;not null;default:'pending'"`
	ContainerID       string    `gorm:"column:container_id;size:64;not null;default:''"`
	SeedSignature     string    `gorm:"column:seed_signature;type:text;not null;default:''"`
	CreatedAt         time.Time `gorm:"column:created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at"`
}

func (AWDDefenseWorkspace) TableName() string {
	return "awd_defense_workspaces"
}
