package application

import (
	"context"

	"go.uber.org/zap"
)

type containerFileRuntime interface {
	WriteFileToContainer(ctx context.Context, containerID, filePath string, content []byte) error
}

// ContainerFileService 收口容器文件写入能力。
type ContainerFileService struct {
	runtime containerFileRuntime
	logger  *zap.Logger
}

// NewContainerFileService 创建容器文件写入服务。
func NewContainerFileService(runtime containerFileRuntime, logger *zap.Logger) *ContainerFileService {
	if logger == nil {
		logger = zap.NewNop()
	}
	if isNilApplicationDependency(runtime) {
		runtime = nil
	}
	return &ContainerFileService{
		runtime: runtime,
		logger:  logger,
	}
}

// WriteFileToContainer 向容器写入文件；未启用运行时时降级跳过。
func (s *ContainerFileService) WriteFileToContainer(ctx context.Context, containerID, filePath string, content []byte) error {
	if s == nil || s.runtime == nil {
		if s != nil && s.logger != nil {
			s.logger.Info("写入容器文件（降级跳过）", zap.String("container_id", containerID), zap.String("path", filePath))
		}
		return nil
	}
	return s.runtime.WriteFileToContainer(normalizeContext(ctx), containerID, filePath, content)
}
