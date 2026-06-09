package contracts

import (
	"time"

	runtimeentity "ctf-platform/internal/module/runtime/entity"
)

type RuntimeManagedInstance struct {
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
	ShareScope     string     `gorm:"column:share_scope;size:16;not null;default:'per_user'"`
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

func (RuntimeManagedInstance) TableName() string {
	return "instances"
}

const (
	RuntimeManagedInstanceStatusPending  = "pending"
	RuntimeManagedInstanceStatusCreating = "creating"
	RuntimeManagedInstanceStatusRunning  = "running"
	RuntimeManagedInstanceStatusStopping = "stopping"
	RuntimeManagedInstanceStatusStopped  = "stopped"
	RuntimeManagedInstanceStatusExpired  = "expired"
	RuntimeManagedInstanceStatusFailed   = "failed"
)

type AWDDefenseWorkspace = runtimeentity.AWDDefenseWorkspace

const (
	AWDDefenseWorkspaceStatusPending      = runtimeentity.AWDDefenseWorkspaceStatusPending
	AWDDefenseWorkspaceStatusProvisioning = runtimeentity.AWDDefenseWorkspaceStatusProvisioning
	AWDDefenseWorkspaceStatusRunning      = runtimeentity.AWDDefenseWorkspaceStatusRunning
	AWDDefenseWorkspaceStatusFailed       = runtimeentity.AWDDefenseWorkspaceStatusFailed
)

type AWDServiceOperation = runtimeentity.AWDServiceOperation

const (
	AWDServiceOperationTypeStart    = runtimeentity.AWDServiceOperationTypeStart
	AWDServiceOperationTypeRestart  = runtimeentity.AWDServiceOperationTypeRestart
	AWDServiceOperationTypeRecover  = runtimeentity.AWDServiceOperationTypeRecover
	AWDServiceOperationTypeRecreate = runtimeentity.AWDServiceOperationTypeRecreate

	AWDServiceOperationRequestedByUser   = runtimeentity.AWDServiceOperationRequestedByUser
	AWDServiceOperationRequestedByAdmin  = runtimeentity.AWDServiceOperationRequestedByAdmin
	AWDServiceOperationRequestedBySystem = runtimeentity.AWDServiceOperationRequestedBySystem

	AWDServiceOperationStatusRequested    = runtimeentity.AWDServiceOperationStatusRequested
	AWDServiceOperationStatusProvisioning = runtimeentity.AWDServiceOperationStatusProvisioning
	AWDServiceOperationStatusRecovering   = runtimeentity.AWDServiceOperationStatusRecovering
	AWDServiceOperationStatusRecovered    = runtimeentity.AWDServiceOperationStatusRecovered
	AWDServiceOperationStatusSucceeded    = runtimeentity.AWDServiceOperationStatusSucceeded
	AWDServiceOperationStatusFailed       = runtimeentity.AWDServiceOperationStatusFailed
)

type AWDScopeControl = runtimeentity.AWDScopeControl

const (
	AWDScopeControlScopeTeam        = runtimeentity.AWDScopeControlScopeTeam
	AWDScopeControlScopeTeamService = runtimeentity.AWDScopeControlScopeTeamService

	AWDScopeControlTypeRetired                    = runtimeentity.AWDScopeControlTypeRetired
	AWDScopeControlTypeServiceDisabled            = runtimeentity.AWDScopeControlTypeServiceDisabled
	AWDScopeControlTypeDesiredReconcileSuppressed = runtimeentity.AWDScopeControlTypeDesiredReconcileSuppressed
)
