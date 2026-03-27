package jobs

import (
	"context"
	"time"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ctf-platform/internal/model"
	rediskeys "ctf-platform/internal/pkg/redis"
)

func (u *StatusUpdater) calculateStatus(contest *model.Contest, now time.Time) string {
	if contest.Status == model.ContestStatusDraft {
		return model.ContestStatusDraft
	}

	if now.Before(contest.StartTime) {
		return model.ContestStatusRegistration
	}

	if !now.Before(contest.EndTime) {
		return model.ContestStatusEnded
	}

	if contest.FreezeTime != nil && !now.Before(*contest.FreezeTime) {
		return model.ContestStatusFrozen
	}

	return model.ContestStatusRunning
}

func (u *StatusUpdater) createFrozenSnapshot(ctx context.Context, contestID int64) {
	if u.redis == nil {
		return
	}
	srcKey := rediskeys.RankContestTeamKey(contestID)
	dstKey := rediskeys.RankContestFrozenKey(contestID)
	if err := u.redis.ZUnionStore(ctx, dstKey, &redislib.ZStore{
		Keys:    []string{srcKey},
		Weights: []float64{1},
	}).Err(); err != nil {
		u.log.Error("create_frozen_snapshot_failed", zap.Int64("contest_id", contestID), zap.Error(err))
	}
}

func (u *StatusUpdater) clearEndedContestRuntimeState(ctx context.Context, contestID int64) {
	if u.redis == nil || contestID <= 0 {
		return
	}
	if err := u.redis.Del(
		ctx,
		rediskeys.AWDCurrentRoundKey(contestID),
		rediskeys.AWDServiceStatusKey(contestID),
	).Err(); err != nil {
		u.log.Error("clear_ended_contest_runtime_state_failed", zap.Int64("contest_id", contestID), zap.Error(err))
	}
}
