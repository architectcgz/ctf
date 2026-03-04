package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	redislib "github.com/redis/go-redis/v9"

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

type TokenService interface {
	IssueTokens(userID int64, username, role string) (*TokenPair, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*dtoRefreshPayload, error)
	RevokeToken(ctx context.Context, jti string, ttl time.Duration) error
	IsRevoked(ctx context.Context, jti string) (bool, error)
	ParseToken(tokenString string) (*jwtpkg.Claims, error)
}

type dtoRefreshPayload struct {
	AccessToken string
	ExpiresIn   int64
}

type tokenService struct {
	config  config.AuthConfig
	cache   *redislib.Client
	manager *jwtpkg.Manager
}

func NewTokenService(cfg config.AuthConfig, cache *redislib.Client, manager *jwtpkg.Manager) TokenService {
	return &tokenService{
		config:  cfg,
		cache:   cache,
		manager: manager,
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

func (s *tokenService) blacklistKey(jti string) string {
	return fmt.Sprintf("%s:%s", s.config.TokenBlacklistPrefix, jti)
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
