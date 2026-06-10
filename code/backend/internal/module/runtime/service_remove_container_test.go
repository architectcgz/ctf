package runtime_test

import (
	"context"
	"errors"
	"testing"

	runtimecmd "ctf-platform/internal/module/container_runtime/application/commands"
	runtimeports "ctf-platform/internal/module/container_runtime/ports"
)

func TestServiceRemoveContainerFailsWhenRuntimeEngineUnavailable(t *testing.T) {
	t.Parallel()

	cleanupService := runtimecmd.NewRuntimeCleanupService(nil, nil, nil)

	err := cleanupService.RemoveContainer(context.Background(), "ctr-missing-engine")
	if err == nil {
		t.Fatal("expected RemoveContainer() to fail when runtime engine is unavailable")
	}
	if !errors.Is(err, runtimeports.ErrRuntimeEngineUnavailable) {
		t.Fatalf("expected runtime engine unavailable error, got %v", err)
	}
}

func TestServiceRemoveContainerHonorsCancellation(t *testing.T) {
	t.Parallel()

	engine := &fakeRuntimeEngine{
		removeContainerFn: func(ctx context.Context, containerID string, force bool) error {
			if containerID != "ctr-ctx" || !force {
				t.Fatalf("unexpected remove container args: id=%s force=%v", containerID, force)
			}
			<-ctx.Done()
			return ctx.Err()
		},
	}
	cleanupService := runtimecmd.NewRuntimeCleanupService(engine, nil, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := cleanupService.RemoveContainer(ctx, "ctr-ctx"); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

func TestServiceRemoveContainerIgnoresMissingContainer(t *testing.T) {
	t.Parallel()

	engine := &fakeRuntimeEngine{
		removeContainerErr: runtimeports.WrapRuntimeContainerNotFound(errors.New("Error response from daemon: No such container: ctr-missing")),
	}
	cleanupService := runtimecmd.NewRuntimeCleanupService(engine, nil, nil)

	if err := cleanupService.RemoveContainer(context.Background(), "ctr-missing"); err != nil {
		t.Fatalf("expected missing container to be ignored, got %v", err)
	}
	if engine.removedContainerID != "ctr-missing" {
		t.Fatalf("expected cleanup to attempt removing ctr-missing, got %s", engine.removedContainerID)
	}
}
