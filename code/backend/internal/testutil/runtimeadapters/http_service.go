package runtimeadapters

import (
	"context"
	"fmt"
	"time"

	"ctf-platform/internal/authctx"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	instanceports "ctf-platform/internal/module/instance/ports"
	runtimehttp "ctf-platform/internal/module/runtime/api/http"
)

type httpInstanceCommandService interface {
	DestroyInstance(ctx context.Context, instanceID, userID int64) error
	ExtendInstance(ctx context.Context, instanceID, userID int64) (*instancecontracts.InstanceResp, error)
	DestroyTeacherInstance(ctx context.Context, instanceID, requesterID int64, requesterRole string) error
}

type httpInstanceQueryService interface {
	GetAccessURL(ctx context.Context, instanceID, userID int64) (string, error)
	GetUserInstances(ctx context.Context, userID int64) ([]*instancecontracts.InstanceInfo, error)
	ListTeacherInstances(ctx context.Context, requesterID int64, requesterRole string, query instancecontracts.TeacherInstanceListQuery) (*instancecontracts.TeacherInstancePageResult, error)
}

type httpProxyTicketService interface {
	IssueTicket(ctx context.Context, user authctx.CurrentUser, instanceID int64) (string, time.Time, error)
	IssueAWDTargetTicket(ctx context.Context, user authctx.CurrentUser, contestID, serviceID, victimTeamID int64) (string, time.Time, error)
	IssueAWDDefenseSSHTicket(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64) (string, time.Time, error)
	ResolveTicket(ctx context.Context, ticket string) (*instanceports.ProxyTicketClaims, error)
	ResolveAWDTargetAccessURL(ctx context.Context, claims *instanceports.ProxyTicketClaims, contestID, serviceID, victimTeamID int64) (string, error)
	MaxAge() int
}

// HTTPService 为测试提供 runtime HTTP handler 所需的 facade。
type HTTPService struct {
	commandService       httpInstanceCommandService
	queryService         httpInstanceQueryService
	proxyTickets         httpProxyTicketService
	proxyBodyPreviewSize int
}

// NewHTTPService 创建 runtime HTTP 测试 facade。
func NewHTTPService(commandService httpInstanceCommandService, queryService httpInstanceQueryService, proxyTickets httpProxyTicketService, proxyBodyPreviewSize int) *HTTPService {
	return &HTTPService{
		commandService:       commandService,
		queryService:         queryService,
		proxyTickets:         proxyTickets,
		proxyBodyPreviewSize: proxyBodyPreviewSize,
	}
}

func (a *HTTPService) DestroyInstance(ctx context.Context, instanceID, userID int64) error {
	return a.commandService.DestroyInstance(ctx, instanceID, userID)
}

func (a *HTTPService) ExtendInstance(ctx context.Context, instanceID, userID int64) (*instancecontracts.InstanceResp, error) {
	return a.commandService.ExtendInstance(ctx, instanceID, userID)
}

func (a *HTTPService) GetAccessURL(ctx context.Context, instanceID, userID int64) (string, error) {
	return a.queryService.GetAccessURL(ctx, instanceID, userID)
}

func (a *HTTPService) GetUserInstances(ctx context.Context, userID int64) ([]*instancecontracts.InstanceInfo, error) {
	return a.queryService.GetUserInstances(ctx, userID)
}

func (a *HTTPService) ListTeacherInstances(ctx context.Context, requesterID int64, requesterRole string, query instancecontracts.TeacherInstanceListQuery) (*instancecontracts.TeacherInstancePageResult, error) {
	return a.queryService.ListTeacherInstances(ctx, requesterID, requesterRole, query)
}

func (a *HTTPService) DestroyTeacherInstance(ctx context.Context, instanceID, requesterID int64, requesterRole string) error {
	return a.commandService.DestroyTeacherInstance(ctx, instanceID, requesterID, requesterRole)
}

func (a *HTTPService) IssueProxyTicket(ctx context.Context, user authctx.CurrentUser, instanceID int64) (string, error) {
	ticket, _, err := a.proxyTickets.IssueTicket(ctx, user, instanceID)
	return ticket, err
}

func (a *HTTPService) IssueAWDTargetProxyTicket(ctx context.Context, user authctx.CurrentUser, contestID, serviceID, victimTeamID int64) (string, error) {
	ticket, _, err := a.proxyTickets.IssueAWDTargetTicket(ctx, user, contestID, serviceID, victimTeamID)
	return ticket, err
}

func (a *HTTPService) IssueAWDDefenseSSHTicket(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64) (*runtimehttp.AWDDefenseSSHAccessResp, error) {
	ticket, expiresAt, err := a.proxyTickets.IssueAWDDefenseSSHTicket(ctx, user, contestID, serviceID)
	if err != nil {
		return nil, err
	}
	username := fmt.Sprintf("%s+%d+%d", user.Username, contestID, serviceID)
	return &runtimehttp.AWDDefenseSSHAccessResp{
		Host:      "127.0.0.1",
		Port:      2222,
		Username:  username,
		Password:  ticket,
		Command:   fmt.Sprintf("ssh %s@127.0.0.1 -p 2222", username),
		ExpiresAt: expiresAt.Format(time.RFC3339),
	}, nil
}

func (a *HTTPService) ReadAWDDefenseFile(context.Context, authctx.CurrentUser, int64, int64, string) (*instancecontracts.AWDDefenseFileResp, error) {
	return &instancecontracts.AWDDefenseFileResp{}, nil
}

func (a *HTTPService) ListAWDDefenseDirectory(context.Context, authctx.CurrentUser, int64, int64, string) (*instancecontracts.AWDDefenseDirectoryResp, error) {
	return &instancecontracts.AWDDefenseDirectoryResp{}, nil
}

func (a *HTTPService) SaveAWDDefenseFile(_ context.Context, _ authctx.CurrentUser, _ int64, _ int64, req instancecontracts.AWDDefenseFileSaveReq) (*instancecontracts.AWDDefenseFileSaveResp, error) {
	return &instancecontracts.AWDDefenseFileSaveResp{
		Path: req.Path,
		Size: len(req.Content),
	}, nil
}

func (a *HTTPService) RunAWDDefenseCommand(_ context.Context, _ authctx.CurrentUser, _ int64, _ int64, req instancecontracts.AWDDefenseCommandReq) (*instancecontracts.AWDDefenseCommandResp, error) {
	return &instancecontracts.AWDDefenseCommandResp{
		Command: req.Command,
	}, nil
}

func (a *HTTPService) ResolveProxyTicket(ctx context.Context, ticket string) (*instanceports.ProxyTicketClaims, error) {
	return a.proxyTickets.ResolveTicket(ctx, ticket)
}

func (a *HTTPService) ResolveAWDTargetAccessURL(ctx context.Context, claims *instanceports.ProxyTicketClaims, contestID, serviceID, victimTeamID int64) (string, error) {
	return a.proxyTickets.ResolveAWDTargetAccessURL(ctx, claims, contestID, serviceID, victimTeamID)
}

func (a *HTTPService) ProxyTicketMaxAge() int {
	return a.proxyTickets.MaxAge()
}

func (a *HTTPService) ProxyBodyPreviewSize() int {
	return a.proxyBodyPreviewSize
}
