package application

import (
	"context"
	"testing"
)

type stubContainerFileRuntime struct {
	calls       int
	containerID string
	filePath    string
	content     []byte
}

func (s *stubContainerFileRuntime) WriteFileToContainer(_ context.Context, containerID, filePath string, content []byte) error {
	s.calls++
	s.containerID = containerID
	s.filePath = filePath
	s.content = append([]byte(nil), content...)
	return nil
}

func TestContainerFileServiceWriteFileToContainerSkipsNilRuntime(t *testing.T) {
	t.Parallel()

	service := NewContainerFileService(nil, nil)
	if err := service.WriteFileToContainer(context.Background(), "container-1", "/flag", []byte("flag")); err != nil {
		t.Fatalf("WriteFileToContainer() error = %v", err)
	}
}

func TestContainerFileServiceWriteFileToContainerDelegatesToRuntime(t *testing.T) {
	t.Parallel()

	runtime := &stubContainerFileRuntime{}
	service := NewContainerFileService(runtime, nil)

	if err := service.WriteFileToContainer(context.Background(), "container-1", "/flag", []byte("flag")); err != nil {
		t.Fatalf("WriteFileToContainer() error = %v", err)
	}
	if runtime.calls != 1 {
		t.Fatalf("WriteFileToContainer() calls = %d, want 1", runtime.calls)
	}
	if runtime.containerID != "container-1" || runtime.filePath != "/flag" || string(runtime.content) != "flag" {
		t.Fatalf("WriteFileToContainer() delegated incorrectly, containerID = %q filePath = %q content = %q", runtime.containerID, runtime.filePath, string(runtime.content))
	}
}
