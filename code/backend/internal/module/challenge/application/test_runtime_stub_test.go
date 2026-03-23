package application

import "context"

type stubImageRuntime struct {
	inspectImageSizeFn func(ctx context.Context, imageRef string) (int64, error)
	removeImageFn      func(ctx context.Context, imageRef string) error
}

func (s *stubImageRuntime) InspectImageSize(ctx context.Context, imageRef string) (int64, error) {
	if s.inspectImageSizeFn == nil {
		return 0, nil
	}
	return s.inspectImageSizeFn(ctx, imageRef)
}

func (s *stubImageRuntime) RemoveImage(ctx context.Context, imageRef string) error {
	if s.removeImageFn == nil {
		return nil
	}
	return s.removeImageFn(ctx, imageRef)
}
