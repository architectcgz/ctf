package application

import (
	"context"
	"testing"
)

type stubImageRuntime struct {
	inspectCalls int
	removeCalls  int
	size         int64
	imageRef     string
	removeRef    string
	inspectCtx   context.Context
	removeCtx    context.Context
}

func (s *stubImageRuntime) InspectImageSize(ctx context.Context, imageRef string) (int64, error) {
	s.inspectCalls++
	s.inspectCtx = ctx
	s.imageRef = imageRef
	return s.size, nil
}

func (s *stubImageRuntime) RemoveImage(ctx context.Context, imageRef string) error {
	s.removeCalls++
	s.removeCtx = ctx
	s.removeRef = imageRef
	return nil
}

func TestImageRuntimeServiceInspectImageSizeSkipsBlankRef(t *testing.T) {
	t.Parallel()

	runtime := &stubImageRuntime{size: 128}
	service := NewImageRuntimeService(runtime)

	size, err := service.InspectImageSize(context.Background(), "   ")
	if err != nil {
		t.Fatalf("InspectImageSize() error = %v", err)
	}
	if size != 0 {
		t.Fatalf("InspectImageSize() size = %d, want 0", size)
	}
	if runtime.inspectCalls != 0 {
		t.Fatalf("InspectImageSize() inspectCalls = %d, want 0", runtime.inspectCalls)
	}
}

func TestImageRuntimeServiceDelegatesToRuntime(t *testing.T) {
	t.Parallel()

	runtime := &stubImageRuntime{size: 256}
	service := NewImageRuntimeService(runtime)

	size, err := service.InspectImageSize(context.Background(), "repo/app:latest")
	if err != nil {
		t.Fatalf("InspectImageSize() error = %v", err)
	}
	if size != 256 {
		t.Fatalf("InspectImageSize() size = %d, want 256", size)
	}
	if runtime.inspectCalls != 1 || runtime.imageRef != "repo/app:latest" {
		t.Fatalf("InspectImageSize() delegated incorrectly, calls = %d ref = %q", runtime.inspectCalls, runtime.imageRef)
	}

	if err := service.RemoveImage(context.Background(), "repo/app:latest"); err != nil {
		t.Fatalf("RemoveImage() error = %v", err)
	}
	if runtime.removeCalls != 1 || runtime.removeRef != "repo/app:latest" {
		t.Fatalf("RemoveImage() delegated incorrectly, calls = %d ref = %q", runtime.removeCalls, runtime.removeRef)
	}
}

func TestImageRuntimeServiceDoesNotCreateBackgroundContext(t *testing.T) {
	t.Parallel()

	runtime := &stubImageRuntime{size: 256}
	service := NewImageRuntimeService(runtime)

	if _, err := service.InspectImageSize(nil, "repo/app:latest"); err != nil {
		t.Fatalf("InspectImageSize() error = %v", err)
	}
	if runtime.inspectCtx != nil {
		t.Fatalf("InspectImageSize() ctx = %v, want nil", runtime.inspectCtx)
	}

	if err := service.RemoveImage(nil, "repo/app:latest"); err != nil {
		t.Fatalf("RemoveImage() error = %v", err)
	}
	if runtime.removeCtx != nil {
		t.Fatalf("RemoveImage() ctx = %v, want nil", runtime.removeCtx)
	}
}
