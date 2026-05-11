package ports

import (
	"context"
	"io"
	"time"

	"ctf-platform/internal/model"
)

// ContainerProvisioningRuntime 定义拓扑创建与资源回滚所需的容器运行时能力。
type ContainerProvisioningRuntime interface {
	CreateNetwork(ctx context.Context, name string, labels map[string]string, internal bool, allowExisting bool) (string, error)
	CreateContainer(ctx context.Context, cfg *model.ContainerConfig) (string, error)
	ResolveServicePort(ctx context.Context, imageRef string, preferredPort int) (int, error)
	ConnectContainerToNetwork(ctx context.Context, containerID, networkName string) error
	InspectContainerNetworkIPs(ctx context.Context, containerID string) (map[string]string, error)
	StartContainer(ctx context.Context, containerID string) error
	StopContainer(ctx context.Context, containerID string, timeout time.Duration) error
	RemoveContainer(ctx context.Context, containerID string, force bool) error
	RemoveNetwork(ctx context.Context, networkID string) error
	ApplyACLRules(ctx context.Context, rules []model.InstanceRuntimeACLRule) error
}

// ContainerCleanupRuntime 定义实例运行时清理所需的容器运行时能力。
type ContainerCleanupRuntime interface {
	StopContainer(ctx context.Context, containerID string, timeout time.Duration) error
	RemoveContainer(ctx context.Context, containerID string, force bool) error
	RemoveNetwork(ctx context.Context, networkID string) error
	RemoveACLRules(ctx context.Context, rules []model.InstanceRuntimeACLRule) error
}

// ContainerFileWriter 定义向容器写入文件的最小能力。
type ContainerFileWriter interface {
	WriteFileToContainer(ctx context.Context, containerID, filePath string, content []byte) error
}

// ContainerFileRuntime 定义 AWD 防守工作台等文件/命令访问能力。
type ContainerFileRuntime interface {
	ContainerFileWriter
	ReadFileFromContainer(ctx context.Context, containerID, filePath string, limit int64) ([]byte, error)
	ListDirectoryFromContainer(ctx context.Context, containerID, dirPath string, limit int) ([]ContainerDirectoryEntry, error)
	ExecContainerCommand(ctx context.Context, containerID string, command []string, stdin []byte, limit int64) ([]byte, error)
}

// ContainerImageRuntime 定义镜像检查与删除能力。
type ContainerImageRuntime interface {
	InspectImageSize(ctx context.Context, imageRef string) (int64, error)
	RemoveImage(ctx context.Context, imageRef string) error
}

// ManagedContainerInventory 定义受管容器盘点能力。
type ManagedContainerInventory interface {
	ListManagedContainers(ctx context.Context) ([]ManagedContainer, error)
	InspectManagedContainer(ctx context.Context, containerID string) (*ManagedContainerState, error)
}

// ManagedContainerStatsReader 定义受管容器指标读取能力。
type ManagedContainerStatsReader interface {
	ListManagedContainerStats(ctx context.Context) ([]ManagedContainerStat, error)
}

// ContainerInteractiveExecutor 定义交互式容器命令执行能力。
type ContainerInteractiveExecutor interface {
	ExecContainerInteractive(ctx context.Context, containerID string, command []string, stdin io.Reader, stdout io.Writer) error
}
