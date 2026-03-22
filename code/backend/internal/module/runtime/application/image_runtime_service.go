package application

import (
	"context"
	"strings"
)

type imageRuntime interface {
	InspectImageSize(ctx context.Context, imageRef string) (int64, error)
	RemoveImage(ctx context.Context, imageRef string) error
}

// ImageRuntimeService 收口镜像检查与删除等运行时能力。
type ImageRuntimeService struct {
	runtime imageRuntime
}

// NewImageRuntimeService 创建镜像运行时服务。
func NewImageRuntimeService(runtime imageRuntime) *ImageRuntimeService {
	if isNilApplicationDependency(runtime) {
		runtime = nil
	}
	return &ImageRuntimeService{runtime: runtime}
}

// InspectImageSize 返回镜像大小；未启用运行时或镜像引用为空时返回 0。
func (s *ImageRuntimeService) InspectImageSize(ctx context.Context, imageRef string) (int64, error) {
	if s == nil || s.runtime == nil || strings.TrimSpace(imageRef) == "" {
		return 0, nil
	}
	return s.runtime.InspectImageSize(normalizeContext(ctx), imageRef)
}

// RemoveImage 删除指定镜像；未启用运行时或镜像引用为空时降级跳过。
func (s *ImageRuntimeService) RemoveImage(ctx context.Context, imageRef string) error {
	if s == nil || s.runtime == nil || strings.TrimSpace(imageRef) == "" {
		return nil
	}
	return s.runtime.RemoveImage(normalizeContext(ctx), imageRef)
}
