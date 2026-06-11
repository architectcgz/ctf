package ports

import (
	"context"
	"time"

	instanceentity "ctf-platform/internal/module/instance/entity"
)

type InstanceLookupRepository interface {
	FindByID(ctx context.Context, id int64) (*instanceentity.Instance, error)
}

type InstanceUserLookupRepository interface {
	FindUserByID(ctx context.Context, userID int64) (*InstanceUser, error)
}

type InstanceAccessRepository interface {
	FindAccessibleByIDForUser(ctx context.Context, instanceID, userID int64) (*instanceentity.Instance, error)
}

type UserVisibleInstanceRepository interface {
	ListVisibleByUser(ctx context.Context, userID int64) ([]UserVisibleInstanceRow, error)
}

type TeacherInstanceQueryRepository interface {
	ListTeacherInstances(ctx context.Context, filter TeacherInstanceFilter) (*TeacherInstancePage, error)
}

type InstanceExtendRepository interface {
	AtomicExtendByID(ctx context.Context, id int64, maxExtends int, duration time.Duration) error
}

type RunningInstanceCountRepository interface {
	CountRunningInstances(ctx context.Context) (int64, error)
}

type RuntimeCleaner interface {
	CleanupRuntime(ctx context.Context, instance *instanceentity.Instance) error
}

type InstanceUser struct {
	ID        int64
	Role      string
	ClassName string
}

const InstanceUserRoleStudent = "student"

type TeacherInstanceFilter struct {
	ClassName string
	Keyword   string
	StudentNo string
	Status    string
	Page      int
	PageSize  int
}

type TeacherInstanceListSummary struct {
	TotalCount        int64
	RunningCount      int64
	ExpiringSoonCount int64
	WarningCount      int64
}

type TeacherInstancePage struct {
	List    []TeacherInstanceRow
	Total   int64
	Summary TeacherInstanceListSummary
}

type UserVisibleInstanceRow struct {
	ID             int64
	ContestMode    string
	ChallengeID    int64
	ChallengeTitle string
	Category       string
	Difficulty     string
	FlagType       string
	Status         string
	ShareScope     instanceentity.ShareScope
	AccessURL      string
	ExpiresAt      time.Time
	ExtendCount    int
	MaxExtends     int
	CreatedAt      time.Time
}

type TeacherInstanceRow struct {
	ID              int64
	StudentID       int64
	StudentName     string
	StudentUsername string
	StudentNo       *string
	ClassName       string
	ChallengeID     int64
	ChallengeTitle  string
	Status          string
	AccessURL       string
	ExpiresAt       time.Time
	ExtendCount     int
	MaxExtends      int
	CreatedAt       time.Time
}

type ActiveAWDContestPause struct {
	ID            int64
	EndTime       time.Time
	PausedSeconds int64
}

type StartupRecoveryLockLease interface {
	Refresh(ctx context.Context, ttl time.Duration) (bool, error)
	Release(ctx context.Context) (bool, error)
	Key(ctx context.Context) string
}

type StartupRuntimeStateStore interface {
	LoadPlatformRuntimeState(ctx context.Context) (string, time.Time, bool, error)
	SavePlatformRuntimeState(ctx context.Context, bootID string, heartbeatAt time.Time) error
	AcquireStartupRecoveryLock(ctx context.Context, ttl time.Duration) (StartupRecoveryLockLease, bool, error)
}

type HostBootIDReader interface {
	ReadBootID(ctx context.Context) (string, error)
}

type ProxyTicketClaims struct {
	UserID               int64                     `json:"user_id"`
	Username             string                    `json:"username"`
	Role                 string                    `json:"role"`
	InstanceID           int64                     `json:"instance_id"`
	ContestID            *int64                    `json:"contest_id,omitempty"`
	ShareScope           instanceentity.ShareScope `json:"share_scope"`
	Purpose              string                    `json:"purpose,omitempty"`
	AWDAttackerTeamID    *int64                    `json:"awd_attacker_team_id,omitempty"`
	AWDVictimTeamID      *int64                    `json:"awd_victim_team_id,omitempty"`
	AWDServiceID         *int64                    `json:"awd_service_id,omitempty"`
	AWDChallengeID       *int64                    `json:"awd_challenge_id,omitempty"`
	AWDWorkspaceRevision *int64                    `json:"awd_workspace_revision,omitempty"`
	IssuedAt             time.Time                 `json:"issued_at"`
}

const (
	ProxyTicketPurposeInstanceAccess = "instance_access"
	ProxyTicketPurposeAWDAttack      = "awd_attack"
	ProxyTicketPurposeAWDDefenseSSH  = "awd_defense_ssh"
)

type AWDTargetProxyScope struct {
	InstanceID     int64
	ContestID      int64
	AttackerTeamID int64
	VictimTeamID   int64
	ServiceID      int64
	AWDChallengeID int64
	ShareScope     instanceentity.ShareScope
	Status         string
	AccessURL      string
	RuntimeDetails string
}

type AWDDefenseSSHScope struct {
	InstanceID        int64
	ContestID         int64
	TeamID            int64
	ServiceID         int64
	AWDChallengeID    int64
	WorkspaceRevision int64
	ContainerID       string
	ShareScope        instanceentity.ShareScope
	EditablePaths     []string `gorm:"-"`
	ProtectedPaths    []string `gorm:"-"`
}

type AWDProxyTrafficEventInput struct {
	ContestID      int64
	AttackerTeamID int64
	VictimTeamID   int64
	ServiceID      int64
	AWDChallengeID int64
	Method         string
	Path           string
	StatusCode     int
}

type AWDDefenseSSHSession struct {
	UserID            int64
	Username          string
	InstanceID        int64
	ContestID         int64
	TeamID            int64
	ServiceID         int64
	ChallengeID       int64
	WorkspaceRevision int64
	ContainerID       string
}

type ContainerDirectoryEntry struct {
	Name string
	Type string
	Size int64
}

type ProxyTicketStore interface {
	SaveProxyTicket(ctx context.Context, ticket string, claims ProxyTicketClaims, ttl time.Duration) error
	FindProxyTicket(ctx context.Context, ticket string) (*ProxyTicketClaims, error)
}

type ProxyTicketInstanceReader interface {
	FindByID(ctx context.Context, id int64) (*instanceentity.Instance, error)
	FindAWDTargetProxyScope(ctx context.Context, userID, contestID, serviceID, victimTeamID int64) (*AWDTargetProxyScope, error)
	FindAWDDefenseSSHScope(ctx context.Context, userID, contestID, serviceID int64) (*AWDDefenseSSHScope, error)
}

type ManagedContainer struct {
	ID        string
	Name      string
	CreatedAt time.Time
}

type ManagedContainerState struct {
	ID      string
	Exists  bool
	Running bool
	Status  string
}

type AWDDefenseWorkspace struct {
	ContainerID string
}

type AWDServiceOperation struct {
	ID            int64
	ContestID     int64
	TeamID        int64
	ServiceID     int64
	InstanceID    int64
	OperationType string
	RequestedBy   string
	RequestedByID *int64
	Reason        string
	SLABillable   bool
	Status        string
	ErrorMessage  string
	StartedAt     time.Time
	FinishedAt    *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

const (
	AWDServiceOperationTypeRecover  = "recover"
	AWDServiceOperationTypeRecreate = "recreate"

	AWDServiceOperationRequestedBySystem = "system"

	AWDServiceOperationStatusProvisioning = "provisioning"
	AWDServiceOperationStatusRecovering   = "recovering"
	AWDServiceOperationStatusRecovered    = "recovered"
	AWDServiceOperationStatusFailed       = "failed"
)
