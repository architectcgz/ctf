package runtime

import (
	"context"
	"net/http"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	runtimeapp "ctf-platform/internal/module/runtime/application"
	runtimeinfra "ctf-platform/internal/module/runtimeinfra"
)

type ManagedContainer = runtimeinfra.ManagedContainer
type ManagedContainerStat = runtimeinfra.ManagedContainerStat
type ProxyTicketClaims = runtimeapp.ProxyTicketClaims

type ProxyCookieConfig struct {
	Secure   bool
	SameSite http.SameSite
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
