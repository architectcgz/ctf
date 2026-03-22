package runtime

import (
	"context"
	"fmt"
	"time"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	runtimeapp "ctf-platform/internal/module/runtime/application"
	"ctf-platform/pkg/errcode"
)

type Module struct {
	*Service
	instanceRepository   instanceRepository
	instanceService      instanceHTTPService
	proxyTickets         proxyTicketService
	proxyBodyPreviewSize int
}

type instanceRepository interface {
	UpdateRuntime(instance *model.Instance) error
	UpdateStatusAndReleasePort(id int64, status string) error
	FindByUserAndChallenge(userID, challengeID int64) (*model.Instance, error)
}

type instanceHTTPService interface {
	DestroyInstanceWithContext(ctx context.Context, instanceID, userID int64) error
	ExtendInstanceWithContext(ctx context.Context, instanceID, userID int64) (*dto.InstanceResp, error)
	GetAccessURLWithContext(ctx context.Context, instanceID, userID int64) (string, error)
	GetUserInstancesWithContext(ctx context.Context, userID int64) ([]*dto.InstanceInfo, error)
	ListTeacherInstances(ctx context.Context, requesterID int64, requesterRole string, query *dto.TeacherInstanceQuery) ([]dto.TeacherInstanceItem, error)
	DestroyTeacherInstance(ctx context.Context, instanceID, requesterID int64, requesterRole string) error
}

type proxyTicketService interface {
	IssueTicket(ctx context.Context, user authctx.CurrentUser, instanceID int64) (string, time.Time, error)
	ResolveTicket(ctx context.Context, ticket string) (*runtimeapp.ProxyTicketClaims, error)
	MaxAge() int
}

func NewModule(service *Service, instanceRepository instanceRepository, instanceService instanceHTTPService, proxyTickets proxyTicketService, proxyBodyPreviewSize int) *Module {
	return &Module{
		Service:              service,
		instanceRepository:   instanceRepository,
		instanceService:      instanceService,
		proxyTickets:         proxyTickets,
		proxyBodyPreviewSize: proxyBodyPreviewSize,
	}
}

func (m *Module) DestroyInstanceWithContext(ctx context.Context, instanceID, userID int64) error {
	if m == nil || m.instanceService == nil {
		return errInstanceServiceUnavailable()
	}
	return m.instanceService.DestroyInstanceWithContext(ctx, instanceID, userID)
}

func (m *Module) ExtendInstanceWithContext(ctx context.Context, instanceID, userID int64) (*dto.InstanceResp, error) {
	if m == nil || m.instanceService == nil {
		return nil, errInstanceServiceUnavailable()
	}
	return m.instanceService.ExtendInstanceWithContext(ctx, instanceID, userID)
}

func (m *Module) GetAccessURLWithContext(ctx context.Context, instanceID, userID int64) (string, error) {
	if m == nil || m.instanceService == nil {
		return "", errInstanceServiceUnavailable()
	}
	return m.instanceService.GetAccessURLWithContext(ctx, instanceID, userID)
}

func (m *Module) GetUserInstancesWithContext(ctx context.Context, userID int64) ([]*dto.InstanceInfo, error) {
	if m == nil || m.instanceService == nil {
		return nil, errInstanceServiceUnavailable()
	}
	return m.instanceService.GetUserInstancesWithContext(ctx, userID)
}

func (m *Module) ListTeacherInstances(ctx context.Context, requesterID int64, requesterRole string, query *dto.TeacherInstanceQuery) ([]dto.TeacherInstanceItem, error) {
	if m == nil || m.instanceService == nil {
		return nil, errInstanceServiceUnavailable()
	}
	return m.instanceService.ListTeacherInstances(ctx, requesterID, requesterRole, query)
}

func (m *Module) DestroyTeacherInstance(ctx context.Context, instanceID, requesterID int64, requesterRole string) error {
	if m == nil || m.instanceService == nil {
		return errInstanceServiceUnavailable()
	}
	return m.instanceService.DestroyTeacherInstance(ctx, instanceID, requesterID, requesterRole)
}

func (m *Module) UpdateRuntime(instance *model.Instance) error {
	if m == nil || m.instanceRepository == nil {
		return errInstanceRepositoryUnavailable()
	}
	return m.instanceRepository.UpdateRuntime(instance)
}

func (m *Module) UpdateStatusAndReleasePort(id int64, status string) error {
	if m == nil || m.instanceRepository == nil {
		return errInstanceRepositoryUnavailable()
	}
	return m.instanceRepository.UpdateStatusAndReleasePort(id, status)
}

func (m *Module) FindByUserAndChallenge(userID, challengeID int64) (*model.Instance, error) {
	if m == nil || m.instanceRepository == nil {
		return nil, errInstanceRepositoryUnavailable()
	}
	return m.instanceRepository.FindByUserAndChallenge(userID, challengeID)
}

func (m *Module) IssueProxyTicket(ctx context.Context, user authctx.CurrentUser, instanceID int64) (string, error) {
	if m == nil || m.proxyTickets == nil {
		return "", errProxyTicketServiceUnavailable()
	}

	ticket, _, err := m.proxyTickets.IssueTicket(ctx, user, instanceID)
	return ticket, err
}

func (m *Module) ResolveProxyTicket(ctx context.Context, ticket string) (*runtimeapp.ProxyTicketClaims, error) {
	if m == nil || m.proxyTickets == nil {
		return nil, errProxyTicketServiceUnavailable()
	}
	return m.proxyTickets.ResolveTicket(ctx, ticket)
}

func (m *Module) ProxyTicketMaxAge() int {
	if m == nil || m.proxyTickets == nil {
		return 0
	}
	return m.proxyTickets.MaxAge()
}

func (m *Module) ProxyBodyPreviewSize() int {
	if m == nil {
		return 0
	}
	return m.proxyBodyPreviewSize
}

func errProxyTicketServiceUnavailable() error {
	return errcode.ErrInternal.WithCause(fmt.Errorf("proxy ticket service is not configured"))
}

func errInstanceServiceUnavailable() error {
	return errcode.ErrInternal.WithCause(fmt.Errorf("instance application service is not configured"))
}

func errInstanceRepositoryUnavailable() error {
	return errcode.ErrInternal.WithCause(fmt.Errorf("instance repository contract is not configured"))
}
