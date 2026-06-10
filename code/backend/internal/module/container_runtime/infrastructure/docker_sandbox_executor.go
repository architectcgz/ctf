package infrastructure

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	networktypes "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"ctf-platform/internal/config"
	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
	runtimeports "ctf-platform/internal/module/container_runtime/ports"
)

type dockerSandboxClient interface {
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *networktypes.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error)
	CopyToContainer(ctx context.Context, containerID string, dstPath string, content io.Reader, options container.CopyToContainerOptions) error
	ContainerStart(ctx context.Context, container string, options container.StartOptions) error
	ContainerWait(ctx context.Context, container string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error)
	ContainerLogs(ctx context.Context, container string, options container.LogsOptions) (io.ReadCloser, error)
	ContainerInspect(ctx context.Context, container string) (types.ContainerJSON, error)
	ContainerRemove(ctx context.Context, container string, options container.RemoveOptions) error
}

type DockerSandboxExecutor struct {
	cli dockerSandboxClient
	cfg config.CheckerSandboxConfig
}

var _ runtimeports.SandboxExecutor = (*DockerSandboxExecutor)(nil)

type dockerSandboxContainerSpec struct {
	ContainerConfig *container.Config
	HostConfig      *container.HostConfig
	NetworkConfig   *networktypes.NetworkingConfig
}

type sandboxLimitedBuffer struct {
	buf      bytes.Buffer
	limit    int64
	exceeded bool
}

func (b *sandboxLimitedBuffer) Write(p []byte) (int, error) {
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

func (b *sandboxLimitedBuffer) String() string {
	if b == nil {
		return ""
	}
	return b.buf.String()
}

func NewDockerSandboxExecutor(cfg config.CheckerSandboxConfig) (*DockerSandboxExecutor, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return NewDockerSandboxExecutorWithClient(cli, cfg), nil
}

func NewDockerSandboxExecutorWithClient(cli dockerSandboxClient, cfg config.CheckerSandboxConfig) *DockerSandboxExecutor {
	return &DockerSandboxExecutor{cli: cli, cfg: cfg}
}

func (r *DockerSandboxExecutor) RunSandboxExec(ctx context.Context, job runtimeports.SandboxExecJob) (runtimeports.SandboxExecResult, error) {
	startedAt := time.Now()
	result := runtimeports.SandboxExecResult{
		Status:    runtimeports.SandboxExecStatusFailed,
		Reason:    runtimeports.SandboxExecReasonSandboxError,
		StartedAt: startedAt.UTC(),
	}
	finish := func() {
		finishedAt := time.Now()
		result.FinishedAt = finishedAt.UTC()
		result.Duration = finishedAt.Sub(startedAt)
	}
	if r == nil || r.cli == nil {
		finish()
		return result, fmt.Errorf("sandbox executor docker client is not configured")
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

	spec, err := r.buildContainerSpec(job)
	if err != nil {
		finish()
		return result, err
	}

	created, err := r.cli.ContainerCreate(runCtx, spec.ContainerConfig, spec.HostConfig, spec.NetworkConfig, nil, "")
	if err != nil {
		finish()
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

	if err := r.copyCheckerFilesToContainer(runCtx, containerID, spec.ContainerConfig.WorkingDir, job.Files); err != nil {
		finish()
		return result, err
	}

	if err := r.cli.ContainerStart(runCtx, containerID, container.StartOptions{}); err != nil {
		finish()
		return result, err
	}

	waitCh, errCh := r.cli.ContainerWait(runCtx, containerID, container.WaitConditionNotRunning)
	var waitResp container.WaitResponse
	select {
	case waitResp = <-waitCh:
	case err := <-errCh:
		finish()
		return result, err
	case <-runCtx.Done():
		result.Reason = runtimeports.SandboxExecReasonTimeout
		result.ResourceLimitHit = "timeout"
		finish()
		return result, nil
	}

	stdout, stderr, outputLimitHit := r.collectLogs(runCtx, containerID, effectiveSandboxOutputLimit(job, r.cfg))
	result.Stdout = stdout
	result.Stderr = stderr
	result.OutputLimitHit = outputLimitHit
	result.ExitCode = waitResp.StatusCode
	finish()

	if outputLimitHit {
		result.Reason = runtimeports.SandboxExecReasonOutputLimitExceeded
		return result, nil
	}
	if waitResp.StatusCode != 0 {
		result.Reason = runtimeports.SandboxExecReasonFailed
		return result, nil
	}
	if strings.EqualFold(strings.TrimSpace(job.OutputMode), "json") {
		status, reason, err := parseSandboxJSONOutput(stdout)
		if err != nil {
			result.Reason = runtimeports.SandboxExecReasonInvalidOutput
			return result, nil
		}
		result.Status = status
		result.Reason = reason
		return result, nil
	}
	result.Status = runtimeports.SandboxExecStatusOK
	result.Reason = runtimeports.SandboxExecReasonPassed
	return result, nil
}

func (r *DockerSandboxExecutor) buildContainerSpec(job runtimeports.SandboxExecJob) (dockerSandboxContainerSpec, error) {
	image := strings.TrimSpace(job.Image)
	if image == "" {
		image = strings.TrimSpace(r.cfg.Image)
	}
	if image == "" {
		return dockerSandboxContainerSpec{}, fmt.Errorf("sandbox executor image is required")
	}

	workDir := strings.TrimSpace(r.cfg.WorkDir)
	if workDir == "" {
		workDir = "/checker"
	}
	entry := strings.TrimSpace(job.Entry)
	if entry == "" {
		return dockerSandboxContainerSpec{}, fmt.Errorf("checker entry is required")
	}
	if !filepath.IsAbs(entry) {
		entry = filepath.ToSlash(filepath.Join(workDir, entry))
	}

	limits := effectiveSandboxLimits(job, r.cfg)
	env := buildSandboxEnv(job)
	cmd := buildSandboxCommand(job.Runtime, entry, job.Args)

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
		Labels:          runtimecontracts.CheckerSandboxLabels(),
		AttachStdout:    true,
		AttachStderr:    true,
	}
	for key, value := range job.Labels {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		containerCfg.Labels[key] = value
	}
	if strings.TrimSpace(r.cfg.User) != "" {
		containerCfg.User = strings.TrimSpace(r.cfg.User)
	}

	return dockerSandboxContainerSpec{
		ContainerConfig: containerCfg,
		HostConfig:      hostCfg,
		NetworkConfig:   &networktypes.NetworkingConfig{},
	}, nil
}

func (r *DockerSandboxExecutor) collectLogs(ctx context.Context, containerID string, limit int64) (string, string, bool) {
	logs, err := r.cli.ContainerLogs(ctx, containerID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return "", err.Error(), false
	}
	defer func() {
		_ = logs.Close()
	}()

	stdout := &sandboxLimitedBuffer{limit: limit}
	stderr := &sandboxLimitedBuffer{limit: limit}
	if _, err := stdcopy.StdCopy(stdout, stderr, logs); err != nil {
		stderr.Write([]byte(err.Error()))
	}
	return stdout.String(), stderr.String(), stdout.exceeded || stderr.exceeded
}

func (r *DockerSandboxExecutor) copyCheckerFilesToContainer(ctx context.Context, containerID, workDir string, files []runtimeports.SandboxExecFile) error {
	if len(files) == 0 {
		return nil
	}
	if strings.TrimSpace(containerID) == "" {
		return fmt.Errorf("sandbox container id is empty")
	}

	archive, err := buildSandboxFilesArchive(files)
	if err != nil {
		return err
	}

	dstDir := strings.TrimSpace(workDir)
	if dstDir == "" {
		dstDir = "/"
	}
	return r.cli.CopyToContainer(ctx, containerID, dstDir, archive, container.CopyToContainerOptions{})
}

func buildSandboxFilesArchive(files []runtimeports.SandboxExecFile) (io.Reader, error) {
	var archive bytes.Buffer
	tw := tar.NewWriter(&archive)
	createdDirs := make(map[string]struct{})

	for _, file := range files {
		rel, err := cleanSandboxFilePath(file.Path)
		if err != nil {
			return nil, err
		}
		rel = filepath.ToSlash(rel)
		if err := writeSandboxArchiveDirectories(tw, rel, createdDirs); err != nil {
			return nil, err
		}

		mode := file.Mode
		if mode == 0 {
			mode = 0o500
		}
		header := &tar.Header{
			Name: rel,
			Mode: int64(mode),
			Size: int64(len(file.Content)),
		}
		if err := tw.WriteHeader(header); err != nil {
			return nil, err
		}
		if _, err := tw.Write(file.Content); err != nil {
			return nil, err
		}
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	return bytes.NewReader(archive.Bytes()), nil
}

func writeSandboxArchiveDirectories(tw *tar.Writer, filePath string, created map[string]struct{}) error {
	dir := path.Dir(filePath)
	if dir == "." || dir == "" {
		return nil
	}

	current := ""
	for _, segment := range strings.Split(dir, "/") {
		if segment == "" || segment == "." {
			continue
		}
		if current == "" {
			current = segment
		} else {
			current = current + "/" + segment
		}
		if _, exists := created[current]; exists {
			continue
		}
		if err := tw.WriteHeader(&tar.Header{
			Name:     current,
			Typeflag: tar.TypeDir,
			Mode:     0o755,
		}); err != nil {
			return err
		}
		created[current] = struct{}{}
	}
	return nil
}

func cleanSandboxFilePath(raw string) (string, error) {
	clean := filepath.Clean(strings.TrimSpace(raw))
	if clean == "." || clean == "" || filepath.IsAbs(clean) || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || clean == ".." {
		return "", fmt.Errorf("invalid sandbox file path: %q", raw)
	}
	return clean, nil
}

func effectiveSandboxLimits(job runtimeports.SandboxExecJob, cfg config.CheckerSandboxConfig) runtimeports.SandboxExecLimits {
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

func effectiveSandboxOutputLimit(job runtimeports.SandboxExecJob, cfg config.CheckerSandboxConfig) int64 {
	return effectiveSandboxLimits(job, cfg).OutputLimitBytes
}

func buildSandboxEnv(job runtimeports.SandboxExecJob) []string {
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

func parseSandboxJSONOutput(stdout string) (runtimeports.SandboxExecStatus, string, error) {
	var payload struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &payload); err != nil {
		return runtimeports.SandboxExecStatusFailed, "", err
	}
	switch strings.ToLower(strings.TrimSpace(payload.Status)) {
	case "ok", "passed", "up":
		reason := strings.TrimSpace(payload.Reason)
		if reason == "" {
			reason = runtimeports.SandboxExecReasonPassed
		}
		return runtimeports.SandboxExecStatusOK, reason, nil
	case "failed", "down", "error":
		reason := strings.TrimSpace(payload.Reason)
		if reason == "" {
			reason = runtimeports.SandboxExecReasonFailed
		}
		return runtimeports.SandboxExecStatusFailed, reason, nil
	default:
		return runtimeports.SandboxExecStatusFailed, "", fmt.Errorf("unknown checker json status: %q", payload.Status)
	}
}

func buildSandboxCommand(runtime, entry string, args []string) []string {
	command := make([]string, 0, len(args)+2)
	if strings.TrimSpace(runtime) != "" {
		command = append(command, strings.TrimSpace(runtime))
	}
	command = append(command, entry)
	command = append(command, args...)
	return command
}
