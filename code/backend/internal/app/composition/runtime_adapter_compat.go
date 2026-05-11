package composition

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
	practiceports "ctf-platform/internal/module/practice/ports"
	runtimeports "ctf-platform/internal/module/runtime/ports"
	"ctf-platform/pkg/errcode"
)

type runtimeHTTPCommandService interface {
	DestroyInstance(ctx context.Context, instanceID, userID int64) error
	ExtendInstance(ctx context.Context, instanceID, userID int64) (*dto.InstanceResp, error)
	DestroyTeacherInstance(ctx context.Context, instanceID, requesterID int64, requesterRole string) error
}

type runtimeHTTPQueryService interface {
	GetAccessURL(ctx context.Context, instanceID, userID int64) (string, error)
	GetUserInstances(ctx context.Context, userID int64) ([]*dto.InstanceInfo, error)
	ListTeacherInstances(ctx context.Context, requesterID int64, requesterRole string, query *dto.TeacherInstanceQuery) ([]dto.TeacherInstanceItem, error)
}

type runtimeDefenseWorkbenchRuntime interface {
	ReadFileFromContainer(ctx context.Context, containerID, filePath string, limit int64) ([]byte, error)
	ListDirectoryFromContainer(ctx context.Context, containerID, dirPath string, limit int) ([]runtimeports.ContainerDirectoryEntry, error)
	WriteFileToContainer(ctx context.Context, containerID, filePath string, content []byte) error
	ExecContainerCommand(ctx context.Context, containerID string, command []string, stdin []byte, limit int64) ([]byte, error)
}

type runtimeHTTPServiceAdapter struct {
	commandService                  runtimeHTTPCommandService
	queryService                    runtimeHTTPQueryService
	proxyTickets                    runtimeHTTPProxyTicketService
	proxyTicketReader               runtimeports.ProxyTicketInstanceReader
	defenseWorkbench                runtimeDefenseWorkbenchRuntime
	proxyBodyPreviewSize            int
	defenseSSHEnabled               bool
	defenseSSHHost                  string
	defenseSSHPort                  int
	defenseWorkbenchReadOnlyEnabled bool
	defenseWorkbenchRoot            string
}

func newRuntimeHTTPServiceAdapter(
	commandService runtimeHTTPCommandService,
	queryService runtimeHTTPQueryService,
	proxyTickets runtimeHTTPProxyTicketService,
	proxyTicketReader runtimeports.ProxyTicketInstanceReader,
	defenseWorkbench runtimeDefenseWorkbenchRuntime,
	proxyBodyPreviewSize int,
	defenseSSHEnabled bool,
	defenseSSHHost string,
	defenseSSHPort int,
	defenseWorkbenchEnabled bool,
	defenseWorkbenchRoot string,
) *runtimeHTTPServiceAdapter {
	adapter := &runtimeHTTPServiceAdapter{
		commandService:                  commandService,
		queryService:                    queryService,
		proxyTickets:                    proxyTickets,
		proxyTicketReader:               proxyTicketReader,
		defenseWorkbench:                defenseWorkbench,
		proxyBodyPreviewSize:            proxyBodyPreviewSize,
		defenseSSHEnabled:               defenseSSHEnabled,
		defenseSSHHost:                  defenseSSHHost,
		defenseSSHPort:                  defenseSSHPort,
		defenseWorkbenchReadOnlyEnabled: defenseWorkbenchEnabled,
	}
	adapter.defenseWorkbenchRoot = strings.TrimSpace(defenseWorkbenchRoot)
	return adapter
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
	scope, root, err := a.resolveDefenseWorkbenchScope(ctx, user, contestID, serviceID)
	if err != nil {
		return nil, err
	}
	cleanPath, containerPath, err := resolveEditableDefenseFilePath(scope, root, filePath)
	if err != nil {
		return nil, err
	}
	content, err := a.defenseWorkbench.ReadFileFromContainer(ctx, scope.ContainerID, containerPath, 256*1024)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return &dto.AWDDefenseFileResp{Path: cleanPath, Content: string(content), Size: len(content)}, nil
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
	entries, err := a.defenseWorkbench.ListDirectoryFromContainer(ctx, scope.ContainerID, containerPath, 300)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return &dto.AWDDefenseDirectoryResp{
		Path:    cleanPath,
		Entries: buildEditableDefenseDirectoryEntries(scope.EditablePaths, cleanPath, entries, 300),
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
	if len(content) > 256*1024 {
		return nil, errcode.ErrInvalidParams.WithCause(fmt.Errorf("awd defense file is too large"))
	}

	backupPath := ""
	if req.Backup {
		existing, readErr := a.defenseWorkbench.ReadFileFromContainer(ctx, scope.ContainerID, containerPath, 256*1024)
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
	if command == "" || len(command) > 2000 {
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
	output, err := a.defenseWorkbench.ExecContainerCommand(runCtx, scope.ContainerID, []string{"/bin/sh", "-lc", command}, nil, 64*1024)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return &dto.AWDDefenseCommandResp{
		Command: command,
		Output:  string(output),
	}, nil
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

func normalizeAWDDefenseDirectoryPath(input string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" || trimmed == "." {
		return ".", nil
	}
	return normalizeAWDDefensePath(trimmed)
}

func normalizeAWDDefensePath(input string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" || strings.HasPrefix(trimmed, "/") {
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

func isSensitiveDefensePath(value string) bool {
	lower := strings.ToLower(strings.TrimSpace(value))
	switch {
	case lower == ".env",
		strings.Contains(lower, ".env"),
		strings.HasPrefix(lower, ".ssh"),
		strings.HasPrefix(lower, "proc/"),
		strings.HasPrefix(lower, "sys/"),
		strings.HasPrefix(lower, "dev/"),
		strings.HasPrefix(lower, "run/secrets"),
		strings.HasPrefix(lower, "var/run/docker.sock"):
		return true
	default:
		return false
	}
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
