package entity

import "time"

type ShareScope string

const (
	ShareScopePerUser ShareScope = "per_user"
	ShareScopePerTeam ShareScope = "per_team"
	ShareScopeShared  ShareScope = "shared"
)

type Instance struct {
	ID             int64      `gorm:"primaryKey"`
	UserID         int64      `gorm:"not null;index"`
	ContestID      *int64     `gorm:"column:contest_id;index"`
	TeamID         *int64     `gorm:"column:team_id;index"`
	ChallengeID    int64      `gorm:"not null;index"`
	ServiceID      *int64     `gorm:"column:service_id;index"`
	NodeID         *int64     `gorm:"column:node_id;index"`
	HostPort       int        `gorm:"column:host_port;index"`
	ContainerID    string     `gorm:"size:64;not null"`
	NetworkID      string     `gorm:"size:64"`
	RuntimeDetails string     `gorm:"column:runtime_details;type:text"`
	ShareScope     ShareScope `gorm:"column:share_scope;size:16;not null;default:'per_user'"`
	Status         string     `gorm:"size:16;not null;index"`
	AccessURL      string     `gorm:"size:255"`
	Nonce          string     `gorm:"size:64"`
	FlagKeyID      string     `gorm:"column:flag_key_id;size:128"`
	ExpiresAt      time.Time  `gorm:"not null;index"`
	DestroyedAt    *time.Time `gorm:"column:destroyed_at;index"`
	ExtendCount    int        `gorm:"default:0"`
	MaxExtends     int        `gorm:"default:2"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

const (
	InstanceStatusPending  = "pending"
	InstanceStatusCreating = "creating"
	InstanceStatusRunning  = "running"
	InstanceStatusStopping = "stopping"
	InstanceStatusStopped  = "stopped"
	InstanceStatusExpired  = "expired"
	InstanceStatusFailed   = "failed"
)
