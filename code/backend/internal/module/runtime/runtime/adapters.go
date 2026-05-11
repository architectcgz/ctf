package runtime

import (
	"context"
	"fmt"
	"path"
	"sort"
	"strings"
	"time"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
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
	commandService                  instancecontracts.InstanceCommandService
	queryService                    instancecontracts.InstanceQueryService
	proxyTickets                    instancecontracts.ProxyTicketService
	proxyTicketReader               runtimeports.ProxyTicketInstanceReader
	defenseWorkbench                runtimeDefenseWorkbenchRuntime
	proxyBodyPreviewSize            int
	proxyTicketMaxAge               int
	defenseWorkbenchReadOnlyEnabled bool
	defenseWorkbenchRoot            string
	defenseSSHEnabled               bool
	defenseSSHHost                  string
	defenseSSHPort                  int
}

func newRuntimeHTTPServiceAdapter(commandService instancecontracts.InstanceCommandService, queryService instancecontracts.InstanceQueryService, proxyTickets instancecontracts.ProxyTicketService, proxyTicketReader runtimeports.ProxyTicketInstanceReader, defenseWorkbench runtimeDefenseWorkbenchRuntime, proxyBodyPreviewSize int, proxyTicketMaxAge int, defenseSSHEnabled bool, defenseSSHHost string, defenseSSHPort int, defenseWorkbenchReadOnlyEnabled bool, defenseWorkbenchRoot string) *runtimeHTTPServiceAdapter {
	return &runtimeHTTPServiceAdapter{
		commandService:                  commandService,
		queryService:                    queryService,
		proxyTickets:                    proxyTickets,
		proxyTicketReader:               proxyTicketReader,
		defenseWorkbench:                defenseWorkbench,
		proxyBodyPreviewSize:            proxyBodyPreviewSize,
		proxyTicketMaxAge:               proxyTicketMaxAge,
		defenseWorkbenchReadOnlyEnabled: defenseWorkbenchReadOnlyEnabled,
		defenseWorkbenchRoot:            strings.TrimSpace(defenseWorkbenchRoot),
		defenseSSHEnabled:               defenseSSHEnabled,
		defenseSSHHost:                  defenseSSHHost,
		defenseSSHPort:                  defenseSSHPort,
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
	scope, root, err := a.resolveDefenseWorkbenchScope(ctx, user, contestID, serviceID)
	if err != nil {
		return nil, err
	}
	cleanPath, containerPath, err := resolveEditableDefenseFilePath(scope, root, filePath)
	if err != nil {
		return nil, err
	}
	content, err := a.defenseWorkbench.ReadFileFromContainer(ctx, scope.ContainerID, containerPath, awdDefenseMaxFileSize)
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
	scope, root, err := a.resolveDefenseWorkbenchScope(ctx, user, contestID, serviceID)
	if err != nil {
		return nil, err
	}
	cleanPath, containerPath, err := resolveEditableDefenseDirectoryPath(scope, root, dirPath)
	if err != nil {
		return nil, err
	}
	entries, err := a.defenseWorkbench.ListDirectoryFromContainer(ctx, scope.ContainerID, containerPath, awdDefenseMaxDirectoryList)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return &dto.AWDDefenseDirectoryResp{
		Path:    cleanPath,
		Entries: buildEditableDefenseDirectoryEntries(scope.EditablePaths, cleanPath, entries, awdDefenseMaxDirectoryList),
	}, nil
}

func (a *runtimeHTTPServiceAdapter) SaveAWDDefenseFile(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64, req dto.AWDDefenseFileSaveReq) (*dto.AWDDefenseFileSaveResp, error) {
	if a == nil || a.proxyTicketReader == nil || a.defenseWorkbench == nil {
		return nil, errRuntimeHTTPProxyTicketServiceUnavailable()
	}
	scope, root, err := a.resolveDefenseWorkbenchScope(ctx, user, contestID, serviceID)
	if err != nil {
		return nil, err
	}
	cleanPath, containerPath, err := resolveEditableDefenseFilePath(scope, root, req.Path)
	if err != nil {
		return nil, err
	}
	content := []byte(req.Content)
	if len(content) > awdDefenseMaxFileSize {
		return nil, errcode.ErrInvalidParams.WithCause(fmt.Errorf("awd defense file is too large"))
	}

	backupPath := ""
	if req.Backup {
		existing, readErr := a.defenseWorkbench.ReadFileFromContainer(ctx, scope.ContainerID, containerPath, awdDefenseMaxFileSize)
		if readErr == nil {
			backupPath = fmt.Sprintf("%s.bak.%d", cleanPath, time.Now().Unix())
			backupContainerPath := defenseWorkbenchContainerPath(root, backupPath)
			if err := a.defenseWorkbench.WriteFileToContainer(ctx, scope.ContainerID, backupContainerPath, existing); err != nil {
				return nil, errcode.ErrInternal.WithCause(err)
			}
		}
	}
	if err := a.defenseWorkbench.WriteFileToContainer(ctx, scope.ContainerID, containerPath, content); err != nil {
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

func mapAWDDefensePackagePathToContainer(pathValue string) string {
	if pathValue == "." || pathValue == "" {
		return "."
	}
	if pathValue == "docker" {
		return "."
	}
	if strings.HasPrefix(pathValue, "docker/") {
		mapped := strings.TrimPrefix(pathValue, "docker/")
		if mapped == "" {
			return "."
		}
		return mapped
	}
	return pathValue
}

func defenseWorkbenchContainerPath(root, contractPath string) string {
	mapped := mapAWDDefensePackagePathToContainer(contractPath)
	if mapped == "." {
		return root
	}
	return path.Join(root, mapped)
}

func normalizeDefenseEditablePaths(paths []string) []string {
	if len(paths) == 0 {
		return nil
	}
	normalized := make([]string, 0, len(paths))
	seen := make(map[string]struct{}, len(paths))
	for _, item := range paths {
		clean, err := normalizeAWDDefensePath(item)
		if err != nil {
			continue
		}
		if _, exists := seen[clean]; exists {
			continue
		}
		seen[clean] = struct{}{}
		normalized = append(normalized, clean)
	}
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

func isEditableDefenseFileAllowed(editablePaths []string, contractPath string) bool {
	for _, item := range normalizeDefenseEditablePaths(editablePaths) {
		if item == contractPath {
			return true
		}
	}
	return false
}

func hasEditableDefenseDirectoryAccess(editablePaths []string, dirPath string) bool {
	normalized := normalizeDefenseEditablePaths(editablePaths)
	if len(normalized) == 0 {
		return false
	}
	if dirPath == "." {
		return true
	}
	prefix := dirPath + "/"
	for _, item := range normalized {
		if strings.HasPrefix(item, prefix) {
			return true
		}
	}
	return false
}

func resolveEditableDefenseFilePath(scope *runtimeports.AWDDefenseSSHScope, root, inputPath string) (string, string, error) {
	cleanPath, err := normalizeAWDDefensePath(inputPath)
	if err != nil {
		return "", "", err
	}
	if !isEditableDefenseFileAllowed(scope.EditablePaths, cleanPath) {
		return "", "", errcode.ErrForbidden
	}
	containerPath := defenseWorkbenchContainerPath(root, cleanPath)
	if isSensitiveDefensePath(containerPath) {
		return "", "", errcode.ErrForbidden
	}
	return cleanPath, containerPath, nil
}

func resolveEditableDefenseDirectoryPath(scope *runtimeports.AWDDefenseSSHScope, root, inputPath string) (string, string, error) {
	cleanPath, err := normalizeAWDDefenseDirectoryPath(inputPath)
	if err != nil {
		return "", "", err
	}
	if !hasEditableDefenseDirectoryAccess(scope.EditablePaths, cleanPath) {
		return "", "", errcode.ErrForbidden
	}
	containerPath := defenseWorkbenchContainerPath(root, cleanPath)
	if isSensitiveDefensePath(containerPath) {
		return "", "", errcode.ErrForbidden
	}
	return cleanPath, containerPath, nil
}

func buildEditableDefenseDirectoryEntries(editablePaths []string, dirPath string, actualEntries []runtimeports.ContainerDirectoryEntry, limit int) []dto.AWDDefenseDirectoryEntryResp {
	normalized := normalizeDefenseEditablePaths(editablePaths)
	if len(normalized) == 0 {
		return []dto.AWDDefenseDirectoryEntryResp{}
	}

	actualByName := make(map[string]runtimeports.ContainerDirectoryEntry, len(actualEntries))
	for _, entry := range actualEntries {
		actualByName[entry.Name] = entry
	}

	type virtualEntry struct {
		name string
		path string
		typ  string
	}

	collected := make(map[string]virtualEntry)
	for _, item := range normalized {
		name, entryPath, entryType, ok := nextEditableDefenseDirectoryEntry(dirPath, item)
		if !ok {
			continue
		}
		existing, exists := collected[entryPath]
		if !exists || (existing.typ == "file" && entryType == "dir") {
			collected[entryPath] = virtualEntry{name: name, path: entryPath, typ: entryType}
		}
	}

	if len(collected) == 0 {
		return []dto.AWDDefenseDirectoryEntryResp{}
	}

	result := make([]dto.AWDDefenseDirectoryEntryResp, 0, len(collected))
	for _, entry := range collected {
		size := int64(0)
		if entry.typ == "file" {
			if actual, ok := actualByName[entry.name]; ok && actual.Type == "file" {
				size = actual.Size
			}
		}
		result = append(result, dto.AWDDefenseDirectoryEntryResp{
			Name: entry.name,
			Path: entry.path,
			Type: entry.typ,
			Size: size,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Type != result[j].Type {
			return result[i].Type == "dir"
		}
		return result[i].Name < result[j].Name
	})
	if len(result) > limit {
		return result[:limit]
	}
	return result
}

func nextEditableDefenseDirectoryEntry(dirPath, filePath string) (string, string, string, bool) {
	if dirPath == "." {
		parts := strings.Split(filePath, "/")
		if len(parts) == 0 || parts[0] == "" {
			return "", "", "", false
		}
		if len(parts) == 1 {
			return parts[0], parts[0], "file", true
		}
		return parts[0], parts[0], "dir", true
	}

	prefix := dirPath + "/"
	if !strings.HasPrefix(filePath, prefix) {
		return "", "", "", false
	}
	remainder := strings.TrimPrefix(filePath, prefix)
	if remainder == "" {
		return "", "", "", false
	}
	parts := strings.Split(remainder, "/")
	childName := parts[0]
	childPath := path.Join(dirPath, childName)
	if len(parts) == 1 {
		return childName, childPath, "file", true
	}
	return childName, childPath, "dir", true
}

func (a *runtimeHTTPServiceAdapter) resolveDefenseWorkbenchScope(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64) (*runtimeports.AWDDefenseSSHScope, string, error) {
	if a == nil || a.proxyTicketReader == nil || a.defenseWorkbench == nil {
		return nil, "", errRuntimeHTTPProxyTicketServiceUnavailable()
	}
	if !a.defenseWorkbenchReadOnlyEnabled {
		return nil, "", errcode.ErrForbidden
	}
	root := strings.TrimSpace(a.defenseWorkbenchRoot)
	if !strings.HasPrefix(root, "/") || root == "/" {
		return nil, "", errcode.ErrInvalidParams
	}
	scope, err := a.proxyTicketReader.FindAWDDefenseSSHScope(ctx, user.UserID, contestID, serviceID)
	if err != nil {
		return nil, "", errcode.ErrInternal.WithCause(err)
	}
	if scope == nil || scope.ContainerID == "" {
		return nil, "", errcode.ErrForbidden
	}
	if len(normalizeDefenseEditablePaths(scope.EditablePaths)) == 0 {
		return nil, "", errcode.ErrForbidden
	}
	return scope, root, nil
}

func isSensitiveDefensePath(value string) bool {
	lower := strings.ToLower(strings.TrimSpace(strings.ReplaceAll(value, "\\", "/")))
	if lower == "" {
		return false
	}
	segments := strings.Split(lower, "/")
	for _, segment := range segments {
		switch {
		case segment == ".env",
			strings.HasPrefix(segment, ".env."),
			segment == ".ssh",
			segment == "id_rsa",
			segment == "id_ed25519",
			segment == "authorized_keys",
			segment == "known_hosts":
			return true
		}
	}
	switch {
	case strings.HasPrefix(lower, "proc/"),
		strings.HasPrefix(lower, "sys/"),
		strings.HasPrefix(lower, "dev/"),
		strings.HasPrefix(lower, "run/secrets"),
		strings.HasPrefix(lower, "var/run/docker.sock"):
		return true
	default:
		return false
	}
}

func (a *runtimeHTTPServiceAdapter) ResolveProxyTicket(ctx context.Context, ticket string) (*instancecontracts.ProxyTicketClaims, error) {
	if a == nil || a.proxyTickets == nil {
		return nil, errRuntimeHTTPProxyTicketServiceUnavailable()
	}
	return a.proxyTickets.ResolveTicket(ctx, ticket)
}

func (a *runtimeHTTPServiceAdapter) ResolveAWDTargetAccessURL(ctx context.Context, claims *instancecontracts.ProxyTicketClaims, contestID, serviceID, victimTeamID int64) (string, error) {
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

func errRuntimeHTTPProxyTicketServiceUnavailable() error {
	return errcode.ErrInternal.WithCause(fmt.Errorf("proxy ticket service is not configured"))
}

type runtimePracticeServiceAdapter struct {
	cleaner     *runtimecmd.RuntimeCleanupService
	provisioner *runtimecmd.ProvisioningService
	engine      Engine
}

func newRuntimePracticeServiceAdapter(cleaner *runtimecmd.RuntimeCleanupService, provisioner *runtimecmd.ProvisioningService, engine Engine) practiceports.RuntimeInstanceService {
	if cleaner == nil && provisioner == nil && engine == nil {
		return nil
	}
	return &runtimePracticeServiceAdapter{
		cleaner:     cleaner,
		provisioner: provisioner,
		engine:      engine,
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

func (a *runtimePracticeServiceAdapter) InspectManagedContainer(ctx context.Context, containerID string) (*practiceports.ManagedContainerState, error) {
	if a == nil || a.engine == nil {
		return nil, nil
	}
	state, err := a.engine.InspectManagedContainer(ctx, containerID)
	if err != nil || state == nil {
		return nil, err
	}
	return &practiceports.ManagedContainerState{
		ID:      state.ID,
		Exists:  state.Exists,
		Running: state.Running,
		Status:  state.Status,
	}, nil
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
			Command:         append([]string(nil), node.Command...),
			WorkingDir:      node.WorkingDir,
			ServicePort:     node.ServicePort,
			ServiceProtocol: node.ServiceProtocol,
			IsEntryPoint:    node.IsEntryPoint,
			NetworkKeys:     append([]string(nil), node.NetworkKeys...),
			NetworkAliases:  append([]string(nil), node.NetworkAliases...),
			Mounts:          append([]model.ContainerMount(nil), node.Mounts...),
			Resources:       cloneRuntimeResourceLimits(node.Resources),
		})
	}

	return &runtimeports.TopologyCreateRequest{
		Networks:                   networks,
		Nodes:                      nodes,
		Policies:                   append([]model.TopologyTrafficPolicy(nil), req.Policies...),
		ReservedHostPort:           req.ReservedHostPort,
		DisableEntryPortPublishing: req.DisableEntryPortPublishing,
		ContainerName:              req.ContainerName,
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
