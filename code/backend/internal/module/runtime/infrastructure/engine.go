package infrastructure

import (
	"bytes"
	"fmt"

	"github.com/docker/docker/client"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
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

func (e *Engine) requireClient() (*client.Client, error) {
	if e == nil || e.cli == nil {
		return nil, runtimeports.ErrRuntimeEngineUnavailable
	}
	return e.cli, nil
}

func (e *Engine) containerDefaults() *config.ContainerConfig {
	if e == nil || e.containerCfg == nil {
		return &config.ContainerConfig{}
	}
	return e.containerCfg
}

func DefaultSecurityConfig(cfg *config.ContainerConfig) *model.SecurityConfig {
	if cfg == nil {
		cfg = &config.ContainerConfig{}
	}
	return &model.SecurityConfig{
		ReadonlyRootfs: cfg.ReadonlyRootfs,
		CapDrop:        []string{"ALL"},
		CapAdd:         append([]string(nil), cfg.AllowedCapabilities...),
		SecurityOpt:    buildSecurityOpts(cfg.Seccomp),
		User:           cfg.RunAsUser,
	}
}

func resolveContainerResourceLimits(input *model.ResourceLimits, cfg *config.ContainerConfig) (*model.ResourceLimits, error) {
	if cfg == nil {
		cfg = &config.ContainerConfig{}
	}

	resolved := &model.ResourceLimits{}
	if input == nil {
		resolved.CPUQuota = cfg.DefaultCPUQuota
		resolved.Memory = cfg.DefaultMemory
		resolved.PidsLimit = cfg.DefaultPidsLimit
	} else {
		resolved.CPUQuota = input.CPUQuota
		resolved.Memory = input.Memory
		resolved.PidsLimit = input.PidsLimit
	}

	if err := validateContainerResourceLimits(resolved); err != nil {
		return nil, err
	}
	return resolved, nil
}

func validateContainerResourceLimits(limits *model.ResourceLimits) error {
	if limits == nil {
		return nil
	}
	if limits.CPUQuota <= 0 || limits.CPUQuota > 16 {
		return fmt.Errorf("invalid cpu quota: %f", limits.CPUQuota)
	}
	if limits.Memory < 64*1024*1024 || limits.Memory > 16*1024*1024*1024 {
		return fmt.Errorf("invalid memory: %d", limits.Memory)
	}
	if limits.PidsLimit <= 0 || limits.PidsLimit > 10000 {
		return fmt.Errorf("invalid pids limit: %d", limits.PidsLimit)
	}
	return nil
}

func resolveContainerSecurityConfig(input *model.SecurityConfig, cfg *config.ContainerConfig) *model.SecurityConfig {
	if input == nil {
		return DefaultSecurityConfig(cfg)
	}
	return &model.SecurityConfig{
		ReadonlyRootfs: input.ReadonlyRootfs,
		CapDrop:        append([]string(nil), input.CapDrop...),
		CapAdd:         append([]string(nil), input.CapAdd...),
		SecurityOpt:    append([]string(nil), input.SecurityOpt...),
		User:           input.User,
	}
}
