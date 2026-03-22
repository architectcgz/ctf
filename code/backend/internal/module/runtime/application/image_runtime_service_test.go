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
}

func (s *stubImageRuntime) InspectImageSize(_ context.Context, imageRef string) (int64, error) {
	s.inspectCalls++
	s.imageRef = imageRef
	return s.size, nil
}

func (s *stubImageRuntime) RemoveImage(_ context.Context, imageRef string) error {
	s.removeCalls++
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
