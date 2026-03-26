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

type RiskRepository interface {
	ListRecentSubmitEvents(ctx context.Context, since time.Time, limit int) ([]RiskAuditEvent, error)
	ListRecentLoginEvents(ctx context.Context, since time.Time, limit int) ([]RiskAuditEvent, error)
}
