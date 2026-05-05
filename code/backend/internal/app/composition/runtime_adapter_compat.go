package composition

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	practiceports "ctf-platform/internal/module/practice/ports"
	runtimeports "ctf-platform/internal/module/runtime/ports"
	"ctf-platform/pkg/errcode"
)

type runtimeHTTPCommandService interface{}

type runtimeHTTPQueryService interface{}

type runtimeDefenseWorkbenchRuntime interface {
	ReadFileFromContainer(ctx context.Context, containerID, filePath string, limit int64) ([]byte, error)
	ListDirectoryFromContainer(ctx context.Context, containerID, dirPath string, limit int) ([]runtimeports.ContainerDirectoryEntry, error)
	WriteFileToContainer(ctx context.Context, containerID, filePath string, content []byte) error
	ExecContainerCommand(ctx context.Context, containerID string, command []string, stdin []byte, limit int64) ([]byte, error)
}

type runtimeHTTPServiceAdapter struct {
	proxyTickets         runtimeHTTPProxyTicketService
	proxyTicketReader    runtimeports.ProxyTicketInstanceReader
	defenseWorkbench     runtimeDefenseWorkbenchRuntime
	proxyBodyPreviewSize int
	defenseSSHEnabled    bool
	defenseSSHHost       string
	defenseSSHPort       int
	defenseWorkbenchRoot string
}

func newRuntimeHTTPServiceAdapter(
	_ runtimeHTTPCommandService,
	_ runtimeHTTPQueryService,
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
		proxyTickets:         proxyTickets,
		proxyTicketReader:    proxyTicketReader,
		defenseWorkbench:     defenseWorkbench,
		proxyBodyPreviewSize: proxyBodyPreviewSize,
		defenseSSHEnabled:    defenseSSHEnabled,
		defenseSSHHost:       defenseSSHHost,
		defenseSSHPort:       defenseSSHPort,
	}
	if defenseWorkbenchEnabled {
		adapter.defenseWorkbenchRoot = strings.TrimSpace(defenseWorkbenchRoot)
	}
	return adapter
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
	scope, cleanPath, err := a.resolveDefenseScope(ctx, user, contestID, serviceID, filePath, false)
	if err != nil {
		return nil, err
	}
	content, err := a.defenseWorkbench.ReadFileFromContainer(ctx, scope.ContainerID, cleanPath, 256*1024)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return &dto.AWDDefenseFileResp{Path: filePath, Content: string(content), Size: len(content)}, nil
}

func (a *runtimeHTTPServiceAdapter) ListAWDDefenseDirectory(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64, dirPath string) (*dto.AWDDefenseDirectoryResp, error) {
	scope, cleanPath, err := a.resolveDefenseScope(ctx, user, contestID, serviceID, dirPath, true)
	if err != nil {
		return nil, err
	}
	entries, err := a.defenseWorkbench.ListDirectoryFromContainer(ctx, scope.ContainerID, cleanPath, 300)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	resp := &dto.AWDDefenseDirectoryResp{Path: dirPath, Entries: make([]dto.AWDDefenseDirectoryEntryResp, 0, len(entries))}
	for _, entry := range entries {
		if isSensitiveDefensePath(entry.Name) {
			continue
		}
		resp.Entries = append(resp.Entries, dto.AWDDefenseDirectoryEntryResp{
			Path: entry.Name,
			Type: entry.Type,
			Size: entry.Size,
		})
	}
	return resp, nil
}

func (a *runtimeHTTPServiceAdapter) resolveDefenseScope(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64, inputPath string, dir bool) (*runtimeports.AWDDefenseSSHScope, string, error) {
	if a == nil || a.proxyTicketReader == nil || a.defenseWorkbench == nil {
		return nil, "", errRuntimeHTTPProxyTicketServiceUnavailable()
	}
	root := strings.TrimSpace(a.defenseWorkbenchRoot)
	if !strings.HasPrefix(root, "/") || root == "/" {
		return nil, "", errcode.ErrInvalidParams
	}
	var cleanPath string
	var err error
	if dir {
		cleanPath, err = normalizeAWDDefenseDirectoryPath(inputPath)
	} else {
		cleanPath, err = normalizeAWDDefensePath(inputPath)
	}
	if err != nil {
		return nil, "", err
	}
	if isSensitiveDefensePath(cleanPath) {
		return nil, "", errcode.ErrForbidden
	}
	scope, err := a.proxyTicketReader.FindAWDDefenseSSHScope(ctx, user.UserID, contestID, serviceID)
	if err != nil {
		return nil, "", errcode.ErrInternal.WithCause(err)
	}
	if scope == nil || scope.ContainerID == "" {
		return nil, "", errcode.ErrForbidden
	}
	if cleanPath == "." {
		return scope, root, nil
	}
	return scope, path.Join(root, cleanPath), nil
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
