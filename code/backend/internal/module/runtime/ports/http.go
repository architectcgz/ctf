package ports

import (
	"context"
	"time"

	"ctf-platform/internal/model"
)

type CountRunningRepository interface {
	CountRunningWithContext(ctx context.Context) (int64, error)
}

type InstanceRepository interface {
	FindByID(ctx context.Context, id int64) (*model.Instance, error)
	FindUserByID(ctx context.Context, userID int64) (*model.User, error)
	FindAccessibleByIDForUser(ctx context.Context, instanceID, userID int64) (*model.Instance, error)
	ListVisibleByUser(ctx context.Context, userID int64) ([]UserVisibleInstanceRow, error)
	ListTeacherInstances(ctx context.Context, filter TeacherInstanceFilter) ([]TeacherInstanceRow, error)
	AtomicExtendByIDWithContext(ctx context.Context, id int64, maxExtends int, duration time.Duration) error
	UpdateStatusAndReleasePortWithContext(ctx context.Context, id int64, status string) error
}

type RuntimeCleaner interface {
	CleanupRuntimeWithContext(ctx context.Context, instance *model.Instance) error
}

type TeacherInstanceFilter struct {
	ClassName string
	Keyword   string
	StudentNo string
}

type UserVisibleInstanceRow struct {
	ID             int64
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
	UserID     int64            `json:"user_id"`
	Username   string           `json:"username"`
	Role       string           `json:"role"`
	InstanceID int64            `json:"instance_id"`
	ContestID  *int64           `json:"contest_id,omitempty"`
	ShareScope model.ShareScope `json:"share_scope"`
	IssuedAt   time.Time        `json:"issued_at"`
}

type ProxyTicketStore interface {
	SaveProxyTicket(ctx context.Context, ticket string, claims ProxyTicketClaims, ttl time.Duration) error
	FindProxyTicket(ctx context.Context, ticket string) (*ProxyTicketClaims, error)
}

type ProxyTicketInstanceReader interface {
	FindByID(ctx context.Context, id int64) (*model.Instance, error)
}

type ProxyTrafficEventRecorder interface {
	RecordRuntimeProxyTrafficEvent(ctx context.Context, instanceID, userID int64, method, requestPath string, statusCode int) error
}
