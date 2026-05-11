package composition

import (
	"context"
	"fmt"
	"time"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	practiceports "ctf-platform/internal/module/practice/ports"
	runtimeports "ctf-platform/internal/module/runtime/ports"
	"ctf-platform/pkg/errcode"
)

type runtimeHTTPServiceAdapter struct {
	commandService       instancecontracts.InstanceCommandService
	queryService         instancecontracts.InstanceQueryService
	proxyTickets         instancecontracts.ProxyTicketService
	defenseWorkbench     instancecontracts.AWDDefenseWorkbenchService
	proxyBodyPreviewSize int
	proxyTicketMaxAge    int
	defenseSSHEnabled    bool
	defenseSSHHost       string
	defenseSSHPort       int
}

func newRuntimeHTTPServiceAdapter(
	commandService instancecontracts.InstanceCommandService,
	queryService instancecontracts.InstanceQueryService,
	proxyTickets instancecontracts.ProxyTicketService,
	defenseWorkbench instancecontracts.AWDDefenseWorkbenchService,
	proxyBodyPreviewSize int,
	proxyTicketMaxAge int,
	defenseSSHEnabled bool,
	defenseSSHHost string,
	defenseSSHPort int,
) *runtimeHTTPServiceAdapter {
	return &runtimeHTTPServiceAdapter{
		commandService:       commandService,
		queryService:         queryService,
		proxyTickets:         proxyTickets,
		defenseWorkbench:     defenseWorkbench,
		proxyBodyPreviewSize: proxyBodyPreviewSize,
		proxyTicketMaxAge:    proxyTicketMaxAge,
		defenseSSHEnabled:    defenseSSHEnabled,
		defenseSSHHost:       defenseSSHHost,
		defenseSSHPort:       defenseSSHPort,
	}
}

func (a *runtimeHTTPServiceAdapter) DestroyInstance(ctx context.Context, instanceID, userID int64) error {
	if a == nil || a.commandService == nil {
		return errRuntimeHTTPInstanceServiceUnavailable()
	}
	return a.commandService.DestroyInstance(ctx, instanceID, userID)
}

func (a *runtimeHTTPServiceAdapter) ExtendInstance(ctx context.Context, instanceID, userID int64) (*dto.InstanceResp, error) {
	if a == nil || a.commandService == nil {
		return nil, errRuntimeHTTPInstanceServiceUnavailable()
	}
	return a.commandService.ExtendInstance(ctx, instanceID, userID)
}

func (a *runtimeHTTPServiceAdapter) GetAccessURL(ctx context.Context, instanceID, userID int64) (string, error) {
	if a == nil || a.queryService == nil {
		return "", errRuntimeHTTPInstanceServiceUnavailable()
	}
	return a.queryService.GetAccessURL(ctx, instanceID, userID)
}

func (a *runtimeHTTPServiceAdapter) GetUserInstances(ctx context.Context, userID int64) ([]*dto.InstanceInfo, error) {
	if a == nil || a.queryService == nil {
		return nil, errRuntimeHTTPInstanceServiceUnavailable()
	}
	return a.queryService.GetUserInstances(ctx, userID)
}

func (a *runtimeHTTPServiceAdapter) ListTeacherInstances(ctx context.Context, requesterID int64, requesterRole string, query *dto.TeacherInstanceQuery) ([]dto.TeacherInstanceItem, error) {
	if a == nil || a.queryService == nil {
		return nil, errRuntimeHTTPInstanceServiceUnavailable()
	}
	return a.queryService.ListTeacherInstances(ctx, requesterID, requesterRole, query)
}

func (a *runtimeHTTPServiceAdapter) DestroyTeacherInstance(ctx context.Context, instanceID, requesterID int64, requesterRole string) error {
	if a == nil || a.commandService == nil {
		return errRuntimeHTTPInstanceServiceUnavailable()
	}
	return a.commandService.DestroyTeacherInstance(ctx, instanceID, requesterID, requesterRole)
}

func (a *runtimeHTTPServiceAdapter) IssueProxyTicket(ctx context.Context, user authctx.CurrentUser, instanceID int64) (string, error) {
	if a == nil || a.proxyTickets == nil {
		return "", errRuntimeHTTPProxyTicketServiceUnavailable()
	}

	ticket, _, err := a.proxyTickets.IssueTicket(ctx, user, instanceID)
	return ticket, err
}

func (a *runtimeHTTPServiceAdapter) IssueAWDTargetProxyTicket(ctx context.Context, user authctx.CurrentUser, contestID, serviceID, victimTeamID int64) (string, error) {
	if a == nil || a.proxyTickets == nil {
		return "", errRuntimeHTTPProxyTicketServiceUnavailable()
	}

	ticket, _, err := a.proxyTickets.IssueAWDTargetTicket(ctx, user, contestID, serviceID, victimTeamID)
	return ticket, err
}

func (a *runtimeHTTPServiceAdapter) IssueAWDDefenseSSHTicket(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64) (*dto.AWDDefenseSSHAccessResp, error) {
	if a == nil || a.proxyTickets == nil {
		return nil, errRuntimeHTTPProxyTicketServiceUnavailable()
	}
	if !a.defenseSSHEnabled || a.defenseSSHHost == "" || a.defenseSSHPort <= 0 {
		return nil, errcode.ErrAWDDefenseSSHUnavailable.WithCause(fmt.Errorf("awd defense ssh gateway is not enabled"))
	}

	ticket, expiresAt, err := a.proxyTickets.IssueAWDDefenseSSHTicket(ctx, user, contestID, serviceID)
	if err != nil {
		return nil, err
	}
	username := fmt.Sprintf("%s+%d+%d", user.Username, contestID, serviceID)
	return &dto.AWDDefenseSSHAccessResp{
		Host:      a.defenseSSHHost,
		Port:      a.defenseSSHPort,
		Username:  username,
		Password:  ticket,
		Command:   fmt.Sprintf("ssh %s@%s -p %d", username, a.defenseSSHHost, a.defenseSSHPort),
		ExpiresAt: expiresAt.Format(time.RFC3339),
	}, nil
}

func (a *runtimeHTTPServiceAdapter) ReadAWDDefenseFile(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64, filePath string) (*dto.AWDDefenseFileResp, error) {
	if a == nil || a.defenseWorkbench == nil {
		return nil, errRuntimeHTTPInstanceServiceUnavailable()
	}
	return a.defenseWorkbench.ReadAWDDefenseFile(ctx, user, contestID, serviceID, filePath)
}

func (a *runtimeHTTPServiceAdapter) ListAWDDefenseDirectory(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64, dirPath string) (*dto.AWDDefenseDirectoryResp, error) {
	if a == nil || a.defenseWorkbench == nil {
		return nil, errRuntimeHTTPInstanceServiceUnavailable()
	}
	return a.defenseWorkbench.ListAWDDefenseDirectory(ctx, user, contestID, serviceID, dirPath)
}

func (a *runtimeHTTPServiceAdapter) SaveAWDDefenseFile(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64, req dto.AWDDefenseFileSaveReq) (*dto.AWDDefenseFileSaveResp, error) {
	if a == nil || a.defenseWorkbench == nil {
		return nil, errRuntimeHTTPInstanceServiceUnavailable()
	}
	return a.defenseWorkbench.SaveAWDDefenseFile(ctx, user, contestID, serviceID, req)
}

func (a *runtimeHTTPServiceAdapter) RunAWDDefenseCommand(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64, req dto.AWDDefenseCommandReq) (*dto.AWDDefenseCommandResp, error) {
	if a == nil || a.defenseWorkbench == nil {
		return nil, errRuntimeHTTPInstanceServiceUnavailable()
	}
	return a.defenseWorkbench.RunAWDDefenseCommand(ctx, user, contestID, serviceID, req)
}

func errRuntimeHTTPProxyTicketServiceUnavailable() error {
	return errcode.ErrInternal.WithCause(fmt.Errorf("proxy ticket service is not configured"))
}

func (a *runtimeHTTPServiceAdapter) ResolveProxyTicket(ctx context.Context, ticket string) (*runtimeports.ProxyTicketClaims, error) {
	if a == nil || a.proxyTickets == nil {
		return nil, errRuntimeHTTPProxyTicketServiceUnavailable()
	}
	return a.proxyTickets.ResolveTicket(ctx, ticket)
}

func (a *runtimeHTTPServiceAdapter) ResolveAWDTargetAccessURL(ctx context.Context, claims *runtimeports.ProxyTicketClaims, contestID, serviceID, victimTeamID int64) (string, error) {
	if a == nil || a.proxyTickets == nil {
		return "", errRuntimeHTTPProxyTicketServiceUnavailable()
	}
	return a.proxyTickets.ResolveAWDTargetAccessURL(ctx, claims, contestID, serviceID, victimTeamID)
}

func (a *runtimeHTTPServiceAdapter) ProxyTicketMaxAge() int {
	if a == nil {
		return 0
	}
	return a.proxyTicketMaxAge
}

func (a *runtimeHTTPServiceAdapter) ProxyBodyPreviewSize() int {
	if a == nil {
		return 0
	}
	return a.proxyBodyPreviewSize
}

func errRuntimeHTTPInstanceServiceUnavailable() error {
	return errcode.ErrInternal.WithCause(fmt.Errorf("instance application service is not configured"))
}

func toRuntimeTopologyCreateRequest(req *practiceports.TopologyCreateRequest) *runtimeports.TopologyCreateRequest {
	if req == nil {
		return nil
	}
	networks := make([]runtimeports.TopologyCreateNetwork, 0, len(req.Networks))
	for _, network := range req.Networks {
		networks = append(networks, runtimeports.TopologyCreateNetwork{
			Key:      network.Key,
			Name:     network.Name,
			Internal: network.Internal,
			Shared:   network.Shared,
		})
	}
	nodes := make([]runtimeports.TopologyCreateNode, 0, len(req.Nodes))
	for _, node := range req.Nodes {
		nodes = append(nodes, runtimeports.TopologyCreateNode{
			Key:             node.Key,
			Image:           node.Image,
			Env:             cloneStringMap(node.Env),
			ServicePort:     node.ServicePort,
			ServiceProtocol: node.ServiceProtocol,
			IsEntryPoint:    node.IsEntryPoint,
			NetworkKeys:     append([]string(nil), node.NetworkKeys...),
			NetworkAliases:  append([]string(nil), node.NetworkAliases...),
			Resources:       cloneResourceLimits(node.Resources),
		})
	}
	return &runtimeports.TopologyCreateRequest{
		Networks:                   networks,
		Nodes:                      nodes,
		Policies:                   append([]model.TopologyTrafficPolicy(nil), req.Policies...),
		ReservedHostPort:           req.ReservedHostPort,
		DisableEntryPortPublishing: req.DisableEntryPortPublishing,
	}
}

func cloneStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	output := make(map[string]string, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}

func cloneResourceLimits(input *model.ResourceLimits) *model.ResourceLimits {
	if input == nil {
		return nil
	}
	return &model.ResourceLimits{
		CPUQuota:  input.CPUQuota,
		Memory:    input.Memory,
		PidsLimit: input.PidsLimit,
	}
}
