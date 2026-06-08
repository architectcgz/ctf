package commands

import (
	"context"
	"fmt"
	"path"
	"strings"

	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	practiceentity "ctf-platform/internal/module/practice/entity"
	practiceports "ctf-platform/internal/module/practice/ports"
	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
)

const (
	awdDefenseWorkspaceShellImage = "python:3.12-alpine"
	awdDefenseWorkspaceWorkingDir = "/workspace"
	// Keep the companion shell usable out of the box instead of relying on the
	// SSH client to negotiate locale/editor state each time.
	awdDefenseWorkspaceBootstrapPrelude = `set -e
missing_tools=""
if ! command -v git >/dev/null 2>&1; then
  missing_tools="$missing_tools git"
fi
if ! command -v vim >/dev/null 2>&1; then
  missing_tools="$missing_tools vim"
fi
if ! command -v nano >/dev/null 2>&1; then
  missing_tools="$missing_tools nano"
fi
if [ -n "$missing_tools" ] && command -v apk >/dev/null 2>&1; then
  apk add --no-cache $missing_tools || true
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
	FindAWDDefenseWorkspace(ctx context.Context, contestID, teamID, serviceID int64) (*runtimecontracts.AWDDefenseWorkspace, error)
	UpsertAWDDefenseWorkspace(ctx context.Context, workspace *runtimecontracts.AWDDefenseWorkspace) error
}

type awdDefenseWorkspacePlan struct {
	contestID                 int64
	teamID                    int64
	serviceID                 int64
	workspaceRevision         int64
	seedSignature             string
	runtimeMounts             []runtimecontracts.ContainerMount
	workspaceMounts           []runtimecontracts.ContainerMount
	workspaceContainerID      string
	staleWorkspaceContainerID string
	workspaceContainerName    string
	checkerTokenEnv           string
	checkerToken              string
	createWorkspace           bool
}

func resolveAWDDefenseWorkspaceRepository(repo any) awdDefenseWorkspaceRepository {
	if repo == nil {
		return nil
	}
	value, _ := repo.(awdDefenseWorkspaceRepository)
	return value
}

func buildAWDDefenseWorkspaceBootstrapCommand(mounts []runtimecontracts.ContainerMount) string {
	var builder strings.Builder
	builder.WriteString(awdDefenseWorkspaceBootstrapPrelude)

	for _, target := range listAWDDefenseWorkspaceWritableTargets(mounts) {
		quotedTarget := shellQuoteForPOSIXSh(target)
		quotedGitDir := shellQuoteForPOSIXSh(path.Join(target, ".git"))
		builder.WriteString("\nif command -v git >/dev/null 2>&1 && [ -d ")
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

func listAWDDefenseWorkspaceWritableTargets(mounts []runtimecontracts.ContainerMount) []string {
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

func (s *Service) prepareAWDDefenseWorkspacePlan(ctx context.Context, instance *instancecontracts.Instance, chal *practiceentity.Challenge) (*awdDefenseWorkspacePlan, error) {
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

	subject, err := s.repo.FindContestAWDServiceRuntimeSubject(ctx, contestID, serviceID)
	if err != nil {
		return nil, err
	}
	if subject == nil {
		return nil, fmt.Errorf("awd service runtime subject is not configured")
	}
	config := subject.WorkspaceConfig
	err = validateAWDDefenseWorkspaceConfig(config)
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
	seedSignature := subject.SeedSignature
	if current != nil && strings.TrimSpace(current.SeedSignature) != "" {
		seedSignature = current.SeedSignature
	}

	volumeBySource := make(map[string]string, len(config.WorkspaceRoots))
	workspaceMounts := make([]runtimecontracts.ContainerMount, 0, len(config.WorkspaceRoots))
	for _, root := range config.WorkspaceRoots {
		relative := relativeAWDDefenseWorkspaceRoot(config.SeedRoot, root.Source)
		volumeName := buildAWDDefenseWorkspaceVolumeName(instance, workspaceRevision, relative)
		volumeBySource[root.Source] = volumeName
		workspaceMounts = append(workspaceMounts, runtimecontracts.ContainerMount{
			Source:   volumeName,
			Target:   buildAWDDefenseWorkspaceTarget(relative),
			ReadOnly: root.ReadOnly,
		})
	}

	runtimeMounts := make([]runtimecontracts.ContainerMount, 0, len(config.RuntimeMounts))
	for _, item := range config.RuntimeMounts {
		volumeName := volumeBySource[item.Source]
		if volumeName == "" {
			return nil, fmt.Errorf("workspace root volume is missing for %s", item.Source)
		}
		runtimeMounts = append(runtimeMounts, runtimecontracts.ContainerMount{
			Source:   volumeName,
			Target:   item.Target,
			ReadOnly: item.ReadOnly,
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
	checkerTokenEnv := strings.TrimSpace(config.CheckerTokenEnv)
	if checkerTokenEnv != "" {
		challengeID := subject.ChallengeID
		if challengeID <= 0 {
			challengeID = instance.ChallengeID
		}
		secret := ""
		if s.config != nil {
			secret = s.config.Container.FlagGlobalSecret
		}
		checkerToken := contestcontracts.BuildAWDCheckerToken(contestID, teamID, serviceID, challengeID, secret)
		if strings.TrimSpace(checkerToken) == "" {
			return nil, fmt.Errorf("awd checker token secret is not configured")
		}
		plan.checkerTokenEnv = checkerTokenEnv
		plan.checkerToken = checkerToken
	}
	if current != nil {
		plan.workspaceContainerID = strings.TrimSpace(current.ContainerID)
		if current.Status != runtimecontracts.AWDDefenseWorkspaceStatusRunning && plan.workspaceContainerID != "" {
			plan.staleWorkspaceContainerID = plan.workspaceContainerID
			plan.workspaceContainerID = ""
		}
	}
	plan.createWorkspace = current == nil || current.Status != runtimecontracts.AWDDefenseWorkspaceStatusRunning || plan.workspaceContainerID == ""
	if !plan.createWorkspace {
		state, err := s.runtimeService.InspectManagedContainer(ctx, plan.workspaceContainerID)
		if err != nil {
			return nil, err
		}
		if state == nil || !state.Exists {
			plan.workspaceContainerID = ""
			plan.createWorkspace = true
		} else if !state.Running {
			plan.staleWorkspaceContainerID = plan.workspaceContainerID
			plan.workspaceContainerID = ""
			plan.createWorkspace = true
		}
	}
	return plan, nil
}

func validateAWDDefenseWorkspaceConfig(config *practiceports.ContestAWDDefenseWorkspaceConfig) error {
	if config == nil {
		return fmt.Errorf("awd runtime config defense_workspace is empty")
	}
	if strings.TrimSpace(config.SeedRoot) == "" {
		return fmt.Errorf("awd defense workspace seed_root is empty")
	}
	if len(config.WorkspaceRoots) == 0 {
		return fmt.Errorf("awd defense workspace roots are empty")
	}
	if len(config.RuntimeMounts) == 0 {
		return fmt.Errorf("awd defense runtime mounts are empty")
	}
	for _, mount := range config.RuntimeMounts {
		if strings.TrimSpace(mount.Source) == "" || strings.TrimSpace(mount.Target) == "" {
			return fmt.Errorf("awd defense runtime mount is incomplete")
		}
	}
	return nil
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

func applyAWDDefenseWorkspaceRuntimeMounts(request *practiceports.TopologyCreateRequest, mounts []runtimecontracts.ContainerMount) {
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
	return workspaceRepo.UpsertAWDDefenseWorkspace(ctx, &runtimecontracts.AWDDefenseWorkspace{
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

func (s *Service) createAWDDefenseWorkspaceCompanion(ctx context.Context, instance *instancecontracts.Instance, plan *awdDefenseWorkspacePlan) (string, error) {
	if s == nil || s.runtimeService == nil || plan == nil {
		return "", fmt.Errorf("awd defense workspace runtime is not configured")
	}
	result, err := s.runtimeService.CreateTopology(ctx, &practiceports.TopologyCreateRequest{
		NodeID:                     runtimeNodeIDValue(instance.NodeID),
		DisableEntryPortPublishing: true,
		ContainerName:              plan.workspaceContainerName,
		Networks: []practiceports.TopologyCreateNetwork{
			{
				Key:    challengecontracts.TopologyDefaultNetworkKey,
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
				ServiceProtocol: runtimecontracts.ChallengeTargetProtocolTCP,
				IsEntryPoint:    true,
				NetworkKeys:     []string{challengecontracts.TopologyDefaultNetworkKey},
				NetworkAliases:  []string{buildAWDDefenseWorkspaceAlias(instance, plan.workspaceRevision)},
				Mounts:          append([]runtimecontracts.ContainerMount(nil), plan.workspaceMounts...),
			},
		},
	})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(result.PrimaryContainerID), nil
}

func (s *Service) cleanupAWDDefenseWorkspaceCompanion(ctx context.Context, containerID string) error {
	if s == nil || s.runtimeService == nil || strings.TrimSpace(containerID) == "" {
		return nil
	}
	runtimeDetails, err := runtimecontracts.EncodeInstanceRuntimeDetails(runtimecontracts.InstanceRuntimeDetails{
		Containers: []runtimecontracts.InstanceRuntimeContainer{
			{ContainerID: strings.TrimSpace(containerID)},
		},
	})
	if err != nil {
		return err
	}
	return s.runtimeService.CleanupRuntime(ctx, &instancecontracts.Instance{RuntimeDetails: runtimeDetails})
}

func resolveAWDDefenseWorkspaceFailureContainerID(plan *awdDefenseWorkspacePlan, containerID string) string {
	if trimmed := strings.TrimSpace(containerID); trimmed != "" {
		return trimmed
	}
	if plan == nil {
		return ""
	}
	return strings.TrimSpace(plan.staleWorkspaceContainerID)
}

func (s *Service) persistAWDDefenseWorkspaceFailure(ctx context.Context, plan *awdDefenseWorkspacePlan, instanceID int64, containerID string) {
	if plan == nil || !plan.createWorkspace {
		return
	}
	_ = s.persistAWDDefenseWorkspaceState(
		ctx,
		plan,
		instanceID,
		runtimecontracts.AWDDefenseWorkspaceStatusFailed,
		resolveAWDDefenseWorkspaceFailureContainerID(plan, containerID),
	)
}

func cloneAWDDefenseWorkspaceShellEnv() map[string]string {
	cloned := make(map[string]string, len(awdDefenseWorkspaceShellEnv))
	for key, value := range awdDefenseWorkspaceShellEnv {
		cloned[key] = value
	}
	return cloned
}
