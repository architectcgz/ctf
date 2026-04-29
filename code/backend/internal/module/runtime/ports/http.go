package ports

import (
	"context"
	"time"

	"ctf-platform/internal/model"
)

type CountRunningRepository interface {
	CountRunning(ctx context.Context) (int64, error)
}

type InstanceRepository interface {
	FindByID(ctx context.Context, id int64) (*model.Instance, error)
	FindUserByID(ctx context.Context, userID int64) (*model.User, error)
	FindAccessibleByIDForUser(ctx context.Context, instanceID, userID int64) (*model.Instance, error)
	ListVisibleByUser(ctx context.Context, userID int64) ([]UserVisibleInstanceRow, error)
	ListTeacherInstances(ctx context.Context, filter TeacherInstanceFilter) ([]TeacherInstanceRow, error)
	AtomicExtendByID(ctx context.Context, id int64, maxExtends int, duration time.Duration) error
	UpdateStatusAndReleasePort(ctx context.Context, id int64, status string) error
}

type RuntimeCleaner interface {
	CleanupRuntime(ctx context.Context, instance *model.Instance) error
}

type TeacherInstanceFilter struct {
	ClassName string
	Keyword   string
	StudentNo string
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
	ShareScope     model.ShareScope
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

type ProxyTicketClaims struct {
	UserID            int64            `json:"user_id"`
	Username          string           `json:"username"`
	Role              string           `json:"role"`
	InstanceID        int64            `json:"instance_id"`
	ContestID         *int64           `json:"contest_id,omitempty"`
	ShareScope        model.ShareScope `json:"share_scope"`
	Purpose           string           `json:"purpose,omitempty"`
	AWDAttackerTeamID *int64           `json:"awd_attacker_team_id,omitempty"`
	AWDVictimTeamID   *int64           `json:"awd_victim_team_id,omitempty"`
	AWDServiceID      *int64           `json:"awd_service_id,omitempty"`
	AWDChallengeID    *int64           `json:"awd_challenge_id,omitempty"`
	IssuedAt          time.Time        `json:"issued_at"`
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
	ShareScope     model.ShareScope
	AccessURL      string
}

type AWDDefenseSSHScope struct {
	InstanceID     int64
	ContestID      int64
	TeamID         int64
	ServiceID      int64
	AWDChallengeID int64
	ContainerID    string
	ShareScope     model.ShareScope
}

type AWDDefenseSSHSession struct {
	UserID      int64
	Username    string
	InstanceID  int64
	ContestID   int64
	TeamID      int64
	ServiceID   int64
	ChallengeID int64
	ContainerID string
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
	FindByID(ctx context.Context, id int64) (*model.Instance, error)
	FindAWDTargetProxyScope(ctx context.Context, userID, contestID, serviceID, victimTeamID int64) (*AWDTargetProxyScope, error)
	FindAWDDefenseSSHScope(ctx context.Context, userID, contestID, serviceID int64) (*AWDDefenseSSHScope, error)
}

type ProxyTrafficEventRecorder interface {
	RecordRuntimeProxyTrafficEvent(ctx context.Context, instanceID, userID int64, method, requestPath string, statusCode int) error
	RecordAWDProxyTrafficEvent(ctx context.Context, event model.AWDProxyTrafficEventInput) error
}
