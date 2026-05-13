package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ctf-platform/internal/config"
	opsports "ctf-platform/internal/module/ops/ports"
)

const defaultDashboardSessionPrefix = "ctf:auth:session"

var _ opsports.DashboardStateStore = (*DashboardStateStore)(nil)

type DashboardStateStore struct {
	cache          *redislib.Client
	cacheKey       string
	cacheTTL       time.Duration
	sessionPattern string
	logger         *zap.Logger
}

type dashboardSessionRecord struct {
	UserID int64 `json:"user_id"`
}

func NewDashboardStateStore(cache *redislib.Client, cfg *config.Config, logger *zap.Logger) *DashboardStateStore {
	if cache == nil || cfg == nil {
		return nil
	}
	if logger == nil {
		logger = zap.NewNop()
	}

	sessionPrefix := cfg.Auth.SessionKeyPrefix
	if strings.TrimSpace(sessionPrefix) == "" {
		sessionPrefix = defaultDashboardSessionPrefix
	}

	return &DashboardStateStore{
		cache:          cache,
		cacheKey:       fmt.Sprintf("%s:stats", cfg.Dashboard.RedisKeyPrefix),
		cacheTTL:       cfg.Dashboard.CacheTTL,
		sessionPattern: sessionPrefix + ":*",
		logger:         logger,
	}
}

func (s *DashboardStateStore) LoadDashboardStats(ctx context.Context) (*opsports.DashboardStatsSnapshot, error) {
	if s == nil || s.cache == nil {
		return nil, nil
	}

	data, err := s.cache.Get(ctx, s.cacheKey).Bytes()
	if err != nil {
		if errors.Is(err, redislib.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("load dashboard stats cache: %w", err)
	}

	var stats opsports.DashboardStatsSnapshot
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, fmt.Errorf("decode dashboard stats cache: %w", err)
	}
	return &stats, nil
}

func (s *DashboardStateStore) SaveDashboardStats(ctx context.Context, stats *opsports.DashboardStatsSnapshot) error {
	if s == nil || s.cache == nil || stats == nil {
		return nil
	}

	data, err := json.Marshal(stats)
	if err != nil {
		return fmt.Errorf("encode dashboard stats cache: %w", err)
	}
	if err := s.cache.Set(ctx, s.cacheKey, data, s.cacheTTL).Err(); err != nil {
		return fmt.Errorf("store dashboard stats cache: %w", err)
	}
	return nil
}

func (s *DashboardStateStore) CountOnlineUsers(ctx context.Context) (int64, error) {
	if s == nil || s.cache == nil {
		return 0, nil
	}

	var cursor uint64
	onlineUserIDs := make(map[int64]struct{})

	for {
		keys, nextCursor, err := s.cache.Scan(ctx, cursor, s.sessionPattern, 100).Result()
		if err != nil {
			return 0, fmt.Errorf("scan online sessions: %w", err)
		}
		if len(keys) > 0 {
			values, err := s.cache.MGet(ctx, keys...).Result()
			if err != nil {
				return 0, fmt.Errorf("load online sessions: %w", err)
			}
			for index, value := range values {
				if value == nil {
					continue
				}

				payload, ok := dashboardPayloadString(value)
				if !ok {
					s.logger.Warn("忽略无法识别的在线会话记录", zap.String("key", keys[index]))
					continue
				}

				var session dashboardSessionRecord
				if err := json.Unmarshal([]byte(payload), &session); err != nil || session.UserID <= 0 {
					s.logger.Warn("忽略无效的在线会话记录", zap.String("key", keys[index]), zap.Error(err))
					continue
				}

				onlineUserIDs[session.UserID] = struct{}{}
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return int64(len(onlineUserIDs)), nil
}

func dashboardPayloadString(value any) (string, bool) {
	switch typed := value.(type) {
	case string:
		return typed, true
	case []byte:
		return string(typed), true
	default:
		return "", false
	}
}
