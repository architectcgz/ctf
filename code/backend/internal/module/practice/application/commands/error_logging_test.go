package commands

import (
	"errors"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"ctf-platform/pkg/errcode"
)

func TestWrappedErrorCauseFieldLogsAppErrorCause(t *testing.T) {
	t.Parallel()

	cause := errors.New("docker create failed")
	err := errcode.ErrContainerCreateFailed.WithCause(cause)
	core, observed := observer.New(zap.WarnLevel)
	logger := zap.New(core)

	logger.Warn("failed", zap.Error(err), wrappedErrorCauseField(err))

	if observed.Len() != 1 {
		t.Fatalf("expected one log entry, got %d", observed.Len())
	}
	fields := observed.All()[0].ContextMap()
	if fields["error"] != "容器创建失败" {
		t.Fatalf("error field = %v, want 容器创建失败", fields["error"])
	}
	if fields["cause"] != "docker create failed" {
		t.Fatalf("cause field = %v, want docker create failed", fields["cause"])
	}
}

func TestWrappedErrorCauseFieldSkipsPlainError(t *testing.T) {
	t.Parallel()

	core, observed := observer.New(zap.WarnLevel)
	logger := zap.New(core)

	logger.Warn("failed", zap.Error(errors.New("plain error")), wrappedErrorCauseField(errors.New("plain error")))

	fields := observed.All()[0].ContextMap()
	if _, ok := fields["cause"]; ok {
		t.Fatalf("plain error should not log cause field, got %v", fields["cause"])
	}
}
