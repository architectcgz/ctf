package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	redislib "github.com/redis/go-redis/v9"

	challengeports "ctf-platform/internal/module/challenge/ports"
	cachekeys "ctf-platform/pkg/cache"
)

var _ challengeports.ChallengeSolvedCountCache = (*SolvedCountCache)(nil)

type SolvedCountCache struct {
	cache *redislib.Client
}

func NewSolvedCountCache(cache *redislib.Client) *SolvedCountCache {
	if cache == nil {
		return nil
	}
	return &SolvedCountCache{cache: cache}
}

func (c *SolvedCountCache) GetSolvedCount(ctx context.Context, challengeID int64) (int64, bool, error) {
	if c == nil || c.cache == nil || challengeID <= 0 {
		return 0, false, nil
	}

	payload, err := c.cache.Get(ctx, cachekeys.ChallengeSolvedCountKey(challengeID)).Result()
	if errors.Is(err, redislib.Nil) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}

	var count int64
	if err := json.Unmarshal([]byte(payload), &count); err != nil {
		return 0, false, nil
	}
	return count, true, nil
}

func (c *SolvedCountCache) StoreSolvedCount(ctx context.Context, challengeID int64, count int64, ttl time.Duration) error {
	if c == nil || c.cache == nil || challengeID <= 0 || ttl <= 0 {
		return nil
	}

	payload, err := json.Marshal(count)
	if err != nil {
		return err
	}
	return c.cache.Set(ctx, cachekeys.ChallengeSolvedCountKey(challengeID), payload, ttl).Err()
}
