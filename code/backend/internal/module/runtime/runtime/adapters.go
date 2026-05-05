package runtime

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
	opsports "ctf-platform/internal/module/ops/ports"
	practiceports "ctf-platform/internal/module/practice/ports"
	runtimeapp "ctf-platform/internal/module/runtime/application"
	runtimecmd "ctf-platform/internal/module/runtime/application/commands"
	runtimeports "ctf-platform/internal/module/runtime/ports"
	"ctf-platform/pkg/errcode"
)

type runtimeOpsStatsProviderAdapter struct {
	service *runtimeapp.ContainerStatsService
}

func newRuntimeOpsStatsProvider(service *runtimeapp.ContainerStatsService) opsports.RuntimeStatsProvider {
	return &runtimeOpsStatsProviderAdapter{service: service}
}

func (p *runtimeOpsStatsProviderAdapter) ListManagedContainerStats(ctx context.Context) ([]opsports.ManagedContainerStat, error) {
	if p == nil || p.service == nil {
		return []opsports.ManagedContainerStat{}, nil
	}

	stats, err := p.service.ListManagedContainerStats(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]opsports.ManagedContainerStat, 0, len(stats))
	for _, item := range stats {
		result = append(result, opsports.ManagedContainerStat{
			ContainerID:   item.ContainerID,
			ContainerName: item.ContainerName,
			CPUPercent:    item.CPUPercent,
			MemoryPercent: item.MemoryPercent,
			MemoryUsage:   item.MemoryUsage,
			MemoryLimit:   item.MemoryLimit,
		})
	}
	return result, nil
}

type runtimeHTTPServiceAdapter struct {
	commandService       runtimeHTTPCommandService
	queryService         runtimeHTTPQueryService
	proxyTickets         runtimeHTTPProxyTicketService
	proxyTicketReader    runtimeports.ProxyTicketInstanceReader
	defenseWorkbench     runtimeDefenseWorkbenchRuntime
	proxyBodyPreviewSize int
	defenseSSHEnabled    bool
	defenseSSHHost       string
	defenseSSHPort       int
}

func newRuntimeHTTPServiceAdapter(commandService runtimeHTTPCommandService, queryService runtimeHTTPQueryService, proxyTickets runtimeHTTPProxyTicketService, proxyTicketReader runtimeports.ProxyTicketInstanceReader, defenseWorkbench runtimeDefenseWorkbenchRuntime, proxyBodyPreviewSize int, defenseSSHEnabled bool, defenseSSHHost string, defenseSSHPort int) *runtimeHTTPServiceAdapter {
	return &runtimeHTTPServiceAdapter{
		commandService:       commandService,
		queryService:         queryService,
		proxyTickets:         proxyTickets,
		proxyTicketReader:    proxyTicketReader,
		defenseWorkbench:     defenseWorkbench,
		proxyBodyPreviewSize: proxyBodyPreviewSize,
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

const (
	awdDefenseMaxFileSize      = 256 * 1024
	awdDefenseMaxDirectoryList = 300
	awdDefenseMaxCommandSize   = 2000
	awdDefenseMaxCommandOutput = 64 * 1024
)

func (a *runtimeHTTPServiceAdapter) ReadAWDDefenseFile(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64, filePath string) (*dto.AWDDefenseFileResp, error) {
	if a == nil || a.proxyTicketReader == nil || a.defenseWorkbench == nil {
		return nil, errRuntimeHTTPProxyTicketServiceUnavailable()
	}
	cleanPath, err := normalizeAWDDefensePath(filePath)
	if err != nil {
		return nil, err
	}
	scope, err := a.proxyTicketReader.FindAWDDefenseSSHScope(ctx, user.UserID, contestID, serviceID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if scope == nil || scope.ContainerID == "" {
		return nil, errcode.ErrForbidden
	}

	content, err := a.defenseWorkbench.ReadFileFromContainer(ctx, scope.ContainerID, cleanPath, awdDefenseMaxFileSize)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return &dto.AWDDefenseFileResp{
		Path:    cleanPath,
		Content: string(content),
		Size:    len(content),
	}, nil
}

func (a *runtimeHTTPServiceAdapter) ListAWDDefenseDirectory(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64, dirPath string) (*dto.AWDDefenseDirectoryResp, error) {
	if a == nil || a.proxyTicketReader == nil || a.defenseWorkbench == nil {
		return nil, errRuntimeHTTPProxyTicketServiceUnavailable()
	}
	cleanPath, err := normalizeAWDDefenseDirectoryPath(dirPath)
	if err != nil {
		return nil, err
	}
	scope, err := a.proxyTicketReader.FindAWDDefenseSSHScope(ctx, user.UserID, contestID, serviceID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if scope == nil || scope.ContainerID == "" {
		return nil, errcode.ErrForbidden
	}

	entries, err := a.defenseWorkbench.ListDirectoryFromContainer(ctx, scope.ContainerID, cleanPath, awdDefenseMaxDirectoryList)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	resp := &dto.AWDDefenseDirectoryResp{
		Path:    cleanPath,
		Entries: make([]dto.AWDDefenseDirectoryEntryResp, 0, len(entries)),
	}
	for _, entry := range entries {
		entryPath := entry.Name
		if cleanPath != "." {
			entryPath = path.Join(cleanPath, entry.Name)
		}
		resp.Entries = append(resp.Entries, dto.AWDDefenseDirectoryEntryResp{
			Name: entry.Name,
			Path: entryPath,
			Type: entry.Type,
			Size: entry.Size,
		})
	}
	return resp, nil
}

func (a *runtimeHTTPServiceAdapter) SaveAWDDefenseFile(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64, req dto.AWDDefenseFileSaveReq) (*dto.AWDDefenseFileSaveResp, error) {
	if a == nil || a.proxyTicketReader == nil || a.defenseWorkbench == nil {
		return nil, errRuntimeHTTPProxyTicketServiceUnavailable()
	}
	cleanPath, err := normalizeAWDDefensePath(req.Path)
	if err != nil {
		return nil, err
	}
	content := []byte(req.Content)
	if len(content) > awdDefenseMaxFileSize {
		return nil, errcode.ErrInvalidParams.WithCause(fmt.Errorf("awd defense file is too large"))
	}
	scope, err := a.proxyTicketReader.FindAWDDefenseSSHScope(ctx, user.UserID, contestID, serviceID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if scope == nil || scope.ContainerID == "" {
		return nil, errcode.ErrForbidden
	}

	backupPath := ""
	if req.Backup {
		existing, readErr := a.defenseWorkbench.ReadFileFromContainer(ctx, scope.ContainerID, cleanPath, awdDefenseMaxFileSize)
		if readErr == nil {
			backupPath = fmt.Sprintf("%s.bak.%d", cleanPath, time.Now().Unix())
			if err := a.defenseWorkbench.WriteFileToContainer(ctx, scope.ContainerID, backupPath, existing); err != nil {
				return nil, errcode.ErrInternal.WithCause(err)
			}
		}
	}
	if err := a.defenseWorkbench.WriteFileToContainer(ctx, scope.ContainerID, cleanPath, content); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return &dto.AWDDefenseFileSaveResp{
		Path:       cleanPath,
		Size:       len(content),
		BackupPath: backupPath,
	}, nil
}

func (a *runtimeHTTPServiceAdapter) RunAWDDefenseCommand(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64, req dto.AWDDefenseCommandReq) (*dto.AWDDefenseCommandResp, error) {
	if a == nil || a.proxyTicketReader == nil || a.defenseWorkbench == nil {
		return nil, errRuntimeHTTPProxyTicketServiceUnavailable()
	}
	command := strings.TrimSpace(req.Command)
	if command == "" || len(command) > awdDefenseMaxCommandSize {
		return nil, errcode.ErrInvalidParams
	}
	scope, err := a.proxyTicketReader.FindAWDDefenseSSHScope(ctx, user.UserID, contestID, serviceID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if scope == nil || scope.ContainerID == "" {
		return nil, errcode.ErrForbidden
	}

	runCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	output, err := a.defenseWorkbench.ExecContainerCommand(runCtx, scope.ContainerID, []string{"/bin/sh", "-lc", command}, nil, awdDefenseMaxCommandOutput)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return &dto.AWDDefenseCommandResp{
		Command: command,
		Output:  string(output),
	}, nil
}

func normalizeAWDDefenseDirectoryPath(input string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" || trimmed == "." {
		return ".", nil
	}
	return normalizeAWDDefensePath(trimmed)
}

func normalizeAWDDefensePath(input string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", errcode.ErrInvalidParams
	}
	if strings.HasPrefix(trimmed, "/") {
		return "", errcode.ErrInvalidParams
	}
	cleaned := path.Clean(trimmed)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.Contains(cleaned, "/../") {
		return "", errcode.ErrInvalidParams
	}
	return cleaned, nil
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
	if a == nil || a.proxyTickets == nil {
		return 0
	}
	return a.proxyTickets.MaxAge()
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

func errRuntimeHTTPProxyTicketServiceUnavailable() error {
	return errcode.ErrInternal.WithCause(fmt.Errorf("proxy ticket service is not configured"))
}

type runtimePracticeServiceAdapter struct {
	cleaner     *runtimecmd.RuntimeCleanupService
	provisioner *runtimecmd.ProvisioningService
}

func newRuntimePracticeServiceAdapter(cleaner *runtimecmd.RuntimeCleanupService, provisioner *runtimecmd.ProvisioningService) practiceports.RuntimeInstanceService {
	if cleaner == nil && provisioner == nil {
		return nil
	}
	return &runtimePracticeServiceAdapter{
		cleaner:     cleaner,
		provisioner: provisioner,
	}
}

func (a *runtimePracticeServiceAdapter) CleanupRuntime(ctx context.Context, instance *model.Instance) error {
	if a == nil || a.cleaner == nil {
		return nil
	}
	return a.cleaner.CleanupRuntime(ctx, instance)
}

func (a *runtimePracticeServiceAdapter) CreateTopology(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error) {
	if a == nil || a.provisioner == nil || req == nil {
		return nil, nil
	}

	result, err := a.provisioner.CreateTopology(ctx, toRuntimeTopologyCreateRequest(req))
	if err != nil {
		return nil, err
	}
	return fromRuntimeTopologyCreateResult(result), nil
}

func (a *runtimePracticeServiceAdapter) CreateContainer(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (containerID, networkID string, hostPort, servicePort int, err error) {
	if a == nil || a.provisioner == nil {
		return "", "", 0, 0, nil
	}
	return a.provisioner.CreateContainer(ctx, imageName, env, reservedHostPort)
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
			Env:             cloneRuntimeStringMap(node.Env),
			ServicePort:     node.ServicePort,
			ServiceProtocol: node.ServiceProtocol,
			IsEntryPoint:    node.IsEntryPoint,
			NetworkKeys:     append([]string(nil), node.NetworkKeys...),
			NetworkAliases:  append([]string(nil), node.NetworkAliases...),
			Resources:       cloneRuntimeResourceLimits(node.Resources),
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

func fromRuntimeTopologyCreateResult(result *runtimeports.TopologyCreateResult) *practiceports.TopologyCreateResult {
	if result == nil {
		return nil
	}
	return &practiceports.TopologyCreateResult{
		PrimaryContainerID: result.PrimaryContainerID,
		NetworkID:          result.NetworkID,
		AccessURL:          result.AccessURL,
		RuntimeDetails:     result.RuntimeDetails,
	}
}

func cloneRuntimeStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	output := make(map[string]string, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}

func cloneRuntimeResourceLimits(input *model.ResourceLimits) *model.ResourceLimits {
	if input == nil {
		return nil
	}
	return &model.ResourceLimits{
		CPUQuota:  input.CPUQuota,
		Memory:    input.Memory,
		PidsLimit: input.PidsLimit,
	}
}

type runtimeChallengeServiceAdapter struct {
	cleaner     *runtimecmd.RuntimeCleanupService
	provisioner *runtimecmd.ProvisioningService
	publicHost  string
}

func newRuntimeChallengeServiceAdapter(cleaner *runtimecmd.RuntimeCleanupService, provisioner *runtimecmd.ProvisioningService, publicHost string) challengeports.ChallengeRuntimeProbe {
	if cleaner == nil && provisioner == nil {
		return nil
	}
	return &runtimeChallengeServiceAdapter{
		cleaner:     cleaner,
		provisioner: provisioner,
		publicHost:  publicHost,
	}
}

func (a *runtimeChallengeServiceAdapter) CreateTopology(ctx context.Context, req *challengeports.RuntimeTopologyCreateRequest) (*challengeports.RuntimeTopologyCreateResult, error) {
	if a == nil || a.provisioner == nil {
		return nil, fmt.Errorf("runtime provisioning service is not configured")
	}
	if req == nil {
		return nil, fmt.Errorf("runtime topology create request is nil")
	}
	result, err := a.provisioner.CreateTopology(ctx, toRuntimeChallengeTopologyCreateRequest(req))
	if err != nil {
		return nil, err
	}
	return &challengeports.RuntimeTopologyCreateResult{
		AccessURL:      result.AccessURL,
		RuntimeDetails: result.RuntimeDetails,
	}, nil
}

func (a *runtimeChallengeServiceAdapter) CreateContainer(ctx context.Context, imageName string, env map[string]string) (string, model.InstanceRuntimeDetails, error) {
	if a == nil || a.provisioner == nil {
		return "", model.InstanceRuntimeDetails{}, fmt.Errorf("runtime provisioning service is not configured")
	}

	containerID, networkID, hostPort, servicePort, err := a.provisioner.CreateContainer(ctx, imageName, env, 0)
	if err != nil {
		return "", model.InstanceRuntimeDetails{}, err
	}

	accessURL := fmt.Sprintf("http://%s:%d", a.publicHost, hostPort)
	return accessURL, model.InstanceRuntimeDetails{
		Networks: []model.InstanceRuntimeNetwork{
			{
				Key:       model.TopologyDefaultNetworkKey,
				Name:      model.TopologyDefaultNetworkKey,
				NetworkID: networkID,
			},
		},
		Containers: []model.InstanceRuntimeContainer{
			{
				NodeKey:         "default",
				ContainerID:     containerID,
				ServicePort:     servicePort,
				ServiceProtocol: model.ChallengeTargetProtocolHTTP,
				HostPort:        hostPort,
				IsEntryPoint:    true,
				NetworkKeys:     []string{model.TopologyDefaultNetworkKey},
			},
		},
	}, nil
}

func (a *runtimeChallengeServiceAdapter) CleanupRuntimeDetails(ctx context.Context, details model.InstanceRuntimeDetails) error {
	if a == nil || a.cleaner == nil {
		return nil
	}

	rawDetails, err := model.EncodeInstanceRuntimeDetails(details)
	if err != nil {
		return err
	}
	instance := &model.Instance{
		RuntimeDetails: rawDetails,
	}
	return a.cleaner.CleanupRuntime(ctx, instance)
}

func toRuntimeChallengeTopologyCreateRequest(req *challengeports.RuntimeTopologyCreateRequest) *runtimeports.TopologyCreateRequest {
	if req == nil {
		return nil
	}
	networks := make([]runtimeports.TopologyCreateNetwork, 0, len(req.Networks))
	for _, network := range req.Networks {
		networks = append(networks, runtimeports.TopologyCreateNetwork{
			Key:      network.Key,
			Internal: network.Internal,
		})
	}

	nodes := make([]runtimeports.TopologyCreateNode, 0, len(req.Nodes))
	for _, node := range req.Nodes {
		nodes = append(nodes, runtimeports.TopologyCreateNode{
			Key:             node.Key,
			Image:           node.Image,
			Env:             cloneRuntimeStringMap(node.Env),
			ServicePort:     node.ServicePort,
			ServiceProtocol: node.ServiceProtocol,
			IsEntryPoint:    node.IsEntryPoint,
			NetworkKeys:     append([]string(nil), node.NetworkKeys...),
			Resources:       cloneRuntimeResourceLimits(node.Resources),
		})
	}
	return &runtimeports.TopologyCreateRequest{
		Networks:                   networks,
		Nodes:                      nodes,
		Policies:                   append([]model.TopologyTrafficPolicy(nil), req.Policies...),
		DisableEntryPortPublishing: true,
	}
}
