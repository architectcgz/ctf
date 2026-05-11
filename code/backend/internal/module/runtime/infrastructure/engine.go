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
	"github.com/docker/docker/api/types/mount"
	networktypes "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"github.com/docker/docker/pkg/stdcopy"
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

type limitedBuffer struct {
	buffer *bytes.Buffer
	limit  int64
}

func (w *limitedBuffer) Write(p []byte) (int, error) {
	if w == nil || w.buffer == nil {
		return len(p), nil
	}
	remaining := int(w.limit) - w.buffer.Len()
	if remaining <= 0 {
		return len(p), nil
	}
	if len(p) > remaining {
		_, _ = w.buffer.Write(p[:remaining])
		return len(p), nil
	}
	_, _ = w.buffer.Write(p)
	return len(p), nil
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
		Cmd:          append([]string(nil), cfg.Command...),
		ExposedPorts: exposedPorts,
		User:         cfg.Security.User,
		Labels:       cfg.Labels,
		WorkingDir:   strings.TrimSpace(cfg.WorkingDir),
	}

	hostCfg := &container.HostConfig{
		PortBindings:   portBindings,
		Resources:      resources,
		Mounts:         buildContainerMounts(cfg.Mounts),
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

	networkCfg := buildContainerNetworkingConfig(cfg.Network, cfg.NetworkAliases)
	resp, err := e.cli.ContainerCreate(ctx, containerCfg, hostCfg, networkCfg, nil, cfg.Name)
	if err != nil {
		if isImageNotFoundError(err) {
			if pullErr := e.pullImage(ctx, cfg.Image); pullErr != nil {
				return "", pullErr
			}
			resp, err = e.cli.ContainerCreate(ctx, containerCfg, hostCfg, networkCfg, nil, cfg.Name)
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

func (e *Engine) CreateNetwork(ctx context.Context, name string, labels map[string]string, internal bool, allowExisting bool) (string, error) {
	resp, err := e.cli.NetworkCreate(ctx, name, networktypes.CreateOptions{
		Labels:   labels,
		Internal: internal,
	})
	if err != nil {
		if errdefs.IsConflict(err) {
			if !allowExisting {
				return "", err
			}
			network, inspectErr := e.cli.NetworkInspect(ctx, name, networktypes.InspectOptions{})
			if inspectErr != nil {
				return "", inspectErr
			}
			if err := validateReusableNetwork(name, labels, internal, network); err != nil {
				return "", err
			}
			return network.ID, nil
		}
		return "", err
	}
	return resp.ID, nil
}

func validateReusableNetwork(name string, labels map[string]string, internal bool, network networktypes.Inspect) error {
	if network.Internal != internal {
		return fmt.Errorf("existing network %q internal=%v does not match requested internal=%v", name, network.Internal, internal)
	}
	for key, expected := range labels {
		if network.Labels[key] != expected {
			return fmt.Errorf("existing network %q is not a managed runtime network", name)
		}
	}
	return nil
}

func buildContainerNetworkingConfig(networkName string, aliases []string) *networktypes.NetworkingConfig {
	networkName = strings.TrimSpace(networkName)
	if networkName == "" {
		return nil
	}
	endpoint := &networktypes.EndpointSettings{}
	if len(aliases) > 0 {
		endpoint.Aliases = append([]string(nil), aliases...)
	}
	return &networktypes.NetworkingConfig{
		EndpointsConfig: map[string]*networktypes.EndpointSettings{
			networkName: endpoint,
		},
	}
}

func buildContainerMounts(mounts []model.ContainerMount) []mount.Mount {
	if len(mounts) == 0 {
		return nil
	}
	result := make([]mount.Mount, 0, len(mounts))
	for _, item := range mounts {
		source := strings.TrimSpace(item.Source)
		target := strings.TrimSpace(item.Target)
		if source == "" || target == "" {
			continue
		}
		result = append(result, mount.Mount{
			Type:     mount.TypeVolume,
			Source:   source,
			Target:   target,
			ReadOnly: item.ReadOnly,
		})
	}
	if len(result) == 0 {
		return nil
	}
	return result
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
	if e == nil || e.cli == nil {
		return fmt.Errorf("runtime engine is not configured")
	}
	if strings.TrimSpace(containerID) == "" {
		return fmt.Errorf("container id is empty")
	}

	resolvedPath, err := e.resolveContainerFilePath(ctx, containerID, filePath)
	if err != nil {
		return err
	}

	dir := path.Dir(resolvedPath)
	if dir == "." || dir == "" {
		dir = "/"
	}

	var archive bytes.Buffer
	tw := tar.NewWriter(&archive)
	header := &tar.Header{
		Name: path.Base(resolvedPath),
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

func (e *Engine) ReadFileFromContainer(ctx context.Context, containerID, filePath string, limit int64) ([]byte, error) {
	if e == nil || e.cli == nil {
		return nil, fmt.Errorf("runtime engine is not configured")
	}
	if strings.TrimSpace(containerID) == "" {
		return nil, fmt.Errorf("container id is empty")
	}
	if limit <= 0 {
		limit = 256 * 1024
	}

	resolvedPath, err := e.resolveContainerFilePath(ctx, containerID, filePath)
	if err != nil {
		return nil, err
	}

	reader, _, err := e.cli.CopyFromContainer(ctx, containerID, resolvedPath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	tr := tar.NewReader(reader)
	for {
		header, err := tr.Next()
		if err != nil {
			return nil, err
		}
		if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeRegA {
			continue
		}
		if header.Size > limit {
			return nil, fmt.Errorf("file exceeds limit")
		}
		var content bytes.Buffer
		if _, err := io.CopyN(&content, tr, limit+1); err != nil && err != io.EOF {
			return nil, err
		}
		if int64(content.Len()) > limit {
			return nil, fmt.Errorf("file exceeds limit")
		}
		return content.Bytes(), nil
	}
}

func (e *Engine) ListDirectoryFromContainer(ctx context.Context, containerID, dirPath string, limit int) ([]runtimeports.ContainerDirectoryEntry, error) {
	if e == nil || e.cli == nil {
		return nil, fmt.Errorf("runtime engine is not configured")
	}
	if strings.TrimSpace(containerID) == "" {
		return nil, fmt.Errorf("container id is empty")
	}
	if limit <= 0 {
		limit = 300
	}

	resolvedPath, err := e.resolveContainerFilePath(ctx, containerID, dirPath)
	if err != nil {
		return nil, err
	}

	reader, _, err := e.cli.CopyFromContainer(ctx, containerID, resolvedPath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	rootName := path.Base(path.Clean(resolvedPath))
	entriesByName := make(map[string]runtimeports.ContainerDirectoryEntry)
	tr := tar.NewReader(reader)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		name, entryType, ok := containerDirectoryEntryFromTar(rootName, header)
		if !ok {
			continue
		}
		entry := runtimeports.ContainerDirectoryEntry{
			Name: name,
			Type: entryType,
			Size: header.Size,
		}
		if existing, exists := entriesByName[name]; !exists || existing.Type != "dir" {
			entriesByName[name] = entry
		}
		if len(entriesByName) >= limit {
			break
		}
	}

	entries := make([]runtimeports.ContainerDirectoryEntry, 0, len(entriesByName))
	for _, entry := range entriesByName {
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Type != entries[j].Type {
			return entries[i].Type == "dir"
		}
		return entries[i].Name < entries[j].Name
	})
	return entries, nil
}

func (e *Engine) ExecContainerCommand(ctx context.Context, containerID string, command []string, stdin []byte, limit int64) ([]byte, error) {
	if e == nil || e.cli == nil {
		return nil, fmt.Errorf("runtime engine is not configured")
	}
	if strings.TrimSpace(containerID) == "" {
		return nil, fmt.Errorf("container id is empty")
	}
	if len(command) == 0 {
		return nil, fmt.Errorf("command is empty")
	}
	if limit <= 0 {
		limit = 64 * 1024
	}
	workingDir, err := e.inspectContainerWorkingDir(ctx, containerID)
	if err != nil {
		return nil, err
	}

	execID, err := e.cli.ContainerExecCreate(ctx, containerID, container.ExecOptions{
		AttachStdin:  len(stdin) > 0,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
		Cmd:          command,
		WorkingDir:   workingDir,
	})
	if err != nil {
		return nil, err
	}

	attach, err := e.cli.ContainerExecAttach(ctx, execID.ID, container.ExecAttachOptions{Tty: false})
	if err != nil {
		return nil, err
	}
	defer attach.Close()

	if len(stdin) > 0 {
		go func() {
			_, _ = attach.Conn.Write(stdin)
			_ = attach.CloseWrite()
		}()
	}

	var output bytes.Buffer
	limited := &limitedBuffer{buffer: &output, limit: limit}
	if _, err := stdcopy.StdCopy(limited, limited, attach.Reader); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func (e *Engine) resolveContainerFilePath(ctx context.Context, containerID, filePath string) (string, error) {
	workingDir, err := e.inspectContainerWorkingDir(ctx, containerID)
	if err != nil {
		return "", err
	}
	return resolveContainerFilePath(workingDir, filePath), nil
}

func (e *Engine) inspectContainerWorkingDir(ctx context.Context, containerID string) (string, error) {
	info, err := e.cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return "", err
	}
	if info.Config == nil {
		return "", nil
	}
	return info.Config.WorkingDir, nil
}

func resolveContainerFilePath(workingDir, filePath string) string {
	cleanFilePath := path.Clean(filePath)
	if path.IsAbs(cleanFilePath) {
		return cleanFilePath
	}

	base := strings.TrimSpace(workingDir)
	if base == "" {
		base = "/"
	}
	if !path.IsAbs(base) {
		base = "/" + base
	}
	return path.Join(path.Clean(base), cleanFilePath)
}

func containerDirectoryEntryFromTar(rootName string, header *tar.Header) (string, string, bool) {
	if header == nil {
		return "", "", false
	}
	name := strings.Trim(path.Clean(header.Name), "/")
	if name == "" || name == "." || name == rootName {
		return "", "", false
	}
	if rootName != "." && strings.HasPrefix(name, rootName+"/") {
		name = strings.TrimPrefix(name, rootName+"/")
	}
	parts := strings.Split(name, "/")
	if len(parts) == 0 || parts[0] == "" || parts[0] == "." {
		return "", "", false
	}

	entryType := "file"
	if len(parts) > 1 || header.Typeflag == tar.TypeDir {
		entryType = "dir"
	} else if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeRegA {
		entryType = "other"
	}
	return parts[0], entryType, true
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
			filters.Arg("label", runtimedomain.ProjectFilter()),
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

func (e *Engine) InspectManagedContainer(ctx context.Context, containerID string) (*runtimeports.ManagedContainerState, error) {
	if e == nil || e.cli == nil {
		return nil, fmt.Errorf("runtime engine is not configured")
	}
	if strings.TrimSpace(containerID) == "" {
		return &runtimeports.ManagedContainerState{Exists: false}, nil
	}

	resp, err := e.cli.ContainerInspect(ctx, containerID)
	if err != nil {
		if errdefs.IsNotFound(err) || strings.Contains(strings.ToLower(err.Error()), "no such container") {
			return &runtimeports.ManagedContainerState{
				ID:     containerID,
				Exists: false,
			}, nil
		}
		return nil, err
	}

	state := &runtimeports.ManagedContainerState{
		ID:     resp.ID,
		Exists: true,
	}
	if resp.State != nil {
		state.Running = resp.State.Running
		state.Status = resp.State.Status
	}
	return state, nil
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
