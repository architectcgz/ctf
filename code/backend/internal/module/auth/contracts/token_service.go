package contracts

import (
	"context"
	"time"

	"ctf-platform/internal/authctx"
)

type Session struct {
	ID        string
	UserID    int64
	Username  string
	Role      string
	ExpiresAt time.Time
}

type WSTicket struct {
	Ticket    string
	ExpiresAt time.Time
}

type TokenService interface {
	CreateSession(ctx context.Context, userID int64, username, role string) (*Session, error)
	GetSession(ctx context.Context, sessionID string) (*Session, error)
	DeleteSession(ctx context.Context, sessionID string) error
	IssueWSTicket(ctx context.Context, user authctx.CurrentUser) (*WSTicket, error)
	ConsumeWSTicket(ctx context.Context, ticket string) (*authctx.CurrentUser, error)
}
