package runtime

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
)

const proxyTicketKeyPrefix = "ctf:instance:proxy:ticket"

type ProxyTicketClaims struct {
	UserID     int64     `json:"user_id"`
	Username   string    `json:"username"`
	Role       string    `json:"role"`
	InstanceID int64     `json:"instance_id"`
	IssuedAt   time.Time `json:"issued_at"`
}

type ProxyTicketService struct {
	cache *redislib.Client
	cfg   *config.ContainerConfig
}

func NewProxyTicketService(cache *redislib.Client, cfg *config.ContainerConfig) *ProxyTicketService {
	return &ProxyTicketService{
		cache: cache,
		cfg:   cfg,
	}
}

func (s *ProxyTicketService) IssueTicket(ctx context.Context, user authctx.CurrentUser, instanceID int64) (string, time.Time, error) {
	if s == nil || s.cache == nil || s.cfg == nil {
		return "", time.Time{}, errcode.ErrInternal.WithCause(fmt.Errorf("proxy ticket service is not configured"))
	}

	ticket, err := generateProxyToken(32)
	if err != nil {
		return "", time.Time{}, errcode.ErrInternal.WithCause(err)
	}

	payload, err := json.Marshal(ProxyTicketClaims{
		UserID:     user.UserID,
		Username:   user.Username,
		Role:       user.Role,
		InstanceID: instanceID,
		IssuedAt:   time.Now().UTC(),
	})
	if err != nil {
		return "", time.Time{}, errcode.ErrInternal.WithCause(err)
	}

	expiresAt := time.Now().Add(s.cfg.ProxyTicketTTL).UTC()
	if err := s.cache.Set(ctx, s.ticketKey(ticket), payload, s.cfg.ProxyTicketTTL).Err(); err != nil {
		return "", time.Time{}, errcode.ErrInternal.WithCause(err)
	}

	return ticket, expiresAt, nil
}

func (s *ProxyTicketService) ResolveTicket(ctx context.Context, ticket string) (*ProxyTicketClaims, error) {
	if s == nil || s.cache == nil {
		return nil, errcode.ErrInternal.WithCause(fmt.Errorf("proxy ticket service is not configured"))
	}
	if ticket == "" {
		return nil, errcode.ErrProxyTicketInvalid
	}

	payload, err := s.cache.Get(ctx, s.ticketKey(ticket)).Result()
	if errors.Is(err, redislib.Nil) {
		return nil, errcode.ErrProxyTicketInvalid
	}
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	var claims ProxyTicketClaims
	if err := json.Unmarshal([]byte(payload), &claims); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if claims.UserID <= 0 || claims.InstanceID <= 0 || claims.Username == "" || claims.Role == "" {
		return nil, errcode.ErrProxyTicketInvalid
	}

	return &claims, nil
}

func (s *ProxyTicketService) ticketKey(ticket string) string {
	return fmt.Sprintf("%s:%s", proxyTicketKeyPrefix, ticket)
}

func generateProxyToken(size int) (string, error) {
	buffer := make([]byte, size)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}
