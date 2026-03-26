package application

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"ctf-platform/internal/authctx"
	runtimeports "ctf-platform/internal/module/runtime/ports"
	"ctf-platform/pkg/errcode"
)

// ProxyTicketClaims 表示实例代理票据载荷。
type ProxyTicketClaims = runtimeports.ProxyTicketClaims

// ProxyTicketStore 定义代理票据持久化端口。
type ProxyTicketStore = runtimeports.ProxyTicketStore

// ProxyTicketService 负责代理票据签发和解析。
type ProxyTicketService struct {
	store     ProxyTicketStore
	ticketTTL time.Duration
}

// NewProxyTicketService 创建代理票据应用服务。
func NewProxyTicketService(store ProxyTicketStore, ticketTTL time.Duration) *ProxyTicketService {
	return &ProxyTicketService{
		store:     store,
		ticketTTL: ticketTTL,
	}
}

// IssueTicket 签发实例代理票据。
func (s *ProxyTicketService) IssueTicket(ctx context.Context, user authctx.CurrentUser, instanceID int64) (string, time.Time, error) {
	if s == nil || s.store == nil || s.ticketTTL <= 0 {
		return "", time.Time{}, errProxyTicketServiceUnavailable()
	}

	ticket, err := generateProxyToken(32)
	if err != nil {
		return "", time.Time{}, errcode.ErrInternal.WithCause(err)
	}

	claims := ProxyTicketClaims{
		UserID:     user.UserID,
		Username:   user.Username,
		Role:       user.Role,
		InstanceID: instanceID,
		IssuedAt:   time.Now().UTC(),
	}
	expiresAt := time.Now().Add(s.ticketTTL).UTC()

	if err := s.store.SaveProxyTicket(ctx, ticket, claims, s.ticketTTL); err != nil {
		return "", time.Time{}, errcode.ErrInternal.WithCause(err)
	}

	return ticket, expiresAt, nil
}

// ResolveTicket 解析实例代理票据。
func (s *ProxyTicketService) ResolveTicket(ctx context.Context, ticket string) (*ProxyTicketClaims, error) {
	if s == nil || s.store == nil {
		return nil, errProxyTicketServiceUnavailable()
	}
	if ticket == "" {
		return nil, errcode.ErrProxyTicketInvalid
	}

	claims, err := s.store.FindProxyTicket(ctx, ticket)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if claims == nil {
		return nil, errcode.ErrProxyTicketInvalid
	}
	if claims.UserID <= 0 || claims.InstanceID <= 0 || claims.Username == "" || claims.Role == "" {
		return nil, errcode.ErrProxyTicketInvalid
	}

	return claims, nil
}

// MaxAge 返回代理票据 cookie 的秒级寿命。
func (s *ProxyTicketService) MaxAge() int {
	if s == nil || s.ticketTTL <= 0 {
		return 0
	}
	return int(s.ticketTTL.Seconds())
}

func errProxyTicketServiceUnavailable() error {
	return errcode.ErrInternal.WithCause(fmt.Errorf("proxy ticket service is not configured"))
}

func generateProxyToken(size int) (string, error) {
	buffer := make([]byte, size)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}
