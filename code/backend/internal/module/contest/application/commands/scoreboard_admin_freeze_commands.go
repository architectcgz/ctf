package commands

import (
	"context"
	"time"

	redislib "github.com/redis/go-redis/v9"

	"ctf-platform/internal/model"
	"ctf-platform/internal/module/contest/domain"
	rediskeys "ctf-platform/internal/pkg/redis"
	"ctf-platform/pkg/errcode"
)

func (s *ScoreboardAdminService) FreezeScoreboard(ctx context.Context, contestID int64, minutesBeforeEnd int) error {
	contest, err := s.repo.FindByID(ctx, contestID)
	if err != nil {
		if err == domain.ErrContestNotFound {
			return errcode.ErrContestNotFound
		}
		return errcode.ErrInternal.WithCause(err)
	}

	now := time.Now()
	if !now.Before(contest.EndTime) {
		return errcode.ErrContestEnded
	}

	freezeTime := contest.EndTime.Add(-time.Duration(minutesBeforeEnd) * time.Minute)
	contest.FreezeTime = &freezeTime
	if !now.Before(freezeTime) {
		contest.Status = model.ContestStatusFrozen
		if err := s.createSnapshotFromLive(ctx, contestID); err != nil {
			return err
		}
	}

	return s.repo.Update(ctx, contest)
}

func (s *ScoreboardAdminService) UnfreezeScoreboard(ctx context.Context, contestID int64) error {
	contest, err := s.repo.FindByID(ctx, contestID)
	if err != nil {
		if err == domain.ErrContestNotFound {
			return errcode.ErrContestNotFound
		}
		return errcode.ErrInternal.WithCause(err)
	}

	if contest.FreezeTime == nil && contest.Status != model.ContestStatusFrozen {
		return errcode.ErrScoreboardNotFrozen
	}

	contest.FreezeTime = nil
	if contest.Status == model.ContestStatusFrozen && time.Now().Before(contest.EndTime) {
		contest.Status = model.ContestStatusRunning
	}
	if err := s.redis.Del(ctx, rediskeys.RankContestFrozenKey(contestID)).Err(); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}

	return s.repo.Update(ctx, contest)
}

func (s *ScoreboardAdminService) createSnapshotFromLive(ctx context.Context, contestID int64) error {
	srcKey := rediskeys.RankContestTeamKey(contestID)
	dstKey := rediskeys.RankContestFrozenKey(contestID)
	if err := s.redis.ZUnionStore(ctx, dstKey, &redislib.ZStore{
		Keys:    []string{srcKey},
		Weights: []float64{1},
	}).Err(); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	return nil
}
