package infrastructure

import (
	"context"
	"fmt"
	"strings"
	"time"

	redislib "github.com/redis/go-redis/v9"

	practiceports "ctf-platform/internal/module/practice/ports"
)

const defaultPracticeSubmissionRateLimitPrefix = "ctf:ratelimit"

var _ practiceports.PracticeFlagSubmitRateLimitStore = (*PracticeFlagSubmitRateLimitStore)(nil)

type PracticeFlagSubmitRateLimitStore struct {
	cache  *redislib.Client
	prefix string
}

func NewFlagSubmitRateLimitStore(cache *redislib.Client, prefix string) *PracticeFlagSubmitRateLimitStore {
	if cache == nil {
		return nil
	}
	return &PracticeFlagSubmitRateLimitStore{
		cache:  cache,
		prefix: prefix,
	}
}

func (s *PracticeFlagSubmitRateLimitStore) AllowFlagSubmit(ctx context.Context, userID, challengeID int64, limit int, window time.Duration) (bool, error) {
	if s == nil || s.cache == nil {
		return true, nil
	}
	if userID <= 0 || challengeID <= 0 {
		return true, nil
	}

	count, err := s.cache.Incr(ctx, practiceFlagSubmitRateLimitKey(s.prefix, userID, challengeID)).Result()
	if err != nil {
		return false, fmt.Errorf("increment practice flag submit rate limit: %w", err)
	}
	if count == 1 {
		_ = s.cache.Expire(ctx, practiceFlagSubmitRateLimitKey(s.prefix, userID, challengeID), window).Err()
	}
	return count <= int64(limit), nil
}

func practiceFlagSubmitRateLimitKey(prefix string, userID, challengeID int64) string {
	trimmedPrefix := strings.TrimSpace(prefix)
	if trimmedPrefix == "" {
		trimmedPrefix = defaultPracticeSubmissionRateLimitPrefix
	}
	return fmt.Sprintf("%s:submit:%d:%d", trimmedPrefix, userID, challengeID)
}
