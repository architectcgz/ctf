package infrastructure

import (
	"context"
	"strings"
	"time"

	redislib "github.com/redis/go-redis/v9"

	rediskeys "ctf-platform/internal/pkg/redis"
)

type PlatformRuntimeStateStore struct {
	cache *redislib.Client
}

func NewPlatformRuntimeStateStore(cache *redislib.Client) *PlatformRuntimeStateStore {
	if cache == nil {
		return nil
	}
	return &PlatformRuntimeStateStore{cache: cache}
}

func (s *PlatformRuntimeStateStore) LoadPlatformRuntimeState(ctx context.Context) (string, time.Time, bool, error) {
	if s == nil || s.cache == nil {
		return "", time.Time{}, false, nil
	}
	values, err := s.cache.HGetAll(ctx, rediskeys.PlatformRuntimeStateKey()).Result()
	if err != nil {
		return "", time.Time{}, false, err
	}
	if len(values) == 0 {
		return "", time.Time{}, false, nil
	}
	bootID := strings.TrimSpace(values["boot_id"])
	heartbeatRaw := strings.TrimSpace(values["last_heartbeat_at"])
	if bootID == "" || heartbeatRaw == "" {
		return "", time.Time{}, false, nil
	}
	heartbeatAt, err := time.Parse(time.RFC3339Nano, heartbeatRaw)
	if err != nil {
		return "", time.Time{}, false, err
	}
	return bootID, heartbeatAt.UTC(), true, nil
}

func (s *PlatformRuntimeStateStore) SavePlatformRuntimeState(ctx context.Context, bootID string, heartbeatAt time.Time) error {
	if s == nil || s.cache == nil || strings.TrimSpace(bootID) == "" || heartbeatAt.IsZero() {
		return nil
	}
	return s.cache.HSet(ctx, rediskeys.PlatformRuntimeStateKey(), map[string]any{
		"boot_id":           strings.TrimSpace(bootID),
		"last_heartbeat_at": heartbeatAt.UTC().Format(time.RFC3339Nano),
	}).Err()
}
