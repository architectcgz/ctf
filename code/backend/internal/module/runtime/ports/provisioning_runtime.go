package ports

import (
	"context"
	"time"

	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
)

// ContainerProvisioningRuntime 定义拓扑创建与资源回滚所需的容器运行时能力。
type ContainerProvisioningRuntime interface {
	CreateNetwork(ctx context.Context, name string, labels map[string]string, internal bool, allowExisting bool, subnet string) (string, error)
	ListNetworkSubnets(ctx context.Context) ([]string, error)
	CreateContainer(ctx context.Context, cfg *runtimecontracts.ContainerConfig) (string, error)
	ResolveServicePort(ctx context.Context, imageRef string, preferredPort int) (int, error)
	ConnectContainerToNetwork(ctx context.Context, containerID, networkName string) error
	InspectContainerNetworkIPs(ctx context.Context, containerID string) (map[string]string, error)
	StartContainer(ctx context.Context, containerID string) error
	StopContainer(ctx context.Context, containerID string, timeout time.Duration) error
	RemoveContainer(ctx context.Context, containerID string, force bool) error
	RemoveNetwork(ctx context.Context, networkID string) error
	ApplyACLRules(ctx context.Context, rules []runtimecontracts.InstanceRuntimeACLRule) error
	ApplyACL(ctx context.Context, handle *runtimecontracts.InstanceRuntimeACLHandle, rules []runtimecontracts.InstanceRuntimeACLRule) error
}
