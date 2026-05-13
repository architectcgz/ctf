package jobs

import (
	"context"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/model"
)

func (u *AWDRoundUpdater) syncRoundFlags(ctx context.Context, contest *model.Contest, activeRound int, now time.Time) error {
	if contest == nil || u.stateStore == nil {
		return nil
	}
	if activeRound <= 0 {
		return u.stateStore.ClearAWDCurrentRoundState(ctx, contest.ID)
	}
	if u.flagSecret == "" {
		u.log.Warn("skip_awd_flag_rotation_due_to_empty_secret", zap.Int64("contest_id", contest.ID))
		return nil
	}

	round, err := u.findRoundByNumber(ctx, contest.ID, activeRound)
	if err != nil {
		return err
	}
	assignments, err := u.buildRoundFlagAssignments(ctx, contest.ID, round)
	if err != nil {
		return err
	}
	if err := u.stateStore.SyncAWDCurrentRoundState(ctx, contest.ID, round, assignments, u.currentRoundTTL(contest, round, now)); err != nil {
		return err
	}

	return u.injector.InjectRoundFlags(ctx, contest, round, assignments)
}
