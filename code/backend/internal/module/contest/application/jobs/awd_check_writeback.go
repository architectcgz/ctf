package jobs

import (
	"context"

	contestentity "ctf-platform/internal/module/contest/entity"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (u *AWDRoundUpdater) persistRoundServiceChecks(ctx context.Context, contest *contestentity.Contest, round *contestentity.AWDRound, records []contestentity.AWDTeamService, statusEntries []contestports.AWDServiceStatusEntry) error {
	if len(records) > 0 {
		if err := u.repo.WithinRoundServiceWritebackTransaction(ctx, func(txRepo contestports.AWDRoundServiceWritebackTxRepository) error {
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

	if u.stateStore != nil && shouldSyncLiveStatusCache {
		if err := u.stateStore.ReplaceAWDServiceStatus(ctx, contest.ID, statusEntries); err != nil {
			return err
		}
	}
	if u.scoreboardCache != nil {
		if err := u.scoreboardCache.RebuildContestScoreboard(ctx, contest.ID); err != nil {
			return err
		}
	}

	return nil
}
