package ports

import (
	"context"
	"time"
)

type RiskAuditEvent struct {
	UserID    *int64
	Username  string
	IPAddress string
	CreatedAt time.Time
}

type RiskSubmitEventRepository interface {
	ListRecentSubmitEvents(ctx context.Context, since time.Time, limit int) ([]RiskAuditEvent, error)
}

type RiskLoginEventRepository interface {
	ListRecentLoginEvents(ctx context.Context, since time.Time, limit int) ([]RiskAuditEvent, error)
}
