package ops

import (
	"context"

	"ctf-platform/internal/auditlog"
)

type Module struct {
	recorder AuditRecorder
}

func NewModule(recorder AuditRecorder) *Module {
	return &Module{recorder: recorder}
}

func (m *Module) Record(ctx context.Context, entry auditlog.Entry) error {
	if m == nil || m.recorder == nil {
		return nil
	}
	return m.recorder.Record(ctx, entry)
}
