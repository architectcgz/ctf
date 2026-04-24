package infrastructure

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
	authcontracts "ctf-platform/internal/module/auth/contracts"
	"ctf-platform/pkg/errcode"
)

type wsTicketPayload struct {
	UserID   int64     `json:"user_id"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
	IssuedAt time.Time `json:"issued_at"`
}

type sessionRecord struct {
	ID        string    `json:"id"`
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	ExpiresAt time.Time `json:"expires_at"`
}

type tokenService struct {
	config   config.AuthConfig
	wsConfig config.WebSocketConfig
	cache    *redislib.Client
}

func NewTokenService(cfg config.AuthConfig, wsConfig config.WebSocketConfig, cache *redislib.Client) authcontracts.TokenService {
	return &tokenService{
		config:   cfg,
		wsConfig: wsConfig,
		cache:    cache,
	}
}

func (s *tokenService) CreateSession(ctx context.Context, userID int64, username, role string) (*authcontracts.Session, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	sessionID, err := generateOpaqueToken(32)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	expiresAt := time.Now().Add(s.config.SessionTTL).UTC()
	record := sessionRecord{
		ID:        sessionID,
		UserID:    userID,
		Username:  username,
		Role:      role,
		ExpiresAt: expiresAt,
	}
	payload, err := json.Marshal(record)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if err := s.cache.Set(ctx, s.sessionKey(sessionID), payload, s.config.SessionTTL).Err(); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return &authcontracts.Session{
		ID:        sessionID,
		UserID:    userID,
		Username:  username,
		Role:      role,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *tokenService) GetSession(ctx context.Context, sessionID string) (*authcontracts.Session, error) {
	if sessionID == "" {
		return nil, errcode.ErrUnauthorized
	}

	payload, err := s.cache.Get(ctx, s.sessionKey(sessionID)).Result()
	if errors.Is(err, redislib.Nil) {
		return nil, errcode.ErrUnauthorized
	}
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	var record sessionRecord
	if err := json.Unmarshal([]byte(payload), &record); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if record.ID == "" || record.UserID <= 0 || record.Username == "" || record.Role == "" {
		return nil, errcode.ErrUnauthorized
	}

	return &authcontracts.Session{
		ID:        record.ID,
		UserID:    record.UserID,
		Username:  record.Username,
		Role:      record.Role,
		ExpiresAt: record.ExpiresAt,
	}, nil
}

func (s *tokenService) DeleteSession(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return nil
	}
	if err := s.cache.Del(ctx, s.sessionKey(sessionID)).Err(); err != nil {
		return err
	}
	return nil
}

func (s *tokenService) IssueWSTicket(ctx context.Context, user authctx.CurrentUser) (*authcontracts.WSTicket, error) {
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

	return &authcontracts.WSTicket{
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

func (s *tokenService) sessionKey(sessionID string) string {
	return fmt.Sprintf("%s:%s", s.config.SessionKeyPrefix, sessionID)
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
