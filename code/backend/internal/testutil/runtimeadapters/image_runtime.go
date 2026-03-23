package runtimeadapters

import "context"

// ImageRuntime 为测试提供 challenge image service 所需的 runtime bridge。
type ImageRuntime struct {
	inspectImageSizeFn func(ctx context.Context, imageRef string) (int64, error)
	removeImageFn      func(ctx context.Context, imageRef string) error
}

// NewImageRuntime 创建 image runtime 测试桥接。
func NewImageRuntime(
	inspectImageSizeFn func(context.Context, string) (int64, error),
	removeImageFn func(context.Context, string) error,
) *ImageRuntime {
	if inspectImageSizeFn == nil && removeImageFn == nil {
		return nil
	}
	return &ImageRuntime{
		inspectImageSizeFn: inspectImageSizeFn,
		removeImageFn:      removeImageFn,
	}
}

func (a *ImageRuntime) InspectImageSize(ctx context.Context, imageRef string) (int64, error) {
	if a == nil || a.inspectImageSizeFn == nil {
		return 0, nil
	}
	return a.inspectImageSizeFn(ctx, imageRef)
}

func (a *ImageRuntime) RemoveImage(ctx context.Context, imageRef string) error {
	if a == nil || a.removeImageFn == nil {
		return nil
	}
	return a.removeImageFn(ctx, imageRef)
}
