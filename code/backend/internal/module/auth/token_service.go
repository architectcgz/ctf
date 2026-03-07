package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	redislib "github.com/redis/go-redis/v9"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/config"
	"ctf-platform/pkg/errcode"
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

type wsTicketPayload struct {
	UserID   int64     `json:"user_id"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
	IssuedAt time.Time `json:"issued_at"`
}

type TokenService interface {
	IssueTokens(userID int64, username, role string) (*TokenPair, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*dtoRefreshPayload, error)
	RevokeToken(ctx context.Context, jti string, ttl time.Duration) error
	IsRevoked(ctx context.Context, jti string) (bool, error)
	ParseToken(tokenString string) (*jwtpkg.Claims, error)
	IssueWSTicket(ctx context.Context, user authctx.CurrentUser) (*WSTicket, error)
	ConsumeWSTicket(ctx context.Context, ticket string) (*authctx.CurrentUser, error)
}

type dtoRefreshPayload struct {
	AccessToken string
	ExpiresIn   int64
}

type tokenService struct {
	config   config.AuthConfig
	wsConfig config.WebSocketConfig
	cache    *redislib.Client
	manager  *jwtpkg.Manager
}

func NewTokenService(cfg config.AuthConfig, wsConfig config.WebSocketConfig, cache *redislib.Client, manager *jwtpkg.Manager) TokenService {
	return &tokenService{
		config:   cfg,
		wsConfig: wsConfig,
		cache:    cache,
		manager:  manager,
	}
}

func (s *tokenService) IssueTokens(userID int64, username, role string) (*TokenPair, error) {
	accessToken, _, err := s.manager.GenerateAccessToken(userID, username, role)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, _, err := s.manager.GenerateRefreshToken(userID, username, role)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:     accessToken,
		RefreshToken:    refreshToken,
		AccessTokenTTL:  s.manager.AccessTokenTTL(),
		RefreshTokenTTL: s.manager.RefreshTokenTTL(),
	}, nil
}

func (s *tokenService) RefreshAccessToken(ctx context.Context, refreshToken string) (*dtoRefreshPayload, error) {
	claims, err := s.manager.ParseToken(refreshToken)
	if err != nil {
		return nil, mapJWTError(err, true)
	}
	if claims.TokenType != jwtpkg.TokenTypeRefresh {
		return nil, errcode.ErrTokenInvalid
	}

	revoked, err := s.IsRevoked(ctx, claims.ID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if revoked {
		return nil, errcode.ErrTokenRevoked
	}

	accessToken, accessClaims, err := s.manager.GenerateAccessToken(claims.UserID, claims.Username, claims.Role)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return &dtoRefreshPayload{
		AccessToken: accessToken,
		ExpiresIn:   accessClaims.ExpiresAt.Time.Unix() - time.Now().Unix(),
	}, nil
}

func (s *tokenService) RevokeToken(ctx context.Context, jti string, ttl time.Duration) error {
	if jti == "" {
		return nil
	}
	return s.cache.Set(ctx, s.blacklistKey(jti), "1", ttl).Err()
}

func (s *tokenService) IsRevoked(ctx context.Context, jti string) (bool, error) {
	if jti == "" {
		return false, nil
	}

	count, err := s.cache.Exists(ctx, s.blacklistKey(jti)).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *tokenService) ParseToken(tokenString string) (*jwtpkg.Claims, error) {
	return s.manager.ParseToken(tokenString)
}

func (s *tokenService) IssueWSTicket(ctx context.Context, user authctx.CurrentUser) (*WSTicket, error) {
	ticket, err := generateOpaqueToken(32)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	payload, err := json.Marshal(wsTicketPayload{
		UserID:   user.UserID,
		Username: user.Username,
		Role:     user.Role,
		IssuedAt: time.Now().UTC(),
	})
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	expiresAt := time.Now().Add(s.wsConfig.TicketTTL).UTC()
	if err := s.cache.Set(ctx, s.wsTicketKey(ticket), payload, s.wsConfig.TicketTTL).Err(); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return &WSTicket{
		Ticket:    ticket,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *tokenService) ConsumeWSTicket(ctx context.Context, ticket string) (*authctx.CurrentUser, error) {
	if ticket == "" {
		return nil, errcode.ErrWSTicketInvalid
	}

	payload, err := s.cache.GetDel(ctx, s.wsTicketKey(ticket)).Result()
	if errors.Is(err, redislib.Nil) {
		return nil, errcode.ErrWSTicketInvalid
	}
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	var claims wsTicketPayload
	if err := json.Unmarshal([]byte(payload), &claims); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if claims.UserID <= 0 || claims.Username == "" || claims.Role == "" {
		return nil, errcode.ErrWSTicketInvalid
	}

	return &authctx.CurrentUser{
		UserID:   claims.UserID,
		Username: claims.Username,
		Role:     claims.Role,
	}, nil
}

func (s *tokenService) blacklistKey(jti string) string {
	return fmt.Sprintf("%s:%s", s.config.TokenBlacklistPrefix, jti)
}

func (s *tokenService) wsTicketKey(ticket string) string {
	return fmt.Sprintf("%s:%s", s.wsConfig.TicketKeyPrefix, ticket)
}

func generateOpaqueToken(size int) (string, error) {
	buffer := make([]byte, size)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func mapJWTError(err error, isRefresh bool) error {
	switch {
	case errors.Is(err, jwtpkg.ErrExpiredToken) && isRefresh:
		return errcode.ErrRefreshTokenExpired
	case errors.Is(err, jwtpkg.ErrExpiredToken):
		return errcode.ErrAccessTokenExpired
	case errors.Is(err, jwtpkg.ErrInvalidToken):
		return errcode.ErrTokenInvalid
	default:
		return err
	}
}
