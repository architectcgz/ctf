package application

import (
	"context"
	"fmt"
	"path"
	"sort"
	"strings"
	"time"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	instanceports "ctf-platform/internal/module/instance/ports"
	"ctf-platform/pkg/errcode"
)

type awdDefenseWorkbenchScopeReader interface {
	FindAWDDefenseSSHScope(ctx context.Context, userID, contestID, serviceID int64) (*instanceports.AWDDefenseSSHScope, error)
}

type awdDefenseWorkbenchRuntime interface {
	ReadFileFromContainer(ctx context.Context, containerID, filePath string, limit int64) ([]byte, error)
	ListDirectoryFromContainer(ctx context.Context, containerID, dirPath string, limit int) ([]instanceports.ContainerDirectoryEntry, error)
	WriteFileToContainer(ctx context.Context, containerID, filePath string, content []byte) error
	ExecContainerCommand(ctx context.Context, containerID string, command []string, stdin []byte, limit int64) ([]byte, error)
}

type AWDDefenseWorkbenchConfig struct {
	ReadOnlyEnabled bool
	Root            string
}

type AWDDefenseWorkbenchService struct {
	scopeReader     awdDefenseWorkbenchScopeReader
	runtime         awdDefenseWorkbenchRuntime
	readOnlyEnabled bool
	root            string
}

func NewAWDDefenseWorkbenchService(
	scopeReader awdDefenseWorkbenchScopeReader,
	runtime awdDefenseWorkbenchRuntime,
	cfg AWDDefenseWorkbenchConfig,
) *AWDDefenseWorkbenchService {
	return &AWDDefenseWorkbenchService{
		scopeReader:     scopeReader,
		runtime:         runtime,
		readOnlyEnabled: cfg.ReadOnlyEnabled,
		root:            strings.TrimSpace(cfg.Root),
	}
}

func (s *AWDDefenseWorkbenchService) ReadAWDDefenseFile(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64, filePath string) (*dto.AWDDefenseFileResp, error) {
	scope, root, err := s.resolveReadOnlyScope(ctx, user, contestID, serviceID)
	if err != nil {
		return nil, err
	}
	cleanPath, containerPath, err := resolveEditableDefenseFilePath(scope, root, filePath)
	if err != nil {
		return nil, err
	}
	content, err := s.runtime.ReadFileFromContainer(ctx, scope.ContainerID, containerPath, 256*1024)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return &dto.AWDDefenseFileResp{Path: cleanPath, Content: string(content), Size: len(content)}, nil
}

func (s *AWDDefenseWorkbenchService) ListAWDDefenseDirectory(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64, dirPath string) (*dto.AWDDefenseDirectoryResp, error) {
	scope, root, err := s.resolveReadOnlyScope(ctx, user, contestID, serviceID)
	if err != nil {
		return nil, err
	}
	cleanPath, containerPath, err := resolveEditableDefenseDirectoryPath(scope, root, dirPath)
	if err != nil {
		return nil, err
	}
	entries, err := s.runtime.ListDirectoryFromContainer(ctx, scope.ContainerID, containerPath, 300)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return &dto.AWDDefenseDirectoryResp{
		Path:    cleanPath,
		Entries: buildEditableDefenseDirectoryEntries(scope.EditablePaths, cleanPath, entries, 300),
	}, nil
}

func (s *AWDDefenseWorkbenchService) SaveAWDDefenseFile(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64, req dto.AWDDefenseFileSaveReq) (*dto.AWDDefenseFileSaveResp, error) {
	scope, root, err := s.resolveReadOnlyScope(ctx, user, contestID, serviceID)
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
		existing, readErr := s.runtime.ReadFileFromContainer(ctx, scope.ContainerID, containerPath, 256*1024)
		if readErr == nil {
			backupPath = fmt.Sprintf("%s.bak.%d", cleanPath, time.Now().Unix())
			backupContainerPath := defenseWorkbenchContainerPath(root, backupPath)
			if err := s.runtime.WriteFileToContainer(ctx, scope.ContainerID, backupContainerPath, existing); err != nil {
				return nil, errcode.ErrInternal.WithCause(err)
			}
		}
	}
	if err := s.runtime.WriteFileToContainer(ctx, scope.ContainerID, containerPath, content); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return &dto.AWDDefenseFileSaveResp{
		Path:       cleanPath,
		Size:       len(content),
		BackupPath: backupPath,
	}, nil
}

func (s *AWDDefenseWorkbenchService) RunAWDDefenseCommand(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64, req dto.AWDDefenseCommandReq) (*dto.AWDDefenseCommandResp, error) {
	if s == nil || s.scopeReader == nil || s.runtime == nil {
		return nil, errAWDDefenseWorkbenchUnavailable()
	}
	command := strings.TrimSpace(req.Command)
	if command == "" || len(command) > 2000 {
		return nil, errcode.ErrInvalidParams
	}
	scope, err := s.scopeReader.FindAWDDefenseSSHScope(ctx, user.UserID, contestID, serviceID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if scope == nil || scope.ContainerID == "" {
		return nil, errcode.ErrForbidden
	}

	runCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	output, err := s.runtime.ExecContainerCommand(runCtx, scope.ContainerID, []string{"/bin/sh", "-lc", command}, nil, 64*1024)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return &dto.AWDDefenseCommandResp{
		Command: command,
		Output:  string(output),
	}, nil
}

func (s *AWDDefenseWorkbenchService) resolveReadOnlyScope(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64) (*instanceports.AWDDefenseSSHScope, string, error) {
	if s == nil || s.scopeReader == nil || s.runtime == nil {
		return nil, "", errAWDDefenseWorkbenchUnavailable()
	}
	if !s.readOnlyEnabled {
		return nil, "", errcode.ErrForbidden
	}
	root := strings.TrimSpace(s.root)
	if !strings.HasPrefix(root, "/") || root == "/" {
		return nil, "", errcode.ErrInvalidParams
	}
	scope, err := s.scopeReader.FindAWDDefenseSSHScope(ctx, user.UserID, contestID, serviceID)
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

func resolveEditableDefenseFilePath(scope *instanceports.AWDDefenseSSHScope, root, inputPath string) (string, string, error) {
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

func resolveEditableDefenseDirectoryPath(scope *instanceports.AWDDefenseSSHScope, root, inputPath string) (string, string, error) {
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

func buildEditableDefenseDirectoryEntries(editablePaths []string, dirPath string, actualEntries []instanceports.ContainerDirectoryEntry, limit int) []dto.AWDDefenseDirectoryEntryResp {
	normalized := normalizeDefenseEditablePaths(editablePaths)
	if len(normalized) == 0 {
		return []dto.AWDDefenseDirectoryEntryResp{}
	}

	actualByName := make(map[string]instanceports.ContainerDirectoryEntry, len(actualEntries))
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

func errAWDDefenseWorkbenchUnavailable() error {
	return errcode.ErrInternal.WithCause(fmt.Errorf("awd defense workbench service is not configured"))
}
