package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	redislib "github.com/redis/go-redis/v9"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
	rediskeys "ctf-platform/internal/pkg/redis"
	"ctf-platform/internal/pkg/redislock"
)

var _ contestports.AWDRoundStateStore = (*AWDRoundStateStore)(nil)

type AWDRoundStateStore struct {
	cache *redislib.Client
}

func NewAWDRoundStateStore(cache *redislib.Client) *AWDRoundStateStore {
	if cache == nil {
		return nil
	}
	return &AWDRoundStateStore{cache: cache}
}

func (s *AWDRoundStateStore) AcquireAWDSchedulerLock(ctx context.Context, ttl time.Duration) (contestports.ContestSchedulerLockLease, bool, error) {
	if s == nil || s.cache == nil || ttl <= 0 {
		return nil, true, nil
	}
	return redislock.Acquire(ctx, s.cache, rediskeys.AWDSchedulerLockKey(), ttl)
}

func (s *AWDRoundStateStore) TryAcquireAWDRoundLock(ctx context.Context, contestID int64, roundNumber int, ttl time.Duration) (bool, error) {
	if s == nil || s.cache == nil {
		return true, nil
	}
	return s.cache.SetNX(ctx, rediskeys.AWDRoundLockKey(contestID, roundNumber), "1", ttl).Result()
}

func (s *AWDRoundStateStore) IsAWDCurrentRound(ctx context.Context, contestID int64, roundNumber int) (bool, error) {
	if s == nil || s.cache == nil {
		return false, nil
	}
	currentRound, err := s.cache.Get(ctx, rediskeys.AWDCurrentRoundKey(contestID)).Result()
	if err == nil {
		return strings.TrimSpace(currentRound) == fmt.Sprintf("%d", roundNumber), nil
	}
	if errors.Is(err, redislib.Nil) {
		return false, nil
	}
	return false, err
}

func (s *AWDRoundStateStore) LoadAWDCurrentRoundNumber(ctx context.Context, contestID int64) (int, bool, error) {
	if s == nil || s.cache == nil {
		return 0, false, nil
	}
	currentRound, err := s.cache.Get(ctx, rediskeys.AWDCurrentRoundKey(contestID)).Result()
	if err == nil {
		roundNumber, convErr := strconv.Atoi(strings.TrimSpace(currentRound))
		if convErr != nil || roundNumber <= 0 {
			return 0, false, nil
		}
		return roundNumber, true, nil
	}
	if errors.Is(err, redislib.Nil) {
		return 0, false, nil
	}
	return 0, false, err
}

func (s *AWDRoundStateStore) LoadAWDRoundFlag(ctx context.Context, contestID, roundID, teamID, serviceID int64) (string, bool, error) {
	if s == nil || s.cache == nil {
		return "", false, nil
	}
	flag, err := s.cache.HGet(ctx, rediskeys.AWDRoundFlagsKey(contestID, roundID), rediskeys.AWDRoundFlagServiceField(teamID, serviceID)).Result()
	if err == nil {
		flag = strings.TrimSpace(flag)
		if flag == "" {
			return "", false, nil
		}
		return flag, true, nil
	}
	if errors.Is(err, redislib.Nil) {
		return "", false, nil
	}
	return "", false, err
}

func (s *AWDRoundStateStore) SyncAWDCurrentRoundState(ctx context.Context, contestID int64, round *model.AWDRound, assignments []contestports.AWDFlagAssignment, ttl time.Duration) error {
	if s == nil || s.cache == nil || round == nil {
		return nil
	}

	if len(assignments) == 0 {
		return s.cache.Set(ctx, rediskeys.AWDCurrentRoundKey(contestID), round.RoundNumber, 0).Err()
	}

	fields := make(map[string]any, len(assignments))
	for _, item := range assignments {
		fields[rediskeys.AWDRoundFlagServiceField(item.TeamID, item.ServiceID)] = item.Flag
	}

	pipe := s.cache.TxPipeline()
	pipe.Set(ctx, rediskeys.AWDCurrentRoundKey(contestID), round.RoundNumber, 0)
	roundKey := rediskeys.AWDRoundFlagsKey(contestID, round.ID)
	pipe.Del(ctx, roundKey)
	pipe.HSet(ctx, roundKey, fields)
	if ttl > 0 {
		pipe.Expire(ctx, roundKey, ttl)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}
	return nil
}

func (s *AWDRoundStateStore) ClearAWDCurrentRoundState(ctx context.Context, contestID int64) error {
	if s == nil || s.cache == nil {
		return nil
	}
	return s.cache.Del(ctx, rediskeys.AWDCurrentRoundKey(contestID)).Err()
}

func (s *AWDRoundStateStore) SetAWDServiceStatus(ctx context.Context, contestID, teamID, serviceID int64, status string) error {
	if s == nil || s.cache == nil || contestID <= 0 || teamID <= 0 || serviceID <= 0 {
		return nil
	}
	return s.cache.HSet(
		ctx,
		rediskeys.AWDServiceStatusKey(contestID),
		rediskeys.AWDRoundFlagServiceField(teamID, serviceID),
		status,
	).Err()
}

func (s *AWDRoundStateStore) ReplaceAWDServiceStatus(ctx context.Context, contestID int64, entries []contestports.AWDServiceStatusEntry) error {
	if s == nil || s.cache == nil {
		return nil
	}

	statusKey := rediskeys.AWDServiceStatusKey(contestID)
	pipe := s.cache.TxPipeline()
	pipe.Del(ctx, statusKey)
	if len(entries) > 0 {
		fields := make(map[string]any, len(entries))
		for _, entry := range entries {
			fields[rediskeys.AWDRoundFlagServiceField(entry.TeamID, entry.ServiceID)] = entry.Status
		}
		pipe.HSet(ctx, statusKey, fields)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}
	return nil
}

func (s *AWDRoundStateStore) ClearAWDServiceStatus(ctx context.Context, contestID int64) error {
	if s == nil || s.cache == nil {
		return nil
	}
	return s.cache.Del(ctx, rediskeys.AWDServiceStatusKey(contestID)).Err()
}
