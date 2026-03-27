package jobs

import (
	"context"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	rediskeys "ctf-platform/internal/pkg/redis"
)

func (u *AWDRoundUpdater) syncRoundServiceChecks(ctx context.Context, contest *model.Contest, activeRound int) error {
	if contest == nil {
		return nil
	}
	if activeRound <= 0 {
		if u.redis == nil {
			return nil
		}
		return u.redis.Del(ctx, rediskeys.AWDServiceStatusKey(contest.ID)).Err()
	}

	round, err := u.findRoundByNumber(ctx, contest.ID, activeRound)
	if err != nil {
		return err
	}
	return u.runRoundServiceChecks(ctx, contest, round, contestdomain.AWDCheckSourceScheduler)
}

// RunRoundServiceChecks 允许后台运维链路手动触发轮次服务检查，并记录巡检来源。
func (u *AWDRoundUpdater) RunRoundServiceChecks(ctx context.Context, contest *model.Contest, round *model.AWDRound, source string) error {
	if contest == nil || round == nil {
		return nil
	}
	return u.runRoundServiceChecks(ctx, contest, round, source)
}
