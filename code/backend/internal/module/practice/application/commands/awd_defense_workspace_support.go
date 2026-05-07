package commands

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path"
	"strings"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	practiceports "ctf-platform/internal/module/practice/ports"
)

const (
	awdDefenseWorkspaceShellImage = "python:3.12-alpine"
	awdDefenseWorkspaceWorkingDir = "/workspace"
	// Keep the companion shell usable out of the box instead of relying on the
	// SSH client to negotiate locale/editor state each time.
	awdDefenseWorkspaceBootstrapPrelude = `set -e
if ! command -v git >/dev/null 2>&1 || ! command -v vim >/dev/null 2>&1 || ! command -v nano >/dev/null 2>&1; then
  apk add --no-cache git vim nano
fi`
	awdDefenseWorkspaceGitUserName          = "workspace"
	awdDefenseWorkspaceGitUserEmail         = "workspace@local"
	awdDefenseWorkspaceInitialCommitMessage = "Initial workspace snapshot"
)

var awdDefenseWorkspaceShellEnv = map[string]string{
	"LANG":   "C.UTF-8",
	"LC_ALL": "C.UTF-8",
	"TERM":   "xterm-256color",
}

type awdDefenseWorkspaceRepository interface {
	FindAWDDefenseWorkspace(ctx context.Context, contestID, teamID, serviceID int64) (*model.AWDDefenseWorkspace, error)
	UpsertAWDDefenseWorkspace(ctx context.Context, workspace *model.AWDDefenseWorkspace) error
}

type awdDefenseWorkspacePlan struct {
	contestID              int64
	teamID                 int64
	serviceID              int64
	workspaceRevision      int64
	seedSignature          string
	runtimeMounts          []model.ContainerMount
	workspaceMounts        []model.ContainerMount
	workspaceContainerID   string
	workspaceContainerName string
	checkerTokenEnv        string
	checkerToken           string
	createWorkspace        bool
}

type awdDefenseWorkspaceConfig struct {
	seedRoot       string
	workspaceRoots []awdDefenseWorkspaceRoot
	runtimeMounts  []awdDefenseRuntimeMount
}

type awdDefenseWorkspaceRoot struct {
	source   string
	readOnly bool
}

type awdDefenseRuntimeMount struct {
	source   string
	target   string
	readOnly bool
}

func resolveAWDDefenseWorkspaceRepository(repo any) awdDefenseWorkspaceRepository {
	if repo == nil {
		return nil
	}
	value, _ := repo.(awdDefenseWorkspaceRepository)
	return value
}

func buildAWDDefenseWorkspaceBootstrapCommand(mounts []model.ContainerMount) string {
	var builder strings.Builder
	builder.WriteString(awdDefenseWorkspaceBootstrapPrelude)

	for _, target := range listAWDDefenseWorkspaceWritableTargets(mounts) {
		quotedTarget := shellQuoteForPOSIXSh(target)
		quotedGitDir := shellQuoteForPOSIXSh(path.Join(target, ".git"))
		builder.WriteString("\nif [ -d ")
		builder.WriteString(quotedTarget)
		builder.WriteString(" ] && [ ! -d ")
		builder.WriteString(quotedGitDir)
		builder.WriteString(" ]; then\n")
		builder.WriteString("  git -C ")
		builder.WriteString(quotedTarget)
		builder.WriteString(" init\n")
		builder.WriteString("  git -C ")
		builder.WriteString(quotedTarget)
		builder.WriteString(" config user.name ")
		builder.WriteString(shellQuoteForPOSIXSh(awdDefenseWorkspaceGitUserName))
		builder.WriteString("\n")
		builder.WriteString("  git -C ")
		builder.WriteString(quotedTarget)
		builder.WriteString(" config user.email ")
		builder.WriteString(shellQuoteForPOSIXSh(awdDefenseWorkspaceGitUserEmail))
		builder.WriteString("\n")
		builder.WriteString("  git -C ")
		builder.WriteString(quotedTarget)
		builder.WriteString(" add --all\n")
		builder.WriteString("  git -C ")
		builder.WriteString(quotedTarget)
		builder.WriteString(" commit --allow-empty -m ")
		builder.WriteString(shellQuoteForPOSIXSh(awdDefenseWorkspaceInitialCommitMessage))
		builder.WriteString("\nfi")
	}

	builder.WriteString("\nexec tail -f /dev/null")
	return builder.String()
}

func listAWDDefenseWorkspaceWritableTargets(mounts []model.ContainerMount) []string {
	if len(mounts) == 0 {
		return nil
	}
	targets := make([]string, 0, len(mounts))
	seen := make(map[string]struct{}, len(mounts))
	for _, mount := range mounts {
		if mount.ReadOnly {
			continue
		}
		target := strings.TrimSpace(mount.Target)
		if target == "" {
			continue
		}
		if _, exists := seen[target]; exists {
			continue
		}
		seen[target] = struct{}{}
		targets = append(targets, target)
	}
	return targets
}

func shellQuoteForPOSIXSh(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}

func (s *Service) prepareAWDDefenseWorkspacePlan(ctx context.Context, instance *model.Instance, chal *model.Challenge) (*awdDefenseWorkspacePlan, error) {
	if !isAWDInstance(instance) || instance.TeamID == nil {
		return nil, nil
	}
	if s.repo == nil {
		return nil, fmt.Errorf("awd service repository is not configured")
	}
	workspaceRepo := resolveAWDDefenseWorkspaceRepository(s.instanceRepo)
	if workspaceRepo == nil {
		return nil, fmt.Errorf("awd defense workspace repository is not configured")
	}

	contestID := *instance.ContestID
	teamID := *instance.TeamID
	serviceID := *instance.ServiceID

	service, err := s.repo.FindContestAWDService(ctx, contestID, serviceID)
	if err != nil {
		return nil, err
	}
	snapshot, err := model.DecodeContestAWDServiceSnapshot(service.ServiceSnapshot)
	if err != nil {
		return nil, err
	}
	config, err := parseAWDDefenseWorkspaceConfig(snapshot.RuntimeConfig)
	if err != nil {
		return nil, err
	}

	current, err := workspaceRepo.FindAWDDefenseWorkspace(ctx, contestID, teamID, serviceID)
	if err != nil {
		return nil, err
	}

	workspaceRevision := int64(1)
	if current != nil && current.WorkspaceRevision > 0 {
		workspaceRevision = current.WorkspaceRevision
	}
	seedSignature := buildAWDDefenseWorkspaceSeedSignature(service.ServiceSnapshot)
	if current != nil && strings.TrimSpace(current.SeedSignature) != "" {
		seedSignature = current.SeedSignature
	}

	volumeBySource := make(map[string]string, len(config.workspaceRoots))
	workspaceMounts := make([]model.ContainerMount, 0, len(config.workspaceRoots))
	for _, root := range config.workspaceRoots {
		relative := relativeAWDDefenseWorkspaceRoot(config.seedRoot, root.source)
		volumeName := buildAWDDefenseWorkspaceVolumeName(instance, workspaceRevision, relative)
		volumeBySource[root.source] = volumeName
		workspaceMounts = append(workspaceMounts, model.ContainerMount{
			Source:   volumeName,
			Target:   buildAWDDefenseWorkspaceTarget(relative),
			ReadOnly: root.readOnly,
		})
	}

	runtimeMounts := make([]model.ContainerMount, 0, len(config.runtimeMounts))
	for _, item := range config.runtimeMounts {
		volumeName := volumeBySource[item.source]
		if volumeName == "" {
			return nil, fmt.Errorf("workspace root volume is missing for %s", item.source)
		}
		runtimeMounts = append(runtimeMounts, model.ContainerMount{
			Source:   volumeName,
			Target:   item.target,
			ReadOnly: item.readOnly,
		})
	}

	plan := &awdDefenseWorkspacePlan{
		contestID:              contestID,
		teamID:                 teamID,
		serviceID:              serviceID,
		workspaceRevision:      workspaceRevision,
		seedSignature:          seedSignature,
		runtimeMounts:          runtimeMounts,
		workspaceMounts:        workspaceMounts,
		workspaceContainerName: buildAWDDefenseWorkspaceContainerName(chal, instance, workspaceRevision),
	}
	checkerTokenEnv := strings.TrimSpace(readStringFromAny(snapshot.RuntimeConfig["checker_token_env"]))
	if checkerTokenEnv != "" {
		challengeID := service.AWDChallengeID
		if challengeID <= 0 {
			challengeID = instance.ChallengeID
		}
		secret := ""
		if s.config != nil {
			secret = s.config.Container.FlagGlobalSecret
		}
		checkerToken := contestdomain.BuildAWDCheckerToken(contestID, teamID, serviceID, challengeID, secret)
		if strings.TrimSpace(checkerToken) == "" {
			return nil, fmt.Errorf("awd checker token secret is not configured")
		}
		plan.checkerTokenEnv = checkerTokenEnv
		plan.checkerToken = checkerToken
	}
	if current != nil {
		plan.workspaceContainerID = strings.TrimSpace(current.ContainerID)
	}
	plan.createWorkspace = current == nil || current.Status != model.AWDDefenseWorkspaceStatusRunning || plan.workspaceContainerID == ""
	if !plan.createWorkspace {
		state, err := s.runtimeService.InspectManagedContainer(ctx, plan.workspaceContainerID)
		if err != nil {
			return nil, err
		}
		if state == nil || !state.Exists || !state.Running {
			plan.workspaceContainerID = ""
			plan.createWorkspace = true
		}
	}
	return plan, nil
}

func parseAWDDefenseWorkspaceConfig(runtimeConfig map[string]any) (*awdDefenseWorkspaceConfig, error) {
	if runtimeConfig == nil {
		return nil, fmt.Errorf("awd runtime config is empty")
	}
	raw, ok := runtimeConfig["defense_workspace"]
	if !ok {
		return nil, fmt.Errorf("awd runtime config defense_workspace is empty")
	}
	payload, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("awd runtime config defense_workspace must be an object")
	}

	seedRoot := strings.TrimSpace(readStringFromAny(payload["seed_root"]))
	if seedRoot == "" {
		return nil, fmt.Errorf("awd defense workspace seed_root is empty")
	}

	workspaceRoots := readStringListFromAny(payload["workspace_roots"])
	if len(workspaceRoots) == 0 {
		return nil, fmt.Errorf("awd defense workspace roots are empty")
	}
	writableRootSet := make(map[string]struct{})
	for _, root := range readStringListFromAny(payload["writable_roots"]) {
		writableRootSet[root] = struct{}{}
	}

	roots := make([]awdDefenseWorkspaceRoot, 0, len(workspaceRoots))
	for _, root := range workspaceRoots {
		_, writable := writableRootSet[root]
		roots = append(roots, awdDefenseWorkspaceRoot{
			source:   root,
			readOnly: !writable,
		})
	}

	runtimeMounts, err := parseAWDDefenseRuntimeMounts(payload["runtime_mounts"])
	if err != nil {
		return nil, err
	}
	return &awdDefenseWorkspaceConfig{
		seedRoot:       seedRoot,
		workspaceRoots: roots,
		runtimeMounts:  runtimeMounts,
	}, nil
}

func parseAWDDefenseRuntimeMounts(raw any) ([]awdDefenseRuntimeMount, error) {
	items, ok := raw.([]any)
	if !ok || len(items) == 0 {
		return nil, fmt.Errorf("awd defense runtime mounts are empty")
	}

	result := make([]awdDefenseRuntimeMount, 0, len(items))
	for _, item := range items {
		payload, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("awd defense runtime mount must be an object")
		}
		source := strings.TrimSpace(readStringFromAny(payload["source"]))
		target := strings.TrimSpace(readStringFromAny(payload["target"]))
		mode := strings.ToLower(strings.TrimSpace(readStringFromAny(payload["mode"])))
		if source == "" || target == "" || mode == "" {
			return nil, fmt.Errorf("awd defense runtime mount is incomplete")
		}
		result = append(result, awdDefenseRuntimeMount{
			source:   source,
			target:   target,
			readOnly: mode == "ro",
		})
	}
	return result, nil
}

func readStringFromAny(raw any) string {
	switch typed := raw.(type) {
	case string:
		return typed
	default:
		return ""
	}
}

func readStringListFromAny(raw any) []string {
	switch typed := raw.(type) {
	case []string:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			value := strings.TrimSpace(item)
			if value == "" {
				continue
			}
			items = append(items, value)
		}
		return items
	case []any:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			value := strings.TrimSpace(readStringFromAny(item))
			if value == "" {
				continue
			}
			items = append(items, value)
		}
		return items
	default:
		return nil
	}
}

func buildAWDDefenseWorkspaceSeedSignature(raw string) string {
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}

func relativeAWDDefenseWorkspaceRoot(seedRoot, root string) string {
	normalizedSeed := strings.Trim(path.Clean(seedRoot), "/")
	normalizedRoot := strings.Trim(path.Clean(root), "/")
	if normalizedRoot == normalizedSeed {
		return ""
	}
	return strings.Trim(strings.TrimPrefix(normalizedRoot, normalizedSeed+"/"), "/")
}

func buildAWDDefenseWorkspaceTarget(relative string) string {
	relative = strings.Trim(relative, "/")
	if relative == "" {
		return awdDefenseWorkspaceWorkingDir
	}
	return awdDefenseWorkspaceWorkingDir + "/" + relative
}

func applyAWDDefenseWorkspaceRuntimeMounts(request *practiceports.TopologyCreateRequest, mounts []model.ContainerMount) {
	if request == nil || len(mounts) == 0 {
		return
	}
	for idx := range request.Nodes {
		if !request.Nodes[idx].IsEntryPoint {
			continue
		}
		request.Nodes[idx].Mounts = append(request.Nodes[idx].Mounts, mounts...)
		return
	}
}

func (s *Service) persistAWDDefenseWorkspaceState(ctx context.Context, plan *awdDefenseWorkspacePlan, instanceID int64, status, containerID string) error {
	if plan == nil {
		return nil
	}
	workspaceRepo := resolveAWDDefenseWorkspaceRepository(s.instanceRepo)
	if workspaceRepo == nil {
		return fmt.Errorf("awd defense workspace repository is not configured")
	}
	return workspaceRepo.UpsertAWDDefenseWorkspace(ctx, &model.AWDDefenseWorkspace{
		ContestID:         plan.contestID,
		TeamID:            plan.teamID,
		ServiceID:         plan.serviceID,
		InstanceID:        instanceID,
		WorkspaceRevision: plan.workspaceRevision,
		Status:            status,
		ContainerID:       containerID,
		SeedSignature:     plan.seedSignature,
	})
}

func (s *Service) createAWDDefenseWorkspaceCompanion(ctx context.Context, instance *model.Instance, plan *awdDefenseWorkspacePlan) (string, error) {
	if s == nil || s.runtimeService == nil || plan == nil {
		return "", fmt.Errorf("awd defense workspace runtime is not configured")
	}
	result, err := s.runtimeService.CreateTopology(ctx, &practiceports.TopologyCreateRequest{
		DisableEntryPortPublishing: true,
		ContainerName:              plan.workspaceContainerName,
		Networks: []practiceports.TopologyCreateNetwork{
			{
				Key:    model.TopologyDefaultNetworkKey,
				Name:   buildAWDContestNetworkName(instance),
				Shared: true,
			},
		},
		Nodes: []practiceports.TopologyCreateNode{
			{
				Key:             "workspace",
				Image:           awdDefenseWorkspaceShellImage,
				Env:             cloneAWDDefenseWorkspaceShellEnv(),
				Command:         []string{"/bin/sh", "-lc", buildAWDDefenseWorkspaceBootstrapCommand(plan.workspaceMounts)},
				WorkingDir:      awdDefenseWorkspaceWorkingDir,
				ServicePort:     22,
				ServiceProtocol: model.ChallengeTargetProtocolTCP,
				IsEntryPoint:    true,
				NetworkKeys:     []string{model.TopologyDefaultNetworkKey},
				NetworkAliases:  []string{buildAWDDefenseWorkspaceAlias(instance, plan.workspaceRevision)},
				Mounts:          append([]model.ContainerMount(nil), plan.workspaceMounts...),
			},
		},
	})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(result.PrimaryContainerID), nil
}

func cloneAWDDefenseWorkspaceShellEnv() map[string]string {
	cloned := make(map[string]string, len(awdDefenseWorkspaceShellEnv))
	for key, value := range awdDefenseWorkspaceShellEnv {
		cloned[key] = value
	}
	return cloned
}
