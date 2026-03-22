package runtimeadapters

import (
	"context"
	"time"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	runtimeapp "ctf-platform/internal/module/runtime/application"
)

type httpInstanceService interface {
	DestroyInstanceWithContext(ctx context.Context, instanceID, userID int64) error
	ExtendInstanceWithContext(ctx context.Context, instanceID, userID int64) (*dto.InstanceResp, error)
	GetAccessURLWithContext(ctx context.Context, instanceID, userID int64) (string, error)
	GetUserInstancesWithContext(ctx context.Context, userID int64) ([]*dto.InstanceInfo, error)
	ListTeacherInstances(ctx context.Context, requesterID int64, requesterRole string, query *dto.TeacherInstanceQuery) ([]dto.TeacherInstanceItem, error)
	DestroyTeacherInstance(ctx context.Context, instanceID, requesterID int64, requesterRole string) error
}

type httpProxyTicketService interface {
	IssueTicket(ctx context.Context, user authctx.CurrentUser, instanceID int64) (string, time.Time, error)
	ResolveTicket(ctx context.Context, ticket string) (*runtimeapp.ProxyTicketClaims, error)
	MaxAge() int
}

// HTTPService 为测试提供 runtime HTTP handler 所需的 facade。
type HTTPService struct {
	instanceService      httpInstanceService
	proxyTickets         httpProxyTicketService
	proxyBodyPreviewSize int
}

// NewHTTPService 创建 runtime HTTP 测试 facade。
func NewHTTPService(instanceService httpInstanceService, proxyTickets httpProxyTicketService, proxyBodyPreviewSize int) *HTTPService {
	return &HTTPService{
		instanceService:      instanceService,
		proxyTickets:         proxyTickets,
		proxyBodyPreviewSize: proxyBodyPreviewSize,
	}
}

func (a *HTTPService) DestroyInstanceWithContext(ctx context.Context, instanceID, userID int64) error {
	return a.instanceService.DestroyInstanceWithContext(ctx, instanceID, userID)
}

func (a *HTTPService) ExtendInstanceWithContext(ctx context.Context, instanceID, userID int64) (*dto.InstanceResp, error) {
	return a.instanceService.ExtendInstanceWithContext(ctx, instanceID, userID)
}

func (a *HTTPService) GetAccessURLWithContext(ctx context.Context, instanceID, userID int64) (string, error) {
	return a.instanceService.GetAccessURLWithContext(ctx, instanceID, userID)
}

func (a *HTTPService) GetUserInstancesWithContext(ctx context.Context, userID int64) ([]*dto.InstanceInfo, error) {
	return a.instanceService.GetUserInstancesWithContext(ctx, userID)
}

func (a *HTTPService) ListTeacherInstances(ctx context.Context, requesterID int64, requesterRole string, query *dto.TeacherInstanceQuery) ([]dto.TeacherInstanceItem, error) {
	return a.instanceService.ListTeacherInstances(ctx, requesterID, requesterRole, query)
}

func (a *HTTPService) DestroyTeacherInstance(ctx context.Context, instanceID, requesterID int64, requesterRole string) error {
	return a.instanceService.DestroyTeacherInstance(ctx, instanceID, requesterID, requesterRole)
}

func (a *HTTPService) IssueProxyTicket(ctx context.Context, user authctx.CurrentUser, instanceID int64) (string, error) {
	ticket, _, err := a.proxyTickets.IssueTicket(ctx, user, instanceID)
	return ticket, err
}

func (a *HTTPService) ResolveProxyTicket(ctx context.Context, ticket string) (*runtimeapp.ProxyTicketClaims, error) {
	return a.proxyTickets.ResolveTicket(ctx, ticket)
}

func (a *HTTPService) ProxyTicketMaxAge() int {
	return a.proxyTickets.MaxAge()
}

func (a *HTTPService) ProxyBodyPreviewSize() int {
	return a.proxyBodyPreviewSize
}
