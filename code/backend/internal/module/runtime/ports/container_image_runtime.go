package ports

import "context"

// ContainerImageRuntime 定义镜像检查与删除能力。
type ContainerImageRuntime interface {
	InspectImageSize(ctx context.Context, imageRef string) (int64, error)
	RemoveImage(ctx context.Context, imageRef string) error
}
