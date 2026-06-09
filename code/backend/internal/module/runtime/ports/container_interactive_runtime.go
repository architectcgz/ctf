package ports

import (
	"context"
	"io"
)

// ContainerInteractiveExecutor 定义交互式容器命令执行能力。
type ContainerInteractiveExecutor interface {
	ExecContainerInteractive(ctx context.Context, containerID string, command []string, stdin io.Reader, stdout io.Writer) error
}
