package container

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"

	"ctf-platform/internal/model"
)

type Engine struct {
	cli *client.Client
}

func NewEngine() (*Engine, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &Engine{cli: cli}, nil
}

func (e *Engine) CreateContainer(ctx context.Context, cfg *model.ContainerConfig) (string, error) {
	portBindings := nat.PortMap{}
	exposedPorts := nat.PortSet{}
	for containerPort, hostPort := range cfg.Ports {
		port, _ := nat.NewPort("tcp", containerPort)
		portBindings[port] = []nat.PortBinding{{HostPort: hostPort}}
		exposedPorts[port] = struct{}{}
	}

	resources := container.Resources{}
	if cfg.Resources != nil {
		resources.NanoCPUs = cfg.Resources.CPUQuota * 10000
		resources.Memory = cfg.Resources.Memory
		resources.PidsLimit = &cfg.Resources.PidsLimit
	}

	containerCfg := &container.Config{
		Image:        cfg.Image,
		Env:          cfg.Env,
		ExposedPorts: exposedPorts,
	}

	hostCfg := &container.HostConfig{
		PortBindings: portBindings,
		Resources:    resources,
		NetworkMode:  container.NetworkMode(cfg.Network),
		Privileged:   false,
	}

	if cfg.Security != nil {
		containerCfg.User = cfg.Security.User
		hostCfg.ReadonlyRootfs = cfg.Security.ReadonlyRootfs
		hostCfg.CapDrop = cfg.Security.CapDrop
		hostCfg.CapAdd = cfg.Security.CapAdd
		hostCfg.SecurityOpt = cfg.Security.SecurityOpt

		if cfg.Security.ReadonlyRootfs {
			hostCfg.Tmpfs = map[string]string{
				"/tmp": "rw,noexec,nosuid,size=65536k",
			}
		}
	}

	resp, err := e.cli.ContainerCreate(ctx, containerCfg, hostCfg, nil, nil, "")
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

func DefaultSecurityConfig() *model.SecurityConfig {
	return &model.SecurityConfig{
		ReadonlyRootfs: false,
		CapDrop:        []string{"ALL"},
		CapAdd:         []string{"CHOWN", "SETUID", "SETGID"},
		SecurityOpt:    []string{"no-new-privileges:true"},
	}
}
