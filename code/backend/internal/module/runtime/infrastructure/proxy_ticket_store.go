package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	redislib "github.com/redis/go-redis/v9"

	runtimeports "ctf-platform/internal/module/runtime/ports"
)

const proxyTicketKeyPrefix = "ctf:instance:proxy:ticket"

// ProxyTicketStore 提供 Redis 版代理票据存储实现。
type ProxyTicketStore struct {
	cache *redislib.Client
}

// NewProxyTicketStore 创建代理票据 Redis 存储。
func NewProxyTicketStore(cache *redislib.Client) *ProxyTicketStore {
	return &ProxyTicketStore{cache: cache}
}

// SaveProxyTicket 保存代理票据。
func (s *ProxyTicketStore) SaveProxyTicket(ctx context.Context, ticket string, claims runtimeports.ProxyTicketClaims, ttl time.Duration) error {
	if s == nil || s.cache == nil {
		return fmt.Errorf("proxy ticket store is not configured")
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		return err
	}
	return s.cache.Set(ctx, s.ticketKey(ticket), payload, ttl).Err()
}

// FindProxyTicket 读取代理票据。
func (s *ProxyTicketStore) FindProxyTicket(ctx context.Context, ticket string) (*runtimeports.ProxyTicketClaims, error) {
	if s == nil || s.cache == nil {
		return nil, fmt.Errorf("proxy ticket store is not configured")
	}

	payload, err := s.cache.Get(ctx, s.ticketKey(ticket)).Result()
	if errors.Is(err, redislib.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var claims runtimeports.ProxyTicketClaims
	if err := json.Unmarshal([]byte(payload), &claims); err != nil {
		return nil, err
	}
	return &claims, nil
}

func (s *ProxyTicketStore) ticketKey(ticket string) string {
	return fmt.Sprintf("%s:%s", proxyTicketKeyPrefix, ticket)
}
