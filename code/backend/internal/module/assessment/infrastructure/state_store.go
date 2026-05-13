package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	redislib "github.com/redis/go-redis/v9"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	assessmentdomain "ctf-platform/internal/module/assessment/domain"
	assessmentports "ctf-platform/internal/module/assessment/ports"
	rediskeys "ctf-platform/internal/pkg/redis"
)

const assessmentProfileLockReleaseScript = `
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("del", KEYS[1])
else
	return 0
end
`

var _ assessmentports.AssessmentProfileLockStore = (*ProfileLockStore)(nil)
var _ assessmentports.AssessmentRecommendationCacheStore = (*RecommendationCacheStore)(nil)

type ProfileLockStore struct {
	cache  *redislib.Client
	prefix string
}

type profileLockLease struct {
	cache *redislib.Client
	key   string
	token string
}

type RecommendationCacheStore struct {
	cache *redislib.Client
}

var _ assessmentports.AssessmentProfileLockLease = (*profileLockLease)(nil)

func NewProfileLockStore(cache *redislib.Client, cfg config.AssessmentConfig) *ProfileLockStore {
	if cache == nil {
		return nil
	}
	normalized := assessmentdomain.NormalizeAssessmentConfig(cfg)
	return &ProfileLockStore{
		cache:  cache,
		prefix: normalized.RedisKeyPrefix,
	}
}

func NewRecommendationCacheStore(cache *redislib.Client) *RecommendationCacheStore {
	if cache == nil {
		return nil
	}
	return &RecommendationCacheStore{cache: cache}
}

func (s *ProfileLockStore) AcquireDimensionUpdateLock(ctx context.Context, userID int64, dimension string, ttl time.Duration) (assessmentports.AssessmentProfileLockLease, bool, error) {
	return s.acquire(ctx, assessmentDimensionUpdateLockKey(s.prefix, userID, dimension), ttl)
}

func (s *ProfileLockStore) AcquireFullProfileRebuildLock(ctx context.Context, userID int64, ttl time.Duration) (assessmentports.AssessmentProfileLockLease, bool, error) {
	return s.acquire(ctx, assessmentFullProfileRebuildLockKey(s.prefix, userID), ttl)
}

func (s *ProfileLockStore) acquire(ctx context.Context, key string, ttl time.Duration) (assessmentports.AssessmentProfileLockLease, bool, error) {
	if s == nil || s.cache == nil || key == "" {
		return nil, true, nil
	}

	lock := &profileLockLease{
		cache: s.cache,
		key:   key,
		token: uuid.NewString(),
	}

	acquired, err := s.cache.SetNX(ctx, key, lock.token, ttl).Result()
	if err != nil {
		return nil, false, fmt.Errorf("acquire assessment profile lock: %w", err)
	}
	if !acquired {
		return nil, false, nil
	}
	return lock, true, nil
}

func (l *profileLockLease) Release(ctx context.Context) (bool, error) {
	if l == nil || l.cache == nil || l.key == "" || l.token == "" {
		return false, nil
	}

	result, err := l.cache.Eval(ctx, assessmentProfileLockReleaseScript, []string{l.key}, l.token).Result()
	if err != nil {
		return false, fmt.Errorf("release assessment profile lock: %w", err)
	}

	released, ok := result.(int64)
	return ok && released > 0, nil
}

func (s *RecommendationCacheStore) LoadRecommendations(ctx context.Context, userID int64) ([]*dto.ChallengeRecommendation, bool, error) {
	if s == nil || s.cache == nil || userID <= 0 {
		return nil, false, nil
	}

	cached, err := s.cache.Get(ctx, rediskeys.RecommendationKey(userID)).Result()
	if err != nil {
		if errors.Is(err, redislib.Nil) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("load recommendation cache: %w", err)
	}

	var recommendations []*dto.ChallengeRecommendation
	if err := json.Unmarshal([]byte(cached), &recommendations); err != nil {
		return nil, false, fmt.Errorf("decode recommendation cache: %w", err)
	}
	return recommendations, true, nil
}

func (s *RecommendationCacheStore) StoreRecommendations(ctx context.Context, userID int64, recommendations []*dto.ChallengeRecommendation, ttl time.Duration) error {
	if s == nil || s.cache == nil || userID <= 0 {
		return nil
	}

	data, err := json.Marshal(recommendations)
	if err != nil {
		return fmt.Errorf("encode recommendation cache: %w", err)
	}
	if err := s.cache.Set(ctx, rediskeys.RecommendationKey(userID), data, ttl).Err(); err != nil {
		return fmt.Errorf("store recommendation cache: %w", err)
	}
	return nil
}

func (s *RecommendationCacheStore) DeleteRecommendations(ctx context.Context, userID int64) error {
	if s == nil || s.cache == nil || userID <= 0 {
		return nil
	}
	if err := s.cache.Del(ctx, rediskeys.RecommendationKey(userID)).Err(); err != nil {
		return fmt.Errorf("delete recommendation cache: %w", err)
	}
	return nil
}

func assessmentDimensionUpdateLockKey(prefix string, userID int64, dimension string) string {
	return fmt.Sprintf("%s:lock:%d:%s", prefix, userID, dimension)
}

func assessmentFullProfileRebuildLockKey(prefix string, userID int64) string {
	return fmt.Sprintf("%s:lock:%d", prefix, userID)
}
