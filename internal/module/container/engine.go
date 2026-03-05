package container

import (
	"context"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"

	"ctf-platform/internal/model"
)

// DefaultSecurityConfig 返回默认安全配置
func DefaultSecurityConfig() *model.SecurityConfig {
	return &model.SecurityConfig{
		Privileged:      false,
		ReadonlyRootfs:  false,
		NoNewPrivileges: true,
		CapDrop:         []string{"ALL"},
		CapAdd:          []string{"CHOWN", "SETUID", "SETGID"},
		SecurityOpt:     []string{"no-new-privileges:true"},
		User:            "",
	}
}


// Engine Docker 引擎封装
type Engine struct {
	cli *client.Client
}

// NewEngine 创建 Docker 引擎实例
func NewEngine(host string) (*Engine, error) {
	var cli *client.Client
	var err error

	if host == "" {
		// Unix Socket 连接
		cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	} else {
		// TCP 连接
		cli, err = client.NewClientWithOpts(client.WithHost(host), client.WithAPIVersionNegotiation())
	}

	if err != nil {
		return nil, err
	}

	return &Engine{cli: cli}, nil
}

// HealthCheck 连接健康检查
func (e *Engine) HealthCheck(ctx context.Context) error {
	_, err := e.cli.Ping(ctx)
	return err
}

// CreateContainer 创建容器
func (e *Engine) CreateContainer(ctx context.Context, cfg *model.ContainerConfig) (string, error) {
	// 构建端口映射
	portBindings := nat.PortMap{}
	exposedPorts := nat.PortSet{}
	for containerPort, hostPort := range cfg.Ports {
		port, _ := nat.NewPort("tcp", containerPort)
		portBindings[port] = []nat.PortBinding{{HostPort: hostPort}}
		exposedPorts[port] = struct{}{}
	}

	// 构建资源限制
	resources := container.Resources{}
	if cfg.Resources != nil {
		resources.NanoCPUs = cfg.Resources.CPUQuota * 10000 // 转换为纳秒
		resources.Memory = cfg.Resources.Memory
		resources.PidsLimit = &cfg.Resources.PidsLimit
	}

	// 构建容器配置
	containerCfg := &container.Config{
		Image:        cfg.Image,
		Env:          cfg.Env,
		ExposedPorts: exposedPorts,
	}

	// 构建主机配置
	hostCfg := &container.HostConfig{
		PortBindings: portBindings,
		Resources:    resources,
		NetworkMode:  container.NetworkMode(cfg.Network),
		Privileged:   false, // 强制禁用特权模式
	}

	// 应用安全配置
	if cfg.Security != nil {
		containerCfg.User = cfg.Security.User
		hostCfg.ReadonlyRootfs = cfg.Security.ReadonlyRootfs
		hostCfg.CapDrop = cfg.Security.CapDrop
		hostCfg.CapAdd = cfg.Security.CapAdd
		hostCfg.SecurityOpt = cfg.Security.SecurityOpt

		// 只读根文件系统需要挂载 tmpfs
		if cfg.Security.ReadonlyRootfs {
			hostCfg.Tmpfs = map[string]string{
				"/tmp": "rw,noexec,nosuid,size=65536k",
			}
		}
	}

	// 创建容器
	resp, err := e.cli.ContainerCreate(ctx, containerCfg, hostCfg, nil, nil, "")
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

// StartContainer 启动容器
func (e *Engine) StartContainer(ctx context.Context, containerID string) error {
	return e.cli.ContainerStart(ctx, containerID, container.StartOptions{})
}

// StopContainer 停止容器
func (e *Engine) StopContainer(ctx context.Context, containerID string, timeout time.Duration) error {
	timeoutSec := int(timeout.Seconds())
	return e.cli.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeoutSec})
}

// RemoveContainer 删除容器
func (e *Engine) RemoveContainer(ctx context.Context, containerID string) error {
	return e.cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true})
}

// GetContainerStatus 获取容器状态
func (e *Engine) GetContainerStatus(ctx context.Context, containerID string) (*model.ContainerStatus, error) {
	inspect, err := e.cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, err
	}

	startAt, _ := time.Parse(time.RFC3339Nano, inspect.State.StartedAt)
	return &model.ContainerStatus{
		ID:      inspect.ID,
		State:   inspect.State.Status,
		StartAt: startAt,
	}, nil
}

// PullImage 拉取镜像
func (e *Engine) PullImage(ctx context.Context, imageName string) error {
	reader, err := e.cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()
	_, err = io.Copy(io.Discard, reader)
	return err
}

// ListImages 列出镜像
func (e *Engine) ListImages(ctx context.Context) ([]*model.ImageInfo, error) {
	images, err := e.cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := make([]*model.ImageInfo, 0, len(images))
	for _, img := range images {
		tag := "none"
		if len(img.RepoTags) > 0 {
			tag = img.RepoTags[0]
		}
		result = append(result, &model.ImageInfo{
			ID:      img.ID,
			RepoTag: tag,
			Size:    img.Size,
			Created: time.Unix(img.Created, 0),
		})
	}
	return result, nil
}

// RemoveImage 删除镜像
func (e *Engine) RemoveImage(ctx context.Context, imageID string) error {
	_, err := e.cli.ImageRemove(ctx, imageID, image.RemoveOptions{Force: true})
	return err
}

// Close 关闭客户端连接
func (e *Engine) Close() error {
	return e.cli.Close()
}

// CreateNetwork 创建 Docker 网络
func (e *Engine) CreateNetwork(ctx context.Context, name string) (string, error) {
	resp, err := e.cli.NetworkCreate(ctx, name, types.NetworkCreate{
		Driver: "bridge",
		CheckDuplicate: true,
	})
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

// RemoveNetwork 删除 Docker 网络
func (e *Engine) RemoveNetwork(ctx context.Context, networkID string) error {
	return e.cli.NetworkRemove(ctx, networkID)
}
