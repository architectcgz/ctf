package runtime

import (
	"context"
	"net/http"
	"time"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	runtimeinfra "ctf-platform/internal/module/runtimeinfra"
)

type ManagedContainer = runtimeinfra.ManagedContainer
type ManagedContainerStat = runtimeinfra.ManagedContainerStat

type ProxyCookieConfig struct {
	Secure   bool
	SameSite http.SameSite
}

type RuntimeQuery interface {
	CountRunning() (int64, error)
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

type RuntimeStatsProvider interface {
	ListManagedContainerStats(ctx context.Context) ([]ManagedContainerStat, error)
}

type RuntimeHTTPService interface {
	DestroyInstanceWithContext(ctx context.Context, instanceID, userID int64) error
	ExtendInstanceWithContext(ctx context.Context, instanceID, userID int64) (*dto.InstanceResp, error)
	GetAccessURLWithContext(ctx context.Context, instanceID, userID int64) (string, error)
	GetUserInstancesWithContext(ctx context.Context, userID int64) ([]*dto.InstanceInfo, error)
	ListTeacherInstances(ctx context.Context, requesterID int64, requesterRole string, query *dto.TeacherInstanceQuery) ([]dto.TeacherInstanceItem, error)
	DestroyTeacherInstance(ctx context.Context, instanceID, requesterID int64, requesterRole string) error
	IssueProxyTicket(ctx context.Context, user authctx.CurrentUser, instanceID int64) (string, error)
	ResolveProxyTicket(ctx context.Context, ticket string) (*ProxyTicketClaims, error)
	ProxyTicketMaxAge() int
	ProxyBodyPreviewSize() int
}

type InstanceRepository interface {
	UpdateRuntime(instance *model.Instance) error
	UpdateStatusAndReleasePort(id int64, status string) error
	FindByUserAndChallenge(userID, challengeID int64) (*model.Instance, error)
}

type RuntimeFacade interface {
	RuntimeStatsProvider
	InspectImageSize(ctx context.Context, imageRef string) (int64, error)
	RemoveImage(ctx context.Context, imageRef string) error
	CleanupRuntime(instance *model.Instance) error
	CreateTopology(ctx context.Context, req *TopologyCreateRequest) (*TopologyCreateResult, error)
	CreateContainer(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (containerID, networkID string, hostPort, servicePort int, err error)
	CleanExpiredInstances(ctx context.Context) error
	CleanupOrphans(ctx context.Context) error
	WriteFileToContainer(ctx context.Context, containerID, filePath string, content []byte) error
}
