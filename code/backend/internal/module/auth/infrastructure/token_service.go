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

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/authctx"
	"ctf-platform/internal/config"
	authcontracts "ctf-platform/internal/module/auth/contracts"
)

type wsTicketPayload struct {
	UserID   int64     `json:"user_id"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
	IssuedAt time.Time `json:"issued_at"`
}

type mcpTokenPayload struct {
	UserID         int64     `json:"user_id"`
	Username       string    `json:"username"`
	Role           string    `json:"role"`
	SessionVersion int64     `json:"session_version"`
	IssuedAt       time.Time `json:"issued_at"`
	ExpiresAt      time.Time `json:"expires_at"`
}

type sessionRecord struct {
	ID             string    `json:"id"`
	UserID         int64     `json:"user_id"`
	Username       string    `json:"username"`
	Role           string    `json:"role"`
	SessionVersion int64     `json:"session_version,omitempty"`
	ExpiresAt      time.Time `json:"expires_at"`
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
	if err := requireContext(ctx); err != nil {
		return nil, err
	}

	sessionID, err := generateOpaqueToken(32)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	sessionVersion, err := s.getUserSessionVersion(ctx, userID)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	expiresAt := time.Now().Add(s.config.SessionTTL).UTC()
	record := sessionRecord{
		ID:             sessionID,
		UserID:         userID,
		Username:       username,
		Role:           role,
		SessionVersion: sessionVersion,
		ExpiresAt:      expiresAt,
	}
	payload, err := json.Marshal(record)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	if err := s.cache.Set(ctx, s.sessionKey(sessionID), payload, s.config.SessionTTL).Err(); err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	// 维护 user→sessions 反向索引，仅用于撤销后的存储清理；安全语义由 session version 判定。
	userSessionsKey := s.userSessionsKey(userID)
	indexTTL := s.config.SessionTTL + time.Hour
	pipe := s.cache.Pipeline()
	pipe.SAdd(ctx, userSessionsKey, sessionID)
	pipe.Expire(ctx, userSessionsKey, indexTTL)
	if _, pipeErr := pipe.Exec(ctx); pipeErr != nil {
		// 索引维护失败不影响鉴权正确性，只会让撤销后的物理清理退化为延迟过期。
		_ = pipeErr
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
	if err := requireContext(ctx); err != nil {
		return nil, err
	}
	if sessionID == "" {
		return nil, authcontracts.ErrTokenInvalid
	}

	payload, err := s.cache.Get(ctx, s.sessionKey(sessionID)).Result()
	if errors.Is(err, redislib.Nil) {
		return nil, authcontracts.ErrAccessTokenExpired
	}
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	var record sessionRecord
	if err := json.Unmarshal([]byte(payload), &record); err != nil {
		return nil, authcontracts.ErrTokenInvalid.WithCause(err)
	}
	if record.ID == "" || record.UserID <= 0 || record.Username == "" || record.Role == "" || record.ExpiresAt.IsZero() {
		return nil, authcontracts.ErrTokenInvalid
	}
	if !record.ExpiresAt.After(time.Now().UTC()) {
		_ = s.cache.Del(ctx, s.sessionKey(sessionID)).Err()
		return nil, authcontracts.ErrAccessTokenExpired
	}
	currentVersion, err := s.getUserSessionVersion(ctx, record.UserID)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	if record.SessionVersion != currentVersion {
		_ = s.cache.Del(ctx, s.sessionKey(sessionID)).Err()
		if record.UserID > 0 {
			_ = s.cache.SRem(ctx, s.userSessionsKey(record.UserID), sessionID).Err()
		}
		return nil, authcontracts.ErrAccessTokenExpired
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
	if err := requireContext(ctx); err != nil {
		return err
	}
	if sessionID == "" {
		return nil
	}

	// 尝试读取会话记录以获取 userID，用于清理反向索引
	payload, err := s.cache.Get(ctx, s.sessionKey(sessionID)).Result()
	if err != nil && !errors.Is(err, redislib.Nil) {
		return err
	}

	pipe := s.cache.Pipeline()
	pipe.Del(ctx, s.sessionKey(sessionID))
	if payload != "" {
		var record sessionRecord
		if json.Unmarshal([]byte(payload), &record) == nil && record.UserID > 0 {
			pipe.SRem(ctx, s.userSessionsKey(record.UserID), sessionID)
		}
	}
	_, err = pipe.Exec(ctx)
	return err
}

func (s *tokenService) RevokeAllUserSessions(ctx context.Context, userID int64) error {
	if err := requireContext(ctx); err != nil {
		return err
	}
	if userID <= 0 {
		return nil
	}

	if _, err := s.cache.Incr(ctx, s.userSessionVersionKey(userID)).Result(); err != nil {
		return err
	}

	userSessionsKey := s.userSessionsKey(userID)
	sessionIDs, err := s.cache.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		return nil
	}
	if len(sessionIDs) == 0 {
		// 即使没有活跃会话，也清理可能残留的空集合
		_ = s.cache.Del(ctx, userSessionsKey).Err()
		return nil
	}

	keys := make([]string, 0, len(sessionIDs)+1)
	for _, sid := range sessionIDs {
		keys = append(keys, s.sessionKey(sid))
	}
	keys = append(keys, userSessionsKey)

	_ = s.cache.Del(ctx, keys...).Err()
	return nil
}

func (s *tokenService) ListUserSessions(ctx context.Context, userID int64) ([]authcontracts.Session, error) {
	if err := requireContext(ctx); err != nil {
		return nil, err
	}
	if userID <= 0 {
		return nil, nil
	}

	userSessionsKey := s.userSessionsKey(userID)
	sessionIDs, err := s.cache.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		return nil, fmt.Errorf("list user sessions: SMembers: %w", err)
	}
	if len(sessionIDs) == 0 {
		return nil, nil
	}

	sessionKeys := make([]string, 0, len(sessionIDs))
	for _, sid := range sessionIDs {
		sessionKeys = append(sessionKeys, s.sessionKey(sid))
	}

	raw, err := s.cache.MGet(ctx, sessionKeys...).Result()
	if err != nil {
		return nil, fmt.Errorf("list user sessions: MGet: %w", err)
	}

	currentVersion, versionErr := s.getUserSessionVersion(ctx, userID)
	if versionErr != nil {
		return nil, fmt.Errorf("list user sessions: version: %w", versionErr)
	}

	now := time.Now().UTC()
	sessions := make([]authcontracts.Session, 0, len(raw))
	for _, val := range raw {
		if val == nil {
			continue
		}
		str, ok := val.(string)
		if !ok {
			continue
		}
		var record sessionRecord
		if err := json.Unmarshal([]byte(str), &record); err != nil {
			continue
		}
		if !record.ExpiresAt.After(now) {
			continue
		}
		if record.SessionVersion != currentVersion {
			continue
		}
		sessions = append(sessions, authcontracts.Session{
			ID:        record.ID,
			UserID:    record.UserID,
			Username:  record.Username,
			Role:      record.Role,
			ExpiresAt: record.ExpiresAt,
		})
	}

	return sessions, nil
}

func (s *tokenService) IssueWSTicket(ctx context.Context, user authctx.CurrentUser) (*authcontracts.WSTicket, error) {
	if err := requireContext(ctx); err != nil {
		return nil, err
	}
	ticket, err := generateOpaqueToken(32)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	payload, err := json.Marshal(wsTicketPayload{
		UserID:   user.UserID,
		Username: user.Username,
		Role:     user.Role,
		IssuedAt: time.Now().UTC(),
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	expiresAt := time.Now().Add(s.wsConfig.TicketTTL).UTC()
	if err := s.cache.Set(ctx, s.wsTicketKey(ticket), payload, s.wsConfig.TicketTTL).Err(); err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	return &authcontracts.WSTicket{
		Ticket:    ticket,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *tokenService) ConsumeWSTicket(ctx context.Context, ticket string) (*authctx.CurrentUser, error) {
	if err := requireContext(ctx); err != nil {
		return nil, err
	}
	if ticket == "" {
		return nil, authcontracts.ErrWSTicketInvalid
	}

	payload, err := s.cache.GetDel(ctx, s.wsTicketKey(ticket)).Result()
	if errors.Is(err, redislib.Nil) {
		return nil, authcontracts.ErrWSTicketInvalid
	}
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	var claims wsTicketPayload
	if err := json.Unmarshal([]byte(payload), &claims); err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	if claims.UserID <= 0 || claims.Username == "" || claims.Role == "" {
		return nil, authcontracts.ErrWSTicketInvalid
	}

	return &authctx.CurrentUser{
		UserID:   claims.UserID,
		Username: claims.Username,
		Role:     claims.Role,
	}, nil
}

func (s *tokenService) IssueMCPToken(ctx context.Context, user authctx.CurrentUser) (*authcontracts.MCPToken, error) {
	if err := requireContext(ctx); err != nil {
		return nil, err
	}
	if user.UserID <= 0 || user.Username == "" || user.Role == "" {
		return nil, authcontracts.ErrMCPTokenInvalid
	}

	token, err := generateOpaqueToken(32)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	sessionVersion, err := s.getUserSessionVersion(ctx, user.UserID)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	now := time.Now().UTC()
	expiresAt := now.Add(s.config.MCPTokenTTL).UTC()
	payload, err := json.Marshal(mcpTokenPayload{
		UserID:         user.UserID,
		Username:       user.Username,
		Role:           user.Role,
		SessionVersion: sessionVersion,
		IssuedAt:       now,
		ExpiresAt:      expiresAt,
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	if err := s.cache.Set(ctx, s.mcpTokenKey(token), payload, s.config.MCPTokenTTL).Err(); err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	return &authcontracts.MCPToken{
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *tokenService) ResolveMCPToken(ctx context.Context, token string) (*authctx.CurrentUser, error) {
	if err := requireContext(ctx); err != nil {
		return nil, err
	}
	if token == "" {
		return nil, authcontracts.ErrMCPTokenInvalid
	}

	key := s.mcpTokenKey(token)
	payload, err := s.cache.Get(ctx, key).Result()
	if errors.Is(err, redislib.Nil) {
		return nil, authcontracts.ErrMCPTokenInvalid
	}
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	var claims mcpTokenPayload
	if err := json.Unmarshal([]byte(payload), &claims); err != nil {
		return nil, authcontracts.ErrMCPTokenInvalid.WithCause(err)
	}
	if claims.UserID <= 0 || claims.Username == "" || claims.Role == "" || claims.ExpiresAt.IsZero() {
		return nil, authcontracts.ErrMCPTokenInvalid
	}
	if !claims.ExpiresAt.After(time.Now().UTC()) {
		_ = s.cache.Del(ctx, key).Err()
		return nil, authcontracts.ErrMCPTokenInvalid
	}
	currentVersion, err := s.getUserSessionVersion(ctx, claims.UserID)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	if claims.SessionVersion != currentVersion {
		_ = s.cache.Del(ctx, key).Err()
		return nil, authcontracts.ErrMCPTokenInvalid
	}

	return &authctx.CurrentUser{
		UserID:    claims.UserID,
		Username:  claims.Username,
		Role:      claims.Role,
		ExpiresAt: claims.ExpiresAt,
	}, nil
}

func (s *tokenService) sessionKey(sessionID string) string {
	return fmt.Sprintf("%s:%s", s.config.SessionKeyPrefix, sessionID)
}

func (s *tokenService) userSessionsKey(userID int64) string {
	return fmt.Sprintf("%s:user_sessions:%d", s.config.SessionKeyPrefix, userID)
}

func (s *tokenService) userSessionVersionKey(userID int64) string {
	return fmt.Sprintf("%s:user_session_version:%d", s.config.SessionKeyPrefix, userID)
}

func (s *tokenService) getUserSessionVersion(ctx context.Context, userID int64) (int64, error) {
	if userID <= 0 {
		return 0, nil
	}
	version, err := s.cache.Get(ctx, s.userSessionVersionKey(userID)).Int64()
	if errors.Is(err, redislib.Nil) {
		return 0, nil
	}
	return version, err
}

func requireContext(ctx context.Context) error {
	if ctx == nil {
		return apperror.ErrInternal.WithCause(fmt.Errorf("context is required"))
	}
	return nil
}

func (s *tokenService) wsTicketKey(ticket string) string {
	return fmt.Sprintf("%s:%s", s.wsConfig.TicketKeyPrefix, ticket)
}

func (s *tokenService) mcpTokenKey(token string) string {
	return fmt.Sprintf("%s:mcp:%s", s.config.SessionKeyPrefix, token)
}

func generateOpaqueToken(size int) (string, error) {
	buffer := make([]byte, size)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}
