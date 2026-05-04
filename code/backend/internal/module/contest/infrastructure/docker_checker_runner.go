package infrastructure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	networktypes "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"ctf-platform/internal/config"
	contestports "ctf-platform/internal/module/contest/ports"
	runtimedomain "ctf-platform/internal/module/runtime/domain"
)

type dockerCheckerClient interface {
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *networktypes.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error)
	ContainerStart(ctx context.Context, container string, options container.StartOptions) error
	ContainerWait(ctx context.Context, container string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error)
	ContainerLogs(ctx context.Context, container string, options container.LogsOptions) (io.ReadCloser, error)
	ContainerInspect(ctx context.Context, container string) (types.ContainerJSON, error)
	ContainerRemove(ctx context.Context, container string, options container.RemoveOptions) error
}

type DockerCheckerRunner struct {
	cli dockerCheckerClient
	cfg config.CheckerSandboxConfig
}

type dockerCheckerContainerSpec struct {
	ContainerConfig *container.Config
	HostConfig      *container.HostConfig
	NetworkConfig   *networktypes.NetworkingConfig
}

type checkerLimitedBuffer struct {
	buf      bytes.Buffer
	limit    int64
	exceeded bool
}

func (b *checkerLimitedBuffer) Write(p []byte) (int, error) {
	if b == nil {
		return len(p), nil
	}
	if b.limit <= 0 {
		b.exceeded = true
		return len(p), nil
	}
	remaining := int(b.limit) - b.buf.Len()
	if remaining <= 0 {
		b.exceeded = true
		return len(p), nil
	}
	if len(p) > remaining {
		_, _ = b.buf.Write(p[:remaining])
		b.exceeded = true
		return len(p), nil
	}
	_, _ = b.buf.Write(p)
	return len(p), nil
}

func (b *checkerLimitedBuffer) String() string {
	if b == nil {
		return ""
	}
	return b.buf.String()
}

func NewDockerCheckerRunner(cfg config.CheckerSandboxConfig) (*DockerCheckerRunner, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return NewDockerCheckerRunnerWithClient(cli, cfg), nil
}

func NewDockerCheckerRunnerWithClient(cli dockerCheckerClient, cfg config.CheckerSandboxConfig) *DockerCheckerRunner {
	return &DockerCheckerRunner{cli: cli, cfg: cfg}
}

func (r *DockerCheckerRunner) RunChecker(ctx context.Context, job contestports.CheckerRunJob) (contestports.CheckerRunResult, error) {
	startedAt := time.Now()
	result := contestports.CheckerRunResult{
		Status:    contestports.CheckerRunStatusFailed,
		Reason:    contestports.CheckerReasonSandboxError,
		StartedAt: startedAt,
	}
	if r == nil || r.cli == nil {
		result.FinishedAt = time.Now()
		result.Duration = result.FinishedAt.Sub(startedAt)
		return result, fmt.Errorf("checker runner docker client is not configured")
	}

	timeout := job.Timeout
	if timeout <= 0 {
		timeout = r.cfg.Timeout
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	workDir, err := materializeCheckerFiles(job.Files, strings.TrimSpace(r.cfg.HostWorkRoot))
	if err != nil {
		result.FinishedAt = time.Now()
		result.Duration = result.FinishedAt.Sub(startedAt)
		return result, err
	}
	defer func() {
		_ = os.RemoveAll(workDir)
	}()

	spec, err := r.buildContainerSpec(job, workDir)
	if err != nil {
		result.FinishedAt = time.Now()
		result.Duration = result.FinishedAt.Sub(startedAt)
		return result, err
	}

	created, err := r.cli.ContainerCreate(runCtx, spec.ContainerConfig, spec.HostConfig, spec.NetworkConfig, nil, "")
	if err != nil {
		result.FinishedAt = time.Now()
		result.Duration = result.FinishedAt.Sub(startedAt)
		return result, err
	}
	containerID := created.ID
	defer func() {
		if containerID != "" {
			cleanupCtx, cleanupCancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			defer cleanupCancel()
			_ = r.cli.ContainerRemove(cleanupCtx, containerID, container.RemoveOptions{Force: true, RemoveVolumes: true})
		}
	}()

	if err := r.cli.ContainerStart(runCtx, containerID, container.StartOptions{}); err != nil {
		result.FinishedAt = time.Now()
		result.Duration = result.FinishedAt.Sub(startedAt)
		return result, err
	}

	waitCh, errCh := r.cli.ContainerWait(runCtx, containerID, container.WaitConditionNotRunning)
	var waitResp container.WaitResponse
	select {
	case waitResp = <-waitCh:
	case err := <-errCh:
		result.FinishedAt = time.Now()
		result.Duration = result.FinishedAt.Sub(startedAt)
		return result, err
	case <-runCtx.Done():
		result.Reason = contestports.CheckerReasonTimeout
		result.ResourceLimitHit = "timeout"
		result.FinishedAt = time.Now()
		result.Duration = result.FinishedAt.Sub(startedAt)
		return result, nil
	}

	stdout, stderr, outputLimitHit := r.collectLogs(runCtx, containerID, effectiveOutputLimit(job, r.cfg))
	result.Stdout = stdout
	result.Stderr = stderr
	result.OutputLimitHit = outputLimitHit
	result.ExitCode = waitResp.StatusCode
	result.FinishedAt = time.Now()
	result.Duration = result.FinishedAt.Sub(startedAt)

	if outputLimitHit {
		result.Reason = contestports.CheckerReasonOutputLimitExceeded
		return result, nil
	}
	if waitResp.StatusCode != 0 {
		result.Reason = contestports.CheckerReasonFailed
		return result, nil
	}
	if strings.EqualFold(strings.TrimSpace(job.OutputMode), "json") {
		status, reason, err := parseCheckerJSONOutput(stdout)
		if err != nil {
			result.Reason = contestports.CheckerReasonInvalidOutput
			return result, nil
		}
		result.Status = status
		result.Reason = reason
		return result, nil
	}
	result.Status = contestports.CheckerRunStatusOK
	result.Reason = contestports.CheckerReasonPassed
	return result, nil
}

func (r *DockerCheckerRunner) buildContainerSpec(job contestports.CheckerRunJob, hostWorkDir string) (dockerCheckerContainerSpec, error) {
	image := strings.TrimSpace(job.Image)
	if image == "" {
		image = strings.TrimSpace(r.cfg.Image)
	}
	if image == "" {
		return dockerCheckerContainerSpec{}, fmt.Errorf("checker sandbox image is required")
	}

	workDir := strings.TrimSpace(r.cfg.WorkDir)
	if workDir == "" {
		workDir = "/checker"
	}
	entry := strings.TrimSpace(job.Entry)
	if entry == "" {
		return dockerCheckerContainerSpec{}, fmt.Errorf("checker entry is required")
	}
	if !filepath.IsAbs(entry) {
		entry = filepath.ToSlash(filepath.Join(workDir, entry))
	}

	limits := effectiveLimits(job, r.cfg)
	env := buildCheckerEnv(job)
	cmd := buildCheckerCommand(job.Runtime, entry, job.Args)

	pidsLimit := limits.PidsLimit
	hostCfg := &container.HostConfig{
		NetworkMode:    container.NetworkMode(strings.TrimSpace(r.cfg.NetworkMode)),
		Privileged:     false,
		ReadonlyRootfs: true,
		CapDrop:        []string{"ALL"},
		SecurityOpt:    []string{"no-new-privileges:true"},
		Resources: container.Resources{
			NanoCPUs:  int64(limits.CPUQuota * 1e9),
			Memory:    limits.MemoryBytes,
			PidsLimit: &pidsLimit,
			Ulimits: []*container.Ulimit{
				{Name: "nofile", Soft: limits.NofileLimit, Hard: limits.NofileLimit},
			},
		},
		Mounts: []mount.Mount{
			{
				Type:     mount.TypeBind,
				Source:   hostWorkDir,
				Target:   workDir,
				ReadOnly: true,
			},
		},
		Tmpfs: map[string]string{
			"/tmp": "rw,noexec,nosuid,size=65536k",
		},
	}
	networkDisabled := true
	if strings.TrimSpace(job.NetworkMode) != "" {
		hostCfg.NetworkMode = container.NetworkMode(strings.TrimSpace(job.NetworkMode))
		networkDisabled = false
	} else if strings.TrimSpace(r.cfg.NetworkMode) == "" {
		hostCfg.NetworkMode = container.NetworkMode("none")
	}

	containerCfg := &container.Config{
		Image:           image,
		Cmd:             cmd,
		Env:             env,
		WorkingDir:      workDir,
		User:            strings.TrimSpace(r.cfg.User),
		NetworkDisabled: networkDisabled,
		Labels: map[string]string{
			runtimedomain.ProjectLabelKey:     runtimedomain.ProjectLabelValue,
			runtimedomain.ManagedByLabelKey:   runtimedomain.ManagedByLabelValue,
			runtimedomain.CheckerRoleLabelKey: runtimedomain.CheckerRoleLabelValue,
			"ctf.checker.contest":             fmt.Sprintf("%d", job.Metadata.ContestID),
			"ctf.checker.service":             fmt.Sprintf("%d", job.Metadata.ServiceID),
			"ctf.checker.team":                fmt.Sprintf("%d", job.Metadata.TeamID),
			"ctf.checker.round":               fmt.Sprintf("%d", job.Metadata.RoundNumber),
		},
		AttachStdout: true,
		AttachStderr: true,
	}
	if strings.TrimSpace(r.cfg.User) != "" {
		containerCfg.User = strings.TrimSpace(r.cfg.User)
	}

	return dockerCheckerContainerSpec{
		ContainerConfig: containerCfg,
		HostConfig:      hostCfg,
		NetworkConfig:   &networktypes.NetworkingConfig{},
	}, nil
}

func (r *DockerCheckerRunner) collectLogs(ctx context.Context, containerID string, limit int64) (string, string, bool) {
	logs, err := r.cli.ContainerLogs(ctx, containerID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return "", err.Error(), false
	}
	defer func() {
		_ = logs.Close()
	}()

	stdout := &checkerLimitedBuffer{limit: limit}
	stderr := &checkerLimitedBuffer{limit: limit}
	if _, err := stdcopy.StdCopy(stdout, stderr, logs); err != nil {
		stderr.Write([]byte(err.Error()))
	}
	return stdout.String(), stderr.String(), stdout.exceeded || stderr.exceeded
}

func materializeCheckerFiles(files []contestports.CheckerRunFile, hostWorkRoot string) (string, error) {
	var (
		root string
		err  error
	)
	if strings.TrimSpace(hostWorkRoot) == "" {
		root, err = os.MkdirTemp("", "ctf-checker-*")
	} else {
		if err := os.MkdirAll(hostWorkRoot, 0o755); err != nil {
			return "", err
		}
		root, err = os.MkdirTemp(hostWorkRoot, "ctf-checker-*")
	}
	if err != nil {
		return "", err
	}
	if err := os.Chmod(root, 0o755); err != nil {
		_ = os.RemoveAll(root)
		return "", err
	}
	for _, file := range files {
		rel, err := cleanCheckerFilePath(file.Path)
		if err != nil {
			_ = os.RemoveAll(root)
			return "", err
		}
		target := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			_ = os.RemoveAll(root)
			return "", err
		}
		mode := file.Mode
		if mode == 0 {
			mode = 0o500
		}
		if err := os.WriteFile(target, file.Content, os.FileMode(mode)); err != nil {
			_ = os.RemoveAll(root)
			return "", err
		}
	}
	return root, nil
}

func cleanCheckerFilePath(raw string) (string, error) {
	clean := filepath.Clean(strings.TrimSpace(raw))
	if clean == "." || clean == "" || filepath.IsAbs(clean) || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || clean == ".." {
		return "", fmt.Errorf("invalid checker file path: %q", raw)
	}
	return clean, nil
}

func effectiveLimits(job contestports.CheckerRunJob, cfg config.CheckerSandboxConfig) contestports.CheckerRunLimits {
	limits := job.Limits
	if limits.CPUQuota <= 0 {
		limits.CPUQuota = cfg.CPUQuota
	}
	if limits.CPUQuota <= 0 {
		limits.CPUQuota = 0.5
	}
	if limits.MemoryBytes <= 0 {
		limits.MemoryBytes = cfg.MemoryBytes
	}
	if limits.MemoryBytes <= 0 {
		limits.MemoryBytes = 128 * 1024 * 1024
	}
	if limits.PidsLimit <= 0 {
		limits.PidsLimit = cfg.PidsLimit
	}
	if limits.PidsLimit <= 0 {
		limits.PidsLimit = 64
	}
	if limits.NofileLimit <= 0 {
		limits.NofileLimit = cfg.NofileLimit
	}
	if limits.NofileLimit <= 0 {
		limits.NofileLimit = 128
	}
	if limits.OutputLimitBytes <= 0 {
		limits.OutputLimitBytes = cfg.OutputLimitBytes
	}
	if limits.OutputLimitBytes <= 0 {
		limits.OutputLimitBytes = 32768
	}
	return limits
}

func effectiveOutputLimit(job contestports.CheckerRunJob, cfg config.CheckerSandboxConfig) int64 {
	return effectiveLimits(job, cfg).OutputLimitBytes
}

func buildCheckerEnv(job contestports.CheckerRunJob) []string {
	env := make([]string, 0, len(job.Env)+1)
	keys := make([]string, 0, len(job.Env))
	for key := range job.Env {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		env = append(env, key+"="+job.Env[key])
	}
	if len(job.TargetAllowlist) > 0 {
		env = append(env, "CHECKER_TARGET_ALLOWLIST="+strings.Join(job.TargetAllowlist, ","))
	}
	return env
}

func parseCheckerJSONOutput(stdout string) (contestports.CheckerRunStatus, string, error) {
	var payload struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &payload); err != nil {
		return contestports.CheckerRunStatusFailed, "", err
	}
	switch strings.ToLower(strings.TrimSpace(payload.Status)) {
	case "ok", "passed", "up":
		reason := strings.TrimSpace(payload.Reason)
		if reason == "" {
			reason = contestports.CheckerReasonPassed
		}
		return contestports.CheckerRunStatusOK, reason, nil
	case "failed", "down", "error":
		reason := strings.TrimSpace(payload.Reason)
		if reason == "" {
			reason = contestports.CheckerReasonFailed
		}
		return contestports.CheckerRunStatusFailed, reason, nil
	default:
		return contestports.CheckerRunStatusFailed, "", fmt.Errorf("unknown checker json status: %q", payload.Status)
	}
}

func buildCheckerCommand(runtime, entry string, args []string) []string {
	command := make([]string, 0, len(args)+2)
	if strings.TrimSpace(runtime) != "" {
		command = append(command, strings.TrimSpace(runtime))
	}
	command = append(command, entry)
	command = append(command, args...)
	return command
}
