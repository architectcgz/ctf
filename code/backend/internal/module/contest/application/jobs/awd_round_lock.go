package jobs

import (
	"context"

	rediskeys "ctf-platform/internal/pkg/redis"
)

func (u *AWDRoundUpdater) acquireRoundLock(ctx context.Context, contestID int64, roundNumber int) (bool, error) {
	if u.redis == nil {
		return true, nil
	}
	return u.redis.SetNX(ctx, rediskeys.AWDRoundLockKey(contestID, roundNumber), "1", u.cfg.RoundLockTTL).Result()
}
