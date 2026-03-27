package jobs

import (
	"context"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
	rediskeys "ctf-platform/internal/pkg/redis"
)

func (u *AWDRoundUpdater) persistRoundServiceChecks(ctx context.Context, contest *model.Contest, round *model.AWDRound, records []model.AWDTeamService, statusFields map[string]any) error {
	if len(records) > 0 {
		if err := u.repo.WithinTransaction(ctx, func(txRepo contestports.AWDRepository) error {
			if err := txRepo.UpsertTeamServices(ctx, records); err != nil {
				return err
			}
			return txRepo.RecalculateContestTeamScores(ctx, contest.ID)
		}); err != nil {
			return err
		}
	}

	shouldSyncLiveStatusCache, err := u.shouldSyncLiveServiceStatusCache(ctx, contest.ID, round)
	if err != nil {
		return err
	}

	if u.redis != nil && shouldSyncLiveStatusCache {
		pipe := u.redis.TxPipeline()
		statusKey := rediskeys.AWDServiceStatusKey(contest.ID)
		pipe.Del(ctx, statusKey)
		if len(statusFields) > 0 {
			pipe.HSet(ctx, statusKey, statusFields)
		}
		if _, err := pipe.Exec(ctx); err != nil {
			return err
		}
	}
	if u.redis != nil {
		if err := u.repo.RebuildContestScoreboardCache(ctx, u.redis, contest.ID); err != nil {
			return err
		}
	}

	return nil
}
