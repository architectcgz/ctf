package contracts

import (
	"context"
	"time"

	"ctf-platform/internal/authctx"
	jwtpkg "ctf-platform/pkg/jwt"
)

type TokenPair struct {
	AccessToken     string
	RefreshToken    string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type WSTicket struct {
	Ticket    string
	ExpiresAt time.Time
}

type RefreshAccessPayload struct {
	AccessToken string
	ExpiresIn   int64
}

type TokenService interface {
	IssueTokens(ctx context.Context, userID int64, username, role string) (*TokenPair, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*RefreshAccessPayload, error)
	RevokeToken(ctx context.Context, jti string, ttl time.Duration) error
	ClearRefreshSession(ctx context.Context, userID int64, refreshJTI string) error
	IsRevoked(ctx context.Context, jti string) (bool, error)
	ParseToken(tokenString string) (*jwtpkg.Claims, error)
	IssueWSTicket(ctx context.Context, user authctx.CurrentUser) (*WSTicket, error)
	ConsumeWSTicket(ctx context.Context, ticket string) (*authctx.CurrentUser, error)
}
