package container

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
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
			SecurityOpt:    []string{fmt.Sprintf("seccomp=%s", e.containerCfg.Seccomp), "no-new-privileges:true"},
			User:           e.containerCfg.RunAsUser,
		}
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

	resp, err := e.cli.ContainerCreate(ctx, containerCfg, hostCfg, nil, nil, "")
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

func DefaultSecurityConfig(cfg *config.ContainerConfig) *model.SecurityConfig {
	return &model.SecurityConfig{
		ReadonlyRootfs: cfg.ReadonlyRootfs,
		CapDrop:        []string{"ALL"},
		CapAdd:         cfg.AllowedCapabilities,
		SecurityOpt:    []string{fmt.Sprintf("seccomp=%s", cfg.Seccomp), "no-new-privileges:true"},
		User:           cfg.RunAsUser,
	}
}
