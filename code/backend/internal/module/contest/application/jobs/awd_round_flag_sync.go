package jobs

import (
	"context"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/model"
	rediskeys "ctf-platform/internal/pkg/redis"
)

func (u *AWDRoundUpdater) syncRoundFlags(ctx context.Context, contest *model.Contest, activeRound int, now time.Time) error {
	if contest == nil || u.redis == nil {
		return nil
	}
	if activeRound <= 0 {
		return u.redis.Del(ctx, rediskeys.AWDCurrentRoundKey(contest.ID)).Err()
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
	if len(assignments) == 0 {
		return u.redis.Set(ctx, rediskeys.AWDCurrentRoundKey(contest.ID), round.RoundNumber, 0).Err()
	}

	fields := make(map[string]any, len(assignments))
	for _, item := range assignments {
		fields[rediskeys.AWDRoundFlagField(item.TeamID, item.ChallengeID)] = item.Flag
	}

	pipe := u.redis.TxPipeline()
	pipe.Set(ctx, rediskeys.AWDCurrentRoundKey(contest.ID), round.RoundNumber, 0)
	roundKey := rediskeys.AWDRoundFlagsKey(contest.ID, round.ID)
	pipe.Del(ctx, roundKey)
	pipe.HSet(ctx, roundKey, fields)
	if ttl := u.currentRoundTTL(contest, round, now); ttl > 0 {
		pipe.Expire(ctx, roundKey, ttl)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	return u.injector.InjectRoundFlags(ctx, contest, round, assignments)
}
