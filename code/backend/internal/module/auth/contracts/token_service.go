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
	// RevokeAllUserSessions 撤销指定用户的所有会话（密码变更、账号禁用等场景）。
	RevokeAllUserSessions(ctx context.Context, userID int64) error
	// ListUserSessions 返回指定用户的活跃会话列表（管理员会话管理场景）。
	ListUserSessions(ctx context.Context, userID int64) ([]Session, error)
	IssueWSTicket(ctx context.Context, user authctx.CurrentUser) (*WSTicket, error)
	ConsumeWSTicket(ctx context.Context, ticket string) (*authctx.CurrentUser, error)
}
