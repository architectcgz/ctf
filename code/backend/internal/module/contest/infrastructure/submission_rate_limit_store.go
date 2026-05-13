package infrastructure

import (
	"context"
	"fmt"
	"strings"
	"time"

	redislib "github.com/redis/go-redis/v9"

	contestports "ctf-platform/internal/module/contest/ports"
)

const defaultContestSubmissionRateLimitPrefix = "ctf:ratelimit"

var _ contestports.ContestSubmissionRateLimitStore = (*ContestSubmissionRateLimitStore)(nil)

type ContestSubmissionRateLimitStore struct {
	cache  *redislib.Client
	prefix string
}

func NewContestSubmissionRateLimitStore(cache *redislib.Client, prefix string) *ContestSubmissionRateLimitStore {
	if cache == nil {
		return nil
	}
	return &ContestSubmissionRateLimitStore{
		cache:  cache,
		prefix: prefix,
	}
}

func (s *ContestSubmissionRateLimitStore) HasIncorrectSubmissionRateLimit(ctx context.Context, userID, contestID, challengeID int64) (bool, error) {
	if s == nil || s.cache == nil {
		return false, fmt.Errorf("contest submission rate limit store unavailable")
	}
	if userID <= 0 || contestID <= 0 || challengeID <= 0 {
		return false, nil
	}
	exists, err := s.cache.Exists(ctx, contestSubmissionRateLimitKey(s.prefix, userID, contestID, challengeID)).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func (s *ContestSubmissionRateLimitStore) SetIncorrectSubmissionRateLimit(ctx context.Context, userID, contestID, challengeID int64, ttl time.Duration) error {
	if s == nil || s.cache == nil {
		return fmt.Errorf("contest submission rate limit store unavailable")
	}
	if userID <= 0 || contestID <= 0 || challengeID <= 0 {
		return nil
	}
	return s.cache.Set(ctx, contestSubmissionRateLimitKey(s.prefix, userID, contestID, challengeID), "1", ttl).Err()
}

func contestSubmissionRateLimitKey(prefix string, userID, contestID, challengeID int64) string {
	trimmedPrefix := strings.TrimSpace(prefix)
	if trimmedPrefix == "" {
		trimmedPrefix = defaultContestSubmissionRateLimitPrefix
	}
	return fmt.Sprintf("%s:contest:submit:rate:%d:%d:%d", trimmedPrefix, userID, contestID, challengeID)
}
