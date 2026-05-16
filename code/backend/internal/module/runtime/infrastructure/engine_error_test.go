package infrastructure

import (
	"errors"
	"testing"

	"github.com/docker/docker/errdefs"

	runtimeports "ctf-platform/internal/module/runtime/ports"
)

func TestNormalizeContainerCreateErrorWrapsPublishedHostPortConflict(t *testing.T) {
	t.Parallel()

	rawErr := errdefs.System(errors.New("Error response from daemon: driver failed programming external connectivity on endpoint test: Bind for 0.0.0.0:30011 failed: port is already allocated"))

	err := normalizeContainerCreateError(rawErr)
	if !errors.Is(err, runtimeports.ErrPublishedHostPortConflict) {
		t.Fatalf("expected published host port conflict, got %v", err)
	}
	if !errors.Is(err, rawErr) {
		t.Fatalf("expected wrapped error to preserve original docker error, got %v", err)
	}
	if errors.Unwrap(err) != rawErr {
		t.Fatalf("expected wrapped error to unwrap to original docker error, got %v", errors.Unwrap(err))
	}
	if !errdefs.IsSystem(err) {
		t.Fatalf("expected wrapped error to preserve docker errdefs classification, got %T: %v", err, err)
	}
}

func TestNormalizeContainerCreateErrorPassesThroughNonConflict(t *testing.T) {
	t.Parallel()

	rawErr := errors.New("Error response from daemon: No such image: ctf/web:v1")

	err := normalizeContainerCreateError(rawErr)
	if !errors.Is(err, rawErr) {
		t.Fatalf("expected non-conflict error passthrough, got %v", err)
	}
	if errors.Is(err, runtimeports.ErrPublishedHostPortConflict) {
		t.Fatalf("expected non-conflict error not to match published host port conflict, got %v", err)
	}
}
