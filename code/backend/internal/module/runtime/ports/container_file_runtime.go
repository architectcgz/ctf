package ports

import "context"

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
