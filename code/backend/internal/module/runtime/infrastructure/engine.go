package infrastructure

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	networktypes "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"github.com/docker/go-connections/nat"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	runtimedomain "ctf-platform/internal/module/runtime/domain"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

type Engine struct {
	cli          *client.Client
	containerCfg *config.ContainerConfig
}

func NewEngine(cfg *config.ContainerConfig) (*Engine, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &Engine{
		cli:          cli,
		containerCfg: cfg,
	}, nil
}

func (e *Engine) CreateContainer(ctx context.Context, cfg *model.ContainerConfig) (string, error) {
	// 运行时参数校验
	if cfg.Resources != nil {
		if cfg.Resources.CPUQuota <= 0 || cfg.Resources.CPUQuota > 16 {
			return "", fmt.Errorf("invalid cpu quota: %f", cfg.Resources.CPUQuota)
		}
		if cfg.Resources.Memory < 64*1024*1024 || cfg.Resources.Memory > 16*1024*1024*1024 {
			return "", fmt.Errorf("invalid memory: %d", cfg.Resources.Memory)
		}
		if cfg.Resources.PidsLimit <= 0 || cfg.Resources.PidsLimit > 10000 {
			return "", fmt.Errorf("invalid pids limit: %d", cfg.Resources.PidsLimit)
		}
	}

	if cfg.Resources == nil {
		cfg.Resources = &model.ResourceLimits{
			CPUQuota:  e.containerCfg.DefaultCPUQuota,
			Memory:    e.containerCfg.DefaultMemory,
			PidsLimit: e.containerCfg.DefaultPidsLimit,
		}
	}

	if cfg.Security == nil {
		cfg.Security = &model.SecurityConfig{
			ReadonlyRootfs: e.containerCfg.ReadonlyRootfs,
			CapDrop:        []string{"ALL"},
			CapAdd:         e.containerCfg.AllowedCapabilities,
			SecurityOpt:    buildSecurityOpts(e.containerCfg.Seccomp),
			User:           e.containerCfg.RunAsUser,
		}
	}

	if err := e.ensureImagePresent(ctx, cfg.Image); err != nil {
		return "", err
	}

	portBindings := nat.PortMap{}
	exposedPorts := nat.PortSet{}
	for containerPort, hostPort := range cfg.Ports {
		port, _ := nat.NewPort("tcp", containerPort)
		portBindings[port] = []nat.PortBinding{{HostPort: hostPort}}
		exposedPorts[port] = struct{}{}
	}

	// CPU 配额转换：核心数 → 纳秒（1 核 = 1e9 纳秒）
	resources := container.Resources{
		NanoCPUs:  int64(cfg.Resources.CPUQuota * 1e9),
		Memory:    cfg.Resources.Memory,
		PidsLimit: &cfg.Resources.PidsLimit,
	}

	containerCfg := &container.Config{
		Image:        cfg.Image,
		Env:          cfg.Env,
		ExposedPorts: exposedPorts,
		User:         cfg.Security.User,
		Labels:       cfg.Labels,
	}

	hostCfg := &container.HostConfig{
		PortBindings:   portBindings,
		Resources:      resources,
		NetworkMode:    container.NetworkMode(cfg.Network),
		Privileged:     false,
		ReadonlyRootfs: cfg.Security.ReadonlyRootfs,
		CapDrop:        cfg.Security.CapDrop,
		CapAdd:         cfg.Security.CapAdd,
		SecurityOpt:    cfg.Security.SecurityOpt,
	}

	if cfg.Security.ReadonlyRootfs {
		hostCfg.Tmpfs = map[string]string{
			"/tmp": "rw,noexec,nosuid,size=65536k",
		}
	}

	resp, err := e.cli.ContainerCreate(ctx, containerCfg, hostCfg, nil, nil, cfg.Name)
	if err != nil {
		if isImageNotFoundError(err) {
			if pullErr := e.pullImage(ctx, cfg.Image); pullErr != nil {
				return "", pullErr
			}
			resp, err = e.cli.ContainerCreate(ctx, containerCfg, hostCfg, nil, nil, cfg.Name)
			if err != nil {
				return "", err
			}
			return resp.ID, nil
		}
		return "", err
	}
	return resp.ID, nil
}

func (e *Engine) ResolveServicePort(ctx context.Context, imageRef string, preferredPort int) (int, error) {
	if err := e.ensureImagePresent(ctx, imageRef); err != nil {
		return 0, err
	}

	resp, _, err := e.cli.ImageInspectWithRaw(ctx, imageRef)
	if err != nil {
		return 0, err
	}
	if resp.Config == nil {
		return preferredPort, nil
	}

	return selectServicePort(resp.Config.ExposedPorts, preferredPort), nil
}

func (e *Engine) CreateNetwork(ctx context.Context, name string, labels map[string]string, internal bool) (string, error) {
	resp, err := e.cli.NetworkCreate(ctx, name, networktypes.CreateOptions{
		Labels:   labels,
		Internal: internal,
	})
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (e *Engine) ConnectContainerToNetwork(ctx context.Context, containerID, networkName string) error {
	return e.cli.NetworkConnect(ctx, networkName, containerID, nil)
}

func (e *Engine) StartContainer(ctx context.Context, containerID string) error {
	return e.cli.ContainerStart(ctx, containerID, container.StartOptions{})
}

func (e *Engine) InspectContainerNetworkIPs(ctx context.Context, containerID string) (map[string]string, error) {
	resp, err := e.cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	if resp.NetworkSettings == nil {
		return result, nil
	}
	for networkName, settings := range resp.NetworkSettings.Networks {
		if settings == nil || settings.IPAddress == "" {
			continue
		}
		result[networkName] = settings.IPAddress
	}
	return result, nil
}

func (e *Engine) StopContainer(ctx context.Context, containerID string, timeout time.Duration) error {
	timeoutSeconds := int(timeout.Seconds())
	return e.cli.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeoutSeconds})
}

func (e *Engine) RemoveContainer(ctx context.Context, containerID string, force bool) error {
	return e.cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: force})
}

func (e *Engine) RemoveNetwork(ctx context.Context, networkID string) error {
	return e.cli.NetworkRemove(ctx, networkID)
}

func (e *Engine) ApplyACLRules(ctx context.Context, rules []model.InstanceRuntimeACLRule) error {
	return applyACLRules(ctx, rules)
}

func (e *Engine) RemoveACLRules(ctx context.Context, rules []model.InstanceRuntimeACLRule) error {
	return removeACLRules(ctx, rules)
}

func (e *Engine) WriteFileToContainer(ctx context.Context, containerID, filePath string, content []byte) error {
	dir := path.Dir(filePath)
	if dir == "." || dir == "" {
		dir = "/"
	}

	var archive bytes.Buffer
	tw := tar.NewWriter(&archive)
	header := &tar.Header{
		Name: path.Base(filePath),
		Mode: 0o644,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	if _, err := tw.Write(content); err != nil {
		return err
	}
	if err := tw.Close(); err != nil {
		return err
	}

	return e.cli.CopyToContainer(ctx, containerID, dir, io.NopCloser(bytes.NewReader(archive.Bytes())), container.CopyToContainerOptions{})
}

func (e *Engine) ExecContainerInteractive(ctx context.Context, containerID string, command []string, stdin io.Reader, stdout io.Writer) error {
	if e == nil || e.cli == nil {
		return fmt.Errorf("runtime engine is not configured")
	}
	if strings.TrimSpace(containerID) == "" {
		return fmt.Errorf("container id is empty")
	}
	if len(command) == 0 {
		command = []string{"/bin/sh"}
	}

	execID, err := e.cli.ContainerExecCreate(ctx, containerID, container.ExecOptions{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          command,
	})
	if err != nil {
		return err
	}

	attach, err := e.cli.ContainerExecAttach(ctx, execID.ID, container.ExecAttachOptions{Tty: true})
	if err != nil {
		return err
	}
	defer attach.Close()

	copyErr := make(chan error, 2)
	go func() {
		_, err := io.Copy(attach.Conn, stdin)
		_ = attach.CloseWrite()
		copyErr <- err
	}()
	go func() {
		_, err := io.Copy(stdout, attach.Reader)
		copyErr <- err
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-copyErr:
		if err != nil && err != io.EOF {
			return err
		}
		return nil
	}
}

func (e *Engine) ListManagedContainers(ctx context.Context) ([]runtimeports.ManagedContainer, error) {
	containers, err := e.cli.ContainerList(ctx, container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", runtimedomain.ManagedByFilter()),
		),
	})
	if err != nil {
		return nil, err
	}

	items := make([]runtimeports.ManagedContainer, 0, len(containers))
	for _, item := range containers {
		name := item.ID[:12]
		if len(item.Names) > 0 {
			name = item.Names[0]
		}
		items = append(items, runtimeports.ManagedContainer{
			ID:        item.ID,
			Name:      name,
			CreatedAt: time.Unix(item.Created, 0),
		})
	}
	return items, nil
}

func DefaultSecurityConfig(cfg *config.ContainerConfig) *model.SecurityConfig {
	return &model.SecurityConfig{
		ReadonlyRootfs: cfg.ReadonlyRootfs,
		CapDrop:        []string{"ALL"},
		CapAdd:         cfg.AllowedCapabilities,
		SecurityOpt:    buildSecurityOpts(cfg.Seccomp),
		User:           cfg.RunAsUser,
	}
}

func (e *Engine) ensureImagePresent(ctx context.Context, imageRef string) error {
	if strings.TrimSpace(imageRef) == "" {
		return fmt.Errorf("image ref is empty")
	}
	_, _, err := e.cli.ImageInspectWithRaw(ctx, imageRef)
	if err == nil {
		return nil
	}
	if !isImageNotFoundError(err) {
		return err
	}
	if pullErr := e.pullImage(ctx, imageRef); pullErr != nil {
		return pullErr
	}
	_, _, err = e.cli.ImageInspectWithRaw(ctx, imageRef)
	return err
}

func (e *Engine) pullImage(ctx context.Context, imageRef string) error {
	options := image.PullOptions{}
	if e != nil && e.containerCfg != nil {
		options.RegistryAuth = buildImagePullRegistryAuth(imageRef, e.containerCfg.Registry)
	}
	reader, err := e.cli.ImagePull(ctx, imageRef, options)
	if err != nil {
		return err
	}
	defer reader.Close()
	_, _ = io.Copy(io.Discard, reader)
	return nil
}

func buildImagePullRegistryAuth(imageRef string, cfg config.ContainerRegistryConfig) string {
	if !cfg.Enabled {
		return ""
	}

	configuredServer := normalizeRegistryServer(cfg.Server)
	if configuredServer == "" || imageRegistryServer(imageRef) != configuredServer {
		return ""
	}

	authConfig := registry.AuthConfig{
		Username:      strings.TrimSpace(cfg.Username),
		Password:      strings.TrimSpace(cfg.Password),
		IdentityToken: strings.TrimSpace(cfg.IdentityToken),
		ServerAddress: configuredServer,
	}
	payload, err := json.Marshal(authConfig)
	if err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(payload)
}

func normalizeRegistryServer(server string) string {
	normalized := strings.TrimSpace(server)
	normalized = strings.TrimPrefix(normalized, "https://")
	normalized = strings.TrimPrefix(normalized, "http://")
	normalized = strings.Trim(normalized, "/")
	if slash := strings.Index(normalized, "/"); slash >= 0 {
		normalized = normalized[:slash]
	}
	return normalized
}

func imageRegistryServer(imageRef string) string {
	ref := strings.TrimSpace(imageRef)
	firstSlash := strings.Index(ref, "/")
	if firstSlash < 0 {
		return "docker.io"
	}
	firstPart := ref[:firstSlash]
	if strings.Contains(firstPart, ".") || strings.Contains(firstPart, ":") || firstPart == "localhost" {
		return firstPart
	}
	return "docker.io"
}

func isImageNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	if errdefs.IsNotFound(err) {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "no such image")
}

func buildSecurityOpts(seccomp string) []string {
	opts := []string{"no-new-privileges:true"}

	normalized := strings.TrimSpace(seccomp)
	switch normalized {
	case "", "default":
		return opts
	default:
		return append([]string{fmt.Sprintf("seccomp=%s", normalized)}, opts...)
	}
}

func selectServicePort(exposedPorts nat.PortSet, preferredPort int) int {
	if len(exposedPorts) == 0 {
		return preferredPort
	}

	available := make([]int, 0, len(exposedPorts))
	preferredKey := strconv.Itoa(preferredPort) + "/tcp"
	for port := range exposedPorts {
		if string(port) == preferredKey {
			return preferredPort
		}

		if port.Proto() != "tcp" {
			continue
		}
		portValue, err := strconv.Atoi(port.Port())
		if err != nil || portValue <= 0 {
			continue
		}
		available = append(available, portValue)
	}

	if len(available) == 0 {
		return preferredPort
	}

	sort.Ints(available)
	for _, candidate := range available {
		if candidate == 80 || candidate == 443 {
			return candidate
		}
	}
	return available[0]
}
