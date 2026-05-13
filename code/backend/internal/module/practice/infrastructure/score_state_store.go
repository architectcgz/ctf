package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	redislib "github.com/redis/go-redis/v9"

	"ctf-platform/internal/dto"
	practiceports "ctf-platform/internal/module/practice/ports"
	"ctf-platform/internal/pkg/cache"
)

const practiceScoreLockReleaseScript = `
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("del", KEYS[1])
else
	return 0
end
`

var _ practiceports.PracticeScoreStateStore = (*ScoreStateStore)(nil)

type ScoreStateStore struct {
	client *redislib.Client
}

type practiceScoreLockLease struct {
	client *redislib.Client
	key    string
	token  string
}

var _ practiceports.PracticeScoreLockLease = (*practiceScoreLockLease)(nil)

func NewScoreStateStore(client *redislib.Client) *ScoreStateStore {
	if client == nil {
		return nil
	}
	return &ScoreStateStore{client: client}
}

func (s *ScoreStateStore) AcquireUserScoreUpdateLock(ctx context.Context, userID int64, ttl time.Duration) (practiceports.PracticeScoreLockLease, bool, error) {
	if s == nil || s.client == nil {
		return nil, true, nil
	}

	lock := &practiceScoreLockLease{
		client: s.client,
		key:    cache.ScoreLockKey(userID),
		token:  uuid.NewString(),
	}

	acquired, err := s.client.SetNX(ctx, lock.key, lock.token, ttl).Result()
	if err != nil {
		return nil, false, fmt.Errorf("acquire user score lock: %w", err)
	}
	if !acquired {
		return nil, false, nil
	}
	return lock, true, nil
}

func (s *ScoreStateStore) LoadUserScoreCache(ctx context.Context, userID int64) (*dto.UserScoreInfo, bool, error) {
	if s == nil || s.client == nil {
		return nil, false, nil
	}

	cached, err := s.client.Get(ctx, cache.UserScoreKey(userID)).Result()
	if err != nil {
		if err == redislib.Nil {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("load user score cache: %w", err)
	}

	var info dto.UserScoreInfo
	if err := json.Unmarshal([]byte(cached), &info); err != nil {
		return nil, false, fmt.Errorf("decode user score cache: %w", err)
	}
	return &info, true, nil
}

func (s *ScoreStateStore) StoreUserScoreCache(ctx context.Context, info *dto.UserScoreInfo, ttl time.Duration) error {
	if s == nil || s.client == nil || info == nil {
		return nil
	}

	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("encode user score cache: %w", err)
	}
	if err := s.client.Set(ctx, cache.UserScoreKey(info.UserID), data, ttl).Err(); err != nil {
		return fmt.Errorf("store user score cache: %w", err)
	}
	return nil
}

func (s *ScoreStateStore) SyncUserScoreState(ctx context.Context, info *dto.UserScoreInfo, ttl time.Duration) error {
	if s == nil || s.client == nil || info == nil {
		return nil
	}

	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("encode user score state: %w", err)
	}

	pipe := s.client.Pipeline()
	pipe.Set(ctx, cache.UserScoreKey(info.UserID), data, ttl)
	pipe.ZAdd(ctx, cache.RankingKey(), redislib.Z{
		Score:  float64(info.TotalScore),
		Member: info.UserID,
	})
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("sync user score state: %w", err)
	}
	return nil
}

func (l *practiceScoreLockLease) Key() string {
	if l == nil {
		return ""
	}
	return l.key
}

func (l *practiceScoreLockLease) Release(ctx context.Context) (bool, error) {
	if l == nil || l.client == nil || l.key == "" || l.token == "" {
		return false, nil
	}

	result, err := l.client.Eval(ctx, practiceScoreLockReleaseScript, []string{l.key}, l.token).Result()
	if err != nil {
		return false, fmt.Errorf("release user score lock: %w", err)
	}

	released, ok := result.(int64)
	return ok && released > 0, nil
}
