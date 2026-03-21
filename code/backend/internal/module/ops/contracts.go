package ops

import (
	"context"

	"ctf-platform/internal/auditlog"
)

type AuditRecorder interface {
	Record(ctx context.Context, entry auditlog.Entry) error
}
