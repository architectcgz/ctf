package jobs

import (
	"context"
	"errors"
	"gorm.io/gorm"

	"ctf-platform/internal/model"
)

func (u *AWDRoundUpdater) shouldSyncLiveServiceStatusCache(ctx context.Context, contestID int64, round *model.AWDRound) (bool, error) {
	if u.stateStore == nil || u.repo == nil || contestID <= 0 || round == nil {
		return false, nil
	}

	currentRound, err := u.repo.FindRunningRound(ctx, contestID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return false, err
		}
		return u.stateStore.IsAWDCurrentRound(ctx, contestID, round.RoundNumber)
	}
	return currentRound.ID == round.ID, nil
}
