package jobs

import "context"

func (u *AWDRoundUpdater) acquireRoundLock(ctx context.Context, contestID int64, roundNumber int) (bool, error) {
	if u.stateStore == nil {
		return true, nil
	}
	return u.stateStore.TryAcquireAWDRoundLock(ctx, contestID, roundNumber, u.cfg.RoundLockTTL)
}
