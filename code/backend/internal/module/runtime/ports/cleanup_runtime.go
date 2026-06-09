package ports

import (
	"context"
	"time"

	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
)

// ContainerCleanupRuntime 定义实例运行时清理所需的容器运行时能力。
type ContainerCleanupRuntime interface {
	StopContainer(ctx context.Context, containerID string, timeout time.Duration) error
	RemoveContainer(ctx context.Context, containerID string, force bool) error
	RemoveNetwork(ctx context.Context, networkID string) error
	RemoveACLRules(ctx context.Context, rules []runtimecontracts.InstanceRuntimeACLRule) error
	RemoveACL(ctx context.Context, handle *runtimecontracts.InstanceRuntimeACLHandle) error
}
