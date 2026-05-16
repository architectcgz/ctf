package infrastructure

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	networktypes "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/errdefs"
	"github.com/docker/go-connections/nat"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
)

func (e *Engine) CreateContainer(ctx context.Context, cfg *model.ContainerConfig) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("container config is nil")
	}

	cli, err := e.requireClient()
	if err != nil {
		return "", err
	}

	resourceLimits, err := resolveContainerResourceLimits(cfg.Resources, e.containerDefaults())
	if err != nil {
		return "", err
	}
	securityCfg := resolveContainerSecurityConfig(cfg.Security, e.containerDefaults())

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

	resources := container.Resources{
		NanoCPUs:  int64(resourceLimits.CPUQuota * 1e9),
		Memory:    resourceLimits.Memory,
		PidsLimit: &resourceLimits.PidsLimit,
	}

	containerCfg := &container.Config{
		Image:        cfg.Image,
		Env:          cfg.Env,
		Cmd:          append([]string(nil), cfg.Command...),
		ExposedPorts: exposedPorts,
		User:         securityCfg.User,
		Labels:       cfg.Labels,
		WorkingDir:   strings.TrimSpace(cfg.WorkingDir),
	}

	hostCfg := &container.HostConfig{
		PortBindings:   portBindings,
		Resources:      resources,
		Mounts:         buildContainerMounts(cfg.Mounts),
		NetworkMode:    container.NetworkMode(cfg.Network),
		Privileged:     false,
		ReadonlyRootfs: securityCfg.ReadonlyRootfs,
		CapDrop:        securityCfg.CapDrop,
		CapAdd:         securityCfg.CapAdd,
		SecurityOpt:    securityCfg.SecurityOpt,
	}

	if securityCfg.ReadonlyRootfs {
		hostCfg.Tmpfs = map[string]string{
			"/tmp": "rw,noexec,nosuid,size=65536k",
		}
	}

	networkCfg := buildContainerNetworkingConfig(cfg.Network, cfg.NetworkAliases)
	resp, err := cli.ContainerCreate(ctx, containerCfg, hostCfg, networkCfg, nil, cfg.Name)
	if err != nil {
		if isImageNotFoundError(err) {
			if pullErr := e.pullImage(ctx, cfg.Image); pullErr != nil {
				return "", pullErr
			}
			resp, err = cli.ContainerCreate(ctx, containerCfg, hostCfg, networkCfg, nil, cfg.Name)
			if err != nil {
				return "", normalizeContainerCreateError(err)
			}
			return resp.ID, nil
		}
		return "", normalizeContainerCreateError(err)
	}
	return resp.ID, nil
}

func (e *Engine) ResolveServicePort(ctx context.Context, imageRef string, preferredPort int) (int, error) {
	cli, err := e.requireClient()
	if err != nil {
		return 0, err
	}

	if err := e.ensureImagePresent(ctx, imageRef); err != nil {
		return 0, err
	}

	resp, _, err := cli.ImageInspectWithRaw(ctx, imageRef)
	if err != nil {
		return 0, err
	}
	if resp.Config == nil {
		return preferredPort, nil
	}

	return selectServicePort(resp.Config.ExposedPorts, preferredPort), nil
}

func (e *Engine) CreateNetwork(ctx context.Context, name string, labels map[string]string, internal bool, allowExisting bool) (string, error) {
	cli, err := e.requireClient()
	if err != nil {
		return "", err
	}

	resp, err := cli.NetworkCreate(ctx, name, networktypes.CreateOptions{
		Labels:   labels,
		Internal: internal,
	})
	if err != nil {
		if errdefs.IsConflict(err) {
			if !allowExisting {
				return "", err
			}
			network, inspectErr := cli.NetworkInspect(ctx, name, networktypes.InspectOptions{})
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
	cli, err := e.requireClient()
	if err != nil {
		return err
	}
	return cli.NetworkConnect(ctx, networkName, containerID, nil)
}

func (e *Engine) StartContainer(ctx context.Context, containerID string) error {
	cli, err := e.requireClient()
	if err != nil {
		return err
	}
	return cli.ContainerStart(ctx, containerID, container.StartOptions{})
}

func (e *Engine) InspectContainerNetworkIPs(ctx context.Context, containerID string) (map[string]string, error) {
	cli, err := e.requireClient()
	if err != nil {
		return nil, err
	}

	resp, err := cli.ContainerInspect(ctx, containerID)
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
	cli, err := e.requireClient()
	if err != nil {
		return err
	}
	timeoutSeconds := int(timeout.Seconds())
	return cli.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeoutSeconds})
}

func (e *Engine) RemoveContainer(ctx context.Context, containerID string, force bool) error {
	cli, err := e.requireClient()
	if err != nil {
		return err
	}
	return normalizeContainerNotFoundError(cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: force}))
}

func (e *Engine) RemoveNetwork(ctx context.Context, networkID string) error {
	cli, err := e.requireClient()
	if err != nil {
		return err
	}
	return normalizeNetworkNotFoundError(cli.NetworkRemove(ctx, networkID))
}

func (e *Engine) ApplyACLRules(ctx context.Context, rules []model.InstanceRuntimeACLRule) error {
	return applyACLRules(ctx, rules)
}

func (e *Engine) RemoveACLRules(ctx context.Context, rules []model.InstanceRuntimeACLRule) error {
	return removeACLRules(ctx, rules)
}

func (e *Engine) ensureImagePresent(ctx context.Context, imageRef string) error {
	cli, err := e.requireClient()
	if err != nil {
		return err
	}
	if strings.TrimSpace(imageRef) == "" {
		return fmt.Errorf("image ref is empty")
	}
	_, _, err = cli.ImageInspectWithRaw(ctx, imageRef)
	if err == nil {
		return nil
	}
	if !isImageNotFoundError(err) {
		return err
	}
	if pullErr := e.pullImage(ctx, imageRef); pullErr != nil {
		return pullErr
	}
	_, _, err = cli.ImageInspectWithRaw(ctx, imageRef)
	return err
}

func (e *Engine) pullImage(ctx context.Context, imageRef string) error {
	cli, err := e.requireClient()
	if err != nil {
		return err
	}

	options := image.PullOptions{}
	options.RegistryAuth = buildImagePullRegistryAuth(imageRef, e.containerDefaults().Registry)

	reader, err := cli.ImagePull(ctx, imageRef, options)
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
